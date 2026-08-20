# English-First Localization and Documentation Migration Plan

**Status:** Phase 7 complete in the automated environment and validated on a
real Android/Termux device; broader device validation remains ongoing.

**Primary locale:** `en-US`

**Additional locale:** `pt-BR`

**Scope:** localize the Mobdesk CLI, TUI, services, tests, scripts and project
documentation. English is the canonical language for code and technical
documentation. Public onboarding documents have an English version and a
Brazilian Portuguese version.

## 1. Goals

- Make English the default language for every user-visible Mobdesk flow.
- Provide native Brazilian Portuguese support for the CLI and TUI.
- Keep command names, flags, JSON keys, state values and paths stable.
- Remove direct user-visible strings from Go presentation code.
- Move user-visible messages into typed message IDs and locale catalogs.
- Rename Portuguese identifiers, comments, test names and documentation paths
  to English where they exist.
- Preserve existing persisted state and JSON schema compatibility.
- Make locale selection work in local Termux and SSH sessions.
- Keep the TUI mouse and keyboard hit-tests correct in both languages.

## 2. Non-Goals

- Translating arbitrary output produced by `apt`, `npm`, `pipx`, Go or other
  external commands.
- Translating machine identifiers, command names, flags, paths or JSON enum
  values.
- Adding a remote translation service or runtime download of translations.
- Adding more locales during this migration.
- Rewriting historical logs or persisted human-readable error text.
- Changing the product scope beyond localization and documentation migration.
- Persisting a language preference in the first implementation. Explicit flags
  and environment variables are the initial contract.

## 3. Locked Localization Contract

### 3.1 Supported locales

- `en-US` is the default and fallback locale.
- `pt-BR` is the supported Brazilian Portuguese locale.
- `en`, `en_US` and `en_US.UTF-8` normalize to `en-US`.
- `pt`, `pt_BR` and `pt_BR.UTF-8` normalize to `pt-BR`.
- `C` and `POSIX` normalize to `en-US`.
- Unsupported environment values fall back to `en-US`.
- An explicitly invalid `--locale` value returns a localized validation error.

### 3.2 Locale precedence

Use the first valid value in this order:

1. Global `--locale` flag.
2. `MOBDESK_LOCALE`.
3. `LC_ALL`.
4. `LC_MESSAGES`.
5. `LANG`.
6. `en-US`.

The TUI must pass the selected locale explicitly to child Mobdesk CLI commands.
This keeps rendering and command output consistent across local and SSH
sessions in the same Termux workstation.

### 3.3 Stable machine contract

The following values remain English and stable:

- Commands: `start`, `stop`, `setup`, `status`, `install`, `uninstall`,
  `config`, `update`, `version`, `tui` and `shell`.
- Flags: `--json`, `--progress`, `--locale` and existing flags.
- JSON keys and schema version.
- State values such as `installed`, `failed`, `detected`, `healthy`,
  `conflict`, `modified` and `completed`.
- Source values such as `mobdesk` and `detected`.
- Error codes, event names, setup phase IDs and profile IDs.
- Persistent paths, environment variables, package names and executables.

Human `message` fields may be localized. Automation must use stable fields such
as `success`, `state`, `command`, `action`, `message_id` and `error_code` rather
than parsing localized prose.

## 4. Proposed Architecture

Create a small `internal/i18n` package with no dependency on Cobra or Bubble
Tea.

```text
internal/i18n
  locale.go       locale parsing and precedence
  message.go      MessageID and Localizer
  catalog.go      embedded catalog loading and validation
  locale/
    en-US.json
    pt-BR.json
```

Recommended API:

```go
type Locale string

const (
    LocaleENUS Locale = "en-US"
    LocalePTBR Locale = "pt-BR"
)

type MessageID string

type Localizer struct {
    Locale Locale
}

func Resolve(explicit string, environ func(string) string) (Locale, error)
func (l Localizer) Text(id MessageID, data any) string
```

Use `go:embed` for immutable catalogs. Use `golang.org/x/text/language` for
BCP 47 parsing and matching if its dependency cost remains acceptable after
the implementation spike. Do not use package-global mutable locale state.

For the current two-locale scope, a typed internal catalog is preferred over
`go-i18n`. `go-i18n` remains a future option if plural rules, translator
workflows or many additional locales justify its tooling.

## 5. Message Rules

- Every user-visible string must be represented by a `MessageID`.
- Message IDs use stable English names such as `tui.apps.title` and
  `error.termux_required`.
- IDs, not source-language prose, are stored in domain models.
- Catalog entries may contain named template data for paths, names, versions
  and counts.
- State values are translated only at the presentation boundary.
- Internal errors expose stable codes and structured details.
- External command output remains original output, with localized context around
  it when necessary.
- Catalog completeness is tested for both locales.
- Missing translations fail catalog validation instead of silently rendering a
  source-language fallback.

## 6. Phase Order

Every phase ends with focused tests, documentation updates, review of the diff,
and exactly one isolated commit. Existing unrelated worktree changes must not
be staged.

### Phase 0: Freeze the Localization Contract

**Commit:** `docs: define english-first localization contract`

Tasks:

- Add this plan and update `.opencode/commands/goal.md`.
- Confirm locale names, precedence and fallback behavior.
- Confirm stable JSON fields and additive localization fields.
- Record the no-persisted-preference decision.
- Inventory Portuguese identifiers, comments, literals, tests, scripts and
  documentation links.
- Reconcile stale documentation status before translating it.

Acceptance criteria:

- The plan is authoritative for localization work.
- No machine-facing enum, command, path or JSON key is scheduled for translation.
- The next incomplete phase is unambiguous.

### Phase 1: Add the Locale and Catalog Core

**Commit:** `feat: add english and brazilian portuguese catalogs`

Files and areas:

- `internal/i18n/`
- `go.mod` and `go.sum` if locale matching is added.
- `internal/cobra/config.go` or a new root-command constructor.

Tasks:

- Implement locale parsing, normalization and precedence.
- Add `en-US` and `pt-BR` catalogs with completeness checks.
- Add tests for explicit flags, environment variables, fallback and invalid
  locale values.
- Keep the existing application behavior unchanged until presentation layers
  start using the localizer.

Acceptance criteria:

- English is selected when no locale is provided.
- Both supported locale identifiers resolve correctly.
- Invalid explicit values fail without changing the process environment.
- Catalog tests prove that both locales contain every required message ID.

#### Phase 1 Result

- Locale resolution, POSIX normalization and explicit/environment precedence are
  implemented in `internal/i18n`.
- English and Brazilian Portuguese catalogs are embedded in the binary and
  validated for completeness at startup and in tests.
- The standard library is sufficient for the current two-locale scope, so no
  locale-matching dependency was added in this phase.

### Phase 2: Localize Cobra and JSON Presentation

**Commit:** `feat: localize cli presentation and json messages`

Files and areas:

- `internal/cobra/*.go`
- `internal/cobra/*_test.go`
- `internal/cobra/json.go`
- `internal/tui/commands.go`

Tasks:

- Replace command help, usage, examples, flag descriptions and validation text.
- Add the global `--locale` flag before help is rendered.
- Make root-command construction testable without shared mutable Cobra state.
- Add `locale`, `message_id` and `error_code` as additive JSON fields where
  needed.
- Keep stdout valid JSON in every locale and on every JSON failure path.
- Make the TUI forward `--locale` to child commands.

Acceptance criteria:

- `mobdesk --help --locale en-US` and `mobdesk --help --locale pt-BR` are
  localized while command names remain unchanged.
- Text output changes language without changing operation state.
- JSON keys, schema version and machine values remain stable.
- JSON stdout contains no human diagnostic outside the JSON document.

#### Phase 2 Result

- Cobra now builds independent command trees with explicit locale resolution,
  localized help, usage, examples, flags and validation messages.
- CLI text presentation uses the selected locale without changing command names,
  flags, JSON keys, schema versions or machine state values.
- Operation JSON responses keep the existing contract and add locale metadata,
  message IDs and error codes where presentation errors require them. JSON
  validation failures remain valid stdout documents.
- The TUI forwards its selected locale explicitly to child Mobdesk CLI commands.
- Focused tests cover bilingual help, text messages, JSON failures, invalid
  locales, independent root state and TUI locale propagation.

### Phase 3: Localize Services, Status and App Profiles

**Commit:** `feat: localize service messages and app profiles`

Files and areas:

- `internal/install/`
- `internal/status/`
- `internal/workstation/`
- `internal/update/`
- `internal/executil/`
- `internal/logs/`
- `internal/paths/` only if a presentation preference is later introduced.

Tasks:

- Replace service-layer prose with message IDs, error codes and structured data.
- Move app and configuration descriptions into the catalog.
- Translate setup, host-boundary, installation, configuration, uninstall,
  update, status and log presentation.
- Preserve old installation and configuration records.
- Add additive `last_error_code` or equivalent structured fields where needed;
  keep old `last_error` readable.
- Translate comments and identifiers to English.

Acceptance criteria:

- No service chooses a language by reading a global mutable variable.
- Existing records load successfully under either locale.
- Changing locale does not change state transitions or safety decisions.
- Detected apps, shared packages, conflicts and modified files keep their stable
  machine states.

#### Phase 3 Result

- Service errors and progress messages now carry typed message IDs and stable
  error codes, with locale selection passed explicitly through service options
  or presentation boundaries.
- Installation and configuration records preserve their existing fields and
  remain readable; additive `last_error_code` fields identify new failures
  without replacing historical `last_error` text.
- App and configuration profile descriptions come from the embedded English and
  Brazilian Portuguese catalogs.
- Status text, workstation warnings/errors, update failures and log failures
  render through the selected locale while JSON states, keys, paths, records
  and external command output remain unchanged.
- Focused tests cover English and Brazilian Portuguese service/status output and
  compatibility with legacy installation records.

### Phase 4: Localize the Bubble Tea Interface

**Commit:** `feat: localize bubble tea interface`

Files and areas:

- `internal/tui/model.go`
- `internal/tui/*.go`
- `internal/tui/*_test.go`
- mock backend and status fixtures.

Tasks:

- Store the localizer in the TUI model.
- Localize every screen, title, card, table, footer, help key, modal, popup,
  button, error, progress message and runtime restriction.
- Use localized labels from the catalog for mouse hit-testing.
- Preserve touch-first behavior and visible close/back actions.
- Keep status values stable internally and localized only when rendered.
- Test the app popup in both locales, including destructive confirmations.

Acceptance criteria:

- English is the default visual experience.
- Portuguese can be selected without rebuilding the binary.
- Enter, Escape, Tab, mouse and touch work in both locales.
- Popup hit-tests use the rendered locale and do not depend on fixed English or
  Portuguese text.
- Narrow terminals do not overflow in either locale.

#### Phase 4 Result

- The Bubble Tea model stores one selected `i18n.Localizer`, and real child
  commands, including the interactive shell, receive that locale explicitly.
- Every TUI screen, popup, modal, status label, operation state, progress
  message, help label, button and Termux/remote restriction renders through
  bilingual message IDs; machine states and action IDs remain stable.
- Mock backends use the selected locale without executing host commands.
- Mouse targets are derived from localized rendered labels and geometry,
  including app rows, popup actions, confirmations and header navigation.
- Focused tests cover both locales, keyboard and mouse confirmations, popup
  actions, remote restrictions and narrow terminals.

### Phase 5: Complete Code, Script and Test Migration

**Commit:** `test: cover bilingual cli and tui behavior`

Tasks:

- Rename remaining Portuguese Go identifiers, test names and helper names.
- Translate all comments to English.
- Translate Makefile help, shell-script comments and script diagnostics.
- Replace exact Portuguese assertions with message-ID or locale-specific tests.
- Add a static check for user-visible literals outside the catalogs where
  practical.
- Add tests for locale propagation across the TUI and SSH boundary.
- Run focused tests for both locales before the complete check.

Acceptance criteria:

- `go test ./...` covers both locales for CLI and TUI presentation.
- Tests assert semantics and stable codes instead of localized prose where
  possible.
- No Portuguese identifier or comment remains in production code.
- No direct user-visible string remains in presentation code outside approved
  catalogs and external command fixtures.

#### Phase 5 Result

- Remaining Go identifiers, helper names, test names and comments are English;
  technical values and persisted machine contracts remain unchanged.
- Makefile and shell-script help, comments and diagnostics are English.
- CLI and TUI presentation paths render through message IDs with English as the
  no-localizer fallback; tests verify stable states, IDs and both locale catalogs.
- `scripts/i18n-check.sh`, exposed as `make i18n-check`, checks source comments
  and literals, script diagnostics and catalog completeness without scanning
  documentation, external command fixtures or machine values.
- TUI child-command locale forwarding and the SSH/runtime restriction remain
  covered by focused bilingual tests.

### Phase 6: Migrate Documentation to English

**Commit:** `docs: migrate repository documentation to english`

Canonical technical documents become English-only. Public onboarding documents
have English and Brazilian Portuguese versions.

Tasks:

- Translate `AGENTS.md` and update stale command lists and validation rules.
- Translate mission, architecture, decisions, roadmap and implementation plans.
- Translate security, SSH, battery, audit and release documents.
- Translate active idea and research documents; label personal or historical
  notes explicitly if they are retained.
- Rename Portuguese document paths to English names and update every reference.
- Make `README.md` English and add `README.pt-BR.md`.
- Make contributor documentation English and add a `pt-BR` mirror.
- Add language links and verify relative links.
- Update `.opencode/commands/goal.md` to use the final English paths.

Recommended path changes:

- `docs/MISSAO.md` -> `docs/MISSION.md`
- `docs/DECISOES.md` -> `docs/DECISIONS.md`
- `docs/ARQUITETURA.md` -> `docs/ARCHITECTURE.md`
- `docs/PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md` ->
  `docs/APP-CONFIGURATION-REFACTOR-PLAN.md`
- `docs/PLANO-IMPLEMENTACAO-APPS-E-CONFIGURACOES.md` ->
  `docs/APP-CONFIGURATION-IMPLEMENTATION-PLAN.md`
- `docs/ideias/` -> `docs/ideas/`
- `teste-vscode.md` -> `vscode-test.md`

Acceptance criteria:

- All maintained technical documentation is English.
- README and contributor onboarding are available in English and pt-BR.
- No internal link points to a removed Portuguese path.
- Product decisions and out-of-scope boundaries are unchanged by translation.

#### Phase 6 Result

- `AGENTS.md`, maintained product plans, security/operations documents and
  active research are English canonical documents; retained personal and
  historical notes are explicitly marked non-authoritative.
- Portuguese document paths named by this phase were renamed with `git mv`:
  mission, architecture, decisions, app plans, `docs/ideas/` research and
  `vscode-test.md`.
- `README.md`, `.github/README.md`, contributor onboarding, support and security
  are English, with public `README.pt-BR.md`, `.github/README.pt-BR.md`,
  `.github/CONTRIBUTING.pt-BR.md`, `.github/SUPPORT.pt-BR.md` and
  `SECURITY.pt-BR.md` mirrors.
- Relative links, README/contributor path inconsistencies and the OpenCode goal
  command now use the final canonical names. `make i18n-check` also validates
  local Markdown links.
- Product decisions, scope boundaries, historical statuses, JSON contracts and
  persisted runtime paths were preserved. TODO files were intentionally not
  edited.

### Phase 7: Release Validation

**Commit:** `test: validate english first localization rollout`

Tasks:

- Add `make i18n-check` or an equivalent validation target.
- Validate catalog completeness and documentation links.
- Run `make check`.
- Run focused CLI and TUI tests in both locales.
- Run `make catalog-test` only if installation, catalog or related
  runtime code changed during the migration.
- Validate the TUI visually with the mock backend in both locales.
- Validate the already tested real Termux flow and expand coverage to additional
  Android devices when available.

Acceptance criteria:

- English is the default in a clean environment.
- `MOBDESK_LOCALE=pt-BR` changes CLI and TUI presentation.
- The TUI and CLI use the same selected locale.
- JSON remains schema-compatible and machine-stable.
- `make check` passes.
- Remaining real-device checks are documented explicitly.

#### Phase 7 Result

- `make i18n-check` passed, including catalog completeness, presentation checks
  and maintained Markdown link validation.
- Focused `internal/i18n`, `internal/cobra` and `internal/tui` tests passed for
  both supported locales.
- `mobdesk --locale en-US --help` and `mobdesk --locale pt-BR --help` produced
  localized help while preserving command names and flags.
- `make check` passed with formatting, i18n validation, `go vet`, all tests and
  the Termux fixture build.
- `make catalog-test` was intentionally not repeated because the migration did
  not change the installation catalog, installation strategies or Termux runtime.
- Visual and operational validation has been completed on a real
  Android/Termux device; validation across a broader device matrix remains
  ongoing before claiming broad device-level release readiness.

## 7. Documentation Inventory

### Canonical project rules and product documents

- `AGENTS.md`
- `docs/MISSION.md`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `docs/ROADMAP.md`
- `docs/APP-CONFIGURATION-REFACTOR-PLAN.md`
- `docs/APP-CONFIGURATION-IMPLEMENTATION-PLAN.md`

### Public documentation

- `README.md`
- `README.pt-BR.md`
- `.github/README.md`
- `.github/README.pt-BR.md`
- `.github/CONTRIBUTING.md`
- `.github/CONTRIBUTING.pt-BR.md`
- `.github/SUPPORT.md`
- `.github/SUPPORT.pt-BR.md`
- `SECURITY.md`
- `SECURITY.pt-BR.md`

### Operational and research documentation

- `docs/SECURITY-AUDIT-2026-07-25.md` — English historical audit with
  follow-up statuses.
- `docs/PRE-RELEASE-FIX-GUIDE.md` — English active remediation guidance.
- `docs/SSH-AUTHENTICATION.md` and `docs/BATTERY-DIAGNOSTICS.md` — English future
  security and operational notes.
- `docs/DOCTOR-PLAN.md` and `docs/PRIORITY-REFACTORING-PLAN.md` — English
  future/historical implementation plans.
- `docs/ideas/TOOL-CATALOG.md`, `REMOTE-BROWSER-RESEARCH.md` and
  `GLOBAL-TUI-REFRESH-PLAN.md` — English active research and ideas.
- `docs/ideas/AROZOS-PROPOSAL.md`, `docs/ideas/CODE-SERVER.md` and
  `vscode-test.md` — English personal or historical notes explicitly marked
  non-authoritative.
- `docs/ideas/TODO.md` and `docs/SECURITY-AUDIT-TODO.md` — English translated
  planning notes retained as non-authoritative documents.

## 8. Validation Matrix

| Area | `en-US` | `pt-BR` | Required result |
|---|---:|---:|---|
| Root help | yes | yes | Same commands and flags, translated prose |
| Text command success | yes | yes | Same operation state |
| Text command failure | yes | yes | Same error code, translated context |
| JSON success | yes | yes | Valid schema 1 JSON |
| JSON failure | yes | yes | Valid schema 1 JSON with stable codes |
| TUI screens | yes | yes | No missing strings or overflow |
| Popup actions | yes | yes | Correct keyboard and mouse hit-tests |
| Persisted records | yes | yes | Old and new records load |
| TUI child commands | yes | yes | Locale is forwarded explicitly |
| Local Termux/SSH session | yes | yes | Same safety restrictions |
| Documentation links | yes | yes | No broken references |

## 9. Research Basis

- [`golang.org/x/text/language`](https://pkg.go.dev/golang.org/x/text/language)
  provides BCP 47 parsing and ordered language matching with a deterministic
  fallback.
- [`go-i18n`](https://github.com/nicksnyder/go-i18n) supports message IDs,
  named variables, plural rules, embedded catalogs and translator workflows. It
  is a future option if the project grows beyond two locales.
- [`Cobra`](https://pkg.go.dev/github.com/spf13/cobra) provides configurable
  help, usage and template functions but has no built-in localization layer.
- [`Bubble Tea`](https://github.com/charmbracelet/bubbletea) has no built-in
  i18n. Its model-based `Init`, `Update` and `View` lifecycle is suitable for
  carrying one immutable localizer per session.

## 10. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Localized prose leaks into automation | Add stable IDs and error codes |
| Existing records contain Portuguese errors | Preserve old fields and add structured codes |
| TUI hit-tests break after translation | Derive hit regions from rendered localized labels |
| Cobra help renders before locale selection | Resolve global locale before help execution |
| SSH session loses Termux environment | Forward locale explicitly to child commands |
| Documentation links break after renames | Rename with `git mv`, then run link validation |
| Translations drift | Require catalog completeness and paired public docs |
| Scope expands into a translation platform | Keep two locales and embedded catalogs for this release |
