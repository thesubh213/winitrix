package cmd

import (
	"fmt"
	"io"
	"runtime"

	"github.com/thesubh213/winitrix/pkg/report"
	"github.com/thesubh213/winitrix/pkg/tui"
)

var (
	showVersion bool
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	goVersion   = runtime.Version()
)

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print build metadata and exit")
}

func printVersionInfo(out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, tui.Brand())
	fmt.Fprintln(out, tui.SubtleStyle.Render("  Unified Windows Package Orchestration Engine"))
	fmt.Fprintln(out, tui.Divider())
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  %s  %s\n", tui.DimBoldStyle.Render("Version"), tui.HighlightStyle.Render(version))
	fmt.Fprintf(out, "  %s   %s\n", tui.DimBoldStyle.Render("Commit"), tui.AccentStyle.Render(commit))
	fmt.Fprintf(out, "  %s    %s\n", tui.DimBoldStyle.Render("Built"), tui.SubtleStyle.Render(date))
	fmt.Fprintf(out, "  %s       %s\n", tui.DimBoldStyle.Render("Go"), tui.SubtleStyle.Render(goVersion))
	fmt.Fprintf(out, "  %s       %s\n", tui.DimBoldStyle.Render("OS"), tui.SubtleStyle.Render(runtime.GOOS+"/"+runtime.GOARCH))
	fmt.Fprintln(out)
}

func buildInfo() report.BuildInfo {
	return report.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: date,
		GoVersion: goVersion,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
