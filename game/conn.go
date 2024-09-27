package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ConnStore struct {
	clients map[int64]*websocket.Conn
	mu      sync.RWMutex
}

func NewConnStore() *ConnStore {
	return &ConnStore{
		clients: make(map[int64]*websocket.Conn),
	}
}

func (cs *ConnStore) AddClient(userID int64, conn *websocket.Conn) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.clients[userID] = conn
}

func (cs *ConnStore) GetClient(userID int64) *websocket.Conn {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	conn := cs.clients[userID]
	return conn
}

func (cs *ConnStore) RemoveClient(userID int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.clients, userID)
}

func (cs *ConnStore) GetAllClients() map[int64]*websocket.Conn {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	clientsCopy := make(map[int64]*websocket.Conn, len(cs.clients))
	for id, conn := range cs.clients {
		clientsCopy[id] = conn
	}
	return clientsCopy
}

func (cs *ConnStore) BroadcastMessage(message []byte) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	for user_id, conn := range cs.clients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			// Handle error (log it and remove the client)
			cs.RemoveClient(user_id)
		}
	}
}

func (cs *ConnStore) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.clients)
}

func (cs *ConnStore) Exists(userID int64) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	conn, exists := cs.clients[userID]
	if !exists {
		return false
	}

	// Check if the connection is still valid
	err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
	if err != nil {
		fmt.Println(err)
	}
	return err == nil
}
