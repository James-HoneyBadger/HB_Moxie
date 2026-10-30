# Contributing

Guidelines for contributing to Honey Ledger.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Branching Strategy](#branching-strategy)
- [Commit Messages](#commit-messages)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Testing Requirements](#testing-requirements)
- [Localization Contributions](#localization-contributions)
- [Reporting Issues](#reporting-issues)
- [License](#license)

---

## Getting Started

1. Fork the repository on GitHub.
2. Clone your fork:
   ```bash
   git clone https://github.com/<your-username>/honey-ledger.git
   cd honey-ledger
   ```
3. Set up the development environment (see [Development Guide](DEVELOPMENT.md)).
4. Create a feature branch and make your changes.
5. Submit a pull request.

---

## Development Setup

See [docs/DEVELOPMENT.md](DEVELOPMENT.md) for complete setup instructions. Quick start:

```bash
go mod download
npm ci
go build -o honey-ledger . && ./honey-ledger server run  # Terminal 1
npm run serve                                           # Terminal 2
```

---

## Branching Strategy

| Branch | Purpose |
|--------|---------|
| `main` | Stable release branch |
| `feature/*` | New features |
| `fix/*` | Bug fixes |
| `docs/*` | Documentation changes |
| `refactor/*` | Code refactoring |

Create your branch from `main`:

```bash
git checkout -b feature/my-feature main
```

---

## Commit Messages

Use clear, descriptive commit messages:

```
<type>: <short description>

<optional longer description>
```

### Types

| Type | Use For |
|------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code formatting (no logic changes) |
| `refactor` | Code restructuring (no behavior change) |
| `test` | Adding or updating tests |
| `chore` | Build, CI, dependency updates |

### Examples

```
feat: add MT940 bank statement import support
fix: prevent duplicate transactions during concurrent imports
docs: update Electron packaging instructions
refactor: extract exchange rate caching into dedicated service
test: add unit tests for QIF date parsing
chore: update Go dependencies to latest patch versions
```

---

## Code Style

### Go

- Follow standard Go conventions (`gofmt`, `go vet`).
- Use meaningful variable and function names.
- Keep functions focused — single responsibility.
- Handle all errors explicitly; do not use `_` to discard errors in production code.
- Add comments for exported functions and types.
- Use the project's error types from `pkg/errs/` rather than raw `error` values for API errors.

### TypeScript / Vue

- Follow the ESLint configuration (`eslint.config.mjs`).
- Use TypeScript strict mode — no `any` unless absolutely necessary.
- Use Vue 3 Composition API with `<script setup lang="ts">`.
- Use Pinia stores for shared state; avoid component-local state for data that crosses component boundaries.
- Use `useI18n()` and `tt()` for all user-facing strings.
- Import paths use the `@/` alias (maps to `src/`).

### CSS / SCSS

- Use Vuetify components and utility classes when possible.
- Custom styles go in `src/styles/`.
- Use SCSS variables for theming consistency.

---

## Pull Request Process

1. **Ensure your code builds** — both backend and frontend must compile without errors:
   ```bash
   go build -o honey-ledger .
   npx vue-tsc --noEmit
   npm run build
   ```

2. **Run tests**:
   ```bash
   go test -short ./...   # skips live third-party API tests
   npm test
   ```
   Run `go test ./...` when you need the full suite; some exchange-rate tests call external APIs and require network access.

3. **Run linters**:
   ```bash
   go vet ./...
   npm run lint        # check only; use `npm run lint:fix` to apply fixes
   ```

   CI runs the same checks on every pull request: `gofmt`, `go vet`, `go test -short -race`, generated OpenAPI verification, a `govulncheck` vulnerability scan, the repository hygiene checks, `npm run lint`, `npm test`, and a frontend build. The `admin/` module gets its own job covering formatting, vet, tests, build, and a vulnerability scan. Run them locally before opening the PR to avoid a round trip.

   Run the vulnerability scan locally with `govulncheck ./...`; it exits non-zero when your code reaches a known vulnerability.

4. **Update documentation** if your change affects:
   - Configuration options → update `docs/CONFIGURATION.md`
   - API endpoints → update `docs/API_REFERENCE.md`
   - User-facing features → update `docs/USER_GUIDE.md`
   - Architecture decisions → update `docs/ARCHITECTURE.md`

5. **Update locale files** if you add or change user-facing strings. At minimum, update `src/locales/en.json`. Other languages can be updated separately.

6. **Submit the PR** with:
   - A clear title describing the change
   - A description of what was changed and why
   - Reference to any related issues

---

## Testing Requirements

### Backend

- New services and API handlers should have corresponding tests.
- Test files live alongside the code they test (e.g., `my_service_test.go`).
- Use table-driven tests for multiple input scenarios.

### Frontend

- Jest is configured via `jest.config.ts` with `tsconfig.jest.json`.
- Component tests are welcome but not strictly required for every change.

### What Must Pass

Before merging, all of the following must pass:

- `go build .` — Backend compiles
- `go vet ./...` — No lint issues
- `go test -short ./...` — Backend unit tests pass (use the full suite for external API tests)
- `npx vue-tsc --noEmit` — TypeScript type checking passes
- `npm run build` — Frontend builds successfully

---

## Localization Contributions

Translation contributions are highly valued.

### Updating Existing Translations

1. Edit the appropriate file in `src/locales/`.
2. Keep key names identical to the English version.
3. Make sure translations fit the UI context (length, formality, etc.).

Every locale must define exactly the same keys as `src/locales/en.json`, which
is the canonical list. Some keys are looked up dynamically — `t('calendar.' +
name)` and ``t(`language.${name}`)`` — so a key missing from `en.json` renders
as a raw identifier in English, and one missing from a translation falls back
to English. Run the check before pushing:

```bash
node scripts/check-locales.mjs
```

It fails on missing keys, keys that `en.json` does not define, and empty
values, then reports how many strings in each locale are still identical to
English — a rough untranslated count (English words that are correct in a
translation, such as product names, count against it too).

### Adding a New Language

1. Create a new locale file: `src/locales/<language_code>.json` with the same
   key set as `en.json`, and translate all values.
2. Register the new locale in the locale configuration.
3. Add the language to `contributors.json` under the `translators` section.

### Current Languages

English, German (de), Spanish (es), French (fr), Italian (it), Japanese (ja), Kannada (kn), Korean (ko), Dutch (nl), Portuguese-Brazil (pt_BR), Russian (ru), Slovenian (sl), Tamil (ta), Thai (th), Turkish (tr), Ukrainian (uk), Vietnamese (vi), Simplified Chinese (zh_Hans), Traditional Chinese (zh_Hant).

---

## Reporting Issues

### Bug Reports

Include:
- Steps to reproduce
- Expected vs. actual behavior
- OS/version information
- Server logs if applicable (redact sensitive data)
- Screenshots if it's a UI issue

### Feature Requests

Describe:
- The use case or problem being solved
- Proposed solution (if you have one)
- Impact on existing functionality

### Security Issues

Never report a vulnerability in a public issue. Follow the process in
[SECURITY.md](../SECURITY.md): open a private advisory or email
<james@honey-badger.org>.

---

## Repository Hygiene Checks

Two checks guard the project's local-first guarantees. CI runs both; run them
before pushing:

```bash
node scripts/check-brand.mjs        # product identity matches across manifests
./scripts/check-no-telemetry.sh     # no analytics or crash-reporting endpoints
```

`check-brand.mjs` keeps the Electron `appId`, the admin tool's Fyne application
ID, both `package.json` versions, and the copyright holder in agreement.
`check-no-telemetry.sh` fails if an analytics or crash-reporting host appears in
`src/`, `electron/`, `pkg/`, or `cmd/`. Extend its pattern list only for
services the user configures themselves (LLM, map, mail, and storage providers
are out of scope).

---

## License

By contributing to Honey Ledger, you agree that your contributions will be licensed under the [MIT License](../LICENSE).
