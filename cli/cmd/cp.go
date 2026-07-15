package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:   "cp <file_id> <to_parent_id>",
	Short: "Copy files to another directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.CopyFiles([]string{args[0]}, args[1]); err != nil {
			return err
		}
		fmt.Println("Copied successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
