package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmPermanent bool

var rmCmd = &cobra.Command{
	Use:   "rm <file_id...>",
	Short: "Move files to trash (or permanently delete with --permanent)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if rmPermanent {
			if err := client.DeleteFiles(args); err != nil {
				return err
			}
			fmt.Println("Permanently deleted.")
		} else {
			if err := client.TrashFiles(args); err != nil {
				return err
			}
			fmt.Println("Moved to trash.")
		}
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmPermanent, "permanent", "P", false, "Permanently delete instead of trash")
	rootCmd.AddCommand(rmCmd)
}
