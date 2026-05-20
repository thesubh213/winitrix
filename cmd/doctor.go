package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/doctor"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics across all detected ecosystems",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Close()

		ctx := context.Background()
		start := time.Now()

		fmt.Println()
		fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  diagnostics"))
		fmt.Println(tui.Divider())
		fmt.Println()

		// Pre-flight checks
		fmt.Println(tui.DimBoldStyle.Render("  Pre-flight Checks"))
		fmt.Println()

		if doctor.CheckInternetConnection(ctx) {
			fmt.Printf("%s %s\n", tui.Checkmark(), "Internet connectivity")
		} else {
			fmt.Printf("%s %s\n", tui.Crossmark(), tui.ErrorStyle.Render("Internet offline — updates will likely fail"))
		}

		if ok, msg := doctor.CheckPathVariables(ctx); ok {
			fmt.Printf("%s %s\n", tui.Checkmark(), "System PATH integrity")
		} else {
			fmt.Printf("%s %s\n", tui.Crossmark(), tui.ErrorStyle.Render(msg))
		}

		fmt.Println()
		fmt.Println(tui.Divider())
		fmt.Println()

		// Ecosystem diagnostics
		settings := ActiveSettings()
		managers := ecosystem.GetDetectedManagers(ctx)
		managers = ecosystem.FilterManagers(managers, settings.Only, settings.Exclude)
		fmt.Printf(tui.DimBoldStyle.Render("  Ecosystem Health")+" %s\n\n",
			tui.SubtleStyle.Render(fmt.Sprintf("(%d detected)", len(managers))))

		healthyCount := 0
		sickCount := 0

		for _, m := range managers {
			res := m.Doctor(ctx)

			if res.Success {
				healthyCount++
				fmt.Printf("%s %s\n", tui.Checkmark(), m.Name())
			} else {
				sickCount++
				fmt.Printf("%s %s  %s\n", tui.Crossmark(), tui.BoldStyle.Render(m.Name()), tui.ErrorStyle.Render(res.Message))
				if res.Error != nil {
					fmt.Printf("     %s\n", tui.SubtleStyle.Render(res.Error.Error()))
				}
			}
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		fmt.Println()
		fmt.Println(tui.Divider())
		fmt.Printf("  %s healthy, %s issues  %s\n",
			tui.SuccessStyle.Render(fmt.Sprintf("%d", healthyCount)),
			tui.ErrorStyle.Render(fmt.Sprintf("%d", sickCount)),
			tui.SubtleStyle.Render(fmt.Sprintf("(%s)", elapsed)))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
