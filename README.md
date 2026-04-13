<div align="center">

```
██╗    ██╗███████╗██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
██║    ██║██╔════╝██╔══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║ █╗ ██║█████╗  ██████╔╝███████╗██║     ███████║██╔██╗ ██║
██║███╗██║██╔══╝  ██╔══██╗╚════██║██║     ██╔══██║██║╚██╗██║
╚███╔███╔╝███████╗██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
 ╚══╝╚══╝ ╚══════╝╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
```

**The Best Recon Tool Written in Go**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen?style=for-the-badge&logo=github-actions)](../../actions)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-blue?style=for-the-badge)]()
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker)](Dockerfile)

*Port scanning · HTTP/HTTPS probing · TLS introspection · DNS enumeration*  
*Subdomain brute-force · WHOIS · TCP banner grabbing · YAML vulnerability templates · HTML reports*

</div>

---

> ⚠️ **Legal Notice:** Port scanning and probing third-party systems may be intrusive and/or prohibited.
> Only scan targets you have **explicit written authorization** to test.

---

## Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Installation](#-installation)
- [Commands](#-commands)
- [Usage Examples](#-usage-examples)
- [YAML Templates](#-yaml-templates)
- [All Flags](#%EF%B8%8F-all-flags)
- [HTML Report](#-html-report)
- [Interactive TUI](#-interactive-tui)
- [CI/CD Pipelines](#-cicd-pipelines)
- [Docker](#-docker)
- [Contributing](#-contributing)

---

## ✨ Features

<table>
<tr>
<td valign="top" width="50%">

**Recon & Discovery**
- 🔍 High-concurrency TCP port scan — worker-pool + token-bucket rate limiter
- 🌐 HTTP/HTTPS probing with TLS introspection (version, cipher suite, ALPN, certs)
- 🧬 Service fingerprinting — 66+ signatures (CDNs, WAFs, frameworks, servers)
- 📡 DNS enumeration: A · AAAA · CNAME · MX · TXT · NS
- 🔤 Subdomain brute-force — 200-word built-in wordlist or custom file
- 🏢 WHOIS — registrar, creation/expiry dates, nameservers
- 🚩 TCP banner grabbing — SSH, FTP, SMTP, POP3, Redis, MySQL and more

</td>
<td valign="top" width="50%">

**Detection & Reporting**
- 🧪 YAML-based vulnerability template engine (Nuclei-inspired)
- 🐛 CVE templates: Log4Shell, Spring4Shell (extensible library)
- 🔓 Exposure checks: `.env`, `.git/config`, Jenkins, phpMyAdmin, Spring Actuator
- 📊 Self-contained dark-mode HTML report — TLS, tech stack, WAF/CDN badges
- 📄 JSON & ANSI text output, multi-target via file or stdin
- 🖥️ Real-time interactive TUI — search, sort, filter, detail pane
- 🐋 Docker-ready · goreleaser · CI matrix (Linux / Windows / macOS)

</td>
</tr>
</table>

---

## 🏗️ Architecture

```
webscan/
├── cmd/                    # Cobra CLI entry points
│   ├── scan.go             # HTTP port scanner  (--list, --html, --json)
│   ├── dns.go              # DNS enum + subdomain brute-force
│   ├── whois.go            # WHOIS lookup
│   ├── banner.go           # TCP banner grabbing
│   ├── vuln.go             # YAML template vulnerability scanner
│   └── tui.go              # Interactive TUI launcher
│
├── internal/
│   ├── scanner/            # Worker-pool engine, rate limiter, banner grabber, benchmarks
│   ├── web/                # HTTP probe, TLS introspection, fingerprinting, signatures loader
│   ├── dns/                # Concurrent DNS resolver + subdomain brute-forcer
│   ├── whois/              # WHOIS TCP client + field parser (IANA → TLD chain)
│   ├── templates/          # YAML template engine (load dir, matcher eval, run all)
│   └── ui/                 # TUI (tview/tcell) + headless runner + persistent config
│
├── pkg/output/             # Output formatters: JSON, ANSI text, HTML report
│
├── templates/              # Built-in detection templates
│   ├── cve/                # CVE-2021-44228 (Log4Shell), CVE-2022-22965 (Spring4Shell)
│   └── exposures/          # .env, .git/config, Jenkins, phpMyAdmin, default-creds, Actuator
│
├── test/
│   ├── signatures/         # JSON fingerprint signatures DB (66+ entries, auto-updatable)
│   └── integration/        # Docker Compose integration harness
│
├── scripts/
│   ├── update_signatures.py   # Merge new CDN/WAF/server vendor signatures
│   └── expand_signatures.py   # Generate expanded 250-entry DB
│
├── ci/
│   └── run-headless-tests.sh  # Xvfb-aware TUI test runner
│
└── .github/workflows/
    ├── ci.yml              # Lint · Test matrix · gosec · Build artifacts · Integration
    ├── docker.yml          # Build & push to GHCR
    ├── headless-ui.yml     # Headless TUI tests (label-gated)
    ├── release.yml         # Goreleaser full release
    └── release_dryrun.yml  # Goreleaser dry-run (no publish)
```

---

## 📦 Installation

**Requirements:** Go 1.21+

```bash
# Clone and build
git clone https://github.com/your-user/webscan.git
cd webscan
go mod tidy
go build -o webscan .         # Linux / macOS
go build -o webscan.exe .     # Windows
```

```bash
# Or install directly with go
go install github.com/your-user/webscan@latest
```

```bash
# Docker (no Go install required)
docker pull ghcr.io/your-user/webscan:latest
docker run --rm ghcr.io/your-user/webscan scan -t example.com -p 80,443 --json
```

---

## 🛠️ Commands

| Command | Description |
|---|---|
| `scan` | HTTP/HTTPS port scan — TLS, fingerprinting, multi-target, HTML report |
| `dns` | DNS record enumeration + subdomain brute-force |
| `whois` | WHOIS lookup — registrar, dates, nameservers |
| `banner` | Raw TCP banner grabbing for service identification |
| `vuln` | YAML-template vulnerability & exposure detection |
| `tui` | Real-time interactive terminal UI |

---

## 🚀 Usage Examples

### Port Scan

```bash
# Quick scan (defaults: ports 80, 443)
webscan scan -t example.com

# All web ports + dark-mode HTML report
webscan scan -t example.com -p 80,443,8080,8443,3000,5000 --html report.html

# JSON output — pipe to jq for filtering
webscan scan -t example.com -p 1-1024 --json | jq '.results[] | select(.open)'

# Multi-target from file
webscan scan -l targets.txt -p 80,443 --threads 200 --rate 100 --json

# Multi-target from stdin (pipeline)
cat targets.txt | webscan scan -l - -p 80,443 --json

# Verbose output with TLS details + fingerprinting
webscan scan -t example.com -p 443 --verbose --signatures test/signatures/popular_signatures.json
```

### DNS Enumeration

```bash
# Enumerate all record types (A, AAAA, CNAME, MX, TXT, NS)
webscan dns -t example.com

# Subdomain brute-force with built-in 200-word wordlist
webscan dns -t example.com --brute

# Custom wordlist + JSON output
webscan dns -t example.com --brute --wordlist /usr/share/seclists/Discovery/DNS/subdomains-top1million-5000.txt --json

# High-concurrency brute-force
webscan dns -t example.com --brute --threads 100 --json > dns_results.json
```

### WHOIS

```bash
webscan whois -t example.com

webscan whois -t example.com --raw           # include full raw WHOIS text
webscan whois -t example.com --json          # machine-readable output
```

### Banner Grabbing

```bash
# Default: probes 16 common service ports
webscan banner -t example.com

# Custom ports, 5s timeout
webscan banner -t example.com -p 22,21,25,110,143,3306,5432,6379,27017 --timeout 5 --json
```

### Vulnerability Scanning

```bash
# Run all built-in templates (templates/ directory)
webscan vuln -t https://example.com

# Filter findings by severity
webscan vuln -t https://example.com --severity high
webscan vuln -t https://example.com --severity critical

# Custom templates directory
webscan vuln -t https://example.com -T /path/to/my-templates/ --json

# Pipe critical findings to a report
webscan vuln -t https://example.com --json | jq '.[] | select(.severity == "critical")'
```

### Interactive TUI

```bash
webscan tui -t example.com -p 1-1024 --threads 100 --timeout 3
```

---

## 🧪 YAML Templates

Templates live in `templates/` and follow a simple, extensible YAML schema:

```yaml
id: git-config-exposed
name: Git Config File Exposed
severity: high
description: |
  Publicly accessible .git/config files may expose credentials,
  remote URLs, and branch names.
tags: [git, exposure, credentials]

requests:
  - method: GET
    path: /.git/config
    matchers:
      - type: word
        words: ["[core]"]
        part: body
      - type: status
        status: [200]
    matcher-condition: and
```

**Matcher types:** `status` · `word` · `regex` · `header`  
**Conditions:** `and` · `or` (per-matcher and top-level)  
**Match scope:** `body` (default) · `header` · `all`

### Built-in Template Library

| ID | Category | Severity |
|---|---|---|
| `cve-2021-44228-log4shell` | CVE | 🔴 Critical |
| `cve-2022-22965-spring4shell` | CVE | 🔴 Critical |
| `env-file` | Secrets Exposure | 🟠 High |
| `git-config` | Secrets Exposure | 🟠 High |
| `jenkins` | Unauthenticated Panel | 🟠 High |
| `default-credentials` | Auth Bypass | 🟠 High |
| `spring-actuator` | Data Exposure | 🟡 Medium |
| `phpmyadmin-panel` | Admin Panel | 🟡 Medium |

> Drop any `.yaml` file into `templates/` and it is automatically picked up at runtime.

---

## ⚙️ All Flags

### `scan`

| Flag | Default | Description |
|---|---|---|
| `-t, --target` | — | Single target domain / IP |
| `-l, --list` | — | File with one target per line (`-` reads stdin) |
| `-p, --ports` | `80,443` | Ports or ranges (`1-1024,8080`) |
| `--threads` | `100` | Concurrent workers |
| `--timeout` | `2` | Network timeout (seconds) |
| `--retries` | `2` | TCP connect retry count |
| `--rate` | `0` | Rate limit conn/sec (`0` = unlimited) |
| `--signatures` | auto | Path to JSON signatures file |
| `--json` | false | JSON output |
| `--html` | — | Save HTML report to file |
| `-v, --verbose` | false | Show headers, fingerprint details |

### `dns`

| Flag | Default | Description |
|---|---|---|
| `-t, --target` | — | Domain to enumerate |
| `--brute` | false | Enable subdomain brute-force |
| `--wordlist` | built-in | Path to custom wordlist |
| `--threads` | `50` | Concurrent DNS resolvers |
| `--json` | false | JSON output |

### `banner`

| Flag | Default | Description |
|---|---|---|
| `-t, --target` | — | Target host |
| `-p, --ports` | 16 ports | Ports to probe |
| `--timeout` | `3` | TCP timeout (seconds) |
| `--threads` | `50` | Concurrent connections |
| `--json` | false | JSON output |

### `whois`

| Flag | Default | Description |
|---|---|---|
| `-t, --target` | — | Domain to query |
| `--timeout` | `15` | WHOIS query timeout (seconds) |
| `--raw` | false | Include full raw WHOIS text |
| `--json` | false | JSON output |

### `vuln`

| Flag | Default | Description |
|---|---|---|
| `-t, --target` | — | Base URL (`https://example.com`) |
| `-T, --templates` | `templates/` | Templates directory |
| `--severity` | — | Filter: `info/low/medium/high/critical` |
| `--timeout` | `10` | HTTP request timeout (seconds) |
| `--json` | false | JSON output |

---

## 📋 HTML Report

Generate a professional self-contained dark-mode HTML report — no server, no dependencies:

```bash
webscan scan -t example.com -p 80,443,8080,8443 --html report.html
```

**Report includes:**
- Summary stats bar — ports scanned, open count, scan duration
- Per-port table with status badge, protocol, TLS version, server header, detected technologies
- WAF / CDN detection badges
- Hover tooltips for truncated titles
- 100% offline — single `.html` file, zero CDN calls

---

## 🖥️ Interactive TUI

```bash
webscan tui -t example.com -p 1-1024 --threads 100
```

| Key | Action |
|---|---|
| `Enter` | Detail pane — full headers, TLS certs, fingerprint |
| `/` | Live search filter across all columns |
| `s` | Cycle sort: port → status → server |
| `o` | Show only **open** ports |
| `f` | Show only **filtered** ports |
| `c` | Show only **closed** ports |
| `a` | Show **all** (reset filter) |
| `h` or `?` | Help overlay |
| `q` / `Esc` | Quit |

Config auto-saved to `.webscan_config.json` — sort preference and style persist across runs.

---

## 🔄 CI/CD Pipelines

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` — Lint | push / PR → `main` | `golangci-lint` static analysis |
| `ci.yml` — Test | push / PR → `main` | `go test ./...` on Ubuntu · Windows · macOS |
| `ci.yml` — Security | push / PR → `main` | `gosec` SAST scan |
| `ci.yml` — Build | push / PR → `main` | Cross-compile linux/windows/darwin artifacts |
| `ci.yml` — Integration | label `run-integration` · `workflow_dispatch` | Docker Compose integration harness |
| `headless-ui.yml` | label `run-headless` · `workflow_dispatch` | Headless TUI tests (Xvfb) |
| `docker.yml` | push `main` · tag `v*` | Build + push to GHCR |
| `release_dryrun.yml` | `workflow_dispatch` | goreleaser dry-run (no publish) |
| `release.yml` | tag `v*` | Full release (needs `GH_TOKEN` + `GPG_PRIVATE_KEY` secrets) |

**Trigger workflows manually:**

```bash
gh workflow run ci.yml
gh workflow run headless-ui.yml
gh workflow run release_dryrun.yml
```

---

## 🐋 Docker

```bash
# Build locally
docker build -t webscan .

# Scan
docker run --rm webscan scan -t example.com -p 80,443 --json

# DNS enum
docker run --rm webscan dns -t example.com --brute --json

# Vuln scan
docker run --rm webscan vuln -t https://example.com --json

# Save HTML report to host
docker run --rm -v "$(pwd):/out" webscan scan -t example.com --html /out/report.html
```

Built from `scratch` — fully static binary, ~8 MB compressed image, minimal attack surface.

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

**Quick start:**

```bash
git clone https://github.com/your-user/webscan.git
cd webscan
go mod tidy
go test ./... -v              # run all tests
golangci-lint run ./...       # lint
```

**Add a detection template** — create `templates/exposures/my-check.yaml`, test it locally:

```bash
webscan vuln -t http://localhost -T templates/my-check.yaml
```

**Update the signatures database:**

```bash
python scripts/update_signatures.py
```

---

## 📄 License

[MIT](LICENSE) © 2026 WebScan Contributors

---

<div align="center">

**Built with ❤️ in Go · tview · cobra · yaml.v3**

*Use responsibly. Only scan targets you own or have explicit written permission to test.*

</div>
