package core

import "encoding/json"

// GameEngine defines the interface for any game implementation
// This follows the hexagonal architecture by decoupling game logic from infrastructure
type GameEngine interface {
	// Init initializes a new game with the given players and options
	Init(players []Player, options json.RawMessage) error

	// ProcessAction handles a player action and updates the game state
	ProcessAction(playerID string, action Action) error

	// CheckWinCondition evaluates if the game has a winner and returns the result
	CheckWinCondition() (bool, string, []string)

	// FilterStateForPlayer provides a player-specific view of the game state (Anti-Cheat)
	FilterStateForPlayer(playerID string) json.RawMessage

	// GetState returns the complete game state (for admin/debug purposes)
	GetState() json.RawMessage
}

// Player represents a player in any game
type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role,omitempty"`
}

// Action represents a game action performed by a player
type Action struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
