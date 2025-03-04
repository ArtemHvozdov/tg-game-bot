package handlers

import (
	//"fmt"

	"gopkg.in/telebot.v3"
)

// Handler for /start
func StartHandler(bot *telebot.Bot, btnCreateGame, btnJoinGame telebot.Btn) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        startMsg := "Оу, привіт, зіронько! 🌟 Хочеш створити гру для своїх найкращих подруг? Натискай кнопку нижче і вперед до пригод!"

		// Create keyboard
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		menu.Reply(menu.Row(btnCreateGame, btnJoinGame))

		return c.Send(startMsg, menu)
    }
}

// Handler create game
func CreateGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		return c.Send("🎲 Гру створено! Надішли код друзям, щоб вони могли приєднатися!")
	}
}

// Handler join to game
func JoinGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		return c.Send("🔑 Введи код гри, щоб приєднатися!")
	}
}