#!/usr/bin/env python3
import sys
import re
from collections import defaultdict

def parse_file(filepath):
    stats = defaultdict(lambda: {"ok": 0, "fail": 0, "tcp": [], "tls": [], "ttfb": [], "codes": defaultdict(int), "errors": defaultdict(int)})
    
    with open(filepath, "r") as f:
        for line in f:
            line = line.strip()
            if line.startswith("OK "):
                m = re.search(r"domain=([^ ]+) ip=([^ ]+) tcp=([^ ]+) tls=([^ ]+) ttfb=([^ ]+) code=([0-9]+)", line)
                if m:
                    domain, ip, tcp, tls, ttfb, code = m.groups()
                    key = (domain, ip)
                    stats[key]["ok"] += 1
                    stats[key]["tcp"].append(float(tcp) * 1000)
                    stats[key]["tls"].append(float(tls) * 1000)
                    stats[key]["ttfb"].append(float(ttfb) * 1000)
                    stats[key]["codes"][code] += 1
            elif line.startswith("FAIL "):
                m = re.search(r"domain=([^ ]+) ip=([^ ]+) rc=([0-9]+)(?: err=(.*))?", line)
                if m:
                    domain, ip, rc, err = m.group(1), m.group(2), m.group(3), m.group(4) or ""
                    key = (domain, ip)
                    stats[key]["fail"] += 1
                    stats[key]["errors"][f"rc={rc}"] += 1

    print(f"\n{'Domain':<42} {'IP':<17} {'Total':<6} {'OK':<4} {'Fail':<5} {'Avg TCP':<9} {'Avg TLS':<9} {'Avg TTFB':<9} {'Codes/Errors'}")
    print("-" * 125)
    for (domain, ip), data in sorted(stats.items()):
        total = data["ok"] + data["fail"]
        avg_tcp = f"{sum(data['tcp'])/len(data['tcp']):.1f}ms" if data["tcp"] else "-"
        avg_tls = f"{sum(data['tls'])/len(data['tls']):.1f}ms" if data["tls"] else "-"
        avg_ttfb = f"{sum(data['ttfb'])/len(data['ttfb']):.1f}ms" if data["ttfb"] else "-"
        
        info = []
        for c, count in sorted(data["codes"].items()):
            info.append(f"{c}:{count}")
        for e, count in sorted(data["errors"].items()):
            info.append(f"{e}:{count}")
        info_str = ", ".join(info)
        
        print(f"{domain:<42} {ip:<17} {total:<6} {data['ok']:<4} {data['fail']:<5} {avg_tcp:<9} {avg_tls:<9} {avg_ttfb:<9} {info_str}")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        parse_file(sys.argv[1])
    else:
        print("Usage: summarize-sweep.py <sweep-results.txt>")
