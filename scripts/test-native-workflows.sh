#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-workflows-$$_${RANDOM}"
export MOBDESK_TERMUX_HOME_VOLUME="${PROJECT}_home"
export MOBDESK_TERMUX_PREFIX_VOLUME="${PROJECT}_prefix"

compose() { "${COMPOSE_ARGS[@]}" -p "${PROJECT}" "$@"; }
cleanup() { compose down --volumes --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

cd "${ROOT_DIR}"
compose build termux
compose run --rm termux bash -s <<'CONTAINER_SCRIPT'
set -Eeuo pipefail

export PATH="/data/data/com.termux/files/usr/bin:/data/data/com.termux/files/home/go/bin:${PATH}"
export TEST_PASSWORD='Mobdesk-workflows-123!'
export TEST_DIR="$HOME/.cache/mobdesk-workflows"
export MOBDESK_TEST_BIN="$TEST_DIR/mobdesk"
export FIXTURES_DIR="$HOME/mobdesk/scripts/fixtures/hello"
export WORKSPACE="$HOME/workspace"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
go build -o "$MOBDESK_TEST_BIN" ./cmd/mobdesk

expect <<'EXPECT_SCRIPT' > "$TEST_DIR/setup.log"
set timeout 600
log_user 0
spawn $env(MOBDESK_TEST_BIN) setup
expect {
    -re {(?i)(new|retype|enter new|nova|redigite).*password} {
        send -- "$env(TEST_PASSWORD)\r"
        exp_continue
    }
    eof {}
    timeout { exit 1 }
}
set result [wait]
if {[lindex $result 3] != 0} { exit [lindex $result 3] }
EXPECT_SCRIPT

"$MOBDESK_TEST_BIN" setup > "$TEST_DIR/setup-second.log" 2>&1
! grep -q '\$ pkg update' "$TEST_DIR/setup-second.log"
test -d "$WORKSPACE"
test ! -L "$WORKSPACE"

profiles=(git neovim tmux go python node c cpp lua)
for profile in "${profiles[@]}"; do
    "$MOBDESK_TEST_BIN" install "$profile" --json > "$TEST_DIR/${profile}-first.json"
    grep -q '"success":true' "$TEST_DIR/${profile}-first.json"
    "$MOBDESK_TEST_BIN" install "$profile" --json > "$TEST_DIR/${profile}-second.json"
    grep -q '"success":true' "$TEST_DIR/${profile}-second.json"
    grep -q '"changed":false' "$TEST_DIR/${profile}-second.json"
done

cp -R "$FIXTURES_DIR/." "$WORKSPACE/"
cd "$WORKSPACE"
test "$(pwd)" = "$WORKSPACE"
mkdir -p .mobdesk-workflows

mkdir git
git -C git init -q
git -C git config user.email 'mobdesk@example.invalid'
git -C git config user.name 'Mobdesk Workflow Test'
printf '%s\n' 'hello-git' > git/README.md
git -C git add README.md
git -C git commit -qm 'Add workflow fixture'
test -z "$(git -C git status --porcelain)"

test "$(go run ./go/main.go)" = 'hello-go'
test "$(python ./python/main.py)" = 'hello-python'
npm --version >/dev/null
test "$(npm --prefix ./node run --silent smoke)" = 'hello-node'
clang ./c/main.c -o .mobdesk-workflows/hello-c
test "$(.mobdesk-workflows/hello-c)" = 'hello-c'
clang++ ./cpp/main.cpp -o .mobdesk-workflows/hello-cpp
test "$(.mobdesk-workflows/hello-cpp)" = 'hello-cpp'
test "$(lua ./lua/main.lua)" = 'hello-lua'
nvim --clean --headless '+edit go/main.go' '+if getline(1) !=# "package main" | cquit 1 | endif' '+qall!'

tmux_socket='mobdesk-workflow'
cleanup_tmux() { tmux -L "$tmux_socket" kill-server >/dev/null 2>&1 || true; }
trap cleanup_tmux EXIT
tmux -L "$tmux_socket" new-session -d -s workflow "printf '%s\\n' hello-tmux; sleep 30"
for _ in $(seq 1 10); do
    if tmux -L "$tmux_socket" capture-pane -pt workflow:0.0 2>/dev/null | grep -q 'hello-tmux'; then
        break
    fi
    sleep 1
done
tmux -L "$tmux_socket" capture-pane -pt workflow:0.0 | grep -q 'hello-tmux'
cleanup_tmux
trap - EXIT

"$MOBDESK_TEST_BIN" status --json > "$TEST_DIR/status.json"
grep -q '"workspace"' "$TEST_DIR/status.json"
for profile in "${profiles[@]}"; do
    grep -q "\"name\": \"${profile}\"" "$TEST_DIR/status.json"
done
! command -v proot-distro >/dev/null 2>&1
printf '%s\n' 'Termux native workflow test: PASS'
CONTAINER_SCRIPT
