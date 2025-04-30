package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config struct for settings bot
type Config struct {
	TelegramToken string // Token by telegram-bot
	DatabaseDir   string // name folder database
	DatabaseFile  string // Name database file
	Mode          string // Mode of bot (dev | prod)
}

// LoadConfig load configuration from .env file
func LoadConfig() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env, environment variables in use")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		panic("You need to set the TELEGRAM_TOKEN environment variable")
	}

	dbDir := os.Getenv("DATABASE_DIR")
	if dbDir == "" {
		panic("You need to set the DATABASE_DIR environment variable")
	}

	dbFile := os.Getenv("DATABASE_FILE")
	if dbFile == "" {
		panic("You need to set the DATABASE_FILE environment variable")
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		panic("You need to set the MODE environment variable")
	}

	return &Config{
		TelegramToken: token,
		DatabaseDir:   dbDir,
		DatabaseFile:  dbFile,
		Mode:		   mode,
	}
}
