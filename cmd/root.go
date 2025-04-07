package cmd

import (
	"cloudflare-dyndns/config"
	"cloudflare-dyndns/helpers"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var cfg config.Config
var configFile string
var logger zerolog.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudflare-dyndns",
	Short: "A brief description of your application",
	Long: `Update your Cloudflare DNS records with the current or specified IP address.
This command simplifies managing dynamic IP changes.
Examples:
  cloudflare-dyns list
  cloudflare-dyns update
  cloudflare-dyns update --name my-record --ip 1.2.3.4`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default searches for ./.cloudflare-dyndns, ~/.cloudflare-dyndns, or ~/.config/cloudflare-dyndns/.cloudflare-dyndns)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		// Use a config file from the flag.
		viper.SetConfigFile(configFile)
		viper.SetConfigType("toml")
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		helpers.FatalError(err)

		// Candidate paths in the specified order.
		candidatePaths := []string{
			"./.cloudflare-dyndns",
			filepath.Join(home, ".cloudflare-dyndns"),
			filepath.Join(home, ".config", "cloudflare-dyndns", ".cloudflare-dyndns"),
		}

		var foundConfigPath string
		for _, path := range candidatePaths {
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				foundConfigPath = path
				break
			}
		}

		if foundConfigPath == "" {
			fmt.Printf("%s\n", color.With(color.Red, fmt.Sprintf("ERROR: Config file not found in paths: %v", candidatePaths)))
			os.Exit(1)
		}

		viper.SetConfigFile(foundConfigPath)
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		msg := color.With(color.Gray, fmt.Sprintf("Using config file: %s\n", viper.ConfigFileUsed()))
		fmt.Printf("%s", msg)
	} else {
		msg := color.With(color.Red, fmt.Sprintf("ERROR: Config file cannot be loaded: %v\n", err))
		fmt.Printf("%s", msg)
		os.Exit(1)
	}

	// Set the configuration defaults.
	viper.SetDefault("main.user_agent", "cloudflare-dyndns/1.0.0")
	viper.SetDefault("main.log_file_path", "./")
	viper.SetDefault("main.home_gateway", "")
	viper.SetDefault("cloudflare.api_token", "")
	viper.SetDefault("cloudflare.base_url", "https://api.cloudflare.com/client/v4")
	viper.SetDefault("cloudflare.zone_id", "")
	viper.SetDefault("cloudflare.update_records", []string{})
	viper.SetDefault("ipify.url", "https://api64.ipify.org")

	// Populate the config struct.
	cfg = config.Config{
		APIToken:      viper.GetString("cloudflare.api_token"),
		BaseURL:       viper.GetString("cloudflare.base_url"),
		ZoneID:        viper.GetString("cloudflare.zone_id"),
		UpdateRecords: viper.GetStringSlice("cloudflare.update_records"),
		UserAgent:     viper.GetString("main.user_agent"),
		LogFilePath:   viper.GetString("main.log_file_path"),
		HomeGateway:   viper.GetString("main.home_gateway"),
		IpifyURL:      viper.GetString("ipify.url"),
	}

	// Required config values.
	if cfg.APIToken == "" || cfg.ZoneID == "" || len(cfg.UpdateRecords) == 0 {
		msg := color.With(color.Red, "Please provide a valid config file at ~/.cloudflare-dyndns or use the --config flag to specify a config file.\n")
		fmt.Printf("%s", msg)
		os.Exit(1)
	}

	// Set up the logger.
	if cfg.LogFilePath != "" {
		logFilePath := filepath.Join(cfg.LogFilePath, "cloudflare-dyndns.log")
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			msg := color.With(color.Red, fmt.Sprintf("Error opening log file: %v\n", err))
			_, _ = fmt.Fprintf(os.Stderr, "%s", msg)
			os.Exit(1)
		}

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        logFile,
			TimeFormat: time.RFC3339,
			NoColor:    true,
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(i.(string) + ":")
			},
			//FormatErrFieldName: func(i interface{}) string {
			//	return color.With(color.Green, "ERROR: ")
			//},
			//FormatErrFieldValue: func(i interface{}) string {
			//	return color.With(color.Green, i.(string))
			//},
		}).With().Timestamp().Logger()
	}
}
