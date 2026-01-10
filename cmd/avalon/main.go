package main

import (
	"fmt"
	"log"
	"os"

	"avalon/internal/handlers"
	"avalon/internal/repositories"
	"avalon/internal/services"
	
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Setup database connection
	db, err := repositories.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Setup repositories
	userRepo := repositories.NewUserRepository(db)
	roomRepo := repositories.NewRoomRepository(db)
	gameSessionRepo := repositories.NewGameSessionRepository(db)

	// Setup services
	gameService := services.NewGameService(roomRepo, gameSessionRepo)
	userService := services.NewUserService(userRepo)
	roomService := services.NewRoomService(roomRepo, userRepo)

	// Setup websocket hub
	hub := services.NewHub()
	go hub.Run()

	// Setup Fiber app
	app := fiber.New()
	
	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Setup handlers
	handlers.SetupRoutes(app, userService, roomService, gameService, hub)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
