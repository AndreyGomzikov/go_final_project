package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat TEXT NOT NULL DEFAULT ''
);
`

func Open(path string) error {
	install := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		install = true
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping sqlite: %w", err)
	}
	if install {
		if _, err := db.Exec(schema); err != nil {
			_ = db.Close()
			return fmt.Errorf("apply schema: %w", err)
		}
	}
	
	var n int
	_ = db.QueryRow("SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name='scheduler'").Scan(&n)
	if n == 0 {
		if _, err := db.Exec(schema); err != nil {
			_ = db.Close()
			return fmt.Errorf("apply schema: %w", err)
		}
	}
DB = db
	return nil
}
