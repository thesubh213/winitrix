package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/config"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/notify"
	"github.com/thesubh213/winitrix/pkg/report"
	"github.com/thesubh213/winitrix/pkg/retry"
	"github.com/thesubh213/winitrix/pkg/state"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var (
	silent          bool
	verbose         bool
	jsonOut         bool
	dryRun          bool
	nonstop         bool
	timeout         int
	retries         int
	retryDelay      int
	onlyManagers    []string
	excludeManagers []string
)

var rootCmd = &cobra.Command{
	Use:               "winitrix",
	Short:             "Unified Windows package orchestration engine",
	Long:              `Winitrix is a modern, offline-first terminal application that unifies and orchestrates updates across multiple Windows package ecosystems seamlessly.`,
	PersistentPreRunE: loadSettings,
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			printVersionInfo(os.Stdout)
			return
		}

		settings := ActiveSettings()

		logger.Init()
		defer logger.Close()

		if err := tryAutoInstall(); err != nil {
			logger.Error("Auto-install failed: %v", err)
		}

		if settings.Verbose {
			logger.Info("Verbose mode enabled")
		}

		ctx := context.Background()
		start := time.Now()
		build := buildInfo()

		prevState, _, err := state.Load("")
		if err != nil {
			logger.Error("Failed to load run state: %v", err)
		}
		previousDurations := prevState.DurationMap()

		if settings.Silent {
			rep := runSilentMode(ctx, settings, build, start)
			finalizeRun(settings, rep)
			return
		}

		progressRelay := tui.NewProgressRelay()

		m := tui.NewModel(ctx, tui.Options{
			DryRun:            settings.DryRun,
			NonStop:           settings.NonStop,
			TimeoutMinutes:    settings.TimeoutMinutes,
			Retries:           settings.Retries,
			RetryDelay:        time.Duration(settings.RetryDelaySeconds) * time.Second,
			RetryBackoff:      settings.RetryBackoff,
			MaxRetryDelay:     time.Duration(settings.MaxRetryDelaySeconds) * time.Second,
			Only:              settings.Only,
			Exclude:           settings.Exclude,
			SelectManagers:    settings.SelectManagers,
			PreviousDurations: previousDurations,
			ProgressRelay:     progressRelay,
		})

		p := tea.NewProgram(m)
		progressRelay.Bind(p.Send)

		finalModel, err := p.Run()
		if err != nil {
			logger.Error("Error running UI: %v", err)
			fmt.Printf("Fatal error: %v\n", err)
			os.Exit(1)
		}

		// Print a final summary after the TUI exits
		if fm, ok := finalModel.(tui.Model); ok {
			results := fm.GetResults()
			rep := buildReportFromTui(results, settings, build, start)
			if settings.JSON {
				_ = report.WriteReport(rep)
			} else if len(results) > 0 {
				printPrettyResults(results)
			}
			finalizeRun(settings, rep)
		}
	},
}

// runSilentMode runs all updates without the TUI.
func runSilentMode(ctx context.Context, settings config.EffectiveSettings, build report.BuildInfo, startedAt time.Time) report.RunReport {
	logger.Info("Running in silent mode")
	if !settings.JSON {
		fmt.Println()
		fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  silent mode"))
		fmt.Println(tui.Divider())
		fmt.Println()
	}

	managers := ecosystem.GetDetectedManagers(ctx)
	managers = ecosystem.FilterManagers(managers, settings.Only, settings.Exclude)

	if len(managers) == 0 {
		if !settings.JSON {
			fmt.Println(tui.WarningStyle.Render("  ⚠ No package managers detected."))
			fmt.Println()
		}
		endedAt := time.Now()
		return report.BuildRunReport("silent", settings.Profile, startedAt, endedAt, build, nil)
	}

	var results []report.ManagerResult
	successCount := 0
	failCount := 0

	for _, m := range managers {
		if settings.DryRun {
			if !settings.JSON {
				fmt.Printf("%s %s  %s\n", tui.Bullet(), tui.BoldStyle.Render(m.Name()), tui.SubtleStyle.Render("(dry-run: skipped)"))
			}
			results = append(results, report.ManagerResult{Manager: m.Name(), Success: true, Message: "dry-run skipped", Updated: 0})
			successCount++
			continue
		}

		managerStart := time.Now()
		res := runSilentWithRetries(ctx, m, settings)
		managerDuration := time.Since(managerStart)

		results = append(results, report.ManagerResult{
			Manager:  m.Name(),
			Success:  res.Success,
			Message:  res.Message,
			Updated:  res.Updated,
			Err:      res.Error,
			Duration: managerDuration,
		})

		if !settings.JSON {
			if res.Success {
				successCount++
				if res.Updated > 0 {
					fmt.Printf("%s %s  %s\n", tui.Checkmark(), tui.BoldStyle.Render(m.Name()), tui.SubtleStyle.Render(fmt.Sprintf("(%d updated)", res.Updated)))
				} else {
					fmt.Printf("%s %s  %s\n", tui.Checkmark(), tui.BoldStyle.Render(m.Name()), tui.SubtleStyle.Render(res.Message))
				}
			} else {
				failCount++
				fmt.Printf("%s %s  %s\n", tui.Crossmark(), tui.BoldStyle.Render(m.Name()), tui.ErrorStyle.Render(res.Message))
				if res.Error != nil {
					fmt.Printf("    %s\n", tui.SubtleStyle.Render(res.Error.Error()))
				}
			}
		}
	}

	endedAt := time.Now()
	rep := report.BuildRunReport("silent", settings.Profile, startedAt, endedAt, build, results)

	if !settings.JSON {
		elapsed := endedAt.Sub(startedAt).Round(time.Millisecond)
		fmt.Println()
		fmt.Println(tui.Divider())
		fmt.Printf("  %s updated, %s failed  %s\n",
			tui.SuccessStyle.Render(fmt.Sprintf("%d", successCount)),
			tui.ErrorStyle.Render(fmt.Sprintf("%d", failCount)),
			tui.SubtleStyle.Render(fmt.Sprintf("(%s)", elapsed)))
		fmt.Println()
	}

	if settings.JSON {
		_ = report.WriteReport(rep)
	}

	return rep
}

func printPrettyResults(results []tui.UpdateMsg) {
	successCount := 0
	failCount := 0
	totalUpdated := 0

	for _, res := range results {
		if res.Result.Success {
			successCount++
			totalUpdated += res.Result.Updated
		} else {
			failCount++
		}
	}

	fmt.Println()
	fmt.Println(tui.Divider())
	fmt.Println()

	for _, res := range results {
		if res.Result.Success {
			if res.Result.Updated > 0 {
				fmt.Printf("%s %s  %s\n", tui.Checkmark(), tui.BoldStyle.Render(res.ManagerName), tui.SubtleStyle.Render(fmt.Sprintf("(%d updated)", res.Result.Updated)))
			} else {
				fmt.Printf("%s %s  %s\n", tui.Checkmark(), tui.BoldStyle.Render(res.ManagerName), tui.SubtleStyle.Render(res.Result.Message))
			}
		} else {
			fmt.Printf("%s %s  %s\n", tui.Crossmark(), tui.BoldStyle.Render(res.ManagerName), tui.ErrorStyle.Render(res.Result.Message))
		}
	}

	fmt.Println()
	fmt.Printf("  %s ecosystems, %s packages updated, %s failed\n",
		tui.HighlightStyle.Render(fmt.Sprintf("%d", len(results))),
		tui.SuccessStyle.Render(fmt.Sprintf("%d", totalUpdated)),
		tui.ErrorStyle.Render(fmt.Sprintf("%d", failCount)))
	fmt.Println()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "Run without showing UI")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output results as JSON")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate execution without making changes")
	rootCmd.PersistentFlags().BoolVar(&nonstop, "nonstop", false, "Continue past manager failures without prompting")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 10, "Timeout in minutes for each ecosystem update")
	rootCmd.PersistentFlags().StringSliceVar(&onlyManagers, "only", nil, "Run only the specified managers (repeatable)")
	rootCmd.PersistentFlags().StringSliceVar(&excludeManagers, "exclude", nil, "Skip the specified managers (repeatable)")
	rootCmd.PersistentFlags().IntVar(&retries, "retries", 0, "Retry failed managers N times before prompting")
	rootCmd.PersistentFlags().IntVar(&retryDelay, "retry-delay", 2, "Delay in seconds between retries")
	rootCmd.PersistentFlags().BoolVar(&retryBackoff, "retry-backoff", false, "Use exponential backoff between retries")
	rootCmd.PersistentFlags().IntVar(&maxRetryDelay, "max-retry-delay", 30, "Max delay in seconds between retries")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path (defaults to %APPDATA%\\Winitrix\\config.toml)")
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "", "Config profile to use")
	rootCmd.PersistentFlags().BoolVar(&selectManagers, "select", false, "Select managers interactively")
	rootCmd.PersistentFlags().BoolVar(&notifyEnabled, "notify", false, "Show a desktop notification on completion")
	rootCmd.PersistentFlags().BoolVar(&notifyOnFailure, "notify-on-failure", true, "Only notify when failures occur")
}

func runSilentWithRetries(ctx context.Context, manager ecosystem.PackageManager, settings config.EffectiveSettings) ecosystem.Result {
	timeoutMinutes := settings.TimeoutMinutes
	if timeoutMinutes <= 0 {
		timeoutMinutes = 10
	}

	attempts := settings.Retries + 1
	if attempts < 1 {
		attempts = 1
	}

	var res ecosystem.Result
	for attempt := 1; attempt <= attempts; attempt++ {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMinutes)*time.Minute)
		res = manager.UpdateAll(timeoutCtx)
		cancel()

		if res.Success || attempt == attempts {
			if !res.Success && attempt > 1 {
				res.Message = fmt.Sprintf("%s (failed after %d attempts)", res.Message, attempt)
			}
			return res
		}

		baseDelay := time.Duration(settings.RetryDelaySeconds) * time.Second
		maxDelay := time.Duration(settings.MaxRetryDelaySeconds) * time.Second
		delay := retry.ComputeDelay(attempt+1, baseDelay, settings.RetryBackoff, maxDelay)
		if delay > 0 {
			time.Sleep(delay)
		}
	}

	return res
}

func buildReportFromTui(results []tui.UpdateMsg, settings config.EffectiveSettings, build report.BuildInfo, startedAt time.Time) report.RunReport {
	managerResults := make([]report.ManagerResult, 0, len(results))
	for _, res := range results {
		managerResults = append(managerResults, report.ManagerResult{
			Manager:  res.ManagerName,
			Success:  res.Result.Success,
			Message:  res.Result.Message,
			Updated:  res.Result.Updated,
			Err:      res.Result.Error,
			Duration: res.Duration,
		})
	}
	return report.BuildRunReport("tui", settings.Profile, startedAt, time.Now(), build, managerResults)
}

func finalizeRun(settings config.EffectiveSettings, rep report.RunReport) {
	if rep.Mode == "" {
		return
	}

	if _, err := report.SaveLastReport(rep); err != nil {
		logger.Error("Failed to save last report: %v", err)
	}

	updatedState := state.UpdateFromReport(rep)
	if _, err := state.Save("", updatedState); err != nil {
		logger.Error("Failed to save run state: %v", err)
	}

	if settings.NotificationsEnabled {
		if settings.NotifyOnlyOnFailure && rep.Summary.Failed == 0 {
			return
		}
		title := "Winitrix update complete"
		message := fmt.Sprintf("%d managers, %d failed, %d packages updated", rep.Summary.Total, rep.Summary.Failed, rep.Summary.Updated)
		if err := notify.Send(title, message); err != nil {
			logger.Error("Failed to send notification: %v", err)
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
