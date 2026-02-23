package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	dbPath := os.Getenv("POOP_DB_PATH")
	if dbPath == "" {
		dbPath = "poop.db"
	}

	expenseDbPath := os.Getenv("EXPENSE_DB_PATH")
	if expenseDbPath == "" {
		expenseDbPath = "expense.db"
	}

	countDbPath := os.Getenv("COUNT_DB_PATH")
	if countDbPath == "" {
		countDbPath = "count.db"
	}

	authPath := os.Getenv("ALLOWED_USERS_PATH")
	if authPath == "" {
		authPath = "allowed_users.cfg"
	}

	authGroupsPath := os.Getenv("ALLOWED_GROUPS_PATH")
	if authGroupsPath == "" {
		authGroupsPath = "allowed_groups.cfg"
	}

	auth, err := NewAuthService(authPath, authGroupsPath)
	if err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}
	log.Printf("Loaded %d allowed users, %d allowed groups", auth.GetAllowedCount(), auth.GetAllowedGroupsCount())

	store, err := NewPoopStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

	expenseStore, err := NewExpenseStore(expenseDbPath)
	if err != nil {
		log.Fatalf("Failed to initialize expense store: %v", err)
	}
	defer expenseStore.Close()

	countStore, err := NewCountStore(countDbPath)
	if err != nil {
		log.Fatalf("Failed to initialize count store: %v", err)
	}
	defer countStore.Close()

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	api.Debug = true
	log.Printf("Authorized on account %s", api.Self.UserName)

	bot := NewBot(api, store, expenseStore, countStore, auth)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := api.GetUpdatesChan(u)

	for update := range updates {
		bot.HandleUpdate(update)
	}
}
