#!/usr/bin/env bash
set -u

ARTIFACT_DIR="${1:-/tmp/debug-artifacts}"
mkdir -p "${ARTIFACT_DIR}/host" "${ARTIFACT_DIR}/vms" "${ARTIFACT_DIR}/probes" "${ARTIFACT_DIR}/pcaps"

MONITOR_LOG="${ARTIFACT_DIR}/monitor.log"
echo "=== Debug Monitor Started: $(date -u -Iseconds) ===" >> "$MONITOR_LOG"


EGRESS_IF=$(ip route show default 2>/dev/null | awk '{print $5}' | head -n 1)
echo "Egress interface: ${EGRESS_IF:-unknown}" >> "$MONITOR_LOG"

if [ -n "$EGRESS_IF" ]; then
    echo "Starting rolling tcpdump on egress interface: $EGRESS_IF" >> "$MONITOR_LOG"
    sudo tcpdump -U -Z root -i "$EGRESS_IF" -s 0 -C 50 -W 3 -w "${ARTIFACT_DIR}/pcaps/egress.pcap" 'tcp port 443 or tcp port 80 or tcp port 3128 or tcp port 54969' >/dev/null 2>&1 &
    echo $! > "${ARTIFACT_DIR}/tcpdump-egress.pid"
fi

lxdbr0_tcpdump_started=false


declare -A FIRST_SEEN_DOING
declare -A PROBED
declare -A CONFIGURED_DEBUG


while true; do
    TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")


    if [ "$lxdbr0_tcpdump_started" = false ] && ip link show lxdbr0 >/dev/null 2>&1; then
        echo "Starting rolling tcpdump on lxdbr0 at $TIMESTAMP" >> "$MONITOR_LOG"
        sudo tcpdump -U -Z root -i lxdbr0 -s 0 -C 50 -W 3 -w "${ARTIFACT_DIR}/pcaps/lxdbr0.pcap" 'tcp port 443 or tcp port 80 or tcp port 3128 or tcp port 54969' >/dev/null 2>&1 &
        echo $! > "${ARTIFACT_DIR}/tcpdump-lxdbr0.pid"
        lxdbr0_tcpdump_started=true
    fi


    {
        echo "=== HOST SNAPSHOT [$TIMESTAMP] ==="
        echo "--- ss -tnpi (port 443 / aproxy 54969 / squid 3128) ---"
        ss -tnpi '( sport = :443 or dport = :443 or sport = :54969 or dport = :54969 or dport = :3128 )' 2>&1 || true
        echo "--- conntrack count ---"
        sudo conntrack -C 2>&1 || true
        echo "--- aproxy logs ---"
        sudo snap logs aproxy.aproxy -n 50 2>&1 || sudo journalctl -u aproxy --since -1min --no-pager 2>&1 || true
    } >> "${ARTIFACT_DIR}/host/host-monitor.log"

    INSTANCES=$(lxc list -c n,s,t --format csv 2>/dev/null || true)
    if [ -n "$INSTANCES" ]; then
        while IFS=, read -r NAME STATUS TYPE; do
            [ -z "$NAME" ] && continue
            [ "$STATUS" != "RUNNING" ] && continue

            VM_DIR="${ARTIFACT_DIR}/vms/${NAME}"
            mkdir -p "$VM_DIR"


            if [ -z "${CONFIGURED_DEBUG[$NAME]:-}" ]; then

                HAS_DROPIN=$(timeout 10s lxc exec "$NAME" -- test -f /etc/systemd/system/snapd.service.d/debug.conf && echo "yes" || echo "no")
                if [ "$HAS_DROPIN" != "yes" ]; then
                    echo "Injecting snapd debug drop-in into VM $NAME at $TIMESTAMP" >> "$MONITOR_LOG"
                    timeout 15s lxc exec "$NAME" -- mkdir -p /etc/systemd/system/snapd.service.d 2>/dev/null || true
                    timeout 15s lxc exec "$NAME" -- bash -c 'cat << "EOF" > /etc/systemd/system/snapd.service.d/debug.conf
[Service]
Environment=SNAPD_DEBUG=1 SNAPD_DEBUG_HTTP=3
EOF' 2>/dev/null || true
                    timeout 15s lxc exec "$NAME" -- systemctl daemon-reload 2>/dev/null || true
                    timeout 15s lxc exec "$NAME" -- systemctl restart snapd.service 2>/dev/null || true
                fi


                timeout 15s lxc exec "$NAME" -- ip -d link > "${VM_DIR}/ip-link.txt" 2>&1 || true
                timeout 15s lxc exec "$NAME" -- ip route > "${VM_DIR}/ip-route.txt" 2>&1 || true
                CONFIGURED_DEBUG[$NAME]=true
                echo "VM $NAME configured and network facts recorded at $TIMESTAMP" >> "$MONITOR_LOG"
            fi


            {
                echo "=== VM $NAME SNAPSHOT [$TIMESTAMP] ==="
                echo "--- snap changes ---"
                timeout 15s lxc exec "$NAME" -- snap changes 2>&1 || true

                echo "--- snap tasks (install or in-progress) ---"

                DOING_ID=$(timeout 15s lxc exec "$NAME" -- snap changes 2>/dev/null | awk '/^[0-9]+/ {if ($2 == "Doing" || $2 == "Do" || $2 == "Undoing") {print $1; exit}}' || true)
                if [ -n "$DOING_ID" ]; then
                    echo "Active change tasks (ID $DOING_ID):"
                    timeout 15s lxc exec "$NAME" -- snap tasks "$DOING_ID" 2>&1 || true
                else
                    echo "Last install change tasks:"
                    timeout 15s lxc exec "$NAME" -- snap tasks --last=install 2>&1 || true
                fi

                echo "--- ss -tnpi ---"
                timeout 15s lxc exec "$NAME" -- ss -tnpi 2>&1 || true
                echo "--- ss -tnpi dport = :443 ---"
                timeout 15s lxc exec "$NAME" -- ss -tnpi 'dport = :443' 2>&1 || true
                echo "--- journalctl snapd (last 1 min) ---"
                timeout 15s lxc exec "$NAME" -- journalctl -u snapd --since -1min --no-pager 2>&1 || true
            } >> "${VM_DIR}/monitor.log"


            DOING_CHANGE=$(timeout 15s lxc exec "$NAME" -- snap changes 2>/dev/null | awk '$2=="Doing" {print $1; exit}' || true)
            NOW=$(date +%s)
            if [ -n "$DOING_CHANGE" ]; then
                if [ -z "${FIRST_SEEN_DOING[$NAME]:-}" ]; then
                    FIRST_SEEN_DOING[$NAME]=$NOW
                    echo "VM $NAME change $DOING_CHANGE entered Doing state at $(date -u -Iseconds)" >> "$MONITOR_LOG"
                fi
                ELAPSED=$(( NOW - FIRST_SEEN_DOING[$NAME] ))
                echo "VM $NAME change $DOING_CHANGE has been Doing for ${ELAPSED}s" >> "$MONITOR_LOG"

                if [ "$ELAPSED" -ge 240 ] && [ -z "${PROBED[$NAME]:-}" ]; then
                    PROBED[$NAME]=true
                    PROBE_LOG="${ARTIFACT_DIR}/probes/probe-${NAME}-${TIMESTAMP}.log"
                    echo ">>> STALL DETECTED on VM $NAME (change $DOING_CHANGE Doing for ${ELAPSED}s)! Initiating active probes..." | tee -a "$MONITOR_LOG" "$PROBE_LOG"


                    CDN_URL=$(timeout 15s lxc exec "$NAME" -- journalctl -u snapd --no-pager 2>/dev/null | grep -oE 'https://[^ "'\''>]*cdn\.snapcraftcontent\.com[^ "'\''>]*|https://[^ "'\''>]*fastly\.cdn\.snapcraft\.io[^ "'\''>]*' | tail -n 1 || true)
                    echo "Identified CDN URL: ${CDN_URL:-none}" | tee -a "$MONITOR_LOG" "$PROBE_LOG"

                    STALLED_FLOW=$(timeout 15s lxc exec "$NAME" -- ss -tnp 'dport = :443' 2>/dev/null | grep -E 'ESTAB|SYN-SENT' | head -n 1 || true)
                    echo "Stalled socket flow: ${STALLED_FLOW:-none}" | tee -a "$MONITOR_LOG" "$PROBE_LOG"
                    CDN_IP=$(echo "$STALLED_FLOW" | awk '{print $5}' | cut -d: -f1 || true)
                    VM_SRC_PORT=$(echo "$STALLED_FLOW" | awk '{print $4}' | cut -d: -f2 || true)

                    echo "--- Interception check (aproxy & squid) ---" >> "$PROBE_LOG"
                    echo "Snap info aproxy:" >> "$PROBE_LOG"
                    snap info aproxy >> "$PROBE_LOG" 2>&1 || true
                    echo "Host nft ruleset:" >> "$PROBE_LOG"
                    sudo nft list ruleset >> "$PROBE_LOG" 2>&1 || true
                    echo "Host ss for aproxy port 54969 and squid port 3128:" >> "$PROBE_LOG"
                    ss -tnpi '( sport = :54969 or dport = :54969 or dport = :3128 )' >> "$PROBE_LOG" 2>&1 || true
                    echo "Aproxy snap logs:" >> "$PROBE_LOG"
                    (sudo snap logs aproxy.aproxy -n 200 || sudo journalctl -u aproxy -n 200) >> "$PROBE_LOG" 2>&1 || true

                    VM_IFACE=$(timeout 15s lxc exec "$NAME" -- ip -o route get 1.1.1.1 2>/dev/null | awk '{print $5}' || echo "enp5s0")
                    ORIG_MTU=$(timeout 15s lxc exec "$NAME" -- cat "/sys/class/net/$VM_IFACE/mtu" 2>/dev/null || echo "1500")
                    echo "VM $NAME interface: $VM_IFACE, default MTU: $ORIG_MTU" | tee -a "$MONITOR_LOG" "$PROBE_LOG"

                    echo "--- MTU and ping probes ---" >> "$PROBE_LOG"
                    echo "Host tcp_mtu_probing: $(sysctl -n net.ipv4.tcp_mtu_probing 2>/dev/null)" >> "$PROBE_LOG"
                    echo "VM tcp_mtu_probing: $(timeout 10s lxc exec "$NAME" -- sysctl -n net.ipv4.tcp_mtu_probing 2>/dev/null)" >> "$PROBE_LOG"
                    if [ -n "$CDN_IP" ]; then
                        echo "Ping DF bit 1422 from VM to $CDN_IP:" >> "$PROBE_LOG"
                        timeout 10s lxc exec "$NAME" -- ping -M do -s 1422 -c 2 "$CDN_IP" >> "$PROBE_LOG" 2>&1 || true
                        echo "Ping DF bit 1472 from VM to $CDN_IP:" >> "$PROBE_LOG"
                        timeout 10s lxc exec "$NAME" -- ping -M do -s 1472 -c 2 "$CDN_IP" >> "$PROBE_LOG" 2>&1 || true
                    fi

                    TARGET_HOST="canonical-bos01.cdn.snapcraftcontent.com"
                    if [ -n "$CDN_URL" ]; then
                        PARSED_HOST=$(echo "$CDN_URL" | awk -F[/:] '{print $4}')
                        [ -n "$PARSED_HOST" ] && TARGET_HOST="$PARSED_HOST"
                    fi
                    SWEEP_URL="$CDN_URL"
                    [ -z "$SWEEP_URL" ] && SWEEP_URL="https://${TARGET_HOST}/"
                    A_RECORDS=$(getent ahostsv4 "$TARGET_HOST" 2>/dev/null | awk '{print $1}' | sort -u || true)
                    echo "CDN Node Sweep for $TARGET_HOST (A records: $A_RECORDS):" >> "$PROBE_LOG"
                    echo "--- CDN Sweep on Host ---" >> "$PROBE_LOG"
                    for ip in $A_RECORDS; do
                        echo -n "Host -> $ip: " >> "$PROBE_LOG"
                        curl -s -o /dev/null --resolve "${TARGET_HOST}:443:${ip}" --connect-timeout 5 --max-time 30 \
                          -w 'ip=%{remote_ip} connect=%{time_connect} ttfb=%{time_starttransfer} code=%{http_code} size=%{size_download}\n' \
                          "$SWEEP_URL" >> "$PROBE_LOG" 2>&1 || echo "curl failed ($?)" >> "$PROBE_LOG"
                    done
                    echo "--- CDN Sweep in VM $NAME ---" >> "$PROBE_LOG"
                    for ip in $A_RECORDS; do
                        echo -n "VM -> $ip: " >> "$PROBE_LOG"
                        timeout 35s lxc exec "$NAME" -- curl -s -o /dev/null --resolve "${TARGET_HOST}:443:${ip}" --connect-timeout 5 --max-time 30 \
                          -w 'ip=%{remote_ip} connect=%{time_connect} ttfb=%{time_starttransfer} code=%{http_code} size=%{size_download}\n' \
                          "$SWEEP_URL" >> "$PROBE_LOG" 2>&1 || echo "curl failed ($?)" >> "$PROBE_LOG"
                    done


                    echo "--- PROBE: VM $NAME (MTU $ORIG_MTU) API and CDN ---" >> "$PROBE_LOG"
                    timeout 65s lxc exec "$NAME" -- curl -v --max-time 60 -o /dev/null https://api.snapcraft.io/ >> "$PROBE_LOG" 2>&1 || true
                    if [ -n "$CDN_URL" ]; then
                        echo "--- In-VM curl CDN ($CDN_URL) ---" >> "$PROBE_LOG"
                        timeout 125s lxc exec "$NAME" -- curl -v --max-time 120 -o /dev/null "$CDN_URL" >> "$PROBE_LOG" 2>&1 || true
                    else
                        echo "--- In-VM snap download core20 ---" >> "$PROBE_LOG"
                        timeout 125s lxc exec "$NAME" -- bash -c 'mkdir -p /tmp/probe-snap && timeout 120 snap download core20 --target-dir /tmp/probe-snap' >> "$PROBE_LOG" 2>&1 || true
                    fi


                    echo "--- PROBE: VM $NAME (MTU 1400) ---" >> "$PROBE_LOG"
                    timeout 15s lxc exec "$NAME" -- ip link set dev "$VM_IFACE" mtu 1400 >> "$PROBE_LOG" 2>&1 || true
                    timeout 65s lxc exec "$NAME" -- curl -v --max-time 60 -o /dev/null https://api.snapcraft.io/ >> "$PROBE_LOG" 2>&1 || true
                    if [ -n "$CDN_URL" ]; then
                        timeout 125s lxc exec "$NAME" -- curl -v --max-time 120 -o /dev/null "$CDN_URL" >> "$PROBE_LOG" 2>&1 || true
                    else
                        timeout 125s lxc exec "$NAME" -- bash -c 'mkdir -p /tmp/probe-snap-1400 && timeout 120 snap download core20 --target-dir /tmp/probe-snap-1400' >> "$PROBE_LOG" 2>&1 || true
                    fi


                    echo "--- PROBE: VM $NAME (MTU 1300) ---" >> "$PROBE_LOG"
                    timeout 15s lxc exec "$NAME" -- ip link set dev "$VM_IFACE" mtu 1300 >> "$PROBE_LOG" 2>&1 || true
                    timeout 65s lxc exec "$NAME" -- curl -v --max-time 60 -o /dev/null https://api.snapcraft.io/ >> "$PROBE_LOG" 2>&1 || true
                    if [ -n "$CDN_URL" ]; then
                        timeout 125s lxc exec "$NAME" -- curl -v --max-time 120 -o /dev/null "$CDN_URL" >> "$PROBE_LOG" 2>&1 || true
                    else
                        timeout 125s lxc exec "$NAME" -- bash -c 'mkdir -p /tmp/probe-snap-1300 && timeout 120 snap download core20 --target-dir /tmp/probe-snap-1300' >> "$PROBE_LOG" 2>&1 || true
                    fi


                    timeout 15s lxc exec "$NAME" -- ip link set dev "$VM_IFACE" mtu "$ORIG_MTU" >> "$PROBE_LOG" 2>&1 || true


                    echo "--- PROBE: Host direct ---" >> "$PROBE_LOG"
                    curl -v --max-time 60 -o /dev/null https://api.snapcraft.io/ >> "$PROBE_LOG" 2>&1 || true
                    if [ -n "$CDN_URL" ]; then
                        echo "--- Host curl CDN ($CDN_URL) ---" >> "$PROBE_LOG"
                        curl -v --max-time 120 -o /dev/null "$CDN_URL" >> "$PROBE_LOG" 2>&1 || true
                    else
                        echo "--- Host snap download core20 ---" >> "$PROBE_LOG"
                        mkdir -p /tmp/host-probe-snap
                        timeout 120 snap download core20 --target-dir /tmp/host-probe-snap >> "$PROBE_LOG" 2>&1 || true
                    fi

                    echo ">>> Active probes completed for VM $NAME. Log written to $PROBE_LOG" | tee -a "$MONITOR_LOG"
                fi
            else
                FIRST_SEEN_DOING[$NAME]=""
            fi
        done <<< "$INSTANCES"
    fi

    sleep 30
done
