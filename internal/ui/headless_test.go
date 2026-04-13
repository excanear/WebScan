package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"webscan/internal/scanner"
	"webscan/pkg/output"
)

func TestRunHeadless(t *testing.T) {
	// simple HTTP server that returns a title
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "unittest")
		io.WriteString(w, "<html><head><title>Hi</title></head><body>Hello</body></html>")
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	host := u.Hostname()
	portStr := u.Port()
	port, _ := strconv.Atoi(portStr)

	cfg := scanner.Config{
		Target:  host,
		Ports:   []int{port},
		Threads: 1,
		Timeout: 2 * time.Second,
		Retries: 0,
	}
	sc := scanner.NewScanner(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	var buf bytes.Buffer
	if err := RunHeadless(ctx, sc, cfg, &buf); err != nil {
		t.Fatalf("RunHeadless failed: %v", err)
	}

	var out output.JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	if out.PortsScanned != 1 {
		t.Fatalf("expected 1 ports scanned, got %d", out.PortsScanned)
	}
	if out.OpenCount != 1 {
		t.Fatalf("expected 1 open, got %d", out.OpenCount)
	}
	if len(out.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out.Results))
	}
	if !strings.Contains(out.Results[0].Title, "Hi") {
		t.Fatalf("expected title to contain Hi, got %q", out.Results[0].Title)
	}
}
