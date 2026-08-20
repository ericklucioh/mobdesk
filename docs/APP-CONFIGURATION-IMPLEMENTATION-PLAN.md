# App and Configuration Implementation Plan

**Status:** superseded historical plan. Application configuration and LazyVim
are outside the current MVP, and its PRoot/Ubuntu design must not guide native
Termux implementation. Use [`MISSION.md`](MISSION.md) and
[`POST-TERMUX-SPRINTS.md`](POST-TERMUX-SPRINTS.md) for current scope.

**Decision document:** [`APP-CONFIGURATION-REFACTOR-PLAN.md`](APP-CONFIGURATION-REFACTOR-PLAN.md)

**Objective:** implement the app center with details, safe installation and
removal, optional configuration, storage estimates and the Neovim/LazyVim
profile without changing the schema 1 contract or persisted runtime paths.

## 1. Mandatory contract

- use `AppProfile` for the catalogue;
- include Neovim as an installable app;
- make LazyVim the only first-delivery opinionated configuration;
- install Neovim before `config apply`;
- refuse existing configuration rather than overwrite it;
- create no automatic configuration backup;
- remove only files and plugins with Mobdesk provenance;
- never automatically remove dependencies in the MVP;
- store configuration separately at `state/configurations/<app>.json`;
- keep schema 1 and additive optional JSON fields;
- always open the app popup on touch or selection;
- confirm destructive actions inside the popup;
- show detected apps as installed without treating them as managed;
- update documentation with every phase.

## 2. Final user flow

The user can open Apps, select an app, inspect name/description/state/storage,
install it, follow the final result, apply optional Mobdesk configuration,
inspect paths and plugins, remove configuration without losing edits, and
uninstall only when provenance makes removal safe. In a TUI opened inside an
Ubuntu SSH session, host operations remain blocked with an explanation.

## 3. Out of scope

This delivery does not add another configuration profile, automatic backups,
dotfile merging, automatic shared-dependency removal, configuration outside
Mobdesk profiles, a web interface, an APK, cloud synchronization, a remote
profile marketplace, automatic updates of every installed app, or forced
reinstallation without confirmation.

## 4. Architecture target

```text
Bubble Tea TUI
  -> real or mock backend
  -> Cobra CLI and schema 1 JSON
  -> internal services
      -> AppProfile catalogue
      -> install / uninstall / configuration
      -> state and provenance
      -> status
  -> Ubuntu through proot-distro
```

The TUI owns interaction, popup, focus, confirmation and presentation. Cobra
adapts flags, runtime, JSON and human output. `internal/install` owns profiles,
app lifecycle, configuration and state; `internal/status` reconciles the
snapshot; `internal/paths` remains the canonical source of persisted paths.
The TUI never calls `apt`, `pipx`, `npm`, `proot-distro` or scripts directly.

## 5. Phase order

| Phase | Delivery | Dependency |
|---|---|---|
| 0 | Preparation and frozen contract | Decision document |
| 1 | `AppProfile` and catalogue | 0 |
| 2 | Installable Neovim | 1 |
| 3 | Storage estimates | 1 and 2 |
| 4 | Provenance and state | 1 and 3 |
| 5 | Safe uninstall | 4 |
| 6 | Configuration engine | 2 and 4 |
| 7 | LazyVim profile | 6 |
| 8 | CLI and JSON contracts | 5, 6 and 7 |
| 9 | Status reconciliation | 4 and 8 |
| 10 | TUI popup | 8 and 9 |
| 11 | Integration, tests and documentation | 10 |

A phase is complete only after its tests and acceptance criteria pass.

## 6. Phase results

### Phase 0 - Preparation and frozen contract

The contract introduced `AppProfile`, `StorageEstimate`, canonical app and
configuration states, the additive `storage_estimate` field and compatibility
with the existing `Language` catalogue and install flow. Compile and JSON
compatibility tests were added without changing production installation.

### Phase 1 - App profile and catalogue

The catalogue now uses `AppProfile`; `Languages`, `Tools` and `Resolve` return
profiles, descriptions live in the catalogue, aliases remain valid, and every
profile has an initial Ubuntu ARM64 storage estimate. Existing installation
strategies and commands remain unchanged.

### Phase 2 - Installable Neovim

Neovim uses the Ubuntu `neovim` package, alias `nvim`, executable `nvim`, and
`nvim --version` verification. It declares optional `lazyvim` configuration and
`/root/.config/nvim`, but installation does not create configuration. Unit
tests cover resolution, aliases and apt commands; the Docker fixture validated
`NVIM v0.11.6` on x86_64. ARM64 validation remains a real-device task.

### Phase 3 - Storage estimates

Profiles expose app, dependency, configuration and total ranges. Install and
status JSON include additive `storage_estimate`; old records and detected apps
receive profile estimates where matched. Estimates never change install state
and remain distinct from free device space.

### Phase 4 - Provenance and persistent state

Installation records persist strategy, dependencies, packages, declared files
and source. Old records remain readable. Future detections use
`source=detected` and are not managed. A separate configuration record stores
profile, state, paths, hashes, conflicts and errors at
`state/configurations/<app>.json`; directories use `0700`, files `0600`, and
atomic rename protects writes. App names are constrained to the private state
root.

### Phase 5 - Safe uninstall

`Uninstall` uses the installation lock and requires `source=mobdesk`.
Detected-only records, shared packages and paths without provenance are
refused. Explicit strategies cover apt, node/npm, pipx, scripts, Go, TTT,
cargo and GitHub extensions without automatic dependency removal. Hashes are
checked; modified files are preserved and marked `modified`. Partial results,
preserved files and failures remain visible in persistent state and JSON.

### Phase 6 - Configuration engine

`ApplyConfig` and `RemoveConfig` use static profiles, validated paths,
declarative files, plugin manifests and estimates. Applying requires an
installed app, refuses conflicts, writes base64 file data through Ubuntu and
uses rename without putting user content into shell syntax. States
`applying`, `applied`, `removing`, `removed`, `modified` and `failed` persist in
the separate record. Removal compares hashes and rollback removes only intact
components created by the current attempt.

### Phase 7 - Neovim/LazyVim profile

The embedded `lazyvim` profile contains versioned Lua files and lock data. It
declares HTTPS repositories, full fixed revisions and managed Ubuntu paths for
`lazy.nvim`, LazyVim and nvim-treesitter. `config apply` clones and checks out
only declared revisions; `config remove` removes only clean managed plugins.
The profile remains optional, requires Mobdesk-installed Neovim, and refuses
existing configuration before any write or clone. Manual validation on a real
Android/Termux device has been performed; broader headless profile validation
remains ongoing.

### Phase 8 - CLI and JSON

The CLI exposes `uninstall`, `config apply` and `config remove`, each with
`--json` and `--progress`. Schema 1 retains old fields such as `language` and
adds `target`, `action`, `changed`, `config_state`, `source`, `paths`,
`conflicts` and `storage_estimate` where applicable. Success, conflict,
partial failure, missing provenance and remote Ubuntu errors produce a final
valid JSON result; progress events remain separate.

### Phase 9 - Status reconciliation

`status --json` separates installations and configurations and associates them
by canonical app. Configurable apps without records are `not_applied`; existing
unregistered paths are `conflict`. Hash inspection identifies `modified` paths
without deleting files. `source=detected` remains unmanaged and does not enable
uninstall.

### Phase 10 - TUI popup

Touch and Enter open a details popup rather than installing directly. The popup
shows state, source, version, dependencies, configuration, paths, plugins,
storage estimate and unavailable-action reasons. Install, uninstall, apply and
remove use the CLI backend; uninstall and configuration removal require an
in-popup keyboard or mouse confirmation. Remote Ubuntu, detected apps and
conflicts remain visible and block unsafe actions. Narrow-terminal and
hit-testing tests cover the flow.

### Phase 11 - Tests and integration

`make check` passed with formatting, vet, tests and fixture build. Tests cover
catalogue resolution and aliases, installation and cancellation, provenance,
uninstall safety, configuration hashes and rollback, JSON contracts, status,
popup focus, confirmations, conflicts, remote runtime and narrow terminals.
The catalogue smoke test had already passed for Neovim, Yazi and TUIFI and was
not repeated because this phase did not alter catalogue, installation or PRoot.

The target real Termux workflow has been manually validated; broader device and
profile coverage remains for follow-up validation.

## 7. Manual validation on real Termux

Before testing, update Mobdesk, confirm Ubuntu and free space, make an external
backup, and move any configuration that must be preserved. Test the clean
Neovim install, the LazyVim apply flow, fixed plugin revisions and headless
startup. Test conflict refusal with an existing `/root/.config/nvim`, then
modify a managed file and verify `config remove` preserves it and reports
`modified`. Through SSH, verify that host operations are blocked and the user
is told to return to Termux.

## 8. Global completion criteria

All profiles have explicit install behavior and an estimate or documented
limitation; the popup works by keyboard and mouse in narrow terminals; app and
configuration state remain separate; existing configuration is never silently
overwritten; modified files and shared dependencies survive removal; detected
apps cannot be uninstalled without provenance; schema 1 JSON is valid on
success and failure; remote Ubuntu blocks host actions; `make check` passes;
the catalogue smoke test passes; and the full flow passes on real Termux/PRoot.

## 9. Documentation and phase commits

Update this plan, [`ARCHITECTURE.md`](ARCHITECTURE.md),
[`DECISIONS.md`](DECISIONS.md) and [`ROADMAP.md`](ROADMAP.md) when a phase
changes their boundaries. Do not reopen discarded alternatives. Each original
implementation phase used one isolated commit; the final app-management
commit was `test: validate app lifecycle and document rollout`.
