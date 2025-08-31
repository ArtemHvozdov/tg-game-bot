package quizdna

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

type Subtask struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Question     string   `json:"question"`
	FolderImages string   `json:"folderImages"`
	Options      []string `json:"options"`
	Images       []string `json:"images"`
}

type QuizNdaSession struct {
	GameID      int
	TaskID      int
	UserID      int64
	Username    string
	Questions   []Subtask
	CurrentStep int
	Answers     map[int]string // questionIndex -> selected option
	StartTime   time.Time
	IsCompleted bool
}

type QuizNdaSessionManager struct {
	sessions map[int]*QuizNdaSession // gameID -> session
	mutex    sync.RWMutex
}

var GlobalSessionManager = &QuizNdaSessionManager{
	sessions: make(map[int]*QuizNdaSession),
}

// Start new quiz dna session
func (qsm *QuizNdaSessionManager) StartSession(gameID, taskID int, userID int64, username string, questions []Subtask) error {
	qsm.mutex.Lock()
	defer qsm.mutex.Unlock()
	
	// Check if another session is active
	if session, exists := qsm.sessions[gameID]; exists && !session.IsCompleted {
		if session.UserID != userID {
			utils.Logger.Infof("Session already active for user %s", session.Username)
			return fmt.Errorf("session_active_other_user")
		}
	}
	
	qsm.sessions[gameID] = &QuizNdaSession{
		GameID:      gameID,
		TaskID:      taskID,
		UserID:      userID,
		Username:    username,
		Questions:   questions,
		CurrentStep: 0,
		Answers:     make(map[int]string),
		StartTime:   time.Now(),
		IsCompleted: false,
	}
	
	return nil
}

// Get active session
func (qsm *QuizNdaSessionManager) GetActiveSession(gameID int) (*QuizNdaSession, bool) {
	qsm.mutex.RLock()
	defer qsm.mutex.RUnlock()
	
	session, exists := qsm.sessions[gameID]
	return session, exists && !session.IsCompleted
}

// Save answer and move to next question
func (qsm *QuizNdaSessionManager) SaveAnswerAndNext(gameID int, selectedOption string, selectedOptionIndex int) (bool, error) {
	qsm.mutex.Lock()
	defer qsm.mutex.Unlock()
	
	session, exists := qsm.sessions[gameID]
	if !exists || session.IsCompleted {
		utils.Logger.Errorf("no active session")
		return false, fmt.Errorf("no active session")
	}
	
	// Save current answer
	session.Answers[session.CurrentStep] = selectedOption
	
	// TODO: Save to database
	err := qsm.saveAnswerToDB(session, selectedOption, selectedOptionIndex)
	if err != nil {
		utils.Logger.Errorf("Failed to save answer to DB: %v", err)
		return false, err
	}
	
	session.CurrentStep++
	
	// Check if all questions answered
	if session.CurrentStep >= len(session.Questions) {
		session.IsCompleted = true
		return true, nil // Session completed
	}
	
	return false, nil // More questions remain
}

// Complete and cleanup session
func (qsm *QuizNdaSessionManager) CompleteSession(gameID int) map[int]string {
	qsm.mutex.Lock()
	defer qsm.mutex.Unlock()
	
	session, exists := qsm.sessions[gameID]
	if !exists {
		return nil
	}
	
	answers := session.Answers
	delete(qsm.sessions, gameID)
	return answers
}

// Load subtasks from JSON and filter by existing images
func LoadSubtasks(taskID int) ([]Subtask, error) {
	filename := fmt.Sprintf("internal/data/tasks/subtasks/subtask_%d/subtask_%d.json", taskID, taskID)
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read subtasks file: %w", err)
	}
	
	var config struct {
		Subtasks []Subtask `json:"subtasks"`
	}
	
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subtasks JSON: %w", err)
	}
	
	// Filter subtasks by existing images
	var validSubtasks []Subtask
	basePath := fmt.Sprintf("internal/data/tasks/subtasks/subtask_%d/images", taskID)
	
	for _, subtask := range config.Subtasks {
		folderPath := filepath.Join(basePath, subtask.FolderImages)
		
		// Check if folder exists and has images
		if hasValidImages(folderPath, subtask.Images) {
			validSubtasks = append(validSubtasks, subtask)
			utils.Logger.Infof("Added subtask %d: %s (has images)", subtask.ID, subtask.Title)
		} else {
			utils.Logger.Infof("Skipped subtask %d: %s (no images found)", subtask.ID, subtask.Title)
		}
	}
	
	return validSubtasks, nil
}

// Check if folder has all required images
func hasValidImages(folderPath string, imageNames []string) bool {
	// Check if folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return false
	}
	
	// Check if all image files exist
	for _, imageName := range imageNames {
		imagePath := filepath.Join(folderPath, imageName)
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return false
		}
	}
	
	return len(imageNames) > 0
}

// Create inline keyboard for options with 2 buttons per row
func createOptionsKeyboard(subtask Subtask, gameID, taskID int) *telebot.ReplyMarkup {
	var rows [][]telebot.InlineButton
	
	// Process options in pairs
	for i := 0; i < len(subtask.Options); i += 2 {
		var row []telebot.InlineButton
		
		// First button in row
		btn1 := telebot.InlineButton{
			Text: subtask.Options[i],
			Data: fmt.Sprintf("quizdna_%d_%d_%d_%d", gameID, taskID, subtask.ID, i),
		}
		row = append(row, btn1)
		
		// Second button if exists
		if i+1 < len(subtask.Options) {
			btn2 := telebot.InlineButton{
				Text: subtask.Options[i+1],
				Data: fmt.Sprintf("quizdna_%d_%d_%d_%d", gameID, taskID, subtask.ID, i+1),
			}
			row = append(row, btn2)
		}
		
		rows = append(rows, row)
	}
	
	return &telebot.ReplyMarkup{InlineKeyboard: rows}
}

// Send current subtask with images and buttons
func SendCurrentSubtask(bot *telebot.Bot, c telebot.Context, gameID, taskID int) error {
	session, exists := GlobalSessionManager.GetActiveSession(gameID)
	if !exists {
		return c.Send("Сесія підзавдань не знайдена")
	}
	
	if session.CurrentStep >= len(session.Questions) {
		return c.Send("Всі питання завершені")
	}
	
	currentSubtask := session.Questions[session.CurrentStep]
	
	// Prepare message text
	// messageText := fmt.Sprintf("%s",
	// 	currentSubtask.Title,
	// 	//currentSubtask.Question,
    // )

    messageText := currentSubtask.Title

    questionText := currentSubtask.Question
	
	// Prepare media group with images
	mediaGroup, err := createMediaGroup(taskID, currentSubtask)
	if err != nil {
		utils.Logger.Errorf("Failed to create media group: %v", err)
		return c.Send("Помилка завантаження зображень")
	}
	
	// Create keyboard
	keyboard := createOptionsKeyboard(currentSubtask, gameID, taskID)
	
	if len(mediaGroup) > 0 {
		// Add caption to first image
		if photo, ok := mediaGroup[0].(*telebot.Photo); ok {
			photo.Caption = messageText
		}
		
		// Send album first
		err = c.SendAlbum(mediaGroup)
		if err != nil {
			utils.Logger.Errorf("Failed to send media group: %v", err)
			return c.Send("Помилка відправки зображень")
		}
		
		// Send keyboard as separate message (will be right after album)
		return c.Send(questionText, keyboard)
	} else {
		// If no images, send just text with keyboard
		return c.Send(messageText, keyboard)
	}
}



// Create media group from subtask images
func createMediaGroup(taskID int, subtask Subtask) (telebot.Album, error) {
	var mediaGroup telebot.Album
	basePath := fmt.Sprintf("internal/data/tasks/subtasks/subtask_%d/images/%s", taskID, subtask.FolderImages)
	
	for _, imageName := range subtask.Images {
		imagePath := filepath.Join(basePath, imageName)
		
		// Check if file exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			continue
		}
		
		// Create photo
		photo := &telebot.Photo{File: telebot.FromDisk(imagePath)}
		mediaGroup = append(mediaGroup, photo)
	}
	
	return mediaGroup, nil
}

// ============ PUBLIC API FUNCTIONS ============

// StartQuizDnaTask - основная функция для запуска Quiz DNA задания
// Вызывается из других пакетов при нажатии "Хочу відповісти"
func StartQuizDnaTask(bot *telebot.Bot) func(c telebot.Context)error {
	return func(c telebot.Context) error {
        utils.Logger.Info("StartQuizDnaTask called")
        taskID := 10
        userID := c.Sender().ID
        username := c.Sender().Username
        game, err := storage_db.GetGameByChatId(c.Chat().ID)
        if err != nil {
            utils.Logger.Errorf("Failed to get game ID: %v", err)
            return c.Send("Помилка отримання інформації про гру")
        }

        gameID := game.ID
        // Проверяем, не отвечает ли уже другой пользователь
        if session, exists := GlobalSessionManager.GetActiveSession(gameID); exists {
            if session.UserID != userID {
                return fmt.Errorf("other_user_active:%s", session.Username)
            }
            // Пользователь продолжает свою сессию - отправляем текущий вопрос
            return SendCurrentSubtask(bot, c, gameID, taskID)
        }
        
        // Загружаем подзадания
        subtasks, err := LoadSubtasks(taskID)
        if err != nil {
            utils.Logger.Errorf("Failed to load subtasks for task %d: %v", taskID, err)
            return fmt.Errorf("failed_to_load_subtasks")
        }
        
        if len(subtasks) == 0 {
            return fmt.Errorf("no_subtasks_available")
        }
        
        utils.Logger.Infof("Loaded %d subtasks for task %d", len(subtasks), taskID)
        
        // Запускаем новую сессию
        err = GlobalSessionManager.StartSession(gameID, taskID, userID, username, subtasks)
        if err != nil {
            utils.Logger.Errorf("Error starting QuizDNA session: %v", err)
            return err
        }
        
        // Отправляем первое подзадание
        return SendCurrentSubtask(bot, c, gameID, taskID)
    }
    
}

// HandleQuizDnaCallback - обработчик нажатий на кнопки ответов
func HandleQuizDnaCallback(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		callback := c.Callback()
		user := c.Sender()
		
		// Parse callback data: subtask_10_gameID_taskID_subtaskID_optionIndex
		var gameID, taskID, subtaskID, optionIndex int
		_, err := fmt.Sscanf(callback.Data, "subtask_10_%d_%d_%d_%d", &gameID, &taskID, &subtaskID, &optionIndex)
		if err != nil {
			utils.Logger.Errorf("Failed to parse callback data: %v", err)
			return c.Respond(&telebot.CallbackResponse{Text: "Помилка обробки відповіді"})
		}
		
		// Check if user has active session
		session, exists := GlobalSessionManager.GetActiveSession(gameID)
		if !exists {
			return c.Respond(&telebot.CallbackResponse{Text: "Сесія не знайдена"})
		}
		
		// Check if it's the right user
		if session.UserID != user.ID {
			return c.Respond(&telebot.CallbackResponse{Text: "Оу киця, почекай поки твоя подружка відповість"})
		}
		
		// Get selected option
		currentSubtask := session.Questions[session.CurrentStep]
		if optionIndex >= len(currentSubtask.Options) {
			return c.Respond(&telebot.CallbackResponse{Text: "Невірний варіант відповіді"})
		}
		
		selectedOption := currentSubtask.Options[optionIndex]
		
		// Save answer and check if session completed
		isCompleted, err := GlobalSessionManager.SaveAnswerAndNext(gameID, selectedOption, optionIndex)
		if err != nil {
			utils.Logger.Errorf("Error saving answer: %v", err)
			return c.Respond(&telebot.CallbackResponse{Text: "Помилка збереження відповіді"})
		}
		
		// Delete current message
		err = c.Delete()
		if err != nil {
			utils.Logger.Errorf("Failed to delete message: %v", err)
		}
		
		// Respond to callback
		err = c.Respond(&telebot.CallbackResponse{Text: fmt.Sprintf("Ви обрали: %s", selectedOption)})
		if err != nil {
			utils.Logger.Errorf("Failed to respond to callback: %v", err)
		}
		
		if isCompleted {
			// All subtasks completed
			GlobalSessionManager.CompleteSession(gameID)
			return c.Send("Дякую, кицю 😽 Коли всі подружки дадуть свої відповіді, я зроблю наш спільний фотоспогад!")
		} else {
			// Send next subtask
			return SendCurrentSubtask(bot, c, gameID, taskID)
		}
	}
}

// GetActiveSessionInfo - получить информацию об активной сессии
func GetActiveSessionInfo(gameID int) (bool, string, int) {
	session, exists := GlobalSessionManager.GetActiveSession(gameID)
	if !exists {
		return false, "", 0
	}
	return true, session.Username, session.CurrentStep + 1
}

// ForceCleanupSession - принудительная очистка сессии (для админов)
func ForceCleanupSession(gameID int) {
	GlobalSessionManager.CompleteSession(gameID)
}

// IsCallbackForQuizDna - проверяет, относится ли callback к Quiz DNA
func IsCallbackForQuizDna(callbackData string) bool {
	return len(callbackData) >= 10 && callbackData[:10] == "subtask_10"
}

// TODO: Implement database saving
func (qsm *QuizNdaSessionManager) saveAnswerToDB(session *QuizNdaSession, selectedOption string, selectedOptionIndex int) error {
	// Here should be database save logic
	// INSERT INTO subtask_image_answers (game_id, task_id, subtask_id, player_id, selected_option, selected_option_index)
	// VALUES (?, ?, ?, ?, ?, ?)
	
	utils.Logger.Infof("Saving answer: GameID=%d, TaskID=%d, SubtaskID=%d, PlayerID=%d, Option=%s, OptionIndex=%d",
		session.GameID, session.TaskID, session.Questions[session.CurrentStep-1].ID, 
		session.UserID, selectedOption, selectedOptionIndex)
	
	return nil
}
