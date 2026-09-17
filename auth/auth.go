package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-cli/db"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	MaxFailedAttempts = 3
	LockoutDuration   = 5 * time.Minute
)

func Register(username, password string) error {
	_, err := db.GetUserByUsername(username)
	if err == nil {
		return errors.New("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	return db.CreateUser(username, string(hashed))
}

func Login(username, password string, promptCode func() string) (*db.User, error) {
	u, err := db.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if u.LockedUntil.Valid && time.Now().Before(u.LockedUntil.Time) {
		return nil, fmt.Errorf("account locked until %v", u.LockedUntil.Time.Format(time.RFC1123))
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		u.FailedAttempts++
		if u.FailedAttempts >= MaxFailedAttempts {
			u.LockedUntil = sql.NullTime{Time: time.Now().Add(LockoutDuration), Valid: true}
			db.UpdateUser(u)
			return nil, fmt.Errorf("account locked due to too many failed attempts")
		}
		db.UpdateUser(u)
		return nil, errors.New("invalid credentials")
	}

	if u.TotpEnabled {
		code := promptCode()
		valid := totp.Validate(code, u.TotpSecret)
		if !valid {
			return nil, errors.New("invalid 2FA code")
		}
	}

	u.FailedAttempts = 0
	u.LockedUntil = sql.NullTime{Valid: false}
	u.LastLogin = sql.NullTime{Time: time.Now(), Valid: true}
	db.UpdateUser(u)

	return u, nil
}

func Enable2FA(u *db.User, promptCode func(string) string) error {
	if u.TotpEnabled {
		return errors.New("2FA is already enabled")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SecureLoginCLI",
		AccountName: u.Username,
	})
	if err != nil {
		return err
	}

	// In a real CLI without GUI, we display the secret and a URL
	fmt.Printf("Scan this URL with Google Authenticator:\n%s\n", key.URL())
	fmt.Printf("Or manually enter this secret: %s\n", key.Secret())

	code := promptCode(key.Secret())
	valid := totp.Validate(code, key.Secret())
	if !valid {
		return errors.New("invalid code, 2FA not enabled")
	}

	u.TotpSecret = key.Secret()
	u.TotpEnabled = true
	return db.UpdateUser(u)
}

func Disable2FA(u *db.User) error {
	if !u.TotpEnabled {
		return errors.New("2FA is not enabled")
	}

	u.TotpSecret = ""
	u.TotpEnabled = false
	return db.UpdateUser(u)
}
