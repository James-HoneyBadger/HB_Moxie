# Electron Desktop Application

Guide for building, running, and packaging the Ratel Ledger standalone desktop application.

---

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Running in Development](#running-in-development)
- [How It Works](#how-it-works)
- [Building Distributable Packages](#building-distributable-packages)
- [Updates](#updates)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Electron shell wraps Ratel Ledger into a native desktop application. It:

- Spawns the Go backend as a child process on a random free port
- Serves the Vue.js frontend via a custom `app://ratel-ledger` protocol
- Proxies API requests to the Go backend
- Provides native window management (menu bar, window controls, system tray)

Users get a self-contained application with no browser or terminal needed.

---

## Prerequisites

| Tool | Version |
|------|---------|
| Node.js | 22.12+ (required by the Electron 45 toolchain) |
| npm | 9+ |
| Go | 1.26+ (to build the backend binary) |
| GCC | For SQLite CGO compilation |

---

## Running in Development

### 1. Build the Backend Binary

From the project root:

```bash
go build -o ratel-ledger .
```

### 2. Build the Frontend

```bash
npm run build
```

### 3. Start the Electron App

```bash
cd electron
npm ci
npm start
```

The Electron app will:
1. Find a free port
2. Start `../ratel-ledger server run` as a child process
3. Wait for the backend TCP port to accept connections
4. Open the main window pointing to `app://ratel-ledger/desktop.html`

---

## How It Works

### Process Architecture

```
Electron Main Process (main.js)
│
├── Child Process: Go Backend (ratel-ledger server run)
│     └── Listens on 127.0.0.1:<random-port>
│
├── Custom Protocol: app://ratel-ledger
│     └── Serves files from dist/ directory
│
├── Request Interception
│     └── Proxies /api/*, /avatar/*, /pictures/*, etc. to Go backend
│
└── BrowserWindow
      └── Loads app://ratel-ledger/desktop.html
```

### Custom Protocol

The `app://ratel-ledger` protocol serves frontend files directly from the filesystem without an HTTP server, improving security and performance.

### API Proxying

These URL paths are intercepted and forwarded to the Go backend:

| Path Pattern | Description |
|-------------|-------------|
| `/api/` | All REST API calls |
| `/server_settings.js` | Server configuration JavaScript |
| `/avatar/` | User avatar images |
| `/pictures/` | Transaction pictures |
| `/proxy/` | Map tile and annotation proxies |
| `/mcp` | MCP JSON-RPC endpoint |
| `/oauth2/` | OAuth2 login/callback |
| `/healthz.json` | Health check |
| `/_AMapService/` | Amap API proxy |

### Preload Script

`electron/preload.js` exposes a limited API to the renderer:

```javascript
window.ratelLedger = {
    platform: process.platform,    // 'linux', 'darwin', 'win32'
    isDesktopApp: true,
    apiBaseUrl: 'app://ratel-ledger'
};
```

### Security

- `contextIsolation: true` — Preload runs in an isolated context
- `nodeIntegration: false` — No Node.js APIs in the renderer
- `sandbox: true` — Renderer process is sandboxed
- Content Security Policy is enforced
- All permission requests (camera, microphone, etc.) are denied
- External links open in the system browser, not in the Electron window

### Window Settings

- Default size: 1400 x 900
- Minimum size: 800 x 600
- Frame: Standard OS window frame

### Backend Lifecycle

- **Startup**: The Go binary is spawned with `RATEL_WORK_DIR=<userData>`, `RATEL_SERVER_HTTP_ADDR=127.0.0.1`, and `RATEL_SERVER_HTTP_PORT=<port>`.
- **Readiness**: The main process waits until the backend TCP port accepts connections.
- **Single instance**: A second launch focuses the existing window instead of starting another backend against the same SQLite file.
- **Shutdown**: On app quit, `SIGTERM` is sent to the Go process. If it doesn't exit within 3 seconds, `SIGKILL` is sent.

---

## Building Distributable Packages

### All Platforms

```bash
cd electron
npm run build
```

### Platform-Specific

```bash
# Linux (AppImage + .deb)
npm run build:linux

# macOS (.dmg)
npm run build:mac

# Windows (NSIS installer)
npm run build:win
```

### Build Output

Packages are output to `electron/dist/`.

### CI Packages

The reusable `.github/workflows/build-electron.yml` workflow builds native packages on GitHub-hosted runners:

| Runner | Artifacts |
|--------|-----------|
| `ubuntu-24.04` | AppImage, `.deb`, `latest-linux.yml` |
| `macos-latest` | `.dmg`, `latest-mac.yml` |
| `windows-latest` | NSIS `.exe`, `latest.yml` |

The release workflow calls it for `v*.*.*` tags and attaches the installers to the draft GitHub release. Snapshot builds attach the same artifacts to the workflow run. Packages are unsigned by default; configure `CSC_LINK`/`CSC_KEY_PASSWORD` (and notarization credentials for macOS) in repository secrets to sign them.

### Build Configuration

The `electron/package.json` contains the `electron-builder` configuration:

```json
{
  "build": {
    "appId": "com.honeybadger.ratelledger",
    "productName": "Ratel Ledger",
    "extraResources": [
      { "from": "../", "to": "backend/", "filter": ["ratel-ledger", "ratel-ledger.exe"] },
      { "from": "../conf", "to": "backend/conf" },
      { "from": "../templates", "to": "backend/templates" },
      { "from": "../dist", "to": "backend/public" }
    ]
  }
}
```

The backend binary, config template, email templates, and frontend assets are bundled as extra resources.

### Pre-Build Checklist

Before building a distributable package:

1. Build the Go backend binary for the target platform
2. Build the frontend (`npm run build` from project root)
3. Ensure `conf/ratel-ledger.ini` has sensible defaults
4. Run `cd electron && npm run build:<platform>`

---

## Updates

Packaged builds check GitHub releases in the background, download a newer
version, and install it the next time the app quits. The user sees a desktop
notification when the download is ready; nothing is installed while the app is
running.

| Install type | Behaviour |
|--------------|-----------|
| AppImage (Linux) | Self-updates |
| NSIS installer (Windows) | Self-updates |
| `.deb` (Linux) | Handled by the package manager; the app only logs that self-update is unavailable |
| `.dmg` (macOS) | Disabled — self-update needs a signed and notarized bundle |
| Development build | Disabled (only packaged builds check) |

How it fits together:

- `build.publish` in `electron/package.json` points at this GitHub repository.
  `electron-builder` turns that into `resources/app-update.yml`, which
  `electron-updater` reads at runtime.
- The release workflow uploads `latest.yml`, `latest-linux.yml`, and
  `latest-mac.yml` alongside the installers. **The release must be published,
  not left as a draft** — `electron-updater` cannot see draft releases.
- Update failures are logged as warnings and never affect the running app.

To disable the check — for packaging tests or offline verification — set:

```bash
RATEL_DISABLE_UPDATE_CHECK=1
```

The tray menu gains a **Check for Updates** entry on platforms where
self-update is supported.

---

## Configuration

The Electron app uses `conf/ratel-ledger.ini` for configuration. On first launch of a packaged build, the template is copied from the read-only application resources into Electron's per-user data directory (`app.getPath('userData')`):

```
<userData>/
  conf/ratel-ledger.ini
  templates/
  data/ratel-ledger.db
  log/ratel-ledger.log
  storage/
```

Subsequent launches reuse the existing files, so generated secrets and user data survive upgrades. On first launch after the rename, an existing `HB Moxie` user-data directory and a legacy `conf/hbmoxie.ini` are migrated automatically. The backend binary itself remains in `resources/backend/`.

---

## Troubleshooting

### Backend Fails to Start

- Check that the backend binary (`ratel-ledger`, or `ratel-ledger.exe` on Windows) exists and is executable.
- Check the log output in the terminal (when running `npm start`).
- Ensure no other process is blocking the randomly selected port.

### Blank Window

- The backend may still be starting. Wait a few seconds.
- Check the DevTools console (View → Toggle Developer Tools) for errors.
- Verify the frontend was built (`dist/` directory exists with assets).

### "Cannot find module" Errors

Run `npm ci` in the `electron/` directory to install Electron dependencies.

### Port Conflicts

The Electron app selects a random free port automatically. If you experience issues, check for firewall rules blocking localhost connections.

### Backend Logs

When running in development, backend logs appear in the terminal where `npm start` was executed. In packaged mode, logs are written to the `log/` directory alongside the data directory.
