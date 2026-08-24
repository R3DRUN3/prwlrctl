package cmd

import (
	"fmt"

	"github.com/r3drun3/prwlrctl/internal/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Save a base URL and API key to the local config file",
	Long: `Persists settings to ~/.config/prwlrctl/config.json so you don't need to
pass --base-url/--api-key on every call. For services and cron jobs, prefer
the PROWLER_API_KEY / PROWLER_BASE_URL environment variables instead of a
config file on disk.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagBaseURL == "" && flagAPIKey == "" {
			return fmt.Errorf("provide at least one of --base-url or --api-key")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if flagBaseURL != "" {
			cfg.BaseURL = flagBaseURL
		}
		if flagAPIKey != "" {
			cfg.APIKey = flagAPIKey
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = config.DefaultBase
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		path, _ := config.Path()
		fmt.Println("Saved configuration to", path)
		return nil
	},
}
