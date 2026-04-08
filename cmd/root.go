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

// ExecuteArgs runs the root command with the given arguments. Exported for testing.
func ExecuteArgs(args []string) error {
	rootCmd.SetArgs(args)
	defer rootCmd.SetArgs(nil)
	return rootCmd.Execute()
}
