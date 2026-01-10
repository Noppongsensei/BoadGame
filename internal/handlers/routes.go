package handlers

import (
	"avalon/internal/services"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(
	app *fiber.App,
	userService *services.UserService,
	roomService *services.RoomService,
	gameService *services.GameService,
	hub *services.Hub,
) {
	// API routes
	api := app.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", registerHandler(userService))
	auth.Post("/login", loginHandler(userService))

	// User routes
	users := api.Group("/users")
	users.Get("/:id", getUserHandler(userService))
	users.Put("/:id", updateUserHandler(userService))
	users.Delete("/:id", deleteUserHandler(userService))

	// Room routes
	rooms := api.Group("/rooms")
	rooms.Post("/", createRoomHandler(roomService))
	rooms.Get("/", listRoomsHandler(roomService))
	rooms.Get("/open", listOpenRoomsHandler(roomService))
	rooms.Get("/:id", getRoomHandler(roomService))
	rooms.Post("/:id/join", joinRoomHandler(roomService, hub))
	rooms.Post("/:id/leave", leaveRoomHandler(roomService, hub))
	rooms.Post("/:id/start", startGameHandler(roomService, hub))
	rooms.Get("/:id/players", getRoomPlayersHandler(roomService))

	// Game routes
	games := api.Group("/games")
	games.Post("/:roomId/init", initGameHandler(gameService))
	games.Get("/:roomId/state", getGameStateHandler(gameService))
	games.Get("/:roomId/history", getGameHistoryHandler(gameService))

	// WebSocket route
	app.Get("/ws", websocketHandler(hub))
}
