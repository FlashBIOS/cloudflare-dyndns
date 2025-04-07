package cmd

import (
	"cloudflare-dyndns/cloudflare"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"text/tabwriter"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display a list of DNS records for your CloudFlare zone.",
	Long: `Display a list of DNS records for your CloudFlare zone. This command only displays A and AAAA records
so that you can easily find the record you want to update.`,
	Run: func(cmd *cobra.Command, args []string) {
		cloudflareClient := cloudflare.New(&cfg)
		dnsRecords, dnsErrors, err := cloudflareClient.ListDnsRecords()
		if err != nil {
			log.Printf("ERROR: Failed to get DNS records: %s", err)
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			for _, dnsError := range dnsErrors {
				log.Printf("ERROR: DNS record error - %s (code: %d)", dnsError.Message, dnsError.Code)
				_, _ = fmt.Fprintf(os.Stderr, "%s (code: %d)\n", dnsError.Message, dnsError.Code)
			}
			os.Exit(1)
		}

		// Setup the tabwriter for aligned columns.
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tIP\tCOMMENT")

		for _, dnsRecord := range dnsRecords {
			if dnsRecord.Type == "A" || dnsRecord.Type == "AAAA" {
				comment := dnsRecord.Comment
				if comment == "" {
					comment = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", dnsRecord.Name, dnsRecord.IP, comment)
			}
		}

		_ = w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("help", "h", false, "Show help for the update command.")
}
