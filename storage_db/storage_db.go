package storage_db

import (
	"database/sql"
	"log"

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
			id INTEGER PRIMARY KEY,
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
