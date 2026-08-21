#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-user-clis-$$_${RANDOM}"
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
export TEST_PASSWORD='Mobdesk-user-clis-123!'
export TEST_DIR="$HOME/.cache/mobdesk-user-clis"
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

profiles=(tuifi bitwarden resterm)
for profile in "${profiles[@]}"; do
    if ! "$MOBDESK_TEST_BIN" install "$profile" --json > "$TEST_DIR/${profile}-install.json"; then
        cat "$TEST_DIR/${profile}-install.json" >&2
        cat "$HOME/.local/share/mobdesk/logs/install/${profile}.log" >&2 || true
        exit 1
    fi
    grep -q '"success":true' "$TEST_DIR/${profile}-install.json"
    "$MOBDESK_TEST_BIN" install "$profile" --json > "$TEST_DIR/${profile}-reinstall.json"
    grep -q '"success":true' "$TEST_DIR/${profile}-reinstall.json"
    grep -q '"changed":false' "$TEST_DIR/${profile}-reinstall.json"
done

"$HOME/.local/bin/tuifi" --version >/dev/null
"$HOME/.local/bin/bw" --version >/dev/null
"$HOME/.local/bin/resterm" --version >/dev/null
"$MOBDESK_TEST_BIN" status --json > "$TEST_DIR/status.json"
for profile in "${profiles[@]}"; do
    grep -q "\"name\": \"${profile}\"" "$TEST_DIR/status.json"
done

for profile in "${profiles[@]}"; do
    if ! "$MOBDESK_TEST_BIN" uninstall "$profile" --json > "$TEST_DIR/${profile}-uninstall.json"; then
        cat "$TEST_DIR/${profile}-uninstall.json" >&2
        exit 1
    fi
    grep -q '"success":true' "$TEST_DIR/${profile}-uninstall.json"
    if grep -q '"conflicts"' "$TEST_DIR/${profile}-uninstall.json"; then
        cat "$TEST_DIR/${profile}-uninstall.json" >&2
        exit 1
    fi
done

for executable in tuifi bw resterm; do
    test ! -e "$HOME/.local/bin/$executable"
done
printf '%s\n' 'Termux private user CLI test: PASS'
CONTAINER_SCRIPT
