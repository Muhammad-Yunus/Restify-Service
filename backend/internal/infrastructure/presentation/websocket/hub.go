package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// Upgrader configures WebSocket upgrade.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all for now; configure CORS in production
	},
}

// Client represents a connected WebSocket client.
type Client struct {
	ID          string
	Conn        *websocket.Conn
	Send        chan []byte
	WorkspaceID string
	UserID      string
	ConnectedAt time.Time
	Topics      []string
	mu          sync.Mutex
}

// Hub manages all WebSocket clients and message routing.
type Hub struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	logger     *slog.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		logger:     slog.Default(),
	}
}

// Run starts the hub main loop.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			h.logger.Info("client connected", "client_id", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.logger.Info("client disconnected", "client_id", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// WebSocketMessage represents a message sent over WebSocket.
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Topic   string      `json:"topic,omitempty"`
	Payload any         `json:"payload,omitempty"`
}

// RegisterClient adds a client to the hub.
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient removes a client to the hub.
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(topic string, payload any) {
	data, err := json.Marshal(WebSocketMessage{
		Type:    "event",
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		h.logger.Error("failed to marshal broadcast message", "error", err)
		return
	}
	h.broadcast <- data
}

// BroadcastToWorkspace sends to clients subscribed to a workspace.
func (h *Hub) BroadcastToWorkspace(workspaceID string, topic string, payload any) {
	data, err := json.Marshal(WebSocketMessage{
		Type:    "event",
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		h.logger.Error("failed to marshal workspace broadcast", "error", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if client.WorkspaceID == workspaceID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// BroadcastToTopic sends to clients subscribed to a specific topic.
func (h *Hub) BroadcastToTopic(topic string, payload any) {
	data, err := json.Marshal(WebSocketMessage{
		Type:    "event",
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		h.logger.Error("failed to marshal topic broadcast", "error", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		client.mu.Lock()
		subscribed := false
		for _, t := range client.Topics {
			if t == topic {
				subscribed = true
				break
			}
		}
		client.mu.Unlock()
		if subscribed {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// SendToClient sends a message directly to a specific client.
func (h *Hub) SendToClient(clientID string, msg WebSocketMessage) {
	h.mu.RLock()
	client, ok := h.clients[clientID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	data, _ := json.Marshal(msg)
	select {
	case client.Send <- data:
	default:
		close(client.Send)
		h.mu.Lock()
		delete(h.clients, clientID)
		h.mu.Unlock()
	}
}

// GetClientCount returns the number of connected clients.
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	count := len(h.clients)
	h.mu.RUnlock()
	return count
}

// Close closes the hub and all connected clients.
func (h *Hub) Close(ctx context.Context) {
	h.closeAll ()
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	for _, client := range h.clients {
		close(client.Send)
		client.Conn.Close()
	}
	h.clients = make(map[string]*Client)
	h.mu.Unlock()
}

// NewClient creates a new WebSocket client.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		ID:          uuid.New().String(),
		Conn:        conn,
		Send:        make(chan []byte, 256),
		ConnectedAt: time.Now(),
	}
}
