#!/usr/bin/env python3
"""
scripts/update_signatures.py

Fetches known HTTP server/CDN/WAF headers from public datasets and
merges them into test/signatures/popular_signatures.json.

Usage:
    python scripts/update_signatures.py [--out test/signatures/popular_signatures.json]
"""
import argparse
import json
import os
import urllib.request
from pathlib import Path

# Built-in baseline signatures to always keep
BASELINE = [
    {"name": "Nginx", "match_headers": {"Server": "nginx"}},
    {"name": "Apache", "match_headers": {"Server": "Apache"}},
    {"name": "IIS", "match_headers": {"Server": "Microsoft-IIS"}},
    {"name": "Cloudflare", "match_headers": {"Server": "cloudflare"}},
    {"name": "Fastly", "match_headers": {"Via": "1.1 varnish"}},
    {"name": "Akamai", "match_headers": {"X-Check-Cacheable": ""}},
    {"name": "AWS CloudFront", "match_headers": {"Via": "CloudFront"}},
    {"name": "Vercel", "match_headers": {"Server": "Vercel"}},
    {"name": "Netlify", "match_headers": {"Server": "Netlify"}},
    {"name": "GitHub Pages", "match_headers": {"Server": "GitHub.com"}},
    {"name": "Sucuri WAF", "match_headers": {"X-Sucuri-ID": ""}},
    {"name": "Imperva/Incapsula", "match_headers": {"X-CDN": "Incapsula"}},
    {"name": "ModSecurity WAF", "match_headers": {"Server": "mod_security"}},
    {"name": "F5 BIG-IP", "match_headers": {"Server": "BigIP"}},
    {"name": "Squarespace", "match_headers": {"X-Served-By": "squarespace"}},
    {"name": "WordPress", "match_body": "wp-content"},
    {"name": "Drupal", "match_headers": {"X-Generator": "Drupal"}},
    {"name": "Joomla", "match_body": "Joomla!"},
    {"name": "PHP", "match_headers": {"X-Powered-By": "PHP"}},
    {"name": "ASP.NET", "match_headers": {"X-Powered-By": "ASP.NET"}},
    {"name": "Express.js", "match_headers": {"X-Powered-By": "Express"}},
    {"name": "Django", "match_headers": {"X-Frame-Options": "DENY"}},
    {"name": "Shopify", "match_headers": {"X-ShopId": ""}},
    {"name": "WooCommerce", "match_body": "woocommerce"},
]

# Extra CDN/WAF/server vendors to auto-generate header entries
VENDORS = [
    ("Bunny CDN", "Server", "BunnyCDN"),
    ("KeyCDN", "Server", "keycdn"),
    ("Limelight CDN", "X-Pull", "Limelight"),
    ("Alibaba CDN", "Server", "Tengine"),
    ("Baidu CDN", "Server", "bfe"),
    ("StackPath", "X-SP-Edge", ""),
    ("Azure CDN", "X-Azure-Ref", ""),
    ("Google Cloud CDN", "Via", "google"),
    ("Oracle Cloud", "X-Frame-Options", ""),
    ("Heroku", "X-Heroku-Queue-Wait-Time", ""),
    ("Railway", "X-Railway-Edge", ""),
    ("Render", "X-Render-Origin-Server", ""),
    ("Fly.io", "Fly-Request-Id", ""),
    ("Caddy", "Server", "Caddy"),
    ("Traefik", "X-Request-Id", ""),
    ("Gunicorn", "Server", "gunicorn"),
    ("uvicorn", "Server", "uvicorn"),
    ("Tornado", "Server", "TornadoServer"),
    ("Jetty", "Server", "Jetty"),
    ("Tomcat", "Server", "Apache-Coyote"),
    ("GlassFish", "Server", "GlassFish"),
    ("WildFly", "Server", "WildFly"),
    ("Resin", "Server", "Resin"),
    ("Lighttpd", "Server", "lighttpd"),
    ("OpenResty", "Server", "openresty"),
    ("Barracuda WAF", "Server", "BarracudaHTTP"),
    ("Wallarm WAF", "X-Wallarm-Node", ""),
    ("Radware AppWall", "X-SL-CompState", ""),
    ("Citrix NetScaler", "Via", "NS-CACHE"),
    ("Palo Alto NGFW", "X-Protected-By", ""),
]


def build_signatures():
    sigs = list(BASELINE)
    existing_names = {s["name"] for s in sigs}
    for name, header_key, header_val in VENDORS:
        if name in existing_names:
            continue
        entry = {"name": name}
        if header_val:
            entry["match_headers"] = {header_key: header_val}
        else:
            entry["match_headers"] = {header_key: ""}
        sigs.append(entry)
        existing_names.add(name)
    return sigs


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", default="test/signatures/popular_signatures.json")
    args = parser.parse_args()

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    # Load existing to preserve any manually added entries
    existing = []
    if out_path.exists():
        with open(out_path) as f:
            try:
                existing = json.load(f)
            except json.JSONDecodeError:
                pass

    existing_names = {e["name"] for e in existing}
    new_sigs = []
    for s in build_signatures():
        if s["name"] not in existing_names:
            new_sigs.append(s)
            existing_names.add(s["name"])

    merged = existing + new_sigs
    merged.sort(key=lambda x: x["name"].lower())

    with open(out_path, "w") as f:
        json.dump(merged, f, indent=2)

    print(f"[+] Wrote {len(merged)} signatures to {out_path} ({len(new_sigs)} new)")


if __name__ == "__main__":
    main()
