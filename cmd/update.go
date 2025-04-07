package cmd

import (
	"cloudflare-dyndns/cloudflare"
	"cloudflare-dyndns/ipify"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/apex/log"
	"github.com/jackpal/gateway"
	"github.com/spf13/cobra"
	"os"
	"slices"
	"strings"
	"time"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update your IP address in Cloudflare",
	Run: func(cmd *cobra.Command, args []string) {
		// Get the gateway information, if configured.
		if homeGateway := cfg.HomeGateway; len(homeGateway) > 0 {
			cg, err := gateway.DiscoverGateway()
			FatalError(err)
			currentGateway := cg.String()

			if homeGateway != currentGateway {
				logger.Info().Msg(fmt.Sprintf("current gateway %s does not match home gateway %s", currentGateway, homeGateway))
				fmt.Print(color.With(color.Yellow,
					fmt.Sprintf("Warning: Your current gateway (%s) does not match your home gateway (%s). Exiting.\n", currentGateway, homeGateway)))
				os.Exit(0)
			}
		}

		// Get the current IP address to use
		var currentIp struct {
			Addr   string
			IsIPv4 bool
		}
		var err error
		if cmd.Flag("ip").Value.String() == "" {
			currentIp.Addr, err = ipify.New(&cfg).GetPublicIP()
			if err != nil {
				message := fmt.Sprintf("Failed to retrieve public IP: %s", err)
				logger.Error().Msg(message)
				fmt.Println(message)
			}
		} else {
			currentIp.Addr = cmd.Flag("ip").Value.String()
		}

		if strings.Contains(currentIp.Addr, ".") {
			currentIp.IsIPv4 = true
		} else {
			currentIp.IsIPv4 = false
		}

		// Update CloudFlare.
		cloudflareClient := cloudflare.New(&cfg)
		dnsRecords, dnsErrors, err := cloudflareClient.GetDnsRecords()
		if err != nil {
			message := fmt.Sprintf("Failed to get DNS records: %s", err)
			log.Error(message)
			fmt.Println(message)

			for _, dnsError := range dnsErrors {
				message := fmt.Sprintf("DNS record error: %s (code: %d)", dnsError.Message, dnsError.Code)
				logger.Error().Msg(message)
				fmt.Println(message)
			}
			os.Exit(1)
		}

		var didFindName = false
		var names []string
		if cmd.Flag("name").Value.String() != "" {
			names = append(names, cmd.Flag("name").Value.String())
		} else {
			names = cfg.UpdateRecords
		}

		for _, dnsRecord := range dnsRecords {
			if slices.Contains(names, dnsRecord.Name) {
				didFindName = true
				if currentIp.Addr != dnsRecord.IP {
					newComment := cmd.Flag("comment").Value.String()
					fmt.Printf("Updating IP address from \"%s\" to \"%s\".\n", dnsRecord.IP, currentIp.Addr)
					dnsRecord.IP = currentIp.Addr
					dnsRecord.Comment = newComment
					dnsRecord.Type = map[bool]string{true: "A", false: "AAAA"}[currentIp.IsIPv4]

					dnsErrors, err = cloudflareClient.UpdateDnsRecord(dnsRecord)
					if err != nil {
						message := fmt.Sprintf("Failed to update DNS record: %s", err)
						logger.Error().Msg(message)
						fmt.Println(message)

						for _, dnsError := range dnsErrors {
							message := fmt.Sprintf("DNS update error - %s (code: %d)", dnsError.Message, dnsError.Code)
							logger.Error().Msg(message)
							fmt.Println(message)
						}
						os.Exit(1)
					}
					message := fmt.Sprintf("IP address for \"%s\" updated.", dnsRecord.Name)
					logger.Info().Msg(message)
					fmt.Println(message)
				} else {
					message := fmt.Sprintf("IP address for \"%s\" is already up to date.", dnsRecord.Name)
					logger.Info().Msg(message)
					fmt.Println(message)
				}
			}
		}

		if !didFindName {
			message := fmt.Sprintf("Could not find DNS record with name \"%s\".", strings.Join(names, "\", \""))
			logger.Warn().Msg(message)
			fmt.Println(message)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("name", "n", "", "The name of the DNS record to update. If not specified, the name will be read from the config file.")
	updateCmd.Flags().StringP("ip", "i", "", "Update the IP address of the DNS record to this value. If not specified, the current public IP address will be used.")
	updateCmd.Flags().StringP("comment", "c", getDefaultComment(), "Update the comment of the DNS record.")
	updateCmd.Flags().BoolP("help", "h", false, "Show help for the update command.")
}

func getDefaultComment() string {
	return "Updated " + time.Now().UTC().Format("2006-01-02T15:04:05")
}
