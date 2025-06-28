-- 001_initial_schema.sql

CREATE TABLE IF NOT EXISTS players (
	id INTEGER,
	username TEXT NOT NULL,
	name TEXT NOT NULL,
	game_id INTEGER,
	status TEXT,
	skipped INT,
	role TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS games (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	game_chat_id INTEGER,
	msg_join_id INTEGER NOT NULL DEFAULT 0,
	current_task_id INTEGER NOT NULL DEFAULT 0,
	total_players INTEGER NOT NULL DEFAULT 0,
	status TEXT CHECK(status IN ('waiting', 'playing', 'finished')) NOT NULL
);

CREATE TABLE IF NOT EXISTS game_players (
	game_id INTEGER,
	player_id INTEGER,
	status TEXT CHECK(status IN ('joined', 'playing', 'finished')) NOT NULL,
	PRIMARY KEY (game_id, player_id),
	FOREIGN KEY (game_id) REFERENCES games(id),
	FOREIGN KEY (player_id) REFERENCES players(id)
);

CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	game_id INTEGER,
	question TEXT NOT NULL,
	answer TEXT NOT NULL,
	FOREIGN KEY (game_id) REFERENCES games(id)
);

CREATE TABLE IF NOT EXISTS player_responses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER,
	game_id INTEGER,
	task_id INTEGER,
	has_answer BOOLEAN,
	skipped BOOLEAN DEFAULT FALSE,
	FOREIGN KEY (player_id) REFERENCES players(id),
	FOREIGN KEY (game_id) REFERENCES games(id),
	FOREIGN KEY (task_id) REFERENCES tasks(id)
);

CREATE TABLE IF NOT EXISTS game_state (
	game_id INTEGER PRIMARY KEY UNIQUE,
	status TEXT NOT NULL,
	FOREIGN KEY (game_id) REFERENCES games(id)
);
