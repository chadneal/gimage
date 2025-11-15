package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployments",
	Long:  `List all tracked Lambda deployments with their status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dm := storage.NewDeploymentManager()
		if err := dm.Load(); err != nil {
			return fmt.Errorf("failed to load deployments: %w", err)
		}

		deployments := dm.List()

		if len(deployments) == 0 {
			fmt.Println("No deployments found.")
			fmt.Println("\nCreate your first deployment with:")
			fmt.Println("  gimage-deploy deploy --id <deployment-id> --stage <stage>")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tREGION\tSTAGE\tSTATUS\tENDPOINT")
		fmt.Fprintln(w, "──\t──────\t─────\t──────\t────────")

		for _, d := range deployments {
			endpoint := d.APIGatewayURL
			if len(endpoint) > 50 {
				endpoint = endpoint[:47] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				d.ID,
				d.Region,
				d.Stage,
				d.Status,
				endpoint,
			)
		}

		w.Flush()
		fmt.Printf("\nTotal: %d deployment(s)\n", len(deployments))
		return nil
	},
}
