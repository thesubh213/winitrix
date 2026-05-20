# Winitrix

<div align="center">
  <img src="https://raw.githubusercontent.com/thesubh213/winitrix/main/.github/banner.png" alt="Winitrix Banner" width="700" />

  <br /><br />

  **Stop running twenty different `update` commands. Let Winitrix handle it.**

  [![Build and Validate](https://github.com/thesubh213/winitrix/actions/workflows/build.yml/badge.svg)](https://github.com/thesubh213/winitrix/actions/workflows/build.yml)
  [![Release](https://github.com/thesubh213/winitrix/actions/workflows/release.yml/badge.svg)](https://github.com/thesubh213/winitrix/actions/workflows/release.yml)
  [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/thesubh213/winitrix)](https://goreportcard.com/report/github.com/thesubh213/winitrix)
</div>

---

## Why Winitrix?

Most developers have Winget, Scoop, NPM, Pip, Cargo, Docker, and WSL scattered across their systems. Updating them requires memorizing 20 different flags, staring at walls of raw subprocess spam, and dealing with cryptically failing network connections.

Winitrix abstracts this away:

- **Beautiful TUI** — Smooth animated interface inspired by modern terminal dashboards (lazygit, btop)
- **Intelligent Error Translation** — Replaces raw errors like `0x80072efd` or `EACCES` with human-readable advice
- **UAC Elevation** — Requests administrator privileges seamlessly when needed
- **Zero Telemetry** — Strictly offline-first. No tracking, no cloud dependencies.

## Supported Ecosystems (21)

Winitrix automatically detects and orchestrates:

| Category | Managers |
| :--- | :--- |
| **Windows Native** | Winget, Scoop, Chocolatey |
| **Node.js** | NPM, pnpm, Yarn, Bun |
| **Python** | Pip, Pipx, uv, Poetry, Conda |
| **Rust** | Cargo, Rustup |
| **DevOps** | Docker, Terraform, Helm |
| **Others** | .NET Tools, Go Tools, VSCode Extensions, WSL (APT) |

---

## Installation

### Method 1: Windows Installer (Recommended)
1. Go to the [Releases](https://github.com/thesubh213/winitrix/releases) page.
2. Download `winitrix-setup-vX.X.X.exe`.
3. Run the installer (per-user, no admin required). It adds Winitrix to PATH and creates an uninstall entry.
4. Open a new terminal and run `winitrix`.

### Method 2: Portable EXE
1. Go to the [Releases](https://github.com/thesubh213/winitrix/releases) page.
2. Download `winitrix.exe` or `winitrix-vX.X.X-windows-amd64.zip`.
3. If you download the ZIP, extract `winitrix.exe`. Double-click it once. Winitrix will copy itself to `%LOCALAPPDATA%\Winitrix\bin` and add that folder to your user `PATH`.
4. Open a new terminal and run `winitrix`.

#### Verify Checksum
We provide SHA256 checksums to ensure supply-chain integrity:
```powershell
Get-FileHash winitrix.exe -Algorithm SHA256
```
Compare the output against the `checksums.txt` file attached to the release.

### Method 3: Compile from Source
Ensure you have Go 1.21+ installed:
```bash
git clone https://github.com/thesubh213/winitrix.git
cd winitrix
go build -ldflags="-s -w" -o winitrix.exe ./main.go
```

To add a custom location to PATH, use:
```bash
winitrix install --path "C:\\Tools\\Winitrix"
```

### Uninstall
Use **Apps & Features** to uninstall if you installed with the setup EXE. For portable installs, remove `%LOCALAPPDATA%\Winitrix\bin` and delete the PATH entry.

---

## Usage

Simply run:
```bash
winitrix
```
Winitrix will scan your system, detect installed managers, and begin the update flow.

### Commands
| Command | Description |
| :--- | :--- |
| `winitrix` | Launch the interactive TUI update flow |
| `winitrix doctor` | Run pre-flight diagnostics (DNS, PATH, ecosystem health) |
| `winitrix clean` | Prune caches across all detected ecosystems |
| `winitrix stats` | View detected ecosystems grouped by category |
| `winitrix logs` | Open the internal debug log file |
| `winitrix logs --bundle` | Create a diagnostic bundle (logs, config, state) |
| `winitrix install` | Install Winitrix to PATH (Windows user profile) |
| `winitrix config` | Manage the Winitrix config file |
| `winitrix schedule` | Print or apply a Task Scheduler entry |
| `winitrix --version` | Display build metadata and version info |

### Flags
| Flag | Description |
| :--- | :--- |
| `--silent` | Run without the TUI (headless mode) |
| `--dry-run` | Simulate updates without executing commands |
| `--json` | Output results as structured JSON |
| `--timeout N` | Set per-ecosystem timeout in minutes (default: 10) |
| `--verbose` | Enable extended debug logging |
| `--nonstop` | Continue past manager failures without prompting |
| `--only <name>` | Run only selected managers (repeatable) |
| `--exclude <name>` | Skip selected managers (repeatable) |
| `--retries N` | Retry failed managers N times before prompting |
| `--retry-delay N` | Delay in seconds between retries |
| `--retry-backoff` | Use exponential backoff between retries |
| `--max-retry-delay N` | Max delay in seconds between retries |
| `--config <path>` | Use a specific config file |
| `--profile <name>` | Use a named profile from the config |
| `--select` | Select managers interactively before updating |
| `--notify` | Show a desktop notification on completion |
| `--notify-on-failure` | Only notify when failures occur |

## Configuration

Create a starter config file:
```bash
winitrix config init
```

Default config location:
- `%APPDATA%\Winitrix\config.toml`

Config highlights:
- Profiles for different manager sets
- Default timeouts, retries, and filters
- Optional TUI selection and notifications

## Scheduling

Print a Task Scheduler command (weekly on Mondays at 09:00):
```bash
winitrix schedule
```

Apply the schedule directly:
```bash
winitrix schedule --apply
```

Customize the schedule:
```bash
winitrix schedule --interval weekly --days Mon,Wed,Fri --time 19:00
```

## Reports and State

When `--json` is used, Winitrix outputs a structured report with per-manager durations and error categories.

Cached files:
- `%LOCALAPPDATA%\Winitrix\state.json`
- `%LOCALAPPDATA%\Winitrix\last_report.json`

---

## Architecture

Winitrix uses a highly modular plugin architecture:

```
winitrix/
├── cmd/                    # Cobra CLI commands
│   ├── root.go             # Main update flow (TUI + silent mode)
│   ├── doctor.go           # Diagnostics command
│   ├── clean.go            # Cache cleaning command
│   ├── stats.go            # Ecosystem statistics
│   ├── version.go          # Build metadata
│   └── logs.go             # Log file viewer
├── pkg/
│   ├── ecosystem/          # Plugin engine (21 managers)
│   │   ├── manager.go      # PackageManager interface
│   │   ├── winget.go       # Winget adapter
│   │   ├── scoop.go        # Scoop adapter
│   │   └── ...             # 19 more adapters
│   ├── tui/                # Bubble Tea state machine
│   │   ├── model.go        # TUI model + rendering
│   │   └── style.go        # Lipgloss color palette
│   ├── translator/         # Error translation engine
│   ├── runner/             # Silent command executor
│   ├── elevation/          # UAC privilege handling
│   ├── logger/             # File-based logging
│   └── doctor/             # Pre-flight diagnostics
├── .github/workflows/      # CI/CD pipelines
├── main.go                 # Entry point
└── go.mod                  # Dependencies
```

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md) for details on adding new package managers.

## Security
See [SECURITY.md](SECURITY.md) for our vulnerability disclosure policy and zero-telemetry philosophy.

## License
MIT License. See [LICENSE](LICENSE) for details.
