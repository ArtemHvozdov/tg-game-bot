package utils

import (
	"fmt"
	"strings"
	//"math/rand"
	//"time"
)

// Function generation unique invite link
// func GenerateInviteLink() string {
// 	rand.Seed(time.Now().UnixNano())
// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// 	linkLength := 10
// 	link := make([]byte, linkLength)
// 	for i := 0; i < linkLength; i++ {
// 		link[i] = charset[rand.Intn(len(charset))]
// 	}
// 	return "https://t.me/bestie_game_bot?start=" + string(link)
// }

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
