# Security Policy

## Supported Versions

Currently, only the latest major release branch is supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability within Winitrix, please DO NOT open a public issue.

Instead, please send an email to the repository owner. All security vulnerabilities will be promptly addressed.

## Architecture Philosophy
- **Zero Telemetry**: Winitrix does not and will never collect usage data, analytics, or IP addresses.
- **Local Only**: Operations are passed directly to your local package managers. Winitrix does not intercept network traffic.
- **UAC Boundaries**: When Winitrix requires Administrator privileges (e.g., to run `choco upgrade all`), it requests an explicit UAC prompt using standard Windows APIs (`ShellExecute`). We do not embed hidden PowerShell bypass scripts.
