package handlers

import (

	"fmt"
	"strconv"
	"strings"

	"time"


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
func StartHandler(bot *telebot.Bot, btnCreateGame, btnJoinGame, btnHelpMe telebot.Btn) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		startMsg := "Оу, привіт, зіронько! 🌟 Хочеш створити гру для своїх найкращих подруг? Натискай кнопку нижче і вперед до пригод!"

		// Create keyboard
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		//.Reply(menu.Row(btnCreateGame, btnJoinGame, btnHelpMe))

		// Buttons on the first line
		row1 := menu.Row(btnCreateGame, btnJoinGame)
		// Button on the second row (all size)
		row2 := menu.Row(btnHelpMe)

		menu.Reply(row1, row2)

		// Get ID game fron invite-link
		inviteData := c.Data() // Get string before /start
		gameID, err := strconv.Atoi(inviteData)
		if err == nil && gameID > 0 { // Check that number is correct
			// Get info about game from DB
			game, err := storage_db.GetGameById(gameID)
			if err != nil || gameID == 0 {
				return c.Send("❌ Не вдалося знайти гру за цим посиланням.") 
			}

			player := &models.Player{
				ID:       user.ID,
				UserName: user.Username,
				Name:     user.FirstName,
				Passes:   0,
				GameID:   game.ID,
				Role:     "player",
			}

			if err := storage_db.AddPlayerToGame(player); err != nil {
				log.Printf("Failed to add player to game: %v", err)
				return c.Send("Ой, не вдалося додати тебе до гри. Спробуй ще раз!")
			}

			notifyPlayerJoined(bot, game.ID, *player)

			successMsg := fmt.Sprintf(
				"Йей! Ти приєдналася до гри '%s'. 🎉 Чекаємо, поки всі зберуться, і можна буде розпочати.",
				game.Name,
			)
			return c.Send(successMsg)
		}


		// If this is not invite-link, send start-message
		c.Send(startMsg, menu)

		bot.Handle(telebot.OnText, func(tc telebot.Context) error {
			return tc.Send("🤔 Зараз я не очікую від тебе текстових повідомлень. Якщо ти хочеш створити гру або приєднатися до гри, натискай на кнопки нижче.")
		})

		return nil
	}
}

/// Handler create game
func CreateGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		userAdmin := c.Sender()
		// Ask tha name game
		if err := c.Send("🎲 Введіть назву гри:"); err != nil {
			return err
		}

		var gameName string

		bot.Handle(telebot.OnText, func(tc telebot.Context) error {
            gameName = tc.Text()

            game, err := storage_db.CreateGame(gameName)
            if err != nil {
                log.Println("Ошибка создания игры:", err)
                return tc.Send("❌ Виникла помилка при створенні гри. Спробуйте ще раз!")
            }

			playerAdmin := &models.Player{
				ID: userAdmin.ID,
				UserName: userAdmin.Username,
				Name: userAdmin.FirstName,
				Passes: 0,
				GameID: game.ID,
				Role: "admin",
			}

			if err := storage_db.AddPlayerToGame(playerAdmin); err != nil {
				log.Printf("Failed to add player-admin to game: %v", err)
				return c.Send("Ой, не вдалося додати тебе до гри. Спробуй ще раз!")
			}

            // Move on to collecting questions
            return askQuestions(bot, tc, game, 2)
        })

        return nil

		// bot.Handle(telebot.OnText, func(tc telebot.Context) error {
		// 	gameName = tc.Text()

		// 	// Создаем игровую комнату (БЕЗ СОЗДАНИЯ В БД)
		// 	gameRoom := models.GameRoom{
		// 		Title: gameRoomTitle,
		// 	}

		// 	// Запрашиваем у пользователя название игры
		// 	if err := tc.Send("📛 Введіть назву гри:"); err != nil {
		// 		return err
		// 	}

		// 	bot.Handle(telebot.OnText, func(tc telebot.Context) error {
		// 		gameTitle := tc.Text()

		// 		// Создаем игру в БД и получаем её ID
		// 		game := models.Game{
		// 			Name:   gameTitle,
		// 			Status: "waiting",
		// 		}

		// 		gameID, err := storage_db.CreateGame(game)
		// 		if err != nil {
		// 			log.Println("Ошибка создания игры:", err)
		// 			return tc.Send("❌ Виникла помилка при створенні гри. Спробуйте ще раз!")
		// 		}

		// 		// Теперь создаем GameRoom и передаем туда GameID
		// 		gameRoom.GameID = &gameID
		// 		_, inviteLink, err := storage_db.CreateGameRoom(gameRoom)
		// 		if err != nil {
		// 			log.Println("Ошибка создания игровой комнаты:", err)
		// 			return tc.Send("❌ Виникла помилка при створенні кімнати. Спробуйте ще раз!")
		// 		}

		// 		// Просим пользователя ввести вопросы
		// 		if err := tc.Send("❓ Введіть 4 питання та відповіді у форматі:\n`Питання 1 | Відповідь 1`"); err != nil {
		// 			return err
		// 		}

		// 		var tasks []models.Task

		// 		bot.Handle(telebot.OnText, func(tc telebot.Context) error {
		// 			text := tc.Text()
		// 			parts := strings.SplitN(text, "|", 2)

		// 			if len(parts) != 2 {
		// 				return tc.Send("⚠ Неправильний формат! Використовуйте `Питання | Відповідь`")
		// 			}

		// 			question := strings.TrimSpace(parts[0])
		// 			answer := strings.TrimSpace(parts[1])

		// 			tasks = append(tasks, models.Task{
		// 				GameID:   gameID,
		// 				Question: question,
		// 				Answer:   answer,
		// 			})

		// 			// Если получили 4 вопроса, записываем их в БД
		// 			if len(tasks) == 2 {
		// 				for _, task := range tasks {
		// 					storage_db.CreateTask(task)
		// 				}

		// 				successMsg := fmt.Sprintf(
		// 					"✅ Гру '%s' створено!\nЗапросіть своїх друзів за цим посиланням: %s",
		// 					gameTitle, inviteLink,
		// 				)
		// 				return tc.Send(successMsg)
		// 			}

		// 			// Просим следующий вопрос
		// 			return tc.Send(fmt.Sprintf("📝 Введіть питання %d:", len(tasks)+1))
		// 		})

		// 		return nil
		// 	})

		// 	return nil
		// })

		//return nil
	}
}

func askQuestions(bot *telebot.Bot, c telebot.Context, game *models.Game, count_questions int) error {
	count := 0
	maxQuestions := count_questions
	inviteLink := game.InviteLink

	// Ask the first question
	if err := c.Send("❓ Введіть два питання та відповіді у форматі:\n`Питання | Відповідь`"); err != nil {
		log.Println("Failed to send initial question prompt:", err)
		return err
	}

	time.Sleep(500 * time.Millisecond) // Wait for 2 seconds before sending the next message

	c.Send("📝 Введіть питання 1:")

	// Temporary slice to hold tasks in memory
	var tasks []models.Task

	bot.Handle(telebot.OnText, func(tc telebot.Context) error {
		text := tc.Text()
		parts := strings.SplitN(text, "|", 2)

		if len(parts) != 2 {
			return tc.Send("⚠ Неправильний формат! Використовуйте `Питання | Відповідь`")
		}

		question := strings.TrimSpace(parts[0])
		answer := strings.TrimSpace(parts[1])

		task := models.Task{
			GameID:  game.ID,
			Question: question,
			Answer:   answer,
		}

		if err := storage_db.CreateTask(task); err != nil {
			log.Println("Failed to save task to the database:", err)
			return tc.Send("❌ Не вдалося зберегти питання. Спробуйте ще раз.")
		}

		tasks = append(tasks, task)
		count++

		log.Printf("Task saved for GameID %d: '%s' -> '%s'", game.ID, question, answer)

		if count >= maxQuestions {
			successMsg := fmt.Sprintf(
				"✅ Гру '%s' створено!\nГра готова до запуску!",
				game.Name,
			)
			bot.Handle(telebot.OnText, nil) // Detach handler
			tc.Send(successMsg)

			time.Sleep(500 * time.Millisecond) // Wait for 2 seconds before sending the next message
			// Send the invite link
			inviteMsg := fmt.Sprintf("Запросіть своїх друзів за цим посиланням: %s", inviteLink)

			menuStartGame := &telebot.ReplyMarkup{}
			btnStartGame := menuStartGame.Data("Розпочати гру", "start_game", strconv.Itoa(game.ID))
			menuStartGame.Inline(menuStartGame.Row(btnStartGame))

			return tc.Send(inviteMsg, menuStartGame)
		}

		return tc.Send(fmt.Sprintf("📝 Введіть питання %d:", count+1))
	})

	return nil
}

func StartGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func (c telebot.Context) error {
		gameID, _ := strconv.Atoi(c.Callback().Data)
		

		user := c.Sender()
		log.Printf("User %s (%d) is starting game %d", user.Username, user.ID, gameID)

		// Check if the game exists
		game, err := storage_db.GetGameById(gameID)
		if err != nil {
		log.Printf("Failed to get game %d: %v", gameID, err)
			return c.Send("Ой, не можу знайти цю гру. Щось пішло не так...")
		}

		// Check if the game already started
		if game.Status == "playing" {
			return c.Send("Ця гра вже розпочата!")
		}

		// Check if there are at least 2 players
		playerCount, err := storage_db.GetPlayerCount(gameID)
		if err != nil {
			log.Printf("Failed to get player count: %v", err)
			return c.Send("Ой, щось пішло не так. Спробуй ще раз!")
		}

		if playerCount < 2 {
			return c.Send("Щоб розпочати гру, потрібно щонайменше 2 гравця. Запроси ще когось!")
		}

		if playerCount >= 2 {
			allTasks, err := storage_db.GetAllTasksByGameID(gameID)
			if err != nil {
				log.Printf("Failed to get tasks for game %d: %v", gameID, err)
				return c.Send("Ой, щось пішло не так. Спробуй ще раз!")
			}

			//countTasks := len(allTasks)
			allPlayers , err := storage_db.GetAllPlayersByGameID(gameID)
			if err != nil {
				log.Printf("Failed to get players for game %d: %v", gameID, err)
				return c.Send("Ой, щось пішло не так. Спробуй ще раз!")
			}

			// Send first task to all players
			for _, player := range allPlayers {
				task := allTasks[0]
				taskMsg := fmt.Sprintf("🎉 Гра почалася! Твоє перше завдання: \n%s", task.Question)
				_, err := bot.Send(&telebot.Chat{ID: player.ID}, taskMsg)
				if err != nil {
					log.Printf("Failed to send task to player %s: %v", player.UserName, err)
					return c.Send("Не вдалося надіслати завдання. Спробуй ще раз!")
				}

			}

		}


		return nil
	}
}


// Handler join to game
func JoinGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		return c.Send("Твої подруги вже створили кімнату для гри? 🏠 Тоді надішли мені посилання!")
	}
}

func HelpMeHandler(bot *telebot.Bot) func (c telebot.Context) error {
	return func(c telebot.Context) error {
		helpText := `
			Привіт, зіронько! 🌟 Я бот для ігор з подругами на відстані. Ось мої команди:

			/start - Почати бота і створити нову гру або доєднатися до існуючої
			/help - Показати це повідомлення

			В грі ти можеш:
			- Відповідати на завдання (текст, фото, відео, голосові повідомлення)
			- Пропустити завдання (максимум 3 рази)
			- Отримувати сповіщення про активність друзів

			Якщо потрібна допомога, натисни кнопку "Хелп мі" в меню!
		`
		return c.Send(helpText)
	}
}

func notifyPlayerJoined(bot *telebot.Bot, gameID int, player models.Player) {
	// Notify all players in the game that a new player has joined
	allPlayers, err := storage_db.GetAllPlayersByGameID(gameID)
	if err != nil {
		log.Printf("Failed to get players for game %d: %v", gameID, err)
		return
	}

	for _, p := range allPlayers {
		if p.ID != player.ID { // Don't notify the new player
			msg := fmt.Sprintf("🎉 Гравець %s приєднався до гри!", player.UserName)
			bot.Send(&telebot.Chat{ID: p.ID}, msg)
		}

	}
}