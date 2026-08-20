package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/r3drun3/prwlrctl/internal/client"
	"github.com/r3drun3/prwlrctl/internal/output"
	"github.com/spf13/cobra"
)

var scansCmd = &cobra.Command{
	Use:   "scans",
	Short: "List, inspect and launch Prowler scans",
}

var (
	scanFilterProvider string
	scanFilterState    string
	scanPage           int
	scanPageSize       int
	scanAll            bool
)

var scanColumns = []output.Column{
	output.IDColumn(),
	output.Attr("NAME", "name"),
	output.Attr("STATE", "state"),
	output.Attr("STARTED", "started_at"),
	output.Attr("COMPLETED", "completed_at"),
}

var scansListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scans",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		filters := map[string]string{
			"provider": scanFilterProvider,
			"state":    scanFilterState,
		}

		if scanAll {
			resources, err := c.ListAll(ctx, "/scans", client.BuildQuery(filters))
			if err != nil {
				return err
			}
			if flagOutput == string(output.JSON) {
				return output.JSONPretty(os.Stdout, resources)
			}
			output.RenderTable(os.Stdout, resources, scanColumns)
			return nil
		}

		doc, err := c.ListScans(ctx, filters, scanPage, scanPageSize)
		if err != nil {
			return err
		}
		if flagOutput == string(output.JSON) {
			return output.JSONPretty(os.Stdout, doc)
		}
		printPaginationHint(doc)
		resources, err := doc.Many()
		if err != nil {
			return err
		}
		output.RenderTable(os.Stdout, resources, scanColumns)
		return nil
	},
}

var scansGetCmd = &cobra.Command{
	Use:   "get <scan-id>",
	Short: "Show a single scan by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		doc, err := c.GetScan(ctx, args[0])
		if err != nil {
			return err
		}
		if flagOutput == string(output.JSON) {
			return output.JSONPretty(os.Stdout, doc)
		}
		res, err := doc.One()
		if err != nil {
			return err
		}
		fmt.Printf("ID:        %s\n", res.ID)
		fmt.Printf("Name:      %s\n", res.Str("name"))
		fmt.Printf("State:     %s\n", res.Str("state"))
		fmt.Printf("Started:   %s\n", res.Str("started_at"))
		fmt.Printf("Completed: %s\n", res.Str("completed_at"))
		return nil
	},
}

var (
	scanCreateProvider string
	scanCreateName     string
	scanCreateWait     bool
	scanCreateQuiet    bool
)

var scansLaunchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a new scan against a provider",
	Long: `Launches an asynchronous scan. Scanning happens in the background on the
Prowler API side; use --wait to block until it reaches a terminal state
(completed/failed/cancelled), which is convenient in CI/CD pipelines and
cron jobs that need a final exit code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if scanCreateProvider == "" {
			return fmt.Errorf("--provider is required")
		}

		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()

		scanID, err := c.CreateScanAndGetID(
			ctx,
			scanCreateProvider,
			scanCreateName,
		)
		if err != nil {
			return err
		}

		if scanCreateWait {
			return waitForScan(c, scanID)
		}

		if scanCreateQuiet {
			fmt.Println(scanID)
			return nil
		}

		if flagOutput == string(output.JSON) {
			return output.JSONPretty(os.Stdout, map[string]any{
				"id": scanID,
			})
		}

		fmt.Printf("Scan launched: id=%s\n", scanID)
		return nil
	},
}

// waitForScan polls GetScan until a terminal state is reached, returning a
// non-nil error (and therefore non-zero exit code) on failure/cancellation,
// which is what cron/CI callers rely on.
func waitForScan(c *client.Client, scanID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), flagTimeout)
		doc, err := c.GetScan(ctx, scanID)
		cancel()
		if err != nil {
			return err
		}
		res, err := doc.One()
		if err != nil {
			return err
		}
		state := res.Str("state")
		switch state {
		case "completed":
			fmt.Printf("Scan %s completed.\n", scanID)
			return nil
		case "failed", "cancelled":
			return fmt.Errorf("scan %s ended with state %q", scanID, state)
		default:
			fmt.Printf("Scan %s state=%s, waiting...\n", scanID, state)
		}
		<-ticker.C
	}
}

func init() {
	scansListCmd.Flags().StringVar(&scanFilterProvider, "provider", "", "Filter by provider ID")
	scansListCmd.Flags().StringVar(&scanFilterState, "state", "", "Filter by scan state")
	scansListCmd.Flags().IntVar(&scanPage, "page", 0, "Page number")
	scansListCmd.Flags().IntVar(&scanPageSize, "size", 0, "Page size")
	scansListCmd.Flags().BoolVar(&scanAll, "all", false, "Fetch every page of results")

	scansLaunchCmd.Flags().StringVar(&scanCreateProvider, "provider", "", "Provider ID to scan (required)")
	scansLaunchCmd.Flags().StringVar(&scanCreateName, "name", "", "Optional scan name")
	scansLaunchCmd.Flags().BoolVar(&scanCreateWait, "wait", false, "Block until the scan reaches a terminal state")
	scansLaunchCmd.Flags().BoolVarP(&scanCreateQuiet, "quiet", "q", false, "Print only the new scan ID (for scripting)")

	scansCmd.AddCommand(scansListCmd, scansGetCmd, scansLaunchCmd)
}
