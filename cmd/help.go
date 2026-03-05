package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const helpText = `wt - Git Worktree Manager

Usage:
  wt <command> [flags]

Commands:
  create <repo> <branch>        Create a new worktree for a repo
  remove <repo> [worktree]      Remove a worktree (interactive if no name given)
  init                           Initialize global configuration
  help                           Show this help message

Create Flags:
  --base <branch>           Override the base branch from config
  --dir <path>              Override the worktree directory
  --no-setup                Skip setup commands
  --dry-run                 Print planned actions without executing

Examples:
  wt create groups feat/new-feature
  wt create groups feat/hotfix --base release/v2
  wt create groups feat/quick-fix --no-setup --dry-run
  wt remove groups feat-new-feature
  wt remove groups
  wt init
`

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help for wt commands",
	Run: func(cmd *cobra.Command, args []string) {
		printHelp(os.Stdout)
	},
}

func init() {
	// Replace Cobra's default help command
	rootCmd.SetHelpCommand(helpCmd)
	rootCmd.AddCommand(helpCmd)
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printHelp(cmd.OutOrStdout())
	})
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, helpText)
}
