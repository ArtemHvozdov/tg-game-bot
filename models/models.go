package models

import "time"

type Player struct {
	ID       int64
	UserName string
	Name     string
	Status   string
	Skipped  int
	GameID   int
	Role     string // "admin", "player"
}

type Game struct {
	ID         int
	Name       string
	GameChatID int64
	//MsgJointID int // ID of the message with the "Join" button
	//InviteLink string
	CurrentTaskID int
	TotalPlayers  int    // max 5
	Status        string // "waiting", "playing", "finished"
}

type GamePlayer struct {
	GameID   int
	PlayerID int
	Status   string // "joined", "playing", "finished"
}

type PlayerResponse struct {
	ID          int
	PlayerID    int64
	GameID      int
	TaskID      int
	HasResponse bool
	Skipped     bool
}

type GameState struct {
	GameID  int64
	Current string // Текущее состояние
}

type SkipStatus struct {
	AlreadyAnswered  bool // true, if player already answer
	AlreadySkipped   bool // true, if player already skipped this task
	SkipLimitReached bool // true, if player already has three skipсли у игрока уже 3 пропуска
	RemainingSkips   int  // количество оставшихся пропусков
}

type AddResponseResult struct {
	AlreadyAnswered bool
	AlreadySkipped  bool
	Success         bool
}

// Struct for storing task information for downloading from JSON
type Task struct {
	ID          int    `json:"id"`
	Tittle      string `json:"title"`
	Description string `json:"description"`
}

// type SubTask struct {
// 	subtasks []string
// }

type SubtaskAnswer struct {
	ID 			  int    `json:"id"`
	GameID       int    `json:"game_id"`
	TaskID       int    `json:"task_id"`
	QuestionIndex uint    `json:"question_index"`
	AnswererUserID int64  `json:"answerer_user_id"` // кто отвечал
	SelectedUserID int64  `json:"selected_user_id"` // кого выбрали
	SelectedUsername string `json:"selected_username"` // username выбранного (для быстрого доступа)
}

// Subtask10Answer represents a user's answer to a subtask 10 question
type Subtask10Answer struct {
	ID               int       `json:"id"`
	GameID           int       `json:"game_id"`
	TaskID           int       `json:"task_id"`
	QuestionIndex    int       `json:"question_index"`
	QuestionID       int       `json:"question_id"`        // ID from JSON file
	AnswererUserID   int64     `json:"answerer_user_id"`
	SelectedOption   string    `json:"selected_option"`    // Data value from JSON
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Const of state
const (
	StatusGameWaiting  = "waiting"
	StatusGamePlaying  = "playing"
	StatusGameFinished = "finished"

	StatusPlayerWaiting   = "waiting_"
	StatusPlayerNoWaiting = "no_waiting"

	// Unique buttons
	UniqueHelp = "help_menu"
	UniqueSupport = "support"
	UniqueExitGame = "exit_game"
	UniqueExactlyExit = "exit_confirm"
	UniqueReturnToGame = "return_to_game"
	UniqueJoinGameBtn = "join_game_btn"
	UniqueStartGame = "start_game"
	UniqueAnswerTask = "answer_task"
	UniqueSkipTask = "skip_task"
	UniqueFeedback = "feedback"

	// Static messages
	MsgInviteToJoinGame          = "invite_to_join_game" // +
	MsgAdminWantToJoinGame       = "admin_want_to_join_game" // +
	MsgUsaulPlayerWantToJoinGame = "usual_player_want_to_join_game" // +
	MsgPlayerExitGame            = "player_exit_game" // +
	MsgAdminExitGame             = "admin_exit_game" // +
	MsgExactlyExitGame           = "exactly_exit_game" // +
	MsgReturnToGame              = "return_to_game" // +
	MsgOnlyAdminCanStartGame     = "only_admin_can_start_game" // +
	MsgPlayerGameeAlreadyStarted = "player_gamee_already_started" // msg no need
	MsgAdminGameAlreadyStarted   = "admin_game_already_started" // +
	MsgUserIsNotInGame 			 = "user_is_not_in_game" // +
	MsgAdminStartGameBtn 	  	 = "admin_start_game_btn" // +
	MsgUserAlreadySkipTask 		 = "user_already_skip_task" // +
	MsgUserAnswerAccepted 		 = "user_answer_accepted" 

	MsgSkipFirstTime          	 = "skip_first_time"
	MsgSkipSecondTime         	 = "skip_second_time"
	MsgSkipThirdTime          	 = "skip_third_time"
	MsgSkipLimitReached       	 = "skip_limit"

	LinkInstagram        		 = "instagram"
	LinkTikTok            		 = "tiktok"
)

    
	// Коли учасник натискаэ "Приэднатися до гри", але він уде в грі - "🎉 @%s, ти вже в грі! Не нервуйся" ++
	// Коли бот приймаэ відповідь учасника - "🎉 @%s, ти вже в грі! Не нервуйся",
	// Коли учасник настискає "Хочу відповісти", але цей учасника вже виконав завдання - "📝 @%s, ти вже виконала це завдання."
	// Коли учасник натискає "Пропустити", але цей учасника вже пропустив це завдання - "⏭️ @%s, ти вже пропустила це завдання."