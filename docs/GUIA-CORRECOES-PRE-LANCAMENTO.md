# Pre-Release Remediation Guide

**Status:** active security/operations guidance; proposed changes do not imply
that the corresponding application behavior is already implemented.

This guide addresses release authenticity, update recovery, SSH defaults and
the Termux/Ubuntu boundary. It complements the historical
[`AUDITORIA-SEGURANCA-2026-07-25.md`](AUDITORIA-SEGURANCA-2026-07-25.md) and
[`AUTENTICACAO-SSH.md`](AUTENTICACAO-SSH.md).

## Recommended order

1. Correct Termux/Ubuntu detection.
2. Make binary replacement recoverable.
3. Sign release artifacts.
4. Define and enforce safe SSH defaults.

Keep these changes separate because SSH UX changes and release signatures have
different product and operational risks.

## 1. Release authenticity

The updater must not trust a binary and checksum downloaded from the same
untrusted release channel. Publish `SHA256SUMS` and `SHA256SUMS.minisig`, sign
the exact manifest, embed the public key in `internal/update`, and verify the
signature before extracting a hash or opening the binary. The private key stays
outside the repository and CI logs. Prefer a maintained Go Minisign verifier
over an external Termux binary.

Release workflow requirements:

- build reproducible `linux/arm64` output;
- calculate hashes from that run's artifacts;
- sign after all hashes are generated;
- fail when any asset or signature is missing;
- pin GitHub Actions by SHA.

Tests must reject missing, invalid, wrong-key and unsigned signatures, and must
prove a binary is not opened or installed when verification fails.

## 2. Recoverable update

Keep the active binary, `.bak` and a same-directory temporary download:

```text
mobdesk       active executable
mobdesk.bak   last known-good version
.mobdesk-new-* temporary download
```

Recover `.bak` when the main binary is absent. Download with private
permissions, verify signature/hash/size, validate `version --json` with a short
context, then replace atomically and keep the backup until safe cleanup. Any
failure restores or preserves a known-good executable. Never delete projects,
logs or Ubuntu to make room. Enforce timeout, manifest and binary-size limits,
and propagate cancellation.

Tests cover normal replacement, failed self-test, failures between renames,
startup with only `.bak`, low storage, permission failure and cancellation.

## 3. Safe SSH defaults

New installations should use keys and `StrictModes yes`, bind to
`127.0.0.1` by default, and require explicit confirmation for LAN exposure or
password fallback. A secure base includes:

```text
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitEmptyPasswords no
PubkeyAuthentication yes
StrictModes yes
MaxAuthTries 3
LoginGraceTime 30
MaxStartups 10:30:60
```

Validate `sshd -t -f` before installing configuration. Store options privately,
accept only validated OpenSSH public keys, use `.ssh` `0700` and
`authorized_keys` `0600`, and never log keys, passwords or temporary codes.
Status should make LAN exposure and password opt-in visible.

PRoot compatibility must be checked on real hardware without disabling
`StrictModes`. Test loopback defaults, explicit LAN, explicit password, invalid
configuration and key login.

## 4. Termux/Ubuntu boundary

`setup`, `start`, `stop`, `install` and `update` are Termux-only. A PRoot SSH
session may inherit `TERMUX_VERSION` or `PREFIX`; those variables alone must
not classify it as Termux. Check `/etc/os-release` for Ubuntu first, then use
Termux markers. Unknown environments fail conservatively and block host actions.

The SSH wrapper may clear host-only variables, but must not casually clear HOME,
USER, TERM or PATH. Unit and integration tests must cover inherited markers,
real host markers, missing markers and remote `status --json`/TUI behavior.

## Final device validation

On real ARM64 Termux, test setup, wake-lock, start/stop, key login, refusal of
default password, remote TUI blocking, signed update, invalid artifact and
interrupted update recovery. Record device, Android, Termux and result details
in the release checklist. Do not call the product ready for third-party remote
use until all four areas pass.
