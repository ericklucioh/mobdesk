# `mobdesk doctor` Implementation Plan

**Status:** future scope; `doctor` is not part of the current MVP command set.

`doctor` will be a deep, read-only diagnostic that reuses `status`, explains
problems with evidence and suggests safe next steps. It must continue reporting
when one check fails and provide human and JSON output.

## Scope

The initial command surface is:

```bash
mobdesk doctor
mobdesk doctor --json
mobdesk doctor --deep
mobdesk doctor --fix
mobdesk doctor --strict
```

Normal mode must not install, remove, kill processes or change configuration.
JSON stdout contains only JSON. `--deep` may inspect the full rootfs, Ubuntu
login, toolchains and SSH in more detail. `--fix` only applies reversible,
explicitly confirmed safe actions. `--strict` treats relevant warnings and
unknown checks as failures.

## Diagnostic model

```go
type CheckStatus string

const (
    CheckOK CheckStatus = "ok"
    CheckInfo CheckStatus = "info"
    CheckWarning CheckStatus = "warning"
    CheckError CheckStatus = "error"
    CheckUnknown CheckStatus = "unknown"
)

type CheckResult struct {
    ID string `json:"id"`
    Category string `json:"category"`
    Status CheckStatus `json:"status"`
    Summary string `json:"summary"`
    Evidence []string `json:"evidence,omitempty"`
    Suggestions []string `json:"suggestions,omitempty"`
    Fixable bool `json:"fixable"`
    FixApplied bool `json:"fix_applied,omitempty"`
    ErrorCode string `json:"error_code,omitempty"`
}
```

Checks cover host/Termux, storage, setup, Ubuntu/PRoot, SSH, network,

Ubuntu and projects are details inside Termux and must not be double-counted in
storage totals. A missing optional language is informational unless explicitly
requested. A failed network check must not stop storage or toolchain checks.

## Safe fixes

After confirmation, future `--fix` may recreate a missing private directory,
non-destructive state file, dedicated SSH configuration or correct launcher;
remove a stale PID after proving the process is gone; and repair known Mobdesk
file permissions. Strong confirmation or separate commands are required for
removing Ubuntu, projects or large caches, revoking all keys, changing a
password, killing user processes, changing SSH ports or reinstalling languages.

## Initial check IDs

```text
host.architecture host.termux host.home host.prefix host.permissions
host.commands host.wakelock
storage.device storage.termux storage.ubuntu storage.mobdesk storage.projects
storage.threshold
ubuntu.installed ubuntu.architecture ubuntu.login ubuntu.command
ubuntu.workspace ubuntu.permissions
ssh.config ssh.host_keys ssh.pid ssh.process ssh.port ssh.banner
ssh.authentication
network.interface network.ip network.local network.internet network.tailscale
toolchain.go toolchain.node toolchain.python toolchain.c toolchain.cpp
toolchain.java toolchain.kotlin toolchain.lua toolchain.php toolchain.ruby
toolchain.rust
```

## Severity and exit codes

`info` does not impede operation, `warning` identifies a limitation or risk,
`error` means an expected component does not work, and `critical` means the
environment should not continue operating. Planned exit codes are `0` for no
critical failure, `1` for an essential error, `2` for invalid arguments or
format, `3` for a strict-mode partial collection, and `4` when a requested fix
was not applied.

## Architecture and order

Reuse `internal/status` collectors through a small `internal/doctor` registry:

```text
internal/status: host, storage, Ubuntu, SSH, network, toolchain collectors
internal/doctor: registry, severity rules, evidence, suggestions,
                 safe fixes and report renderer
```

Implementation order is contract and redaction, host/storage/setup, Ubuntu and
SSH, network and toolchains, report rendering, safe fixes, then integration
with future sessions/projects/services.

## Tests and completion

Unit tests cover every status, partial failure, severity, suggestions, exit
codes, JSON, redaction, confirmation and destructive-fix prevention.
Integration tests cover clean/incomplete setup, missing/inaccessible Ubuntu,
SSH states, occupied ports, stale PIDs, invalid configuration, missing IP,
low storage, absent and incomplete toolchains, `--json` and `--strict`.
Real-device tests cover ARM64 Termux, interruption, Termux:API absence,
HyperOS suspension, Wi-Fi changes and safe fixes.

The command is complete only when it diagnoses host, storage, setup, Ubuntu,
SSH, network and toolchains, survives partial failures, explains evidence,
produces valid JSON, does not reveal secrets or modify normal mode, and passes
on real Termux ARM64.
