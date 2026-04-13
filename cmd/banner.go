package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webscan/internal/scanner"

	"github.com/spf13/cobra"
)

var (
	flagBannerTarget  string
	flagBannerPorts   string
	flagBannerJSON    bool
	flagBannerTimeout int
	flagBannerThreads int
)

var bannerCmd = &cobra.Command{
	Use:   "banner",
	Short: "TCP banner grabbing for service identification",
	Long:  "Connects raw TCP to each port and reads the server banner for service identification.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagBannerTarget == "" {
			return fmt.Errorf("target is required (-t)")
		}

		ports, err := parsePorts(flagBannerPorts)
		if err != nil {
			return err
		}

		timeout := time.Duration(flagBannerTimeout) * time.Second
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		results := scanner.GrabBanners(ctx, flagBannerTarget, ports, flagBannerThreads, timeout)

		if flagBannerJSON {
			// filter only open
			var open []scanner.BannerResult
			for _, r := range results {
				if r.Open {
					open = append(open, r)
				}
			}
			b, _ := json.MarshalIndent(open, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		fmt.Printf("\n  Banner Grab: %s\n", flagBannerTarget)
		for _, r := range results {
			if r.Open {
				svc := r.Service
				if svc == "" {
					svc = "unknown"
				}
				fmt.Printf("  [OPEN] %-5d  %-12s  %s\n", r.Port, svc, truncate(r.Banner, 80))
			}
		}
		fmt.Println()
		return nil
	},
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func init() {
	rootCmd.AddCommand(bannerCmd)
	bannerCmd.Flags().StringVarP(&flagBannerTarget, "target", "t", "", "Target host (required)")
	bannerCmd.Flags().StringVarP(&flagBannerPorts, "ports", "p", "21,22,23,25,80,110,143,443,445,587,3306,3389,5432,6379,8080,8443,27017", "Ports to banner-grab")
	bannerCmd.Flags().BoolVar(&flagBannerJSON, "json", false, "Output as JSON")
	bannerCmd.Flags().IntVar(&flagBannerTimeout, "timeout", 3, "Connection timeout in seconds")
	bannerCmd.Flags().IntVar(&flagBannerThreads, "threads", 50, "Concurrent workers")
}
