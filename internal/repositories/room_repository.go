package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Room status constants
const (
	RoomStatusOpen     = "waiting"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"
)

// Room represents a game room
type Room struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	HostID     string    `json:"host_id"`
	Status     string    `json:"status"`
	MaxPlayers int       `json:"max_players"`
	Players    []User    `json:"players,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RoomRepository defines the interface for room data access
type RoomRepository interface {
	Create(name string, hostID string, maxPlayers int) (*Room, error)
	GetByID(id string) (*Room, error)
	List(limit, offset int) ([]*Room, error)
	ListOpenRooms(limit, offset int) ([]*Room, error)
	Update(room *Room) error
	Delete(id string) error
	AddPlayer(roomID, userID string) error
	RemovePlayer(roomID, userID string) error
	GetPlayers(roomID string) ([]User, error)
	UpdateStatus(roomID, status string) error
}

// PostgresRoomRepository implements RoomRepository for PostgreSQL
type PostgresRoomRepository struct {
	db *PostgresDB
}

// NewRoomRepository creates a new PostgreSQL room repository
func NewRoomRepository(db *PostgresDB) RoomRepository {
	return &PostgresRoomRepository{db: db}
}

// Create adds a new room to the database
func (r *PostgresRoomRepository) Create(name string, hostID string, maxPlayers int) (*Room, error) {
	room := &Room{
		ID:         uuid.New().String(),
		Name:       name,
		HostID:     hostID,
		Status:     RoomStatusOpen,
		MaxPlayers: maxPlayers,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create the room
	_, err = tx.Exec(
		`INSERT INTO rooms (id, name, host_id, status, max_players, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		room.ID, room.Name, room.HostID, room.Status, room.MaxPlayers, room.CreatedAt, room.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Add the host as a player in the room
	_, err = tx.Exec(
		`INSERT INTO room_users (room_id, user_id) VALUES ($1, $2)`,
		room.ID, room.HostID,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return room, nil
}

// GetByID retrieves a room by ID
func (r *PostgresRoomRepository) GetByID(id string) (*Room, error) {
	var room Room
	err := r.db.QueryRow(
		`SELECT id, name, host_id, status, max_players, created_at, updated_at FROM rooms WHERE id = $1`,
		id,
	).Scan(&room.ID, &room.Name, &room.HostID, &room.Status, &room.MaxPlayers, &room.CreatedAt, &room.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get players in the room
	players, err := r.GetPlayers(room.ID)
	if err != nil {
		return nil, err
	}
	room.Players = players

	return &room, nil
}

// List retrieves a list of rooms with pagination
func (r *PostgresRoomRepository) List(limit, offset int) ([]*Room, error) {
	rows, err := r.db.Query(
		`SELECT id, name, host_id, status, max_players, created_at, updated_at 
		FROM rooms ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.HostID, &room.Status, &room.MaxPlayers, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

// ListOpenRooms retrieves a list of open rooms with pagination
func (r *PostgresRoomRepository) ListOpenRooms(limit, offset int) ([]*Room, error) {
	rows, err := r.db.Query(
		`SELECT id, name, host_id, status, max_players, created_at, updated_at 
		FROM rooms WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		RoomStatusOpen, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.HostID, &room.Status, &room.MaxPlayers, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

// Update updates room information
func (r *PostgresRoomRepository) Update(room *Room) error {
	room.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE rooms SET name = $1, host_id = $2, status = $3, max_players = $4, updated_at = $5 WHERE id = $6`,
		room.Name, room.HostID, room.Status, room.MaxPlayers, room.UpdatedAt, room.ID,
	)
	return err
}

// Delete removes a room from the database
func (r *PostgresRoomRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all player associations
	_, err = tx.Exec(`DELETE FROM room_users WHERE room_id = $1`, id)
	if err != nil {
		return err
	}

	// Delete the room
	_, err = tx.Exec(`DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AddPlayer adds a player to a room
func (r *PostgresRoomRepository) AddPlayer(roomID, userID string) error {
	_, err := r.db.Exec(
		`INSERT INTO room_users (room_id, user_id) VALUES ($1, $2)`,
		roomID, userID,
	)
	return err
}

// RemovePlayer removes a player from a room
func (r *PostgresRoomRepository) RemovePlayer(roomID, userID string) error {
	_, err := r.db.Exec(
		`DELETE FROM room_users WHERE room_id = $1 AND user_id = $2`,
		roomID, userID,
	)
	return err
}

// GetPlayers retrieves all players in a room
func (r *PostgresRoomRepository) GetPlayers(roomID string) ([]User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.username, u.created_at, u.updated_at
		FROM users u
		JOIN room_users ru ON u.id = ru.user_id
		WHERE ru.room_id = $1`,
		roomID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []User
	for rows.Next() {
		var player User
		err := rows.Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}

// UpdateStatus updates the status of a room
func (r *PostgresRoomRepository) UpdateStatus(roomID, status string) error {
	_, err := r.db.Exec(
		`UPDATE rooms SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), roomID,
	)
	return err
}
