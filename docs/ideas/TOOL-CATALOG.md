# Mobdesk Tool Catalogue

**Status:** active editorial research; this catalogue is not an installation
contract. Only profiles present in the application catalogue are supported by
Mobdesk installation commands.

Each entry follows `name - type - description` and is grouped by likely value
inside Ubuntu through PRoot.

## Essential tools

### Version control

- Git - git - version control for projects.

### Terminal UI and multiplexing

- Zellij - multiplexer - tabs, panes and layouts.
- tmux - multiplexer - persistent sessions and panes.
- VTM - multiplexer - experimental textual desktop with windows and panes.
- TUIOS - multiplexer - candidate for a simpler terminal desktop.

The choice between tmux, Zellij, VTM and TUIOS remains an open UX question:
tmux is mature but difficult for beginners; Zellij is mature and friendlier;
VTM and TUIOS may be more approachable but require validation.

### Development, Git and files

- micro - editor - simple terminal text editor.
- TTT - editor - minimal terminal text editor.
- GitHub CLI (`gh`) - git - official GitHub command-line client.
- gh-dash - git - GitHub CLI dashboard.
- lazygit - git - visual Git operations.
- tig - git - terminal history viewer.
- git-scope - git - multi-repository panel.
- gitv - git - terminal GitHub issue client.
- tree - files - directory tree display.
- fsel - files - interactive file selector; environment validation pending.
- Yazi - files - file manager with previews.
- TUIFI Manager - files - visual terminal file explorer.
- ripgrep - search - fast text search.
- fd - search - fast file and directory search.
- fzf - search - fuzzy selection for files and commands.
- television - search - configurable terminal fuzzy finder.
- Zsh - shell - interactive shell with configurable completions.
- fzf-tab - shell - fuzzy Zsh completion selector.
- Atuin - shell - interactive command history.
- Carapace - shell - completions for shells and commands.

### Data, APIs and databases

`jq`, `fx`, `csvlens`, `Tabiew`, `OTree`, `jqp`, `Twig`, `ATAC`, `resterm`
and `Posting` are candidates for JSON, CSV, structured-data and HTTP work.
`Harlequin`, `Rainfrog`, `Lazysql`, `pgcli`, `mycli`, `litecli`, `sqlite3`,
`IRedis` and `termdbms` are terminal database candidates.

### Operations and infrastructure

`btop`, `ncdu`, `gdu`, `bottom`, `glances`, `diskonaut` and `inxi` cover
resource and storage inspection. `lnav`, `logradar` and `lazyjournal` cover
logs. `bandwhich`, `termshark`, `trippy`, `mitmproxy`, `speedtest-cli`,
`curl`, `wget`, `ftp`, `ping`, `ip` and `netstat` cover network work.
`K9s`, `ktop` and `termscp` are infrastructure candidates.

### Assistance and documentation

- OpenCode, clin and glint - programming assistance for the terminal.
- leetgo - programming exercise client.
- dialog - terminal dialogs and menus.
- glow - terminal Markdown reader.
- less - text pager.
- mermaid-ascii - ASCII diagram renderer.

The Mermaid reference is <https://github.com/AlexanderGrooff/mermaid-ascii>.

## Non-essential or host-dependent tools

Email and communication candidates include aerc, NeoMutt, Mutt, Alpine, meli,
Discord clients, nchat, gomuks, Irssi, Profanity, gurk-rs, tuir and Devzat.
Productivity candidates include calcurse, taskwarrior-tui, hledger-ui, taskline,
Cal and keydex.

Media and web candidates include cmus, GopherTube, invidtui, Streamlink,
twitch-tui, ncspot, spotify-player, spotatui, ncmpcpp, kew, termusic, rmpc,
pyradio, soundcloud2000, newsboat, Jellyfin TUI, managarr, epy, tdf,
manga-tui, Elinks, FFmpeg, VLC and related clients.

Tools that normally require host capabilities or are poorly suited to PRoot
include WibWob-DOS, Wi-Fi/Bluetooth managers, NetworkManager tools, partition
editors, Distrobox, container monitors and image writers. They remain catalogue
ideas, not Mobdesk promises.
