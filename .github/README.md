# Mobdesk

[English](README.md) | [Portugues](README.pt-BR.md)

> **MVP / experimental:** Mobdesk works and has been tested on a real Android
> device. Validation across a broader device matrix is still in progress. Use
> it for learning, development and lightweight local services, not production
> workloads.

## Your Linux workstation in your pocket

Mobdesk turns an Android phone into a personal Ubuntu development workstation:

```text
Android
  Termux -> Mobdesk -> Ubuntu via PRoot
                      -> local shell or SSH on port 8022
```

Ubuntu persists on the phone. You can work locally or connect from another
computer on a trusted network. Mobdesk does not require root, a virtual
machine, Docker on the phone, systemd or a graphical desktop.

## What is available

- persistent Ubuntu through PRoot-Distro;
- dedicated SSH server on port `8022`;
- local shell access with `mobdesk shell`;
- touch/mouse/keyboard TUI;
- status and JSON output for automation;
- Go, Python, Node.js, C, C++ and Lua installation profiles;
- Neovim/LazyVim configuration profiles;
- rollback-aware binary updates;
- English (`en-US`) and Brazilian Portuguese (`pt-BR`) presentation.

Projects, persistent sessions, services and a web interface remain future
roadmap stages.

## Requirements

- ARM64 Android phone;
- Termux from [F-Droid](https://f-droid.org/packages/com.termux/) or the
  [official releases](https://github.com/termux/termux-app/releases);
- approximately 1.5 GB of free storage for the base Ubuntu installation;
- a trusted local network for remote SSH access.

## Installation

The recommended method uses the published ARM64 binary and does not require
Go. The latest stable binary is available from the
[releases page](https://github.com/ericklucioh/mobdesk/releases).

### Released ARM64 binary

```bash
pkg update
pkg install -y curl coreutils

BASE_URL="https://github.com/ericklucioh/mobdesk/releases/latest/download"
mkdir -p "$HOME/.local/bin"
cd "$HOME/.local/bin"
curl -fL -o mobdesk-linux-arm64 "${BASE_URL}/mobdesk-linux-arm64"
curl -fL -o SHA256SUMS "${BASE_URL}/SHA256SUMS"
sha256sum -c SHA256SUMS
mv mobdesk-linux-arm64 mobdesk
chmod 0755 mobdesk
"$HOME/.local/bin/mobdesk" setup
```

The checksum verifies integrity. Releases are not independently signed yet,
so the checksum does not authenticate their origin.

### Build with Go

The project requires Go `1.26.5` or newer:

```bash
pkg update
pkg install -y golang git
go version
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
~/go/bin/mobdesk setup
```

`@latest` means the latest stable semantic-version tag. It does not mean an
untagged commit or a `test-v*` prerelease. Pin `@v0.6.0`, or another explicit
release tag, when reproducibility matters.

## First run

Setup installs the required Termux packages, downloads Ubuntu, creates the
persistent workspace, configures SSH and installs the `mobdesk` launcher.
After the first run:

```bash
mobdesk status
mobdesk start
mobdesk shell
mobdesk stop
```

`mobdesk start` starts SSH and displays connection details. It does not open a
local Ubuntu shell automatically. Use `mobdesk shell` for local access or the
displayed SSH command from another computer.

## TUI and language

Run `mobdesk tui` in Termux. `Tab` changes focus, `Enter` activates an action,
`Esc` goes back and `q` starts the quit confirmation. The same TUI can run
inside Ubuntu over SSH; host-only actions are blocked and explained there.

English is the default. Select Brazilian Portuguese with:

```bash
mobdesk tui --locale pt-BR
MOBDESK_LOCALE=pt-BR mobdesk tui
```

The TUI has no in-app language button yet; restart it with the desired locale.

## Security and limitations

Use SSH only on trusted networks or through a private tunnel. Never expose port
`8022` directly to the public internet. The current MVP uses password
authentication and local-network listening.

PRoot is not a virtual machine and does not provide a separate kernel. Android
may suspend Termux, and the project is not designed for heavy production
workloads, real Docker, systemd, kernel modules or privileged device access.

## More information

The [root README](../README.md) contains the complete technical documentation.

- [Mission](../docs/MISSION.md)
- [Architecture](../docs/ARCHITECTURE.md)
- [Roadmap](../docs/ROADMAP.md)
- [Changelog](../CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](../CODE_OF_CONDUCT.md)
- [Support](SUPPORT.md)
- [Security policy](../SECURITY.md)

## License

Mobdesk is distributed under the [MIT license](../LICENSE).
