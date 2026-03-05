package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const defaultWorktreeDir = "~/Worktrees/{repo}/{name}"

const globalConfigTemplate = `# Global wt configuration
# These defaults apply to all repos unless overridden in per-repo configs.

# Default directory for worktrees
# {repo} = repo name, {name} = sanitized branch name
worktree_dir = "%s"

# Default base branch
base_branch = "main"
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize wt global configuration",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}
	return initConfig(filepath.Join(home, ".config", "wt"), os.Stdin, os.Stderr)
}

func initConfig(configDir string, input io.Reader, output io.Writer) error {
	reposDir := filepath.Join(configDir, "repos")
	configPath := filepath.Join(configDir, "config.toml")

	if err := os.MkdirAll(reposDir, 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(output, "Global config already exists at %s, skipping.\n", configPath)
		return nil
	}

	worktreeDir := promptWorktreeDir(input, output)

	content := fmt.Sprintf(globalConfigTemplate, worktreeDir)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("could not write global config: %w", err)
	}

	fmt.Fprintf(output, "Created global config at %s\n", configPath)
	fmt.Fprintf(output, "Created repos config directory at %s\n", reposDir)
	return nil
}

func promptWorktreeDir(input io.Reader, output io.Writer) string {
	fmt.Fprintf(output, "Where should worktrees be created? [%s]: ", defaultWorktreeDir)
	scanner := bufio.NewScanner(input)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			return line
		}
	}
	return defaultWorktreeDir
}
