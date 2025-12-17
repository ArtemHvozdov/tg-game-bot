package storage_db

import (
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
)

// Update current task ID in game
func UpdateCurrentTaskID(gameID int, taskID int, timeUpdate int64) error {
	query := `UPDATE games SET current_task_id = ?, time_update_task = ? WHERE id = ?`
	_, err := Db.Exec(query, taskID, timeUpdate, gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: UpdateCurrentTaskID",
			"game_id": gameID,
			"task_id": taskID,
			"time_update_task": timeUpdate,
			"error": err,
		}).Error("Error updating current task ID for game")
	
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: UpdateCurrentTaskID",
		"game_id": gameID,
		"task_id": taskID,
		"time_update_task": timeUpdate,
	}).Info("Current task ID for game updated successfully")
	
	return nil
}

// Add answer to subtask
func AddSubtaskAnswer(answer *models.SubtaskAnswer) error {
	query := `INSERT INTO subtask_answers (game_id, task_id, question_index, answerer_user_id, selected_user_id, selected_username) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := Db.Exec(query, answer.GameID, answer.TaskID, answer.QuestionIndex, answer.AnswererUserID, answer.SelectedUserID, answer.SelectedUsername)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: AddSubtaskAnswer",
			"game_id": answer.GameID,
			"task_id": answer.TaskID,
			"error": err,
		}).Error("Failed to add subtask answer")
		
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: AddSubtaskAnswer",
		"game_id": answer.GameID,
		"task_id": answer.TaskID,
	}).Info("Subtask answer added successfully")
	
	return nil
}

// Get subtask results from database
func GetSubtaskResults(gameID, taskID int) (map[int]map[string]int, error) {
    query := `
        SELECT question_index, selected_username, COUNT(*) as vote_count
        FROM subtask_answers 
        WHERE game_id = ? AND task_id = ?
        GROUP BY question_index, selected_username
        ORDER BY question_index, vote_count DESC
    `
    
    rows, err := Db.Query(query, gameID, taskID)
    if err != nil {
        utils.Logger.WithFields(logrus.Fields{
            "source":  "Db: GetSubtaskResults",
            "game_id": gameID,
            "task_id": taskID,
            "error":   err,
        }).Error("Failed to query subtask results")
        return nil, err
    }
    defer rows.Close()
    
    // results[questionIndex][username] = voteCount
    results := make(map[int]map[string]int)
    
    for rows.Next() {
        var questionIndex int
        var selectedUsername string
        var voteCount int
        
        err := rows.Scan(&questionIndex, &selectedUsername, &voteCount)
        if err != nil {
            utils.Logger.WithFields(logrus.Fields{
                "source":  "Db: GetSubtaskResults",
                "game_id": gameID,
                "task_id": taskID,
                "error":   err,
            }).Error("Error scanning subtask result row")
            continue
        }
        
        if results[questionIndex] == nil {
            results[questionIndex] = make(map[string]int)
        }
        
        results[questionIndex][selectedUsername] = voteCount
    }
    
    utils.Logger.WithFields(logrus.Fields{
        "source":           "Db: GetSubtaskResults",
        "game_id":          gameID,
        "task_id":          taskID,
        "questions_found":  len(results),
    }).Info("Subtask results retrieved successfully")
    
    return results, rows.Err()
}

// AddSubtask10Answer adds a new subtask 10 answer to the database
func AddSubtask10Answer(answer *models.Subtask10Answer) error {
	query := `INSERT INTO subtask_10_answers (game_id, task_id, question_index, question_id, answerer_user_id, selected_option)
		VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := Db.Exec(query, answer.GameID, answer.TaskID, answer.QuestionIndex, answer.QuestionID, answer.AnswererUserID, answer.SelectedOption)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source":           "Db: AddSubtask10Answer",
			"game_id":          answer.GameID,
			"task_id":          answer.TaskID,
			"question_index":   answer.QuestionIndex,
			"question_id":      answer.QuestionID,
			"answerer_user_id": answer.AnswererUserID,
			"selected_option":  answer.SelectedOption,
			"error":            err,
		}).Error("Failed to add subtask 10 answer")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source":           "Db: AddSubtask10Answer",
		"game_id":          answer.GameID,
		"task_id":          answer.TaskID,
		"question_index":   answer.QuestionIndex,
		"question_id":      answer.QuestionID,
		"answerer_user_id": answer.AnswererUserID,
		"selected_option":  answer.SelectedOption,
	}).Info("Subtask 10 answer added successfully")
	
	return nil
}