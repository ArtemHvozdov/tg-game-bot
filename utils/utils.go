package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	
	//"math/rand"
	//"time"
)


func GenerateInviteLink(gameRoomID int) string {
	return "https://t.me/bestie_game_bot?start=" + fmt.Sprintf("%d", gameRoomID)
}


// Функция для извлечения ID из инвайт-ссылки
func ExtractGameRoomID(link string) string {
	parts := strings.Split(link, "start=")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func GetWaitingTaskID(status string) (int, error) {
	if !strings.HasPrefix(status, "waiting_") {
		return 0, fmt.Errorf("status does not start with 'waiting_'")
	}

	parts := strings.Split(status, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid status format")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid task ID: %v", err)
	}

	return id, nil
}

func GetSkipTaskID(status string) (int, error) {
	if !strings.HasPrefix(status, "skip_") {
		return 0, fmt.Errorf("status does not start with 'skip_'")
	}

	parts := strings.Split(status, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid status format")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid task ID: %v", err)
	}

	return id, nil
}

func LoadTasks(path string) ([]models.Task, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var tasks []models.Task
    err = json.Unmarshal(file, &tasks)
    if err != nil {
        return nil, err
    }

    return tasks, nil
}

func LoadJoinMessagges(path string) ([]string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var messages []string
	err = json.Unmarshal(file, &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func CleanChatID(chatID int64) string {
	idStr := strconv.FormatInt(chatID, 10)
	return strings.TrimPrefix(idStr, "-100")
}