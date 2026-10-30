package models

// SchemaMigration records a database migration that has been applied.
// Migration ids are ordered lexicographically, so they carry a UTC date prefix
// and a sequence number: 20261030_01_shared_expenses.
type SchemaMigration struct {
	MigrationId     string `xorm:"PK VARCHAR(255) NOT NULL"`
	AppliedUnixTime int64  `xorm:"NOT NULL"`
}
