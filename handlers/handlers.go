package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"

	//"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	"gopkg.in/telebot.v3"
)

// type AwaiteState struct {
// 	StartStateAwait bool
// 	NameGameRoomAwait bool
// 	NameGameAwait bool
// 	QuestionsAwait bool
// }

// var botState = AwaiteState{}

// Handler for /start
func StartHandler(bot *telebot.Bot, btnCreateGame, btnHelpMe telebot.Btn) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		startMsg := "Оу, привіт, зіронько! 🌟 Хочеш створити гру для своїх найкращих подруг? Натискай кнопку нижче і вперед до пригод!"

		// Create keyboard
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		//.Reply(menu.Row(btnCreateGame, btnJoinGame, btnHelpMe))

		// Buttons on the first line
		menuBtns := menu.Row(btnCreateGame, btnHelpMe)
		// Button on the second row (all size)
		//row2 := menu.Row(btnHelpMe)

		menu.Reply(menuBtns)

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
				GameID:   int64(game.ID),
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
		//userAdmin := c.Sender()
		gameStartMsg := `Для початку тобі потрібно:
		1. Створити супергрупу
		2. Додати мене в цю групу з правами адміна
		3. У групі викликати команду /check_admin_bot`


		// Ask tha name game
		if err := c.Send(gameStartMsg); err != nil {
			return err
		}

		//var gameName string

		// bot.Handle(telebot.OnText, func(tc telebot.Context) error {
        //     gameName = tc.Text()

        //     game, err := storage_db.CreateGame(gameName, "")
        //     if err != nil {
        //         log.Println("Ошибка создания игры:", err)
        //         return tc.Send("❌ Виникла помилка при створенні гри. Спробуйте ще раз!")
        //     }

		// 	playerAdmin := &models.Player{
		// 		ID: userAdmin.ID,
		// 		UserName: userAdmin.Username,
		// 		Name: userAdmin.FirstName,
		// 		Passes: 0,
		// 		GameID: int64(game.ID),
		// 		Role: "admin",
		// 	}

		// 	if err := storage_db.AddPlayerToGame(playerAdmin); err != nil {
		// 		log.Printf("Failed to add player-admin to game: %v", err)
		// 		return c.Send("Ой, не вдалося додати тебе до гри. Спробуй ще раз!")
		// 	}

        //     // Move on to collecting questions
        //     return askQuestions(bot, tc, game, 2)
        // })

        return nil
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

// func StartGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
// 	return func (c telebot.Context) error {
// 		gameID, _ := strconv.Atoi(c.Callback().Data)
		

// 		user := c.Sender()
// 		log.Printf("User %s (%d) is starting game %d", user.Username, user.ID, gameID)

// 		// Check if the game exists
// 		game, err := storage_db.GetGameById(gameID)
// 		if err != nil {
// 		log.Printf("Failed to get game %d: %v", gameID, err)
// 			return c.Send("Ой, не можу знайти цю гру. Щось пішло не так...")
// 		}

// 		// Check if the game already started
// 		if game.Status == "playing" {
// 			return c.Send("Ця гра вже розпочата!")
// 		}

// 		// Check if there are at least 2 players
// 		playerCount, err := storage_db.GetPlayerCount(gameID)
// 		if err != nil {
// 			log.Printf("Failed to get player count: %v", err)
// 			return c.Send("Ой, щось пішло не так. Спробуй ще раз!")
// 		}

// 		if playerCount < 2 {
// 			return c.Send("Щоб розпочати гру, потрібно щонайменше 2 гравця. Запроси ще когось!")
// 		}

// 		if playerCount >= 2 {
// 			allTasks, err := storage_db.GetAllTasksByGameID(gameID)
// 			if err != nil {
// 				log.Printf("Failed to get tasks for game %d: %v", gameID, err)
// 				return c.Send("Ой, щось пішло не так. Спробуй ще раз!")
// 			}

// 			//countTasks := len(allTasks)
// 			allPlayers , err := storage_db.GetAllPlayersByGameID(gameID)
// 			if err != nil {
// 				log.Printf("Failed to get players for game %d: %v", gameID, err)
// 				return c.Send("Ой, щось пішло не так. Спробуй ще раз!")
// 			}

// 			// Send first task to all players
// 			for _, player := range allPlayers {
// 				task := allTasks[0]
// 				taskMsg := fmt.Sprintf("🎉 Гра почалася! Твоє перше завдання: \n%s", task.Question)
// 				_, err := bot.Send(&telebot.Chat{ID: player.ID}, taskMsg)
// 				if err != nil {
// 					log.Printf("Failed to send task to player %s: %v", player.UserName, err)
// 					return c.Send("Не вдалося надіслати завдання. Спробуй ще раз!")
// 				}

// 			}

// 		}


// 		return nil
// 	}
// }


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

// CheckAdminBotHandler handles the /check_admin_bot command
func CheckAdminBotHandler(bot *telebot.Bot, btnStartGame telebot.Btn) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		// Step 1: Ensure the command is used in a group chat
		if c.Chat().Type == telebot.ChatPrivate {
			return c.Send("Цю команду можна викликати тільки у груповому чаті ✋")
		}

		chat := c.Chat()
		user := c.Sender()
		//chatID := chat.ID
		//userID := user.ID
		username := user.Username

		// Step 2: Check if the user is an admin in the group
		memberUser, err := bot.ChatMemberOf(chat, user)
		if err != nil {
			log.Printf("Error fetching user's role in the group: %v", err)
			return nil
		}

		if memberUser.Role != telebot.Administrator && memberUser.Role != telebot.Creator {
			// Notify the group the user is not an admin
			warnMsg := fmt.Sprintf("@%s, цю команду може викликати тільки адмін групи 🚫", username)
			groupMsg, err := bot.Send(chat, warnMsg)
			if err != nil {
				log.Printf("Error sending non-admin warning: %v", err)
				return err
			}

			// Try deleting the messages after 30 seconds
			go func() {
				time.Sleep(30 * time.Second)
				_ = bot.Delete(groupMsg)
				err = bot.Delete(c.Message())
				if err != nil {
					log.Printf("Error deleting non-admin warning: %v", err)
				}
			}()

			return nil
		}

		// Step 3: Check if the bot itself is an admin
		memberBot, err := bot.ChatMemberOf(chat, &telebot.User{ID: bot.Me.ID})
		if err != nil {
			log.Printf("Error fetching bot's role in the group: %v", err)
			bot.Send(chat, "Я не можу перевірити свою роль у групі. Переконайся, що в мене є права адміна 🤖")
			return nil
		}

		if memberBot.Role != telebot.Administrator && memberBot.Role != telebot.Creator {
			notAdminMsg, err := bot.Send(chat, "Я не адміністратор у цій групі. Додай мене як адміна, будь ласка 🙏")
			if err != nil {
				log.Printf("Error sending bot admin warning: %v", err)
			}

			time.Sleep(30 * time.Second)
			err = bot.Delete(c.Message())
			if err != nil {
				log.Printf("Error deleting user message: %v", err)
			}
			_ = bot.Delete(notAdminMsg)

			return nil
		}

		// Step 4: All checks passed, notify in group and proceed in private
		groupSuccessMsg := fmt.Sprintf("@%s, я все перевірив ✅ Повернись до приватного чату зі мною, щоб продовжити створення гри. Чекаю тебе... 🌟", username)
		groupMsg, err := bot.Send(chat, groupSuccessMsg)
		if err != nil {
			log.Printf("Error sending success message to group: %v", err)
			return err
		}

		// Try deleting the group messages after 1 minute
		go func() {
			time.Sleep(1 * time.Minute)
			_ = bot.Delete(groupMsg)
			_ = bot.Delete(c.Message())
		}()

		// Continue interaction in private chat
		privateMsg := "Ухх, все в порядку! Групу створено і я маю права адміністратора 🛡️\nЙдемо далі..."
		_, err = bot.Send(user, privateMsg)
		if err != nil {
			log.Printf("Error sending private message to user: %v", err)
			return err
		}

		gameName := chat.Title
		inviteChatLink, err := GenerateChatInviteLink(bot, chat)
		if err != nil {
			log.Printf("Error generating invite link: %v", err)
		}

		bot.Send(user, fmt.Sprintf("Тепер я можу створити гру '%s' з інвайт-ссилкою: %s", gameName, inviteChatLink))

		// game, err := storage_db.CreateGame(gameName, inviteChatLink) // If we neen instance of game
		_, err = storage_db.CreateGame(gameName, inviteChatLink)
		if err != nil {
			log.Printf("Error creating game: %v", err)
		}

		playerAdmin := &models.Player{
			ID: user.ID,
			UserName: user.Username,
			Name: user.FirstName,
			Passes: 0,
			GameID: chat.ID,
			Role: "admin",
		}
		
		if err := storage_db.AddPlayerToGame(playerAdmin); err != nil {
			log.Printf("Failed to add player-admin to game: %v", err)
			return c.Send("Ой, не вдалося додати тебе до гри. Спробуй ще раз!")
		}

		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		row1 := menu.Row(btnStartGame)
		menu.Reply(row1)

		time.Sleep(700 * time.Millisecond)	

		
		bot.Send(chat, "Тепер натисни кнопку нижче, коли будеш готовий почату гру! 🎮", menu)

		return nil
	}
}

func CreateGame(groupChat *telebot.Chat, user *telebot.User) error {
	return nil
}

func GenerateChatInviteLink(bot *telebot.Bot, chat *telebot.Chat) (string, error) {
	params := map[string]interface{}{
		"chat_id": chat.ID,
	}

	raw, err := bot.Raw("exportChatInviteLink", params)
	if err != nil {
		log.Printf("Error generating invite link: %v", err)
		return "", fmt.Errorf("failed to export chat invite link: %w", err)
	}

	// Структура ответа Telegram API
	var result struct {
		Result string `json:"result"`
	}

	err = json.Unmarshal(raw, &result)
	if err != nil {
		log.Printf("Error parsing invite link response: %v", err)
		return "", fmt.Errorf("failed to parse invite link response: %w", err)
	}

	return result.Result, nil
}


// StartGameHandlerFoo handles the "start_game" button press in a group chat
func StartGameHandlerFoo(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		log.Println("StartGameHandlerFoo called")
		
		chat := c.Chat()
		user := c.Sender()

		log.Println("StartGameHandlerFoo logs: User:", user.Username, "Chat Name:", chat.Title)

		// Проверка: это должен быть групповой чат
		if chat.Type == telebot.ChatPrivate {
			return c.Send("Ця кнопка працює лише у груповому чаті 🧑‍🤝‍🧑")
		}

		// // Проверка: была ли нажата правильная кнопка
		// if c.Callback() == nil || c.Callback().Data != "start_game" {
		// 	return nil // просто игнорируем
		// }

		// // Удалим всплытие кнопки (если есть)
		// if err := bot.Respond(c.Callback(), &telebot.CallbackResponse{}); err != nil {
		// 	log.Printf("Error responding to callback: %v", err)
		// }

		// Формируем сообщение о старте
		startMsg := fmt.Sprintf("🎉 Гру розпочато! Учасник @%s натиснув кнопку старту. Готуйтесь до запитань!", user.Username)
		_, err := bot.Send(chat, startMsg)
		if err != nil {
			log.Printf("Error sending start message: %v", err)
			return err
		}

		// Здесь можно будет продолжить: запуск вопросов и т.д.
		return nil
	}
}
