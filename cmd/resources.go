package cmd

import (
	"fmt"
	"os"

	"github.com/r3drun3/prwlrctl/internal/jsonapi"
	"github.com/r3drun3/prwlrctl/internal/output"
	"github.com/spf13/cobra"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Inspect resources",
}

var resourcesGetCmd = &cobra.Command{
	Use:   "get <resource-id>",
	Short: "Show a single resource",
	Long: `Retrieve a single resource by ID.

The JSON output contains the complete JSON:API response, including:
  - resource attributes
  - provider relationship
  - findings relationship`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()

		doc, err := c.GetResource(ctx, args[0])
		if err != nil {
			return err
		}

		if flagOutput == string(output.JSON) {
			return output.JSONPretty(os.Stdout, doc)
		}

		return renderResourceDetails(doc)
	},
}

func renderResourceDetails(doc jsonapi.Document) error {
	resource, err := doc.One()
	if err != nil {
		return fmt.Errorf("decoding resource: %w", err)
	}

	fmt.Printf("ID:              %s\n", resource.ID)
	fmt.Printf("Type:            %s\n", resource.Type)
	fmt.Printf("Name:            %s\n", resource.Str("name"))
	fmt.Printf("UID:             %s\n", resource.Str("uid"))
	fmt.Printf("Region:          %s\n", resource.Str("region"))
	fmt.Printf("Service:         %s\n", resource.Str("service"))
	fmt.Printf("Resource Type:   %s\n", resource.Str("type"))
	fmt.Printf("Partition:       %s\n", resource.Str("partition"))
	fmt.Printf("Inserted:        %s\n", resource.Str("inserted_at"))
	fmt.Printf("Updated:         %s\n", resource.Str("updated_at"))

	if details := resource.Str("details"); details != "" {
		fmt.Printf("Details:         %s\n", details)
	}

	if tags := resource.Any("tags"); tags != nil {
		fmt.Printf("Tags:            %v\n", tags)
	}

	if groups := resource.Any("groups"); groups != nil {
		fmt.Printf("Groups:          %v\n", groups)
	}

	if failed := resource.Any("failed_findings_count"); failed != nil {
		fmt.Printf("Failed Findings: %v\n", failed)
	}

	fmt.Println()

	if providerIDs := resource.RelatedIDs("provider"); len(providerIDs) > 0 {
		fmt.Println("Provider:")
		fmt.Printf("  ID:   %s\n", providerIDs[0].ID)
		fmt.Printf("  Type: %s\n", providerIDs[0].Type)
		fmt.Println()
	}

	if findingIDs := resource.RelatedIDs("findings"); len(findingIDs) > 0 {
		fmt.Printf("Findings: %d\n", len(findingIDs))
		for _, finding := range findingIDs {
			fmt.Printf("  %s\n", finding.ID)
		}
	} else {
		fmt.Println("Findings: 0")
	}

	return nil
}

func init() {
	resourcesCmd.AddCommand(resourcesGetCmd)
}
