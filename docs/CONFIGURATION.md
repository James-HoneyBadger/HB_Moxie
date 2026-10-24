# Configuration Reference

Complete reference for all Ratel Ledger configuration options.

Configuration is read from `conf/ratel-ledger.ini` (INI format). Every setting can be overridden via environment variables.

---

## Table of Contents

- [Environment Variable Overrides](#environment-variable-overrides)
- [Global](#global)
- [Server](#server)
- [MCP (Model Context Protocol)](#mcp-model-context-protocol)
- [Database](#database)
- [Mail](#mail)
- [Log](#log)
- [Storage](#storage)
- [LLM (Large Language Models)](#llm-large-language-models)
- [LLM Image Recognition](#llm-image-recognition)
- [UUID](#uuid)
- [Duplicate Checker](#duplicate-checker)
- [Cron Jobs](#cron-jobs)
- [Security](#security)
- [Authentication](#authentication)
- [User](#user)
- [Data](#data)
- [Tips](#tips)
- [Notifications](#notifications)
- [Map](#map)
- [Exchange Rates](#exchange-rates)

---

## Environment Variable Overrides

Every configuration key in `ratel-ledger.ini` can be overridden by an environment variable without modifying the file.

### Direct Value Override

Pattern: `RATEL_{SECTION}_{KEY}`

```bash
# Override [server] http_port
export RATEL_SERVER_HTTP_PORT=9090

# Override [database] type
export RATEL_DATABASE_TYPE=mysql
```

### File-Based Override

Pattern: `RATELCFP_{SECTION}_{KEY}` — the value is read from the file path specified.

```bash
# Read database password from a file
export RATELCFP_DATABASE_PASSWD=/run/secrets/db_password
```

### Working Directory Override

```bash
export RATEL_WORK_DIR=/opt/ratel-ledger
```

Section and key names are **uppercase** in the environment variable. Underscores in key names remain underscores.

---

## Global

```ini
[global]
# Application mode: production or development
# Development mode enables verbose logging and debug features.
mode = production
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mode` | string | `production` | `production` or `development` |

---

## Server

```ini
[server]
http_addr = 127.0.0.1
http_port = 8080
domain = localhost
root_url = %(protocol)s://%(domain)s:%(http_port)s/
log_request = true
request_id_header = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `http_addr` | string | `127.0.0.1` | Listen address. Non-loopback values are coerced to `127.0.0.1` |
| `http_port` | int | `8080` | HTTP listen port (Electron assigns a random port at runtime) |
| `domain` | string | `localhost` | Domain name |
| `root_url` | string | *(derived)* | Full URL. Supports `%(protocol)s`, `%(domain)s`, `%(http_port)s` interpolation. |
| `log_request` | bool | `true` | Log HTTP requests. `token`, `code`, `state`, and `password` query values are redacted |
| `request_id_header` | bool | `true` | Add X-Request-Id header to responses |

---

## MCP (Model Context Protocol)

```ini
[mcp]
enable_mcp = false
mcp_allowed_remote_ips =
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_mcp` | bool | `false` | Enable the MCP JSON-RPC endpoint at `/mcp` |
| `mcp_allowed_remote_ips` | string | *(empty)* | Comma-separated list of allowed client IPs. Empty = all allowed. |

---

## Database

```ini
[database]
type = sqlite3
host = 127.0.0.1:3306
name = ratel-ledger
user = root
passwd =
ssl_mode = disable
db_path = data/ratel-ledger.db
max_idle_conn = 2
max_open_conn = 0
conn_max_lifetime = 14400
log_query = false
auto_update_database = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `type` | string | `sqlite3` | `sqlite3`, `mysql`, or `postgres` |
| `host` | string | `127.0.0.1:3306` | Database host and port (MySQL/PostgreSQL) |
| `name` | string | `ratel-ledger` | Database name (MySQL/PostgreSQL) |
| `user` | string | `root` | Database user (MySQL/PostgreSQL) |
| `passwd` | string | *(empty)* | Database password |
| `ssl_mode` | string | `disable` | SSL mode for PostgreSQL (`disable`, `require`, `verify-full`) |
| `db_path` | string | `data/ratel-ledger.db` | SQLite database file path |
| `max_idle_conn` | int | `2` | Maximum idle database connections |
| `max_open_conn` | int | `0` | Maximum open connections. `0` means unlimited for MySQL/PostgreSQL, but is treated as `1` for SQLite |
| `conn_max_lifetime` | int | `14400` | Connection max lifetime in seconds |
| `log_query` | bool | `false` | Log all database queries (development only) |
| `auto_update_database` | bool | `true` | Automatically sync schema on startup |

---

## Mail

```ini
[mail]
enable_smtp = false
smtp_host = 127.0.0.1:25
smtp_user =
smtp_passwd =
smtp_skip_tls_verify = false
from_address =
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_smtp` | bool | `false` | Enable SMTP mail delivery |
| `smtp_host` | string | `127.0.0.1:25` | SMTP server host:port |
| `smtp_user` | string | *(empty)* | SMTP username |
| `smtp_passwd` | string | *(empty)* | SMTP password |
| `smtp_skip_tls_verify` | bool | `false` | Skip TLS certificate verification |
| `from_address` | string | *(empty)* | Sender email address |

---

## Log

```ini
[log]
mode = console
level = info
log_path = log/ratel-ledger.log
request_log_path =
query_log_path =
log_file_rotate = true
log_file_max_size = 104857600
log_file_max_days = 7
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mode` | string | `console` | `console` or `file` |
| `level` | string | `info` | `debug`, `info`, `warn`, or `error` |
| `log_path` | string | `log/ratel-ledger.log` | Main log file path |
| `request_log_path` | string | *(empty)* | Separate request log file (empty = use main log) |
| `query_log_path` | string | *(empty)* | Separate database query log file |
| `log_file_rotate` | bool | `true` | Enable log file rotation |
| `log_file_max_size` | int | `104857600` | Max log file size in bytes (100 MB) |
| `log_file_max_days` | int | `7` | Days to retain rotated log files |

---

## Storage

```ini
[storage]
type = local_filesystem
local_filesystem_path = storage
minio_endpoint =
minio_location =
minio_access_key_id =
minio_secret_access_key =
minio_use_ssl = false
minio_skip_tls_verify = false
minio_bucket =
minio_root_path =
webdav_endpoint =
webdav_root_path =
webdav_user =
webdav_passwd =
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `type` | string | `local_filesystem` | `local_filesystem`, `minio`, or `webdav` |
| `local_filesystem_path` | string | `storage` | Local storage root directory |
| `minio_endpoint` | string | *(empty)* | MinIO server endpoint |
| `minio_location` | string | *(empty)* | MinIO region/location |
| `minio_access_key_id` | string | *(empty)* | MinIO access key |
| `minio_secret_access_key` | string | *(empty)* | MinIO secret key |
| `minio_use_ssl` | bool | `false` | Use SSL for MinIO connections |
| `minio_skip_tls_verify` | bool | `false` | Skip TLS verification for MinIO |
| `minio_bucket` | string | *(empty)* | MinIO bucket name |
| `minio_root_path` | string | *(empty)* | Root path within the bucket |
| `webdav_endpoint` | string | *(empty)* | WebDAV server URL |
| `webdav_root_path` | string | *(empty)* | Root path on WebDAV server |
| `webdav_user` | string | *(empty)* | WebDAV username |
| `webdav_passwd` | string | *(empty)* | WebDAV password |

---

## LLM (Large Language Models)

```ini
[llm]
transaction_from_ai_image_recognition = false
max_ai_recognition_picture_size = 524288
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `transaction_from_ai_image_recognition` | bool | `false` | Enable AI receipt image recognition |
| `max_ai_recognition_picture_size` | int | `524288` | Max image size in bytes (512 KB) |

---

## LLM Image Recognition

```ini
[llm_image_recognition]
llm_provider =
openai_api_key =
openai_api_model_id = gpt-4o
openai_api_url = https://api.openai.com/v1/chat/completions
openai_compatible_api_key =
openai_compatible_api_model_id =
openai_compatible_api_url =
anthropic_api_key =
anthropic_api_model_id = claude-sonnet-4-20250514
anthropic_api_url = https://api.anthropic.com/v1/messages
anthropic_compatible_api_key =
anthropic_compatible_api_model_id =
anthropic_compatible_api_url =
openrouter_api_key =
openrouter_api_model_id =
openrouter_api_url = https://openrouter.ai/api/v1/chat/completions
ollama_api_model_id =
ollama_api_url = http://127.0.0.1:11434/api/chat
lm_studio_api_model_id =
lm_studio_api_url = http://127.0.0.1:1234/v1/chat/completions
google_ai_api_key =
google_ai_api_model_id = gemini-2.0-flash
google_ai_api_url = https://generativelanguage.googleapis.com/v1beta/models
request_timeout = 30
proxy =
skip_tls_verify = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `llm_provider` | string | *(empty)* | Provider: `openai`, `openai_compatible`, `anthropic`, `anthropic_compatible`, `openrouter`, `ollama`, `lm_studio`, `google_ai` |
| `*_api_key` | string | *(empty)* | API key for the selected provider |
| `*_api_model_id` | string | *(varies)* | Model ID. Defaults: `gpt-4o` (OpenAI), `claude-sonnet-4-20250514` (Anthropic), `gemini-2.0-flash` (Google) |
| `*_api_url` | string | *(varies)* | API endpoint URL |
| `request_timeout` | int | `30` | Request timeout in seconds |
| `proxy` | string | *(empty)* | HTTP proxy URL |
| `skip_tls_verify` | bool | `false` | Skip TLS certificate verification |

---

## UUID

```ini
[uuid]
generator_type = internal
server_id = 0
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `generator_type` | string | `internal` | UUID generation strategy |
| `server_id` | int | `0` | Server identifier |

---

## Duplicate Checker

```ini
[duplicate_checker]
checker_type = in_memory
cleanup_interval = 60
duplicate_submissions_interval = 5
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `checker_type` | string | `in_memory` | Duplicate request detection backend |
| `cleanup_interval` | int | `60` | Cleanup interval in seconds |
| `duplicate_submissions_interval` | int | `5` | Window in seconds for duplicate detection |

---

## Cron Jobs

```ini
[cron]
enable_remove_expired_tokens = true
enable_create_scheduled_transaction = true
enable_auto_backup = true
auto_backup_interval_hours = 24
auto_backup_max_count = 10
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_remove_expired_tokens` | bool | `true` | Periodically remove expired session tokens |
| `enable_create_scheduled_transaction` | bool | `true` | Automatically create scheduled (recurring) transactions |
| `enable_auto_backup` | bool | `true` | Periodically create backups (SQLite databases only) |
| `auto_backup_interval_hours` | int | `24` | Hours between automatic backups |
| `auto_backup_max_count` | int | `10` | Number of backups to retain |

---

## Security

```ini
[security]
secret_key =
token_expired_time = 2592000
token_min_refresh_interval = 86400
temporary_token_expired_time = 300
email_verify_token_expired_time = 3600
password_reset_token_expired_time = 3600
enable_api_token = true
api_token_allowed_remote_ips =
max_failures_per_ip_per_minute = 10
max_failures_per_user_per_minute = 5
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `secret_key` | string | *(empty)* | Encrypts 2FA secrets and derives server identifiers. When empty, a random key is generated on first start and persisted with `0600` permissions. Generate manually with `./ratel-ledger security gen-secret-key`. |
| `token_expired_time` | int | `2592000` | Session token lifetime in seconds (30 days) |
| `token_min_refresh_interval` | int | `86400` | Minimum interval between token refreshes (24 hours) |
| `temporary_token_expired_time` | int | `300` | Temporary token lifetime in seconds (5 minutes) |
| `email_verify_token_expired_time` | int | `3600` | Email verification token lifetime (1 hour) |
| `password_reset_token_expired_time` | int | `3600` | Password reset token lifetime (1 hour) |
| `enable_api_token` | bool | `true` | Allow users to generate API access tokens |
| `api_token_allowed_remote_ips` | string | *(empty)* | Restrict API token usage to these IPs (comma-separated) |
| `max_failures_per_ip_per_minute` | int | `10` | Rate limit: max auth failures per IP per minute |
| `max_failures_per_user_per_minute` | int | `5` | Rate limit: max auth failures per user per minute |

---

## Authentication

```ini
[auth]
enable_internal_auth = true
enable_desktop_auto_login = true
enable_oauth2_auth = false
enable_two_factor = true
enable_forget_password = true
forget_password_require_email_verify = false
oauth2_provider_name =
oauth2_provider_type =
oauth2_client_id =
oauth2_client_secret =
oauth2_auth_url =
oauth2_token_url =
oauth2_user_info_url =
oauth2_scope =
oauth2_user_identifier = email
oauth2_enable_pkce = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_internal_auth` | bool | `true` | Enable username/password login |
| `enable_desktop_auto_login` | bool | `true` | Skip the login screen when exactly one user exists and that user has not enabled 2FA |
| `enable_oauth2_auth` | bool | `false` | Enable OAuth2/OIDC login |
| `enable_two_factor` | bool | `true` | Allow users to enable TOTP 2FA |
| `enable_forget_password` | bool | `true` | Enable password reset flow |
| `forget_password_require_email_verify` | bool | `false` | Require verified email for password reset |
| `oauth2_provider_name` | string | *(empty)* | Display name for the OAuth2 provider |
| `oauth2_provider_type` | string | *(empty)* | `oidc`, `nextcloud`, `gitea`, or `github` |
| `oauth2_client_id` | string | *(empty)* | OAuth2 client ID |
| `oauth2_client_secret` | string | *(empty)* | OAuth2 client secret |
| `oauth2_auth_url` | string | *(empty)* | Authorization endpoint URL |
| `oauth2_token_url` | string | *(empty)* | Token endpoint URL |
| `oauth2_user_info_url` | string | *(empty)* | User info endpoint URL |
| `oauth2_scope` | string | *(empty)* | OAuth2 scopes (space-separated) |
| `oauth2_user_identifier` | string | `email` | Field to match users: `email` or `username` |
| `oauth2_enable_pkce` | bool | `false` | Enable PKCE for OAuth2 flow |

---

## User

```ini
[user]
enable_register = true
enable_email_verify = false
enable_force_email_verify = false
enable_transaction_picture = true
max_transaction_picture_size = 524288
enable_scheduled_transaction = true
avatar_provider = internal
max_user_avatar_size = 1048576
default_feature_restrictions =
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_register` | bool | `true` | Allow new user self-registration |
| `enable_email_verify` | bool | `false` | Enable email verification flow |
| `enable_force_email_verify` | bool | `false` | Block unverified users from most features |
| `enable_transaction_picture` | bool | `true` | Allow attaching pictures to transactions |
| `max_transaction_picture_size` | int | `524288` | Max picture upload size in bytes (512 KB) |
| `enable_scheduled_transaction` | bool | `true` | Enable scheduled/recurring transactions |
| `avatar_provider` | string | `internal` | `internal` (uploaded) or `gravatar` |
| `max_user_avatar_size` | int | `1048576` | Max avatar upload size in bytes (1 MB) |
| `default_feature_restrictions` | string | *(empty)* | Default feature restriction bitmask for new users |

### Feature Restriction Types

Restrictions are specified as a comma-separated list of type numbers:

| Type | Description |
|------|-------------|
| 1 | Update Password |
| 2 | Update Email |
| 3 | Update Profile Basic Info |
| 4 | Update Avatar |
| 5 | Logout Other Sessions |
| 6 | Enable Two-Factor Authentication |
| 7 | Disable Two-Factor Authentication |
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

---

## Data

```ini
[data]
enable_export = true
enable_import = true
max_import_file_size = 2097152
enable_backup = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_export` | bool | `true` | Allow transaction data export |
| `enable_import` | bool | `true` | Allow transaction data import |
| `max_import_file_size` | int | `2097152` | Max import file size in bytes (2 MB) |
| `enable_backup` | bool | `true` | Enable backup creation, download, restore, and deletion |

---

## Tips

```ini
[tip]
enable_tips_in_login_page = false
login_page_tips_content =
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_tips_in_login_page` | bool | `false` | Show a tips banner on the login page |
| `login_page_tips_content` | string | *(empty)* | HTML content for the login page tips banner |

---

## Notifications

Multi-language notification messages shown after specific events.

```ini
[notification]
after_register_notification_content =
after_login_notification_content =
after_open_notification_content =
```

Each key supports language suffixes: `after_login_notification_content_en`, `after_login_notification_content_zh_hans`, etc. The unsuffixed key is the default fallback.

| Key | Type | Description |
|-----|------|-------------|
| `after_register_notification_content` | string | Message shown after user registration |
| `after_login_notification_content` | string | Message shown after login |
| `after_open_notification_content` | string | Message shown when the app opens |

---

## Map

```ini
[map]
map_provider = openstreetmap
enable_map_data_fetch_proxy = false
proxy =
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `map_provider` | string | `openstreetmap` | See supported providers below |
| `enable_map_data_fetch_proxy` | bool | `false` | Proxy map tile requests through the backend |
| `proxy` | string | *(empty)* | HTTP proxy for map tile fetching |

### Supported Map Providers

| Provider | Key | Additional Config |
|----------|-----|-------------------|
| OpenStreetMap | `openstreetmap` | None |
| OpenStreetMap (Humanitarian) | `openstreetmap_humanitarian` | None |
| OpenTopoMap | `opentopomap` | None |
| OPNVKarte | `opnvkarte` | None |
| CyclOSM | `cyclosm` | None |
| CartoDB | `cartodb` | None |
| TomTom | `tomtom` | `tomtom_map_api_key` |
| Tianditu | `tianditu` | `tianditu_map_api_key` |
| Google Maps | `googlemap` | None (uses public tiles) |
| Baidu Map | `baidumap` | None |
| Amap (Gaode) | `amap` | `amap_application_key`, `amap_application_secret`, `amap_api_security_verification_method` |
| Custom Tile Server | `custom` | `custom_map_tile_server_url`, `custom_map_tile_server_min_zoom`, `custom_map_tile_server_max_zoom`, `custom_map_tile_server_default_zoom` |
| Custom with Annotation | `custom_with_annotation_layer` | Custom tile + annotation layer URLs |

---

## Exchange Rates

```ini
[exchange_rates]
data_source = euro_central_bank
request_timeout = 10
proxy =
skip_tls_verify = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `data_source` | string | `euro_central_bank` | Exchange rate data provider |
| `request_timeout` | int | `10` | Request timeout in seconds |
| `proxy` | string | *(empty)* | HTTP proxy for rate fetching |
| `skip_tls_verify` | bool | `false` | Skip TLS verification |

### Supported Exchange Rate Sources

| Source | Key |
|--------|-----|
| European Central Bank | `euro_central_bank` |
| Bank of Canada | `bank_of_canada` |
| Czech National Bank | `czech_national_bank` |
| Danish National Bank | `danish_national_bank` |
| National Bank of Georgia | `national_bank_of_georgia` |
| Hungarian National Bank | `hungarian_national_bank` |
| Bank of Israel | `bank_of_israel` |
| National Bank of Kazakhstan | `national_bank_of_kazakhstan` |
| National Bank of Poland | `national_bank_of_poland` |
| Central Bank of Russia | `central_bank_of_russia` |
| Swiss National Bank | `swiss_national_bank` |
| National Bank of Ukraine | `national_bank_of_ukraine` |
| Reserve Bank of Australia | `reserve_bank_of_australia` |
| International Monetary Fund | `international_monetary_fund` |
| Fixer.io | `fixer_io` (requires `fixer_io_api_key`) |
| User Custom | `user_custom` |
