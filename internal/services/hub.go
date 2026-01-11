package services

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// Client represents a connected websocket client
type Client struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	RoomID   string
	Hub      *Hub
	Send     chan []byte
	mu       sync.Mutex
}

// Message represents the structure of a WebSocket message
type Message struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"room_id"`
	UserID  string          `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients, mapped by userID
	clients map[string]*Client

	// Rooms maps roomID to a set of client IDs
	rooms map[string]map[string]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific room
	broadcast chan *Message

	// Direct message to specific client
	direct chan *Message

	// Lock to protect concurrent access
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		direct:     make(chan *Message),
	}
}

// Run starts the Hub to process events
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client

			// Add client to room if roomID is specified
			if client.RoomID != "" {
				if _, ok := h.rooms[client.RoomID]; !ok {
					h.rooms[client.RoomID] = make(map[string]bool)
				}
				h.rooms[client.RoomID][client.ID] = true
			}
			h.mu.Unlock()

			// Notify client that they are connected
			connectMsg := &Message{
				Type:   "system.connected",
				UserID: client.ID,
			}
			jsonMsg, _ := json.Marshal(connectMsg)
			select {
			case client.Send <- jsonMsg:
			default:
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				// Remove client from room
				if client.RoomID != "" && h.rooms[client.RoomID] != nil {
					delete(h.rooms[client.RoomID], client.ID)
					// If room is empty, remove it
					if len(h.rooms[client.RoomID]) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}

				// Close channel and delete client
				close(client.Send)
				delete(h.clients, client.ID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			jsonMsg, _ := json.Marshal(message)
			h.mu.RLock()
			if roomClients, ok := h.rooms[message.RoomID]; ok {
				for clientID := range roomClients {
					if client, ok := h.clients[clientID]; ok {
						select {
						case client.Send <- jsonMsg:
						default:
						}
					}
				}
			}
			h.mu.RUnlock()

		case message := <-h.direct:
			jsonMsg, _ := json.Marshal(message)
			h.mu.RLock()
			if client, ok := h.clients[message.UserID]; ok {
				select {
				case client.Send <- jsonMsg:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new client to the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(message *Message) {
	h.broadcast <- message
}

// SendToClient sends a direct message to a specific client
func (h *Hub) SendToClient(message *Message) {
	h.direct <- message
}

// ClientExists checks if a client with the given ID exists
func (h *Hub) ClientExists(clientID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.clients[clientID]
	return exists
}

// GetClientsInRoom returns all clients in a room
func (h *Hub) GetClientsInRoom(roomID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if roomClients, ok := h.rooms[roomID]; ok {
		for clientID := range roomClients {
			if client, exists := h.clients[clientID]; exists {
				clients = append(clients, client)
			}
		}
	}
	return clients
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(clientID, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.clients[clientID]; ok {
		// Remove from previous room if any
		if client.RoomID != "" && h.rooms[client.RoomID] != nil {
			delete(h.rooms[client.RoomID], client.ID)
			// If room is empty, remove it
			if len(h.rooms[client.RoomID]) == 0 {
				delete(h.rooms, client.RoomID)
			}
		}

		// Add to new room
		client.RoomID = roomID
		if _, ok := h.rooms[roomID]; !ok {
			h.rooms[roomID] = make(map[string]bool)
		}
		h.rooms[roomID][clientID] = true
	}
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.clients[clientID]; ok && client.RoomID != "" {
		if h.rooms[client.RoomID] != nil {
			delete(h.rooms[client.RoomID], clientID)
			// If room is empty, remove it
			if len(h.rooms[client.RoomID]) == 0 {
				delete(h.rooms, client.RoomID)
			}
		}
		client.RoomID = ""
	}
}

// Run handles WebSocket connection for each client
func (c *Client) Run() {
	defer c.Conn.Close()
	for message := range c.Send {
		c.mu.Lock()
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing message to client %s: %v", c.ID, err)
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()
	}
}
