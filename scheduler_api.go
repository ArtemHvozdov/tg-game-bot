package main

import (
	"encoding/json"
	"fmt"

	//"fmt"
	"log"
	"net/http"

	//"time"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"

	telebot "gopkg.in/telebot.v3"
)

// Структура для запросов от планировщика
type TaskRequest struct {
    GameID int64 `json:"game_id"`
    TaskID int64 `json:"task_id"`
}

// Run HTTP server
func startHTTPServer(bot *telebot.Bot) {
    mux := http.NewServeMux()

    // Эндпоинт для подведения итогов
    mux.HandleFunc("/api/create-summary", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
            return
        }

        var req TaskRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            log.Printf("Failed to decode request: %v", err)
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        log.Printf("Received create-summary request for game %d, task %d", req.GameID, req.TaskID)

        // Вызываем функцию создания итогов
        if err := createTaskSummary(bot, req.GameID, req.TaskID); err != nil {
            log.Printf("Failed to create summary: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })

    // Эндпоинт для отправки следующей таски
    mux.HandleFunc("/api/send-next-task", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
            return
        }

        var req TaskRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            log.Printf("Failed to decode request: %v", err)
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        log.Printf("Received send-next-task request for game %d", req.GameID)

        // Вызываем функцию отправки новой таски
        // if err := sendNextTask(bot, db, req.GameID); err != nil {
        //     log.Printf("Failed to send next task: %v", err)
        //     http.Error(w, err.Error(), http.StatusInternalServerError)
        //     return
        // }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })

    //Healthcheck эндпоинт
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    log.Println("HTTP API server listening on :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal("Failed to start HTTP server:", err)
    }
}

// Task summary creation function
func createTaskSummary(bot *telebot.Bot, gameID, taskID int64) error {
    // Get game info
    game, err := storage_db.GetGameById(int(gameID))
    if err != nil {
        return err
    }

	utils.Logger.WithFields(logrus.Fields{
		"game_id": game.ID,
		"game_name": game.Name,
		"chat_id": game.GameChatID,
	}).Info("Info by game during API req")

	playerHasAnswered, err := storage_db.GetPlayersWithAnswer(gameID, taskID)
	if err != nil {
		return err
	}

	playerSkipped, err := storage_db.GetPlayersWhoSkipped(gameID, taskID)
	if err != nil {
		return err
	}

	txt1 := fmt.Sprintf("Самарі гри %s:\nГравці, що відповіли:", game.Name)
	txt2 := "Гравці, що пропустили:"
	for _, player := range playerHasAnswered {
		txt1 += fmt.Sprintf("@%s", player.UserName)
	}
	for _, player := range playerSkipped {
		txt2 += fmt.Sprintf("@%s", player.UserName)
	}

	summary := txt1 + "\n\n" + txt2

    // Create message with summary
    // summary := "📊 *Підсумки завдання*\n\n"
    // summary += "Всього відповідей: 10\n"
    // summary += "Пропустили: 2\n\n"
    // summary += "⏰ Через 1 годину буде наступне завдання!"

    // Send summary messag to game chat
	utils.Logger.Infof("Sending task summary to game chat %d", game.GameChatID)
    _, err = bot.Send(&telebot.Chat{ID: game.GameChatID}, summary)
    return err
}

// Функция отправки следующей таски (ваша существующая логика)
// func sendNextTask(bot *telebot.Bot, db *sql.DB, gameID int64) error {
//     // Get game info
//     game, err := storage_db.GetGameById(int(gameID))
//     if err != nil {
//         return err
//     }

//     // Получаем следующую таску
//     nextTask, err := db.GetNextTask(gameID, game.CurrentTaskID)
//     if err != nil {
//         return err
//     }

//     // Обновляем current_task_id и time_update_task
//     timeUpdate := time.Now().Unix()
//     if err := storage_db.UpdateCurrentTaskID(int(gameID), nextTask.ID, timeUpdate); err != nil {
//         return err
//     }

//     // Очищаем уведомления
//    // db.ClearNotificationsForGame(gameID, game.GameChatID)

//     // Формируем сообщение
//     msg := nextTask.Title + "\n\n" + nextTask.Description

//     // Создаем кнопки
//     inlineKeys := &telebot.ReplyMarkup{}
//     answerBtn := inlineKeys.Data("Хочу відповісти", "answer_task", fmt.Sprintf("%d", nextTask.ID))
//     skipBtn := inlineKeys.Data("Пропустити", "skip_task", fmt.Sprintf("%d", nextTask.ID))
//     inlineKeys.Inline(inlineKeys.Row(answerBtn, skipBtn))

//     // Отправляем
//     _, err = bot.Send(&telebot.Chat{ID: game.GameChatID}, msg, inlineKeys, telebot.ModeMarkdown)
//     return err
// }
