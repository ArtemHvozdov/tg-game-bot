// Bot handler for displaying subtask 10 results
package handlers

import (
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

// ShowSubtask10Results displays the results of subtask 10 for the current game
func ShowSubtask10Results(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		// Get game by chat ID
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
			return c.Send("Помилка отримання гри")
		}

		// Process subtask 10 results
		resultMessage, err := storage_db.ProcessSubtask10Results(game.ID)
		if err != nil {
			utils.Logger.Errorf("Failed to process subtask 10 results for game %d: %v", game.ID, err)
			return c.Send("Помилка обробки результатів підзавдання 10")
		}

		// Send results message
		return c.Send(resultMessage)
	}
}

// Alternative function that can be called from other handlers
func SendSubtask10ResultsToChat(bot *telebot.Bot, chatID int64, gameID int) error {
	resultMessage, err := storage_db.ProcessSubtask10Results(gameID)
	if err != nil {
		utils.Logger.Errorf("Failed to process subtask 10 results for game %d: %v", gameID, err)
		return err
	}

	_, err = bot.Send(&telebot.Chat{ID: chatID}, resultMessage)
	if err != nil {
		utils.Logger.Errorf("Failed to send subtask 10 results to chat %d: %v", chatID, err)
		return err
	}

	utils.Logger.Infof("Successfully sent subtask 10 results to chat %d for game %d", chatID, gameID)
	return nil
}