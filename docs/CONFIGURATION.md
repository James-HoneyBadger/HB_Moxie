# Configuration Reference

Complete reference for every Honey Ledger configuration option.

Settings live in `conf/honey-ledger.ini` (INI format) inside the working directory. The file is not generated at startup — the server exits with `cannot get default configuration path` when it is missing. The Electron app copies the bundled template into the user data directory on first launch, packaged releases ship the template as `conf/honey-ledger.ini`, and `./run.sh` uses the one in the repository.

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
- [LLM AI Assistant](#llm-ai-assistant)
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

Every configuration key in `honey-ledger.ini` can be overridden by an environment variable without modifying the file.

### Direct Value Override

Pattern: `HONEY_{SECTION}_{KEY}`

```bash
# Override [server] http_port
export HONEY_SERVER_HTTP_PORT=9090

# Override [database] type
export HONEY_DATABASE_TYPE=mysql
```

### File-Based Override

Pattern: `HONEYCFP_{SECTION}_{KEY}` — the value is read from the file path specified.

```bash
# Read database password from a file
export HONEYCFP_DATABASE_PASSWD=/run/secrets/db_password
```

### Working Directory Override

```bash
export HONEY_WORK_DIR=/opt/honey-ledger
```

### Legacy Compatibility

Installs created before the rename to Honey Ledger keep working. Each name is checked in order, newest first:

| Setting | Lookup order |
|---------|--------------|
| Working directory | `HONEY_WORK_DIR`, `RATEL_WORK_DIR`, `HBM_WORK_DIR` |
| Value override | `HONEY_{SECTION}_{KEY}`, `RATEL_{SECTION}_{KEY}`, `HBM_{SECTION}_{KEY}` |
| File-based override | `HONEYCFP_{SECTION}_{KEY}`, `RATELCFP_{SECTION}_{KEY}`, `HBMCFP_{SECTION}_{KEY}` |
| Config file | `conf/honey-ledger.ini`, `conf/ratel-ledger.ini`, `conf/hbmoxie.ini` |

Section and key names are **uppercase** in the environment variable. Underscores in key names remain underscores.

---

## Global

```ini
[global]
mode = development
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mode` | string | `development` | `production` or `development`. Development mode enables verbose boot logging |

---

## Server

```ini
[server]
http_port = 8080
domain = localhost
root_url = %(protocol)s://%(domain)s:%(http_port)s/
log_request = true
request_id_header = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `http_addr` | string | `127.0.0.1` | Listen address. Not present in the template; loopback only — any non-loopback value is forced back to `127.0.0.1` for desktop safety |
| `http_port` | int | `8080` | HTTP listen port. Electron assigns a free port at runtime and passes it to the frontend |
| `domain` | string | `localhost` | Domain used to build absolute URLs |
| `root_url` | string | `%(protocol)s://%(domain)s:%(http_port)s/` | Full URL. Supports the `%(protocol)s`, `%(domain)s`, and `%(http_port)s` placeholders. The protocol is always `http` |
| `log_request` | bool | `true` | Log every request with its duration. `token`, `access_token`, `id_token`, `code`, `state`, and `password` query values are redacted |
| `request_id_header` | bool | `true` | Add an `X-Request-Id` response header so a request or error can be traced |

---

## MCP (Model Context Protocol)

```ini
[mcp]
enable_mcp = false
mcp_allowed_remote_ips = 
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_mcp` | bool | `false` | Serve the MCP JSON-RPC endpoint at `/mcp` for AI agent access |
| `mcp_allowed_remote_ips` | string | *(empty)* | Comma-separated list of allowed client IPs, `*` matches any address (for example `192.168.1.*`). Empty allows every IP |

---

## Database

```ini
[database]
type = sqlite3
host = 127.0.0.1:3306
name = honey_ledger
user = root
passwd = 
ssl_mode = disable
db_path = data/honey-ledger.db
max_idle_conn = 2
max_open_conn = 0
conn_max_lifetime = 14400
log_query = false
auto_update_database = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `type` | string | `sqlite3` | `sqlite3`, `mysql`, or `postgres` |
| `host` | string | `127.0.0.1:3306` | Connection address, for `mysql` and `postgres` |
| `name` | string | `honey_ledger` | Database name, for `mysql` and `postgres` |
| `user` | string | `root` | Database user, for `mysql` and `postgres` |
| `passwd` | string | *(empty)* | Database password, for `mysql` and `postgres` |
| `ssl_mode` | string | `disable` | For `postgres` only: `disable`, `require`, or `verify-full` |
| `db_path` | string | `data/honey-ledger.db` | For `sqlite3` only: database file path, relative to the working directory or absolute |
| `max_idle_conn` | int | `2` | Max idle connections (0 - 65535, 0 retains none) |
| `max_open_conn` | int | `0` | Max open connections (0 - 65535, 0 is unlimited). For SQLite an empty or zero value is treated as 1 because SQLite serializes writers |
| `conn_max_lifetime` | int | `14400` | Max connection lifetime in seconds (0 - 4294967295) |
| `log_query` | bool | `false` | Log every SQL statement with its execution time |
| `auto_update_database` | bool | `true` | Sync the table structure on start and apply pending schema migrations |

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
| `enable_smtp` | bool | `false` | Send mail through an SMTP server |
| `smtp_host` | string | `127.0.0.1:25` | SMTP server address including the port |
| `smtp_user` | string | *(empty)* | SMTP user name |
| `smtp_passwd` | string | *(empty)* | SMTP password |
| `smtp_skip_tls_verify` | bool | `false` | Skip TLS certificate verification when connecting |
| `from_address` | string | *(empty)* | Sender address, either `user@domain.com` or `"Name" <user@domain.com>` |

---

## Log

```ini
[log]
mode = console file
level = info
log_path = log/honey-ledger.log
request_log_path = 
query_log_path = 
log_file_rotate = false
log_file_max_size = 104857600
log_file_max_days = 7
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mode` | string | `console file` | `console`, `file`, or both separated by a space |
| `level` | string | `info` | `debug`, `info`, `warn`, or `error` |
| `log_path` | string | `log/honey-ledger.log` | For `file` mode: log file path, relative to the working directory or absolute |
| `request_log_path` | string | *(empty)* | For `file` mode: separate request log path. Empty writes requests into the main log |
| `query_log_path` | string | *(empty)* | For `file` mode: separate query log path. Empty writes queries into the main log |
| `log_file_rotate` | bool | `false` | Rotate log files |
| `log_file_max_size` | int | `104857600` | Size in bytes (1 - 4294967295) at which a log file is rotated |
| `log_file_max_days` | int | `7` | Days to retain rotated files, `0` keeps all of them |

---

## Storage

Attachment storage for avatars and transaction pictures.

```ini
[storage]
type = local_filesystem
local_filesystem_path = storage/
minio_endpoint = 127.0.0.1:9000
minio_location = 
minio_access_key_id = 
minio_secret_access_key = 
minio_use_ssl = false
minio_skip_tls_verify = false
minio_bucket = honey-ledger
minio_root_path = /
webdav_url = 
webdav_username = 
webdav_password = 
webdav_root_path = /
webdav_request_timeout = 10000
webdav_proxy = system
webdav_skip_tls_verify = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `type` | string | `local_filesystem` | `local_filesystem`, `minio`, or `webdav` |
| `local_filesystem_path` | string | `storage/` | For `local_filesystem` only: storage root, relative or absolute |
| `minio_endpoint` | string | `127.0.0.1:9000` | For `minio` only: server endpoint |
| `minio_location` | string | *(empty)* | For `minio` only: region/location |
| `minio_access_key_id` | string | *(empty)* | For `minio` only: access key |
| `minio_secret_access_key` | string | *(empty)* | For `minio` only: secret key |
| `minio_use_ssl` | bool | `false` | For `minio` only: use SSL |
| `minio_skip_tls_verify` | bool | `false` | For `minio` only: skip TLS verification |
| `minio_bucket` | string | `honey-ledger` | For `minio` only: bucket name |
| `minio_root_path` | string | `/` | For `minio` only: root path inside the bucket |
| `webdav_url` | string | *(empty)* | For `webdav` only: server URL |
| `webdav_username` | string | *(empty)* | For `webdav` only: user name |
| `webdav_password` | string | *(empty)* | For `webdav` only: password |
| `webdav_root_path` | string | `/` | For `webdav` only: root path to store files |
| `webdav_request_timeout` | int | `10000` | For `webdav` only: request timeout in milliseconds (0 - 4294967295), `0` disables the timeout |
| `webdav_proxy` | string | `system` | For `webdav` only: `system`, `none`, or a proxy URL starting with `http://`, `https://`, or `socks5://` |
| `webdav_skip_tls_verify` | bool | `false` | For `webdav` only: skip TLS verification |

---


## LLM (Large Language Models)

Switches and size limits shared by both LLM features.

```ini
[llm]
transaction_from_ai_image_recognition = false
max_ai_recognition_picture_size = 10485760
ai_assistant = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `transaction_from_ai_image_recognition` | bool | `false` | Enable creating transactions from receipt images |
| `max_ai_recognition_picture_size` | int | `10485760` | Max receipt image size in bytes (10 MB) |
| `ai_assistant` | bool | `false` | Enable `POST /api/v1/ai/ask.json`, which answers questions about the user's finances with a configured model |

---

## LLM Image Recognition

Reads receipts and turns them into transaction drafts. Both LLM sections have the same shape: set `llm_provider` and fill in only that provider's keys.

```ini
[llm_image_recognition]
llm_provider = 
openai_api_key = 
openai_model_id = 
openai_compatible_base_url = 
openai_compatible_api_key = 
openai_compatible_model_id = 
anthropic_api_key = 
anthropic_model_id = 
anthropic_max_tokens = 1024
anthropic_compatible_base_url = 
anthropic_compatible_api_version = 
anthropic_compatible_api_key = 
anthropic_compatible_model_id = 
anthropic_compatible_max_tokens = 1024
openrouter_api_key = 
openrouter_model_id = 
ollama_server_url = 
ollama_model_id = 
lm_studio_server_url = 
lm_studio_token = 
lm_studio_model_id = 
google_ai_api_key = 
google_ai_model_id = 
request_timeout = 60000
proxy = system
skip_tls_verify = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `llm_provider` | string | *(empty)* | `openai`, `openai_compatible`, `anthropic`, `anthropic_compatible`, `openrouter`, `ollama`, `lm_studio`, or `google_ai`. Empty disables the feature |
| `openai_api_key` | string | *(empty)* | For `openai` only: API key |
| `openai_model_id` | string | *(empty)* | For `openai` only: model ID |
| `openai_compatible_base_url` | string | *(empty)* | For `openai_compatible` only: base URL of the OpenAI-compatible API |
| `openai_compatible_api_key` | string | *(empty)* | For `openai_compatible` only: API key |
| `openai_compatible_model_id` | string | *(empty)* | For `openai_compatible` only: model ID |
| `anthropic_api_key` | string | *(empty)* | For `anthropic` only: API key |
| `anthropic_model_id` | string | *(empty)* | For `anthropic` only: model ID |
| `anthropic_max_tokens` | int | `1024` | For `anthropic` and `anthropic_compatible`: max response tokens |
| `anthropic_compatible_base_url` | string | *(empty)* | For `anthropic_compatible` only: base URL |
| `anthropic_compatible_api_version` | string | *(empty)* | For `anthropic_compatible` only: `anthropic-version` header value |
| `anthropic_compatible_api_key` | string | *(empty)* | For `anthropic_compatible` only: API key |
| `anthropic_compatible_model_id` | string | *(empty)* | For `anthropic_compatible` only: model ID |
| `anthropic_compatible_max_tokens` | int | `1024` | For `anthropic_compatible` only: max response tokens |
| `openrouter_api_key` | string | *(empty)* | For `openrouter` only: API key |
| `openrouter_model_id` | string | *(empty)* | For `openrouter` only: model ID |
| `ollama_server_url` | string | *(empty)* | For `ollama` only: server URL |
| `ollama_model_id` | string | *(empty)* | For `ollama` only: model ID |
| `lm_studio_server_url` | string | *(empty)* | For `lm_studio` only: server URL |
| `lm_studio_token` | string | *(empty)* | For `lm_studio` only: access token, if the server requires one |
| `lm_studio_model_id` | string | *(empty)* | For `lm_studio` only: model ID |
| `google_ai_api_key` | string | *(empty)* | For `google_ai` only: API key |
| `google_ai_model_id` | string | *(empty)* | For `google_ai` only: model ID |
| `request_timeout` | int | `60000` | Request timeout in milliseconds |
| `proxy` | string | `system` | `system`, `none`, or a proxy URL starting with `http://`, `https://`, or `socks5://` |
| `skip_tls_verify` | bool | `false` | Skip TLS certificate verification |

---

## LLM AI Assistant

Answers questions about the user's finances through `POST /api/v1/ai/ask.json`. Only an aggregated, de-identified summary is sent to the model: totals per currency, expense totals per category name, account balances, and shared expense balances labelled by position. Each response reports what was sent in its `privacyNote`.

```ini
[llm_ai_assistant]
llm_provider = 
openai_api_key = 
openai_model_id = 
openai_compatible_base_url = 
openai_compatible_api_key = 
openai_compatible_model_id = 
anthropic_api_key = 
anthropic_model_id = 
anthropic_max_tokens = 1024
anthropic_compatible_base_url = 
anthropic_compatible_api_version = 
anthropic_compatible_api_key = 
anthropic_compatible_model_id = 
anthropic_compatible_max_tokens = 1024
openrouter_api_key = 
openrouter_model_id = 
ollama_server_url = 
ollama_model_id = 
lm_studio_server_url = 
lm_studio_token = 
lm_studio_model_id = 
google_ai_api_key = 
google_ai_model_id = 
request_timeout = 60000
proxy = system
skip_tls_verify = false
```

The keys are identical to [LLM Image Recognition](#llm-image-recognition). The assistant is only reachable when `[llm] ai_assistant` is `true`, and an empty `llm_provider` keeps the endpoint disabled even then.

---

## UUID

```ini
[uuid]
generator_type = internal
server_id = 0
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `generator_type` | string | `internal` | UUID generator, `internal` is the only supported value |
| `server_id` | int | `0` | Unique server id (0 - 255) that keeps generated ids apart when several servers share a database |

---

## Duplicate Checker

Rejects repeated submissions of the same form, such as double-clicking Save.

```ini
[duplicate_checker]
checker_type = in_memory
cleanup_interval = 60
duplicate_submissions_interval = 300
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `checker_type` | string | `in_memory` | `in_memory` is the only supported value |
| `cleanup_interval` | int | `60` | Interval in seconds (1 - 4294967295) at which expired entries are dropped |
| `duplicate_submissions_interval` | int | `300` | Seconds (0 - 4294967295) during which a repeat submission of the same form is rejected. Leaving and re-entering an edit page counts as a new session. `0` disables the check |

---

## Cron Jobs

Background jobs that run inside the server process.

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
| `enable_remove_expired_tokens` | bool | `true` | Remove expired login, verify, and reset tokens |
| `enable_create_scheduled_transaction` | bool | `true` | Create the transactions of due scheduled templates |
| `enable_auto_backup` | bool | `true` | Create periodic database snapshots |
| `auto_backup_interval_hours` | int | `24` | Hours between automatic backups |
| `auto_backup_max_count` | int | `10` | Number of automatic backups to keep, older ones are deleted |

---

## Security

```ini
[security]
token_expired_time = 2592000
token_min_refresh_interval = 86400
temporary_token_expired_time = 300
email_verify_token_expired_time = 3600
password_reset_token_expired_time = 3600
enable_api_token = false
api_token_allowed_remote_ips = 
max_failures_per_ip_per_minute = 5
max_failures_per_user_per_minute = 5
secret_key = 
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `token_expired_time` | int | `2592000` | Login token lifetime in seconds (60 - 4294967295), 30 days by default |
| `token_min_refresh_interval` | int | `86400` | Minimum seconds between token refreshes. Must be smaller than `token_expired_time`; `0` refreshes on every frontend refresh |
| `temporary_token_expired_time` | int | `300` | Lifetime of the temporary token issued while 2FA is pending |
| `email_verify_token_expired_time` | int | `3600` | Email verification token lifetime in seconds |
| `password_reset_token_expired_time` | int | `3600` | Password reset token lifetime in seconds |
| `enable_api_token` | bool | `false` | Allow generating API tokens for programmatic access |
| `api_token_allowed_remote_ips` | string | *(empty)* | Comma-separated list of client IPs allowed to use API tokens, `*` matches any address. Empty allows every IP |
| `max_failures_per_ip_per_minute` | int | `5` | Failed password/token checks per IP per minute (0 - 4294967295) before further attempts are rejected; `0` disables the limit |
| `max_failures_per_user_per_minute` | int | `5` | Failed password/token checks per user per minute (0 - 4294967295); `0` disables the limit |
| `secret_key` | string | *(generated)* | Encrypts two-factor secrets and derives server identifiers. When empty a random key is generated on first start and written back to the file, so 2FA secrets stay valid across restarts. Changing it invalidates existing two-factor enrollments |

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
oauth2_provider = 
oauth2_client_id = 
oauth2_client_secret = 
oauth2_user_identifier = email
oauth2_use_pkce = false
oauth2_auto_register = true
oauth2_state_expired_time = 300
oauth2_request_timeout = 10000
oauth2_proxy = system
oauth2_skip_tls_verify = false
oidc_provider_base_url = 
oidc_provider_check_issuer_url = true
enable_oidc_display_name = false
oidc_custom_display_name = 
nextcloud_base_url = 
gitea_base_url = 
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_internal_auth` | bool | `true` | Allow username/email and password login |
| `enable_desktop_auto_login` | bool | `true` | The desktop app logs in without a password when exactly one user exists, that user has not enabled 2FA, and email verification is either disabled or complete |
| `enable_oauth2_auth` | bool | `false` | Allow OAuth2/OIDC login |
| `enable_two_factor` | bool | `true` | Allow users to enable TOTP two-factor authentication |
| `enable_forget_password` | bool | `true` | Enable the password reset flow |
| `forget_password_require_email_verify` | bool | `false` | Only allow a reset for users with a verified email address |
| `oauth2_provider` | string | *(empty)* | `oidc`, `nextcloud`, `gitea`, or `github`. Empty disables OAuth2 login even when `enable_oauth2_auth` is `true` |
| `oauth2_client_id` | string | *(empty)* | OAuth2 client ID |
| `oauth2_client_secret` | string | *(empty)* | OAuth2 client secret |
| `oauth2_user_identifier` | string | `email` | Which provider field identifies the user: `email` or `username` |
| `oauth2_use_pkce` | bool | `false` | Use PKCE for the authorization code flow |
| `oauth2_auto_register` | bool | `true` | Create a local user on the first successful third-party login |
| `oauth2_state_expired_time` | int | `300` | Lifetime of the OAuth2 state parameter in seconds |
| `oauth2_request_timeout` | int | `10000` | Timeout in milliseconds for token and user info requests |
| `oauth2_proxy` | string | `system` | `system`, `none`, or a proxy URL starting with `http://`, `https://`, or `socks5://` |
| `oauth2_skip_tls_verify` | bool | `false` | Skip TLS verification when talking to the provider |
| `oidc_provider_base_url` | string | *(empty)* | For `oidc` only: issuer base URL used for discovery |
| `oidc_provider_check_issuer_url` | bool | `true` | For `oidc` only: verify that the issuer in the discovery document matches `oidc_provider_base_url` |
| `enable_oidc_display_name` | bool | `false` | For `oidc` only: show a custom name on the login button instead of the discovery document name |
| `oidc_custom_display_name` | string | *(empty)* | For `oidc` only: the text to show when `enable_oidc_display_name` is `true` |
| `nextcloud_base_url` | string | *(empty)* | For `nextcloud` only: instance base URL |
| `gitea_base_url` | string | *(empty)* | For `gitea` only: instance base URL |

---

## User

```ini
[user]
enable_register = true
enable_email_verify = false
enable_force_email_verify = false
enable_transaction_picture = true
max_transaction_picture_size = 10485760
enable_scheduled_transaction = true
avatar_provider = internal
max_user_avatar_size = 1048576
default_feature_restrictions = 
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_register` | bool | `true` | Allow self-registration |
| `enable_email_verify` | bool | `false` | Send a verification mail after registration |
| `enable_force_email_verify` | bool | `false` | Block unverified users from the features that require a verified address |
| `enable_transaction_picture` | bool | `true` | Allow attaching pictures to transactions |
| `max_transaction_picture_size` | int | `10485760` | Max picture size in bytes (10 MB) |
| `enable_scheduled_transaction` | bool | `true` | Enable scheduled/recurring transactions |
| `avatar_provider` | string | `internal` | `internal` for uploaded avatars or `gravatar` |
| `max_user_avatar_size` | int | `1048576` | Max avatar upload size in bytes (1 MB) |
| `default_feature_restrictions` | string | *(empty)* | Comma-separated restriction types applied to every new user; `0` applies none |

### Feature Restriction Types

Restrictions are given as a comma-separated list of type numbers:

| Type | Description |
|------|-------------|
| 1 | Update Password |
| 2 | Update Email |
| 3 | Update Profile Basic Info |
| 4 | Update Avatar |
| 5 | Revoke Other Sessions |
| 6 | Enable Two-Factor Authentication |
| 7 | Disable Two-Factor Authentication |
| 8 | Forget Password |
| 9 | Import Transactions |
| 10 | Export Transactions |
| 11 | Clear All Data |
| 12 | Sync Application Settings |
| 13 | MCP Access |
| 14 | Create Transaction From AI Image Recognition |
| 15 | OAuth 2.0 Login |
| 16 | Unlink Third-Party Login |
| 17 | Generate API Token |

---

## Data

```ini
[data]
enable_export = true
enable_import = true
max_import_file_size = 10485760
enable_backup = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_export` | bool | `true` | Allow exporting transaction data as CSV, TSV, or a ZIP archive |
| `enable_import` | bool | `true` | Allow importing transaction data |
| `max_import_file_size` | int | `10485760` | Max import file size in bytes (10 MB) |
| `enable_backup` | bool | `true` | Allow creating, downloading, restoring, and deleting database snapshots |

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
| `login_page_tips_content` | string | *(empty)* | Banner content. HTML is allowed |

---

## Notifications

Messages shown in the app after a specific event.

```ini
[notification]
enable_notification_after_register = false
after_register_notification_content = 
enable_notification_after_login = false
after_login_notification_content = 
enable_notification_after_open = false
after_open_notification_content = 
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enable_notification_after_register` | bool | `false` | Show a notification after registration |
| `after_register_notification_content` | string | *(empty)* | Notification content after registration |
| `enable_notification_after_login` | bool | `false` | Show a notification after login |
| `after_login_notification_content` | string | *(empty)* | Notification content after login |
| `enable_notification_after_open` | bool | `false` | Show a notification when the app opens |
| `after_open_notification_content` | string | *(empty)* | Notification content when the app opens |

Each content key accepts language suffixes: `after_login_notification_content_en`, `after_login_notification_content_zh_hans`, and so on. The unsuffixed key is the fallback for every language without its own text.

---

## Map

```ini
[map]
map_provider = openstreetmap
map_data_fetch_proxy = false
proxy = system
tomtom_map_api_key = 
tianditu_map_app_key = 
google_map_api_key = 
baidu_map_ak = 
amap_application_key = 
amap_security_verification_method = internal_proxy
amap_application_secret = 
amap_api_external_proxy_url = 
custom_map_tile_server_url = 
custom_map_tile_server_annotation_url = 
custom_map_tile_server_min_zoom_level = 1
custom_map_tile_server_max_zoom_level = 18
custom_map_tile_server_default_zoom_level = 14
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `map_provider` | string | `openstreetmap` | See the provider table below. Empty disables the map |
| `map_data_fetch_proxy` | bool | `false` | Fetch map data through the Honey Ledger server instead of directly from the browser. Supported for `openstreetmap`, `openstreetmap_humanitarian`, `opentopomap`, `opnvkarte`, `cyclosm`, `cartodb`, `tomtom`, `tianditu`, and `custom` |
| `proxy` | string | `system` | Proxy used when `map_data_fetch_proxy` is on: `system`, `none`, or a proxy URL starting with `http://`, `https://`, or `socks5://` |
| `tomtom_map_api_key` | string | *(empty)* | For `tomtom` only: API key |
| `tianditu_map_app_key` | string | *(empty)* | For `tianditu` only: application key |
| `google_map_api_key` | string | *(empty)* | For `googlemap` only: JavaScript API key |
| `baidu_map_ak` | string | *(empty)* | For `baidumap` only: JavaScript API application key |
| `amap_application_key` | string | *(empty)* | For `amap` only: JavaScript API application key |
| `amap_security_verification_method` | string | `internal_proxy` | For `amap` only: `internal_proxy`, `external_proxy`, or `plain_text` |
| `amap_application_secret` | string | *(empty)* | For `amap` only: application secret. Required for `internal_proxy` and `plain_text` |
| `amap_api_external_proxy_url` | string | *(empty)* | For `amap` only: external proxy URL. Required for `external_proxy` |
| `custom_map_tile_server_url` | string | *(empty)* | For `custom` only: tile layer URL with `{x}`, `{y}`, and `{z}` placeholders |
| `custom_map_tile_server_annotation_url` | string | *(empty)* | For `custom` only: optional annotation layer URL with the same placeholders |
| `custom_map_tile_server_min_zoom_level` | int | `1` | For `custom` only: minimum zoom level (0 - 255) |
| `custom_map_tile_server_max_zoom_level` | int | `18` | For `custom` only: maximum zoom level (0 - 255) |
| `custom_map_tile_server_default_zoom_level` | int | `14` | For `custom` only: zoom level used when a transaction has coordinates but no saved zoom |

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
| Tianditu | `tianditu` | `tianditu_map_app_key` |
| Google Maps | `googlemap` | `google_map_api_key` (public tiles work without one) |
| Baidu Map | `baidumap` | `baidu_map_ak` |
| Amap (Gaode) | `amap` | `amap_application_key`, `amap_security_verification_method`, plus `amap_application_secret` or `amap_api_external_proxy_url` depending on the method |
| Custom tile server | `custom` | `custom_map_tile_server_url` and the optional annotation layer and zoom levels |

---

## Exchange Rates

```ini
[exchange_rates]
data_source = euro_central_bank
request_timeout = 10000
proxy = system
skip_tls_verify = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `data_source` | string | `euro_central_bank` | Data source, see the table below |
| `request_timeout` | int | `10000` | Request timeout in milliseconds (0 - 4294967295), `0` disables the timeout |
| `proxy` | string | `system` | `system`, `none`, or a proxy URL starting with `http://`, `https://`, or `socks5://` |
| `skip_tls_verify` | bool | `false` | Skip TLS verification when fetching rates |

### Supported Exchange Rate Sources

| Source | Key |
|--------|-----|
| European Central Bank | `euro_central_bank` |
| Bank of Canada | `bank_of_canada` |
| Czech National Bank | `czech_national_bank` |
| Danmarks Nationalbank | `danmarks_national_bank` |
| National Bank of Georgia | `national_bank_of_georgia` |
| Central Bank of Hungary | `central_bank_of_hungary` |
| Bank of Israel | `bank_of_israel` |
| Central Bank of Myanmar | `central_bank_of_myanmar` |
| Norges Bank | `norges_bank` |
| National Bank of Poland | `national_bank_of_poland` |
| National Bank of Romania | `national_bank_of_romania` |
| Bank of Russia | `bank_of_russia` |
| Swiss National Bank | `swiss_national_bank` |
| National Bank of Ukraine | `national_bank_of_ukraine` |
| Central Bank of Uzbekistan | `central_bank_of_uzbekistan` |
| User custom | `user_custom` (rates entered in the app) |
