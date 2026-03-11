package handlers

import (
	"fmt"
	"os"
	"strings"
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

    tasks, err := utils.LoadTasks("internal/data/tasks/tasks_v3.json")
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
	fileGif := telebot.FromDisk(fmt.Sprintf("internal/data/tasks/media_for_tasks/%d.gif", firstTask.ID))

	weight, height, err := utils.GetImageDimensions(fileGif.FileLocal)
	if err != nil {
		utils.Logger.Errorf("Error getting GIF dimensions: %v", err)
		return err
	}

	photo := &telebot.Animation{
		File:   fileGif,
		Caption: msg,
		Width:  weight,
		Height: height,
	}
		
		// create buttons Answer and Skip
	inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

	answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, firstTask.ID)
	skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, firstTask.ID)

	inlineKeys.Inline(
		inlineKeys.Row(answerBtn, skipBtn),
	)

	// utils.Logger.Info("Starting subtask 5 voiting fot test test")
	// 		err = voting.StartSubtask5VotingDirect(bot, chatID, msg, inlineKeys)
	// 		if err != nil {
	// 			utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
	// 			// Можете решить, продолжать ли выполнение или вернуть ошибку
	// 		} else {
	// 			utils.Logger.Info("Successfully started subtask 5 voting")
	// 		}

	// Создаем объект фото
	// photo := &telebot.Photo{
	// 	File:    telebot.FromDisk("internal/data/tasks/media_for_tasks/1.jpg"),
	// 	Caption: msg,
	// }

	_, err = bot.Send(&telebot.Chat{ID: chatID}, photo, inlineKeys, telebot.ModeMarkdown)
	if err != nil {
		utils.Logger.Errorf("SendTasks logs: Error sending first task with media to chat %d: %v", chatID, err)
		_, err = bot.Send(&telebot.Chat{ID: chatID}, msg, inlineKeys, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.Errorf("SendTasks logs: Error sending first task without media to chat %d: %v", chatID, err)
			return err
		}
    }

	// ticker := time.NewTicker(2 * time.Minute) // создаём тикер с интервалом 10 секунд
	// defer ticker.Stop()                        // останавливаем тикер при выходе из main

	// for {
	// 	<-ticker.C // ждём, пока пройдёт 10 секунд
	// 	SendNextTask(bot, int64(game.ID))
	// }

	for {
		// Получаем текущую игру чтобы узнать актуальный task ID
		currentGame, err := storage_db.GetGameById(int(game.ID))
		if err != nil {
			utils.Logger.Errorf("Error getting game: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		var interval time.Duration
		if currentGame.CurrentTaskID == 12 {
			interval = 5 * time.Minute
		} else {
			interval = 30 * time.Second
		}

		utils.Logger.Infof("Next task in %v (current task ID: %d)", interval, currentGame.CurrentTaskID)
		time.Sleep(interval)
		SendNextTask(bot, int64(game.ID))
	}

	// return nil

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

    tasks, err := utils.LoadTasks("internal/data/tasks/tasks_v3.json")
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

	fileGif := telebot.FromDisk(fmt.Sprintf("internal/data/tasks/media_for_tasks/%d.gif", currentTask.ID))

	weight, height, err := utils.GetImageDimensions(fileGif.FileLocal)
	if err != nil {
		utils.Logger.Errorf("Error getting GIF dimensions: %v", err)
	}

	photo := &telebot.Animation{
		File:   fileGif,
		Caption: msg,
		Width:  weight,
		Height: height,
	}
		
		// create buttons Answer and Skip
	inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

	answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, currentTask.ID)
	skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, currentTask.ID)

	inlineKeys.Inline(
		inlineKeys.Row(answerBtn, skipBtn),
	)

	switch currentTask.ID {
		case 10:
			err := voting.StartSubtask5VotingDirect(bot, chat.ID, msg, inlineKeys)
			if err != nil {
				utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
				// Можете решить, продолжать ли выполнение или вернуть ошибку
			} else {
				utils.Logger.Info("Successfully started subtask 5 voting")
			}
		case 12:
			parts := strings.SplitN(currentTask.Description, "\n---PART2---\n", 2)

			if len(parts) == 2 {
				// Part 1 — GIF without buttons
				photo.Caption = currentTask.Tittle + "\n\n" + parts[0]
				_, err = bot.Send(chat, photo, telebot.ModeMarkdown)
				if err != nil {
					utils.Logger.Errorf("Error sending task %d part 1: %v", currentTask.ID, err)
					return err
				}
				// Part 2 — text with buttons
				_, err = bot.Send(chat, parts[1], inlineKeys, telebot.ModeMarkdown)
				if err != nil {
					utils.Logger.Errorf("Error sending task %d part 2: %v", currentTask.ID, err)
					return err
				}
			} else {
				// Default sending — GIF with buttons
				_, err = bot.Send(chat, photo, inlineKeys, telebot.ModeMarkdown)
				if err != nil {
					utils.Logger.Errorf("Error sending task %d: %v", currentTask.ID, err)
					return err
				}
			}
			
		default:
			_, err = bot.Send(chat, photo, inlineKeys, telebot.ModeMarkdown )
			if err != nil {
						
					utils.Logger.Errorf("Error sending task %d as text: %v", currentTask.ID, err)
						
				return err
			}
	}
	// 	case 1:
	// 		utils.Logger.Info("Starting subtask 5 voiting fot test test")
	// 		err := voting.StartSubtask5VotingDirect(bot, chat.ID, msg, inlineKeys)
	// 		if err != nil {
	// 			utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
	// 			// Можете решить, продолжать ли выполнение или вернуть ошибку
	// 		} else {
	// 			utils.Logger.Info("Successfully started subtask 5 voting")
	// 		}
	// 	case 2:
	// 		// video := &telebot.Video{
	// 		// 	File:    telebot.FromDisk("internal/data/tasks/media_for_tasks/2.mp4"),
	// 		// 	Caption: msg,
	// 		// }
	// 		// _, err = bot.Send(chat, video, inlineKeys, telebot.ModeMarkdown)
	// 		// if err != nil {
	// 		// 	utils.Logger.Errorf("Error sending task %d with video: %v", currentTask.ID,err)
	// 		// 	_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown)
	// 		// 	if err != nil {
	// 		// 		return err
	// 		// 	}
	// 		// }
	// 	case 3:
	// 		photo := &telebot.Photo{
	// 			File:	telebot.FromDisk("internal/data/tasks/media_for_tasks/3.jpg"),
	// 			Caption: msg,
	// 		}
	// 		_, err = bot.Send(chat, photo, inlineKeys, telebot.ModeMarkdown)
	// 		if err != nil {
	// 			utils.Logger.Errorf("Error sending task %d with photo: %v", currentTask.ID,err)
	// 			_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 	case 5:
	// 		err := voting.StartSubtask5VotingDirect(bot, chat.ID, msg, inlineKeys)
	// 		if err != nil {
	// 			utils.Logger.Errorf("Error starting subtask 5 voting: %v", err)
	// 			// Можете решить, продолжать ли выполнение или вернуть ошибку
	// 		} else {
	// 			utils.Logger.Info("Successfully started subtask 5 voting")
	// 		}
	// 	default:
	// 		_, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown )
	// 		if err != nil {
	// 			return err
	// 		}
	// }

	if _, err := os.Stat(fileGif.FileLocal); os.IsNotExist(err) {
		utils.Logger.Warnf("GIF file not found for task %d: %s", currentTask.ID, fileGif.FileLocal)
	}

	// main functionality for sending task with media
	// _, err = bot.Send(chat, photo, inlineKeys, telebot.ModeMarkdown )
	// 		if err != nil {
				 
	// 				utils.Logger.Errorf("Error sending task %d as text: %v", currentTask.ID, err)
				
	// 			return err
	// 		}

// 	if err != nil {
//     utils.Logger.Errorf("Error sending task %d: %v", currentTask.ID, err)
//     _, err = bot.Send(chat, msg, inlineKeys, telebot.ModeMarkdown)
//     if err != nil {
//         utils.Logger.Errorf("Error sending task %d as text: %v", currentTask.ID, err)
//     }
// }
	
	utils.Logger.WithFields(logrus.Fields{
        "game_id": game.ID,
        "task_id": currentTask.ID,
        "chat_id": game.GameChatID,
    }).Info("Successfully sent next task")

	return  nil
}
