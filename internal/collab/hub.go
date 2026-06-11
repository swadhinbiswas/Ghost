package collab

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 65536
	// maxConnections is the maximum number of concurrent WebSocket connections.
	maxConnections = 65536
)

// isLocalhostOrigin returns true if the origin is a localhost variant.
func isLocalhostOrigin(origin string) bool {
	return origin == "http://localhost" ||
		origin == "https://localhost" ||
		origin == "http://127.0.0.1" ||
		origin == "https://127.0.0.1" ||
		origin == "http://[::1]" ||
		origin == "https://[::1]"
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// SetAllowedOrigins configures the origin allowlist for WebSocket connections.
// If empty, all origins are allowed (with a warning for non-localhost).
func SetAllowedOrigins(origins []string) {
	allowedOrigins = origins
}

var allowedOrigins []string

// checkOrigin validates WebSocket origin against an allowlist.
// In production, this should match the application's served origins.
// For local development, localhost origins are always allowed.
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Browsers always send Origin; non-browser clients may not.
		// Allow connections without an origin header.
		return true
	}

	// Allow localhost variants for development.
	if isLocalhostOrigin(origin) {
		return true
	}

	// Check against configured allowed origins.
	if len(allowedOrigins) > 0 {
		for _, a := range allowedOrigins {
			if a == origin {
				return true
			}
		}
		slog.Warn("WebSocket connection from disallowed origin", "origin", origin)
		return false
	}

	// No origins configured, allow all but log a warning for non-localhost.
	slog.Warn("WebSocket connection from non-localhost origin (no allowed origins configured)", "origin", origin)
	return true
}

// Hub manages all collaboration rooms.
type Hub struct {
	rooms      map[string]*Room
	mu         sync.RWMutex
	shutdown   chan struct{}
	shutdownWg sync.WaitGroup

	// Connection counting with atomic operations for lock-free reads.
	connCount int64
}

// NewHub creates a new collaboration hub.
func NewHub() *Hub {
	return &Hub{
		rooms:    make(map[string]*Room),
		shutdown: make(chan struct{}),
	}
}

// ConnectionCount returns the current number of active WebSocket connections.
func (h *Hub) ConnectionCount() int {
	return int(atomic.LoadInt64(&h.connCount))
}

// GetOrCreateRoom gets an existing room or creates a new one.
func (h *Hub) GetOrCreateRoom(roomID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		room = NewRoom(roomID)
		h.rooms[roomID] = room
		slog.Info("Collaboration room created", "room", roomID)
	}
	return room
}

// GetRoom returns a room by ID.
func (h *Hub) GetRoom(roomID string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, ok := h.rooms[roomID]
	return room, ok
}

// DeleteRoom removes a room and disconnects all clients.
func (h *Hub) DeleteRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, exists := h.rooms[roomID]; exists {
		for _, client := range room.Clients {
			client.Close()
		}
		delete(h.rooms, roomID)
		slog.Info("Collaboration room deleted", "room", roomID)
	}
}

// ListRooms returns all active rooms.
func (h *Hub) ListRooms() []*Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]*Room, 0, len(h.rooms))
	for _, r := range h.rooms {
		rooms = append(rooms, r)
	}
	return rooms
}

// ClientCount returns the total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, room := range h.rooms {
		count += len(room.Clients)
	}
	return count
}

// ServeHTTP implements http.Handler for WebSocket connections.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the server is shutting down.
	select {
	case <-h.shutdown:
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	default:
	}

	// Enforce maximum connection limit.
	if atomic.LoadInt64(&h.connCount) >= maxConnections {
		http.Error(w, "maximum connections reached", http.StatusServiceUnavailable)
		return
	}

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "room parameter required", http.StatusBadRequest)
		return
	}

	// Validate roomID format to prevent path traversal or injection.
	if !isValidRoomID(roomID) {
		http.Error(w, "invalid room ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err, "room", roomID, "remote_addr", r.RemoteAddr)
		return
	}

	atomic.AddInt64(&h.connCount, 1)

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	room := h.GetOrCreateRoom(roomID)
	room.AddClient(client)

	h.shutdownWg.Add(2)
	go func() {
		defer h.shutdownWg.Done()
		h.writePump(client)
		atomic.AddInt64(&h.connCount, -1)
	}()
	go func() {
		defer h.shutdownWg.Done()
		h.readPump(client, room)
	}()

	slog.Info("WebSocket client connected", "room", roomID, "client", client.ID, "total_connections", h.ConnectionCount())
}

// isValidRoomID validates that a room ID is alphanumeric and reasonably sized.
func isValidRoomID(roomID string) bool {
	if len(roomID) == 0 || len(roomID) > 128 {
		return false
	}
	for _, c := range roomID {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return false
		}
	}
	return true
}

func (h *Hub) readPump(client *Client, room *Room) {
	defer func() {
		room.RemoveClient(client.ID)
		client.Close()
		h.broadcastPresence(room, client, "left")
	}()

	conn := client.Conn
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Error("WebSocket read error", "error", err, "room", room.ID, "client", client.ID)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Error("Invalid message received", "error", err, "room", room.ID, "client", client.ID)
			continue
		}

		msg.ClientID = client.ID

		switch msg.Type {
		case "join":
			var join JoinMessage
			if err := json.Unmarshal(msg.Payload, &join); err != nil {
				slog.Error("Failed to parse join message", "error", err, "room", room.ID, "client", client.ID)
				continue
			}
			client.ID = join.ClientID
			client.Name = join.ClientName
			client.Color = join.Color
			client.LastSeen = time.Now()

			// Send current document state to the joining client
			if err := client.SendJSON(Message{
				Type: "sync",
				Payload: mustMarshal(map[string]interface{}{
					"content":    room.Document.GetContent(),
					"version":    room.Document.GetVersion(),
					"cursors":    room.Document.GetCursors(),
					"selections": room.Document.Selections,
					"clients":    room.GetClients(),
				}),
			}); err != nil {
				slog.Error("Failed to send sync message to joining client", "error", err, "room", room.ID, "client", client.ID)
			}

			h.broadcastPresence(room, client, "joined")

		case "operation":
			var op Operation
			if err := json.Unmarshal(msg.Payload, &op); err != nil {
				slog.Error("Failed to parse operation message", "error", err, "room", room.ID, "client", client.ID)
				continue
			}
			op.ClientID = client.ID
			if err := room.ApplyAndBroadcast(op); err != nil {
				slog.Error("Operation apply failed", "error", err, "room", room.ID, "client", client.ID)
			}

		case "cursor":
			var cursorOp Operation
			if err := json.Unmarshal(msg.Payload, &cursorOp); err != nil {
				slog.Error("Failed to parse cursor message", "error", err, "room", room.ID, "client", client.ID)
				continue
			}
			cursorOp.ClientID = client.ID
			cursorOp.Type = OpCursor
			_ = room.ApplyAndBroadcast(cursorOp)

		case "chat":
			var chatMsg struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(msg.Payload, &chatMsg); err != nil {
				slog.Error("Failed to parse chat message", "error", err, "room", room.ID, "client", client.ID)
				continue
			}
			room.Broadcast(client.ID, Message{
				Type: "chat",
				Payload: mustMarshal(map[string]interface{}{
					"client_id":   client.ID,
					"client_name": client.Name,
					"message":     chatMsg.Message,
					"timestamp":   time.Now().Unix(),
				}),
			})

		case "ping":
			_ = client.SendJSON(Message{Type: "pong"})

		default:
			slog.Warn("Unknown message type received", "type", msg.Type, "room", room.ID, "client", client.ID)
		}
	}
}

func (h *Hub) writePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Batch pending messages
			n := len(client.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) broadcastPresence(room *Room, client *Client, action string) {
	room.Broadcast(client.ID, Message{
		Type: "presence",
		Payload: mustMarshal(map[string]interface{}{
			"action":    action,
			"client_id": client.ID,
			"clients":   room.GetClients(),
		}),
	})
}

// Shutdown gracefully closes all rooms and waits for goroutines.
func (h *Hub) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down collaboration server...")

	close(h.shutdown)

	// Close all rooms and their clients.
	h.mu.Lock()
	for roomID, room := range h.rooms {
		for _, client := range room.Clients {
			client.Close()
		}
		delete(h.rooms, roomID)
	}
	h.mu.Unlock()

	// Wait for all pump goroutines to finish with a timeout.
	done := make(chan struct{})
	go func() {
		h.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Collaboration server shutdown complete")
		return nil
	case <-ctx.Done():
		slog.Warn("Collaboration server shutdown timed out, forcing exit")
		return ctx.Err()
	}
}
