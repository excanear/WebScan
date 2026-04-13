package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"webscan/internal/dns"

	"github.com/spf13/cobra"
)

var (
	flagDNSTarget   string
	flagDNSJSON     bool
	flagDNSWordlist string
	flagDNSThreads  int
	flagDNSBrute    bool
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS enumeration and subdomain discovery",
	Long:  "Enumerate A, AAAA, CNAME, MX, TXT and NS records. Optionally brute-force subdomains.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDNSTarget == "" {
			return fmt.Errorf("target is required (-t)")
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// DNS record enumeration
		result := dns.Enumerate(ctx, flagDNSTarget)
		if flagDNSJSON {
			b, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(b))
		} else {
			printDNSResult(result)
		}

		// Subdomain brute-force
		if flagDNSBrute {
			fmt.Printf("\n[*] Brute-forcing subdomains of %s ...\n", flagDNSTarget)
			subs, err := dns.BruteForce(ctx, flagDNSTarget, flagDNSWordlist, flagDNSThreads)
			if err != nil {
				return fmt.Errorf("bruteforce: %w", err)
			}
			if flagDNSJSON {
				b, _ := json.MarshalIndent(subs, "", "  ")
				fmt.Println(string(b))
			} else {
				fmt.Printf("\n[+] Found %d subdomains:\n", len(subs))
				for _, s := range subs {
					fmt.Printf("  %-45s  %v\n", s.Subdomain, s.IPs)
				}
			}
		}
		return nil
	},
}

func printDNSResult(r dns.DNSResult) {
	fmt.Printf("\n DNS Enumeration: %s\n", r.Hostname)
	fmt.Println(strings.Repeat("─", 50))
	if len(r.A) > 0 {
		fmt.Printf("  A      : %v\n", r.A)
	}
	if len(r.AAAA) > 0 {
		fmt.Printf("  AAAA   : %v\n", r.AAAA)
	}
	if r.CNAME != "" {
		fmt.Printf("  CNAME  : %s\n", r.CNAME)
	}
	if len(r.NS) > 0 {
		fmt.Printf("  NS     : %v\n", r.NS)
	}
	if len(r.MX) > 0 {
		fmt.Printf("  MX     : %v\n", r.MX)
	}
	if len(r.TXT) > 0 {
		fmt.Printf("  TXT    : \n")
		for _, t := range r.TXT {
			fmt.Printf("    %q\n", t)
		}
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(dnsCmd)
	dnsCmd.Flags().StringVarP(&flagDNSTarget, "target", "t", "", "Domain to enumerate (required)")
	dnsCmd.Flags().BoolVar(&flagDNSJSON, "json", false, "Output as JSON")
	dnsCmd.Flags().BoolVar(&flagDNSBrute, "brute", false, "Enable subdomain brute-force")
	dnsCmd.Flags().StringVar(&flagDNSWordlist, "wordlist", "", "Path to wordlist for brute-force (built-in if omitted)")
	dnsCmd.Flags().IntVar(&flagDNSThreads, "threads", 50, "Concurrent threads for brute-force")
}
