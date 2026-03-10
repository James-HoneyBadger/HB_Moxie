# HB Moxie

A standalone desktop personal finance management application.

HB Moxie is a self-contained Electron desktop application for individuals who want full control over their financial data. All data is stored locally on your machine — no cloud, no browser, no external services required.

## Features

- **Desktop Only** — Native Electron application for Windows, macOS, and Linux. No browser needed.
- **Self-Contained** — Go backend runs as a child process inside the Electron app.
- **System Tray** — Minimize to tray, global keyboard shortcuts, desktop notifications.
- **Auto-Login** — Single-user mode skips the login screen automatically.
- **Automatic Backups** — Scheduled backups with configurable retention.
- **Multiple Databases** — SQLite (default), MySQL, or PostgreSQL.
- **Dark Mode** — Full dark theme support with automatic system detection.
- **Multi-Currency** — 170+ currencies with automatic exchange rate updates from 16 data sources.
- **Multi-Language** — 19 languages including English, German, French, Spanish, Italian, Japanese, Korean, Chinese, and more.
- **Multi-Timezone** — Per-user timezone configuration with timezone-aware statistics.
- **Data Import** — Import from CSV, TSV, Excel, OFX, QFX, QIF, IIF, Camt.052/053, MT940, GnuCash, Firefly III, Beancount, and more.
- **Data Export** — Export transactions to CSV, TSV, or ZIP archives.
- **Backups** — Create, download, restore, and manage full database backups.
- **AI Receipt Recognition** — Extract transaction data from receipt images using LLM providers (OpenAI, Anthropic, Google AI, Ollama, LM Studio, and others).
- **MCP Support** — Model Context Protocol integration for AI agent access to your financial data.
- **Scheduled Transactions** — Recurring transaction automation via built-in cron jobs.
- **Transaction Templates** — Save and reuse common transaction patterns.
- **Insights Explorer** — Custom data queries with chart and table visualization.
- **Account Reconciliation** — Reconciliation statement support for account balancing.
- **Map Integration** — Geotagging with 13 map providers (OpenStreetMap, Google Maps, Baidu, Amap, and others).
- **Cloud Settings Sync** — Sync application preferences across sessions.
- **API Tokens** — Generate tokens for programmatic access and automation.

## Quick Start

### Run the Desktop App

```bash
cd electron
npm install
npm start
```

### Build from Source

```bash
./scripts/build.sh package
```

Then extract the archive and run the binary. See [docs/INSTALLATION.md](docs/INSTALLATION.md) for complete instructions.

## Default Credentials

On first launch with no existing database, HB Moxie creates a default administrator account with a random password displayed in the startup log. If auto-login is enabled (default), the login screen is skipped entirely.

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Shift+N` / `Cmd+Shift+N` | Quick-add transaction |
| `Ctrl+Shift+M` / `Cmd+Shift+M` | Show/focus main window |

## Admin Tool

A standalone Fyne-based desktop admin tool is available in the `admin/` directory for database management, user administration, and diagnostics.

```bash
cd admin
go run .
```

## Documentation

| Document | Description |
|----------|-------------|
| [Installation Guide](docs/INSTALLATION.md) | Installation methods: binary, source, Electron packaging |
| [Configuration Reference](docs/CONFIGURATION.md) | Complete configuration file reference |
| [User Guide](docs/USER_GUIDE.md) | End-user guide for all application features |
| [API Reference](docs/API_REFERENCE.md) | REST API endpoint documentation |
| [Architecture Overview](docs/ARCHITECTURE.md) | System design, project structure, and technology stack |
| [Development Guide](docs/DEVELOPMENT.md) | Setting up a development environment, building, testing |
| [Electron Desktop App](docs/ELECTRON.md) | Desktop application build, packaging, and usage |
| [CLI Reference](docs/CLI_REFERENCE.md) | Command-line interface for server, database, and user management |
| [Contributing](docs/CONTRIBUTING.md) | Contribution guidelines, code style, and workflow |

## Technology Stack

| Layer    | Technology |
|----------|------------|
| Backend  | Go 1.25, Gin, xorm |
| Frontend | Vue 3, Vuetify 3, TypeScript, Vite |
| Desktop  | Electron 33 |
| Database | SQLite 3 / MySQL / PostgreSQL |
| Build    | Vite 7, Go toolchain, electron-builder |

## License

[MIT License](LICENSE) — Copyright (c) 2020-2026 Honey Badger Universe
