package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const currentVersion = 1

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open creates a new database connection and runs any pending migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return d, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) migrate() error {
	var version int
	if err := d.conn.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}

	if version >= currentVersion {
		return nil
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("beginning migration transaction: %w", err)
	}
	defer tx.Rollback()

	if version < 1 {
		if err := migrateV1(tx); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", currentVersion)); err != nil {
		return fmt.Errorf("updating schema version: %w", err)
	}

	return tx.Commit()
}

func migrateV1(tx *sql.Tx) error {
	statements := []string{
		`CREATE TABLE media_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			media_type TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE media_labels (
			media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			label_id INTEGER NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
			PRIMARY KEY (media_file_id, label_id)
		)`,
		`CREATE TABLE keyframes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			timestamp_ms INTEGER NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			pinned INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE keyframe_labels (
			keyframe_id INTEGER NOT NULL REFERENCES keyframes(id) ON DELETE CASCADE,
			label_id INTEGER NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
			PRIMARY KEY (keyframe_id, label_id)
		)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration v1: %w", err)
		}
	}

	return nil
}
