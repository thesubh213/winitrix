package translator

import "testing"

func TestParseError_GenericTimeout(t *testing.T) {
	msg := ParseError("Winget", "some output context deadline exceeded and stuff")
	if msg != "Operation timed out. The package manager took too long to respond." {
		t.Errorf("Expected timeout message, got: %s", msg)
	}
}

func TestParseError_WingetNetworkErrors(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{"Firewall error 0x80072efd", "Error 0x80072efd something", "Network connection failed. Please check your firewall or internet connection."},
		{"Timeout error 0x80072ee2", "Error 0x80072EE2 blah", "Network connection failed. Please check your firewall or internet connection."},
		{"Access denied 0x80070005", "Error 0x80070005 denied", "Administrator privileges required to perform this action."},
		{"Access denied text", "Access is denied for this action", "Administrator privileges required to perform this action."},
		{"Windows Update issue", "Error 0x80240438 blah", "Windows Update service or network issue preventing download."},
		{"Package not found", "No installed package found matching input criteria", "Package not found or already up to date."},
		{"Generic winget error", "some random winget output", "Winget encountered an error. Check logs for details."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseError("Winget", tt.output)
			if got != tt.expected {
				t.Errorf("ParseError(\"Winget\", %q)\n  got:  %q\n  want: %q", tt.output, got, tt.expected)
			}
		})
	}
}

func TestParseError_ScoopErrors(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{"Folder in use", "Error: folder in use by another process", "An application folder is currently in use. Please close the application and try again."},
		{"Hash check failed", "Error: Hash check failed for package", "Package hash mismatch. The upstream provider may have updated the file unexpectedly. Try running 'winitrix clean'."},
		{"Git DNS failure", "fatal: unable to access the repository", "Network failure. Scoop couldn't reach the Git repository."},
		{"Git fast-forward", "error: cannot fast-forward merge", "Scoop bucket Git repository is corrupted. Try running 'winitrix clean'."},
		{"Generic scoop error", "some random scoop output", "Scoop encountered an error. Check logs for details."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseError("Scoop", tt.output)
			if got != tt.expected {
				t.Errorf("ParseError(\"Scoop\", %q)\n  got:  %q\n  want: %q", tt.output, got, tt.expected)
			}
		})
	}
}

func TestParseError_NpmAndVariants(t *testing.T) {
	tests := []struct {
		manager  string
		output   string
		expected string
	}{
		{"Npm", "Error: EACCES permission denied", "NPM requires administrator permissions to update global packages."},
		{"Npm", "error EPERM operation not permitted", "NPM requires administrator permissions to update global packages."},
		{"Npm", "error ECONNREFUSED 127.0.0.1", "NPM registry is unreachable. Check your internet connection."},
		{"Npm", "error ENOTFOUND registry.npmjs.org", "NPM registry is unreachable. Check your internet connection."},
		{"Npm", "cb() never called", "NPM encountered an internal cache or network timeout error. Try running 'winitrix clean'."},
		{"Yarn", "error EACCES permission denied", "NPM requires administrator permissions to update global packages."},
		{"Bun", "error ECONNREFUSED 127.0.0.1", "NPM registry is unreachable. Check your internet connection."},
		{"pnpm", "generic pnpm error", "NPM encountered an error. Check logs for details."},
	}

	for _, tt := range tests {
		name := tt.manager + "_" + tt.output
		if len(name) > 40 {
			name = name[:40]
		}
		t.Run(name, func(t *testing.T) {
			got := ParseError(tt.manager, tt.output)
			if got != tt.expected {
				t.Errorf("ParseError(%q, %q)\n  got:  %q\n  want: %q", tt.manager, tt.output, got, tt.expected)
			}
		})
	}
}

func TestParseError_PipAndVariants(t *testing.T) {
	tests := []struct {
		manager  string
		output   string
		expected string
	}{
		{"Pip", "error: externally-managed-environment", "Pip environment is externally managed (PEP 668). Global updates are blocked. Use virtual environments or pipx."},
		{"Pip", "error: Permission denied during install", "Administrator privileges required to modify global Python packages."},
		{"Pip", "error: access denied to site-packages", "Administrator privileges required to modify global Python packages."},
		{"pipx", "some pipx error", "Pip encountered an error. Check logs for details."},
		{"Conda", "some conda error", "Pip encountered an error. Check logs for details."},
		{"Poetry", "some poetry error", "Pip encountered an error. Check logs for details."},
		{"uv", "some uv error", "Pip encountered an error. Check logs for details."},
	}

	for _, tt := range tests {
		name := tt.manager + "_" + tt.output
		if len(name) > 40 {
			name = name[:40]
		}
		t.Run(name, func(t *testing.T) {
			got := ParseError(tt.manager, tt.output)
			if got != tt.expected {
				t.Errorf("ParseError(%q, %q)\n  got:  %q\n  want: %q", tt.manager, tt.output, got, tt.expected)
			}
		})
	}
}

func TestParseError_OtherManagers(t *testing.T) {
	tests := []struct {
		manager  string
		output   string
		expected string
	}{
		{"Chocolatey", "not recognized as an internal command", "Administrator privileges required or Chocolatey not configured properly."},
		{"Chocolatey", "access denied to choco directory", "Administrator privileges required or Chocolatey not configured properly."},
		{"Chocolatey", "random error", "Chocolatey encountered an error. Check logs for details."},
		{"Cargo", "error: could not compile package", "Cargo compilation failed for a package. Check C++ build tools and logs."},
		{"Cargo", "random cargo error", "Cargo/Rustup encountered an error."},
		{"Rustup", "random rustup error", "Cargo/Rustup encountered an error."},
		{"Docker", "error during connect to daemon", "Docker engine is not running. Start Docker Desktop first."},
		{"Docker", "Cannot connect to Docker daemon", "Docker engine is not running. Start Docker Desktop first."},
		{"Docker", "random docker error", "Docker encountered an error."},
		{"WSL", "sudo: a password is required", "WSL apt upgrade requires passwordless sudo setup for Winitrix."},
		{"WSL", "E: could not get lock /var/lib/dpkg/lock", "Another APT process is running in WSL (e.g., background updates). Try again later."},
		{"WSL", "random wsl error", "WSL apt encountered an error."},
	}

	for _, tt := range tests {
		name := tt.manager + "_" + tt.output
		if len(name) > 40 {
			name = name[:40]
		}
		t.Run(name, func(t *testing.T) {
			got := ParseError(tt.manager, tt.output)
			if got != tt.expected {
				t.Errorf("ParseError(%q, %q)\n  got:  %q\n  want: %q", tt.manager, tt.output, got, tt.expected)
			}
		})
	}
}

func TestParseError_UnknownManager(t *testing.T) {
	got := ParseError("FooBar", "some error")
	expected := "FooBar encountered an unknown error during execution."
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}
