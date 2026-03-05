package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/andrefelizardo/worktree-cli/internal/config"
	"github.com/andrefelizardo/worktree-cli/internal/worktree"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <repo> [worktree-name]",
	Short: "Remove a worktree for a repo",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	repoName := args[0]

	cfg, err := config.LoadRepoConfig(repoName)
	if err != nil {
		return err
	}

	if len(args) == 2 {
		return removeByName(cfg, repoName, args[1])
	}

	return removeInteractive(cfg, repoName, os.Stdin, os.Stderr)
}

func removeByName(cfg *config.RepoConfig, repoName, wtName string) error {
	entries, err := worktree.List(cfg.RepoPath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		dirName := dirBaseName(e.Path)
		if dirName == wtName || e.Branch == wtName {
			fmt.Fprintf(os.Stderr, "Removing worktree %s...\n", e.Path)
			return worktree.RemoveWithBranch(cfg.RepoPath, e.Path, e.Branch)
		}
	}

	return fmt.Errorf("worktree %q not found for repo %q", wtName, repoName)
}

func removeInteractive(cfg *config.RepoConfig, repoName string, input io.Reader, output io.Writer) error {
	entries, err := worktree.List(cfg.RepoPath)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintf(output, "No worktrees found for %q.\n", repoName)
		return nil
	}

	fmt.Fprintf(output, "Worktrees for %q:\n", repoName)
	for i, e := range entries {
		fmt.Fprintf(output, "  [%d] %s (%s)\n", i+1, dirBaseName(e.Path), e.Branch)
	}
	fmt.Fprintf(output, "\nSelect worktree to remove (1-%d): ", len(entries))

	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		return fmt.Errorf("no selection made")
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > len(entries) {
		return fmt.Errorf("invalid selection: %q", scanner.Text())
	}

	selected := entries[choice-1]
	fmt.Fprintf(output, "Removing worktree %s...\n", selected.Path)
	return worktree.RemoveWithBranch(cfg.RepoPath, selected.Path, selected.Branch)
}

func dirBaseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
