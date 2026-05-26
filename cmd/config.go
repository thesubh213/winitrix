package cmd

import (
	"errors"
	"fmt"
	"os"

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

func init() {
	configInitCmd.Flags().StringVar(&configInitPath, "path", "", "Destination path for the config file")
	configInitCmd.Flags().BoolVar(&configInitForce, "force", false, "Overwrite an existing config file")
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
