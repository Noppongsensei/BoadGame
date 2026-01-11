package handlers

import (
	"avalon/internal/services"

	"github.com/gofiber/fiber/v2"
)

// getUserHandler handles GET requests for user information
func getUserHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from URL
		userID := c.Params("id")

		tokenUserID, ok := c.Locals("userID").(string)
		if !ok || tokenUserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Only allow users to access their own information
		if tokenUserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Get user
		user, err := userService.GetUser(userID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.Status(fiber.StatusOK).JSON(user)
	}
}

// updateUserHandler handles PUT requests for updating user information
func updateUserHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from URL
		userID := c.Params("id")

		tokenUserID, ok := c.Locals("userID").(string)
		if !ok || tokenUserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Only allow users to update their own information
		if tokenUserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Parse request body
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Update user
		user, err := userService.UpdateUser(userID, req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(user)
	}
}

// deleteUserHandler handles DELETE requests for deleting a user
func deleteUserHandler(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from URL
		userID := c.Params("id")

		tokenUserID, ok := c.Locals("userID").(string)
		if !ok || tokenUserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Only allow users to delete their own account
		if tokenUserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Delete user
		if err := userService.DeleteUser(userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete user",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User deleted successfully",
		})
	}
}
