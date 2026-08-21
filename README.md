# Mobdesk

Turn an Android phone into a personal Termux development workstation.

**[Open the landing page](https://ericklucioh.github.io/mobdesk/)**

> **MVP / experimental:** Mobdesk is functional and has been tested on a real
> Android device. Validation across a broader device matrix is still in
> progress. Use it for learning, development and lightweight local services,
> not production workloads.

Mobdesk uses Termux as both the Android integration layer and the sole
development environment:

```text
Android
  Termux -> Mobdesk -> local shell or SSH on port 8022
```

The project does not require root, a virtual machine, Docker on the phone,
systemd, a graphical desktop or kernel modules.

## What is available

- repeatable setup of Termux, SSH, networking and workspace state;
- a dedicated Mobdesk SSH server on port `8022`;
- local Termux shell access with `mobdesk shell`;
- human-readable and JSON output for automation and the TUI;
- idempotent `pkg` installation of Git, Neovim, tmux, Go, Python, Node.js,
  Clang/C++, Lua, GitHub CLI, tree, htop, ncdu and Micro;
- curated private user CLI profiles for TUIFI, Bitwarden CLI and Resterm;
- Docker-validated offline Git, Go, Python, Node/npm, C, C++, Lua, Neovim and
  tmux workflows in the Termux workspace;
- status, setup, tools, shell, system and update screens in the TUI;
- verified binary updates with rollback and recovery handling;
- English (`en-US`) and Brazilian Portuguese (`pt-BR`) presentation.

Projects, persistent sessions, services and a web interface remain outside the
current MVP and belong to later roadmap stages.

Application configuration, including Neovim/LazyVim, is deferred for the first
sprint. Managed JVM and Spring Boot tooling are also not current sprint scope.

## Requirements

- Android phone with ARM64 architecture;
- Termux from a trusted source, preferably [F-Droid](https://f-droid.org/packages/com.termux/)
  or the [official releases](https://github.com/termux/termux-app/releases);
- storage for Termux, projects and installed tools; Mobdesk warns below
  20 GB free and blocks new installations below 10 GB;
- a trusted local network if connecting from another computer.

Mobdesk does not require root. Performance and reliability depend on the
phone's memory, temperature, battery and Android/HyperOS background-process
policies.

## Installation

Install Termux from a trusted source before choosing one of the installation
methods below. The recommended method uses the published ARM64 binary and does
not require Go.

### Option 1: released ARM64 binary

The latest stable binary is available through the
[releases page](https://github.com/ericklucioh/mobdesk/releases). In Termux,
run:

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

The checksum verifies file integrity. Release signatures are not available
yet, so this does not independently authenticate a release. The first setup
installs the required Termux packages, creates the persistent workspace,
configures SSH and installs the `mobdesk` launcher.

### Existing PRoot installations

The Termux-only first sprint does not migrate PRoot-Distro or Ubuntu
installations. Back up anything you need, perform a full Termux reset, then
install Mobdesk fresh. Do not attempt an in-place upgrade or migration.

### Option 2: build with Go

Use this method when you want to build from the latest stable Go module tag.
The project currently requires Go `1.26.5` or newer:

```bash
pkg update
pkg install -y golang git
go version
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
~/go/bin/mobdesk setup
```

`@latest` means the latest stable semantic-version tag; it does not mean the
latest untagged commit or a `test-v*` prerelease. For a reproducible
installation, pin the version explicitly, for example:

```bash
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@v0.6.0
```

Running setup through `~/go/bin/mobdesk` on the first run is intentional. Setup
creates the Termux launcher at `$PREFIX/bin/mobdesk`, so subsequent commands
can use `mobdesk` normally. Setup can be run again after an interruption
without deleting the workspace.

## Basic use

Start and inspect the workstation:

```bash
mobdesk status
mobdesk start
```

`mobdesk start` starts the SSH server and displays the connection details. To
open Termux locally without SSH, use:

```bash
mobdesk shell
```

To connect from another computer on the same trusted network, use the SSH
command shown by Mobdesk, for example:

```bash
ssh -p 8022 android@192.168.1.50
```

Stop the SSH server when you are done:

```bash
mobdesk stop
```

Use `mobdesk status --json` for automation. Use `mobdesk logs --name <name>` to
read a bounded installation log when a tool installation fails.

## Security and network use

The current MVP uses password authentication and listens on the local network
to support SSH from another computer. Use it only on a trusted network or
through a private tunnel. Never expose port `8022` directly to the public
internet. Keep important projects backed up outside the phone.

Binary updates verify SHA-256 integrity, but release authenticity is not yet
independently signed. Treat downloaded releases as experimental until release
signing is implemented.

See the [security policy](SECURITY.md) for private vulnerability reports.

## TUI

Run `mobdesk tui` in Termux. The TUI provides visible actions for status, setup,
tools, shell and system updates. Important actions work with touch/mouse and
keyboard. `Tab` changes focus, `Enter` activates an action, `Esc` goes back and
`q` starts the quit confirmation.

The TUI can also be opened through SSH. That session uses the same Termux
workstation, not a separate Ubuntu or PRoot environment.

### Language

Mobdesk uses English (`en-US`) by default and currently supports Brazilian
Portuguese (`pt-BR`). Select the language when starting the CLI or TUI:

```bash
mobdesk tui --locale pt-BR
mobdesk --locale en-US status
```

You can also set the language with `MOBDESK_LOCALE`:

```bash
MOBDESK_LOCALE=pt-BR mobdesk tui
```

The selected locale is passed to TUI operations and child commands. The TUI
does not yet have an in-app language button, so restart it with the desired
locale to change languages. Technical identifiers such as commands, flags,
JSON keys and state values remain in English.

## Limitations

- Docker, systemd, kernel modules and privileged device access are not available
  through Mobdesk;
- Android may suspend or terminate Termux; exempt Termux from battery
  optimization when appropriate;
- Docker validation does not reproduce Android permissions, battery behavior,
  wake-lock behavior, HyperOS process suspension or a real ARM64 device;
- the project is not intended for heavy production workloads or public SSH
  exposure.

## Development

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
make build-image
make check
```

Before contributing, read [CONTRIBUTING](.github/CONTRIBUTING.md). The
repository uses Docker for repeatable checks, but Termux/SSH changes also
require validation on real Termux when a device is available.

## Documentation and community

- [Landing page and GitHub Pages deployment](docs/GITHUB-PAGES.md)
- [Mission](docs/MISSION.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Decisions](docs/DECISIONS.md)
- [Roadmap](docs/ROADMAP.md)
- [Superseded app configuration refactor plan](docs/APP-CONFIGURATION-REFACTOR-PLAN.md)
- [Superseded app configuration implementation plan](docs/APP-CONFIGURATION-IMPLEMENTATION-PLAN.md)
- [Next features plan](docs/PLAN-NEXT-FEATURES.md)
- [Tool catalog](docs/ideas/TOOL-CATALOG.md)
- [Changelog](CHANGELOG.md)
- [Contributing](.github/CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Support](.github/SUPPORT.md)
- [Security policy](SECURITY.md)
- [Portuguese README](README.pt-BR.md)

## License

Mobdesk is distributed under the [MIT license](LICENSE).
