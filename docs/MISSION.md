# Mission

This document describes the product problem, user value and intended audience
before choosing tools or implementation details.

## Objective

Mobdesk should let a computer-science student or developer go to class, travel
or work elsewhere with only a phone and still have a personal development
environment. The user should not need to sign personal GitHub, messaging or
email accounts into shared computers, and should retain control of projects,
tools and data.

The environment is intended for study and small or medium development work:
C, JavaScript, HTML, React, Java, Go and Python projects; compiling programs;
running a local server such as `npm run dev`; and viewing that server from a
browser on the same network. It is not intended for production-scale load or
performance testing.

## Audience

### Students

The primary experience should eventually be a simple flow inside a TUI:
configure, start, open a tool or editor, work, and stop. Data remains under the
user's control so changing classrooms or returning home does not require
rebuilding accounts and configuration.

### Professionals

The simple flow must not remove access to the underlying Termux environment.
Advanced users should still be able to use the workstation through the phone and
SSH.

## Main challenge

The largest challenge is visual interaction without requiring a conventional
desktop. A full graphical desktop would be too heavy and unreliable on the
target hardware. Browser-based applications may eventually expose individual
tools, but that is a later direction rather than an MVP requirement.

## MVP-1 boundary

The first MVP is intentionally a 100% TUI experience exposed through one SSH
entry point:

```text
Termux -> Mobdesk -> local shell or SSH
```

Termux is the sole workstation and development environment. PRoot-Distro and
Ubuntu are removed from the active product scope. Application configuration,
including LazyVim, is deferred. It does not add a code-server port or a complete
graphical desktop. User applications may expose their own development ports, but
the Mobdesk control surface remains the dedicated SSH flow.

Existing PRoot-based installations cannot be migrated. They require a full
Termux reset followed by a fresh Termux-only installation.
