#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Building wt-bin..."
(cd "$SCRIPT_DIR" && go build -o wt-bin .)
echo "    Built $SCRIPT_DIR/wt-bin"

echo ""
echo "==> Initializing global config..."
"$SCRIPT_DIR/wt-bin" init

# Detect shell config file
SHELL_NAME="$(basename "$SHELL")"
case "$SHELL_NAME" in
  zsh)  SHELL_RC="$HOME/.zshrc" ;;
  bash) SHELL_RC="$HOME/.bashrc" ;;
  *)    SHELL_RC="" ;;
esac

SOURCE_LINE="source \"$SCRIPT_DIR/shell/wt.sh\""

if [[ -n "$SHELL_RC" ]]; then
  if grep -qF "shell/wt.sh" "$SHELL_RC" 2>/dev/null; then
    echo ""
    echo "==> Shell wrapper already sourced in $SHELL_RC, skipping."
  else
    echo ""
    printf "Add wt shell wrapper to %s? [Y/n]: " "$SHELL_RC"
    read -r answer
    if [[ "${answer:-Y}" =~ ^[Yy]$ ]]; then
      echo "" >> "$SHELL_RC"
      echo "# wt - Git Worktree Manager" >> "$SHELL_RC"
      echo "$SOURCE_LINE" >> "$SHELL_RC"
      echo "    Added to $SHELL_RC"
    else
      echo "    Skipped. Add this line manually:"
      echo "    $SOURCE_LINE"
    fi
  fi
else
  echo ""
  echo "==> Could not detect shell config file. Add this line to your shell rc:"
  echo "    $SOURCE_LINE"
fi

echo ""
echo "==> Done! Restart your shell or run:"
echo "    source $SHELL_RC"
