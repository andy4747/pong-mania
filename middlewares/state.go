package middlewares

import (
	"context"
	"fmt"
	"pong-htmx/utils"

	"github.com/labstack/echo/v4"
)

func (m *Middlewares) UserStateMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session_token")
			var sessionToken string
			if err == nil {
				sessionToken = cookie.Value
			}

			state := utils.UserState{
				IsAuthenticated: false,
				CurrentPath:     c.Request().URL.Path,
			}

			if sessionToken != "" {
				session, err := m.sessionRepo.GetByToken(sessionToken)
				if err == nil {
					user, err := m.userRepo.GetByID(session.UserID)
					if err == nil {
						state.IsAuthenticated = true
						state.Username = user.Username
						state.UserID = user.ID
						state.ImageUrl = fmt.Sprintf("/user/profile/image?image_key=%s", user.ImageUrl)
						state.Email = user.Email
						state.Error = ""
					}
				}
			} else {
				state.Error = "user not authenticated"
			}

			ctx := context.WithValue(c.Request().Context(), utils.UserStateKey{}, state)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set("user_state", state)
			return next(c)
		}
	}
}
