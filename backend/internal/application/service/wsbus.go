package service

import (
	"context"

	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/websocket"
)

// WSBus provides WebSocket broadcasting capabilities.
type WSBus struct {
	hub *websocket.Hub
}

// NewWSBus creates a new WebSocket broadcast hub.
func NewWSBus(hub *websocket.Hub) *WSBus {
	return &WSBus{hub: hub}
}

// Broadcast sends a message to all connected WebSocket clients.
func (w *WSBus) Broadcast(topic string, payload any) {
	w.hub.Broadcast(topic, payload)
}

// BroadcastToWorkspace sends a message to all clients in a workspace.
func (w *WSBus) BroadcastToWorkspace(workspaceID string, topic string, payload any) {
	w.hub.BroadcastToWorkspace(workspaceID, topic, payload)
}

// BroadcastToTopic sends a message to clients subscribed to a topic.
func (w *WSBus) BroadcastToTopic(topic string, payload any) {
	w.hub.BroadcastToTopic(topic, payload)
}

// SendToClient sends a message directly to a specific client.
func (w *WSBus) SendToClient(clientID string, msg websocket.WebSocketMessage) {
	w.hub.SendToClient(clientID, msg)
}

// ClientCount returns the number of connected clients.
func (w *WSBus) ClientCount() int {
	return w.hub.GetClientCount()
}

// Start runs the hub main loop in the background.
func (w *WSBus) Start(ctx context.Context) {
	go w.hub.Run(ctx)
}
