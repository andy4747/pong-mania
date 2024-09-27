package models

import (
	"time"
)

type User struct {
	ID         int64
	Email      string
	Username   string
	Provider   string
	ImageUrl   string
	IsActive   bool
	IsVerified bool
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}
