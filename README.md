<div align="center">

```
██╗    ██╗███████╗██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
██║    ██║██╔════╝██╔══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║ █╗ ██║█████╗  ██████╔╝███████╗██║     ███████║██╔██╗ ██║
██║███╗██║██╔══╝  ██╔══██╗╚════██║██║     ██╔══██║██║╚██╗██║
╚███╔███╔╝███████╗██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
 ╚══╝╚══╝ ╚══════╝╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
```

**A Melhor Ferramenta de Recon Escrita em Go**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen?style=for-the-badge&logo=github-actions)](../../actions)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-blue?style=for-the-badge)]()
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker)](Dockerfile)

*Port scanning · Sondagem HTTP/HTTPS · Inspeção TLS · Enumeração DNS*  
*Brute-force de subdomínios · WHOIS · Banner grabbing TCP · Templates YAML de vulnerabilidades · Relatórios HTML*

</div>

---

> ⚠️ **Aviso Legal:** Varreduras de portas e sondagens em sistemas de terceiros podem ser intrusivas e/ou proibidas.
> Execute varreduras apenas em alvos que você possui **autorização escrita explícita** para testar.

---

## Índice

- [Funcionalidades](#-funcionalidades)
- [Arquitetura](#-arquitetura)
- [Instalação](#-instalação)
- [Comandos](#-comandos)
- [Exemplos de Uso](#-exemplos-de-uso)
- [Templates YAML](#-templates-yaml)
- [Todas as Flags](#%EF%B8%8F-todas-as-flags)
- [Relatório HTML](#-relatório-html)
- [TUI Interativa](#-tui-interativa)
- [Pipelines CI/CD](#-pipelines-cicd)
- [Docker](#-docker)
- [Contribuindo](#-contribuindo)
- [Créditos](#-créditos)

---

## ✨ Funcionalidades

<table>
<tr>
<td valign="top" width="50%">

**Recon & Descoberta**
- 🔍 Varredura TCP de alta concorrência — worker-pool + rate limiter por token-bucket
- 🌐 Sondagem HTTP/HTTPS com inspeção TLS (versão, cipher suite, ALPN, certificados)
- 🧬 Fingerprinting de serviços — 66+ assinaturas (CDNs, WAFs, frameworks, servidores)
- 📡 Enumeração DNS: A · AAAA · CNAME · MX · TXT · NS
- 🔤 Brute-force de subdomínios — wordlist embutida de 200 palavras ou arquivo customizado
- 🏢 WHOIS — registrador, datas de criação/expiração, nameservers
- 🚩 Banner grabbing TCP — SSH, FTP, SMTP, POP3, Redis, MySQL e muito mais

</td>
<td valign="top" width="50%">

**Detecção & Relatórios**
- 🧪 Engine de templates YAML para detecção de vulnerabilidades (inspirado no Nuclei)
- 🐛 Templates de CVE: Log4Shell, Spring4Shell (biblioteca extensível)
- 🔓 Verificação de exposições: `.env`, `.git/config`, Jenkins, phpMyAdmin, Spring Actuator
- 📊 Relatório HTML dark-mode auto-contido — TLS, stack de tecnologias, badges WAF/CDN
- 📄 Saída em JSON e texto ANSI, multi-alvos via arquivo ou stdin
- 🖥️ TUI interativa em tempo real — busca, ordenação, filtros, painel de detalhes
- 🐋 Docker-ready · goreleaser · Matriz CI (Linux / Windows / macOS)

</td>
</tr>
</table>

---

## 🏗️ Arquitetura

```
webscan/
├── cmd/                    # Pontos de entrada da CLI (Cobra)
│   ├── scan.go             # Scanner de portas HTTP  (--list, --html, --json)
│   ├── dns.go              # Enumeração DNS + brute-force de subdomínios
│   ├── whois.go            # Consulta WHOIS
│   ├── banner.go           # Banner grabbing TCP
│   ├── vuln.go             # Scanner de vulnerabilidades via templates YAML
│   └── tui.go              # Lançador da TUI interativa
│
├── internal/
│   ├── scanner/            # Engine worker-pool, rate limiter, banner grabber, benchmarks
│   ├── web/                # Sonda HTTP, inspeção TLS, fingerprinting, carregador de assinaturas
│   ├── dns/                # Resolver DNS concorrente + brute-forcer de subdomínios
│   ├── whois/              # Cliente TCP WHOIS + parser de campos (cadeia IANA → TLD)
│   ├── templates/          # Engine de templates YAML (carrega dir, avalia matchers, executa todos)
│   └── ui/                 # TUI (tview/tcell) + runner headless + config persistente
│
├── pkg/output/             # Formatadores de saída: JSON, texto ANSI, relatório HTML
│
├── templates/              # Templates de detecção embutidos
│   ├── cve/                # CVE-2021-44228 (Log4Shell), CVE-2022-22965 (Spring4Shell)
│   └── exposures/          # .env, .git/config, Jenkins, phpMyAdmin, default-creds, Actuator
│
├── test/
│   ├── signatures/         # Banco de assinaturas JSON (66+ entradas, atualizável)
│   └── integration/        # Harness de integração com Docker Compose
│
├── scripts/
│   ├── update_signatures.py   # Mescla novas assinaturas de vendors CDN/WAF/servidor
│   └── expand_signatures.py   # Gera banco expandido com 250 entradas
│
├── ci/
│   └── run-headless-tests.sh  # Runner de testes TUI com suporte a Xvfb
│
└── .github/workflows/
    ├── ci.yml              # Lint · Matriz de testes · gosec · Build · Integração
    ├── docker.yml          # Build & push para GHCR
    ├── headless-ui.yml     # Testes TUI headless (ativado por label)
    ├── release.yml         # Release completo via Goreleaser
    └── release_dryrun.yml  # Dry-run do Goreleaser (sem publicar)
```

---

## 📦 Instalação

**Requisitos:** Go 1.21+

```bash
# Clonar e compilar
git clone https://github.com/excanear/WebScan.git
cd WebScan
go mod tidy
go build -o webscan .         # Linux / macOS
go build -o webscan.exe .     # Windows
```

```bash
# Ou instalar diretamente via go
go install github.com/excanear/WebScan@latest
```

```bash
# Docker (sem necessidade de instalar Go)
docker pull ghcr.io/excanear/webscan:latest
docker run --rm ghcr.io/excanear/webscan scan -t example.com -p 80,443 --json
```

---

## 🛠️ Comandos

| Comando | Descrição |
|---|---|
| `scan` | Varredura de portas HTTP/HTTPS — TLS, fingerprinting, multi-alvo, relatório HTML |
| `dns` | Enumeração de registros DNS + brute-force de subdomínios |
| `whois` | Consulta WHOIS — registrador, datas, nameservers |
| `banner` | Banner grabbing TCP bruto para identificação de serviços |
| `vuln` | Detecção de vulnerabilidades e exposições via templates YAML |
| `tui` | Interface de terminal interativa em tempo real |

---

## 🚀 Exemplos de Uso

### Varredura de Portas

```bash
# Varredura rápida (padrão: portas 80, 443)
webscan scan -t example.com

# Todas as portas web + relatório HTML dark-mode
webscan scan -t example.com -p 80,443,8080,8443,3000,5000 --html report.html

# Saída JSON — pipe para jq filtrar resultados
webscan scan -t example.com -p 1-1024 --json | jq '.results[] | select(.open)'

# Multi-alvo a partir de arquivo
webscan scan -l targets.txt -p 80,443 --threads 200 --rate 100 --json

# Multi-alvo via stdin (pipeline)
cat targets.txt | webscan scan -l - -p 80,443 --json

# Saída detalhada com TLS + fingerprinting
webscan scan -t example.com -p 443 --verbose --signatures test/signatures/popular_signatures.json
```

### Enumeração DNS

```bash
# Enumerar todos os tipos de registro (A, AAAA, CNAME, MX, TXT, NS)
webscan dns -t example.com

# Brute-force de subdomínios com wordlist embutida de 200 palavras
webscan dns -t example.com --brute

# Wordlist customizada + saída JSON
webscan dns -t example.com --brute --wordlist /usr/share/seclists/Discovery/DNS/subdomains-top1million-5000.txt --json

# Brute-force de alta concorrência
webscan dns -t example.com --brute --threads 100 --json > dns_results.json
```

### WHOIS

```bash
webscan whois -t example.com

webscan whois -t example.com --raw           # inclui texto WHOIS bruto completo
webscan whois -t example.com --json          # saída legível por máquina
```

### Banner Grabbing

```bash
# Padrão: sonda 16 portas de serviços comuns
webscan banner -t example.com

# Portas customizadas, timeout de 5s
webscan banner -t example.com -p 22,21,25,110,143,3306,5432,6379,27017 --timeout 5 --json
```

### Varredura de Vulnerabilidades

```bash
# Executar todos os templates embutidos (diretório templates/)
webscan vuln -t https://example.com

# Filtrar achados por severidade
webscan vuln -t https://example.com --severity high
webscan vuln -t https://example.com --severity critical

# Diretório de templates customizado
webscan vuln -t https://example.com -T /caminho/para/meus-templates/ --json

# Pipe achados críticos para relatório
webscan vuln -t https://example.com --json | jq '.[] | select(.severity == "critical")'
```

### TUI Interativa

```bash
webscan tui -t example.com -p 1-1024 --threads 100 --timeout 3
```

---

## 🧪 Templates YAML

Os templates ficam em `templates/` e seguem um schema YAML simples e extensível:

```yaml
id: git-config-exposed
name: Arquivo Git Config Exposto
severity: high
description: |
  Arquivos .git/config acessíveis publicamente podem expor credenciais,
  URLs remotas e nomes de branches.
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

**Tipos de matcher:** `status` · `word` · `regex` · `header`  
**Condições:** `and` · `or` (por matcher e no nível superior)  
**Escopo de busca:** `body` (padrão) · `header` · `all`

### Biblioteca de Templates Embutidos

| ID | Categoria | Severidade |
|---|---|---|
| `cve-2021-44228-log4shell` | CVE | 🔴 Crítico |
| `cve-2022-22965-spring4shell` | CVE | 🔴 Crítico |
| `env-file` | Exposição de Segredos | 🟠 Alto |
| `git-config` | Exposição de Segredos | 🟠 Alto |
| `jenkins` | Painel Não Autenticado | 🟠 Alto |
| `default-credentials` | Bypass de Autenticação | 🟠 Alto |
| `spring-actuator` | Exposição de Dados | 🟡 Médio |
| `phpmyadmin-panel` | Painel Admin | 🟡 Médio |

> Adicione qualquer arquivo `.yaml` em `templates/` e ele será carregado automaticamente em tempo de execução.

---

## ⚙️ Todas as Flags

### `scan`

| Flag | Padrão | Descrição |
|---|---|---|
| `-t, --target` | — | Domínio / IP alvo único |
| `-l, --list` | — | Arquivo com um alvo por linha (`-` lê stdin) |
| `-p, --ports` | `80,443` | Portas ou intervalos (`1-1024,8080`) |
| `--threads` | `100` | Workers concorrentes |
| `--timeout` | `2` | Timeout de rede (segundos) |
| `--retries` | `2` | Tentativas de reconexão TCP |
| `--rate` | `0` | Limite de conexões/seg (`0` = ilimitado) |
| `--signatures` | auto | Caminho para arquivo JSON de assinaturas |
| `--json` | false | Saída em JSON |
| `--html` | — | Salvar relatório HTML em arquivo |
| `-v, --verbose` | false | Exibir cabeçalhos e detalhes de fingerprint |

### `dns`

| Flag | Padrão | Descrição |
|---|---|---|
| `-t, --target` | — | Domínio para enumerar |
| `--brute` | false | Ativar brute-force de subdomínios |
| `--wordlist` | embutida | Caminho para wordlist customizada |
| `--threads` | `50` | Resolvers DNS concorrentes |
| `--json` | false | Saída em JSON |

### `banner`

| Flag | Padrão | Descrição |
|---|---|---|
| `-t, --target` | — | Host alvo |
| `-p, --ports` | 16 portas | Portas para sondar |
| `--timeout` | `3` | Timeout TCP (segundos) |
| `--threads` | `50` | Conexões concorrentes |
| `--json` | false | Saída em JSON |

### `whois`

| Flag | Padrão | Descrição |
|---|---|---|
| `-t, --target` | — | Domínio para consultar |
| `--timeout` | `15` | Timeout da consulta WHOIS (segundos) |
| `--raw` | false | Incluir texto WHOIS bruto completo |
| `--json` | false | Saída em JSON |

### `vuln`

| Flag | Padrão | Descrição |
|---|---|---|
| `-t, --target` | — | URL base (`https://example.com`) |
| `-T, --templates` | `templates/` | Diretório de templates |
| `--severity` | — | Filtro: `info/low/medium/high/critical` |
| `--timeout` | `10` | Timeout de requisição HTTP (segundos) |
| `--json` | false | Saída em JSON |

---

## 📋 Relatório HTML

Gere um relatório HTML dark-mode profissional e auto-contido — sem servidor, sem dependências:

```bash
webscan scan -t example.com -p 80,443,8080,8443 --html report.html
```

**O relatório inclui:**
- Barra de estatísticas — portas varridas, total de portas abertas, duração da varredura
- Tabela por porta com badge de status, protocolo, versão TLS, cabeçalho do servidor, tecnologias detectadas
- Badges de detecção WAF / CDN
- Tooltips ao passar o mouse em títulos truncados
- 100% offline — arquivo `.html` único, zero chamadas a CDNs externos

---

## 🖥️ TUI Interativa

```bash
webscan tui -t example.com -p 1-1024 --threads 100
```

| Tecla | Ação |
|---|---|
| `Enter` | Painel de detalhes — cabeçalhos completos, certificados TLS, fingerprint |
| `/` | Filtro de busca ao vivo em todas as colunas |
| `s` | Alternar ordenação: porta → status → servidor |
| `o` | Exibir apenas portas **abertas** |
| `f` | Exibir apenas portas **filtradas** |
| `c` | Exibir apenas portas **fechadas** |
| `a` | Exibir **todas** (resetar filtro) |
| `h` ou `?` | Overlay de ajuda |
| `q` / `Esc` | Sair |

Configuração salva automaticamente em `.webscan_config.json` — preferência de ordenação e estilo persistem entre execuções.

---

## 🔄 Pipelines CI/CD

| Workflow | Gatilho | O que faz |
|---|---|---|
| `ci.yml` — Lint | push / PR → `main` | Análise estática com `golangci-lint` |
| `ci.yml` — Test | push / PR → `main` | `go test ./...` no Ubuntu · Windows · macOS |
| `ci.yml` — Security | push / PR → `main` | Varredura SAST com `gosec` |
| `ci.yml` — Build | push / PR → `main` | Cross-compile de artefatos linux/windows/darwin |
| `ci.yml` — Integration | label `run-integration` · `workflow_dispatch` | Harness de integração com Docker Compose |
| `headless-ui.yml` | label `run-headless` · `workflow_dispatch` | Testes TUI headless (Xvfb) |
| `docker.yml` | push `main` · tag `v*` | Build + push para GHCR |
| `release_dryrun.yml` | `workflow_dispatch` | Dry-run do goreleaser (sem publicar) |
| `release.yml` | tag `v*` | Release completo (requer secrets `GH_TOKEN` + `GPG_PRIVATE_KEY`) |

**Disparar workflows manualmente:**

```bash
gh workflow run ci.yml
gh workflow run headless-ui.yml
gh workflow run release_dryrun.yml
```

---

## 🐋 Docker

```bash
# Build local
docker build -t webscan .

# Varredura
docker run --rm webscan scan -t example.com -p 80,443 --json

# Enumeração DNS
docker run --rm webscan dns -t example.com --brute --json

# Varredura de vulnerabilidades
docker run --rm webscan vuln -t https://example.com --json

# Salvar relatório HTML no host
docker run --rm -v "$(pwd):/out" webscan scan -t example.com --html /out/report.html
```

Construído a partir do `scratch` — binário totalmente estático, imagem comprimida de ~8 MB, superfície de ataque mínima.

---

## 🤝 Contribuindo

Veja o [CONTRIBUTING.md](CONTRIBUTING.md) para o guia completo.

**Início rápido:**

```bash
git clone https://github.com/excanear/WebScan.git
cd WebScan
go mod tidy
go test ./... -v              # executar todos os testes
golangci-lint run ./...       # lint
```

**Adicionar um template de detecção** — crie `templates/exposures/meu-check.yaml` e teste localmente:

```bash
webscan vuln -t http://localhost -T templates/meu-check.yaml
```

**Atualizar o banco de assinaturas:**

```bash
python scripts/update_signatures.py
```

---

## 📄 Licença

[MIT](LICENSE) © 2026 WebScan Contributors

---

## 🏆 Créditos

<div align="center">

**Feito por: Escanearcplx**

[![GitHub](https://img.shields.io/badge/GitHub-excanear-181717?style=for-the-badge&logo=github)](https://github.com/excanear)

</div>

---

<div align="center">

**Desenvolvido com ❤️ em Go · tview · cobra · yaml.v3**

*Use com responsabilidade. Escaneie apenas alvos que você possui ou tem permissão escrita explícita para testar.*

</div>
