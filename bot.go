package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ConversationState int

const (
	StateNone ConversationState = iota
	StateAwaitingPoopTexture
)

type ChatSession struct {
	State ConversationState
	Data  map[string]interface{}
}

type Bot struct {
	api        *tgbotapi.BotAPI
	store      *PoopStore
	sessions   map[int64]*ChatSession
	sessionsMu sync.RWMutex
}

func NewBot(api *tgbotapi.BotAPI, store *PoopStore) *Bot {
	return &Bot{
		api:      api,
		store:    store,
		sessions: make(map[int64]*ChatSession),
	}
}

func (b *Bot) getSession(chatID int64) *ChatSession {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()

	if _, exists := b.sessions[chatID]; !exists {
		b.sessions[chatID] = &ChatSession{
			State: StateNone,
			Data:  make(map[string]interface{}),
		}
	}
	return b.sessions[chatID]
}

func (b *Bot) resetSession(chatID int64) {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()
	b.sessions[chatID] = &ChatSession{
		State: StateNone,
		Data:  make(map[string]interface{}),
	}
}

func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (b *Bot) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	session := b.getSession(chatID)

	if session.State == StateAwaitingPoopTexture {
		b.handleAwaitingTexture(update.Message)
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	b.HandleCommand(update.Message)
}

func (b *Bot) handleAwaitingTexture(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession(chatID)
	texture := strings.TrimSpace(msg.Text)

	if texture == "" {
		b.SendMessage(chatID, "Texture cannot be empty. Please enter the poop texture:")
		return
	}

	chatIDFloat, ok := session.Data["chat_id"].(int64)
	if !ok {
		b.SendMessage(chatID, "Session error. Please start over with /poop_add")
		b.resetSession(chatID)
		return
	}

	record, err := b.store.Create(chatIDFloat, texture)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to save record: %v", err))
		b.resetSession(chatID)
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Record saved!\n%s", FormatRecord(record)))
	b.resetSession(chatID)
}

func (b *Bot) HandleCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	command := msg.Command()
	args := msg.CommandArguments()

	switch command {
	case "start":
		b.SendMessage(chatID, "Welcome! I'm your pet poop tracker bot.\n\nUse /help to see available commands.")

	case "help":
		b.SendMessage(chatID, `Available commands:

Basic:
/start - Start the bot
/help - Show this help message
/ping - Check if bot is alive

Poop Tracking:
/poop_add - Add a new poop record
/poop_list - List all records
/poop_recent - Show 10 most recent records
/poop_check - Check if pet hasn't pooped in 3+ days
/poop_update <id> <texture> - Update a record
/poop_delete <id> - Delete a record`)

	case "ping":
		b.SendMessage(chatID, "Pong!")

	case "poop_add":
		b.handlePoopAdd(chatID)

	case "poop_list":
		b.handlePoopList(chatID)

	case "poop_recent":
		b.handlePoopRecent(chatID)

	case "poop_check":
		b.handlePoopCheck(chatID)

	case "poop_update":
		b.handlePoopUpdate(chatID, args)

	case "poop_delete":
		b.handlePoopDelete(chatID, args)

	default:
		b.SendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handlePoopAdd(chatID int64) {
	session := b.getSession(chatID)
	session.State = StateAwaitingPoopTexture
	session.Data["chat_id"] = chatID
	b.SendMessage(chatID, "Please enter the poop texture:")
}

func (b *Bot) handlePoopList(chatID int64) {
	records, err := b.store.GetAll(chatID)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to get records: %v", err))
		return
	}

	if len(records) == 0 {
		b.SendMessage(chatID, "No poop records found. Use /poop_add to add one.")
		return
	}

	result := fmt.Sprintf("All Poop Records (%d total):\n\n%s", len(records), FormatRecordsList(records))
	b.SendMessage(chatID, result)
}

func (b *Bot) handlePoopRecent(chatID int64) {
	records, err := b.store.GetRecent(chatID, 10)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to get records: %v", err))
		return
	}

	if len(records) == 0 {
		b.SendMessage(chatID, "No poop records found. Use /poop_add to add one.")
		return
	}

	result := fmt.Sprintf("10 Most Recent Poop Records:\n\n%s", FormatRecordsList(records))
	b.SendMessage(chatID, result)
}

func (b *Bot) handlePoopCheck(chatID int64) {
	record, isOverdue, err := b.store.CheckIfOverdue(chatID)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to check records: %v", err))
		return
	}

	if record == nil {
		b.SendMessage(chatID, "No poop records found. Use /poop_add to add one.")
		return
	}

	timeSince := timeSinceRecord(record.Datetime)

	if isOverdue {
		b.SendMessage(chatID, fmt.Sprintf(
			"WARNING: Your pet hasn't pooped in %s!\n\nLast record:\n%s\n\nA vet visit may be needed.",
			timeSince,
			FormatRecord(record),
		))
	} else {
		b.SendMessage(chatID, fmt.Sprintf(
			"Good news! Last poop was %s ago.\n\nLast record:\n%s",
			timeSince,
			FormatRecord(record),
		))
	}
}

func (b *Bot) handlePoopUpdate(chatID int64, args string) {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		b.SendMessage(chatID, "Usage: /poop_update <id> <texture>")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.SendMessage(chatID, "Invalid ID. Please provide a numeric ID.")
		return
	}

	texture := strings.Join(parts[1:], " ")
	if texture == "" {
		b.SendMessage(chatID, "Texture cannot be empty.")
		return
	}

	record, err := b.store.Update(id, chatID, texture)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to update record: %v", err))
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Record updated!\n%s", FormatRecord(record)))
}

func (b *Bot) handlePoopDelete(chatID int64, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		b.SendMessage(chatID, "Usage: /poop_delete <id>")
		return
	}

	id, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		b.SendMessage(chatID, "Invalid ID. Please provide a numeric ID.")
		return
	}

	if err := b.store.Delete(id, chatID); err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to delete record: %v", err))
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Record %d deleted successfully.", id))
}

func timeSinceRecord(t time.Time) string {
	duration := time.Since(t)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%d days and %d hours", days, hours)
	}
	return fmt.Sprintf("%d hours", hours)
}
