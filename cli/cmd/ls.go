package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var lsLong bool

var lsCmd = &cobra.Command{
	Use:   "ls [parent_id]",
	Short: "List directory contents",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parentID := ""
		if len(args) > 0 {
			parentID = args[0]
		}
		list, err := client.ListFiles(parentID, 50)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(list)
			return nil
		}
		for _, f := range list.Files {
			kind := "📄"
			if f.Kind == "drive#folder" {
				kind = "📁"
			}
			size := ""
			if f.Size != "" && f.Kind != "drive#folder" {
				s, _ := strconv.ParseInt(f.Size, 10, 64)
				size = " " + formatSize(s)
			}
			if lsLong {
				fmt.Printf("%s %-30s %s (ID: %s)\n", kind, f.Name, size, f.ID)
			} else {
				fmt.Printf("%s %s\n", kind, f.Name)
			}
		}
		return nil
	},
}

func init() {
	lsCmd.Flags().BoolVarP(&lsLong, "long", "l", false, "Show detailed listing")
	rootCmd.AddCommand(lsCmd)
}
