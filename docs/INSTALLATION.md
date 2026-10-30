# Installation Guide

Complete instructions for installing and running Ratel Ledger as a desktop application.

---

## Table of Contents

- [System Requirements](#system-requirements)
- [Desktop Application (Electron)](#desktop-application-electron)
- [Build from Source](#build-from-source)
- [Upgrading](#upgrading)
- [Uninstalling](#uninstalling)

---

## System Requirements

| Component | Minimum |
|-----------|---------|
| OS | Linux (amd64, arm64), macOS, Windows |
| RAM | 256 MB |
| Disk | 200 MB (application) + database storage |
| Go | 1.26+ (build from source only) |
| Node.js | 20.19+ or 22.12+ (build from source only) |

---

## Desktop Application (Electron)

### Pre-built Package

Download the latest desktop installer for your platform from the [Releases](https://github.com/James-HoneyBadger/ratel-ledger/releases) page:

| Platform | Format |
|----------|--------|
| Windows | `.exe` (NSIS installer) |
| macOS | `.dmg` |
| Linux | `.AppImage`, `.deb` |

Run the installer and launch Ratel Ledger from your applications menu.

### Run from Source (Development)

```bash
# Clone the repository
git clone https://github.com/James-HoneyBadger/ratel-ledger.git
cd ratel-ledger

# Build the Go backend
go build -o ratel-ledger -tags timetzdata .

# Build the frontend
npm install
npm run build

# Launch the Electron app
cd electron
npm install
npm start
```

See [ELECTRON.md](ELECTRON.md) for full Electron build and packaging details.

---

## Build from Source

### Prerequisites

- Go 1.26 or later
- Node.js 24 or later
- npm
- GCC (for SQLite CGO compilation)

### Full Build (Backend + Frontend)

```bash
git clone https://github.com/James-HoneyBadger/ratel-ledger.git
cd ratel-ledger
./scripts/build.sh package
```

This produces a `.tar.gz` archive containing the binary, frontend assets, configuration template, email templates, and license.

### Backend Only

```bash
./scripts/build.sh backend
```

Produces the `ratel-ledger` binary in the project root.

### Frontend Only

```bash
./scripts/build.sh frontend
```

Produces optimized assets in the `dist/` directory.

### Build Options

| Flag | Description |
|------|-------------|
| `--release` / `-r` | Set release version string |
| `--output` / `-o` | Output directory (default: current directory) |
| `--no-lint` | Skip lint checks during build |
| `--no-test` | Skip test execution during build |

### Manual Build Steps

If you prefer not to use `scripts/build.sh`:

```bash
# Backend
go mod download
go build -o ratel-ledger \
  -tags timetzdata \
  -ldflags "-X main.Version=dev -X main.CommitHash=$(git rev-parse HEAD)" \
  .

# Frontend
npm ci
npm run build
```

### Package the Electron App

```bash
cd electron
npm install
npm run build    # or: npx electron-builder --linux --mac --win
```

---

## Upgrading

1. Back up your data directory and configuration file.
2. Download and install the new desktop release.
3. Database schema updates run automatically on startup when `auto_update_database = true` (the default).

---

## Uninstalling

### Windows

Use **Settings → Apps & features** or **Control Panel → Programs** to uninstall Ratel Ledger.

### macOS

Drag Ratel Ledger from `/Applications` to the Trash.

### Linux

- **AppImage:** Delete the `.AppImage` file.
- **Deb package:** `sudo apt remove ratel-ledger`
- **Manual:** Remove the application directory.

### Data Cleanup

After uninstalling, optionally remove application data:

| Platform | Data Location |
|----------|---------------|
| Windows | `%APPDATA%/Ratel Ledger/` |
| macOS | `~/Library/Application Support/Ratel Ledger/` |
| Linux | `~/.config/Ratel Ledger/` |
