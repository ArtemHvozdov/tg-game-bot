package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"

	//"log"
	//"os"
	"strconv"

	"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/config"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"

	//"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	"gopkg.in/telebot.v3"
)

// type Task struct {
// 	ID 		int    `json:"id"`
// 	Tittle string `json:"title"`
// 	Description string `json:"description"`
// }

// func LoadTasks(path string) ([]Task, error) {
//     file, err := os.ReadFile(path)
//     if err != nil {
//         return nil, err
//     }

//     var tasks []Task
//     err = json.Unmarshal(file, &tasks)
//     if err != nil {
//         return nil, err
//     }

//     return tasks, nil
// }

var processedAlbums = make(map[string]time.Time) // processedAlbums keeps track of AlbumIDs that were already handled,
												 // to prevent sending multiple acknowledgments for a single album.

var cfg = config.LoadConfig()

func StartHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		user := c.Sender()
		
		utils.Logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"username": user.Username,
		}).Info("User started the bot")

		if chat.Type == telebot.ChatPrivate {
			utils.Logger.WithFields(logrus.Fields{
				"source": "StartHandler",
				"user_id": user.ID,
				"username": user.Username,
				"type_chat": chat.Type,
			}).Infof("User (%d | %s) clicked /start in private chat wit bot", user.ID, user.Username)

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

			bot.Handle(&btnHelp, HelpMeHandler(bot))

			return c.Send(startMsg, menu)
		}

		payload := c.Message().Payload
		if payload == "" {
			return c.Send("Щось пішло не так. 😔 Спробуй створити гру ще раз через особисте повідомлення боту.")
		}

		creatorID, err := strconv.ParseInt(payload, 10, 64)
		if err != nil {
		  utils.Logger.Errorf("Не вдалося розпізнати ID користувача: %v", err)
			return c.Send("Помилка при запуску гри. Спробуй ще раз.")
		}
    
		utils.Logger.WithFields(logrus.Fields{
			"source": "StartHandler",
			"group": chat.Title,
			"group_id": chat.ID,
			"admin_id:": creatorID,
			"admin": user.Username,
		}).Info("The bot was added to the group via a button in a private chat with the bot")
		
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
		utils.Logger.Info("CreateGameHandler buton's logs: User:", user.Username, user.ID)
    	
		if err := c.Send(gameStartMsg); err != nil {
			return err
		}

    return nil
	}
}

func HandleAddedToGroup(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		user := c.Sender()
		
		utils.Logger.WithFields(logrus.Fields{
			"source": "HandleAddedToGroup",
			"user:": user.Username,
			"user_id": user.ID,
			"group": chat.Title,
			"group_id": chat.ID,
		}).Info("The user added the bot to the group manually")
		
		// btnStartGame := telebot.Btn{Text: "Почати гру"}

		CheckAdminBotHandler(bot)(c)

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
		utils.Logger.Errorf("Failed to get players for game %d: %v", gameID, err)
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
func CheckAdminBotHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		user := c.Sender()

		utils.Logger.WithFields(logrus.Fields{
			"source": "CheckAdminBotHandler",
			"user_id": user.ID,
			"username": user.Username,
			"group": chat.Title,
			"group_id": chat.ID,
		}).Infof("Start creating game in the group (%d | %s)", chat.ID, chat.Title)
		
		gameName := chat.Title
		
		game, err := storage_db.CreateGame(gameName, chat.ID)
		if err != nil {
			utils.Logger.Errorf("Error creating game %s in the group %s: %v", gameName, chat.Title, err)
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
			utils.Logger.Errorf("Failed to add player-admin (%d | %s) to game %s: %v", playerAdmin.ID, playerAdmin.UserName, gameName, err)
			return c.Send("Ой, не вдалося додати тебе до гри. Спробуй ще раз!")
		}

		joinBtn := telebot.InlineButton{
			Unique: "join_game_btn",
			Text:   "🎲 Приєднатися до гри",
		}
		inline := &telebot.ReplyMarkup{}
		inline.InlineKeyboard = [][]telebot.InlineButton{
			{joinBtn},
		}

		//msgJoin, _ := bot.Send(chat, "Хочеш приєднатися до гри? 🏠 Тицяй кнопку", inline)
		bot.Send(chat, "Хочеш приєднатися до гри? 🏠 Тицяй кнопку", inline)
		//joinMsgId := msgJoin.ID
		//storage_db.UpdateMsgJoinID(game.ID, joinMsgId)
		
		// Delay pause between start game msg and join msg 
		time.Sleep(cfg.Durations.TimePauseMsgStartGameAndMsgJoinGame)

		// Version with Markup Button
		// menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		// btnStartGame := telebot.Btn{Text: "Почати гру"}
		// row1 := menu.Row(btnStartGame)
		// menu.Reply(row1)

		// Version with Inline Button
		menu := &telebot.ReplyMarkup{}
		btnStartGame := menu.Data("Почати гру", "start_game")
		menu.Inline(menu.Row(btnStartGame))

		bot.Handle(&btnStartGame, func(c telebot.Context) error {
			StartGameHandlerFoo(bot)(c)

			return nil
		})

		//time.Sleep(700 * time.Millisecond)	

		bot.Send(chat, "Тепер натисни кнопку нижче, коли будеш готовий почати гру! 🎮", menu)
				
		JoinBtnHandler(bot, joinBtn)

		return nil
	}
}

func JoinBtnHandler(bot *telebot.Bot, btn telebot.InlineButton) {
	bot.Handle(&btn, func(c telebot.Context) error {
			user := c.Sender()
			chat := c.Chat()

			utils.Logger.Info("Join btn handler was called. New funcion")
      
			utils.Logger.WithFields(logrus.Fields{
				"user_id": user.ID,
				"username": user.Username,
				"group": chat.Title,
				"group_id": chat.ID,
		  	}).Info("Inline button was called for joined to game")

			// Get game by chat ID
			game, err := storage_db.GetGameByChatId(chat.ID)
			if err != nil {
				utils.Logger.Errorf("Game not found for chat %d: %v", chat.ID, err)
				return c.Respond(&telebot.CallbackResponse{Text: "Гру не знайдено 😢"})
			}

			userIsInGame, err := storage_db.IsUserInGame(user.ID, game.ID)
			if err != nil {
				utils.Logger.Errorf("Failed to check if user %s is in game: %v", chat.Username, err)
				return nil
			}

			if userIsInGame {
				msg, err := bot.Send(chat, fmt.Sprintf("🎉 @%s, ти вже в грі! Не нервуйся", user.Username))
				if err != nil {
					utils.Logger.Errorf("Failed to send message for user %s: %v", user.Username, err)
					return nil
				}

				// Delay delete msg user is in game aready. Future: change time to 5 cseconds
				time.Sleep(cfg.Durations.TimeDeleteMsgUserIsAlreadyInGame)

				err = bot.Delete(msg)
				if err != nil {
					utils.Logger.Errorf("Failed to delete message for user %s: %v", user.Username, err)
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
				utils.Logger.WithFields(logrus.Fields{
					"source": "CheckAdminHandler",
					"player:": player.UserName,
					"player_id": player.ID,
					"game_id": game.ID,
				}).Error("Failed to add player to game")
				return c.Respond(&telebot.CallbackResponse{Text: "Не вдалося приєднатися 😢"})
			}

			statusGame := game.Status
			if statusGame == models.StatusGamePlaying {
				err := bot.Delete(c.Callback().Message)
				if err != nil {
					utils.Logger.Errorf("Failed to delete message with join button: %v", err)
					return nil
				}
			}

			joinedMessages, err := utils.LoadJoinMessagges("internal_data/hello_messages/hello_messages.json")
			if err != nil {
				utils.Logger.Errorf("Failed to load join messages: %v", err)
				return nil
			}

			//msg, err := bot.Send(chat, fmt.Sprintf("✨ @%s приєднався до гри!", user.Username))
			_, err = bot.Send(chat, fmt.Sprintf(joinedMessages[rand.Intn(len(joinedMessages))], user.Username))
			if err != nil {
				utils.Logger.Errorf("Failed to send join message for user %s: %v", user.Username, err)
				return nil
			}

			return c.Respond(&telebot.CallbackResponse{Text: "Ти в грі! 🎉"})
		})
}

func SendJoinGameReminder(bot *telebot.Bot) func (c telebot.Context) error {
	return func (c telebot.Context) error {
		joinBtn := telebot.InlineButton{
			Unique: "join_game_btn",
			Text:   "🎲 Приєднатися до гри",
		}
		inline := &telebot.ReplyMarkup{}
		inline.InlineKeyboard = [][]telebot.InlineButton{
			{joinBtn},
		}

		msgText := fmt.Sprintf(`🎉 @%s, ти ще не в грі! Натисни на кнопку щоб приєднатися і повертайся до завдання.`, c.Sender().Username)
		msgJoinGamerReminder, err := bot.Send(c.Chat(), msgText, inline)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "SendJoinGameReminder",
				"group": c.Chat().Title,
				"group_id": c.Chat().ID,
				"user_id": c.Sender().ID,
				"username": c.Sender().Username,
			}).Errorf("Failed to send join game reminder: %v", err)
		}

		JoinBtnHandler(bot, joinBtn)

		time.AfterFunc(cfg.Durations.TimeDeleteMsgJoinGamerReminder, func() {
			if msgJoinGamerReminder != nil {
				err := bot.Delete(msgJoinGamerReminder)
				if err != nil {
					if strings.Contains(err.Error(), "message to delete not found") {
						utils.Logger.WithFields(logrus.Fields{
							"source": "SendJoinGameReminder",
							"group": c.Chat().Title,
							"group_id": c.Chat().ID,
							"user_id": c.Sender().ID,
							"username": c.Sender().Username,
						}).Info("Message was already deleted earlier, skip deleting")
					} else {
						utils.Logger.WithFields(logrus.Fields{
							"source": "SendJoinGameReminder",
							"group": c.Chat().Title,
							"group_id": c.Chat().ID,
							"user_id": c.Sender().ID,
							"username": c.Sender().Username,
						}).Errorf("Failed to delete join game reminder message: %v", err)
					}
				}
			}
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
		utils.Logger.Errorf("Error generating invite link for the group %s: %v", chat.Title, err)
		return "", fmt.Errorf("failed to export chat invite link: %w", err)
	}

	// Struct of response Telegram API
	var result struct {
		Result string `json:"result"`
	}

	err = json.Unmarshal(raw, &result)
	if err != nil {
		utils.Logger.Errorf("Error parsing invite link for the group %s response: %v", chat.Title, err)
		return "", fmt.Errorf("failed to parse invite link response: %w", err)
	}

	return result.Result, nil
}


// StartGameHandlerFoo handles the "start_game" button press in a group chat
func StartGameHandlerFoo(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		user := c.Sender()

		utils.Logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"username": user.Username,
			"group": chat.Title,
		}).Infof("Start game handler called by %s", user.Username)

		memberUser, _ := bot.ChatMemberOf(chat, user)

		if memberUser.Role != telebot.Administrator && memberUser.Role != telebot.Creator {
			warningMsg := fmt.Sprintf("@%s, розпочати гру може тільки адмін групи. Трохи терпіння і почнемо.", user.Username)
			
			utils.Logger.WithFields(logrus.Fields{
				"user_id": user.ID,
				"username": user.Username,
				"group": chat.Title,
			}).Warn("Click to button /start_game, user is not admin in the group, tha can't start game")
			
			
			warningMsgSend, err := bot.Send(chat, warningMsg)
			if err != nil {
				utils.Logger.Errorf("Error sending warning message about start game in the chat: %v", err)
			}

			// Delay delete msg only admin can start game
			time.Sleep(cfg.Durations.TimeDeleteMsgOnlyAdmniCanStartGame)
			err = bot.Delete(warningMsgSend)
			if err != nil {
				utils.Logger.Errorf("Error deleting message warning message for user %s: %v", user.Username, err)
			}
			return nil
		}

		// Delete message with button "Start game" after click
		go func ()  {
			err := bot.Delete(c.Message())
			if err != nil {
			utils.Logger.Errorf("Failed to delete message with button: %v", err)
			}
		}()

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Error getting game by chat ID(%d): %v", chat.ID, err)
			return c.Send("❌ Не вдалося знайти гру для цього чату.")
		}

		utils.Logger.Infof("Game (%s) status: %s", game.Name, game.Status)

		// Checking: this have to be a group chat
		if chat.Type == telebot.ChatPrivate {
			c.Send("Ця кнопка працює лише у груповому чаті 🧑‍🤝‍🧑")
			return nil
		}

		if game.Status == models.StatusGamePlaying {
			msgText := fmt.Sprintf("@%s, ти вже розпочав гру!", user.Username)
			msg, err := bot.Send(chat, msgText)
			if err != nil {
				utils.Logger.Errorf(
					"Error sending message that game %s has already started for user (%d | %s): %v", game.Name, user.ID, user.Username, err,
				)
			}

			// Delay delete msg you already srarted game
			time.Sleep(cfg.Durations.TimeDeleteMsgYouAlreadyStartedGame)
			err = bot.Delete(msg)
			if err != nil {
				utils.Logger.Errorf(
					"Error deleting message that game %s has already started for user (%d | %s): %v", game.Name, user.ID, user.Username, err,
				)
			}

			return nil

		}

		startGameMsg := `ПРИВІТ, мене звати Фібі 😊, і наступні три тижні я буду вашим провідником у грі ✨ Грі, з якої вийдуть переможницями всі, якщо поділяться одна з одною своїм особливим скарбом – увагою. Від вас вимагається трошки часу і готове до досліджень серденько, від мене – цікава пригода, яку я загорнула у розроблені спеціально для вас спільні завдання.

Кожна дружба - неповторна, як булочка, повна родзинок 🍇 Ми будемо відщипувати шматочок за шматочком, виконуючи завдання. На кожне у вас буде 48 годин і незліченна кількість підтримки ваших бесті. Якщо якась родзинка вам не до смаку, ви можете пропустити це завдання. Але таких пропусків за всю гру кожній учасниці дозволяється лише 3.

Також є аварійна кнопка, щоб покинути цю гру раніше (але я вам точно не скажу, де вона, бо дуже хочу, щоб ви танцювали на цій вечірці до ранку). А якщо раптом щось пішло не так, ви можете дописатися до ді-джея, який ставить музику на тому боці (техпідтримка).

Вже зовсім скоро я надішлю вам перше завдання, де прийняття і чесність ми помножимо на спогади і гумор. А поки що тримайте в голові найважливіші правила гри – хев фан - і насолоджуйтеся часом, проведеним разом!`

		time.Sleep(600 * time.Millisecond) // Wait for 2 seconds before sending the next message
		removeKeyboard := &telebot.ReplyMarkup{RemoveKeyboard: true}
		_, err = bot.Send(chat, startGameMsg, removeKeyboard)
		if err != nil {
			utils.Logger.Errorf("Error sending welcome start game message go the chat %s: %v", chat.Title, err)
			
		}

		storage_db.UpdateGameStatus(int64(game.ID), models.StatusGamePlaying)

		// Delay pause before sending tasks
		time.Sleep(cfg.Durations.TimePauseBeforeStartSendingTask)

		// Start sending tasks
		return SendTasks(bot, chat.ID)(c)
		//return utils.SafeHandlerWithMessage(FinishGameHandler(bot))(c)

	}
}

func HandlerPlayerResponse(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()
		msg := c.Message()

		// ПCheck: if the message is part of an album and has already been processed, ignore it
		if msg.AlbumID != "" {
			if _, exists := processedAlbums[msg.AlbumID]; exists {
				return nil
			}

			// Register the album and set it to clear after 2 minutes
			processedAlbums[msg.AlbumID] = time.Now()

			// Delay delete album ID for group media msg
			time.AfterFunc(cfg.Durations.TimeDeleteAlbumId, func() {
				delete(processedAlbums, msg.AlbumID)
			})
		}

		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandleUHandlerPlayerResponseserJoined",
				"chat_id": chat.ID,
				"user_called": user.Username,
			}).Errorf("Error getting game by chat ID: %v", err)
			
			return nil
		}

		statusUser, err := storage_db.GetStatusPlayer(user.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandlerPlayerResponse",
				"user_id": user.ID,
				"username": user.Username,
				"group": chat.Title,
			}).Errorf("Error getting status player: %v", err)
		
			return nil
		}

		utils.Logger.WithFields(logrus.Fields{
				"source": "HandlerPlayerResponse",
				"username": user.Username,
				"group": chat.Title,
				"status_uer_from_DB": statusUser,
				"status_user_in_blocK_if": models.StatusPlayerWaiting+strconv.Itoa(game.CurrentTaskID),
			}).Info("Info about player and his status")
		
		userTaskID, _ := utils.GetWaitingTaskID(statusUser)

		playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
				GameID: 	game.ID,
				TaskID:		userTaskID,
				HasResponse: true,
				Skipped: false,
			}

			storage_db.AddPlayerResponse(playerResponse)

			bot.Send(chat, fmt.Sprintf("Дякую, @%s! Твоя відповідь на завдання %d прийнята.", user.Username, userTaskID))

			storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)

		return nil
	}
}

// SendFirstTasks send all tasks in group chat
func SendTasks(bot *telebot.Bot, chatID int64) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		game, err := storage_db.GetGameByChatId(chatID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"siource": "SendTasks",
			"chat_id": chatID,
			"chat_name": game.Name,
		}).Errorf("Error getting game by chat ID: %v", err)
	
		return err
	}

    tasks, err := utils.LoadTasks("internal_data/tasks/tasks.json")
    if err != nil {
		    utils.Logger.Errorf("SendTasks logs: Error loading tasks: %v", err)
        return err
    }

    if len(tasks) == 0 {
		utils.Logger.Error("SendTasks logs: No tasks to send. Tasks's array is empty" )
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
			// Delay pause between sending tasks
			time.Sleep(cfg.Durations.TimePauseBetweenSendingTasks) // await some minutes or hours before sending the next task
		}

    }

	return FinishGameHandler(bot)(c)

	}
	
}

func FinishGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "FinishGameHandler",
				"group": chat.Title,
				"err": err,
			}).Error("Finish game handler called")
			
			return nil
		}
		
		utils.Logger.WithFields(logrus.Fields{
			"source": "FinishGameHandler",
			"group": chat.Title,
		}).Info("Finish game handler called")

		// Final game. Future - function of final game will be here run
		finishMsg := `✨ Оʼкей, богині дружби, це офіційно — ВИ ПРОЙШЛИ ЦЕЙ ШЛЯХ РАЗОМ! ✨

Я хочу, щоб ви зараз на секунду відірвалися від екрану, зробили глибокий вдих і усвідомили: ВИ НЕЙОВІРНІ! Не тому, що виконали всі завдання (хоча це теж круто!), а тому, що ви створюєте простір, де можна бути собою. Де можна нити, мріяти, реготати, підтримувати, відкриватися і бути справжньою. Ви даєте одна одній свою увагу, час і ментальні обнімашки. 
	
І це точно найкращий момент, щоб подякувати всесвіту за ВАС! Серйозно, в світі 8 мільярдів людей, а ви зустріли своїх сестер по духу і змогли пронести цю дружбу крізь роки попри все! Це магія, це досягнення і це вдячність. Бережіть цю булочку з родзинками — вона унікальна.💛
	
Я сподіваюся, що цей досвід залишиться з вами не просто у вигляді чатику, а як тепле тріпотіння всередині: у мене є мої люди. І це — безцінно.
І, звісно, цей квест не має закінчення! Тому що дружба — це безперервна і прекрасна пригода.
	
Тепер питання: коли і де ви зустрічаєтеся, щоб відсвяткувати вашу перемогу, зіроньки? 🥂 😉`
		
		_, err = bot.Send(&telebot.Chat{ID: chat.ID}, finishMsg)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "FinishGameHandler",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending final message to the group")
		}

		storage_db.UpdateGameStatus(int64(game.ID), models.StatusGameFinished)

		return nil
	}
}

// Handler for answering a task
func OnAnswerTaskBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		user := c.Sender()
		chat := c.Chat()
		dataButton := c.Data()
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			utils.Logger.Errorf("Error getting game by chat ID (%d): %v", chat.ID, err)
			return nil
		}

		utils.Logger.WithFields(logrus.Fields{
			"source": "OnAnswerTaskBtnHandler",
			"username": user.Username,
			"group": chat.Title,
			"data_button": dataButton,
		}).Infof("User click to button WantAnswer to task %v", dataButton)

		userIsInGame, err := storage_db.IsUserInGame(user.ID, game.ID)
		if err != nil {
			utils.Logger.Errorf("Error checking if user is in game: %v", err)
			return nil
		}
		if !userIsInGame {
			SendJoinGameReminder(bot)(c)

			return nil
		}

		idTask, err := utils.GetWaitingTaskID(dataButton)
		if err != nil {
			utils.Logger.Errorf("Error getting task ID from data button: %v", err)
		}

		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, idTask)
		if err != nil {
			utils.Logger.Errorf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			textYouAlreadyAnswered := fmt.Sprintf("@%s, ти вже відповіла на це завдання 😅", user.Username)
			msgYouAlreadyAnswered, err := bot.Send(chat, textYouAlreadyAnswered)
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s already answered task %d: %v", user.Username, idTask, err)
			}

			time.AfterFunc(cfg.Durations.TimeDeleteMsgYouAlreadyAnswered, func() {
				err = bot.Delete(msgYouAlreadyAnswered)
				if err != nil {
					utils.Logger.WithFields(logrus.Fields{
						"source": "OnAnswerTaskBtnHandler",
						"username": user.Username,
						"group": chat.Title,
						"data_button": dataButton,
						"task_id": idTask,
					}).Errorf("Error deleting message that user %s already answered task %d: %v", user.Username, idTask, err)
				}
			})

			//return c.Send(fmt.Sprintf("@%s, ти вже відповідала на це завдання 😉", user.Username))
		case status.AlreadySkipped:
			return c.Send(fmt.Sprintf("@%s, це завдання ти вже пропустила 😅", user.Username))
		}

		storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(idTask))

		msg := fmt.Sprintf("@%s, чекаю від тебе відповідь на завдання %d", user.Username, idTask)
		awaitingAnswerMsg, err := bot.Send(chat, msg)
		if err != nil {
			utils.Logger.Errorf("Error sending message: %v", err)
		}

		// Delay delete msg awaiting answer
		time.AfterFunc(cfg.Durations.TimeDeleteMsgAwaitingAnswer, func() {
			err = bot.Delete(awaitingAnswerMsg)
			if err != nil {
				utils.Logger.WithFields(logrus.Fields{
					"source": "OnAnswerTaskBtnHandler",
					"username": user.Username,
					"group": chat.Title,
					"data_button": dataButton,
					"task_id": idTask,
				}).Errorf("Error deleting answer task message for user %s in the group %s: %v", chat.Username, chat.Title, err)
			}
		})

		return nil
	}
}

// Handler for skipping a task
func OnSkipTaskBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		utils.Logger.Info("OnSkipTaskHandler called")

		user := c.Sender()
		chat := c.Chat()
		dataButton := c.Data()
		game, _ := storage_db.GetGameByChatId(chat.ID)
		//statusUser, err := storage_db.GetStatusPlayer(user.ID)
		userTaskID, err := utils.GetSkipTaskID(dataButton)
		if err != nil {
			utils.Logger.Errorf("Error getting skip task ID from data button: %v", err)
			return nil
		}

		utils.Logger.WithFields(logrus.Fields{
			"source": "OnSkipTaskBtnHandler",
			"user": user.Username,
			"group": chat.Title,
			"data_button": dataButton,
			"skip_task_id": userTaskID,
		}).Infof("User click to button SkipTask from tasl %v", dataButton)

		userIsInGame, err := storage_db.IsUserInGame(user.ID, game.ID)
		if err != nil {
			utils.Logger.Errorf("Error checking if user is in game: %v", err)
			return nil
		}
		if !userIsInGame {
			SendJoinGameReminder(bot)(c)

			return nil
		}

		status, err := storage_db.SkipPlayerResponse(user.ID, game.ID, userTaskID)
		if err != nil {
			utils.Logger.Errorf("Error skipping task %d bu user: %v. %v", userTaskID, user.Username, err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			bot.Send(chat, fmt.Sprintf("📝 @%s, ти вже виконала це завдання.", user.Username))
		case status.AlreadySkipped:
			bot.Send(chat, fmt.Sprintf("⏭️ @%s, ти вже пропустила це завдання.", user.Username))
		case status.SkipLimitReached:
			msg, _ := bot.Send(chat, fmt.Sprintf("🚫 @%s, ти вже пропустила максимальну дозволену кількість завдань.", user.Username))
			
			// Delay delete the message max skip tasks
			time.AfterFunc(cfg.Durations.TimeDeleteMsgMaxSkipTasks, func() {
				err = bot.Delete(msg)
				if err != nil {
					utils.Logger.Errorf("Error deleting skip limit reached message for user %s: %v", user.Username, err)
				}
			})
		default:
			bot.Send(chat, fmt.Sprintf("✅ @%s, завдання пропущено! У тебе залишилось %d пропуск(ів).", user.Username, status.RemainingSkips-1))
		}

		return nil
	}
}