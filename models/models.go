package models

type Player struct {
    ID int64
    UserName string
    Name string
	  Status string
	  Skipped int
    GameID int
    Role string // "admin", "player"
}

type Game struct {
	ID int
	Name string
	GameChatID int64
	//MsgJointID int // ID of the message with the "Join" button
    //InviteLink string
	CurrentTaskID int
    TotalPlayers int // max 5
	Status string // "waiting", "playing", "finished"
}

type GamePlayer struct {
    GameID   int
    PlayerID int
    Status   string // "joined", "playing", "finished"
}

// Struct for storing task information in DB
// type Task struct {
// 	ID int
// 	GameID int
// 	Question string
// 	Answer string
// }

type PlayerResponse struct {
    ID        int    
    PlayerID  int64    
    GameID    int    
    TaskID    int    
    HasResponse bool 
    Skipped bool   
}

type GameState struct {
	GameID int64 
	Current  string // Текущее состояние
}

type SkipStatus struct {
	AlreadyAnswered     bool // true, if player already answer
	AlreadySkipped      bool // true, if player already skipped this task
	SkipLimitReached    bool // true, if player already has three skipсли у игрока уже 3 пропуска
	RemainingSkips      int  // количество оставшихся пропусков
}

type AddResponseResult struct {
	AlreadyAnswered bool
	AlreadySkipped  bool
	Success         bool
}

// Struct for storing task information for downloading from JSON
type Task struct {
	ID 		int    `json:"id"`
	Tittle string `json:"title"`
	Description string `json:"description"`
}

// Const of state
const (
	StatusGameWaiting	   = "waiting"
	StatusGamePlaying	   = "playing"
	StatusGameFinished	   = "finished"

	StatusPlayerWaiting	   = "waiting_"
	StatusPlayerNoWaiting  = "no_waiting"
)