package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go-cli/auth"
	"go-cli/db"

	"github.com/chzyer/readline"
)

var currentUser *db.User
var lastActivity time.Time
const SessionTimeout = 15 * time.Minute

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/cli.db"
	}
	db.InitDB(dbPath)
	defer db.DB.Close()

	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m»\033[0m ",
		HistoryFile:       "/tmp/readline.tmp",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	fmt.Println("Welcome to Secure CLI Login System!")
	fmt.Println("Type 'help' to see available commands.")

	for {
		if currentUser != nil {
			if time.Since(lastActivity) > SessionTimeout {
				fmt.Println("\n[!] Session expired due to inactivity. Please log in again.")
				currentUser = nil
			} else {
				l.SetPrompt(fmt.Sprintf("\033[32m%s»\033[0m ", currentUser.Username))
			}
		} else {
			l.SetPrompt("\033[31m»\033[0m ")
		}

		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lastActivity = time.Now()
		args := strings.Split(line, " ")
		cmd := args[0]

		if cmd == "exit" {
			break
		}

		if currentUser == nil {
			handleUnauthCmd(cmd, l)
		} else {
			handleAuthCmd(cmd, l)
		}
	}
}

func handleUnauthCmd(cmd string, l *readline.Instance) {
	switch cmd {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("  register - create a new user")
		fmt.Println("  login    - login with username/password")
		fmt.Println("  help     - show available commands")
		fmt.Println("  exit     - quit program")
	case "register":
		username := promptStr(l, "Username: ")
		password := promptPassword(l, "Password: ")
		
		err := auth.Register(username, password)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("User registered successfully! You can now log in.")
		}
	case "login":
		username := promptStr(l, "Username: ")
		password := promptPassword(l, "Password: ")
		
		u, err := auth.Login(username, password, func() string {
			return promptStr(l, "Enter 2FA code: ")
		})
		
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			currentUser = u
			lastActivity = time.Now()
			fmt.Printf("Welcome back, %s!\n", u.Username)
			printUserDetails(u)
		}
	default:
		fmt.Println("Unknown command. Type 'help' for available commands.")
	}
}

func handleAuthCmd(cmd string, l *readline.Instance) {
	switch cmd {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("  whoami      - show current user details")
		fmt.Println("  enable-2fa  - enable TOTP-based MFA")
		fmt.Println("  disable-2fa - disable MFA")
		fmt.Println("  logout      - end session")
		fmt.Println("  help        - show available commands")
		fmt.Println("  exit        - quit program")
	case "whoami":
		printUserDetails(currentUser)
	case "enable-2fa":
		err := auth.Enable2FA(currentUser, func(secret string) string {
			return promptStr(l, "Enter the 6-digit code to verify: ")
		})
		if err != nil {
			fmt.Printf("Failed to enable 2FA: %v\n", err)
		} else {
			fmt.Println("2FA successfully enabled!")
		}
	case "disable-2fa":
		err := auth.Disable2FA(currentUser)
		if err != nil {
			fmt.Printf("Failed to disable 2FA: %v\n", err)
		} else {
			fmt.Println("2FA successfully disabled.")
		}
	case "logout":
		currentUser = nil
		fmt.Println("Logged out successfully.")
	default:
		fmt.Println("Unknown command. Type 'help' for available commands.")
	}
}

func printUserDetails(u *db.User) {
	fmt.Println("--- User Details ---")
	fmt.Printf("Username: %s\n", u.Username)
	fmt.Printf("Registered: %s\n", u.CreatedAt.Format(time.RFC1123))
	mfaStatus := "Disabled"
	if u.TotpEnabled {
		mfaStatus = "Enabled"
	}
	fmt.Printf("MFA Status: %s\n", mfaStatus)
	
	if u.LastLogin.Valid {
		fmt.Printf("Last Login: %s\n", u.LastLogin.Time.Format(time.RFC1123))
	}
	fmt.Printf("Session Expires: %s\n", lastActivity.Add(SessionTimeout).Format(time.Kitchen))
	fmt.Println("--------------------")
}

func promptStr(l *readline.Instance, prompt string) string {
	l.SetPrompt(prompt)
	defer l.SetPrompt("")
	line, err := l.Readline()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

func promptPassword(l *readline.Instance, prompt string) string {
	cfg := l.Config.Clone()
	cfg.EnableMask = true
	cfg.MaskRune = '*'
	l.SetConfig(cfg)
	defer func() {
		cfg.EnableMask = false
		l.SetConfig(cfg)
	}()
	
	l.SetPrompt(prompt)
	line, err := l.Readline()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}
