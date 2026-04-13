package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// BannerResult holds a TCP banner grabbed from a port.
type BannerResult struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Banner  string `json:"banner,omitempty"`
	Service string `json:"service,omitempty"`
}

// GrabBanner dials host:port via raw TCP and reads the initial server banner.
// It sends a protocol-appropriate probe string where useful.
func GrabBanner(ctx context.Context, host string, port int, timeout time.Duration) BannerResult {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return BannerResult{Port: port, Open: false}
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	// Send a probe for ports that require client-first
	if probe := probeFor(port); probe != "" {
		fmt.Fprint(conn, probe) // #nosec G104
	}

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	raw := string(buf[:n])
	banner := sanitizeBanner(raw)

	return BannerResult{
		Port:    port,
		Open:    true,
		Banner:  banner,
		Service: inferService(port, raw),
	}
}

// GrabBanners concurrently grabs banners for a list of ports.
func GrabBanners(ctx context.Context, host string, ports []int, threads int, timeout time.Duration) []BannerResult {
	if threads <= 0 {
		threads = 50
	}

	type job struct{ port int }
	jobs := make(chan job, threads)
	out := make(chan BannerResult, len(ports))

	for i := 0; i < threads; i++ {
		go func() {
			for j := range jobs {
				out <- GrabBanner(ctx, host, j.port, timeout)
			}
		}()
	}
	for _, p := range ports {
		jobs <- job{p}
	}
	close(jobs)

	var results []BannerResult
	for range ports {
		results = append(results, <-out)
	}
	close(out)
	return results
}

func probeFor(port int) string {
	switch port {
	case 25, 587: // SMTP — wait for banner
		return ""
	case 21: // FTP — wait
		return ""
	case 110: // POP3 — wait
		return ""
	case 143: // IMAP — wait
		return ""
	case 22: // SSH — wait
		return ""
	case 3306: // MySQL — wait
		return ""
	case 6379: // Redis
		return "*1\r\n$4\r\nPING\r\n"
	default: // HTTP-like ports
		return "HEAD / HTTP/1.0\r\nHost: localhost\r\n\r\n"
	}
}

func sanitizeBanner(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(' ')
		} else if r >= 32 && r < 127 {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func inferService(port int, banner string) string {
	bl := strings.ToLower(banner)
	switch {
	case strings.HasPrefix(bl, "ssh-"):
		return "SSH"
	case port == 21, strings.HasPrefix(bl, "220") && strings.Contains(bl, "ftp"):
		return "FTP"
	case port == 25 || port == 587, strings.HasPrefix(bl, "220") && strings.Contains(bl, "smtp"):
		return "SMTP"
	case port == 110 || strings.HasPrefix(bl, "+ok"):
		return "POP3"
	case port == 143 || strings.HasPrefix(bl, "* ok"):
		return "IMAP"
	case port == 3306 || (len(banner) > 4 && banner[4] == 0x0a):
		return "MySQL"
	case port == 5432:
		return "PostgreSQL"
	case port == 6379 || strings.HasPrefix(bl, "+pong"):
		return "Redis"
	case port == 27017:
		return "MongoDB"
	case port == 11211:
		return "Memcached"
	case strings.Contains(bl, "http/") || strings.Contains(bl, "http "):
		return "HTTP"
	case port == 8080 || port == 8443 || port == 8000:
		return "HTTP-ALT"
	default:
		return ""
	}
}
