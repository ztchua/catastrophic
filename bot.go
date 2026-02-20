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
	StateAwaitingUpdateField
	StateAwaitingUpdateValue
	StateAwaitingExpenseItemName
	StateAwaitingExpenseCategory
	StateAwaitingExpenseQuantity
	StateAwaitingExpensePrice
	StateAwaitingExpenseUpdateField
	StateAwaitingExpenseUpdateValue
)

type ChatSession struct {
	State ConversationState
	Data  map[string]interface{}
}

type Bot struct {
	api          *tgbotapi.BotAPI
	store        *PoopStore
	expenseStore *ExpenseStore
	auth         *AuthService
	sessions     map[int64]*ChatSession
	sessionsMu   sync.RWMutex
}

func NewBot(api *tgbotapi.BotAPI, store *PoopStore, expenseStore *ExpenseStore, auth *AuthService) *Bot {
	return &Bot{
		api:          api,
		store:        store,
		expenseStore: expenseStore,
		auth:         auth,
		sessions:     make(map[int64]*ChatSession),
	}
}

// SharedSessionKey is used for centralized session storage
// Since access is controlled by auth, all authorized users share the same session
const SharedSessionKey int64 = 0

func (b *Bot) getSession() *ChatSession {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()

	if _, exists := b.sessions[SharedSessionKey]; !exists {
		b.sessions[SharedSessionKey] = &ChatSession{
			State: StateNone,
			Data:  make(map[string]interface{}),
		}
	}
	return b.sessions[SharedSessionKey]
}

func (b *Bot) resetSession() {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()
	b.sessions[SharedSessionKey] = &ChatSession{
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
	username := update.Message.From.UserName
	chatType := update.Message.Chat.Type
	chatTitle := update.Message.Chat.Title

	if !b.auth.IsAllowed(username, chatID, chatType, chatTitle) {
		b.SendMessage(chatID, "Access denied. You are not authorized to use this bot.")
		return
	}

	// Use shared session since all authorized users share the same data
	session := b.getSession()

	switch session.State {
	case StateAwaitingPoopTexture:
		b.handleAwaitingTexture(update.Message)
		return
	case StateAwaitingUpdateField:
		b.handleAwaitingUpdateField(update.Message)
		return
	case StateAwaitingUpdateValue:
		b.handleAwaitingUpdateValue(update.Message)
		return
	case StateAwaitingExpenseItemName:
		b.handleAwaitingExpenseItemName(update.Message)
		return
	case StateAwaitingExpenseCategory:
		b.handleAwaitingExpenseCategory(update.Message)
		return
	case StateAwaitingExpenseQuantity:
		b.handleAwaitingExpenseQuantity(update.Message)
		return
	case StateAwaitingExpensePrice:
		b.handleAwaitingExpensePrice(update.Message)
		return
	case StateAwaitingExpenseUpdateField:
		b.handleAwaitingExpenseUpdateField(update.Message)
		return
	case StateAwaitingExpenseUpdateValue:
		b.handleAwaitingExpenseUpdateValue(update.Message)
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	b.HandleCommand(update.Message)
}

func (b *Bot) handleAwaitingTexture(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	texture := strings.TrimSpace(msg.Text)

	if texture == "" {
		b.SendMessage(chatID, "Texture cannot be empty. Please enter the poop texture:")
		return
	}

	record, err := b.store.Create(texture)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to save record: %v", err))
		b.resetSession()
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Record saved!\n%s", FormatRecord(record)))
	b.resetSession()
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
/poop_update <id> - Update a record
/poop_delete <id> - Delete a record

Expense Tracking:
/expense_add - Add a new expense
/expense_list - List all expenses
/expense_month - Show total spent this month
/expense_category <category> - Filter by category (past 30 days)
/expense_update <id> - Update an expense
/expense_delete <id> - Delete an expense`)

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

	case "expense_add":
		b.handleExpenseAdd(chatID)

	case "expense_list":
		b.handleExpenseList(chatID)

	case "expense_month":
		b.handleExpenseMonth(chatID)

	case "expense_category":
		b.handleExpenseCategory(chatID, args)

	case "expense_update":
		b.handleExpenseUpdate(chatID, args)

	case "expense_delete":
		b.handleExpenseDelete(chatID, args)

	default:
		b.SendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handlePoopAdd(chatID int64) {
	session := b.getSession()
	session.State = StateAwaitingPoopTexture
	b.SendMessage(chatID, "Please enter the poop texture:")
}

func (b *Bot) handlePoopList(chatID int64) {
	records, err := b.store.GetAll()
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
	records, err := b.store.GetRecent(10)
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
	record, isOverdue, err := b.store.CheckIfOverdue()
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
	args = strings.TrimSpace(args)
	if args == "" {
		b.SendMessage(chatID, "Usage: /poop_update <id>")
		return
	}

	id, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		b.SendMessage(chatID, "Invalid ID. Please provide a numeric ID.")
		return
	}

	record, err := b.store.GetByID(id)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to get record: %v", err))
		return
	}
	if record == nil {
		b.SendMessage(chatID, "Record not found.")
		return
	}

	session := b.getSession()
	session.State = StateAwaitingUpdateField
	session.Data["update_id"] = id

	b.SendMessage(chatID, fmt.Sprintf(`Current record:
%s

What would you like to update?
Reply with:
1 - Datetime
2 - Texture`, FormatRecord(record)))
}

func (b *Bot) handleAwaitingUpdateField(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()
	text := strings.TrimSpace(msg.Text)

	switch text {
	case "1":
		session.Data["update_field"] = "datetime"
		session.State = StateAwaitingUpdateValue
		b.SendMessage(chatID, "Please enter the new datetime (format: YYYY-MM-DD HH:MM):")
	case "2":
		session.Data["update_field"] = "texture"
		session.State = StateAwaitingUpdateValue
		b.SendMessage(chatID, "Please enter the new texture:")
	default:
		b.SendMessage(chatID, "Invalid option. Reply with:\n1 - Datetime\n2 - Texture")
	}
}

func (b *Bot) handleAwaitingUpdateValue(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()

	id, ok := session.Data["update_id"].(int64)
	if !ok {
		b.SendMessage(chatID, "Session error. Please start over with /poop_update <id>")
		b.resetSession()
		return
	}

	field, ok := session.Data["update_field"].(string)
	if !ok {
		b.SendMessage(chatID, "Session error. Please start over with /poop_update <id>")
		b.resetSession()
		return
	}

	var record *PoopRecord
	var err error

	switch field {
	case "datetime":
		datetime, parseErr := time.Parse("2006-01-02 15:04", strings.TrimSpace(msg.Text))
		if parseErr != nil {
			b.SendMessage(chatID, "Invalid datetime format. Please use: YYYY-MM-DD HH:MM")
			return
		}
		record, err = b.store.UpdateDatetime(id, datetime)
	case "texture":
		texture := strings.TrimSpace(msg.Text)
		if texture == "" {
			b.SendMessage(chatID, "Texture cannot be empty. Please enter the new texture:")
			return
		}
		record, err = b.store.UpdateTexture(id, texture)
	default:
		b.SendMessage(chatID, "Session error. Please start over with /poop_update <id>")
		b.resetSession()
		return
	}

	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to update record: %v", err))
		b.resetSession()
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Record updated!\n%s", FormatRecord(record)))
	b.resetSession()
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

	if err := b.store.Delete(id); err != nil {
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

func (b *Bot) handleExpenseAdd(chatID int64) {
	session := b.getSession()
	session.State = StateAwaitingExpenseItemName
	b.SendMessage(chatID, "Please enter the item name:")
}

func (b *Bot) handleAwaitingExpenseItemName(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()
	itemName := strings.TrimSpace(msg.Text)

	if itemName == "" {
		b.SendMessage(chatID, "Item name cannot be empty. Please enter the item name:")
		return
	}

	session.Data["item_name"] = itemName
	session.State = StateAwaitingExpenseCategory
	b.SendMessage(chatID, "Please enter the category (e.g., food, litter, vet, toys):")
}

func (b *Bot) handleAwaitingExpenseCategory(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()
	category := strings.TrimSpace(msg.Text)

	if category == "" {
		b.SendMessage(chatID, "Category cannot be empty. Please enter the category:")
		return
	}

	session.Data["category"] = category
	session.State = StateAwaitingExpenseQuantity
	b.SendMessage(chatID, "Please enter the quantity:")
}

func (b *Bot) handleAwaitingExpenseQuantity(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()
	quantityStr := strings.TrimSpace(msg.Text)

	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil || quantity <= 0 {
		b.SendMessage(chatID, "Invalid quantity. Please enter a positive number:")
		return
	}

	session.Data["quantity"] = quantity
	session.State = StateAwaitingExpensePrice
	b.SendMessage(chatID, "Please enter the price per item:")
}

func (b *Bot) handleAwaitingExpensePrice(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()
	priceStr := strings.TrimSpace(msg.Text)

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		b.SendMessage(chatID, "Invalid price. Please enter a non-negative number:")
		return
	}

	itemName, _ := session.Data["item_name"].(string)
	category, _ := session.Data["category"].(string)
	quantity, _ := session.Data["quantity"].(float64)

	record, err := b.expenseStore.Create(itemName, category, quantity, price)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to save expense: %v", err))
		b.resetSession()
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Expense saved!\n%s", FormatExpenseRecord(record)))
	b.resetSession()
}

func (b *Bot) handleExpenseList(chatID int64) {
	records, err := b.expenseStore.GetAll()
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to get expenses: %v", err))
		return
	}

	if len(records) == 0 {
		b.SendMessage(chatID, "No expense records found. Use /expense_add to add one.")
		return
	}

	result := fmt.Sprintf("All Expense Records (%d total):\n\n%s", len(records), FormatExpenseRecordsList(records))
	b.SendMessage(chatID, result)
}

func (b *Bot) handleExpenseMonth(chatID int64) {
	total, err := b.expenseStore.GetTotalSpentCurrentMonth()
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to calculate total: %v", err))
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Total spent this month: %.2f", total))
}

func (b *Bot) handleExpenseCategory(chatID int64, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		b.SendMessage(chatID, "Usage: /expense_category <category>")
		return
	}

	records, err := b.expenseStore.GetByCategoryPast30Days(args)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to get expenses: %v", err))
		return
	}

	if len(records) == 0 {
		b.SendMessage(chatID, fmt.Sprintf("No expenses found in category '%s' in the past 30 days.", args))
		return
	}

	result := fmt.Sprintf("Expenses in '%s' (past 30 days):\n\n%s", args, FormatExpenseRecordsList(records))
	b.SendMessage(chatID, result)
}

func (b *Bot) handleExpenseUpdate(chatID int64, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		b.SendMessage(chatID, "Usage: /expense_update <id>")
		return
	}

	id, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		b.SendMessage(chatID, "Invalid ID. Please provide a numeric ID.")
		return
	}

	record, err := b.expenseStore.GetByID(id)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to get expense: %v", err))
		return
	}
	if record == nil {
		b.SendMessage(chatID, "Expense record not found.")
		return
	}

	session := b.getSession()
	session.State = StateAwaitingExpenseUpdateField
	session.Data["update_id"] = id

	b.SendMessage(chatID, fmt.Sprintf(`Current record:
%s

What would you like to update?
Reply with:
1 - Item Name
2 - Category
3 - Quantity
4 - Price`, FormatExpenseRecord(record)))
}

func (b *Bot) handleAwaitingExpenseUpdateField(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()
	text := strings.TrimSpace(msg.Text)

	switch text {
	case "1":
		session.Data["update_field"] = "item_name"
		session.State = StateAwaitingExpenseUpdateValue
		b.SendMessage(chatID, "Please enter the new item name:")
	case "2":
		session.Data["update_field"] = "category"
		session.State = StateAwaitingExpenseUpdateValue
		b.SendMessage(chatID, "Please enter the new category:")
	case "3":
		session.Data["update_field"] = "quantity"
		session.State = StateAwaitingExpenseUpdateValue
		b.SendMessage(chatID, "Please enter the new quantity:")
	case "4":
		session.Data["update_field"] = "price"
		session.State = StateAwaitingExpenseUpdateValue
		b.SendMessage(chatID, "Please enter the new price:")
	default:
		b.SendMessage(chatID, "Invalid option. Reply with:\n1 - Item Name\n2 - Category\n3 - Quantity\n4 - Price")
	}
}

func (b *Bot) handleAwaitingExpenseUpdateValue(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession()

	id, ok := session.Data["update_id"].(int64)
	if !ok {
		b.SendMessage(chatID, "Session error. Please start over with /expense_update <id>")
		b.resetSession()
		return
	}

	field, ok := session.Data["update_field"].(string)
	if !ok {
		b.SendMessage(chatID, "Session error. Please start over with /expense_update <id>")
		b.resetSession()
		return
	}

	record, err := b.expenseStore.GetByID(id)
	if err != nil || record == nil {
		b.SendMessage(chatID, "Failed to get expense record")
		b.resetSession()
		return
	}

	newItemName := record.ItemName
	newCategory := record.Category
	newQuantity := record.Quantity
	newPrice := record.Price

	switch field {
	case "item_name":
		newItemName = strings.TrimSpace(msg.Text)
		if newItemName == "" {
			b.SendMessage(chatID, "Item name cannot be empty. Please enter the new item name:")
			return
		}
	case "category":
		newCategory = strings.TrimSpace(msg.Text)
		if newCategory == "" {
			b.SendMessage(chatID, "Category cannot be empty. Please enter the new category:")
			return
		}
	case "quantity":
		quantity, parseErr := strconv.ParseFloat(strings.TrimSpace(msg.Text), 64)
		if parseErr != nil || quantity <= 0 {
			b.SendMessage(chatID, "Invalid quantity. Please enter a positive number:")
			return
		}
		newQuantity = quantity
	case "price":
		price, parseErr := strconv.ParseFloat(strings.TrimSpace(msg.Text), 64)
		if parseErr != nil || price < 0 {
			b.SendMessage(chatID, "Invalid price. Please enter a non-negative number:")
			return
		}
		newPrice = price
	default:
		b.SendMessage(chatID, "Session error. Please start over with /expense_update <id>")
		b.resetSession()
		return
	}

	updated, err := b.expenseStore.Update(id, newItemName, newCategory, newQuantity, newPrice)
	if err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to update expense: %v", err))
		b.resetSession()
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Expense updated!\n%s", FormatExpenseRecord(updated)))
	b.resetSession()
}

func (b *Bot) handleExpenseDelete(chatID int64, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		b.SendMessage(chatID, "Usage: /expense_delete <id>")
		return
	}

	id, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		b.SendMessage(chatID, "Invalid ID. Please provide a numeric ID.")
		return
	}

	if err := b.expenseStore.Delete(id); err != nil {
		b.SendMessage(chatID, fmt.Sprintf("Failed to delete expense: %v", err))
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Expense %d deleted successfully.", id))
}
