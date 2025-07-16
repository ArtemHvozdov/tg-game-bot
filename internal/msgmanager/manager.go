package msgmanager

import (
	"fmt"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

// ActiveMessage is active message with timer delete
type ActiveMessage struct {
	Message   *telebot.Message
	Timer     *time.Timer
	CreatedAt time.Time
}

// Manager nanage active user's messages
type Manager struct {
	mu       sync.RWMutex
	messages map[string]*ActiveMessage // key: "chatID:userID:messageType"
	bot      *telebot.Bot
}

// New create new manager of messages
func New(bot *telebot.Bot) *Manager {
	mm := &Manager{
		messages: make(map[string]*ActiveMessage),
		bot:      bot,
	}
	
	// Run period clean every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mm.CleanupOldMessages()
		}
	}()
	
	return mm
}

// generateKey create unique key for message
func (mm *Manager) generateKey(chatID, userID int64, messageType string) string {
	return fmt.Sprintf("%d:%d:%s", chatID, userID, messageType)
}

// HasActive check if there is active message for user
func (mm *Manager) HasActive(chatID, userID int64, messageType string) bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	key := mm.generateKey(chatID, userID, messageType)
	_, exists := mm.messages[key]
	return exists
}

// SendTemporary sends a temporary message with duplication protection
func (mm *Manager) SendTemporaryMessage(chatID, userID int64, messageType, text string, deleteDuration time.Duration, options ...interface{}) (*telebot.Message, error) {
	// Check if there is already an active message of this type
	if mm.HasActive(chatID, userID, messageType) {
		utils.Logger.WithFields(logrus.Fields{
			"userID":     userID,
			"chatID":      chatID,
			"messageType": messageType,
		}).Info("Ignore the duplicate message")
		return nil, nil 
	}
	
	// Create recipient
	recipient := &telebot.Chat{ID: chatID}
	
	// Send message
	msg, err := mm.bot.Send(recipient, text, options...)
	if err != nil {
		return nil, err
	}
	
	// Add message in manager with auto delete
	mm.addMessage(chatID, userID, messageType, msg, deleteDuration)
	
	return msg, nil
}

// addMessage add new msg with auto delete
func (mm *Manager) addMessage(chatID, userID int64, messageType string, msg *telebot.Message, deleteDuration time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	key := mm.generateKey(chatID, userID, messageType)
	
	// if there is already an active message, cancel its timer
	if existing, exists := mm.messages[key]; exists {
		existing.Timer.Stop()
		// Trying to delete the previous message
		go func() {
			mm.bot.Delete(existing.Message)
		}()
	}
	
	// Create timer for delete msg
	timer := time.AfterFunc(deleteDuration, func() {
		mm.deleteMessage(key, msg)
	})
	
	mm.messages[key] = &ActiveMessage{
		Message:   msg,
		Timer:     timer,
		CreatedAt: time.Now(),
	}
}

// deleteMessage deletes the message and removes it from the cache
func (mm *Manager) deleteMessage(key string, msg *telebot.Message) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	// Delete msg from chat
	if err := mm.bot.Delete(msg); err != nil {
		utils.Logger.Errorf("Error delete message: %v", err)
	}
	
	// Delete from cache
	delete(mm.messages, key)
}

// Cancel undoes message deletion
func (mm *Manager) Cancel(chatID, userID int64, messageType string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	key := mm.generateKey(chatID, userID, messageType)
	if activeMsg, exists := mm.messages[key]; exists {
		activeMsg.Timer.Stop()
		delete(mm.messages, key)
	}
}

// CleanupOldMessages clears stuck messages
func (mm *Manager) CleanupOldMessages() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	now := time.Now()
	for key, activeMsg := range mm.messages {
		if now.Sub(activeMsg.CreatedAt) > 10*time.Second {
			activeMsg.Timer.Stop()
			delete(mm.messages, key)
		}
	}
}

// GetActiveCount returns the number of active messages
func (mm *Manager) GetActiveCount() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return len(mm.messages)
}

// GetActiveByUser returns active messages for a specific user
func (mm *Manager) GetActiveByUser(chatID, userID int64) []string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	var messageTypes []string
	prefix := fmt.Sprintf("%d:%d:", chatID, userID)
	
	for key := range mm.messages {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			messageType := key[len(prefix):]
			messageTypes = append(messageTypes, messageType)
		}
	}
	
	return messageTypes
}