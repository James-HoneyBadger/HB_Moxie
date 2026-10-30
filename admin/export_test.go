package main

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestAdminApp(t *testing.T) *AdminApp {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "ratel-ledger.db")

	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	_, err = db.Exec(`
		CREATE TABLE user (uid INTEGER PRIMARY KEY, username TEXT);
		INSERT INTO user (uid, username) VALUES (1, 'alice'), (2, 'o''brien');
		CREATE TABLE account (account_id INTEGER PRIMARY KEY, name TEXT);
	`)

	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	return &AdminApp{
		db:     db,
		dbPath: dbPath,
	}
}

func TestExportDatabaseWritesSchemaAndRows(t *testing.T) {
	app := newTestAdminApp(t)

	var buffer bytes.Buffer

	if err := app.exportDatabase(&buffer); err != nil {
		t.Fatalf("exportDatabase returned an error: %v", err)
	}

	output := buffer.String()

	expectedFragments := []string{
		"-- Table: account",
		"-- Table: user",
		"CREATE TABLE user (uid INTEGER PRIMARY KEY, username TEXT)",
		"INSERT INTO `user` VALUES (1, 'alice');",
		// Single quotes must be escaped so the dump stays valid SQL.
		"INSERT INTO `user` VALUES (2, 'o''brien');",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(output, fragment) {
			t.Errorf("export output does not contain %q\n--- output ---\n%s", fragment, output)
		}
	}
}

func TestExportDatabaseReturnsErrorWhenQueryFails(t *testing.T) {
	app := newTestAdminApp(t)

	// Dropping the connection makes every subsequent query fail.
	if err := app.db.Close(); err != nil {
		t.Fatalf("close test database: %v", err)
	}

	var buffer bytes.Buffer

	if err := app.exportDatabase(&buffer); err == nil {
		t.Fatal("expected exportDatabase to return an error after the database was closed")
	}
}
