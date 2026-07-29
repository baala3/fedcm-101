package idp

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL,
	given_name TEXT NOT NULL,
	email TEXT NOT NULL,
	picture_url TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS grants (
	user_id INTEGER NOT NULL,
	client_id TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, client_id)
);
`

// OpenDB opens (creating if needed) the sqlite file at path, applies the
// schema, and seeds two demo users the first time the users table is empty.
func OpenDB(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := seedUsers(db); err != nil {
		return nil, fmt.Errorf("seed users: %w", err)
	}
	return db, nil
}

func seedUsers(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	seeds := []struct {
		username, password, name, givenName, email string
	}{
		{"alice", "password123", "Alice Adams", "Alice", "alice@idp-demo.local"},
		{"bob", "password123", "Bob Baker", "Bob", "bob@idp-demo.local"},
	}
	for _, s := range seeds {
		hash, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec(
			`INSERT INTO users (username, password_hash, name, given_name, email, picture_url) VALUES (?, ?, ?, ?, ?, ?)`,
			s.username, string(hash), s.name, s.givenName, s.email,
			fmt.Sprintf("https://api.dicebear.com/7.x/identicon/svg?seed=%s", s.username),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
