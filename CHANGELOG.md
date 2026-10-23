# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-10-16

### Added
- **Animated gradient dashboard TUI** (Bubble Tea v2 + Lip Gloss v2): animated gradient wordmark, gradient progress bars with sub-cell precision, state-colored panels with transition pulses, responsive layout down to narrow terminals, alternate-screen mode, a native terminal progress bar (where supported), and a live detection progress view.
- **Hidden `--demo` mode**: exercises the full UI against simulated package managers (streamed progress plus a failing manager) without touching real packages.
- **New UI settings**: `[ui] animations`, `[ui] alt_screen`, and `[ui] ascii` (plus `WINITRIX_ASCII=1` for legacy consoles); `NO_COLOR` is respected throughout.
- **Live output pane**: press `d` while updating to expand a panel showing the running ecosystem's actual command output (last ~200 lines, live-updating, with light severity highlighting); `esc` (or `d`) collapses it. Also available on the failure prompt, where it shows the output that led to the error.

### Changed
- **Much faster startup**: ecosystem detection now probes managers in parallel with a per-probe timeout, skips probes for tools that are not on `PATH`, and the artificial 800 ms scan delay is gone. `winitrix doctor` runs its ecosystem health checks with the same bounded parallelism (30s → ~8s on a typical machine), and `--only`/`--exclude` filters are applied before probing.
- **Cheaper progress pipeline**: progress lines are throttled *before* parsing, per-line regex compilation is gone, and streamed output is captured in bounded buffers (2 MiB per stream) instead of growing without limit.
- **Accurate update counts**: streamed output is fully drained before `Wait`, so trailing "Successfully installed …" summary lines are no longer dropped.
- **Real cancellation**: Ctrl+C stops the in-flight package manager and terminates its process tree — parent and children — instead of leaving it running in the background. Non-TUI commands use a signal-aware context as well, and quitting during the scan cancels pending probes.
- **Clean machine-readable output**: `--json` and piped (non-TTY) runs force silent mode, so JSON is never interleaved with TUI frames.
- **Deterministic `--only`/`--exclude` matching**: exact name/alias matches now take precedence (`--only pip` no longer selects pipx).
- **Run-state history is merged**, so ETA estimates survive filtered runs such as `--only winget`.
- **Atomic writes** for state, report, and config files.
- Corrupt config files produce a warning and fall back to defaults instead of breaking every command; unknown `--profile` values warn instead of silently doing nothing.
- Toolchain bumped to Go 1.26 (required by the Charm v2 libraries); CI now reads the Go version from `go.mod`.

### Fixed
- **PATH safety**: the installer and auto-install now abort on a registry read error and preserve the `REG_EXPAND_SZ` type of the user PATH instead of replacing it, which could previously wipe the PATH.
- **Cancel dialog correctness**: the prompt keys (`y`/`n`) now match the handlers, the dialog no longer freezes while an update finishes, and cancelling records the interrupted manager in the report.
- Fixed a potential panic in progress parsing on exotic Unicode output.
- `clean` no longer reports success when the underlying command failed (Bun, Conda, uv, pnpm, Yarn, Poetry, .NET Tools, Go Tools, Scoop).
- Scoop detection no longer depends on the literal string "Scoop" in its version output; Scoop doctor handles warnings and nil errors correctly.
- Retry backoff can no longer overflow into nonsensical delays; the log file closes idempotently; `schedule` refuses to build a task from an unresolved executable path; notifications are bounded by a timeout so they cannot hang exit.

## [1.0.0] - 2026-05-20

### Added
- **Core Architecture**: Initial release of the Winitrix Unified Windows Package Orchestration Engine.
- **21 Supported Ecosystems**: Complete detection and execution engines for Winget, Scoop, Chocolatey, NPM, pnpm, Yarn, Bun, Pip, Pipx, uv, Poetry, Conda, Cargo, Rustup, Docker, Terraform, Helm, .NET tools, Go tools, VSCode extensions, and WSL (APT).
- **Beautiful TUI**: Smooth, animated Bubble Tea interface featuring a Deep Sky Blue color palette and elegant state transitions.
- **Intelligent Error Translation**: Replaced raw subprocess spam (e.g., `0x80072efd`, `EACCES`) with human-friendly, actionable advice across all 21 ecosystems.
- **Edge Case Hardening**: Added explicit handling for Python PEP 668, Yarn 2+ deprecations, Chocolatey pending reboots (Exit Code 3010), and Docker daemon disconnects.
- **Interactive Fallback**: The UI now gracefully pauses and prompts `(y/N)` if an ecosystem fails to update, preventing terminal crashes.
- **Silent Execution Runner**: Utilized Windows `SysProcAttr` with `HideWindow: true` to prevent console flashing during background updates.
- **Pre-flight Diagnostics**: `winitrix doctor` now performs a DNS network connectivity ping and checks the `PATH` variable for corruption.
- **Security & Privacy**: Offline-first architecture. Zero telemetry. Zero hidden scripts.
- **CI/CD Integration**: Automated GitHub Actions pipelines for deterministic builds and release packaging.
- **Windows Installer**: Built Inno Setup compiler integration with auto-PATH configuration, desktop/menu shortcuts, and clean uninstallation.
- **Portable first-run auto-install**: Auto-copies the EXE into `%LOCALAPPDATA%\Winitrix\bin` and adds it to user PATH on first double-click.
- **Manager filters & retries**: Command-line controls for `--only`, `--exclude`, `--retries`, `--retry-delay`, and exponential backoff (`--retry-backoff`).
- **TOML configuration**: Local custom profiles and configurations via `winitrix config init`.
- **Structured JSON reports**: Auto-generated run-state history (`state.json`) and run reports (`last_report.json`).
- **Task Scheduler**: Built-in scheduling setup via `winitrix schedule`.
- **Desktop notifications**: Integration with Windows Action Center (`--notify`).
- **Diagnostic bundle**: Run `winitrix logs --bundle` to generate compressed logs for diagnostics.

### Changed
- Structured progress parsing across more provider outputs, shown in the TUI.
- Release workflow now builds and publishes the setup EXE.

### Fixed
- Windows-only code paths now isolated with build tags and non-Windows stubs to avoid cross-platform compile errors.

### Security
- Integrated UAC elevation detection utilizing `ShellExecute` to automatically elevate the application safely when modifying global package managers.

## [1.0.2] - 2026-06-03

### Added
- **Config Edit Subcommand**: Added `winitrix config edit` to open the active Winitrix config file in the default system text editor (or notepad).
- **Execution Report CLI**: Added a styled `winitrix report` command using Lipgloss to view detailed outcomes of the last update run.

### Changed
- **WSL Passwordless Sudo Diagnostic**: Enhanced `winitrix doctor` to verify passwordless `sudo` capability for WSL apt-get commands.

## [1.0.1] - 2026-05-26

### Added
- **Configuration Viewer**: Added `winitrix config show` to print the merged runtime settings, with JSON output for automation.

### Changed
- **Windows Launcher**: The Start Menu and desktop shortcuts now open Winitrix through a terminal wrapper so double-clicking from Windows launches a real console session.
- **Documentation**: Clarified how to launch Winitrix from Windows shortcuts and updated the release guide for patch releases.

### Fixed
- **Accurate Update Counts**: Yarn and pipx now report real package update counts instead of placeholder values.
- **CI Validation**: Build and release workflows now check go.mod/go.sum with a supported tidy/diff step and only run `-race` on amd64 runners.
