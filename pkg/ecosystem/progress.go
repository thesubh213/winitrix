package ecosystem

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	ansiEscapeRe        = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	percentageRe        = regexp.MustCompile(`(?i)(\d{1,3})%`)
	versionTransitionRe = regexp.MustCompile(`(?i)^(.+?)\s+([^\s]+)\s*->\s*([^\s]+)$`)
	wingetProgressRe    = regexp.MustCompile(`(?i)^(?:installing|updating|upgrading):\s+(.+?)\s+\[.*\]$`)
	wingetDownloadRe    = regexp.MustCompile(`(?i)^downloading:\s+(.+?)\s+\[.*\]$`)
	wingetFoundRe       = regexp.MustCompile(`(?i)^found\s+(.+?)\s+\[.*\]$`)
	wingetSuccessRe     = regexp.MustCompile(`(?i)^successfully installed\s+(.+?)\s+\[.*\]$`)
	scoopUpdateRe       = regexp.MustCompile(`(?i)^updating\s+'?(.+?)'?\s+\((.+?)\s*->\s*(.+?)\)$`)
	poetryPackageOpRe   = regexp.MustCompile(`(?i)^\s*[-*]\s+(installing|updating|removing)\s+(.+?)\s*(?:\((.+?)\))?$`)
	aptUpgradeListRe    = regexp.MustCompile(`(?i)^the following packages will be (updated|upgraded|installed):$`)
	aptSummaryRe        = regexp.MustCompile(`(?i)^(\d+) upgraded, (\d+) newly installed, (\d+) to remove`)
	packageCountRe      = regexp.MustCompile(`(?i)^(\d+) packages? (?:updated|upgraded|installed|removed|changed)$`)
	componentInstallRe  = regexp.MustCompile(`(?i)^installing component ['\"]?(.+?)['\"]?$`)
	repoUpdateSuccessRe = regexp.MustCompile(`(?i)^\.\.\.successfully got an update from the ["“](.+?)["”] chart repository$`)
	npmVersionChangeRe  = regexp.MustCompile(`(?i)^(.+?)\s+([^\s]+)\s*->\s*([^\s]+)$`)
	chocoPackageOpRe    = regexp.MustCompile(`(?i)^(upgrading|installing)\s+([^\s]+)`)
	pipCollectingRe     = regexp.MustCompile(`(?i)^collecting\s+(.+)$`)
	pipInstallingRe     = regexp.MustCompile(`(?i)^installing collected packages: (.+)$`)
)

// ParseProgressLine converts raw provider output into structured progress data.
func ParseProgressLine(managerName, line string) (ProgressUpdate, bool) {
	normalized := normalizeProgressLine(line)
	if normalized == "" {
		return ProgressUpdate{}, false
	}

	if update, ok := parseManagerProgress(managerName, normalized); ok {
		return update, true
	}

	if update, ok := parseCommonProgress(normalized); ok {
		return update, true
	}

	return ProgressUpdate{}, false
}

func normalizeProgressLine(line string) string {
	line = ansiEscapeRe.ReplaceAllString(line, "")
	line = strings.ReplaceAll(line, "\r", "")
	return strings.TrimSpace(line)
}

func parseManagerProgress(managerName, line string) (ProgressUpdate, bool) {
	switch strings.ToLower(managerName) {
	case "winget":
		if matches := wingetProgressRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Installing", matches[1], line), true
		}
		if matches := wingetDownloadRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Downloading", matches[1], line), true
		}
		if matches := wingetFoundRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Found", matches[1], line), true
		}
		if matches := wingetSuccessRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Installed", matches[1], line), true
		}
	case "scoop":
		if matches := scoopUpdateRe.FindStringSubmatch(line); len(matches) == 4 {
			return progressFromLine("Updating", matches[1], matches[1]+" "+matches[2]+" -> "+matches[3]), true
		}
	case "chocolatey":
		if strings.Contains(strings.ToLower(line), "upgrading the following packages") {
			return ProgressUpdate{Status: "Preparing package upgrades"}, true
		}
		if matches := chocoPackageOpRe.FindStringSubmatch(line); len(matches) == 3 {
			status := strings.ToLower(matches[1])
			if len(status) > 0 {
				status = strings.ToUpper(status[:1]) + status[1:]
			}
			return progressFromLine(status, matches[2], line), true
		}
	case "npm", "pnpm", "yarn", "bun":
		if matches := npmVersionChangeRe.FindStringSubmatch(line); len(matches) == 4 {
			return progressFromLine("Updating", matches[1], matches[1]+" "+matches[2]+" -> "+matches[3]), true
		}
	case "pip":
		if strings.HasPrefix(strings.ToLower(line), "successfully installed ") {
			return ProgressUpdate{Status: "Installed packages", PackageName: strings.TrimSpace(line[len("Successfully installed "):])}, true
		}
		if matches := pipCollectingRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Collecting", matches[1], line), true
		}
		if matches := pipInstallingRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Installing", matches[1], line), true
		}
	case "pipx":
		if strings.Contains(strings.ToLower(line), "upgrading") {
			return progressFromLine("Upgrading", extractSubjectAfterKeyword(line, "upgrading"), line), true
		}
	case "poetry":
		if matches := poetryPackageOpRe.FindStringSubmatch(line); len(matches) == 4 {
			status := matches[1]
			if len(status) > 0 {
				status = strings.ToUpper(status[:1]) + status[1:]
			}
			return progressFromLine(status, matches[2], line), true
		}
	case "conda":
		if aptUpgradeListRe.MatchString(line) {
			return ProgressUpdate{Status: "Selecting packages for update"}, true
		}
	case "cargo":
		if strings.HasPrefix(strings.ToLower(line), "installing ") {
			return progressFromLine("Installing", extractSubjectAfterKeyword(line, "installing"), line), true
		}
		if strings.HasPrefix(strings.ToLower(line), "updating ") {
			return progressFromLine("Updating", extractSubjectAfterKeyword(line, "updating"), line), true
		}
	case "rustup":
		if matches := componentInstallRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Installing component", matches[1], line), true
		}
	case "helm":
		if matches := repoUpdateSuccessRe.FindStringSubmatch(line); len(matches) == 2 {
			return progressFromLine("Updated repository", matches[1], line), true
		}
	case "wsl (apt)":
		if aptUpgradeListRe.MatchString(line) {
			return ProgressUpdate{Status: "Preparing apt upgrade"}, true
		}
	case "vscode extensions":
		if strings.Contains(strings.ToLower(line), "updating extension") {
			return progressFromLine("Updating extension", extractSubjectAfterKeyword(line, "updating extension"), line), true
		}
	}

	return ProgressUpdate{}, false
}

func parseCommonProgress(line string) (ProgressUpdate, bool) {
	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "already up to date"), strings.Contains(lower, "no updates available"), strings.Contains(lower, "nothing to do"):
		return ProgressUpdate{Status: "Already up to date"}, true
	case strings.Contains(lower, "preparing transaction"):
		return ProgressUpdate{Status: "Preparing transaction"}, true
	case strings.Contains(lower, "verifying transaction"):
		return ProgressUpdate{Status: "Verifying transaction"}, true
	case strings.Contains(lower, "executing transaction"):
		return ProgressUpdate{Status: "Applying transaction"}, true
	case strings.Contains(lower, "building dependency tree"):
		return ProgressUpdate{Status: "Building dependency tree"}, true
	case strings.Contains(lower, "reading package lists"):
		return ProgressUpdate{Status: "Reading package lists"}, true
	case strings.Contains(lower, "resolving"):
		return ProgressUpdate{Status: "Resolving dependencies"}, true
	case strings.Contains(lower, "fetching"):
		return progressFromLine("Fetching", extractSubjectAfterKeyword(line, "fetching"), line), true
	case strings.Contains(lower, "downloading"):
		return progressFromLine("Downloading", extractSubjectAfterKeyword(line, "downloading"), line), true
	case strings.Contains(lower, "installing"):
		return progressFromLine("Installing", extractSubjectAfterKeyword(line, "installing"), line), true
	case strings.Contains(lower, "updating"):
		if matches := versionTransitionRe.FindStringSubmatch(line); len(matches) == 4 {
			return progressFromLine("Updating", matches[1], matches[2]+" -> "+matches[3]), true
		}
		return progressFromLine("Updating", extractSubjectAfterKeyword(line, "updating"), line), true
	case strings.Contains(lower, "linking"):
		return progressFromLine("Linking", extractSubjectAfterKeyword(line, "linking"), line), true
	case strings.Contains(lower, "building"):
		return progressFromLine("Building", extractSubjectAfterKeyword(line, "building"), line), true
	case strings.Contains(lower, "checking for updates"):
		return ProgressUpdate{Status: "Checking for updates"}, true
	case aptSummaryRe.MatchString(line):
		matches := aptSummaryRe.FindStringSubmatch(line)
		return ProgressUpdate{Status: matches[1] + " upgraded, " + matches[2] + " installed, " + matches[3] + " removed"}, true
	case packageCountRe.MatchString(line):
		return ProgressUpdate{Status: line}, true
	}

	if matches := versionTransitionRe.FindStringSubmatch(line); len(matches) == 4 {
		return progressFromLine("Updating", matches[1], matches[2]+" -> "+matches[3]), true
	}

	if pct := extractPercentage(line); pct >= 0 {
		if strings.Contains(lower, "download") {
			return ProgressUpdate{Status: "Downloading", Percentage: pct}, true
		}
		if strings.Contains(lower, "install") {
			return ProgressUpdate{Status: "Installing", Percentage: pct}, true
		}
	}

	return ProgressUpdate{}, false
}

func progressFromLine(status, packageName, detail string) ProgressUpdate {
	update := ProgressUpdate{Status: strings.TrimSpace(status)}
	packageName = strings.TrimSpace(packageName)
	if packageName != "" {
		update.PackageName = packageName
	}
	if detail != "" && update.Status == "" {
		update.Status = strings.TrimSpace(detail)
	}
	if pct := extractPercentage(detail); pct >= 0 {
		update.Percentage = pct
	}
	return update
}

func extractSubjectAfterKeyword(line, keyword string) string {
	lowerLine := strings.ToLower(line)
	keywordLower := strings.ToLower(keyword)
	idx := strings.Index(lowerLine, keywordLower)
	if idx < 0 {
		return ""
	}

	subject := strings.TrimSpace(line[idx+len(keyword):])
	subject = strings.TrimPrefix(subject, ":")
	return strings.TrimSpace(subject)
}

func extractPercentage(line string) int {
	matches := percentageRe.FindStringSubmatch(line)
	if len(matches) != 2 {
		return -1
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1
	}

	if value < 0 || value > 100 {
		return -1
	}

	return value
}
