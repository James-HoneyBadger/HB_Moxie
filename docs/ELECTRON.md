# Electron Desktop Application

Guide for building, running, and packaging the HB Moxie standalone desktop application.

---

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Running in Development](#running-in-development)
- [How It Works](#how-it-works)
- [Building Distributable Packages](#building-distributable-packages)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Electron shell wraps HB Moxie into a native desktop application. It:

- Spawns the Go backend as a child process on a random free port
- Serves the Vue.js frontend via a custom `app://hbmoxie` protocol
- Proxies API requests to the Go backend
- Provides native window management (menu bar, window controls, system tray)

Users get a self-contained application with no browser or terminal needed.

---

## Prerequisites

| Tool | Version |
|------|---------|
| Node.js | 24+ |
| npm | 10+ |
| Go | 1.25+ (to build the backend binary) |
| GCC | For SQLite CGO compilation |

---

## Running in Development

### 1. Build the Backend Binary

From the project root:

```bash
go build -o hbmoxie_bin .
```

### 2. Build the Frontend

```bash
npm run build
cp dist/desktop.html dist/index.html
```

### 3. Start the Electron App

```bash
cd electron
npm install
npm start
```

The Electron app will:
1. Find a free port
2. Start `../hbmoxie_bin server run` as a child process
3. Wait for the backend health check to respond
4. Open the main window pointing to `app://hbmoxie/desktop.html`

---

## How It Works

### Process Architecture

```
Electron Main Process (main.js)
│
├── Child Process: Go Backend (hbmoxie_bin server run)
│     └── Listens on 127.0.0.1:<random-port>
│
├── Custom Protocol: app://hbmoxie
│     └── Serves files from dist/ directory
│
├── Request Interception
│     └── Proxies /api/*, /avatar/*, /pictures/*, etc. to Go backend
│
└── BrowserWindow
      └── Loads app://hbmoxie/desktop.html
```

### Custom Protocol

The `app://hbmoxie` protocol serves frontend files directly from the filesystem without an HTTP server, improving security and performance.

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
window.hbmoxie = {
    platform: process.platform,    // 'linux', 'darwin', 'win32'
    isDesktopApp: true,
    apiBaseUrl: 'app://hbmoxie'
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

- **Startup**: The Go binary is spawned with environment variables `HBM_SERVER_HTTP_ADDR=127.0.0.1` and `HBM_SERVER_HTTP_PORT=<port>`.
- **Health polling**: The main process polls `/healthz.json` until the backend responds.
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

### Build Configuration

The `electron/package.json` contains the `electron-builder` configuration:

```json
{
  "build": {
    "appId": "com.hbmoxie.desktop",
    "productName": "HB Moxie",
    "extraResources": [
      { "from": "../hbmoxie_bin", "to": "backend/hbmoxie" },
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
3. Copy `dist/desktop.html` to `dist/index.html`
4. Ensure `conf/hbmoxie.ini` has sensible defaults
5. Run `cd electron && npm run build:<platform>`

---

## Configuration

The Electron app uses `conf/hbmoxie.ini` for configuration. When packaged, the config file is located at:

| Platform | Path |
|----------|------|
| Linux | `<install-dir>/resources/backend/conf/hbmoxie.ini` |
| macOS | `HB Moxie.app/Contents/Resources/backend/conf/hbmoxie.ini` |
| Windows | `<install-dir>/resources/backend/conf/hbmoxie.ini` |

The backend data directory (SQLite database, logs, storage) defaults to the `data/` folder relative to the binary location.

---

## Troubleshooting

### Backend Fails to Start

- Check that `hbmoxie_bin` (or `hbmoxie` when packaged) exists and is executable.
- Check the log output in the terminal (when running `npm start`).
- Ensure no other process is blocking the randomly selected port.

### Blank Window

- The backend may still be starting. Wait a few seconds.
- Check the DevTools console (View → Toggle Developer Tools) for errors.
- Verify the frontend was built (`dist/` directory exists with assets).

### "Cannot find module" Errors

Run `npm install` in the `electron/` directory to install Electron dependencies.

### Port Conflicts

The Electron app selects a random free port automatically. If you experience issues, check for firewall rules blocking localhost connections.

### Backend Logs

When running in development, backend logs appear in the terminal where `npm start` was executed. In packaged mode, logs are written to the `log/` directory alongside the data directory.
