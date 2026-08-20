# Code-Server on Android PRoot

**Status:** personal historical setup note; non-authoritative. It predates the
Mobdesk MVP and must not be treated as an application profile or supported
runtime path. The PRoot steps below are archived and must not be used for
Mobdesk.

## Intended topology

```text
POCO F6
└── Termux
    ├── Mobdesk SSH :8022
    ├── tmux
    └── Ubuntu PRoot
        ├── code-server :1212
        └── /root/workspace
```

The phone performs the work; a computer browser is only the client. PRoot does
not provide Android root. Mobdesk's supported control surface remains its own
CLI/TUI and SSH boundary.

## Manual experiment

The original experiment installed `proot-distro`, OpenSSH and tmux in Termux,
entered Ubuntu, installed curl, certificates, Git, OpenSSL and nano, and used
the official code-server installer after inspecting its dry run. It verified
`uname -m`, `code-server --version`, and a workspace at `/root/workspace`.

The first code-server launch creates `/root/.config/code-server/config.yaml`.
For a local-only setup the note used:

```yaml
bind-addr: 127.0.0.1:1212
cert: false
```

The password file must be private. Binding `0.0.0.0` exposes the service to the
LAN and is not recommended; an SSH tunnel is preferred.

## SSH tunnel

Start the SSH server in Termux, not inside Ubuntu. A client can forward the
local browser port with:

```bash
ssh -p 8022 -N \
  -o ServerAliveInterval=15 \
  -o ServerAliveCountMax=3 \
  -o ExitOnForwardFailure=yes \
  -L 1212:127.0.0.1:1212 \
  USER@PHONE_IP
```

The browser then opens `http://127.0.0.1:1212` on the computer. Additional
application ports can be forwarded with more `-L` options, but local ports
must not collide. SSH config aliases may store the host, port, keepalive and
forwarding values. Prefer public-key authentication.

## Persistence and extensions

PRoot does not normally provide systemd. The experiment used tmux to keep
code-server running and `termux-wake-lock` to reduce Android suspension risk.
Neither prevents Android or HyperOS from terminating processes. Code-server
extensions and user data are stored in the Ubuntu user's directories, for
example `/root/.local/share/code-server/extensions`; VS Code desktop and
code-server are separate installations and use different marketplaces.

Exporting desktop extensions and copying settings is manual and may fail when
an extension is unavailable on Open VSX or lacks an ARM64/browser build.
Backups should use PRoot-Distro backup and must be made before destructive
reset; never assume the computer has a local copy of `/root/workspace`.

## Diagnostics

```bash
curl http://127.0.0.1:1212/healthz
code-server --version
uname -m
ldd --version
```

If authentication fails, inspect certificates, DNS, date and proxy variables
inside Ubuntu. If the phone works but the computer cannot connect, inspect the
SSH server, tunnel, network and Android background restrictions. `systemctl`
failure is expected in PRoot; use a deliberately managed foreground/tmux
process instead.

This note is retained for history only. It does not change Mobdesk persisted
paths, application behavior, or the MVP scope.
