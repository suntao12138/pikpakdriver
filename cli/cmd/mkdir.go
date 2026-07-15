package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <parent_id> <name>",
	Short: "Create a new folder",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := client.Mkdir(args[0], args[1])
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(f)
			return nil
		}
		fmt.Printf("Folder created: %s (ID: %s)\n", f.Name, f.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkdirCmd)
}
