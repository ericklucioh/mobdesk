# Security Audit - 2026-07-25

**Status:** historical audit with follow-up results; non-authoritative for new
architecture decisions. [`SECURITY.md`](../SECURITY.md) and the current code
and tests are authoritative.

## Executive summary

Mobdesk is a Go CLI/TUI for Termux that provisions Ubuntu through PRoot-Distro,
manages SSH on port `8022`, installs profiles and updates its binary. The audit
found **0 critical, 3 high, 5 medium, 2 low and 2 informational findings** at
the original review point. It was not ready for production on an untrusted
network: the main blockers were password SSH on all interfaces, updater
authenticity, and TUI coordination/cancellation.

## Attack surface

Inputs include CLI flags, local JSON state, persisted logs, HOME/PREFIX,
GitHub responses, Termux API output and SSH traffic. Sensitive data includes
Termux passwords, `authorized_keys`, command logs and installation state.
Privileged operations include `pkg`, Ubuntu `apt-get`, SSH process lifecycle,
configuration writes and replacing the executable.

## Findings

### High

- **H-01: SSH LAN exposure.** The reviewed configuration used
  `0.0.0.0`, password and keyboard-interactive authentication, and
  `StrictModes no`. Recommended fix: loopback and keys by default, explicit
  LAN/password opt-in, `StrictModes yes`, `MaxAuthTries`, `LoginGraceTime` and
  ownership-aware users. This remains a release blocker until the policy is
  implemented and tested.
- **H-02: Binary and checksum can be compromised together.** SHA-256 detects
  accidental corruption but does not authenticate a release. Sign the checksum
  manifest with a trusted key embedded in the binary and reject unsigned or
  untrusted signatures before downloading/installing the binary.
- **H-03: Interrupted update availability.** A process interruption could leave
  the launcher without the main executable between renames. **Status at the
  audit date: corrected.** Atomic replacement and legacy `.bak` recovery were
  added with regression tests.

### Medium

- **M-01: TUI concurrency.** Refreshes and operations could complete out of
  order. **Corrected:** host operations are blocked while busy, and monotonic
  IDs discard stale snapshots.
- **M-02: TUI child cancellation.** Subprocesses lacked lifecycle cancellation
  and deadlines. **Corrected:** the real backend uses a cancellable TUI context
  and `exec.CommandContext` for operations, status and shell.
- **M-03: Update download limits.** The original client lacked global timeout,
  manifest limits and binary-size limits. Add these before release.
- **M-04: Unbounded logs and persisted path.** The original reader trusted a
  persisted `LogPath` and loaded full files. **Partially corrected:** canonical
  catalogue names now derive paths below `InstallLogsDir`; bounded reading is
  still a policy task.
- **M-05: CLI JSON errors lost in TUI.** Non-zero child exits could hide the
  structured error. The JSON presentation and TUI parsing work from the final
  structured result; retain regression coverage.

### Low and informational

- **L-01:** commands with no positional arguments should reject unexpected args.
- **L-02:** human log rendering should neutralize terminal control sequences.
- **L-03:** relevant file and PTY close errors should be handled deliberately.
- **I-01:** dead symbols and deprecated APIs reduce maintenance signal.
  **Corrected:** symbols were removed, APIs updated and relevant write errors
  propagated; the linter subsequently passed.
- **I-02:** dependency updates require changelog and compatibility triage, not
  blind upgrades.

## Supply chain and test gaps

Actions were not pinned by SHA and development images/tools used `latest` at the
review point. Recommended defenses are pinned Actions, release signatures,
`go test -race`, `govulncheck`, coverage tracking and reproducible images.
Remaining test gaps include interrupted updates, slow/large downloads, log
limits, secure SSH defaults, concurrent TUI operations and child cancellation.

The dynamic follow-up recorded successful module verification, vet, normal,
race, shuffled and integration tests; coverage was 57.4% at that time. The
available linter initially reported diagnostics and later passed after the
listed cleanup. Docker still does not reproduce Android permissions,
Termux:API, HyperOS suspension or real ARM64 PRoot.

## Prioritized work

1. Make SSH safe by default and authenticate releases independently.
2. Keep update replacement recoverable and bounded.
3. Finish TUI serialization/cancellation, download and log limits, and JSON
   error coverage.
4. Add race, vulnerability, filesystem-failure and provenance tests; pin CI
   actions and development images.

No secrets were found in the repository. This historical audit did not perform
real-device validation and should not be read as a current security approval.
