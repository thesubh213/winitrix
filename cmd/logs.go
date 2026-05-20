package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/paths"
)

var (
	logsBundle     bool
	logsBundlePath string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Open Winitrix logs in the default text editor",
	Run: func(cmd *cobra.Command, args []string) {
		if logsBundle {
			bundlePath, err := createLogBundle(logsBundlePath)
			if err != nil {
				fmt.Printf("Failed to create bundle: %v\n", err)
				return
			}
			fmt.Printf("Bundle created: %s\n", bundlePath)
			return
		}

		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to determine executable location: %v\n", err)
			return
		}

		logFile := filepath.Join(filepath.Dir(exePath), "logs", "winitrix.log")

		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			fmt.Println("No logs found. Run an update first.")
			return
		}

		fmt.Println("Opening logs...")
		// Use notepad as fallback or default file association
		execCmd := exec.Command("cmd", "/c", "start", "", logFile)
		err = execCmd.Start()
		if err != nil {
			fmt.Printf("Failed to open logs: %v\n", err)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVar(&logsBundle, "bundle", false, "Create a diagnostic bundle with logs and config")
	logsCmd.Flags().StringVar(&logsBundlePath, "output", "", "Path to write the bundle zip")
	rootCmd.AddCommand(logsCmd)
}

func createLogBundle(outputPath string) (string, error) {
	if outputPath == "" {
		cacheDir, err := paths.CacheDir()
		if err != nil {
			return "", err
		}
		if err := paths.EnsureDir(cacheDir); err != nil {
			return "", err
		}
		timestamp := time.Now().Format("20060102_150405")
		outputPath = filepath.Join(cacheDir, "winitrix_bundle_"+timestamp+".zip")
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	logFile, _ := resolveLogPath()
	configPath, _ := paths.DefaultConfigPath()
	statePath, _ := paths.DefaultStatePath()
	reportPath, _ := paths.DefaultReportPath()

	_ = addFileToZip(zipWriter, logFile, "logs/winitrix.log")
	_ = addFileToZip(zipWriter, configPath, "config/config.toml")
	_ = addFileToZip(zipWriter, statePath, "state/state.json")
	_ = addFileToZip(zipWriter, reportPath, "state/last_report.json")
	_ = addStringToZip(zipWriter, "metadata.txt", fmt.Sprintf("bundle_created=%s\n", time.Now().Format(time.RFC3339)))

	return outputPath, nil
}

func resolveLogPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "logs", "winitrix.log"), nil
}

func addFileToZip(zipWriter *zip.Writer, path string, name string) error {
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	entry, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, file)
	return err
}

func addStringToZip(zipWriter *zip.Writer, name string, content string) error {
	entry, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = entry.Write([]byte(content))
	return err
}
