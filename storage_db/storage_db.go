package storage_db

import (
	"database/sql"
	
	//"log"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var Db *sql.DB // Global variable for database connection

// InitDb initializate database SQLite with path DbPath
func InitDB(DbPath string) (*sql.DB, error) {
	var err error
	// Connect to database
	Db, err = sql.Open("sqlite3", DbPath)
	if err != nil {
		utils.Logger.Fatalf("Error connection database: %v", err)
		return nil, err
	}

	// Check connection
	if err := Db.Ping(); err != nil {
		utils.Logger.Fatalf("Error checking connect to database: %v", err)
		return nil, err
	}

	utils.Logger.Info("The database has been initialized successfully.")

	// Create tables
	if err := createTables(); err != nil {
		return nil, err
	}

	return Db, nil
}

// CloseDb close connect to database
func CloseDB(Db *sql.DB) {
	if Db != nil {
		if err := Db.Close(); err != nil {
			utils.Logger.Errorf("Error closing database connection: %v", err)
		} else {
			utils.Logger.Info("The database connection was closed successfully.")
		}
	}
}

// CreateGame добавляет новую игру в базу данных и возвращает ее ID
func CreateGame(gameName string, gameGroupChatId int64) (*models.Game, error) {
	game := &models.Game{
		Name: gameName,
		GameChatID: gameGroupChatId,
		//MsgJointID: 0,
		//InviteLink: "",
		CurrentTaskID: 0,
		TotalPlayers: 0,
		Status: models.StatusGameWaiting, // "waiting"
	}

	query := `INSERT INTO games (name, game_chat_id, status ) VALUES (?, ?, ?)`
	res, err := Db.Exec(query, game.Name, game.GameChatID , game.Status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: CreateGame",
			"game": game.Name,
			"game_id": game.ID,
			"error": err,
		}).Error("Failed to add game to Db")
		return nil, err
	}

	// Получаем ID созданной игры
	gameID, err := res.LastInsertId()
	if err != nil {
		utils.Logger.Errorf("Failed to get ID created game: %v", err)
		return nil, err
	}

	game.ID = int(gameID)
	
	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: CreateGame",
		"game": game.Name,
		"game_id": game.ID,
		"group_id": game.GameChatID,
	}).Info("Game has been created successfully")
	
	return game, nil
}

// UpdateGameStatus update status game in Db
func UpdateGameStatus(gameID int64, status string) error {
	query := `UPDATE games SET status = ? WHERE id = ?`
	_, err := Db.Exec(query, status, gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: UpdateGameStatus",
			"game_id": gameID,
			"error": err,
		}).Error("Failed to update game status")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: UpdateGameStatus",
		"game_id": gameID,
		"status": status,
	}).Info("Game status has been updated successfully")
	return nil
}

// Update MsgJoinID in game
// func UpdateMsgJoinID(gameID int, msgJoinID int) error {
// 	query := `UPDATE games SET msg_join_id = ? WHERE id = ?`
// 	_, err := Db.Exec(query, msgJoinID, gameID)
// 	if err != nil {
// 		utils.Logger.WithFields(logrus.Fields{
// 			"source": "Db: UpdateMsgJointID",
// 			"game_id": gameID,
// 			"msg_joint_id": msgJoinID,
// 			"error": err,
// 		}).Error("Failed to update MsgJointID in game")
// 		return err
// 	}

// 	utils.Logger.WithFields(logrus.Fields{
// 		"source": "Db: UpdateMsgJointID",
// 		"game_id": gameID,
// 		"msg_joint_id": msgJoinID,
// 	}).Info("MsgJointID has been updated successfully")
// 	return nil
// }

// Get MsgJointID by game ID
// func GetMsgJoinID(gameID int) (int, error) {
// 	query := `SELECT msg_join_id FROM games WHERE id = ?`
// 	row := Db.QueryRow(query, gameID)

// 	var msgJoinID int
// 	err := row.Scan(&msgJoinID)
// 	if err != nil {
// 		utils.Logger.WithFields(logrus.Fields{
// 			"source": "Db: GetMsgJointID",
// 			"game_id": gameID,
// 			"error": err,
// 		}).Error("Failed to get MsgJointID by game ID")
// 		return 0, err
// 	}

// 	utils.Logger.WithFields(logrus.Fields{
// 		"source": "Db: GetMsgJointID",
// 		"game_id": gameID,
// 		"msg_joint_id": msgJoinID,
// 	}).Info("MsgJointID has been retrieved successfully")

// 	return msgJoinID, nil
// }

// GetCurrentGameStatus gett current status game by ID
func GetCurrentGameStatus(gameID int) (string, error) {
	query := `SELECT status FROM games WHERE id = ?`
	row := Db.QueryRow(query, gameID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetCurrentGameStatus",
			"game_id": gameID,
			"error": err,
		}).Error("Failed to get current game status")
		return "", err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: GetCurrentGameStatus",
		"game_id": gameID,
		"status": status,
	}).Info("Current game status has been retrieved successfully")

	return status, nil
}

// CreateTask add a new task (question | answer) to Db
// func CreateTask(task models.Task) error {
// 	query := `INSERT INTO tasks (game_id, question, answer) VALUES (?, ?, ?)`
// 	_, err := Db.Exec(query, task.GameID, task.Question, task.Answer)
// 	if err != nil {
// 		utils.Logger.WithFields(logrus.Fields{
// 			"source": "Db: CreateTask",
// 			"game_id": task.GameID,
// 			"task_id": task.ID,
// 			"error": err,
// 		}).Error("Failed to add task to Db")
// 		return err
// 	}

// 	utils.Logger.Infof("Db: CreateTask: Task для GameID %d добавлен: '%s' -> '%s'", task.GameID, task.Question, task.Answer)
// 	return nil
// }

// GetGameById getting a game by ID
func GetGameById(gameID int) (*models.Game, error) {
	query := `SELECT id, name, current_task_id, total_players, status FROM games WHERE id = ?`
	row := Db.QueryRow(query, gameID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetGameById",
			"game_id": gameID,
			"error": err,
		}).Error("Failed to get game by ID")
		return nil, err
	}

	return game, nil
}

// GetGameByChatId getting a game by chat ID
func GetGameByChatId(chatID int64) (*models.Game, error) {
	query := `SELECT id, name, current_task_id, total_players, status FROM games WHERE game_chat_id = ?`
	row := Db.QueryRow(query, chatID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetGameByChatId",
			"chat_id": chatID,
			"error": err,
		}).Error("Failed to get game by chat ID")
		return nil, err
	}

	return game, nil
}

// AddPlayerToGame add player to game
func AddPlayerToGame(player *models.Player) error {
	query := `INSERT INTO players (id, username, name, game_id, status,
				skipped, role) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := Db.Exec(query, player.ID, player.UserName, player.Name, player.GameID, player.Status, player.Skipped, player.Role)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: AddPlayerToGame",
			"player": player.UserName,
			"player_id": player.ID,
			"game_id": player.GameID,
			"err": err,
		}).Error("Failed to add player to game")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: AddPlayerToGame",
		"player": player.UserName,
		"player_id": player.ID,
		"role": player.Role,
		"game_id": player.GameID,
	}).Info("Player has been added to game successfully")

	return nil
}

// DeletePlayerFromGame delete player from game
func DeletePlayerFromGame(playerID int64, gameID int) error {
	query := `DELETE FROM players WHERE id = ? AND game_id = ?`
	_, err := Db.Exec(query, playerID, gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: DeletePlayerFromGame",
			"player_id": playerID,
			"game_id": gameID,
			"error": err,
		}).Error("Failed to delete player from game")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: DeletePlayerFromGame",
		"player_id": playerID,
		"game_id": gameID,
	}).Info("Player has been deleted from game successfully")

	return nil
}

func GetPlayerCount(gameId int) (int, error) {
	query := `SELECT COUNT(*) FROM players WHERE game_id = ?`
	row := Db.QueryRow(query, gameId)

	var count int
	err := row.Scan(&count)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetPlayerCount",
			"game_id": gameId,
			"error": err,
		}).Error("Failed to get player count")
		return 0, err
	}

	return count, nil
}

func GetAllPlayersByGameID(gameId int) ([]models.Player, error) {
	query := `SELECT id, username, name, game_id, status, skipped, role FROM players WHERE game_id = ?`
	rows, err := Db.Query(query, gameId)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetAllPlayersByGameID",
			"game_id": gameId,
			"err": err,
		}).Error("Failed to get all players by game ID")
		return nil, err
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.Player
		err := rows.Scan(&player.ID, &player.UserName, &player.Name, &player.GameID, &player.Status, &player.Skipped, &player.Role)
		if err != nil {
			utils.Logger.Errorf("Error scanning player: %v", err)
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func GetCountTasksByGameID(gameId int) (int, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE game_id = ?`
	row := Db.QueryRow(query, gameId)

	var count int
	err := row.Scan(&count)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetCountTasksByGameID",
			"game_id": gameId,
			"error": err,
		}).Error("Error fetching task count")
		return 0, err
	}

	return count, nil
}

// UpdatePlayerStatus update player status in Db
func UpdatePlayerStatus(playerID int64, status string) error {
	query := `UPDATE players SET status = ? WHERE id = ?`
	_, err := Db.Exec(query, status, playerID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: UpdatePlayerStatus",
			"player_id": playerID,
			"status": status,
			"err": err,
		}).Error("Failed to update player status")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: UpdatePlayerStatus",
		"player_id": playerID,
		"status": status,
	}).Info("Player status updated successfully")

	return nil
}

// GetPlayerStatus get player status by ID
func GetStatusPlayer(playerID int64) (string, error) {
	query := `SELECT status FROM players WHERE id = ?`
	row := Db.QueryRow(query, playerID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetStatusPlayer",
			"player_id": playerID,
			"error": err,
		}).Error("Failed to get player status")

		return "", err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: GetStatusPlayer",
		"player_id": playerID,
		"status": status,
	}).Info("Player status get successfully")

	return status, nil
}

// Get player role by ID in game using player ID and game ID
func GetPlayerRoleByUserIDAndGameID(playerID int64, gameID int) (string, error) {
	query := `SELECT role FROM players WHERE id = ? AND game_id = ?`
	row := Db.QueryRow(query, playerID, gameID)

	var role string
	err := row.Scan(&role)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetPlayerRoleByUserIDAndGameID",
			"player_id": playerID,
			"game_id": gameID,
			"error": err,
		}).Error("Failed to get player role")
		return "", err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: GetPlayerGetPlayerRoleByUserIDAndGameIDRole",
		"player_id": playerID,
		"game_id": gameID,
		"role": role,
	}).Info("Player role retrieved successfully")

	return role, nil
}

// GetAdminPlayerByGameID get admin player by game ID
func GetAdminPlayerByGameID(gameID int) (*models.Player, error) {
	query := `SELECT id, username, name, game_id, status, skipped, role FROM players WHERE game_id = ? AND role = 'admin'`
	row := Db.QueryRow(query, gameID)

	player := &models.Player{}

	err := row.Scan(&player.ID, &player.UserName, &player.Name, &player.GameID, &player.Status, &player.Skipped, &player.Role)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetAdminPlayerByGameID",
			"game_id": gameID,
			"error": err,
		}).Error("Failed to get admin player by game ID")
		return nil, err
	}

	return player, nil
}

// AddPlayerAnswer add player answer to Db
func AddPlayerResponse(playerResponse *models.PlayerResponse) error {
	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: AddPlayerResponse",
		"player_id": playerResponse.PlayerID,
		"game_id": playerResponse.GameID,
		"task_id": playerResponse.TaskID,
	}).Info("Db: AddPlayerResponse - AddPlayerResponse was called")
	
	query := `INSERT INTO player_responses (player_id, game_id, task_id, has_answer, skipped, notification_sent) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := Db.Exec(query, playerResponse.PlayerID, playerResponse.GameID, playerResponse.TaskID, playerResponse.HasResponse, playerResponse.Skipped, playerResponse.NotificationSent)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: AddPlayerResponse",
			"player_id": playerResponse.PlayerID,
			"game_id": playerResponse.GameID,
			"task_id": playerResponse.TaskID,
			//"date_create": playerResponse.DateCreate,
			"notification_sent": playerResponse.NotificationSent,
			"err": err,
		}).Error("Failed to add player answer")
		
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "Db: AddPlayerResponse",
		"player_id": playerResponse.PlayerID,
		"game_id": playerResponse.GameID,
		"task_id": playerResponse.TaskID,
		//"date_create": playerResponse.DateCreate,
		"notification_sent": playerResponse.NotificationSent,
	}).Info("Player response added successfully")
	
	return nil
}

func CheckPlayerResponseStatus(playerID int64, gameID int, taskID int) (*models.AddResponseResult, error) {
	var hasAnswer, skipped bool

	err := Db.QueryRow(`
		SELECT has_answer, skipped FROM player_responses 
		WHERE player_id = ? AND game_id = ? AND task_id = ?
	`, playerID, gameID, taskID).Scan(&hasAnswer, &skipped)

	if err == sql.ErrNoRows {
		return &models.AddResponseResult{}, nil // Ничего ещё нет — всё ок
	}
	if err != nil {
		return nil, err
	}

	return &models.AddResponseResult{
		AlreadyAnswered: hasAnswer,
		AlreadySkipped:  skipped,
	}, nil
}

// SkipPlayerResponse handles skip logic for a specific task by a player
func SkipPlayerResponse(playerID int64, gameID int, taskID int) (*models.SkipStatus, error) {
	status := &models.SkipStatus{}

	// Check how many times player has already skipped tasks
	var skipCount int
	err := Db.QueryRow(`SELECT COUNT(*) FROM player_responses WHERE player_id = ? AND game_id = ? AND skipped = 1`, playerID, gameID  ).Scan(&skipCount)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: SkipPlayerResponse",
			"player_id": playerID,
			"game_id": gameID,
			"task_id": taskID,
			"error": err,
		}).Error("Error checking skip count")
		
		return nil, err
	}

	if skipCount >= 3 {
		status.SkipLimitReached = true
		status.RemainingSkips = 0
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: SkipPlayerResponse",
			"player_id": playerID,
			"game_id": gameID,
			"task_id": taskID,
		}).Info("Player has reached the skip limit")
		
		return status, nil
	}
	status.RemainingSkips = 3 - skipCount

	// Check if a response exists for this player and task
	var hasAnswer bool
	var skipped bool
	err = Db.QueryRow(`
		SELECT has_answer, skipped FROM player_responses 
		WHERE player_id = ? AND game_id = ? AND task_id = ?
	`, playerID, gameID, taskID).Scan(&hasAnswer, &skipped)

	switch {
	case err == sql.ErrNoRows:
		// No response yet — insert a new skipped record
		_, err := Db.Exec(`
			INSERT INTO player_responses (player_id, game_id, task_id, has_answer, skipped)
			VALUES (?, ?, ?, 0, 1)
		`, playerID, gameID, taskID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "Db: SkipPlayerResponse",
				"player_id": playerID,
				"game_id": gameID,
				"task_id": taskID,
				"error": err,
			}).Error("Error inserting skipped response")
			
			return nil, err
		}

		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: SkipPlayerResponse",
			"player_id": playerID,
			"game_id": gameID,
			"task_id": taskID,
			"skipped": true,
		}).Info("Player skipped task (new entry)")
		

	case err != nil:
		// Any other Db error
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: SkipPlayerResponse",
			"player_id": playerID,
			"game_id": gameID,
			"task_id": taskID,
			"error": err,
		}).Error("Error checking existing response")
		
		return nil, err

	default:
		// Response exists
		if hasAnswer {
			status.AlreadyAnswered = true
			utils.Logger.WithFields(logrus.Fields{
				"source": "Db: SkipPlayerResponse",
				"player_id": playerID,
				"game_id": gameID,
				"task_id": taskID,
				"skipped": skipped,
				"has_answer": hasAnswer,
			}).Warn("Player has already answered the task")
			
			return status, nil
		}
		if skipped {
			status.AlreadySkipped = true
			utils.Logger.WithFields(logrus.Fields{
				"source": "Db: SkipPlayerResponse",
				"player_id": playerID,
				"game_id": gameID,
				"task_id": taskID,
				"skipped": skipped,
				"has_answer": hasAnswer,
			}).Info("Player has already skipped the task")
			
			return status, nil
		}

		// Update existing entry to mark as skipped
		_, err := Db.Exec(`UPDATE player_responses SET skipped = 1 WHERE player_id = ? AND game_id = ? AND task_id = ?`, playerID, gameID, taskID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "Db: SkipPlayerResponse",
				"player_id": playerID,
				"game_id": gameID,
				"task_id": taskID,
				"error": err,
			}).Error("Error updating skipped flag")

			return nil, err
		}
		
	}

	return status, nil
}


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

// Check user is in the game by ID
func IsUserInGame(playerID int64, gameID int) (bool, error) {
	query := `SELECT COUNT(*) FROM players WHERE id = ? AND game_id = ?`
	row := Db.QueryRow(query, playerID, gameID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: IsUserInGame",
			"player_id": playerID,
			"game_id": gameID,
			"error": err,
		}).Error("Error checking user in game")
		
		return false, err
	}

	return count > 0, nil
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

// Get all games with status "playing"
func GetAllActiveGames() ([]models.Game, error) {
	query := `SELECT id, name, current_task_id, total_players, status FROM games WHERE status = ?`
	rows, err := Db.Query(query, models.StatusGamePlaying)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetAllActiveGames",
			"error": err,
		}).Error("Failed to get all active games")
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(&game.ID, &game.Name, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
		if err != nil {
			utils.Logger.Errorf("Error scanning game: %v", err)
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}

// reposistorys 