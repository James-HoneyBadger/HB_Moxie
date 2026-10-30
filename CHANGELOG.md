# Changelog

All notable changes to Ratel Ledger are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Releases before 1.4.2 are not catalogued here; see the commit history for
details.

## [Unreleased]

### Added

- Pull request CI covering `gofmt`, `go vet`, generated OpenAPI verification, a
  backend build, `go test -short -race`, frontend lint/tests/build, Electron
  entry-point syntax checks, a `govulncheck` vulnerability scan, and repository
  hygiene checks.
- `SECURITY.md` with scope, threat model, and private disclosure process.
- Dependabot configuration for the Go modules, `admin/`, npm packages, and
  GitHub Actions, with minor and patch updates grouped.
- Export tests for the admin tool, covering dump contents, SQL quote escaping,
  and error propagation.

### Changed

- Builds are reproducible: `go mod download` replaces `go get .`, `npm ci`
  replaces `npm install`, and the pre-test `go clean -cache` is gone.
- `npm run lint` no longer rewrites files; `npm run lint:fix` applies fixes.
- Minimum Go version is 1.26, required by the updated `golang.org/x/net`.

### Fixed

- The backend shuts down gracefully on `SIGTERM`/`SIGINT`: it stops accepting
  connections, drains in-flight requests, stops the cron scheduler, and closes
  database connections instead of exiting mid-write.
- The admin tool reports database iteration, scan, export, and file-close
  errors instead of silently producing a partial SQL dump that appeared
  successful.
- Data race in the interval cron test, exposed by enabling the race detector.

### Security

- The generated default-user password is printed to the console only and is no
  longer written to the log file.
- Updated `golang.org/x/net`, `golang.org/x/crypto`, `xuri/excelize/v2`,
  `quic-go`, and `go-jose/v4`, clearing all vulnerabilities reachable from the
  application code according to `govulncheck`.

## [1.4.3] - 2026-10-24

### Added

- Linux package metadata and consistent installer file names.
- The release workflow can publish a GitHub release for a version tag.

## [1.4.2] - 2026-10-24

### Added

- Multi-platform application icons for the Linux, macOS, and Windows builds.

[Unreleased]: https://github.com/James-HoneyBadger/ratel-ledger/compare/v1.4.3...HEAD
[1.4.3]: https://github.com/James-HoneyBadger/ratel-ledger/releases/tag/v1.4.3
[1.4.2]: https://github.com/James-HoneyBadger/ratel-ledger/releases/tag/v1.4.2
