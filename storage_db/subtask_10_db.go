// Functions for calculating and displaying subtask 10 results
package storage_db

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	//"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/sirupsen/logrus"
)

// Subtask10ResultItem represents vote count for an option
type Subtask10ResultItem struct {
	QuestionIndex  int    `json:"question_index"`
	QuestionID     int    `json:"question_id"`
	SelectedOption string `json:"selected_option"`
	VoteCount      int    `json:"vote_count"`
}

// GetSubtask2ResultsByGame retrieves all subtask 10 answers grouped by question for a specific game
func GetSubtask2ResultsByGame(gameID int) ([]Subtask10ResultItem, error) {
	query := `SELECT question_index, question_id, selected_option, COUNT(*) as vote_count 
			  FROM subtask_2_answers 
			  WHERE game_id = ? AND task_id = ? 
			  GROUP BY question_index, question_id, selected_option 
			  ORDER BY question_index ASC, vote_count DESC, selected_option ASC`

	rows, err := Db.Query(query, gameID, 2)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source":  "Db: GetSubtask2ResultsByGame",
			"game_id": gameID,
			"error":   err,
		}).Error("Failed to get subtask 2 results")
		return nil, err
	}
	defer rows.Close()

	var results []Subtask10ResultItem
	for rows.Next() {
		var result Subtask10ResultItem
		err := rows.Scan(&result.QuestionIndex, &result.QuestionID, &result.SelectedOption, &result.VoteCount)
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"source":  "Db: GetSubtask2ResultsByGame",
				"game_id": gameID,
				"error":   err,
			}).Error("Failed to scan subtask 10 result row")
			return nil, err
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source":  "Db: GetSubtask2ResultsByGame",
			"game_id": gameID,
			"error":   err,
		}).Error("Error iterating subtask 10 result rows")
		return nil, err
	}

	utils.Logger.WithFields(logrus.Fields{
		"source":       "Db: GetSubtask2ResultsByGame",
		"game_id":      gameID,
		"result_count": len(results),
	}).Info("Successfully retrieved subtask 2 results")

	return results, nil
}

// CalculateSubtask2Winners determines the winning option for each question
func CalculateSubtask2Winners(gameID int) ([]string, error) {
	results, err := GetSubtask2ResultsByGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask 2 results: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no subtask 2 results found for game %d", gameID)
	}

	// Group results by question index
	questionResults := make(map[int][]Subtask10ResultItem)
	for _, result := range results {
		questionResults[result.QuestionIndex] = append(questionResults[result.QuestionIndex], result)
	}

	var winners []string

	// Process each question in order
	questionIndices := make([]int, 0, len(questionResults))
	for questionIndex := range questionResults {
		questionIndices = append(questionIndices, questionIndex)
	}
	sort.Ints(questionIndices)

	for _, questionIndex := range questionIndices {
		options := questionResults[questionIndex]
		
		if len(options) == 0 {
			utils.Logger.Warnf("No options found for question %d", questionIndex)
			continue
		}

		// Sort options by vote count (descending), then by option name (ascending) for tiebreaking
		sort.Slice(options, func(i, j int) bool {
			if options[i].VoteCount == options[j].VoteCount {
				// If vote counts are equal, choose option with smaller second number
				return compareOptionsBySecondNumber(options[i].SelectedOption, options[j].SelectedOption)
			}
			return options[i].VoteCount > options[j].VoteCount
		})

		winner := options[0]
		winners = append(winners, winner.SelectedOption)

		utils.Logger.WithFields(logrus.Fields{
			"source":          "CalculateSubtask2Winners",
			"game_id":         gameID,
			"question_index":  questionIndex,
			"winning_option":  winner.SelectedOption,
			"vote_count":      winner.VoteCount,
			"total_options":   len(options),
		}).Info("Question winner determined")
	}

	utils.Logger.WithFields(logrus.Fields{
		"source":        "CalculateSubtask2Winners",
		"game_id":       gameID,
		"total_winners": len(winners),
		"winners":       winners,
	}).Info("All subtask 2 winners calculated")

	return winners, nil
}

// compareOptionsBySecondNumber compares two options by their second number for tiebreaking
// Returns true if option1 should come before option2
func compareOptionsBySecondNumber(option1, option2 string) bool {
	num1 := extractSecondNumber(option1)
	num2 := extractSecondNumber(option2)
	
	if num1 == num2 {
		// If second numbers are equal, fallback to alphabetical order
		return option1 < option2
	}
	
	return num1 < num2
}

// extractSecondNumber extracts the second number from option string like "04_02.png" -> 2
func extractSecondNumber(option string) int {
	// Remove file extension if present
	option = strings.TrimSuffix(option, ".png")
	
	// Split by underscore
	parts := strings.Split(option, "_")
	if len(parts) < 2 {
		return 0
	}
	
	// Parse second number
	if num, err := strconv.Atoi(parts[1]); err == nil {
		return num
	}
	
	return 0
}

// FormatSubtask10Results formats the winning results into a readable message
func FormatSubtask2Results(winners []string) string {
	if len(winners) == 0 {
		return "Результати підзавдання 2:\n\nНемає даних для відображення"
	}

	var message strings.Builder
	message.WriteString("🏆 Результати підзавдання 2:\n\n")

	for i, winner := range winners {
		message.WriteString(fmt.Sprintf("%d. %s\n", i+1, winner))
	}

	return message.String()
}

// ProcessSubtask2Results - main function to calculate and format results
func ProcessSubtask2Results(gameID int) (string, error) {
	winners, err := CalculateSubtask2Winners(gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source":  "ProcessSubtask2Results",
			"game_id": gameID,
			"error":   err,
		}).Error("Failed to calculate subtask 10 winners")
		return "", fmt.Errorf("failed to calculate winners: %w", err)
	}

	message := FormatSubtask2Results(winners)
	
	utils.Logger.WithFields(logrus.Fields{
		"source":        "ProcessSubtask2Results",
		"game_id":       gameID,
		"winners_count": len(winners),
	}).Info("Subtask 2 results processed successfully")

	return message, nil
}

// GetSubtask10WinnersArray returns array of winning image names for subtask 10
func GetSubtask2WinnersArray(gameID int) ([]string, error) {
	winners, err := CalculateSubtask2Winners(gameID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source":  "GetSubtask2WinnersArray",
			"game_id": gameID,
			"error":   err,
		}).Error("Failed to calculate subtask 2 winners")
		return nil, fmt.Errorf("failed to calculate winners: %w", err)
	}

	if len(winners) == 0 {
		utils.Logger.WithFields(logrus.Fields{
			"source":  "GetSubtask2WinnersArray",
			"game_id": gameID,
		}).Warn("No winners found for subtask 2")
		return nil, fmt.Errorf("no winners found for game %d", gameID)
	}

	// Convert .png extensions to .jpg for collage assets
	var winnersForCollage []string
	for _, winner := range winners {
		// Remove .png extension and add .jpg
		imageName := winner
		if len(imageName) >= 4 && imageName[len(imageName)-4:] == ".png" {
			imageName = imageName[:len(imageName)-4] + ".png"
		}
		winnersForCollage = append(winnersForCollage, imageName)
	}

	utils.Logger.WithFields(logrus.Fields{
		"source":          "GetSubtask2WinnersArray",
		"game_id":         gameID,
		"winners_count":   len(winnersForCollage),
		"winners":         winnersForCollage,
	}).Info("Successfully retrieved subtask 2 winners array")

	return winnersForCollage, nil
}