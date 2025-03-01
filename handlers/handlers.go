package handlers

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

// Handler for /start
func StartHandler(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        userName := c.Sender().Username
       
        msg := fmt.Sprintf(
            "Привіт, %s! Я чат-бот для гри з подругами!\n Якщо ти хочеш створити нову гру, викличи команду /newgame",
            userName,
        )
        return c.Send(msg)
    }
}