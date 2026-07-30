# Contributing to Mobdesk

Thank you for considering a contribution. Mobdesk is building a small,
verifiable Ubuntu workstation for Android.

## Before you start

Read:

- [README](../README.md) for installation and usage;
- [Mission](../docs/MISSION.md) for the product problem and value;
- [Architecture](../docs/ARCHITECTURE.md) for the Termux/Ubuntu boundary;
- [Roadmap](../docs/ROADMAP.md) for future scope;
- [Decisions](../docs/DECISIONS.md) for current project choices.

## Development environment

Recommended requirements:

- Go `1.26.5`;
- Docker with Docker Compose;
- Git;
- a terminal with TTY support;
- Android/Termux for final integration validation.

Prepare and run the project with:

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
make build-image
make dev
```

Use `make shell` for a separate shell in the environment.

## Required checks

Before submitting a change:

```bash
make check
```

For Docker changes, also run `docker compose config` and `make build-image`.
For Termux, SSH or PRoot changes, validate on real Termux as well. Docker does
not reproduce Android permissions, networking, battery behavior or kernel
restrictions. `make integration-test` validates the disposable Docker flow but
does not replace a device test.

## Code organization and rules

- `cmd/mobdesk/`: executable entry point;
- `internal/cobra/`: CLI commands and routing;
- `internal/tui/`: Bubble Tea screens and components;
- `internal/status/`: environment state collection;
- `internal/install/`: idempotent tool installation;
- `internal/update/`: update checks and application;
- `docs/`: mission, architecture, decisions, roadmap and technical plans.

Keep operations idempotent, preserve user data, separate Termux commands from
Ubuntu commands, use context cancellation for long operations, validate input
before forming commands, and never write passwords, tokens or keys to code or
logs. Keep user-facing prose in the i18n catalogs. Update documentation when
scope or architecture changes.

## Commits and pull requests

Use short descriptive commits, preferably `type: short description`. A pull
request should explain the problem, behavior change, tests, affected Termux,
Docker or Ubuntu environments, and remaining limitations. Do not combine
unrelated refactors, architecture changes and fixes.

## Current scope

The MVP-1 flow is:

```text
Termux -> Mobdesk -> SSH -> Ubuntu via PRoot
```

Projects, services, persistent sessions, Tailscale workflows, web interfaces
and other expansion areas remain future scope. Contributions in those areas
must follow the roadmap or explicitly update the scope decision.

## Reporting issues

Include the phone model and Android version, Termux source and version,
Mobdesk version, command executed, complete error output, and whether the
problem occurred in Termux, Ubuntu, SSH or Docker. Never publish passwords,
private keys, tokens or personal data in logs.

See [the Brazilian Portuguese contributor guide](CONTRIBUTING.pt-BR.md).
