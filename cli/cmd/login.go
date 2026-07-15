package cmd

import (
	"fmt"

	pikpak "github.com/suntao12138/pikpakdriver/pkg/pikpak"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with PikPak (email/password)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailFlag == "" || passwordFlag == "" {
			return fmt.Errorf("--email and --password are required")
		}

		cliClient := pikpak.NewLoginClient(emailFlag, proxyFlag)
		_, err := cliClient.Login(emailFlag, passwordFlag)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if err := pikpak.SaveCredentials(emailFlag, passwordFlag, proxyFlag); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Println("Login successful! Credentials saved.")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&emailFlag, "email", "", "PikPak account email")
	loginCmd.Flags().StringVar(&passwordFlag, "password", "", "PikPak account password")
	rootCmd.AddCommand(loginCmd)
}
