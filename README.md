# Secure CLI Login System

A containerized command-line login system written in Go. It supports user registration, secure authentication (bcrypt), optional TOTP-based 2FA (Google Authenticator compatible), account lockouts after multiple failed attempts, and session management.

## Requirements

- Docker
- Docker Compose

## Setup & Running

1. **Build and start the container in the background:**
   ```bash
   docker-compose up -d --build
   ```

2. **Attach to the interactive CLI:**
   Since this is an interactive CLI app, you need to attach to the container's standard input/output:
   ```bash
   docker attach secure-cli
   ```
   *(To detach without stopping the container, press `Ctrl+P`, then `Ctrl+Q`)*

   Alternatively, you can just run it temporarily with:
   ```bash
   docker-compose run --rm cli
   ```

## Usage

### Before Login
- `register`: Create a new user account.
- `login`: Log in with your username and password. If 2FA is enabled, you'll be prompted for a TOTP code.
- `help`: Show available commands.
- `exit`: Quit the program.

### After Login
- `whoami`: Show current user details (username, registration date, MFA status, last login, and session expiration).
- `enable-2fa`: Enable TOTP-based Multi-Factor Authentication. It will display a secret you can enter into Google Authenticator or Authy.
- `disable-2fa`: Disable MFA.
- `logout`: End your current session.
- `help`: Show available commands.

## Architecture & Security

- **Database**: SQLite (via `modernc.org/sqlite` pure-Go driver to ensure seamless cross-compilation and easy Dockerization). Data is persisted in the `./data` directory on your host machine.
- **Passwords**: Hashed securely using `bcrypt`.
- **2FA**: TOTP using `github.com/pquerna/otp`.
- **Lockout**: Accounts are locked for 5 minutes after 3 consecutive failed login attempts.
- **Sessions**: The CLI maintains an active session memory block that expires after 15 minutes of inactivity.
- **UI**: Interactive prompt implemented using `github.com/chzyer/readline`, providing history (up/down arrows) and robust input handling.
