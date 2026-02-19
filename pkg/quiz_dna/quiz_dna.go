// Package subtask2 provides functionality for handling subtask 2 - choice-based questions
package quizdna

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	//"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

// Subtask2Item represents a single subtask from JSON
type Subtask2Item struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Options []string `json:"options"`
	Data    []string `json:"data"`
	Image   string   `json:"image"`
}

// Subtask2Data represents the entire JSON structure
type Subtask2Data struct {
	Subtasks []Subtask2Item `json:"subtasks"`
}

// Subtask2Session manages user session for subtask 2
type Subtask2Session struct {
	GameID        int
	TaskID        int
	UserID        int64
	Username      string
	Subtasks      []Subtask2Item
	CurrentStep   int
	Answers       map[int]string // questionIndex -> selectedOption
	StartTime     time.Time
	IsCompleted   bool
}

// Subtask2SessionManager manages all active sessions
type Subtask2SessionManager struct {
	sessions map[int]*Subtask2Session // gameID -> session
	mutex    sync.RWMutex
}

// Global session manager instance
var GlobalSubtask2SessionManager = &Subtask2SessionManager{
	sessions: make(map[int]*Subtask2Session),
}

// StartSession creates a new subtask 2 session
func (sm *Subtask2SessionManager) StartSession(gameID, taskID int, userID int64, username string, subtasks []Subtask2Item) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if another session is active
	if session, exists := sm.sessions[gameID]; exists && !session.IsCompleted {
		if session.UserID != userID {
			return fmt.Errorf("session already active for user %s", session.Username)
		}
	}

	sm.sessions[gameID] = &Subtask2Session{
		GameID:      gameID,
		TaskID:      taskID,
		UserID:      userID,
		Username:    username,
		Subtasks:    subtasks,
		CurrentStep: 0,
		Answers:     make(map[int]string),
		StartTime:   time.Now(),
		IsCompleted: false,
	}

	return nil
}

// GetActiveSession returns active session for given game
func (sm *Subtask2SessionManager) GetActiveSession(gameID int) (*Subtask2Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[gameID]
	return session, exists && !session.IsCompleted
}

// SaveAnswerAndNext saves current answer and moves to next question
func (sm *Subtask2SessionManager) SaveAnswerAndNext(gameID int, selectedOption string) (bool, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[gameID]
	if !exists || session.IsCompleted {
		return false, fmt.Errorf("no active session")
	}

	// Save current answer
	session.Answers[session.CurrentStep] = selectedOption
	session.CurrentStep++

	// Check if all questions answered
	if session.CurrentStep >= len(session.Subtasks) {
		session.IsCompleted = true
		return true, nil // Session completed
	}

	return false, nil // More questions remain
}

// CompleteSession completes and cleans up the session
func (sm *Subtask2SessionManager) CompleteSession(gameID int) map[int]string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[gameID]
	if !exists {
		return nil
	}

	answers := session.Answers
	delete(sm.sessions, gameID)
	return answers
}

// createSubtask2Keyboard creates inline keyboard with 4 option buttons (2x2 layout)
func createSubtask2Keyboard(subtask Subtask2Item, taskID int, questionIndex int, userID int64) *telebot.ReplyMarkup {
	var rows [][]telebot.InlineButton

	// Create each button on its own row for full width
	for i := 0; i < len(subtask.Options); i++ {
		btn := telebot.InlineButton{
			Text: subtask.Options[i],
			Data: fmt.Sprintf("subtask_2_%d_%d_%s", userID, questionIndex, subtask.Data[i]),
		}
		
		// Each button gets its own row
		row := []telebot.InlineButton{btn}
		rows = append(rows, row)
	}

	return &telebot.ReplyMarkup{InlineKeyboard: rows}
}

// WhoIsUsSubTask2 handles the main subtask 2 flow
func WhoIsUsSubTask2(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
			return c.Send("Помилка отримання гри")
		}

		// Check if another user is already answering
		if session, exists := GlobalSubtask2SessionManager.GetActiveSession(game.ID); exists {
			if session.UserID != user.ID {
				msgTextSubtask2OtherUserAlreadyAnswer := fmt.Sprintf("@%s почекай, люба 🌸 Твоя подружка зараз відповідає. Ти зможеш відповісти, як тільки вона завершить!", user.Username)

				_, err := msgmanager.SendTemporaryMessage(
					chat.ID,
					user.ID,
					msgmanager.TypeNotInGame,
					msgTextSubtask2OtherUserAlreadyAnswer,
					2*time.Second,
				)
				if err != nil {
					utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
				}

				return nil
			}
			// User is continuing their session - send current question
			return SendCurrentSubtask2Question(bot, c, game.ID)
		}

		// Load subtasks from JSON file
		subtask2Data, err := LoadSubTask2("internal/data/tasks/subtasks/subtask_2/subtask_2.json")
		if err != nil {
			utils.Logger.Errorf("Failed to load subtask 2: %v", err)
			return c.Send("Помилка завантаження підзавдань")
		}

		if len(subtask2Data.Subtasks) == 0 {
			return c.Send("Підзавдання порожнє")
		}

		utils.Logger.Infof("Loaded subtask 2: %d questions", len(subtask2Data.Subtasks))

		// Start new session
		err = GlobalSubtask2SessionManager.StartSession(game.ID, 2, user.ID, user.Username, subtask2Data.Subtasks)
		if err != nil {
			utils.Logger.Errorf("Error starting subtask 2 session: %v", err)
			return c.Send("Помилка запуску підзавдань")
		}

		utils.Logger.Infof("Started subtask 2 session for user %s in game %d", user.Username, game.ID)

		// Send first question
		utils.Logger.Infof("About to send first subtask 2 question to user %s", user.Username)
		return SendCurrentSubtask2Question(bot, c, game.ID)
	}
}

// SendCurrentSubtask2Question sends current question with image and option buttons
func SendCurrentSubtask2Question(bot *telebot.Bot, c telebot.Context, gameID int) error {
	session, exists := GlobalSubtask2SessionManager.GetActiveSession(gameID)
	if !exists {
		return c.Send("Сесія підзавдань не знайдена")
	}

	if session.CurrentStep >= len(session.Subtasks) {
		return c.Send("Всі питання завершені")
	}

	currentSubtask := session.Subtasks[session.CurrentStep]

	// Create keyboard
	keyboard := createSubtask2Keyboard(currentSubtask, session.TaskID, session.CurrentStep, session.UserID)

	// Create message text
	messageText := currentSubtask.Title

	// Create photo with caption and keyboard
	imagePath := fmt.Sprintf("internal/data/tasks/subtasks/subtask_2/%s", currentSubtask.Image)
	
	// Log the attempt to send photo
	utils.Logger.Infof("Sending subtask 2 question %d to user %s, image path: %s", 
		session.CurrentStep+1, session.Username, imagePath)
	
	photo := &telebot.Photo{
		File:    telebot.FromDisk(imagePath),
		Caption: messageText,
	}

	err := c.Send(photo, keyboard)
	if err != nil {
		utils.Logger.Errorf("Failed to send subtask 2 question: %v", err)
		// Try sending without image as fallback
		utils.Logger.Warnf("Attempting to send without image...")
		return c.Send(messageText, keyboard)
	}
	
	utils.Logger.Infof("Successfully sent subtask 2 question %d", session.CurrentStep+1)
	return nil
}

// LoadSubTask2 loads subtask 2 data from JSON file
func LoadSubTask2(filename string) (*Subtask2Data, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read subtask 2 file %s: %w", filename, err)
	}

	// Parse JSON
	var subtask2Data Subtask2Data
	err = json.Unmarshal(data, &subtask2Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subtask 2 JSON: %w", err)
	}

	// Validate data
	if len(subtask2Data.Subtasks) == 0 {
		return nil, fmt.Errorf("no subtasks found in file %s", filename)
	}

	// Validate each subtask
	for _, subtask := range subtask2Data.Subtasks {
		if len(subtask.Options) != 4 {
			return nil, fmt.Errorf("subtask %d must have exactly 4 options, got %d", subtask.ID, len(subtask.Options))
		}
		if len(subtask.Data) != 4 {
			return nil, fmt.Errorf("subtask %d must have exactly 4 data items, got %d", subtask.ID, len(subtask.Data))
		}
		if subtask.Title == "" {
			return nil, fmt.Errorf("subtask %d has empty title", subtask.ID)
		}
		if subtask.Image == "" {
			return nil, fmt.Errorf("subtask %d has empty image", subtask.ID)
		}

		// Verify image file exists
		imagePath := filepath.Join(filepath.Dir(filename), subtask.Image)
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			utils.Logger.Warnf("Image file not found for subtask %d: %s", subtask.ID, imagePath)
		}

		utils.Logger.Debugf("Loaded subtask %d: %s with %d options", 
			subtask.ID, subtask.Title, len(subtask.Options))
	}

	utils.Logger.Infof("Successfully loaded %d subtasks from %s", len(subtask2Data.Subtasks), filename)
	return &subtask2Data, nil
}
