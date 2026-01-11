package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"avalon/internal/handlers"
	"avalon/internal/middleware"
	"avalon/internal/repositories"
	"avalon/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	if err := handlers.InitJWTSecretFromEnv(); err != nil {
		log.Fatalf("JWT configuration error: %v", err)
	}

	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}

	// Setup database connection
	db, err := repositories.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize schema (retry a few times in case Postgres is still starting)
	for i := 0; i < 10; i++ {
		if err := db.InitSchema(); err == nil {
			break
		} else if i == 9 {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}
		time.Sleep(2 * time.Second)
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
	app := fiber.New(fiber.Config{
		// Allow panic recovery to return 500 instead of crashing empty
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add middlewares
	app.Use(recover.New()) // Recover from panic
	app.Use(middleware.RequestIDMiddleware())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

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
