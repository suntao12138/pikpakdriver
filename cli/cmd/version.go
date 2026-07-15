package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set via -ldflags at build time
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("pikpakdriver version %s\n", Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
