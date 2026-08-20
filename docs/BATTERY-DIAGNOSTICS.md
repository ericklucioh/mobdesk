# Battery Consumption Diagnostics

**Status:** operational investigation; no cleanup behavior is changed by this
document. Mobdesk runs only in native Termux; PRoot and Ubuntu are not part of
the process model.

Mobdesk starts Termux SSH and may enable `termux-wake-lock`. SSH sessions,
interactive shells and development tools can create child processes that remain
after the initiating TUI or terminal closes.

## Current behavior

- `mobdesk start` validates native setup, configures SSH, enables wake-lock and
  starts `sshd` on port `8022`;
- `mobdesk stop` signals the main `sshd` PID and releases wake-lock;
- `mobdesk shell` starts an interactive Termux shell in the workspace;
- the TUI cancels commands it starts through `context.Context`;
- installed packages consume no CPU or battery without processes.

## Risk

The stop path directly controls the main SSH PID but has no complete inventory
of children. Shells, tmux/Zellij sessions, `nohup` processes, development
servers and long builds may survive the TUI or SSH session and consume CPU,
memory, network and battery. A global process kill is unsafe because unrelated
processes may be affected.

Inspect processes and status in Termux:

```sh
ps -ef | grep -E 'sshd|tmux|zellij|go build|node|python' | grep -v grep
mobdesk status --json
cat "$HOME/.local/share/mobdesk/ssh/sshd.pid" 2>/dev/null
mobdesk stop
```

Close SSH, tmux and Zellij sessions explicitly.

## Proposed safe correction

Future `stop` should prove that the registered PID is Mobdesk `sshd`, identify
only descendants and their sessions, signal them in order, wait for the port to
close, release wake-lock on every error path, and report processes
that could not be stopped. It must continue protecting external processes.

The current conclusion is a lifecycle gap, not permission to kill all user
processes. Confirm the device process list and wake-lock state before changing
cleanup behavior.
