package models

import "time"

type Score struct {
	ID           int64
	Player1ID    int64
	Player2ID    int64
	Player1Score int64
	Player2Score int64
	GameEndedAt  time.Time
}
