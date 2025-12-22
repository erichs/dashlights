#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${IMAGE:-golang:1.25-rc-bullseye}"

docker run --rm -it \
  -v "${REPO_DIR}:/work:ro" \
  -w /work \
  "${IMAGE}" \
  bash -lc "$(cat <<'INNER'
set -euo pipefail
trap 'echo "FAILED: $BASH_COMMAND" >&2' ERR

export PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
echo "STEP: install deps"
apt-get update
apt-get install -y zsh fish ripgrep util-linux

echo "STEP: build dashlights"
go build -o /tmp/dashlights ./src
export PATH="/tmp:$PATH"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

begin_test() {
  printf 'TEST: %s -- ' "$1"
}

end_test() {
  echo "OK"
}

assert_contains() {
  local hay="$1"
  local needle="$2"
  echo "$hay" | rg -F -- "$needle" >/dev/null || fail "Expected output to contain: $needle"
}

assert_file_contains() {
  local file="$1"
  local needle="$2"
  rg -F -- "$needle" "$file" >/dev/null || fail "Expected ${file} to contain: $needle"
}

expect_success() {
  local desc="$1"
  shift
  local output
  if ! output="$("$@" 2>&1)"; then
    echo "$output" >&2
    fail "$desc failed"
  fi
  echo "$output"
}

expect_failure() {
  local desc="$1"
  shift
  local output
  if output="$("$@" 2>&1)"; then
    echo "$output" >&2
    fail "$desc expected failure"
  fi
  echo "$output"
}

run_piped_cmd() {
  local desc="$1"
  local cmd="$2"
  local output
  local status
  set +e
  output="$(bash -c "$cmd" 2>&1)"
  status=$?
  set -e
  echo "$output"
  echo "$status"
}

reset_home() {
  rm -rf "$HOME"
  mkdir -p "$HOME"
}

export HOME="/tmp/dashlights-home"

begin_test "version check"
output="$(expect_success "version check" dashlights --version)"
assert_contains "$output" "dashlights"
end_test

begin_test "bash install"
reset_home
export SHELL="/bin/bash"
echo 'export BASH_TEST=1' > "$HOME/.bashrc"
output="$(expect_success "bash install" dashlights --installprompt -y)"
assert_contains "$output" "Installed dashlights into $HOME/.bashrc"
assert_file_contains "$HOME/.bashrc" "# BEGIN dashlights"
assert_file_contains "$HOME/.bashrc" "# END dashlights"
assert_file_contains "$HOME/.bashrc" "PS1="
test -f "$HOME/.bashrc.dashlights-backup" || fail "Expected backup to exist for bash install"
backup_mtime="$(stat -c %Y "$HOME/.bashrc.dashlights-backup")"
end_test
begin_test "bash idempotency"
output="$(expect_success "bash idempotency" dashlights --installprompt -y)"
assert_contains "$output" "already installed"
test "$(stat -c %Y "$HOME/.bashrc.dashlights-backup")" = "$backup_mtime" || fail "Backup changed on idempotent bash install"
end_test

begin_test "zsh install"
reset_home
export SHELL="/bin/zsh"
echo 'export ZSH_TEST=1' > "$HOME/.zshrc"
output="$(expect_success "zsh install" dashlights --installprompt -y)"
assert_contains "$output" "Installed dashlights into $HOME/.zshrc"
assert_file_contains "$HOME/.zshrc" "setopt prompt_subst"
assert_file_contains "$HOME/.zshrc" "PROMPT="
test -f "$HOME/.zshrc.dashlights-backup" || fail "Expected backup to exist for zsh install"
backup_mtime="$(stat -c %Y "$HOME/.zshrc.dashlights-backup")"
end_test
begin_test "zsh idempotency"
output="$(expect_success "zsh idempotency" dashlights --installprompt -y)"
assert_contains "$output" "already installed"
test "$(stat -c %Y "$HOME/.zshrc.dashlights-backup")" = "$backup_mtime" || fail "Backup changed on idempotent zsh install"
end_test

begin_test "p10k install with elements"
reset_home
export SHELL="/bin/zsh"
cat > "$HOME/.p10k.zsh" <<'EOF'
# p10k sample
typeset -g POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
  dir
  vcs
)
EOF
output="$(expect_success "p10k install with elements" dashlights --installprompt -y)"
assert_contains "$output" "Installed dashlights into $HOME/.p10k.zsh"
assert_file_contains "$HOME/.p10k.zsh" "prompt_dashlights"
assert_file_contains "$HOME/.p10k.zsh" "POWERLEVEL9K_LEFT_PROMPT_ELEMENTS"
assert_file_contains "$HOME/.p10k.zsh" "dashlights"
backup_mtime="$(stat -c %Y "$HOME/.p10k.zsh.dashlights-backup")"
end_test
begin_test "p10k idempotency with elements"
output="$(expect_success "p10k idempotency with elements" dashlights --installprompt -y)"
assert_contains "$output" "already installed"
test "$(stat -c %Y "$HOME/.p10k.zsh.dashlights-backup")" = "$backup_mtime" || fail "Backup changed on idempotent p10k install with elements"
end_test

begin_test "p10k install without elements"
reset_home
export SHELL="/bin/zsh"
cat > "$HOME/.p10k.zsh" <<'EOF'
# p10k sample without prompt elements array
typeset -g POWERLEVEL9K_MODE=nerdfont-complete
EOF
output="$(expect_success "p10k install without elements" dashlights --installprompt -y)"
assert_contains "$output" "Could not locate POWERLEVEL9K_LEFT_PROMPT_ELEMENTS"
assert_file_contains "$HOME/.p10k.zsh" "prompt_dashlights"
backup_mtime="$(stat -c %Y "$HOME/.p10k.zsh.dashlights-backup")"
end_test
begin_test "p10k idempotency without elements"
output="$(expect_success "p10k idempotency without elements" dashlights --installprompt -y)"
assert_contains "$output" "already installed"
test "$(stat -c %Y "$HOME/.p10k.zsh.dashlights-backup")" = "$backup_mtime" || fail "Backup changed on idempotent p10k install without elements"
end_test

begin_test "fish install"
reset_home
mkdir -p "$HOME/.config/fish"
export SHELL="/usr/bin/fish"
output="$(expect_success "fish install" dashlights --installprompt -y)"
assert_contains "$output" "Installed dashlights into $HOME/.config/fish/config.fish"
assert_file_contains "$HOME/.config/fish/config.fish" "# BEGIN dashlights"
assert_file_contains "$HOME/.config/fish/config.fish" "function __dashlights_prompt"
test ! -f "$HOME/.config/fish/config.fish.dashlights-backup" || fail "Did not expect fish backup for new file"
end_test
begin_test "fish idempotency"
output="$(expect_success "fish idempotency" dashlights --installprompt -y)"
assert_contains "$output" "already installed"
test ! -f "$HOME/.config/fish/config.fish.dashlights-backup" || fail "Did not expect fish backup on idempotent install"
end_test

begin_test "configpath install"
reset_home
export SHELL="/bin/zsh"
echo '# custom bashrc' > "$HOME/custom.bashrc"
output="$(expect_success "configpath install" dashlights --installprompt --configpath "$HOME/custom.bashrc" -y)"
assert_contains "$output" "Installed dashlights into $HOME/custom.bashrc"
assert_file_contains "$HOME/custom.bashrc" "PS1="
test -f "$HOME/custom.bashrc.dashlights-backup" || fail "Expected backup to exist for configpath install"
end_test

begin_test "unsupported shell"
reset_home
export SHELL="/bin/unknownshell"
output="$(expect_failure "unsupported shell" dashlights --installprompt -y)"
assert_contains "$output" "unsupported shell"
end_test

begin_test "configpath directory"
reset_home
mkdir -p "$HOME/dir"
export SHELL="/bin/bash"
output="$(expect_failure "configpath directory" dashlights --installprompt --configpath "$HOME/dir" -y)"
assert_contains "$output" "--configpath must be a file, not a directory"
end_test

begin_test "corrupted install"
reset_home
export SHELL="/bin/bash"
cat > "$HOME/.bashrc" <<'EOF'
# BEGIN dashlights
# corrupted block
EOF
output="$(expect_failure "corrupted install" dashlights --installprompt -y)"
assert_contains "$output" "Corrupted dashlights installation detected"
end_test

begin_test "dry run"
reset_home
export SHELL="/bin/bash"
echo 'export BASH_TEST=1' > "$HOME/.bashrc"
before_hash="$(sha256sum "$HOME/.bashrc" | awk '{print $1}')"
output="$(expect_success "dry run" dashlights --installprompt --dry-run)"
assert_contains "$output" "[DRY-RUN]"
test ! -f "$HOME/.bashrc.dashlights-backup" || fail "Dry run should not create backup"
after_hash="$(sha256sum "$HOME/.bashrc" | awk '{print $1}')"
test "$before_hash" = "$after_hash" || fail "Dry run should not modify bashrc"
end_test

begin_test "claude install"
reset_home
output="$(expect_success "claude install" dashlights --installagent claude -y)"
assert_contains "$output" "Installed dashlights into $HOME/.claude/settings.json"
assert_file_contains "$HOME/.claude/settings.json" "dashlights --agentic"
assert_file_contains "$HOME/.claude/settings.json" "PreToolUse"
end_test
begin_test "claude idempotency"
output="$(expect_success "claude idempotency" dashlights --installagent claude -y)"
assert_contains "$output" "already installed"
end_test

begin_test "cursor install"
reset_home
output="$(expect_success "cursor install" dashlights --installagent cursor -y)"
assert_contains "$output" "Installed dashlights into $HOME/.cursor/hooks.json"
assert_file_contains "$HOME/.cursor/hooks.json" "beforeShellExecution"
assert_file_contains "$HOME/.cursor/hooks.json" "dashlights --agentic"
end_test
begin_test "cursor idempotency"
output="$(expect_success "cursor idempotency" dashlights --installagent cursor -y)"
assert_contains "$output" "already installed"
end_test

begin_test "cursor conflict"
reset_home
mkdir -p "$HOME/.cursor"
cat > "$HOME/.cursor/hooks.json" <<'EOF'
{"beforeShellExecution":{"command":"echo existing"}}
EOF
output_status="$(run_piped_cmd "cursor decline" 'printf "n\n" | script -q -e -c "dashlights --installagent cursor" /dev/null')"
output="$(echo "$output_status" | sed '$d')"
status="$(echo "$output_status" | tail -n 1)"
test "$status" -eq 0 || test "$status" -eq 1 || fail "Unexpected exit code for interactive decline: $status"
assert_contains "$output" "Installation cancelled"
assert_file_contains "$HOME/.cursor/hooks.json" "echo existing"

output="$(expect_failure "cursor conflict non-interactive" dashlights --installagent cursor -y)"
assert_contains "$output" "already has a beforeShellExecution hook"

output_status="$(run_piped_cmd "cursor accept" 'printf "y\ny\n" | script -q -e -c "dashlights --installagent cursor" /dev/null')"
output="$(echo "$output_status" | sed '$d')"
status="$(echo "$output_status" | tail -n 1)"
test "$status" -eq 0 || test "$status" -eq 1 || fail "Unexpected exit code for interactive accept: $status"
assert_file_contains "$HOME/.cursor/hooks.json" "dashlights --agentic"
end_test

begin_test "invalid agent"
output="$(expect_failure "invalid agent" dashlights --installagent not-an-agent -y)"
assert_contains "$output" "unsupported agent"
end_test

begin_test "configpath with installagent"
output="$(expect_failure "configpath with installagent" dashlights --installagent claude --configpath /tmp/fake -y)"
assert_contains "$output" "--configpath cannot be used with --installagent"
end_test

echo "All dockerized install tests passed."
INNER
)"
