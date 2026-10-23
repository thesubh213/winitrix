package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/report"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show details of the last update execution report",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Close()

		rep, err := report.LoadLastReport()
		if err != nil {
			tui.Println()
			tui.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  last report"))
			tui.Println(tui.Divider())
			tui.Println()
			tui.Println(tui.WarningStyle.Render("  ⚠ No previous update execution report found."))
			tui.Println(tui.SubtleStyle.Render("    Run winitrix to execute updates first."))
			tui.Println()
			return
		}

		tui.Println()
		tui.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  last report"))
		tui.Println(tui.Divider())
		tui.Println()

		// Metadata
		startedLabel := rep.StartedAt
		if parsedStart, err := time.Parse(time.RFC3339, rep.StartedAt); err == nil {
			startedLabel = parsedStart.Local().Format("2006-01-02 15:04:05")
		}
		duration := time.Duration(rep.DurationMs) * time.Millisecond

		tui.Printf("  %s      %s\n", tui.DimBoldStyle.Render("Ran at"), tui.HighlightStyle.Render(startedLabel))
		if rep.Profile != "" {
			tui.Printf("  %s    %s\n", tui.DimBoldStyle.Render("Profile"), tui.AccentStyle.Render(rep.Profile))
		}
		tui.Printf("  %s       %s\n", tui.DimBoldStyle.Render("Mode"), tui.SubtleStyle.Render(rep.Mode))
		tui.Printf("  %s   %s\n", tui.DimBoldStyle.Render("Duration"), tui.SubtleStyle.Render(duration.Round(time.Millisecond).String()))
		tui.Println()

		// Summary stats
		tui.Printf("  Summary: %s total ecosystems, %s packages updated, %s failed\n",
			tui.HighlightStyle.Render(fmt.Sprintf("%d", rep.Summary.Total)),
			tui.SuccessStyle.Render(fmt.Sprintf("%d", rep.Summary.Updated)),
			tui.ErrorStyle.Render(fmt.Sprintf("%d", rep.Summary.Failed)))
		tui.Println()
		tui.Println(tui.Divider())
		tui.Println()

		// Details for each manager
		for _, m := range rep.Managers {
			managerDuration := time.Duration(m.DurationMs) * time.Millisecond
			durationStr := fmt.Sprintf("(%s)", managerDuration.Round(time.Millisecond))

			if m.Success {
				if m.Updated > 0 {
					tui.Printf("%s %-20s %s %s\n",
						tui.Checkmark(),
						tui.BoldStyle.Render(m.Manager),
						tui.SuccessStyle.Render(fmt.Sprintf("%d packages updated", m.Updated)),
						tui.SubtleStyle.Render(durationStr))
				} else {
					tui.Printf("%s %-20s %s %s\n",
						tui.Checkmark(),
						tui.BoldStyle.Render(m.Manager),
						tui.SubtleStyle.Render("already up-to-date"),
						tui.SubtleStyle.Render(durationStr))
				}
			} else {
				tui.Printf("%s %-20s %s %s\n",
					tui.Crossmark(),
					tui.BoldStyle.Render(m.Manager),
					tui.ErrorStyle.Render("failed"),
					tui.SubtleStyle.Render(durationStr))
				if m.Message != "" {
					tui.Printf("     %s\n", tui.ErrorStyle.Render(m.Message))
				}
				if m.Error != "" && m.Error != m.Message {
					tui.Printf("     %s\n", tui.SubtleStyle.Render(m.Error))
				}
			}
		}

		tui.Println()
		tui.Println(tui.Divider())
		tui.Println()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
