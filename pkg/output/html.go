package output

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"webscan/internal/scanner"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>WebScan Report — {{.Target}}</title>
  <style>
    :root {
      --bg: #0d1117; --bg2: #161b22; --bg3: #21262d;
      --fg: #e6edf3; --fg2: #8b949e;
      --green: #3fb950; --red: #f85149; --yellow: #d29922;
      --blue: #58a6ff; --purple: #bc8cff; --orange: #ffa657;
      --border: #30363d;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { background: var(--bg); color: var(--fg); font-family: 'Segoe UI', system-ui, monospace; font-size: 14px; }
    header { background: var(--bg2); border-bottom: 1px solid var(--border); padding: 20px 32px; display: flex; align-items: center; gap: 16px; }
    header h1 { font-size: 22px; font-weight: 700; color: var(--blue); letter-spacing: 1px; }
    header .meta { color: var(--fg2); font-size: 12px; }
    .container { max-width: 1200px; margin: 0 auto; padding: 24px 32px; }
    .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px; margin-bottom: 24px; }
    .stat-card { background: var(--bg2); border: 1px solid var(--border); border-radius: 8px; padding: 16px; text-align: center; }
    .stat-card .value { font-size: 28px; font-weight: 700; }
    .stat-card .label { color: var(--fg2); font-size: 11px; text-transform: uppercase; letter-spacing: 1px; margin-top: 4px; }
    .open .value { color: var(--green); }
    .total .value { color: var(--blue); }
    .time .value { color: var(--orange); }
    table { width: 100%; border-collapse: collapse; background: var(--bg2); border-radius: 8px; overflow: hidden; border: 1px solid var(--border); }
    thead { background: var(--bg3); }
    th { padding: 10px 14px; text-align: left; font-size: 11px; text-transform: uppercase; letter-spacing: 1px; color: var(--fg2); border-bottom: 1px solid var(--border); }
    td { padding: 10px 14px; border-bottom: 1px solid var(--border); vertical-align: top; }
    tr:last-child td { border-bottom: none; }
    tr:hover td { background: var(--bg3); }
    .badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 11px; font-weight: 600; }
    .badge-open { background: rgba(63,185,80,.15); color: var(--green); }
    .badge-https { background: rgba(88,166,255,.15); color: var(--blue); }
    .badge-http { background: rgba(255,166,87,.15); color: var(--orange); }
    .badge-tls { background: rgba(188,140,255,.15); color: var(--purple); }
    .tech { background: var(--bg3); border: 1px solid var(--border); border-radius: 4px; padding: 1px 6px; font-size: 11px; margin: 1px; display: inline-block; color: var(--fg2); }
    .title-cell { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .waf { color: var(--red); font-weight: 600; }
    .cdn { color: var(--purple); }
    footer { text-align: center; color: var(--fg2); font-size: 11px; padding: 32px; border-top: 1px solid var(--border); margin-top: 32px; }
    .section-title { font-size: 16px; font-weight: 600; color: var(--fg); margin: 0 0 12px 0; }
  </style>
</head>
<body>
  <header>
    <div>
      <h1>&#x1F50E; WebScan</h1>
      <div class="meta">Generated {{.GeneratedAt}} · Target: <strong>{{.Target}}</strong> · Duration: {{.Duration}}</div>
    </div>
  </header>
  <div class="container">
    <div class="stats">
      <div class="stat-card total"><div class="value">{{.TotalPorts}}</div><div class="label">Ports Scanned</div></div>
      <div class="stat-card open"><div class="value">{{.OpenPorts}}</div><div class="label">Open Ports</div></div>
      <div class="stat-card time"><div class="value">{{.Duration}}</div><div class="label">Scan Time</div></div>
    </div>

    {{if .Results}}
    <p class="section-title">Open Ports</p>
    <table>
      <thead>
        <tr>
          <th>Port</th><th>Proto</th><th>Status</th><th>Server</th>
          <th>Title</th><th>TLS</th><th>Technologies</th><th>WAF/CDN</th>
        </tr>
      </thead>
      <tbody>
        {{range .Results}}
        {{if .Open}}
        <tr>
          <td><strong>{{.Port}}</strong></td>
          <td>
            {{if eq .Protocol "https"}}<span class="badge badge-https">HTTPS</span>
            {{else if eq .Protocol "http"}}<span class="badge badge-http">HTTP</span>
            {{else}}<span class="badge">{{upper .Protocol}}</span>{{end}}
          </td>
          <td><span class="badge badge-open">{{.Status}}</span></td>
          <td>{{or .Server "—"}}</td>
          <td class="title-cell" title="{{.Title}}">{{or .Title "—"}}</td>
          <td>
            {{if .TLSVersion}}<span class="badge badge-tls">{{.TLSVersion}}</span>{{else}}—{{end}}
          </td>
          <td>{{range .Technologies}}<span class="tech">{{.}}</span>{{end}}</td>
          <td>
            {{if .WAF}}<span class="waf">&#x26A0; {{.WAF}}</span><br>{{end}}
            {{if .CDN}}<span class="cdn">&#x2601; {{.CDN}}</span>{{end}}
          </td>
        </tr>
        {{end}}
        {{end}}
      </tbody>
    </table>
    {{else}}
    <p style="color:var(--fg2);text-align:center;padding:40px">No open ports found.</p>
    {{end}}
  </div>
  <footer>WebScan · Scan Report · {{.GeneratedAt}}</footer>
</body>
</html>
`

type htmlData struct {
	Target      string
	GeneratedAt string
	Duration    string
	TotalPorts  int
	OpenPorts   int
	Results     []scanner.PortResult
}

// FormatHTML generates a full HTML report from scan results.
func FormatHTML(results []scanner.PortResult, target string, duration time.Duration) (string, error) {
	openCount := 0
	for _, r := range results {
		if r.Open {
			openCount++
		}
	}

	data := htmlData{
		Target:      target,
		GeneratedAt: time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		Duration:    duration.Round(time.Millisecond).String(),
		TotalPorts:  len(results),
		OpenPorts:   openCount,
		Results:     results,
	}

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"or": func(a, b string) string {
			if a != "" {
				return a
			}
			return b
		},
		"printf": fmt.Sprintf,
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}
