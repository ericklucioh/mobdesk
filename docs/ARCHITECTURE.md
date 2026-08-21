# Mobdesk Architecture

This document records the technical foundation: where software runs, how
layers relate, which execution boundaries exist, and which limitations come
from Android and Termux. Business rules, app catalogues, roadmap material and
product decisions are documented separately.

## Execution topology

```text
Android / HyperOS
└── Termux
    ├── Mobdesk Go binary
    ├── development tools and processes
    ├── workspace and private Mobdesk state
    └── Mobdesk-managed SSH
```

Android supplies the kernel, network, storage, battery and suspension policy.
HyperOS may terminate or restrict Termux while a process is running.

Termux is the Mobdesk host and sole development userland. It provides the Go
runtime, native packages and PATH, Android-facing commands, Mobdesk-managed
SSH, the workspace and private application state. PRoot-Distro and Ubuntu are
removed from the active architecture.

Existing PRoot-based installations cannot be migrated to this layout. A full
Termux reset and fresh installation are required.

## Code layers

```text
cmd/mobdesk
└── internal/cobra       CLI entry and commands
    ├── internal/status  state collection and model
    ├── internal/install app profile installation
    ├── internal/update  update lookup and application
    ├── internal/logs    operation log reading
    └── internal/version binary metadata

internal/tui             Bubble Tea presentation
└── communication backend
    ├── real backend: invokes the executable and reads JSON
    └── mock backend: simulates responses for visual tests
```

App catalog profiles separate user-facing metadata from installation data. A
profile's localized description, concise usage form and optional catalog
version feed the TUI presentation; package, executable, verification arguments,
installation strategy and storage estimates remain operational metadata. The
TUI must never render raw verification output as app metadata, especially when
an app uses `--help` only as an installation check.

`cmd/mobdesk` starts the application. `internal/cobra` registers commands,
parses arguments and coordinates services. Install and uninstall share app
services and provide human output or schema 1 JSON; requested progress uses
separate JSON events. Application configuration, including LazyVim, is deferred
for the first sprint.

Services must not depend on TUI rendering. Human and JSON output are
presentation-layer responsibilities.

The TUI uses Bubble Tea, Bubbles and Lip Gloss. It does not duplicate
installation, collection, update or safety rules. The real backend consumes the
CLI's JSON contract; the mock backend implements the same interface for visual
scenarios. App popups keep presentation state only and turn actions into CLI
operations. They never execute package-manager commands directly.

SSH sessions run in the same Termux workstation. The TUI therefore has one
runtime model rather than a separate Ubuntu/PRoot remote mode.

## Internal services

- `internal/status` collects a shared environment snapshot and reconciles
  installations;
- `internal/install` resolves profiles, performs idempotent installation and
  writes records;
- `internal/update` checks and applies Mobdesk updates;
- `internal/logs` reads persisted records without owning a separate screen;
- `internal/version` provides binary metadata.

Create new layers only when real behavior requires separation.

## Execution boundaries

```text
Mobdesk in Termux
    └── Termux commands
        └── pkg, sshd and development tools
```

Every command targets Termux. Simple processes use `os/exec` with context and
cancellation. Human setup and tool installation, shells and editors require PTY
input, output forwarding and safe terminal restoration so package-manager
prompts remain answerable. The TUI suspends through `tea.ExecProcess` instead of
writing package-manager commands itself. JSON and progress operations remain
headless and use deterministic package configuration defaults. User input is
never concatenated into commands without validation.

## State and storage

Persistent Mobdesk state lives in private user directories in Termux and may
contain setup state, installation records, operation logs and SSH files.

- private files have restrictive permissions;
- secrets never enter code, Git or logs;
- all managed state and dependencies remain in Termux;
- projects and user data survive repeated operations;
- Android external storage is not assumed to be a complete Unix filesystem.

Native profiles remain under `$PREFIX`. Curated user-profile tools publish only
Mobdesk-owned executable links in `$HOME/.local/bin`; generated shell
configuration adds that path. Their pipx runtime and each profile's environment
remain private under `$HOME/.local/share/mobdesk/tools`. Installation and
removal refuse a pre-existing or replaced link rather than modifying an
unmanaged `$HOME/.local/bin` file. Control and installation commands remain
Termux operations.

New catalog apps must declare the same presentation contract as existing apps:
localized description, concise usage, installation profile and storage
estimate. Application configuration profiles, including LazyVim, are deferred.
The Java profile additionally resolves `java.home` at runtime and records it
only when it is an absolute child of the Termux prefix; it never writes a global
`JAVA_HOME` shell setting. The Maven profile is a native `pkg` profile requiring
Java. A managed Java prerequisite cannot be removed while its Maven record is
installed.

## Layer contracts

CLI commands are the public execution boundary. The real TUI backend expects a
final result, normally JSON on stdout, on both success and failure. Auxiliary
messages must not corrupt JSON.

`status --json`, `logs --json` and `version --json` include the common
`schema_version`, `command`, `success`, `state` and `message` envelope as
additive fields alongside their command-specific data. The TUI validates the
known schema and command before applying a response. Setup and package
installation are PTY handoffs so users can interact with native package prompts;
their post-operation state is reconciled by a validated status response.

```text
TUI event
  -> backend
  -> Cobra command or mock
  -> internal service
  -> final result
  -> Bubble Tea message
  -> rendered state
```

The real flow does not depend on progress streaming or continuous polling. The
TUI runs at most one host operation at a time. Setup and tool installation are
explicit terminal handoffs; the TUI waits outside the alternate screen and
refreshes status after the child exits. Operations and status snapshots carry
monotonic IDs so stale responses are discarded, and backend subprocesses are
cancelled when the TUI exits.

## Platform limitations

Do not assume systemd, cgroups, seccomp, kernel modules, real Docker, privileged
Android devices, guaranteed graphics acceleration, continued execution after
HyperOS suspension, or heavy production performance.

These limitations must remain visible instead of being hidden behind an
abstraction that promises unavailable capabilities.

## Structural security

- SSH controls only the Mobdesk instance;
- unrelated ports and processes are not stopped without ownership proof;
- remote access prefers a local network or secure tunnel;
- configuration, state and credentials remain private;
- destructive operations require explicit confirmation;
- long operations accept context and cancellation;
- partial failures do not delete data or orphan processes.

Application configuration profiles and plugin management are deferred. In
particular, LazyVim is not installed or managed during the first sprint.

## Verification

```bash
make check
```

Local tests validate logic and contracts. Final Android, Termux and HyperOS
integration requires a real device.

`make catalog-test-fast` validates native package installation, catalog records
and safe removal. `make workflow-test` runs the tracked offline fixtures from
`~/workspace` to validate the installed development tools. Both use Docker only
as a repeatable Termux fixture and do not replace ARM64 device validation.

`make spring-test` is a separate, bounded networked workflow: it installs the
native Maven profile, builds and tests a pinned Spring Boot fixture, verifies a
loopback-only health endpoint and stops the child JAR process. Maven dependency
download, cache size and Android behavior still require POCO F6 validation.

`make user-cli-test` validates the curated pipx profile in a clean Termux
fixture: install, idempotent reinstall, executable, `status --json` and safe
uninstall. It does not replace ARM64 device validation.

## Superseded architecture

The former Termux-host plus Ubuntu-through-PRoot topology, its `proot-distro`
execution boundary, Ubuntu-owned JVM toolchain and configuration-profile engine
are retained only as historical context. They are not part of the active
architecture.
