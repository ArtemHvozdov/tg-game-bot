package repo

import (
	"./github.com/ArtemHvozdov/tg-game-bot.git/models"
)

// Get all games with status "playing"
func GetAllActiveGames() ([]models.Game, error) {
	query := `SELECT id, name, current_task_id, total_players, status FROM games WHERE status = ?`
	rows, err := Db.Query(query, models.StatusGamePlaying)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"source": "Db: GetAllActiveGames",
			"error": err,
		}).Error("Failed to get all active games")
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(&game.ID, &game.Name, &game.CurrentTaskID, &game.TotalPlayers, &game.Status)
		if err != nil {
			utils.Logger.Errorf("Error scanning game: %v", err)
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}