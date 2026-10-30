# Architecture Overview

System design, project structure, and technology decisions for Honey Ledger.

---

## Table of Contents

- [System Overview](#system-overview)
- [High-Level Architecture](#high-level-architecture)
- [Backend Architecture](#backend-architecture)
- [Frontend Architecture](#frontend-architecture)
- [Electron Desktop Shell](#electron-desktop-shell)
- [Data Flow](#data-flow)
- [Database Schema](#database-schema)
- [Authentication & Authorization](#authentication--authorization)
- [File Storage](#file-storage)
- [External Integrations](#external-integrations)
- [Build & Deployment Pipeline](#build--deployment-pipeline)
- [Security Model](#security-model)

---

## System Overview

Honey Ledger is a desktop personal finance application built with Electron:

The Go backend runs as a child process of Electron, which hosts the frontend via a custom `app://` protocol.

```
┌──────────────────────────────────────────────────┐
│                   User Interface                  │
│           Vue 3 + Vuetify 3 (SPA)                │
├──────────────────────────────────────────────────┤
│              Electron Shell (desktop host)         │
├──────────────────────────────────────────────────┤
│                  Go HTTP Server                   │
│             Gin Framework + Middleware             │
├──────────────┬───────────────┬───────────────────┤
│   API Layer  │ Service Layer │  Background Jobs   │
│  (Handlers)  │   (Logic)     │  (Cron Scheduler)  │
├──────────────┴───────────────┴───────────────────┤
│              Data Access Layer (xorm)              │
├──────────────┬───────────────┬───────────────────┤
│   SQLite 3   │    MySQL      │   PostgreSQL       │
└──────────────┴───────────────┴───────────────────┘
```

---

## High-Level Architecture

### Request Flow

```
Electron
    │
    ▼
Gin Router ──► Middleware Chain ──► API Handler ──► Service ──► DataStore ──► Database
    │               │                     │            │
    │          (Recovery,             (Parse        (Business      (xorm ORM,
    │           JWT Auth,             Request,      Logic,         SQL queries)
    │           GZip,                 Validate)     Validation)
    │           Rate Limit)
    │
    ├──► Static File Server (dist/)
    ├──► Proxy Routes (map tiles, Amap)
    └──► MCP JSON-RPC Handler
```

### Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **Gin Router** | HTTP routing, middleware pipeline, static file serving |
| **Middleware** | Authentication, rate limiting, error recovery |
| **API Handlers** (`pkg/api/`) | Request parsing, response formatting, input validation |
| **Services** (`pkg/services/`) | Business logic, authorization checks, cross-cutting concerns |
| **DataStore** (`pkg/datastore/`) | Database abstraction, ORM queries, transactions |
| **Models** (`pkg/models/`) | Data transfer objects, request/response structures |
| **Core** (`pkg/core/`) | Interfaces, constants, context wrappers |
| **Settings** (`pkg/settings/`) | INI file parsing, environment variable override, config struct |

---

## Backend Architecture

### Entry Point

`main.go` registers CLI commands via `urfave/cli/v3`:

| Command | File | Description |
|---------|------|-------------|
| `server run` | `cmd/webserver.go` | Start the HTTP server |
| `database update` | `cmd/database.go` | Sync database schema |
| `userdata *` | `cmd/user_data.go` | User management (add, modify, delete, import, export) |
| `security gen-secret-key` | `cmd/security.go` | Generate a JWT secret key |
| `cron list / run` | `cmd/cron_jobs.go` | List or run cron jobs |
| `utility *` | `cmd/utility.go` | Parse request IDs, send test emails |

### Initialization Sequence

`cmd/initializer.go` runs on every startup:

1. Load configuration (INI + environment overrides)
2. Validate secret key (required for production mode)
3. Initialize subsystems in order:
   - DataStore (database connection + schema sync)
   - Logger (file or console output)
   - Storage (local filesystem, MinIO, or WebDAV)
   - LLM provider (if AI recognition enabled)
   - UUID generator
   - Duplicate checker
   - Avatar provider
   - Mail sender
   - Exchange rates data source
4. Ensure default admin user exists (first run only)

### API Layer Pattern

Every API handler follows the same pattern:

```go
func (a *MyApi) MyHandler(c *core.WebContext) (any, *errs.Error) {
    var request models.MyRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
    }
    result, err := a.myService.DoSomething(c, &request)
    if err != nil {
        return nil, err
    }
    return result, nil
}
```

Route binding in `cmd/webserver.go` wraps handlers:

| Wrapper | Purpose |
|---------|---------|
| `bindApi()` | Standard JSON API — returns `{success, result}` or `{success, errorCode, errorMessage}` |
| `bindApiWithTokenUpdate()` | Standard JSON API plus Amap proxy auth-cookie handling |
| `bindJSONRPCApi()` | JSON-RPC 2.0 wrapper (used for MCP) |
| `bindCachedJs()` | Cached JavaScript response (server_settings.js) |
| `bindCsv() / bindTsv() / bindZip()` | File download responses |
| `bindImage()` | Image serving |
| `bindProxy()` | Reverse proxy passthrough |

### Middleware Stack

Applied in order per request:

1. **Recovery** — Catches panics, returns 500
2. **GZip** — Optional response compression
3. **JWT Authorization** — Token validation, user context injection
4. **API Token IP Restriction** — Validates API token source IP
5. **Rate Limiter** — Per-IP and per-user failure rate limiting

### File Format Converters

`pkg/converters/` contains parsers for every supported import format:

| Package | Format |
|---------|--------|
| `csv` | Generic CSV/TSV/SSV |
| `excel` | XLSX, XLS |
| `ofx` | OFX, QFX (Open Financial Exchange) |
| `qif` | QIF (Quicken Interchange Format) |
| `iif` | IIF (Intuit Interchange Format) |
| `camt` | Camt.052, Camt.053 (ISO 20022 XML) |
| `mt` | MT940 (SWIFT) |
| `gnucash` | GnuCash XML |
| `fireflyIII` | Firefly III CSV |
| `beancount` | Beancount plain text |
| `alipay` | Alipay CSV |
| `wechat` | WeChat Pay CSV/XLSX |
| `jdcom` | JD.com Finance |
| `feidee` | Feidee MyMoney |
| `sgml` | SGML-based format parsing |
| `datatable` | Internal tabular data type |
| `custom` | User-defined column mappings |
| `converter` | Converter registry and factory |
| `default` | Default Honey Ledger CSV/TSV format |

---

## Frontend Architecture

### Build System

| Tool | Configuration | Purpose |
|------|---------------|---------|
| Vite 7 | `vite.config.ts` | Dev server, production bundling, HMR |
| TypeScript 5.9 | `tsconfig.json` | Type checking |
| vue-tsc | Via vite-plugin-checker | Vue SFC type checking |
| PostCSS | `postcss.config.js` | CSS post-processing (autoprefixer, etc.) |

### Application Bootstrap

`src/desktop-main.ts` initializes the application:

1. Create Vue app instance
2. Install plugins: Vuetify, Pinia, Router, vue-i18n, vue-echarts
3. Register global components
4. Mount to `#app` in `desktop.html`

### Component Organization

```
src/components/
  desktop/           # Desktop-specific components
    SidebarNav.vue
    HeaderToolbar.vue
    ...
  common/            # Shared between views
    cards/           # Reusable card components
    ...
```

### View Organization

```
src/views/
  desktop/           # Page-level components (routed)
    LoginPage.vue
    MainLayout.vue   # Main app shell (sidebar + content area)
    HomePage.vue
    accounts/        # Account management pages
    transactions/    # Transaction pages + import dialog
    categories/      # Category management
    tags/            # Tag management
    templates/       # Template management
    statistics/      # Statistics & charts
    insights/        # Insights explorer
    exchangerates/   # Exchange rate management
    app/             # Application settings pages
    user/            # User profile & security settings
  base/              # Shared page logic
    LoginPageBase.ts
    transactions/    # Transaction composables
    settings/        # Settings composables
```

### State Management Architecture

```
Pinia Stores
├── index.ts          (auth, notifications, app state)
├── setting.ts        (app settings, locale, cloud sync)
├── user.ts           (current user profile)
├── transaction.ts    (transaction list, pagination)
├── account.ts        (account list, balances)
├── overview.ts       (home page summary data)
├── statistics.ts     (chart data, filters)
├── exchangeRates.ts  (rate cache, auto-update)
├── explorer.ts       (insights query state)
├── token.ts          (session/API token management)
└── system.ts         (version checking)
```

### Routing

`src/router/desktop.ts` defines all routes with:
- **Navigation guards** for authentication and app-lock enforcement
- **Static page imports** (Vite splits shared dependencies into vendor chunks)
- **Hash mode** routing (`/#/path`)

### API Client

`src/lib/services.ts` wraps axios with:
- Base URL configuration
- Auth token injection (via interceptor)
- Token refresh logic (when token nears expiration)
- Error response handling
- File upload support

The machine-readable API contract lives in `docs/openapi.json` and the shared TypeScript types are generated into `src/models/generated/api.ts`. Regenerate both with `npm run openapi:generate`; `npm run openapi:check` is enforced in CI.

### Chunk Splitting Strategy

Vite is configured to split output into optimized chunks:

| Chunk | Contents |
|-------|----------|
| `vendor-desktop` | Vuetify, vue-router, vue-echarts |
| `vendor-common` | Vue, Pinia, axios, crypto-js, vue-i18n |
| `common` | Shared application code |
| `desktop` | Desktop-specific code |
| `locales` | All translation bundles |
| `moment` | Moment.js + timezone data |
| `leaflet` | Map rendering library |

---

## Electron Desktop Shell

### Architecture

```
Electron Main Process (electron/main.js)
├── Spawn Go backend as child process
├── Register app:// protocol scheme
├── Create BrowserWindow
├── Proxy API paths to Go backend
└── Serve static files from dist/
```

### Protocol Scheme

Electron registers a custom `app://honey-ledger` protocol to serve frontend assets without a local HTTP server for the UI layer. API calls are intercepted and proxied to the Go backend.

### Backend Lifecycle

1. Acquire a single-instance lock (a second launch focuses the existing window)
2. Find a free TCP port
3. Seed `conf/honey-ledger.ini` and `templates/` into Electron's per-user data directory on first launch
4. Spawn the Go binary with `HONEY_WORK_DIR=<userData>`, `HONEY_SERVER_HTTP_ADDR=127.0.0.1`, and `HONEY_SERVER_HTTP_PORT=<port>`
5. Wait for the backend TCP port to accept connections
6. Load the frontend in a BrowserWindow
7. On app quit, send SIGTERM to the Go process (SIGKILL after 3s timeout)

### Security Sandbox

- `contextIsolation: true`
- `nodeIntegration: false`
- `sandbox: true`
- Content Security Policy enforced
- All permission requests denied
- External links open in system browser

### Build Targets

| Platform | Format |
|----------|--------|
| Linux | AppImage, .deb |
| macOS | .dmg |
| Windows | NSIS installer |

---

## Data Flow

### Transaction Creation

```
User fills form → Vue component
    │
    ▼
Pinia store action
    │
    ▼
services.ts → POST /api/v1/transactions/add.json
    │
    ▼
API Handler (transactions.go) → validates input
    │
    ▼
Service (transactions.go) → business logic, duplicate check
    │
    ▼
DataStore → INSERT into transaction table
    │
    ▼
Response → store update → UI refresh
```

### Data Import

```
User uploads file → import dialog
    │
    ▼
POST /api/v1/transactions/parse_import.json
    │
    ▼
Converter (auto-detected by file type) → parsed transactions
    │
    ▼
User reviews, adjusts mapping, batch replaces
    │
    ▼
POST /api/v1/transactions/import.json
    │
    ▼
JSON polling (GET /api/v1/transactions/import/process.json) → progress updates
    │
    ▼
Transactions inserted → import complete
```

---

## Database Schema

### Entity Relationships

```
User ─────┬──── Account ──── Transaction
          │         │              │
          │         │              ├── TransactionCategory
          │         │              ├── TransactionTag (via TagIndex)
          │         │              └── TransactionPictureInfo
          │         │
          │         └── SubAccount (self-referencing)
          │
          ├──── TwoFactor + RecoveryCodes
          ├──── TokenRecord (sessions + API tokens)
          ├──── UserCustomExchangeRate
          ├──── UserApplicationCloudSetting
          ├──── UserExternalAuth (OAuth2)
          ├──── TransactionTemplate
          ├──── TransactionTagGroup ──── TransactionTag
          └──── InsightsExplorer
```

### Key Design Decisions

- **Soft deletes**: Records are marked as deleted rather than physically removed.
- **Ordered entities**: Accounts, categories, tags, and templates have a `displayOrder` field for user-controlled sorting.
- **Multi-currency**: Transactions store amounts in minor currency units (cents) as integers.
- **Tag flexibility**: Tags are many-to-many via a join table (`TransactionTagIndex`).

---

## Authentication & Authorization

### Token Types

| Type | Lifetime | Purpose |
|------|----------|---------|
| Session token | 30 days (configurable) | Desktop app login sessions |
| API token | No expiration | Programmatic API access |
| MCP token | No expiration | AI agent access |
| Temporary token | 5 minutes | 2FA intermediate step |
| Email verify token | 1 hour | Email address verification |
| Password reset token | 1 hour | Password reset flow |

### JWT Claims

Each token is a JWT signed with a random per-token secret stored in the `token_record` table. Claims include user ID, token ID, token type, and issuance time. Validation checks the database record, so revoking a token takes effect immediately.

### OAuth2 Flow

1. Frontend redirects to `/oauth2/login`
2. Backend redirects to provider's auth URL
3. Provider redirects back to `/oauth2/callback`
4. Backend exchanges code for tokens, matches user by email/username
5. Backend issues session JWT and redirects to frontend

### Feature Restrictions

Per-user bitmask (`UserFeatureRestrictions`) controls access to 17 features. Administrators can restrict specific users from operations like importing data, clearing data, or using MCP.

---

## File Storage

### Storage Backends

| Backend | Use Case | Configuration |
|---------|----------|---------------|
| **Local filesystem** | Default, single-server | `storage/` directory |
| **MinIO** | S3-compatible object storage | Endpoint, bucket, access keys |
| **WebDAV** | Network file storage | Endpoint, credentials |

### Stored Files

- **Avatars**: `storage/avatar/` — User profile pictures
- **Transaction pictures**: `storage/transaction/` — Receipt images

---

## External Integrations

### Exchange Rate Providers

16 central bank and financial data sources. Rates are fetched periodically and cached with configurable TTL.

### LLM Providers

8 supported providers for AI receipt recognition. The backend sends receipt images to the configured LLM API and parses structured transaction data from the response using a prompt template (`templates/prompt/receipt_image_recognition.tmpl`).

### Map Providers

13 tile-based map providers. The backend can optionally proxy tile requests to add server-side API keys and work around CORS restrictions.

### Email Templates

`templates/email/` contains Go templates for:
- `verify_email.tmpl` — Email address verification
- `password_reset.tmpl` — Password reset link

---

## Build & Deployment Pipeline

### Build Stages

```
Source
  │
  ├── ./scripts/build.sh backend
  │       │
  │       ├── go mod download
  │       ├── go vet (lint)
  │       ├── go test
  │       └── go build (-tags timetzdata, CGO_ENABLED=1)
  │
  ├── ./scripts/build.sh frontend
  │       │
  │       ├── npm ci
  │       ├── npm run lint
  │       ├── npm test
  │       └── npm run build (vite)
  │
  └── ./scripts/build.sh package
          │
          └── tar.gz archive:
              ├── honey-ledger (binary)
              ├── public/ (frontend build output)
              ├── conf/ (config template)
              ├── templates/ (email + prompt)
              └── LICENSE
```

---

## Security Model

### Defense in Depth

| Layer | Mechanism |
|-------|-----------|
| Transport | Loopback HTTP only (`127.0.0.1` or `localhost`) |
| Authentication | JWT tokens, OAuth2/OIDC, 2FA (TOTP) |
| Authorization | Per-user feature restriction bitmask |
| Rate limiting | Per-IP and per-user failure throttling |
| Input validation | Gin binding validators + custom validators |
| Secret management | Env var + file-based secret injection, `0600` config on first write |
| Electron sandbox | contextIsolation, no nodeIntegration, CSP |
| Request deduplication | In-memory duplicate submission checker |
| Cryptographic | Per-token HMAC-SHA256 secrets, PBKDF2-HMAC-SHA256 password hashing (600,000 iterations), AES-GCM 2FA secrets |
| Log hygiene | `token`, `code`, `state`, `password` query parameters are redacted |
| File permissions | Process umask `0077`; config, database (including WAL/SHM), logs, and stored objects are `0600`; data directories are `0750` |

### Secret Key

The `secret_key` in `[security]` is used to encrypt two-factor secrets and derive server/instance identifiers:
- It is **not** the JWT signing key; each token has its own random secret in the database
- If empty, a random key is generated on first start and written back to the config file with `0600` permissions
- Generate a value explicitly with: `./honey-ledger security gen-secret-key`
- Changing it makes existing two-factor secrets unreadable (users must re-enroll 2FA)
