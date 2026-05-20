# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-05-15

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

### Security
- Integrated UAC elevation detection utilizing `ShellExecute` to automatically elevate the application safely when modifying global package managers.

## [Unreleased]

### Added
- Windows installer (Inno Setup) with PATH task, shortcuts, and uninstall entry.
- Portable first-run auto-install to copy the EXE into `%LOCALAPPDATA%\Winitrix\bin` and add it to user PATH.
- Manager filters (`--only`, `--exclude`) and retry controls (`--retries`, `--retry-delay`).
- TOML configuration with profiles and defaults (`winitrix config init`).
- Structured JSON reports with per-manager durations and error categories.
- Run state cache for ETAs (`state.json`) and last report cache (`last_report.json`).
- Interactive manager selection UI (`--select`).
- Task Scheduler helper command (`winitrix schedule`).
- Desktop completion notifications (`--notify`).
- Diagnostic bundle support (`winitrix logs --bundle`).

### Changed
- Structured progress parsing across more provider outputs, shown in the TUI.
- Release workflow now builds and publishes the setup EXE.
- Retry logic now supports optional exponential backoff (`--retry-backoff`).

### Fixed
- Windows-only code paths now isolated with build tags and non-Windows stubs to avoid cross-platform compile errors.
