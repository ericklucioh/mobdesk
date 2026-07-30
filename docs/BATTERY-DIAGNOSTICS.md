# Battery Consumption Diagnostics

**Status:** operational investigation; no cleanup behavior is changed by this
document.

Ubuntu through PRoot is a persistent userland, not an independently running
service. A `proot-distro login ubuntu ...` process exists while its command,
shell or session remains active. Mobdesk starts Termux SSH and may enable
`termux-wake-lock`; SSH/PRoot sessions can create child processes.

## Current behavior

- `mobdesk start` checks Ubuntu, configures SSH, enables wake-lock and starts
  `sshd` on port `8022`;
- `mobdesk stop` signals the main `sshd` PID and releases wake-lock;
- `mobdesk shell` starts an interactive PRoot session;
- the TUI cancels commands it starts through `context.Context`;
- an installed Ubuntu filesystem consumes no CPU or battery without processes.

## Risk

The stop path directly controls the main SSH PID but has no complete inventory
of children. SSH/PRoot shells, tmux/Zellij sessions, `nohup` processes,
development servers and long builds may survive the TUI or SSH session and
consume CPU, memory, network and battery. A global `pkill proot` is unsafe
because unrelated processes may be affected.

Inspect processes and status in Termux:

```sh
ps -ef | grep -E 'sshd|proot|ubuntu|tmux|zellij|go build|node|python' | grep -v grep
mobdesk status --json
cat "$HOME/.local/share/mobdesk/ssh/sshd.pid" 2>/dev/null
mobdesk stop
```

Close SSH, tmux and Zellij sessions explicitly.

## Proposed safe correction

Future `stop` should prove that the registered PID is Mobdesk `sshd`, identify
only descendants and their sessions, signal them in order, wait for PRoot and
the port to close, release wake-lock on every error path, and report processes
that could not be stopped. It must continue protecting external processes.

The current conclusion is a lifecycle gap, not permission to kill all PRoot
processes. Confirm the device process list and wake-lock state before changing
cleanup behavior.
