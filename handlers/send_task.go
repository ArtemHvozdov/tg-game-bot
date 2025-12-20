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
	utils.Logger.Infof("SendTasks called")

	//chat := &telebot.Chat{ID: chatID}

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

	utils.Logger.Infof("Sendint first tasks to the chat %d game %d", chatID, game.ID)

	firstTask := tasks[0]

	timeUpdate := time.Now().Unix()
	storage_db.UpdateCurrentTaskID(game.ID, firstTask.ID, timeUpdate)

	msg := firstTask.Tittle + "\n\n" + firstTask.Description
		
		// create buttons Answer and Skip
	inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

	answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, firstTask.ID)
	skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, firstTask.ID)

	inlineKeys.Inline(
		inlineKeys.Row(answerBtn, skipBtn),
	)

	// Создаем объект фото
	photo := &telebot.Photo{
		File:    telebot.FromDisk("internal/data/tasks/media_for_tasks/1.jpg"),
		Caption: msg,
	}

	_, err = bot.Send(&telebot.Chat{ID: chatID}, photo, inlineKeys, telebot.ModeMarkdown)
	if err != nil {
		utils.Logger.Errorf("SendTasks logs: Error sending first task with media to chat %d: %v", chatID, err)
		_, err = bot.Send(&telebot.Chat{ID: chatID}, msg, inlineKeys, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.Errorf("SendTasks logs: Error sending first task without media to chat %d: %v", chatID, err)
			return err
		}
    }

	return nil

	}
	
}

func SendNextTask(bot *telebot.Bot, gameID int64) error {
	game, err := storage_db.GetGameById(int(gameID))
    if err != nil {
        utils.Logger.Errorf("failed to get game: %v", err)
		return err
    }
	
	chat := &telebot.Chat{ID: game.GameChatID}

	utils.Logger.Info(chat.Title)

    tasks, err := utils.LoadTasks("internal/data/tasks/tasks.json")
    if err != nil {
        utils.Logger.Errorf("failed to load tasks: %v", err)
		return err
    }

	nextTaskID := game.CurrentTaskID + 1
    
    if nextTaskID > len(tasks) {
        utils.Logger.Info("All tasks completed, finishing game")
        return FinishGameHandler(bot, chat)
    }

	currentTask := tasks[nextTaskID-1]

	timeUpdate := time.Now().Unix()
	storage_db.UpdateCurrentTaskID(game.ID, nextTaskID, timeUpdate)

	msg := currentTask.Tittle + "\n\n" + currentTask.Description
		
		// create buttons Answer and Skip
	inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

	answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, currentTask.ID)
	skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, currentTask.ID)

	inlineKeys.Inline(
		inlineKeys.Row(answerBtn, skipBtn),
	)

	switch currentTask.ID {
		case 2:
			video := &telebot.Video{
				File:    telebot.FromDisk("internal/data/tasks/media_for_tasks/2.mp4"),
				Caption: msg,
			}
			_, err = bot.Send(chat, video, inlineKeys, telebot.ModeMarkdown)
			if err != nil {
				utils.Logger.Errorf("Error sending task %d with video: %v", currentTask.ID,err)
				_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown)
				if err != nil {
					return err
				}
			}
		case 3:
			photo := &telebot.Photo{
				File:	telebot.FromDisk("internal/data/tasks/media_for_tasks/3.jpg"),
				Caption: msg,
			}
			_, err = bot.Send(chat, photo, inlineKeys, telebot.ModeMarkdown)
			if err != nil {
				utils.Logger.Errorf("Error sending task %d with photo: %v", currentTask.ID,err)
				_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown)
				if err != nil {
					return err
				}
			}
		case 5:
			err := voting.StartSubtask5VotingDirect(bot, chat.ID, msg, inlineKeys)
			if err != nil {
				utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
				// Можете решить, продолжать ли выполнение или вернуть ошибку
			} else {
				utils.Logger.Info("Successfully started subtask 5 voting")
			}
		default:
			_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown )
			if err != nil {
				return err
			}
	}
	
	utils.Logger.WithFields(logrus.Fields{
        "game_id": game.ID,
        "task_id": currentTask.ID,
        "chat_id": game.GameChatID,
    }).Info("Successfully sent next task")

	return  nil
}
