package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func ShowSubtaskResults(bot *telebot.Bot, gameID, taskID int) error {
    // Load subtask questions
    subtasks, err := utils.LoadSubTasks("internal/data/tasks/subtasks/subtask_3.json")
    if err != nil {
        return fmt.Errorf("failed to load subtasks: %w", err)
    }
    
    if len(subtasks) == 0 {
        return fmt.Errorf("no subtasks found")
    }
    
    // Get results from database
    results, err := storage_db.GetSubtaskResults(gameID, taskID)
    if err != nil {
        return fmt.Errorf("failed to get results from database: %w", err)
    }
    
    // Build message
    var messageBuilder strings.Builder
    messageBuilder.WriteString("Круасанчики, ось ваш рейтинг відповідей. Усі згодні з результатами? 😃\n\n")
    
    // Process each question
    for questionIndex, question := range subtasks {
        messageBuilder.WriteString(fmt.Sprintf("%s ", question))
        
        // Get winners for this question
        winners := getWinnersForQuestion(results, questionIndex)
        
        if len(winners) == 0 {
            messageBuilder.WriteString("(немає відповідей)")
        } else {
            for _, winner := range winners {
                messageBuilder.WriteString(fmt.Sprintf("@%s ", winner))
            }
        }
        
        messageBuilder.WriteString("\n")
    }
    
    return nil
}

// Get winners (users with most votes) for specific question
func getWinnersForQuestion(results map[int]map[string]int, questionIndex int) []string {
    questionResults := results[questionIndex]
    if len(questionResults) == 0 {
        return []string{}
    }
    
    // Find maximum vote count
    maxVotes := 0
    for _, voteCount := range questionResults {
        if voteCount > maxVotes {
            maxVotes = voteCount
        }
    }
    
    // Get all users with maximum votes
    var winners []string
    for username, voteCount := range questionResults {
        if voteCount == maxVotes {
            winners = append(winners, username)
        }
    }
    
    // Sort winners alphabetically for consistent output
    sort.Strings(winners)
    
    return winners
}

// Handler function to send results to chat
// Variant called as handler in the chat with context
// func SendSubtaskResultsToChat(bot *telebot.Bot, chatID int64) func(c telebot.Context) error {
//     return func(c telebot.Context) error {
//         // user := c.Sender()
//         // chat := c.Chat()
        
//         // Get game info
//         game, err := storage_db.GetGameByChatId(chatID)
//         if err != nil {
//             utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chatID, err)
//             return c.Send("Помилка отримання гри")
//         }
        
//         // Load subtask questions
//         subtasks, err := utils.LoadSubTasks("internal/data/tasks/subtasks/subtask_3.json")
//         if err != nil {
//             utils.Logger.Errorf("Failed to load subtasks: %v", err)
//             return c.Send("Помилка завантаження підзавдань")
//         }
        
//         if len(subtasks) == 0 {
//             return c.Send("Підзавдання не знайдені")
//         }
        
//         // Get results from database
//         taskID := 3 // for subtask 3
//         results, err := storage_db.GetSubtaskResults(game.ID, taskID)
//         if err != nil {
//             utils.Logger.Errorf("Failed to get subtask results: %v", err)
//             return c.Send("Помилка отримання результатів")
//         }
        
//         if len(results) == 0 {
//             return c.Send("Результати підзавдань не знайдені. Можливо, ніхто ще не відповідав на питання.")
//         }
        
//         // Build message
//         var messageBuilder strings.Builder
//         messageBuilder.WriteString("Круасанчики, ось ваш рейтинг відповідей. Усі згодні з результатами? 😃\n\n")
        
//         // Process each question
//         for questionIndex, question := range subtasks {
//             messageBuilder.WriteString(fmt.Sprintf("%s ", question))
            
//             // Get winners for this question
//             winners := getWinnersForQuestion(results, questionIndex)
            
//             if len(winners) == 0 {
//                 messageBuilder.WriteString("(немає відповідей)")
//             } else {
//                 for _, winner := range winners {
//                     messageBuilder.WriteString(fmt.Sprintf("@%s \n", winner))
//                 }
//             }
            
//             messageBuilder.WriteString("\n")
//         }
        
//         // utils.Logger.WithFields(logrus.Fields{
//         //     "source":    "SendSubtaskResultsToChat",
//         //     "user":      user.Username,
//         //     "game_id":   game.ID,
//         //     "task_id":   taskID,
//         //     "questions": len(subtasks),
//         // }).Info("Sending subtask results")
        
//         //return c.Send(messageBuilder.String())
//         _, err = bot.Send(&telebot.Chat{ID: chatID}, messageBuilder.String())
//         return err
//     }
// }


// Variant called as external function from another packages without context
func SendSubtaskResultsToChat(bot *telebot.Bot, chatID int64) error {
    // Убираем обёртку func(c telebot.Context)
    
    // Get game info
    game, err := storage_db.GetGameByChatId(chatID)
        if err != nil {
            utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chatID, err)
            _, err = bot.Send(&telebot.Chat{ID: chatID}, "Помилка отримання гри")
            return err
        }
    
    // Load subtask questions
    subtasks, err := utils.LoadSubTasks("internal/data/tasks/subtasks/subtask_3.json")
    if err != nil {
        utils.Logger.Errorf("Failed to load subtasks: %v", err)
        return fmt.Errorf("failed to load subtasks: %w", err)
    }
    
    if len(subtasks) == 0 {
        return fmt.Errorf("no subtasks found")
    }
    
    // Get results from database
    taskID := 3 // for subtask 3
    results, err := storage_db.GetSubtaskResults(game.ID, taskID)
    if err != nil {
        utils.Logger.Errorf("Failed to get subtask results: %v", err)
        return fmt.Errorf("failed to get subtask results: %w", err)
    }
    
    if len(results) == 0 {
        utils.Logger.Warn("No results found for subtask 3")
        // Всё равно отправляем сообщение
        _, err = bot.Send(&telebot.Chat{ID: chatID}, 
            "Результати підзавдань не знайдені. Можливо, ніхто ще не відповідав на питання.")
        return err
    }
    
    // Build message
    var messageBuilder strings.Builder
    messageBuilder.WriteString("Круасанчики, ось ваш рейтинг відповідей. Усі згодні з результатами? 😃\n\n")
    
    // Process each question
    for questionIndex, question := range subtasks {
        messageBuilder.WriteString(fmt.Sprintf("%s ", question))
        
        // Get winners for this question
        winners := getWinnersForQuestion(results, questionIndex)
        
        if len(winners) == 0 {
            messageBuilder.WriteString("(немає відповідей)")
        } else {
            for _, winner := range winners {
                messageBuilder.WriteString(fmt.Sprintf("@%s\n", winner))
            }
        }
        messageBuilder.WriteString("\n")
    }
    
    utils.Logger.WithFields(logrus.Fields{
        "source":    "SendSubtaskResultsToChat",
        "game_id":   game.ID,
        "task_id":   taskID,
        "chat_id":   chatID,
        "questions": len(subtasks),
    }).Info("Sending subtask results to chat")
    
    // Send message
    _, err = bot.Send(&telebot.Chat{ID: chatID}, messageBuilder.String())
    if err != nil {
        utils.Logger.Errorf("Failed to send subtask results: %v", err)
        return fmt.Errorf("failed to send message: %w", err)
    }
    
    utils.Logger.Info("Subtask results sent successfully")
    return nil
}