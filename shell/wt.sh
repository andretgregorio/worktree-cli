# Resolve the binary path relative to this script
__WT_BIN="$(cd "$(dirname "${(%):-%x}")/.." && pwd)/wt-bin"

if [[ ! -x "$__WT_BIN" ]]; then
  echo "wt: binary not found at $__WT_BIN (run 'go build -o wt-bin .' in the worktree-cli directory)" >&2
  return 1 2>/dev/null || exit 1
fi

wt() {
  local output
  output=$("$__WT_BIN" "$@")
  local exit_code=$?

  local cd_path
  cd_path=$(echo "$output" | grep '^__WT_CD__:' | cut -d: -f2-)

  # Export any environment variables
  local env_line
  while IFS= read -r env_line; do
    if [[ "$env_line" == __WT_ENV__:* ]]; then
      local kv="${env_line#__WT_ENV__:}"
      export "$kv"
    fi
  done <<< "$output"

  # Print everything except markers
  echo "$output" | grep -v '^__WT_CD__:' | grep -v '^__WT_ENV__:'

  if [[ $exit_code -eq 0 && -n "$cd_path" && -d "$cd_path" ]]; then
    cd "$cd_path" || return 1
  fi

  return $exit_code
}
