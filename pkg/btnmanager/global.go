package btnmanager

import (
	"fmt"
	"log"

	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

// ManagerBtns - global instance of button manager
var ManagerBtns *Manager

// Init initializes the global button manager from a JSON file
func Init(path string) error {
	ManagerBtns = NewManager()
	
	if err := ManagerBtns.loadFromFile(path); err != nil {
		return fmt.Errorf("failed to initialize button manager: %w", err)
	}
	
	utils.Logger.Infof("Button manager initialized with %d buttons from %s", 
		len(ManagerBtns.buttons), path)
	
	return nil
}

// MustInit initializes the button manager or panics on error
func MustInit(path string) {
	if err := Init(path); err != nil {
		log.Fatalf("Button manager initialization failed: %v", err)
	}
}

// Get creates an inline button (global wrapper function)
func Get(markup *telebot.ReplyMarkup, unique string, args ...any) telebot.Btn {
	if ManagerBtns == nil {
		log.Printf("WARNING: Button manager not initialized, call Init() first")
		return markup.Data("ERROR: Manager not initialized", "error")
	}
	return ManagerBtns.Get(markup, unique, args...)
}

// GetInlineButton creates an InlineButton structure (global wrapper function)
func GetInlineButton(unique string, args ...any) telebot.InlineButton {
	if ManagerBtns == nil {
		log.Printf("WARNING: Button manager not initialized, call Init() first")
		return telebot.InlineButton{
			Text: "ERROR: Manager not initialized",
			Data: "error",
		}
	}
	return ManagerBtns.GetInlineButton(unique, args...)
}

// HasButton checks if a button exists (global wrapper function)
func HasButton(unique string) bool {
	if ManagerBtns == nil {
		return false
	}
	return ManagerBtns.HasButton(unique)
}

// GetConfig returns the button configuration (global wrapper function)
func GetConfig(unique string) (ButtonConfig, bool) {
	if ManagerBtns == nil {
		return ButtonConfig{}, false
	}
	return ManagerBtns.GetConfig(unique)
}