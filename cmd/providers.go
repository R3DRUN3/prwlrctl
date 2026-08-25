package cmd

import (
	"fmt"
	"os"

	"github.com/r3drun3/prwlrctl/internal/output"
	prowler "github.com/r3drun3/prwlrctl/pkg/prowler"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Manage and inspect cloud provider connections",
}

var (
	providerFilterType string
	providerPage       int
	providerPageSize   int
	providerAll        bool
)

var providerColumns = []output.Column{
	output.IDColumn(),
	output.Attr("PROVIDER", "provider"),
	output.Attr("ALIAS", "alias"),
	output.Attr("UID", "uid"),
	output.NestedBoolStatus("STATUS", "connection", "connected", "connected", "disconnected"),
}

var providersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List provider accounts (AWS/Azure/GCP/Kubernetes/...)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		filters := map[string]string{"provider": providerFilterType}

		if providerAll {
			resources, err := c.ListAll(ctx, "/providers", prowler.BuildQuery(filters))
			if err != nil {
				return err
			}
			if flagOutput == string(output.JSON) {
				return output.JSONPretty(os.Stdout, resources)
			}
			output.RenderTable(os.Stdout, resources, providerColumns)
			return nil
		}

		doc, err := c.ListProviders(ctx, filters, providerPage, providerPageSize)
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
		output.RenderTable(os.Stdout, resources, providerColumns)
		return nil
	},
}

var providersGetCmd = &cobra.Command{
	Use:   "get <provider-id>",
	Short: "Show a single provider by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		doc, err := c.GetProvider(ctx, args[0])
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
		status := "-"
		if v, ok := res.NestedBool("connection", "connected"); ok {
			if v {
				status = "connected"
			} else {
				status = "disconnected"
			}
		}
		fmt.Printf("ID:       %s\n", res.ID)
		fmt.Printf("Provider: %s\n", res.Str("provider"))
		fmt.Printf("Alias:    %s\n", res.Str("alias"))
		fmt.Printf("UID:      %s\n", res.Str("uid"))
		fmt.Printf("Status:   %s\n", status)
		return nil
	},
}

func init() {
	providersListCmd.Flags().StringVar(&providerFilterType, "type", "", "Filter by provider type (aws|azure|gcp|kubernetes|...)")
	providersListCmd.Flags().IntVar(&providerPage, "page", 0, "Page number")
	providersListCmd.Flags().IntVar(&providerPageSize, "size", 0, "Page size")
	providersListCmd.Flags().BoolVar(&providerAll, "all", false, "Fetch every page of results")
	providersCmd.AddCommand(providersListCmd, providersGetCmd)
}
