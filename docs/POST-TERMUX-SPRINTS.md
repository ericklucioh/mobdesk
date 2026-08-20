# Post-Termux Sprint Plan

**Status:** proposed

This plan extends the native Termux workstation after the Termux-only
foundation is accepted on a reset ARM64 device. It does not restore PRoot,
Ubuntu, APT, `/root`, systemd, or a second runtime.

## Delivery Order

| Order | Sprint | Outcome |
|---|---|---|
| Gate | ARM64 acceptance | Verify the current native workstation on the POCO F6 |
| 4 | Native JVM | Install and configure OpenJDK 21 for Java development |
| 5 | Spring Boot API | Build and run a local Spring Boot API |
| 6 | Doctor | Diagnose the native workstation without changing it |
| 7 | Projects | Manage workspace-contained project metadata safely |
| 8 | React and Vite | Create and validate a React/Vite project |
| 9 | Services | Manage project processes through identified tmux sessions |
| 10 | User CLIs | Add safe npm-global and pipx strategies for curated profiles |
| 11 | LazyVim | Offer optional, safe Neovim configuration |

The order is intentional. Java must work before Spring, a project must exist
before it owns a service, and services must be stable before adding more
developer tooling or editor configuration.

## Gate: ARM64 Device Acceptance

The current foundation is not complete until it is exercised on a clean POCO F6
Termux installation.

1. Reset Termux and do not install PRoot-Distro or Ubuntu.
2. Run `mobdesk setup` twice and confirm the native workspace and shell setup.
3. Install, reinstall and exercise every active catalog profile.
4. Verify `status --json`, the TUI, SSH, workspace behavior and safe uninstall.
5. Verify SSH reconnection and basic wake-lock/screen-lock behavior.
6. Record installed versions, storage use and ARM64-specific failures.

No subsequent profile is considered supported until it has equivalent device
evidence.

## Sprint 4: Native JVM

### Goal

Provide a native Java 21 JDK that can compile and run programs in
`$HOME/workspace`.

### Scope

- Audit `openjdk-21` in the Termux fixture and on the POCO F6.
- Add a curated `java` profile backed by `pkg install openjdk-21`.
- Require `java`, `javac` and `jar` to be available.
- Verify `java --version`, `javac --version` and `jar --version`.
- Discover the runtime `java.home` after installation and accept it only when
  it is an absolute path below `$PREFIX`.
- Record the JDK version, required executables and discovered Java home in
  private Mobdesk state.
- Add Java toolchain information to `status --json` as an additive field.
- Activate the Java hello fixture: compile a class, run it, build a JAR and run
  the JAR from the workspace.
- Cover install, reinstall, status reconciliation, missing executables,
  cancellation, storage blocking and safe uninstall.

### JVM Environment Policy

Mobdesk must not set a static `JAVA_HOME` in `.bashrc`. Package-managed Java
paths can change during upgrades. Instead, Mobdesk resolves Java home from the
installed runtime and supplies `JAVA_HOME` only to a managed Maven, Gradle or
Java service child process that needs it. Users retain control of their own
interactive shell configuration.

### Out of Scope

- Maven, Gradle, Kotlin and Spring Boot;
- downloaded JDK distributions;
- Java paths outside Termux `$PREFIX`;
- global shell environment changes.

### Acceptance

The Java fixture compiles and runs in Docker and on the POCO F6, with no PRoot
process, external JDK or permanent Java shell configuration.

## Sprint 5: Spring Boot API

### Goal

Prove a small Spring Boot API can build, start, answer locally and stop without
an orphan process.

### Scope

- Audit the native Termux Maven package. If it is not supportable, audit only a
  project-local executable Maven Wrapper; do not download a Maven distribution
  globally.
- Add a `maven` profile only after its package or wrapper strategy passes the
  native audit and device test.
- Make Maven depend on the Java profile.
- Reuse `ResolveBuildCommand` to prefer an executable project `mvnw`, then use
  the native Maven command when appropriate.
- Add a pinned Spring Boot Maven fixture with a `GET /health` endpoint and a
  simple test.
- Add `make spring-test`: build the fixture, start the generated JAR, wait for
  the endpoint, call it with `curl`, and stop only the process created by the
  test.
- Bound network, build and startup operations with context cancellation and
  useful private logs.
- Measure Maven cache and build storage on the POCO F6.

### Out of Scope

- Gradle, Kotlin, database provisioning and production deployment;
- persistent API management and LAN exposure;
- reuse of the superseded Ubuntu Spring fixtures as evidence.

### Acceptance

A pinned Spring Boot fixture builds and its local health endpoint responds on a
clean Termux ARM64 device. The test cleans up its child process reliably.

## Sprint 6: Doctor

### Goal

Explain native workstation problems before project and service management add
more state.

### Command Contract

```text
mobdesk doctor [--json] [--strict]
```

The JSON response preserves `schema_version`, `command`, `success`, `state`
and `message`, then adds a stable list of checks containing an ID, severity,
sanitized evidence and a suggested action.

### Scope

- Create a read-only diagnostic service from the existing status collectors.
- Check Termux markers, `PREFIX`, setup phases, workspace, storage, PATH,
  SSH ownership, wake-lock availability, Termux:API availability and catalog
  executable reconciliation.
- Add Java and Maven checks only when those profiles are installed.
- Use contexts and timeouts for every external probe and continue after an
  individual failure.
- Add a touch-first TUI diagnostics screen with visible close/back controls.

### Out of Scope

- `doctor --fix`;
- writing shell configuration, reinstalling packages or stopping processes;
- logging shell environments, credentials or tokens.

### Acceptance

Doctor reports partial failures without failing to produce valid JSON or
changing the workstation.

## Sprint 7: Projects

### Goal

Add a safe project model rooted exclusively in `$HOME/workspace`.

### Proposed Commands

```text
mobdesk project list --json
mobdesk project inspect <name> --json
mobdesk project create <name> --template <template> --json
```

### Scope

- Validate project names and canonicalize all paths.
- Refuse traversal, paths outside the workspace and symlink escapes.
- Persist private metadata for project name, directory, template, created time
  and detected toolchains.
- Reconcile existing workspace directories without deleting user data.
- Present projects in the TUI with keyboard and touch navigation.
- Keep project removal out of this first project sprint.

### Acceptance

No project command can operate outside the workspace, and repeated operations
preserve project files and metadata.

## Sprint 8: React and Vite

### Goal

Create a React/Vite project through the project model and validate local build
and development startup.

### Scope

- Add a `react-vite` project template, using the already native Node/npm
  profile.
- Run the audited Vite creator with structured arguments, never concatenated
  shell input.
- Pin and record the creator version used for reproducibility.
- Install dependencies only in the project directory.
- Validate `npm install`, `npm run build` and an ephemeral `npm run dev`.
- Verify local loopback access only. LAN binding requires a later explicit,
  security-reviewed option.
- Add cancellation, network failure and cleanup coverage.

### Policy

Vite is a project dev dependency. It is not a global npm profile.

### Acceptance

A React/Vite project can be created, built and served locally from the
workspace on the target device.

## Sprint 9: Managed Services

### Goal

Start, observe, reconnect to and stop project processes without claiming
systemd or Android background-execution guarantees.

### Proposed Commands

```text
mobdesk service start <project> -- <command> [args...]
mobdesk service stop <id> --json
mobdesk service status [--json]
mobdesk service logs <id>
```

### Scope

- Require an explicit argv array after `--`; do not invoke a shell.
- Run each service in an identifiable Mobdesk tmux session.
- Persist project, argv, working directory, session, timestamps and private
  log path.
- Stop only a session or process proven to be owned by Mobdesk.
- Surface exited, failed and partial states through status, doctor and the TUI.
- Support Spring `java -jar` and React `npm run dev` as first workloads.
- Treat `termux-services` as a separately audited experiment, not a dependency
  of service management.

### Acceptance

Mobdesk can start a project service, show its logs, reconnect to it by SSH and
stop it without affecting an unrelated tmux session.

## Sprint 10: Curated npm-global and pipx CLIs

### Goal

Install selected user CLIs without corrupting package-managed paths or removing
user-owned files.

### Scope

- Add explicit native strategies for audited profiles only; never implement an
  arbitrary `npm` or `pipx` installer command.
- Install each npm CLI in a profile-specific user-owned directory and publish
  only verified, Mobdesk-owned links under `$HOME/.local/bin`.
- Keep npm package files out of `$PREFIX` and do not alter the user's global
  npm prefix.
- Bootstrap pipx only after a native Python, pip and virtual-environment audit.
- Use private `PIPX_HOME` and `PIPX_BIN_DIR` paths for Mobdesk-owned profiles.
- Record package version, source, expected executables, files and hashes.
- Refuse executable-name conflicts and preserve modified or unmanaged files on
  removal.
- Audit candidate profiles separately, including native module compilation and
  authentication behavior. Potential candidates include Posting for pipx and
  selected AI CLIs for npm.

### Acceptance

Each approved CLI installs, reinstalls, reports status, cancels safely and
uninstalls without modifying an unrelated npm, pipx or `$HOME/.local/bin` file.

## Sprint 11: LazyVim

### Goal

Offer optional LazyVim setup without taking ownership of an existing Neovim
configuration.

### Scope

- Require audited native Neovim, Git and C compiler versions.
- Require the exact LazyVim Neovim minimum version before offering the action.
- Use `$HOME/.config/nvim` and `$HOME/.local/share/nvim/lazy` only.
- Refuse a pre-existing Neovim configuration rather than overwrite it.
- Install the LazyVim starter at a fixed, verified revision, never `main`.
- Store configuration provenance, revision and file hashes privately.
- Separate initial configuration creation from plugin synchronization.
- Validate native headless startup and cancellation handling.
- Remove only clean, Mobdesk-provenance files; preserve modified configuration
  and plugin files with an explicit report.

### Acceptance

Apply is idempotent, existing configuration is protected, and removal cannot
delete a user modification.

## Common Requirements

Every sprint must:

- keep Termux as the only runtime;
- validate arguments before process construction;
- use contexts and cancellation for network, build and long-running work;
- keep state and logs in private Termux paths;
- preserve JSON command contracts and add fields compatibly;
- provide text, JSON, runtime-error and cancellation tests for new commands;
- provide keyboard and mouse/touch equivalents for new TUI actions;
- pass `make check`, relevant Docker workflow tests and `git diff --check`;
- receive a clean ARM64 device validation before being presented as supported.
