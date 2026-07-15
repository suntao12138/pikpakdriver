package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Manage share links",
}

var shareCreateCmd = &cobra.Command{
	Use:   "create <file_id...>",
	Short: "Create share links for files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expireDays, _ := cmd.Flags().GetInt("expire")
		passCode, _ := cmd.Flags().GetString("pass-code")
		resp, err := client.CreateShare(args, expireDays, passCode)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
			return nil
		}
		fmt.Printf("Share created:\n")
		fmt.Printf("  URL:      %s\n", resp.ShareURL)
		fmt.Printf("  ShareID:  %s\n", resp.ShareID)
		if resp.PassCode != "" {
			fmt.Printf("  PassCode: %s\n", resp.PassCode)
		}
		return nil
	},
}

var shareListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List your share links",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := client.ListShares()
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(list)
			return nil
		}
		for _, s := range list.Data {
			fmt.Printf("  🔗 %s (status: %s)\n", s.ShareURL, s.ShareStatus)
		}
		return nil
	},
}

var shareRmCmd = &cobra.Command{
	Use:   "rm <share_id...>",
	Short: "Delete share links",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteShares(args); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(map[string]bool{"deleted": true})
			return nil
		}
		fmt.Println("Share(s) deleted.")
		return nil
	},
}

var shareInfoCmd = &cobra.Command{
	Use:   "info <share_id> [pass_code]",
	Short: "Get share information",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		passCode := ""
		if len(args) > 1 {
			passCode = args[1]
		}
		info, err := client.GetShareInfo(args[0], passCode)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(info)
			return nil
		}
		fmt.Printf("Share status: %s\n", info.ShareStatus)
		for _, f := range info.Files {
			fmt.Printf("  %s\n", f.Name)
		}
		return nil
	},
}

var shareSaveCmd = &cobra.Command{
	Use:   "save <share_id> <to_parent_id>",
	Short: "Save a shared file to your account",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		passCodeToken, _ := cmd.Flags().GetString("pass-code-token")
		if err := client.SaveShare(args[0], passCodeToken, nil, args[1]); err != nil {
			return err
		}
		fmt.Println("Share saved successfully.")
		return nil
	},
}

func init() {
	shareCreateCmd.Flags().IntP("expire", "e", 0, "Expiration in days (0 = never)")
	shareCreateCmd.Flags().StringP("pass-code", "p", "", "Password protection")
	shareSaveCmd.Flags().String("pass-code-token", "", "Pass code token from share info (for encrypted shares)")
	shareCmd.AddCommand(shareCreateCmd)
	shareCmd.AddCommand(shareListCmd)
	shareCmd.AddCommand(shareRmCmd)
	shareCmd.AddCommand(shareInfoCmd)
	shareCmd.AddCommand(shareSaveCmd)
	rootCmd.AddCommand(shareCmd)
}
