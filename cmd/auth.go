package cmd

import (
	"fmt"

	"github.com/r3drun3/prwlrctl/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage JWT authentication (alternative to a static API key)",
}

var (
	loginEmail    string
	loginPassword string
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with email/password and store the resulting JWT tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginEmail == "" || loginPassword == "" {
			return fmt.Errorf("--email and --password are required")
		}
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		access, refresh, err := c.Login(ctx, loginEmail, loginPassword)
		if err != nil {
			return err
		}
		cfg, _ := config.Load()
		if flagBaseURL != "" {
			cfg.BaseURL = flagBaseURL
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = config.DefaultBase
		}
		cfg.AccessToken = access
		cfg.RefreshToken = refresh
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Println("Logged in and saved tokens.")
		return nil
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the stored JWT access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.RefreshToken == "" {
			return fmt.Errorf("no refresh token stored; run 'prwlrctl auth login' first")
		}
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		access, err := c.RefreshToken(ctx, cfg.RefreshToken)
		if err != nil {
			return err
		}
		cfg.AccessToken = access
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Println("Access token refreshed.")
		return nil
	},
}

func init() {
	authLoginCmd.Flags().StringVar(&loginEmail, "email", "", "Account email")
	authLoginCmd.Flags().StringVar(&loginPassword, "password", "", "Account password")
	authCmd.AddCommand(authLoginCmd, authRefreshCmd)
}
