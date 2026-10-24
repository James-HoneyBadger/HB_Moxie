package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/ini.v1"
)

// AdminApp holds all application state
type AdminApp struct {
	fyneApp    fyne.App
	window     fyne.Window
	configPath string
	dbPath     string
	db         *sql.DB
}

// UserRecord represents a user row from the database
type UserRecord struct {
	Uid               int64
	Username          string
	Email             string
	Nickname          string
	DefaultCurrency   string
	Language          string
	Disabled          bool
	Deleted           bool
	EmailVerified     bool
	CreatedUnixTime   int64
	UpdatedUnixTime   int64
	LastLoginUnixTime int64
}

// TableInfo holds table name and row count
type TableInfo struct {
	Name     string
	RowCount int64
}

func resolveConfigPath(workDir string) string {
	candidates := []string{
		filepath.Join(workDir, "conf", "ratel-ledger.ini"),
		filepath.Join(workDir, "..", "conf", "ratel-ledger.ini"),
		filepath.Join(workDir, "conf", "hbmoxie.ini"),
		filepath.Join(workDir, "..", "conf", "hbmoxie.ini"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return candidates[0]
}

func main() {
	workDir, _ := os.Getwd()
	configPath := resolveConfigPath(workDir)

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, configPath)
	if err != nil {
		log.Fatalf("Cannot load config from %s: %v", configPath, err)
	}

	dbType := cfg.Section("database").Key("type").MustString("sqlite3")
	if dbType != "sqlite3" {
		log.Fatalf("Only sqlite3 supported (found: %s)", dbType)
	}

	dbPath := cfg.Section("database").Key("db_path").MustString("data/ratel-ledger.db")
	if !filepath.IsAbs(dbPath) {
		configDir := filepath.Dir(filepath.Dir(configPath))
		dbPath = filepath.Join(configDir, dbPath)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("Cannot open database: %v", err)
	}
	defer db.Close()

	a := &AdminApp{
		configPath: configPath,
		dbPath:     dbPath,
		db:         db,
	}

	a.fyneApp = app.NewWithID("org.honeybadger.ratelledger.admin")
	a.window = a.fyneApp.NewWindow("Ratel Ledger — System Administration")
	a.window.Resize(fyne.NewSize(1400, 850))
	a.window.CenterOnScreen()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Dashboard", theme.HomeIcon(), a.makeDashboard()),
		container.NewTabItemWithIcon("Configuration", theme.SettingsIcon(), a.makeConfigTab()),
		container.NewTabItemWithIcon("Users", theme.AccountIcon(), a.makeUsersTab()),
		container.NewTabItemWithIcon("Database", theme.StorageIcon(), a.makeDatabaseTab()),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	a.window.SetContent(tabs)
	a.window.ShowAndRun()
}

// --- Data access helpers ---

func (a *AdminApp) queryInt(query string) int64 {
	var count int64
	if err := a.db.QueryRow(query).Scan(&count); err != nil {
		log.Printf("queryInt failed for %q: %v", query, err)
		return 0
	}
	return count
}

func (a *AdminApp) loadUsers() []UserRecord {
	rows, err := a.db.Query(`
		SELECT uid, username, email, nickname, default_currency, language,
		       disabled, deleted, email_verified,
		       created_unix_time, updated_unix_time, last_login_unix_time
		FROM user ORDER BY uid ASC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var users []UserRecord
	for rows.Next() {
		var u UserRecord
		if err := rows.Scan(&u.Uid, &u.Username, &u.Email, &u.Nickname,
			&u.DefaultCurrency, &u.Language,
			&u.Disabled, &u.Deleted, &u.EmailVerified,
			&u.CreatedUnixTime, &u.UpdatedUnixTime, &u.LastLoginUnixTime); err != nil {
			log.Printf("loadUsers scan error: %v", err)
			continue
		}
		users = append(users, u)
	}
	return users
}

func (a *AdminApp) loadTables() []TableInfo {
	rows, err := a.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var name string
		rows.Scan(&name)
		var count int64
		a.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", name)).Scan(&count)
		tables = append(tables, TableInfo{Name: name, RowCount: count})
	}
	return tables
}

func (a *AdminApp) exportDatabase(w io.Writer) {
	fmt.Fprintf(w, "-- Ratel Ledger Database Export\n")
	fmt.Fprintf(w, "-- Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "-- Database: %s\n\n", a.dbPath)

	rows, err := a.db.Query("SELECT name, sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, createSQL string
		rows.Scan(&name, &createSQL)
		fmt.Fprintf(w, "-- Table: %s\n%s;\n\n", name, createSQL)

		dataRows, err := a.db.Query(fmt.Sprintf("SELECT * FROM `%s`", name))
		if err != nil {
			continue
		}
		cols, _ := dataRows.Columns()

		for dataRows.Next() {
			values := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range values {
				ptrs[i] = &values[i]
			}
			dataRows.Scan(ptrs...)

			var valStrs []string
			for _, v := range values {
				switch val := v.(type) {
				case nil:
					valStrs = append(valStrs, "NULL")
				case int64:
					valStrs = append(valStrs, fmt.Sprintf("%d", val))
				case float64:
					valStrs = append(valStrs, fmt.Sprintf("%g", val))
				case string:
					valStrs = append(valStrs, fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''")))
				case []byte:
					valStrs = append(valStrs, fmt.Sprintf("'%s'", strings.ReplaceAll(string(val), "'", "''")))
				default:
					valStrs = append(valStrs, fmt.Sprintf("'%v'", val))
				}
			}
			fmt.Fprintf(w, "INSERT INTO `%s` VALUES (%s);\n", name, strings.Join(valStrs, ", "))
		}
		dataRows.Close()
		fmt.Fprintln(w)
	}
}

// --- Utility functions ---

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatUnixTime(unix int64) string {
	if unix == 0 {
		return "Never"
	}
	return time.Unix(unix, 0).Format("2006-01-02 15:04:05")
}

func isPasswordKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "password") || strings.Contains(lower, "secret")
}

func scanRowStrings(rows *sql.Rows, colCount int) []string {
	values := make([]interface{}, colCount)
	ptrs := make([]interface{}, colCount)
	for i := range values {
		ptrs[i] = &values[i]
	}
	rows.Scan(ptrs...)

	result := make([]string, colCount)
	for i, v := range values {
		switch val := v.(type) {
		case nil:
			result[i] = "NULL"
		case int64:
			result[i] = fmt.Sprintf("%d", val)
		case float64:
			result[i] = fmt.Sprintf("%g", val)
		case string:
			result[i] = val
		case []byte:
			result[i] = string(val)
		default:
			result[i] = fmt.Sprintf("%v", val)
		}
	}
	return result
}
