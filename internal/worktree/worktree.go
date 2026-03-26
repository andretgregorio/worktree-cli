package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/andrefelizardo/worktree-cli/internal/config"
)

type CreateOptions struct {
	RepoName   string
	BranchName string
	Config     *config.RepoConfig
	DryRun     bool
	NoSetup    bool
}

func SanitizeDirName(branch string) string {
	r := strings.NewReplacer("/", "-", "\\", "-", " ", "-")
	return r.Replace(branch)
}

func ResolveWorktreeDir(cfg *config.RepoConfig, repoName, branchName string) string {
	sanitized := SanitizeDirName(branchName)
	dir := cfg.WorktreeDir
	dir = strings.ReplaceAll(dir, "{repo}", repoName)
	dir = strings.ReplaceAll(dir, "{name}", sanitized)
	return dir
}

func Create(opts CreateOptions) (string, error) {
	cfg := opts.Config
	repoPath := cfg.RepoPath
	worktreeDir := ResolveWorktreeDir(cfg, opts.RepoName, opts.BranchName)

	if opts.DryRun {
		printDryRun(opts, worktreeDir)
		return worktreeDir, nil
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return "", fmt.Errorf("repo path does not exist: %s", repoPath)
	}

	fmt.Printf("Fetching origin in %s...\n", repoPath)
	if err := runGit(repoPath, "fetch", "origin"); err != nil {
		return "", fmt.Errorf("git fetch failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Creating worktree at %s...\n", worktreeDir)
	if branchExists(repoPath, opts.BranchName) {
		fmt.Fprintf(os.Stderr, "Branch %q already exists, resetting to origin/%s...\n", opts.BranchName, cfg.BaseBranch)
		if err := runGit(repoPath, "branch", "-f", opts.BranchName, "origin/"+cfg.BaseBranch); err != nil {
			return "", fmt.Errorf("git branch reset failed: %w", err)
		}
		if err := runGit(repoPath, "worktree", "add", worktreeDir, opts.BranchName); err != nil {
			return "", fmt.Errorf("git worktree add failed: %w", err)
		}
	} else {
		if err := runGit(repoPath, "worktree", "add", "-b", opts.BranchName, worktreeDir, "origin/"+cfg.BaseBranch); err != nil {
			return "", fmt.Errorf("git worktree add failed: %w", err)
		}
	}

	return worktreeDir, nil
}

type WorktreeEntry struct {
	Path   string
	Branch string
}

func List(repoPath string) ([]WorktreeEntry, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var entries []WorktreeEntry
	var current WorktreeEntry
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			current = WorktreeEntry{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if line == "" && current.Path != "" {
			entries = append(entries, current)
			current = WorktreeEntry{}
		}
	}

	// Filter out the main worktree (the repo itself)
	if len(entries) > 0 {
		entries = entries[1:]
	}
	return entries, nil
}

func Remove(repoPath, worktreePath string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	if err := runGit(repoPath, args...); err != nil {
		return fmt.Errorf("git worktree remove failed: %w", err)
	}
	return nil
}

func RemoveWithBranch(repoPath, worktreePath, branch string, force bool) error {
	if err := Remove(repoPath, worktreePath, force); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Deleting branch %q...\n", branch)
	if err := runGit(repoPath, "branch", "-D", branch); err != nil {
		return fmt.Errorf("git branch delete failed: %w", err)
	}
	return nil
}

func branchExists(repoPath, branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = repoPath
	return cmd.Run() == nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printDryRun(opts CreateOptions, worktreeDir string) {
	cfg := opts.Config
	fmt.Fprintln(os.Stderr, "[dry-run] Would execute the following:")
	fmt.Fprintf(os.Stderr, "  1. git fetch origin          (in %s)\n", cfg.RepoPath)
	fmt.Fprintf(os.Stderr, "  2. git worktree add -b %s %s origin/%s\n", opts.BranchName, worktreeDir, cfg.BaseBranch)
	if !opts.NoSetup && len(cfg.Setup) > 0 {
		fmt.Fprintln(os.Stderr, "  3. Run setup commands:")
		for i, s := range cfg.Setup {
			fmt.Fprintf(os.Stderr, "     %d. %s\n", i+1, s)
		}
	}
	if len(cfg.Env) > 0 {
		fmt.Fprintln(os.Stderr, "  4. Environment variables:")
		for k, v := range cfg.Env {
			fmt.Fprintf(os.Stderr, "     %s=%s\n", k, v)
		}
	}
	fmt.Fprintf(os.Stderr, "  cd → %s\n", worktreeDir)
}
