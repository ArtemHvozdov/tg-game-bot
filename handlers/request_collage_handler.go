package handlers

import (
	"fmt"
	"strings"

	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

func RequestCollageHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()
		adminsIds := cfg.AdminIDs

		isAdminUser := utils.IsAdminByBot(adminsIds, user.ID)
		if !isAdminUser {
			msg := "Вибачте, ця команда доступна лише адміністраторам."
			_, err := bot.Send(chat, msg)
			if err != nil {
				utils.Logger.Errorf("Failed to send admin restriction message to chat %d: %v", chat.ID, err)
			}
			return nil
		}

		// get request collage from db with status waiting
		requests, err := storage_db.GetAllCollageRequests()
		if err != nil {
			utils.Logger.Errorf("Error getting collage requests with status waiting: %v", err)
			return nil
		}

		if len(requests) == 0 {
			msg := "Немає запитів на створення колажу."
			_, err := bot.Send(chat, msg)
			if err != nil {
				utils.Logger.Errorf("Failed to send no requests message to chat %d: %v", chat.ID, err)
			}
			return nil
		}

		// Build markdown message
        var msgBuilder strings.Builder
        msgBuilder.WriteString("*Requests for collage:*\n\n")
        
        for i, request := range requests {
            msgBuilder.WriteString(fmt.Sprintf("%d. %d : %s\n", i+1, request.ChatID, request.Status))
        }

        _, err = bot.Send(chat, msgBuilder.String(), &telebot.SendOptions{
            ParseMode: telebot.ModeMarkdown,
        })
        if err != nil {
            utils.Logger.Errorf("Failed to send requests list to chat %d: %v", chat.ID, err)
        }


		return nil
	}
}