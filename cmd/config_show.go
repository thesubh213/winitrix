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
			fmt.Printf("Failed to load config: %v\n", err)
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
		fmt.Printf("Failed to render config as JSON: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

func printConfigShowText(output configShowOutput) {
	settings := output.Settings

	fmt.Println()
	fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  configuration"))
	fmt.Println(tui.Divider())
	fmt.Println()

	printConfigRow("Source", output.Source)
	printConfigRow("Profile", valueOrDefault(settings.Profile, "(none)"))
	printConfigRow("Config path", valueOrDefault(settings.ConfigPath, "(default location)"))

	fmt.Println()
	fmt.Println(tui.DimBoldStyle.Render("  Runtime"))
	fmt.Println()

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

	fmt.Println()
	fmt.Println(tui.DimBoldStyle.Render("  Filters"))
	fmt.Println()

	printConfigRow("Only", valueOrDefault(strings.Join(settings.Only, ", "), "(none)"))
	printConfigRow("Exclude", valueOrDefault(strings.Join(settings.Exclude, ", "), "(none)"))

	fmt.Println()
	fmt.Println(tui.DimBoldStyle.Render("  Notifications"))
	fmt.Println()

	printConfigRow("Enabled", boolLabel(settings.NotificationsEnabled))
	printConfigRow("Only on failure", boolLabel(settings.NotifyOnlyOnFailure))

	fmt.Println()
	fmt.Println(tui.Divider())
	fmt.Println()
}

func printConfigRow(label, value string) {
	fmt.Printf("  %-18s %s\n", tui.DimBoldStyle.Render(label+":"), value)
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
