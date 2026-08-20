# Next Features and User Experience Plan

**Status:** superseded historical plan. It assumes the retired PRoot/Ubuntu
runtime and application-configuration scope. Use
[`POST-TERMUX-SPRINTS.md`](POST-TERMUX-SPRINTS.md) for planned native Termux
work.

**Primary language:** English

**Supported second language:** Brazilian Portuguese (`pt-BR`)

**Scope:** improve the current Mobdesk MVP after the English-first localization
work, add clipboard support over SSH, improve the mobile TUI, expand the app
catalogue with the first requested languages and tools, and add an optional
Go configuration profile.

## 1. Product Intent

Mobdesk should make a phone-based development environment practical for class,
travel and small development projects. The next work must improve the daily
experience before expanding into a large catalogue or a complete project
manager.

The implementation must preserve the current architecture:

```text
Termux host
  -> Mobdesk CLI/TUI
  -> SSH or local shell
  -> Ubuntu through PRoot-Distro
  -> development tools and user projects
```

The next release should make three things better:

1. Move text from an SSH session to the connected computer clipboard safely.
2. Make the TUI comfortable on a phone-sized terminal.
3. Install the most relevant development languages and tools without creating
   unsafe or non-reproducible installation paths.

## 2. Current Baseline

The current code already provides:

- Termux/Ubuntu runtime separation;
- resumable setup and persistent Ubuntu through PRoot;
- SSH on port `8022`;
- CLI and TUI using the same JSON operation contract;
- an app catalogue based on `AppProfile`;
- installation, status, provenance and safe uninstallation;
- optional Neovim/LazyVim configuration;
- English-first localization with `en-US` and `pt-BR`;
- bilingual CLI and Bubble Tea TUI;
- `make i18n-check` and `make check`.

The following validation remains external:

- expand manual validation across real Termux/Android devices;
- clipboard behavior with the target desktop SSH terminal;
- mobile touch behavior on the target terminal dimensions;
- ARM64 package availability and installation time for every new profile.

## 3. Priorities

| Priority | Work | Reason |
|---|---|---|
| P0 | Expand validation across real Termux/Android devices | Do not expand the next roadmap stage without measuring the current flow across supported device variants |
| P1A | SSH clipboard with OSC 52 | Fixes a concrete daily workflow problem reported in the TODO |
| P1B | Mobile TUI redesign | The current interface is visually dense and difficult to use by touch |
| P2 | Multi-package and multi-executable catalogue foundation | Required by Java, Rust, Rails and other profiles |
| P3 | SQLite, Ruby and Rails | Smallest useful expansion and Rails depends on Ruby |
| P4 | Rust | High educational value and validates multi-package installation |
| P5 | Java | Required for the stated computer-science course use case |
| P6 | Zig | Requires a verified architecture-specific archive instead of apt |
| P7 | Composer | Useful for PHP/Laravel workflows, but dependent on PHP package behavior |
| P8 | Optional Go pure-Go configuration | Makes the CGO limitation explicit without silently changing user behavior |

The four first requested languages are interpreted as:

1. Ruby
2. Rust
3. Java
4. Zig

Composer and Rails are apps. SQLite is a standalone database command-line tool.

## 4. Global Rules

- English remains the default user-facing language.
- Every new user-visible message uses an existing or new `MessageID`.
- Portuguese translations are added to both locale catalogues.
- Commands, flags, JSON keys, schema versions, states, paths and profile IDs
  remain stable machine values.
- Installation runs inside Ubuntu through `proot-distro`.
- The TUI never calls `apt`, `gem`, `cargo`, `proot-distro` or scripts directly.
- User input is validated before it becomes a command argument.
- Destructive actions require confirmation.
- Installation is idempotent and records provenance.
- Shared packages are not removed automatically.
- External command output is not translated or persisted as a translated
  message.
- New profiles require storage estimates, aliases, version checks and tests.
- Broader real Termux validation is required before claiming support across a
  wider device matrix.
- `make catalog-test` is run when catalogue, installation strategy, PRoot
  behavior or catalogue smoke scripts change.

## 5. P0 - Real Device Validation

### Objective

Expand validation of the already implemented MVP across real Android devices
before adding large new runtime behavior.

### Manual scenarios

#### Base environment

1. Run setup from a clean or disposable Termux environment.
2. Confirm Ubuntu ARM64 installation.
3. Confirm workspace creation and private state permissions.
4. Start and stop the Mobdesk SSH server.
5. Open `mobdesk shell` locally.
6. Connect from another computer through SSH.
7. Test the TUI inside Termux and inside the SSH/Ubuntu session.

#### App lifecycle

1. Open the Apps screen.
2. Open the Neovim details view without installing from the row.
3. Install Neovim.
4. Apply LazyVim from the details flow.
5. Verify headless Neovim startup.
6. Modify a managed file.
7. Remove the configuration.
8. Confirm the modified file is preserved.
9. Uninstall Neovim.
10. Confirm detected applications cannot be removed without provenance.

#### Localization

1. Run the TUI with `--locale en-US`.
2. Run the TUI with `--locale pt-BR`.
3. Run with `MOBDESK_LOCALE=pt-BR`.
4. Confirm child CLI operations keep the selected locale.
5. Test a narrow phone terminal in both languages.

### Acceptance criteria

- No data loss occurs during install, configuration removal or uninstall.
- Host-only operations are blocked inside Ubuntu/SSH with a localized message.
- The TUI is usable by touch and keyboard.
- All failures are recorded with enough information for follow-up.

## 6. P1A - Clipboard Over SSH

### Problem

Text copied from OpenCode or another program inside an SSH session does not
reach the clipboard of the connected computer. The current PTY path forwards
terminal bytes but has no explicit clipboard primitive.

### Proposed command

```bash
printf 'clipboard test\n' | mobdesk copy --stdin
```

The command emits an OSC 52 sequence. The SSH connection transports the bytes
unchanged, and the local terminal emulator decides whether to write them to the
computer clipboard.

### Contract

Command name:

```text
mobdesk copy
```

Supported input:

- `--stdin` is required for the first implementation.
- Positional text is intentionally not supported initially because it leaks
  copied content into shell history and process arguments.
- The command is allowed inside Ubuntu and SSH; it must not require the Termux
  host runtime.
- `--json` returns the operation result without writing control sequences to
  JSON stdout. The OSC 52 sequence is emitted only in normal text mode.
- Default locale and `--locale` follow the existing CLI contract.

### Implementation

Add:

```text
internal/clipboard/osc52.go
internal/clipboard/osc52_test.go
internal/cobra/copy.go
```

Use `github.com/aymanbagabas/go-osc52/v2`, already listed as an intended
project dependency. Add the pinned dependency to `go.mod` and `go.sum`.

The clipboard package should:

- accept an `io.Reader` and `io.Writer`;
- validate UTF-8;
- reject empty input;
- reject input over 4096 bytes;
- emit Base64 UTF-8 payloads only;
- never emit raw ESC or BEL bytes from the input inside the encoded payload;
- report writer errors and short writes;
- never log or persist copied content.

### Terminal modes

Initial mode:

- direct SSH and normal terminals use OSC 52 directly.

Later explicit modes may support:

- tmux wrapping;
- GNU screen wrapping.

Do not guess terminal mode from `$TMUX` in the first implementation. Do not
implement clipboard query support because that would send the computer
clipboard back into the remote session.

### Security and limitations

- Any trusted process in the SSH session can request an OSC 52 clipboard write.
- The command must explain that success means the sequence was emitted, not
  that the terminal definitely accepted it.
- Unsupported terminals may ignore the sequence or display it as text.
- Termux, tmux, SSH client and desktop terminal versions must be tested on the
  target device.

### Tests

- ASCII payload.
- Unicode payload.
- Newlines and terminal punctuation.
- Empty input rejection.
- Invalid UTF-8 rejection.
- 4096-byte accepted input.
- Oversized input rejection without output.
- Writer failure and short write.
- JSON stdout purity.
- Operation allowed from Ubuntu/SSH runtime.
- Interactive SSH smoke test.

## 7. P1B - Mobile TUI Redesign

### Problem

The current TUI contains too many explanatory cards, compact inline actions and
small controls. This makes the interface difficult to read and operate by touch
on a phone.

### Design goals

- Make the primary action obvious within one screen.
- Use larger vertically stacked controls.
- Separate neighboring actions with visible blank space.
- Remove explanatory text that does not help the current decision.
- Keep required boundary, conflict and safety explanations.
- Preserve keyboard equivalence for every mouse/touch action.
- Keep all labels localized.

### Responsive tiers

| Width | Layout |
|---:|---|
| 20-23 columns | Compact labels, one column, no secondary descriptions |
| 24-31 columns | Full-width stacked controls, abbreviated metadata |
| 32-39 columns | Readable full-width controls and compact details |
| 40-63 columns | Full labels and selected secondary details |
| 64+ columns | Optional informational cards; primary actions remain stacked |

Treat 20 columns as the rendering floor and 32 columns as the minimum
comfortable target.

### Home screen

Keep:

- header and workstation state;
- one large Start/Stop button;
- separated navigation buttons for Setup, Status, Apps, Shell and System;
- compact SSH information only when the workstation is running.

Remove:

- repeated descriptions under navigation cards;
- redundant explanatory paragraphs;
- secondary cards that duplicate the button label.

Fix the existing focus-count issue where the home screen exposes a focus index
that has no corresponding action.

### Apps screen

Each app row should be a large, separated, three-row target:

```text
┌──────────────────────────────┐
│ Neovim                       │
│ Installed                    │
│                              │
└──────────────────────────────┘
```

Remove descriptions from the list. Keep descriptions in the details view.
Tapping or pressing Enter opens details and never installs directly.

### App details view

Order content as:

1. App name and one-line description.
2. State, source and version.
3. Primary action.
4. Optional configuration action.
5. Destructive actions in a separate section.
6. Close/Back action.

Move dependencies, paths, plugins and storage into a compact Details section or
hide secondary values on narrow terminals. Disabled actions show their reason
under the control but must not look enabled.

### Confirmation view

Replace compact inline text such as `[ Y ] Yes [ N ] No` with two full-width
controls:

```text
┌──────────────────────────────┐
│ Confirm                      │
└──────────────────────────────┘

┌──────────────────────────────┐
│ Cancel                       │
└──────────────────────────────┘
```

Mouse and keyboard must activate the same action IDs.

### Other screens

- Setup: keep actions and progress steps; remove default advanced explanations.
- Status: keep overall state, key cards, refresh and back; collapse metadata on
  narrow terminals.
- Shell: keep two large actions; remove duplicate descriptions.
- System: keep update actions and result; remove repeated update guidance.
- Operation: keep title, live progress and one concise wait message.

### Hit-testing architecture

Replace text-search hit-testing with explicit regions:

```go
type HitRegion struct {
    ID      string
    X       int
    Y       int
    Width   int
    Height  int
    Enabled bool
}
```

Each screen produces regions from the layout it renders. Mouse coordinates map
to action IDs, and keyboard focus uses the same IDs. This prevents localized
wrapping, borders, padding and viewport offsets from breaking interaction.

### TUI tests

- widths 20, 24, 32, 40, 64 and 80;
- heights around 20, 24 and 30 rows;
- English and Portuguese labels;
- button interior, border, padding and blank gaps;
- popup actions and confirmations;
- disabled action reasons;
- row opens details without installing;
- keyboard focus order and count;
- scroll offset and mouse coordinates;
- drag versus tap behavior;
- remote runtime restrictions.

## 8. P2 - Catalogue Foundation

### Required model changes

The current `AppProfile` has one `Package` and one `Executable`. Add additive
metadata:

```go
Packages            []string `json:"packages,omitempty"`
RequiredExecutables []string `json:"required_executables,omitempty"`
```

Compatibility rules:

- If `Packages` is empty, use the existing `Package` field.
- If `RequiredExecutables` is empty, use the existing `Executable` field.
- Persist every installed package in installation records.
- Uninstall only packages recorded as Mobdesk-managed.
- Shared-package protection compares the complete package list.
- Status marks a profile installed only when all required executables are
  present, unless the profile explicitly declares an optional executable.

### Catalogue schema tests

- Single-package legacy profile.
- Multi-package profile.
- Multiple required executables.
- Alias resolution.
- Missing executable detection.
- Shared package protection.
- Legacy record loading.
- JSON additive compatibility.

## 9. P3 - SQLite, Ruby and Rails

### SQLite profile

| Field | Value |
|---|---|
| Canonical name | `sqlite` |
| Package | `sqlite3` |
| Executable | `sqlite3` |
| Version check | `sqlite3 --version` |
| Strategy | `apt` |
| Estimate | 3-12 MB total planning range |

Smoke test:

```bash
sqlite3 test.db 'create table hello (value text); insert into hello values ("ok"); select value from hello;'
```

The database must be created in a temporary fixture path and removed after the
test.

### Ruby profile

| Field | Value |
|---|---|
| Canonical name | `ruby` |
| Aliases | `ruby-full` |
| Package | `ruby-full` |
| Required executables | `ruby`, `gem` |
| Version check | `ruby --version` |
| Strategy | `apt` |
| Estimate | 65-170 MB total planning range |

Smoke test:

```bash
printf 'puts 2 + 2\n' > main.rb
ruby main.rb
```

### Rails profile

| Field | Value |
|---|---|
| Canonical name | `rails` |
| Aliases | `ruby-on-rails` |
| Package | `ruby-rails` |
| Required executables | `ruby`, `rails` |
| Version check | `rails --version` |
| Strategy | `apt` |
| Dependency | `ruby` |
| Estimate | 65-220 MB total planning range |

Use the Ubuntu package in the first version. Do not run `gem install rails`
without a pinned gem strategy, native build validation and a dedicated removal
policy.

## 10. P4 - Rust

| Field | Value |
|---|---|
| Canonical name | `rust` |
| Aliases | `rustc` |
| Packages | `rustc`, `cargo` |
| Required executables | `rustc`, `cargo` |
| Version check | `rustc --version` |
| Strategy | `apt` |
| Estimate | 70-230 MB total planning range |

Smoke test:

```rust
fn main() {
    println!("hello from rust");
}
```

Compile and run the fixture inside Ubuntu. Record installation time and
dependency packages on ARM64.

## 11. P5 - Java

| Field | Value |
|---|---|
| Canonical name | `java` |
| Aliases | `openjdk` |
| Package | `openjdk-21-jdk` |
| Required executables | `java`, `javac` |
| Version check | `java --version` |
| Strategy | `apt` |
| Estimate | 160-340 MB total planning range |

Smoke test:

```java
class Main {
    public static void main(String[] args) {
        System.out.println("hello from java");
    }
}
```

Compile with `javac Main.java` and run with `java Main` inside Ubuntu.

## 12. P6 - Zig

Ubuntu does not provide a reliable `zig` package for the target baseline. Use a
verified release archive instead of an unpinned download.

Requirements:

- pinned Zig version;
- separate `aarch64` and `x86_64` archive names;
- SHA-256 checksum for every supported architecture;
- architecture validation using `uname -m`;
- HTTPS-only repository and release URL;
- managed executable path under the Ubuntu user profile;
- recorded installed file hash;
- safe uninstall preserving modified files.

Smoke test:

```zig
const std = @import("std");

pub fn main() !void {
    try std.io.getStdOut().writer().print("hello from zig\n", .{});
}
```

Compile and run on both x86_64 fixture and ARM64 device when available.

## 13. P7 - Composer

| Field | Value |
|---|---|
| Canonical name | `composer` |
| Package | `composer` |
| Executable | `composer` |
| Version check | `composer --version` |
| Strategy | `apt` |
| Estimate | 28-90 MB total planning range |

The Ubuntu Composer package brings the PHP runtime dependencies required for
Composer. A separate PHP application profile is not required for this phase,
but the documentation must state that Composer does not install a complete
Laravel runtime by itself.

Smoke test:

```bash
composer --version
```

## 14. P8 - Optional Go Pure-Go Configuration

### Problem

The Go toolchain may use cgo by default. Minimal Ubuntu/PRoot environments may
not contain the C headers required by cgo, causing network or native packages to
fail during compilation.

### Decision

Do not apply `CGO_ENABLED=0` automatically during Go installation. It changes
the behavior of every Go build and can break packages that require C.

Expose an optional Mobdesk configuration profile named `go-pure` or
`go-no-cgo`:

- description explains the tradeoff;
- writes `CGO_ENABLED=0` to the Go environment file;
- validates with `go env CGO_ENABLED`;
- refuses to overwrite an existing user configuration;
- records provenance and file hashes;
- preserves a manually modified file during removal;
- is available from CLI and the app popup only after Go is installed.

Expected user flow:

```text
Install Go
  -> show optional recommendation
  -> user chooses Add pure-Go configuration
  -> Mobdesk writes ~/.config/go/env
```

The install result may mention the recommendation, but must not silently change
the user's Go environment.

## 15. Smoke Test Expansion

Update the catalogue smoke tests to cover:

- SQLite installation and SQL execution;
- Ruby installation and script execution;
- Rails installation and version output;
- Rust installation, compilation and execution;
- Java installation, compilation and execution;
- Zig architecture-specific installation, compilation and execution;
- Composer version output;
- repeated installation of every new profile;
- status JSON and provenance records;
- uninstall safety and shared dependencies.

Run the expanded catalogue test only after the catalogue or installation code
changes. Use a longer timeout if cold ARM64 installation exceeds the current
default.

## 16. Documentation Updates

Update after implementation:

- `README.md` and `README.pt-BR.md` with supported tools and examples;
- `.github/README.md` and `.github/README.pt-BR.md` for new-user discovery;
- `docs/ARCHITECTURE.md` for multi-package installation and verified archives;
- `docs/DECISIONS.md` for apt package strategy and no-initial-gem decision;
- `docs/ROADMAP.md` for expanded Stage 2 support;
- `docs/ideas/TOOL-CATALOG.md` to promote supported tools;
- `docs/ideas/TODO.md` to mark completed requests and link this plan;
- `docs/APP-CONFIGURATION-REFACTOR-PLAN.md` for Go configuration behavior;
- `docs/APP-CONFIGURATION-IMPLEMENTATION-PLAN.md` for catalogue phases;
- `docs/DOCTOR-PLAN.md` for future diagnostics of new profiles.

## 17. Proposed Commit Sequence

Each phase must have focused tests, documentation updates, a reviewed diff and
one isolated commit.

1. `feat: add ssh clipboard command`
2. `feat: redesign mobile tui controls`
3. `feat: support multi-package app profiles`
4. `feat: add sqlite ruby and rails profiles`
5. `feat: add rust app profile`
6. `feat: add java app profile`
7. `feat: add zig app profile`
8. `feat: add composer app profile`
9. `feat: add optional go pure configuration`
10. `test: validate expanded app catalogue`

Do not combine the TUI redesign with catalogue behavior in the same commit.
Do not combine clipboard security behavior with unrelated app profiles.

## 18. Final Acceptance Criteria

### Clipboard

- `printf 'test\n' | mobdesk copy --stdin` emits a valid OSC 52 sequence.
- Input is bounded, validated and never persisted.
- The command works inside Ubuntu/SSH.
- JSON output never contains terminal control bytes.

### TUI

- Important controls are visually large and separated.
- The home screen has no dead focus position.
- App rows open details without installing directly.
- Destructive actions remain clearly separated and confirmed.
- Keyboard, mouse and touch use the same action IDs.
- English and pt-BR fit supported narrow widths.

### Catalogue

- Ruby, Rust, Java and Zig are installable profiles.
- SQLite, Composer and Rails are installable profiles.
- Multi-package and multi-executable profiles are supported.
- Java and Rust compile fixtures successfully.
- Zig uses pinned architecture-specific checksums.
- Rails does not depend on unpinned `gem install`.
- Shared packages are protected during uninstall.
- Status and JSON remain schema-compatible.

### Go configuration

- The pure-Go recommendation is visible after Go installation.
- No automatic `CGO_ENABLED=0` change occurs.
- Optional configuration apply/remove is safe and provenance-aware.
- Existing user Go configuration is never overwritten.

### Release validation

- `make check` passes.
- `make i18n-check` passes.
- Expanded catalogue smoke tests pass where applicable.
- Real Termux validation and the remaining device matrix are documented.
