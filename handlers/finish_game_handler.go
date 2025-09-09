package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/btnmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func FinishGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "FinishGameHandler",
				"group": chat.Title,
				"err": err,
			}).Error("Finish game handler called")
			
			return nil
		}
		
		utils.Logger.WithFields(logrus.Fields{
			"source": "FinishGameHandler",
			"group": chat.Title,
		}).Info("Finish game handler called")
		
		_, err = bot.Send(&telebot.Chat{ID: chat.ID}, finishMessage, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "FinishGameHandler",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending final message to the group")
		}

		storage_db.UpdateGameStatus(int64(game.ID), models.StatusGameFinished)

		time.Sleep(5 * time.Second)

		SendReferalMsg(bot)(c)

		time.Sleep(5 * time.Second)

		SendFeedbackMsg(bot)(c)

		time.Sleep(5 * time.Second)

		SendBuyMeCoffeeMsg(bot)(c)

		return nil
	}
}

func SendReferalMsg(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		userCalled := c.Sender()

		var refLink string

		// Get game by chat ID
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source":   "SendReferalMsg",
				"group":    chat.Title,
				"group_id": chat.ID,
				"user_id":  userCalled.ID,
				"username": userCalled.Username,
			}).Warnf("Game not found or error occurred: %v", err)

			// Use link to the user who called the command if game not found
			refLink = utils.GenerateInviteLink(int(userCalled.ID))
		} else {
			// Get admin player to generate referral link
			adminPlayer, err := storage_db.GetAdminPlayerByGameID(game.ID)
			if err != nil {
				utils.Logger.WithFields(logrus.Fields{
					"source":   "SendReferalMsg",
					"group":    chat.Title,
					"group_id": chat.ID,
					"user_id":  userCalled.ID,
					"username": userCalled.Username,
				}).Warnf("Admin not found, fallback to sender: %v", err)

				refLink = utils.GenerateInviteLink(int(userCalled.ID))
			} else {
				refLink = utils.GenerateInviteLink(int(adminPlayer.ID))
			}
		}

		// Create message with social media links
		msg := referalMsg
		msg = strings.ReplaceAll(msg, "Instagram", fmt.Sprintf(`<a href="%s">Instagram</a>`, utils.GetStaticMessage(socialMediaLinks, models.LinkInstagram)))
		msg = strings.ReplaceAll(msg, "TikTok", fmt.Sprintf(`<a href="%s">TikTok</a>`, utils.GetStaticMessage(socialMediaLinks, models.LinkTikTok)))
		msg = strings.ReplaceAll(
			msg,
			"Ось твоє Космічне посилання, за яким подружки і подружки подружок зможуть зіграти у власну гру BESTIEVERSE",
			fmt.Sprintf(`<a href="%s">Ось твоє Космічне посилання, за яким подружки і подружки подружок зможуть зіграти у власну гру BESTIEVERSE</a>`, refLink),
		)

		// Sending message
		_, err = bot.Send(chat, msg, &telebot.SendOptions{
			ParseMode:             telebot.ModeHTML,
			DisableWebPagePreview: true,
		})
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source":   "SendReferalMsg",
				"group":    chat.Title,
				"group_id": chat.ID,
				"err":      err,
			}).Error("Error sending referral message to the group")
			return err
		}

		utils.Logger.WithFields(logrus.Fields{
			"group": chat.Title,
			"user":  userCalled.Username,
			"link":  refLink,
		}).Info("Referral message sent successfully")

		return nil
	}
}

// SendFeedbackMsg sends a feedback message to the users
func SendFeedbackMsg(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		feedbackMenu := &telebot.ReplyMarkup{}

		feedbackBtn := btnmanager.Get(feedbackMenu, models.UniqueFeedback)

		feedbackMenu.Inline(
			feedbackMenu.Row(feedbackBtn),
		)

		// startMenu.Inline(
		// 	startMenu.Row(startBtnSupport),
		// )
		_, err := bot.Send(chat, feedbackMsg, feedbackMenu)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "SendFeedbackMsg",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending feedback message to the group")
			return err
		}

		utils.Logger.Info("Feedback message sent successfully")

		return nil
	}
}

func SendBuyMeCoffeeMsg(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		_, err := bot.Send(chat, buyMeCoffeeMsg, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "SendBuyMeCoffeeMsg",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending Buy Me Coffee message to the group")
			return err
		}

		utils.Logger.Info("Buy Me Coffee message sent successfully")

		return nil
	}
}