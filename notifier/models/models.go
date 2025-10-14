package models

type Game struct {
	ID         int
	Name       string
	GameChatID int64
	CurrentTaskID int
	TotalPlayers  int    // max 5
	Status        string // "waiting", "playing", "finished"
}