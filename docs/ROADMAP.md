# Mobdesk Roadmap

Mobdesk evolves through four stages. Termux remains the control host and
Ubuntu through PRoot remains the primary development environment.

## Overview

| Stage | Category | Name | Result |
|---|---|---|---|
| 1 | MVP | Ubuntu bootstrap | Install and access persistent Ubuntu by shell and SSH |
| 2 | MVP | TUI workstation | Work with organized terminal tools |
| 3 | MVP | Persistent environment | Recover sessions, services and remote access |
| 4 | Application | Mobdesk Manager | Manage projects, sessions and services |

## Stage 1 - Ubuntu bootstrap

**Status:** initial implementation complete and tested on a real Android/Termux
device; validation across a broader device matrix is still ongoing.

The goal is to take a Termux installation with Go to persistent Ubuntu
accessible on the phone and through SSH. Scope includes PRoot-Distro, OpenSSH,
diagnostic tools, Ubuntu ARM64, workspace and local state, dedicated SSH,
safe start/stop, `mobdesk shell`, `mobdesk status`, initial tool profiles,
local addresses, wake-lock when available, and resumable setup.

Out of scope: TUI, projects, services, persistent sessions, `doctor`, Tailscale
and port forwarding.

## Stage 2 - TUI workstation

**Status:** initial implementation in progress and tested on a real
Android/Termux device; validation across a broader device matrix is still
ongoing.

The goal is an organized text interface for working in Ubuntu on the phone or
through SSH. Scope includes status, setup, start, stop and update screens,
initial tool profiles, keyboard/mouse/mobile terminal support, remote-session
identification and host-action blocking, optional app configuration profiles
starting with Neovim/LazyVim, versioned JSON CLI operations, separate app and
configuration states, and touch-first details/actions/confirmation popups.

The criterion is that users can study and develop without a long sequence of
internal commands, and that interactive package installation can be completed
from the TUI without losing access to terminal prompts.

## Stage 3 - Persistent and remote environment

The goal is to reconnect and continue after a network change, disconnect or
screen-off event. Planned scope includes tmux sessions, automatic recovery,
complete `status` and `doctor`, logs and health checks, optional Tailscale, port
forwarding, backups, and battery/background guidance for HyperOS.

The criterion is that a project can be started, stopped, reconnected and
continued without rebuilding its environment.

## Stage 4 - Mobdesk Manager

The goal is a local management center for Termux, PRoot and Ubuntu. Planned
scope includes projects, environments, installed tools, sessions, services,
ports, tunnels, logs, diagnostics, recovery, backups, controlled updates and
observable persisted configuration.

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
