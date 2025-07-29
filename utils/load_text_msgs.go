package utils

import (
	"encoding/json"
	"os"
)

func LoadTextMessagges(path string) ([]string, error) {
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