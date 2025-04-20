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

// Handler for /start
func StartHandler(bot *telebot.Bot, btnCreateGame, btnHelpMe telebot.Btn) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		startMsg := "Оу, привіт, зіронько! 🌟 Хочеш створити гру для своїх найкращих подруг? Натискай кнопку нижче і вперед до пригод!"

		// Create keyboard
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

		// Buttons on the first line
		menuBtns := menu.Row(btnCreateGame, btnHelpMe)
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
				Status:   models.StatusPlayerNoWaiting,
				Skipped:  0,
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
		// Step 1: Ensure the command is used in a group chat
		if c.Chat().Type == telebot.ChatPrivate {
			return c.Send("Цю команду можна викликати тільки у груповому чаті ✋")
		}

		chat := c.Chat()
		user := c.Sender()
		
		// Step 2: Check if the user is an admin in the group
		memberUser, err := bot.ChatMemberOf(chat, user)
		if err != nil {
			log.Printf("Error fetching user's role in the group: %v", err)
			return nil
		}

		if memberUser.Role != telebot.Administrator && memberUser.Role != telebot.Creator {
			// Notify the group the user is not an admin
			warnMsg := fmt.Sprintf("@%s, цю команду може викликати тільки адмін групи 🚫", user.Username)
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

		// Try deleting the group messages after 1 minute
		go func() {
			time.Sleep(1 * time.Minute)
			// _ = bot.Delete(groupMsg)
			_ = bot.Delete(c.Message())
		}()

		gameName := chat.Title
		inviteChatLink, err := GenerateChatInviteLink(bot, chat)
		if err != nil {
			log.Printf("Error generating invite link: %v", err)
		}

		msgInviteLink, err := bot.Send(chat, fmt.Sprintf("Ухх, все в порядку! Гра створена, ось твоє магічне посилання: %s", inviteChatLink))
		if err != nil {
			log.Printf("Error sending invite link message: %v", err)
		}

		pinnedMsg := chat.PinnedMessage
		if pinnedMsg != nil {
			log.Printf("Deleting previous pinned message: %s", pinnedMsg.Text)
			err = bot.Delete(pinnedMsg)
			if err != nil {
				log.Printf("Error deleting previous pinned message: %v", err)
			}
		}

		// Pin message with invite link in the group chat
		err = bot.Pin(msgInviteLink)
		if err != nil {
			log.Printf("Error pinning message in group chat: %v", err)
		}

		game, err := storage_db.CreateGame(gameName, inviteChatLink, chat.ID)
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

		
		bot.Send(chat, "Тепер натисни кнопку нижче, коли будеш готовий почату гру! 🎮", menu)

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

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			log.Printf("Error getting game by chat ID: %v", err)
			return c.Send("❌ Не вдалося знайти гру для цього чату.")
		}

		memberUser, _ := bot.ChatMemberOf(chat, user)

		log.Println("StartGameHandlerFoo logs: User:", user.Username, "Chat Name:", chat.Title, "Game status:", game.Status)

		// Checking: this have to be a group chat
		if chat.Type == telebot.ChatPrivate {
			c.Send("Ця кнопка працює лише у груповому чаті 🧑‍🤝‍🧑")
			return nil
		}

		if chat.Type == telebot.ChatGroup && memberUser.Role != telebot.Administrator && memberUser.Role != telebot.Creator {
			warningMsg := fmt.Sprintf("@%s, не треба тицяти на кнопку, зараз тестуються нові фічі! 🚫", user.Username)
		
			_, err := bot.Send(chat, warningMsg)
			if err != nil {
				log.Println("Error sending warning message in the chat:", err)
			}
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

		time.Sleep(10 * time.Second)

		// Start sending tasks
		return SendTasks(bot, chat.ID)
	}
}

func HandleUserJoined(bot *telebot.Bot) telebot.HandlerFunc {
    return func(c telebot.Context) error {
        user := c.Sender()
        chat := c.Chat()

        log.Printf("User %s (%d) joined to chat %s (%d)\n",
            user.Username, user.ID, chat.Title, chat.ID)

		// Getting gamer by chat ID
        game, err := storage_db.GetGameByChatId(chat.ID)
        if err != nil {
            log.Printf("❌ Не удалось найти игру для чата %d: %v", chat.ID, err)
            return nil
        }

        log.Printf("Add user to game: %s (id: %d)", game.Name, game.ID)

        player := &models.Player{
            ID:       user.ID,
            UserName: user.Username,
            Name:     user.FirstName,
            Status:   models.StatusPlayerNoWaiting,
			Skipped:  0,
            GameID:   game.ID,
            Role:     "player",
        }

		// Add player to game
        if err := storage_db.AddPlayerToGame(player); err != nil {
            log.Printf("Failed to add player to game: %v", err)
            warningMsg := fmt.Sprintf("@%s, не вдалося додати тебе до гри. Спробуй ще раз!", user.Username)
            bot.Send(chat, warningMsg)
			return nil
        }

		bot.Send(chat, fmt.Sprintf("🎉Привіт @%s. Чекаємо ще подруг і скоро почнемо гру!", user.Username))

        return nil
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
		
		log.Printf("OnTextMsgHandler logs: User: %s, Chat Name: %s", user.Username, chat.Title)
		log.Print("OnTextMsgHandler logs: User status: ", statusUser)
		log.Print("OnTextMsgHandler logs: User status in block if:: ", models.StatusPlayerWaiting+strconv.Itoa(game.CurrentTaskID))

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

		answerBtn := inlineKeys.Data("Відповісти", "answer_task", fmt.Sprintf("waiting_%d", task.ID))
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
			time.Sleep(30 * time.Second) // await some minutes or hours before sending the next task
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
			return c.Send(fmt.Sprintf("@%s, ти вже відповідав на це завдання 😉", user.Username))
		case status.AlreadySkipped:
			return c.Send(fmt.Sprintf("@%s, це завдання ти вже пропустив 😅", user.Username))
		}

		storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(idTask))

		
		msg := fmt.Sprintf("@%s, чекаю від тебе відповідь на завдання %d", user.Username, idTask)
		_, err = bot.Send(chat, msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

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
