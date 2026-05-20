# Contributing to Winitrix

First off, thank you for considering contributing to Winitrix! It's people like you that make Winitrix such a great tool for developers.

## Adding a New Package Manager

Winitrix uses a highly modular plugin architecture. To add a new package manager:

1. Create a new file in `pkg/ecosystem/` (e.g., `mytool.go`).
2. Implement the `PackageManager` interface:
   - `Name() string`
   - `Detect(ctx context.Context) bool`
   - `UpdateAll(ctx context.Context) Result`
   - `Clean(ctx context.Context) Result`
   - `Doctor(ctx context.Context) Result`
3. Add any specific error edge cases to `pkg/translator/translator.go`.
4. Register your manager in `pkg/ecosystem/manager.go` under `GetAllManagers()`.

## Development Guidelines

- **No Output Spam**: Never use `fmt.Print` or raw `exec.Command(...).Run()` that dumps to stdout. Always use `runner.RunSilent()` to preserve the TUI.
- **Go Style**: Adhere to standard Go formatting. Run `go fmt ./...` before committing.
- **Testing**: If you add new translator logic, ensure you test the error parsing behavior.

## Submitting Pull Requests

1. Fork the repository and create your branch from `main`.
2. Ensure your code passes all linting (`golangci-lint run`).
3. Describe your changes clearly in the PR description.
