---
description: Execute the Mobdesk English-first localization plan
agent: build
---

Execute the Mobdesk English-first localization and documentation migration plan
end to end.

Read `AGENTS.md` and `docs/PLAN-I18N-EN-PTBR.md` before changing code. Until the
documentation migration phase renames the existing Portuguese paths, also read
the current mission, architecture, decisions and app implementation documents:
`docs/MISSAO.md`, `docs/ARQUITETURA.md`, `docs/DECISOES.md`,
`docs/PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md`, and
`docs/PLANO-IMPLEMENTACAO-APPS-E-CONFIGURACOES.md`.

Treat the localization plan, existing locked product decisions and JSON
contract rules as authoritative. Do not reopen decisions already marked as
closed. English is the primary locale; `pt-BR` is the supported second locale.

If arguments were provided, interpret them as the requested scope, such as a
phase number or a focused task. Without arguments, inspect the localization plan
status, phase results, current tests, implementation, documentation and recent
commits; resume at the first incomplete localization phase in dependency order.
Do not reimplement or recreate completed app-management phases. If the
documentation is stale, reconcile it with the code, tests and commits before
choosing the next phase.
Work in small, coherent increments and preserve any unrelated pre-existing
worktree changes.

For every phase:

1. Inspect the current implementation and the phase acceptance criteria before
   editing.
2. Implement the smallest correct change that satisfies the phase. Keep
   user-visible text in message catalogs and use message IDs in code.
3. Add or update focused tests, including regression coverage for safety and
   Termux/Ubuntu boundary behavior where relevant. Test both `en-US` and
   `pt-BR` when presentation is affected.
4. Update the project documentation required by the phase.
5. Run the phase validation and fix failures before moving on.
6. Review the diff and report the phase result explicitly.
7. Create exactly one isolated commit for the completed phase before moving on.

Respect the project boundaries: host actions run only in Termux, Ubuntu runs
through PRoot, the TUI uses the CLI JSON contract, user input must not become
shell syntax, destructive actions require confirmation, existing user files
must not be overwritten or removed silently, long operations must support
context cancellation, and locale selection must not change machine-facing
states, paths or safety decisions.

Localization rules:

- Default to `en-US`; support `pt-BR` through `--locale` and
  `MOBDESK_LOCALE`.
- Preserve command names, flags, JSON keys, schema versions, state values,
  source values, error codes, paths and profile IDs in English.
- Do not place user-visible prose directly in Go presentation code.
- Do not translate arbitrary output from external commands.
- Keep the localizer immutable per CLI/TUI session; do not use mutable package
  global locale state.
- Localize TUI hit-test labels and test both locales at narrow widths.
- Preserve old persisted records and add structured compatibility fields instead
  of changing existing machine values.
- Make English canonical for technical documentation. Maintain pt-BR mirrors
  for public onboarding documents required by the localization plan.

After validation, inspect `git status`, `git diff`, and recent commits, stage
only files belonging to the current phase, and commit using the phase message
defined in the implementation plan. Never include unrelated pre-existing
changes. Do not amend, push, reset, checkout, or remove unrelated changes. If
the phase is not valid, do not commit it. Do not claim Termux or PRoot
validation was performed unless it actually ran in that environment. Stop only
for a genuine technical blocker or an unsafe ambiguity; otherwise choose the
minimal option consistent with the locked decisions.

Before finishing, run `make check`, summarize changed files and validation
results, and list any checks that still require a real Termux/POCO F6 run. Run
the expensive `make catalog-test` only when the phase changes the catalog,
installation strategies, PRoot boundary, catalog smoke script, or related
runtime behavior; otherwise use `make check` and focused tests. For localization
phases, also run the catalog, locale, JSON and bilingual TUI checks defined in
`docs/PLAN-I18N-EN-PTBR.md`.
