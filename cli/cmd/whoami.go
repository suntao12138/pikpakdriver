package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authenticated user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := client.GetUserInfo()
		if err != nil {
			return err
		}
		quota, _ := client.GetQuota()
		vip, _ := client.GetVipInfo()

		if jsonOutput {
			printJSON(map[string]interface{}{
				"user":  user,
				"quota": quota,
				"vip":   vip,
			})
			return nil
		}
		fmt.Printf("Name:  %s\n", user.Name)
		fmt.Printf("Email: %s\n", user.Email)
		fmt.Printf("Phone: %s\n", user.PhoneNumber)
		if quota != nil {
			usage, _ := strconv.ParseInt(quota.Quota.Usage, 10, 64)
			lim, _ := strconv.ParseInt(quota.Quota.Limit, 10, 64)
			fmt.Printf("Storage: %d GB / %d GB\n", usage/(1024*1024*1024), lim/(1024*1024*1024))
		}
		if vip != nil && vip.Data != nil {
			fmt.Printf("VIP: %s (status: %s, expire: %s)\n", vip.Data.VipType, vip.Data.Status, vip.Data.Expire)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
