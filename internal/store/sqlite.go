package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"memorialcandle/internal/memorial"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func Open(path string) (*SQLiteStore, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &SQLiteStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func OpenMemory() (*SQLiteStore, error) {
	return Open("file:memorial?mode=memory&cache=shared")
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS memorials (id TEXT PRIMARY KEY, title TEXT NOT NULL, dedication TEXT NOT NULL, created_at TEXT NOT NULL, candle_count INTEGER NOT NULL, quiet INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS candles (id TEXT PRIMARY KEY, memorial_id TEXT NOT NULL, visitor_id TEXT NOT NULL, message TEXT NOT NULL, color TEXT NOT NULL, intensity INTEGER NOT NULL, lit_at TEXT NOT NULL, extinguished INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS sparks (id TEXT PRIMARY KEY, candle_id TEXT NOT NULL, sequence INTEGER NOT NULL, angle REAL NOT NULL, radius REAL NOT NULL, brightness REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS animation_jobs (id TEXT PRIMARY KEY, candle_id TEXT NOT NULL, status TEXT NOT NULL, generated INTEGER NOT NULL, started_at TEXT NOT NULL, stopped_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS sessions (id TEXT PRIMARY KEY, memorial_id TEXT NOT NULL, started_at TEXT NOT NULL, last_seen_at TEXT NOT NULL, angle REAL NOT NULL, zoom REAL NOT NULL, active INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS reflections (id TEXT PRIMARY KEY, memorial_id TEXT NOT NULL, visitor_id TEXT NOT NULL, text TEXT NOT NULL, moderated INTEGER NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS audits (id TEXT PRIMARY KEY, memorial_id TEXT NOT NULL, action TEXT NOT NULL, subject TEXT NOT NULL, detail TEXT NOT NULL, created_at TEXT NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func parseStamp(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intBool(value int) bool { return value != 0 }

func missing(kind, id string) error { return fmt.Errorf("%s %s: %w", kind, id, memorial.ErrNotFound) }
