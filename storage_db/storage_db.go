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
		{"players", `CREATE TABLE IF NOT EXISTS players (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL,
			name TEXT NOT NULL,
			passes INTEGER DEFAULT 0,
			game_room_id INTEGER,
			FOREIGN KEY (game_room_id) REFERENCES game_rooms(id)
		)`},
		{"game_rooms", `CREATE TABLE IF NOT EXISTS game_rooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			invite_link TEXT NOT NULL UNIQUE,
			game_id INTEGER,
			FOREIGN KEY (game_id) REFERENCES games(id)
		)`},
		{"games", `CREATE TABLE IF NOT EXISTS games (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			game_room_id INTEGER,
			current_task_id INTEGER,
			status TEXT CHECK(status IN ('waiting', 'playing', 'finished')) NOT NULL,
			FOREIGN KEY (game_room_id) REFERENCES game_rooms(id)
		)`},
		{"game_players", `CREATE TABLE IF NOT EXISTS game_players (
			game_id INTEGER,
			player_id INTEGER,
			status TEXT CHECK(status IN ('joined', 'playing', 'finished')) NOT NULL,
			PRIMARY KEY (game_id, player_id),
			FOREIGN KEY (game_id) REFERENCES games(id),
			FOREIGN KEY (player_id) REFERENCES players(id)
		)`},
		{"tasks", `CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY,
			game_id INTEGER,
			question TEXT NOT NULL,
			answer TEXT NOT NULL,
			FOREIGN KEY (game_id) REFERENCES games(id)
		)`},
		{"player_answers", `CREATE TABLE IF NOT EXISTS player_answers (
			id INTEGER PRIMARY KEY,
			player_id INTEGER,
			game_id INTEGER,
			task_id INTEGER,
			answer TEXT NOT NULL,
			is_correct BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (player_id) REFERENCES players(id),
			FOREIGN KEY (game_id) REFERENCES games(id),
			FOREIGN KEY (task_id) REFERENCES tasks(id)
		)`},
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
func CreateGameRoom(gameRoom models.GameRoom) (int, string, error) {
	// Вставка данных в таблицу game_rooms
	query := `INSERT INTO game_rooms (title, invite_link) VALUES (?, ?)`
	res, err := db.Exec(query, gameRoom.Title, gameRoom.InviteLink)
	if err != nil {
		return 0, "", err
	}

	// Получаем ID созданной игровой комнаты
	gameRoomID, err := res.LastInsertId()
	if err != nil {
		return 0, "", err
	}

	// Создаем инвайт-ссылку с использованием ID игровой комнаты
	inviteLink := utils.GenerateInviteLink(int(gameRoomID))

	// Обновляем запись с инвайт-ссылкой
	_, err = db.Exec(`UPDATE game_rooms SET invite_link = ? WHERE id = ?`, inviteLink, gameRoomID)
	if err != nil {
		return 0, "", err
	}

	log.Printf("Game room '%s' created with ID %d and invite link %s", gameRoom.Title, gameRoomID, inviteLink)

	return int(gameRoomID), inviteLink, nil
}



// CreateGame добавляет новую игру в базу данных и возвращает ее ID
func CreateGame(game models.Game) (int, error) {
	query := `INSERT INTO games (name, status) VALUES (?, ?)`
	res, err := db.Exec(query, game.Name, game.Status)
	if err != nil {
		log.Println("Ошибка при добавлении игры в БД:", err)
		return 0, err
	}

	// Получаем ID созданной игры
	gameID, err := res.LastInsertId()
	if err != nil {
		log.Println("Ошибка получения ID созданной игры:", err)
		return 0, err
	}

	log.Printf("Game '%s' создана с ID %d", game.Name, gameID)
	return int(gameID), nil
}

// CreateTask добавляет новую задачу (вопрос и ответ) в базу данных
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
func GetGameRoomByID(gameRoomID int) (*models.GameRoom, error) {
	query := `SELECT id, title, invite_link, game_id FROM game_rooms WHERE id = ?`
	row := db.QueryRow(query, gameRoomID)

	var gameRoom models.GameRoom
	var gameID sql.NullInt64

	err := row.Scan(&gameRoom.ID, &gameRoom.Title, &gameRoom.InviteLink, &gameID)
	if err != nil {
		log.Printf("Error fetching game room with ID %d: %v", gameRoomID, err)
		return nil, err
	}

	// Если gameID не NULL, записываем его
	if gameID.Valid {
		gameRoom.GameID = new(int)
		*gameRoom.GameID = int(gameID.Int64)
	} else {
		gameRoom.GameID = nil
	}

	return &gameRoom, nil
}
