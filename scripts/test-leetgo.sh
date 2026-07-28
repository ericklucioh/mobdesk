#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-leetgo-$$_${RANDOM}"

compose() {
    "${COMPOSE_ARGS[@]}" -p "${PROJECT}" "$@"
}

cleanup() {
    compose down --remove-orphans >/dev/null 2>&1 || true
}

trap cleanup EXIT

cd "${ROOT_DIR}"
compose build termux-catalog
compose run --rm termux-catalog bash -s <<'CONTAINER_SCRIPT'
set -Eeuo pipefail

MOBDESK=/data/data/com.termux/files/home/mobdesk/bin/mobdesk
proot-distro login ubuntu -- true
"$MOBDESK" install leetgo
proot-distro login ubuntu -- leetgo --help
"$MOBDESK" install leetgo
CONTAINER_SCRIPT
