package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var offlineCmd = &cobra.Command{
	Use:   "offline",
	Short: "Manage offline downloads",
}

var offlineAddCmd = &cobra.Command{
	Use:   "add <url> [name]",
	Short: "Add an offline download task (magnet or URL)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		name := ""
		parentID := ""
		if len(args) > 1 {
			name = args[1]
		}
		resp, err := client.AddOfflineTask(url, parentID, name)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
			return nil
		}
		t := resp.Task
		fmt.Printf("Offline task created:\n")
		fmt.Printf("  File: %s\n", t.FileName)
		fmt.Printf("  ID:   %s\n", t.ID)
		fmt.Printf("  Phase: %s\n", t.Phase)
		return nil
	},
}

var offlineListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List offline download tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := client.ListOfflineTasks(50, nil)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(list)
			return nil
		}
		for _, t := range list.Tasks {
			fmt.Printf("  %s [%s] %d%%\n", t.FileName, t.Phase, t.Progress)
		}
		return nil
	},
}

var offlineInfoCmd = &cobra.Command{
	Use:   "info <task_id>",
	Short: "Get offline task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task, err := client.GetOfflineTask(args[0])
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(task)
			return nil
		}
		fmt.Printf("Name:     %s\n", task.FileName)
		fmt.Printf("ID:       %s\n", task.ID)
		fmt.Printf("Phase:    %s\n", task.Phase)
		fmt.Printf("Progress: %d%%\n", task.Progress)
		fmt.Printf("Size:     %s\n", task.FileSize)
		fmt.Printf("Message:  %s\n", task.Message)
		return nil
	},
}

var offlineRmCmd = &cobra.Command{
	Use:   "rm <task_id>",
	Short: "Delete an offline download task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteOfflineTask(args[0], false); err != nil {
			return err
		}
		fmt.Println("Task deleted.")
		return nil
	},
}

var offlineRetryCmd = &cobra.Command{
	Use:   "retry <task_id>",
	Short: "Retry a failed offline download task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.RetryOfflineTask(args[0]); err != nil {
			return err
		}
		fmt.Println("Task retry initiated.")
		return nil
	},
}

func init() {
	offlineCmd.AddCommand(offlineAddCmd)
	offlineCmd.AddCommand(offlineListCmd)
	offlineCmd.AddCommand(offlineInfoCmd)
	offlineRmCmd.Flags().BoolP("delete-files", "D", false, "Also delete downloaded files")
	offlineCmd.AddCommand(offlineRmCmd)
	offlineCmd.AddCommand(offlineRetryCmd)
	rootCmd.AddCommand(offlineCmd)
}
