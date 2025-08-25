package voting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

// Subtask structure for loading from JSON
type Subtask struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Update PollSession version
type PollSession struct {
	GameID        int
	Subtasks      []Subtask
	PollID        string
	PollMessage   *telebot.Message
	MainTaskMsg   *telebot.Message // Added: message with the main task text
	StartTime     time.Time
	IsActive      bool
	IsCompleted   bool
	mutex         sync.RWMutex
}

// PollManager manages poll sessions
type PollManager struct {
	sessions map[int]*PollSession    // gameID -> session
	pollGame map[string]int          // pollID -> gameID (for quick lookup)
	mutex    sync.RWMutex
}

var GlobalPollManager = &PollManager{
	sessions: make(map[int]*PollSession),
	pollGame: make(map[string]int),
}

func processTaskDescription(subtaskID int, description string) string {
	// Map of subtask ID to YouTube links
	youtubeLinks := map[int]string{
		1: "https://www.youtube.com/shorts/Cb5ljmm3820?si=qWu3oIgz890cV0xo", // Lady Gaga - Paparazzi
		2: "https://www.youtube.com/watch?v=Mkuw7vdi-VA", // Phoebe Buffay - Smelly Cat
		3: "", // YouTube is not needed for memes
	}

	// Processing each subtask individually
	switch subtaskID {
	case 1: // Dancing to Lady Gaga
		processed := description
		if link, exists := youtubeLinks[subtaskID]; exists && link != "" {
			// Replacing "цього" with a link
			processed = strings.Replace(processed, "цього", `<a href="https://www.youtube.com/shorts/Cb5ljmm3820?si=qWu3oIgz890cV0xo">цього</a>`, 1)
		}
		return processed

	case 2: // Singing Smelly Cat
		processed := description
		if link, exists := youtubeLinks[subtaskID]; exists && link != "" {
			// Replacing "тут" with a link
			processed = strings.Replace(processed, "тут", `<a href="https://www.youtube.com/watch?v=Mkuw7vdi-VA">тут</a>`, 1)
		}

		// Searching for song text after emoji and colon
		// Trying different search options
		patterns := []string{"🥹:\\n", "🥹:\n", "🥹:"}
		songTextStart := -1
		patternLength := 0
		
		for _, pattern := range patterns {
			if idx := strings.Index(processed, pattern); idx != -1 {
				songTextStart = idx
				patternLength = len(pattern)
				break
			}
		}
		
		songTextEnd := strings.Index(processed, "\n\nПотренуйтеся")
		
		if songTextStart != -1 && songTextEnd != -1 && songTextEnd > songTextStart {
			// Splitting the text
			beforeSong := processed[:songTextStart+len("🥹:")] // only emoji and colon
			songText := processed[songTextStart+patternLength:songTextEnd]
			afterSong := processed[songTextEnd:]

			// Cleaning the beginning of the song text from unnecessary characters
			songText = strings.TrimLeft(songText, "\\n\n ")

			// Splitting the song text into lines
			songLines := strings.Split(songText, "\n")
			var formattedLines []string
			
			// Adding an empty line
			formattedLines = append(formattedLines, "")
			
			for _, line := range songLines {
				line = strings.TrimSpace(line)
				if line != "" {
					formattedLines = append(formattedLines, "<em>"+line+"</em>")
				}
			}

			// Putting it back together
			formattedSongText := strings.Join(formattedLines, "\n")
			processed = beforeSong + "\n" +formattedSongText + afterSong
		}
		
		return processed

	case 3: // Memes - YouTube is not needed
		return description

	default:
		return description
	}
}

// Load subtasks from JSON file
func loadSubtask5() ([]Subtask, error) {
	data, err := ioutil.ReadFile("internal/data/tasks/subtasks/subtask_5.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read subtask_5.json: %v", err)
	}

	var subtasks []Subtask
	err = json.Unmarshal(data, &subtasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subtasks: %v", err)
	}

	return subtasks, nil
}

// StartPollVoting - starts voting using Telegram Poll
func (pm *PollManager) StartPollVoting(gameID int) (*PollSession, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Checking if there is an active vote
	if session, exists := pm.sessions[gameID]; exists && session.IsActive {
		return nil, fmt.Errorf("voting already active for game %d", gameID)
	}

	// Loading options for voting
	subtasks, err := loadSubtask5()
	if err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %v", err)
	}

	if len(subtasks) == 0 {
		return nil, fmt.Errorf("no subtasks found")
	}

	// Creating a new session with nulled fields for messages
	session := &PollSession{
		GameID:      gameID,
		Subtasks:    subtasks,
		StartTime:   time.Now(),
		IsActive:    true,
		IsCompleted: false,
		MainTaskMsg: nil, // Will be set in CreateTelegramPoll
		PollMessage: nil, // Will be set in CreateTelegramPoll
	}

	pm.sessions[gameID] = session

	utils.Logger.Infof("Created poll session for game %d with %d options", gameID, len(subtasks))
	return session, nil
}

// CreateTelegramPoll - creates a Telegram poll and sends it
func (pm *PollManager) CreateTelegramPoll(bot *telebot.Bot, chatID int64, session *PollSession, mainTaskText string, keyboard *telebot.ReplyMarkup) error {
	// Preparing options from titles
	var options []string
	for _, subtask := range session.Subtasks {
		options = append(options, subtask.Title)
	}

	question := "🗳️ Оберіть завдання для виконання:"
	
	// Creating a poll using the Raw method
	data := map[string]interface{}{
		"chat_id":                chatID,
		"question":               question,
		"options":                options,
		"is_anonymous":           false,
		"type":                   "regular",
		"allows_multiple_answers": false,
	}

	// Sending the main task message and SAVING it
	chat := &telebot.Chat{ID: chatID}
	mainTaskMsg, err := bot.Send(chat, mainTaskText, telebot.ModeMarkdown)
	if err != nil {
		return fmt.Errorf("failed to send main task message: %v", err)
	}

	// Saving the main task message in the session
	session.MainTaskMsg = mainTaskMsg

	time.Sleep(200 * time.Millisecond)
	
	resp, err := bot.Raw("sendPoll", data)
	if err != nil {
		return fmt.Errorf("failed to send poll via Raw API: %v", err)
	}

	// Processing the response
	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
			Poll      struct {
				ID      string `json:"id"`
				Question string `json:"question"`
				Options []struct {
					Text       string `json:"text"`
					VoterCount int    `json:"voter_count"`
				} `json:"options"`
			} `json:"poll"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to unmarshal poll response: %v", err)
	}

	if !result.Ok {
		return fmt.Errorf("API returned not ok")
	}

	// Creating a Message object for the poll
	pollMsg := &telebot.Message{
		ID: result.Result.MessageID,
		Chat: &telebot.Chat{ID: chatID},
		Poll: &telebot.Poll{
			ID:       result.Result.Poll.ID,
			Question: result.Result.Poll.Question,
		},
	}

	// Saving information about the poll
	pm.mutex.Lock()
	session.PollID = pollMsg.Poll.ID
	session.PollMessage = pollMsg
	pm.pollGame[pollMsg.Poll.ID] = session.GameID
	pm.mutex.Unlock()

	utils.Logger.Infof("Created Telegram poll %s for game %d", pollMsg.Poll.ID, session.GameID)

	// Starting a timer
	go pm.startPollTimer(bot, chatID, session, keyboard)

	return nil
}


// startPollTimer - starts the timer and ends the voting
func (pm *PollManager) startPollTimer(bot *telebot.Bot, chatID int64, session *PollSession, keyboard *telebot.ReplyMarkup) {
	// time.Sleep(15 * time.Second)
	time.Sleep(1 * time.Minute)

	// Do NOT call bot.StopPoll here, as it is done in ProcessPollResults
	utils.Logger.Infof("Poll timer expired for game %d, processing results...", session.GameID)

	// Processing the results (stopPoll will be called inside)
	winner, err := pm.ProcessPollResultsWithBot(bot, session.GameID)
	if err != nil {
		utils.Logger.Errorf("Failed to process poll results: %v", err)
		return
	}

	// The rest of the code remains the same...
	processedDescription := processTaskDescription(winner.ID, winner.Description)
	
	utils.Logger.Infof("Processing subtask ID %d, original description length: %d, processed length: %d", 
		winner.ID, len(winner.Description), len(processedDescription))

	chat := &telebot.Chat{ID: chatID}
	winnerMessage := fmt.Sprintf("🎉 Переможець голосування:\n\n%s\n\n%s",
		winner.Title, processedDescription)

	_, err = bot.Send(chat, winnerMessage, keyboard, telebot.ModeHTML)
	if err != nil {
		utils.Logger.Errorf("Failed to send winner message: %v", err)
	}

	utils.Logger.Infof("Poll voting completed for game %d, selected subtask ID: %d", session.GameID, winner.ID)
}

// ProcessPollResultsWithBot - version with access to bot for API requests
func (pm *PollManager) ProcessPollResultsWithBot(bot *telebot.Bot, gameID int) (*Subtask, error) {
	// Just calling the existing function, but now we have access to the bot
	return pm.ProcessPollResults(bot, gameID)
}

// ProcessPollResults - processes the voting results and selects a winner
func (pm *PollManager) ProcessPollResults(bot *telebot.Bot, gameID int) (*Subtask, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	session, exists := pm.sessions[gameID]
	if !exists {
		return nil, fmt.Errorf("no poll session found for game %d", gameID)
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.IsActive = false
	session.IsCompleted = true

	utils.Logger.Infof("=== POLL RESULTS FOR GAME %d ===", gameID)

	// Getting results via stopPoll response
	stopPollData := map[string]interface{}{
		"chat_id":    session.PollMessage.Chat.ID,
		"message_id": session.PollMessage.ID,
	}

	resp, err := bot.Raw("stopPoll", stopPollData)
	if err != nil {
		utils.Logger.Errorf("Failed to stop poll and get results: %v", err)
		return pm.fallbackProcessResults(session, gameID)
	}

	var pollResult struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID      string `json:"id"`
			Question string `json:"question"`
			Options []struct {
				Text       string `json:"text"`
				VoterCount int    `json:"voter_count"`
			} `json:"options"`
			TotalVoterCount int  `json:"total_voter_count"`
			IsClosed        bool `json:"is_closed"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &pollResult); err != nil {
		utils.Logger.Errorf("Failed to unmarshal poll results: %v", err)
		return pm.fallbackProcessResults(session, gameID)
	}

	if !pollResult.Ok {
		utils.Logger.Errorf("API returned not ok for poll results")
		return pm.fallbackProcessResults(session, gameID)
	}

	maxVotes := 0
	var winningIndexes []int

	// Analyzing the voting results
	for i, option := range pollResult.Result.Options {
		voteCount := option.VoterCount
		if i < len(session.Subtasks) {
			utils.Logger.Infof("Option %d ('%s'): %d votes", i, session.Subtasks[i].Title, voteCount)

			if voteCount > maxVotes {
				maxVotes = voteCount
				winningIndexes = []int{i}
			} else if voteCount == maxVotes {
				winningIndexes = append(winningIndexes, i)
			}
		}
	}

	// Selecting a winner (in case of a tie - with the smallest ID)
	var winner Subtask
	if len(winningIndexes) > 0 {
		winnerIndex := winningIndexes[0]
		winner = session.Subtasks[winnerIndex]

		// In case of a tie, we choose the one with the smallest ID
		for _, index := range winningIndexes[1:] {
			if session.Subtasks[index].ID < winner.ID {
				winner = session.Subtasks[index]
				winnerIndex = index
			}
		}

		utils.Logger.Infof("Winner: Option %d - Subtask %d ('%s') with %d votes",
			winnerIndex, winner.ID, winner.Title, maxVotes)
	} else {
		// If no one voted, we select the first option
		winner = session.Subtasks[0]
		utils.Logger.Infof("No votes received, selecting first option: Subtask %d ('%s')",
			winner.ID, winner.Title)
	}

	utils.Logger.Infof("Total votes in poll: %d", pollResult.Result.TotalVoterCount)
	utils.Logger.Infof("=== END POLL RESULTS ===")

	// Cleaning up the session
	delete(pm.pollGame, session.PollID)
	delete(pm.sessions, gameID)

	return &winner, nil
}

// fallbackProcessResults - fallback method for processing results
func (pm *PollManager) fallbackProcessResults(session *PollSession, gameID int) (*Subtask, error) {
	utils.Logger.Infof("Using fallback method for poll results processing")
	
	// If we can't get the results, we return the first option
	if len(session.Subtasks) > 0 {
		winner := session.Subtasks[0]
		utils.Logger.Infof("Fallback: selecting first option: Subtask %d ('%s')", winner.ID, winner.Title)
		
		// Cleaning up the session
		delete(pm.pollGame, session.PollID)
		delete(pm.sessions, gameID)
		
		return &winner, nil
	}
	
	return nil, fmt.Errorf("no subtasks available for fallback")
}

// GetSessionByPollID - gets the session by the poll ID
func (pm *PollManager) GetSessionByPollID(pollID string) (*PollSession, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if gameID, exists := pm.pollGame[pollID]; exists {
		if session, sessionExists := pm.sessions[gameID]; sessionExists {
			return session, true
		}
	}
	return nil, false
}

// IsActive - checks if the vote is active for the game
func (pm *PollManager) IsActive(gameID int) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	session, exists := pm.sessions[gameID]
	return exists && session.IsActive
}

// StartSubtask5VotingDirect -  starts voting directly (for calling from SendTasks)
func StartSubtask5VotingDirect(bot *telebot.Bot, chatID int64, mainTaskText string, keyboard *telebot.ReplyMarkup) error {
	// Getting the game from the database
	game, err := storage_db.GetGameByChatId(chatID)
	if err != nil {
		utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chatID, err)
		return fmt.Errorf("помилка отримання гри: %v", err)
	}

	// Creating a voting session
	session, err := GlobalPollManager.StartPollVoting(game.ID)
	if err != nil {
		utils.Logger.Errorf("Failed to start poll voting: %v", err)
		return fmt.Errorf("помилка запуску голосування: %v", err)
	}

	// Creating and sending a Telegram poll
	err = GlobalPollManager.CreateTelegramPoll(bot, chatID, session, mainTaskText, keyboard)
	if err != nil {
		utils.Logger.Errorf("Failed to create Telegram poll: %v", err)
		return fmt.Errorf("помилка створення опитування: %v", err)
	}

	utils.Logger.Infof("Successfully started Telegram poll voting for game %d", game.ID)
	return nil
}

// HandlePollAnswer - poll answer handler (if additional logic is needed)
func HandlePollAnswer(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		// Simple logging without detailed processing
		utils.Logger.Info("Poll answer received")
		return nil
	}
}

