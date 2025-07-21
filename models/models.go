package models

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
)


	// menuIntro = &telebot.ReplyMarkup{}
	// menuExit = &telebot.ReplyMarkup{}

	// introBtnHelp = menuIntro.Data("Хелп", "help_menu")
	// introBtnSupport = menuIntro.URL("Техпідтримка", "https://t.me/Jay_jayss")
	// introBtnExit = menuIntro.Data("Вийти з гри", fmt.Sprintf("exit_%d", gameID))

	// btnExactlyExit = menuExit.Data("Вийти з гри", fmt.Sprintf("exit_game_%d", gameID))
	// btnReturnToGame = menuExit.Data(" << Повернутися до гри", "return_to_game")

	// startMenu := &telebot.ReplyMarkup{}
	// startBtnSupport := startMenu.URL("🕹️ Техпідтримка", "https://t.me/Jay_jayss")

	// menu := &telebot.ReplyMarkup{}
	// btnStartGame := menu.Data("Почати гру", "start_game")

	// joinBtn := telebot.InlineButton{
	// 		Unique: "join_game_btn",
	// 		Text:   "🎲 Приєднатися до гри",
	// 	}
	// 	inline := &telebot.ReplyMarkup{}
	// 	inline.InlineKeyboard = [][]telebot.InlineButton{
	// 		{joinBtn},
	// 	}

	// inlineKeys := &telebot.ReplyMarkup{} // initialize inline keyboard

	// 	answerBtn := inlineKeys.Data("Хочу відповісти", "answer_task", fmt.Sprintf("waiting_%d", task.ID))
	// 	skipBtn := inlineKeys.Data("Пропустити", "skip_task", fmt.Sprintf("skip_%d", task.ID))

	// 	inlineKeys.Inline(
	// 		inlineKeys.Row(answerBtn, skipBtn),
	// 	)	