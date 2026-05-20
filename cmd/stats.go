package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about detected ecosystems",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Close()

		ctx := context.Background()

		fmt.Println()
		fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  statistics"))
		fmt.Println(tui.Divider())
		fmt.Println()

		all := ecosystem.GetAllManagers()
		managers := ecosystem.GetDetectedManagers(ctx)

		fmt.Printf("  %s  %s / %s ecosystems detected\n\n",
			tui.HighlightStyle.Render(fmt.Sprintf("%d", len(managers))),
			tui.SuccessStyle.Render("active"),
			tui.SubtleStyle.Render(fmt.Sprintf("%d supported", len(all))))

		// Group by category
		categories := []struct {
			name     string
			managers []string
		}{
			{"Windows Native", []string{"Winget", "Scoop", "Chocolatey"}},
			{"Node.js", []string{"Npm", "pnpm", "Yarn", "Bun"}},
			{"Python", []string{"Pip", "pipx", "uv", "Poetry", "Conda"}},
			{"Rust", []string{"Cargo", "Rustup"}},
			{"DevOps", []string{"Docker", "Terraform", "Helm"}},
			{"Others", []string{".NET Tools", "Go Tools", "VSCode Extensions", "WSL (Apt)"}},
		}

		detectedNames := make(map[string]bool)
		for _, m := range managers {
			detectedNames[m.Name()] = true
		}

		for _, cat := range categories {
			hasAny := false
			for _, name := range cat.managers {
				if detectedNames[name] {
					hasAny = true
					break
				}
			}
			if !hasAny {
				continue
			}

			fmt.Printf("  %s\n", tui.DimBoldStyle.Render(cat.name))
			for _, name := range cat.managers {
				if detectedNames[name] {
					fmt.Printf("%s %s\n", tui.Checkmark(), name)
				}
			}
			fmt.Println()
		}

		fmt.Println(tui.Divider())
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
