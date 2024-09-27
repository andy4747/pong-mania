package game

import (
	"encoding/json"
	"math"
	"math/rand"
	"time"
)

const (
	Width        = 800
	Height       = 400
	PaddleHeight = 100
	PaddleWidth  = 10
	BallSize     = 10
	BallSpeed    = 300
)

var (
	gameTime = 300 * time.Second
)

type Game struct {
	Player1        *Player
	Player2        *Player
	Ball           Ball
	LastUpdateTime time.Time
	StartTime      time.Time
	RemainingTime  time.Duration
}

type Player struct {
	Y     float64
	Score int64
	ID    int64
}

type Ball struct {
	X, Y   float64
	VX, VY float64
}

func NewGame(player1ID, player2ID int64) *Game {
	return &Game{
		Player1:        &Player{Y: Height/2 - PaddleHeight/2, ID: player1ID},
		Player2:        &Player{Y: Height/2 - PaddleHeight/2, ID: player2ID},
		Ball:           Ball{X: Width / 2, Y: Height / 2, VX: BallSpeed, VY: 0},
		LastUpdateTime: time.Now(),
		StartTime:      time.Now(),
	}
}

func (g *Game) Update() (gameEnded bool) {
	now := time.Now()
	dt := now.Sub(g.LastUpdateTime).Seconds()
	g.LastUpdateTime = now

	g.RemainingTime = gameTime - now.Sub(g.StartTime)

	if g.RemainingTime <= 0 {
		// Game has ended, store the final score in the database
		// and reset the game
		gameEnded = true
		return gameEnded
	}

	// Update ball position
	g.Ball.X += g.Ball.VX * dt
	g.Ball.Y += g.Ball.VY * dt

	// Ball collision with top and bottom walls
	if g.Ball.Y <= 0 || g.Ball.Y >= Height {
		g.Ball.VY = -g.Ball.VY
	}

	// Ball collision with paddles
	if g.Ball.X <= PaddleWidth && g.Ball.Y >= g.Player1.Y && g.Ball.Y <= g.Player1.Y+PaddleHeight {
		g.ballHitPaddle(g.Player1)
	}
	if g.Ball.X >= Width-PaddleWidth && g.Ball.Y >= g.Player2.Y && g.Ball.Y <= g.Player2.Y+PaddleHeight {
		g.ballHitPaddle(g.Player2)
	}

	// Score points
	if g.Ball.X <= 0 {
		g.Player2.Score++
		g.ResetBall()
	}
	if g.Ball.X >= Width {
		g.Player1.Score++
		g.ResetBall()
	}
	gameEnded = false
	return gameEnded
}

func (g *Game) ballHitPaddle(player *Player) {
	// Calculate the hit position relative to the paddle center
	relativeIntersectY := (player.Y + PaddleHeight/2) - g.Ball.Y

	// Normalize the relative intersection (-1 to 1)
	normalizedRelativeIntersectionY := relativeIntersectY / (PaddleHeight / 2)

	// Calculate the bounce angle (up to 75 degrees)
	bounceAngle := normalizedRelativeIntersectionY * (5 * math.Pi / 12) // 75 degrees in radians

	// Calculate new velocity components
	speed := math.Sqrt(g.Ball.VX*g.Ball.VX + g.Ball.VY*g.Ball.VY)
	g.Ball.VX = -g.Ball.VX // Reverse horizontal direction
	g.Ball.VY = speed * math.Sin(bounceAngle)

	// Ensure the horizontal speed doesn't become too low
	minHorizontalSpeed := 0.5 * BallSpeed
	if math.Abs(g.Ball.VX) < minHorizontalSpeed {
		g.Ball.VX = minHorizontalSpeed * math.Copysign(1, g.Ball.VX)
	}
}

func (g *Game) ResetBall() {
	g.Ball = Ball{
		X:  Width / 2,
		Y:  Height / 2,
		VX: BallSpeed * math.Copysign(1, rand.Float64()-0.5),
		VY: (rand.Float64() - 0.5) * BallSpeed,
	}
}

func (g *Game) MovePlayer(playerNumber int, dy float64) {
	var player *Player
	switch playerNumber {
	case 1:
		player = g.Player1
	case 2:
		player = g.Player2
	default:
		return
	}

	player.Y += dy
	if player.Y < 0 {
		player.Y = 0
	}
	if player.Y > Height-PaddleHeight {
		player.Y = Height - PaddleHeight
	}
}

func (g *Game) ToJSON() []byte {
	state := struct {
		Player1Y      float64 `json:"player1Y"`
		Player2Y      float64 `json:"player2Y"`
		BallX         float64 `json:"ballX"`
		BallY         float64 `json:"ballY"`
		Score1        int64   `json:"score1"`
		Score2        int64   `json:"score2"`
		RemainingTime int64   `json:"remainingTime"`
	}{
		Player1Y:      g.Player1.Y,
		Player2Y:      g.Player2.Y,
		BallX:         g.Ball.X,
		BallY:         g.Ball.Y,
		Score1:        g.Player1.Score,
		Score2:        g.Player2.Score,
		RemainingTime: int64(g.RemainingTime.Seconds()),
	}

	jsonData, _ := json.Marshal(state)
	return jsonData
}

func (g *Game) GetPlayerY(playerNumber int) float64 {
	switch playerNumber {
	case 1:
		return g.Player1.Y
	case 2:
		return g.Player2.Y
	default:
		return 0
	}
}
