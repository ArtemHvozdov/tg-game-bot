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

	_, err = bot.Send(&telebot.Chat{ID: chatID}, msg, inlineKeys, telebot.ModeMarkdown )
	if err != nil {
		return err
    }

    // for i, task := range tasks {
    //     //task := tasks[i]
	// 	timeUpdate := time.Now().Unix()
	// 	storage_db.UpdateCurrentTaskID(game.ID, task.ID, timeUpdate)
    //     // msg := "🌟 *" + task.Tittle + "*\n" + task.Description

	// 	msg := task.Tittle + "\n\n" + task.Description
		
	// 	// create buttons Answer and Skip
	// 	inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

	// 	//answerBtn := inlineKeys.Data("Хочу відповісти", "answer_task", fmt.Sprintf("waiting_%d", task.ID))
	// 	//skipBtn := inlineKeys.Data("Пропустити", "skip_task", fmt.Sprintf("skip_%d", task.ID))
	// 	answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, task.ID)
	// 	skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, task.ID)

	// 	inlineKeys.Inline(
	// 		inlineKeys.Row(answerBtn, skipBtn),
	// 	)

	// 	if i == 4 {
	// 		err := voting.StartSubtask5VotingDirect(bot, chatID, msg, inlineKeys)
	// 		if err != nil {
	// 			utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
	// 			// Можете решить, продолжать ли выполнение или вернуть ошибку
	// 		} else {
	// 			utils.Logger.Info("Successfully started subtask 5 voting")
	// 		}

	// 		//return nil
	// 	} else {
	// 		_, err := bot.Send(
	// 			&telebot.Chat{ID: chatID},
	// 			msg,
	// 			inlineKeys,
	// 			telebot.ModeMarkdown,
	// 		)
	// 		if err != nil {
	// 			return err
    //     	}
	// 	}

    //     // _, err := bot.Send(
    //     //     &telebot.Chat{ID: chatID},
    //     //     msg,
	// 	// 	inlineKeys,
    //     //     telebot.ModeMarkdown,
    //     // )
    //     // if err != nil {
    //     //     return err
    //     // }

	// 	if i < len(tasks)-1 {
	// 		// i == 2 || i == 4 || i == 9
	// 		// if i == 4 {
	// 		// 	time.Sleep(25 * time.Minute) // Wait for 5 seconds before sending the next task
	// 		// } else {
	// 		// 	time.Sleep(cfg.Durations.TimePauseBetweenSendingTasks)
	// 		// }
	// 		// Delay pause between sending tasks
	// 		//time.Sleep(15 * time.Minute)
	// 		if (cfg.Durations.TimePauseBetweenSendingTasks == 15*time.Minute) {
	// 			utils.Logger.Info("Value of time delay from config is 15 minute! The all is the ok!")
	// 		} else {
	// 			utils.Logger.Warn("Value of time delay from config is not 3 minute! Check the config!")
	// 		}
	// 		time.Sleep(cfg.Durations.TimePauseBetweenSendingTasks) // await some minutes or hours before sending the next task
	// 	}

    // }

	//return FinishGameHandler(bot, chat)

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

	if currentTask.ID == 5 {
		err := voting.StartSubtask5VotingDirect(bot, chat.ID, msg, inlineKeys)
		if err != nil {
			utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
			// Можете решить, продолжать ли выполнение или вернуть ошибку
		} else {
			utils.Logger.Info("Successfully started subtask 5 voting")
		}

		//return nil
	} else {
		_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown )
		if err != nil {
			return err
		}
	}

	// _, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown )
	// if err != nil {
	// 	return err
    // }

	utils.Logger.WithFields(logrus.Fields{
        "game_id": game.ID,
        "task_id": currentTask.ID,
        "chat_id": game.GameChatID,
    }).Info("Successfully sent next task")

	return  nil
}



// // handlers/task_sender.go (обновленная версия)
// func SendNextTask1(bot *telebot.Bot, gameID int64) error {
//     game, err := storage_db.GetGameById(int(gameID))
//     if err != nil {
//         return fmt.Errorf("failed to get game: %w", err)
//     }

//     tasks, err := utils.LoadTasks("internal/data/tasks/tasks.json")
//     if err != nil {
//         return fmt.Errorf("failed to load tasks: %w", err)
//     }

//     nextTaskID := game.CurrentTaskID + 1
    
//     if nextTaskID > len(tasks) {
//         utils.Logger.Info("All tasks completed, finishing game")
//         return finishGame(bot, game.GameChatID)
//     }

//     var currentTask *models.Task
//     for _, task := range tasks {
//         if task.ID == nextTaskID {
//             currentTask = &task
//             break
//         }
//     }

//     if currentTask == nil {
//         return fmt.Errorf("task with ID %d not found", nextTaskID)
//     }

//     // Обновляем текущую таску
//     timeUpdate := time.Now().Unix()
//     err = storage_db.UpdateCurrentTaskID(game.ID, currentTask.ID, timeUpdate)
//     if err != nil {
//         return fmt.Errorf("failed to update current task: %w", err)
//     }

//     // Отправляем таску
//     err = sendTaskToChat(bot, game.GameChatID, currentTask)
//     if err != nil {
//         return fmt.Errorf("failed to send task: %w", err)
//     }

//     utils.Logger.WithFields(logrus.Fields{
//         "game_id": game.ID,
//         "task_id": currentTask.ID,
//         "chat_id": game.GameChatID,
//     }).Info("Successfully sent next task")

//     return nil
// }