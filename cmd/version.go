package cmd

import (
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
	tui.Fprintln(out)
	tui.Fprintln(out, tui.Brand())
	tui.Fprintln(out, tui.SubtleStyle.Render("  Unified Windows Package Orchestration Engine"))
	tui.Fprintln(out, tui.Divider())
	tui.Fprintln(out)
	tui.Fprintf(out, "  %s  %s\n", tui.DimBoldStyle.Render("Version"), tui.HighlightStyle.Render(version))
	tui.Fprintf(out, "  %s   %s\n", tui.DimBoldStyle.Render("Commit"), tui.AccentStyle.Render(commit))
	tui.Fprintf(out, "  %s    %s\n", tui.DimBoldStyle.Render("Built"), tui.SubtleStyle.Render(date))
	tui.Fprintf(out, "  %s       %s\n", tui.DimBoldStyle.Render("Go"), tui.SubtleStyle.Render(goVersion))
	tui.Fprintf(out, "  %s       %s\n", tui.DimBoldStyle.Render("OS"), tui.SubtleStyle.Render(runtime.GOOS+"/"+runtime.GOARCH))
	tui.Fprintln(out)
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
