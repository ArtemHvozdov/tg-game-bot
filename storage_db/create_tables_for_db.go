package storage_db

import (
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
)

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
				msg_join_id INTEGER NOT NULL DEFAULT 0,
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
		{
			"subtask_answers",
			`CREATE TABLE IF NOT EXISTS subtask_answers (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				game_id INTEGER,
				task_id INTEGER,
				question_index INTEGER,
				answerer_user_id INTEGEER,
				selected_user_id INTEGER,
				selected_username TEXT
			)`,
		},
		{
			"subtask_10_answers",
			`CREATE TABLE IF NOT EXISTS subtask_10_answers (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				game_id INTEGER NOT NULL,
				task_id INTEGER NOT NULL,
				question_index INTEGER NOT NULL,
				question_id INTEGER NOT NULL,
				answerer_user_id INTEGER NOT NULL,
				selected_option TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
		},
	}

	for _, q := range queries {
		if _, err := Db.Exec(q.query); err != nil {
			return err
		}
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: createTables",
			"table": q.tableName,
		}).Info("Table has been created or already exists.")
	}

	return nil
}