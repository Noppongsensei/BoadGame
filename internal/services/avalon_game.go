package services

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"avalon/internal/core"
)

// Role constants
const (
	RoleMerlin     = "merlin"
	RolePercival   = "percival"
	RoleLoyal      = "loyal"
	RoleMorgana    = "morgana"
	RoleMordred    = "mordred"
	RoleAssassin   = "assassin"
	RoleOberon     = "oberon"
	RoleMinion     = "minion"
)

// Team constants
const (
	TeamGood = "good"
	TeamEvil = "evil"
)

// Action constants
const (
	ActionSelectQuest    = "select_quest"
	ActionVoteQuest      = "vote_quest"
	ActionPerformQuest   = "perform_quest"
	ActionAssassinateMerlin = "assassinate_merlin"
)

// AvalonState represents the game state for Avalon
type AvalonState struct {
	Players       []AvalonPlayer       `json:"players"`
	CurrentRound  int                  `json:"current_round"`
	Leader        int                  `json:"leader"` // Index of current leader
	QuestTracker  []AvalonQuest        `json:"quest_tracker"`
	CurrentQuest  *AvalonQuest         `json:"current_quest,omitempty"`
	VoteTrack     int                  `json:"vote_track"`  // Failed team proposals in current round
	GameOver      bool                 `json:"game_over"`
	WinningTeam   string               `json:"winning_team,omitempty"`
	Winners       []string             `json:"winners,omitempty"`
	Phase         string               `json:"phase"` // setup, quest_selection, quest_voting, quest_performance, assassination, game_over
	Options       AvalonOptions        `json:"options"`
	QuestRequirements []QuestRequirement `json:"quest_requirements"`
}

// AvalonPlayer extends the core Player with Avalon-specific fields
type AvalonPlayer struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Team     string `json:"team"`
	IsLeader bool   `json:"is_leader"`
}

// AvalonQuest represents a quest in the game
type AvalonQuest struct {
	Round           int      `json:"round"`
	RequiredPlayers int      `json:"required_players"`
	SelectedPlayers []string `json:"selected_players,omitempty"`
	Votes           map[string]bool `json:"votes,omitempty"` // playerID -> approve (true) or reject (false)
	Results         []bool   `json:"results,omitempty"` // true = success, false = fail
	Status          string   `json:"status"` // pending, approved, rejected, success, fail
}

// QuestRequirement defines the number of players required for quests based on player count
type QuestRequirement struct {
	PlayerCount      int `json:"player_count"`
	Quest1Players    int `json:"quest1_players"`
	Quest2Players    int `json:"quest2_players"`
	Quest3Players    int `json:"quest3_players"`
	Quest4Players    int `json:"quest4_players"`
	Quest5Players    int `json:"quest5_players"`
	FailsRequiredForQuest4 int `json:"fails_required_for_quest4"`
}

// AvalonGame implements the GameEngine interface for Avalon
type AvalonGame struct {
	state AvalonState
	rand  *rand.Rand // For deterministic randomness in tests
}

// NewAvalonGame creates a new Avalon game
func NewAvalonGame() *AvalonGame {
	src := rand.NewSource(time.Now().UnixNano())
	return &AvalonGame{
		rand: rand.New(src),
	}
}

// Init initializes a new Avalon game
func (g *AvalonGame) Init(players []core.Player, options json.RawMessage) error {
	if len(players) < 5 || len(players) > 10 {
		return errors.New("avalon requires 5-10 players")
	}

	// Parse options
	var avalonOptions AvalonOptions
	if err := json.Unmarshal(options, &avalonOptions); err != nil {
		return err
	}

	// Initialize state
	g.state = AvalonState{
		CurrentRound: 1,
		Leader:       g.rand.Intn(len(players)), // Random starting leader
		VoteTrack:    0,
		GameOver:     false,
		Phase:        "setup",
		Options:      avalonOptions,
		QuestRequirements: g.getQuestRequirements(),
	}

	// Initialize players
	g.state.Players = make([]AvalonPlayer, len(players))
	for i, p := range players {
		g.state.Players[i] = AvalonPlayer{
			ID:       p.ID,
			Username: p.Username,
			IsLeader: i == g.state.Leader,
		}
	}

	// Assign roles
	g.assignRoles()

	// Initialize quest tracker
	g.initQuestTracker()

	// Move to first phase
	g.state.Phase = "quest_selection"

	return nil
}

// ProcessAction handles a player action
func (g *AvalonGame) ProcessAction(playerID string, action core.Action) error {
	if g.state.GameOver {
		return errors.New("game is over")
	}

	playerIndex := g.findPlayerIndex(playerID)
	if playerIndex == -1 {
		return errors.New("player not found")
	}

	switch action.Type {
	case ActionSelectQuest:
		return g.handleQuestSelection(playerID, action.Payload)
	case ActionVoteQuest:
		return g.handleQuestVote(playerID, action.Payload)
	case ActionPerformQuest:
		return g.handleQuestPerformance(playerID, action.Payload)
	case ActionAssassinateMerlin:
		return g.handleAssassination(playerID, action.Payload)
	default:
		return errors.New("unknown action type")
	}
}

// CheckWinCondition evaluates if the game has a winner
func (g *AvalonGame) CheckWinCondition() (bool, string, []string) {
	if !g.state.GameOver {
		return false, "", nil
	}
	return true, g.state.WinningTeam, g.state.Winners
}

// FilterStateForPlayer provides a player-specific view of the game state
func (g *AvalonGame) FilterStateForPlayer(playerID string) json.RawMessage {
	playerIndex := g.findPlayerIndex(playerID)
	if playerIndex == -1 {
		// Player not in game, return limited info
		limitedState := g.getPublicGameState()
		jsonState, _ := json.Marshal(limitedState)
		return jsonState
	}

	// Clone the state to avoid modifying the original
	filteredState := g.state

	// Filter information based on the player's role
	playerRole := g.state.Players[playerIndex].Role
	// playerTeam := g.state.Players[playerIndex].Team - ตัดออกเพื่อแก้ปัญหา unused variable

	// Filter player roles - each player only knows their own role and what their role allows them to know
	for i := range filteredState.Players {
		if i == playerIndex {
			// Player knows their own role
			continue
		}

		otherPlayerRole := g.state.Players[i].Role
		otherPlayerTeam := g.state.Players[i].Team

		// Hide role information based on player's role
		switch playerRole {
		case RoleMerlin:
			// Merlin knows all evil players except Mordred
			if otherPlayerRole == RoleMordred {
				filteredState.Players[i].Role = ""
				filteredState.Players[i].Team = ""
			} else if otherPlayerTeam == TeamEvil {
				// Merlin knows evil team but not specific roles
				filteredState.Players[i].Team = TeamEvil
				filteredState.Players[i].Role = ""
			} else {
				filteredState.Players[i].Role = ""
				filteredState.Players[i].Team = ""
			}
		case RolePercival:
			// Percival knows Merlin and Morgana but can't tell them apart
			if otherPlayerRole == RoleMerlin || otherPlayerRole == RoleMorgana {
				filteredState.Players[i].Role = "merlin_or_morgana"
				filteredState.Players[i].Team = ""
			} else {
				filteredState.Players[i].Role = ""
				filteredState.Players[i].Team = ""
			}
		case RoleAssassin, RoleMinion, RoleMorgana:
			// Evil players (except Oberon) know all evil players but not specific roles
			if otherPlayerTeam == TeamEvil {
				filteredState.Players[i].Team = TeamEvil
				if otherPlayerRole != RoleOberon {
					filteredState.Players[i].Role = ""
				} else {
					// Don't know Oberon
					filteredState.Players[i].Role = ""
					filteredState.Players[i].Team = ""
				}
			} else {
				filteredState.Players[i].Role = ""
				filteredState.Players[i].Team = ""
			}
		case RoleMordred:
			// Mordred knows all evil players except Oberon
			if otherPlayerTeam == TeamEvil {
				filteredState.Players[i].Team = TeamEvil
				if otherPlayerRole != RoleOberon {
					filteredState.Players[i].Role = ""
				} else {
					// Don't know Oberon
					filteredState.Players[i].Role = ""
					filteredState.Players[i].Team = ""
				}
			} else {
				filteredState.Players[i].Role = ""
				filteredState.Players[i].Team = ""
			}
		case RoleOberon:
			// Oberon knows no one, not even other evil players
			filteredState.Players[i].Role = ""
			filteredState.Players[i].Team = ""
		default:
			// Regular loyal servants know nothing
			filteredState.Players[i].Role = ""
			filteredState.Players[i].Team = ""
		}
	}

	// Filter quest votes - only show the player's own vote
	if filteredState.CurrentQuest != nil && filteredState.CurrentQuest.Votes != nil {
		playerVotes := make(map[string]bool)
		if vote, exists := filteredState.CurrentQuest.Votes[playerID]; exists {
			playerVotes[playerID] = vote
		}
		filteredState.CurrentQuest.Votes = playerVotes
	}

	// Filter quest results - only show the results, not who succeeded/failed
	// This is already handled as results are anonymous

	jsonState, _ := json.Marshal(filteredState)
	return jsonState
}

// GetState returns the complete game state
func (g *AvalonGame) GetState() json.RawMessage {
	jsonState, _ := json.Marshal(g.state)
	return jsonState
}

// LoadState loads a game state from JSON
func (g *AvalonGame) LoadState(state json.RawMessage, players []core.Player) error {
	if err := json.Unmarshal(state, &g.state); err != nil {
		return err
	}
	return nil
}

// Private helper methods

// findPlayerIndex finds the index of a player by ID
func (g *AvalonGame) findPlayerIndex(playerID string) int {
	for i, p := range g.state.Players {
		if p.ID == playerID {
			return i
		}
	}
	return -1
}

// assignRoles assigns roles to players
func (g *AvalonGame) assignRoles() {
	playerCount := len(g.state.Players)
	
	// Determine number of evil players based on player count
	// According to official Avalon rules
	evilCount := playerCount / 3
	if playerCount == 9 {
		evilCount = 3
	} else if playerCount == 10 {
		evilCount = 4
	}

	// Create indices and shuffle
	indices := make([]int, playerCount)
	for i := range indices {
		indices[i] = i
	}
	g.shuffle(indices)

	// Assign good team roles first
	goodIndices := indices[:playerCount-evilCount]
	evilIndices := indices[playerCount-evilCount:]

	// Assign Merlin and Percival if enabled
	merlinIndex := -1
	percivalIndex := -1
	
	if g.state.Options.EnableMerlin && len(goodIndices) > 0 {
		merlinIndex = goodIndices[0]
		g.state.Players[merlinIndex].Role = RoleMerlin
		g.state.Players[merlinIndex].Team = TeamGood
		goodIndices = goodIndices[1:]
	}
	
	if g.state.Options.EnablePercival && len(goodIndices) > 0 {
		percivalIndex = goodIndices[0]
		g.state.Players[percivalIndex].Role = RolePercival
		g.state.Players[percivalIndex].Team = TeamGood
		goodIndices = goodIndices[1:]
	}

	// Assign remaining good players as loyal servants
	for _, idx := range goodIndices {
		g.state.Players[idx].Role = RoleLoyal
		g.state.Players[idx].Team = TeamGood
	}

	// Assign evil team roles
	// Always have an Assassin
	assassinIndex := evilIndices[0]
	g.state.Players[assassinIndex].Role = RoleAssassin
	g.state.Players[assassinIndex].Team = TeamEvil
	evilIndices = evilIndices[1:]

	// Assign optional evil roles
	if g.state.Options.EnableMorgana && len(evilIndices) > 0 {
		morganaIndex := evilIndices[0]
		g.state.Players[morganaIndex].Role = RoleMorgana
		g.state.Players[morganaIndex].Team = TeamEvil
		evilIndices = evilIndices[1:]
	}
	
	if g.state.Options.EnableMordred && len(evilIndices) > 0 {
		mordredIndex := evilIndices[0]
		g.state.Players[mordredIndex].Role = RoleMordred
		g.state.Players[mordredIndex].Team = TeamEvil
		evilIndices = evilIndices[1:]
	}
	
	if g.state.Options.EnableOberon && len(evilIndices) > 0 {
		oberonIndex := evilIndices[0]
		g.state.Players[oberonIndex].Role = RoleOberon
		g.state.Players[oberonIndex].Team = TeamEvil
		evilIndices = evilIndices[1:]
	}

	// Assign remaining evil players as generic minions
	for _, idx := range evilIndices {
		g.state.Players[idx].Role = RoleMinion
		g.state.Players[idx].Team = TeamEvil
	}
}

// initQuestTracker initializes the quest tracker based on number of players
func (g *AvalonGame) initQuestTracker() {
	playerCount := len(g.state.Players)
	
	// Find appropriate quest requirements
	var questReq QuestRequirement
	for _, req := range g.state.QuestRequirements {
		if req.PlayerCount == playerCount {
			questReq = req
			break
		}
	}
	
	// Initialize quest tracker
	g.state.QuestTracker = []AvalonQuest{
		{Round: 1, RequiredPlayers: questReq.Quest1Players, Status: "pending"},
		{Round: 2, RequiredPlayers: questReq.Quest2Players, Status: "pending"},
		{Round: 3, RequiredPlayers: questReq.Quest3Players, Status: "pending"},
		{Round: 4, RequiredPlayers: questReq.Quest4Players, Status: "pending"},
		{Round: 5, RequiredPlayers: questReq.Quest5Players, Status: "pending"},
	}
}

// handleQuestSelection handles the selection of players for a quest
func (g *AvalonGame) handleQuestSelection(playerID string, payload json.RawMessage) error {
	// Check if it's quest selection phase
	if g.state.Phase != "quest_selection" {
		return errors.New("not in quest selection phase")
	}

	// Check if player is the leader
	playerIndex := g.findPlayerIndex(playerID)
	if playerIndex == -1 || !g.state.Players[playerIndex].IsLeader {
		return errors.New("only the leader can select quest team")
	}

	// Parse selected players
	var selection struct {
		SelectedPlayers []string `json:"selected_players"`
	}
	
	if err := json.Unmarshal(payload, &selection); err != nil {
		return err
	}

	// Validate selection count
	currentQuest := g.state.QuestTracker[g.state.CurrentRound-1]
	if len(selection.SelectedPlayers) != currentQuest.RequiredPlayers {
		return errors.New("incorrect number of players selected for quest")
	}

	// Validate that selected players exist
	for _, id := range selection.SelectedPlayers {
		if g.findPlayerIndex(id) == -1 {
			return errors.New("invalid player selected")
		}
	}

	// Update current quest
	currentQuest.SelectedPlayers = selection.SelectedPlayers
	currentQuest.Votes = make(map[string]bool)
	currentQuest.Status = "voting"
	
	g.state.CurrentQuest = &currentQuest
	g.state.QuestTracker[g.state.CurrentRound-1] = currentQuest
	
	// Move to voting phase
	g.state.Phase = "quest_voting"
	
	return nil
}

// handleQuestVote handles a vote for a proposed quest team
func (g *AvalonGame) handleQuestVote(playerID string, payload json.RawMessage) error {
	// Check if it's quest voting phase
	if g.state.Phase != "quest_voting" {
		return errors.New("not in quest voting phase")
	}

	// Check if player exists
	playerIndex := g.findPlayerIndex(playerID)
	if playerIndex == -1 {
		return errors.New("player not found")
	}

	// Check if player has already voted
	if _, voted := g.state.CurrentQuest.Votes[playerID]; voted {
		return errors.New("player has already voted")
	}

	// Parse vote
	var vote struct {
		Approve bool `json:"approve"`
	}
	
	if err := json.Unmarshal(payload, &vote); err != nil {
		return err
	}

	// Record vote
	g.state.CurrentQuest.Votes[playerID] = vote.Approve
	
	// Check if all players have voted
	if len(g.state.CurrentQuest.Votes) == len(g.state.Players) {
		// Count votes
		approveCount := 0
		for _, approve := range g.state.CurrentQuest.Votes {
			if approve {
				approveCount++
			}
		}
		
		// Majority needed to approve
		if approveCount > len(g.state.Players)/2 {
			// Team approved, move to quest performance
			g.state.CurrentQuest.Status = "approved"
			g.state.VoteTrack = 0
			g.state.Phase = "quest_performance"
		} else {
			// Team rejected
			g.state.CurrentQuest.Status = "rejected"
			g.state.VoteTrack++
			
			// Check if vote track reached 5 (automatic evil win)
			if g.state.VoteTrack >= 5 {
				g.state.GameOver = true
				g.state.WinningTeam = TeamEvil
				g.setWinners(TeamEvil)
				g.state.Phase = "game_over"
				return nil
			}
			
			// Move leader token
			g.advanceLeader()
			
			// Reset for new proposal
			g.state.CurrentQuest = nil
			g.state.Phase = "quest_selection"
		}
		
		// Update quest tracker
		g.state.QuestTracker[g.state.CurrentRound-1] = *g.state.CurrentQuest
	}
	
	return nil
}

// handleQuestPerformance handles the performance of a quest by the selected team
func (g *AvalonGame) handleQuestPerformance(playerID string, payload json.RawMessage) error {
	// Check if it's quest performance phase
	if g.state.Phase != "quest_performance" {
		return errors.New("not in quest performance phase")
	}

	// Check if player exists and is part of the quest
	playerIndex := g.findPlayerIndex(playerID)
	if playerIndex == -1 {
		return errors.New("player not found")
	}

	isOnQuest := false
	for _, id := range g.state.CurrentQuest.SelectedPlayers {
		if id == playerID {
			isOnQuest = true
			break
		}
	}
	if !isOnQuest {
		return errors.New("player not selected for quest")
	}

	// Parse quest action
	var action struct {
		Success bool `json:"success"`
	}
	
	if err := json.Unmarshal(payload, &action); err != nil {
		return err
	}

	// Only evil players can fail quests
	if !action.Success {
		if g.state.Players[playerIndex].Team != TeamEvil {
			return errors.New("good team players cannot fail quests")
		}
	}

	// Add result
	g.state.CurrentQuest.Results = append(g.state.CurrentQuest.Results, action.Success)
	
	// Check if all selected players have performed the quest
	if len(g.state.CurrentQuest.Results) == len(g.state.CurrentQuest.SelectedPlayers) {
		// Count failures
		failCount := 0
		for _, success := range g.state.CurrentQuest.Results {
			if !success {
				failCount++
			}
		}
		
		// Determine if quest was successful or failed
		// Special rule for quest 4 with 7+ players - requires 2 fails
		failsRequired := 1
		if g.state.CurrentRound == 4 && len(g.state.Players) >= 7 {
			// Check specific requirement for this player count
			for _, req := range g.state.QuestRequirements {
				if req.PlayerCount == len(g.state.Players) {
					failsRequired = req.FailsRequiredForQuest4
					break
				}
			}
		}
		
		if failCount >= failsRequired {
			g.state.CurrentQuest.Status = "fail"
		} else {
			g.state.CurrentQuest.Status = "success"
		}
		
		// Update quest tracker
		g.state.QuestTracker[g.state.CurrentRound-1] = *g.state.CurrentQuest
		
		// Check win conditions
		successCount := 0
		failCount = 0
		for _, quest := range g.state.QuestTracker {
			if quest.Status == "success" {
				successCount++
			} else if quest.Status == "fail" {
				failCount++
			}
		}
		
		if successCount >= 3 {
			// Good team has won 3 quests, move to assassination phase
			g.state.Phase = "assassination"
		} else if failCount >= 3 {
			// Evil team has won 3 quests, game over
			g.state.GameOver = true
			g.state.WinningTeam = TeamEvil
			g.setWinners(TeamEvil)
			g.state.Phase = "game_over"
		} else {
			// Move to next round
			g.advanceRound()
		}
	}
	
	return nil
}

// handleAssassination handles the assassination attempt on Merlin
func (g *AvalonGame) handleAssassination(playerID string, payload json.RawMessage) error {
	// Check if it's assassination phase
	if g.state.Phase != "assassination" {
		return errors.New("not in assassination phase")
	}

	// Check if player is the assassin
	playerIndex := g.findPlayerIndex(playerID)
	if playerIndex == -1 || g.state.Players[playerIndex].Role != RoleAssassin {
		return errors.New("only the assassin can perform assassination")
	}

	// Parse assassination target
	var target struct {
		TargetID string `json:"target_id"`
	}
	
	if err := json.Unmarshal(payload, &target); err != nil {
		return err
	}

	// Check if target exists
	targetIndex := g.findPlayerIndex(target.TargetID)
	if targetIndex == -1 {
		return errors.New("target player not found")
	}

	// Determine winner based on whether Merlin was correctly identified
	if g.state.Players[targetIndex].Role == RoleMerlin {
		// Assassin correctly identified Merlin, evil wins
		g.state.GameOver = true
		g.state.WinningTeam = TeamEvil
		g.setWinners(TeamEvil)
	} else {
		// Assassin failed, good wins
		g.state.GameOver = true
		g.state.WinningTeam = TeamGood
		g.setWinners(TeamGood)
	}
	
	g.state.Phase = "game_over"
	
	return nil
}

// advanceLeader moves the leader token to the next player
func (g *AvalonGame) advanceLeader() {
	// Remove leader status from current leader
	g.state.Players[g.state.Leader].IsLeader = false
	
	// Advance leader
	g.state.Leader = (g.state.Leader + 1) % len(g.state.Players)
	
	// Set new leader
	g.state.Players[g.state.Leader].IsLeader = true
}

// advanceRound moves to the next round
func (g *AvalonGame) advanceRound() {
	g.state.CurrentRound++
	g.state.CurrentQuest = nil
	g.state.VoteTrack = 0
	
	// Move leader token
	g.advanceLeader()
	
	// Move to quest selection phase
	g.state.Phase = "quest_selection"
}

// setWinners determines which players are winners based on the winning team
func (g *AvalonGame) setWinners(winningTeam string) {
	var winners []string
	for _, p := range g.state.Players {
		if p.Team == winningTeam {
			winners = append(winners, p.ID)
		}
	}
	g.state.Winners = winners
}

// getPublicGameState returns a public view of the game state with no hidden information
func (g *AvalonGame) getPublicGameState() AvalonState {
	publicState := g.state
	
	// Remove all roles and team information
	for i := range publicState.Players {
		publicState.Players[i].Role = ""
		publicState.Players[i].Team = ""
	}
	
	// Only include quest vote counts, not individual votes
	if publicState.CurrentQuest != nil && publicState.CurrentQuest.Votes != nil {
		publicState.CurrentQuest.Votes = nil
	}
	
	return publicState
}

// getQuestRequirements returns the quest requirements based on player count
func (g *AvalonGame) getQuestRequirements() []QuestRequirement {
	return []QuestRequirement{
		{
			PlayerCount:      5,
			Quest1Players:    2,
			Quest2Players:    3,
			Quest3Players:    2,
			Quest4Players:    3,
			Quest5Players:    3,
			FailsRequiredForQuest4: 1,
		},
		{
			PlayerCount:      6,
			Quest1Players:    2,
			Quest2Players:    3,
			Quest3Players:    4,
			Quest4Players:    3,
			Quest5Players:    4,
			FailsRequiredForQuest4: 1,
		},
		{
			PlayerCount:      7,
			Quest1Players:    2,
			Quest2Players:    3,
			Quest3Players:    3,
			Quest4Players:    4,
			Quest5Players:    4,
			FailsRequiredForQuest4: 2,
		},
		{
			PlayerCount:      8,
			Quest1Players:    3,
			Quest2Players:    4,
			Quest3Players:    4,
			Quest4Players:    5,
			Quest5Players:    5,
			FailsRequiredForQuest4: 2,
		},
		{
			PlayerCount:      9,
			Quest1Players:    3,
			Quest2Players:    4,
			Quest3Players:    4,
			Quest4Players:    5,
			Quest5Players:    5,
			FailsRequiredForQuest4: 2,
		},
		{
			PlayerCount:      10,
			Quest1Players:    3,
			Quest2Players:    4,
			Quest3Players:    4,
			Quest4Players:    5,
			Quest5Players:    5,
			FailsRequiredForQuest4: 2,
		},
	}
}

// shuffle randomly permutes the elements of the slice
func (g *AvalonGame) shuffle(slice []int) {
	for i := len(slice) - 1; i > 0; i-- {
		j := g.rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
