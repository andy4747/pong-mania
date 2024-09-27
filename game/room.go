package game

import (
	"encoding/json"
	"pong-htmx/models"
	"pong-htmx/repository"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID              string
	Game            *Game
	Players         map[int64]int
	Register        chan int64
	Ready           chan int64
	Unregister      chan int64
	Broadcast       chan []byte
	ScoreRepo       *repository.ScoresRepository
	NumPlayers      int
	ReadyPlayers    int
	GameStartTime   time.Time
	CountdownTicker *time.Ticker
}

func NewRoom(player1ID, player2ID int64, scoreRepo *repository.ScoresRepository) *Room {
	return &Room{
		ID:           generateRoomID(),
		Game:         NewGame(player1ID, player2ID),
		Players:      make(map[int64]int),
		Register:     make(chan int64),
		Ready:        make(chan int64),
		Unregister:   make(chan int64),
		Broadcast:    make(chan []byte),
		ScoreRepo:    scoreRepo,
		NumPlayers:   0,
		ReadyPlayers: 0,
	}
}

func (r *Room) Run() {
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		select {
		case playerID := <-r.Register:
			r.registerPlayer(playerID)
		case playerID := <-r.Ready:
			r.playerReady(playerID)
		case playerID := <-r.Unregister:
			r.unregisterPlayer(playerID)
		case <-ticker.C:
			r.update()
		}
	}
}

func (r *Room) registerPlayer(playerID int64) {
	if r.NumPlayers < 2 {
		r.NumPlayers++
		r.Players[playerID] = r.NumPlayers
	}
}

func (r *Room) playerReady(playerID int64) {
	if _, ok := r.Players[playerID]; ok {
		if r.ReadyPlayers < 2 {
			r.ReadyPlayers++
		}
		if r.ReadyPlayers == 2 {
			r.startCountdown()
		}
	}
}

func (r *Room) startCountdown() {
	r.CountdownTicker = time.NewTicker(time.Second)
	countdown := 15

	go func() {
		for range r.CountdownTicker.C {
			if countdown > 0 {
				r.broadcastCountdown(countdown)
				countdown--
			} else {
				r.CountdownTicker.Stop()
				r.startGame()
				return
			}
		}
	}()
}

func (r *Room) broadcastCountdown(seconds int) {
	message := map[string]interface{}{
		"type":      "countdown",
		"countdown": seconds,
	}
	jsonMessage, _ := json.Marshal(message)
	r.Broadcast <- jsonMessage
}

func (r *Room) startGame() {
	r.GameStartTime = time.Now()
	message := map[string]interface{}{
		"type":        "gameStart",
		"gameStarted": true,
	}
	jsonMessage, _ := json.Marshal(message)
	r.Broadcast <- jsonMessage
}

func (r *Room) update() {
	if r.Game.Player1 != nil && r.Game.Player2 != nil && !r.GameStartTime.IsZero() {
		gameEnded := r.Game.Update()
		if gameEnded {
			err := r.ScoreRepo.Create(models.Score{
				Player1ID:    r.Game.Player1.ID,
				Player2ID:    r.Game.Player2.ID,
				Player1Score: r.Game.Player1.Score,
				Player2Score: r.Game.Player2.Score,
			})
			if err != nil {
				panic(err.Error())
			}
		}
		state := r.Game.ToJSON()
		r.Broadcast <- state

		if time.Since(r.GameStartTime) >= gameTime {
			r.endGame()
		}
	}
}

func (r *Room) endGame() {
	score := models.Score{
		Player1ID:    r.Game.Player1.ID,
		Player2ID:    r.Game.Player2.ID,
		Player1Score: r.Game.Player1.Score,
		Player2Score: r.Game.Player2.Score,
		GameEndedAt:  time.Now(),
	}
	r.ScoreRepo.Create(score)

	message := map[string]interface{}{
		"type":      "gameEnd",
		"gameEnded": true,
		"score1":    r.Game.Player1.Score,
		"score2":    r.Game.Player2.Score,
	}
	jsonMessage, _ := json.Marshal(message)
	r.Broadcast <- jsonMessage
}

func generateRoomID() string {
	room_uuid := uuid.NewString()
	return room_uuid
}

func (r *Room) IsFinished() bool {
	return time.Since(r.GameStartTime) >= gameTime
}

func (r *Room) unregisterPlayer(playerID int64) {
	if playerNumber, exists := r.Players[playerID]; exists {
		delete(r.Players, playerID)
		r.NumPlayers--
		r.ReadyPlayers--

		// Reset the corresponding player in the game
		if playerNumber == 1 {
			r.Game.Player1 = &Player{Y: Height/2 - PaddleHeight/2, ID: -1}
		} else if playerNumber == 2 {
			r.Game.Player2 = &Player{Y: Height/2 - PaddleHeight/2, ID: -1}
		}

		// If all players have left, end the game
		if r.NumPlayers == 0 {
			r.endGame()
		}
	}
}
