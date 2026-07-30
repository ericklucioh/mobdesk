# Mobdesk

Turn an Android phone into a personal Ubuntu development workstation.

Mobdesk uses Termux as the Android control host and persistent Ubuntu through
PRoot-Distro as the development environment. The supported MVP flow is:

```text
Termux -> Mobdesk -> SSH -> Ubuntu via PRoot
```

The project is intentionally small. It does not require root, a virtual
machine, Docker on the phone, systemd, a graphical desktop, or kernel modules.

## Current MVP

- repeatable setup of Ubuntu, SSH, networking and workspace state;
- dedicated Mobdesk SSH server on port `8022`;
- `setup`, `start`, `stop`, `shell`, `status`, `install`, `update`, `version`
  and `tui` commands;
- idempotent installation of Go, Python, Node.js, C, C++ and Lua profiles;
- human-readable and JSON output for automation and the TUI;
- status, setup, tools, shell and update screens in the TUI;
- verified binary updates with rollback/recovery handling;
- local use on the phone and remote access through SSH.

Projects, persistent sessions, services and a web interface remain outside the
current MVP and belong to later roadmap stages.

## Quick start

Install Termux from a trusted source, then run:

```bash
pkg update
pkg install -y golang git
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
~/go/bin/mobdesk setup
mobdesk start
```

Use `mobdesk status` for a read-only environment check, or
`mobdesk status --json` for automation. The first setup creates the persistent
Ubuntu environment, configures SSH, and prepares direct access to Ubuntu.

Open the local shell with:

```bash
mobdesk shell
```

To connect from another computer on the same trusted network, use the SSH
command shown by Mobdesk, for example:

```bash
ssh -p 8022 android@192.168.1.50
```

Do not expose port `8022` directly to the public internet. Prefer a trusted
local network, a private network such as Tailscale, or an SSH tunnel.

## TUI and runtime boundary

Run `mobdesk tui` in Termux. The TUI can also be opened through an SSH session,
where it is running inside Ubuntu. In that remote mode it can show the
workspace and open the local shell, but host-only actions such as setup, SSH
control, installation and binary updates are blocked and explained.

Tools installed by Mobdesk belong to Ubuntu. Run control and installation
commands in Termux; use the installed tools directly inside Ubuntu.

## Development

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
make build-image
make check
```

The Docker fixture validates project logic and a simulated Termux/Ubuntu
userland. It does not reproduce Android permissions, battery policies,
wake-lock behavior, HyperOS process suspension, or a real ARM64 device.

See [CONTRIBUTING](.github/CONTRIBUTING.md) before submitting changes.

## Documentation

- [Mission](docs/MISSION.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Decisions](docs/DECISIONS.md)
- [Roadmap](docs/ROADMAP.md)
- [App configuration refactor plan](docs/APP-CONFIGURATION-REFACTOR-PLAN.md)
- [App configuration implementation plan](docs/APP-CONFIGURATION-IMPLEMENTATION-PLAN.md)
- [Tool catalog](docs/ideas/TOOL-CATALOG.md)
- [Security policy](SECURITY.md)
- [Portuguese README](README.pt-BR.md)

## License

Mobdesk is distributed under the [MIT license](LICENSE).
