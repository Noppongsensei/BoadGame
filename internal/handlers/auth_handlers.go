package handlers

import (
	"time"

	"avalon/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// JWT secret key
var jwtSecret = []byte("your-secret-key") // In production, use environment variables

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response for authentication endpoints
type AuthResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// registerHandler handles user registration
func registerHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Register user
		user, err := userService.RegisterUser(req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Generate JWT
		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(AuthResponse{
			User:  user,
			Token: token,
		})
	}
}

// loginHandler handles user login
func loginHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Authenticate user
		user, err := userService.AuthenticateUser(req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Generate JWT
		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.Status(fiber.StatusOK).JSON(AuthResponse{
			User:  user,
			Token: token,
		})
	}
}

// generateJWT generates a JWT token for authentication
func generateJWT(userID, username string) (string, error) {
	// Create claims with user information
	claims := &Claims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	return token.SignedString(jwtSecret)
}

// validateJWT validates a JWT token and returns the user ID
func validateJWT(tokenString string) (string, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	// Validate token
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return "", jwt.ErrSignatureInvalid
}
