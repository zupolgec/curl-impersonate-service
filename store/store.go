// Package store provides SQLite-backed persistence for API tokens, settings
// (such as CORS origins) and request usage logs.
package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database.
type Store struct {
	db *sql.DB
}

// Token is an API token record.
type Token struct {
	ID         int64
	Name       string
	Token      string
	Enabled    bool
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

// LogEntry is a single usage-log record.
type LogEntry struct {
	ID         int64
	TS         time.Time
	TokenName  string
	Browser    string
	Method     string
	TargetHost string
	StatusCode int
	Success    bool
	DurationMs int64
	ErrorType  string
}

const schema = `
CREATE TABLE IF NOT EXISTS api_tokens (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT    NOT NULL,
    token        TEXT    NOT NULL UNIQUE,
    enabled      INTEGER NOT NULL DEFAULT 1,
    created_at   INTEGER NOT NULL,
    last_used_at INTEGER
);
CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS usage_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    ts          INTEGER NOT NULL,
    token_name  TEXT,
    browser     TEXT,
    method      TEXT,
    target_host TEXT,
    status_code INTEGER,
    success     INTEGER,
    duration_ms INTEGER,
    error_type  TEXT
);
CREATE INDEX IF NOT EXISTS idx_usage_logs_ts ON usage_logs(ts);
`

// Open opens (and migrates) the SQLite database at path.
func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: serialize writes to avoid "database is locked"
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// generateToken returns a cryptographically random token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateToken creates a new random API token with the given name.
func (s *Store) CreateToken(name string) (*Token, error) {
	tok, err := generateToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO api_tokens (name, token, enabled, created_at) VALUES (?, ?, 1, ?)`,
		name, tok, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Token{ID: id, Name: name, Token: tok, Enabled: true, CreatedAt: time.Unix(now, 0)}, nil
}

// SeedToken inserts a token with an explicit value if it does not already
// exist. Used to preserve the legacy TOKEN env value across restarts.
func (s *Store) SeedToken(name, value string) error {
	var exists int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM api_tokens WHERE token = ?`, value).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	_, err := s.db.Exec(
		`INSERT INTO api_tokens (name, token, enabled, created_at) VALUES (?, ?, 1, ?)`,
		name, value, time.Now().Unix(),
	)
	return err
}

// ValidateToken returns the token name if the value matches an enabled token,
// updating its last-used timestamp.
func (s *Store) ValidateToken(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	var name string
	err := s.db.QueryRow(
		`SELECT name FROM api_tokens WHERE token = ? AND enabled = 1`, value,
	).Scan(&name)
	if err != nil {
		return "", false
	}
	_, _ = s.db.Exec(`UPDATE api_tokens SET last_used_at = ? WHERE token = ?`, time.Now().Unix(), value)
	return name, true
}

// ListTokens returns all API tokens ordered by creation time.
func (s *Store) ListTokens() ([]Token, error) {
	rows, err := s.db.Query(
		`SELECT id, name, token, enabled, created_at, last_used_at FROM api_tokens ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Token
	for rows.Next() {
		var t Token
		var enabled int
		var created int64
		var lastUsed sql.NullInt64
		if err := rows.Scan(&t.ID, &t.Name, &t.Token, &enabled, &created, &lastUsed); err != nil {
			return nil, err
		}
		t.Enabled = enabled == 1
		t.CreatedAt = time.Unix(created, 0)
		if lastUsed.Valid {
			u := time.Unix(lastUsed.Int64, 0)
			t.LastUsedAt = &u
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// DeleteToken removes a token by id.
func (s *Store) DeleteToken(id int64) error {
	_, err := s.db.Exec(`DELETE FROM api_tokens WHERE id = ?`, id)
	return err
}

// SetTokenEnabled enables or disables a token by id.
func (s *Store) SetTokenEnabled(id int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec(`UPDATE api_tokens SET enabled = ? WHERE id = ?`, v, id)
	return err
}

// GetSetting returns a setting value, or def if unset.
func (s *Store) GetSetting(key, def string) string {
	var v string
	if err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v); err != nil {
		return def
	}
	return v
}

// SetSetting upserts a setting value.
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// AddLog inserts a usage-log entry. The timestamp is set to now.
func (s *Store) AddLog(e LogEntry) error {
	success := 0
	if e.Success {
		success = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO usage_logs (ts, token_name, browser, method, target_host, status_code, success, duration_ms, error_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now().Unix(), e.TokenName, e.Browser, e.Method, e.TargetHost,
		e.StatusCode, success, e.DurationMs, e.ErrorType,
	)
	return err
}

// ListLogs returns the most recent usage logs, up to limit.
func (s *Store) ListLogs(limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, ts, token_name, browser, method, target_host, status_code, success, duration_ms, error_type
		 FROM usage_logs ORDER BY ts DESC, id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []LogEntry
	for rows.Next() {
		var e LogEntry
		var ts int64
		var success int
		if err := rows.Scan(&e.ID, &ts, &e.TokenName, &e.Browser, &e.Method, &e.TargetHost,
			&e.StatusCode, &success, &e.DurationMs, &e.ErrorType); err != nil {
			return nil, err
		}
		e.TS = time.Unix(ts, 0)
		e.Success = success == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

// PurgeLogsOlderThan deletes usage logs older than the given duration and
// returns the number of rows removed.
func (s *Store) PurgeLogsOlderThan(d time.Duration) (int64, error) {
	cutoff := time.Now().Add(-d).Unix()
	res, err := s.db.Exec(`DELETE FROM usage_logs WHERE ts < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}
