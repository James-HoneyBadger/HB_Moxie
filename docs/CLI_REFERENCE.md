# CLI Reference

Complete reference for the Honey Ledger command-line interface.

---

## Table of Contents

- [Global Options](#global-options)
- [server run](#server-run)
- [database update](#database-update)
- [userdata Commands](#userdata-commands)
- [security gen-secret-key](#security-gen-secret-key)
- [cron Commands](#cron-commands)
- [utility Commands](#utility-commands)
- [honeytools API Script](#honeytools-api-script)

---

## Global Options

These options apply to all commands:

| Flag | Description | Default |
|------|-------------|---------|
| `--conf-path <path>` | Path to configuration file | `conf/honey-ledger.ini` |
| `--no-boot-log` | Suppress startup log messages | `false` |

Example:

```bash
./honey-ledger --conf-path /etc/honey-ledger/honey-ledger.ini server run
```

---

## server run

Start the HTTP server.

```bash
./honey-ledger server run
```

The backend binds to the address and port specified in the configuration file (`[server]` section). Default: `127.0.0.1:8080`. In the desktop app, the port is assigned automatically by Electron.

### Behavior

1. Loads configuration from INI file (with environment variable overrides).
2. Loads the secret key; if unset, generates one and persists it with `0600` permissions.
3. Initializes all subsystems (database, storage, mail, etc.).
4. Creates the default admin user if no users exist.
5. Starts the HTTP listener.
6. Runs background cron jobs (token cleanup, scheduled transactions).

---

## database update

Synchronize the database schema with the current application version.

```bash
./honey-ledger database update
```

This creates or updates all tables:

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
| `transaction_tag_index` | Tag-transaction join table |
| `transaction_template` | Transaction templates |
| `transaction_picture_info` | Picture metadata |
| `user_custom_exchange_rate` | Custom exchange rates |
| `user_application_cloud_setting` | Cloud-synced app settings |
| `user_external_auth` | OAuth2 linked accounts |
| `insights_explorer` | Saved insight queries |

This runs automatically on server startup when `auto_update_database = true` (the default).

---

## userdata Commands

User management commands for the CLI. All require `--username` / `-n` to identify the user.

### user-add

Create a new user.

```bash
./honey-ledger userdata user-add \
  --username alice \
  --email alice@example.com \
  --nickname "Alice Smith" \
  --password "securepassword" \
  --default-currency USD
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--username` | `-n` | Yes | Username |
| `--email` | `-m` | Yes | Email address |
| `--nickname` | `-i` | Yes | Display name |
| `--password` | `-p` | Yes | Password |
| `--default-currency` | `-c` | Yes | Default currency code |

### user-get

Display user information.

```bash
./honey-ledger userdata user-get --username alice
```

### user-modify-password

Change a user's password.

```bash
./honey-ledger userdata user-modify-password --username alice --password "newpassword"
```

### user-enable / user-disable

Enable or disable a user account.

```bash
./honey-ledger userdata user-enable --username alice
./honey-ledger userdata user-disable --username alice
```

### user-set-restrict-features

Set the complete feature restriction bitmask for a user.

```bash
./honey-ledger userdata user-set-restrict-features --username alice --feature-types "9,10,11"
```

### user-add-restrict-features

Add restrictions to a user's existing set.

```bash
./honey-ledger userdata user-add-restrict-features --username alice --feature-types "13"
```

### user-remove-restrict-features

Remove specific restrictions from a user.

```bash
./honey-ledger userdata user-remove-restrict-features --username alice --feature-types "9,10"
```

| Feature Type | Description |
|-------------|-------------|
| 1 | Update Password |
| 2 | Update Email |
| 3 | Update Profile Basic Info |
| 4 | Update Avatar |
| 5 | Logout Other Sessions |
| 6 | Enable 2FA |
| 7 | Disable 2FA |
| 8 | Forget Password |
| 9 | Import Transactions |
| 10 | Export Transactions |
| 11 | Clear All Data |
| 12 | Sync Application Settings |
| 13 | MCP Access |
| 14 | AI Image Recognition |
| 15 | OAuth 2.0 Login |
| 16 | Unlink Third-Party Login |
| 17 | Generate API Token |

### user-resend-verify-email

Resend the email verification message.

```bash
./honey-ledger userdata user-resend-verify-email --username alice
```

### user-set-email-verified / user-set-email-unverified

Manually set a user's email verification status.

```bash
./honey-ledger userdata user-set-email-verified --username alice
./honey-ledger userdata user-set-email-unverified --username alice
```

### user-delete

Permanently delete a user and all their data.

```bash
./honey-ledger userdata user-delete --username alice
```

### user-2fa-disable

Disable two-factor authentication for a user (admin recovery).

```bash
./honey-ledger userdata user-2fa-disable --username alice
```

### user-session-list

List all active sessions for a user.

```bash
./honey-ledger userdata user-session-list --username alice
```

### user-session-new

Create a new session token for a user (outputs the token).

```bash
./honey-ledger userdata user-session-new --username alice
```

### user-session-revoke

Revoke a specific session.

```bash
./honey-ledger userdata user-session-revoke --username alice --session-id <id>
```

### user-session-clear

Revoke all sessions for a user.

```bash
./honey-ledger userdata user-session-clear --username alice
```

### send-password-reset-mail

Send a password reset email to a user.

```bash
./honey-ledger userdata send-password-reset-mail --username alice
```

### transaction-check

Verify transaction and account data integrity for a user.

```bash
./honey-ledger userdata transaction-check --username alice
```

### transaction-tag-index-fix-transaction-time

Fix transaction time values in the tag index table.

```bash
./honey-ledger userdata transaction-tag-index-fix-transaction-time --username alice
```

### transaction-import

Import transactions from a file via the CLI.

```bash
./honey-ledger userdata transaction-import \
  --username alice \
  --file /path/to/transactions.csv \
  --type honey_ledger_csv
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--username` | `-n` | Yes | Username |
| `--file` | `-f` | Yes | Path to import file |
| `--type` | `-t` | Yes | File type identifier |

### transaction-export

Export transactions to a file via the CLI.

```bash
./honey-ledger userdata transaction-export \
  --username alice \
  --file /path/to/export.csv \
  --type csv
```

---

## security gen-secret-key

Generate a random secret key for JWT signing.

```bash
./honey-ledger security gen-secret-key
```

| Flag | Description | Default |
|------|-------------|---------|
| `--length` | Key length in bytes | `32` |

Output: A hex-encoded random string. Copy this into `conf/honey-ledger.ini`:

```ini
[security]
secret_key = <generated-key>
```

---

## cron Commands

### cron list

List all registered cron jobs.

```bash
./honey-ledger cron list
```

### cron run

Manually trigger a cron job.

```bash
./honey-ledger cron run --name <job-name>
```

| Job Name | Description |
|----------|-------------|
| `RemoveExpiredTokens` | Clean up expired session tokens |
| `CreateScheduledTransaction` | Create pending scheduled transactions |

---

## utility Commands

### parse-default-request-id

Decode a request ID into its component parts for debugging.

```bash
./honey-ledger utility parse-default-request-id --id <request-id>
```

Output includes: server ID, instance ID, timestamp, sequence number, and client IP address.

### send-test-mail

Send a test email to verify SMTP configuration.

```bash
./honey-ledger utility send-test-mail --to user@example.com
```

Requires SMTP to be configured in `[mail]` section.

---

## honeytools API Script

`scripts/honeytools.sh` (Linux/macOS) and `scripts/honeytools.ps1` (Windows) provide a CLI tool for calling Honey Ledger API endpoints from the terminal or AI agents.

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `HONEYTOOL_SERVER_BASEURL` | Yes | Server URL (e.g., `http://localhost:8080`) |
| `HONEYTOOL_TOKEN` | Yes | API or session token |

### Usage

```bash
export HONEYTOOL_SERVER_BASEURL="http://localhost:8080"
export HONEYTOOL_TOKEN="your-api-token"

# List available commands
./scripts/honeytools.sh list

# Get help for a command
./scripts/honeytools.sh help accounts-list

# Run a command
./scripts/honeytools.sh accounts-list
./scripts/honeytools.sh transactions-list --count 10
```

### Available Commands

The script provides CLI wrappers for the most common API operations:

| Command | API Endpoint |
|---------|-------------|
| `tokens-list` | `GET /api/v1/tokens/list.json` |
| `tokens-revoke` | `POST /api/v1/tokens/revoke.json` |
| `accounts-list` | `GET /api/v1/accounts/list.json` |
| `accounts-add` | `POST /api/v1/accounts/add.json` |
| `transactions-list` | `GET /api/v1/transactions/list.json` |
| `transactions-add` | `POST /api/v1/transactions/add.json` |
| `categories-list` | `GET /api/v1/transaction/categories/list.json` |
| `tags-list` | `GET /api/v1/transaction/tags/list.json` |
| `exchange-rates-latest` | `GET /api/v1/exchange_rates/latest.json` |

Run `./scripts/honeytools.sh list` for the full list of available commands and `help <command>` for detailed parameter information.

### MCP / AI Agent Integration

The honeytools script is designed for use as an AI skill. AI agents can use it to:

- Read financial data (accounts, transactions, balances)
- Create transactions
- Query statistics
- Manage exchange rates

Provide the script path and environment variables to your AI agent configuration.
