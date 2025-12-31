package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func HandlerPlayerResponse(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()
		msg := c.Message()

		isUserAdminByBot := utils.IsAdminByBot(cfg.AdminIDs, user.ID)

		// Check if the message is in a private chat
		if chat.Type == telebot.ChatPrivate {
			if isUserAdminByBot {
				// Check if there is a current working chat for this admin
				targetChatID, exists := currentAdminWorkingChat[user.ID]
				if !exists {
					msgText := "Спочатку виберіть запрос за допомогою /select_request"
					bot.Send(chat, msgText)
					return nil
				}

				// Check if there is a photo in the message
				if msg.Photo == nil {
					msgText := "Будь ласка, надішліть фото колажу."
					bot.Send(chat, msgText)
					return nil
				}

				game, err := storage_db.GetGameByChatId(targetChatID)
				if err != nil {
					utils.Logger.WithFields(logrus.Fields{
						"source": "HandlerPlayerResponse",
						"chat_id": targetChatID,
						"user_called": user.Username,
					}).Errorf("Error getting game by chat ID: %v", err)
					
					msgText := "Помилка: гру не знайдено для вказаного чату."
					bot.Send(chat, msgText)
					return nil
				}

				targetChat := &telebot.Chat{ID: targetChatID}
				
				// // Send the collage photo to the target chat
				_, err = bot.Send(targetChat, msg.Photo)
				if err != nil {
					utils.Logger.Errorf("Failed to send collage to target chat %d: %v", targetChatID, err)
					msgText := "Помилка при відправці колажу в чат."
					bot.Send(chat, msgText)
					return nil
				}

				// Update the request status to "done"
				err = storage_db.UpdateCollageRequestStatus(int64(game.ID), targetChatID, models.StatusReqCollageDone)
				if err != nil {
					utils.Logger.Errorf("Error updating request status to done: %v", err)
				}

				// Clear the current working chat
				delete(currentAdminWorkingChat, user.ID)

				// Delete the folder with images
				chatIDStr := strings.TrimPrefix(strconv.FormatInt(targetChatID, 10), "-")
				folderPath := fmt.Sprintf("./storage_temp/%s", chatIDStr)
				err = os.RemoveAll(folderPath)
				if err != nil {
					utils.Logger.Errorf("Error removing folder %s: %v", folderPath, err)
				}

				msgText := fmt.Sprintf("Колаж успішно відправлено в чат %d!", targetChatID)
				bot.Send(chat, msgText)

				return nil
			} else {
				// Infom user that the bot works only in groups
				msgText := "Цей бот працює тільки в групах. Приєднайтесь до групи для участі в грі."
				bot.Send(chat, msgText)
				return nil
			}
		}
		
		// Check if the message is part of an album
		isAlbumProcessed := false
		if msg.AlbumID != "" {
			if _, exists := processedAlbums[msg.AlbumID]; exists {
				isAlbumProcessed = true
			} else {
				// Register the album and set it to clear after 2 minutes
				processedAlbums[msg.AlbumID] = time.Now()
				// Delay delete album ID for group media msg
				time.AfterFunc(cfg.Durations.TimeDeleteAlbumId, func() {
					delete(processedAlbums, msg.AlbumID)
				})
			}
		}

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandlePlayerResponse",
				"chat_id": chat.ID,
				"user_called": user.Username,
			}).Errorf("Error getting game by chat ID: %v", err)
			
			return nil
		}

		statusUser, err := storage_db.GetStatusPlayer(user.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandlerPlayerResponse",
				"user_id": user.ID,
				"username": user.Username,
				"group": chat.Title,
			}).Errorf("Error getting status player: %v", err)
		
			return nil
		}

		utils.Logger.Infof("HandlerPlayerResponse called by user %s in group %s with status %s", user.Username, chat.Title, statusUser)

		utils.Logger.WithFields(logrus.Fields{
				"source": "HandlerPlayerResponse",
				"username": user.Username,
				"group": chat.Title,
				"status_uer_from_DB": statusUser,
				"status_user_in_blocK_if": models.StatusPlayerWaiting+strconv.Itoa(game.CurrentTaskID),
			}).Info("Info about player and his status")

		if statusUser == models.StatusPlayerNoWaiting {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandlerPlayerResponse",
				"user_id": user.ID,
				"username": user.Username,		
				"group": chat.Title,
			}).Warnf("User %s is not waiting for task %d, current status: %s", user.Username, game.CurrentTaskID, statusUser)
			
			// Skip message of user he already answered
			return nil
		}
		
		userTaskID, _ := utils.GetWaitingTaskID(statusUser)

		// Switch case for different task IDs
        switch userTaskID {
		case 1: // for the task 5
			HandleSubTask5Response(bot)(msg)
			return nil
        case 13:
            // Handle photo saving for task 1
            if msg.Photo != nil {
                err := savePhotosForTask13(bot, chat.ID, msg)
                if err != nil {
                    utils.Logger.WithFields(logrus.Fields{
                        "source":   "HandlerPlayerResponse",
                        "user_id":  user.ID,
                        "username": user.Username,
                        "chat_id":  chat.ID,
                    }).Errorf("Error saving photos for task 1: %v", err)
                    return nil
                }

                if msg.AlbumID == "" || !isAlbumProcessed {
                    storage_db.AddCollageRequest(int64(game.ID), chat.ID, models.StatusReqCollageWaiting)
                    
                    utils.Logger.WithFields(logrus.Fields{
                        "source":   "HandlerPlayerResponse",
                        "game_id":  game.ID,
                        "chat_id":  chat.ID,
                        "album_id": msg.AlbumID,
                        "is_album": msg.AlbumID != "",
                    }).Info("Collage request added to database")
                }
                
                // If this is part of an album but not the first message, don't process response
                if isAlbumProcessed {
                    return nil
                }
            }

        case 3:
            // Skip messages from user. User answered subtask
            return nil

        default:
            // For non-photo messages in albums, skip if already processed
            if isAlbumProcessed {
                return nil
            }
        }

		playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
				UserName: user.Username,
				GameID: 	game.ID,
				TaskID:		userTaskID,
				HasResponse: true,
				Skipped: false,
				//DateCreate: time.Now().Unix(),
				NotificationSent: 0,
			}

			storage_db.AddPlayerResponse(playerResponse)

			bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAnswerAccepted), user.Username, userTaskID))

			storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)

		return nil
	}
}


func savePhotosForTask13(bot *telebot.Bot, chatID int64, msg *telebot.Message) error {
    chatIDStr := strings.TrimPrefix(strconv.FormatInt(chatID, 10), "-")
    dirPath := filepath.Join("storage_temp", chatIDStr)
    
    if err := os.MkdirAll(dirPath, 0755); err != nil {
        return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
    }

    photo := msg.Photo
    if photo == nil {
        return fmt.Errorf("no photo in message")
    }

    fileReader, err := bot.File(&photo.File)
    if err != nil {
        return fmt.Errorf("failed to get file: %w", err)
    }
    defer fileReader.Close()
	
    filename := fmt.Sprintf("%d_%s.jpg", time.Now().UnixNano(), photo.File.UniqueID)
    filePath := filepath.Join(dirPath, filename)

    counter := 1
    originalPath := filePath
    for {
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            break
        }
        filePath = fmt.Sprintf("%s_%d%s", 
            strings.TrimSuffix(originalPath, ".jpg"), 
            counter, 
            ".jpg")
        counter++
    }

    outFile, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create file %s: %w", filePath, err)
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, fileReader)
    if err != nil {
        return fmt.Errorf("failed to save file content: %w", err)
    }

    utils.Logger.WithFields(logrus.Fields{
        "source":        "savePhotosForTask1",
        "chat_id":       chatID,
        "file_path":     filePath,
        "filename":      filename,
        "clean_chat_id": chatIDStr,
        "unique_id":     photo.File.UniqueID,
    }).Info("Photo saved successfully")

    return nil
}