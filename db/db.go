package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

type User struct {
	ID             int
	Username       string
	PasswordHash   string
	TotpSecret     string
	TotpEnabled    bool
	CreatedAt      time.Time
	LastLogin      sql.NullTime
	FailedAttempts int
	LockedUntil    sql.NullTime
}

func InitDB(dbPath string) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create db directory: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTables()
}

func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT,
		totp_enabled BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME,
		failed_attempts INTEGER DEFAULT 0,
		locked_until DATETIME
	);
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func GetUserByUsername(username string) (*User, error) {
	u := &User{}
	var lastLogin, lockedUntil sql.NullTime
	var createdAt string
	
	row := DB.QueryRow("SELECT id, username, password_hash, totp_secret, totp_enabled, created_at, last_login, failed_attempts, locked_until FROM users WHERE username = ?", username)
	
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.TotpSecret, &u.TotpEnabled, &createdAt, &lastLogin, &u.FailedAttempts, &lockedUntil)
	if err != nil {
		return nil, err
	}
	
	// parse time
	if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		u.CreatedAt = t
	}
	
	u.LastLogin = lastLogin
	u.LockedUntil = lockedUntil
	
	return u, nil
}

func CreateUser(username, passwordHash string) error {
	_, err := DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	return err
}

func UpdateUser(u *User) error {
	_, err := DB.Exec(`
		UPDATE users 
		SET password_hash = ?, totp_secret = ?, totp_enabled = ?, last_login = ?, failed_attempts = ?, locked_until = ?
		WHERE id = ?`,
		u.PasswordHash, u.TotpSecret, u.TotpEnabled, u.LastLogin, u.FailedAttempts, u.LockedUntil, u.ID,
	)
	return err
}
