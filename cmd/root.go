package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "k3p",
		Short: "Package Manager for k3s",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(purgeCmd)
}

func initConfig() {
}
