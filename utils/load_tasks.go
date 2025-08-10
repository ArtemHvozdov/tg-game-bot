package utils

import (
	"encoding/json"
	"os"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
)

func LoadTasks(path string) ([]models.Task, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var tasks []models.Task
    err = json.Unmarshal(file, &tasks)
    if err != nil {
        return nil, err
    }

    return tasks, nil
}

func LoadSubTasks(path string) ([]string, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var subtasks []string
    err = json.Unmarshal(file, &subtasks)
    if err != nil {
        return nil, err
    }

    return subtasks, nil
}