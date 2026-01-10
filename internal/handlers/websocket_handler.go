package handlers

import (
	"encoding/json"
	"log"
	"strings"
	
	"avalon/internal/services"
	
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// websocketHandler handles WebSocket connections
func websocketHandler(hub *services.Hub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Extract JWT token from query parameters
		token := c.Query("token")
		if token == "" {
			log.Println("WebSocket connection attempt without token")
			return
		}
		
		// Validate token
		userID, err := validateJWT(token)
		if err != nil {
			log.Printf("Invalid token in WebSocket connection: %v", err)
			return
		}
		
		// Extract username from token
		claims := &Claims{}
		_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			log.Printf("Failed to parse token claims: %v", err)
			return
		}
		
		// Extract room ID from query parameters (optional)
		roomID := c.Query("room_id")
		
		// Create a new client
		client := &services.Client{
			ID:       userID,
			Username: claims.Username,
			Conn:     c,
			RoomID:   roomID,
			Hub:      hub,
			Send:     make(chan []byte, 256),
		}
		
		// Register client with hub
		hub.RegisterClient(client)
		
		// Handle WebSocket connection
		go client.Run()
		
		// Send connection confirmation
		welcomeMsg := map[string]interface{}{
			"type":    "system.welcome",
			"message": "Connected to Avalon game server",
			"user_id": userID,
		}
		
		welcomeJSON, _ := json.Marshal(welcomeMsg)
		client.Send <- welcomeJSON
		
		// Handle incoming messages
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			
			// Parse message
			var message services.Message
			if err := json.Unmarshal(msg, &message); err != nil {
				log.Printf("Error parsing WebSocket message: %v", err)
				continue
			}
			
			// Validate message
			message.UserID = userID // Ensure correct user ID
			
			// Process message based on type
			switch message.Type {
			case "chat.message":
				// Broadcast chat message to room
				if message.RoomID != "" {
					hub.BroadcastToRoom(&message)
				}
			case "game.action":
				// Process game action (handled by game service)
				if message.RoomID != "" {
					hub.BroadcastToRoom(&message)
				}
			case "room.join":
				// Join room
				if message.RoomID != "" {
					hub.JoinRoom(userID, message.RoomID)
					
					// Notify room about new player
					joinMsg := &services.Message{
						Type:    "room.player_joined",
						RoomID:  message.RoomID,
						UserID:  userID,
						Payload: []byte(`{"username":"` + claims.Username + `"}`),
					}
					hub.BroadcastToRoom(joinMsg)
				}
			case "room.leave":
				// Leave room
				if client.RoomID != "" {
					// Notify room about player leaving
					leaveMsg := &services.Message{
						Type:    "room.player_left",
						RoomID:  client.RoomID,
						UserID:  userID,
						Payload: []byte(`{"username":"` + claims.Username + `"}`),
					}
					hub.BroadcastToRoom(leaveMsg)
					
					// Leave room
					hub.LeaveRoom(userID)
				}
			}
		}
		
		// Unregister client when connection is closed
		hub.UnregisterClient(client)
	})
}
