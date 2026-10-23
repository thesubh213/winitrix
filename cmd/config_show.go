package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/config"
	"github.com/thesubh213/winitrix/pkg/tui"
)

type configShowOutput struct {
	Loaded   bool                     `json:"loaded"`
	Source   string                   `json:"source"`
	Settings config.EffectiveSettings `json:"settings"`
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the effective configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, resolvedPath, found, err := config.Load(configPath)
		if err != nil {
			tui.Printf("Failed to load config: %v\n", err)
			return
		}

		profile := config.ResolveProfile(cfg, profileName)
		settings := config.Effective(cfg, profile)
		settings.Profile = profile
		settings.ConfigPath = resolvedPath

		applyFlagOverrides(cmd, &settings)

		output := buildConfigShowOutput(settings, found)
		if settings.JSON {
			printConfigShowJSON(output)
			return
		}

		printConfigShowText(output)
	},
}

func buildConfigShowOutput(settings config.EffectiveSettings, loaded bool) configShowOutput {
	source := "default settings"
	if loaded {
		source = "loaded config"
	}
	if settings.ConfigPath != "" {
		source = fmt.Sprintf("%s (%s)", source, settings.ConfigPath)
	}

	return configShowOutput{
		Loaded:   loaded,
		Source:   source,
		Settings: settings,
	}
}

func printConfigShowJSON(output configShowOutput) {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		tui.Printf("Failed to render config as JSON: %v\n", err)
		return
	}

	tui.Println(string(data))
}

func printConfigShowText(output configShowOutput) {
	settings := output.Settings

	tui.Println()
	tui.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  configuration"))
	tui.Println(tui.Divider())
	tui.Println()

	printConfigRow("Source", output.Source)
	printConfigRow("Profile", valueOrDefault(settings.Profile, "(none)"))
	printConfigRow("Config path", valueOrDefault(settings.ConfigPath, "(default location)"))

	tui.Println()
	tui.Println(tui.DimBoldStyle.Render("  Runtime"))
	tui.Println()

	printConfigRow("Dry run", boolLabel(settings.DryRun))
	printConfigRow("Silent", boolLabel(settings.Silent))
	printConfigRow("JSON", boolLabel(settings.JSON))
	printConfigRow("Verbose", boolLabel(settings.Verbose))
	printConfigRow("Nonstop", boolLabel(settings.NonStop))
	printConfigRow("Timeout", fmt.Sprintf("%d minutes", settings.TimeoutMinutes))
	printConfigRow("Retries", fmt.Sprintf("%d", settings.Retries))
	printConfigRow("Retry delay", fmt.Sprintf("%d seconds", settings.RetryDelaySeconds))
	printConfigRow("Retry backoff", boolLabel(settings.RetryBackoff))
	printConfigRow("Max retry delay", fmt.Sprintf("%d seconds", settings.MaxRetryDelaySeconds))
	printConfigRow("Select managers", boolLabel(settings.SelectManagers))

	tui.Println()
	tui.Println(tui.DimBoldStyle.Render("  Interface"))
	tui.Println()

	printConfigRow("Animations", boolLabel(settings.Animations))
	printConfigRow("Alt screen", boolLabel(settings.AltScreen))
	printConfigRow("ASCII glyphs", boolLabel(settings.ASCII))

	tui.Println()
	tui.Println(tui.DimBoldStyle.Render("  Filters"))
	tui.Println()

	printConfigRow("Only", valueOrDefault(strings.Join(settings.Only, ", "), "(none)"))
	printConfigRow("Exclude", valueOrDefault(strings.Join(settings.Exclude, ", "), "(none)"))

	tui.Println()
	tui.Println(tui.DimBoldStyle.Render("  Notifications"))
	tui.Println()

	printConfigRow("Enabled", boolLabel(settings.NotificationsEnabled))
	printConfigRow("Only on failure", boolLabel(settings.NotifyOnlyOnFailure))

	tui.Println()
	tui.Println(tui.Divider())
	tui.Println()
}

func printConfigRow(label, value string) {
	tui.Printf("  %-18s %s\n", tui.DimBoldStyle.Render(label+":"), value)
}

func boolLabel(value bool) string {
	if value {
		return tui.SuccessStyle.Render("enabled")
	}

	return tui.SubtleStyle.Render("disabled")
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
