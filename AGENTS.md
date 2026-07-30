# Mobdesk

Mobdesk turns an Android phone into a development workstation. Termux is the
control host; persistent Ubuntu ARM64 through PRoot-Distro is the development
environment. PRoot is not a VM or Docker: do not assume systemd, cgroups,
complete namespaces, real root, kernel modules or graphics acceleration.

## Architecture

```text
Termux (Android host, SSH, wake-lock, PRoot)
└── Ubuntu via PRoot (workspace and development tools)
```

- Entry point: `cmd/mobdesk/main.go`; module:
  `github.com/ericklucioh/mobdesk`.
- Cobra implements the CLI; Bubble Tea implements the TUI.
- Services do not depend on TUI rendering.
- The TUI uses the CLI JSON contract for real operations.
- Host actions (`setup`, SSH, PRoot, installation and updates) run only in
  Termux. From an SSH session inside Ubuntu, the TUI must explain the boundary.

## Current scope

Current commands are `start`, `stop`, `setup`, `shell`, `install`, `status`,
`update`, `version` and `tui`. `doctor`, projects, services, a web interface,
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
2. Keep Termux and Ubuntu commands explicitly separate; use `os/exec` for
   simple processes and PTY for interactive shells.
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
   installation, status, configuration and host operations belong to internal
   services, not command handlers.
5. Termux is the Android host and Ubuntu via PRoot is the development userland.
   Never assume Ubuntu has host access, real root, systemd or full namespaces.
6. Host-only actions validate the runtime and, from Ubuntu or SSH, return an
   objective explanation telling the user to return to Termux.
7. Do not detect Termux with `runtime.GOOS` alone. Use project markers and
   canonical paths such as `PREFIX` and `paths.Current()`.
8. Simple processes use `executil`; Ubuntu commands cross the declared
   `proot-distro` boundary. The TUI never calls `apt`, `pipx`, `npm`,
   `proot-distro` or scripts directly.
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
5. Tapping an app row opens details before install, removal or configuration.
   Destructive actions require in-screen confirmation.
6. Busy, error, conflict, completed and blocked states are visible and prevent
   duplicate or incompatible actions.
7. Screens work in narrow terminals without fixed widths or clipped controls.
8. New mouse/touch flows need hit-test, navigation and keyboard-equivalence
   tests, plus narrow-terminal validation.

## Validation and documentation

Run `make check` before completing changes. Docker validates logic and the
simulated userland; final integration requires real Termux on the POCO F6.

Read `docs/MISSION.md` before changing architecture or scope. Update:

- `docs/DECISIONS.md` for decisions;
- `docs/ARCHITECTURE.md` for technical boundaries;
- `docs/ROADMAP.md` for scope and stages.
