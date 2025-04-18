package models

type Player struct {
    ID int64
    UserName string
    Name string
	Passes uint8
    GameID int
    Role string // "admin", "player"
}

type Game struct {
	ID int
	Name string
	GameChatID int64
    InviteLink string
	CurrentTaskID int
    TotalPlayers int // max 5
	Status string // "waiting", "playing", "finished"
}

type GamePlayer struct {
    GameID   int
    PlayerID int
    Status   string // "joined", "playing", "finished"
}

type Task struct {
	ID int
	GameID int
	Question string
	Answer string
}

type PlayerAnswer struct {
    ID        int    
    PlayerID  int    
    GameID    int    
    TaskID    int    
    Answer    string 
    IsCorrect bool   
}
