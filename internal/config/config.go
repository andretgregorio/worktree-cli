package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type GlobalConfig struct {
	WorktreeDir string `toml:"worktree_dir"`
	BaseBranch  string `toml:"base_branch"`
}

type RepoConfig struct {
	RepoPath    string            `toml:"repo_path"`
	BaseBranch  string            `toml:"base_branch"`
	WorktreeDir string            `toml:"worktree_dir"`
	Setup       []string          `toml:"setup"`
	Env         map[string]string `toml:"env"`
}

func wtConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "wt"), nil
}

func loadGlobalConfig() (*GlobalConfig, error) {
	dir, err := wtConfigDir()
	if err != nil {
		return nil, err
	}
	return loadGlobalConfigFrom(filepath.Join(dir, "config.toml"))
}

func loadGlobalConfigFrom(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &GlobalConfig{}, nil
		}
		return nil, fmt.Errorf("could not read global config: %w", err)
	}

	cfg := &GlobalConfig{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse global config: %w", err)
	}
	return cfg, nil
}

func LoadRepoConfig(name string) (*RepoConfig, error) {
	dir, err := wtConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "repos", name+".toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config for repo %q: %w", name, err)
	}

	cfg := &RepoConfig{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config for repo %q: %w", name, err)
	}

	global, err := loadGlobalConfig()
	if err != nil {
		return nil, err
	}

	applyDefaults(cfg, global, name)
	return cfg, nil
}

func expandTilde(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func applyDefaults(cfg *RepoConfig, global *GlobalConfig, repoName string) {
	home, _ := os.UserHomeDir()
	defaultWorktreeDir := filepath.Join(home, "Workspace", "worktrees", "{repo}", "{name}")

	if cfg.BaseBranch == "" {
		if global.BaseBranch != "" {
			cfg.BaseBranch = global.BaseBranch
		} else {
			cfg.BaseBranch = "main"
		}
	}
	if cfg.RepoPath == "" {
		cfg.RepoPath = filepath.Join(home, "Workspace", repoName)
	} else {
		cfg.RepoPath = expandTilde(cfg.RepoPath, home)
	}
	if cfg.WorktreeDir == "" {
		if global.WorktreeDir != "" {
			cfg.WorktreeDir = expandTilde(global.WorktreeDir, home)
		} else {
			cfg.WorktreeDir = defaultWorktreeDir
		}
	} else {
		cfg.WorktreeDir = expandTilde(cfg.WorktreeDir, home)
	}
}
