# Security Audit TODO

Based on `docs/SECURITY-AUDIT-2026-07-25.md`.

## Pending Decisions

### H-01: SSH access policy

**Decision:** define network exposure and the default authentication method.

| Option | Benefits | Costs and risks |
| --- | --- | --- |
| Loopback + SSH keys by default (recommended) | Smallest attack surface; removes remote brute force; preserves tunnel access. | Requires initial key creation and management; does not provide immediate LAN access. |
| LAN + SSH keys by default | Allows another device on the same network to connect without a password. | The port remains exposed on the LAN; requires clear guidance about untrusted networks. |
| LAN + opt-in password | Lower friction for beginners and preserves the current flow when explicitly selected. | Passwords can face brute force; requires warnings, attempt limits and explicit confirmation. |

**Why this matters:** the choice changes first-access behavior and the accepted
risk level on university, public or shared Wi-Fi.

### H-02: Release authenticity

**Decision:** choose the signing mechanism and custody model for the private
release key.

| Option | Benefits | Costs and risks |
| --- | --- | --- |
| Minisign (recommended) | Small, simple for a single binary and easy to verify with an embedded public key. | Requires protecting and backing up the private key; adds signing to the release workflow. |
| Cosign/Sigstore | Integrates well with provenance and CI OIDC identities. | More complex ecosystem for users and a local binary; may be excessive for the MVP. |
| GPG | Familiar and widely supported. | More complex key UX and management; higher operational error risk. |

**Why this matters:** a checksum without a signature does not protect against a
compromised release or GitHub account. The choice defines asset formats and the
publishing workflow.

### M-03: Updater limits

**Decision:** define timeout, maximum download size and redirect policy.

| Option | Benefits | Costs and risks |
| --- | --- | --- |
| 2 minutes total, 64 MiB binary, 1 MiB checksums, up to 3 redirects (recommended) | Protects storage and avoids a blocked TUI; suitable for a mobile ARM64 binary. | May reject a future release above the limit; requires adjustment as the binary grows. |
| Caller context timeout only, without a size limit | Minimal and flexible implementation. | Retains infinite-connection and storage-exhaustion risks. |
| Very strict limits such as 30 seconds and 16 MiB | Strong defense against abuse and resource consumption. | Fails on slow mobile networks or normal releases. |

**Why this matters:** these values are product policy and must balance mobile
connectivity, release size and protection of phone resources.

### M-04: Log read limit

**Decision:** define how many lines and bytes `mobdesk logs` may read per file.

| Option | Benefits | Costs and risks |
| --- | --- | --- |
| 200 lines and 1 MiB per log (recommended) | Keeps diagnostics useful without loading huge logs into memory. | May truncate context from older failures. |
| Lines-only limit | Simple UX. | The current implementation still reads the entire file; memory use remains unresolved. |
| 1,000 lines and 8 MiB | Retains more diagnostic context. | Increases memory use and terminal output. |

**Why this matters:** the limit defines diagnostic UX and protection from logs
created by installers or tools with excessive output.

### CI: Validation and supply-chain policy

**Decision:** define mandatory gates and provenance requirements for automation.

| Option | Benefits | Costs and risks |
| --- | --- | --- |
| `go test -race`, `govulncheck`, linter, SHA-pinned Actions and informative coverage (recommended) | Adds meaningful defense without blocking PRs on an arbitrary coverage target. | CI is slower; Action SHAs require periodic maintenance. |
| Same set plus a 70% minimum coverage | Prevents measurable coverage regression. | Current coverage is lower; may encourage superficial tests to hit a number. |
| Keep the current CI | Lowest cost and runtime. | Does not detect races in CI and leaves Action supply-chain risk unmitigated. |

**Why this matters:** this defines the maintenance cost accepted in exchange for
detecting regressions and vulnerabilities before release.
