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

// PollSession manages native Telegram poll voting for subtasks
// type PollSession struct {
// 	GameID      int
// 	Subtasks    []Subtask
// 	PollID      string        // ID созданного poll'а
// 	PollMessage *telebot.Message // Сообщение с poll'ом
// 	StartTime   time.Time
// 	IsActive    bool
// 	IsCompleted bool
// 	mutex       sync.RWMutex
// }

// Update PollSession version
type PollSession struct {
	GameID        int
	Subtasks      []Subtask
	PollID        string
	PollMessage   *telebot.Message
	MainTaskMsg   *telebot.Message // Добавлено: сообщение с основным текстом задания
	StartTime     time.Time
	IsActive      bool
	IsCompleted   bool
	mutex         sync.RWMutex
}

// PollManager manages poll sessions
type PollManager struct {
	sessions map[int]*PollSession    // gameID -> session
	pollGame map[string]int          // pollID -> gameID (для быстрого поиска)
	mutex    sync.RWMutex
}

var GlobalPollManager = &PollManager{
	sessions: make(map[int]*PollSession),
	pollGame: make(map[string]int),
}

// youtubeLinks := map[int]string{
// 		1: "https://www.youtube.com/shorts/Cb5ljmm3820?si=qWu3oIgz890cV0xo", // Lady Gaga - Paparazzi
// 		2: "https://www.youtube.com/watch?v=Mkuw7vdi-VA", // Phoebe Buffay - Smelly Cat
// 		3: "", // Для мемов YouTube не нужен
// 	}


func processTaskDescription(subtaskID int, description string) string {
	// Карта соответствий ID сабтаски к YouTube ссылкам
	youtubeLinks := map[int]string{
		1: "https://www.youtube.com/shorts/Cb5ljmm3820?si=qWu3oIgz890cV0xo", // Lady Gaga - Paparazzi
		2: "https://www.youtube.com/watch?v=Mkuw7vdi-VA", // Phoebe Buffay - Smelly Cat
		3: "", // Для мемов YouTube не нужен
	}

	// Обрабатываем каждую сабтаску индивидуально
	switch subtaskID {
	case 1: // Танцы под Lady Gaga
		processed := description
		if link, exists := youtubeLinks[subtaskID]; exists && link != "" {
			// Заменяем "цього" на ссылку
			processed = strings.Replace(processed, "цього", `<a href="https://www.youtube.com/shorts/Cb5ljmm3820?si=qWu3oIgz890cV0xo">цього</a>`, 1)
		}
		return processed

	case 2: // Пение Smelly Cat
		processed := description
		if link, exists := youtubeLinks[subtaskID]; exists && link != "" {
			// Заменяем "тут" на ссылку
			processed = strings.Replace(processed, "тут", `<a href="https://www.youtube.com/watch?v=Mkuw7vdi-VA">тут</a>`, 1)
		}

		// Ищем текст песни после эмодзи и двоеточия
		// Пробуем разные варианты поиска
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
			// Разделяем текст
			beforeSong := processed[:songTextStart+len("🥹:")] // только эмодзи и двоеточие
			songText := processed[songTextStart+patternLength:songTextEnd]
			afterSong := processed[songTextEnd:]

			// Очищаем начало текста песни от лишних символов
			songText = strings.TrimLeft(songText, "\\n\n ")

			// Разбиваем текст песни на строки
			songLines := strings.Split(songText, "\n")
			var formattedLines []string
			
			// Добавляем пустую строку
			formattedLines = append(formattedLines, "")
			
			for _, line := range songLines {
				line = strings.TrimSpace(line)
				if line != "" {
					formattedLines = append(formattedLines, "<em>"+line+"</em>")
				}
			}

			// Собираем обратно
			formattedSongText := strings.Join(formattedLines, "\n")
			processed = beforeSong + "\n" +formattedSongText + afterSong
		}
		
		return processed

	case 3: // Мемы - YouTube не нужен
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

// StartPollVoting - начинает голосование с использованием Telegram Poll
// StartPollVoting - обновленная версия с инициализацией новых полей
func (pm *PollManager) StartPollVoting(gameID int) (*PollSession, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Проверяем, есть ли активное голосование
	if session, exists := pm.sessions[gameID]; exists && session.IsActive {
		return nil, fmt.Errorf("voting already active for game %d", gameID)
	}

	// Загружаем варианты для голосования
	subtasks, err := loadSubtask5()
	if err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %v", err)
	}

	if len(subtasks) == 0 {
		return nil, fmt.Errorf("no subtasks found")
	}

	// Создаем новую сессию с обнуленными полями для сообщений
	session := &PollSession{
		GameID:      gameID,
		Subtasks:    subtasks,
		StartTime:   time.Now(),
		IsActive:    true,
		IsCompleted: false,
		MainTaskMsg: nil, // Будет установлено в CreateTelegramPoll
		PollMessage: nil, // Будет установлено в CreateTelegramPoll
	}

	pm.sessions[gameID] = session

	utils.Logger.Infof("Created poll session for game %d with %d options", gameID, len(subtasks))
	return session, nil
}

// CreateTelegramPoll - создает Telegram poll и отправляет его
func (pm *PollManager) CreateTelegramPoll(bot *telebot.Bot, chatID int64, session *PollSession, mainTaskText string, keyboard *telebot.ReplyMarkup) error {
	// Подготавливаем варианты ответов из titles
	var options []string
	for _, subtask := range session.Subtasks {
		options = append(options, subtask.Title)
	}

	question := "🗳️ Оберіть завдання для виконання:"
	
	// Создаем poll через метод Raw
	data := map[string]interface{}{
		"chat_id":                chatID,
		"question":               question,
		"options":                options,
		"is_anonymous":           false,
		"type":                   "regular",
		"allows_multiple_answers": false,
	}

	// Отправляем основное сообщение с заданием и СОХРАНЯЕМ его
	chat := &telebot.Chat{ID: chatID}
	mainTaskMsg, err := bot.Send(chat, mainTaskText, telebot.ModeMarkdown)
	if err != nil {
		return fmt.Errorf("failed to send main task message: %v", err)
	}

	// Сохраняем сообщение основного задания в сессии
	session.MainTaskMsg = mainTaskMsg

	time.Sleep(200 * time.Millisecond)
	
	resp, err := bot.Raw("sendPoll", data)
	if err != nil {
		return fmt.Errorf("failed to send poll via Raw API: %v", err)
	}

	// Обрабатываем ответ
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

	// Создаем объект Message для poll'а
	pollMsg := &telebot.Message{
		ID: result.Result.MessageID,
		Chat: &telebot.Chat{ID: chatID},
		Poll: &telebot.Poll{
			ID:       result.Result.Poll.ID,
			Question: result.Result.Poll.Question,
		},
	}

	// Сохраняем информацию о poll'е
	pm.mutex.Lock()
	session.PollID = pollMsg.Poll.ID
	session.PollMessage = pollMsg
	pm.pollGame[pollMsg.Poll.ID] = session.GameID
	pm.mutex.Unlock()

	utils.Logger.Infof("Created Telegram poll %s for game %d", pollMsg.Poll.ID, session.GameID)

	// Запускаем таймер на 30 секунд
	go pm.startPollTimer(bot, chatID, session, keyboard)

	return nil
}


// startPollTimer - запускает таймер на 30 секунд и завершает голосование
// startPollTimer - обновленная версия с передачей bot для API запросов
// startPollTimer - обновленная версия с передачей bot для API запросов
func (pm *PollManager) startPollTimer(bot *telebot.Bot, chatID int64, session *PollSession, keyboard *telebot.ReplyMarkup) {
	// time.Sleep(15 * time.Second)
	time.Sleep(1 * time.Minute)

	// НЕ вызываем bot.StopPoll здесь, так как это делается в ProcessPollResults
	utils.Logger.Infof("Poll timer expired for game %d, processing results...", session.GameID)

	// Обрабатываем результаты (stopPoll будет вызван внутри)
	winner, err := pm.ProcessPollResultsWithBot(bot, session.GameID)
	if err != nil {
		utils.Logger.Errorf("Failed to process poll results: %v", err)
		return
	}

	// Остальной код остается тот же...
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

	go pm.cleanupMessages(bot, session)

	utils.Logger.Infof("Poll voting completed for game %d, selected subtask ID: %d", session.GameID, winner.ID)
}

// ProcessPollResultsWithBot - версия с доступом к bot для API запросов
func (pm *PollManager) ProcessPollResultsWithBot(bot *telebot.Bot, gameID int) (*Subtask, error) {
	// Просто вызываем существующую функцию, но теперь у нас есть доступ к bot
	return pm.ProcessPollResults(bot, gameID)
}

// cleanupMessages - удаляет сообщения основного задания и голосования
func (pm *PollManager) cleanupMessages(bot *telebot.Bot, session *PollSession) {
	// Небольшая задержка, чтобы пользователи успели прочитать результат
	time.Sleep(3 * time.Second)

	// Удаляем сообщение с основным заданием
	if session.MainTaskMsg != nil {
		err := bot.Delete(session.MainTaskMsg)
		if err != nil {
			utils.Logger.Errorf("Failed to delete main task message: %v", err)
		} else {
			utils.Logger.Infof("Deleted main task message for game %d", session.GameID)
		}
	}

	// Удаляем сообщение с голосованием
	if session.PollMessage != nil {
		err := bot.Delete(session.PollMessage)
		if err != nil {
			utils.Logger.Errorf("Failed to delete poll message: %v", err)
		} else {
			utils.Logger.Infof("Deleted poll message for game %d", session.GameID)
		}
	}
}

// ProcessPollResults - обрабатывает результаты голосования и выбирает победителя
// ProcessPollResults - обрабатывает результаты голосования и выбирает победителя
// ProcessPollResults - обрабатывает результаты голосования и выбирает победителя
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

	// Получаем результаты через stopPoll response
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

	// Анализируем результаты голосования
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

	// Выбираем победителя (при равенстве - с наименьшим ID)
	var winner Subtask
	if len(winningIndexes) > 0 {
		winnerIndex := winningIndexes[0]
		winner = session.Subtasks[winnerIndex]

		// При равенстве выбираем с наименьшим ID
		for _, index := range winningIndexes[1:] {
			if session.Subtasks[index].ID < winner.ID {
				winner = session.Subtasks[index]
				winnerIndex = index
			}
		}

		utils.Logger.Infof("Winner: Option %d - Subtask %d ('%s') with %d votes",
			winnerIndex, winner.ID, winner.Title, maxVotes)
	} else {
		// Если никто не голосовал, выбираем первый вариант
		winner = session.Subtasks[0]
		utils.Logger.Infof("No votes received, selecting first option: Subtask %d ('%s')",
			winner.ID, winner.Title)
	}

	utils.Logger.Infof("Total votes in poll: %d", pollResult.Result.TotalVoterCount)
	utils.Logger.Infof("=== END POLL RESULTS ===")

	// Очищаем сессию
	delete(pm.pollGame, session.PollID)
	delete(pm.sessions, gameID)

	return &winner, nil
}

// fallbackProcessResults - резервный метод обработки результатов
func (pm *PollManager) fallbackProcessResults(session *PollSession, gameID int) (*Subtask, error) {
	utils.Logger.Infof("Using fallback method for poll results processing")
	
	// Если не можем получить результаты, возвращаем первый вариант
	if len(session.Subtasks) > 0 {
		winner := session.Subtasks[0]
		utils.Logger.Infof("Fallback: selecting first option: Subtask %d ('%s')", winner.ID, winner.Title)
		
		// Очищаем сессию
		delete(pm.pollGame, session.PollID)
		delete(pm.sessions, gameID)
		
		return &winner, nil
	}
	
	return nil, fmt.Errorf("no subtasks available for fallback")
}

// GetSessionByPollID - получает сессию по ID poll'а
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

// IsActive - проверяет, активно ли голосование для игры
func (pm *PollManager) IsActive(gameID int) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	session, exists := pm.sessions[gameID]
	return exists && session.IsActive
}

// ==========================================
// ПУБЛИЧНЫЕ ФУНКЦИИ ДЛЯ ИНТЕГРАЦИИ
// ==========================================

// StartSubtask5VotingDirect - запускает голосование напрямую (для вызова из SendTasks)
func StartSubtask5VotingDirect(bot *telebot.Bot, chatID int64, mainTaskText string, keyboard *telebot.ReplyMarkup) error {
	// Получаем игру из базы данных
	game, err := storage_db.GetGameByChatId(chatID)
	if err != nil {
		utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chatID, err)
		return fmt.Errorf("помилка отримання гри: %v", err)
	}

	// Создаем сессию голосования
	session, err := GlobalPollManager.StartPollVoting(game.ID)
	if err != nil {
		utils.Logger.Errorf("Failed to start poll voting: %v", err)
		return fmt.Errorf("помилка запуску голосування: %v", err)
	}

	// Создаем и отправляем Telegram poll
	err = GlobalPollManager.CreateTelegramPoll(bot, chatID, session, mainTaskText, keyboard)
	if err != nil {
		utils.Logger.Errorf("Failed to create Telegram poll: %v", err)
		return fmt.Errorf("помилка створення опитування: %v", err)
	}

	utils.Logger.Infof("Successfully started Telegram poll voting for game %d", game.ID)
	return nil
}

// HandlePollAnswer - обработчик ответов на poll (если нужна дополнительная логика)
func HandlePollAnswer(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		// Простое логирование без детальной обработки
		utils.Logger.Info("Poll answer received")
		return nil
	}
}

