package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/internal/subtasks"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func HandleSubTask3(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        user := c.Sender()
        chat := c.Chat()
        msg := c.Message()
        data := c.Data()
        
        // Check if this is a subtask callback
        if !strings.HasPrefix(data, "subtask_") {
            return nil
        }
        
        game, err := storage_db.GetGameByChatId(chat.ID)
        if err != nil {
            utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
            return c.Send("Помилка отримання гри")
        }
        
        // Parse callback data
        // ... (ваш существующий код парсинга) ...
        
        // Remove prefix "subtask_" first
        dataWithoutPrefix := strings.TrimPrefix(data, "subtask_")
        
        // Parse data (your existing parsing code)
        firstUnderscore := strings.Index(dataWithoutPrefix, "_")
        taskIDStr := dataWithoutPrefix[:firstUnderscore]
        remainder := dataWithoutPrefix[firstUnderscore+1:]
        
        secondUnderscore := strings.Index(remainder, "_")
        questionIndexStr := remainder[:secondUnderscore]
        userPart := remainder[secondUnderscore+1:]
        
        pipeIndex := strings.Index(userPart, "|")
        selectedUserIDStr := userPart[:pipeIndex]
        selectedUsername := userPart[pipeIndex+1:]
        
        // Convert to types
        taskID, _ := strconv.Atoi(taskIDStr)
        questionIndex, _ := strconv.ParseUint(questionIndexStr, 10, 32)
        selectedUserID, _ := strconv.ParseInt(selectedUserIDStr, 10, 64)
        
		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, taskID)
		if err != nil {
			utils.Logger.Errorf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			//textYouAlreadyAnswered := fmt.Sprintf("@%s, ти вже відповіла на це завдання 😅", user.Username)
			msgYouAlreadyAnswered, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s already answered task %d: %v", user.Username, taskID, err)
			}

			time.AfterFunc(cfg.Durations.TimeDeleteMsgYouAlreadyAnswered, func() {
				err = bot.Delete(msgYouAlreadyAnswered)
				if err != nil {
					utils.Logger.WithFields(logrus.Fields{
						"source": "OnAnswerTaskBtnHandler",
						"username": user.Username,
						"group": chat.Title,
						"data_button": data,
						"task_id": taskID,
					}).Errorf("Error deleting message that user %s already answered task %d: %v", user.Username, taskID, err)
				}
			})

			// return c.Send(fmt.Sprintf("@%s, ти вже відповідала на це завдання 😉", user.Username))
			return nil
			case status.AlreadySkipped:
			return c.Send(fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		}

		//storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(taskID))

        // Check if user has active session
        session, exists := subtasks.GlobalSessionManager.GetActiveSession(game.ID)
        if !exists || session.UserID != user.ID {
			msgTextOtherUserAnswer := fmt.Sprintf("@%s донт пуш зе хорсес! Інша зірочка зараз відповідає.", user.Username)

			_, err = msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				msgmanager.TypeNotInGame,
				msgTextOtherUserAnswer,
				10 * time.Second,
			)
			if err != nil {
					utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
			}

            return nil
        }
        
        utils.Logger.WithFields(logrus.Fields{
            "source":            "HandleSubTask3",
            "username":          user.Username,
            "task_id":           taskID,
            "question_index":    uint(questionIndex),
            "selected_user_id":  selectedUserID,
            "selected_username": selectedUsername,
        }).Infof("User %s selected user %s", user.Username, selectedUsername)
        
        // Delete the question message
        err = bot.Delete(msg)
        if err != nil {
            utils.Logger.Errorf("Failed to delete message: %v", err)
        }
        
        // Save answer and check if completed
        completed, err := subtasks.GlobalSessionManager.SaveAnswerAndNext(game.ID, selectedUsername)
        if err != nil {
            utils.Logger.Errorf("Error saving subtask answer: %v", err)
            return c.Send("Помилка збереження відповіді")
        }

		subTaskAnswer := &models.SubtaskAnswer{
			GameID: game.ID,
			TaskID: taskID,
			QuestionIndex: uint(questionIndex),
			AnswererUserID: user.ID,
			SelectedUserID: selectedUserID,
			SelectedUsername: selectedUsername,
		}

		err = storage_db.AddSubtaskAnswer(subTaskAnswer)
		if err != nil {
			utils.Logger.Errorf("Error add subtask answer to DB: %v", err)
		} else {
			utils.Logger.Infof("Answe of subtask add to DB: succes")
		}
        
        if completed {
            // All questions answered
            answers := subtasks.GlobalSessionManager.CompleteSession(game.ID)
            
            utils.Logger.WithFields(logrus.Fields{
                "source":        "HandleSubTask3",
                "username":      user.Username,
                "total_answers": len(answers),
                "task_id":       taskID,
            }).Info("Subtask completed")

			playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
				UserName: 	user.Username,
				GameID: 	game.ID,
				TaskID:		taskID,
				HasResponse: true,
				Skipped: false,
			}

			storage_db.AddPlayerResponse(playerResponse)

			storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)
            
            return c.Send(fmt.Sprintf("@%s, дякую за відповідь, кицю 🐈Очікуй результатів, коли всі подружки поділяться своєю думкою 💁‍♀️", user.Username))
        }
        
        // Send next question
        return subtasks.SendCurrentQuestion(bot, c, game.ID)
    }
}