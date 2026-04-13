package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"webscan/internal/scanner"
	"webscan/internal/web"
	"webscan/pkg/output"

	"github.com/spf13/cobra"
)

var (
	flagTarget     string
	flagPorts      string
	flagThreads    int
	flagTimeout    int
	flagJSON       bool
	flagVerbose    bool
	flagRetries    int
	flagRate       int
	flagSignatures string
	flagList       string
	flagHTML       string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan target(s) for web services",
	RunE: func(cmd *cobra.Command, args []string) error {
		targets, err := resolveTargets(flagTarget, flagList)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return fmt.Errorf("target is required (-t) or provide a list (-l)")
		}
		ports, err := parsePorts(flagPorts)
		if err != nil {
			return err
		}
		// load optional signatures DB
		if err := web.LoadSignatures(flagSignatures); err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		for _, target := range targets {
			cfg := scanner.Config{
				Target:    target,
				Ports:     ports,
				Threads:   flagThreads,
				Timeout:   time.Duration(flagTimeout) * time.Second,
				Verbose:   flagVerbose,
				JSON:      flagJSON,
				Retries:   flagRetries,
				RateLimit: flagRate,
			}
			sc := scanner.NewScanner(cfg)
			start := time.Now()
			results, err := sc.Start(ctx)
			duration := time.Since(start)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[!] %s: %v\n", target, err)
				continue
			}
			if flagHTML != "" {
				html, err := output.FormatHTML(results, target, duration)
				if err != nil {
					return err
				}
				outFile := flagHTML
				if len(targets) > 1 {
					outFile = strings.ReplaceAll(target, ".", "_") + "_" + flagHTML
				}
				if err := os.WriteFile(outFile, []byte(html), 0600); err != nil {
					return err
				}
				fmt.Printf("[+] HTML report saved to %s\n", outFile)
				continue
			}
			if flagJSON {
				b, err := output.FormatJSON(results, target, duration)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				continue
			}
			fmt.Print(output.FormatText(results, flagVerbose, target, duration))
		}
		return nil
	},
}

// resolveTargets returns the list of targets from -t flag, -l file, or stdin.
func resolveTargets(single, listPath string) ([]string, error) {
	var targets []string
	if single != "" {
		targets = append(targets, single)
	}
	if listPath == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				targets = append(targets, line)
			}
		}
		return targets, scanner.Err()
	}
	if listPath != "" {
		f, err := os.Open(listPath) // #nosec G304
		if err != nil {
			return nil, err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				targets = append(targets, line)
			}
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}
	return targets, nil
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVarP(&flagTarget, "target", "t", "", "Target domain or IP (required)")
	scanCmd.Flags().StringVarP(&flagPorts, "ports", "p", "80,443", "Comma-separated ports or ranges (e.g., 1-1024,80,443)")
	scanCmd.Flags().IntVar(&flagThreads, "threads", 100, "Number of concurrent workers")
	scanCmd.Flags().IntVar(&flagTimeout, "timeout", 2, "Timeout in seconds for network ops")
	scanCmd.Flags().IntVar(&flagRetries, "retries", 2, "Number of retries for TCP connect (default 2)")
	scanCmd.Flags().IntVar(&flagRate, "rate", 0, "Rate limit (connections per second). 0 = unlimited")
	scanCmd.Flags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	scanCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose output")
	scanCmd.Flags().StringVar(&flagSignatures, "signatures", "", "Path to signatures JSON file for fingerprinting")
	scanCmd.Flags().StringVarP(&flagList, "list", "l", "", "File with list of targets (one per line). Use '-' for stdin")
	scanCmd.Flags().StringVar(&flagHTML, "html", "", "Save HTML report to specified file (e.g., report.html)")
}

// parsePorts parses comma separated ports and ranges into a slice of ints.
func parsePorts(spec string) ([]int, error) {
	var ports []int
	if spec == "" {
		return ports, nil
	}
	tokens := strings.Split(spec, ",")
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if strings.Contains(t, "-") {
			parts := strings.SplitN(t, "-", 2)
			start, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
			continue
		}
		p, err := strconv.Atoi(t)
		if err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	return ports, nil
}
