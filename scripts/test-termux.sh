#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"

PROJECT="mobdesk-integration-$$_${RANDOM}"
HOME_VOLUME="${PROJECT}_home"
PREFIX_VOLUME="${PROJECT}_prefix"
export MOBDESK_TERMUX_HOME_VOLUME="${HOME_VOLUME}"
export MOBDESK_TERMUX_PREFIX_VOLUME="${PREFIX_VOLUME}"

compose() {
	"${COMPOSE_ARGS[@]}" -p "${PROJECT}" "$@"
}

cleanup() {
	compose down --volumes --remove-orphans >/dev/null 2>&1 || true
}

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
set timeout 1800
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
if {[lindex $result 3] != 0} {
    exit [lindex $result 3]
}
EXPECT_SCRIPT

test -f "$HOME/.local/share/mobdesk/setup.done"
test -f "$HOME/.config/mobdesk/ssh/sshd_config"
test -f "$HOME/.local/share/mobdesk/ssh/mobdesk-ssh-shell"

if test -f "$PREFIX/etc/ssh/sshd_config"; then
    ! grep -q 'mobdesk-ssh-shell' "$PREFIX/etc/ssh/sshd_config"
fi

if grep -q '\$ pkg upgrade' "$TEST_DIR/mobdesk-setup.log" 2>/dev/null; then
    printf '%s\n' 'setup executou pkg upgrade inesperadamente' >&2
    exit 1
fi

# A segunda execução deve usar as etapas persistidas e não reinstalar tudo.
mobdesk setup > "$TEST_DIR/mobdesk-setup-second.log" 2>&1
! grep -q '\$ pkg update' "$TEST_DIR/mobdesk-setup-second.log"
! grep -q '\$ pkg upgrade' "$TEST_DIR/mobdesk-setup-second.log"

mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"
ssh-keygen -q -t ed25519 -N '' -f "$TEST_DIR/mobdesk-integration-key"
cat "$TEST_DIR/mobdesk-integration-key.pub" >> "$HOME/.ssh/authorized_keys"
chmod 600 "$HOME/.ssh/authorized_keys"

printf 'exit\n' | mobdesk start > "$TEST_DIR/mobdesk-start.log" 2>&1 &
start_pid=$!

ssh_ready=false
for _ in $(seq 1 60); do
    if ssh-keyscan -T 1 -p 8022 127.0.0.1 >"$TEST_DIR/mobdesk-host-key" 2>/dev/null; then
        ssh_ready=true
        break
    fi
    sleep 1
done

if test "$ssh_ready" != true; then
    cat "$TEST_DIR/mobdesk-start.log" >&2
    exit 1
fi

wait "$start_pid"

printf 'test -f /etc/os-release && printf "ubuntu-ok\\n"\nexit\n' |
    ssh -i "$TEST_DIR/mobdesk-integration-key" \
        -o BatchMode=yes \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile="$TEST_DIR/mobdesk-known-hosts" \
        -p 8022 "$(id -un)@127.0.0.1" > "$TEST_DIR/mobdesk-ssh.log"
grep -q 'ubuntu-ok' "$TEST_DIR/mobdesk-ssh.log"

mobdesk stop
if ssh-keyscan -T 1 -p 8022 127.0.0.1 >/dev/null 2>&1; then
    printf '%s\n' 'a porta SSH continuou aberta após mobdesk stop' >&2
    exit 1
fi

# Um PID obsoleto não deve ser sinalizado nem bloquear a limpeza do estado.
printf '%s\n' "$$" > "$HOME/.local/share/mobdesk/ssh/sshd.pid"
mobdesk stop
test ! -e "$HOME/.local/share/mobdesk/ssh/sshd.pid"

printf '%s\n' 'Termux integration smoke test: PASS'
CONTAINER_SCRIPT
