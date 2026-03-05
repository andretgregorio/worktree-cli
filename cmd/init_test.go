package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfig_CreatesFilesWithDefault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	configPath := filepath.Join(dir, "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config.toml not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `worktree_dir = "~/Worktrees/{repo}/{name}"`) {
		t.Errorf("config.toml missing default worktree_dir, got:\n%s", content)
	}
	if !strings.Contains(content, `base_branch = "main"`) {
		t.Errorf("config.toml missing base_branch, got:\n%s", content)
	}

	reposDir := filepath.Join(dir, "repos")
	info, err := os.Stat(reposDir)
	if err != nil {
		t.Fatalf("repos/ directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("repos/ is not a directory")
	}
}

func TestInitConfig_CreatesFilesWithCustomDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	input := strings.NewReader("/custom/worktrees/{repo}/{name}\n")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "config.toml"))
	content := string(data)
	if !strings.Contains(content, `worktree_dir = "/custom/worktrees/{repo}/{name}"`) {
		t.Errorf("config.toml has wrong worktree_dir, got:\n%s", content)
	}
}

func TestInitConfig_CreatesFilesWithTildeDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	input := strings.NewReader("~/my-worktrees/{repo}/{name}\n")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "config.toml"))
	content := string(data)
	if !strings.Contains(content, `worktree_dir = "~/my-worktrees/{repo}/{name}"`) {
		t.Errorf("config.toml has wrong worktree_dir, got:\n%s", content)
	}
}

func TestInitConfig_SkipsExistingConfig(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	os.MkdirAll(dir, 0o755)

	configPath := filepath.Join(dir, "config.toml")
	existing := `worktree_dir = "/my/custom/path/{repo}/{name}"` + "\n"
	os.WriteFile(configPath, []byte(existing), 0o644)

	input := strings.NewReader("/should/not/be/used\n")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	if string(data) != existing {
		t.Errorf("existing config was overwritten, got:\n%s\nwant:\n%s", string(data), existing)
	}

	if !strings.Contains(output.String(), "already exists") {
		t.Errorf("expected skip message, got: %s", output.String())
	}
}

func TestInitConfig_CreatesNestedDirectories(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "nested", "wt")
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Errorf("config.toml not created in nested path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "repos")); err != nil {
		t.Errorf("repos/ not created in nested path: %v", err)
	}
}

func TestInitConfig_PromptsUser(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}

	initConfig(dir, input, output)

	if !strings.Contains(output.String(), "Where should worktrees be created?") {
		t.Errorf("expected prompt in output, got: %s", output.String())
	}
	if !strings.Contains(output.String(), defaultWorktreeDir) {
		t.Errorf("expected default dir in prompt, got: %s", output.String())
	}
}

func TestInitConfig_TrimsWhitespaceFromInput(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	input := strings.NewReader("  /trimmed/path/{repo}/{name}  \n")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "config.toml"))
	content := string(data)
	if !strings.Contains(content, `worktree_dir = "/trimmed/path/{repo}/{name}"`) {
		t.Errorf("whitespace not trimmed, got:\n%s", content)
	}
}

func TestInitConfig_EmptyInputUsesDefault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "wt")
	// Empty reader simulates user just pressing enter with no input
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	err := initConfig(dir, input, output)
	if err != nil {
		t.Fatalf("initConfig() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "config.toml"))
	content := string(data)
	if !strings.Contains(content, `worktree_dir = "~/Worktrees/{repo}/{name}"`) {
		t.Errorf("expected default worktree_dir on empty input, got:\n%s", content)
	}
}
