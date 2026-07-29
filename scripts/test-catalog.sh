#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_COMMAND="${COMPOSE:-docker compose}"
read -r -a COMPOSE_ARGS <<< "${COMPOSE_COMMAND}"
PROJECT="mobdesk-catalog-$$_${RANDOM}"

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

export PATH="/data/data/com.termux/files/usr/bin:/data/data/com.termux/files/home/go/bin:${PATH}"
MOBDESK=/data/data/com.termux/files/home/mobdesk/bin/mobdesk

# The fixture already contains a PRoot Ubuntu. Do not run setup here: the
# separate integration test owns the clean-install user journey.
proot-distro login ubuntu -- true

for tool in go python node c cpp lua git gh tmux micro lazygit tree ttt htop ncdu inxi speedtest-cli posting yazi tuifi neovim opencode-cli codex-cli claudecode-cli; do
    "$MOBDESK" install "$tool"
done

if test "$(proot-distro login ubuntu -- uname -m)" = "aarch64"; then
    "$MOBDESK" install zellij
fi

proot-distro login ubuntu -- git --version
proot-distro login ubuntu -- gh --version
proot-distro login ubuntu -- tmux -V
if test "$(proot-distro login ubuntu -- uname -m)" = "aarch64"; then
    proot-distro login ubuntu -- sh -ec 'PATH="$HOME/.local/bin:$PATH"; zellij --version'
fi
proot-distro login ubuntu -- micro --version
proot-distro login ubuntu -- lazygit --version
proot-distro login ubuntu -- tree --version
proot-distro login ubuntu -- ttt --help
proot-distro login ubuntu -- htop --version
proot-distro login ubuntu -- ncdu --version
proot-distro login ubuntu -- inxi --version
proot-distro login ubuntu -- speedtest-cli --version
proot-distro login ubuntu -- posting --help
proot-distro login ubuntu -- sh -ec 'PATH="$HOME/.local/bin:$PATH"; yazi --version'
proot-distro login ubuntu -- sh -ec 'PATH="$HOME/.local/bin:$PATH"; ya --version'
proot-distro login ubuntu -- tuifi --version
proot-distro login ubuntu -- nvim --version
test ! -e /root/.config/nvim
proot-distro login ubuntu -- sh -ec 'PATH="$HOME/.local/bin:$PATH"; opencode --version'
proot-distro login ubuntu -- sh -ec 'PATH="$HOME/.local/bin:$PATH"; codex --version'
proot-distro login ubuntu -- sh -ec 'PATH="$HOME/.local/bin:$PATH"; claude --version'

# A second pass proves profiles remain idempotent without exercising login or
# network benchmarks.
for tool in git yazi tuifi nvim opencode-cli codex-cli claudecode-cli; do
    "$MOBDESK" install "$tool"
done
CONTAINER_SCRIPT
