# Mobdesk Decisions

This document records current decisions, deferred alternatives and hypotheses
that still require validation. Closed decisions should not be reopened during
implementation without an explicit scope change.

## Current decisions

### Termux is the sole workstation

Termux owns Android integration, networking, startup, wake-lock, SSH,
Termux:API, development tools and the user workspace. It is the only active
development userland; PRoot-Distro and Ubuntu are removed.

Existing PRoot-based installations have no migration path. Users must perform a
full Termux reset and install the Termux-only version fresh.

### Go is the control-plane language

Go fits a single binary, process execution, concurrency, logging, ARM64
distribution and a shared CLI/TUI.

### TUI precedes a web interface

The first product works in a terminal. A web interface or APK comes only after
installation and the workflow are proven.

### Installation is guided and idempotent

Users should not need to know `pkg`, scripts or internal paths.
The desired flow is install Mobdesk, run `mobdesk start`, and select tools in
the TUI. Termux tools, projects and state persist across runs.

### App presentation uses curated metadata

The app catalog owns concise user-facing descriptions and usage forms. Runtime
verification may collect a version, but raw command output is never a TUI
metadata field; commands such as `--help` are not displayed as versions. New
profiles follow the same metadata contract so the tools screen remains
consistent across apps.

### Core profiles require workflow validation

Catalog installation and version checks are not sufficient evidence that a
developer tool works. Core profiles receive a small offline workflow test in a
clean Termux fixture, executed from the Mobdesk workspace. The Node profile
requires both `node` and `npm` from its native package; dependency installation,
application configuration and framework setup remain deferred.

### The TUI remains in control after start

`mobdesk start` starts the workstation without automatically opening a shell.
The TUI suspends the terminal only when the user explicitly opens the shell.

### The TUI uses the Termux boundary

The TUI runs against the Termux workstation. SSH connects to the same Termux
environment; there is no Ubuntu remote mode or PRoot boundary to manage.

### Operations have one JSON result

`setup`, `start` and `stop` provide one structured result on stdout with
`--json`. The TUI consumes that contract and stderr remains for auxiliary
messages.

### Human installations own the terminal

Text-mode setup and tool installation run through a PTY and keep command output
visible so users can answer package-manager and installer prompts. The TUI
suspends temporarily through `tea.ExecProcess` for these operations and resumes
after the command exits. JSON and progress modes remain non-interactive and
receive deterministic package configuration defaults instead of reading stdin.

### Application configuration is deferred

Application configuration profiles are not part of the Termux-only first
sprint. Neovim configuration and LazyVim are deferred rather than installed or
managed by Mobdesk.

### Updates preserve an executable version

The updater uses a private temporary file, verifies SHA-256 and runs
`version --json` before activation. The previous version can be recovered after
an interrupted replacement. SHA-256 detects corruption but is not independent
release authenticity.

### Setup is safely repeatable

Concurrent setup is serialized, state is private, and a failed setup can be
repeated without deleting the Termux workspace or projects.

### Storage thresholds are global

Installations warn below 20 GB of free storage and block new changes below
10 GB. The rule is applied before package changes, downloads or removals and is
visible in CLI, JSON, status and TUI flows.

## Superseded decisions

The following decisions describe the retired PRoot/Ubuntu product line. They
are retained for release history and must not guide Termux-only work.

- Ubuntu through PRoot as the primary environment, including its timezone,
  `dpkg` recovery, generated shell and JVM ownership.
- PRoot-specific pure-Go defaults, project-wrapper behavior and Spring fixtures.
- Versioned LazyVim profiles and other Mobdesk-managed application
  configuration.
- The Termux/Ubuntu remote-mode boundary and Ubuntu-owned state.

## Deferred alternatives

- PRoot/Ubuntu development is superseded, not an active alternative in this
  sprint. Reintroduction requires a new explicit decision and migration design.
- Nix-on-Droid may help declarative configuration and rollback, but is outside
  the core because of complexity, storage and diagnostic cost.
- A graphical desktop through X11, VNC or a full desktop increases consumption,
  latency and failure modes; TUI and later individual web applications remain
  preferred.
- Neko is an experimental remote-browser line outside this sprint. Any future
  evaluation requires a new Termux-native feasibility decision.
- Real Docker, VM and kernel features remain outside the educational MVP.

Remote-browser research and the current baseline are recorded in
[`docs/ideas/REMOTE-BROWSER-RESEARCH.md`](ideas/REMOTE-BROWSER-RESEARCH.md).

## Hypotheses to validate

- Termux-native tools are acceptable for classes and small projects;
- the complete Termux-only installation fits comfortably on the POCO;
- HyperOS keeps Termux active during use;
- the TUI is comfortable on a virtual keyboard and SSH;
- selected tools are available or installable in Termux;
- Mobdesk can follow processes and sessions without losing output;
- Termux:Widget shortcuts are sufficient before an APK exists.

## Security rules

- never expose SSH directly to the internet;
- prefer SSH keys;
- use Tailscale for external access;
- keep development servers on localhost when possible;
- preserve authentication in future applications;
- never store passwords in the repository;
- confirm before deleting Termux-managed profiles or projects;
- keep backups outside the phone.
