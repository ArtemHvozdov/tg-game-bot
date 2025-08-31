package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/subtasks"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/quiz_dna"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

// Handler for answering a task
func OnAnswerTaskBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()
		dataButton := c.Data()
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Error getting game by chat ID (%d): %v", chat.ID, err)
			return nil
		}

		utils.Logger.WithFields(logrus.Fields{
			"source": "OnAnswerTaskBtnHandler",
			"username": user.Username,
			"group": chat.Title,
			"data_button": dataButton,
		}).Infof("User click to button WantAnswer to task %v", dataButton)

		userIsInGame, err := storage_db.IsUserInGame(user.ID, game.ID)
		if err != nil {
			utils.Logger.Errorf("Error checking if user is in game: %v", err)
			return nil
		}
		if !userIsInGame {
			SendJoinGameReminder(bot)(c)

			return nil
		}

		idTask, err := utils.GetWaitingTaskID(dataButton)
		if err != nil {
			utils.Logger.Errorf("Error getting task ID from data button: %v", err)
		}

		// switch idTask {
		// case 3:
		// 	subtasks.WhoIsUsSubTask(bot)(c)
		// 	return nil
		// case 7:
		// 	// call function for subtask for task 7
		// case 12:
		// 	// call function for subtask for task 12
		// }

		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, idTask)
		if err != nil {
			utils.Logger.Errorf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			//textYouAlreadyAnswered := fmt.Sprintf("@%s, ти вже відповіла на це завдання 😅", user.Username)
			msgYouAlreadyAnswered, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s already answered task %d: %v", user.Username, idTask, err)
			}

			time.AfterFunc(cfg.Durations.TimeDeleteMsgYouAlreadyAnswered, func() {
				err = bot.Delete(msgYouAlreadyAnswered)
				if err != nil {
					utils.Logger.WithFields(logrus.Fields{
						"source": "OnAnswerTaskBtnHandler",
						"username": user.Username,
						"group": chat.Title,
						"data_button": dataButton,
						"task_id": idTask,
					}).Errorf("Error deleting message that user %s already answered task %d: %v", user.Username, idTask, err)
				}
			})

			// return c.Send(fmt.Sprintf("@%s, ти вже відповідала на це завдання 😉", user.Username))
			return nil
		case status.AlreadySkipped:
			return c.Send(fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		}

		storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(idTask))

		switch idTask {
		case 3:
			subtasks.WhoIsUsSubTask(bot)(c)
			return nil
		case 7:
			// call function for subtask for task 7
		case 10:
			quizdna.StartQuizDnaTask(bot)(c)
			return nil
		}

		awaitingAnswerMsg, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(wantAnswerMessages), user.Username))
		if err != nil {
			utils.Logger.Errorf("Error sending message: %v", err)
		}

		// Delay delete msg awaiting answer
		time.AfterFunc(cfg.Durations.TimeDeleteMsgAwaitingAnswer, func() {
			err = bot.Delete(awaitingAnswerMsg)
			if err != nil {
				utils.Logger.WithFields(logrus.Fields{
					"source": "OnAnswerTaskBtnHandler",
					"username": user.Username,
					"group": chat.Title,
					"data_button": dataButton,
					"task_id": idTask,
				}).Errorf("Error deleting answer task message for user %s in the group %s: %v", chat.Username, chat.Title, err)
			}
		})

		return nil
	}
}