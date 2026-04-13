package ui

import (
	"context"
	"io"
	"time"

	"webscan/internal/scanner"
	"webscan/pkg/output"
)

// RunHeadless performs a non-interactive scan using the provided Scanner and
// writes JSON output to the provided writer. Useful for CI/smoke tests.
func RunHeadless(ctx context.Context, sc *scanner.Scanner, cfg scanner.Config, w io.Writer) error {
	start := time.Now()
	ch := make(chan scanner.PortResult)

	// Start streaming (non-blocking)
	go func() {
		_ = sc.StartStream(ctx, ch)
	}()

	var results []scanner.PortResult
	for pr := range ch {
		results = append(results, pr)
	}

	dur := time.Since(start)
	b, err := output.FormatJSON(results, cfg.Target, dur)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
