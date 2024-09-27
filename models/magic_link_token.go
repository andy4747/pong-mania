package models

import "time"

type MagicLinkToken struct {
	ID        int64
	Email     string
	Token     string
	ExpiresAt *time.Time
}
