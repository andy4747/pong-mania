package config

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
)

type ManiaLogger struct {
	env *Env
}

func NewManiaLogger(env *Env) *ManiaLogger {
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              env.SENTRY_DSN,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	return &ManiaLogger{
		env: env,
	}
}

func (ml *ManiaLogger) LogError(key string, err error, ctx *echo.Context) {
	if hub := sentryecho.GetHubFromContext(*ctx); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetTag(key, err.Error())
			scope.SetExtra(key, err.Error())
			hub.CaptureMessage(err.Error())
		})
	}
}
