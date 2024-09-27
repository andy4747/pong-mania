package utils

import (
	"log"
	"pong-htmx/repository"
	"time"
)

func StartSessionCleanupRoutine(repo *repository.SessionsRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			err := repo.CleanupExpiredSessions()
			if err != nil {
				log.Println("Error cleaning up sessions: ", err)
			}
		}
	}()
}
