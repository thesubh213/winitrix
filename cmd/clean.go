package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean caches for all detected ecosystems",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Close()

		ctx := context.Background()
		start := time.Now()

		fmt.Println()
		fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  cache cleaner"))
		fmt.Println(tui.Divider())
		fmt.Println()

		settings := ActiveSettings()
		managers := ecosystem.GetDetectedManagers(ctx)
		managers = ecosystem.FilterManagers(managers, settings.Only, settings.Exclude)

		if len(managers) == 0 {
			fmt.Println(tui.WarningStyle.Render("  ⚠ No package managers detected."))
			fmt.Println()
			return
		}

		successCount := 0
		failCount := 0

		for _, m := range managers {
			fmt.Printf("%s %s\n", tui.Bullet(), tui.BoldStyle.Render(m.Name()))
			res := m.Clean(ctx)

			if res.Success {
				successCount++
				fmt.Printf("%s %s\n", tui.Checkmark(), tui.SubtleStyle.Render(res.Message))
			} else {
				failCount++
				fmt.Printf("%s %s\n", tui.Crossmark(), tui.ErrorStyle.Render(res.Message))
			}
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		fmt.Println()
		fmt.Println(tui.Divider())
		fmt.Printf("  %s cleaned, %s failed  %s\n",
			tui.SuccessStyle.Render(fmt.Sprintf("%d", successCount)),
			tui.ErrorStyle.Render(fmt.Sprintf("%d", failCount)),
			tui.SubtleStyle.Render(fmt.Sprintf("(%s)", elapsed)))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
