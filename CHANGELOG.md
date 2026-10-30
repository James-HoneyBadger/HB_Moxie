# Changelog

All notable changes to Honey Ledger are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Releases before 1.4.2 are not catalogued here; see the commit history for
details.

## [Unreleased]

### Added

- Shared expenses: split an expense transaction with up to 20 people, review what
  each person owes per currency, and drill into the individual expenses behind a
  total. Backed by `GET /api/v1/transactions/splits/{list,summary,person}.json`
  and `POST /api/v1/transactions/splits/update.json`.
- A `Shared Expenses` page under Basis Data, with a split editor that shows the
  transaction amount and the unallocated remainder while editing. Shared
  expenses are also marked directly in the transaction list, where a chip shows
  the amount others owe and opens the split editor, and expense rows offer the
  same editor next to the description.
- Versioned database migrations with a `schema_migration` ledger, applied at
  startup after the automatic schema sync. Each migration runs once, inside a
  transaction, and is recorded only on success.
- Pull request CI covering `gofmt`, `go vet`, generated OpenAPI verification, a
  backend build, `go test -short -race`, frontend lint/tests/build, Electron
  entry-point syntax checks, a `govulncheck` vulnerability scan, and repository
  hygiene checks.
- `SECURITY.md` with scope, threat model, and private disclosure process.
- Dependabot configuration for the Go modules, `admin/`, npm packages, and
  GitHub Actions, with minor and patch updates grouped.
- Export tests for the admin tool, covering dump contents, SQL quote escaping,
  and error propagation.
- Background update checks for packaged builds: the app downloads a newer
  release and installs it on quit, notifying the user when it is ready. Only
  AppImage and Windows installer builds self-update; set
  `HONEY_DISABLE_UPDATE_CHECK=1` to skip the check.

### Changed

- Renamed the application from Ratel Ledger to **Honey Ledger**. The rename covers
  the product name, package names, the application id
  (`com.honeybadger.honeyledger`), installers, exported file names, the config
  file (`honey-ledger.ini`), the SQLite database (`honey-ledger.db`) and the app
  icons, and is translated across all 19 locales.
- Existing installations upgrade in place: the previous data directory
  (`Ratel Ledger`, and `HB Moxie` before it) is migrated on first launch, the
  `ratel-ledger.ini`/`hbmoxie.ini` config files are still found and adopted, and
  the `RATEL_*`/`HBM_*` environment variables and `RATEL_DISABLE_UPDATE_CHECK`
  remain supported alongside the new `HONEY_*` names.
- Import mapping files exported before the rename still load through the legacy
  `ratelLedger…` JSON root fields.
- The GitHub repository moved to `James-HoneyBadger/honey-ledger`, and the Go
  module path followed it (`github.com/James-HoneyBadger/honey-ledger`), so
  `go install` and imports use the new location. GitHub redirects the previous
  repository and clone URLs.
- Upgrade the admin tool to Fyne 2.8.2 (from 2.5.4). Building it on Linux now
  requires the Wayland development headers in addition to X11.
- Upgrade the desktop shell to Electron 45.1 (from 33, which reached
  end-of-life in April 2025) and electron-builder 26, dropping the unused
  `electron-is-dev` dependency. The Electron 45 toolchain requires Node 22.12+.
- Builds are reproducible: `go mod download` replaces `go get .`, `npm ci`
  replaces `npm install`, and the pre-test `go clean -cache` is gone.
- `npm run lint` no longer rewrites files; `npm run lint:fix` applies fixes.
- Minimum Go version is 1.26, required by the updated `golang.org/x/net`.

### Fixed

- English was missing the translation keys for the Persian, Chinese, and Hebrew
  calendar display options and for the names of every language in the language
  picker, so those menus rendered raw identifiers. Every locale now defines the
  same key set, enforced by `scripts/check-locales.mjs` in CI.
- The backend shuts down gracefully on `SIGTERM`/`SIGINT`: it stops accepting
  connections, drains in-flight requests, stops the cron scheduler, and closes
  database connections instead of exiting mid-write.
- The admin tool reports database iteration, scan, export, and file-close
  errors instead of silently producing a partial SQL dump that appeared
  successful.
- Data race in the interval cron test, exposed by enabling the race detector.

### Security

- Release artifacts now carry a SLSA build-provenance attestation, and are
  signed with minisign when a `MINISIGN_SECRET_KEY` secret is configured. The
  verification commands are documented in `SECURITY.md`.
- The generated default-user password is printed to the console only and is no
  longer written to the log file.
- Updated `golang.org/x/net`, `golang.org/x/crypto`, `xuri/excelize/v2`,
  `quic-go`, and `go-jose/v4`, clearing all vulnerabilities reachable from the
  application code according to `govulncheck`.
- Updated `golang.org/x/image` in the admin tool, clearing the four reachable
  image-decoding vulnerabilities reported by `govulncheck`.

## [1.4.3] - 2026-10-24

### Added

- Linux package metadata and consistent installer file names.
- The release workflow can publish a GitHub release for a version tag.

## [1.4.2] - 2026-10-24

### Added

- Multi-platform application icons for the Linux, macOS, and Windows builds.

[Unreleased]: https://github.com/James-HoneyBadger/honey-ledger/compare/v1.4.3...HEAD
[1.4.3]: https://github.com/James-HoneyBadger/honey-ledger/releases/tag/v1.4.3
[1.4.2]: https://github.com/James-HoneyBadger/honey-ledger/releases/tag/v1.4.2
