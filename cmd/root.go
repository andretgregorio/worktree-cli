package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "Git worktree manager",
	Long:  "A CLI tool to create and manage git worktrees with per-repo setup automation.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Hide auto-generated commands
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}
