package idp

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// User is a demo IdP account.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Name         string
	GivenName    string
	Email        string
	Picture      string
}

// Store wraps the sqlite database with the queries the IdP handlers need.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

var ErrInvalidCredentials = errors.New("invalid username or password")

// Authenticate checks a username/password pair against the seeded demo users.
func (s *Store) Authenticate(username, password string) (*User, error) {
	u, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	return s.scanUser(s.db.QueryRow(
		`SELECT id, username, password_hash, name, given_name, email, picture_url FROM users WHERE username = ?`,
		username,
	))
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	return s.scanUser(s.db.QueryRow(
		`SELECT id, username, password_hash, name, given_name, email, picture_url FROM users WHERE id = ?`,
		id,
	))
}

func (s *Store) scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Name, &u.GivenName, &u.Email, &u.Picture)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// HasGrant reports whether userID has previously consented to sharing their
// profile with clientID (used to populate approved_clients so the browser
// can skip the re-consent dialog on return visits).
func (s *Store) HasGrant(userID int64, clientID string) (bool, error) {
	var exists int
	err := s.db.QueryRow(`SELECT 1 FROM grants WHERE user_id = ? AND client_id = ?`, userID, clientID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// ClientsGrantedFor returns every client_id userID has an active grant for.
func (s *Store) ClientsGrantedFor(userID int64) ([]string, error) {
	rows, err := s.db.Query(`SELECT client_id FROM grants WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var clients []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

// AddGrant records consent for userID sharing their profile with clientID.
func (s *Store) AddGrant(userID int64, clientID string) error {
	_, err := s.db.Exec(
		`INSERT INTO grants (user_id, client_id) VALUES (?, ?) ON CONFLICT (user_id, client_id) DO NOTHING`,
		userID, clientID,
	)
	return err
}

// RemoveGrant revokes consent, e.g. when the RP calls IdentityCredential.disconnect().
// It reports whether a grant existed and was removed.
func (s *Store) RemoveGrant(userID int64, clientID string) (bool, error) {
	res, err := s.db.Exec(`DELETE FROM grants WHERE user_id = ? AND client_id = ?`, userID, clientID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
