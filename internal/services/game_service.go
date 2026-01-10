package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"avalon/internal/core"
	"avalon/internal/repositories"
)

// GameService handles game-related business logic
type GameService struct {
	roomRepo        repositories.RoomRepository
	gameSessionRepo repositories.GameSessionRepository
	games           map[string]core.GameEngine // Map of roomID to game instance
}

// NewGameService creates a new game service
func NewGameService(roomRepo repositories.RoomRepository, gameSessionRepo repositories.GameSessionRepository) *GameService {
	return &GameService{
		roomRepo:        roomRepo,
		gameSessionRepo: gameSessionRepo,
		games:           make(map[string]core.GameEngine),
	}
}

// InitGame initializes a new game for a room
func (s *GameService) InitGame(roomID, gameType string) error {
	// Check if room exists and is in playing status
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}
	if room.Status != repositories.RoomStatusPlaying {
		return errors.New("room is not in playing status")
	}

	// Get players in the room
	players, err := s.roomRepo.GetPlayers(roomID)
	if err != nil {
		return err
	}
	if len(players) < 5 || len(players) > 10 {
		return errors.New("game requires 5-10 players")
	}

	// Convert to core.Player objects
	corePlayers := make([]core.Player, len(players))
	for i, p := range players {
		corePlayers[i] = core.Player{
			ID:       p.ID,
			Username: p.Username,
		}
	}

	// Create initial game state based on game type
	var gameEngine core.GameEngine
	var initialState json.RawMessage
	
	switch gameType {
	case "avalon":
		// Initialize Avalon game (will be implemented separately in avalon_game.go)
		avalonGame := NewAvalonGame()
		options := AvalonOptions{
			EnableMerlin:    true,
			EnablePercival:  true,
			EnableMorgana:   len(players) >= 7,
			EnableOberon:    len(players) >= 8,
			EnableMordred:   len(players) >= 9,
			UseStrictRules:  true,
		}
		
		optionsJson, _ := json.Marshal(options)
		if err := avalonGame.Init(corePlayers, optionsJson); err != nil {
			return err
		}
		
		gameEngine = avalonGame
		initialState = avalonGame.GetState()
	default:
		return fmt.Errorf("unsupported game type: %s", gameType)
	}
	
	// Save game state to database
	initialHistory := json.RawMessage("[]")
	gameSession, err := s.gameSessionRepo.Create(roomID, gameType, initialState, initialHistory)
	if err != nil {
		return err
	}

	// Store game instance in memory
	s.games[roomID] = gameEngine

	return nil
}

// ProcessGameAction processes a game action
func (s *GameService) ProcessGameAction(roomID, playerID string, actionType string, payload json.RawMessage) error {
	// Check if game exists for the room
	gameEngine, exists := s.games[roomID]
	if !exists {
		// Try to load from database
		if err := s.loadGame(roomID); err != nil {
			return err
		}
		gameEngine = s.games[roomID]
	}

	// Create action object
	action := core.Action{
		Type:    actionType,
		Payload: payload,
	}

	// Process the action
	if err := gameEngine.ProcessAction(playerID, action); err != nil {
		return err
	}

	// Check win condition
	isGameOver, winningTeam, winners := gameEngine.CheckWinCondition()

	// Update game state in database
	gameState := gameEngine.GetState()
	gameSession, err := s.gameSessionRepo.GetByRoomID(roomID)
	if err != nil {
		return err
	}
	
	if err := s.gameSessionRepo.UpdateGameState(gameSession.ID, gameState); err != nil {
		return err
	}

	// Update history with the action
	var history []map[string]interface{}
	if err := json.Unmarshal(gameSession.History, &history); err != nil {
		return err
	}
	
	historyEntry := map[string]interface{}{
		"playerID":   playerID,
		"actionType": actionType,
		"payload":    payload,
		"timestamp":  time.Now(),
	}
	
	history = append(history, historyEntry)
	historyJson, _ := json.Marshal(history)
	
	if err := s.gameSessionRepo.UpdateHistory(gameSession.ID, historyJson); err != nil {
		return err
	}

	// If game is over, update room status
	if isGameOver {
		// Add win information to history
		winInfo := map[string]interface{}{
			"type":        "game.over",
			"winningTeam": winningTeam,
			"winners":     winners,
			"timestamp":   time.Now(),
		}
		
		history = append(history, winInfo)
		historyJson, _ := json.Marshal(history)
		
		if err := s.gameSessionRepo.UpdateHistory(gameSession.ID, historyJson); err != nil {
			return err
		}
		
		// Update room status
		if err := s.roomRepo.UpdateStatus(roomID, repositories.RoomStatusFinished); err != nil {
			return err
		}
		
		// Remove game from memory
		delete(s.games, roomID)
	}

	return nil
}

// GetFilteredGameState returns the player-specific game state
func (s *GameService) GetFilteredGameState(roomID, playerID string) (json.RawMessage, error) {
	// Check if game exists for the room
	gameEngine, exists := s.games[roomID]
	if !exists {
		// Try to load from database
		if err := s.loadGame(roomID); err != nil {
			return nil, err
		}
		gameEngine = s.games[roomID]
	}

	// Get filtered state for the player (implements anti-cheat)
	return gameEngine.FilterStateForPlayer(playerID), nil
}

// GetGameHistory returns the game history
func (s *GameService) GetGameHistory(roomID string) (json.RawMessage, error) {
	gameSession, err := s.gameSessionRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}
	if gameSession == nil {
		return nil, errors.New("no game session found for room")
	}
	
	return gameSession.History, nil
}

// loadGame loads a game from the database
func (s *GameService) loadGame(roomID string) error {
	// Get room information
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}

	// Get game session
	gameSession, err := s.gameSessionRepo.GetByRoomID(roomID)
	if err != nil {
		return err
	}
	if gameSession == nil {
		return errors.New("no game session found for room")
	}

	// Get players in the room
	players, err := s.roomRepo.GetPlayers(roomID)
	if err != nil {
		return err
	}
	
	// Convert to core.Player objects
	corePlayers := make([]core.Player, len(players))
	for i, p := range players {
		corePlayers[i] = core.Player{
			ID:       p.ID,
			Username: p.Username,
		}
	}

	// Create game instance based on game type
	var gameEngine core.GameEngine
	
	switch gameSession.GameType {
	case "avalon":
		avalonGame := NewAvalonGame()
		// Load game state from database
		if err := avalonGame.LoadState(gameSession.GameState, corePlayers); err != nil {
			return err
		}
		gameEngine = avalonGame
	default:
		return fmt.Errorf("unsupported game type: %s", gameSession.GameType)
	}

	// Store game instance in memory
	s.games[roomID] = gameEngine
	
	return nil
}

// AvalonOptions represents the options for an Avalon game
type AvalonOptions struct {
	EnableMerlin    bool `json:"enable_merlin"`
	EnablePercival  bool `json:"enable_percival"`
	EnableMorgana   bool `json:"enable_morgana"`
	EnableOberon    bool `json:"enable_oberon"`
	EnableMordred   bool `json:"enable_mordred"`
	UseStrictRules  bool `json:"use_strict_rules"`
}
