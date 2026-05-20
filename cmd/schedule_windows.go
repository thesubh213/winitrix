//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	scheduleTime     string
	scheduleDays     string
	scheduleInterval string
	scheduleApply    bool
	scheduleRemove   bool
	scheduleTaskName string
	scheduleSilent   bool
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Create or print a Task Scheduler entry",
	Run: func(cmd *cobra.Command, args []string) {
		settings := ActiveSettings()
		if scheduleTaskName == "" {
			scheduleTaskName = "Winitrix Update"
		}

		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to resolve executable path: %v\n", err)
			return
		}
		exePath, _ = filepath.Abs(exePath)

		argsList := []string{}
		if scheduleSilent {
			argsList = append(argsList, "--silent")
		}
		if settings.Profile != "" {
			argsList = append(argsList, "--profile", settings.Profile)
		}
		if configPath != "" {
			argsList = append(argsList, "--config", configPath)
		}

		scheduledCommand := buildCommandLine(exePath, argsList)
		createArgs, err := buildScheduleArgs(scheduleTaskName, scheduledCommand)
		if err != nil {
			fmt.Printf("Invalid schedule settings: %v\n", err)
			return
		}

		if scheduleRemove {
			deleteArgs := []string{"/Delete", "/TN", scheduleTaskName, "/F"}
			if scheduleApply {
				if err := runSchtasks(deleteArgs); err != nil {
					fmt.Printf("Failed to remove task: %v\n", err)
					return
				}
				fmt.Println("Task removed.")
			} else {
				fmt.Printf("schtasks %s\n", strings.Join(deleteArgs, " "))
			}
			return
		}

		if scheduleApply {
			if err := runSchtasks(createArgs); err != nil {
				fmt.Printf("Failed to create task: %v\n", err)
				return
			}
			fmt.Println("Task created.")
			return
		}

		fmt.Printf("schtasks %s\n", strings.Join(createArgs, " "))
	},
}

func buildScheduleArgs(taskName string, command string) ([]string, error) {
	if scheduleTime == "" {
		scheduleTime = "09:00"
	}
	interval := strings.ToLower(strings.TrimSpace(scheduleInterval))
	if interval == "" {
		interval = "weekly"
	}
	if interval != "weekly" && interval != "daily" {
		return nil, fmt.Errorf("interval must be 'daily' or 'weekly'")
	}

	args := []string{"/Create", "/F", "/TN", taskName, "/TR", command, "/ST", scheduleTime}

	if interval == "daily" {
		args = append(args, "/SC", "DAILY")
		return args, nil
	}

	if scheduleDays == "" {
		scheduleDays = "Mon"
	}
	dayTokens := parseScheduleDays(scheduleDays)
	if len(dayTokens) == 0 {
		return nil, fmt.Errorf("invalid --days value")
	}

	args = append(args, "/SC", "WEEKLY", "/D", strings.Join(dayTokens, ","))
	return args, nil
}

func parseScheduleDays(input string) []string {
	parts := strings.Split(input, ",")
	var out []string
	for _, part := range parts {
		d := strings.TrimSpace(strings.ToLower(part))
		switch d {
		case "mon", "monday":
			out = append(out, "MON")
		case "tue", "tues", "tuesday":
			out = append(out, "TUE")
		case "wed", "wednesday":
			out = append(out, "WED")
		case "thu", "thur", "thurs", "thursday":
			out = append(out, "THU")
		case "fri", "friday":
			out = append(out, "FRI")
		case "sat", "saturday":
			out = append(out, "SAT")
		case "sun", "sunday":
			out = append(out, "SUN")
		}
	}
	return out
}

func buildCommandLine(exePath string, args []string) string {
	parts := []string{quoteArg(exePath)}
	for _, arg := range args {
		parts = append(parts, quoteArg(arg))
	}
	return strings.Join(parts, " ")
}

func quoteArg(arg string) string {
	if strings.ContainsAny(arg, " \t") {
		escaped := strings.ReplaceAll(arg, "\"", "\\\"")
		return "\"" + escaped + "\""
	}
	return arg
}

func runSchtasks(args []string) error {
	cmd := exec.Command("schtasks", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	scheduleCmd.Flags().StringVar(&scheduleTime, "time", "09:00", "Start time in 24h format (HH:MM)")
	scheduleCmd.Flags().StringVar(&scheduleDays, "days", "Mon", "Days for weekly schedule (e.g., Mon,Wed,Fri)")
	scheduleCmd.Flags().StringVar(&scheduleInterval, "interval", "weekly", "Schedule interval: daily or weekly")
	scheduleCmd.Flags().BoolVar(&scheduleApply, "apply", false, "Apply changes via schtasks")
	scheduleCmd.Flags().BoolVar(&scheduleRemove, "remove", false, "Remove the scheduled task")
	scheduleCmd.Flags().StringVar(&scheduleTaskName, "task-name", "Winitrix Update", "Task Scheduler task name")
	scheduleCmd.Flags().BoolVar(&scheduleSilent, "silent", true, "Run scheduled updates in silent mode")
	rootCmd.AddCommand(scheduleCmd)
}
