//go:build windows

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/tui"
	"golang.org/x/sys/windows/registry"
)

var (
	installPath  string
	installForce bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Winitrix and add it to PATH",
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to determine executable path: %v\n", err)
			return
		}

		exePath, err = filepath.Abs(exePath)
		if err != nil {
			fmt.Printf("Failed to resolve executable path: %v\n", err)
			return
		}

		destDir, destPath, err := resolveInstallPath(installPath)
		if err != nil {
			fmt.Printf("Failed to resolve install path: %v\n", err)
			return
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			fmt.Printf("Failed to create install directory: %v\n", err)
			return
		}

		if !strings.EqualFold(exePath, destPath) {
			if err := copyExecutable(exePath, destPath, installForce); err != nil {
				fmt.Printf("Failed to install Winitrix: %v\n", err)
				return
			}
		}

		updated, err := ensurePathContains(destDir)
		if err != nil {
			fmt.Printf("Failed to update PATH: %v\n", err)
			return
		}

		if updated {
			broadcastEnvironmentChange()
		}

		fmt.Println()
		fmt.Println(tui.MiniLogo() + tui.SubtleStyle.Render("  installed"))
		fmt.Println(tui.Divider())
		fmt.Println()
		fmt.Printf("  %s %s\n", tui.Checkmark(), tui.BoldStyle.Render("Winitrix installed"))
		fmt.Printf("  %s %s\n", tui.Checkmark(), tui.SubtleStyle.Render("Path: "+destPath))
		if updated {
			fmt.Printf("  %s %s\n", tui.Checkmark(), tui.SubtleStyle.Render("PATH updated (new terminals will pick this up)"))
		} else {
			fmt.Printf("  %s %s\n", tui.Checkmark(), tui.SubtleStyle.Render("PATH already contains install directory"))
		}
		fmt.Println()
		fmt.Println(tui.SubtleStyle.Render("Try: winitrix --help"))
	},
}

func init() {
	installCmd.Flags().StringVar(&installPath, "path", "", "Install directory (defaults to %LOCALAPPDATA%\\Winitrix\\bin)")
	installCmd.Flags().BoolVar(&installForce, "force", false, "Overwrite existing executable")
	rootCmd.AddCommand(installCmd)
}

func tryAutoInstall() error {
	if !shouldAutoInstall() {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return err
	}

	destDir, destPath, err := resolveInstallPath("")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	if !strings.EqualFold(exePath, destPath) {
		if _, err := os.Stat(destPath); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err := copyExecutable(exePath, destPath, false); err != nil {
				return err
			}
		}
	}

	updated, err := ensurePathContains(destDir)
	if err != nil {
		return err
	}

	if updated {
		broadcastEnvironmentChange()
	}

	return nil
}

func shouldAutoInstall() bool {
	if strings.EqualFold(os.Getenv("CI"), "true") {
		return false
	}
	if strings.EqualFold(os.Getenv("GITHUB_ACTIONS"), "true") {
		return false
	}
	return true
}

func resolveInstallPath(input string) (string, string, error) {
	if input == "" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			userProfile := os.Getenv("USERPROFILE")
			if userProfile == "" {
				return "", "", fmt.Errorf("LOCALAPPDATA and USERPROFILE are not set")
			}
			localAppData = filepath.Join(userProfile, "AppData", "Local")
		}
		input = filepath.Join(localAppData, "Winitrix", "bin")
	}

	destDir := input
	destPath := filepath.Join(destDir, "winitrix.exe")
	if strings.HasSuffix(strings.ToLower(input), ".exe") {
		destPath = input
		destDir = filepath.Dir(input)
	}

	return destDir, destPath, nil
}

func copyExecutable(src, dst string, force bool) error {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination already exists (use --force to overwrite)")
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

func ensurePathContains(dir string) (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	existing, _, _ := key.GetStringValue("Path")
	if pathContains(existing, dir) {
		return false, nil
	}

	newPath := existing
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += dir

	if err := key.SetStringValue("Path", newPath); err != nil {
		return false, err
	}

	return true, nil
}

func pathContains(pathList, dir string) bool {
	normalizedDir := strings.ToLower(filepath.Clean(dir))
	for _, entry := range strings.Split(pathList, ";") {
		if strings.ToLower(filepath.Clean(entry)) == normalizedDir {
			return true
		}
	}
	return false
}

func broadcastEnvironmentChange() {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SendMessageTimeoutW")

	const (
		hwndBroadcast   = 0xffff
		wmSettingChange = 0x1a
		smtoAbortIfHung = 0x0002
	)

	envPtr, _ := syscall.UTF16PtrFromString("Environment")
	var result uintptr
	_, _, _ = proc.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(envPtr)),
		smtoAbortIfHung,
		5000,
		uintptr(unsafe.Pointer(&result)),
	)
}
