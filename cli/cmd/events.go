package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events [limit]",
	Short: "List recent file events",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := 10
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &limit)
		}
		events, err := client.ListEvents(limit)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(events)
			return nil
		}
		for _, e := range events.Events {
			fmt.Printf("  🔔 %s — %s (%s)\n", e.TypeName, e.FileName, e.CreatedTime)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(eventsCmd)
}
