# Mobdesk

[Português](README.md) | [English](README.en.md)

## Your Linux workstation, in your pocket

Mobdesk turns an Android phone into a personal development environment. Instead of depending on university computers, shared machines, or leaving personal accounts signed in on someone else’s device, you carry your projects, tools, and data with you.

Mobdesk turns the phone into a small development server:

```text
Android
└── Termux — device control layer
    └── Persistent Ubuntu — development environment
```

Ubuntu runs on the phone through PRoot-Distro. You can work directly in Termux or connect from another computer on the same network over SSH. Your files remain on your own device.

## What is it for?

Mobdesk is designed for students, developers, and professionals who need a portable Linux environment to:

- study C, JavaScript, HTML, React, Java, Go, or Python;
- create and run small and medium-sized projects for learning and development;
- start local development servers, such as `npm run dev`, and open them in a browser;
- use a phone as a personal workstation during classes, travel, or commutes;
- access the same environment from the phone or from a computer on the same network;
- keep code, configuration, and sessions without signing in to shared computers.

Mobdesk is not intended to replace a production machine, a full virtual machine, or a graphical desktop. It is a lightweight, user-controlled mobile workstation for development, learning, and local servers.

## Why Mobdesk?

### Your environment travels with you

The Ubuntu environment persists on the phone. You do not have to rebuild your setup every time you change rooms, networks, or computers.

### Your data stays yours

Projects and configuration remain on the device. This reduces the need to leave GitHub, email, messaging, or other personal accounts signed in on shared computers.

### One phone, many workflows

Edit code on the phone, open an SSH session from another computer, and publish a local development server for browser testing — all from the same environment.

### No root and no Docker on the phone

Mobdesk uses Termux and PRoot-Distro. It does not require root access, a virtual machine, or real Docker on Android.

## What is available today?

The current MVP focuses on bootstrapping the environment:

- persistent Ubuntu installation through PRoot-Distro;
- OpenSSH installation and configuration in Termux;
- SSH access on port `8022`;
- sessions opened directly inside Ubuntu;
- local IP address detection;
- password setup for SSH access;
- repeatable operations that do not reinstall existing components;
- `setup`, `start`, and `stop` commands.

The full TUI, assisted tool installation, project management, and web control center are planned for later product stages. See the [roadmap](../docs/project/ROADMAP.md) for the planned evolution.

## Installation for end users

### Requirements

- an Android phone with ARM64 architecture, as found on most current devices;
- Termux installed from a trusted source, preferably [F-Droid](https://f-droid.org/packages/com.termux/) or the [official releases](https://github.com/termux/termux-app/releases);
- approximately 1.5 GB of free storage for the base Ubuntu installation, plus additional space for your projects;
- a regular Wi-Fi network if you want to connect from another computer.

Mobdesk does not require root. Performance depends on the phone’s memory, temperature, battery, and Android/HyperOS background-process limits.

### 1. Install Mobdesk

Open Termux and run:

```bash
pkg update
pkg install -y golang git
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
```

### 2. Set up the environment

On the first run, use the binary installed by Go:

```bash
~/go/bin/mobdesk setup
```

Setup installs the required Termux components, downloads Ubuntu, creates the persistent workspace, and asks for the password used for SSH access. At the end, the `mobdesk` command is available globally.

### 3. Start your workstation

```bash
mobdesk start
```

Mobdesk starts SSH on port `8022`, keeps the device awake while it is in use, and opens an Ubuntu session directly in Termux.

To connect from another computer on the same network, use the SSH command displayed by Mobdesk, for example:

```bash
ssh -p 8022 android@192.168.1.50
```

Replace the address with the IP shown on your phone and enter the password configured during setup. The SSH connection is forwarded directly into Ubuntu.

### 4. Stop when you are done

To leave only the Ubuntu session, run:

```bash
exit
```

To stop the SSH server:

```bash
mobdesk stop
```

## How the flow works

```text
mobdesk setup
    ↓
Termux + PRoot-Distro + persistent Ubuntu
    ↓
mobdesk start
    ↓
SSH :8022 → Ubuntu session
    ↓
projects, editors, and local development servers
```

Termux is the control host. Ubuntu is the development environment. PRoot improves Linux userland compatibility, but it does not create a separate kernel or provide the isolation of a virtual machine or real container.

## Important limitations

Mobdesk is suitable for learning, development, and lightweight servers. It is not designed for:

- heavy production workloads;
- large-scale load or performance testing;
- real Docker, systemd, or a complete Linux VM;
- a full graphical desktop with guaranteed acceleration;
- privileged device access or kernel modules.

Android may suspend or terminate Termux. For a more stable experience, allow Termux to run in the background in the phone’s battery settings.

## Security

Use SSH only on trusted networks. Do not expose port `8022` directly to the public internet. For remote access outside the local network, prefer a private network such as Tailscale or an SSH tunnel. Keep important projects backed up outside the phone.

## Documentation

- [Product mission](../docs/project/MISSAO.md)
- [Current MVP](../docs/project/MVP.md)
- [Roadmap](../docs/project/ROADMAP.md)
- [Architecture and limitations](../docs/project/ARQUITETURA.md)
- [Contributing](CONTRIBUTING.md)

## License

Mobdesk is distributed under the [MIT license](../LICENSE).
