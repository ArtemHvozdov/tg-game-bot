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
    has_response bool 
    skipped bool   
}

type GameState struct {
	GameID int64 
	Current  string // Текущее состояние
}


// Const of state
const (
	StateIdle              = "idle"
	StateWaitingChat       = "await_creating_chat"
	StateCheckAdmin        = "await_check_admin"
	StateJoin              = "await_joined_player"
	StateGameStarted       = "game_started"

	StateWaitingAnswer01   = "waiting_answer_01"
	StateWaitingAnswer02   = "waiting_answer_02"
	StateWaitingAnswer03   = "waiting_answer_03"
	StateWaitingAnswer04   = "waiting_answer_04"
	StateWaitingAnswer05   = "waiting_answer_05"
	StateWaitingAnswer06   = "waiting_answer_06"
	StateWaitingAnswer07   = "waiting_answer_07"
	StateWaitingAnswer08   = "waiting_answer_08"
	StateWaitingAnswer09   = "waiting_answer_09"
	StateWaitingAnswer10   = "waiting_answer_10"
	StateWaitingAnswer11   = "waiting_answer_11"
	StateWaitingAnswer12   = "waiting_answer_12"

	StateGameFinished      = "game_finished"

	StatusGameWaiting	   = "waiting"
	StatusGamePlaying	   = "playing"
	StatusGameFinished	   = "finished"
)

// Set устанавливает новое состояние
func (gs *GameState) Set(state string) {
	gs.Current = state
}

// Get возвращает текущее состояние
func (gs *GameState) Get() string {
	return gs.Current
}

// AdvanceTaskState переходит к следующему таску или завершает игру
func (gs *GameState) AdvanceTaskState() {
	switch gs.Current {
	case StateWaitingAnswer01:
		gs.Current = StateWaitingAnswer02
	case StateWaitingAnswer02:
		gs.Current = StateWaitingAnswer03
	case StateWaitingAnswer03:
		gs.Current = StateWaitingAnswer04
	case StateWaitingAnswer04:
		gs.Current = StateWaitingAnswer05
	case StateWaitingAnswer05:
		gs.Current = StateWaitingAnswer06
	case StateWaitingAnswer06:
		gs.Current = StateWaitingAnswer07
	case StateWaitingAnswer07:
		gs.Current = StateWaitingAnswer08
	case StateWaitingAnswer08:
		gs.Current = StateWaitingAnswer09
	case StateWaitingAnswer09:
		gs.Current = StateWaitingAnswer10
	case StateWaitingAnswer10:
		gs.Current = StateWaitingAnswer11
	case StateWaitingAnswer11:
		gs.Current = StateWaitingAnswer12
	case StateWaitingAnswer12:
		gs.Current = StateGameFinished
	default:
		// нет перехода, остаёмся как есть
	}
}