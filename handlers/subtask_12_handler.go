package handlers

import (
	"fmt"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

type Subtask12State struct {
    CurrentIndex int
    MsgToDelete  []int
    WaitingAnswer bool // флаг: ждём ли сейчас ответа от админа
}

var subtask12States sync.Map // key: chatID (int64), value: *Subtask12State

func getSubtask12State(chatID int64) (*Subtask12State, bool) {
    val, ok := subtask12States.Load(chatID)
    if !ok {
        return nil, false
    }
    return val.(*Subtask12State), true
}

func HandleSubTask12(c telebot.Context) error {
	//user := c.Sender()
	chat := c.Chat()
	//game, err := storage_db.GetGameByChatId(chat.ID)
	// if err != nil {
	// 	utils.Logger.Errorf("Error getting game by chat ID: %v", err)
	// 	return err
	// }

	// subtasks, err := utils.LoadArrayMsgs("internal/data/tasks/subtasks/subtask_12.json")
	// if err != nil {
	// 	utils.Logger.Errorf("Error loading subtask 12 messages: %v", err)
	// 	return err
	// }

	subtask12States.Store(chat.ID, &Subtask12State{
        CurrentIndex:  0,
        WaitingAnswer: false,
    })

	// awaitMsgs, err := utils.LoadArrayMsgs("internal/data/messages/group/subtask_12/await_msgs.json")
	// if err != nil {
	// 	utils.Logger.Errorf("Error loading subtask 12 await messages: %v", err)
	// 	return err
	// }

	// replyMsgs, err := utils.LoadArrayMsgs("internal/data/messages/group/subtask_12/reply_msgs.json")
	// if err != nil {
	// 	utils.Logger.Errorf("Error loading subtask 12 reply messages: %v", err)
	// 	return err
	// }		
	return sendSubtask12Question(c.Bot(), chat.ID, 0)
}

func sendSubtask12Question(bot *telebot.Bot, chatID int64, index int) error {
    chat := &telebot.Chat{ID: chatID}

	subtasks, err := utils.LoadArrayMsgs("internal/data/tasks/subtasks/subtask_12.json")
	if err != nil {
		utils.Logger.Errorf("Error loading subtask 12 messages: %v", err)
		return err
	}

    markup := &telebot.ReplyMarkup{}
    shareBtn := markup.Data("Поділитися мрією 🌙", "subtask12_share")
    markup.Inline(markup.Row(shareBtn))

    sent, err := bot.Send(chat, subtasks[index], markup, telebot.ModeMarkdown)
    if err != nil {
        return err
    }

    if state, ok := getSubtask12State(chatID); ok {
        state.MsgToDelete = append(state.MsgToDelete, sent.ID)
    }

    return nil
}

func HandleSubtask12ShareBtn(c telebot.Context) error {
    user := c.Sender()
    chat := c.Chat()
    bot := c.Bot()

    game, err := storage_db.GetGameByChatId(chat.ID)
    if err != nil {
        utils.Logger.Errorf("Error getting game: %v", err)
        return c.Respond(&telebot.CallbackResponse{Text: "Помилка отримання гри"})
    }

	userRole, err := storage_db.GetPlayerRoleByUserIDAndGameID(user.ID, game.ID)
	if err != nil {
		utils.Logger.Infof("Error getting player role for user %s in game %d during answering task %d: %v", user.Username, game.ID, game.CurrentTaskID, err)
	}

    // Не админ — показываем алерт
    if userRole != "admin" {
        c.Respond() // убираем часики

        warning, err := bot.Send(c.Chat(),fmt.Sprintf("@%s, тільки адмін може відповідати на це завдання 😊", user.Username))
        if err != nil {
            utils.Logger.Errorf("Error sending warning: %v", err)
            return nil
        }

        time.AfterFunc(5*time.Second, func() {
            bot.Delete(warning)
        })
        return nil
    }

    state, ok := getSubtask12State(chat.ID)
    if !ok {
        return c.Respond(&telebot.CallbackResponse{Text: "Помилка стану гри"})
    }

    // Уже ждём ответа — игнорируем повторное нажатие
    if state.WaitingAnswer {
        // return c.Respond(&telebot.CallbackResponse{
        //     Text:      "Очікую твою відповідь ✍️",
        //     ShowAlert: false,
        // })
        return c.Respond()
    }

    c.Respond() // убираем "часики"

    awaitMsgs, err := utils.LoadArrayMsgs("internal/data/messages/group/subtask_12/await_msgs.json")
    if err != nil {
        return err
    }

    // Отправляем await сообщение по индексу текущего вопроса
    awaitSent, err := bot.Send(&telebot.Chat{ID: chat.ID}, awaitMsgs[state.CurrentIndex], telebot.ModeMarkdown)
    if err != nil {
        return err
    }
    state.MsgToDelete = append(state.MsgToDelete, awaitSent.ID)
    state.WaitingAnswer = true

    return nil
}

func HandleSubtask12Answer(c telebot.Context) bool {
    chat := c.Chat()
    user := c.Sender()
    bot := c.Bot()

    state, ok := getSubtask12State(chat.ID)
    if !ok || !state.WaitingAnswer {
        return false
    }

    game, err := storage_db.GetGameByChatId(chat.ID)
    if err != nil {
        utils.Logger.Errorf("Error getting game: %v", err)
        return true
    }

    userAdmin, err := storage_db.GetAdminPlayerByGameID(game.ID)
    if err != nil {
        utils.Logger.Errorf("Error getting admin player: %v", err)
        return true
    }

    if user.ID != userAdmin.ID {
        return false
    }

    // *** СРАЗУ блокируем повторные ответы ***
    state.WaitingAnswer = false

    userMsg := c.Message()
    answer := userMsg.Text

    // Сохраняем ответ в БД
    if err := storage_db.SaveTask12Answer(int64(game.ID), chat.ID, state.CurrentIndex+1, answer); err != nil {
        utils.Logger.Errorf("Error saving answer: %v", err)
    } else {
        utils.Logger.Infof("Answer saved for game %d, question %d: %s", game.ID, state.CurrentIndex+1, answer)
    }

    replyMsgs, err := utils.LoadArrayMsgs("internal/data/messages/group/subtask_12/reply_msgs.json")
    if err != nil {
        utils.Logger.Errorf("Error loading reply msgs: %v", err)
        return true
    }

    // *** Проверка границ перед обращением к слайсу ***
    // Надо — отправляем итоги если вышли за границы:
    if state.CurrentIndex >= len(replyMsgs) {
        utils.Logger.Warnf("CurrentIndex %d out of range, replyMsgs len: %d — finishing subtask12", state.CurrentIndex, len(replyMsgs))
        subtask12States.Delete(chat.ID)
        storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)
        chatObj := &telebot.Chat{ID: chat.ID}

        if err := sendSubtask12Results(bot, chatObj, int64(game.ID)); err != nil {
            utils.Logger.Errorf("Error sending results: %v", err)
        }
        return true
    }

    replySent, err := bot.Reply(userMsg, replyMsgs[state.CurrentIndex], telebot.ModeMarkdown)
    if err != nil {
        utils.Logger.Errorf("Error sending reply: %v", err)
        return true
    }

    // Сохраняем локально до горутины
    currentIndex := state.CurrentIndex
    msgsToDelete := make([]int, len(state.MsgToDelete))
    copy(msgsToDelete, state.MsgToDelete)
    msgsToDelete = append(msgsToDelete, userMsg.ID, replySent.ID)
    state.MsgToDelete = nil

    go func() {
        time.Sleep(5 * time.Second)

        chatObj := &telebot.Chat{ID: chat.ID}
        for _, msgID := range msgsToDelete {
            bot.Delete(&telebot.Message{ID: msgID, Chat: chatObj})
        }

        subtasks, err := utils.LoadArrayMsgs("internal/data/tasks/subtasks/subtask_12.json")
        if err != nil {
            utils.Logger.Errorf("Error loading subtasks: %v", err)
            return
        }

        nextIndex := currentIndex + 1
        state.CurrentIndex = nextIndex

        if nextIndex >= len(subtasks) {
            subtask12States.Delete(chat.ID)
            // *** Сбрасываем статус юзера в БД ***
            playerResponse := &models.PlayerResponse{
                PlayerID:     user.ID,
                UserName:     user.Username,
                GameID:       game.ID,
                TaskID:       12,
                HasResponse:  true,
                Skipped:      false,
                NotificationSent: 0,
            }
            storage_db.AddPlayerResponse(playerResponse)

            storage_db.UpdatePlayerStatus(user.ID, models.StatusPlayerNoWaiting)
            if err := sendSubtask12Results(bot, chatObj, int64(game.ID)); err != nil {
                utils.Logger.Errorf("Error sending results: %v", err)
            }
            return
        }

        if err := sendSubtask12Question(bot, chat.ID, nextIndex); err != nil {
            utils.Logger.Errorf("Error sending next question: %v", err)
        }
    }()

    return true
}

func sendSubtask12Results(bot *telebot.Bot, chat *telebot.Chat, gameID int64) error {
    answers, err := storage_db.GetTask12Answers(gameID, chat.ID)
    if err != nil {
        return err
    }

    text := "🌟 *Наші мрії на наступний рік:*\n\n"
    for i, a := range answers {
        text += fmt.Sprintf("%d️⃣ %s\n\n", i+1, a.Answer)
    }

    _, err = bot.Send(chat, text, telebot.ModeMarkdown)
    return err
}