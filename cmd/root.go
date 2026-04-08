package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "evoloop",
	Short: "A CLI tool that runs self-improvement loops on local Git projects",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
