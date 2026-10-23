package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/doctor"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/tui"
)

// Health-check parallelism and per-probe budget.
const (
	doctorWorkers = 6
	doctorTimeout = 30 * time.Second
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics across all detected ecosystems",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Close()

		ctx, stop := interruptContext()
		defer stop()
		start := time.Now()

		tui.Println()
		tui.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  diagnostics"))
		tui.Println(tui.Divider())
		tui.Println()

		// Pre-flight checks
		tui.Println(tui.DimBoldStyle.Render("  Pre-flight Checks"))
		tui.Println()

		if doctor.CheckInternetConnection(ctx) {
			tui.Printf("%s %s\n", tui.Checkmark(), "Internet connectivity")
		} else {
			tui.Printf("%s %s\n", tui.Crossmark(), tui.ErrorStyle.Render("Internet offline — updates will likely fail"))
		}

		if ok, msg := doctor.CheckPathVariables(ctx); ok {
			tui.Printf("%s %s\n", tui.Checkmark(), "System PATH integrity")
		} else {
			tui.Printf("%s %s\n", tui.Crossmark(), tui.ErrorStyle.Render(msg))
		}

		tui.Println()
		tui.Println(tui.Divider())
		tui.Println()

		// Ecosystem diagnostics
		settings := ActiveSettings()
		managers := ecosystem.DetectFiltered(ctx, settings.Only, settings.Exclude, nil)
		tui.Printf(tui.DimBoldStyle.Render("  Ecosystem Health")+" %s\n\n",
			tui.SubtleStyle.Render(fmt.Sprintf("(%d detected)", len(managers))))

		healthyCount := 0
		sickCount := 0

		// Health checks are independent; run them concurrently (bounded) so a
		// slow probe (npm doctor, WSL sudo, Docker) doesn't serialize the rest.
		type probeResult struct {
			name string
			res  ecosystem.Result
		}
		results := make([]probeResult, len(managers))
		var wg sync.WaitGroup
		sem := make(chan struct{}, doctorWorkers)
		for i, m := range managers {
			results[i].name = m.Name()
			wg.Add(1)
			sem <- struct{}{}
			go func(idx int, mgr ecosystem.PackageManager) {
				defer wg.Done()
				defer func() { <-sem }()

				probeCtx, cancel := context.WithTimeout(ctx, doctorTimeout)
				results[idx].res = mgr.Doctor(probeCtx)
				cancel()
			}(i, m)
		}
		wg.Wait()

		for _, r := range results {
			if r.res.Success {
				healthyCount++
				tui.Printf("%s %s\n", tui.Checkmark(), r.name)
			} else {
				sickCount++
				tui.Printf("%s %s  %s\n", tui.Crossmark(), tui.BoldStyle.Render(r.name), tui.ErrorStyle.Render(r.res.Message))
				if r.res.Error != nil {
					tui.Printf("     %s\n", tui.SubtleStyle.Render(r.res.Error.Error()))
				}
			}
		}

		elapsed := time.Since(start).Round(time.Millisecond)
		tui.Println()
		tui.Println(tui.Divider())
		tui.Printf("  %s healthy, %s issues  %s\n",
			tui.SuccessStyle.Render(fmt.Sprintf("%d", healthyCount)),
			tui.ErrorStyle.Render(fmt.Sprintf("%d", sickCount)),
			tui.SubtleStyle.Render(fmt.Sprintf("(%s)", elapsed)))
		tui.Println()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
