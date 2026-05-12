package collab

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 65536
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub manages all collaboration rooms.
type Hub struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewHub creates a new collaboration hub.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
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

// DeleteRoom removes a room.
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
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "room parameter required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	room := h.GetOrCreateRoom(roomID)
	room.AddClient(client)

	go h.writePump(client)
	go h.readPump(client, room)
}

func (h *Hub) readPump(client *Client, room *Room) {
	defer func() {
		room.RemoveClient(client.ID)
		client.Close()
		h.broadcastPresence(room, client, "left")
	}()

	conn := client.Conn
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Error("WebSocket read error", "error", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Error("Invalid message", "error", err)
			continue
		}

		msg.ClientID = client.ID

		switch msg.Type {
		case "join":
			var join JoinMessage
			if err := json.Unmarshal(msg.Payload, &join); err != nil {
				continue
			}
			client.ID = join.ClientID
			client.Name = join.ClientName
			client.Color = join.Color
			client.LastSeen = time.Now()

			// Send current document state to the joining client
			client.SendJSON(Message{
				Type: "sync",
				Payload: mustMarshal(map[string]interface{}{
					"content":    room.Document.GetContent(),
					"version":    room.Document.GetVersion(),
					"cursors":    room.Document.GetCursors(),
					"selections": room.Document.Selections,
					"clients":    room.GetClients(),
				}),
			})

			h.broadcastPresence(room, client, "joined")

		case "operation":
			var op Operation
			if err := json.Unmarshal(msg.Payload, &op); err != nil {
				continue
			}
			op.ClientID = client.ID
			if err := room.ApplyAndBroadcast(op); err != nil {
				slog.Error("Operation apply failed", "error", err)
			}

		case "cursor":
			var cursorOp Operation
			if err := json.Unmarshal(msg.Payload, &cursorOp); err != nil {
				continue
			}
			cursorOp.ClientID = client.ID
			cursorOp.Type = OpCursor
			room.ApplyAndBroadcast(cursorOp)

		case "chat":
			var chatMsg struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(msg.Payload, &chatMsg); err != nil {
				continue
			}
			room.Broadcast(client.ID, Message{
				Type: "chat",
				Payload: mustMarshal(map[string]interface{}{
					"client_id":   client.ID,
					"client_name": client.Name,
					"message":     chatMsg.Message,
					"timestamp":   time.Now(),
				}),
			})

		case "ping":
			client.SendJSON(Message{Type: "pong"})
		}
	}
}

func (h *Hub) writePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch pending messages
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
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
