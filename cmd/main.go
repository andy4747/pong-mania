package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pong-htmx/config"
	"pong-htmx/game"
	"pong-htmx/handlers"
	"pong-htmx/middlewares"
	"pong-htmx/repository"
	"pong-htmx/utils"
	"strings"
	"syscall"
	"time"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".exported_vars.env"); err != nil {
		log.Println("Warning: Error loading .env file")
	}

	env := config.NewEnv().InitOSEnv()

	// Check if we're in production or development
	isProduction := env.GOENV == "production"

	// Set up database connection
	db := config.PSQLConn(env)
	defer db.Close()

	// Set up Echo
	e := echo.New()

	// Set up migrations
	if _, err := config.PSQLMigrate(db, env); err != nil {
		log.Fatalf("Failed to set up migrations: %v", err)
	}

	//Initialize Mania Logger
	mLog := config.NewManiaLogger(env)

	// Initialize game stores
	gameStore := game.NewGameStore()
	connStore := game.NewConnStore()

	// Initialize repositories and services
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionsRepository(db)
	magicLinkRepo := repository.NewMagicLinkTokenRepository(db)
	scoreRepo := repository.NewScoresRepository(db)

	awsConfig := config.NewAWSConfig()
	s3Client := config.NewS3Client(awsConfig)
	sesClient := config.NewSESConfig(awsConfig)

	// Start session cleanup routine
	utils.StartSessionCleanupRoutine(sessionRepo)

	// Initialize configurations and handlers
	googleConfig := config.NewGoogleOAuth2Config(env)
	pagesHandler := handlers.NewPagesHandler(scoreRepo)
	authHandler := handlers.NewAuthHandler(googleConfig, userRepo, sessionRepo, magicLinkRepo, s3Client, sesClient, env, mLog)
	appMiddlewares := middlewares.NewMiddlewares(userRepo, sessionRepo)
	gamesHandler := handlers.NewGamesHandler(gameStore, scoreRepo, &upgrader, connStore)

	// Set up middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))

	e.Use(appMiddlewares.SentryMiddleware)

	if isProduction {
		// Production-specific middleware
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			HSTSMaxAge:            31536000,
			HSTSExcludeSubdomains: false,
			ContentSecurityPolicy: "default-src 'self'",
		}))
		e.Use(middleware.BodyLimit("2M"))
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
	} else {
		// Development-specific middleware
		e.Debug = true
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			userState := c.Get("user_state").(utils.UserState)
			requestLog := map[string]any{
				"REQUEST":  v.URI,
				"METHOD":   c.Request().Method,
				"STATUS":   v.Status,
				"USER_ID":  userState.UserID,
				"USERNAME": strings.Split(userState.Username, "@")[0],
			}
			jsonLog, _ := json.MarshalIndent(requestLog, "", "  ")
			fmt.Println(string(jsonLog))
			return nil
		},
	}))
	e.Use(middleware.CORS())
	e.Use(appMiddlewares.UserStateMiddleware())

	// Set up routes
	e.Static("/static", "static")
	e.GET("/ping", pagesHandler.Ping)
	e.GET("/", pagesHandler.Index)
	e.GET("/about", pagesHandler.About)
	e.GET("/contact", pagesHandler.Contact)
	e.GET("/rooms", gamesHandler.AdminRooms, appMiddlewares.AuthMiddleware())
	e.GET("/user/profile", pagesHandler.Profile, appMiddlewares.AuthMiddleware())
	e.POST("/user/profile", authHandler.UpdateProfile, appMiddlewares.AuthMiddleware())
	e.POST("/user/profile/image", authHandler.UpdateProfileImage, appMiddlewares.AuthMiddleware(), middlewares.TransactionMiddleware(db))
	e.GET("/user/profile/image", authHandler.ProfileImage, appMiddlewares.AuthMiddleware())
	e.GET("/game/random/start", gamesHandler.StartRandomGame, appMiddlewares.AuthMiddleware())
	e.GET("/wait", gamesHandler.WaitForPlayers, appMiddlewares.AuthMiddleware())
	e.GET("/wait/conn", gamesHandler.WaitForConn, appMiddlewares.AuthMiddleware())
	e.GET("/play/:room_id", gamesHandler.PlayGamePage, appMiddlewares.AuthMiddleware())
	e.GET("/play/conn/:room_id", gamesHandler.PlayConnHandler, appMiddlewares.AuthMiddleware())
	e.GET("/start", pagesHandler.StartGamePage, appMiddlewares.AuthMiddleware())
	e.GET("/score", pagesHandler.Score, appMiddlewares.AuthMiddleware())
	e.GET("/login", pagesHandler.Login)
	e.GET("/logout", authHandler.Logout)
	e.GET("/auth/google/login", authHandler.GoogleLogin)
	e.GET("/auth/google/callback", authHandler.GoogleLoginCallback)
	e.POST("/auth/magic/send", authHandler.SendMagicLink)
	e.GET("/auth/magic/verify", authHandler.VerifyMagicLinkTokenPage)
	e.GET("/auth/magic/login/:token", authHandler.LoginWithMagicLink)

	// Start server
	go func() {
		if err := e.Start(env.SERVER_PORT); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
