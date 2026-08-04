# Mobdesk Decisions

This document records current decisions, deferred alternatives and hypotheses
that still require validation. Closed decisions should not be reopened during
implementation without an explicit scope change.

## Current decisions

### Ubuntu through PRoot is the primary environment

Persistent Ubuntu prioritizes compatibility with conventional Linux tools,
`apt`, glibc, standard paths and common dependencies.

### Termux is the host

Termux owns Android integration, networking, startup, wake-lock, SSH, PRoot and
Termux:API. It is not the user's primary development userland.

### Go is the control-plane language

Go fits a single binary, process execution, concurrency, logging, ARM64
distribution and a shared CLI/TUI.

### TUI precedes a web interface

The first product works in a terminal. A web interface or APK comes only after
installation and the workflow are proven.

### Installation is guided and idempotent

Users should not need to know `pkg`, `proot-distro`, `apt`, mounts or scripts.
The desired flow is install Mobdesk, run `mobdesk start`, and select tools in
the TUI. Ubuntu, tools, projects and configuration persist across runs.

### App presentation uses curated metadata

The app catalog owns concise user-facing descriptions and usage forms. Runtime
verification may collect a version, but raw command output is never a TUI
metadata field; commands such as `--help` are not displayed as versions. New
profiles follow the same metadata contract so the tools screen remains
consistent across apps.

### The TUI remains in control after start

`mobdesk start` starts the workstation without automatically opening a shell.
The TUI suspends the terminal only when the user explicitly opens the shell.

### The TUI respects the Termux/Ubuntu boundary

The status identifies Termux host mode versus an SSH session inside Ubuntu. In
remote Ubuntu mode, SSH, PRoot, setup, installation and binary update actions
are unavailable; workspace information and the local Ubuntu shell remain
available.

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

### Ubuntu timezone follows Android

Setup reads the Android timezone, validates it against Ubuntu's zoneinfo
database and persists `/etc/localtime` and `/etc/timezone`; repeated setup runs
reconcile changes. Non-interactive APT operations rely on that persisted value
instead of forcing UTC.

### APT repairs interrupted dpkg state first

APT operations run `dpkg --configure -a` before changing packages. This is a
safe, idempotent repair for interrupted package configuration; broader
dependency repair is not performed automatically. Interactive PTY stdin
forwarding is cancellable and must finish before the terminal is restored.

### Go tools default to pure Go builds

The generated Ubuntu shell exports `CGO_ENABLED=0`, and Mobdesk also sets it
explicitly for Go-based tool installs. This avoids requiring a C toolchain and
headers in the PRoot userland; tools that genuinely require CGO remain outside
the current default installation profile.

### LazyVim is an optional versioned profile

LazyVim is separate from Neovim installation. Embedded files and fixed plugin
revisions are used, existing configuration is refused, and manually modified
plugin checkouts are preserved on removal.

### Updates preserve an executable version

The updater uses a private temporary file, verifies SHA-256 and runs
`version --json` before activation. The previous version can be recovered after
an interrupted replacement. SHA-256 detects corruption but is not independent
release authenticity.

### Setup is safely repeatable

Concurrent setup is serialized, state is private, and a failed setup can be
repeated without deleting Ubuntu, the workspace or projects.

## Deferred alternatives

- Termux-native development as the primary runtime may be faster but has glibc,
  manylinux, path and native-dependency compatibility costs.
- Nix-on-Droid may help declarative configuration and rollback, but is outside
  the core because of complexity, storage and diagnostic cost.
- A graphical desktop through X11, VNC or a full desktop increases consumption,
  latency and failure modes; TUI and later individual web applications remain
  preferred.
- Neko is an experimental remote-browser line and depends on validating
  Firefox, Xorg, PulseAudio, GStreamer, WebRTC and PRoot on real hardware.
- PRoot is not Docker. Real Docker, VM and kernel features remain outside the
  educational MVP.

Remote-browser research and the current baseline are recorded in
[`docs/ideas/REMOTE-BROWSER-RESEARCH.md`](ideas/REMOTE-BROWSER-RESEARCH.md).

## Hypotheses to validate

- Ubuntu ARM64 through PRoot is acceptable for classes and small projects;
- the complete installation fits comfortably on the POCO;
- HyperOS keeps Termux and Ubuntu active during use;
- the TUI is comfortable on a virtual keyboard and SSH;
- selected tools are available or installable in Ubuntu;
- the Termux/Ubuntu project mount preserves the intended workflow;
- Mobdesk can follow processes and sessions without losing output;
- Termux:Widget shortcuts are sufficient before an APK exists.

## Security rules

- never expose SSH directly to the internet;
- prefer SSH keys;
- use Tailscale for external access;
- keep development servers on localhost when possible;
- preserve authentication in future applications;
- never store passwords in the repository;
- confirm before deleting Ubuntu, profiles or projects;
- keep backups outside the phone.
