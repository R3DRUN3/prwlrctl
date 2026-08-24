package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/r3drun3/prwlrctl/internal/client"
	"github.com/r3drun3/prwlrctl/internal/config"
	"github.com/r3drun3/prwlrctl/internal/jsonapi"
	"github.com/spf13/cobra"
)

var (
	flagBaseURL string
	flagAPIKey  string
	flagToken   string
	flagOutput  string
	flagTimeout time.Duration
)

var rootCmd = &cobra.Command{
	Use:           "prwlrctl",
	Short:         "Talk to the Prowler Server API",
	SilenceUsage:  true,
	SilenceErrors: false,
	Version:       version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "Prowler API base URL (env PROWLER_BASE_URL, default "+config.DefaultBase+")")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Prowler API key (env PROWLER_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "JWT access token, overrides API key (env PROWLER_TOKEN)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table|json")
	rootCmd.PersistentFlags().DurationVar(&flagTimeout, "timeout", 60*time.Second, "Per-request timeout")

	rootCmd.AddCommand(configureCmd, authCmd, providersCmd, scansCmd, findingsCmd)
}

// newClient resolves config precedence (flags > env > file) and builds a
// ready-to-use API client for a command.
func newClient() *client.Client {
	cfg := config.Resolve(flagBaseURL, flagAPIKey, flagToken)
	return client.New(cfg.BaseURL, cfg.APIKey, cfg.AccessToken, flagTimeout)
}

// cmdContext gives every command a bounded context so hung requests never
// block a cron job forever, independent of the per-HTTP-call timeout.
func cmdContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), flagTimeout*2)
}

// printPaginationHint writes a one-line stderr note when a list response is
// paginated and there's more than one page, so humans notice a truncated
// view and know to pass --all; scripts consuming stdout are unaffected.
func printPaginationHint(doc jsonapi.Document) {
	if p, ok := doc.Pagination(); ok && p.Pages > 1 {
		fmt.Fprintf(os.Stderr, "showing page %d of %d (total %d results) — pass --all to fetch everything\n", p.Page, p.Pages, p.Count)
	}
}
