package handlers

import (
	"strings"

	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func RegisterCallbackHandlers(bot *telebot.Bot) {
	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := c.Callback().Data

		utils.Logger.WithFields(logrus.Fields{
			"source": "RegisterCallbackHandlers",
			"data": data,
			"user_id": c.Sender().ID,
			"username": c.Sender().Username,
			"group": c.Chat().Title,
		}).Info("Callback handler called")

		switch {
		case strings.HasPrefix(data, "\fexit_game_"):
			return handleExitGame(bot, c)
		case strings.HasPrefix(data, "\fexit_"):
			return handleExitConfirm(bot, c)
		// case data == "support_menu":
		// 	return handleSupportMenu(bot, c)
		case data == "\fhelp_menu":
			return handleHelpMenu(bot, c)
		case data == "\freturn_to_game":
			return handleReturnToGame(bot, c)
		// case strings.HasPrefix(data, "\fphoto_choice_"):
		// 	return HandlePhotoChoice(bot)(c)
		case strings.HasPrefix(data,"\fwaiting_" ):
			return OnAnswerTaskBtnHandler(bot)(c)
		case strings.HasPrefix(data,"\fskip_"):
			return OnSkipTaskBtnHandler(bot)(c)
		case strings.HasPrefix(data, "subtask_3_"):
			return HandleSubTask3(bot)(c)
		default:
			return nil
		}
	})
}