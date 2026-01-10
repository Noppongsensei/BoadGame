package services

import (
	"errors"

	"avalon/internal/repositories"
)

// RoomService handles room-related business logic
type RoomService struct {
	roomRepo repositories.RoomRepository
	userRepo repositories.UserRepository
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo repositories.RoomRepository, userRepo repositories.UserRepository) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

// CreateRoom creates a new room
func (s *RoomService) CreateRoom(name string, hostID string, maxPlayers int) (*repositories.Room, error) {
	// Validate input
	if name == "" || hostID == "" {
		return nil, errors.New("name and hostID are required")
	}

	if maxPlayers < 5 || maxPlayers > 10 {
		return nil, errors.New("game requires 5-10 players")
	}

	// Check if host exists
	host, err := s.userRepo.GetByID(hostID)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, errors.New("host user not found")
	}

	// Create the room
	room, err := s.roomRepo.Create(name, hostID, maxPlayers)
	if err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom gets a room by ID
func (s *RoomService) GetRoom(id string) (*repositories.Room, error) {
	room, err := s.roomRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room not found")
	}

	return room, nil
}

// ListRooms lists all rooms with pagination
func (s *RoomService) ListRooms(limit, offset int) ([]*repositories.Room, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.roomRepo.List(limit, offset)
}

// ListOpenRooms lists all open rooms with pagination
func (s *RoomService) ListOpenRooms(limit, offset int) ([]*repositories.Room, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.roomRepo.ListOpenRooms(limit, offset)
}

// JoinRoom adds a player to a room
func (s *RoomService) JoinRoom(roomID, userID string) error {
	// Check if room exists and is open
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}
	if room.Status != repositories.RoomStatusOpen {
		return errors.New("room is not open")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if the room is full
	players, err := s.roomRepo.GetPlayers(roomID)
	if err != nil {
		return err
	}
	if len(players) >= room.MaxPlayers {
		return errors.New("room is full")
	}

	// Check if player is already in the room
	for _, p := range players {
		if p.ID == userID {
			return errors.New("user already in room")
		}
	}

	// Add player to room
	return s.roomRepo.AddPlayer(roomID, userID)
}

// LeaveRoom removes a player from a room
func (s *RoomService) LeaveRoom(roomID, userID string) error {
	// Check if room exists
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}

	// Check if player is in the room
	players, err := s.roomRepo.GetPlayers(roomID)
	if err != nil {
		return err
	}
	
	found := false
	for _, p := range players {
		if p.ID == userID {
			found = true
			break
		}
	}
	
	if !found {
		return errors.New("user not in room")
	}

	// If player is the host and game hasn't started, assign new host or delete room
	if userID == room.HostID && room.Status == repositories.RoomStatusOpen {
		// If there are other players, assign the first one as the new host
		var newHost string
		for _, p := range players {
			if p.ID != userID {
				newHost = p.ID
				break
			}
		}
		
		if newHost != "" {
			room.HostID = newHost
			if err := s.roomRepo.Update(room); err != nil {
				return err
			}
		} else {
			// If no other players, delete the room
			return s.roomRepo.Delete(roomID)
		}
	}

	// Remove player from room
	return s.roomRepo.RemovePlayer(roomID, userID)
}

// StartGame changes room status to playing
func (s *RoomService) StartGame(roomID, userID string) error {
	// Check if room exists
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}

	// Check if user is the host
	if room.HostID != userID {
		return errors.New("only the host can start the game")
	}

	// Check if room is open
	if room.Status != repositories.RoomStatusOpen {
		return errors.New("game is already in progress or finished")
	}

	// Check if there are enough players
	players, err := s.roomRepo.GetPlayers(roomID)
	if err != nil {
		return err
	}
	
	numPlayers := len(players)
	if numPlayers < 5 {
		return errors.New("at least 5 players required to start")
	}
	
	if numPlayers > room.MaxPlayers {
		return errors.New("too many players")
	}

	// Update room status to playing
	return s.roomRepo.UpdateStatus(roomID, repositories.RoomStatusPlaying)
}

// EndGame changes room status to finished
func (s *RoomService) EndGame(roomID string) error {
	// Check if room exists
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}

	// Check if room is playing
	if room.Status != repositories.RoomStatusPlaying {
		return errors.New("game is not in progress")
	}

	// Update room status to finished
	return s.roomRepo.UpdateStatus(roomID, repositories.RoomStatusFinished)
}
