# API Reference

Complete REST API documentation for Honey Ledger.

Most API endpoints return JSON. Export and download endpoints return CSV, TSV, or ZIP files instead. Authenticated endpoints require a `Bearer` token in the `Authorization` header or a valid session cookie.

The machine-readable OpenAPI document is generated at [`openapi.json`](./openapi.json). Frontend request/response types are generated into [`src/models/generated/api.ts`](../src/models/generated/api.ts). Regenerate both with `npm run openapi:generate`; CI verifies them with `npm run openapi:check`. The generated document currently covers the authentication, backup, and token endpoints (plus shared statistics models); this page remains the full endpoint reference.

---

## Table of Contents

- [Base URL](#base-url)
- [Authentication](#authentication)
- [Response Format](#response-format)
- [Error Codes](#error-codes)
- [Public Endpoints](#public-endpoints)
- [Authorization](#authorization)
- [Tokens](#tokens)
- [Users](#users)
- [Cloud Settings](#cloud-settings)
- [Two-Factor Authentication](#two-factor-authentication)
- [Accounts](#accounts)
- [Transactions](#transactions)
- [Transaction Pictures](#transaction-pictures)
- [Transaction Categories](#transaction-categories)
- [Transaction Tag Groups](#transaction-tag-groups)
- [Transaction Tags](#transaction-tags)
- [Transaction Templates](#transaction-templates)
- [Insights Explorers](#insights-explorers)
- [AI Assistant](#ai-assistant)
- [Shared Expenses](#shared-expenses)
- [Exchange Rates](#exchange-rates)
- [Data Management](#data-management)
- [Backups](#backups)
- [LLM](#llm)
- [System](#system)
- [Health Check](#health-check)
- [MCP (Model Context Protocol)](#mcp-model-context-protocol)
- [OAuth2 Endpoints](#oauth2-endpoints)
- [Proxy Endpoints](#proxy-endpoints)
- [Rate Limiting](#rate-limiting)

---

## Base URL

In the desktop app, the backend API is available at:

```
http://127.0.0.1:<port>/api
```

The port is assigned automatically by Electron. For development or API scripts, the default port is `8080`.

Versioned endpoints use the `/api/v1/` prefix.

---

## Authentication

### Session Authentication

Obtain a token via the `/api/authorize.json` endpoint and include it in subsequent requests:

```
Authorization: Bearer <token>
```

### API Token Authentication

Generate an API token from the user interface (Settings → Security) and use it the same way:

```
Authorization: Bearer <api-token>
```

API tokens may be restricted to specific IP addresses by the administrator.

### Token Refresh

Tokens are automatically refreshed when near expiration. The response header `X-New-Token` contains the refreshed token when applicable.

---

## Response Format

### Success Response

```json
{
    "success": true,
    "result": { ... }
}
```

### Error Response

```json
{
    "success": false,
    "errorCode": 10001,
    "errorMessage": "description of the error"
}
```

---

## Error Codes

Errors are returned with appropriate HTTP status codes and a numeric `errorCode` for programmatic handling.

| HTTP Status | Meaning |
|-------------|---------|
| 200 | Success |
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — missing or invalid token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

---

## Public Endpoints

These endpoints do not require authentication.

### Register

```
POST /api/register.json
```

Create a new user account (if registration is enabled).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | string | Yes | Username (alphanumeric, 2-32 characters) |
| `email` | string | Yes | Email address |
| `nickname` | string | Yes | Display name |
| `password` | string | Yes | Password |
| `defaultCurrency` | string | Yes | Default currency code (e.g., `USD`) |

### Server Settings

```
GET /server_settings.js
```

Returns server configuration as JavaScript (cached). Used by the frontend to detect enabled features.

---

## Authorization

### Login

```
POST /api/authorize.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `loginName` | string | Yes | Username or email |
| `password` | string | Yes | Password |

**Response:** Token object with `token`, `need2FA`, `user`, and optional `applicationCloudSettings` / `notificationContent`. If 2FA is enabled, returns a `need2FA: true` flag and a temporary `token` to use for the 2FA step.

### Desktop Auto-Login

```
POST /api/auto_login.json
```

No request body. Succeeds only when `enable_desktop_auto_login = true`, exactly one non-deleted user exists, that user has not enabled 2FA, and email verification is either disabled or already complete. Returns the same response shape as `POST /api/authorize.json`. The request must originate from a loopback address; this endpoint is intended for the local Electron shell.

### Two-Factor Login

```
POST /api/2fa/authorize.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `passcode` | string | Yes | 6-digit TOTP code |
| `token` | string | Yes | Temporary token from login response |

### Two-Factor Recovery

```
POST /api/2fa/recovery.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `recoveryCode` | string | Yes | One-time recovery code |
| `token` | string | Yes | Temporary token from login response |

### Logout

```
GET /api/logout.json
```

Revokes the current session token.

### Email Verification

```
POST /api/verify_email.json
```

```
POST /api/verify_email/resend.json
```

### Password Reset

```
POST /api/forget_password.json
POST /api/forget_password/reset.json
```

---

## Tokens

All token endpoints require authentication.

### List Tokens

```
GET /api/v1/tokens/list.json
```

Returns all active session and API tokens for the current user.

### Generate API Token

```
POST /api/v1/tokens/generate/api.json
```

Requires `enable_api_token = true` in configuration.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `password` | string | Yes | Account password |
| `expiresInSeconds` | int | No | Token lifetime in seconds; `0` uses the server default |

### Generate MCP Token

```
POST /api/v1/tokens/generate/mcp.json
```

Requires `enable_mcp = true` in configuration.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `password` | string | Yes | Account password |
| `expiresInSeconds` | int | No | Token lifetime in seconds; `0` uses the server default |

### Revoke Token

```
POST /api/v1/tokens/revoke.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tokenId` | string | Yes | Token ID to revoke |

### Revoke All Tokens

```
POST /api/v1/tokens/revoke_all.json
```

Revokes all tokens except the current session.

### Refresh Token

```
POST /api/v1/tokens/refresh.json
```

---

## Users

### Get Profile

```
GET /api/v1/users/profile/get.json
```

### Update Profile

```
POST /api/v1/users/profile/update.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `email` | string | No | New email address |
| `nickname` | string | No | New display name |
| `password` | string | No | New password |
| `oldPassword` | string | Conditional | Current password (required when changing the password) |
| `defaultCurrency` | string | No | Default currency code |
| `language` | string | No | Interface language |
| `defaultAccountId` | int | No | Default account for transactions |
| `transactionEditScope` | int | No | Transaction edit permission scope |
| `expenseAmountColor` | string | No | Hex color for expense amounts |
| `incomeAmountColor` | string | No | Hex color for income amounts |

### Update Avatar

```
POST /api/v1/users/avatar/update.json
```

Multipart form upload with an image file.

### Remove Avatar

```
POST /api/v1/users/avatar/remove.json
```

### Verify Email

```
POST /api/v1/users/verify_email.json
```

### External Auth Info

```
GET /api/v1/users/external_auth/list.json
POST /api/v1/users/external_auth/unlink.json
```

---

## Cloud Settings

### Get Cloud Settings

```
GET /api/v1/users/settings/cloud/get.json
```

### Update Cloud Settings

```
POST /api/v1/users/settings/cloud/update.json
```

### Disable Cloud Sync

```
POST /api/v1/users/settings/cloud/disable.json
```

---

## Two-Factor Authentication

### Get Status

```
GET /api/v1/users/2fa/status.json
```

### Enable — Request

```
POST /api/v1/users/2fa/enable/request.json
```

Returns a TOTP secret and provisioning URI for QR code generation.

### Enable — Confirm

```
POST /api/v1/users/2fa/enable/confirm.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `passcode` | string | Yes | 6-digit TOTP code to verify setup |

**Response:** Recovery codes (store securely).

### Disable

```
POST /api/v1/users/2fa/disable.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `password` | string | Yes | Account password |

### Regenerate Recovery Codes

```
POST /api/v1/users/2fa/recovery/regenerate.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `password` | string | Yes | Account password |

---

## Accounts

### List Accounts

```
GET /api/v1/accounts/list.json
```

### Get Account

```
GET /api/v1/accounts/get.json?id=<accountId>
```

### Add Account

```
POST /api/v1/accounts/add.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Account name |
| `category` | int | Yes | Account type/category |
| `type` | int | Yes | Account sub-type |
| `currency` | string | Yes | Currency code |
| `balance` | int | No | Initial balance (in minor units) |
| `balanceTime` | int | No | Balance timestamp |
| `comment` | string | No | Description |
| `icon` | int | Yes | Icon identifier |
| `color` | string | Yes | Hex color code |
| `creditCardStatementDate` | int | No | Credit card statement day (0–28) |
| `subAccounts` | array | No | Nested sub-accounts |
| `clientSessionId` | string | No | Client-side deduplication ID |

### Modify Account

```
POST /api/v1/accounts/modify.json
```

### Hide Account

```
POST /api/v1/accounts/hide.json
```

### Move Account (Reorder)

```
POST /api/v1/accounts/move.json
```

### Delete Account

```
POST /api/v1/accounts/delete.json
```

### Delete Sub-Account

```
POST /api/v1/accounts/sub_account/delete.json
```

---

## Transactions

### Transaction Count

```
GET /api/v1/transactions/count.json
```

### List Transactions

```
GET /api/v1/transactions/list.json
```

Query parameters: `page`, `count` (1–50), `account_ids`, `category_ids`, `tag_filter`, `amount_filter`, `type`, `keyword`, `max_time`, `min_time`, `with_count`, `with_pictures`, `trim_account`, `trim_category`, `trim_tag`.

### List by Month

```
GET /api/v1/transactions/list/by_month.json
```

### List All

```
GET /api/v1/transactions/list/all.json
```

### Reconciliation Statements

```
GET /api/v1/transactions/reconciliation_statements.json
```

### Statistics

```
GET /api/v1/transactions/statistics.json
```

### Statistics Trends

```
GET /api/v1/transactions/statistics/trends.json
```

### Asset Trends

```
GET /api/v1/transactions/statistics/asset_trends.json
```

### Transaction Amounts

```
GET /api/v1/transactions/amounts.json
```

### Get Transaction

```
GET /api/v1/transactions/get.json?id=<transactionId>
```

### Add Transaction

```
POST /api/v1/transactions/add.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | int | Yes | 1=Expense, 2=Income, 3=Transfer |
| `categoryId` | int | Conditional | Category ID (not used for balance modification) |
| `time` | int | Yes | Unix timestamp |
| `utcOffset` | int | No | Client UTC offset in minutes (−720 to 840) |
| `sourceAccountId` | int | Yes | Source account ID |
| `destinationAccountId` | int | Conditional | Destination account (transfers only) |
| `sourceAmount` | int | Yes | Amount in minor currency units |
| `destinationAmount` | int | Conditional | Destination amount (for cross-currency transfers) |
| `hideAmount` | bool | No | Hide the amount in the UI |
| `tagIds` | array<string> | No | Tag IDs |
| `pictureIds` | array<string> | No | Transaction picture IDs |
| `comment` | string | No | Description |
| `geoLocation` | object | No | `{ latitude, longitude }` |
| `clientSessionId` | string | No | Client-side deduplication ID |

### Modify Transaction

```
POST /api/v1/transactions/modify.json
```

### Move All Transactions

```
POST /api/v1/transactions/move/all.json
```

Move all transactions from one account to another.

### Delete Transaction

```
POST /api/v1/transactions/delete.json
```

### Parse Custom File

```
POST /api/v1/transactions/parse_custom_file.json
```

Upload and parse a custom-format file for import preview.

### Parse Import

```
POST /api/v1/transactions/parse_import.json
```

Upload and parse a standard-format file for import preview.

### Import Transactions

```
POST /api/v1/transactions/import.json
```

### Import Progress

```
GET /api/v1/transactions/import/process.json
```

Returns the current import progress as a JSON number from `0` to `100` (or `null` when no task is running). The client polls this endpoint while an import is in progress.

---

## Transaction Pictures

### Upload Picture

```
POST /api/v1/transaction/pictures/upload.json
```

Multipart form upload.

### Remove Unused Pictures

```
POST /api/v1/transaction/pictures/remove_unused.json
```

---

## Transaction Categories

### List Categories

```
GET /api/v1/transaction/categories/list.json
```

### Get Category

```
GET /api/v1/transaction/categories/get.json?id=<categoryId>
```

### Add Category

```
POST /api/v1/transaction/categories/add.json
```

### Add Batch (Presets)

```
POST /api/v1/transaction/categories/add_batch.json
```

### Modify Category

```
POST /api/v1/transaction/categories/modify.json
```

### Hide Category

```
POST /api/v1/transaction/categories/hide.json
```

### Move Category (Reorder)

```
POST /api/v1/transaction/categories/move.json
```

### Delete Category

```
POST /api/v1/transaction/categories/delete.json
```

---

## Transaction Tag Groups

### List Tag Groups

```
GET /api/v1/transaction/tags/groups/list.json
```

### Get Tag Group

```
GET /api/v1/transaction/tags/groups/get.json?id=<groupId>
```

### Add Tag Group

```
POST /api/v1/transaction/tags/groups/add.json
```

### Modify Tag Group

```
POST /api/v1/transaction/tags/groups/modify.json
```

### Move Tag Group (Reorder)

```
POST /api/v1/transaction/tags/groups/move.json
```

### Delete Tag Group

```
POST /api/v1/transaction/tags/groups/delete.json
```

---

## Transaction Tags

### List Tags

```
GET /api/v1/transaction/tags/list.json
```

### Get Tag

```
GET /api/v1/transaction/tags/get.json?id=<tagId>
```

### Add Tag

```
POST /api/v1/transaction/tags/add.json
```

### Add Batch

```
POST /api/v1/transaction/tags/add_batch.json
```

### Modify Tag

```
POST /api/v1/transaction/tags/modify.json
```

### Hide Tag

```
POST /api/v1/transaction/tags/hide.json
```

### Move Tag (Reorder)

```
POST /api/v1/transaction/tags/move.json
```

### Delete Tag

```
POST /api/v1/transaction/tags/delete.json
```

---

## Transaction Templates

### List Templates

```
GET /api/v1/transaction/templates/list.json
```

### Get Template

```
GET /api/v1/transaction/templates/get.json?id=<templateId>
```

### Add Template

```
POST /api/v1/transaction/templates/add.json
```

### Modify Template

```
POST /api/v1/transaction/templates/modify.json
```

### Hide Template

```
POST /api/v1/transaction/templates/hide.json
```

### Move Template (Reorder)

```
POST /api/v1/transaction/templates/move.json
```

### Delete Template

```
POST /api/v1/transaction/templates/delete.json
```

---

## Insights Explorers

### List Explorers

```
GET /api/v1/insights/explorers/list.json
```

### Get Explorer

```
GET /api/v1/insights/explorers/get.json?id=<explorerId>
```

### Add Explorer

```
POST /api/v1/insights/explorers/add.json
```

### Modify Explorer

```
POST /api/v1/insights/explorers/modify.json
```

### Hide Explorer

```
POST /api/v1/insights/explorers/hide.json
```

### Move Explorer (Reorder)

```
POST /api/v1/insights/explorers/move.json
```

### Delete Explorer

```
POST /api/v1/insights/explorers/delete.json
```

---

## AI Assistant

Answers a question about the user's finances with the configured language model.

```
POST /api/v1/ai/ask.json
```

```json
{
  "question": "How much did I spend on dining this year?",
  "months": 12
}
```

The endpoint is disabled unless `[llm] ai_assistant = true` and a provider is
configured in `[llm_ai_assistant]`.

Only an aggregated summary leaves the machine, and every response describes what
was sent in its `privacyNote` field:

- income and expense totals per currency, plus the transaction count
- expense totals per category name, capped at the eight largest
- account balances labelled `Account 1`, `Account 2`, … with the account category
- shared expense balances labelled `Person 1`, `Person 2`, …

Account names, people names, transaction descriptions, tags and pictures are
never included. Amounts are integers in the smallest currency unit.

```json
{
  "answer": "You spent 75.00 USD on dining.",
  "provider": "openai_compatible",
  "model": "<model id>",
  "startTime": 1761864428,
  "endTime": 1793400428,
  "transactionCount": 2,
  "privacyNote": "Only the aggregated summary ..."
}
```

---

## Shared Expenses

Split expense transactions with other people and track what each of them owes.

### List Transaction Splits

```
GET /api/v1/transactions/splits/list.json?transactionId=<transactionId>
```

Returns the transaction amount, its currency, and the shares owed by other
people.

### Update Transaction Splits

```
POST /api/v1/transactions/splits/update.json
```

```json
{
  "transactionId": "3851278476451512320",
  "splits": [
    { "personName": "Priya", "amount": 3200 },
    { "personName": "Sam", "amount": 2000 }
  ]
}
```

Replaces all shares of the transaction; send an empty `splits` array to clear
them. Only expense transactions can be split, person names are trimmed and must
be unique per transaction, amounts must be positive, and the shares may not add
up to more than the transaction amount. Up to 20 people may share one
transaction.

### Shared Expense Summary

```
GET /api/v1/transactions/splits/summary.json
```

Returns one entry per person and currency with the outstanding total, the number
of shared expenses, and the time of the most recent one. Splits of deleted
transactions are excluded.

### Shared Expenses by Person

```
GET /api/v1/transactions/splits/person.json?personName=<personName>
```

Returns the individual expenses behind a person's total, newest first.

---

## Exchange Rates

### Get Latest Rates

```
GET /api/v1/exchange_rates/latest.json
```

### Update Custom Rate

```
POST /api/v1/exchange_rates/user_custom/update.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `currency` | string | Yes | Currency code |
| `rate` | string | Yes | Exchange rate value |

### Delete Custom Rate

```
POST /api/v1/exchange_rates/user_custom/delete.json
```

---

## Data Management

### Data Statistics

```
GET /api/v1/data/statistics.json
```

Returns counts of accounts, transactions, categories, tags, and templates.

### Clear All Data

```
POST /api/v1/data/clear/all.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `password` | string | Yes | Account password to confirm |

### Clear Transactions

```
POST /api/v1/data/clear/transactions.json
```

### Clear by Account

```
POST /api/v1/data/clear/by_account.json
```

### Export CSV

```
GET /api/v1/data/export.csv
```

### Export TSV

```
GET /api/v1/data/export.tsv
```

---

## Backups

Backups are supported for SQLite databases. The archive contains a consistent SQLite snapshot under `database/` and the local storage files under `storage/`. Configuration files are not included. Backup files are created with `0600` permissions.

### List Backups

```
GET /api/v1/backups/list.json
```

Returns an array of backup files with filename, size, and creation timestamp.

### Create Backup

```
POST /api/v1/backups/create.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `password` | string | Yes | Account password |

### Download Backup

```
GET /api/v1/backups/download.zip?file_name=<filename>
```

Returns the backup file as a ZIP download. Backups are SQLite-only.

### Restore Backup

```
POST /api/v1/backups/restore.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fileName` | string | Yes | Backup filename |
| `password` | string | Yes | Account password |

### Delete Backup

```
POST /api/v1/backups/delete.json
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fileName` | string | Yes | Backup filename |

---

## LLM

### Recognize Receipt Image

```
POST /api/v1/llm/transactions/recognize_receipt_image.json
```

Multipart form upload with a receipt image. Returns extracted transaction data (amount, date, description, category suggestion).

Requires `transaction_from_ai_image_recognition = true` in configuration.

---

## System

### Get Version

```
GET /api/v1/systems/version.json
```

Returns the application version, commit hash, and build timestamp.

---

## Health Check

```
GET /healthz.json
```

Returns `200 OK` if the server is running. No authentication required.

---

## MCP (Model Context Protocol)

The MCP endpoint enables AI agent access to financial data via JSON-RPC 2.0.

### Endpoint

```
POST /mcp
```

Requires `enable_mcp = true` and a valid MCP token in the `Authorization` header.

### Supported Methods

| Method | Description |
|--------|-------------|
| `initialize` | Initialize MCP session |
| `resources/list` | List available data resources |
| `resources/read` | Read a specific resource |
| `tools/list` | List available tools |
| `tools/call` | Execute a tool |
| `ping` | Health check |

The `query_transactions` tool defaults to 100 records per page and caps `count` at 1000.

### Authentication

Generate an MCP token via **Settings → Security → Generate MCP Token** and include it:

```
Authorization: Bearer <mcp-token>
```

### Example

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list",
    "params": {}
}
```

---

## OAuth2 Endpoints

### Login Redirect

```
GET /oauth2/login
```

Redirects to the configured OAuth2 provider's authorization page.

### Callback

```
GET /oauth2/callback
```

Handles the OAuth2 callback and exchanges the authorization code for tokens.

---

## Proxy Endpoints

### Map Tile Proxy

```
GET /proxy/map/tile/:tileImageUrl
```

Proxies map tile requests (when `enable_map_data_fetch_proxy = true`).

### Map Annotation Proxy

```
GET /proxy/map/annotation/:annotationUrl
```

### Amap Service Proxy

```
GET /_AMapService/*action
```

Proxies Amap (Gaode) API requests with server-side authentication.

---

## Rate Limiting

Authentication failures are rate-limited:

- **Per IP:** 5 failures per minute (default, configurable)
- **Per User:** 5 failures per minute (default, configurable)

Exceeding the limit returns `429 Too Many Requests`.
