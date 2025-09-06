package handlers

import (
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/btnmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/voting"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

// SendFirstTasks send all tasks in group chat
func SendTasks(bot *telebot.Bot, chatID int64) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		game, err := storage_db.GetGameByChatId(chatID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"siource": "SendTasks",
			"chat_id": chatID,
			"chat_name": game.Name,
		}).Errorf("Error getting game by chat ID: %v", err)
	
		return err
	}

    tasks, err := utils.LoadTasks("internal/data/tasks/tasks.json")
    if err != nil {
		    utils.Logger.Errorf("SendTasks logs: Error loading tasks: %v", err)
        return err
    }

    if len(tasks) == 0 {
		utils.Logger.Error("SendTasks logs: No tasks to send. Tasks's array is empty" )
		return nil
	}

    for i, task := range tasks {
        //task := tasks[i]
		storage_db.UpdateCurrentTaskID(game.ID, task.ID)
        // msg := "🌟 *" + task.Tittle + "*\n" + task.Description

		msg := task.Tittle + "\n\n" + task.Description
		
		// create buttons Answer and Skip
		inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

		//answerBtn := inlineKeys.Data("Хочу відповісти", "answer_task", fmt.Sprintf("waiting_%d", task.ID))
		//skipBtn := inlineKeys.Data("Пропустити", "skip_task", fmt.Sprintf("skip_%d", task.ID))
		answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, task.ID)
		skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, task.ID)

		inlineKeys.Inline(
			inlineKeys.Row(answerBtn, skipBtn),
		)

		if i == 4 {
			err := voting.StartSubtask5VotingDirect(bot, chatID, msg, inlineKeys)
			if err != nil {
				utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
				// Можете решить, продолжать ли выполнение или вернуть ошибку
			} else {
				utils.Logger.Info("Successfully started subtask 5 voting")
			}

			//return nil
		} else {
			_, err := bot.Send(
				&telebot.Chat{ID: chatID},
				msg,
				inlineKeys,
				telebot.ModeMarkdown,
			)
			if err != nil {
				return err
        	}
		}

        // _, err := bot.Send(
        //     &telebot.Chat{ID: chatID},
        //     msg,
		// 	inlineKeys,
        //     telebot.ModeMarkdown,
        // )
        // if err != nil {
        //     return err
        // }

		if i < len(tasks)-1 {
			// i == 2 || i == 4 || i == 9
			if i == 9 {
				time.Sleep(5 * time.Minute) // Wait for 5 seconds before sending the next task
			}
			// Delay pause between sending tasks
			time.Sleep(cfg.Durations.TimePauseBetweenSendingTasks) // await some minutes or hours before sending the next task
		}

    }

	return FinishGameHandler(bot)(c)

	}
	
}