package main

import (
	"log"
	"os"

	"github.com/ArtemHvozdov/tg-game-bot.git/config"
	"github.com/ArtemHvozdov/tg-game-bot.git/handlers"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"gopkg.in/telebot.v3"
)
func main() {
	cfg := config.LoadConfig() // Loading the configuration from a file or environment variable

	// Initialize the database
	dataDir := cfg.DatabaseDir
	dataFile := cfg.DatabaseFile

	dbPath := dataDir + dataFile

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Error creating folder %s: %v", dataDir, err)
	}
	
	db, err := storage_db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer storage_db.CloseDB(db)

	pref := telebot.Settings{
		Token: cfg.TelegramToken,
		Poller: &telebot.LongPoller{
			Timeout: 10,
		},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Failed to create a new bot: %v", err)
	}

	bot.Handle("/start", handlers.StartHandler(bot))

	log.Println("Bot is running...")
	bot.Start()
}
