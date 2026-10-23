//go:build !windows

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Create or print a Task Scheduler entry",
	Run: func(cmd *cobra.Command, args []string) {
		tui.Println("Scheduling is only supported on Windows.")
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}
