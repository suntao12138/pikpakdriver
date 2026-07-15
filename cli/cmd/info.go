package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <file_id>",
	Short: "Get detailed file or folder info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := client.GetFileInfo(args[0])
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(info)
			return nil
		}
		fmt.Printf("Name: %s\n", info.File.Name)
		fmt.Printf("ID:   %s\n", info.File.ID)
		fmt.Printf("Kind: %s\n", info.File.Kind)
		if info.File.Size != "" {
			s, _ := strconv.ParseInt(info.File.Size, 10, 64)
			fmt.Printf("Size: %s\n", formatSize(s))
		}
		fmt.Printf("Created: %s\n", info.File.CreatedTime)
		fmt.Printf("Modified: %s\n", info.File.ModifiedTime)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
