package cmd

import (
	"cloudflare-dyndns/helpers"
	"cloudflare-dyndns/ipify"
	"fmt"
	"github.com/spf13/cobra"
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Print your public IP address.",
	Run: func(cmd *cobra.Command, args []string) {
		ip, err := ipify.New(&cfg).GetPublicIP()
		helpers.FatalError(err)

		fmt.Printf("%s\n", ip)
	},
}

func init() {
	rootCmd.AddCommand(ipCmd)

	ipCmd.Flags().BoolP("help", "h", false, "Show help for the update command.")
}
