package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "log"
)

// RequestIDMiddleware generates a unique request ID for each incoming HTTP request
// and sets it in the response header "X-Request-ID" as well as the context locals.
func RequestIDMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Generate a UUID v4
        id := uuid.New().String()
        // Store in locals for downstream handlers
        c.Locals("requestID", id)
        // Set response header
        c.Set("X-Request-ID", id)
        // Optionally log the request start
        log.Printf("[RequestID %s] %s %s", id, c.Method(), c.Path())
        // Continue to next handler
        err := c.Next()
        // Log response status
        log.Printf("[RequestID %s] Completed with status %d", id, c.Response().StatusCode())
        return err
    }
}
