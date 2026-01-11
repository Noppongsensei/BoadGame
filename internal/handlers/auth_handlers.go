// d:\Boadgame\internal\handlers\auth_handlers.go
package handlers

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"avalon/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
	jwtSecretErr  error
)

func InitJWTSecretFromEnv() error {
	jwtSecretOnce.Do(func() {
		secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
		if secret == "" {
			jwtSecretErr = errors.New("JWT_SECRET is not set")
			return
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecretErr
}

func getJWTSecret() ([]byte, error) {
	if err := InitJWTSecretFromEnv(); err != nil {
		return nil, err
	}
	return jwtSecret, nil
}

func parseJWTClaims(tokenString string) (*Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if tokenString == "" {
		return nil, errors.New("missing token")
	}

	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func authRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := parseJWTClaims(c.Get("Authorization"))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}
		c.Locals("userID", claims.UserID)
		c.Locals("claims", claims)
		return c.Next()
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the register request body
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

func registerHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse request body",
			})
		}

		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "username and password are required",
			})
		}

		user, err := userService.RegisterUser(req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		response := AuthResponse{Token: token}
		response.User.ID = user.ID
		response.User.Username = user.Username

		return c.Status(fiber.StatusCreated).JSON(response)
	}
}

func loginHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse request body",
			})
		}

		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "username and password are required",
			})
		}

		user, err := userService.AuthenticateUser(req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		response := AuthResponse{Token: token}
		response.User.ID = user.ID
		response.User.Username = user.Username

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// generateJWT generates a JWT token for a user
func generateJWT(userID, username string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// validateJWT validates a JWT token and returns the user ID
func validateJWT(tokenString string) (string, error) {
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if tokenString == "" {
		return "", errors.New("missing token")
	}

	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", jwt.ErrSignatureInvalid
}
