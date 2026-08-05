package whatsmeow_service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"go.mau.fi/whatsmeow/store/sqlstore"
	_ "modernc.org/sqlite"
)

func TestSQLStoreUpgradeCreatesPrivacyTokenAndNCTSaltTables(t *testing.T) {
	db, err := sql.Open("sqlite", "file:whatsmeow_schema_test?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	container := sqlstore.NewWithDB(db, "sqlite", nil)
	if err = container.Upgrade(context.Background()); err != nil {
		t.Fatalf("upgrade sqlstore schema: %v", err)
	}

	assertTableExists(t, db, "whatsmeow_privacy_tokens")
	assertTableExists(t, db, "whatsmeow_nct_salt")
	assertTableExists(t, db, "whatsmeow_lid_map")
	assertColumnExists(t, db, "whatsmeow_privacy_tokens", "sender_timestamp")
}

func assertTableExists(t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	var found string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&found)
	if err != nil {
		t.Fatalf("table %s not found: %v", tableName, err)
	}
}

func assertColumnExists(t *testing.T, db *sql.DB, tableName, columnName string) {
	t.Helper()

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		t.Fatalf("read columns for %s: %v", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		if err = rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &pk); err != nil {
			t.Fatalf("scan column for %s: %v", tableName, err)
		}
		if name == columnName {
			return
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("iterate columns for %s: %v", tableName, err)
	}
	t.Fatalf("column %s.%s not found", tableName, columnName)
}
