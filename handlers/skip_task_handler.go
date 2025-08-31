package handlers

import (
	"fmt"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

// Handler for skipping a task
func OnSkipTaskBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		utils.Logger.Info("OnSkipTaskHandler called")

		user := c.Sender()
		chat := c.Chat()
		dataButton := c.Data()
		game, _ := storage_db.GetGameByChatId(chat.ID)
		//statusUser, err := storage_db.GetStatusPlayer(user.ID)
		userTaskID, err := utils.GetSkipTaskID(dataButton)
		if err != nil {
			utils.Logger.Errorf("Error getting skip task ID from data button: %v", err)
			return nil
		}

		utils.Logger.WithFields(logrus.Fields{
			"source": "OnSkipTaskBtnHandler",
			"user": user.Username,
			"group": chat.Title,
			"data_button": dataButton,
			"skip_task_id": userTaskID,
		}).Infof("User click to button SkipTask from tasl %v", dataButton)

		userIsInGame, err := storage_db.IsUserInGame(user.ID, game.ID)
		if err != nil {
			utils.Logger.Errorf("Error checking if user is in game: %v", err)
			return nil
		}
		if !userIsInGame {
			SendJoinGameReminder(bot)(c)

			return nil
		}

		status, err := storage_db.SkipPlayerResponse(user.ID, game.ID, userTaskID)
		if err != nil {
			utils.Logger.Errorf("Error skipping task %d bu user: %v. %v", userTaskID, user.Username, err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
		case status.AlreadySkipped:
			bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		case status.SkipLimitReached:
			msg, _ := bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipLimitReached), user.Username))
			
			// Delay delete the message max skip tasks
			time.AfterFunc(cfg.Durations.TimeDeleteMsgMaxSkipTasks, func() {
				err = bot.Delete(msg)
				if err != nil {
					utils.Logger.Errorf("Error deleting skip limit reached message for user %s: %v", user.Username, err)
				}
			})
		default:
			switch status.RemainingSkips-1 {
			case 0:
				bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipThirdTime), user.Username))
			case 1:
				bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipSecondTime), user.Username))
			case 2:
				bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipFirstTime), user.Username))
			}
			// Skip messages
			//bot.Send(chat, fmt.Sprintf("✅ @%s, завдання пропущено! У тебе залишилось %d пропуск(ів).", user.Username, status.RemainingSkips-1))
		}

		return nil
	}
}