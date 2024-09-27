package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pong-htmx/game"
	"pong-htmx/models"
	"pong-htmx/repository"
	"pong-htmx/utils"
	"pong-htmx/views/pages"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type GamesHandler struct {
	gameStore      *game.GameStore
	scoreRepo      *repository.ScoresRepository
	socketUpgrader *websocket.Upgrader
	connStore      *game.ConnStore
}

func NewGamesHandler(gameStore *game.GameStore, scoreRepo *repository.ScoresRepository, socketUpgrader *websocket.Upgrader, connStore *game.ConnStore) *GamesHandler {
	return &GamesHandler{
		gameStore:      gameStore,
		scoreRepo:      scoreRepo,
		socketUpgrader: socketUpgrader,
		connStore:      connStore,
	}
}

func (h *GamesHandler) StartRandomGame(c echo.Context) error {
	user := c.Get("user").(models.User)

	room := h.gameStore.GetAvailableRoom()
	if room == nil {
		room = game.NewRoom(user.ID, -1, h.scoreRepo)
		h.gameStore.AddRoom(room)
	} else {
		room.Game.Player2.ID = user.ID
	}

	ws, err := h.socketUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer ws.Close()

	data := map[string]any{
		"type":         "open",
		"data":         "client connected",
		"redirect_url": fmt.Sprintf("/wait?room_id=%s", room.ID),
	}

	jsonData, _ := json.Marshal(data)

	err = ws.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	h.connStore.AddClient(user.ID, ws)
	return nil
}

func (h *GamesHandler) WaitForPlayers(c echo.Context) error {
	roomID := c.QueryParam("room_id")
	if roomID == "" {
		return c.String(http.StatusBadRequest, "Room ID is required")
	}

	_, exists := h.gameStore.GetRoom(roomID)
	if !exists {
		return c.String(http.StatusNotFound, "Room not found")
	}

	return utils.Render(c, http.StatusOK, pages.Waiting(c.Request().Context(), roomID))
}

func (h *GamesHandler) WaitForConn(ctx echo.Context) error {
	roomID := ctx.QueryParam("room_id")
	if roomID == "" {
		return ctx.String(http.StatusBadRequest, "Room ID and User ID are required")
	}
	user := ctx.Get("user").(models.User)
	room, exists := h.gameStore.GetRoom(roomID)
	if !exists {
		return ctx.String(http.StatusNotFound, "Room not found")
	}

	ws, err := h.socketUpgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	h.connStore.AddClient(user.ID, ws)

	go room.Run()

	room.Register <- user.ID
	room.Ready <- user.ID

	go h.handleWebSocketConnection(ws, room, user.ID)

	return nil
}

func (h *GamesHandler) handleWebSocketConnection(ws *websocket.Conn, room *game.Room, userID int64) {
	defer ws.Close()
	defer h.connStore.RemoveClient(userID)

	redirectMessage := map[string]string{
		"type":         "redirect",
		"redirect_url": "/play/" + room.ID,
	}
	redirectJsonData, _ := json.Marshal(redirectMessage)

	waitingMessage := map[string]string{
		"type": "waiting",
		"data": "waiting for both players to connect",
	}
	waitJsonData, _ := json.Marshal(waitingMessage)

	for {
		if room.NumPlayers > 1 && room.ReadyPlayers > 1 {
			ws.WriteMessage(websocket.TextMessage, redirectJsonData)
			break
		}

		ws.WriteMessage(websocket.TextMessage, waitJsonData)
		time.Sleep(1 * time.Second)
	}
}

func (h *GamesHandler) PlayGamePage(c echo.Context) error {
	roomID := c.Param("room_id")
	room, exists := h.gameStore.GetRoom(roomID)
	if !exists {
		return c.String(http.StatusNotFound, "Room not found")
	}

	user := c.Get("user").(models.User)
	playerNumber, ok := room.Players[user.ID]
	if !ok {
		return c.String(http.StatusForbidden, "You are not a player in this game")
	}

	return utils.Render(c, http.StatusOK, pages.Play(c.Request().Context(), roomID, playerNumber))
}

func (h *GamesHandler) PlayConnHandler(ctx echo.Context) error {
	roomID := ctx.Param("room_id")
	room, exists := h.gameStore.GetRoom(roomID)
	if !exists {
		return ctx.String(http.StatusNotFound, "Room not found")
	}

	user := ctx.Get("user").(models.User)
	ws, err := h.socketUpgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	room.Ready <- user.ID
	h.connStore.AddClient(user.ID, ws)

	playerNumber := room.Players[user.ID]
	otherPlayerID := room.Game.Player1.ID
	if room.Game.Player1.ID == user.ID {
		otherPlayerID = room.Game.Player2.ID
	}

	// Start a goroutine to send game state updates
	go h.sendGameUpdates(ws, room)

	// Handle disconnection
	defer func() {
		room.Unregister <- user.ID
		h.connStore.RemoveClient(user.ID)
		ws.Close()
	}()

	// Keep the connection alive and handle incoming messages
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err.Error())
			break
		}

		var movement struct {
			Type string  `json:"type"`
			Y    float64 `json:"y"`
		}
		if err := json.Unmarshal(msg, &movement); err != nil {
			fmt.Println("Error unmarshaling message:", err.Error())
			continue
		}

		if movement.Type == "move" {
			room.Game.MovePlayer(playerNumber, movement.Y-room.Game.GetPlayerY(playerNumber))

			// Send updated game state to both players
			gameState := room.Game.ToJSON()
			h.sendToPlayer(user.ID, gameState)
			h.sendToPlayer(otherPlayerID, gameState)
		}
	}

	return nil
}

func (h *GamesHandler) sendGameUpdates(ws *websocket.Conn, room *game.Room) {
	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			room.Game.Update()
			gameState := room.Game.ToJSON()

			// Create a new map to hold the game state and type
			stateWithType := map[string]interface{}{
				"type":      "gameState",
				"gameState": json.RawMessage(gameState),
			}

			// Marshal the new map to JSON
			stateWithTypeJSON, err := json.Marshal(stateWithType)
			if err != nil {
				// Handle error
				continue
			}

			err = ws.WriteMessage(websocket.TextMessage, stateWithTypeJSON)
			if err != nil {
				fmt.Println("Error sending game update:", err.Error())
				return
			}
		case <-room.Unregister:
			return
		}
	}
}

func (h *GamesHandler) sendToPlayer(userID int64, message []byte) {
	ws := h.connStore.GetClient(userID)
	if ws != nil {
		err := ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Println("Error sending message to player:", err.Error())
		}
	}
}

func (h *GamesHandler) AdminRooms(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage := 5

	rooms, totalPages := h.gameStore.GetAllRoomsPaginated(page, perPage)
	return utils.Render(c, http.StatusOK, pages.Rooms(c.Request().Context(), rooms, page, totalPages))
}
