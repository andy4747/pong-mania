package utils

import (
	"os"
	"time"
)

const (
	GOOGLE_PROVIDER = "google"
	EMAIL_PROVIDER  = "Email"
)

var (
	COOKIE_NAME         = os.Getenv("COOKIE_NAME")
	SESSION_EXPIRY_TIME = time.Now().Add(24 * time.Hour)
)
