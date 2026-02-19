package storage_db

import (
	"database/sql"
	"fmt"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
)

// AddPlayerToGame add player to game
func AddPlayerToGame(player *models.Player) error {
	query := `INSERT INTO players (user_id, username, name, game_id, status,
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
	query := `DELETE FROM players WHERE user_id = ? AND game_id = ?`
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

func GetAllPlayersByGameID(gameId int) ([]models.Player, error) {
	query := `SELECT user_id, username, name, game_id, status, skipped, role FROM players WHERE game_id = ?`
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

// UpdatePlayerStatus update player status in Db
func UpdatePlayerStatus(playerID int64, status string) error {
	query := `UPDATE players SET status = ? WHERE user_id = ?`
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
	query := `SELECT status FROM players WHERE user_id = ?`
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
	query := `SELECT role FROM players WHERE user_id = ? AND game_id = ?`
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
	query := `SELECT user_id, username, name, game_id, status, skipped, role FROM players WHERE game_id = ? AND role = 'admin'`
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

    var playerExists bool
    checkQuery := `SELECT EXISTS(SELECT 1 FROM players WHERE user_id = ?)`
    err := Db.QueryRow(checkQuery, playerResponse.PlayerID).Scan(&playerExists)
    if err != nil {
        utils.Logger.WithFields(logrus.Fields{
            "source": "Db: AddPlayerResponse",
            "player_id": playerResponse.PlayerID,
            "err": err,
        }).Error("Failed to check if player exists")
        return err
    }
    
    if !playerExists {
        utils.Logger.WithFields(logrus.Fields{
            "source": "Db: AddPlayerResponse",
            "player_id": playerResponse.PlayerID,
        }).Error("Player does not exist in players table!")
        return fmt.Errorf("player %d does not exist in players table", playerResponse.PlayerID)
    }
    
    utils.Logger.WithFields(logrus.Fields{
        "source": "Db: AddPlayerResponse",
        "player_id": playerResponse.PlayerID,
    }).Info("Player exists in database, proceeding to add response")
    
    var gameExists bool
    err = Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM games WHERE id = ?)`, playerResponse.GameID).Scan(&gameExists)
    if err != nil {
        utils.Logger.WithFields(logrus.Fields{
            "source": "Db: AddPlayerResponse",
            "game_id": playerResponse.GameID,
            "err": err,
        }).Error("Failed to check if game exists")
        return err
    }
    
    if !gameExists {
        utils.Logger.WithFields(logrus.Fields{
            "source": "Db: AddPlayerResponse",
            "game_id": playerResponse.GameID,
        }).Error("Game does not exist in games table!")
        return fmt.Errorf("game %d does not exist in games table", playerResponse.GameID)
    }
    
    utils.Logger.WithFields(logrus.Fields{
        "source": "Db: AddPlayerResponse",
        "game_id": playerResponse.GameID,
    }).Info("Game exists in database")
    
    utils.Logger.WithFields(logrus.Fields{
        "source": "Db: AddPlayerResponse",
        "task_id": playerResponse.TaskID,
    }).Info("Task exists in database")
	
	query := `INSERT INTO player_responses (player_id, user_name, game_id, task_id, has_answer, skipped, notification_sent) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = Db.Exec(query, playerResponse.PlayerID, playerResponse.UserName, playerResponse.GameID, playerResponse.TaskID, playerResponse.HasResponse, playerResponse.Skipped, playerResponse.NotificationSent)
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
func SkipPlayerResponse(playerID int64, userName string, gameID, taskID int) (*models.SkipStatus, error) {
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
			INSERT INTO player_responses (player_id, user_name, game_id, task_id, has_answer, skipped)
			VALUES (?, ?, ?, ?, 0, 1)
		`, playerID, userName, gameID, taskID)
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
		_, err := Db.Exec(`UPDATE player_responses SET skipped = 1 WHERE player_id = ? AND user_name = ? AND game_id = ? AND task_id = ?`, playerID, userName, gameID, taskID)
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

// Check user is in the game by ID
func IsUserInGame(playerID int64, gameID int) (bool, error) {
	query := `SELECT COUNT(*) FROM players WHERE user_id = ? AND game_id = ?`
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

func GetPlayersWithAnswer(gameID, taskID int64) ([]models.PlayerNot, error) {
    query := `
        SELECT player_id, user_name
        FROM player_responses
        WHERE game_id = ? AND task_id = ? AND has_answer = 1
    `
    
    rows, err := Db.Query(query, gameID, taskID)
    if err != nil {
        return nil, fmt.Errorf("ошибка выполнения запроса GetPlayersWithAnswer: %w", err)
    }
    defer rows.Close()
    
    var players []models.PlayerNot
    for rows.Next() {
        var player models.PlayerNot
        err := rows.Scan(
            &player.ID,
            &player.UserName,
        )
        if err != nil {
            return nil, fmt.Errorf("ошибка сканирования строки в GetPlayersWithAnswer: %w", err)
        }
        players = append(players, player)
    }
    
    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("ошибка итерации по строкам в GetPlayersWithAnswer: %w", err)
    }
    
    return players, nil
}

// GetPlayersWhoSkipped возвращает игроков, которые не дали ответ (skipped = true)
func GetPlayersWhoSkipped(gameID, taskID int64) ([]models.PlayerNot, error) {
	query := `
		SELECT player_id, user_name
        FROM player_responses
		WHERE game_id = ? AND task_id = ? AND skipped = true
	`
	
	rows, err := Db.Query(query, gameID, taskID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса GetPlayersWhoSkipped: %w", err)
	}
	defer rows.Close()
	
	var players []models.PlayerNot
	for rows.Next() {
		var player models.PlayerNot
		err := rows.Scan(
			&player.ID,
			&player.UserName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки в GetPlayersWhoSkipped: %w", err)
		}
		players = append(players, player)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам в GetPlayersWhoSkipped: %w", err)
	}
	
	return players, nil
}