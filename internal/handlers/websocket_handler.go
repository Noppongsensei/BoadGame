package handlers

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	"avalon/internal/services"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var (
	wsAllowedOriginsOnce sync.Once
	wsAllowedOrigins     []string
)

func loadWSAllowedOrigins() {
	wsAllowedOriginsOnce.Do(func() {
		origins := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
		if origins == "" {
			return
		}
		for _, o := range strings.Split(origins, ",") {
			o = strings.TrimSpace(o)
			if o == "" {
				continue
			}
			wsAllowedOrigins = append(wsAllowedOrigins, o)
		}
	})
}

// websocketHandler handles WebSocket connections
func websocketHandler(hub *services.Hub, gameService *services.GameService) fiber.Handler {
	loadWSAllowedOrigins()
	cfg := websocket.Config{}
	if len(wsAllowedOrigins) > 0 {
		cfg.Origins = wsAllowedOrigins
	}
	return websocket.New(func(c *websocket.Conn) {
		// Extract JWT token from query parameters
		token := c.Query("token")
		if token == "" {
			log.Println("WebSocket connection attempt without token")
			return
		}

		claims, err := parseJWTClaims(token)
		if err != nil {
			log.Printf("Invalid token in WebSocket connection: %v", err)
			return
		}
		userID := claims.UserID

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
		select {
		case client.Send <- welcomeJSON:
		default:
		}

		sendError := func(roomID string, msg string) {
			payload, _ := json.Marshal(map[string]interface{}{
				"message": msg,
			})
			hub.SendToClient(&services.Message{
				Type:    "system.error",
				RoomID:  roomID,
				UserID:  userID,
				Payload: payload,
			})
		}

		sendGameStateToRoom := func(roomID string) {
			if gameService == nil {
				return
			}
			clients := hub.GetClientsInRoom(roomID)
			for _, roomClient := range clients {
				state, err := gameService.GetFilteredGameState(roomID, roomClient.ID)
				if err != nil {
					continue
				}
				hub.SendToClient(&services.Message{
					Type:    "game.state_update",
					RoomID:  roomID,
					UserID:  roomClient.ID,
					Payload: state,
				})
			}
		}

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
				if message.RoomID == "" {
					sendError("", "room_id is required")
					continue
				}
				if client.RoomID == "" {
					sendError(message.RoomID, "client is not in a room")
					continue
				}
				if client.RoomID != message.RoomID {
					sendError(message.RoomID, "client is not in this room")
					continue
				}
				if gameService == nil {
					sendError(message.RoomID, "game service is not available")
					continue
				}

				var actionEnvelope struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(message.Payload, &actionEnvelope); err != nil {
					sendError(message.RoomID, "invalid action payload")
					continue
				}
				if actionEnvelope.Type == "" {
					sendError(message.RoomID, "action payload.type is required")
					continue
				}

				var payloadMap map[string]interface{}
				if err := json.Unmarshal(message.Payload, &payloadMap); err != nil {
					sendError(message.RoomID, "invalid action payload")
					continue
				}
				delete(payloadMap, "type")
				actionPayload, _ := json.Marshal(payloadMap)

				if err := gameService.ProcessGameAction(message.RoomID, userID, actionEnvelope.Type, actionPayload); err != nil {
					sendError(message.RoomID, err.Error())
					continue
				}

				sendGameStateToRoom(message.RoomID)
			case "room.join", "room.leave":
				sendError(message.RoomID, "room join/leave must be done via HTTP API")
			}
		}

		// Unregister client when connection is closed
		hub.UnregisterClient(client)
	}, cfg)
}
