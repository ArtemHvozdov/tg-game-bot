package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
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

		utils.Logger.Infof("User %s is answering to task %d in game %d", user.Username, idTask, game.ID)


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
		// this case is for quick test of subtask memes
		// case 1:
		// 	handleSubTask53(c)
		// 	return nil
		case 2:
			session, exists := quizdna.GlobalSubtask2SessionManager.GetActiveSession(game.ID)
			if exists && session.UserID == user.ID {
				msgTextOtherUserAnswer := fmt.Sprintf("@%s ти вже відповідаєш на це питання", user.Username)

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

			quizdna.WhoIsUsSubTask2(bot)(c)
			return nil
		case 4:
			session, exists := subtasks.GlobalSessionManager.GetActiveSession(game.ID)
			if exists && session.UserID == user.ID {
				msgTextOtherUserAnswer := fmt.Sprintf("@%s ти вже відповідаєш на це питання", user.Username)

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
			subtasks.WhoIsUsSubTask(bot)(c)
			return nil
		case 10:
			if dataButton == "\fwaiting_10_3" {
				handleSubTask13(c)
				return nil
			}
		case 12:
			userRole, err := storage_db.GetPlayerRoleByUserIDAndGameID(user.ID, game.ID)
			if err != nil {
				utils.Logger.Infof("Error getting player role for user %s in game %d during answering task %d: %v", user.Username, game.ID, idTask, err)
			}

			utils.Logger.Infof("User %s has role %s in game %d during answering task %d", user.Username, userRole, game.ID, idTask)
			if userRole != "admin" {
				msgTextOnlyAdmin := "⛔ *Стоп-кадр!*\nФінальне слово у цьому завданні – за *Адміном* 💼 \nАле твоя роль не менш важлива – продовжуй мріяти у чаті 💬"
				_, err = msgmanager.SendTemporaryMessage(
					chat.ID,
					user.ID,
					msgmanager.TypeNotInGame,
					msgTextOnlyAdmin,
					5*time.Second,
					telebot.ModeMarkdown,
				)
				if err != nil {
					utils.Logger.Errorf("Error sending message that only admin can answer task for user %s: %v", user.Username, err)
				}
				return nil
			}

			// if userRole == "admin" {
			// 	HandleSubTask12(c)
			// } 
			// else {
			// 	msgTextOnlyAdmin := "⛔ *Стоп-кадр!*\nФінальне слово у цьому завданні – за *Адміном* 💼 \nАле твоя роль не менш важлива – продовжуй мріяти у чаті 💬"

			// 	_, err = msgmanager.SendTemporaryMessage(
			// 		chat.ID,
			// 		user.ID,
			// 		msgmanager.TypeNotInGame,
			// 		msgTextOnlyAdmin,
			// 		10 * time.Second,
			// 		telebot.ModeMarkdown,
			// 	)
			// 	if err != nil {
			// 		utils.Logger.Errorf("Error sending message that only admin can answer task for user %s: %v", user.Username, err)
			// 	}
			// }

			// Проверяем: все 7 ответов уже есть
			allAnswered, err := storage_db.HasAllTask12Answers(int64(game.ID), chat.ID)
			if err != nil {
				utils.Logger.Errorf("Error checking task 12 answers: %v", err)
				return nil
			}
			if allAnswered {
				msg, _ := bot.Send(chat, fmt.Sprintf("@%s, ти вже відповів на всі питання цього завдання 😊", user.Username))
				time.AfterFunc(5*time.Second, func() {
					bot.Delete(msg)
				})
				return nil
			}

			// Проверяем: уже отвечают прямо сейчас (есть state в памяти)
			if _, exists := getSubtask12State(chat.ID); exists {
				msg, _ := bot.Send(chat, fmt.Sprintf("@%s, ти прямо зараз відповідаєш на це завдання ⏳", user.Username))
				time.AfterFunc(5*time.Second, func() {
					bot.Delete(msg)
				})
				return nil
			}

			HandleSubTask12(c)

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