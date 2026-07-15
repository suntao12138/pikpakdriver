package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "Manage trash",
}

var trashListCmd = &cobra.Command{
	Use:   "ls [limit]",
	Short: "List trash contents",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := 50
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &limit)
		}
		list, err := client.ListTrash(limit)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(list)
			return nil
		}
		for _, f := range list.Files {
			fmt.Printf("  🗑️ %s (ID: %s)\n", f.Name, f.ID)
		}
		return nil
	},
}

var trashRestoreCmd = &cobra.Command{
	Use:   "restore <file_id...>",
	Short: "Restore files from trash",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.UntrashFiles(args); err != nil {
			return err
		}
		fmt.Println("Restored from trash.")
		return nil
	},
}

var trashEmptyCmd = &cobra.Command{
	Use:   "empty",
	Short: "Permanently empty the entire trash",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.EmptyTrash(); err != nil {
			return err
		}
		fmt.Println("Trash emptied.")
		return nil
	},
}

func init() {
	trashCmd.AddCommand(trashListCmd)
	trashCmd.AddCommand(trashRestoreCmd)
	trashCmd.AddCommand(trashEmptyCmd)
	rootCmd.AddCommand(trashCmd)
}
