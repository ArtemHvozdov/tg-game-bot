package storage_db

import (
	"database/sql"
	"log"
	
	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB // Global variable for database connection

// InitDB initializate database SQLite with path dbPath
func InitDB(dbPath string) (*sql.DB, error) {
	var err error
	// Connect to database
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error connection database: %v", err)
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error checking connect to database: %v", err)
		return nil, err
	}

	log.Println("The database has been initialized successfully.")

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
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("The database connection was closed successfully.")
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
				id INTEGER PRIMARY KEY,
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
				invite_link TEXT NOT NULL UNIQUE,
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
		log.Printf("Storage_db logs: Table '%s' has been created or already exists.", q.tableName)
	}

	return nil
}

// CreateGame добавляет новую игру в базу данных и возвращает ее ID
func CreateGame(gameName, inviteChatLink string, gameGroupChatId int64) (*models.Game, error) {
	game := &models.Game{
		Name: gameName,
		GameChatID: gameGroupChatId,
		InviteLink: utils.GenerateInviteLink(1),
		CurrentTaskID: 0,
		TotalPlayers: 0,
		Status: "waiting",
	}

	query := `INSERT INTO games (name, game_chat_id ,invite_link, status ) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, game.Name, game.GameChatID ,game.InviteLink, game.Status)
	if err != nil {
		log.Println("Ошибка при добавлении игры в БД:", err)
		return nil, err
	}

	// Получаем ID созданной игры
	gameID, err := res.LastInsertId()
	if err != nil {
		log.Println("Ошибка получения ID созданной игры:", err)
		return nil, err
	}

	game.ID = int(gameID)
	log.Printf("DB-Create-game-log: Game '%s' создана с ID %d и ID группового чата%d ", game.Name, game.ID, game.GameChatID)

	return game, nil

}

// UpdateGameStatus update status game in DB
func UpdateGameStatus(gameID int64, status string) error {
	query := `UPDATE games SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, gameID)
	if err != nil {
		log.Printf("Error updating game status for game ID %d: %v", gameID, err)
		return err
	}

	log.Printf("DB logs: (UpdateGameStatus) Game status for game ID %d updated to '%s'", gameID, status)
	return nil
}

// GetCurrentGameStatus gett current status game by ID
func GetCurrentGameStatus(gameID int) (string, error) {
	query := `SELECT status FROM games WHERE id = ?`
	row := db.QueryRow(query, gameID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		log.Printf("Error fetching current game status for game ID %d: %v", gameID, err)
		return "", err
	}

	log.Printf("Current game status for game ID %d is '%s'", gameID, status)
	return status, nil
}

// CreateTask add a new task (question | answer) to DB
func CreateTask(task models.Task) error {
	query := `INSERT INTO tasks (game_id, question, answer) VALUES (?, ?, ?)`
	_, err := db.Exec(query, task.GameID, task.Question, task.Answer)
	if err != nil {
		log.Println("Ошибка при добавлении задания в БД:", err)
		return err
	}

	log.Printf("Task для GameID %d добавлен: '%s' -> '%s'", task.GameID, task.Question, task.Answer)
	return nil
}

// GetGameById getting a game by ID
func GetGameById(gameID int) (*models.Game, error) {
	query := `SELECT id, name, invite_link, current_task_id, total_players, status FROM games WHERE id = ?`
	row := db.QueryRow(query, gameID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.InviteLink, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
	if err != nil {
		log.Printf("Error fetching game with ID %d: %v", gameID, err)
		return nil, err
	}

	return game, nil
}

// GetGameByChatId getting a game by chat ID
func GetGameByChatId(chatID int64) (*models.Game, error) {
	query := `SELECT id, name, invite_link, current_task_id, total_players, status FROM games WHERE game_chat_id = ?`
	row := db.QueryRow(query, chatID)

	game := &models.Game{}

	err := row.Scan(&game.ID, &game.Name, &game.InviteLink, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
	if err != nil {
		log.Printf("Error fetching game with chat ID %d: %v", chatID, err)
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
		log.Println("Failed to add player:", err)
		return err
	}

	log.Printf("DB-add-player-log:Player %s as role %s added to game %d", player.UserName, player.Role, player.GameID)

	return nil
}

func GetPlayerCount(gameId int) (int, error) {
	query := `SELECT COUNT(*) FROM players WHERE game_id = ?`
	row := db.QueryRow(query, gameId)

	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Printf("Error fetching player count for game ID %d: %v", gameId, err)
		return 0, err
	}

	return count, nil
}

func GetAllPlayersByGameID(gameId int) ([]models.Player, error) {
	query := `SELECT id, username, name, game_id, passes, role FROM players WHERE game_id = ?`
	rows, err := db.Query(query, gameId)
	if err != nil {
		log.Printf("Error fetching players for game ID %d: %v", gameId, err)
		return nil, err
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.Player
		err := rows.Scan(&player.ID, &player.UserName, &player.Name, &player.GameID, &player.Status, &player.Skipped, &player.Role)
		if err != nil {
			log.Printf("Error scanning player: %v", err)
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
		log.Printf("Error fetching task count for game ID %d: %v", gameId, err)
		return 0, err
	}

	return count, nil
}

func GetAllTasksByGameID(gameId int) ([]models.Task, error) {
	query := `SELECT id, game_id, question, answer FROM tasks WHERE game_id = ?`
	rows, err := db.Query(query, gameId)
	if err != nil {
		log.Printf("Error fetching tasks for game ID %d: %v", gameId, err)
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.GameID, &task.Question, &task.Answer)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func GetGameByUserId(playerID int64) (*models.Game, error) {
	query := `
		SELECT g.id, g.name, g.game_chat_id, g.invite_link, g.current_task_id, g.total_players, g.status
		FROM players p
		JOIN games g ON p.game_id = g.id
		WHERE p.id = ?
	`

	var game models.Game
	err := db.QueryRow(query, playerID).Scan(
		&game.ID,
		&game.Name,
		&game.GameChatID,
		&game.InviteLink,
		&game.CurrentTaskID,
		&game.TotalPlayers,
		&game.Status,
	)
	if err != nil {
		log.Printf("failed to get game for player_id %d: %v", playerID, err)
		return nil, err
	}

	return &game, nil
}

// UpdatePlayerStatus update player status in DB
func UpdatePlayerStatus(playerID int64, status string) error {
	query := `UPDATE players SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, playerID)
	if err != nil {
		log.Printf("Error updating player status for player ID %d: %v", playerID, err)
		return err
	}

	log.Printf("DB logs: (UpdatePlayerStatus) Player status for player ID %d updated to '%s'", playerID, status)
	return nil
}

// GetPlayerStatus get player status by ID
func GetStatusPlayer(playerID int64) (string, error) {
	query := `SELECT status FROM players WHERE id = ?`
	row := db.QueryRow(query, playerID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		log.Printf("Error fetching player status for player ID %d: %v", playerID, err)
		return "", err
	}

	log.Printf("DB logs: (GetStatusPlayer) Player status for player ID %d is '%s'", playerID, status)
	return status, nil
}

// AddPlayerAnswer add player answer to DB
func AddPlayerResponse(playerResponse *models.PlayerResponse) error {
	log.Println("DB logs: (AadPlayerResponse) AddPlayerResponse was called")
	query := `INSERT INTO player_responses (player_id, game_id, task_id, has_answer, skipped) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, playerResponse.PlayerID, playerResponse.GameID, playerResponse.TaskID, playerResponse.HasResponse, playerResponse.Skipped)
	if err != nil {
		log.Println("Failed to add player answer:", err)
		return err
	}

	log.Printf("DB logs: (AddPlayerResponse) Player response for game ID %d and task ID %d added", playerResponse.GameID, playerResponse.TaskID)
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
		log.Printf("Error checking skip count: %v", err)
		return nil, err
	}

	if skipCount >= 3 {
		status.SkipLimitReached = true
		status.RemainingSkips = 0
		log.Printf("Player ID %d has reached the skip limit", playerID)
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
			log.Printf("Error inserting skipped response: %v", err)
			return nil, err
		}
		log.Printf("Player ID %d skipped task ID %d (new entry)", playerID, taskID)

	case err != nil:
		// Any other DB error
		log.Printf("Error checking existing response: %v", err)
		return nil, err

	default:
		// Response exists
		if hasAnswer {
			status.AlreadyAnswered = true
			log.Printf("DB logs: (SkipPlayerResponse) Player ID %d already answered task ID %d", playerID, taskID)
			return status, nil
		}
		if skipped {
			status.AlreadySkipped = true
			log.Printf("Player ID %d already skipped task ID %d", playerID, taskID)
			return status, nil
		}

		// Update existing entry to mark as skipped
		_, err := db.Exec(`
			UPDATE player_responses SET skipped = 1 
			WHERE player_id = ? AND game_id = ? AND task_id = ?
		`, playerID, gameID, taskID)
		if err != nil {
			log.Printf("Error updating skipped flag: %v", err)
			return nil, err
		}
		log.Printf("Player ID %d skipped task ID %d (updated existing entry)", playerID, taskID)
	}

	return status, nil
}


// Update current task ID in game
func UpdateCurrentTaskID(gameID int, taskID int) error {
	query := `UPDATE games SET current_task_id = ? WHERE id = ?`
	_, err := db.Exec(query, taskID, gameID)
	if err != nil {
		log.Printf("Error updating current task ID for game ID %d: %v", gameID, err)
		return err
	}

	log.Printf("DB logs: (UpdateCurrentTaskID) Current task ID for game ID %d updated to %d", gameID, taskID)
	return nil
}