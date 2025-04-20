package utils

import (
	"fmt"
	"strconv"
	"strings"
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
