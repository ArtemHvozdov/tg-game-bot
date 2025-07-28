package utils

import (
	"strings"

	
	//"math/rand"
	//"time"
)


// Функция для извлечения ID из инвайт-ссылки
func ExtractGameRoomID(link string) string {
	parts := strings.Split(link, "start=")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
