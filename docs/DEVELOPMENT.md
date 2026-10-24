# Development Guide

Instructions for setting up a development environment, building, testing, and debugging Ratel Ledger.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Repository Setup](#repository-setup)
- [Development Workflow](#development-workflow)
- [Backend Development](#backend-development)
- [Frontend Development](#frontend-development)
- [Running the Full Stack](#running-the-full-stack)
- [Testing](#testing)
- [Linting](#linting)
- [Build Scripts](#build-scripts)
- [Database Migrations](#database-migrations)
- [Adding a New API Endpoint](#adding-a-new-api-endpoint)
- [Adding a New Frontend Page](#adding-a-new-frontend-page)
- [Localization](#localization)
- [Debugging](#debugging)

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25+ | Backend compilation |
| Node.js | 24+ | Frontend build |
| npm | 10+ | Package management |
| GCC | Any recent | SQLite CGO compilation |
| Git | Any recent | Version control |

### Install on Fedora / RHEL

```bash
sudo dnf install golang nodejs npm gcc git
```

### Install on Ubuntu / Debian

```bash
sudo apt install golang-go nodejs npm gcc git
```

### Install on macOS

```bash
brew install go node gcc git
```

---

## Repository Setup

```bash
git clone https://github.com/James-HoneyBadger/ratel-ledger.git
cd ratel-ledger
```

### Backend Dependencies

```bash
go get .
```

### Frontend Dependencies

```bash
npm install
```

---

## Development Workflow

The typical development loop:

1. **Start the backend** in one terminal:
   ```bash
   go build -o ratel-ledger . && ./ratel-ledger server run
   ```

2. **Start the frontend dev server** in another terminal:
   ```bash
   npm run serve
   ```

3. The Vite dev server runs at `http://localhost:8081` and proxies API calls to the backend on port 8080. For full desktop testing, run the Electron app instead (see [Electron Guide](ELECTRON.md)).

4. Edit frontend code — Vite provides hot module replacement (HMR) for instant updates.

5. Edit backend code — rebuild and restart the binary.

---

## Backend Development

### Project Structure

```
main.go          # Main entry point, CLI command registration
cmd/
  base.go           # CLI helper utilities
  webserver.go      # HTTP server setup, routes, middleware
  initializer.go    # System initialization sequence
  database.go       # Database CLI commands
  user_data.go      # User management CLI commands
  security.go       # Security utility CLI commands
  cron_jobs.go      # Cron job CLI commands
  utility.go        # Miscellaneous CLI utilities
pkg/
  api/              # HTTP request handlers (controllers)
  auth/             # OAuth2 provider implementations
  avatars/          # Avatar provider implementations
  cli/              # CLI helper types
  converters/       # File format converters (import/export)
  core/             # Core types, interfaces, constants
  cron/             # Cron job scheduler
  datastore/        # Database abstraction (xorm)
  duplicatechecker/ # Request deduplication
  errs/             # Error definitions
  exchangerates/    # Exchange rate data sources
  httpclient/       # HTTP client utilities
  llm/              # LLM provider integrations
  locales/          # Server-side localization
  log/              # Logging infrastructure
  mail/             # Email sending
  mcp/              # Model Context Protocol handlers
  middlewares/      # HTTP middleware (auth, CORS, etc.)
  models/           # Data models and DTOs
  requestid/        # Request ID generation
  services/         # Business logic layer
  settings/         # Configuration loading
  storage/          # File storage backends
  templates/        # Template rendering
  utils/            # Shared utilities
  uuid/             # UUID generation
  validators/       # Custom validation rules
```

### Key Patterns

**Request flow:** HTTP Request → Gin Router → Middleware → API Handler → Service → DataStore → Database

**Adding an API handler:**

1. Create or modify a handler in `pkg/api/`.
2. Add the route in `cmd/webserver.go`.
3. Add business logic in `pkg/services/`.
4. Define models in `pkg/models/`.
5. Define errors in `pkg/errs/`.

### Configuration Loading

Configuration is loaded in `pkg/settings/setting.go`:
- Reads from INI file (`conf/ratel-ledger.ini`)
- Environment variables take precedence (`RATEL_{SECTION}_{KEY}`)
- File-path variables for secrets (`RATELCFP_{SECTION}_{KEY}`)

### Database Access

The project uses xorm as the ORM. Database operations are in `pkg/datastore/`. Table definitions are synced automatically on startup.

**Supported databases:** SQLite 3, MySQL, PostgreSQL.

---

## Frontend Development

### Project Structure

```
src/
  desktop-main.ts    # Application bootstrap
  desktop.html       # HTML entry point
  DesktopApp.vue     # Root Vue component
  global.d.ts        # Global TypeScript declarations
  vue-shim.d.ts      # Vue SFC type shim
  components/        # Reusable Vue components
    desktop/         # Desktop-specific components
    common/          # Shared components
  consts/            # Constants (API paths, currencies, etc.)
  core/              # Core types (settings, themes, etc.)
  directives/        # Vue custom directives
  icons/             # Icon definitions
  lib/               # Utility libraries
    services.ts      # API client (axios-based)
    ui/              # UI helpers
    map/             # Map provider integration
  locales/           # Translation files (19 languages)
  models/            # TypeScript data models
  router/            # Vue Router configuration
    desktop.ts       # Desktop route definitions
  stores/            # Pinia state stores
  styles/            # SCSS stylesheets
  views/             # Page components
    desktop/         # Desktop pages
    base/            # Shared page logic (composables)
```

### Key Technologies

| Library | Purpose |
|---------|---------|
| Vue 3 | UI framework (Composition API) |
| Vuetify 3 | Material Design component library |
| Pinia | State management |
| Vue Router | Client-side routing |
| vue-i18n | Internationalization |
| axios | HTTP client |
| ECharts | Chart visualization |
| Leaflet | Map rendering |
| Moment.js | Date/time handling |
| CryptoJS | Client-side encryption |

### State Management

Pinia stores are in `src/stores/`:
- `index.ts` — Root store (auth state, notifications)
- `setting.ts` — Application settings
- `user.ts` — Current user profile
- `transaction.ts` — Transaction list state
- `exchangeRates.ts` — Exchange rate cache
- `account.ts` — Account list
- `overview.ts` — Home page data
- `statistics.ts` — Statistics data
- `explorer.ts` — Insights explorer state
- `system.ts` — Version checking
- `token.ts` — Token management

### API Client

All API calls are centralized in `src/lib/services.ts`. The axios instance includes:
- Automatic `Authorization` header injection
- Token refresh interceptor
- Error handling interceptor

### Routing

Routes are defined in `src/router/desktop.ts` with authentication guards. The router uses hash mode (`/#/`).

---

## Running the Full Stack

### Option 1: One Command

From the repository root:

```bash
./run.sh
```

This builds the backend, builds the frontend when needed, installs missing dependencies, and launches the Electron app. Other modes:

```bash
./run.sh --rebuild    # force a frontend rebuild
./run.sh --server     # backend only
./run.sh --build-only # build without launching
./run.sh --help
```

### Option 2: Separate Processes

```bash
# Terminal 1 — Backend
go build -o ratel-ledger . && ./ratel-ledger server run

# Terminal 2 — Frontend (dev server with HMR)
npm run serve
```

Access at `http://localhost:8081` for frontend development (or run the Electron app for full desktop testing).

### Option 3: Production Build

```bash
# Build frontend
npm run build

# Build and run backend
go build -o ratel-ledger . && ./ratel-ledger server run
```

Access via the Electron app, or at `http://localhost:8080` for backend-only testing.

### Option 4: Full Package

```bash
./scripts/build.sh package
```

---

## API Contract Generation

The OpenAPI document and frontend API types are generated from the Go models:

```bash
go run ./cmd/openapi          # regenerate docs/openapi.json and src/models/generated/api.ts
go run ./cmd/openapi --check  # verify generated files are current (used by CI)
```

Equivalent npm scripts are available as `npm run openapi:generate` and `npm run openapi:check`.

---

## Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/converters/ofx/

# Run a specific test
go test ./pkg/models/ -run TestUserCanEditTransaction

# Verbose output
go test -v ./pkg/services/

# With race detection
go test -race ./...
```

### Frontend Tests

```bash
# Run all tests
npm test

# Watch mode
npx jest --watch
```

### Test Configuration

- Backend: Standard Go test framework
- Frontend: Jest with `tsconfig.jest.json` and `jest.config.ts`

---

## Linting

### Backend

```bash
go vet ./...
```

### Frontend

```bash
npm run lint
```

ESLint configuration is in `eslint.config.mjs`.

### TypeScript Type Checking

```bash
npx vue-tsc --noEmit
```

---

## Build Scripts

The `scripts/build.sh` script automates the build process:

| Command | Description |
|---------|-------------|
| `./scripts/build.sh backend` | Build Go binary only |
| `./scripts/build.sh frontend` | Build frontend assets only |
| `./scripts/build.sh package` | Build both and create a release archive |

### Build Options

| Flag | Description |
|------|-------------|
| `--release` / `-r` | Set version string (e.g., `-r 1.4.1`) |
| `--output` / `-o` | Output directory |
| `--no-lint` | Skip lint checks |
| `--no-test` | Skip tests |

### Manual Backend Build

```bash
CGO_ENABLED=1 go build \
  -tags timetzdata \
  -ldflags "-X main.Version=dev \
            -X main.CommitHash=$(git rev-parse --short HEAD) \
            -X main.BuildUnixTime=$(date +%s)" \
  -o ratel-ledger .
```

### Manual Frontend Build

```bash
npm run build
```

---

## Database Migrations

Database schema updates are handled automatically by xorm on startup when `auto_update_database = true`.

### Manual Update

```bash
./ratel-ledger database update
```

### Tables

| Table | Description |
|-------|-------------|
| `user` | User accounts |
| `two_factor` | 2FA configuration |
| `two_factor_recovery_code` | 2FA recovery codes |
| `token_record` | Session and API tokens |
| `account` | Financial accounts |
| `transaction` | Financial transactions |
| `transaction_category` | Transaction categories |
| `transaction_tag_group` | Tag groups |
| `transaction_tag` | Tags |
| `transaction_tag_index` | Tag-transaction associations |
| `transaction_template` | Transaction templates |
| `transaction_picture_info` | Transaction picture metadata |
| `user_custom_exchange_rate` | Custom exchange rates |
| `user_application_cloud_setting` | Cloud-synced settings |
| `user_external_auth` | OAuth2 linked accounts |
| `insights_explorer` | Saved insight queries |

---

## Adding a New API Endpoint

1. **Define the model** in `pkg/models/`:
   ```go
   type MyRequest struct {
       Name string `json:"name" binding:"required,notBlank"`
   }
   type MyResponse struct {
       ID   int64  `json:"id"`
       Name string `json:"name"`
   }
   ```

2. **Define errors** in `pkg/errs/` if needed:
   ```go
   var ErrMyThingNotFound = NewNormalError(NormalSubcategoryDefault, 100, http.StatusNotFound, "thing not found")
   ```

3. **Add service logic** in `pkg/services/`:
   ```go
   type MyService struct {
       ServiceUsingDB
       ServiceUsingConfig
   }
   func (s *MyService) DoThing(c core.Context, req *models.MyRequest) (*models.MyResponse, error) {
       // business logic
   }
   ```

4. **Add the API handler** in `pkg/api/`:
   ```go
   type MyApi struct {
       ApiUsingConfig
       myThings *services.MyService
   }
   func (a *MyApi) MyHandler(c *core.WebContext) (any, *errs.Error) {
       // parse request, call service, return response
   }
   ```

5. **Register the route** in `cmd/webserver.go`:
   ```go
   v1Route.GET("/my/thing.json", bindApi(api.MyThing.MyHandler))
   ```

6. **Add frontend API call** in `src/lib/services.ts` if needed.

---

## Adding a New Frontend Page

1. **Create the view** in `src/views/desktop/`:
   ```vue
   <template>
     <v-container>
       <h1>{{ tt('My Page Title') }}</h1>
     </v-container>
   </template>
   <script setup lang="ts">
   import { useI18n } from '@/locales/helpers.ts';
   const { tt } = useI18n();
   </script>
   ```

2. **Add the route** in `src/router/desktop.ts`:
   ```typescript
   {
       path: '/my-page',
       component: () => import('@/views/desktop/MyPage.vue'),
       meta: { requireAuth: true }
   }
   ```

3. **Add translations** to all locale files in `src/locales/`.

---

## Localization

### Backend Locales

Server-side translations are in `pkg/locales/`. These are used for error messages and notification content.

### Frontend Locales

Translation files are in `src/locales/` — one JSON file per language (19 languages total):

```
en.json, de.json, es.json, fr.json, it.json, ja.json, kn.json, ko.json,
nl.json, pt_BR.json, ru.json, sl.json, ta.json, th.json, tr.json,
uk.json, vi.json, zh_Hans.json, zh_Hant.json
```

### Adding a Translation Key

1. Add the key and English text to `src/locales/en.json`.
2. Add translations to all other locale files (or leave them empty for fallback to English).
3. Use `tt('key.path')` in Vue templates or `useI18n()` in script setup.

---

## Debugging

### Backend Debugging

Set development mode for verbose logging:

```ini
[global]
mode = development

[log]
level = debug

[database]
log_query = true
```

Or via environment:

```bash
RATEL_GLOBAL_MODE=development RATEL_LOG_LEVEL=debug ./ratel-ledger server run
```

### Frontend Debugging

Use Electron DevTools (View → Toggle Developer Tools). The Vite dev server provides source maps for TypeScript debugging during development.

Vue DevTools extension is recommended for inspecting component state, Pinia stores, and routing.

### Request ID Parsing

Every request gets a unique ID. Parse it for debugging:

```bash
./ratel-ledger utility parse-default-request-id --id <request-id>
```

This decodes the server ID, instance, timestamp, sequence, and client IP from the request ID.

### Test Email

Verify SMTP configuration:

```bash
./ratel-ledger utility send-test-mail --to user@example.com
```
