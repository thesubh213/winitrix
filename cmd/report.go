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
			fmt.Println()
			fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  last report"))
			fmt.Println(tui.Divider())
			fmt.Println()
			fmt.Println(tui.WarningStyle.Render("  ⚠ No previous update execution report found."))
			fmt.Println(tui.SubtleStyle.Render("    Run winitrix to execute updates first."))
			fmt.Println()
			return
		}

		fmt.Println()
		fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  last report"))
		fmt.Println(tui.Divider())
		fmt.Println()

		// Metadata
		parsedStart, _ := time.Parse(time.RFC3339, rep.StartedAt)
		duration := time.Duration(rep.DurationMs) * time.Millisecond

		fmt.Printf("  %s      %s\n", tui.DimBoldStyle.Render("Ran at"), tui.HighlightStyle.Render(parsedStart.Local().Format("2006-01-02 15:04:05")))
		if rep.Profile != "" {
			fmt.Printf("  %s    %s\n", tui.DimBoldStyle.Render("Profile"), tui.AccentStyle.Render(rep.Profile))
		}
		fmt.Printf("  %s       %s\n", tui.DimBoldStyle.Render("Mode"), tui.SubtleStyle.Render(rep.Mode))
		fmt.Printf("  %s   %s\n", tui.DimBoldStyle.Render("Duration"), tui.SubtleStyle.Render(duration.Round(time.Millisecond).String()))
		fmt.Println()

		// Summary stats
		fmt.Printf("  Summary: %s total ecosystems, %s packages updated, %s failed\n",
			tui.HighlightStyle.Render(fmt.Sprintf("%d", rep.Summary.Total)),
			tui.SuccessStyle.Render(fmt.Sprintf("%d", rep.Summary.Updated)),
			tui.ErrorStyle.Render(fmt.Sprintf("%d", rep.Summary.Failed)))
		fmt.Println()
		fmt.Println(tui.Divider())
		fmt.Println()

		// Details for each manager
		for _, m := range rep.Managers {
			managerDuration := time.Duration(m.DurationMs) * time.Millisecond
			durationStr := fmt.Sprintf("(%s)", managerDuration.Round(time.Millisecond))

			if m.Success {
				if m.Updated > 0 {
					fmt.Printf("%s %-20s %s %s\n",
						tui.Checkmark(),
						tui.BoldStyle.Render(m.Manager),
						tui.SuccessStyle.Render(fmt.Sprintf("%d packages updated", m.Updated)),
						tui.SubtleStyle.Render(durationStr))
				} else {
					fmt.Printf("%s %-20s %s %s\n",
						tui.Checkmark(),
						tui.BoldStyle.Render(m.Manager),
						tui.SubtleStyle.Render("already up-to-date"),
						tui.SubtleStyle.Render(durationStr))
				}
			} else {
				fmt.Printf("%s %-20s %s %s\n",
					tui.Crossmark(),
					tui.BoldStyle.Render(m.Manager),
					tui.ErrorStyle.Render("failed"),
					tui.SubtleStyle.Render(durationStr))
				if m.Message != "" {
					fmt.Printf("     %s\n", tui.ErrorStyle.Render(m.Message))
				}
				if m.Error != "" && m.Error != m.Message {
					fmt.Printf("     %s\n", tui.SubtleStyle.Render(m.Error))
				}
			}
		}

		fmt.Println()
		fmt.Println(tui.Divider())
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
