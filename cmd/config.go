package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/config"
	"github.com/thesubh213/winitrix/pkg/paths"
)

var (
	configInitPath  string
	configInitForce bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Winitrix configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a starter config file",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.WriteDefaultConfig(configInitPath, configInitForce)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				fmt.Println("Config already exists. Use --force to overwrite.")
				return
			}
			fmt.Printf("Failed to write config: %v\n", err)
			return
		}
		fmt.Printf("Config written to: %s\n", path)
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the default config path",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := paths.DefaultConfigPath()
		if err != nil {
			fmt.Printf("Failed to resolve config path: %v\n", err)
			return
		}
		fmt.Println(path)
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the config file in the default text editor",
	Run: func(cmd *cobra.Command, args []string) {
		settings := ActiveSettings()
		path := settings.ConfigPath
		if path == "" {
			var err error
			path, err = paths.DefaultConfigPath()
			if err != nil {
				fmt.Printf("Failed to resolve config path: %v\n", err)
				return
			}
		}

		// Ensure config exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println("Config file does not exist. Creating default starter config...")
			_, err = config.WriteDefaultConfig(path, false)
			if err != nil {
				fmt.Printf("Failed to write default config: %v\n", err)
				return
			}
		}

		// Resolve editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "notepad.exe"
		}

		fmt.Printf("Opening config file at: %s with '%s'...\n", path, editor)

		editorCmd := exec.Command(editor, path)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			fmt.Printf("Failed to launch editor: %v\n", err)
		}
	},
}

func init() {
	configInitCmd.Flags().StringVar(&configInitPath, "path", "", "Destination path for the config file")
	configInitCmd.Flags().BoolVar(&configInitForce, "force", false, "Overwrite an existing config file")
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	rootCmd.AddCommand(configCmd)
}
