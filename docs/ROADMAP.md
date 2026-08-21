# Mobdesk Roadmap

Mobdesk begins with a Termux-only first sprint. Termux is the sole workstation;
PRoot-Distro and Ubuntu are removed. Existing PRoot-based installations require
a full Termux reset and fresh installation because no migration is provided.

## Overview

| Stage | Category | Name | Result |
|---|---|---|---|
| 1 | MVP | Termux-only foundation | Establish the sole workstation and reset boundary |
| 2 | MVP | TUI workstation | Work with organized Termux tools |
| 3 | MVP | Persistent environment | Recover sessions, services and remote access |
| 4 | Application | Mobdesk Manager | Manage projects, sessions and services |

Application configuration and LazyVim are deferred. Native Java 21, Maven,
Kotlin, Gradle and curated user CLI profiles are delivered separately from
project wrappers and configuration profiles; all remain subject to device
acceptance.

The curated user CLI catalog currently includes TUIFI, Bitwarden CLI and
Resterm. Their private installation paths preserve user-managed npm, pip and Go
locations while real-device acceptance remains pending.

## Stage 1 - Termux-only foundation

**Status:** active.

The goal is a single Termux development workstation accessible on the phone and
through SSH. Scope includes Termux packages, OpenSSH, workspace and local state,
safe start/stop, `mobdesk shell`, `mobdesk status`, initial tool profiles, local
addresses, wake-lock when available and repeatable setup.

The current validation sprint executes small native Git, Go, Java, Python,
Node/npm, C, C++, Lua, Neovim and tmux workflows in `~/workspace` before the required
real-device acceptance. It remains part of Stage 1; it does not start the
persistent-environment Stage 3.

Out of scope: PRoot-Distro, Ubuntu, application configuration, LazyVim, projects,
services, persistent sessions, `doctor`, Tailscale and port forwarding.

## Stage 2 - TUI workstation

**Status:** planned after Stage 1 validation.

The goal is an organized text interface for working in Termux on the phone or
through SSH. Scope includes status, setup, start, stop and update screens,
initial tool profiles, keyboard/mouse/mobile terminal support, versioned JSON
CLI operations, standardized app metadata, and touch-first
details/actions/confirmation popups. Application configuration, including
Neovim/LazyVim, remains deferred.

The criterion is that users can study and develop without a long sequence of
internal commands, that every catalog app presents the same useful metadata,
and that interactive package installation can be completed from the TUI without
losing access to terminal prompts.

## Stage 3 - Persistent and remote environment

The goal is to reconnect and continue after a network change, disconnect or
screen-off event. Planned scope includes tmux sessions, automatic recovery,
complete `status` and `doctor`, logs and health checks, optional Tailscale, port
forwarding, backups, and battery/background guidance for HyperOS.

The criterion is that a project can be started, stopped, reconnected and
continued without rebuilding its environment.

## Stage 4 - Mobdesk Manager

The goal is a local management center for Termux. Planned scope includes
projects, installed tools, sessions, services, ports, tunnels, logs,
diagnostics, recovery, backups, controlled updates and observable persisted
state.

## Evolution principles

1. Each stage preserves the previous stage.
2. CLI and TUI use the same internal services.
3. Plugins, Nix and multiple users are not anticipated without validated need.
4. The next stage starts only after the current flow is validated on real
   Termux.

## Outside the core

- real Docker;
- a Linux VM;
- a complete graphical desktop;
- Nix as a requirement;
- Neko as a requirement;
- multiple users;
- heavy production workloads.

## Superseded roadmap material

The previous Ubuntu bootstrap and PRoot-based workstation stages are superseded
by the Termux-only first sprint. They are retained in repository history but do
not define active work.
