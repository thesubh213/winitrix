package translator

import (
	"strings"
)

// ParseError analyzes the raw output from a package manager and returns a human-friendly string.
func ParseError(manager string, output string) string {
	lowerOutput := strings.ToLower(output)

	// Generic Timeout Error from context
	if strings.Contains(lowerOutput, "context deadline exceeded") {
		return "Operation timed out. The package manager took too long to respond."
	}

	switch manager {
	case "Winget":
		return parseWingetError(lowerOutput, output)
	case "Scoop":
		return parseScoopError(lowerOutput, output)
	case "Npm", "Yarn", "Bun", "pnpm":
		return parseNpmError(lowerOutput, output)
	case "Pip", "pipx", "Conda", "Poetry", "uv":
		return parsePipError(lowerOutput, output)
	case "Chocolatey":
		if strings.Contains(lowerOutput, "not recognized") || strings.Contains(lowerOutput, "access denied") {
			return "Administrator privileges required or Chocolatey not configured properly."
		}
		return "Chocolatey encountered an error. Check logs for details."
	case "Cargo", "Rustup":
		if strings.Contains(lowerOutput, "could not compile") {
			return "Cargo compilation failed for a package. Check C++ build tools and logs."
		}
		return "Cargo/Rustup encountered an error."
	case "Docker":
		if strings.Contains(lowerOutput, "daemon") || strings.Contains(lowerOutput, "error during connect") {
			return "Docker engine is not running. Start Docker Desktop first."
		}
		return "Docker encountered an error."
	case "WSL":
		if strings.Contains(lowerOutput, "sudo") {
			return "WSL apt upgrade requires passwordless sudo setup for Winitrix."
		}
		if strings.Contains(lowerOutput, "could not get lock") || strings.Contains(lowerOutput, "/var/lib/dpkg/lock") {
			return "Another APT process is running in WSL (e.g., background updates). Try again later."
		}
		return "WSL apt encountered an error."
	default:
		return manager + " encountered an unknown error during execution."
	}
}

func parseWingetError(lowerOutput, raw string) string {
	if strings.Contains(lowerOutput, "0x80072efd") || strings.Contains(lowerOutput, "0x80072ee2") {
		return "Network connection failed. Please check your firewall or internet connection."
	}
	if strings.Contains(lowerOutput, "0x80070005") || strings.Contains(lowerOutput, "access is denied") {
		return "Administrator privileges required to perform this action."
	}
	if strings.Contains(lowerOutput, "0x80240438") {
		return "Windows Update service or network issue preventing download."
	}
	if strings.Contains(lowerOutput, "no installed package found matching input criteria") {
		return "Package not found or already up to date."
	}
	// Fallback to a generic winget message
	return "Winget encountered an error. Check logs for details."
}

func parseScoopError(lowerOutput, raw string) string {
	if strings.Contains(lowerOutput, "folder in use") || strings.Contains(lowerOutput, "permission denied") {
		return "An application folder is currently in use. Please close the application and try again."
	}
	if strings.Contains(lowerOutput, "hash check failed") {
		return "Package hash mismatch. The upstream provider may have updated the file unexpectedly. Try running 'winitrix clean'."
	}
	if strings.Contains(lowerOutput, "fatal: unable to access") || strings.Contains(lowerOutput, "could not resolve host") {
		return "Network failure. Scoop couldn't reach the Git repository."
	}
	if strings.Contains(lowerOutput, "cannot fast-forward") {
		return "Scoop bucket Git repository is corrupted. Try running 'winitrix clean'."
	}
	return "Scoop encountered an error. Check logs for details."
}

func parseNpmError(lowerOutput, raw string) string {
	if strings.Contains(lowerOutput, "eacces") || strings.Contains(lowerOutput, "eperm") {
		return "NPM requires administrator permissions to update global packages."
	}
	if strings.Contains(lowerOutput, "econnrefused") || strings.Contains(lowerOutput, "enotfound") {
		return "NPM registry is unreachable. Check your internet connection."
	}
	if strings.Contains(lowerOutput, "cb() never called") {
		return "NPM encountered an internal cache or network timeout error. Try running 'winitrix clean'."
	}
	return "NPM encountered an error. Check logs for details."
}

func parsePipError(lowerOutput, raw string) string {
	if strings.Contains(lowerOutput, "externally-managed-environment") {
		return "Pip environment is externally managed (PEP 668). Global updates are blocked. Use virtual environments or pipx."
	}
	if strings.Contains(lowerOutput, "permission denied") || strings.Contains(lowerOutput, "access denied") {
		return "Administrator privileges required to modify global Python packages."
	}
	return "Pip encountered an error. Check logs for details."
}
