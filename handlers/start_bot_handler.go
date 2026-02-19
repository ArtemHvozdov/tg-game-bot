package handlers

import (
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/btnmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func StartBotHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		//chat := c.Chat()
		user := c.Sender()
		
		utils.Logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"username": user.Username,
		}).Info("User started the bot")

		startMsg, err := utils.LoadSingleMessage("./internal/data/messages/personal/start_message.json")
		if err != nil {
			utils.Logger.Errorf("Error loading start message for the private chat with bot: %v", err)
		}

		startMenu := &telebot.ReplyMarkup{}
		//startBtnSupport := startMenu.URL("🕹️ Техпідтримка", "https://t.me/Jay_jayss")

		startBtnSupport := btnmanager.Get(startMenu, models.UniqueSupport)

		startMenu.Inline(
			startMenu.Row(startBtnSupport),
		)

		return c.Send(startMsg, startMenu, telebot.ModeHTML)
 	}
}
