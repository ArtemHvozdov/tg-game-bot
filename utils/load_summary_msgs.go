package utils

import (
	"encoding/json"
	"os"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
)

func LoadSummaryMsgs(path string) ([]models.SummaryMsg, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var tasks []models.SummaryMsg
    err = json.Unmarshal(file, &tasks)
    if err != nil {
        return nil, err
    }

    return tasks, nil
}