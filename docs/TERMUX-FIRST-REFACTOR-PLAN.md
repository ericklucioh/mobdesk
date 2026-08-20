# Termux-First Refactor Plan

**Status:** proposed implementation plan

**Decision:** Mobdesk is a native Termux development workstation. PRoot-Distro,
Ubuntu, guests, runtime selection and compatibility fallback do not exist in the
target product.

**Precondition:** the Termux installation will be reset completely before this
work is used. The reset removes the previous Mobdesk state, PRoot rootfs,
Ubuntu workspace, caches, installation records and configuration. This plan
therefore has no migration, compatibility or cleanup path for the former
Ubuntu-first product.

**Objective:** make `mobdesk setup` produce a useful, persistent Termux
workstation with native SSH, `$HOME/workspace`, curated native tools and a TUI
that manages only the Termux environment.

## Sprint 1 Implementation Scope

The first implementation run delivers the native workstation foundation. It
does not attempt to port every historical profile.

Included:

- native setup, workspace, shell and SSH;
- status schema version 2 and native TUI presentation;
- direct `pkg` installation and removal for Git, Neovim, tmux, Go, Python,
  Node.js, C and C++;
- removal of PRoot, Ubuntu, app configuration and LazyVim code;
- native Docker integration coverage.

Deferred:

- Java, Kotlin, Gradle, Maven and all other unverified profiles;
- managed Go, npm, pipx and download installation strategies;
- app configuration and LazyVim;
- native fast/full catalog validation until profiles are audited on Termux
  ARM64.

## Sprint 2 Implementation Scope

Sprint 2 expands the native package-manager catalog without adding a second
installation strategy. It adds Lua (`lua54` package and `lua` executable),
GitHub CLI, tree, htop, ncdu and Micro. Each profile uses `InstallKind: "pkg"`
and was installed and version-checked in the Termux Docker fixture.

`make catalog-test-fast` now provisions a clean Termux fixture, installs every
active profile, verifies its executable, repeats installation and checks the
persisted status. ARM64 device validation remains required before treating this
catalog as validated for the target phone.

When profiles share a native package, uninstalling one profile releases only
that profile's ownership record and reports `preserved_packages` in JSON. The
package is removed only when the final managed profile is uninstalled.

## Sprint 3 Implementation Scope

Sprint 3 validates the native developer workflows already promised by the
catalog. `make workflow-test` provisions a clean fixture, copies the tracked
hello programs into `~/workspace`, and executes local Git, Go, Python, Node/npm,
C, C++, Lua, Neovim and tmux workflows. The test does not download project
dependencies or configure applications.

The Node profile requires both `node` and `npm`, which the audited `nodejs`
Termux package supplies. SSH smoke validation also proves that remote commands
run in the same Termux home, prefix and workspace. This is a Stage 1 validation
sprint, not the roadmap's persistent-environment Stage 3.

## 1. Product Boundary

### 1.1 Target topology

```text
Android / HyperOS
└── Termux
    ├── Mobdesk
    ├── sshd
    ├── $HOME/workspace
    ├── $HOME/.config/mobdesk
    ├── $HOME/.local/share/mobdesk
    ├── pkg
    └── native development tools
```

Termux is both the Android integration layer and the user's development
environment. SSH starts a native Termux session. The workspace, toolchains,
configuration and Mobdesk state all belong to the same userland.

### 1.2 Explicitly excluded

- `proot-distro` is not installed or invoked by Mobdesk.
- Ubuntu is not installed, detected, configured, shown or documented as an
  available environment.
- No `guest` command exists.
- No `runtime` type, `--runtime` flag, target selection or fallback policy
  exists.
- No PRoot workspace bind, shared home or `/workspace` path exists.
- No legacy Ubuntu state is migrated, reconciled, archived or deleted by
  Mobdesk. The external Termux reset is responsible for removing it.
- A tool that cannot be verified as working natively in Termux is unavailable
  from the catalog. It is not redirected to a Linux compatibility layer.
- Docker, systemd, kernel features, graphical desktop and production workloads
  remain outside the product boundary.

### 1.3 Closed implementation choices

| Topic | Decision |
|---|---|
| Development environment | Native Termux only |
| Persistent project path | `$HOME/workspace` |
| SSH destination | User's normal Termux shell |
| Shell initially supported by setup | Bash |
| Managed user executable path | `$HOME/.local/bin` |
| Termux package executable path | `$PREFIX/bin` |
| Setup package manager | `pkg` |
| Default tool installation policy | Native package or native verified strategy only |
| Installation/configuration state | Termux private Mobdesk state only |
| Status JSON | New documented status schema version 2 |
| Install/config/uninstall JSON | Preserve schema 1 fields and semantics where possible |
| Existing Ubuntu-first environments | Unsupported; reset Termux before use |

## 2. Current-State Inventory

The refactor must remove the following existing Ubuntu-first behavior rather
than merely hiding it.

| Area | Current behavior | Required result |
|---|---|---|
| Paths | `UbuntuWorkspace()` returns `/root/workspace` | `Workspace()` returns `$HOME/workspace` |
| Setup | Installs `proot-distro`, Ubuntu, APT packages and Ubuntu shell | Installs only native Termux prerequisites and creates the workspace |
| SSH | `ForceCommand` wrapper enters Ubuntu | No `ForceCommand`; normal Termux SSH session |
| Start | Requires `proot-distro login ubuntu` to succeed | Requires only completed native setup and configured SSH |
| Shell | `mobdesk shell` opens Ubuntu through a PTY | Opens Bash in `$HOME/workspace` |
| Installer | Every command wraps `proot-distro login ubuntu` | Executes native Termux commands directly |
| Catalog | Package names, scripts and paths assume APT and `/root` | Every retained profile is validated for Termux |
| Config | LazyVim is rooted in `/root/.config/nvim` | Configuration is rooted in `$HOME/.config/nvim` |
| Status | Ubuntu is mandatory for setup health | Workspace is mandatory; no Ubuntu state exists |
| TUI | Remote Ubuntu mode blocks host operations | SSH TUI is native Termux and can manage the workstation |
| Tests | Docker and catalog validation expect PRoot/Ubuntu | Tests exercise the Termux-only contract |

Principal implementation locations before the refactor:

- `internal/paths/paths.go`
- `internal/workstation/setup.go`
- `internal/workstation/ssh_config.go`
- `internal/workstation/start_stop.go`
- `internal/cobra/shell.go`
- `internal/install/install.go`
- `internal/install/config.go`
- `internal/install/uninstall.go`
- `internal/install/lazyvim.go`
- `internal/status/model.go`
- `internal/status/collect.go`
- `internal/tui/screen_status.go`
- `internal/tui/screen_setup.go`
- `internal/tui/screen_shell.go`
- `internal/tui/app_popup.go`
- `internal/tui/tui.go`
- `internal/i18n/message.go`
- `internal/i18n/locale/en-US.json`
- `internal/i18n/locale/pt-BR.json`
- `scripts/test-termux.sh`
- `scripts/test-catalog.sh`
- `docker-compose.yml`
- `Makefile`

## 3. Public Contract After The Refactor

### 3.1 Commands

The supported command surface remains intentionally small:

```text
mobdesk setup [--upgrade-system] [--json]
mobdesk start [--json]
mobdesk stop [--json]
mobdesk status [--json]
mobdesk shell
mobdesk install <tool> [--json] [--progress]
mobdesk uninstall <tool> [--json] [--progress]
mobdesk config apply <tool> [--json] [--progress]
mobdesk config remove <tool> [--json] [--progress]
mobdesk tui
mobdesk logs --name <tool>
mobdesk update [--json]
mobdesk version [--json]
```

Commands that must not be introduced in this work:

```text
mobdesk guest ...
mobdesk runtime ...
mobdesk shell --guest ...
mobdesk install <tool> --runtime ...
```

### 3.2 Setup contract

`mobdesk setup` must be idempotent and leave a native workstation ready without
requiring any rootfs download. Its successful result means all required native
setup phases are complete:

```text
directories
packages-updated
packages-installed
workspace-created
password-configured
ssh-configured
shell-configured
launcher-installed
```

`system-upgraded` remains optional and is complete only when
`--upgrade-system` requested it successfully. It must not determine whether the
workstation is usable.

The minimum managed package installation is:

```text
pkg install -y openssh net-tools
```

The final package set must be confirmed against a reset real Termux device.
`net-tools` remains optional only if network address discovery is rewritten to
use another available native interface.

### 3.3 Shell contract

`mobdesk shell` opens an interactive native Bash session with its working
directory set to `$HOME/workspace`. It must not start another userland,
download anything or alter the current process environment after it exits.

Setup manages a small Bash source file below `$HOME/.config/mobdesk` and an
idempotent, clearly delimited source line in `$HOME/.bashrc`. The source file
may add `$HOME/.local/bin` to `PATH`. It must not replace the user's `.bashrc`,
set a guest-specific `SHELL`, force `CGO_ENABLED=0`, set a custom prompt or
invent Java paths.

The managed block must be safe to repeat. It must be removed only by an
explicit future cleanup feature, not by an unrelated setup failure. If the
existing `.bashrc` cannot be changed safely, setup must return an objective
error instead of overwriting it.

### 3.4 SSH contract

Mobdesk runs a dedicated native `sshd` on port `8022`, using its private
configuration, PID, log and lock paths. It preserves current ownership checks:
Mobdesk must never stop an SSH process it cannot prove uses Mobdesk's config.

The generated `sshd_config` must not contain `ForceCommand`. An SSH connection
therefore reaches the configured Termux account and its normal shell. A client
may also use normal SSH remote commands and SFTP if Termux OpenSSH supports
them.

### 3.5 JSON contract

`setup`, `start`, `stop`, `install`, `uninstall`, `config`, `update` and
`version` retain their documented JSON mode. JSON writes only structured JSON
to stdout. Progress stays newline-delimited JSON events followed by the final
result.

`status --json` changes intentionally from schema version 1 to schema version
2 because it removes the mandatory `ubuntu` object. The schema 2 payload must
include at least:

```json
{
  "schema_version": 2,
  "generated_at": "2026-01-01T00:00:00Z",
  "overall": "healthy",
  "host": {},
  "setup": {},
  "workspace": {},
  "storage": {},
  "ssh": {},
  "network": {},
  "battery": {},
  "wifi": {},
  "installations": [],
  "configurations": [],
  "alerts": {}
}
```

The exact field definitions are finalized in Phase 7. The changelog and both
READMEs must call out this breaking status-schema change.

## 4. Delivery Sequence

The phases below must be implemented in order. Do not start catalog conversion
before the Termux-only execution boundary is complete. Do not update product
claims ahead of the behavior they describe, except for documents explicitly
marked as this plan.

Each completed phase requires:

- focused unit tests for changed behavior;
- `go test` for affected packages;
- `git diff --check`;
- no unrelated formatting or generated-file changes;
- an implementation commit with the proposed message or an equivalent concise
  message.

`make check` is required at the end of every phase that changes Go, catalogs,

## 5. Phase 0: Freeze The Old Direction And Establish A Clean Baseline

**Goal:** make the reset assumption and the replacement architecture explicit
before product code changes begin.

### Step 0.1: Record the replacement plan

- Add this document under `docs/`.
- Link it from the documentation index in both READMEs when implementation
  begins.
- Treat this document as the implementation authority for the refactor.

**Acceptance criteria:**

- The document states that Termux reset is required.
- The document does not promise a compatibility guest or migration.
- The document lists the removal of PRoot as an explicit result, not a deferred
  optimization.

### Step 0.2: Preserve a factual baseline

Before modifying behavior, capture the following from a clean development
fixture and, when available, a real device:

- `mobdesk version --json`;
- `mobdesk status --json`;
- current `make check` result;
- current `make integration-test` result;
- current setup command sequence from unit tests;
- currently supported catalog profile names;
- current SSH configuration ownership and port behavior.

Do not run the old full Ubuntu catalog test solely for this refactor plan. Its
purpose is invalidated by the new architecture and the existing goal explicitly
defers it during planning.

### Step 0.3: Reset the device before product validation

The user resets Termux outside Mobdesk. After reset, install only the minimum
requirements needed to obtain and run Mobdesk. Do not install PRoot-Distro or
Ubuntu manually as a test shortcut.

**Proposed commit:** `docs: add Termux-first refactor plan`

## 6. Phase 1: Replace Architecture And Product Documentation

**Goal:** remove Ubuntu-first claims from active product documentation and make
the Termux-first contract testable and understandable.

### Step 1.1: Update the mission

Edit `docs/MISSION.md`:

- Replace the MVP path `Termux -> Mobdesk -> SSH -> Ubuntu via PRoot` with a
  native Termux SSH entry point.
- Replace the professional-user wording that promises an underlying Ubuntu host
  with native Termux shell and SSH access.
- Keep the user outcome unchanged: phone-only study and small-to-medium work in
  C, JavaScript, React, Java, Go and Python.

### Step 1.2: Replace current architectural decisions

Edit `docs/DECISIONS.md`:

- Replace “Ubuntu through PRoot is the primary environment” with “Termux is
  the primary development environment.”
- Replace “Termux is the host” with a decision that it owns Android integration
  and the user workspace as one environment.
- Remove the deferred Termux-native alternative and the Ubuntu-owned JVM
  decision.
- Add a decision that incompatible software is not supported in the current
  product rather than automatically executed in a guest.
- Retain platform limitations that still apply to Termux and Android.

### Step 1.3: Replace technical architecture

Edit `docs/ARCHITECTURE.md`:

- Remove PRoot from the execution topology, execution boundaries and state
  description.
- State that `pkg`, native commands and Mobdesk run in Termux.
- Replace `/root` configuration examples with `$HOME` and `$PREFIX` examples.
- Define native command execution, cancellation, private state, package-manager
  boundaries and shell configuration ownership.
- Remove claims that a TUI in SSH is necessarily remote Ubuntu mode.
- Document status schema version 2 as part of the contract change.

### Step 1.4: Replace roadmap and public documentation

Edit `docs/ROADMAP.md`, `README.md` and `README.pt-BR.md`:

- Rename the initial stage from Ubuntu bootstrap to Termux bootstrap.
- Update installation storage requirements; do not claim 1.5 GB for Ubuntu.
- Update shell examples so `mobdesk shell` means native Termux workspace.
- Update SSH examples to describe a normal Termux shell.
- Describe only catalog profiles validated for native Termux after Phase 5.
- Remove guest, Ubuntu, PRoot and APT references from active usage sections.
- Keep historical release notes historical, but do not use them as product
  instructions.

### Step 1.5: Supersede the Ubuntu catalog optimization goal

Edit `goal.md` to mark it superseded before its implementation begins. Add a
short link to this plan and explain that its PRoot-session and Ubuntu-cache
optimizations are no longer valid work. Do not rewrite it as if it were
completed.

**Acceptance criteria:**

- Active architecture documents have one unambiguous primary runtime: Termux.
- No active command example tells a new user to install or enter Ubuntu.
- Documentation links pass `make i18n-check`.

**Proposed commit:** `docs: define Termux-first workstation architecture`

## 7. Phase 2: Remove Ubuntu Paths And PRoot Execution From The Core

**Goal:** eliminate the old userland from canonical paths, process helpers and
runtime detection before changing higher-level behavior.

### Step 2.1: Make the workspace a native path

Edit `internal/paths/paths.go`:

- Add `Workspace() string`, returning `filepath.Join(p.Home, "workspace")`.
- Retain current private Mobdesk paths: data, config, state, logs, locks, SSH
  configuration and launcher.
- Remove `UbuntuWorkspace`, `UbuntuConfigDir`, `UbuntuShellConfig`,
  `UbuntuShellLauncher` and `UbuntuDataDir`.
- Add native shell configuration paths only if they have a concrete consumer,
  such as `TermuxShellConfig()` below `ConfigDir()`.

Update `internal/paths/paths_test.go` to assert the new layout and assert that
no path API exposes an Ubuntu root path.

### Step 2.2: Simplify runtime detection

Edit `internal/status/collect.go` and `internal/cobra/runtime.go`:

- Preserve Termux detection through canonical Termux markers and paths.
- Remove the special `/etc/os-release` check for Ubuntu sessions.
- Keep the host-action guard so a binary run from an unsupported non-Termux
  environment still receives an objective Termux-required error.
- Make runtime detection directly testable through injected environment/path
  inputs rather than package-private global process assumptions where practical.

Do not introduce a `Runtime` enum. There is one supported environment.

### Step 2.3: Remove PRoot command construction

- Delete `runUbuntu` and `ensureUbuntu` from `internal/workstation/setup.go`.
- Remove `runUbuntuLogged`, the Ubuntu PATH constant and helpers whose only
  behavior is PRoot command wrapping from `internal/install/install.go`.
- Delete Ubuntu-specific runner behavior in `internal/install/runner.go`.
- Replace direct consumers with native command helpers only after Phase 4
  defines their logging and non-interactive behavior.

This phase may temporarily leave catalog installation unimplemented behind
failing compile-time changes only within the same atomic implementation branch;
the final commit must build and test successfully.

### Step 2.4: Delete obsolete code and tests

Remove or rewrite code and tests that have no Termux-only meaning:

- Ubuntu timezone script and validation usage;
- generated Ubuntu Bash launcher;
- Ubuntu shell environment such as guest `JAVA_HOME` discovery and forced
  `CGO_ENABLED=0`;
- test fixtures that assert `proot-distro login ubuntu` command shapes;
- helpers that validate only `/root` or `/usr/local/bin` paths.

**Acceptance criteria:**

- `rg 'proot-distro|UbuntuWorkspace|runUbuntu|/root/workspace' internal cmd`
  has no production-code matches.
- `Paths.Workspace()` is the only Mobdesk workspace path.
- Unit tests cover positive Termux detection and unsupported-environment
  detection.
- `make check` passes.

**Proposed commit:** `refactor: remove Ubuntu runtime primitives`

## 8. Phase 3: Make Setup, Shell And SSH Native

**Goal:** a clean `mobdesk setup` must produce a useful Termux workstation with
no PRoot dependency.

### Step 3.1: Redefine setup phases

Edit `internal/workstation/setup.go`:

- Remove the `ubuntu-installed` phase.
- Replace the old `workspace-created` implementation with native creation of
  `Paths.Workspace()` using safe directory handling.
- Remove Ubuntu timezone, APT repair, APT update and `bash-completion`
  installation.
- Change native prerequisites to `pkg update` followed by `pkg install -y`
  for the approved setup package list.
- Preserve setup lock, private directories, private marker writes, repeatable
  phases, password setup, SSH setup and launcher setup.
- Do not mark setup complete when a required native phase failed or was
  cancelled.

The initial setup should not install compilers, editors, Java or language
toolchains. Those belong to explicit catalog installation.

### Step 3.2: Add native shell configuration safely

Add a renderer for a small Mobdesk-owned Bash source file, for example below
`$HOME/.config/mobdesk`. It must:

- export `PATH="$HOME/.local/bin:$PATH"` only when that directory is not
  already present;
- contain no hard-coded Termux package paths when `$PREFIX/bin` is already in
  the normal Termux PATH;
- avoid setting prompt, Java, Go, CGO or unrelated environment variables;
- be created with private permissions;
- be sourced by an idempotent, clearly bounded block in `$HOME/.bashrc`.

The `.bashrc` editor must preserve unrelated content. It must refuse unsafe
symlink handling and surface a diagnostic rather than replacing an unknown
file. Add a focused unit-tested helper instead of embedding mutable shell-file
logic in setup orchestration.

### Step 3.3: Open a native workspace shell

Edit `internal/cobra/shell.go`:

- Require completed native setup.
- Ensure `Paths.Workspace()` exists before opening the shell.
- Launch native Bash through the existing PTY-safe interactive flow.
- Set only the child process working directory to the workspace.
- Do not print or execute any `proot-distro` command.
- Preserve terminal raw-mode restoration and cancellation behavior.

If `runInteractive` cannot set a working directory, minimally extend it with a
command-construction input that supports `Cmd.Dir`; do not add a general
process framework solely for this shell change.

### Step 3.4: Remove SSH forced entry

Edit `internal/workstation/ssh_config.go`:

- Remove `SSHWrapper()` from paths and delete wrapper generation.
- Remove `ForceCommand` from rendered SSH configuration.
- Retain port, listen address, PID file, host keys, password policy and SFTP
  subsystem handling.
- Rename comments and errors so they describe native Mobdesk SSH, not an Ubuntu
  shell.

Edit `internal/workstation/start_stop.go`:

- Remove the Ubuntu availability check from `Start()`.
- Preserve setup/password checks, SSH ownership checks, port conflict handling,
  wake-lock behavior and start/stop locking.

### Step 3.5: Update setup, shell and SSH tests

Rewrite the relevant tests to verify:

- setup never installs or invokes PRoot;
- setup creates `$HOME/workspace` with the intended permissions;
- setup's phase list is exactly the native list;
- repeated setup does not duplicate the Bash source block;
- `.bashrc` user content is preserved;
- shell command receives native Bash and workspace working directory;
- rendered SSH configuration has no `ForceCommand`;
- start succeeds without a PRoot executable;
- start/stop continue rejecting an unowned SSH PID or occupied port.

**Acceptance criteria:**

- A clean test fixture reaches `setup.done` without PRoot.
- `mobdesk shell` uses a Termux-native interactive process rooted in the
  workspace.
- SSH configuration passes `sshd -t` and no longer changes the login command.
- `make check` passes.

**Proposed commit:** `feat: make setup and SSH native to Termux`

## 9. Phase 4: Build A Native Termux Installation Engine

**Goal:** all installation, verification, logging, repair and removal code runs
directly in Termux without retaining hidden APT assumptions.

### Step 4.1: Define native execution helpers

Refactor `internal/install` around one direct-command helper, conceptually:

```text
runTermuxLogged(context, runner, timeout, logPath, name, args...)
```

Required behavior:

- create a command with context and timeout;
- execute the requested native command directly;
- append command, stdout, stderr and error to a private log;
- redact nothing by default because commands must never contain user secrets;
- retain deterministic non-interactive handling for `pkg` operations;
- preserve cancellation and leave the installation record in a failed or
  partial observable state.

Log headers must name the actual command. They must no longer claim every
operation was `proot-distro`.

### Step 4.2: Replace APT semantics with package semantics

Remove `repairDpkg`, `runAptLogged`, APT lock timeout configuration and Ubuntu
PATH construction. Add only the native behavior that is actually necessary:

- `pkg install -y <packages...>` for repository profiles;
- `pkg uninstall -y <packages...>` only when safe removal is proven;
- native direct commands for Go, npm, pip/pipx and verified scripts;
- profile-specific prerequisite commands when a package must be installed
  before a native strategy runs.

Do not perform a global `pkg update` during every tool installation. Setup owns
the baseline update. A future explicit refresh policy can be designed only when
there is evidence it is necessary.

### Step 4.3: Make managed paths native and safe

Edit `declaredInstalledFiles`, uninstall validation and hash helpers:

- packages installed through `pkg` have no claimed binary file path because the
  package manager owns them;
- Mobdesk-managed downloaded or generated executables may live only below
  `$HOME/.local/bin`;
- use the actual configured Home when validating managed paths;
- do not write generated binaries into `$PREFIX/bin`;
- preserve hash comparison before a managed file is removed;
- retain shared-package protection, adapted to native package names.

### Step 4.4: Preserve state behavior

Installation state remains below:

```text
$HOME/.local/share/mobdesk/state/installations/<app>.json
$HOME/.local/share/mobdesk/state/configurations/<app>.json
$HOME/.local/share/mobdesk/logs/install/<app>.log
```

Because the Termux reset creates a fresh environment, no state migration logic
is added. New records must no longer contain an Ubuntu path, APT package
assumption or runtime label.

### Step 4.5: Test installation engine behavior

Add or revise tests for:

- direct native command shape and log rendering;
- cancellation and timeout;
- JSON and progress output with no stdout pollution;
- low-storage warning and block behavior;
- dependency installation and idempotent reinstallation;
- safe file hash removal under `$HOME/.local/bin`;
- package ownership and shared dependency protection;
- unsupported profile behavior without guest fallback.

**Acceptance criteria:**

- No installer command or log references PRoot, APT, dpkg or `/root`.
- `install`, `uninstall` and configuration operations remain serialized by the
  Mobdesk lock.
- Existing operation JSON remains valid schema 1 JSON.
- `make check` passes.

**Proposed commit:** `refactor: run catalog tools natively in Termux`

## 10. Phase 5: Audit And Convert The Catalog

**Goal:** retain only tools with an explicit, tested native Termux strategy.

### Step 5.1: Audit protocol

For every current catalog profile, record this information before editing the
profile:

| Field | Requirement |
|---|---|
| Termux source | Exact package repository or official verified download source |
| Package or strategy | Exact `pkg` package, Go command, npm command, pip/pipx command or script |
| Architecture | ARM64/Bionic compatibility confirmed |
| Executables | Every required executable and version argument confirmed |
| Dependencies | Native package/profile prerequisites confirmed |
| Install location | Package manager, `$PREFIX/bin` or `$HOME/.local/bin` |
| Removal path | Safe package or managed-file removal behavior confirmed |
| Configuration | Native config location and headless validation, if applicable |
| Device evidence | Command tested in a clean Termux ARM64 environment |
| Result | supported, deferred or removed from the active catalog |

Do not infer compatibility from Linux/ARM64 alone. A glibc binary, a manylinux
wheel or a script that invokes APT is not a Termux strategy.

### Step 5.2: Audit matrix

The following matrix is a work checklist, not an assertion that a native
package currently exists. Names in the “current strategy” column describe the
old Ubuntu catalog and must be replaced or removed.

| Profile | Current strategy | Native decision required |
|---|---|---|
| Go | APT `golang` | Verify Termux `golang`, `go version`, Go install destination |
| Java | APT OpenJDK 21 | Verify native OpenJDK 21, `java`, `javac`, `jar` and `JAVA_HOME` needs |
| Kotlin | Downloaded JVM compiler | Verify native package or pinned native-compatible archive and JDK integration |
| Gradle | Downloaded JVM archive | Verify native package/archive, wrapper behavior and JDK integration |
| Maven | APT `maven` | Verify native package and project wrapper behavior |
| Python | APT `python3` | Verify Termux package name, Python executable and native package builds |
| Node | APT Node/npm | Verify Termux Node/npm package names and global prefix behavior |
| C | APT Clang | Verify native `clang`, headers and compilation fixture |
| C++ | APT Clang | Verify native `clang++`, headers and compilation fixture |
| Lua | APT Lua 5.4 | Verify package and executable naming |
| Git | APT Git | Verify native package and executable |
| GitHub CLI | APT `gh` | Verify native package and executable |
| tmux | APT tmux | Verify native package and terminal behavior over SSH |
| Zellij | Downloaded musl archive | Verify a supported Termux build; otherwise remove/defer |
| Micro | APT micro | Verify native package and executable |
| Lazygit | Downloaded Linux archive | Prefer verified native package; otherwise verify Bionic-compatible release |
| tree | APT tree | Verify native package |
| TTT | Go install | Verify native Go installation and dependencies |
| htop | APT htop | Verify native package and `/proc` behavior |
| ncdu | APT ncdu | Verify native package |
| inxi | APT inxi | Verify native package and Android limitations |
| speedtest-cli | APT package | Verify native package or supported Python strategy |
| Posting | pipx | Verify pip/pipx availability and native dependencies |
| Yazi | downloaded GNU release | Verify a Termux package or Bionic-compatible build; otherwise remove/defer |
| TUIFI | Python build script | Verify native build dependencies and pip strategy |
| Neovim | APT neovim | Verify native package and minimum version |
| OpenCode CLI | npm | Verify npm global prefix and executable location |
| Codex CLI | npm | Verify npm global prefix and executable location |
| Claude Code CLI | npm | Verify npm global prefix and executable location |
| Leetgo | downloaded Linux archive | Verify native Go build or Bionic-compatible binary |

### Step 5.3: Convert in cohorts

Convert the catalog in this order. Each cohort is independently testable and
may remove/defer a profile that fails the audit.

1. Base workstation: Git, tree, tmux, Micro, Neovim, htop and ncdu.
2. Core languages: Go, Python, Node, C, C++, Lua and Java.
3. JVM tools: Kotlin, Gradle and Maven after Java is proven native.
4. Developer tools: GitHub CLI, TTT, Lazygit and Leetgo.
5. User tools: inxi, speedtest-cli, Posting, Yazi and TUIFI.
6. AI clients: OpenCode CLI, Codex CLI and Claude Code CLI.
7. Deferred tools: any profile that lacks a native supportable strategy,
   including Zellij if no compatible source is verified.

### Step 5.4: Update profile data and strategies

For each retained profile in `internal/install/install.go`:

- replace APT package names with verified Termux package names;
- replace `InstallKind: "apt"` with a native package strategy name such as
  `pkg`;
- remove scripts that install APT dependencies or write to `/usr/local/bin`;
- ensure required executables and version arguments reflect the actual Termux
  package;
- update dependencies to retained native profile names;
- update storage estimates after measuring native Termux ARM64, including the
  source, version, architecture and measurement date;
- retain localized descriptions and concise usage forms;
- remove a profile entirely when its strategy is not supportable.

### Step 5.5: Validate profiles on a clean device

For each cohort, after reset/setup and before considering it complete:

```text
mobdesk install <tool>
mobdesk install <tool>
mobdesk status --json
<tool> <version arguments>
mobdesk uninstall <tool>
```

For tools that create or compile a fixture, run the actual fixture. This is
mandatory for C/C++, Go, Python package use, Node, Java, Kotlin, Gradle and
Maven. A package that installs but cannot build the documented workload is not
validated.

**Acceptance criteria:**

- Every active profile has audit evidence and a native strategy.
- No active profile declares APT, `/root`, `/usr/local/bin`, PRoot or a guest.
- Deferred profiles are absent from the user-visible catalog.
- Each retained profile is idempotent in the Termux fixture and device sample.
- `make check` passes.

**Proposed commits:**

```text
feat: add native Termux base tool profiles
feat: add native Termux language profiles
feat: add native Termux JVM profiles
feat: convert remaining native tool profiles
```

## 11. Phase 6: Move Configuration And Safe Removal Into Termux HOME

**Goal:** configuration profiles and their cleanup rules use only the native
Termux user home.

### Step 6.1: Make configuration paths dynamic

Refactor configuration profile construction so paths can be based on
`paths.Paths.Home`. Static `/root` values must disappear from:

- `internal/install/lazyvim.go`;
- `internal/install/config.go`;
- `internal/install/uninstall.go`;
- configuration tests;
- status configuration hash checks.

The Neovim/LazyVim pilot must use:

```text
$HOME/.config/nvim
$HOME/.local/share/nvim/lazy
```

Do not share configuration with another runtime because no second runtime
exists.

### Step 6.2: Preserve safety invariants

Native paths do not relax safety rules:

- configuration destinations must be cleaned absolute paths below the expected
  Termux home;
- existing user configuration is a conflict, not a silent overwrite;
- generated files use safe temporary-write and rename behavior;
- plugin repositories use HTTPS and fixed revisions;
- removal compares stored hashes before deleting a file;
- modified files and plugins are preserved and reported;
- unmanaged packages and files are never deleted.

### Step 6.3: Validate native Neovim and LazyVim

- Install native Neovim through the catalog.
- Apply LazyVim only after Neovim's required minimum version is verified.
- Validate configuration using the native `nvim` executable in headless mode.
- Test a pre-existing `$HOME/.config/nvim` conflict.
- Test modified managed-file and modified-plugin preservation during removal.

**Acceptance criteria:**

- No configuration validator accepts `/root` paths.
- LazyVim works with native Neovim and never invokes PRoot.
- Apply, repeated apply, conflict, removal and modified-file tests pass.
- `make check` passes.

**Proposed commit:** `refactor: manage app configuration in Termux home`

## 12. Phase 7: Redesign Native Status And Cobra Boundaries

**Goal:** status represents the native workstation truthfully and all Cobra
commands describe a single supported environment.

### Step 7.1: Define status schema version 2

Edit `internal/status/model.go`:

- change `SchemaVersion` to `2`;
- remove `UbuntuStatus` and the `ubuntu` field;
- add `WorkspaceStatus` with at least state, path, existence and error;
- retain Host, Setup, Storage, SSH, Network, Battery, Wi-Fi, Installations,
  Configurations and Alerts where they remain meaningful;
- keep `Host.Termux` until an explicit future schema removes it, because it
  still communicates whether host control is available;
- remove fields whose only meaning was PRoot availability.

### Step 7.2: Collect native state

Edit `internal/status/collect.go`:

- replace `collectUbuntu` with `collectWorkspace`;
- make completed native setup and existing workspace determine setup health;
- remove all `proot-distro` command probes;
- discover catalog executables directly in the Termux environment;
- check `$HOME/.local/bin` for managed user executables;
- reconcile installations and configurations against native commands and
  native paths;
- make an absent workspace actionable and visible;
- ensure absence of any compatibility environment cannot degrade status because
  no such environment is modeled.

### Step 7.3: Keep Cobra as an adapter

- Retain host-runtime validation for destructive operations.
- Remove messages that tell a caller to return from Ubuntu to Termux.
- Ensure text, JSON and progress output describe Termux-native actions.
- Update command help, examples and localized errors.
- Do not add runtime flags or runtime-specific result fields.

### Step 7.4: Test the schema and command contract

Add tests that assert:

- schema version 2 JSON has workspace and has no `ubuntu` key;
- missing workspace is visible as a warning/error according to the chosen
  status policy;
- a clean completed setup is healthy without PRoot;
- direct executable detection recognizes native packages and managed user
  binaries;
- CLI JSON remains valid on success, validation error, runtime error and
  cancellation;
- no status path invokes `proot-distro`.

**Acceptance criteria:**

- Status represents a complete native setup as healthy.
- Status schema is documented as version 2 in code and public docs.
- The TUI backend can decode schema version 2 before Phase 8 begins.
- `make check` passes.

**Proposed commit:** `feat: report native Termux workstation status`

## 13. Phase 8: Simplify The TUI To One Environment

**Goal:** the TUI presents and controls the native Termux workstation directly,
including when opened over Mobdesk SSH.

### Step 8.1: Remove remote Ubuntu presentation

Edit TUI model and screens:

- remove the special remote Ubuntu session behavior from `internal/tui/tui.go`;
- remove host-action restrictions whose only cause was execution inside Ubuntu;
- remove Ubuntu status cards, labels, setup phase labels and explanations;
- remove the broken remote-shell assumption where a guest TUI tries to invoke
  host PRoot commands;
- retain generic behavior for genuinely unsupported runtime execution if it is
  still reachable.

### Step 8.2: Render native system state

Update `screen_status.go`, `screen_setup.go`, `screen_shell.go` and relevant
components to show:

```text
Termux            ready or actionable state
Workspace         $HOME/workspace state
SSH               running or stopped
Storage           native device storage state
Tools             installed, partial, available or failed
```

The shell action must be labelled and described as a native Termux workspace
shell. It uses the existing `tea.ExecProcess` terminal handoff and resumes the
TUI after shell exit.

### Step 8.3: Render native tool provenance

Update `app_popup.go` and tool-list state:

- remove runtime/Ubuntu alternative messaging;
- show native package or managed installation metadata only when it is useful;
- preserve install/reinstall, safe uninstall, configuration and close actions;
- preserve storage blocking, busy state, destructive confirmation, keyboard
  equivalence and touch hit regions;
- explain unavailable/deferred tools through the catalog rather than offering a
  compatibility action.

### Step 8.4: Update localization

Edit `internal/i18n/message.go` and both locale JSON catalogs:

- remove or stop requiring Ubuntu-specific active strings;
- add messages for native workspace, native shell and unavailable native tools
  only when presentation requires them;
- retain translated error and status vocabulary;
- never introduce presentation literals directly into Cobra, status or TUI Go
  files.

### Step 8.5: Verify TUI interaction

Add or update tests for:

- status rendering at narrow terminal widths;
- visible workspace state;
- native shell button keyboard and mouse/touch behavior;
- install, uninstall and configuration actions from an SSH-originated Termux
  session;
- disabled action explanation for storage, busy, conflict and unsupported-tool
  states;
- close/back controls on every new or changed popup;
- status schema version 2 backend decoding.

**Acceptance criteria:**

- No active TUI screen describes Ubuntu, PRoot, guests or runtime choice.
- SSH and local Termux TUI behavior is equivalent for supported actions.
- All important actions remain reachable by both keyboard and touch/mouse.
- `make check` passes.

**Proposed commit:** `feat: simplify TUI for native Termux workflows`

## 14. Phase 9: Replace Ubuntu Test Infrastructure

**Goal:** automated validation mirrors the one-userland architecture and does
not conceal PRoot regressions behind an obsolete fixture.

### Step 9.1: Audit current fixtures and scripts

Inspect and update:

- `Dockerfile.termux`;
- `docker-compose.yml`;
- `scripts/test-termux.sh`;
- `scripts/test-catalog.sh`;
- any catalog image, PRoot volume, Ubuntu cache or rootfs preparation;
- `Makefile` help and targets.

Delete Ubuntu fixture setup, PRoot package setup and Ubuntu cache volumes. Do
not retain unused Docker services “for future compatibility.”

### Step 9.2: Define native test tiers

The replacement test commands should be explicit:

| Target | Purpose |
|---|---|
| `make test` | Go unit tests in the Termux fixture |
| `make integration-test` | Setup, native shell and dedicated SSH lifecycle in the fixture |
| `make catalog-test-fast` | Representative native profiles and strategies for normal development |
| `make catalog-test-full` | Every active native profile, repeated install and required fixtures |
| `make check` | Formatting, i18n, vet, unit tests and build |

The exact fast representative set is chosen only after Phase 5 establishes the
native profile strategies. It must include a package profile, a managed user
binary profile, a Go/npm/Python strategy if each remains active, a dependency,
a configuration profile and an idempotent second installation.

### Step 9.3: Make fixture limits explicit

Docker is not Android or a real Termux phone. The automated fixture may prove
command shape, package availability, state management, SSH process ownership
and catalog logic. It cannot prove:

- Android Bionic behavior for all native binaries;
- HyperOS battery/suspension behavior;
- Termux app permissions and storage behavior;
- Wi-Fi and Termux:API behavior;
- real ARM64 package availability unless the ARM64 fixture is used;
- interactive keyboard and touch ergonomics.

These remain mandatory device-validation items.

### Step 9.4: Update developer documentation

- Update `Makefile` help descriptions.
- Document fast versus full catalog coverage.
- Remove claims that catalog validation uses Ubuntu.
- Document how to reset the Docker Termux fixture separately from the required
  real-device Termux reset.

**Acceptance criteria:**

- No Docker configuration creates, mounts or tests Ubuntu/PRoot state.
- Fast and full catalog validation describe and cover active native profiles.
- `make check`, `make integration-test` and the applicable catalog tests pass.

**Proposed commit:** `test: replace Ubuntu fixtures with native Termux validation`

## 15. Phase 10: Real Device Acceptance And Release Preparation

**Goal:** prove the product works from a completely reset ARM64 Termux
installation and publish the breaking architectural change accurately.

### Step 10.1: Fresh-device installation

On a reset real Termux device:

1. Install Termux from a supported source.
2. Install only the documented bootstrap prerequisites.
3. Install the Mobdesk binary by the documented release or source method.
4. Confirm neither `proot-distro` nor an Ubuntu rootfs is installed as part of
   Mobdesk setup.
5. Run `mobdesk setup` interactively.
6. Run `mobdesk status --json` and verify schema version 2.
7. Confirm `$HOME/workspace`, private state, launcher and Bash source setup.
8. Run `mobdesk setup` again and verify idempotency.

### Step 10.2: Native SSH validation

1. Run `mobdesk start`.
2. Confirm the reported username, address and port are correct.
3. Connect from a trusted LAN computer using the documented SSH command.
4. Confirm the shell is Termux, not a guest userland.
5. Confirm `$HOME`, `$PREFIX`, `pwd`, `command -v pkg` and workspace behavior.
6. Confirm a requested SSH remote command behaves normally.
7. Confirm SFTP behavior if it remains advertised by the generated config.
8. Open `mobdesk tui` locally and via SSH; verify equivalent native actions.
9. Run `mobdesk stop` and confirm it stops only the owned SSH daemon.

### Step 10.3: Catalog acceptance

For each retained catalog cohort:

1. Install every profile on the device.
2. Reinstall every profile and verify idempotency.
3. Run required executable/version checks.
4. Execute language/build fixtures where applicable.
5. Apply and remove supported configuration profiles.
6. Verify modified configuration preservation.
7. Verify logs contain useful command diagnostics and no secrets.
8. Confirm low-storage behavior when it can be safely simulated.

### Step 10.4: Documentation and release review

- Review every active Markdown reference to Ubuntu, PRoot, APT, `/root` and
  guest runtime claims. Keep only justified historical mentions.
- Run `make check`.
- Run `make integration-test`.
- Run `make catalog-test-fast` and `make catalog-test-full` after those targets
  exist.
- Run `git diff --check`.
- Review the full diff for accidental removal of SSH ownership, cancellation,
  private-permission or destructive-confirmation safeguards.
- Update `CHANGELOG.md` with the breaking move to Termux-only behavior and the
  reset requirement.
- Publish release notes that tell current Ubuntu-first users to reset Termux;
  do not imply that projects or tools will be migrated automatically.

**Acceptance criteria:**

- A reset ARM64 Termux device completes setup with no PRoot/Ubuntu dependency.
- SSH opens a normal Termux session.
- Workspace, shell, TUI and every active catalog profile work natively.
- No user-facing active path refers to a compatibility environment.
- All automated checks and documented device checks pass.

**Proposed commit:** `docs: release Termux-only workstation`

## 16. Security And Reliability Requirements

These requirements apply in every phase and cannot be traded for refactor
simplicity.

- Validate arguments before invoking a process; never concatenate user input
  into shell syntax.
- Use `context.Context` and timeouts for long-running native package and tool
  operations.
- Preserve terminal state after PTY shell/install handoffs and cancellation.
- Keep Mobdesk state, logs, SSH configuration, PID files and locks private.
- Do not log credentials, tokens, passwords or shell environment dumps.
- Keep the dedicated SSH port and PID ownership proof before stopping anything.
- Do not overwrite user shell or app configuration silently.
- Require in-screen confirmation for destructive TUI actions.
- Keep all important TUI actions available by both touch/mouse and keyboard.
- Do not claim systemd, Docker, root, namespaces, kernel modules or Android
  background execution guarantees.

## 17. Completion Checklist

The Termux-first refactor is complete only when every item below is true.

- [ ] Active architecture, mission, roadmap and READMEs describe Termux as the workstation.
- [ ] The required Termux reset and lack of migration are documented.
- [ ] No production code invokes or installs `proot-distro`.
- [ ] No production path, status field, TUI screen or active help text models Ubuntu.
- [ ] `mobdesk setup` creates a complete native workstation without PRoot.
- [ ] `mobdesk shell` opens native Bash in `$HOME/workspace`.
- [ ] Mobdesk SSH has no `ForceCommand` and opens Termux normally.
- [ ] Installation, verification, logs, configuration and safe removal are native.
- [ ] Every active catalog profile has native audit evidence and device validation.
- [ ] Unsupported profiles are removed/deferred instead of redirected to another runtime.
- [ ] LazyVim and all configuration paths are rooted in Termux HOME.
- [ ] `status --json` schema version 2 documents workspace and has no Ubuntu object.
- [ ] Local and SSH TUI flows operate on the same native Termux state.
- [ ] Docker fixtures and catalog tests contain no Ubuntu/PRoot setup.
- [ ] `make check` passes.
- [ ] `make integration-test` passes.
- [ ] Native fast and full catalog tests pass after they are introduced.
- [ ] Fresh real-Termux ARM64 acceptance passes.
