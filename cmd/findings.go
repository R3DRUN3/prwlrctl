package cmd

import (
	"os"
	"time"

	"github.com/r3drun3/prwlrctl/internal/client"
	"github.com/r3drun3/prwlrctl/internal/output"
	"github.com/spf13/cobra"
)

var findingsCmd = &cobra.Command{
	Use:   "findings",
	Short: "List findings produced by scans",
}

var (
	findingScan     string
	findingProvider string
	findingSeverity string
	findingStatus   string
	findingSince    time.Duration
	findingPage     int
	findingPageSize int
	findingAll      bool
)

var findingColumns = []output.Column{
	output.IDColumn(),
	output.Attr("CHECK", "check_id"),
	output.Attr("SEVERITY", "severity"),
	output.Attr("STATUS", "status"),
	output.Attr("REGION", "region"),
	output.Attr("RESOURCE", "resource_uid"),
}

var findingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List findings, optionally filtered by scan/severity/status",
	Long: `Prowler requires at least one date filter on findings (to avoid
unbounded queries). This command always sends filter[updated\_at.gte],
computed as now - --since (default 24h). Widen --since for older results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		filters := map[string]string{
			"scan":           findingScan,
			"provider":       findingProvider,
			"severity":       findingSeverity,
			"status":         findingStatus,
			"updated_at.gte": time.Now().Add(-findingSince).UTC().Format(time.RFC3339),
		}

		if findingAll {
			resources, err := c.ListAll(ctx, "/findings", client.BuildQuery(filters))
			if err != nil {
				return err
			}
			if flagOutput == string(output.JSON) {
				return output.JSONPretty(os.Stdout, resources)
			}
			output.RenderTable(os.Stdout, resources, findingColumns)
			return nil
		}

		doc, err := c.ListFindings(ctx, filters, findingPage, findingPageSize)
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
		output.RenderTable(os.Stdout, resources, findingColumns)
		return nil
	},
}

func init() {
	findingsListCmd.Flags().StringVar(&findingScan, "scan", "", "Filter by scan ID")
	findingsListCmd.Flags().StringVar(&findingProvider, "provider", "", "Filter by provider ID")
	findingsListCmd.Flags().StringVar(&findingSeverity, "severity", "", "Filter by severity (critical|high|medium|low|informational)")
	findingsListCmd.Flags().StringVar(&findingStatus, "status", "", "Filter by status (PASS|FAIL|MANUAL)")
	findingsListCmd.Flags().DurationVar(&findingSince, "since", 24*time.Hour, "Only findings updated within this duration (Prowler requires a date filter)")
	findingsListCmd.Flags().IntVar(&findingPage, "page", 0, "Page number")
	findingsListCmd.Flags().IntVar(&findingPageSize, "size", 0, "Page size")
	findingsListCmd.Flags().BoolVar(&findingAll, "all", false, "Fetch every page of results")
	findingsCmd.AddCommand(findingsListCmd)
}
