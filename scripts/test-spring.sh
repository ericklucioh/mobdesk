#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-spring-$$_${RANDOM}"
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
export TEST_PASSWORD='Mobdesk-spring-123!'
export TEST_DIR="$HOME/.cache/mobdesk-spring"
export MOBDESK_TEST_BIN="$TEST_DIR/mobdesk"
export FIXTURE_DIR="$HOME/mobdesk/scripts/fixtures/spring"
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

"$MOBDESK_TEST_BIN" install maven --json > "$TEST_DIR/maven.json"
grep -q '"success":true' "$TEST_DIR/maven.json"
java --version >/dev/null
mvn --version >/dev/null
! grep -q 'JAVA_HOME' "$HOME/.bashrc" 2>/dev/null

cp -R "$FIXTURE_DIR" "$WORKSPACE/spring"
cd "$WORKSPACE/spring"
mvn --batch-mode --no-transfer-progress verify

server_pid=''
stop_server() {
    if test -n "$server_pid" && kill -0 "$server_pid" 2>/dev/null; then
        kill "$server_pid"
    fi
    if test -n "$server_pid"; then
        wait "$server_pid" 2>/dev/null || true
    fi
}
trap stop_server EXIT

java -jar target/spring-health-1.0.0.jar > "$TEST_DIR/spring.log" 2>&1 &
server_pid=$!
for _ in $(seq 1 30); do
    if test "$(curl --fail --silent --max-time 1 http://127.0.0.1:8080/health)" = '{"status":"ok"}'; then
        break
    fi
    if ! kill -0 "$server_pid" 2>/dev/null; then
        cat "$TEST_DIR/spring.log" >&2
        exit 1
    fi
    sleep 1
done
test "$(curl --fail --silent --max-time 1 http://127.0.0.1:8080/health)" = '{"status":"ok"}'
stop_server
server_pid=''
trap - EXIT

printf '%s\n' 'Termux Spring Boot test: PASS'
CONTAINER_SCRIPT
