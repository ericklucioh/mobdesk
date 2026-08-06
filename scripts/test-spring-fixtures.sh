#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MOBDESK="${MOBDESK:-/data/data/com.termux/files/home/mobdesk/bin/mobdesk}"

export MOBDESK_TEST_MODE=1
export MOBDESK_TEST_STORAGE_FREE_BYTES=$((25 * 1024 * 1024 * 1024))
export PATH="/data/data/com.termux/files/usr/bin:/data/data/com.termux/files/home/go/bin:${PATH}"

"${MOBDESK}" install java
"${MOBDESK}" install kotlin
"${MOBDESK}" install gradle
"${MOBDESK}" install maven

proot-distro login ubuntu -- mkdir -p /root/workspace/spring-fixtures
tar -C "${ROOT_DIR}/scripts/fixtures" -cf - spring-java-gradle spring-java-maven spring-kotlin-gradle |
    proot-distro login ubuntu -- tar -xf - -C /root/workspace/spring-fixtures

proot-distro login ubuntu -- sh -ec '
    set -eu
    export PATH="$HOME/.local/bin:$PATH"
    export JAVA_HOME="$(dirname "$(dirname "$(readlink -f "$(command -v javac)")")")"

    run_process() {
        expected=$1
        port=$2
        shift 2
        log=$(mktemp)
        "$@" >"$log" 2>&1 &
        pid=$!
        cleanup() { kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; rm -f "$log"; }
        for attempt in $(seq 1 30); do
            if test "$(curl -fsS "http://127.0.0.1:$port/health" 2>/dev/null || true)" = "$expected"; then
                cleanup
                return 0
            fi
            sleep 1
        done
        cat "$log" >&2
        exit 1
    }

    cd /root/workspace/spring-fixtures/spring-java-gradle
    gradle --no-daemon clean test bootJar
    run_process spring-java-gradle 18080 gradle --no-daemon bootRun --args=--server.port=18080
    run_process spring-java-gradle 18081 java -jar "$(find build/libs -maxdepth 1 -type f -name "*.jar" ! -name "*-plain.jar" -print -quit)" --server.port=18081
    cd ../spring-kotlin-gradle
    gradle --no-daemon clean test bootJar
    run_process spring-kotlin-gradle 18082 java -jar "$(find build/libs -maxdepth 1 -type f -name "*.jar" ! -name "*-plain.jar" -print -quit)" --server.port=18082
    cd ../spring-java-maven
    mvn --batch-mode test package
    run_process spring-java-maven 18083 java -jar "$(find target -maxdepth 1 -type f -name "*.jar" ! -name "*.original" -print -quit)" --server.port=18083
    run_process spring-java-maven 18084 mvn --batch-mode spring-boot:run -Dspring-boot.run.arguments=--server.port=18084
'
