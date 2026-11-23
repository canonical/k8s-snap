#!/usr/bin/env bash
set -u

MODE="${1:-public}"
SQUID_URL="${2:-}"
OUT_DIR="${3:-/tmp/tls-investigation}"

mkdir -p "$OUT_DIR/logs" "$OUT_DIR/pcap"

echo "=== TLS Investigation Mode: $MODE (Started: $(date -u -Iseconds)) ===" | tee "$OUT_DIR/summary.txt"

PUB_IP=$(curl -s --max-time 5 https://ifconfig.me 2>/dev/null || curl -s --max-time 5 https://api.ipify.org 2>/dev/null || echo "unknown")
echo "Runner Public IP: $PUB_IP" | tee -a "$OUT_DIR/summary.txt"
GEO_INFO=$(curl -s --max-time 5 "https://ipinfo.io/${PUB_IP}/json" 2>/dev/null || echo "{}")
echo "Runner Geo: $GEO_INFO" | tee -a "$OUT_DIR/summary.txt"

CURL_PROXY_ARGS=()
OPENSSL_PROXY_ARGS=()
if [ "$MODE" = "squid" ]; then
    CURL_PROXY_ARGS=(-x "$SQUID_URL")
    SQUID_HOSTPORT=$(echo "$SQUID_URL" | sed -e 's|http://||' -e 's|/.*||')
    OPENSSL_PROXY_ARGS=(-proxy "$SQUID_HOSTPORT")
    echo "Using Squid proxy: $SQUID_URL ($SQUID_HOSTPORT)" | tee -a "$OUT_DIR/summary.txt"
    SQUID_HOST=$(echo "$SQUID_HOSTPORT" | cut -d: -f1)
    echo "Squid resolved IPs:" | tee -a "$OUT_DIR/summary.txt"
    getent ahosts "$SQUID_HOST" 2>/dev/null | tee -a "$OUT_DIR/summary.txt" || true
fi

PCAP_FILE="$OUT_DIR/pcap/capture.pcap"
if [ "$MODE" = "public" ]; then
    sudo tcpdump -U -Z root -i any -s 0 -w "$PCAP_FILE" 'tcp port 443' >/dev/null 2>&1 &
    TCPDUMP_PID=$!
elif [ "$MODE" = "squid" ]; then
    sudo tcpdump -U -Z root -i any -s 0 -w "$PCAP_FILE" 'tcp port 3128' >/dev/null 2>&1 &
    TCPDUMP_PID=$!
fi

echo "--- 1. 100 Trials standard curl https://api.snapcraft.io/ ---" | tee -a "$OUT_DIR/summary.txt"
FAIL_COUNT_STD=0
OK_COUNT_STD=0
declare -A FAIL_TYPES_STD
for i in $(seq 1 100); do
    LOG_FILE="$OUT_DIR/logs/std_trial_${i}.log"
    rc=0
    curl -sv --max-time 15 "${CURL_PROXY_ARGS[@]}" "https://api.snapcraft.io/" -o /dev/null > "$LOG_FILE" 2>&1 || rc=$?
    if [ $rc -ne 0 ]; then
        FAIL_COUNT_STD=$((FAIL_COUNT_STD + 1))
        ERR_MSG=$(grep -E 'error:|OpenSSL SSL_connect:|curl: \([0-9]+\)|SSL_ERROR|Alert|reset|EOF|timed out|50[0-9]' "$LOG_FILE" | tail -n 2 | tr '\n' ' ' || echo "rc=$rc")
        FAIL_TYPES_STD["$ERR_MSG"]=$(( ${FAIL_TYPES_STD["$ERR_MSG"]:-0} + 1 ))
        echo "Trial $i: FAILED (rc=$rc): $ERR_MSG" | tee -a "$OUT_DIR/summary.txt"
    else
        OK_COUNT_STD=$((OK_COUNT_STD + 1))
        rm -f "$LOG_FILE"
    fi
done
echo "Standard curl: $OK_COUNT_STD OK, $FAIL_COUNT_STD FAILED / 100" | tee -a "$OUT_DIR/summary.txt"

echo "--- 2. 50 Trials openssl s_client -brief ---" | tee -a "$OUT_DIR/summary.txt"
FAIL_COUNT_SSL=0
OK_COUNT_SSL=0
for i in $(seq 1 50); do
    LOG_FILE="$OUT_DIR/logs/ssl_trial_${i}.log"
    rc=0
    openssl s_client -connect api.snapcraft.io:443 -servername api.snapcraft.io "${OPENSSL_PROXY_ARGS[@]}" -brief </dev/null > "$LOG_FILE" 2>&1 || rc=$?
    if [ $rc -ne 0 ] || grep -qiE 'error|alert|failure' "$LOG_FILE"; then
        FAIL_COUNT_SSL=$((FAIL_COUNT_SSL + 1))
        ERR_MSG=$(grep -E 'error|alert|errno|EOF|failure' "$LOG_FILE" | head -n 2 | tr '\n' ' ' || echo "rc=$rc")
        echo "OpenSSL Trial $i: FAILED (rc=$rc): $ERR_MSG" | tee -a "$OUT_DIR/summary.txt"
    else
        OK_COUNT_SSL=$((OK_COUNT_SSL + 1))
        rm -f "$LOG_FILE"
    fi
done
echo "OpenSSL s_client: $OK_COUNT_SSL OK, $FAIL_COUNT_SSL FAILED / 50" | tee -a "$OUT_DIR/summary.txt"

echo "--- 3. 50 Trials curl --tls-max 1.2 ---" | tee -a "$OUT_DIR/summary.txt"
FAIL_COUNT_TLS12=0
OK_COUNT_TLS12=0
for i in $(seq 1 50); do
    LOG_FILE="$OUT_DIR/logs/tls12_trial_${i}.log"
    rc=0
    curl -sv --max-time 15 --tls-max 1.2 "${CURL_PROXY_ARGS[@]}" "https://api.snapcraft.io/" -o /dev/null > "$LOG_FILE" 2>&1 || rc=$?
    if [ $rc -ne 0 ]; then
        FAIL_COUNT_TLS12=$((FAIL_COUNT_TLS12 + 1))
        ERR_MSG=$(grep -E 'error:|OpenSSL SSL_connect:|curl: \([0-9]+\)|SSL_ERROR|Alert|reset|EOF|timed out|50[0-9]' "$LOG_FILE" | tail -n 2 | tr '\n' ' ' || echo "rc=$rc")
        echo "TLS 1.2 Trial $i: FAILED (rc=$rc): $ERR_MSG" | tee -a "$OUT_DIR/summary.txt"
    else
        OK_COUNT_TLS12=$((OK_COUNT_TLS12 + 1))
        rm -f "$LOG_FILE"
    fi
done
echo "TLS 1.2 curl: $OK_COUNT_TLS12 OK, $FAIL_COUNT_TLS12 FAILED / 50" | tee -a "$OUT_DIR/summary.txt"

echo "--- 4. 50 Trials curl --http1.1 ---" | tee -a "$OUT_DIR/summary.txt"
FAIL_COUNT_H11=0
OK_COUNT_H11=0
for i in $(seq 1 50); do
    LOG_FILE="$OUT_DIR/logs/h11_trial_${i}.log"
    rc=0
    curl -sv --max-time 15 --http1.1 "${CURL_PROXY_ARGS[@]}" "https://api.snapcraft.io/" -o /dev/null > "$LOG_FILE" 2>&1 || rc=$?
    if [ $rc -ne 0 ]; then
        FAIL_COUNT_H11=$((FAIL_COUNT_H11 + 1))
        ERR_MSG=$(grep -E 'error:|OpenSSL SSL_connect:|curl: \([0-9]+\)|SSL_ERROR|Alert|reset|EOF|timed out|50[0-9]' "$LOG_FILE" | tail -n 2 | tr '\n' ' ' || echo "rc=$rc")
        echo "HTTP/1.1 Trial $i: FAILED (rc=$rc): $ERR_MSG" | tee -a "$OUT_DIR/summary.txt"
    else
        OK_COUNT_H11=$((OK_COUNT_H11 + 1))
        rm -f "$LOG_FILE"
    fi
done
echo "HTTP/1.1 curl: $OK_COUNT_H11 OK, $FAIL_COUNT_H11 FAILED / 50" | tee -a "$OUT_DIR/summary.txt"

echo "--- 5. 50 Trials openssl -no_tls1_3 (minimal ClientHello) ---" | tee -a "$OUT_DIR/summary.txt"
FAIL_COUNT_NOTLS13=0
OK_COUNT_NOTLS13=0
for i in $(seq 1 50); do
    LOG_FILE="$OUT_DIR/logs/notls13_trial_${i}.log"
    rc=0
    openssl s_client -connect api.snapcraft.io:443 -servername api.snapcraft.io "${OPENSSL_PROXY_ARGS[@]}" -no_tls1_3 -brief </dev/null > "$LOG_FILE" 2>&1 || rc=$?
    if [ $rc -ne 0 ] || grep -qiE 'error|alert|failure' "$LOG_FILE"; then
        FAIL_COUNT_NOTLS13=$((FAIL_COUNT_NOTLS13 + 1))
        ERR_MSG=$(grep -E 'error|alert|errno|EOF|failure' "$LOG_FILE" | head -n 2 | tr '\n' ' ' || echo "rc=$rc")
        echo "OpenSSL -no_tls1_3 Trial $i: FAILED (rc=$rc): $ERR_MSG" | tee -a "$OUT_DIR/summary.txt"
    else
        OK_COUNT_NOTLS13=$((OK_COUNT_NOTLS13 + 1))
        rm -f "$LOG_FILE"
    fi
done
echo "OpenSSL -no_tls1_3: $OK_COUNT_NOTLS13 OK, $FAIL_COUNT_NOTLS13 FAILED / 50" | tee -a "$OUT_DIR/summary.txt"

if [ -n "${TCPDUMP_PID:-}" ]; then
    sudo kill "$TCPDUMP_PID" 2>/dev/null || true
    sleep 2
fi

echo "--- PCAP Analysis (RST / FIN / Control Packets) ---" | tee -a "$OUT_DIR/summary.txt"
sudo tcpdump -nn -r "$PCAP_FILE" 'tcp[tcpflags] & (tcp-rst|tcp-fin) != 0' 2>/dev/null | head -n 40 | tee -a "$OUT_DIR/summary.txt" || true

if [ "$MODE" = "squid" ]; then
    echo "--- Squid CONNECT status codes in PCAP ---" | tee -a "$OUT_DIR/summary.txt"
    sudo tcpdump -nn -A -r "$PCAP_FILE" 'tcp port 3128' 2>/dev/null | grep -E 'HTTP/1\.[01] [0-9]{3}' | sort | uniq -c | tee -a "$OUT_DIR/summary.txt" || true
fi
