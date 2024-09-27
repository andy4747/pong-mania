package middlewares

import (
	"pong-htmx/repository"
)

type Middlewares struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionsRepository
}

func NewMiddlewares(userRepo *repository.UserRepository, sessionRepo *repository.SessionsRepository) *Middlewares {
	return &Middlewares{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}
