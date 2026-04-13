package scanner

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkScanPort_Open(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	host := srv.Listener.Addr().(*net.TCPAddr).IP.String()
	port := srv.Listener.Addr().(*net.TCPAddr).Port

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GrabBanner(context.Background(), host, port, time.Second)
		}
	})
}

func BenchmarkGrabBanner_Closed(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GrabBanner(context.Background(), "127.0.0.1", 19999, 100*time.Millisecond)
	}
}
