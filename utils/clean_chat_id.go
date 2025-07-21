package utils

import (
	"strconv"
	"strings"
)

func CleanChatID(chatID int64) string {
	idStr := strconv.FormatInt(chatID, 10)
	return strings.TrimPrefix(idStr, "-100")
}