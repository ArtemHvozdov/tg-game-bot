package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	//"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	//"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	"gopkg.in/telebot.v3"
)

type Task struct {
	ID 		int    `json:"id"`
	Tittle string `json:"title"`
	Description string `json:"description"`
}

func LoadTasks(path string) ([]Task, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var tasks []Task
    err = json.Unmarshal(file, &tasks)
    if err != nil {
        return nil, err
    }

    return tasks, nil
}

func StartHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		if chat.Type == telebot.ChatPrivate {
			
			startMsg := "Оу, привіт, зіронько! 🌟 Хочеш створити гру для своїх найкращих подруг? Натискай кнопку нижче і вперед до пригод!"

			creatorID := fmt.Sprintf("%d", c.Sender().ID)
			deepLink := "https://t.me/bestie_game_bot?startgroup=" + creatorID

			menu := &telebot.ReplyMarkup{}
			btnDeepLink := menu.URL("➕ Створити гру", deepLink)
			btnHelp := menu.Data("❓ Help Me", "help_me")

			menu.Inline(
				menu.Row(btnDeepLink),
				menu.Row(btnHelp),
			)

			return c.Send(startMsg, menu)
		}

		payload := c.Message().Payload
		if payload == "" {
			return c.Send("Щось пішло не так. 😔 Спробуй створити гру ще раз через особисте повідомлення боту.")
		}

		creatorID, err := strconv.ParseInt(payload, 10, 64)
		if err != nil {
			log.Printf("❌ Не вдалося розпізнати ID користувача: %v", err)
			return c.Send("Помилка при запуску гри. Спробуй ще раз.")
		}

		log.Printf("Bot was join to group: %s (ID: %d), creatorID: %d", chat.Title, chat.ID, creatorID)

		return c.Send("🎉 Гру створено! Додайте своїх подруг і вперед до веселощів!")
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

		user := c.Sender()
		log.Println("CreateGameHandler butonns logs: User:", user.Username, user.ID)
    	
		if err := c.Send(gameStartMsg); err != nil {
			return err
		}

    return nil
	}
}

func HandleAddedToGroup(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		//chat := c.Chat()
		user := c.Sender()
		
		log.Printf("User: %d | %s", user.ID, user.Username)


		btnStartGame := telebot.Btn{Text: "Почати гру"}

		CheckAdminBotHandler(bot, btnStartGame)(c)

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

// CheckAdminBotHandler handles the /check_admin_bot command
func CheckAdminBotHandler(bot *telebot.Bot, btnStartGame telebot.Btn) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		log.Println("CheckAdminBotHandler called")

		chat := c.Chat()
		user := c.Sender()
		
		gameName := chat.Title
		
		game, err := storage_db.CreateGame(gameName, chat.ID)
		if err != nil {
			log.Printf("Error creating game: %v", err)
		}

		playerAdmin := &models.Player{
			ID: user.ID,
			UserName: user.Username,
			Name: user.FirstName,
			Status:   models.StatusPlayerNoWaiting,
			Skipped:  0,
			GameID: game.ID,
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

		
		bot.Send(chat, "Тепер натисни кнопку нижче, коли будеш готовий почати гру! 🎮", menu)

		time.Sleep(5 * time.Second)

		joinBtn := telebot.InlineButton{
			Unique: "join_game_btn",
			Text:   "🎲 Приєднатися до гри",
		}
		inline := &telebot.ReplyMarkup{}
		inline.InlineKeyboard = [][]telebot.InlineButton{
			{joinBtn},
		}

		bot.Send(chat, "Хочеш приєднатися до гри? 🏠 Тицяй кнопку", inline)		

		bot.Handle(&joinBtn, func(c telebot.Context) error {
			user := c.Sender()
			chat := c.Chat()

			log.Printf("Inline button was called for joined to game(DM): %s (%d) in chat %s (%d)\n", user.Username, user.ID, chat.Title, chat.ID)

			// Get game by chat ID
			game, err := storage_db.GetGameByChatId(chat.ID)
			if err != nil {
				log.Printf("Game not found for chat %d: %v", chat.ID, err)
				return c.Respond(&telebot.CallbackResponse{Text: "Гру не знайдено 😢"})
			}

			userIsInGame, err := storage_db.IsUserInGame(user.ID, game.ID)
			if err != nil {
				log.Printf("Failed to check if user is in game: %v", err)
				return nil
			}

			if userIsInGame {
				msg, err := bot.Send(chat, fmt.Sprintf("🎉 @%s, ти вже в грі! Не нервуйся", user.Username))
				if err != nil {
					log.Printf("Failed to send message: %v", err)
					return nil
				}

				time.Sleep(30 * time.Second)

				err = bot.Delete(msg)
				if err != nil {
					log.Printf("Failed to delete message: %v", err)
					return nil
				}
				return nil
			}

			player := &models.Player{
				ID:       user.ID,
				UserName: user.Username,
				Name:     user.FirstName,
				Status:   models.StatusPlayerNoWaiting,
				Skipped:  0,
				GameID:   game.ID,
				Role:     "player",
			}

			if err := storage_db.AddPlayerToGame(player); err != nil {
				log.Printf("Failed to add player: %v", err)
				return c.Respond(&telebot.CallbackResponse{Text: "Не вдалося приєднатися 😢"})
			}

			msg, err := bot.Send(chat, fmt.Sprintf("✨ @%s приєднався до гри!", user.Username))
			if err != nil {
				log.Printf("Failed to send join message: %v", err)
				return nil
			}

			// Delete message fate 1 minutes
			go func() {
				time.Sleep(60 * time.Second)
				bot.Delete(msg)
			}()

			return c.Respond(&telebot.CallbackResponse{Text: "Ти в грі! 🎉"})
		})

		return nil
	}
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

	// Struct of response Telegram API
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

		memberUser, _ := bot.ChatMemberOf(chat, user)

		if memberUser.Role != telebot.Administrator && memberUser.Role != telebot.Creator {
			warningMsg := fmt.Sprintf("@%s, розпочати гру може тільки адмін групи. Трохи терпіння і почнемо.", user.Username)
		
			warningMsgSend, err := bot.Send(chat, warningMsg)
			if err != nil {
				log.Println("Error sending warning message in the chat:", err)
			}

			time.Sleep(30 * time.Second)
			err = bot.Delete(warningMsgSend)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
			return nil
		}

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			log.Printf("Error getting game by chat ID: %v", err)
			return c.Send("❌ Не вдалося знайти гру для цього чату.")
		}

		log.Println("StartGameHandlerFoo logs: User:", user.Username, "Chat Name:", chat.Title, "Game status:", game.Status)

		// Checking: this have to be a group chat
		if chat.Type == telebot.ChatPrivate {
			c.Send("Ця кнопка працює лише у груповому чаті 🧑‍🤝‍🧑")
			return nil
		}

		if game.Status == models.StatusGamePlaying {
			msgText := fmt.Sprintf("@%s, ти вже розпочав гру!", user.Username)
			msg, err := bot.Send(chat, msgText)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}

			time.Sleep(1 * time.Minute)
			err = bot.Delete(msg)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}

			return nil

		}

		startGameMsg := `ПРИВІТ, мене звати Фібі 😊, і наступні три тижні я буду вашим провідником у грі ✨ Грі, з якої вийдуть переможницями всі, якщо поділяться одна з одною своїм особливим скарбом – увагою. Від вас вимагається трошки часу і готове до досліджень серденько, від мене – цікава пригода, яку я загорнула у розроблені спеціально для вас спільні завдання.

Кожна дружба - неповторна, як булочка, повна родзинок 🍇 Ми будемо відщипувати шматочок за шматочком, виконуючи завдання. На кожне у вас буде 48 годин і незліченна кількість підтримки ваших бесті. Якщо якась родзинка вам не до смаку, ви можете пропустити це завдання. Але таких пропусків за всю гру кожній учасниці дозволяється лише 3.

Також є аварійна кнопка, щоб покинути цю гру раніше (але я вам точно не скажу, де вона, бо дуже хочу, щоб ви танцювали на цій вечірці до ранку). А якщо раптом щось пішло не так, ви можете дописатися до ді-джея, який ставить музику на тому боці (техпідтримка).

Вже зовсім скоро я надішлю вам перше завдання, де прийняття і чесність ми помножимо на спогади і гумор. А поки що тримайте в голові найважливіші правила гри – хев фан - і насолоджуйтеся часом, проведеним разом!`

		time.Sleep(600 * time.Millisecond) // Wait for 2 seconds before sending the next message
		_, err = bot.Send(chat, startGameMsg)
		if err != nil {
			log.Printf("Error sending welcome start game message: %v", err)
		}

		storage_db.UpdateGameStatus(int64(game.ID), models.StatusGamePlaying)

		time.Sleep(1 * time.Hour)

		// Start sending tasks
		return SendTasks(bot, chat.ID)
	}
}

func HandlerPlayerResponse(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			log.Printf("Error getting game by chat ID: %v", err)
			return nil
		}

		statusUser, err := storage_db.GetStatusPlayer(user.ID)
		if err != nil {
			log.Printf("Error getiing status player: %v", err)
			return nil
		}
		
		log.Printf("HandlerPlayerResponse logs: User: %s, Chat Name: %s", user.Username, chat.Title)
		log.Print("HandlerPlayerResponse logs: User status: ", statusUser)
		log.Print("HandlerPlayerResponse logs: User status in block if: ", models.StatusPlayerWaiting+strconv.Itoa(game.CurrentTaskID))

		if statusUser == models.StatusPlayerWaiting+strconv.Itoa(game.CurrentTaskID) {
			playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
				GameID: 	game.ID,
				TaskID:		game.CurrentTaskID,
				HasResponse: true,
				Skipped: false,
			}

			storage_db.AddPlayerResponse(playerResponse)

			bot.Send(chat, fmt.Sprintf("Дякую, @%s! Твоя відповідь на завдання %d прийнята.", user.Username, game.CurrentTaskID))

			storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)
		}

		return nil
	}
}

// SendFirstTasks send all tasks in group chat
func SendTasks(bot *telebot.Bot, chatID int64) error {
	game, err := storage_db.GetGameByChatId(chatID)
	if err != nil {
		log.Printf("SendTasks logs: Error getting game by chat ID: %v", err)
		return err
	}

    tasks, err := LoadTasks("tasks/tasks.json")
    if err != nil {
		log.Printf("SendTasks logs: Error loading tasks: %v", err)
        return err
    }

    if len(tasks) == 0 {
		log.Println("SendTasks logs: No tasks to send.")
		return nil
	}

    for i, task := range tasks {
        //task := tasks[i]
		storage_db.UpdateCurrentTaskID(game.ID, task.ID)
        msg := "🌟 *" + task.Tittle + "*\n" + task.Description

		// create buttons Answer and Skip
		inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

		answerBtn := inlineKeys.Data("Хочу відповісти", "answer_task", fmt.Sprintf("waiting_%d", task.ID))
		skipBtn := inlineKeys.Data("Пропустити", "skip_task", fmt.Sprintf("skip_%d", task.ID))

		inlineKeys.Inline(
			inlineKeys.Row(answerBtn, skipBtn),
		)

        _, err := bot.Send(
            &telebot.Chat{ID: chatID},
            msg,
			inlineKeys,
            telebot.ModeMarkdown,
        )
        if err != nil {
            return err
        }

		if i < len(tasks)-1 {
			time.Sleep(3 * time.Minute) // await some minutes or hours before sending the next task
		}

    }

	// Final game. Future - function of final game will be here run
	finalMsg := `✨ Оʼкей, богині дружби, це офіційно — ВИ ПРОЙШЛИ ЦЕЙ ШЛЯХ РАЗОМ! ✨

Я хочу, щоб ви зараз на секунду відірвалися від екрану, зробили глибокий вдих і усвідомили: ВИ НЕЙОВІРНІ! Не тому, що виконали всі завдання (хоча це теж круто!), а тому, що ви створюєте простір, де можна бути собою. Де можна нити, мріяти, реготати, підтримувати, відкриватися і бути справжньою. Ви даєте одна одній свою увагу, час і ментальні обнімашки. 

І це точно найкращий момент, щоб подякувати всесвіту за ВАС! Серйозно, в світі 8 мільярдів людей, а ви зустріли своїх сестер по духу і змогли пронести цю дружбу крізь роки попри все! Це магія, це досягнення і це вдячність. Бережіть цю булочку з родзинками — вона унікальна.💛

Я сподіваюся, що цей досвід залишиться з вами не просто у вигляді чатику, а як тепле тріпотіння всередині: у мене є мої люди. І це — безцінно.
І, звісно, цей квест не має закінчення! Тому що дружба — це безперервна і прекрасна пригода.

Тепер питання: коли і де ви зустрічаєтеся, щоб відсвяткувати вашу перемогу, зіроньки? 🥂 😉`
	_, err = bot.Send(&telebot.Chat{ID: chatID}, finalMsg)
	if err != nil {
		log.Println("Error sending final message:", err)
	}

    return nil
}

// Handler for answering a task
func OnAnswerTaskBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		log.Println("OnAnswerTaskHandler called")

		user := c.Sender()
		chat := c.Chat()
		dataButton := c.Data()
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			log.Printf("Error getting game by chat ID: %v", err)
			return nil
		}

		idTask, err := utils.GetWaitingTaskID(dataButton)
		if err != nil {
			log.Printf("Error getting task ID from data button: %v", err)
		}

		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, idTask)
		if err != nil {
			log.Printf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			return c.Send(fmt.Sprintf("@%s, ти вже відповідала на це завдання 😉", user.Username))
		case status.AlreadySkipped:
			return c.Send(fmt.Sprintf("@%s, це завдання ти вже пропустила 😅", user.Username))
		}

		storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(idTask))

		return nil
	}
}

// Handler for skipping a task
func OnSkipTaskBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		log.Println("OnSkipTaskHadler called")

		user := c.Sender()
		chat := c.Chat()
		dataButton := c.Data()
		game, _ := storage_db.GetGameByChatId(chat.ID)

		log.Println("OnSkipTaskHandler logs: User:", user.Username, "Chat Name:", chat.Title, "Data Button:", dataButton, "Current Task ID:", game.CurrentTaskID)

		status, err := storage_db.SkipPlayerResponse(user.ID, game.ID, game.CurrentTaskID)
		if err != nil {
			log.Printf("Error skipping task: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			bot.Send(chat, fmt.Sprintf("📝 @%s, ти вже виконала це завдання.", user.Username))
		case status.AlreadySkipped:
			bot.Send(chat, fmt.Sprintf("⏭️ @%s, ти вже пропустила це завдання.", user.Username))
		case status.SkipLimitReached:
			bot.Send(chat, fmt.Sprintf("🚫 @%s, ти вже пропустила максимальну дозволену кількість завдань.", user.Username))
		default:
			bot.Send(chat, fmt.Sprintf("✅ @%s, завдання пропущено! У тебе залишилось %d пропуск(ів).", user.Username, status.RemainingSkips-1))
		}

		return nil
	}
}