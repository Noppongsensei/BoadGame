package handlers

import (
	"strconv"
	"strings"
	
	"avalon/internal/services"
	
	"github.com/gofiber/fiber/v2"
)

// Room request/response structs
type CreateRoomRequest struct {
	Name       string `json:"name"`
	MaxPlayers int    `json:"max_players"`
}

// createRoomHandler handles POST requests for creating a new room
func createRoomHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authenticate request
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization token required",
			})
		}
		
		userID, err := validateJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}
		
		// Parse request body
		var req CreateRoomRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
		
		// Validate input
		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Room name is required",
			})
		}
		
		if req.MaxPlayers < 5 || req.MaxPlayers > 10 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Game requires 5-10 players",
			})
		}
		
		// Create room
		room, err := roomService.CreateRoom(req.Name, userID, req.MaxPlayers)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		
		return c.Status(fiber.StatusCreated).JSON(room)
	}
}

// listRoomsHandler handles GET requests for listing all rooms
func listRoomsHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		
		// Get rooms
		rooms, err := roomService.ListRooms(limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list rooms",
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"rooms":  rooms,
			"limit":  limit,
			"offset": offset,
		})
	}
}

// listOpenRoomsHandler handles GET requests for listing open rooms
func listOpenRoomsHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		
		// Get open rooms
		rooms, err := roomService.ListOpenRooms(limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list open rooms",
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"rooms":  rooms,
			"limit":  limit,
			"offset": offset,
		})
	}
}

// getRoomHandler handles GET requests for getting a room by ID
func getRoomHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("id")
		
		// Get room
		room, err := roomService.GetRoom(roomID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Room not found",
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(room)
	}
}

// joinRoomHandler handles POST requests for joining a room
func joinRoomHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("id")
		
		// Authenticate request
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization token required",
			})
		}
		
		userID, err := validateJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}
		
		// Join room
		if err := roomService.JoinRoom(roomID, userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		
		// Get updated room information
		room, err := roomService.GetRoom(roomID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get room information",
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(room)
	}
}

// leaveRoomHandler handles POST requests for leaving a room
func leaveRoomHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("id")
		
		// Authenticate request
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization token required",
			})
		}
		
		userID, err := validateJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}
		
		// Leave room
		if err := roomService.LeaveRoom(roomID, userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Successfully left the room",
		})
	}
}

// startGameHandler handles POST requests for starting a game in a room
func startGameHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("id")
		
		// Authenticate request
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization token required",
			})
		}
		
		userID, err := validateJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}
		
		// Start game
		if err := roomService.StartGame(roomID, userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		
		// Get updated room information
		room, err := roomService.GetRoom(roomID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get room information",
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(room)
	}
}

// getRoomPlayersHandler handles GET requests for getting players in a room
func getRoomPlayersHandler(roomService *services.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get room ID from URL
		roomID := c.Params("id")
		
		// Get room
		room, err := roomService.GetRoom(roomID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Room not found",
			})
		}
		
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"players": room.Players,
		})
	}
}
