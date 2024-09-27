package middlewares

import (
	"fmt"
	"net/http"
	"pong-htmx/models"
	"time"

	"github.com/labstack/echo/v4"
)

func (m *Middlewares) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session_token")
			if err != nil {
				fmt.Println(err)
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			sessionToken := cookie.Value

			session, err := m.sessionRepo.GetByToken(sessionToken)
			if err != nil || time.Now().After(*session.ExpiresAt) {
				// Session not found or expired
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			var user models.User
			if sessionToken != "" {
				session, err := m.sessionRepo.GetByToken(sessionToken)
				if err == nil {
					user, err = m.userRepo.GetByID(session.UserID)
					if err != nil {
						return c.Redirect(http.StatusSeeOther, "/login")
					}
				} else {
					return c.Redirect(http.StatusSeeOther, "/login")
				}
			}
			c.Set("user", user)
			return next(c)
		}
	}
}
