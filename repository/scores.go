package repository

import (
	"database/sql"
	"pong-htmx/models"
)

type ScoresRepository struct {
	DB *sql.DB
}

func NewScoresRepository(db *sql.DB) *ScoresRepository {
	return &ScoresRepository{DB: db}
}

func (r *ScoresRepository) Create(score models.Score) error {
	_, err := r.DB.Exec("INSERT INTO scores (player1_id, player2_id, player1_score, player2_score, game_ended_at) VALUES ($1, $2, $3, $4, $5)",
		score.Player1ID, score.Player2ID, score.Player1Score, score.Player2Score, score.GameEndedAt)
	return err
}

func (r *ScoresRepository) GetAll() ([]models.Score, error) {
	rows, err := r.DB.Query("SELECT id, player1_id, player2_id, player1_score, player2_score, game_ended_at FROM scores")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []models.Score
	for rows.Next() {
		var score models.Score
		err := rows.Scan(&score.ID, &score.Player1ID, &score.Player2ID, &score.Player1Score, &score.Player2Score, &score.GameEndedAt)
		if err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return scores, nil
}
