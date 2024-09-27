package config

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewGoogleOAuth2Config(env *Env) *oauth2.Config {
	redirectUrl := "https://pongmania.com/auth/google/callback"
	return &oauth2.Config{
		RedirectURL:  redirectUrl,
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}
}
