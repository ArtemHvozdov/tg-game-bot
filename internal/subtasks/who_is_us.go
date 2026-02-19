package subtasks

import (
	"fmt"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

type SubtaskSession struct {
    GameID        int
    TaskID        int
    UserID        int64
    Username      string
    Questions     []string
    CurrentStep   int
    Answers       map[int]string // questionIndex -> selectedUsername
    StartTime     time.Time
    IsCompleted   bool
}

type SessionManager struct {
    sessions map[int]*SubtaskSession // gameID -> session
    mutex    sync.RWMutex
}

var GlobalSessionManager = &SessionManager{
    sessions: make(map[int]*SubtaskSession),
}

// Start new subtask session
func (sm *SessionManager) StartSession(gameID, taskID int, userID int64, username string, questions []string) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    // Check if another session is active
    if session, exists := sm.sessions[gameID]; exists && !session.IsCompleted {
        if session.UserID != userID {
            return fmt.Errorf("session already active for user %s", session.Username)
        }
    }
    
    sm.sessions[gameID] = &SubtaskSession{
        GameID:      gameID,
        TaskID:      taskID,
        UserID:      userID,
        Username:    username,
        Questions:   questions,
        CurrentStep: 0,
        Answers:     make(map[int]string),
        StartTime:   time.Now(),
        IsCompleted: false,
    }
    
    return nil
}

// Get active session
func (sm *SessionManager) GetActiveSession(gameID int) (*SubtaskSession, bool) {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()
    
    session, exists := sm.sessions[gameID]
    return session, exists && !session.IsCompleted
}

// Save answer and move to next question
func (sm *SessionManager) SaveAnswerAndNext(gameID int, selectedUsername string) (bool, error) {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    session, exists := sm.sessions[gameID]
    if !exists || session.IsCompleted {
        return false, fmt.Errorf("no active session")
    }
    
    // Save current answer
    session.Answers[session.CurrentStep] = selectedUsername
    session.CurrentStep++
    
    // Check if all questions answered
    if session.CurrentStep >= len(session.Questions) {
        session.IsCompleted = true
        return true, nil // Session completed
    }
    
    return false, nil // More questions remain
}

// Complete and cleanup session
func (sm *SessionManager) CompleteSession(gameID int) map[int]string {
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

func filterOutCurrentUser(players []models.Player, currentUserID int64) []models.Player {
	filtered := make([]models.Player, 0, len(players)-1)
	for _, player := range players {
		if player.ID != currentUserID {
			filtered = append(filtered, player)
		}
	}
	return filtered
}

// Create inline keyboard from players with 2 buttons per row
func createPlayersKeyboard(players []models.Player, taskID int, questionIndex int) *telebot.ReplyMarkup {
    var rows [][]telebot.InlineButton
    
    // Process players in pairs
    for i := 0; i < len(players); i += 2 {
        var row []telebot.InlineButton
        
        // First button in row
        btn1 := telebot.InlineButton{
            Text: fmt.Sprintf("@%s", players[i].UserName),
            Data: fmt.Sprintf("subtask_%d_%d_%d|%s", taskID, questionIndex, players[i].ID ,players[i].UserName),
        }
        row = append(row, btn1)
        
        // Second button if exists
        if i+1 < len(players) {
            btn2 := telebot.InlineButton{
                Text: fmt.Sprintf("@%s", players[i+1].UserName),
                Data: fmt.Sprintf("subtask_%d_%d_%d|%s", taskID, questionIndex, players[i+1].ID, players[i+1].UserName),
            }
            row = append(row, btn2)
        }
        
        rows = append(rows, row)
    }
    
    return &telebot.ReplyMarkup{InlineKeyboard: rows}
}

// func filterOutCurrentUser(players []models.Player, currentUserID int64) []models.Player {
//     filtered := make([]models.Player, 0, len(players)-1)
//     for _, player := range players {
//         if player.UserID != currentUserID {
//             filtered = append(filtered, player)
//         }
//     }
//     return filtered
// }

func WhoIsUsSubTask(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        user := c.Sender()
        chat := c.Chat()
        
        game, err := storage_db.GetGameByChatId(chat.ID)
        if err != nil {
            utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
            return c.Send("Помилка отримання гри")
        }
        
        // Check if another user is already answering
        if session, exists := GlobalSessionManager.GetActiveSession(game.ID); exists {
            if session.UserID != user.ID {
                //return c.Send("Почекай, люба 🌸 Твоя подружка зараз відповідає. Ти зможеш відповісти, як тільки вона завершить!")

				msgTextSubtaskOtheruserAlreadyAnswer := fmt.Sprintf("@%s почекай, люба 🌸 Твоя подружка зараз відповідає. Ти зможеш відповісти, як тільки вона завершить!", user.Username)

				_, err := msgmanager.SendTemporaryMessage(
					chat.ID,
					user.ID,
					msgmanager.TypeNotInGame, // unique message type
					msgTextSubtaskOtheruserAlreadyAnswer,
					10 * time.Second,
				)
				if err != nil {
					utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
				}

				return nil
            }
            // User is continuing their session - send current question
            return SendCurrentQuestion(bot, c, game.ID)
        }
        
        // Load subtasks from JSON file
        subtasks, err := utils.LoadSubTasks("internal/data/tasks/subtasks/subtask_4.json")
        if err != nil {
            utils.Logger.Errorf("Failed to load subtasks: %v", err)
            return c.Send("Помилка завантаження підзавдань")
        }
        
        if len(subtasks) == 0 {
            return c.Send("Підзавдання порожнє")
        }
        
        utils.Logger.Infof("Loaded subtasks for task 4: %d questions", len(subtasks))
        
        // Start new session
        err = GlobalSessionManager.StartSession(game.ID, 4, user.ID, user.Username, subtasks)
        if err != nil {
            utils.Logger.Errorf("Error starting subtask session: %v", err)
            return c.Send("Помилка запуску підзавдань")
        }
        
        // Send first question
        return SendCurrentQuestion(bot, c, game.ID)
    }
}

// Send current question with player buttons
func SendCurrentQuestion(bot *telebot.Bot, c telebot.Context, gameID int) error {
    session, exists := GlobalSessionManager.GetActiveSession(gameID)
    if !exists {
        return c.Send("Сесія підзавдань не знайдена")
    }
    
    if session.CurrentStep >= len(session.Questions) {
        return c.Send("Всі питання завершені")
    }
    
    // Get all players from game
    allPlayerGame, err := storage_db.GetAllPlayersByGameID(gameID)
    if err != nil {
        utils.Logger.Errorf("Failed to get players for game %d: %v", gameID, err)
        return c.Send("Помилка отримання гравців")
    }
    
    // // Filter out current user
    // otherPlayers := filterOutCurrentUser(allPlayerGame, session.UserID)
    
    // if len(otherPlayers) == 0 {
    //     return c.Send("В грі немає інших гравців для опитування")
    // }
    
    // Create keyboard
    keyboard := createPlayersKeyboard(allPlayerGame, session.TaskID, session.CurrentStep)
    
    // Create message text
    currentQuestion := session.Questions[session.CurrentStep]
    messageText := fmt.Sprintf("Питання %d з %d:\n\n%s\n\nОберіть відповідь:",
        session.CurrentStep+1, len(session.Questions), currentQuestion)
    
    return c.Send(messageText, keyboard)
}