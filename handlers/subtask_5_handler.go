package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

var pathMemes string

func HandleSubTask5(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
		utils.Logger.Info("HandleSubTask5 called")
        // Get callback data from the button press
        //data := c.Callback().Data

		data := c.Callback().Data
        utils.Logger.WithFields(logrus.Fields{
            "callback_data": data,
        }).Info("HandleSubTask5 received data")
        
        parts := strings.Split(data, "_")
        utils.Logger.WithFields(logrus.Fields{
            "parts": parts,
            "parts_count": len(parts),
        }).Info("Data parts")
        
        // Split the string by "_" delimiter
       // parts := strings.Split(data, "_")
        
        // Check if we have the correct format (waiting_5_X)
        if len(parts) != 3 || parts[0] != "\fwaiting" || parts[1] != "1" {
            return c.Respond(&telebot.CallbackResponse{Text: "Invalid data format"})
        }
        
        // Get the third part (digit after "5_")
        subTaskNum, err := strconv.Atoi(parts[2])
        if err != nil {
            return c.Respond(&telebot.CallbackResponse{Text: "Invalid subtask number"})
        }
        
        // Call the appropriate function based on the digit
        switch subTaskNum {
        case 1:
            return handleSubTask51(c)
        case 2:
            return handleSubTask52(c)
        case 3:
            return handleSubTask53(c)
        default:
            return c.Respond(&telebot.CallbackResponse{Text: "Unknown subtask"})
        }
    }
}

// Functions for handling specific subtasks
func handleSubTask51(c telebot.Context) error {
    // Logic for subtask 5.1
    return c.Respond(&telebot.CallbackResponse{Text: "Processing subtask 5.1"})
}

func handleSubTask52(c telebot.Context) error {
    // Logic for subtask 5.2
    return c.Respond(&telebot.CallbackResponse{Text: "Processing subtask 5.2"})
}

// func handleSubTask53(c telebot.Context) error {
//     // Logic for subtask 5.3
// 	pathMemes := "internal/data/tasks/subtasks/subtask_5"
// 	return nil
// }


// SubTask5Session represents an active session for subtask 5
type SubTask5Session struct {
	UserID        int64
	Username      string
	GameID        int64
	TaskID        int
	CurrentMeme   int
	TotalMemes    int
	Responses     []string
	StartTime     time.Time
	Bot           *telebot.Bot
	Chat          *telebot.Chat
	IsActive      bool
	mu            sync.RWMutex
}

// Global session manager for subtask 5
var (
	subTask5Sessions = make(map[int64]*SubTask5Session) // key - GameID
	sessionsMutex    sync.RWMutex
)

// Get active session for game
func getSubTask5Session(gameID int64) (*SubTask5Session, bool) {
	sessionsMutex.RLock()
	defer sessionsMutex.RUnlock()
	session, exists := subTask5Sessions[gameID]
	return session, exists
}

// Create new session
func createSubTask5Session(userID int64, username string, gameID int64, taskID int, bot *telebot.Bot, chat *telebot.Chat) *SubTask5Session {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	
	session := &SubTask5Session{
		UserID:      userID,
		Username:    username,
		GameID:      gameID,
		TaskID:      taskID,
		CurrentMeme: 1,
		TotalMemes:  5,
		Responses:   make([]string, 0, 4),
		StartTime:   time.Now(),
		Bot:         bot,
		Chat:        chat,
		IsActive:    true,
	}
	
	subTask5Sessions[gameID] = session
	return session
}

// Remove session
func removeSubTask5Session(gameID int64) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	delete(subTask5Sessions, gameID)
}

func handleSubTask53(c telebot.Context) error {
	user := c.Sender()
	chat := c.Chat()
	
	utils.Logger.WithFields(logrus.Fields{
		"source":   "handleSubTask53",
		"username": user.Username,
		"user_id":  user.ID,
		"chat_id":  chat.ID,
	}).Info("User wants to answer SubTask 5.3")

	// Get game data
	game, err := storage_db.GetGameByChatId(chat.ID)
	if err != nil {
		utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
		return c.Respond(&telebot.CallbackResponse{Text: "Помилка отримання гри"})
	}

	// Check player status
	status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, 1) // taskID = 5
	if err != nil {
		utils.Logger.Errorf("Error checking player response status: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Помилка перевірки статусу"})
	}

	// Check if user already answered
	if status.AlreadyAnswered {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("@%s, ти вже відповіла на це завдання 😅", user.Username),
		})
	}

	if status.AlreadySkipped {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("@%s, ти вже пропустила це завдання", user.Username),
		})
	}

	// Check if there's an active session
	if existingSession, exists := getSubTask5Session(int64(game.ID)); exists {
		if existingSession.UserID != user.ID {
			// Another user is already answering
			msgTextOtherUserAnswer := fmt.Sprintf("@%s донт пуш зе хорсес! Інша зірочка зараз відповідає на мемчики.", user.Username)
			
			_, err = msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				msgmanager.TypeNotInGame,
				msgTextOtherUserAnswer,
				10*time.Second,
			)
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
			}
			
			return c.Respond(&telebot.CallbackResponse{Text: "Зачекай своєї черги! 😊"})
		}
	}

	// Create new session
	session := createSubTask5Session(user.ID, user.Username, int64(game.ID), 1, c.Bot(), chat) // таск айди нужно указать 1
	utils.Logger.WithFields(logrus.Fields{
		"source":   "handleSubTask53",
		"user_id":  session.UserID,
		"game_id":  session.GameID,
		"task_id":  session.TaskID,
		"current_meme": session.CurrentMeme,
		"total_memes": session.TotalMemes,
		"responses": session.Responses,
		"start_time": session.StartTime,
		"is_active": session.IsActive,
	}).Info("Created new SubTask 5 session")
	
	msg := fmt.Sprintf("@%s, ти починаєш відповідати на мемчики! 🎭", user.Username)
	_, err = c.Bot().Send(chat, msg)
	if err != nil {
		utils.Logger.Errorf("Error notifying user %s about starting SubTask 5: %v", user.Username, err)
		removeSubTask5Session(int64(game.ID))
		return c.Respond(&telebot.CallbackResponse{Text: "Помилка початку завдання"})
	}

	time.Sleep(1 * time.Second)

	// Send first meme
	err = sendMeme(session, c.Bot())
	if err != nil {
		utils.Logger.Errorf("Error sending first meme: %v", err)
		removeSubTask5Session(int64(game.ID))
		return c.Respond(&telebot.CallbackResponse{Text: "Помилка відправки мема"})
	}

	return c.Respond(&telebot.CallbackResponse{Text: "Почнемо з мемчиками! 🎭"})
}

// sendMeme sends the current meme to user
func sendMeme(session *SubTask5Session, bot *telebot.Bot) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.CurrentMeme > session.TotalMemes {
		return fmt.Errorf("all memes already sent")
	}

	countAnswer, err := storage_db.GetVoiceMemeAnswersCount(session.GameID)
	if err != nil {
		utils.Logger.Errorf("Error getting voice meme answers count for game %s: %v", session.Chat.Title, err)
	}

	pathMemes = fmt.Sprintf("internal/data/tasks/subtasks/subtask_5/%d", countAnswer+1)

	// if countAnswer == 0 {
	// 	pathMemes = fmt.Sprintf("internal/data/tasks/subtasks/subtask_5/%d", countAnswer)
	// } else {
	// 	pathMemes = fmt.Sprintf("internal/data/tasks/subtasks/subtask_5/%d", countAnswer+1)
	// }

	//pathMemes = "internal/data/tasks/subtasks/subtask_5"
	memeFilename := fmt.Sprintf("meme_%d.gif", session.CurrentMeme)
	memePath := filepath.Join(pathMemes, memeFilename)

	utils.Logger.WithFields(logrus.Fields{
		"source":      "sendMeme",
		"user_id":     session.UserID,
		"game_id":     session.GameID,
		"current_meme": session.CurrentMeme,
		"meme_path":   memePath,
	}).Info("Sending meme")

	// Check if file exists
	if _, err := os.Stat(memePath); os.IsNotExist(err) {
		utils.Logger.Errorf("Meme file not found: %s", memePath)
		return fmt.Errorf("meme file not found: %s", memePath)
	}

	// Create photo object
	photo := &telebot.Animation{
		File: telebot.FromDisk(memePath),
		Width: 480,    // Укажите ширину
    	Height: 270,
		//Caption: fmt.Sprintf("Озвуч мем №%d", session.CurrentMeme),
	}

	// Send meme
	_, err = bot.Send(session.Chat, photo)
	if err != nil {
		utils.Logger.Errorf("Error sending meme: %v", err)
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source":      "sendMeme",
		"user_id":     session.UserID,
		"meme_number": session.CurrentMeme,
	}).Info("Meme sent successfully")

	return nil
}

// HandleSubTask5Response handles user responses to memes
func HandleSubTask5Response(bot *telebot.Bot) func(m *telebot.Message) {
	return func(m *telebot.Message) {
		utils.Logger.Info("HandleSubTask5Response called")
		user := m.Sender
		chat := m.Chat

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Error getting game by chat ID: %v", err)
			return
		}

		utils.Logger.WithFields(logrus.Fields{
			"source":   "HandleSubTask5Response",
			"username": user.Username,
			"user_id":  user.ID,
			"chat_id":  chat.ID,
		}).Info("User is responding to SubTask 5")
		// Check if there's an active session for this chat
		session, exists := getSubTask5Session(int64(game.ID))
		if !exists {
			utils.Logger.Info("Session is not exists")

			return // No active session, ignore
		}

		// Check if the same user is responding
		if session.UserID != user.ID {
			return // Wrong user responding, ignore
		}

		session.mu.Lock()
		defer session.mu.Unlock()

		if !session.IsActive {
			return // Session inactive
		}

		// Get response text
		var responseText string
		if m.Text != "" {
			responseText = m.Text
		} else if m.Voice != nil {
			responseText = "[Голосове повідомлення]"
		} else {
			return // Unsupported message type
		}

		// Save response
		session.Responses = append(session.Responses, responseText)

		utils.Logger.WithFields(logrus.Fields{
			"source":       "HandleSubTask5Response",
			"user_id":      session.UserID,
			"meme_number":  session.CurrentMeme,
			"response":     responseText,
			"responses_count": len(session.Responses),
		}).Info("Response received")

		// // Send confirmation
		// confirmationMessages := []string{
		// 	"Класно! 👍",
		// 	"Супер! 🔥",
		// 	"Круто! ✨",
		// 	"Чудово! 🎉",
		// }
		// confirmationText := utils.GetRandomMsg(confirmationMessages)
		
		// _, err = bot.Send(chat, confirmationText)
		// if err != nil {
		// 	utils.Logger.Errorf("Error sending confirmation: %v", err)
		// }

		// Check if all memes are processed
		if session.CurrentMeme >= session.TotalMemes {
			// All memes processed, complete session
			time.Sleep(2*time.Second)
			err := completeSubTask5(session, bot)
			if err != nil {
				utils.Logger.Errorf("Error completing SubTask 5: %v", err)
			}
			return
		}

		// Move to next meme
		session.CurrentMeme++
		
		// Send next meme after small delay
		time.AfterFunc(2*time.Second, func() {
			err := sendMeme(session, bot)
			if err != nil {
				utils.Logger.Errorf("Error sending next meme: %v", err)
				removeSubTask5Session(session.GameID)
			}
		})
	}
}

// completeSubTask5 completes subtask 5
func completeSubTask5(session *SubTask5Session, bot *telebot.Bot) error {
	utils.Logger.WithFields(logrus.Fields{
		"source":         "completeSubTask5",
		"user_id":        session.UserID,
		"game_id":        session.GameID,
		"responses_count": len(session.Responses),
	}).Info("Completing SubTask 5")

	// Save player response to database
	playerResponse := &models.PlayerResponse{
		PlayerID:    session.UserID,
		UserName:    session.Username, // Now using username from session
		GameID:      int(session.GameID),
		TaskID:      session.TaskID,
		HasResponse: true,
		Skipped:     false,
	}

	err := storage_db.AddPlayerResponse(playerResponse)
	if err != nil {
		utils.Logger.Errorf("Error adding player response to DB: %v", err)
		return err
	}

	// Update player status
	err = storage_db.UpdatePlayerStatus(session.UserID, models.StatusPlayerNoWaiting)
	if err != nil {
		utils.Logger.Errorf("Error updating player status: %v", err)
	}

	// Send final message
	finalMessage := fmt.Sprintf("@%s, дякую за відповіді на всі мемчики! 🎭✨ Скоро інші подружки підтянуться 💁‍♀️", session.Username)
	_, err = bot.Send(session.Chat, finalMessage)
	if err != nil {
		utils.Logger.Errorf("Error sending final message: %v", err)
	}

	storage_db.IncrementVoiceMemeAnswers(session.GameID)

	// Mark session as inactive
	session.IsActive = false
	
	// Remove session after some time
	time.AfterFunc(5*time.Second, func() {
		removeSubTask5Session(session.GameID)
	})

	utils.Logger.WithFields(logrus.Fields{
		"source":  "completeSubTask5",
		"user_id": session.UserID,
		"game_id": session.GameID,
	}).Info("SubTask 5 completed successfully")

	return nil
}

// GetActiveSubTask5Session returns active session for game (for use in other parts of code)
func GetActiveSubTask5Session(gameID int64) (*SubTask5Session, bool) {
	return getSubTask5Session(gameID)
}