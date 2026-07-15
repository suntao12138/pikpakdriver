package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <file_id>",
	Short: "Get a direct download URL for a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		link, err := client.GetDownloadLink(args[0])
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(link)
			return nil
		}
		fmt.Println("Download URL:", link.WebContentLink)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
