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

# As três primeiras linguagens devem ser instaláveis no Ubuntu e utilizáveis.
mobdesk install go > "$TEST_DIR/install-go.log"
mobdesk install python > "$TEST_DIR/install-python.log"
mobdesk install node > "$TEST_DIR/install-node.log"
mobdesk install c > "$TEST_DIR/install-c.log"
mobdesk install cpp > "$TEST_DIR/install-cpp.log"
mobdesk install lua > "$TEST_DIR/install-lua.log"
grep -q 'Lua' "$TEST_DIR/install-lua.log"

proot-distro login ubuntu -- tee /root/workspace/hello.go < scripts/fixtures/hello/go/main.go >/dev/null
proot-distro login ubuntu -- tee /root/workspace/hello.py < scripts/fixtures/hello/python/main.py >/dev/null
proot-distro login ubuntu -- tee /root/workspace/hello.js < scripts/fixtures/hello/node/main.js >/dev/null
proot-distro login ubuntu -- tee /root/workspace/hello.c < scripts/fixtures/hello/c/main.c >/dev/null
proot-distro login ubuntu -- tee /root/workspace/hello.cpp < scripts/fixtures/hello/cpp/main.cpp >/dev/null
proot-distro login ubuntu -- tee /root/workspace/hello.lua < scripts/fixtures/hello/lua/main.lua >/dev/null

proot-distro login ubuntu -- go build -o /tmp/mobdesk-hello-go /root/workspace/hello.go
proot-distro login ubuntu -- clang /root/workspace/hello.c -o /tmp/mobdesk-hello-c
proot-distro login ubuntu -- clang++ /root/workspace/hello.cpp -o /tmp/mobdesk-hello-cpp
test "$(proot-distro login ubuntu -- /tmp/mobdesk-hello-go)" = "hello-go"
test "$(proot-distro login ubuntu -- python3 /root/workspace/hello.py)" = "hello-python"
test "$(proot-distro login ubuntu -- node /root/workspace/hello.js)" = "hello-node"
test "$(proot-distro login ubuntu -- /tmp/mobdesk-hello-c)" = "hello-c"
test "$(proot-distro login ubuntu -- /tmp/mobdesk-hello-cpp)" = "hello-cpp"
test "$(proot-distro login ubuntu -- lua5.4 /root/workspace/hello.lua)" = "hello-lua"

# Repetir a instalação não deve atualizar índices nem reinstalar pacotes.
mobdesk install go > "$TEST_DIR/install-go-second.log"
mobdesk install python > "$TEST_DIR/install-python-second.log"
mobdesk install node > "$TEST_DIR/install-node-second.log"
mobdesk install c > "$TEST_DIR/install-c-second.log"
mobdesk install cpp > "$TEST_DIR/install-cpp-second.log"
mobdesk install lua > "$TEST_DIR/install-lua-second.log"
grep -q 'já estava instalada' "$TEST_DIR/install-go-second.log"
grep -q 'já estava instalada' "$TEST_DIR/install-python-second.log"
grep -q 'já estava instalada' "$TEST_DIR/install-node-second.log"
grep -q 'já estava instalada' "$TEST_DIR/install-c-second.log"
grep -q 'já estava instalada' "$TEST_DIR/install-cpp-second.log"
grep -q 'já estava instalada' "$TEST_DIR/install-lua-second.log"
test -f "$HOME/.local/share/mobdesk/state/installations/go.json"
test -f "$HOME/.local/share/mobdesk/state/installations/python.json"
test -f "$HOME/.local/share/mobdesk/state/installations/node.json"
test -f "$HOME/.local/share/mobdesk/state/installations/c.json"
test -f "$HOME/.local/share/mobdesk/state/installations/cpp.json"
test -f "$HOME/.local/share/mobdesk/state/installations/lua.json"
test -f "$HOME/.local/share/mobdesk/logs/install/go.log"
test -f "$HOME/.local/share/mobdesk/logs/install/python.log"
test -f "$HOME/.local/share/mobdesk/logs/install/node.log"
test -f "$HOME/.local/share/mobdesk/logs/install/c.log"
test -f "$HOME/.local/share/mobdesk/logs/install/cpp.log"
test -f "$HOME/.local/share/mobdesk/logs/install/lua.log"
mobdesk status --json > "$TEST_DIR/mobdesk-status.json"
grep -q '"installations"' "$TEST_DIR/mobdesk-status.json"
grep -q '"name": "go"' "$TEST_DIR/mobdesk-status.json"
grep -q '"name": "python"' "$TEST_DIR/mobdesk-status.json"
grep -q '"name": "node"' "$TEST_DIR/mobdesk-status.json"
grep -q '"name": "c"' "$TEST_DIR/mobdesk-status.json"
grep -q '"name": "cpp"' "$TEST_DIR/mobdesk-status.json"
grep -q '"name": "lua"' "$TEST_DIR/mobdesk-status.json"

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
