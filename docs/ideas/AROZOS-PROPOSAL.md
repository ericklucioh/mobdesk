# ArozOS Proposal

**Status:** personal historical research; non-authoritative and not part of the
Mobdesk MVP or roadmap.

ArozOS does not add impossible Linux powers. Its value is a browser-based layer
for tasks where the client browser provides graphics, drag-and-drop, media,
sharing links, multiple users and access without an SSH client.

## Potential advantages over a pure CLI/TUI

- visual file transfer between laptop and phone;
- full-size image galleries and thumbnails;
- video transcoding and browser playback;
- browser music playback;
- shareable file URLs with permissions;
- multiple users and storage isolation;
- access from a browser without SSH installation;
- several WebApps in windows, such as code-server, Grafana, terminal and
  Mobdesk panels;
- Markdown editing with rendered preview;
- visual static-site editing and publishing;
- graphical storage and disk-operation views.

These workflows turn commands such as `scp`, `rsync`, `ffmpeg`, `sftp`, `ls`,
`du`, `chmod` and HTTP servers into browser windows, buttons, links and forms.

## What remains better in CLI/TUI

Programming in Go, Git, compilation, scripting, advanced process management,
automation, Neovim, Lazygit, `htop`, environment configuration and keyboard-
first work remain better suited to the Mobdesk TUI. ArozOS does not
automatically open arbitrary Linux GUI programs, Firefox, VS Code desktop or
Android applications; each needs a WebApp, URL, adaptation or VNC/xpra.

## Non-authoritative composition idea

Keep Termux, Git, Go, Node, Neovim, Zellij and projects as the
development core. A separate future browser layer could provide files, media,
sharing, code-server, Zellij Web and HTTP panels. It would complement rather
than replace the TUI, and would require explicit scope, authentication,
multi-user and Android/Termux validation.

Reference: [ArozOS](https://os.aroz.org/),
[ArozOS releases](https://github.com/tobychui/arozos/releases).
