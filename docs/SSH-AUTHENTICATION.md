# SSH Authentication and Connection Approval

**Status:** active future design note; the current MVP still uses its existing
Termux password flow and must not expose SSH directly to the internet.

This document compares two future authorization layers: approving a connection
on the phone and a short-lived one-time code. Neither replaces basic network,
key, credential and logging protections.

## Current context

Mobdesk runs a dedicated SSH server on port `8022` in the same native Termux
workstation used by local commands. Current authentication is the Termux password.
The future goal is to reduce dependence on a fixed password and let the phone
owner decide which computers may connect.

## Phone approval

When a computer starts an SSH connection, the phone TUI shows the device, IP,
time and fingerprint. The user can approve, reject or block it. Approval must
bind to the exact connection attempt, expire quickly, default to denial, and
never be reusable for another device or session. An SSH `Banner` alone is not
authentication; approval must participate in a real authentication or
authorization step.

## One-time code

The phone generates a short random code with a short lifetime, one use, a
cryptographically secure generator, attempt limits, cancellation/expiry
invalidations and a binding to user, device and connection. The value is shown
only on the phone and never logged. A random challenge tied to the connection
is preferred over independent TOTP because clock synchronization and capture
are risks.

## Comparison and recommendation

Phone approval has the clearest normal-session UX and best context. A one-time
code is a useful pairing and recovery fallback but is more exposed during entry.
Prioritize phone approval, use the code for first pairing/recovery, then install
an Ed25519 key and optionally retain approval as a second layer.

Suggested policy: deny new devices by default, identify keys by fingerprint,
expire pending requests, rate-limit failures, keep SSH on local network or
Tailscale, and never reuse approval.

## Stage boundary

MVP-1 keeps dedicated SSH on `8022`, password authentication and the warning
not to expose it publicly. Post-MVP work may add device pairing, automatic keys,
TUI requests, approval/rejection, one-time fallback, revocation, expiry,
rate-limiting and auditing. OpenSSH methods and `AuthenticationMethods` are
documented in [`sshd_config(5)`](https://man.openbsd.org/sshd_config).
