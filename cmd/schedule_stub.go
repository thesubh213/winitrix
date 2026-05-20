//go:build !windows

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Create or print a Task Scheduler entry",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scheduling is only supported on Windows.")
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}
