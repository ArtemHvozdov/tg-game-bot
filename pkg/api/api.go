package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/handlers"
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
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

var (
	summaryMsgs []models.SummaryMsg
    subSummaryMsgs1 []string
    noAnswersMsgs []string
	err error
)


func init() {
	summaryMsgs, err = utils.LoadSummaryMsgs("internal/data/messages/group/summary_msgs/summary_msgs.json")
	if err != nil {
		log.Printf("Error loading summary messages: %v", err)
	} else {
		log.Printf("Loaded %d summary messages", len(summaryMsgs))
	}

    noAnswersMsgs, err = utils.LoadArrayMsgs("internal/data/messages/group/summary_msgs/no_answers_msgs.json")
    if err != nil {
        log.Printf("Error loading no answers messages: %v", err)
    } else {
        log.Printf("Load %d no answers messages", len(noAnswersMsgs))
    }
}

// Run HTTP server
func StartHTTPServer(bot *telebot.Bot) {
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
       if err := handlers.SendNextTask(bot, req.GameID); err != nil {
            log.Printf("Failed to send next task: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

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
	}).Info("Info by game during API requese")

	chat := &telebot.Chat{ID: game.GameChatID}

	var summary string

	// if taskID == 3 {
	// 	err := handlers.SendSubtaskResultsToChat(bot, game.GameChatID)
	// 	if err != nil {
	// 		log.Printf("Failed to send subtask 3 results: %v", err)
	// 	}
	// }

	// if taskID == 5 {
	// 	message := `Зіроньки, тепер ви РЕАЛЬНО зіроньки! ✨ Не забувайте про хештег #Bestieverse якщо захочете прославитися в <a href="https://instagram.com/">Інстаграмах</a> і <a href="https://tiktok.com/">ТікТоках</a>.`

	// 	bot.Send(chat, message,  &telebot.SendOptions{
	// 		ParseMode:             telebot.ModeHTML,
	// 		DisableWebPagePreview: true,
	// 	})
	// }

	// if taskID == 10 {
	// 	handlers.CreateSubtask10Collage(bot, game.GameChatID)
	// }

	switch taskID {
    case 3:
        utils.Logger.Info("Task 3: Sending subtask results")

        hasAnswers, err := storage_db.HasResponses(gameID, taskID)
        if err != nil {
            utils.Logger.Errorf("Error checking no responses in the game %d for task %d: %v", gameID, taskID, err)
        }

        if !hasAnswers {
            message := utils.GetRandomMsg(noAnswersMsgs)
            
            _, err := bot.Send(chat, message)
            if err != nil {
                utils.Logger.Errorf("Failed to send no answers message for task 3: %v", err)
                return fmt.Errorf("failed to send no answers message for task 3: %w", err)
            }
            storage_db.MarkSummaryAsSent(gameID, taskID)
            return nil
        } else {
            if err := handlers.SendSubtaskResultsToChat(bot, game.GameChatID); err != nil {
                utils.Logger.Errorf("Failed to send subtask 3 results: %v", err)
                return fmt.Errorf("failed to send subtask 3 results: %w", err)
            }
            storage_db.MarkSummaryAsSent(gameID, taskID)
            return nil
        }
        
    case 5:
        utils.Logger.Info("Task 5: Sending congratulations message")

        hasAnswers, err := storage_db.HasResponses(gameID, taskID)
        if err != nil {
            utils.Logger.Errorf("Error checking no responses in the game %d for task %d: %v", gameID, taskID, err)
        }

        if !hasAnswers {
            message := utils.GetRandomMsg(noAnswersMsgs)
            
            _, err := bot.Send(chat, message)
            if err != nil {
                utils.Logger.Errorf("Failed to send no answers message for task 3: %v", err)
                return fmt.Errorf("failed to send no answers message for task 3: %w", err)
            }
            storage_db.MarkSummaryAsSent(gameID, taskID)
            return nil
        } else {
             message := `Зіроньки, тепер ви РЕАЛЬНО зіроньки! ✨ Не забувайте про хештег #Bestieverse якщо захочете прославитися в <a href="https://instagram.com/">Інстаграмах</a> і <a href="https://tiktok.com/">ТікТоках</a>.`
        
            _, err := bot.Send(chat, message, &telebot.SendOptions{
                ParseMode:             telebot.ModeHTML,
                DisableWebPagePreview: true,
            })
            if err != nil {
                utils.Logger.Errorf("Failed to send task 5 message: %v", err)
                return fmt.Errorf("failed to send task 5 message: %w", err)
            }
            storage_db.MarkSummaryAsSent(gameID, taskID)
            return nil
        }
        
    case 10:
        utils.Logger.Info("Task 10: Creating subtask collage")

        hasAnswers, err := storage_db.HasResponses(gameID, taskID)
        if err != nil {
            utils.Logger.Errorf("Error checking no responses in the game %d for task %d: %v", gameID, taskID, err)
        }

        if !hasAnswers {
            message := utils.GetRandomMsg(noAnswersMsgs)
            
            _, err := bot.Send(chat, message)
            if err != nil {
                utils.Logger.Errorf("Failed to send no answers message for task 3: %v", err)
                return fmt.Errorf("failed to send no answers message for task 3: %w", err)
            }
            storage_db.MarkSummaryAsSent(gameID, taskID)
            return nil
        } else {
            if err := handlers.CreateSubtask10Collage(bot, game.GameChatID); err != nil {
                utils.Logger.Errorf("Failed to create subtask 10 collage: %v", err)
                return fmt.Errorf("failed to create subtask 10 collage: %w", err)
            }
            storage_db.MarkSummaryAsSent(gameID, taskID)
            return nil
        }
    }

	for _, msg := range summaryMsgs {
		if msg.ID == int(taskID) {
            if msg.ID == 1 {
                utils.Logger.Info("Sending sub summary for task 1")

                hasAnswers, err := storage_db.HasResponses(gameID, taskID)
                if err != nil {
                    utils.Logger.Errorf("Error checking no responses in the game %d for task %d: %v", gameID, taskID, err)
                }

                if !hasAnswers {
                    message := utils.GetRandomMsg(noAnswersMsgs)
                    
                    _, err := bot.Send(chat, message)
                    if err != nil {
                        utils.Logger.Errorf("Failed to send no answers message for task 3: %v", err)
                        return fmt.Errorf("failed to send no answers message for task 3: %w", err)
                    }
                    storage_db.MarkSummaryAsSent(gameID, taskID)
                    return nil
                }

                summary = msg.Summary
                _, err = bot.Send(chat, summary)
                if err != nil {
                    utils.Logger.Errorf("Failed to send summary message to chat %d: %v", game.GameChatID, err)
                    return err
                }

                subSummaryMsgs1, err = utils.LoadArrayMsgs("internal/data/messages/group/summary_msgs/sub_summary_1.json")
                if err != nil {
                    utils.Logger.Errorf("Error loading sub summary messages for task 1: %v", err)
                } else {
                    utils.Logger.Infof("Loaded %d sub summary messages for task 1", len(subSummaryMsgs1))
                }

                players, err := storage_db.GetAllPlayersByGameID(int(gameID))
                if err != nil {
                    utils.Logger.Errorf("Error gettinf player from game %d: %v", gameID, err)
                }

                // for _, player := range players {
                //     for _, subMsg := range subSummaryMsgs1 {
                //         msg := fmt.Sprintf(subMsg, player.UserName)
                //         _, err = bot.Send(chat, msg)
                //         if err != nil {
                //             utils.Logger.Errorf("Failed to send sub summary message to chat %d: %v", game.GameChatID, err)
                //         }
                //         time.Sleep(5 * time.Second)
                //     }
                // }

                for i, player := range players {
                    // Проверяем, что у нас есть соответствующее сообщение
                    if i < len(subSummaryMsgs1) {
                        msg := fmt.Sprintf(subSummaryMsgs1[i], player.UserName)
                        _, err = bot.Send(chat, msg)
                        if err != nil {
                            utils.Logger.Errorf("Failed to send sub summary message to chat %d: %v", game.GameChatID, err)
                        }
                        time.Sleep(5 * time.Second)
                    } else {
                        utils.Logger.Warnf("No more sub summary messages available for player %s (index %d)", player.UserName, i)
                    }
                }

                return nil
            }

            hasAnswers, err := storage_db.HasResponses(gameID, taskID)
            if err != nil {
                utils.Logger.Errorf("Error checking no responses in the game %d for task %d: %v", gameID, taskID, err)
            }

            if !hasAnswers {
                message := utils.GetRandomMsg(noAnswersMsgs)
                
                _, err := bot.Send(chat, message)
                if err != nil {
                    utils.Logger.Errorf("Failed to send no answers message for task 3: %v", err)
                    return fmt.Errorf("failed to send no answers message for task 3: %w", err)
                }
                storage_db.MarkSummaryAsSent(gameID, taskID)
                return nil
            }

			summary = msg.Summary
			// Send summary messag to game chat
			utils.Logger.Infof("Sending task summary to game chat %d", game.GameChatID)
			_, err = bot.Send(chat, summary)
			break
		}
	}

    // // Send summary messag to game chat
	// utils.Logger.Infof("Sending task summary to game chat %d", game.GameChatID)
    // _, err = bot.Send(&telebot.Chat{ID: game.GameChatID}, summary)
	storage_db.MarkSummaryAsSent(gameID, taskID)
    return err
}

