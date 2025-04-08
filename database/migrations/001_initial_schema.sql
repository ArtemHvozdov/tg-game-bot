-- +++ Create players table
CREATE TABLE IF NOT EXISTS players (
	id INTEGER PRIMARY KEY,
	username TEXT NOT NULL,
	name TEXT NOT NULL,
	passes INTEGER DEFAULT 0,
	game_room_id INTEGER,
	role TEXT NOT NULL,
	FOREIGN KEY (game_room_id) REFERENCES game_rooms(id)
);

-- +++ Create game_rooms table
CREATE TABLE IF NOT EXISTS game_rooms (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	invite_link TEXT NOT NULL UNIQUE,
	game_id INTEGER,
	FOREIGN KEY (game_id) REFERENCES games(id)
);

-- +++ Create games table
CREATE TABLE IF NOT EXISTS games (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	invite_link TEXT NOT NULL UNIQUE,
	current_task_id INTEGER,
	total_players INTEGER,
	status TEXT CHECK(status IN ('waiting', 'playing', 'finished')) NOT NULL,
	FOREIGN KEY (game_room_id) REFERENCES game_rooms(id)
);

-- +++ Create game_players table
CREATE TABLE IF NOT EXISTS game_players (
	game_id INTEGER,
	player_id INTEGER,
	status TEXT CHECK(status IN ('joined', 'playing', 'finished')) NOT NULL,
	PRIMARY KEY (game_id, player_id),
	FOREIGN KEY (game_id) REFERENCES games(id),
	FOREIGN KEY (player_id) REFERENCES players(id)
);

-- +++ Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	game_id INTEGER,
	question TEXT NOT NULL,
	answer TEXT NOT NULL,
	FOREIGN KEY (game_id) REFERENCES games(id)
);

-- +++ Create player_answers table
CREATE TABLE IF NOT EXISTS player_answers (
	id INTEGER PRIMARY KEY,
	player_id INTEGER,
	game_id INTEGER,
	task_id INTEGER,
	answer TEXT NOT NULL,
	is_correct BOOLEAN DEFAULT FALSE,
	FOREIGN KEY (player_id) REFERENCES players(id),
	FOREIGN KEY (game_id) REFERENCES games(id),
	FOREIGN KEY (task_id) REFERENCES tasks(id)
);
