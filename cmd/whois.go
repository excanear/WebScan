package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"webscan/internal/whois"

	"github.com/spf13/cobra"
)

var (
	flagWhoisTarget  string
	flagWhoisJSON    bool
	flagWhoisTimeout int
	flagWhoisRaw     bool
)

var whoisCmd = &cobra.Command{
	Use:   "whois",
	Short: "WHOIS lookup with registrar, dates and nameservers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagWhoisTarget == "" {
			return fmt.Errorf("target is required (-t)")
		}

		timeout := time.Duration(flagWhoisTimeout) * time.Second
		result, err := whois.Lookup(flagWhoisTarget, timeout)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[!] WHOIS warning: %v\n", err)
		}

		if flagWhoisJSON {
			if !flagWhoisRaw {
				result.Raw = "" // omit raw in JSON unless requested
			}
			b, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		}

		fmt.Printf("\n WHOIS: %s\n", result.Domain)
		fmt.Println(strings.Repeat("─", 50))
		if result.Registrar != "" {
			fmt.Printf("  Registrar   : %s\n", result.Registrar)
		}
		if result.CreatedDate != "" {
			fmt.Printf("  Created     : %s\n", result.CreatedDate)
		}
		if result.ExpiryDate != "" {
			fmt.Printf("  Expires     : %s\n", result.ExpiryDate)
		}
		if len(result.NameServers) > 0 {
			fmt.Printf("  Nameservers : %s\n", strings.Join(result.NameServers, ", "))
		}
		if len(result.Status) > 0 {
			fmt.Printf("  Status      : %s\n", strings.Join(result.Status, ", "))
		}
		if flagWhoisRaw && result.Raw != "" {
			fmt.Printf("\n--- Raw WHOIS ---\n%s\n", result.Raw)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoisCmd)
	whoisCmd.Flags().StringVarP(&flagWhoisTarget, "target", "t", "", "Domain to look up (required)")
	whoisCmd.Flags().BoolVar(&flagWhoisJSON, "json", false, "Output as JSON")
	whoisCmd.Flags().BoolVar(&flagWhoisRaw, "raw", false, "Include raw WHOIS text")
	whoisCmd.Flags().IntVar(&flagWhoisTimeout, "timeout", 15, "WHOIS query timeout in seconds")
}
