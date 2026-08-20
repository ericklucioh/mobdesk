#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-integration-$$_${RANDOM}"
export MOBDESK_TERMUX_HOME_VOLUME="${PROJECT}_home"
export MOBDESK_TERMUX_PREFIX_VOLUME="${PROJECT}_prefix"

compose() { "${COMPOSE_ARGS[@]}" -p "${PROJECT}" "$@"; }
cleanup() { compose down --volumes --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

cd "${ROOT_DIR}"
compose build termux
compose run --rm --service-ports termux bash -s <<'CONTAINER_SCRIPT'
set -Eeuo pipefail
export PATH="/data/data/com.termux/files/usr/bin:/data/data/com.termux/files/home/go/bin:${PATH}"
export TEST_PASSWORD='Mobdesk-test-123!'
export TEST_DIR="$HOME/.cache/mobdesk-integration"
export MOBDESK_TEST_BIN="$TEST_DIR/mobdesk"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
go build -o "$MOBDESK_TEST_BIN" ./cmd/mobdesk

expect <<'EXPECT_SCRIPT' | tee "$TEST_DIR/mobdesk-setup.log"
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

test -f "$HOME/.local/share/mobdesk/setup.done"
test -d "$HOME/workspace"
test -f "$HOME/.config/mobdesk/ssh/sshd_config"
test -f "$HOME/.config/mobdesk/shell.bash"
! grep -q 'ForceCommand' "$HOME/.config/mobdesk/ssh/sshd_config"
! command -v proot-distro >/dev/null 2>&1

"$MOBDESK_TEST_BIN" setup > "$TEST_DIR/mobdesk-setup-second.log" 2>&1
! grep -q '\$ pkg update' "$TEST_DIR/mobdesk-setup-second.log"

"$MOBDESK_TEST_BIN" install git > "$TEST_DIR/install-git.log"
"$MOBDESK_TEST_BIN" install go > "$TEST_DIR/install-go.log"
"$MOBDESK_TEST_BIN" install git > "$TEST_DIR/install-git-second.log"
command -v git
command -v go
test -f "$HOME/.local/share/mobdesk/state/installations/git.json"
"$MOBDESK_TEST_BIN" status --json > "$TEST_DIR/mobdesk-status.json"
grep -q '"schema_version": 2' "$TEST_DIR/mobdesk-status.json"
grep -q '"workspace"' "$TEST_DIR/mobdesk-status.json"
! grep -q '"ubuntu"' "$TEST_DIR/mobdesk-status.json"

mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"
ssh-keygen -q -t ed25519 -N '' -f "$TEST_DIR/mobdesk-integration-key"
cat "$TEST_DIR/mobdesk-integration-key.pub" >> "$HOME/.ssh/authorized_keys"
chmod 600 "$HOME/.ssh/authorized_keys"
"$MOBDESK_TEST_BIN" start > "$TEST_DIR/mobdesk-start.log" 2>&1 &
start_pid=$!
for _ in $(seq 1 60); do
    if ssh-keyscan -T 1 -p 8022 127.0.0.1 >"$TEST_DIR/mobdesk-host-key" 2>/dev/null; then break; fi
    sleep 1
done
wait "$start_pid"
ssh -i "$TEST_DIR/mobdesk-integration-key" -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile="$TEST_DIR/mobdesk-known-hosts" -p 8022 "$(id -un)@127.0.0.1" 'test -n "$HOME" && test -n "$PREFIX" && command -v pkg >/dev/null && test -d "$HOME/workspace" && cd "$HOME/workspace" && go version >/dev/null'
"$MOBDESK_TEST_BIN" stop
printf '%s\n' 'Termux integration smoke test: PASS'
CONTAINER_SCRIPT
