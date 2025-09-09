 package utils

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"path/filepath"

// 	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/qiuz_dna"
// )

// // LoadSubTask10 loads subtask 10 data from JSON file
// func LoadSubTask10(filename string) (*subtask10.Subtask10Data, error) {
// 	// Read file
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read subtask 10 file %s: %w", filename, err)
// 	}

// 	// Parse JSON
// 	var subtask10Data subtask10.Subtask10Data
// 	err = json.Unmarshal(data, &subtask10Data)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse subtask 10 JSON: %w", err)
// 	}

// 	// Validate data
// 	if len(subtask10Data.Subtasks) == 0 {
// 		return nil, fmt.Errorf("no subtasks found in file %s", filename)
// 	}

// 	// Validate each subtask
// 	for i, subtask := range subtask10Data.Subtasks {
// 		if len(subtask.Options) != 4 {
// 			return nil, fmt.Errorf("subtask %d must have exactly 4 options, got %d", subtask.ID, len(subtask.Options))
// 		}
// 		if len(subtask.Data) != 4 {
// 			return nil, fmt.Errorf("subtask %d must have exactly 4 data items, got %d", subtask.ID, len(subtask.Data))
// 		}
// 		if subtask.Title == "" {
// 			return nil, fmt.Errorf("subtask %d has empty title", subtask.ID)
// 		}
// 		if subtask.Image == "" {
// 			return nil, fmt.Errorf("subtask %d has empty image", subtask.ID)
// 		}

// 		// Verify image file exists
// 		imagePath := filepath.Join(filepath.Dir(filename), subtask.Image)
// 		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
// 			Logger.Warnf("Image file not found for subtask %d: %s", subtask.ID, imagePath)
// 		}

// 		Logger.Debugf("Loaded subtask %d: %s with %d options", 
// 			subtask.ID, subtask.Title, len(subtask.Options))
// 	}

// 	Logger.Infof("Successfully loaded %d subtasks from %s", len(subtask10Data.Subtasks), filename)
// 	return &subtask10Data, nil
// }
