# User Guide

A complete guide to using Honey Ledger for personal finance management.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Accounts](#accounts)
- [Transactions](#transactions)
- [Scheduled Transactions](#scheduled-transactions)
- [Transaction Templates](#transaction-templates)
- [Categories](#categories)
- [Tags](#tags)
- [Data Import](#data-import)
- [Data Export](#data-export)
- [Backups](#backups)
- [Statistics & Overview](#statistics--overview)
- [Insights Explorer](#insights-explorer)
- [Shared Expenses](#shared-expenses)
- [Exchange Rates](#exchange-rates)
- [AI Receipt Recognition](#ai-receipt-recognition)
- [Application Settings](#application-settings)
- [User Profile & Security](#user-profile--security)
- [Two-Factor Authentication](#two-factor-authentication)
- [OAuth2 / Single Sign-On](#oauth2--single-sign-on)
- [API Tokens](#api-tokens)
- [Cloud Settings Sync](#cloud-settings-sync)
- [Application Lock](#application-lock)
- [Keyboard Shortcuts](#keyboard-shortcuts)

---

## Getting Started

### First Login

1. Launch the Honey Ledger desktop application.
2. Log in with the default credentials:
   - **Username:** `admin`
   - **Password:** the random password printed to the console on first launch. It is shown once and is deliberately not written to the log file.
3. Change your password immediately via **Settings → Security**.

   If auto-login is enabled (the default) the login screen is skipped, so no password is needed. If you missed the one-time console output, set a new password from the command line:

   ```bash
   ./honey-ledger userdata user-modify-password --username admin --password "newpassword"
   ```

4. Set your preferred currency, timezone, and language in **Settings → Basic**.

### Initial Setup Checklist

- [ ] Change default password
- [ ] Set your default currency (Settings → Basic)
- [ ] Set your timezone (Settings → Basic)
- [ ] Create your accounts (bank accounts, credit cards, cash, etc.)
- [ ] Set up transaction categories or import the presets
- [ ] Optionally configure scheduled transactions for recurring bills

---

## Accounts

Accounts represent your bank accounts, credit cards, cash wallets, investment accounts, and any other financial containers.

### Account Types

| Type | Description |
|------|-------------|
| **Cash** | Physical cash |
| **Checking Account** | Bank checking/current account |
| **Savings Account** | Savings/deposit account |
| **Credit Card** | Credit card |
| **Virtual Account** | Digital wallet, gift card, prepaid card |
| **Debt Account** | Loans, mortgages |
| **Receivables** | Money owed to you |
| **Investment Account** | Stocks, bonds, investment portfolios |

### Creating an Account

1. Navigate to **Accounts** in the sidebar.
2. Click **Add Account**.
3. Fill in:
   - **Name** — descriptive name (e.g., "Chase Checking")
   - **Type** — select the account type
   - **Currency** — the account's currency
   - **Icon** — choose an icon
   - **Color** — pick a display color
   - **Balance** — current balance (optional, for initial setup)
   - **Comment** — optional notes
4. Click **Save**.

### Sub-Accounts

Accounts can have sub-accounts for organization (e.g., a "Bank" parent with "Checking" and "Savings" sub-accounts). Sub-accounts are created by selecting a parent account when adding a new account.

### Hiding & Reordering

- **Hide**: Right-click or use the menu to hide accounts you no longer use. Hidden accounts retain their data but don't appear in transaction forms.
- **Reorder**: Drag accounts to change their display order.

### Account Reconciliation

For bank accounts, you can perform reconciliation:

1. Open an account.
2. Click **Reconciliation Statement**.
3. Enter the statement balance and date.
4. Mark individual transactions as cleared.

---

## Transactions

Transactions are the core data unit — every movement of money is a transaction.

### Transaction Types

| Type | Description |
|------|-------------|
| **Expense** | Money spent from an account |
| **Income** | Money received into an account |
| **Transfer** | Money moved between your own accounts |

### Adding a Transaction

1. Click the **+** button in the navigation bar or go to **Transactions → Add**.
2. Select the transaction type (Expense, Income, or Transfer).
3. Fill in:
   - **Amount** — the transaction amount
   - **Account** — source account (and destination for transfers)
   - **Category** — transaction category
   - **Date/Time** — when the transaction occurred
   - **Tags** — optional tags for grouping
   - **Comment** — optional description
   - **Location** — optional geographic location (if maps are configured)
   - **Picture** — optional receipt photo attachment
4. Click **Save**.

### Editing & Deleting

- Click any transaction in the list to edit it.
- Use the delete button to remove a transaction.
- Bulk operations are available from the transaction list view.

### Transaction Pictures

If enabled by your administrator, you can attach pictures (receipts, invoices) to transactions:

1. When adding or editing a transaction, click the picture attachment area.
2. Upload an image file (max size configured by admin, default 512 KB).
3. Pictures are stored in the configured storage backend.

### Draft Auto-Save

Honey Ledger can automatically save transaction drafts as you type. Enable this in **Settings → Basic → Auto-save transaction draft**.

---

## Scheduled Transactions

Automate recurring transactions (rent, subscriptions, salary, etc.).

### Creating a Schedule

1. When adding a transaction, set it as a scheduled transaction.
2. Configure the recurrence pattern:
   - **Frequency** — daily, weekly, monthly, yearly
   - **Interval** — every N periods
   - **Start date** — when to begin
3. The system's cron job creates transactions automatically at the scheduled times.

### Managing Schedules

View and manage all scheduled transactions from the transaction list. You can pause, modify, or delete schedules.

---

## Transaction Templates

Save frequently-used transaction patterns as templates for quick entry.

### Creating a Template

1. Navigate to **Templates** in the sidebar.
2. Click **Add Template**.
3. Fill in the template fields (type, amount, account, category, tags, etc.).
4. Give the template a name and save.

### Using a Template

When adding a new transaction, you can select a template to pre-fill the form fields.

---

## Categories

Transaction categories organize your spending and income into a hierarchical structure.

### Category Structure

Categories have two levels:
- **Primary categories** (e.g., "Food & Dining", "Transportation")
- **Sub-categories** (e.g., "Groceries", "Restaurants" under "Food & Dining")

### Preset Categories

Honey Ledger includes preset category sets you can import during initial setup:
1. Go to **Categories**.
2. Click **Import Preset**.
3. Select a preset group and confirm.

### Custom Categories

1. Go to **Categories**.
2. Click **Add Category**.
3. Set name, icon, color, and parent category.
4. Reorder by dragging.

---

## Tags

Tags provide flexible cross-cutting labels for transactions, independent of categories.

### Tag Groups

Tags are organized into groups:
1. Go to **Tags** in the sidebar.
2. Create a **Tag Group** (e.g., "Projects", "People", "Trips").
3. Add tags within each group.

### Using Tags

When creating or editing a transaction, select one or more tags. Tags can be used as filters in the transaction list and statistics.

---

## Data Import

Import transactions from external financial software and bank exports.

### Supported Formats

| Category | Formats |
|----------|---------|
| **Honey Ledger** | CSV, TSV |
| **Custom** | CSV/TSV/SSV (configurable), Excel (.xlsx, .xls) |
| **Quicken** | QFX, QIF (Y/M/D, M/D/Y, D/M/Y date formats), IIF |
| **Microsoft Money** | OFX |
| **General Data Exchange** | OFX, QFX, QIF, IIF |
| **Bank Statements** | Camt.052 (XML), Camt.053 (XML), MT940 (SWIFT) |
| **Chinese Payment** | Alipay (App/Web CSV), WeChat Pay (Excel/CSV), JD.com Finance |
| **Finance Apps** | GnuCash XML, Firefly III CSV, Beancount, Feidee MyMoney |

### Import Process

1. Navigate to **Transactions**.
2. Click **Import** (toolbar button).
3. Select the file format from the dropdown.
4. Upload your file.
5. **For custom formats:** Define column mappings (which column maps to amount, date, category, etc.).
6. **Review:** Check the parsed transactions. Use batch replace to fix categories or accounts.
7. Click **Import** to save the transactions.

### Custom Import Scripts

For non-standard file formats, you can:
1. Upload the file.
2. Use the **Custom Script** tab to write a transformation script.
3. Map the transformed data to Honey Ledger's transaction fields.

### Import Tips

- Set the correct file encoding if auto-detection fails (70+ encodings supported).
- Use **Batch Replace** to fix account/category assignments before importing.
- Duplicate detection prevents importing the same transaction twice.

---

## Data Export

Export your transactions to standard file formats.

### Export Formats

| Format | Description |
|--------|-------------|
| **CSV** | Comma-separated values |
| **TSV** | Tab-separated values |

### How to Export

1. Open the **Statistics** or **Transactions** page.
2. Apply any filters (date range, accounts, categories).
3. Click **Export**.
4. Select the format and download.

---

## Backups

Create and manage full database backups directly from the application.

### Creating a Backup

1. Go to **Settings → Data Management**.
2. In the **Backups** card, click **Create Backup**.
3. Enter your account password to confirm.
4. The backup is created and appears in the backup list.

### Downloading a Backup

Click the download icon next to a backup to download it as a ZIP file.

### Restoring a Backup

1. In the backup list, click the restore icon for the desired backup.
2. Enter your account password.
3. **Warning:** Restoring a backup replaces all current data.

### Deleting a Backup

Click the delete icon to remove a backup file.

---

## Statistics & Overview

### Home Page Overview

The home page shows:
- **Monthly income and expense** summary
- **Account balances** overview
- Configurable filters for accounts and categories

### Statistics Page

Navigate to **Statistics** for detailed analysis:
- **Categorical charts** — pie/bar charts showing spending by category
- **Trend charts** — line/bar charts showing income/expense over time
- **Asset trends** — track net worth and account balances over time
- **Date range** — filter by custom date ranges
- **Account filters** — include/exclude specific accounts

---

## Insights Explorer

Build custom data queries and visualize results.

### Creating an Explorer

1. Navigate to **Insights** in the sidebar.
2. Click **Add Explorer**.
3. Configure:
   - **Query parameters** — date range, accounts, categories, tags
   - **Grouping** — how to aggregate data
   - **Visualization** — table view or chart view (bar, line, pie)
4. Save the explorer for reuse.

### Managing Explorers

Saved explorers appear in the Insights sidebar. You can reorder, edit, or delete them.

---

## Transaction Review

Reviews recent transactions for possible duplicate charges, unusual spending and
recurring subscriptions.

- **Dismiss** hides a finding you have already checked, and it stops reappearing
  without changing any transaction.
- Clicking the transaction details of a finding opens that transaction, so a real
  duplicate can be corrected straight away.
- The duplicate and unusual thresholds can be tuned through the API
  (`z_score`, `duplicate_max_bucket`) when the defaults surface too much noise or
  miss something.

---

## Shared Expenses

Track money other people owe you by splitting expense transactions with them.

### Splitting an Expense

1. Navigate to **Shared Expenses** in the sidebar.
2. Click **Split an Expense**.
3. Pick the expense transaction to split.
4. Enter each person's name and the amount they owe, using **Add Person** for
   more people (up to 20 per transaction).
5. Click **Save**.

The dialog shows the transaction amount and how much is still unallocated, so a
split can never add up to more than the transaction cost. Only expense
transactions can be split, and each person may appear once per transaction.

Splitting is also available from the transaction list: rows that are already
split show a chip with the amount other people owe, and other expense rows show
a split button next to the description. Both open the same editor.

To see only the transactions you share, click the people icon next to the
transaction list search box. Click it again to show every transaction; the
choice is kept in the page address, so a filtered list can be bookmarked.

### Tracking What Is Owed

The summary lists every person with the total they owe, grouped by currency, and
the date of the most recent shared expense. Click a person row to see the
individual expenses behind the total, where you can:

- **Edit** — change the people or amounts of that transaction.
- **Clear** — remove the split entirely.

Splits follow their transaction: deleting or editing a transaction removes it
from the summary. Each split records the account currency at creation time, so
totals in different currencies are never mixed.

---

## Exchange Rates

Honey Ledger automatically fetches exchange rates for multi-currency support.

### Automatic Updates

Exchange rates are fetched from the configured data source (default: European Central Bank) and cached locally. The cache expiration is configurable in application settings.

### Custom Exchange Rates

1. Navigate to **Exchange Rates** in the sidebar.
2. Click **Update** to manually set a custom rate between any two currencies.
3. Custom rates override the automatic rates.

### Viewing Rates

The exchange rate page shows all available currency pairs and their current rates, with sorting options.

---

## AI Receipt Recognition

Extract transaction data from receipt photographs using AI.

### Prerequisites

- An LLM provider must be configured by the administrator (see [Configuration Reference](CONFIGURATION.md)).
- The feature must be enabled in the configuration (`transaction_from_ai_image_recognition = true`).

### Usage

1. When adding a new transaction, click the **AI Recognition** button.
2. Upload or photograph a receipt.
3. The AI extracts: amount, date, merchant/description, and category.
4. Review and adjust the extracted data.
5. Save the transaction.

### Supported Providers

OpenAI (GPT-4o), Anthropic (Claude), Google AI (Gemini), OpenRouter, Ollama, LM Studio, and OpenAI/Anthropic-compatible APIs.

---

## Application Settings

Access via the gear icon or **Settings** in the sidebar.

### Basic Settings

| Setting | Description |
|---------|-------------|
| Theme | Light, Dark, or Auto (follows system) |
| Font Size | Adjust the UI font size |
| Timezone | Your local timezone for date display |
| Auto-update exchange rates | Fetch rates automatically on app open |
| Show account balance | Show/hide balances on the home page |
| Items per page | Number of transactions per page |
| Show total amount | Show sum in transaction list |
| Show tags | Show tags in transaction list |
| Auto-save draft | Save transaction form state automatically |
| Auto-get location | Request GPS location for new transactions |

### Statistics Settings

Configure default chart types, date ranges, sorting, and filter preferences for the statistics and overview pages.

### Local Cache

Manage the local map tile cache and exchange rate cache.

---

## User Profile & Security

### Profile

Edit your display name, email, avatar, and default currency via **Settings → Basic**.

### Change Password

1. Go to **Settings → Security**.
2. Enter your current password and new password.
3. Click **Save**.

### Sessions

View and revoke active login sessions from **Settings → Security**.

---

## Two-Factor Authentication

### Enabling 2FA

1. Go to **Settings → Two-Factor Authentication**.
2. Click **Enable**.
3. Scan the QR code with your authenticator app (Google Authenticator, Authy, etc.).
4. Enter the verification code to confirm.
5. **Save your recovery codes** — these are one-time-use codes in case you lose your authenticator.

### Logging In with 2FA

After entering your username and password, you'll be prompted for the 6-digit code from your authenticator app.

### Recovery

If you lose access to your authenticator:
1. Use a recovery code on the 2FA prompt.
2. Each recovery code can only be used once.
3. After logging in, regenerate recovery codes or disable 2FA.

### Disabling 2FA

Go to **Settings → Two-Factor Authentication** and click **Disable**.

---

## OAuth2 / Single Sign-On

If configured by the administrator, you can log in using an external identity provider.

### Supported Providers

- **OpenID Connect** (any OIDC provider)
- **GitHub**
- **Gitea**
- **Nextcloud**

### Linking an Account

1. Go to **Settings → Security**.
2. Click **Link** next to the OAuth2 provider.
3. Authorize the application in the provider's interface.
4. Your Honey Ledger account is now linked.

### Unlinking

Go to **Settings → Security** and click **Unlink** to disconnect the external provider.

---

## API Tokens

Generate tokens for programmatic access to the Honey Ledger API.

### Generating a Token

1. Go to **Settings → Security**.
2. Click **Generate API Token**.
3. Give the token a name.
4. **Copy the token immediately** — it won't be shown again.

### MCP Tokens

For MCP (Model Context Protocol) access, generate a dedicated MCP token from the same interface. MCP tokens allow AI agents to interact with your financial data via the `/mcp` endpoint.

### Revoking Tokens

View all active tokens in **Settings → Security** and click **Revoke** to deactivate any token.

---

## Cloud Settings Sync

Sync your application preferences across multiple sessions or devices.

### How It Works

1. Go to **Settings → Cloud Sync**.
2. Toggle the settings you want to sync.
3. Synced settings are stored locally and applied automatically when you log in.

### Syncable Settings

Application theme, font size, timezone, chart preferences, filter settings, display options, and more.

---

## Application Lock

Protect the application with a PIN code or WebAuthn (fingerprint/security key).

### Setting Up a Lock

1. Go to **Settings → Lock**.
2. Set a PIN code.
3. Optionally enable WebAuthn for biometric/hardware key unlock.

### Using the Lock

When the lock is active, you'll need to enter your PIN or use WebAuthn after a period of inactivity.

---

## Navigation Shortcuts

The desktop interface supports the following navigation shortcuts:

| Shortcut | Action |
|----------|--------|
| Sidebar items | Navigate between sections |
| Tab / Shift+Tab | Move between form fields |
| Enter | Confirm dialogs and forms |
| Escape | Close dialogs |
