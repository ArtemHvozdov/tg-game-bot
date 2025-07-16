package msgmanager

import (
	"fmt"
	"time"

	"gopkg.in/telebot.v3"
)

// Global Manager Instance
var globalManager *Manager

// Init initializes the global message manager
func Init(bot *telebot.Bot) {
	globalManager = New(bot)
}

// GetManager returns the global manager (for direct access to methods)
func GetManager() *Manager {
	return globalManager
}

// HasActive checks active messages via global manager
func HasActive(chatID, userID int64, messageType string) bool {
	if globalManager == nil {
		return false
	}
	return globalManager.HasActive(chatID, userID, messageType)
}

// SendTemporaryMessage sends a temporary message via the global manager
func SendTemporaryMessage(chatID, userID int64, messageType, text string, deleteDuration time.Duration, options ...interface{}) (*telebot.Message, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("message manager not initialized")
	}
	return globalManager.SendTemporaryMessage(chatID, userID, messageType, text, deleteDuration, options...)
}

// Cancel undoes deletion of message via global manager
func Cancel(chatID, userID int64, messageType string) {
	if globalManager != nil {
		globalManager.Cancel(chatID, userID, messageType)
	}
}

// GetActiveCount returns the number of active messages
func GetActiveCount() int {
	if globalManager == nil {
		return 0
	}
	return globalManager.GetActiveCount()
}
