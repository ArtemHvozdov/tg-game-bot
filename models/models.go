package models

type Player struct {
    ID int
    UserName string
    Name string
	Passes uint8
	GameRoomID int
}

type GameRoom struct{
    ID int
    Title string
	InviteLink string
    GameID int
}

type Game struct {
	ID int
	Name string
	GameRoomID int
	CurrentTaskID int
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
