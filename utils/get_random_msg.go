package utils

import "math/rand"

func GetRandomMsg(messages []string) string {
	if len(messages) == 0 {
		return "Array of messages is empty"
	}
	return messages[rand.Intn(len(messages))]
}