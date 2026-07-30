# Global TUI Refresh Plan

**Status:** active implementation idea; not a shipped contract by itself.

## Objective

Let `r` and `R` refresh the displayed state on every TUI screen while keeping
Cobra as the source of truth. Refresh is a static read and never performs
setup, installation, start, stop or update.

The real backend uses only existing commands:

```text
mobdesk status --json
mobdesk version --json
```

The update check remains an explicit action. Refresh updates workstation, SSH,
environment, installations, setup phases, system version and shell state while
preserving screen, scroll position and focus where possible.

Ignore refresh during an active operation, confirmation modal or shutdown. The
real flow has no polling, invented percentages or fake stages. It shows a fixed
busy message, parses the final JSON, displays success/error, refreshes shared
state and updates dependent screens. The mock may delay and simulate success,
failure, degraded state and installed/uninstalled tools, but follows the same
contract.

## Tests and acceptance

Test `r`/`R` on Home, Apps, Setup, Status, System and Shell; state updates after
simulated installation/start/stop; CLI version loading; refresh blocking during
operations and confirmation; screen preservation; no polling; no fabricated
progress; and narrow and wide terminals. The result must never execute a
destructive action or alter the final Cobra JSON format.
