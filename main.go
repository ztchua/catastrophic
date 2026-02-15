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

	store, err := NewPoopStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	api.Debug = true
	log.Printf("Authorized on account %s", api.Self.UserName)

	bot := NewBot(api, store)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := api.GetUpdatesChan(u)

	for update := range updates {
		bot.HandleUpdate(update)
	}
}
