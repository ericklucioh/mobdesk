# Mobdesk

Mobdesk turns an Android phone into a development workstation. Termux is the
sole host and development environment. PRoot-Distro and Ubuntu are removed from
the active scope. Existing PRoot-based installs require a full Termux reset and
fresh installation; no migration is supported.

## Architecture

```text
Termux (Android host, SSH, wake-lock, workspace and development tools)
```

- Entry point: `cmd/mobdesk/main.go`; module:
  `github.com/ericklucioh/mobdesk`.
- Cobra implements the CLI; Bubble Tea implements the TUI.
- Services do not depend on TUI rendering.
- The TUI uses the CLI JSON contract for real operations.
- All actions run in Termux. SSH sessions use the same Termux workstation.

## Current scope

Current commands are `start`, `stop`, `setup`, `shell`, `install`, `uninstall`,
`status`, `update`, `version`, `tui` and `logs`. `doctor`, projects, services, a web interface,
an APK, a graphical desktop, real Docker, Nix and multiple users remain outside
the MVP.

## Dependencies and roles

- `charm.land/bubbletea/v2`: event loop and TUI application;
- `charm.land/bubbles/v2`: lists, inputs, tables and spinners;
- `charm.land/lipgloss/v2`: styles and layout;
- `github.com/aymanbagabas/go-osc52/v2`: OSC 52 clipboard;
- `github.com/spf13/cobra`: CLI commands;
- `github.com/spf13/pflag`: CLI flags;
- `golang.org/x/sync`: concurrency coordination;
- `golang.org/x/sys`: low-level integration when required;
- `charmbracelet/x`, terminfo and terminal packages: terminal support.

## Implementation rules

1. Prefer small changes and the standard library before new dependencies.
2. Run commands in Termux; use `os/exec` for simple processes and PTY for
   interactive shells.
3. Validate inputs before forming commands. Never concatenate user input into
   shell syntax.
4. Repeated operations preserve data and state; destructive actions require
   confirmation.
5. Use context and cancellation for long processes and never block the TUI.
6. Keep state, logs and configuration in private paths. Never log secrets.
7. Do not create packages in advance; extract a package only for real reusable
   behavior.

## CLI boundary and contract

1. Every non-interactive Cobra command consumed by the TUI or automation must
   offer `--json` and keep its versioned schema. `shell` and `tui` are
   interactive exceptions.
2. JSON mode writes only valid JSON to stdout. Human messages, progress and
   diagnostics must not pollute it; progress uses the command's documented
   event format.
3. JSON preserves `schema_version`, `command`, `success`, `state` and `message`.
   New fields are additive and compatible with the existing schema.
4. Cobra adapts flags, arguments, context and output. Business rules,
   installation, status and host operations belong to internal
   services, not command handlers.
5. Termux is the Android host and the development userland. Do not introduce a
   PRoot/Ubuntu execution boundary without an explicit scope decision.
6. SSH uses the same Termux environment; do not add remote Ubuntu-mode blocking.
7. Do not detect Termux with `runtime.GOOS` alone. Use project markers and
   canonical paths such as `PREFIX` and `paths.Current()`.
8. Simple processes use `executil`. The TUI never calls package managers or
   scripts directly.
9. Every long process receives `cmd.Context()` or equivalent, supports
   cancellation, and leaves state and logs consistent after partial failure.
10. New commands need tests for invalid arguments, text mode, JSON mode when
    applicable, runtime errors and cancellation for long operations.

## TUI UX rules

1. The TUI is touch-first: every important action has a visible clickable
   target and does not depend on keyboard discovery.
2. Clickable hit regions match the rendered control. Do not make an entire row a
   slow, destructive or surprising button when a smaller target is possible.
3. Every new screen or popup has a visible clickable Back, Close or X action and
   a keyboard equivalent.
4. Every mouse/touch action also works by keyboard; unavailable actions explain
   why in the same flow.
5. Tapping an app row opens details before install or removal. Destructive
   actions require in-screen confirmation. Application configuration and
   LazyVim are deferred.
6. Busy, error, conflict, completed and blocked states are visible and prevent
   duplicate or incompatible actions.
7. Screens work in narrow terminals without fixed widths or clipped controls.
8. New mouse/touch flows need hit-test, navigation and keyboard-equivalence
   tests, plus narrow-terminal validation.

## Validation and documentation

Run `make check` before completing changes. Docker validates logic; final
integration requires real Termux on the POCO F6.

Read `docs/MISSION.md` before changing architecture or scope. Update:

- `docs/DECISIONS.md` for decisions;
- `docs/ARCHITECTURE.md` for technical boundaries;
- `docs/ROADMAP.md` for scope and stages.
