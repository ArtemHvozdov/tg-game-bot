package subtasks

import (
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"gopkg.in/telebot.v3"
)

// Common interface for all subtask processors
type SubtaskProcessor interface {
    // Start subtask for user
    Start(bot *telebot.Bot, c telebot.Context, game *models.Game, user *telebot.User) error
    
    // Handle user interaction (callback, message, photo, etc.)
    HandleInteraction(bot *telebot.Bot, c telebot.Context) error
    
    // Check if user has active session
    HasActiveSession(gameID int, userID int64) bool
    
    // Complete and cleanup session
    Complete(gameID int) error
    
    // Get processor name for logging
    Name() string
}

// Common session data
type SessionData struct {
    GameID    int
    TaskID    int
    UserID    int64
    Username  string
    StartTime time.Time
}