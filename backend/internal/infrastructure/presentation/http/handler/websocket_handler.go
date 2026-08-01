package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	gorillaWS "github.com/gorilla/websocket"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/websocket"
)

// Upgrader for WebSocket connections.
var upgrader = gorillaWS.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure in production with specific origins
	},
}

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler struct {
	hub *websocket.Hub
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// Handle handles WebSocket connections at /ws.
func (h *WebSocketHandler) Handle(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	client := websocket.NewClient(conn)
	h.hub.RegisterClient(client)
	h.readMessages(c.Request.Context(), client)
}

func (h *WebSocketHandler) readMessages(ctx interface{}, client *websocket.Client) {
	defer func() {
		// Client disconnects
		deleteClient := &websocket.Client{
			ID: client.ID,
		}
		h.hub.UnregisterClient(deleteClient)
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg websocket.WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			h.sendError(client, "invalid message format")
			continue
		}

		h.handleMessage(ctx, client, &msg)
	}
}

func (h *WebSocketHandler) handleMessage(ctx interface{}, client *websocket.Client, msg *websocket.WebSocketMessage) {
	switch msg.Type {
	case "subscribe":
		h.handleSubscribe(client, msg)
	case "unsubscribe":
		h.handleUnsubscribe(client, msg)
	case "ping":
		h.send(client, websocket.WebSocketMessage{Type: "pong", Payload: msg.Payload})
	default:
		h.sendError(client, "unknown message type")
	}
}

func (h *WebSocketHandler) handleSubscribe(client *websocket.Client, msg *websocket.WebSocketMessage) {
	topic, ok := msg.Payload.(map[string]interface{})["topic"].(string)
	if !ok || topic == "" {
		h.sendError(client, "topic required")
		return
	}
	client.Topics = append(client.Topics, topic)
	h.send(client, websocket.WebSocketMessage{Type: "subscribed", Payload: gin.H{"topic": topic}})
}

func (h *WebSocketHandler) handleUnsubscribe(client *websocket.Client, msg *websocket.WebSocketMessage) {
	topic, ok := msg.Payload.(map[string]interface{})["topic"].(string)
	if !ok || topic == "" {
		return
	}
	// Remove topic from client's subscriptions
	filtered := make([]string, 0, len(client.Topics))
	for _, t := range client.Topics {
		if t != topic {
			filtered = append(filtered, t)
		}
	}
	client.Topics = filtered
	h.send(client, websocket.WebSocketMessage{Type: "unsubscribed", Payload: gin.H{"topic": topic}})
}

func (h *WebSocketHandler) send(client *websocket.Client, msg websocket.WebSocketMessage) {
	data, _ := json.Marshal(msg)
	select {
	case client.Send <- data:
	default:
		close(client.Send)
	}
}

func (h *WebSocketHandler) sendError(client *websocket.Client, message string) {
	h.send(client, websocket.WebSocketMessage{Type: "error", Payload: gin.H{"message": message}})
}
