
# ⚡ drop

> **Instant, zero-config, bidirectional P2P file transfers between Linux and any mobile device.**

[![Go Report Card](https://goreportcard.com/badge/github.com/raunakranjann/drop)](https://goreportcard.com/report/github.com/raunakranjann/drop)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform: Linux](https://img.shields.io/badge/Platform-Linux-orange.svg)]()
[![Architecture: amd64](https://img.shields.io/badge/Arch-amd64-green.svg)]()

`drop` is an ultra-lightweight CLI utility for Ubuntu/Debian designed to transfer files between your computer and mobile devices effortlessly. It prints a high-contrast UTF-8 QR code directly inside your terminal—scan it with any phone camera to stream files over your local Wi-Fi or Hotspot with zero companion apps required.

---

## ✨ Features

- 📱 **Zero Mobile Apps:** Transfers work out-of-the-box using standard mobile browsers (Safari, Chrome, Firefox).
- 📶 **100% Offline & Private:** Operates directly over local Wi-Fi or Personal Hotspot. No telemetry, no external accounts, no cloud relays.
- 🔄 **Bidirectional Transfers:**
  - Send from Linux to phone: `drop movie.mp4`
  - Receive from phone to Linux: `drop -r`
- 🎨 **Glassmorphic Mobile UI:** Modern, responsive upload interface with file selection and live progress.
- 📊 **Real-time Terminal Feedback:** Displays real-time transfer speed (MB/s), bytes transmitted, and ETA using an interactive terminal progress bar.
- 🌐 **Online Mode Fallback:** Add the `-online` flag to establish an encrypted public HTTPS tunnel via Cloudflare Quick Tunnels for transfers across cellular/5G networks.
- 🛡️ **Clean Teardown:** Handles termination signals (`Ctrl+C`) gracefully to ensure ports and child processes are freed immediately.

---

## 🚀 Installation

### Option 1: Debian Package (`.deb`) — Recommended

Download the pre-compiled `.deb` file from your repository releases and run:

```bash
sudo dpkg -i drop_1.0.0_amd64.deb

```

*To cleanly uninstall at any time:*

```bash
sudo apt remove drop

```

---

### Option 2: One-Line Installer Script

```bash
curl -sSL [https://raw.githubusercontent.com/raunakranjann/drop/main/install.sh](https://raw.githubusercontent.com/raunakranjann/drop/main/install.sh) | bash

```

---

### Option 3: Build from Source

```bash
git clone [https://github.com/raunakranjann/drop.git](https://github.com/raunakranjann/drop.git)
cd drop
go build -ldflags="-s -w" -o drop main.go
sudo mv drop /usr/local/bin/

```

---

## 📖 Usage

### 1. Send a File to Mobile (Default Local Mode)

```bash
drop ~/Documents/notes.pdf

```

1. The terminal displays a QR code alongside a direct LAN link.
2. Scan the QR code using your phone camera (or open the displayed URL in your mobile browser).
3. The download begins immediately while your terminal monitors the transfer progress.

---

### 2. Receive Files from Mobile to Linux

```bash
drop -r

```

By default, uploaded files are saved in your current working directory. To choose a specific destination directory:

```bash
drop -r -dir ~/Downloads

```

1. Scan the QR code with your phone.
2. Select or drop any file onto the web page.
3. The file streams directly to your machine and saves to the target folder.

---

### 3. Remote / Cellular Mode (Online)

To send or receive files across different networks (e.g., phone on mobile data, laptop on Wi-Fi):

```bash
drop -online report.zip

```

*(Requires `cloudflared` installed on your Linux machine).*

---

### 4. Custom Port

```bash
drop -port 9090 presentation.pdf

```

---

## 🛠️ CLI Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-r` | `false` | Receive mode (upload from mobile to Linux) |
| `-dir` | `.` | Destination directory for received files |
| `-online` | `false` | Launch a secure Cloudflare HTTPS tunnel for non-local transfers |
| `-port` | `8989` | Port used by the local transfer server |

---

