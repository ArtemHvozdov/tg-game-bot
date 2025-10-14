package handlers

import (
	"fmt"
	"strconv"
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

		// ПCheck: if the message is part of an album and has already been processed, ignore it
		if msg.AlbumID != "" {
			if _, exists := processedAlbums[msg.AlbumID]; exists {
				return nil
			}

			// Register the album and set it to clear after 2 minutes
			processedAlbums[msg.AlbumID] = time.Now()

			// Delay delete album ID for group media msg
			time.AfterFunc(cfg.Durations.TimeDeleteAlbumId, func() {
				delete(processedAlbums, msg.AlbumID)
			})
		}

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandleUHandlerPlayerResponseserJoined",
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

		// Skip messges from user. User answered subtask
		if userTaskID == 3 {
			return nil
		}

		playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
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