# wt - Git Worktree Manager CLI

A CLI tool that automates git worktree creation with per-repo setup commands. Create a worktree, install dependencies, copy env files, and `cd` into it — all in one command.

## Install

### Prerequisites

- Go 1.21+
- Git

### Quick Install

```bash
git clone git@github.com:andrefelizardo/worktree-cli.git
cd worktree-cli
./install.sh
```

The install script will:

1. Build the `wt-bin` binary
2. Ask where you want worktrees created (defaults to `~/Worktrees/{repo}/{name}`)
3. Write the global config to `~/.config/wt/config.toml`
4. Offer to add the shell wrapper to your `~/.zshrc` or `~/.bashrc`

After install, restart your shell or run `source ~/.zshrc`.

### Manual Install

If you prefer to set things up yourself:

```bash
git clone git@github.com:andrefelizardo/worktree-cli.git
cd worktree-cli
go build -o wt-bin .
./wt-bin init
```

Then add to your `~/.zshrc` (or `~/.bashrc`):

```bash
source /path/to/worktree-cli/shell/wt.sh
```

The shell wrapper resolves the `wt-bin` binary relative to itself, so no PATH changes are needed.

## Configuration

### Global Config

Set defaults for all repos in `~/.config/wt/config.toml`:

```toml
# Default worktree output directory for all repos
worktree_dir = "~/Worktrees/{repo}/{name}"

# Default base branch for all repos
base_branch = "main"
```

### Per-Repo Config

Per-repo configs live at `~/.config/wt/repos/<repo-name>.toml` and override global settings:

```toml
# Path to the repo (supports ~/, defaults to ~/Workspace/<repo-name>)
repo_path = "~/Workspace/groups"

# Branch to create worktrees from (overrides global, default: "main")
base_branch = "main"

# Where to place worktrees (overrides global)
# {repo} = repo name, {name} = sanitized branch name
worktree_dir = "~/Workspace/worktrees/{repo}/{name}"

# Setup commands run in order after worktree creation
setup = [
  "nvm use",
  "pnpm install",
  "cp .env.example .env",
]

# Environment variables exported in the shell session
[env]
NODE_ENV = "development"
```

### Precedence

All fields are optional. Values are resolved in this order:

1. **Per-repo config** (highest priority)
2. **Global config**
3. **Built-in defaults** (lowest priority)

| Field | Built-in Default |
|---|---|
| `repo_path` | `~/Workspace/<repo-name>` |
| `base_branch` | `main` |
| `worktree_dir` | `~/Workspace/worktrees/{repo}/{name}` |

## Usage

```
wt <command> [flags]

Commands:
  create <repo> <branch>        Create a new worktree for a repo
  remove <repo> [worktree]      Remove a worktree (interactive if no name given)
  init                           Initialize global configuration
  help                           Show help message
```

### `wt create <repo> <branch-name>`

This will:

1. Load `~/.config/wt/repos/<repo>.toml`
2. `git fetch origin` in the repo
3. Create a worktree with a new branch based on `origin/<base-branch>`
4. Run setup commands in the worktree directory
5. `cd` into the worktree

### Examples

```bash
# Create a worktree for the "groups" repo
wt create groups feat/new-feature

# Override the base branch
wt create groups feat/hotfix --base release/v2

# Preview what would happen without executing
wt create groups feat/new-feature --dry-run

# Skip setup commands
wt create groups feat/quick-fix --no-setup
```

Branch names are sanitized for directory paths: `feat/foo` becomes `feat-foo`.

If a branch with the same name already exists (e.g. from a previously removed worktree), `wt` resets it to the base branch and reuses it.

### Flags

| Flag | Description |
|---|---|
| `--base <branch>` | Override the base branch from config |
| `--dir <path>` | Override the worktree directory |
| `--no-setup` | Skip setup commands |
| `--dry-run` | Print planned actions without executing |

### `wt remove <repo> [worktree-name]`

Removes a worktree and deletes its branch.

- With a name: `wt remove groups feat-new-feature` — removes that worktree directly. Matches by directory name or branch name.
- Without a name: `wt remove groups` — lists all worktrees for the repo and prompts you to select one.

```bash
# Remove by name
wt remove groups feat-new-feature

# Interactive selection
wt remove groups
# Worktrees for "groups":
#   [1] feat-new-feature (feat/new-feature)
#   [2] fix-bug (fix/bug)
#
# Select worktree to remove (1-2):
```

### `wt init`

Initializes the global configuration. Prompts for the worktree directory and writes `~/.config/wt/config.toml`.

### `wt help`

Shows all available commands, flags, and examples.

## Running Tests

```bash
go test ./...
```

Integration tests create real git repos in temp directories and clean up after themselves.
