package utils

import "fmt"

func GetStaticMessage(messages map[string]string, key string) string {
	if msg, ok := messages[key]; ok {
		return msg
	}
	return fmt.Sprintf("[missing message for key: %s]", key)
}
