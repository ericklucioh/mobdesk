# Remote Browser Research

**Research date:** 2026-07-17

**Status:** historical research; no alternative is approved for the MVP. Any
future reassessment must use a native Termux strategy, not PRoot or Ubuntu.

## Problem

Future Mobdesk work may run a graphical Firefox in Termux and stream
its desktop to another browser. The target is the complete browser surface,
including tabs, address bar, menus, keyboard, mouse and possibly audio, not the
HTML contents alone.

```text
HTTP       loads the web UI
WebSocket  or WebRTC carries screen, audio and events
X server   provides Firefox's virtual display
Firefox    runs in Termux
```

An apparent single HTTP port may still require WebSocket or WebRTC transport.

## Alternatives

| Alternative | Role | PRoot fit | Initial assessment |
|---|---|---|---|
| [Selkies](https://github.com/selkies-project/selkies) | HTML5 desktop streaming | promising | first standalone spike candidate |
| [Neko](https://github.com/m1k1o/neko) | virtual browser with WebRTC | experimental | best browser UX, Docker-first |
| [jlesage/docker-firefox](https://github.com/jlesage/docker-firefox) | Firefox/VNC web container | low/medium | simple reference, not standalone |
| [noVNC](https://github.com/novnc/noVNC) | HTML5 VNC client | high as a component | simple baseline with Xvfb/VNC |
| [Apache Guacamole](https://guacamole.apache.org/) | HTML5 VNC/RDP/SSH gateway | medium | useful future gateway, not video-first |
| [Kasm Workspaces](https://www.kasmweb.com/) | Docker-oriented VDI | low | too large for the phone and PRoot |

Selkies documents standalone packaging and ARM64 and can use one WebSocket
port. It still requires Xvfb/Xorg, Firefox, GStreamer and a compatible capture
pipeline, and its CPU cost and HyperOS/PRoot compatibility are unproven.

Neko offers packaged Firefox, audio/video, authentication and reconnection but
officially prioritizes Docker, X, PulseAudio, GStreamer and WebRTC. It should
return to consideration only with a real container runtime or validated
non-Docker execution.

The Docker Firefox image is a useful simplicity reference. It exposes a web UI
such as port `5800`, but its container assumptions do not transfer to PRoot.
noVNC is only a browser-side VNC client and must be paired with Xvfb, Firefox,
VNC and websockify.

## Current research decision

1. Baseline: Firefox + Xvfb + VNC + noVNC.
2. Main follow-up candidate: standalone Selkies.
3. Best UX candidate if real containers become available: Neko.
4. Do not adopt Kasm or Guacamole as the Mobdesk core now.

The POCO F6 spike must measure startup/reconnect time, Firefox/Xvfb/encoder
memory and CPU, touch/keyboard latency, persistent profile stability, audio,
clipboard, upload/download, background Termux behavior, wake-lock needs and
localhost/LAN/Tailscale exposure.

Sources: [Selkies standalone start](https://selkies-project.github.io/selkies/start/),
[Neko installation](https://neko.m1k1o.net/docs/v3/installation),
[noVNC](https://github.com/novnc/noVNC),
[jlesage/docker-firefox](https://github.com/jlesage/docker-firefox),
[Guacamole architecture](https://guacamole.apache.org/doc/1.5.2/gug/guacamole-architecture.html),
[Kasm requirements](https://www.kasmweb.com/docs/latest/install/system_requirements.html),
and [PRoot-Distro](https://github.com/termux/proot-distro).
