package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestExpandTilde(t *testing.T) {
	home := "/Users/testuser"

	tests := []struct {
		input string
		want  string
	}{
		{"~/Workspace/repo", "/Users/testuser/Workspace/repo"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~/", "/Users/testuser"},
		{"~nope/foo", "~nope/foo"},
	}

	for _, tt := range tests {
		got := expandTilde(tt.input, home)
		if got != tt.want {
			t.Errorf("expandTilde(%q, %q) = %q, want %q", tt.input, home, got, tt.want)
		}
	}
}

func TestApplyDefaults(t *testing.T) {
	home, _ := os.UserHomeDir()
	emptyGlobal := &GlobalConfig{}

	t.Run("all defaults", func(t *testing.T) {
		cfg := &RepoConfig{}
		applyDefaults(cfg, emptyGlobal, "my-repo")

		if cfg.BaseBranch != "main" {
			t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "main")
		}
		wantRepo := filepath.Join(home, "Workspace", "my-repo")
		if cfg.RepoPath != wantRepo {
			t.Errorf("RepoPath = %q, want %q", cfg.RepoPath, wantRepo)
		}
		wantWt := filepath.Join(home, "Workspace", "worktrees", "{repo}", "{name}")
		if cfg.WorktreeDir != wantWt {
			t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, wantWt)
		}
	})

	t.Run("preserves explicit values", func(t *testing.T) {
		cfg := &RepoConfig{
			BaseBranch:  "develop",
			RepoPath:    "/custom/path",
			WorktreeDir: "/custom/worktrees/{name}",
		}
		applyDefaults(cfg, emptyGlobal, "my-repo")

		if cfg.BaseBranch != "develop" {
			t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "develop")
		}
		if cfg.RepoPath != "/custom/path" {
			t.Errorf("RepoPath = %q, want %q", cfg.RepoPath, "/custom/path")
		}
		if cfg.WorktreeDir != "/custom/worktrees/{name}" {
			t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, "/custom/worktrees/{name}")
		}
	})

	t.Run("expands tilde in repo_path", func(t *testing.T) {
		cfg := &RepoConfig{RepoPath: "~/Workspace/groups"}
		applyDefaults(cfg, emptyGlobal, "groups")

		want := filepath.Join(home, "Workspace", "groups")
		if cfg.RepoPath != want {
			t.Errorf("RepoPath = %q, want %q", cfg.RepoPath, want)
		}
	})

	t.Run("expands tilde in worktree_dir", func(t *testing.T) {
		cfg := &RepoConfig{WorktreeDir: "~/worktrees/{repo}/{name}"}
		applyDefaults(cfg, emptyGlobal, "test")

		want := filepath.Join(home, "worktrees", "{repo}", "{name}")
		if cfg.WorktreeDir != want {
			t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, want)
		}
	})
}

func TestApplyDefaults_GlobalConfig(t *testing.T) {
	home, _ := os.UserHomeDir()

	t.Run("global worktree_dir used when repo has none", func(t *testing.T) {
		global := &GlobalConfig{WorktreeDir: "~/custom-worktrees/{repo}/{name}"}
		cfg := &RepoConfig{}
		applyDefaults(cfg, global, "my-repo")

		want := filepath.Join(home, "custom-worktrees", "{repo}", "{name}")
		if cfg.WorktreeDir != want {
			t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, want)
		}
	})

	t.Run("global base_branch used when repo has none", func(t *testing.T) {
		global := &GlobalConfig{BaseBranch: "develop"}
		cfg := &RepoConfig{}
		applyDefaults(cfg, global, "my-repo")

		if cfg.BaseBranch != "develop" {
			t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "develop")
		}
	})

	t.Run("repo config overrides global worktree_dir", func(t *testing.T) {
		global := &GlobalConfig{WorktreeDir: "~/global-worktrees/{repo}/{name}"}
		cfg := &RepoConfig{WorktreeDir: "/repo-specific/{name}"}
		applyDefaults(cfg, global, "my-repo")

		if cfg.WorktreeDir != "/repo-specific/{name}" {
			t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, "/repo-specific/{name}")
		}
	})

	t.Run("repo config overrides global base_branch", func(t *testing.T) {
		global := &GlobalConfig{BaseBranch: "develop"}
		cfg := &RepoConfig{BaseBranch: "release"}
		applyDefaults(cfg, global, "my-repo")

		if cfg.BaseBranch != "release" {
			t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "release")
		}
	})
}

func TestLoadRepoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "repos")
	os.MkdirAll(configPath, 0o755)

	content := `
repo_path = "/tmp/my-repo"
base_branch = "develop"

setup = [
  "npm install",
  "cp .env.example .env",
]

[env]
NODE_ENV = "test"
`
	os.WriteFile(filepath.Join(configPath, "my-repo.toml"), []byte(content), 0o644)

	t.Run("parses valid config", func(t *testing.T) {
		data, _ := os.ReadFile(filepath.Join(configPath, "my-repo.toml"))
		cfg := &RepoConfig{}
		if err := tomlUnmarshal(data, cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		applyDefaults(cfg, &GlobalConfig{}, "my-repo")

		if cfg.RepoPath != "/tmp/my-repo" {
			t.Errorf("RepoPath = %q, want %q", cfg.RepoPath, "/tmp/my-repo")
		}
		if cfg.BaseBranch != "develop" {
			t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "develop")
		}
		if len(cfg.Setup) != 2 {
			t.Fatalf("Setup len = %d, want 2", len(cfg.Setup))
		}
		if cfg.Setup[0] != "npm install" {
			t.Errorf("Setup[0] = %q, want %q", cfg.Setup[0], "npm install")
		}
		if cfg.Setup[1] != "cp .env.example .env" {
			t.Errorf("Setup[1] = %q, want %q", cfg.Setup[1], "cp .env.example .env")
		}
		if cfg.Env["NODE_ENV"] != "test" {
			t.Errorf("Env[NODE_ENV] = %q, want %q", cfg.Env["NODE_ENV"], "test")
		}
	})

	t.Run("missing repo config returns error", func(t *testing.T) {
		_, err := LoadRepoConfig("nonexistent-repo")
		if err == nil {
			t.Fatal("expected error for missing config, got nil")
		}
	})
}

func TestLoadGlobalConfig(t *testing.T) {
	t.Run("returns empty config when file missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nonexistent", "config.toml")
		cfg, err := loadGlobalConfigFrom(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.WorktreeDir != "" {
			t.Errorf("WorktreeDir = %q, want empty", cfg.WorktreeDir)
		}
		if cfg.BaseBranch != "" {
			t.Errorf("BaseBranch = %q, want empty", cfg.BaseBranch)
		}
	})

	t.Run("parses existing global config", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		os.WriteFile(path, []byte(`
worktree_dir = "~/CustomWorktrees/{repo}/{name}"
base_branch = "develop"
`), 0o644)

		cfg, err := loadGlobalConfigFrom(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.WorktreeDir != "~/CustomWorktrees/{repo}/{name}" {
			t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, "~/CustomWorktrees/{repo}/{name}")
		}
		if cfg.BaseBranch != "develop" {
			t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "develop")
		}
	})
}

func tomlUnmarshal(data []byte, v interface{}) error {
	_, err := toml.Decode(string(data), v)
	return err
}
