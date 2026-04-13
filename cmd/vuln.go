package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"webscan/internal/templates"

	"github.com/spf13/cobra"
)

var (
	flagVulnTarget   string
	flagVulnDir      string
	flagVulnJSON     bool
	flagVulnTimeout  int
	flagVulnSeverity string
)

var vulnCmd = &cobra.Command{
	Use:   "vuln",
	Short: "Run YAML-based vulnerability/exposure detection templates",
	Long:  "Executes YAML detection templates against a base URL (similar to Nuclei).",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagVulnTarget == "" {
			return fmt.Errorf("target is required (-t)")
		}
		if flagVulnDir == "" {
			flagVulnDir = "templates"
		}

		tmplList, err := templates.LoadDir(flagVulnDir)
		if err != nil {
			return fmt.Errorf("loading templates from %s: %w", flagVulnDir, err)
		}
		if len(tmplList) == 0 {
			return fmt.Errorf("no templates found in %s", flagVulnDir)
		}

		fmt.Printf("[*] Loaded %d templates. Scanning %s ...\n", len(tmplList), flagVulnTarget)

		timeout := time.Duration(flagVulnTimeout) * time.Second
		results := templates.RunAll(tmplList, flagVulnTarget, timeout)

		// filter by severity if requested
		if flagVulnSeverity != "" {
			var filtered []templates.MatchResult
			for _, r := range results {
				if strings.EqualFold(r.Severity, flagVulnSeverity) {
					filtered = append(filtered, r)
				}
			}
			results = filtered
		}

		if flagVulnJSON {
			b, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		if len(results) == 0 {
			fmt.Println("[*] No findings.")
			return nil
		}

		fmt.Printf("\n[!] %d finding(s):\n\n", len(results))
		for _, r := range results {
			sev := strings.ToUpper(r.Severity)
			color := severityColor(r.Severity)
			fmt.Printf("  %s[%s]%s  %-40s  %s\n", color, sev, "\033[0m", r.Name, r.URL)
			if r.Description != "" {
				fmt.Printf("          %s\n", r.Description)
			}
		}
		fmt.Println()
		return nil
	},
}

func severityColor(sev string) string {
	switch strings.ToLower(sev) {
	case "critical":
		return "\033[35m" // magenta
	case "high":
		return "\033[31m" // red
	case "medium":
		return "\033[33m" // yellow
	case "low":
		return "\033[34m" // blue
	default:
		return "\033[36m" // cyan for info
	}
}

func init() {
	rootCmd.AddCommand(vulnCmd)
	vulnCmd.Flags().StringVarP(&flagVulnTarget, "target", "t", "", "Base URL to test (e.g., https://example.com)")
	vulnCmd.Flags().StringVarP(&flagVulnDir, "templates", "T", "templates", "Directory containing YAML templates")
	vulnCmd.Flags().BoolVar(&flagVulnJSON, "json", false, "Output as JSON")
	vulnCmd.Flags().IntVar(&flagVulnTimeout, "timeout", 10, "HTTP request timeout in seconds")
	vulnCmd.Flags().StringVar(&flagVulnSeverity, "severity", "", "Filter results by severity (info/low/medium/high/critical)")
}
