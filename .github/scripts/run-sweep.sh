#!/usr/bin/env bash
set -u

MODE="${1:-container}"
SQUID_URL="${2:-}"
OUTPUT_FILE="${3:-/tmp/sweep-results.txt}"
LXC_CONTAINER="${4:-c1}"

DOMAINS=(
    "api.snapcraft.io"
    "canonical-bos01.cdn.snapcraftcontent.com"
    "canonical-lgw01.cdn.snapcraftcontent.com"
)

mkdir -p "$(dirname "$OUTPUT_FILE")"
echo "=== Sweep Mode: $MODE (Started: $(date -u -Iseconds)) ===" | tee -a "$OUTPUT_FILE"

run_trial() {
    local domain="$1"
    local ip="$2"
    local trial="$3"

    local curl_cmd=(curl -s -o /dev/null --resolve "${domain}:443:${ip}" --connect-timeout 5 --max-time 15 -w "ip=%{remote_ip} tcp=%{time_connect} tls=%{time_appconnect} ttfb=%{time_starttransfer} code=%{http_code}\n")

    local res=""
    local rc=0

    if [ "$MODE" = "container" ]; then
        res=$(lxc exec "$LXC_CONTAINER" -- "${curl_cmd[@]}" "https://${domain}/" 2>&1) || rc=$?
    elif [ "$MODE" = "host-squid" ]; then
        res=$("${curl_cmd[@]}" -x "$SQUID_URL" "https://${domain}/" 2>&1) || rc=$?
    elif [ "$MODE" = "host-direct" ] || [ "$MODE" = "public" ]; then
        res=$("${curl_cmd[@]}" "https://${domain}/" 2>&1) || rc=$?
    fi

    if [ $rc -ne 0 ]; then
        echo "FAIL trial=$trial domain=$domain ip=$ip rc=$rc err=${res//$'\n'/ }" | tee -a "$OUTPUT_FILE"
    else
        echo "OK trial=$trial domain=$domain $res" | tee -a "$OUTPUT_FILE"
    fi
}

for domain in "${DOMAINS[@]}"; do
    echo "--- Resolving $domain ---" | tee -a "$OUTPUT_FILE"
    IPS=$(getent ahostsv4 "$domain" 2>/dev/null | awk '{print $1}' | sort -u || true)
    echo "IPs for $domain: $IPS" | tee -a "$OUTPUT_FILE"

    for ip in $IPS; do
        echo "Testing $domain on $ip (10 trials)..."
        for trial in $(seq 1 10); do
            run_trial "$domain" "$ip" "$trial"
        done
    done
done

echo "=== Sweep Finished: $(date -u -Iseconds) ===" | tee -a "$OUTPUT_FILE"
