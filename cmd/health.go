package cmd

import (
	"fmt"
	"os"

	"github.com/r3drun3/prwlrctl/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check whether the Prowler Server is healthy",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := cmdContext()
		defer cancel()

		c := newClient()
		health, err := c.Health(ctx)

		if flagOutput == string(output.JSON) {
			if health != nil {
				return output.JSONPretty(os.Stdout, health)
			}

			return output.JSONPretty(os.Stdout, map[string]any{
				"status": "fail",
				"error":  err.Error(),
			})
		}

		if err != nil {
			return err
		}

		fmt.Printf("Status:      %s\n", health.Status)
		fmt.Printf("Version:     %s\n", health.Version)
		fmt.Printf("Release ID:  %s\n", health.ReleaseID)
		fmt.Printf("Service ID:  %s\n", health.ServiceID)
		fmt.Printf("Description: %s\n", health.Description)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
