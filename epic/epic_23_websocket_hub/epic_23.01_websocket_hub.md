# Epic 23 — WebSocket Hub

**Goal:** Implement WebSocket hub for real-time bi-directional communication, event broadcasting, and client management.
**Dependencies:** Epic 22 (MQTT Event Bus), Epic 13 (HTTP Router)
**Commit:** `feat: add WebSocket hub for real-time communication`

---

## Step 23.01 — WebSocket Hub Implementation

**Build:** Create `backend/internal/infrastructure/presentation/websocket/hub.go`:

```go
package websocket

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
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
    ID          string          `json:"id"`
    Conn        *websocket.Conn `json:"-"`
    Send        chan []byte     `json:"-"`
    WorkspaceID string          `json:"workspace_id,omitempty"`
    UserID      string          `json:"user_id,omitempty"`
    ConnectedAt time.Time       `json:"connected_at"`
    Topics      []string        `json:"-"`
    mu          sync.Mutex
}

// Hub manages all WebSocket clients and message routing.
type Hub struct {
    clients    map[string]*Client
    mu         sync.RWMutex
    register   chan *Client
    unregister chan *Client
    broadcast  chan []byte
    logger     repository.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger repository.Logger) *Hub {
    return &Hub{
        clients:    make(map[string]*Client),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan []byte, 256),
        logger:     logger,
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
            h.logger.Info(ctx, "client connected", "client_id", client.ID)

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client.ID]; ok {
                delete(h.clients, client.ID)
                close(client.Send)
            }
            h.mu.Unlock()
            h.logger.Info(ctx, "client disconnected", "client_id", client.ID)

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

// HandleWebSocket upgrades HTTP to WebSocket.
func (h *Hub) HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        h.logger.Error(c.Request.Context(), "websocket upgrade failed", "error", err)
        return
    }

    client := &Client{
        ID:          uuid.New().String(),
        Conn:        conn,
        Send:        make(chan []byte, 256),
        ConnectedAt: time.Now(),
    }

    h.register <- client
    h.readMessages(c.Request.Context(), client)
}

func (h *Hub) readMessages(ctx context.Context, client *Client) {
    defer func() {
        h.unregister <- client
        client.Conn.Close()
    }()

    for {
        _, message, err := client.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
                h.logger.Error(ctx, "websocket read error", "client_id", client.ID, "error", err)
            }
            break
        }

        // Parse and route message
        var msg WebSocketMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            h.sendError(client, "invalid message format")
            continue
        }

        h.handleMessage(ctx, client, &msg)
    }
}

func (h *Hub) handleMessage(ctx context.Context, client *Client, msg *WebSocketMessage) {
    switch msg.Type {
    case "subscribe":
        h.handleSubscribe(client, msg)
    case "unsubscribe":
        h.handleUnsubscribe(client, msg)
    case "ping":
        h.send(client, WebSocketMessage{Type: "pong", Payload: msg.Payload})
    default:
        h.send(client, WebSocketMessage{Type: "error", Payload: gin.H{"message": "unknown message type"}})
    }
}

func (h *Hub) handleSubscribe(client *Client, msg *WebSocketMessage) {
    topic, ok := msg.Payload["topic"].(string)
    if !ok {
        h.sendError(client, "topic required")
        return
    }
    client.mu.Lock()
    client.Topics = append(client.Topics, topic)
    client.mu.Unlock()
    h.send(client, WebSocketMessage{Type: "subscribed", Payload: gin.H{"topic": topic}})
}

func (h *Hub) handleUnsubscribe(client *Client, msg *WebSocketMessage) {
    topic, ok := msg.Payload["topic"].(string)
    if !ok {
        return
    }
    client.mu.Lock()
    for i, t := range client.Topics {
        if t == topic {
            client.Topics = append(client.Topics[:i], client.Topics[i+1:]...)
            break
        }
    }
    client.mu.Unlock()
    h.send(client, WebSocketMessage{Type: "unsubscribed", Payload: gin.H{"topic": topic}})
}

func (h *Hub) send(client *Client, msg WebSocketMessage) {
    data, _ := json.Marshal(msg)
    select {
    case client.Send <- data:
    default:
        close(client.Send)
        delete(h.clients, client.ID)
    }
}

func (h *Hub) sendError(client *Client, message string) {
    h.send(client, WebSocketMessage{Type: "error", Payload: gin.H{"message": message}})
}

func (h *Hub) Broadcast(topic string, payload any) {
    data, _ := json.Marshal(WebSocketMessage{
        Type:    "event",
        Topic:   topic,
        Payload: payload,
    })
    h.broadcast <- data
}

// BroadcastToWorkspace sends to clients subscribed to workspace.
func (h *Hub) BroadcastToWorkspace(workspaceID string, topic string, payload any) {
    data, _ := json.Marshal(WebSocketMessage{
        Type:    "event",
        Topic:   topic,
        Payload: payload,
    })
    h.mu.RLock()
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
    h.mu.RUnlock()
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

// WebSocketMessage represents a message sent over WebSocket.
type WebSocketMessage struct {
    Type    string         `json:"type"`
    Topic   string         `json:"topic,omitempty"`
    Payload any            `json:"payload,omitempty"`
}

// Compile-time check.
// NOTE: Hub implements a subset of MQTTBroker interface.
// The full interface compliance is checked in the MQTT adapter.
var _ repository.MQTTBroker = (*Hub)(nil)
```

**Test cases:**
- [ ] Unit: `HandleWebSocket()` upgrades connection
- [ ] Unit: `Broadcast()` sends to all clients
- [ ] Unit: `BroadcastToWorkspace()` sends only to workspace clients
- [ ] Unit: Client disconnect is handled gracefully
- [ ] Integration: Full WebSocket connect/subscribe/broadcast cycle

---

## Step 23.02 — WebSocket Handler

**Build:** Create `backend/internal/presentation/http/handler/websocket_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/infrastructure/presentation/websocket"
)

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler struct {
    hub *websocket.Hub
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
    return &WebSocketHandler{hub: hub}
}

// Handle handles GET /ws
func (h *WebSocketHandler) Handle(c *gin.Context) {
    h.hub.HandleWebSocket(c)
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add WebSocket hub for real-time bi-directional communication"
```
