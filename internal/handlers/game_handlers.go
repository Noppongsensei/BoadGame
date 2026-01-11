package handlers

import (
	"encoding/json"

	"avalon/internal/services"

	"github.com/gofiber/fiber/v2"
)

// initGameHandler handles POST requests for initializing a game in a room
func initGameHandler(gameService *services.GameService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("roomId")

		// Parse request body
		var req struct {
			GameType string `json:"game_type"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate game type
		if req.GameType == "" {
			req.GameType = "avalon" // Default to avalon if not specified
		}

		// Initialize game
		if err := gameService.InitGame(roomID, req.GameType); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":   "Game initialized successfully",
			"room_id":   roomID,
			"game_type": req.GameType,
		})
	}
}

// getGameStateHandler handles GET requests for getting the current game state
func getGameStateHandler(gameService *services.GameService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("roomId")

		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Get game state filtered for the player (anti-cheat)
		gameState, err := gameService.GetFilteredGameState(roomID, userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Parse game state
		var stateMap map[string]interface{}
		if err := json.Unmarshal(gameState, &stateMap); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse game state",
			})
		}

		return c.Status(fiber.StatusOK).JSON(stateMap)
	}
}

// getGameHistoryHandler handles GET requests for getting game history
func getGameHistoryHandler(gameService *services.GameService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("roomId")

		// Get game history
		history, err := gameService.GetGameHistory(roomID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Parse history
		var historyArray []interface{}
		if err := json.Unmarshal(history, &historyArray); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse game history",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"history": historyArray,
		})
	}
}
