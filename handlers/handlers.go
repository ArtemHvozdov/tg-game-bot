package handlers

import (
	"encoding/json"
	"fmt"

	//"log"
	//"image"

	// "image/color"
	// "image/draw"
	//"image/jpeg"
	//"io"
	//"math"
	//"path/filepath"

	//"os"
	"strconv"

	"strings"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/config"
	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/internal/subtasks"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/btnmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	//"github.com/fogleman/gg"

	//"github.com/fogleman/gg"
	"github.com/sirupsen/logrus"

	//"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	"gopkg.in/telebot.v3"
)

// Photo task session state
type PhotoTaskSession struct {
    UserID          int64
    CurrentQuestion int
    IsActive        bool
    QuestionMsgID   int
    AlbumMsgIDs     []int
    StartedAt       time.Time
}

// Photo questions
// var photoQuestions = []string{
//     "📸 Надішли фото місця, де ти найчастіше проводиш час з подругами",
//     "🌅 Покажи фото, яке передає настрій вашої дружби",
//     "💝 Надішли фото речі, яка нагадує тобі про найкращі моменти з подругами",
// }

var processedAlbums = make(map[string]time.Time) // processedAlbums keeps track of AlbumIDs that were already handled,
												 // to prevent sending multiple acknowledgments for a single album.

var cfg = config.LoadConfig()

var (
	//joinReminderMessages = make(map[int64]int)
	//joinReminderMutex = sync.Mutex{}

	menuIntro *telebot.ReplyMarkup
	menuExit  *telebot.ReplyMarkup

	introBtnHelp     telebot.Btn
	introBtnSupport  telebot.Btn
	introBtnExit     telebot.Btn
	btnExactlyExit   telebot.Btn
	btnReturnToGame  telebot.Btn

	joinedMessages []string
	finishMessage string
	buyMeCoffeeMsg string
	feedbackMsg string
	referalMsg string
	socialMediaLinks map[string]string
	wantAnswerMessages []string
	alreadyAnswerMessages []string
	staticMessages map[string]string
	skipMessages map[string]string
	startGameMessages []string
)

func InitLoaderMessages() {
	var err error
	joinedMessages, err = utils.LoadTextMessagges("internal/data/messages/group/hello_msgs/hello_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load join messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d join messages joinedMessages", len(joinedMessages))	
	}

	wantAnswerMessages, err = utils.LoadTextMessagges("internal/data/messages/group/want_answer_msgs/want_answer_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load want answer messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d want answer messages wantAnswerMessages", len(wantAnswerMessages))	
	}

	alreadyAnswerMessages, err = utils.LoadTextMessagges("internal/data/messages/group/already_answer_msgs/already_answer_msgs.json")	
	if err != nil {
		utils.Logger.Errorf("Failed to load already answer messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d already answer messages alreadyAnswerMessages", len(alreadyAnswerMessages))	
	}

	staticMessages, err = utils.LoadMessageMap("internal/data/messages/group/static_msgs/static_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load static messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d static messages", len(staticMessages))
	}

	skipMessages, err = utils.LoadMessageMap("internal/data/messages/group/skip_msgs/skip_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load skip messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d skip messages", len(skipMessages))
	}

	startGameMessages, err = utils.LoadTextMessagges("internal/data/messages/group/start_game_msgs/start_game_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load start game messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d start game messages", len(startGameMessages))
	}

	finishMessage, err = utils.LoadSingleMessage("internal/data/messages/group/finish_game_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load finish message: %v", err)
	} else {
		utils.Logger.Info("Loaded finish message: succes")
	}

	buyMeCoffeeMsg, err = utils.LoadSingleMessage("internal/data/messages/group/buy_me_coffee_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load buy me coffee message: %v", err)
	} else {
		utils.Logger.Info("Loaded buy me coffee message: succes")
	}

	feedbackMsg, err = utils.LoadSingleMessage("internal/data/messages/group/feedback_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load feedback message: %v", err)
	} else {
		utils.Logger.Info("Loaded feedback message: succes")
	}

	referalMsg, err = utils.LoadSingleMessage("internal/data/messages/group/referal_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load referal message: %v", err)
	} else {
		utils.Logger.Info("Loaded referal message: succes")
	}

	socialMediaLinks, err = utils.LoadMessageMap("internal/data/messages/group/social_media_links.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load social media links: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d social media links", len(socialMediaLinks))
	}
}

func InitButtons(gameID int) {
	menuIntro = &telebot.ReplyMarkup{}
	menuExit = &telebot.ReplyMarkup{}

	introBtnHelp = menuIntro.Data("Хелп", "help_menu")
	//introBtnSupport = menuIntro.URL("Техпідтримка", "https://t.me/Jay_jayss")
	introBtnSupport = btnmanager.Get(menuIntro, models.UniqueSupport)
	//introBtnExit = menuIntro.Data("Вийти з гри", fmt.Sprintf("exit_%d", gameID))
	introBtnExit = btnmanager.Get(menuIntro, models.UniqueExitGame, gameID)

	//btnExactlyExit = menuExit.Data("Точно вийти", fmt.Sprintf("exit_game_%d", gameID))
	btnExactlyExit = btnmanager.Get(menuExit, models.UniqueExactlyExit, gameID)
	//btnReturnToGame = menuExit.Data(" << Повернутися до гри", "return_to_game")
	btnReturnToGame = btnmanager.Get(menuExit, models.UniqueReturnToGame)

	menuIntro.Inline(menuIntro.Row(introBtnHelp))
	menuExit.Inline(menuExit.Row(btnExactlyExit), menuExit.Row(btnReturnToGame))
}

func StartHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		//chat := c.Chat()
		user := c.Sender()
		
		utils.Logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"username": user.Username,
		}).Info("User started the bot")

		startMsg := `ОУ, ПРИВІТ ЗІРОНЬКО! 🌟

Хочеш створити гру для своїх найкращих подруг? Лови інструкцію, як запустити магію✨:

➊ Створи групу з УСІМА подругами, з якими хочеш грати!
(Не забудь нікого! Пізніше додати вже не вийде 😬)

➋ Додай також і мене — @bestie_game_bot — я твоя ведуча, хе-хе 😎

➌ Можеш обрати фото і назву для групи! Це не must-have, але так фановіше 🤪

➍ Дочекайся, поки всі подружки натиснуть “Приєднатись до гри” 
(не тисни “Почати гру”, поки не зібрались усі ❗️)

➎ Коли УСІ приєднаються — тисни “Почати гру”! 🚀
Це можеш зробити тільки ти, бо ти тут — босс! 💅👑

І… let the madness begin! 💃🎉

ps  Маєше труднощі? Тоді пиши сюди`

		startMenu := &telebot.ReplyMarkup{}
		//startBtnSupport := startMenu.URL("🕹️ Техпідтримка", "https://t.me/Jay_jayss")

		startBtnSupport := btnmanager.Get(startMenu, models.UniqueSupport)

		startMenu.Inline(
			startMenu.Row(startBtnSupport),
		)

		return c.Send(startMsg, startMenu)
 	}
}

func TestRunHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		utils.Logger.Info("Test mode is running")

		SetupGameHandler(bot)(c)

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

		user := c.Sender()
		utils.Logger.Info("CreateGameHandler buton's logs: User:", user.Username, user.ID)
    	
		if err := c.Send(gameStartMsg); err != nil {
			return err
		}

    return nil
	}
}

func HandleBotAddedToGroup(bot *telebot.Bot) func(c telebot.Context) error {
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

		SetupGameHandler(bot)(c)

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

// SetupGameHandler handles the /check_admin_bot command
func SetupGameHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		user := c.Sender()
	
		utils.Logger.WithFields(logrus.Fields{
			"source": "SetupGameHandler",
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

		InitButtons(game.ID)

		// joinBtn := telebot.InlineButton{
		// 	Unique: "join_game_btn",
		// 	Text:   "🎲 Приєднатися до гри",
		// }

		//joinBtn := menuIntro.Data("🎲 Приєднатися до гри", "join_game_btn")
		joinBtn := btnmanager.Get(menuIntro, models.UniqueJoinGameBtn)
		
		//inline := &telebot.ReplyMarkup{}
		// inline.InlineKeyboard = [][]telebot.InlineButton{
		// 	{joinBtn},
		// 	{introBtnSupport},
		// 	{introBtnExit},
		// }

		menuIntro.Inline(
			menuIntro.Row(joinBtn),     // Join button
			menuIntro.Row(introBtnSupport),  // Support button  
			menuIntro.Row(introBtnExit),     // Exit button
		)

		//msgJoin, _ := bot.Send(chat, "Хочеш приєднатися до гри? 🏠 Тицяй кнопку", inline)
		//bot.Send(chat, "Хочеш приєднатися до гри? 🏠 Тицяй кнопку", inline)

		//bot.Send(chat, "Хочеш приєднатися до гри? 🏠 Тицяй кнопку", menuIntro)
		bot.Send(chat, utils.GetStaticMessage(staticMessages, models.MsgInviteToJoinGame), menuIntro)
				
		// Delay pause between start game msg and join msg 
		time.Sleep(cfg.Durations.TimePauseMsgStartGameAndMsgJoinGame)

		// Version with Inline Button
		menu := &telebot.ReplyMarkup{}
		//btnStartGame := menu.Data("Почати гру", "start_game")
		btnStartGame := btnmanager.Get(menu, models.UniqueStartGame)
		menu.Inline(menu.Row(btnStartGame))

		// bot.Handle(&btnStartGame, func(c telebot.Context) error {
		// 	StartGameHandlerFoo(bot)(c)

		// 	return nil
		// })

		bot.Send(chat, utils.GetStaticMessage(staticMessages, models.MsgAdminStartGameBtn), menu)

		return nil
	}
}

func JoinBtnHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return  func(c telebot.Context) error {
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
				roleUserInGame, err := storage_db.GetPlayerRoleByUserIDAndGameID(user.ID, game.ID)
				if err != nil {
					utils.Logger.Errorf("Failed to get player role for user %s in game %d: %v", user.Username, game.ID, err)
					return nil
				}

				switch roleUserInGame {
				case "admin":
					msg, err := bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgAdminWantToJoinGame), user.Username))
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
				case "player":
					msg, err := bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUsaulPlayerWantToJoinGame), user.Username))
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

				// msg, err := bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUsaulPlayerWantToJoinGame), user.Username))
				// if err != nil {
				// 	utils.Logger.Errorf("Failed to send message for user %s: %v", user.Username, err)
				// 	return nil
				// }

				// // Delay delete msg user is in game aready. Future: change time to 5 cseconds
				// time.Sleep(cfg.Durations.TimeDeleteMsgUserIsAlreadyInGame)

				// err = bot.Delete(msg)
				// if err != nil {
				// 	utils.Logger.Errorf("Failed to delete message for user %s: %v", user.Username, err)
				// 	return nil
				// }
				// return nil
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

			// statusGame := game.Status
			// if statusGame == models.StatusGamePlaying {
			// 	err := bot.Delete(c.Callback().Message)
			// 	if err != nil {
			// 		utils.Logger.Errorf("Failed to delete message with join button: %v", err)
			// 		return nil
			// 	}
			// }

			// joinedMessages1, err := utils.LoadTextMessagges("internal/data/messages/group/hello_messages/hello_messages.json")
			// if err != nil {
			// 	utils.Logger.Errorf("Failed to load join messages: %v", err)
			// 	return nil
			// }

			//msg, err := bot.Send(chat, fmt.Sprintf("✨ @%s приєднався до гри!", user.Username))
			_, err = bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(joinedMessages), user.Username))
			if err != nil {
				utils.Logger.Errorf("Failed to send join message for user %s: %v", user.Username, err)
				return nil
			}

			return c.Respond(&telebot.CallbackResponse{Text: "Ти в грі! 🎉"})
	}
}

func SendJoinGameReminder(bot *telebot.Bot) func (c telebot.Context) error {
	return func (c telebot.Context) error {
		//userID := c.Sender().ID
		chat := c.Chat()
		user := c.Sender()

		// joinReminderMutex.Lock()
		// defer joinReminderMutex.Unlock()

		// // Checking to see if there has already been a message
		// if msgID, ok := joinReminderMessages[user.ID]; ok && msgID != 0 {
		// 	// Trying to get this message (or just not sending it again)
		// 	// In Telebot, you can't get a message by ID, so we'll check by trying to delete it
		// 	err := bot.Delete(&telebot.Message{ID: msgID, Chat: c.Chat()})
		// 	if err == nil {
		// 		// There was an old message - it's still alive, let's exit
		// 		utils.Logger.Infof("Join reminder message already exists for user %d", user.ID)
		// 		return nil
		// 	} else if !strings.Contains(err.Error(), "message to delete not found") {
		// 		utils.Logger.Warnf("Failed to delete existing join reminder: %v", err)
		// 		// In any case, we remove it from the map
		// 		delete(joinReminderMessages, user.ID)
		// 	}
		// }

		// joinBtn := telebot.InlineButton{
		// 	Unique: "join_game_btn",
		// 	Text:   "🎲 Приєднатися до гри",
		// }
		// inline := &telebot.ReplyMarkup{}
		// inline.InlineKeyboard = [][]telebot.InlineButton{
		// 	{joinBtn},
		// }

		msgText := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserIsNotInGame), user.Username)
		_, err := msgmanager.SendTemporaryMessage(
			chat.ID, 
			user.ID, 
			msgmanager.TypeReminderJoinGame, 
			msgText, 
			cfg.Durations.TimeDeleteMsgJoinGamerReminder,
		)
		if err != nil {
			utils.Logger.Errorf("Error sending reminder joint to game msg for user %s in the chat (%s | %d)", user.Username, chat.Title, chat.ID)
		}

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

// SendStartGameMessages sends the start game messages to the game chat
func SendStartGameMessages(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		msg1 := startGameMessages[0]
		msg2 := startGameMessages[1]

		if _, err := bot.Send(chat, msg1, telebot.ModeMarkdown); err != nil {
			return err
		}

		if _, err := bot.Send(chat, msg2, telebot.ModeMarkdown); err != nil {
			return err
		}

		return nil
	}
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
			warningMsgText := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgOnlyAdminCanStartGame), user.Username)
			
			utils.Logger.WithFields(logrus.Fields{
				"user_id": user.ID,
				"username": user.Username,
				"group": chat.Title,
			}).Warn("Click to button /start_game, user is not admin in the group, tha can't start game")

			_, err := msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				msgmanager.TypeGameStart,
				warningMsgText,
				cfg.Durations.TimeDeleteMsgOnlyAdmniCanStartGame,
			)
			if err != nil {
				utils.Logger.Errorf("Error sending warning message about start game in the chat: %v", err)
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
			msgText := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgAdminGameAlreadyStarted), user.Username)
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

		//InitButtons(game.ID)

// 		msgTextStartGame := `ПРИВІТ, мене звати Фібі 😊, і наступні три тижні я буду вашим провідником у грі ✨ Грі, з якої вийдуть переможницями всі, якщо поділяться одна з одною своїм особливим скарбом – увагою. Від вас вимагається трошки часу і готове до досліджень серденько, від мене – цікава пригода, яку я загорнула у розроблені спеціально для вас спільні завдання.

// Кожна дружба - неповторна, як булочка, повна родзинок 🍇 Ми будемо відщипувати шматочок за шматочком, виконуючи завдання. На кожне у вас буде 48 годин і незліченна кількість підтримки ваших бесті. Якщо якась родзинка вам не до смаку, ви можете пропустити це завдання. Але таких пропусків за всю гру кожній учасниці дозволяється лише 3.

// Також є аварійна кнопка, щоб покинути цю гру раніше (але я вам точно не скажу, де вона, бо дуже хочу, щоб ви танцювали на цій вечірці до ранку). А якщо раптом щось пішло не так, ви можете дописатися до ді-джея, який ставить музику на тому боці (техпідтримка).

// Вже зовсім скоро я надішлю вам перше завдання, де прийняття і чесність ми помножимо на спогади і гумор. А поки що тримайте в голові найважливіші правила гри – хев фан - і насолоджуйтеся часом, проведеним разом!`

		
		// menuIntro := &telebot.ReplyMarkup{}
		// menuExit := &telebot.ReplyMarkup{}

		// introBtnHelp := menuIntro.Data("🕹️ Хелп", "help_menu")
		// introBtnSupport := menuIntro.URL("🕹️ Техпідтримка", "https://t.me/Jay_jayss")
		// introBtnExit := menuIntro.Data("🕹️ Вийти з гри", fmt.Sprintf("exit_%d", game.ID))

		// btnExactlyExit := menuIntro.Data("Вийти з гри", fmt.Sprintf("exit_game_%d", game.ID))
		// btnReturnToGame := menuIntro.Data(" << Повернутися до гри", "return_to_game")

		time.Sleep(600 * time.Millisecond) // Wait for 2 seconds before sending the next message
		//removeKeyboard := &telebot.ReplyMarkup{RemoveKeyboard: true}
		// menuIntro.Inline(
		// 	menuIntro.Row(introBtnHelp),
		// )

		SendStartGameMessages(bot)(c)
		 
		// _, err = bot.Send(chat, msgTextStartGame)
		// if err != nil {
		// 	utils.Logger.Errorf("Error sending welcome start game message go the chat %s: %v", chat.Title, err)
			
		// }

		storage_db.UpdateGameStatus(int64(game.ID), models.StatusGamePlaying)

		// Delay pause before sending tasks
		time.Sleep(cfg.Durations.TimePauseBeforeStartSendingTask)

		// Start sending tasks
		return SendTasks(bot, chat.ID)(c)
		//return utils.SafeHandlerWithMessage(FinishGameHandler(bot))(c)

	}
}

func handleHelpMenu(bot *telebot.Bot, c telebot.Context) error {
	user := c.Sender()
	chat := c.Chat()
	utils.Logger.WithFields(logrus.Fields{
		"source": "handleHelpMenu",
		"user_id": user.ID,
		"username": user.Username,
		"group": chat.Title,
	}).Info("Help menu button was pressed")

	menuIntro.Inline(
		menuIntro.Row(introBtnSupport),
		menuIntro.Row(introBtnExit),
	)
	// bot.EditReplyMarkup(c.Callback().Message, menuIntro)

	// time.Sleep(5 * time.Second) // Delay to allow user to read the message

	// 	menuIntro.Inline(
	// 		menuIntro.Row(introBtnHelp),
	// 	)

	// 	_, err := bot.EditReplyMarkup(msgStartGame, menuIntro)
	// 	if err != nil {
	// 		utils.Logger.WithFields(logrus.Fields{
	// 			"source": "StartGameHhandleHelpMenuandlerFoo",
	// 			"group": chat.Title,
	// 			"group_id": chat.ID,
	// 			"user_id": user.ID,
	// 			"username": user.Username,
	// 		}).Errorf("Failed to edit reply markup after exit game: %v", err)
	// 	}

	return nil
}

func handleExitConfirm(bot *telebot.Bot, c telebot.Context) error {
	utils.Logger.WithFields(logrus.Fields{
		"source": "handleExitConfirm",
		"user_id": c.Sender().ID,
		"username": c.Sender().Username,
		"group": c.Chat().Title,
	}).Info("Exit confirm button was pressed")

	user := c.Sender()
	chat := c.Chat()
	data := c.Callback().Data

	if strings.HasPrefix(data, "\fexit_") {
		gameIDStr := strings.TrimPrefix(data, "\fexit_")
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			return nil
		}

		isUserInGame , err := storage_db.IsUserInGame(user.ID, gameID)
		if err != nil {
			utils.Logger.Errorf("Error checking if user %s is in game: %v", user.Username, err)
			return nil
		}

		if !isUserInGame {
			msgTextUserIsNotInGame := fmt.Sprintf("Ти не в грі, @%s. Тому не можеш вийти з неї 🤷‍♂️", user.Username)

			_, err := msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				msgmanager.TypeNotInGame, // unique message type
				msgTextUserIsNotInGame,
				cfg.Durations.TimeDeleteMsgYouAreNotInGame,
			)
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
			}


			return nil
		}

		roleUserInGame, err := storage_db.GetPlayerRoleByUserIDAndGameID(user.ID, gameID)
		if err != nil {
			utils.Logger.Errorf("Error getting player role for user %s in game %d: %v", user.Username, gameID, err)
			return nil
		}

		if roleUserInGame == "admin" {
			msgTextAdminExit := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgAdminExitGame), user.Username)
			
			// Using MessageManager to Send with Anti-Duplicate Protection
			_, err := msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				"admin_exit", // unique message type
				msgTextAdminExit,
				cfg.Durations.TimeDeleteMsgAdminExit,
			)

		
			if err != nil {
				utils.Logger.Errorf("Error sending message that admin %s cannot exit game: %v", user.Username, err)
			}

			return nil
		}

		msgTextExtit := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgPlayerExitGame), user.Username)

						
		menuExit.Inline(
			menuExit.Row(btnExactlyExit),
			menuExit.Row(btnReturnToGame),
		)

		_, err = msgmanager.SendTemporaryMessage(
			chat.ID,
			user.ID,
			msgmanager.TypeExitSuccess, // unique message type
			msgTextExtit,
			cfg.Durations.TimeDeleteMsgAdminExit,
			menuExit,
		)
		if err != nil {
			utils.Logger.Errorf("Error sending exit game message to the chat %s: %v", chat.Title, err)
		}
	}
			
	return nil
}

func handleExitGame(bot *telebot.Bot, c telebot.Context) error {
	user := c.Sender()
	chat := c.Chat()
	data := c.Callback().Data

	if strings.HasPrefix(data, "\fexit_game_") {
		gameIDStr := strings.TrimPrefix(data, "\fexit_game_")
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			return nil
		}

		isUserInGame , err := storage_db.IsUserInGame(user.ID, gameID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "handleExitGame",
				"group": chat.Title,
				"group_id": chat.ID,
				"user_id": user.ID,
				"username": user.Username,
			}).Errorf("Error checking if user %s is in game: %v", user.Username, err)
		}
		
		if !isUserInGame {
			msgTextUserIsNotInGame := fmt.Sprintf("@%s ти ж вже вийшла з гри 🤷‍♂️", chat.Username)
			msgUserIsNotInGame, err := bot.Send(chat, msgTextUserIsNotInGame)
			if err != nil {
				utils.Logger.WithFields(logrus.Fields{
					"source": "handleExitGame",
					"group": chat.Title,
					"group_id": chat.ID,
					"user_id": user.ID,
					"username": user.Username,
				}).Errorf("Error sending message that user %s is not in game: %v", chat.Username, err)
			}

			time.AfterFunc(cfg.Durations.TimeDeleteMsgYouAreNotInGame, func() {
				err := bot.Delete(msgUserIsNotInGame)
				if err != nil {
					utils.Logger.WithFields(logrus.Fields{
						"source": "handleExitGame",
						"group": chat.Title,
						"group_id": chat.ID,
						"user_id": user.ID,
						"username": user.Username,
					}).Errorf("Failed to delete message that user is not in game: %v", err)
				}
			})
		}

		storage_db.DeletePlayerFromGame(user.ID, gameID)

		err = bot.Delete(c.Callback().Message)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
					"source": "handleExitGame",
					"group": chat.Title,
					"group_id": chat.ID,
					"user_id": user.ID,
					"username": user.Username,		
				}).Errorf("Failed to delete message with exit confirmation: %v", err)
			return nil
		}

		msgTextExit := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgExactlyExitGame), user.Username)
		_, err = bot.Send(chat, msgTextExit)
		if err != nil {
			utils.Logger.Errorf("Error sending exit message to the chat %s: %v", chat.Title, err)
		}

	}
			
	return nil
}

func handleReturnToGame(bot *telebot.Bot, c telebot.Context) error {
	user := c.Sender()
	chat := c.Chat()

	err := bot.Delete(c.Callback().Message)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "handleReturnToGame",
			"group": chat.Title,
			"group_id": chat.ID,
			"user_id": user.ID,
			"username": user.Username,
		}).Errorf("Failed to delete return to game message: %v", err)
		return nil
	}

	msgTextReturnToGame := fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgReturnToGame), user.Username)
	msgReturnToGame, err := bot.Send(chat, msgTextReturnToGame)
	if err != nil {
		utils.Logger.Errorf("Error sending return to game message to the chat %s: %v", chat.Title, err)
	}

	time.AfterFunc(cfg.Durations.TimeDeleteMsgReturnToGame, func() {
		err := bot.Delete(msgReturnToGame)
		if err != nil {
			if strings.Contains(err.Error(), "message to delete not found") {
				utils.Logger.WithFields(logrus.Fields{
					"source": "handleReturnToGame",
					"group": chat.Title,
					"group_id": chat.ID,
					"user_id": user.ID,
					"username": user.Username,
				}).Info("Message was already deleted earlier, skip deleting")
			} else {
				utils.Logger.WithFields(logrus.Fields{		
					"source": "handleReturnToGame",
					"group": chat.Title,
					"group_id": chat.ID,
					"user_id": user.ID,
					"username": user.Username,
				}).Errorf("Failed to delete return to game message: %v", err)
			}
		}
	})
	// menuIntro.Inline(
	// 	menuIntro.Row(introBtnHelp),
	// )

	// _, err = bot.EditReplyMarkup(msgStartGame, menuIntro)
	// if err != nil {
	// 	utils.Logger.WithFields(logrus.Fields{
	// 		"source": "handleReturnToGame",
	// 		"group": chat.Title,
	// 		"group_id": chat.ID,
	// 		"user_id": user.ID,
	// 		"username": user.Username,
	// 	}).Errorf("Failed to edit reply markup after return to game: %v", err)
	// 	return nil
	// }
			
	return nil
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

		if statusUser == models.StatusPlayerNoWaiting {
			utils.Logger.WithFields(logrus.Fields{
				"source": "HandlerPlayerResponse",
				"user_id": user.ID,
				"username": user.Username,		
				"group": chat.Title,
			}).Warnf("User %s is not waiting for task %d, current status: %s", user.Username, game.CurrentTaskID, statusUser)
			
			// Skip message of user he already answered
			return nil
		}
		
		userTaskID, _ := utils.GetWaitingTaskID(statusUser)

		// Skip messges from user. User answered subtask
		if userTaskID == 3 {
			return nil
		}

		playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
				GameID: 	game.ID,
				TaskID:		userTaskID,
				HasResponse: true,
				Skipped: false,
			}

			storage_db.AddPlayerResponse(playerResponse)

			bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAnswerAccepted), user.Username, userTaskID))

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

    tasks, err := utils.LoadTasks("internal/data/tasks/tasks.json")
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
        // msg := "🌟 *" + task.Tittle + "*\n" + task.Description

		msg := task.Tittle + "\n\n" + task.Description
		
		// create buttons Answer and Skip
		inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

		//answerBtn := inlineKeys.Data("Хочу відповісти", "answer_task", fmt.Sprintf("waiting_%d", task.ID))
		//skipBtn := inlineKeys.Data("Пропустити", "skip_task", fmt.Sprintf("skip_%d", task.ID))
		answerBtn := btnmanager.Get(inlineKeys, models.UniqueAnswerTask, task.ID)
		skipBtn := btnmanager.Get(inlineKeys, models.UniqueSkipTask, task.ID)

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
			if i == 2 {
				time.Sleep(5 * time.Minute) // Wait for 5 seconds before sending the next task
			}
			// Delay pause between sending tasks
			time.Sleep(cfg.Durations.TimePauseBetweenSendingTasks) // await some minutes or hours before sending the next task
		}

    }

	return FinishGameHandler(bot)(c)

	}
	
}

// func GetReferalLinkHandler(bot *telebot.Bot) func(c telebot.Context) error {
// 	return func(c telebot.Context) error {
// 		chat := c.Chat()
// 		game , err := storage_db.GetGameByChatId(chat.ID)
// 		if err != nil {
// 			utils.Logger.WithFields(logrus.Fields{
// 				"source": "GetReferalLinkHandler",
// 				"group": chat.Title,
// 				"group_id": chat.ID,	
// 				"user_id": c.Sender().ID,
// 				"username": c.Sender().Username,
// 			}).Errorf("Error getting game by chat ID: %v", err)
// 			return nil
// 		}

// 		userAdminGame, err := storage_db.GetAdminPlayerByGameID(game.ID)
// 		if err != nil {
// 			utils.Logger.WithFields(logrus.Fields{
// 				"source": "GetReferalLinkHandler",
// 				"group": chat.Title,
// 				"group_id": chat.ID,
// 				"user_id": c.Sender().ID,
// 				"username": c.Sender().Username,
// 			}).Errorf("Error getting admin player by game ID: %v", err)
// 		}

// 		refLink1 := utils.GenerateInviteLink(int(userAdminGame.ID))

// 		bot.Send(&telebot.Chat{ID: chat.ID}, 
// 		fmt.Sprintf("Твоє реферальне посилання: <a href=\"%s\">%s</a>", refLink1, refLink1),
// 		&telebot.SendOptions{
// 			ParseMode: telebot.ModeHTML,
// 			DisableWebPagePreview: true,
// 		},
// )

// 		return nil
// 	}
// }

func SendReferalMsg(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		userCalled := c.Sender()

		var refLink string

		// Пробуем получить игру
		game, err := storage_db.GetGameByChatId(chat.ID)
		if err != nil {
			// Логируем ошибку, но НЕ выходим из функции
			utils.Logger.WithFields(logrus.Fields{
				"source":   "SendReferalMsg",
				"group":    chat.Title,
				"group_id": chat.ID,
				"user_id":  userCalled.ID,
				"username": userCalled.Username,
			}).Warnf("Game not found or error occurred: %v", err)

			// Используем ссылку на вызывающего пользователя
			refLink = utils.GenerateInviteLink(int(userCalled.ID))
		} else {
			// Игра найдена, пробуем получить админа
			adminPlayer, err := storage_db.GetAdminPlayerByGameID(game.ID)
			if err != nil {
				utils.Logger.WithFields(logrus.Fields{
					"source":   "SendReferalMsg",
					"group":    chat.Title,
					"group_id": chat.ID,
					"user_id":  userCalled.ID,
					"username": userCalled.Username,
				}).Warnf("Admin not found, fallback to sender: %v", err)

				refLink = utils.GenerateInviteLink(int(userCalled.ID))
			} else {
				refLink = utils.GenerateInviteLink(int(adminPlayer.ID))
			}
		}

		// Подготовка HTML-сообщения
		msg := referalMsg
		msg = strings.ReplaceAll(msg, "Instagram", fmt.Sprintf(`<a href="%s">Instagram</a>`, utils.GetStaticMessage(socialMediaLinks, models.LinkInstagram)))
		msg = strings.ReplaceAll(msg, "TikTok", fmt.Sprintf(`<a href="%s">TikTok</a>`, utils.GetStaticMessage(socialMediaLinks, models.LinkTikTok)))
		msg = strings.ReplaceAll(
			msg,
			"Ось твоє Космічне посилання, за яким подружки і подружки подружок зможуть зіграти у власну гру BESTIEVERSE",
			fmt.Sprintf(`<a href="%s">Ось твоє Космічне посилання, за яким подружки і подружки подружок зможуть зіграти у власну гру BESTIEVERSE</a>`, refLink),
		)

		// Отправка
		_, err = bot.Send(chat, msg, &telebot.SendOptions{
			ParseMode:             telebot.ModeHTML,
			DisableWebPagePreview: true,
		})
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source":   "SendReferalMsg",
				"group":    chat.Title,
				"group_id": chat.ID,
				"err":      err,
			}).Error("Error sending referral message to the group")
			return err
		}

		utils.Logger.WithFields(logrus.Fields{
			"group": chat.Title,
			"user":  userCalled.Username,
			"link":  refLink,
		}).Info("Referral message sent successfully")

		return nil
	}
}

// SendFeedbackMsg sends a feedback message to the users
func SendFeedbackMsg(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		feedbackMenu := &telebot.ReplyMarkup{}

		feedbackBtn := btnmanager.Get(feedbackMenu, models.UniqueFeedback)

		feedbackMenu.Inline(
			feedbackMenu.Row(feedbackBtn),
		)

		// startMenu.Inline(
		// 	startMenu.Row(startBtnSupport),
		// )
		_, err := bot.Send(chat, feedbackMsg, feedbackMenu)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "SendFeedbackMsg",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending feedback message to the group")
			return err
		}

		utils.Logger.Info("Feedback message sent successfully")

		return nil
	}
}

func SendBuyMeCoffeeMsg(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()

		_, err := bot.Send(chat, buyMeCoffeeMsg, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "SendBuyMeCoffeeMsg",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending Buy Me Coffee message to the group")
			return err
		}

		utils.Logger.Info("Buy Me Coffee message sent successfully")

		return nil
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
// 		finishMsg := `✨ Оʼкей, богині дружби, це офіційно — ВИ ПРОЙШЛИ ЦЕЙ ШЛЯХ РАЗОМ! ✨

// Я хочу, щоб ви зараз на секунду відірвалися від екрану, зробили глибокий вдих і усвідомили: ВИ НЕЙОВІРНІ! Не тому, що виконали всі завдання (хоча це теж круто!), а тому, що ви створюєте простір, де можна бути собою. Де можна нити, мріяти, реготати, підтримувати, відкриватися і бути справжньою. Ви даєте одна одній свою увагу, час і ментальні обнімашки. 
	
// І це точно найкращий момент, щоб подякувати всесвіту за ВАС! Серйозно, в світі 8 мільярдів людей, а ви зустріли своїх сестер по духу і змогли пронести цю дружбу крізь роки попри все! Це магія, це досягнення і це вдячність. Бережіть цю булочку з родзинками — вона унікальна.💛
	
// Я сподіваюся, що цей досвід залишиться з вами не просто у вигляді чатику, а як тепле тріпотіння всередині: у мене є мої люди. І це — безцінно.
// І, звісно, цей квест не має закінчення! Тому що дружба — це безперервна і прекрасна пригода.
	
// Тепер питання: коли і де ви зустрічаєтеся, щоб відсвяткувати вашу перемогу, зіроньки? 🥂 😉`
		
		_, err = bot.Send(&telebot.Chat{ID: chat.ID}, finishMessage, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "FinishGameHandler",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending final message to the group")
		}

		// _, err := bot.Send(
        //     &telebot.Chat{ID: chatID},
        //     msg,
		// 	inlineKeys,
        //     telebot.ModeMarkdown,
        // )

		storage_db.UpdateGameStatus(int64(game.ID), models.StatusGameFinished)

		time.Sleep(5 * time.Second)

		SendReferalMsg(bot)(c)

		time.Sleep(5 * time.Second)

		SendFeedbackMsg(bot)(c)

		time.Sleep(5 * time.Second)

		SendBuyMeCoffeeMsg(bot)(c)

		return nil
	}
}

func FinishTestHandler(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		_, err := bot.Send(&telebot.Chat{ID: chat.ID}, finishMessage, telebot.ModeMarkdown)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "FinishGameHandler",
				"group": chat.Title,
				"err": err,
			}).Error("Error sending final message to the group")
		}
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

		// switch idTask {
		// case 3:
		// 	subtasks.WhoIsUsSubTask(bot)(c)
		// 	return nil
		// case 7:
		// 	// call function for subtask for task 7
		// case 12:
		// 	// call function for subtask for task 12
		// }

		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, idTask)
		if err != nil {
			utils.Logger.Errorf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			//textYouAlreadyAnswered := fmt.Sprintf("@%s, ти вже відповіла на це завдання 😅", user.Username)
			msgYouAlreadyAnswered, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
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

			// return c.Send(fmt.Sprintf("@%s, ти вже відповідала на це завдання 😉", user.Username))
			return nil
		case status.AlreadySkipped:
			return c.Send(fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		}

		storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(idTask))

		switch idTask {
		case 3:
			subtasks.WhoIsUsSubTask(bot)(c)
			return nil
		case 7:
			// call function for subtask for task 7
		case 12:
			// call function for subtask for task 12
		}

		awaitingAnswerMsg, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(wantAnswerMessages), user.Username))
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
			bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
		case status.AlreadySkipped:
			bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		case status.SkipLimitReached:
			msg, _ := bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipLimitReached), user.Username))
			
			// Delay delete the message max skip tasks
			time.AfterFunc(cfg.Durations.TimeDeleteMsgMaxSkipTasks, func() {
				err = bot.Delete(msg)
				if err != nil {
					utils.Logger.Errorf("Error deleting skip limit reached message for user %s: %v", user.Username, err)
				}
			})
		default:
			switch status.RemainingSkips-1 {
			case 0:
				bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipThirdTime), user.Username))
			case 1:
				bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipSecondTime), user.Username))
			case 2:
				bot.Send(chat, fmt.Sprintf(utils.GetStaticMessage(skipMessages, models.MsgSkipFirstTime), user.Username))
			}
			// Skip messages
			//bot.Send(chat, fmt.Sprintf("✅ @%s, завдання пропущено! У тебе залишилось %d пропуск(ів).", user.Username, status.RemainingSkips-1))
		}

		return nil
	}
}

// func HandleSubTask3(bot *telebot.Bot) func(c telebot.Context) error {
//     return func(c telebot.Context) error {
//         utils.Logger.Infof("HandleSubTask3 called by user %s in group %s", c.Sender().Username, c.Chat().Title)
        
//         user := c.Sender()
// 		chat := c.Chat()
//         msg := c.Message()
//         data := c.Data()

// 		game, _ := storage_db.GetGameByChatId(chat.ID)
        
//         // Check if this is a subtask callback
//         if !strings.HasPrefix(data, "subtask_") {
//             return nil
//         }
        
//         // Remove prefix "subtask_" first
//         dataWithoutPrefix := strings.TrimPrefix(data, "subtask_")
//         // Now: "3_0_123456789|Jay_jayss"
        
//         // Find first underscore (taskID separator)
//         firstUnderscore := strings.Index(dataWithoutPrefix, "_")
//         if firstUnderscore == -1 {
//             utils.Logger.Errorf("Invalid callback data format: %s", data)
//             return c.Send("Помилка формату даних кнопки")
//         }
        
//         taskIdStr := dataWithoutPrefix[:firstUnderscore]
//         remainder := dataWithoutPrefix[firstUnderscore+1:]
       
// 		// Convert strings to appropriate types
//         taskID, err := strconv.Atoi(taskIdStr)
//         if err != nil {
//             utils.Logger.Errorf("Invalid task ID: %s", taskIdStr)
//             return c.Send("Некоректний ID завдання")
//         }
        
//         // Find second underscore (questionIndex separator)
//         secondUnderscore := strings.Index(remainder, "_")
//         if secondUnderscore == -1 {
//             utils.Logger.Errorf("Invalid callback data format: %s", data)
//             return c.Send("Помилка формату даних кнопки")
//         }
        
//         questionIndexStr := remainder[:secondUnderscore]
//         userPart := remainder[secondUnderscore+1:]
		
// 		questionIndex1, _ := strconv.ParseInt(questionIndexStr, 10, 64)
// 		questionIndex := uint(questionIndex1)
//         // Parse user part: {userID}|{userName}
//         pipeIndex := strings.Index(userPart, "|")
//         if pipeIndex == -1 {
//             utils.Logger.Errorf("Invalid user data format in callback: %s", userPart)
//             return c.Send("Помилка формату даних користувача")
//         }
        
//         selectedUserIdStr := userPart[:pipeIndex]    
		
// 		selectedUserID, err := strconv.ParseInt(selectedUserIdStr, 10, 64)
//         if err != nil {
//             utils.Logger.Errorf("Invalid user ID: %s", selectedUserIdStr)
//             return c.Send("Некоректний ID користувача")
//         }// "123456789"
//         selectedUsername := userPart[pipeIndex+1:]   // "Jay_jayss"
        
//         utils.Logger.WithFields(logrus.Fields{
//             "source":            "HandleSubTask3",
//             "user_id":           user.ID,
//             "username":          user.Username,
//             "group":             c.Chat().Title,
//             "task_id":           taskID,
//             "question_index":    questionIndex,
//             "selected_user_id":  selectedUserID,
//             "selected_username": selectedUsername,
//         }).Infof("User %s selected user %s in group %s", user.Username, selectedUsername, c.Chat().Title)

// 		subTaskAnswer := &models.SubtaskAnswer{
// 			GameID: game.ID,
// 			TaskID: taskID,
// 			QuestionIndex: questionIndex,
// 			AnswererUserID: user.ID,
// 			SelectedUserID: selectedUserID,
// 			SelectedUsername: selectedUsername,
// 		}

// 		err = storage_db.AddSubtaskAnswer(subTaskAnswer)
// 		if err != nil {
// 			utils.Logger.Errorf("Error add subtask answer to DB: %v", err)
// 		} else {
// 			utils.Logger.Infof("Answe of subtask add to DB: succes")
// 		}
        
//         err = bot.Delete(msg)
//         if err != nil {
//             utils.Logger.WithFields(logrus.Fields{
//                 "source":   "HandleSubTask3",
//                 "user_id":  user.ID,
//                 "username": user.Username,
//                 "group":    c.Chat().Title,
//             }).Errorf("Failed to delete message with subtask 3: %v", err)
//             return nil
//         }
        
//         return nil
//     }
// }

func HandleSubTask3(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        user := c.Sender()
        chat := c.Chat()
        msg := c.Message()
        data := c.Data()
        
        // Check if this is a subtask callback
        if !strings.HasPrefix(data, "subtask_") {
            return nil
        }
        
        game, err := storage_db.GetGameByChatId(chat.ID)
        if err != nil {
            utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
            return c.Send("Помилка отримання гри")
        }
        
        // Parse callback data
        // ... (ваш существующий код парсинга) ...
        
        // Remove prefix "subtask_" first
        dataWithoutPrefix := strings.TrimPrefix(data, "subtask_")
        
        // Parse data (your existing parsing code)
        firstUnderscore := strings.Index(dataWithoutPrefix, "_")
        taskIDStr := dataWithoutPrefix[:firstUnderscore]
        remainder := dataWithoutPrefix[firstUnderscore+1:]
        
        secondUnderscore := strings.Index(remainder, "_")
        questionIndexStr := remainder[:secondUnderscore]
        userPart := remainder[secondUnderscore+1:]
        
        pipeIndex := strings.Index(userPart, "|")
        selectedUserIDStr := userPart[:pipeIndex]
        selectedUsername := userPart[pipeIndex+1:]
        
        // Convert to types
        taskID, _ := strconv.Atoi(taskIDStr)
        questionIndex, _ := strconv.ParseUint(questionIndexStr, 10, 32)
        selectedUserID, _ := strconv.ParseInt(selectedUserIDStr, 10, 64)
        
		status, err := storage_db.CheckPlayerResponseStatus(user.ID, game.ID, taskID)
		if err != nil {
			utils.Logger.Errorf("Error checking player response status: %v", err)
			return nil
		}

		switch {
		case status.AlreadyAnswered:
			//textYouAlreadyAnswered := fmt.Sprintf("@%s, ти вже відповіла на це завдання 😅", user.Username)
			msgYouAlreadyAnswered, err := bot.Send(chat, fmt.Sprintf(utils.GetRandomMsg(alreadyAnswerMessages), user.Username))
			if err != nil {
				utils.Logger.Errorf("Error sending message that user %s already answered task %d: %v", user.Username, taskID, err)
			}

			time.AfterFunc(cfg.Durations.TimeDeleteMsgYouAlreadyAnswered, func() {
				err = bot.Delete(msgYouAlreadyAnswered)
				if err != nil {
					utils.Logger.WithFields(logrus.Fields{
						"source": "OnAnswerTaskBtnHandler",
						"username": user.Username,
						"group": chat.Title,
						"data_button": data,
						"task_id": taskID,
					}).Errorf("Error deleting message that user %s already answered task %d: %v", user.Username, taskID, err)
				}
			})

			// return c.Send(fmt.Sprintf("@%s, ти вже відповідала на це завдання 😉", user.Username))
			return nil
			case status.AlreadySkipped:
			return c.Send(fmt.Sprintf(utils.GetStaticMessage(staticMessages, models.MsgUserAlreadySkipTask), user.Username))
		}

		//storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerWaiting+strconv.Itoa(taskID))

        // Check if user has active session
        session, exists := subtasks.GlobalSessionManager.GetActiveSession(game.ID)
        if !exists || session.UserID != user.ID {
			msgTextOtherUserAnswer := fmt.Sprintf("@%s донт пуш зе хорсес! Інша зірочка зараз відповідає.", user.Username)

			_, err = msgmanager.SendTemporaryMessage(
				chat.ID,
				user.ID,
				msgmanager.TypeNotInGame,
				msgTextOtherUserAnswer,
				10 * time.Second,
			)
			if err != nil {
					utils.Logger.Errorf("Error sending message that user %s is not in game: %v", user.Username, err)
			}

            return nil
        }
        
        utils.Logger.WithFields(logrus.Fields{
            "source":            "HandleSubTask3",
            "username":          user.Username,
            "task_id":           taskID,
            "question_index":    uint(questionIndex),
            "selected_user_id":  selectedUserID,
            "selected_username": selectedUsername,
        }).Infof("User %s selected user %s", user.Username, selectedUsername)
        
        // Delete the question message
        err = bot.Delete(msg)
        if err != nil {
            utils.Logger.Errorf("Failed to delete message: %v", err)
        }
        
        // Save answer and check if completed
        completed, err := subtasks.GlobalSessionManager.SaveAnswerAndNext(game.ID, selectedUsername)
        if err != nil {
            utils.Logger.Errorf("Error saving subtask answer: %v", err)
            return c.Send("Помилка збереження відповіді")
        }
        
        if completed {
            // All questions answered
            answers := subtasks.GlobalSessionManager.CompleteSession(game.ID)
            
            utils.Logger.WithFields(logrus.Fields{
                "source":        "HandleSubTask3",
                "username":      user.Username,
                "total_answers": len(answers),
                "task_id":       taskID,
            }).Info("Subtask completed")

			playerResponse := &models.PlayerResponse{
				PlayerID:   user.ID,
				GameID: 	game.ID,
				TaskID:		taskID,
				HasResponse: true,
				Skipped: false,
			}

			storage_db.AddPlayerResponse(playerResponse)

			storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)
            
            return c.Send(fmt.Sprintf("@%s, дякую за відповідь, кицю 🐈Очікуй результатів, коли всі подружки поділяться своєю думкою 💁‍♀️", user.Username))
        }
        
        // Send next question
        return subtasks.SendCurrentQuestion(bot, c, game.ID)
    }
}

// func RegisterCallbackHandlers(bot *telebot.Bot) {
// 	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
// 		data := c.Callback().Data

// 		utils.Logger.WithFields(logrus.Fields{
// 			"source": "RegisterCallbackHandlers",
// 			"data": data,
// 			"user_id": c.Sender().ID,
// 			"username": c.Sender().Username,
// 			"group": c.Chat().Title,
// 		}).Info("Callback handler called")

// 		switch {
// 		case strings.HasPrefix(data, "\fexit_game_"):
// 			return handleExitGame(bot, c)
// 		case strings.HasPrefix(data, "\fexit_"):
// 			return handleExitConfirm(bot, c)
// 		// case data == "support_menu":
// 		// 	return handleSupportMenu(bot, c)
// 		case data == "\fhelp_menu":
// 			return handleHelpMenu(bot, c)
// 		case data == "\freturn_to_game":
// 			return handleReturnToGame(bot, c)
// 		// case strings.HasPrefix(data, "\fphoto_choice_"):
// 		// 	return HandlePhotoChoice(bot)(c)
// 		case strings.HasPrefix(data,"\fwaiting_" ):
// 			return OnAnswerTaskBtnHandler(bot)(c)
// 		case strings.HasPrefix(data,"\fskip_"):
// 			return OnSkipTaskBtnHandler(bot)(c)
// 		case strings.HasPrefix(data, "subtask_3_"):
// 			return HandleSubTask3(bot)(c)
// 		default:
// 			return nil
// 		}
// 	})
// }

// func CreateCollageFromResultsImage(bot *telebot.Bot) func(c telebot.Context) error {
// 	return func(c telebot.Context) error {
// 		return CreateCollageWithGG()
// 	}
// }

// func CreateCollageWithGG() error {
//     fmt.Println("=== START CREATE COLLAGE ===")
    
//     // Find all images in results folder
//     imagePaths, err := filepath.Glob("results_1/*.jpg")
//     if err != nil {
//         return fmt.Errorf("error reading results folder: %w", err)
//     }
    
//     if len(imagePaths) == 0 {
//         return fmt.Errorf("no images found in results folder")
//     }
    
//     // Limit to 6 images
//     if len(imagePaths) > 6 {
//         imagePaths = imagePaths[:6]
//     }
    
//     fmt.Printf("Finded %d images for collage\n", len(imagePaths))
    
//     // Image cell sizes
//     imageWidth := 540
//     imageHeight := 640
    
//     totalWidth := imageWidth * 2    // 1080
//     totalHeight := imageHeight * 3  // 1920
    
//     fmt.Printf("Create canvas: %dx%d\n", totalWidth, totalHeight)
    
//     dc := gg.NewContext(totalWidth, totalHeight)
    
//     // Very dark background
//     dc.SetRGB(0.05, 0.05, 0.05)
//     dc.Clear()
    
//     // Positions for 2x3 grid - NO MARGINS
//     positions := [][2]int{
//         {0, 0},
//         {imageWidth, 0},
//         {0, imageHeight},
//         {imageWidth, imageHeight},
//         {0, imageHeight * 2},
//         {imageWidth, imageHeight * 2},
//     }
    
//     // Process each image
//     for i, path := range imagePaths {
//         if i >= 6 { break }
        
//         fmt.Printf("Image processing %d: %s\n", i+1, path)
        
//         img, err := gg.LoadImage(path)
//         if err != nil {
//             fmt.Printf("ERROR download of image %s: %v\n", path, err)
//             continue
//         }
        
//         // Enlarging the image scale
//         processedImg := showBiggerImageWithLessMargin(img, imageWidth, imageHeight)
        
//         x := positions[i][0]
//         y := positions[i][1]
        
//         fmt.Printf("Placing an image in position: %d, %d\n", x, y)
        
//         dc.DrawImageAnchored(processedImg, x+imageWidth/2, y+imageHeight/2, 0.5, 0.5)
//     }
    
//     // Reduced text by 20%
//     centerX := totalWidth / 2
//     centerY := totalHeight / 2
    
//     fmt.Println("Add text (reduced by 20%)...")
    
//     // We draw the text graphically - each letter separately
//     drawSmallerText(dc, centerX, centerY)
    
//     fmt.Println("Save file...")
    
//     // Save as JPEG
//     img := dc.Image()
    
//     file, err := os.Create("collage.jpg")
//     if err != nil {
//         return fmt.Errorf("failed to create collage file: %w", err)
//     }
//     defer file.Close()
    
//     err = jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
//     if err != nil {
//         return fmt.Errorf("failed to encode JPEG: %w", err)
//     }
    
//     fmt.Println("=== COLLAGE SAVED AS collage.jpg ===")
//     return nil
// }

// // Show MORE of the image with less padding
// func showBiggerImageWithLessMargin(img image.Image, targetWidth, targetHeight int) image.Image {
//     bounds := img.Bounds()
//     imgWidth := bounds.Dx()
//     imgHeight := bounds.Dy()
    
//     fmt.Printf("Original size: %dx%d\n", imgWidth, imgHeight)
    
//     // Calculate the scale so that the image fills more space
//     scaleX := float64(targetWidth) / float64(imgWidth)
//     scaleY := float64(targetHeight) / float64(imgHeight)
    
//     // Increase fill from 95% to 98% for large images
//     scale := math.Min(scaleX, scaleY) * 0.98
    
//     fmt.Printf("Zoom in: %.2f\n", scale)
    
//     newWidth := int(float64(imgWidth) * scale)
//     newHeight := int(float64(imgHeight) * scale)
    
//     fmt.Printf("Zoom size: %dx%d\n", newWidth, newHeight)
    
//     // Create context scope size
//     dc := gg.NewContext(targetWidth, targetHeight)
    
//     // Darker background for less contrast with padding
//     dc.SetRGB(0.08, 0.08, 0.08)
//     dc.Clear()
    
//     // Center the image
//     offsetX := (targetWidth - newWidth) / 2
//     offsetY := (targetHeight - newHeight) / 2
    
//     fmt.Printf("Bias: %d, %d\n", offsetX, offsetY)
    
//     // Draw a scaled image Offset
//     dc.DrawImageAnchored(img, offsetX+newWidth/2, offsetY+newHeight/2, 0.5, 0.5)
    
//     return dc.Image()
// }

// func drawSmallerText(dc *gg.Context, centerX, centerY int) {
//     // Sizes for text
//     letterHeight := 90.0  // old:150
//     letterWidth := 50.0    // old: 100
//     spacing := 14.0        // old: 20
    
//     // text "MEMORIES"
//     text := "MEMORIES"
//     totalWidth := float64(len(text)) * (letterWidth + spacing) - spacing
    
//     startX := float64(centerX) - totalWidth/2
//     startY := float64(centerY)
    
//     // Reduced background for text
//     bgPadding := 45.0  // old 30
//     dc.SetRGBA(0, 0, 0, 0.4)
//     dc.DrawRoundedRectangle(
//         startX - bgPadding,
//         startY - letterHeight/2 - bgPadding,
//         totalWidth + bgPadding*2,
//         letterHeight + bgPadding*2,
//         16,  // old 20
//     )
//     dc.Fill()
    
//     // draw each letter as simple lines
//     dc.SetRGB(1, 1, 1) // White color
//     dc.SetLineWidth(6)  // old 8
    
//     for i, char := range text {
//         x := startX + float64(i) * (letterWidth + spacing)
//         drawLetter(dc, char, x, startY, letterWidth, letterHeight)
//     }
// }


// // draw each letter as simple lines
// func drawLetter(dc *gg.Context, char rune, x, y, width, height float64) {
//     halfWidth := width / 2
//     halfHeight := height / 2
    
//     switch char {
//     case 'M':
//         // letter M
//         dc.DrawLine(x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight) 
//         dc.DrawLine(x+halfWidth, y+halfHeight, x+halfWidth, y-halfHeight) 
//         dc.DrawLine(x-halfWidth, y-halfHeight, x, y) 
//         dc.DrawLine(x, y, x+halfWidth, y-halfHeight) 
        
//     case 'E':
//         // letter E
//         dc.DrawLine(x-halfWidth, y-halfHeight, x-halfWidth, y+halfHeight) 
//         dc.DrawLine(x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight)
//         dc.DrawLine(x-halfWidth, y, x+halfWidth*0.7, y)
//         dc.DrawLine(x-halfWidth, y+halfHeight, x+halfWidth, y+halfHeight)
//         dc.Stroke()
        
//     case 'O':
//         // letter O - овал
//         dc.DrawEllipse(x, y, halfWidth, halfHeight)
//         dc.Stroke()
        
//     case 'R':
//         // letter R
//         dc.DrawLine(x-halfWidth, y-halfHeight, x-halfWidth, y+halfHeight)
//         dc.DrawLine(x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight)
//         dc.DrawLine(x+halfWidth, y-halfHeight, x+halfWidth, y)
//         dc.DrawLine(x+halfWidth, y, x-halfWidth, y)
//         dc.DrawLine(x, y, x+halfWidth, y+halfHeight)
//         dc.Stroke()
        
//     case 'I':
//         // letter I
//         dc.DrawLine(x, y-halfHeight, x, y+halfHeight)
//         dc.DrawLine(x-halfWidth*0.5, y-halfHeight, x+halfWidth*0.5, y-halfHeight)
//         dc.DrawLine(x-halfWidth*0.5, y+halfHeight, x+halfWidth*0.5, y+halfHeight)
//         dc.Stroke()
        
//     case 'S':
//         // letter S 
//         dc.DrawLine(x+halfWidth, y-halfHeight, x-halfWidth, y-halfHeight)
//         dc.DrawLine(x-halfWidth, y-halfHeight, x-halfWidth, y)
//         dc.DrawLine(x-halfWidth, y, x+halfWidth, y)
//         dc.DrawLine(x+halfWidth, y, x+halfWidth, y+halfHeight)
//         dc.DrawLine(x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight)
//         dc.Stroke()
        
//     default:
//         // Unknown letter - draw a rectangle
//         dc.DrawRectangle(x-halfWidth, y-halfHeight, width, height)
//         dc.Stroke()
//     }
// }


// // Helper function to process image for square format
// func processImageForSquare(img image.Image, targetSize int) image.Image {
//     bounds := img.Bounds()
//     imgWidth := bounds.Dx()
//     imgHeight := bounds.Dy()
    
//     // Create new square context
//     dc := gg.NewContext(targetSize, targetSize)
//     dc.SetRGB(1, 1, 1) // white background
//     dc.Clear()
    
//     // Calculate scaling to fit image in square while maintaining aspect ratio
//     scale := math.Min(float64(targetSize)/float64(imgWidth), float64(targetSize)/float64(imgHeight))*0.7
    
//     newWidth := int(float64(imgWidth) * scale)
//     newHeight := int(float64(imgHeight) * scale)
    
//     // Center the scaled image
//     offsetX := (targetSize - newWidth) / 2
//     offsetY := (targetSize - newHeight) / 2
    
//     // Draw the scaled image
//     dc.DrawImageAnchored(img, offsetX+newWidth/2, offsetY+newHeight/2, 0.5, 0.5)
    
//     return dc.Image()
// }

// // Bot handler function for creating collage - updated for JPEG
// func CreateCollageFromResults(bot *telebot.Bot) func(c telebot.Context) error {
//     return func(c telebot.Context) error {
//         chat := c.Chat()
        
//         // Get all images from results folder
//         imagePaths, err := filepath.Glob("results/*.jpg")
//         if err != nil {
//             return c.Send("❌ Помилка читання папки з результатами")
//         }
        
//         if len(imagePaths) == 0 {
//             return c.Send("❌ Не знайдено зображень для створення колажу")
//         }
        
//         // Limit to 6 images
//         if len(imagePaths) > 6 {
//             imagePaths = imagePaths[:6]
//         }
        
//         c.Send(fmt.Sprintf("🎨 Створюю колаж з %d зображень...", len(imagePaths)))
        
//         // Create collage
//         err = CreateCollageWithGG()
//         if err != nil {
//             return c.Send(fmt.Sprintf("❌ Помилка створення колажу: %v", err))
//         }
        
//         // Check if collage file was created
//         if _, err := os.Stat("collage.jpg"); os.IsNotExist(err) {
//             return c.Send("❌ Файл колажу не був створений")
//         }
        
//         // Send as photo (JPEG is better supported by Telegram)
//         photo := &telebot.Photo{
//             File:    telebot.FromDisk("collage.jpg"),
//             Caption: "🎨 Ваш колаж готовий!\n\n📱 Ідеально підходить для обоїв телефону",
//         }
        
//         _, err = bot.Send(chat, photo)
//         if err != nil {
//             return c.Send(fmt.Sprintf("❌ Помилка відправки колажу: %v", err))
//         }
        
//         // Also send as document for highest quality
//         document := &telebot.Document{
//             File:     telebot.FromDisk("collage.jpg"),
//             MIME:     "image/jpeg",
//             FileName: "wallpaper_collage.jpg",
//             Caption:  "📱 Високоякісна версія для завантаження як обої",
//         }
        
//         bot.Send(chat, document)
        
//         // Clean up - remove the collage file after sending
//         go func() {
//             // Wait a bit to ensure file was sent, then clean up
//             time.Sleep(5 * time.Second)
//             os.Remove("collage.jpg")
//             fmt.Println("Temporary collage file cleaned up")
//         }()
        
//         return nil
//     }
// }

// // Optional: Function to create a quick preview collage (smaller size for testing)
// func CreatePreviewCollage() error {
//     imagePaths, err := filepath.Glob("results/*.jpg")
//     if err != nil {
//         return fmt.Errorf("error reading results folder: %w", err)
//     }
    
//     if len(imagePaths) == 0 {
//         return fmt.Errorf("no images found in results folder")
//     }
    
//     // Limit to 6 images
//     if len(imagePaths) > 6 {
//         imagePaths = imagePaths[:6]
//     }
    
//     // Create smaller canvas for preview
//     dc := gg.NewContext(540, 960) // Half size of main collage
    
//     // Simple white background for preview
//     dc.SetRGB(1, 1, 1)
//     dc.Clear()
    
//     // Calculate smaller layout
//     imageSize := 160
//     spacing := 20
//     marginTop := 50
//     marginSide := (540 - (2*imageSize + spacing)) / 2
    
//     // Positions for 2x3 grid
//     positions := [][2]int{
//         {marginSide, marginTop},
//         {marginSide + imageSize + spacing, marginTop},
//         {marginSide, marginTop + imageSize + spacing},
//         {marginSide + imageSize + spacing, marginTop + imageSize + spacing},
//         {marginSide, marginTop + 2*(imageSize + spacing)},
//         {marginSide + imageSize + spacing, marginTop + 2*(imageSize + spacing)},
//     }
    
//     for i, path := range imagePaths {
//         if i >= 6 { break }
        
//         img, err := gg.LoadImage(path)
//         if err != nil {
//             continue
//         }
        
//         processedImg := processImageForSquare(img, imageSize)
        
//         x := positions[i][0]
//         y := positions[i][1]
        
//         dc.DrawImageAnchored(processedImg, x+imageSize/2, y+imageSize/2, 0.5, 0.5)
//     }
    
//     // Save preview as JPEG
//     img := dc.Image()
    
//     file, err := os.Create("collage_preview.jpg")
//     if err != nil {
//         return fmt.Errorf("failed to create preview file: %w", err)
//     }
//     defer file.Close()
    
//     err = jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
//     if err != nil {
//         return fmt.Errorf("failed to encode preview JPEG: %w", err)
//     }
    
//     fmt.Println("Preview collage saved as collage_preview.jpg")
//     return nil
// }