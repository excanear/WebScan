package whois

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Result holds parsed WHOIS data for a domain.
type Result struct {
	Domain      string   `json:"domain"`
	Registrar   string   `json:"registrar,omitempty"`
	CreatedDate string   `json:"created_date,omitempty"`
	ExpiryDate  string   `json:"expiry_date,omitempty"`
	NameServers []string `json:"name_servers,omitempty"`
	Status      []string `json:"status,omitempty"`
	Raw         string   `json:"raw"`
}

// Lookup queries WHOIS for domain.
// It first contacts whois.iana.org to discover the TLD's WHOIS server,
// then queries that server for detailed information.
func Lookup(domain string, timeout time.Duration) (Result, error) {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	result := Result{Domain: domain}

	// Step 1: ask IANA for the authoritative WHOIS server
	rawIANA, err := query("whois.iana.org", domain, timeout)
	if err != nil {
		return result, fmt.Errorf("iana whois: %w", err)
	}

	whoisServer := extractField(rawIANA, "whois")
	if whoisServer == "" {
		result.Raw = rawIANA
		return result, nil
	}

	// Step 2: query the TLD WHOIS server
	raw, err := query(whoisServer, domain, timeout)
	if err != nil {
		// Fall back to IANA data
		result.Raw = rawIANA
		return result, nil
	}

	result.Raw = raw
	result = parseFields(result, raw)
	return result, nil
}

// query sends a WHOIS request to host:43 and returns the raw response.
func query(host, domain string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("tcp", host+":43", timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	fmt.Fprintf(conn, "%s\r\n", domain) // #nosec G104
	b, err := io.ReadAll(io.LimitReader(conn, 256*1024))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseFields(r Result, raw string) Result {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		switch {
		case hasPrefix(lower, "registrar:"):
			r.Registrar = extractField(line, "registrar")
		case hasPrefix(lower, "creation date:"), hasPrefix(lower, "created:"), hasPrefix(lower, "registered:"):
			if r.CreatedDate == "" {
				r.CreatedDate = extractField(line, "")
			}
		case hasPrefix(lower, "expiry date:"), hasPrefix(lower, "expiration date:"), hasPrefix(lower, "registrar registration expiration date:"):
			if r.ExpiryDate == "" {
				r.ExpiryDate = extractField(line, "")
			}
		case hasPrefix(lower, "name server:"), hasPrefix(lower, "nserver:"):
			if ns := extractField(line, ""); ns != "" {
				r.NameServers = append(r.NameServers, strings.ToLower(ns))
			}
		case hasPrefix(lower, "domain status:"):
			if s := extractField(line, ""); s != "" {
				r.Status = append(r.Status, s)
			}
		}
	}
	return r
}

func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func extractField(line, _ string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
