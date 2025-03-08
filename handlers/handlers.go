package handlers

import (

	"fmt"
	"strconv"
	"strings"

	"log"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	//"github.com/ArtemHvozdov/tg-game-bot.git/utils"


	"gopkg.in/telebot.v3"
)

type AwaiteState struct {
	StartStateAwait bool
	NameGameRoomAwait bool
	NameGameAwait bool
	QuestionsAwait bool
}


var botState = AwaiteState{}

// Handler for /start
func StartHandler(bot *telebot.Bot, btnCreateGame, btnJoinGame telebot.Btn) func(c telebot.Context) error {

	return func(c telebot.Context) error {
		startMsg := "Оу, привіт, зіронько! 🌟 Хочеш створити гру для своїх найкращих подруг? Натискай кнопку нижче і вперед до пригод!"


		// Create keyboard
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		menu.Reply(menu.Row(btnCreateGame, btnJoinGame))


		// Получаем ID игровой комнаты из инвайт-ссылки
		inviteData := c.Data() // Получаем строку, переданную после /start
		gameRoomID, err := strconv.Atoi(inviteData)
		if err == nil && gameRoomID > 0 { // Проверяем, что получили корректное число
			// Получаем информацию о комнате из базы данных
			gameRoom, err := storage_db.GetGameRoomByID(gameRoomID)
			if err != nil || gameRoom == nil {
				return c.Send("❌ Не вдалося знайти гру за цим посиланням.")
			}

			// Отправляем информацию о комнате пользователю
			successMsg := fmt.Sprintf(
				"Вітаємо! Ви приєдналися до гри '%s'. Запрошуємо до участі!\nІнвайт-лінк: %s",
				gameRoom.Title, gameRoom.InviteLink,
			)
			return c.Send(successMsg)
		}

		// Если это не инвайт-ссылка, показываем стартовое сообщение
		return c.Send(startMsg, menu)
	}
}


// Функция для извлечения ID из инвайт-ссылки
func extractGameRoomID(link string) string {
	parts := strings.Split(link, "start=")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}


// Handler create game
// func CreateGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
// 	return func(c telebot.Context) error {
// 		// Запрашиваем у пользователя название игры
// 		if err := c.Send("🎲 Введіть назву ігрової кінати:"); err != nil {
// 			return err
// 		}

// 		// Ожидаем следующий текстовый ввод с названием игры
// 		bot.Handle(telebot.OnText, func(tc telebot.Context) error {
// 			gameTitle := tc.Text()

// 			// Генерируем уникальную ссылку для приглашения
// 			inviteLink := utils.GenerateInviteLink()

// 			// Создаём экземпляр структуры GameRoom
// 			gameRoom := models.GameRoom{
// 				Title:      gameTitle,
// 				InviteLink: inviteLink,
// 			}

// 			// Записываем игровую комнату в БД
// 			err := storage_db.CreateGameRoom(gameRoom)
// 			if err != nil {
// 				log.Println("Ошибка создания игровой комнаты:", err)
// 				return tc.Send("❌ Виникла помилка при створенні гри. Спробуйте ще раз!")
// 			}

// 			// Отправляем пользователю ссылку на игру
// 			successMsg := "✅ Гру '" + gameTitle + "' створено!\n" +
// 				"Запросіть своїх друзів за цим посиланням: " + inviteLink
// 			return tc.Send(successMsg)
// 		})

// 		return nil
// 	}
	
// }

// Handler create game
// Handler для создания игры
func CreateGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		// Запрашиваем у пользователя название игровой комнаты
		if err := c.Send("🎲 Введіть назву ігрової кімнати:"); err != nil {
			return err
		}

		bot.Handle(telebot.OnText, func(tc telebot.Context) error {
			gameRoomTitle := tc.Text()

			// Создаем игровую комнату (БЕЗ СОЗДАНИЯ В БД)
			gameRoom := models.GameRoom{
				Title: gameRoomTitle,
			}

			// Запрашиваем у пользователя название игры
			if err := tc.Send("📛 Введіть назву гри:"); err != nil {
				return err
			}

			bot.Handle(telebot.OnText, func(tc telebot.Context) error {
				gameTitle := tc.Text()

				// Создаем игру в БД и получаем её ID
				game := models.Game{
					Name:   gameTitle,
					Status: "waiting",
				}

				gameID, err := storage_db.CreateGame(game)
				if err != nil {
					log.Println("Ошибка создания игры:", err)
					return tc.Send("❌ Виникла помилка при створенні гри. Спробуйте ще раз!")
				}

				// Теперь создаем GameRoom и передаем туда GameID
				gameRoom.GameID = &gameID
				_, inviteLink, err := storage_db.CreateGameRoom(gameRoom)
				if err != nil {
					log.Println("Ошибка создания игровой комнаты:", err)
					return tc.Send("❌ Виникла помилка при створенні кімнати. Спробуйте ще раз!")
				}

				// Просим пользователя ввести вопросы
				if err := tc.Send("❓ Введіть 4 питання та відповіді у форматі:\n`Питання 1 | Відповідь 1`"); err != nil {
					return err
				}

				var tasks []models.Task

				bot.Handle(telebot.OnText, func(tc telebot.Context) error {
					text := tc.Text()
					parts := strings.SplitN(text, "|", 2)

					if len(parts) != 2 {
						return tc.Send("⚠ Неправильний формат! Використовуйте `Питання | Відповідь`")
					}

					question := strings.TrimSpace(parts[0])
					answer := strings.TrimSpace(parts[1])

					tasks = append(tasks, models.Task{
						GameID:   gameID,
						Question: question,
						Answer:   answer,
					})

					// Если получили 4 вопроса, записываем их в БД
					if len(tasks) == 2 {
						for _, task := range tasks {
							storage_db.CreateTask(task)
						}

						successMsg := fmt.Sprintf(
							"✅ Гру '%s' створено!\nЗапросіть своїх друзів за цим посиланням: %s",
							gameTitle, inviteLink,
						)
						return tc.Send(successMsg)
					}

					// Просим следующий вопрос
					return tc.Send(fmt.Sprintf("📝 Введіть питання %d:", len(tasks)+1))
				})

				return nil
			})

			return nil
		})

		return nil
	}
}





// Handler join to game
func JoinGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		return c.Send("🔑 Введи код гри, щоб приєднатися!")
	}
}