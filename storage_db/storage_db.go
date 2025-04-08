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
				passes INTEGER DEFAULT 0,
				role TEXT NOT NULL
			)`,
		},
		{
			"games",
			`CREATE TABLE IF NOT EXISTS games (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL,
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
			"player_answers",
			`CREATE TABLE IF NOT EXISTS player_answers (
				id INTEGER PRIMARY KEY,
				player_id INTEGER,
				game_id INTEGER,
				task_id INTEGER,
				answer TEXT NOT NULL,
				is_correct BOOLEAN DEFAULT FALSE,
				FOREIGN KEY (player_id) REFERENCES players(id),
				FOREIGN KEY (game_id) REFERENCES games(id),
				FOREIGN KEY (task_id) REFERENCES tasks(id)
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


// CreateGameRoom добавляет новую игровую комнату в базу данных и возвращает её ID и инвайт-ссылку
// CreateGameRoom добавляет новую игровую комнату в базу данных и возвращает её ID и инвайт-ссылку
// func CreateGameRoom(gameRoom models.GameRoom) (int, string, error) {
// 	// Вставка данных в таблицу game_rooms
// 	query := `INSERT INTO game_rooms (title, invite_link) VALUES (?, ?)`
// 	res, err := db.Exec(query, gameRoom.Title, gameRoom.InviteLink)
// 	if err != nil {
// 		return 0, "", err
// 	}

// 	// Получаем ID созданной игровой комнаты
// 	gameRoomID, err := res.LastInsertId()
// 	if err != nil {
// 		return 0, "", err
// 	}

// 	// Создаем инвайт-ссылку с использованием ID игровой комнаты
// 	inviteLink := utils.GenerateInviteLink(int(gameRoomID))

// 	// Обновляем запись с инвайт-ссылкой
// 	_, err = db.Exec(`UPDATE game_rooms SET invite_link = ? WHERE id = ?`, inviteLink, gameRoomID)
// 	if err != nil {
// 		return 0, "", err
// 	}

// 	log.Printf("Game room '%s' created with ID %d and invite link %s", gameRoom.Title, gameRoomID, inviteLink)

// 	return int(gameRoomID), inviteLink, nil
// }



// CreateGame добавляет новую игру в базу данных и возвращает ее ID
func CreateGame(gameName string) (*models.Game, error) {
	game := &models.Game{
		ID: 1,
		Name: gameName,
		InviteLink: utils.GenerateInviteLink(1),
		CurrentTaskID: 0,
		TotalPlayers: 0,
		Status: "waiting",
	}

	query := `INSERT INTO games (name, invite_link, status ) VALUES (?, ?, ?)`
	res, err := db.Exec(query, game.Name, game.InviteLink, game.Status)
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
	log.Printf("Game '%s' создана с ID %d", game.Name, game.ID)

	return game, nil

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

// GetGameRoomByID получает игровую комнату по её ID
// func GetGameRoomByID(gameRoomID int) (*models.GameRoom, error) {
// 	query := `SELECT id, title, invite_link, game_id FROM game_rooms WHERE id = ?`
// 	row := db.QueryRow(query, gameRoomID)

// 	var gameRoom models.GameRoom
// 	var gameID sql.NullInt64

// 	err := row.Scan(&gameRoom.ID, &gameRoom.Title, &gameRoom.InviteLink, &gameID)
// 	if err != nil {
// 		log.Printf("Error fetching game room with ID %d: %v", gameRoomID, err)
// 		return nil, err
// 	}

// 	// Если gameID не NULL, записываем его
// 	if gameID.Valid {
// 		gameRoom.GameID = new(int)
// 		*gameRoom.GameID = int(gameID.Int64)
// 	} else {
// 		gameRoom.GameID = nil
// 	}

// 	return &gameRoom, nil
// }

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

// AddPlayerToGame add player to game
func AddPlayerToGame(player *models.Player) error {
	query := `INSERT INTO players (id, username, name, game_id, passes, role) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, player.ID, player.UserName, player.Name, player.GameID, player.Passes, player.Role)
	if err != nil {
		log.Println("Failed to add player:", err)
		return err
	}

	log.Printf("Player %s as role %s added to game %d", player.UserName, player.Role, player.GameID)

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
		err := rows.Scan(&player.ID, &player.UserName, &player.Name, &player.GameID, &player.Passes, &player.Role)
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