package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/quiz_dna"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

var (
	greetings = []string{
		"кицю 🐱", 
		"пандочко 🐼", 
		"лисичко 🦊",  
		"мишко 🐭", 
		"левице 🦁",
	}
	
	// Map to track used greetings per game: gameID -> map[userID -> greeting]
	gameGreetings = make(map[int]map[int64]string)
	greetingsMutex = sync.Mutex{}
)


// HandleSubTask10 handles button clicks for subtask 10
func HandleSubTask10(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()
		msg := c.Message()
		data := c.Data()

		utils.Logger.Infof("HandleSubTask10 called by user %s in chat %s with data: %s", user.Username, chat.Title, data)

		// Check if this is a subtask 10 callback
		if !strings.HasPrefix(data, "subtask_10_") {
			return nil
		}

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
			return c.Send("Помилка отримання гри")
		}

		// Parse callback data: "subtask_10_{userID}_{questionIndex}_{selectedData}"
		dataWithoutPrefix := strings.TrimPrefix(data, "subtask_10_")
		parts := strings.Split(dataWithoutPrefix, "_")

		if len(parts) < 3 {
			utils.Logger.Errorf("Invalid callback data format: %s", data)
			return nil
		}

		// Parse components
		userIDStr := parts[0]
		questionIndexStr := parts[1]
		selectedData := strings.Join(parts[2:], "_") // In case data contains underscores

		// Convert to types
		expectedUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			utils.Logger.Errorf("Invalid userID in callback data: %s", userIDStr)
			return nil
		}

		questionIndex, err := strconv.Atoi(questionIndexStr)
		if err != nil {
			utils.Logger.Errorf("Invalid questionIndex in callback data: %s", questionIndexStr)
			return nil
		}

		// Verify that the clicking user matches the session user
		if user.ID != expectedUserID {
			msgTextWrongUser := fmt.Sprintf("@%s донт пуш зе хорсес! Інша зірочка зараз відповідає.", user.Username)

			_, err = msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				msgmanager.TypeNotInGame,
				msgTextWrongUser,
				10*time.Second,
			)
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
			}
			return nil
		}

		// Check player response status
		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, 10)
		if err != nil {
			utils.Logger.Errorf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			msgYouAlreadyAnswered, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s already answered task 10: %v", user.Username, err)
			}

			time.AfterFunc(cfg.Durations.TimeDeleteMsgYouAlreadyAnswered, func() {
				err = bot.Delete(msgYouAlreadyAnswered)
				if err != nil {
					utils.Logger.Errorf("Error deleting message that user %s already answered task 10: %v", user.Username, err)
				}
			})

			return nil
		case status.AlreadySkipped:
			return c.Send(fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		}

		// Check if user has active session
		session, exists := quizdna.GlobalSubtask10SessionManager.GetActiveSession(game.ID)
		if !exists || session.UserID != user.ID {
			msgTextOtherUserAnswer := fmt.Sprintf("@%s донт пуш зе хорсес! Інша зірочка зараз відповідає.", user.Username)

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

			return nil
		}

		// Verify question index matches current step
		if questionIndex != session.CurrentStep {
			utils.Logger.Errorf("Question index mismatch: expected %d, got %d", session.CurrentStep, questionIndex)
			return nil
		}

		utils.Logger.WithFields(logrus.Fields{
			"source":           "HandleSubTask10",
			"username":         user.Username,
			"task_id":          10,
			"question_index":   questionIndex,
			"selected_data":    selectedData,
		}).Infof("User %s selected option %s for question %d", user.Username, selectedData, questionIndex)

		// Delete the question message
		err = bot.Delete(msg)
		if err != nil {
			utils.Logger.Errorf("Failed to delete message: %v", err)
		}

		// Save answer and check if completed
		completed, err := quizdna.GlobalSubtask10SessionManager.SaveAnswerAndNext(game.ID, selectedData)
		if err != nil {
			utils.Logger.Errorf("Error saving subtask 10 answer: %v", err)
			return c.Send("Помилка збереження відповіді")
		}

		// Save to database
		subtask10Answer := &models.Subtask10Answer{
			GameID:           game.ID,
			TaskID:           10,
			QuestionIndex:    questionIndex,
			AnswererUserID:   user.ID,
			SelectedOption:   selectedData,
			QuestionID:       session.Subtasks[questionIndex].ID,
		}

		err = storage_db.AddSubtask10Answer(subtask10Answer)
		if err != nil {
			utils.Logger.Errorf("Error add subtask 10 answer to DB: %v", err)
		} else {
			utils.Logger.Infof("Answer of subtask 10 add to DB: success")
		}

		if completed {
			// All questions answered
			answers := quizdna.GlobalSubtask10SessionManager.CompleteSession(game.ID)

			utils.Logger.WithFields(logrus.Fields{
				"source":        "HandleSubTask10",
				"username":      user.Username,
				"total_answers": len(answers),
				"task_id":       10,
			}).Info("Subtask 10 completed")

			playerResponse := &models.PlayerResponse{
				PlayerID:    user.ID,
				GameID:      game.ID,
				TaskID:      10,
				HasResponse: true,
				Skipped:     false,
			}

			storage_db.AddPlayerResponse(playerResponse)
			storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)

			//greetingMsg := utils.GetRandomMsg(greetings)
			greetingMsg := getUniqueGreeting(game.ID, user.ID)

			return c.Send(fmt.Sprintf("@%s дякую %s Коли усі подружки дадуть свої відповіді, я зроблю ваш спільний колаж!", user.Username, greetingMsg))
		}

		// Send next question
		return quizdna.SendCurrentSubtask10Question(bot, c, game.ID)
	}
}

// getUniqueGreeting returns a unique greeting for the user in specific game
func getUniqueGreeting(gameID int, userID int64) string {
	greetingsMutex.Lock()
	defer greetingsMutex.Unlock()
	
	// Initialize game map if doesn't exist
	if gameGreetings[gameID] == nil {
		gameGreetings[gameID] = make(map[int64]string)
	}
	
	// Check if user already has a greeting in this game
	if greeting, exists := gameGreetings[gameID][userID]; exists {
		return greeting
	}
	
	// Find available greetings for this game
	availableGreetings := make([]string, 0)
	usedSet := make(map[string]bool)
	
	// Mark used greetings in this game
	for _, used := range gameGreetings[gameID] {
		usedSet[used] = true
	}
	
	// Find available ones
	for _, greeting := range greetings {
		if !usedSet[greeting] {
			availableGreetings = append(availableGreetings, greeting)
		}
	}
	
	// If all are used in this game, reset and use all
	if len(availableGreetings) == 0 {
		gameGreetings[gameID] = make(map[int64]string) // Reset for this game
		availableGreetings = greetings
	}
	
	// Pick random from available
	rand.Seed(time.Now().UnixNano())
	selectedGreeting := availableGreetings[rand.Intn(len(availableGreetings))]
	
	// Assign to user in this game
	gameGreetings[gameID][userID] = selectedGreeting
	
	fmt.Printf("Assigned greeting '%s' to user %d in game %d\n", selectedGreeting, userID, gameID)
	
	return selectedGreeting
}
