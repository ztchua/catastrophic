# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Telegram bot for tracking pet poop records and expenses with SQLite storage. Uses `telegram-bot-api/v5` for Telegram integration and `go-sqlite3` for persistence.

## Development Commands

```bash
# Build
go build -o catastrophic .

# Run (requires TELEGRAM_BOT_TOKEN)
TELEGRAM_BOT_TOKEN=your_token ./catastrophic

# Test all
go test -v ./...

# Test single function
go test -v -run TestFunctionName ./...

# Lint
go vet ./...
gofmt -w .
```

## Architecture

**Single package (`main`)** with these core components:

- **main.go** - Entry point, initializes stores, auth, and starts update loop
- **bot.go** - `Bot` struct handles Telegram updates via state machine pattern
- **poop.go** - `PoopStore` for poop record CRUD with SQLite
- **expense.go** - `ExpenseStore` for expense tracking with SQLite
- **auth.go** - `AuthService` validates users against `allowed_users.cfg` and groups against `allowed_groups.cfg`

**Key patterns:**

- **Per-chat isolation**: All data queries include `chat_id` to isolate records between Telegram chats
- **Conversation state machine**: `Bot.sessions` tracks multi-step flows (e.g., adding a record requires multiple messages). States defined in `ConversationState` enum
- **Test setup**: Tests use temp SQLite files via `setupTestStore(t)` helper pattern with `t.Cleanup()`

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | Yes | - | Bot token from @BotFather |
| `POOP_DB_PATH` | No | `poop.db` | SQLite file for poop records |
| `EXPENSE_DB_PATH` | No | `expense.db` | SQLite file for expense records |
| `ALLOWED_USERS_PATH` | No | `allowed_users.cfg` | Allowed usernames file |
| `ALLOWED_GROUPS_PATH` | No | `allowed_groups.cfg` | Allowed group IDs/names file |

## Configuration

Copy `allowed_users.cfg.example` to `allowed_users.cfg` and add Telegram usernames (one per line, with or without `@` prefix). Lines starting with `#` are comments. Similarly, use `allowed_groups.cfg` for group IDs (negative numbers) or group names.
