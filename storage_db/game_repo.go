package storage_db

import (
	"database/sql"
	"fmt"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
)

// Add new game to DB and return ID of game
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

// GetGameById getting a game by ID
func GetGameById(gameID int) (*models.Game, error) {
	query := `SELECT id, name, game_chat_id, current_task_id, total_players, status FROM games WHERE id = ?`
	row := Db.QueryRow(query, gameID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.GameChatID, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
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

// ClearNotificationsForGame removes all notification records for a specific game and chat
func ClearNotificationsForGame(gameID, gameChatID int64) error {
    query := `
        DELETE FROM notifications 
        WHERE game_id = ? 
            AND game_chat_id = ?
    `
    
	result, err := Db.Exec(query, gameID, gameChatID)
	if err != nil {
		utils.Logger.Errorf("failed to clear notifications: %v", err)
		return err
	}
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        utils.Logger.Errorf("failed to get rows affected: %v", err)
		return  err
    }
    
    utils.Logger.Infof("Cleared %d notification records for game %d, chat %d", rowsAffected, gameID, gameChatID)
    
    return nil
}

func MarkSummaryAsSent(gameID, taskID int64) error {
    query := `
        INSERT INTO summary_notifications (game_id, task_id, summary_sent)
        VALUES (?, ?, 1)
    `
    
    _, err := Db.Exec(query, gameID, taskID)
    if err != nil {
        return fmt.Errorf("failed to mark summary as sent: %w", err)
    }
    
    utils.Logger.WithFields(logrus.Fields{
        "source":  "Db: MarkSummaryAsSent",
        "game_id": gameID,
        "task_id": taskID,
    }).Info("Summary marked as sent successfully")
    
    return nil
}

func HasSummaryBeenSent(gameID, taskID int64) (bool, error) {
    query := `
        SELECT EXISTS(
            SELECT 1 
            FROM summary_notifications 
            WHERE game_id = ? 
                AND task_id = ?
                AND summary_sent = 1
        )
    `
    
    var exists bool
    err := Db.QueryRow(query, gameID, taskID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check if summary was sent: %w", err)
    }
    
    return exists, nil
}

func ClearSummaryNotifications(gameID int64) error {
    query := `
        DELETE FROM summary_notifications 
        WHERE game_id = ?
    `
    
    result, err := Db.Exec(query, gameID)
    if err != nil {
        return fmt.Errorf("failed to clear summary notifications: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    utils.Logger.WithFields(logrus.Fields{
        "source":        "Db: ClearSummaryNotifications",
        "game_id":       gameID,
        "rows_affected": rowsAffected,
    }).Info("Summary notifications cleared successfully")
    
    return nil
}

func HasResponses(gameID, taskID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM player_responses 
			WHERE game_id = ? AND task_id = ?
			LIMIT 1
		);
	`

	var exists bool
	err := Db.QueryRow(query, gameID, taskID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		utils.Logger.WithFields(logrus.Fields{
			"source":        "Db: HasResponses",
			"game_id":       gameID,
			"task_id": 		 taskID,
		}).Errorf("error checking player_responses existence: %v", err)

		return false, err
	}

	return exists, nil
}