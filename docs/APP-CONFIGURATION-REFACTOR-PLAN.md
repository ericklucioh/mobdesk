# App and Configuration Refactor Plan

**Status:** product decisions recorded; implementation completed through the
documented app lifecycle work

**Objective:** turn the current tools screen into an app center with details,
installation, safe removal and optional Mobdesk-owned configuration profiles.

This is a decision record and executable plan. Its closed choices must not be
reopened during implementation without an explicit scope change.

## 1. Primary decision

Each app is a Mobdesk profile. A profile may declare installation, removal,
dependencies, optional configuration, affected paths, plugins, application and
safe removal behavior. Installing an app and applying its configuration are
independent operations. The suggested configuration is optional and is always
identified as a Mobdesk configuration.

## 2. Closed decisions

### Product

- The TUI is the MVP's primary interface.
- The app screen opens a details popup instead of installing on row activation.
- The popup shows name, explanation, state, storage estimate and available
  actions: install, safe uninstall, add/remove optional configuration and close.
- Configuration is a suggestion, never a requirement, and is never applied
  silently.

### Architecture

- Termux remains the control host.
- Ubuntu through PRoot remains the user's primary environment.
- Apps and configuration are applied inside Ubuntu.
- Host actions remain blocked in an Ubuntu SSH session.
- The TUI consumes the CLI JSON contract and owns no installation rules.
- Cobra remains a flags/input/output adapter; services remain independent of
  rendering.
- Long operations use context and cancellation, with at most one host
  operation active in the TUI.

### Safety and persistence

- Destructive actions require explicit confirmation.
- Existing user files are never silently overwritten or removed.
- Existing configuration is detected before application.
- User-modified files remain during removal.
- Shared packages are not blindly removed.
- Private operation state and logs persist across repeated operations.
- Logs contain no secrets and partial failures do not delete unrelated data.
- Final validation includes `make check` and real Termux testing.

## 3. Current implementation boundary

The existing flow resolves an app, calls `mobdesk install`, crosses the
`proot-distro` boundary, persists state and refreshes the TUI. The app popup,
uninstall and configuration operations use the shared backend and service
contracts rather than executing Ubuntu commands from the TUI.

## 4. Desired experience

Every card shows the primary name, short description, app state and optional
configuration state. Enter or touch opens a popup containing version,
dependencies, affected paths, plugins, storage estimate, available actions and
an explicit close control.

Popup behavior:

| State | Behavior |
|---|---|
| Not installed | Install is available; configuration explains its prerequisite |
| Detected outside Mobdesk | Shows installed, but removal is unavailable |
| Installed without configuration | App use and optional configuration are available |
| Configuration applied | Removal is available after confirmation |
| Busy | Other actions are blocked |
| Conflict | Automatic application is blocked with an explanation |
| No safe removal | Removal is disabled with an objective reason |
| Remote Ubuntu session | Explains that the action must run in Termux |

Uninstall and configuration removal require confirmation by keyboard and mouse,
including narrow terminals.

## 5. Domain model

The catalogue profile contains the following conceptual fields:

```text
name, aliases, description, kind, package, executable, version_arg,
install_kind, requires, user_bin, install_profile, uninstall_profile,
config_profile, profile_version, storage_estimate
```

Profiles never accept commands from user input. Commands and paths come from
the versioned Mobdesk catalogue.

A configuration profile declares a stable ID and version, associated app,
description, Ubuntu destinations, generated files/directories, managed plugins,
dependencies, conflict policy and apply/remove strategies.

Installation state remains in `state/installations/<app>.json`; configuration
state is separate in `state/configurations/<app>.json`. Both are private and
persistent, and old installation records remain readable.

Storage estimates are split into `app_mb`, `dependencies_mb` and `config_mb`.
They are planning ranges for Ubuntu ARM64 and must not be added across apps
when dependencies are shared. Profiles record source, evaluated version,
architecture and measurement date.

| App or language | App | Dependencies | Configuration | Isolated total |
|---|---:|---:|---:|---:|
| Go | 180-300 MB | 0-50 MB | 0-5 MB | 180-355 MB |
| Python | 35-60 MB | 0-20 MB | 0-5 MB | 35-85 MB |
| Node.js | 70-130 MB | 20-60 MB | 0-10 MB | 90-200 MB |
| C/C++ with shared Clang | 250-450 MB | 20-80 MB | 0-10 MB | 270-540 MB |
| Lua | 2-6 MB | 0-5 MB | 0-2 MB | 2-13 MB |
| Git | 35-60 MB | 0-10 MB | 0-5 MB | 35-75 MB |
| GitHub CLI | 30-50 MB | 0-10 MB | 0-5 MB | 30-65 MB |
| tmux | 2-5 MB | 0-2 MB | 0-2 MB | 2-9 MB |
| Zellij | 20-30 MB | 0-5 MB | 0-5 MB | 20-40 MB |
| Micro | 15-25 MB | 0-5 MB | 0-5 MB | 15-35 MB |
| Lazygit | 15-25 MB | 0-5 MB | 0-5 MB | 15-35 MB |
| tree | <1 MB | 0-1 MB | 0-1 MB | <3 MB |
| TTT | 10-20 MB | 0-10 MB | 0-2 MB | 10-32 MB |
| Posting | 20-60 MB | 10-40 MB | 0-5 MB | 30-105 MB |
| Yazi with previews | 25-40 MB | 300-550 MB | 1-20 MB | 326-610 MB |
| TUIFI Manager | 20-40 MB | 90-180 MB | 1-5 MB | 111-225 MB |
| Neovim without configuration | 15-30 MB | 0-20 MB | 0-2 MB | 15-52 MB |
| LazyVim on Neovim | 0 MB | 0-20 MB | 100-300 MB | 100-320 MB |
| OpenCode CLI | 60-150 MB | 0-100 MB | 5-30 MB | 65-280 MB |
| Codex CLI | 60-150 MB | 0-100 MB | 5-30 MB | 65-280 MB |
| Claude Code CLI | 80-200 MB | 0-100 MB | 5-30 MB | 85-330 MB |
| Leetgo | 10-20 MB | 0-20 MB | 0-5 MB | 10-45 MB |

## 6. Configuration safety

Before applying, the service resolves the canonical app, confirms installation
and profile existence, validates paths below the expected Ubuntu HOME,
inspects existing files, classifies ownership and emptiness, checks conflicts,
persists an attempt, applies files/plugins, validates the result, and writes
hashes only after success.

Application uses same-filesystem temporary files, atomic rename where possible,
component manifests, first-unrecoverable-error stop behavior and rollback of
the current operation. Shell commands never contain unvalidated names.

Removal compares recorded hashes: unchanged managed files may be removed,
changed files are preserved and marked `modified`, missing files are recorded,
and unknown files are preserved. Shared dependencies remain installed.

## 7. CLI and JSON contract

The planned operations are:

```text
mobdesk install <app>
mobdesk uninstall <app>
mobdesk config apply <app>
mobdesk config remove <app>
```

They accept `--json` and, for long operations, `--progress`. Schema 1 remains
stable and additive. Results may contain `target`, `action`, `changed`,
`config_state`, `storage_estimate`, `log_path`, `conflicts` and `paths`, while
existing fields retain their meaning.

Canonical app states are `available`, `installing`, `installed`,
`uninstalling`, `uninstalled`, `partial` and `failed`. Configuration states are
`unavailable`, `not_applied`, `applying`, `applied`, `removing`, `removed`,
`modified`, `conflict` and `failed`.

## 8. Neovim/LazyVim pilot

Neovim is installed first through the Ubuntu `neovim` package and verified by
`nvim --version`. LazyVim is the only initial optional configuration. It uses
fixed HTTPS repositories and complete revisions, embeds its static files, and
validates headlessly. Existing `/root/.config/nvim` content creates a conflict;
no automatic backup is made.

## 9. Registered decisions

| ID | Decision |
|---|---|
| D1 | Use `AppProfile` for languages, tools, editors and clients |
| D2 | Only Neovim/LazyVim is configurable in the first delivery |
| D3 | Require Neovim before applying LazyVim |
| D4 | Refuse existing configuration and report conflict |
| D5 | Do not create automatic configuration backups |
| D6 | Remove only provably managed plugins |
| D7 | Never remove shared dependencies automatically in the MVP |
| D8 | Store configuration in `state/configurations/<app>.json` |
| D9 | Keep JSON schema 1 with additive optional fields |
| D10 | Disable unavailable uninstall and explain why |
| D11 | Confirm destructive actions inside the popup |
| D12 | Represent partial safe removal with `modified` details |
| D13 | Install plugins during explicit `config apply` |
| D14 | Always open the popup on a list touch or Enter |
| D15 | Treat detected executables as installed but not automatically managed |
| D16 | Update documentation with each implementation phase |
| D17 | Embed the profile and clone only fixed plugin revisions |

## 10. Risks and completion

Main risks are mixed installation strategies, shared packages, unknown
provenance, pre-existing user configuration, manually modified plugins, narrow
mobile screens, long operations and Android interruption. Required mitigations
are explicit catalogue strategies, provenance manifests, hashes, conservative
dependency handling, confirmation, shared locks, rollback where possible,
observable partial state and real-Termux tests.

The refactor is complete only when the popup, shared services, separate states,
optional LazyVim profile, safe removal, schema-compatible JSON, keyboard/mouse
flows, remote-host blocking and full tests all pass, including real Termux
validation.
