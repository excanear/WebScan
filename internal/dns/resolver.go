package dns

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"sync"
)

// DNSResult holds all DNS records for a hostname.
type DNSResult struct {
	Hostname string   `json:"hostname"`
	A        []string `json:"a,omitempty"`
	AAAA     []string `json:"aaaa,omitempty"`
	CNAME    string   `json:"cname,omitempty"`
	MX       []string `json:"mx,omitempty"`
	TXT      []string `json:"txt,omitempty"`
	NS       []string `json:"ns,omitempty"`
}

// String returns a compact JSON representation.
func (r DNSResult) String() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

// Enumerate resolves all standard DNS record types for host concurrently.
func Enumerate(ctx context.Context, host string) DNSResult {
	result := DNSResult{Hostname: host}
	resolver := net.DefaultResolver

	var mu sync.Mutex
	var wg sync.WaitGroup

	// A + AAAA via LookupHost
	wg.Add(1)
	go func() {
		defer wg.Done()
		addrs, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		for _, a := range addrs {
			if strings.Contains(a, ":") {
				result.AAAA = append(result.AAAA, a)
			} else {
				result.A = append(result.A, a)
			}
		}
	}()

	// CNAME
	wg.Add(1)
	go func() {
		defer wg.Done()
		cname, err := resolver.LookupCNAME(ctx, host)
		if err != nil {
			return
		}
		cname = strings.TrimSuffix(cname, ".")
		// Only set CNAME if it differs from the queried host
		if !strings.EqualFold(cname, host) {
			mu.Lock()
			result.CNAME = cname
			mu.Unlock()
		}
	}()

	// MX
	wg.Add(1)
	go func() {
		defer wg.Done()
		mxs, err := resolver.LookupMX(ctx, host)
		if err != nil {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		for _, mx := range mxs {
			result.MX = append(result.MX, mx.Host)
		}
	}()

	// TXT
	wg.Add(1)
	go func() {
		defer wg.Done()
		txts, err := resolver.LookupTXT(ctx, host)
		if err != nil {
			return
		}
		mu.Lock()
		result.TXT = txts
		mu.Unlock()
	}()

	// NS
	wg.Add(1)
	go func() {
		defer wg.Done()
		nss, err := resolver.LookupNS(ctx, host)
		if err != nil {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		for _, ns := range nss {
			result.NS = append(result.NS, strings.TrimSuffix(ns.Host, "."))
		}
	}()

	wg.Wait()
	return result
}
