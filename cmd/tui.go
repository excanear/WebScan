package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webscan/internal/scanner"
	"webscan/internal/ui"
	"webscan/internal/web"

	"github.com/spf13/cobra"
)

var (
	flagTUITarget     string
	flagTUIPorts      string
	flagTUIThreads    int
	flagTUITimeout    int
	flagTUIStyle      string
	flagTUISignatures string
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Run interactive TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagTUITarget == "" {
			return fmt.Errorf("target is required")
		}
		ports, err := parsePorts(flagTUIPorts)
		if err != nil {
			return err
		}
		cfg := scanner.Config{
			Target:  flagTUITarget,
			Ports:   ports,
			Threads: flagTUIThreads,
			Timeout: time.Duration(flagTUITimeout) * time.Second,
		}

		// load optional signatures DB for fingerprinting
		if err := web.LoadSignatures(flagTUISignatures); err != nil {
			return err
		}
		sc := scanner.NewScanner(cfg)
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		return ui.RunTUI(ctx, sc, cfg, flagTUIStyle)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().StringVarP(&flagTUITarget, "target", "t", "", "Target domain or IP (required)")
	tuiCmd.Flags().StringVarP(&flagTUIPorts, "ports", "p", "80,443", "Comma-separated ports or ranges (e.g., 1-1024,80,443)")
	tuiCmd.Flags().IntVar(&flagTUIThreads, "threads", 100, "Number of concurrent workers")
	tuiCmd.Flags().IntVar(&flagTUITimeout, "timeout", 2, "Timeout in seconds for network ops")
	tuiCmd.Flags().StringVar(&flagTUIStyle, "style", "default", "TUI style preset (vaporwave, default)")
	tuiCmd.Flags().StringVar(&flagTUISignatures, "signatures", "", "Path to signatures JSON file for fingerprinting")
}
