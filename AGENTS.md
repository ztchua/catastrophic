# AGENTS.md

Coding agent guidelines for this repository.

## !IMPORTANT

Do not add additional rules to this file, only provide suggestions if necessary.

## Project Overview

This is a Telegram bot server written in Go for tracking pet poop records. It uses the `telegram-bot-api/v5` library to interact with the Telegram Bot API and SQLite for persistent storage.

## Build Commands

```bash
# Build the binary
go build -o catastrophic .

# Run the server (requires TELEGRAM_BOT_TOKEN env var)
TELEGRAM_BOT_TOKEN=your_token ./catastrophic

# Run directly without building
TELEGRAM_BOT_TOKEN=your_token go run .
```

## Lint Commands

```bash
# Run go vet (static analysis)
go vet ./...

# Run gofmt check (formatting)
gofmt -l .

# Format code
gofmt -w .

# Run golangci-lint (if installed)
golangci-lint run
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test -v ./...

# Run a single test by name
go test -v -run TestFunctionName ./...

# Run a single test in a specific file
go test -v -run TestFunctionName ./path/to/package

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Dependency Management

```bash
# Download dependencies
go mod download

# Tidy dependencies (add missing, remove unused)
go mod tidy

# Verify dependencies
go mod verify
```

## Code Style Guidelines

### Imports

- Use standard library imports first, separated by a blank line
- Third-party imports come second, separated by a blank line
- Local imports come last
- Use import aliases sparingly, only when necessary to avoid conflicts

```go
import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
```

### Formatting

- Always run `gofmt -w .` before committing
- Use tabs for indentation (Go standard)
- Line length should be reasonable, but Go has no strict limit

### Naming Conventions

- Use CamelCase (not snake_case)
- Exported names start with uppercase
- Private names start with lowercase
- Interfaces: typically end with `-er` (e.g., `Handler`, `Reader`)
- Acronyms should be consistent: `URL`, `HTTP`, `ID` (not `Url`, `Http`, `Id`)

### Error Handling

- Always handle errors explicitly
- Never ignore errors with `_`
- Use `log.Fatal` or `log.Fatalf` for startup errors
- Use `log.Printf` for runtime errors that shouldn't crash the app
- Return errors from functions when appropriate

```go
if err != nil {
	log.Printf("Failed to send message: %v", err)
	return err
}
```

### Types

- Prefer `string.Builder` for efficient string concatenation
- Use `fmt.Sprintf` for formatted strings
- Prefer concrete types over interfaces when possible
- Define types for domain concepts

### Functions

- Keep functions focused on a single responsibility
- Return early to reduce nesting
- Use descriptive function names that indicate what they do

### Comments

- Comments should explain "why", not "what"
- Package comments should start with "Package <name>"
- Exported functions should have doc comments
- Avoid redundant comments that just repeat the code

### Project Structure

For larger projects, follow standard Go project layout:

```
/cmd/           # Main applications
/internal/      # Private application code
/pkg/           # Public library code
/go.mod         # Module definition
/go.sum         # Dependency checksums
```

## Environment Variables

- `TELEGRAM_BOT_TOKEN`: Required. Your Telegram bot token from @BotFather.
- `POOP_DB_PATH`: Optional. Path to SQLite database file. Defaults to `poop.db` in current directory.

## Key Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5`: Telegram Bot API client
- `github.com/mattn/go-sqlite3`: SQLite driver for Go

## Agent Behavior

- Run `go build` to verify code compiles after changes
- Run `go vet ./...` to check for common errors
- Run `go mod tidy` after adding new imports
- Run `go test ./...` to ensure all tests pass before committing
- Never commit the binary (`catastrophic`)
- Ensure TELEGRAM_BOT_TOKEN is never hardcoded or committed
