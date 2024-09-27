package middlewares

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"

	sentryecho "github.com/getsentry/sentry-go/echo"
)

func (m *Middlewares) SentryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// Execute the next handler
		err := next(ctx)
		if err != nil {
			// Capture the error with Sentry
			if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
				hub.Scope().SetTag("endpoint", ctx.Path())
				hub.Scope().SetTag("Request-IP", ctx.RealIP())
				hub.CaptureException(err)
			} else {
				sentry.CaptureException(err)
			}

			// Respond with a generic error message
			if !ctx.Response().Committed {
				ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
			}

		}
		return err
	}
}
