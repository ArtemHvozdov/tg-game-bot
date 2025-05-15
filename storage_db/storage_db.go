package storage_db

import (
	"database/sql"
	//"log"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB // Global variable for database connection

// InitDB initializate database SQLite with path dbPath
func InitDB(dbPath string) (*sql.DB, error) {
	var err error
	// Connect to database
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		utils.Logger.Fatalf("Error connection database: %v", err)
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		utils.Logger.Fatalf("Error checking connect to database: %v", err)
		return nil, err
	}

	utils.Logger.Info("The database has been initialized successfully.")

	// Create tables
	if err := createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

// CloseDB close connect to database
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			utils.Logger.Errorf("Error closing database connection: %v", err)
		} else {
			utils.Logger.Info("The database connection was closed successfully.")
		}
	}
}

// createTables creates necessary tables in the database
func createTables() error {
	queries := []struct {
		tableName string
		query     string
	}{
		{
			"players",
			`CREATE TABLE IF NOT EXISTS players (
				id INTEGER,
				username TEXT NOT NULL,
				name TEXT NOT NULL,
				game_id INTEGER,
				status TEXT,
				skipped INT,
				role TEXT NOT NULL
			)`,
		},
		{
			"games",
			`CREATE TABLE IF NOT EXISTS games (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				game_chat_id INTEGER,
				
				current_task_id INTEGER NOT NULL DEFAULT 0,
				total_players INTEGER NOT NULL DEFAULT 0,
				status TEXT CHECK(status IN ('waiting', 'playing', 'finished')) NOT NULL
			)`,
		},
		{
			"game_players",
			`CREATE TABLE IF NOT EXISTS game_players (
				game_id INTEGER,
				player_id INTEGER,
				status TEXT CHECK(status IN ('joined', 'playing', 'finished')) NOT NULL,
				PRIMARY KEY (game_id, player_id),
				FOREIGN KEY (game_id) REFERENCES games(id),
				FOREIGN KEY (player_id) REFERENCES players(id)
			)`,
		},
		{
			"tasks",
			`CREATE TABLE IF NOT EXISTS tasks (
				id INTEGER PRIMARY KEY,
				game_id INTEGER,
				question TEXT NOT NULL,
				answer TEXT NOT NULL,
				FOREIGN KEY (game_id) REFERENCES games(id)
			)`,
		},
		{
			"player_responses",
			`CREATE TABLE IF NOT EXISTS player_responses (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				player_id INTEGER,
				game_id INTEGER,
				task_id INTEGER,
				has_answer BOOLEAN,
				skipped BOOLEAN DEFAULT FALSE,
				FOREIGN KEY (player_id) REFERENCES players(id),
				FOREIGN KEY (game_id) REFERENCES games(id),
				FOREIGN KEY (task_id) REFERENCES tasks(id)
			)`,
		},
		{
			"game_state",
			`CREATE TABLE IF NOT EXISTS game_state (
				game_id INTEGER PRIMARY KEY UNIQUE ,
				status TEXT NOT NULL,
				FOREIGN KEY (game_id) REFERENCES games(id)
			)`,
		},
	}

	for _, q := range queries {
		if _, err := db.Exec(q.query); err != nil {
			return err
		}
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: createTables",
			"table": q.tableName,
		}).Info("Table has been created or already exists.")
	}

	return nil
}

// CreateGame добавляет новую игру в базу данных и возвращает ее ID
func CreateGame(gameName string, gameGroupChatId int64) (*models.Game, error) {
	game := &models.Game{
		Name: gameName,
		GameChatID: gameGroupChatId,
		//InviteLink: "",
		CurrentTaskID: 0,
		TotalPlayers: 0,
		Status: "waiting",
	}

	query := `INSERT INTO games (name, game_chat_id, status ) VALUES (?, ?, ?)`
	res, err := db.Exec(query, game.Name, game.GameChatID , game.Status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: CreateGame",
			"game": game.Name,
			"game_id": game.ID,
			"error": err,
		}).Error("Failed to add game to DB")
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
		"source": "DB: CreateGame",
		"game": game.Name,
		"game_id": game.ID,
		"group_id": game.GameChatID,
	}).Info("Game has been created successfully")
	
	return game, nil
}

// UpdateGameStatus update status game in DB
func UpdateGameStatus(gameID int64, status string) error {
	query := `UPDATE games SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: UpdateGameStatus",
			"game_id": gameID,
			"error": err,
		}).Error("Failed to update game status")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: UpdateGameStatus",
		"game_id": gameID,
		"status": status,
	}).Info("Game status has been updated successfully")
	return nil
}

// GetCurrentGameStatus gett current status game by ID
func GetCurrentGameStatus(gameID int) (string, error) {
	query := `SELECT status FROM games WHERE id = ?`
	row := db.QueryRow(query, gameID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetCurrentGameStatus",
			"game_id": gameID,
			"error": err,
		}).Error("Failed to get current game status")
		return "", err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: GetCurrentGameStatus",
		"game_id": gameID,
		"status": status,
	}).Info("Current game status has been retrieved successfully")

	return status, nil
}

// CreateTask add a new task (question | answer) to DB
// func CreateTask(task models.Task) error {
// 	query := `INSERT INTO tasks (game_id, question, answer) VALUES (?, ?, ?)`
// 	_, err := db.Exec(query, task.GameID, task.Question, task.Answer)
// 	if err != nil {
// 		utils.Logger.WithFields(logrus.Fields{
// 			"source": "DB: CreateTask",
// 			"game_id": task.GameID,
// 			"task_id": task.ID,
// 			"error": err,
// 		}).Error("Failed to add task to DB")
// 		return err
// 	}

// 	utils.Logger.Infof("DB: CreateTask: Task для GameID %d добавлен: '%s' -> '%s'", task.GameID, task.Question, task.Answer)
// 	return nil
// }

// GetGameById getting a game by ID
func GetGameById(gameID int) (*models.Game, error) {
	query := `SELECT id, name, current_task_id, total_players, status FROM games WHERE id = ?`
	row := db.QueryRow(query, gameID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetGameById",
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
	row := db.QueryRow(query, chatID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetGameByChatId",
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
	_, err := db.Exec(query, player.ID, player.UserName, player.Name, player.GameID, player.Status, player.Skipped, player.Role)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: AddPlayerToGame",
			"player": player.UserName,
			"player_id": player.ID,
			"game_id": player.GameID,
			"err": err,
		}).Error("Failed to add player to game")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: AddPlayerToGame",
		"player": player.UserName,
		"player_id": player.ID,
		"role": player.Role,
		"game_id": player.GameID,
	}).Info("Player has been added to game successfully")

	return nil
}

func GetPlayerCount(gameId int) (int, error) {
	query := `SELECT COUNT(*) FROM players WHERE game_id = ?`
	row := db.QueryRow(query, gameId)

	var count int
	err := row.Scan(&count)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetPlayerCount",
			"game_id": gameId,
			"error": err,
		}).Error("Failed to get player count")
		return 0, err
	}

	return count, nil
}

func GetAllPlayersByGameID(gameId int) ([]models.Player, error) {
	query := `SELECT id, username, name, game_id, passes, role FROM players WHERE game_id = ?`
	rows, err := db.Query(query, gameId)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetAllPlayersByGameID",
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
	row := db.QueryRow(query, gameId)

	var count int
	err := row.Scan(&count)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetCountTasksByGameID",
			"game_id": gameId,
			"error": err,
		}).Error("Error fetching task count")
		return 0, err
	}

	return count, nil
}

// UpdatePlayerStatus update player status in DB
func UpdatePlayerStatus(playerID int64, status string) error {
	query := `UPDATE players SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, playerID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: UpdatePlayerStatus",
			"player_id": playerID,
			"status": status,
			"err": err,
		}).Error("Failed to update player status")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: UpdatePlayerStatus",
		"player_id": playerID,
		"status": status,
	}).Info("Player status updated successfully")

	return nil
}

// GetPlayerStatus get player status by ID
func GetStatusPlayer(playerID int64) (string, error) {
	query := `SELECT status FROM players WHERE id = ?`
	row := db.QueryRow(query, playerID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: GetStatusPlayer",
			"player_id": playerID,
			"error": err,
		}).Error("Failed to get player status")

		return "", err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: GetStatusPlayer",
		"player_id": playerID,
		"status": status,
	}).Info("Player status get successfully")

	return status, nil
}

// AddPlayerAnswer add player answer to DB
func AddPlayerResponse(playerResponse *models.PlayerResponse) error {
	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: AddPlayerResponse",
		"player_id": playerResponse.PlayerID,
		"game_id": playerResponse.GameID,
		"task_id": playerResponse.TaskID,
	}).Info("DB: AddPlayerResponse - AddPlayerResponse was called")
	
	query := `INSERT INTO player_responses (player_id, game_id, task_id, has_answer, skipped) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, playerResponse.PlayerID, playerResponse.GameID, playerResponse.TaskID, playerResponse.HasResponse, playerResponse.Skipped)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: AddPlayerResponse",
			"player_id": playerResponse.PlayerID,
			"game_id": playerResponse.GameID,
			"task_id": playerResponse.TaskID,
			"err": err,
		}).Error("Failed to add player answer")
		
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: AddPlayerResponse",
		"player_id": playerResponse.PlayerID,
		"game_id": playerResponse.GameID,
		"task_id": playerResponse.TaskID,
	}).Info("Player response added successfully")
	
	return nil
}

func CheckPlayerResponseStatus(playerID int64, gameID int, taskID int) (*models.AddResponseResult, error) {
	var hasAnswer, skipped bool

	err := db.QueryRow(`
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
	err := db.QueryRow(`
		SELECT COUNT(*) FROM player_responses 
		WHERE player_id = ? AND skipped = 1
	`, playerID).Scan(&skipCount)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: SkipPlayerResponse",
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
			"source": "DB: SkipPlayerResponse",
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
	err = db.QueryRow(`
		SELECT has_answer, skipped FROM player_responses 
		WHERE player_id = ? AND game_id = ? AND task_id = ?
	`, playerID, gameID, taskID).Scan(&hasAnswer, &skipped)

	switch {
	case err == sql.ErrNoRows:
		// No response yet — insert a new skipped record
		_, err := db.Exec(`
			INSERT INTO player_responses (player_id, game_id, task_id, has_answer, skipped)
			VALUES (?, ?, ?, 0, 1)
		`, playerID, gameID, taskID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "DB: SkipPlayerResponse",
				"player_id": playerID,
				"game_id": gameID,
				"task_id": taskID,
				"error": err,
			}).Error("Error inserting skipped response")
			
			return nil, err
		}

		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: SkipPlayerResponse",
			"player_id": playerID,
			"game_id": gameID,
			"task_id": taskID,
			"skipped": true,
		}).Info("Player skipped task (new entry)")
		

	case err != nil:
		// Any other DB error
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: SkipPlayerResponse",
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
				"source": "DB: SkipPlayerResponse",
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
				"source": "DB: SkipPlayerResponse",
				"player_id": playerID,
				"game_id": gameID,
				"task_id": taskID,
				"skipped": skipped,
				"has_answer": hasAnswer,
			}).Info("Player has already skipped the task")
			
			return status, nil
		}

		// Update existing entry to mark as skipped
		_, err := db.Exec(`
			UPDATE player_responses SET skipped = 1 
			WHERE player_id = ? AND game_id = ? AND task_id = ?
		`, playerID, gameID, taskID)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source": "DB: SkipPlayerResponse",
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
func UpdateCurrentTaskID(gameID int, taskID int) error {
	query := `UPDATE games SET current_task_id = ? WHERE id = ?`
	_, err := db.Exec(query, taskID, gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: UpdateCurrentTaskID",
			"game_id": gameID,
			"task_id": taskID,
			"error": err,
		}).Error("Error updating current task ID for game")
	
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source": "DB: UpdateCurrentTaskID",
		"game_id": gameID,
		"task_id": taskID,
	}).Info("Current task ID for game updated successfully")
	
	return nil
}

// Check user is in the game by ID
func IsUserInGame(playerID int64, gameID int) (bool, error) {
	query := `SELECT COUNT(*) FROM players WHERE id = ? AND game_id = ?`
	row := db.QueryRow(query, playerID, gameID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "DB: IsUserInGame",
			"player_id": playerID,
			"game_id": gameID,
			"error": err,
		}).Error("Error checking user in game")
		
		return false, err
	}

	return count > 0, nil
}