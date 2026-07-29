---
description: Execute the Mobdesk apps and configurations implementation plan
agent: build
---

Execute the implementation plan for Mobdesk end to end.

Read `AGENTS.md`, `docs/MISSAO.md`,
`docs/PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md`, and
`docs/PLANO-IMPLEMENTACAO-APPS-E-CONFIGURACOES.md` before changing code.
Treat the decision document and implementation plan as authoritative. Do not
reopen decisions that those documents mark as closed.

If arguments were provided, interpret them as the requested scope, such as a
phase number or a focused task. Without arguments, execute phases 0 through 11
in dependency order. Work in small, coherent increments and preserve any
unrelated pre-existing worktree changes.

For every phase:

1. Inspect the current implementation and the phase acceptance criteria before
   editing.
2. Implement the smallest correct change that satisfies the phase.
3. Add or update focused tests, including regression coverage for safety and
   Termux/Ubuntu boundary behavior where relevant.
4. Update the project documentation required by the phase.
5. Run the phase validation and fix failures before moving on.
6. Review the diff and report the phase result explicitly.

Respect the project boundaries: host actions run only in Termux, Ubuntu runs
through PRoot, the TUI uses the CLI JSON contract, user input must not become
shell syntax, destructive actions require confirmation, existing user files
must not be overwritten or removed silently, and long operations must support
context cancellation.

Do not commit, amend, push, reset, checkout, or remove unrelated changes unless
the user explicitly requests it. Do not claim Termux or PRoot validation was
performed unless it actually ran in that environment. Stop only for a genuine
technical blocker or an unsafe ambiguity; otherwise choose the minimal option
consistent with the locked decisions.

Before finishing, run `make check`, summarize changed files and validation
results, and list any checks that still require a real Termux/POCO F6 run.
