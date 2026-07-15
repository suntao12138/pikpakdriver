package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <file_id> <to_parent_id>",
	Short: "Move files to another directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.MoveFiles([]string{args[0]}, args[1]); err != nil {
			return err
		}
		fmt.Println("Moved successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
