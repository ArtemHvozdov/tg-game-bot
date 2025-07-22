package utils

import (
	"encoding/json"
	"os"
)

// LoadSingleMessage загружает одно сообщение из файла JSON
func LoadSingleMessage(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	var message string
	if err := json.Unmarshal(data, &message); err != nil {
		return "", err
	}
	return message, nil
}