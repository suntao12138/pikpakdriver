package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var starCmd = &cobra.Command{
	Use:   "star <file_id...>",
	Short: "Star (bookmark) files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.StarFiles(args); err != nil {
			return err
		}
		fmt.Println("Starred successfully.")
		return nil
	},
}

var unstarCmd = &cobra.Command{
	Use:   "unstar <file_id...>",
	Short: "Unstar (remove bookmark from) files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.UnstarFiles(args); err != nil {
			return err
		}
		fmt.Println("Unstarred successfully.")
		return nil
	},
}

var starredCmd = &cobra.Command{
	Use:   "starred [limit]",
	Short: "List starred files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := 50
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &limit)
		}
		list, err := client.ListStarred(limit)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(list)
			return nil
		}
		for _, f := range list.Files {
			fmt.Printf("  📄 %s\n", f.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(starCmd)
	rootCmd.AddCommand(unstarCmd)
	rootCmd.AddCommand(starredCmd)
}
