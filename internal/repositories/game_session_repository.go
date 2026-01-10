package repositories

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GameSession represents a game session
type GameSession struct {
	ID        string          `json:"id"`
	RoomID    string          `json:"room_id"`
	GameType  string          `json:"game_type"`
	GameState json.RawMessage `json:"game_state"`
	History   json.RawMessage `json:"history"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GameSessionRepository defines the interface for game session data access
type GameSessionRepository interface {
	Create(roomID, gameType string, initialState, initialHistory json.RawMessage) (*GameSession, error)
	GetByID(id string) (*GameSession, error)
	GetByRoomID(roomID string) (*GameSession, error)
	UpdateGameState(id string, gameState json.RawMessage) error
	UpdateHistory(id string, history json.RawMessage) error
	Delete(id string) error
}

// PostgresGameSessionRepository implements GameSessionRepository for PostgreSQL
type PostgresGameSessionRepository struct {
	db *PostgresDB
}

// NewGameSessionRepository creates a new PostgreSQL game session repository
func NewGameSessionRepository(db *PostgresDB) GameSessionRepository {
	return &PostgresGameSessionRepository{db: db}
}

// Create adds a new game session to the database
func (r *PostgresGameSessionRepository) Create(roomID, gameType string, initialState, initialHistory json.RawMessage) (*GameSession, error) {
	// If initialHistory is nil, create an empty array
	if initialHistory == nil {
		initialHistory = json.RawMessage("[]")
	}

	session := &GameSession{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		GameType:  gameType,
		GameState: initialState,
		History:   initialHistory,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.db.Exec(
		`INSERT INTO game_sessions (id, room_id, game_type, game_state, history, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		session.ID, session.RoomID, session.GameType, session.GameState, session.History, 
		session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetByID retrieves a game session by ID
func (r *PostgresGameSessionRepository) GetByID(id string) (*GameSession, error) {
	var session GameSession
	err := r.db.QueryRow(
		`SELECT id, room_id, game_type, game_state, history, created_at, updated_at 
		FROM game_sessions WHERE id = $1`,
		id,
	).Scan(
		&session.ID, &session.RoomID, &session.GameType, &session.GameState, 
		&session.History, &session.CreatedAt, &session.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetByRoomID retrieves the latest game session for a room
func (r *PostgresGameSessionRepository) GetByRoomID(roomID string) (*GameSession, error) {
	var session GameSession
	err := r.db.QueryRow(
		`SELECT id, room_id, game_type, game_state, history, created_at, updated_at 
		FROM game_sessions WHERE room_id = $1 ORDER BY created_at DESC LIMIT 1`,
		roomID,
	).Scan(
		&session.ID, &session.RoomID, &session.GameType, &session.GameState, 
		&session.History, &session.CreatedAt, &session.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateGameState updates the game state of a session
func (r *PostgresGameSessionRepository) UpdateGameState(id string, gameState json.RawMessage) error {
	_, err := r.db.Exec(
		`UPDATE game_sessions SET game_state = $1, updated_at = $2 WHERE id = $3`,
		gameState, time.Now(), id,
	)
	return err
}

// UpdateHistory updates the history of a session
func (r *PostgresGameSessionRepository) UpdateHistory(id string, history json.RawMessage) error {
	_, err := r.db.Exec(
		`UPDATE game_sessions SET history = $1, updated_at = $2 WHERE id = $3`,
		history, time.Now(), id,
	)
	return err
}

// Delete removes a game session from the database
func (r *PostgresGameSessionRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM game_sessions WHERE id = $1`, id)
	return err
}
