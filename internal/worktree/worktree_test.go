package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/andrefelizardo/worktree-cli/internal/config"
)

func TestSanitizeDirName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"feat/foo", "feat-foo"},
		{"feat/foo/bar", "feat-foo-bar"},
		{"fix\\backslash", "fix-backslash"},
		{"space branch", "space-branch"},
		{"simple", "simple"},
		{"chore/test-branch", "chore-test-branch"},
	}

	for _, tt := range tests {
		got := SanitizeDirName(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeDirName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveWorktreeDir(t *testing.T) {
	cfg := &config.RepoConfig{
		WorktreeDir: "/worktrees/{repo}/{name}",
	}

	got := ResolveWorktreeDir(cfg, "my-repo", "feat/new-thing")
	want := "/worktrees/my-repo/feat-new-thing"
	if got != want {
		t.Errorf("ResolveWorktreeDir() = %q, want %q", got, want)
	}
}

func TestCreate_DryRun(t *testing.T) {
	cfg := &config.RepoConfig{
		RepoPath:    "/nonexistent/path",
		BaseBranch:  "main",
		WorktreeDir: "/tmp/worktrees/{repo}/{name}",
		Setup:       []string{"echo hello"},
	}

	opts := CreateOptions{
		RepoName:   "test",
		BranchName: "feat/dry",
		Config:     cfg,
		DryRun:     true,
	}

	dir, err := Create(opts)
	if err != nil {
		t.Fatalf("dry run should not error, got: %v", err)
	}
	want := "/tmp/worktrees/test/feat-dry"
	if dir != want {
		t.Errorf("Create() dry run dir = %q, want %q", dir, want)
	}
}

func TestCreate_RepoNotFound(t *testing.T) {
	cfg := &config.RepoConfig{
		RepoPath:    "/nonexistent/path",
		BaseBranch:  "main",
		WorktreeDir: "/tmp/worktrees/{repo}/{name}",
	}

	opts := CreateOptions{
		RepoName:   "test",
		BranchName: "feat/test",
		Config:     cfg,
	}

	_, err := Create(opts)
	if err == nil {
		t.Fatal("expected error for missing repo path, got nil")
	}
}

// --- Integration tests ---
// These create real git repos and worktrees, then clean up after.

func initBareAndClone(t *testing.T) (bareDir, cloneDir string) {
	t.Helper()
	tmpDir := t.TempDir()

	bareDir = filepath.Join(tmpDir, "bare.git")
	cloneDir = filepath.Join(tmpDir, "clone")

	// Create a bare repo to act as "origin"
	run(t, tmpDir, "git", "init", "--bare", bareDir)

	// Clone it
	run(t, tmpDir, "git", "clone", bareDir, cloneDir)

	// Create an initial commit on main so origin/main exists
	run(t, cloneDir, "git", "config", "user.email", "test@test.com")
	run(t, cloneDir, "git", "config", "user.name", "Test")
	dummyFile := filepath.Join(cloneDir, "README.md")
	os.WriteFile(dummyFile, []byte("# test"), 0o644)
	run(t, cloneDir, "git", "add", ".")
	run(t, cloneDir, "git", "commit", "-m", "initial commit")
	run(t, cloneDir, "git", "push", "origin", "main")

	return bareDir, cloneDir
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestIntegration_CreateWorktree(t *testing.T) {
	_, cloneDir := initBareAndClone(t)
	wtDir := filepath.Join(t.TempDir(), "wt-test")

	cfg := &config.RepoConfig{
		RepoPath:    cloneDir,
		BaseBranch:  "main",
		WorktreeDir: wtDir,
	}

	opts := CreateOptions{
		RepoName:   "test-repo",
		BranchName: "feat/integration-test",
		Config:     cfg,
	}

	dir, err := Create(opts)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if dir != wtDir {
		t.Errorf("Create() returned dir = %q, want %q", dir, wtDir)
	}

	// Verify the worktree directory exists
	if _, err := os.Stat(wtDir); os.IsNotExist(err) {
		t.Fatalf("worktree directory was not created at %s", wtDir)
	}

	// Verify we're on the correct branch in the worktree
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = wtDir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git branch --show-current failed: %v", err)
	}
	branch := string(out[:len(out)-1]) // trim newline
	if branch != "feat/integration-test" {
		t.Errorf("worktree branch = %q, want %q", branch, "feat/integration-test")
	}

	// Verify the README from the initial commit exists
	if _, err := os.Stat(filepath.Join(wtDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md not found in worktree — files not checked out")
	}

	// Cleanup: remove worktree
	run(t, cloneDir, "git", "worktree", "remove", wtDir)
}

func TestIntegration_CreateWorktree_BranchAlreadyExists(t *testing.T) {
	_, cloneDir := initBareAndClone(t)

	// Pre-create the branch so it already exists
	run(t, cloneDir, "git", "branch", "feat/existing-branch", "origin/main")

	wtDir := filepath.Join(t.TempDir(), "wt-existing")

	cfg := &config.RepoConfig{
		RepoPath:    cloneDir,
		BaseBranch:  "main",
		WorktreeDir: wtDir,
	}

	opts := CreateOptions{
		RepoName:   "test-repo",
		BranchName: "feat/existing-branch",
		Config:     cfg,
	}

	dir, err := Create(opts)
	if err != nil {
		t.Fatalf("Create() with existing branch failed: %v", err)
	}

	if dir != wtDir {
		t.Errorf("Create() returned dir = %q, want %q", dir, wtDir)
	}

	// Verify the worktree was created
	if _, err := os.Stat(wtDir); os.IsNotExist(err) {
		t.Fatalf("worktree directory was not created at %s", wtDir)
	}

	// Verify branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = wtDir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git branch --show-current failed: %v", err)
	}
	branch := string(out[:len(out)-1])
	if branch != "feat/existing-branch" {
		t.Errorf("worktree branch = %q, want %q", branch, "feat/existing-branch")
	}

	// Cleanup
	run(t, cloneDir, "git", "worktree", "remove", wtDir)
}

func TestIntegration_CreateWorktree_RecreateAfterRemoval(t *testing.T) {
	_, cloneDir := initBareAndClone(t)
	tmpDir := t.TempDir()
	wtDir1 := filepath.Join(tmpDir, "wt-first")
	wtDir2 := filepath.Join(tmpDir, "wt-second")

	cfg := &config.RepoConfig{
		RepoPath:   cloneDir,
		BaseBranch: "main",
	}

	// Create first worktree
	cfg.WorktreeDir = wtDir1
	opts := CreateOptions{
		RepoName:   "test-repo",
		BranchName: "feat/recreate-test",
		Config:     cfg,
	}
	_, err := Create(opts)
	if err != nil {
		t.Fatalf("first Create() failed: %v", err)
	}

	// Remove the worktree but leave the branch
	run(t, cloneDir, "git", "worktree", "remove", wtDir1)

	// Verify branch still exists
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/feat/recreate-test")
	cmd.Dir = cloneDir
	if cmd.Run() != nil {
		t.Fatal("branch should still exist after worktree removal")
	}

	// Create a second worktree with the same branch name
	cfg.WorktreeDir = wtDir2
	opts2 := CreateOptions{
		RepoName:   "test-repo",
		BranchName: "feat/recreate-test",
		Config:     cfg,
	}
	dir, err := Create(opts2)
	if err != nil {
		t.Fatalf("second Create() with existing branch failed: %v", err)
	}

	if dir != wtDir2 {
		t.Errorf("Create() returned dir = %q, want %q", dir, wtDir2)
	}

	if _, err := os.Stat(wtDir2); os.IsNotExist(err) {
		t.Fatalf("second worktree was not created at %s", wtDir2)
	}

	// Cleanup
	run(t, cloneDir, "git", "worktree", "remove", wtDir2)
}

func TestIntegration_ListWorktrees(t *testing.T) {
	_, cloneDir := initBareAndClone(t)
	tmpDir := t.TempDir()

	// Create two worktrees
	cfg := &config.RepoConfig{
		RepoPath:   cloneDir,
		BaseBranch: "main",
	}

	cfg.WorktreeDir = filepath.Join(tmpDir, "wt-alpha")
	Create(CreateOptions{RepoName: "test", BranchName: "feat/alpha", Config: cfg})

	cfg.WorktreeDir = filepath.Join(tmpDir, "wt-beta")
	Create(CreateOptions{RepoName: "test", BranchName: "feat/beta", Config: cfg})

	entries, err := List(cloneDir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("List() returned %d entries, want 2", len(entries))
	}

	branches := map[string]bool{}
	for _, e := range entries {
		branches[e.Branch] = true
		if e.Path == "" {
			t.Error("entry has empty path")
		}
	}
	if !branches["feat/alpha"] {
		t.Error("missing feat/alpha in list")
	}
	if !branches["feat/beta"] {
		t.Error("missing feat/beta in list")
	}

	// Cleanup
	run(t, cloneDir, "git", "worktree", "remove", filepath.Join(tmpDir, "wt-alpha"))
	run(t, cloneDir, "git", "worktree", "remove", filepath.Join(tmpDir, "wt-beta"))
}

func TestIntegration_ListWorktrees_Empty(t *testing.T) {
	_, cloneDir := initBareAndClone(t)

	entries, err := List(cloneDir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("List() returned %d entries, want 0", len(entries))
	}
}

func TestIntegration_Remove(t *testing.T) {
	_, cloneDir := initBareAndClone(t)
	wtDir := filepath.Join(t.TempDir(), "wt-remove")

	cfg := &config.RepoConfig{
		RepoPath:    cloneDir,
		BaseBranch:  "main",
		WorktreeDir: wtDir,
	}

	Create(CreateOptions{RepoName: "test", BranchName: "feat/to-remove", Config: cfg})

	// Verify worktree exists
	if _, err := os.Stat(wtDir); os.IsNotExist(err) {
		t.Fatal("worktree not created")
	}

	err := Remove(cloneDir, wtDir, false)
	if err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	// Verify worktree directory is gone
	if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
		t.Error("worktree directory still exists after Remove()")
	}

	// Branch should still exist after Remove (no branch deletion)
	branchCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/feat/to-remove")
	branchCmd.Dir = cloneDir
	if branchCmd.Run() != nil {
		t.Error("branch should still exist after Remove()")
	}
}

func TestIntegration_RemoveWithBranch(t *testing.T) {
	_, cloneDir := initBareAndClone(t)
	wtDir := filepath.Join(t.TempDir(), "wt-removebranch")

	cfg := &config.RepoConfig{
		RepoPath:    cloneDir,
		BaseBranch:  "main",
		WorktreeDir: wtDir,
	}

	Create(CreateOptions{RepoName: "test", BranchName: "feat/remove-branch", Config: cfg})

	err := RemoveWithBranch(cloneDir, wtDir, "feat/remove-branch", false)
	if err != nil {
		t.Fatalf("RemoveWithBranch() error: %v", err)
	}

	// Verify worktree directory is gone
	if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
		t.Error("worktree directory still exists after RemoveWithBranch()")
	}

	// Branch should be gone too
	branchCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/feat/remove-branch")
	branchCmd.Dir = cloneDir
	if branchCmd.Run() == nil {
		t.Error("branch should be deleted after RemoveWithBranch()")
	}
}
