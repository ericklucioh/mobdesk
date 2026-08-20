# Priority Refactoring Plan

**Status:** superseded historical plan. It describes the retired PRoot/Ubuntu
runtime and must not guide current implementation. Use
[`MISSION.md`](MISSION.md) and [`ARCHITECTURE.md`](ARCHITECTURE.md) instead.

This historical implementation plan records incremental refactors that preserve
existing commands and tests. Packages are extracted only for real reusable
behavior.

## 1. Separate Termux and Ubuntu in the TUI

**Status:** completed in the current TUI.

The project has two process environments: Termux, which owns PRoot, SSH and
wake-lock, and Ubuntu through PRoot, which is the development environment and
SSH session destination. A TUI started through SSH therefore cannot offer
`start`, `stop`, `setup` or `update` as host actions.

The completed direction detects runtime explicitly, keeps workspace and local
shell actions available remotely, blocks host operations with an objective
explanation, preserves the CLI JSON contract, and tests both runtimes.

## 2. Centralize paths and persistent state

**Status:** completed. `internal/paths` is the canonical source for the current
layout and migrated consumers preserve existing files and directories.

Paths include `$HOME/.local/share/mobdesk`, `$HOME/.config/mobdesk`, logs,
setup markers, installation records, SSH configuration and `/root/workspace`.
Consumers receive explicit HOME, PREFIX and path dependencies where needed.
The persisted layout is not changed by this refactor, and temporary-directory
tests do not depend on the real HOME.

## 3. Move start and setup orchestration out of Cobra

**Status:** completed. `internal/workstation.Service` orchestrates `start`,
`stop` and setup phases with explicit paths and dependencies. Cobra adapts
flags, streams and human/JSON rendering.

The service boundary keeps PID, port, lock, SSH, wake-lock, filesystem and
process rules testable without real `sshd`, PRoot or Termux. External behavior
of `mobdesk start`, `stop` and `setup` remains compatible.

## Historical sequence

The original work was ordered as: fix the TUI/documentation mismatch, separate
Termux and Ubuntu behavior, then extract `start`/`stop` and apply the same
design to setup. This record is retained for project history; the current
architecture and [`ROADMAP.md`](ROADMAP.md) are authoritative.
