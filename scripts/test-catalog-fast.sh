#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-catalog-$$_${RANDOM}"
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
export TEST_PASSWORD='Mobdesk-catalog-123!'
export TEST_DIR="$HOME/.cache/mobdesk-catalog"
export MOBDESK_TEST_BIN="$TEST_DIR/mobdesk"
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

profiles=(git neovim tmux go java maven python node c cpp lua gh tree htop ncdu micro)
for profile in "${profiles[@]}"; do
    "$MOBDESK_TEST_BIN" install "$profile" --json > "$TEST_DIR/${profile}-first.json"
    grep -q '"success":true' "$TEST_DIR/${profile}-first.json"
    "$MOBDESK_TEST_BIN" install "$profile" --json > "$TEST_DIR/${profile}-second.json"
    grep -q '"success":true' "$TEST_DIR/${profile}-second.json"
    grep -q '"changed":false' "$TEST_DIR/${profile}-second.json"
    test -f "$HOME/.local/share/mobdesk/state/installations/${profile}.json"
done

git --version >/dev/null
nvim --version >/dev/null
tmux -V >/dev/null
go version >/dev/null
java --version >/dev/null
javac --version >/dev/null
jar --version >/dev/null
! grep -q 'JAVA_HOME' "$HOME/.bashrc" 2>/dev/null
mvn --version >/dev/null
python --version >/dev/null
node --version >/dev/null
npm --version >/dev/null
clang --version >/dev/null
clang++ --version >/dev/null
lua -v >/dev/null 2>&1
gh --version >/dev/null
tree --version >/dev/null
htop --version >/dev/null
ncdu --version >/dev/null
micro --version >/dev/null

"$MOBDESK_TEST_BIN" status --json > "$TEST_DIR/status.json"
for profile in "${profiles[@]}"; do
	grep -q "\"name\": \"${profile}\"" "$TEST_DIR/status.json"
done
grep -q '"java_home":' "$TEST_DIR/status.json"
"$MOBDESK_TEST_BIN" logs --json > "$TEST_DIR/logs.json"
grep -q '"schema_version": 1' "$TEST_DIR/logs.json"
grep -q '"command": "logs"' "$TEST_DIR/logs.json"
grep -q '"success": true' "$TEST_DIR/logs.json"
grep -q '"state":' "$TEST_DIR/logs.json"
grep -q '"message":' "$TEST_DIR/logs.json"

# A standalone pkg profile may be removed after Mobdesk records ownership.
"$MOBDESK_TEST_BIN" uninstall tree --json > "$TEST_DIR/tree-uninstall.json"
grep -q '"success":true' "$TEST_DIR/tree-uninstall.json"
grep -q '"state":"uninstalled"' "$TEST_DIR/tree-uninstall.json"
grep -q '"state": "uninstalled"' "$HOME/.local/share/mobdesk/state/installations/tree.json"
! command -v tree >/dev/null 2>&1

# C and C++ share clang. Removing one releases its record, and removing the
# final owner removes the package.
"$MOBDESK_TEST_BIN" uninstall c --json > "$TEST_DIR/c-uninstall.json"
grep -q '"success":true' "$TEST_DIR/c-uninstall.json"
grep -q '"preserved_packages":\["clang"\]' "$TEST_DIR/c-uninstall.json"
command -v clang
command -v clang++
"$MOBDESK_TEST_BIN" uninstall cpp --json > "$TEST_DIR/cpp-uninstall.json"
grep -q '"success":true' "$TEST_DIR/cpp-uninstall.json"
! command -v clang >/dev/null 2>&1
! command -v clang++ >/dev/null 2>&1
! command -v proot-distro >/dev/null 2>&1
printf '%s\n' 'Termux native pkg catalog test: PASS'
CONTAINER_SCRIPT
