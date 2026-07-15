package cmd

import (
	"fmt"
	"os"

	pikpak "github.com/suntao12138/pikpakdriver/pkg/pikpak"
	"github.com/spf13/cobra"
)

var (
	proxyFlag   string
	jsonOutput  bool
	emailFlag   string
	passwordFlag string
)

var client *pikpak.Client

var rootCmd = &cobra.Command{
	Use:           "pikpakdriver",
	Short:         "CLI tool for PikPak cloud storage",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		name := cmd.Name()
		switch name {
		case "login", "help", "completion", "version", "__complete", "__completeNoDesc":
			return nil
		}
		if parent := cmd.Parent(); parent != nil {
			pname := parent.Name()
			if pname == "completion" {
				return nil
			}
		}

		var err error
		client, err = pikpak.NewClient(proxyFlag)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		return nil
	},
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func init() {
	rootCmd.PersistentFlags().StringVar(&proxyFlag, "proxy", "", "HTTP proxy URL (e.g. http://127.0.0.1:7890)")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
}
