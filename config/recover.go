package config

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func CustomRecover() middleware.LogErrorFunc {
	return func(c echo.Context, err error, stack []byte) error {
		formattedStack := formatStack(stack)

		log.Printf("PANIC RECOVER: %v\n%s", err, formattedStack)

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal Server Error",
		})
	}
}

func formatStack(stack []byte) string {
	lines := strings.Split(string(stack), "\n")
	var formattedLines []string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "goroutine") {
			formattedLines = append(formattedLines, "\n"+line)
			i++ // Skip the next line as it's usually empty
		} else if strings.Contains(line, ".go:") {
			formattedLines = append(formattedLines, line)
		}
	}

	return strings.Join(formattedLines, "\n")
}
