package game

import (
	"sync"
)

type GameStore struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewGameStore() *GameStore {
	return &GameStore{
		rooms: make(map[string]*Room),
	}
}

func (gs *GameStore) AddRoom(room *Room) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.rooms[room.ID] = room
}

func (gs *GameStore) GetRoom(roomID string) (*Room, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	room, exists := gs.rooms[roomID]
	return room, exists
}

func (gs *GameStore) RemoveRoom(roomID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.rooms, roomID)
}

func (gs *GameStore) GetAvailableRoom() *Room {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	for _, room := range gs.rooms {
		if room.NumPlayers < 2 && room.Game.Player2.ID == -1 {
			return room
		}
	}
	return nil
}

func (gs *GameStore) CleanupFinishedGames() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for id, room := range gs.rooms {
		if room.IsFinished() {
			delete(gs.rooms, id)
		}
	}
}

func (gs *GameStore) GetAllRoomsPaginated(page, perPage int) ([]*Room, int) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	totalRooms := len(gs.rooms)
	totalPages := (totalRooms + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage
	if end > totalRooms {
		end = totalRooms
	}

	rooms := make([]*Room, 0, perPage)
	i := 0
	for _, room := range gs.rooms {
		if i >= start && i < end {
			rooms = append(rooms, room)
		}
		i++
		if i >= end {
			break
		}
	}

	return rooms, totalPages
}
