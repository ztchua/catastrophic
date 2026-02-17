# Catastrophic

A Telegram bot for tracking pet poop records and expenses with SQLite storage.

[![CI](https://github.com/ztchua/catastrophic/actions/workflows/ci.yml/badge.svg)](https://github.com/ztchua/catastrophic/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ztchua/catastrophic)](https://goreportcard.com/report/github.com/ztchua/catastrophic)
[![GoDoc](https://godoc.org/github.com/ztchua/catastrophic?status.svg)](https://godoc.org/github.com/ztchua/catastrophic)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ztchua/catastrophic)](https://go.dev/)

## Features

- Record pet poop events with texture descriptions
- View all records or the 10 most recent
- Check if pet hasn't pooped in 3+ days (health warning)
- Update records (datetime or texture)
- Delete records
- Track cat expenses with categories (food, litter, vet, toys, etc.)
- View total spent this month
- Filter expenses by category for past 30 days
- Per-chat isolation of records
- Private bot - only allowed users can interact

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Start the bot |
| `/help` | Show available commands |
| `/ping` | Check if bot is alive |
| `/poop_add` | Add a new poop record |
| `/poop_list` | List all records |
| `/poop_recent` | Show 10 most recent records |
| `/poop_check` | Check if pet hasn't pooped in 3+ days |
| `/poop_update <id>` | Update a record |
| `/poop_delete <id>` | Delete a record |
| `/expense_add` | Add a new expense |
| `/expense_list` | List all expenses |
| `/expense_month` | Show total spent this month |
| `/expense_category <category>` | Filter by category (past 30 days) |
| `/expense_update <id>` | Update an expense |
| `/expense_delete <id>` | Delete an expense |

## Setup

### Prerequisites

- Go 1.23+
- SQLite3

### Installation

```bash
# Clone the repository
git clone https://github.com/ztchua/catastrophic.git
cd catastrophic

# Download dependencies
go mod download

# Build
go build -o catastrophic .
```

### Configuration

1. Create a bot token from [@BotFather](https://t.me/BotFather)

2. Create `allowed_users.cfg` with allowed Telegram usernames:
   ```
   cp allowed_users.cfg.example allowed_users.cfg
   # Edit with your username(s)
   ```

3. Run the bot:
   ```bash
   TELEGRAM_BOT_TOKEN=your_token ./catastrophic
   ```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | Yes | - | Telegram bot token from @BotFather |
| `POOP_DB_PATH` | No | `poop.db` | Path to SQLite database file for poop records |
| `EXPENSE_DB_PATH` | No | `expense.db` | Path to SQLite database file for expense records |
| `ALLOWED_USERS_PATH` | No | `allowed_users.cfg` | Path to allowed users config |

## Development

### Build

```bash
go build -o catastrophic .
```

### Test

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Lint

```bash
# Static analysis
go vet ./...

# Format check
gofmt -l .

# Format code
gofmt -w .
```

## License

MIT
