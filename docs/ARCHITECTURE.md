# Mobdesk Architecture

This document records the technical foundation: where software runs, how
layers relate, which execution boundaries exist, and which limitations come
from Android and PRoot. Business rules, app catalogues, roadmap material and
product decisions are documented separately.

## Execution topology

```text
Android / HyperOS
└── Termux
    ├── Mobdesk Go binary
    ├── host tools and processes
    └── PRoot-Distro
        └── persistent Ubuntu ARM64
            └── user Linux processes
```

Android supplies the kernel, network, storage, battery and suspension policy.
HyperOS may terminate or restrict Termux while a process is running.

Termux is the Mobdesk host. It provides the Go runtime, native packages and
PATH, Android-facing commands, Mobdesk-managed SSH, PRoot entry, and private
application state. Ubuntu is a persistent Linux userland with its own
filesystem, libraries and distribution tools, but it has no separate kernel.

PRoot is not a VM, a genuinely isolated container or an independent operating
system. Processes remain subject to Android's kernel and policies.

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
parses arguments and coordinates services. Install, uninstall and
`config apply/remove` share app services and provide human output or schema 1
JSON; requested progress uses separate JSON events.

Services must not depend on TUI rendering. Human and JSON output are
presentation-layer responsibilities.

The TUI uses Bubble Tea, Bubbles and Lip Gloss. It does not duplicate
installation, collection, update or safety rules. The real backend consumes the
CLI's JSON contract; the mock backend implements the same interface for visual
scenarios. App popups keep presentation state only and turn actions into CLI
operations. They never execute Ubuntu commands directly.

When launched inside a Mobdesk SSH session, the TUI runs in Ubuntu/PRoot rather
than Termux. It can show the workspace but cannot inspect or control Termux,
`sshd` or `proot-distro`. It must identify remote mode and block host actions:
setup, start, stop, tool installation and binary update.

## Internal services

- `internal/status` collects a shared environment snapshot and reconciles
  installations and configurations;
- `internal/install` resolves profiles, performs idempotent installation,
  applies embedded configuration and writes records;
- `internal/update` checks and applies Mobdesk updates;
- `internal/logs` reads persisted records without owning a separate screen;
- `internal/version` provides binary metadata.

Create new layers only when real behavior requires separation.

## Execution boundaries

```text
Mobdesk in Termux
    ├── host commands
    │   └── pkg, sshd and Termux tools
    └── Ubuntu commands
        └── proot-distro login ubuntu -- ...
```

Every command identifies its target environment. The application must not treat
an Ubuntu process as a native Termux process. Simple processes use `os/exec`
with context and cancellation. Human setup and tool installation, shells and
editors require PTY input, output forwarding and safe terminal restoration so
package-manager prompts remain answerable. The TUI suspends through
`tea.ExecProcess` instead of writing Ubuntu commands itself. JSON and progress
operations remain headless and use deterministic package configuration
defaults. Setup synchronizes Android's validated timezone into Ubuntu, and APT
operations repair pending `dpkg` configuration before changing packages. User
input is never concatenated into commands without validation.

## State and storage

Persistent Mobdesk state lives in private user directories in Termux and may
contain setup state, installation records, operation logs and SSH files.

- private files have restrictive permissions;
- secrets never enter code, Git or logs;
- host state and Ubuntu state remain distinguishable;
- Termux and Ubuntu dependencies are not mixed;
- projects and user data survive repeated operations;
- Android external storage is not assumed to be a complete Unix filesystem.

User-profile tools such as Zellij live in `$HOME/.local/bin`; generated shell
configuration adds that path. Control and installation commands remain Termux
operations. Mobdesk also supplies a Ubuntu shell launcher through `$SHELL` so
Zellij panels retain the main shell configuration.

The generated Ubuntu shell exports `CGO_ENABLED=0`. Go-based catalog tools also
receive the variable explicitly during installation so they do not depend on a
a C compiler and development headers that may be absent from the PRoot userland.

New catalog apps must declare the same presentation contract as existing apps:
localized description, concise usage, installation profile and storage
estimate. Optional configuration and dependencies are rendered only when they
apply to the current state.

## JVM toolchain

Java 21 is installed and verified inside Ubuntu through the declared PRoot
boundary. The generated shell configuration discovers the resolved `javac`
path, exports `JAVA_HOME` and prepends its `bin` directory without importing
Termux environment variables. Kotlin/JVM 2.2.20 and Gradle 8.14.3 use pinned
official archives and checksums; Maven remains an independent Ubuntu APT
profile. All three build tools inherit the same `JAVA_HOME`.

Installation records retain required executables and status reconciliation can
report missing executables or dependencies as `partial`. The TUI consumes those
fields and does not offer storage-blocked installation actions. Project
wrappers are selected before global `gradle` or `mvn` commands and are never
rewritten.

## Layer contracts

CLI commands are the public execution boundary. The real TUI backend expects a
final result, normally JSON on stdout, on both success and failure. Auxiliary
messages must not corrupt JSON.

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

Do not assume systemd, namespaces, cgroups, seccomp, kernel modules, real
Docker, privileged Android devices, guaranteed graphics acceleration, continued
execution after HyperOS suspension, or heavy production performance.

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

Configuration profiles may declare plugins only with HTTPS repositories, fixed
revisions and paths inside the Ubuntu HOME. The engine performs clone and
checkout through `proot-distro`; network content is not treated as a script.
Modified checkouts are preserved on removal.

## Verification

```bash
make check
```

Local tests validate logic and contracts. Final Android, Termux, PRoot and
HyperOS integration requires a real device.
