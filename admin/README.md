# Ratel Ledger System Administration Tool

A standalone desktop application for managing Ratel Ledger configuration, users, and database. Built with [Fyne](https://fyne.io/) — a native Go GUI toolkit.

## Features

### Dashboard
- User counts (total, active, disabled, deleted)
- Data counts (accounts, transactions, categories, tags)
- System info (config path, database path, database size)

### Configuration Management
- Browse all `ratel-ledger.ini` sections in a sidebar list
- Edit any configuration key with live form fields
- Password/secret fields are masked by default
- **Save** creates a timestamped backup before writing
- **Reset to File** discards changes and reloads from disk

### User Management
- Full user table with column headers (UID, username, email, status, etc.)
- Select a row then use toolbar buttons:
  - **Enable / Disable** user accounts
  - **Delete** (soft-delete) / **Restore** users
  - **Verify Email** manually
- Confirmation dialogs for destructive actions

### Database Browser
- **Browse tab** — Select any table from the sidebar; paginated data grid with column headers, 50 rows per page
- **Query tab** — Execute read-only SELECT/PRAGMA queries (max 1,000 rows returned)
- **Execute tab** — Run write statements (INSERT, UPDATE, DELETE, etc.) with confirmation dialog
- **Export tab** — Save a full SQL dump of the database via native file dialog

## Building

### Prerequisites

Fyne draws through OpenGL and links against both the X11 and Wayland
backends, so the desktop development headers are required.

**Fedora/RHEL:**
```bash
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel \
  libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel \
  wayland-devel libxkbcommon-devel
```

**Ubuntu/Debian:**
```bash
sudo apt install -y gcc libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev
```

### Compile

```bash
cd admin
go build -o ratel-ledger-admin .
```

## Running

Run from the Ratel Ledger project root so the tool can find `conf/ratel-ledger.ini`:

```bash
./admin/ratel-ledger-admin
```

The tool will:
1. Load `conf/ratel-ledger.ini` to find database settings
2. Open the SQLite database directly
3. Launch a native desktop window (1400x850)

## Security Notes

- Runs locally as a native desktop application — no network ports opened
- Configuration saves always create timestamped backups (`.bak.YYYYMMDD-HHMMSS`)
- Destructive operations (disable, delete, SQL execute) require confirmation dialogs
- Password and secret config values are masked in the editor

## File Structure

```
admin/
├── main.go            # Entry point, types, data access, utilities
├── dashboard.go       # Dashboard statistics tab
├── config_tab.go      # Configuration INI editor tab
├── users_tab.go       # User management tab
├── database_tab.go    # Database browser, SQL console, export
├── go.mod             # Go module dependencies
├── go.sum             # Dependency checksums
└── README.md          # This file
```

## Dependencies

- **[Fyne v2](https://fyne.io/)** — Native Go GUI toolkit
- **[go-sqlite3](https://github.com/mattn/go-sqlite3)** — SQLite3 driver (CGO)
- **[go-ini](https://gopkg.in/ini.v1)** — INI file parser
- Go 1.25+, CGO enabled
