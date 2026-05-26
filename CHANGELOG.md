# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2026-05-26

### Fixed
- **CI Validation**: Build and release workflows now run `-race` only on amd64 runners, avoiding failures on unsupported Windows architectures.

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

## [1.0.1] - 2026-05-26

### Added
- **Configuration Viewer**: Added `winitrix config show` to print the merged runtime settings, with JSON output for automation.

### Changed
- **Windows Launcher**: The Start Menu and desktop shortcuts now open Winitrix through a terminal wrapper so double-clicking from Windows launches a real console session.
- **Documentation**: Clarified how to launch Winitrix from Windows shortcuts and updated the release guide for patch releases.

### Fixed
- **Accurate Update Counts**: Yarn and pipx now report real package update counts instead of placeholder values.
