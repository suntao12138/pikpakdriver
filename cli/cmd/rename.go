package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <file_id> <new_name>",
	Short: "Rename a file or folder",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Rename(args[0], args[1]); err != nil {
			return err
		}
		fmt.Println("Renamed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
