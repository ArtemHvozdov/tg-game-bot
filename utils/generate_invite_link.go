package utils

import "fmt"

func GenerateInviteLink(userID int) string {
	return "https://t.me/bestie_game_bot?start=" + fmt.Sprintf("%d", userID)
}