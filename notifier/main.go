package main

import (
	"log"
	"os"
	"time"
	"database/sql"

	//"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("=== Notifier started ===")
	log.Printf("Execution time: %s", time.Now().Format("2006-01-02 15:04:05"))

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Failed to load .env file, using environment variables")
	}

	// Get environment variables
	dbDir := os.Getenv("DATABASE_DIR")
	if dbDir == "" {
		log.Fatal("DATABASE_DIR environment variable is required")
	}

	dbFile := os.Getenv("DATABASE_FILE")
	if dbFile == "" {
		log.Fatal("DATABASE_FILE environment variable is required")
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Println("Warning: TELEGRAM_TOKEN not set")
	}

	// Initialization Telegram bot
	pref := telebot.Settings{
		Token:  telegramToken,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	dbPath := dbDir + dbFile
	log.Printf("Database path: %s", dbPath)

	// Create database directory if it doesn't exist
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Error creating database directory: %v", err)
	}

	// Open database connection	
	db, err := sql.Open("sqlite3", dbPath+
		"?_journal_mode=WAL"+
		"&_foreign_keys=on"+
		"&_busy_timeout=5000") 
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Setting pool of connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
}