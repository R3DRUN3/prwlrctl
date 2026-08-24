package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/r3drun3/prwlrctl/internal/client"
	"github.com/r3drun3/prwlrctl/internal/jsonapi"
	"github.com/r3drun3/prwlrctl/internal/output"
	"github.com/spf13/cobra"
)

var findingsCmd = &cobra.Command{
	Use:   "findings",
	Short: "List and inspect findings produced by scans",
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
}

var findingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List findings, optionally filtered by scan/severity/status",
	Long: `Prowler requires at least one date filter on findings (to avoid
unbounded queries). This command always sends filter[updated_at.gte],
computed as now - --since (default 7 days). Widen --since for older results.`,
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

var findingsGetCmd = &cobra.Command{
	Use:   "get <finding-id>",
	Short: "Show a single finding, including affected resources",
	Long: `Retrieve a single finding by ID, shows all related informazions, including its related resource.

The JSON output contains the complete JSON:API response, including:
  - finding attributes
  - scan relationship
  - resource relationship
  - included resource details
  - finding check metadata
  - remediation information`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()

		doc, err := c.GetFinding(ctx, args[0])
		if err != nil {
			return err
		}

		if flagOutput == string(output.JSON) {
			return output.JSONPretty(os.Stdout, doc)
		}

		return renderFindingDetails(doc)
	},
}

// renderFindingDetails renders the most useful finding information in a
// human-readable format. JSON output remains available when callers need
// the complete API response.
func renderFindingDetails(doc jsonapi.Document) error {
	finding, err := doc.One()
	if err != nil {
		return fmt.Errorf("decoding finding: %w", err)
	}

	fmt.Printf("ID:             %s\n", finding.ID)
	fmt.Printf("UID:             %s\n", finding.Str("uid"))
	fmt.Printf("Check:           %s\n", finding.Str("check_id"))
	fmt.Printf("Severity:        %s\n", finding.Str("severity"))
	fmt.Printf("Status:          %s\n", finding.Str("status"))
	fmt.Printf("Status Extended: %s\n", finding.Str("status_extended"))
	fmt.Printf("Inserted:        %s\n", finding.Str("inserted_at"))
	fmt.Printf("Updated:         %s\n", finding.Str("updated_at"))
	fmt.Printf("First Seen:      %s\n", finding.Str("first_seen_at"))

	if v := finding.Any("muted"); v != nil {
		fmt.Printf("Muted:           %v\n", v)
	}

	if reason := finding.Any("muted_reason"); reason != nil {
		fmt.Printf("Muted Reason:    %v\n", reason)
	}

	fmt.Println()

	// Scan relationship.
	if scanIDs := finding.RelatedIDs("scan"); len(scanIDs) > 0 {
		fmt.Println("Scan:")
		fmt.Printf("  ID: %s\n", scanIDs[0].ID)
		fmt.Printf("  Type: %s\n", scanIDs[0].Type)
		fmt.Println()
	}

	// Check metadata.
	if metadata, ok := finding.Any("check_metadata").(map[string]any); ok {
		fmt.Println("Check Metadata:")

		printAnyField(metadata, "checktitle", "Title")
		printAnyField(metadata, "provider", "Provider")
		printAnyField(metadata, "servicename", "Service")
		printAnyField(metadata, "subservicename", "Subservice")
		printAnyField(metadata, "resourcetype", "Resource Type")
		printAnyField(metadata, "resourcegroup", "Resource Group")
		printAnyField(metadata, "description", "Description")
		printAnyField(metadata, "risk", "Risk")
		printAnyField(metadata, "notes", "Notes")
		printAnyField(metadata, "relatedurl", "Related URL")

		fmt.Println()
	}

	// Related resources.
	if resources, err := doc.IncludedResources(); err != nil {
		return fmt.Errorf("decoding included resources: %w", err)
	} else if len(resources) > 0 {
		fmt.Println("Related Resources:")

		for _, resource := range resources {
			fmt.Printf("  ID:       %s\n", resource.ID)
			fmt.Printf("  Type:     %s\n", resource.Type)
			fmt.Printf("  Name:     %s\n", resource.Str("name"))
			fmt.Printf("  UID:      %s\n", resource.Str("uid"))
			fmt.Printf("  Region:   %s\n", resource.Str("region"))
			fmt.Printf("  Service:  %s\n", resource.Str("service"))
			fmt.Printf("  Partition:%s\n", resource.Str("partition"))

			if details := resource.Str("details"); details != "" {
				fmt.Printf("  Details:  %s\n", details)
			}

			fmt.Println()
		}
	}

	// Remediation.
	if metadata, ok := finding.Any("check_metadata").(map[string]any); ok {
		if remediation, ok := metadata["remediation"].(map[string]any); ok {
			fmt.Println("Remediation:")

			if recommendation, ok := remediation["recommendation"].(map[string]any); ok {
				printAnyField(recommendation, "text", "Recommendation")
				printAnyField(recommendation, "url", "URL")
			}

			if code, ok := remediation["code"].(map[string]any); ok {
				fmt.Println()
				fmt.Println("Remediation Code:")

				printAnyField(code, "cli", "CLI")
				printAnyField(code, "other", "Other")
				printAnyField(code, "nativeiac", "Native IaC")
				printAnyField(code, "terraform", "Terraform")
			}
		}
	}

	return nil
}

func printAnyField(values map[string]any, key, label string) {
	value, ok := values[key]
	if !ok || value == nil {
		return
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return
		}
		fmt.Printf("%s: %s\n", label, v)
	default:
		fmt.Printf("%s: %v\n", label, v)
	}
}

func init() {
	findingsListCmd.Flags().StringVar(&findingScan, "scan", "", "Filter by scan ID")
	findingsListCmd.Flags().StringVar(&findingProvider, "provider", "", "Filter by provider ID")
	findingsListCmd.Flags().StringVar(&findingSeverity, "severity", "", "Filter by severity (critical|high|medium|low|informational)")
	findingsListCmd.Flags().StringVar(&findingStatus, "status", "", "Filter by status (PASS|FAIL|MANUAL)")
	findingsListCmd.Flags().DurationVar(&findingSince, "since", 7*24*time.Hour, "Only findings updated within this duration (Prowler requires a date filter)")
	findingsListCmd.Flags().IntVar(&findingPage, "page", 0, "Page number")
	findingsListCmd.Flags().IntVar(&findingPageSize, "size", 0, "Page size")
	findingsListCmd.Flags().BoolVar(&findingAll, "all", false, "Fetch every page of results")

	findingsCmd.AddCommand(findingsListCmd, findingsGetCmd)
}
