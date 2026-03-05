package cmd

import (
	"fmt"
	"os"

	"github.com/andrefelizardo/worktree-cli/internal/config"
	"github.com/andrefelizardo/worktree-cli/internal/shell"
	"github.com/andrefelizardo/worktree-cli/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	flagBase    string
	flagDir     string
	flagNoSetup bool
	flagDryRun  bool
)

var createCmd = &cobra.Command{
	Use:   "create <repo> <branch-name>",
	Short: "Create a new worktree for a repo",
	Args:  cobra.ExactArgs(2),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&flagBase, "base", "", "Override base branch")
	createCmd.Flags().StringVar(&flagDir, "dir", "", "Override worktree directory")
	createCmd.Flags().BoolVar(&flagNoSetup, "no-setup", false, "Skip setup commands")
	createCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print actions without executing")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	repoName := args[0]
	branchName := args[1]

	cfg, err := config.LoadRepoConfig(repoName)
	if err != nil {
		return err
	}

	if flagBase != "" {
		cfg.BaseBranch = flagBase
	}
	if flagDir != "" {
		cfg.WorktreeDir = flagDir
	}

	opts := worktree.CreateOptions{
		RepoName:   repoName,
		BranchName: branchName,
		Config:     cfg,
		DryRun:     flagDryRun,
		NoSetup:    flagNoSetup,
	}

	wtDir, err := worktree.Create(opts)
	if err != nil {
		return err
	}

	if !flagDryRun && !flagNoSetup && len(cfg.Setup) > 0 {
		if err := shell.RunSetupCommands(wtDir, cfg.Setup); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	if len(cfg.Env) > 0 {
		shell.PrintEnvExports(cfg.Env)
	}

	fmt.Fprintf(os.Stderr, "\nWorktree ready: %s\n", wtDir)
	shell.PrintCdMarker(wtDir)
	return nil
}
