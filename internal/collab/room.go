package collab

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// OperationType defines the type of collaborative operation.
type OperationType string

const (
	OpInsert OperationType = "insert"
	OpDelete OperationType = "delete"
	OpUpdate OperationType = "update"
	OpCursor OperationType = "cursor"
	OpSelect OperationType = "select"
	OpChat   OperationType = "chat"
)

// Operation represents a single collaborative operation.
type Operation struct {
	ID        string        `json:"id"`
	Type      OperationType `json:"type"`
	ClientID  string        `json:"client_id"`
	Timestamp time.Time     `json:"timestamp"`

	// For text operations
	Position int    `json:"position,omitempty"`
	Content  string `json:"content,omitempty"`
	Length   int    `json:"length,omitempty"`

	// For cursor/selection
	CursorPos   int `json:"cursor_pos,omitempty"`
	SelectStart int `json:"select_start,omitempty"`
	SelectEnd   int `json:"select_end,omitempty"`

	// For chat messages
	Message string `json:"message,omitempty"`

	// Version vector for conflict resolution
	Version VectorClock `json:"version"`
}

// VectorClock implements a simple vector clock for causal ordering.
type VectorClock map[string]int

// Increment increments the clock for a given client.
func (vc VectorClock) Increment(clientID string) {
	vc[clientID]++
}

// Merge merges another vector clock into this one.
func (vc VectorClock) Merge(other VectorClock) {
	for k, v := range other {
		if vc[k] < v {
			vc[k] = v
		}
	}
}

// Compare compares two vector clocks.
// Returns -1 if vc < other, 1 if vc > other, 0 if concurrent, 2 if equal.
func (vc VectorClock) Compare(other VectorClock) int {
	vcLess := false
	otherLess := false

	allKeys := make(map[string]bool)
	for k := range vc {
		allKeys[k] = true
	}
	for k := range other {
		allKeys[k] = true
	}

	for k := range allKeys {
		v1 := vc[k]
		v2 := other[k]
		if v1 < v2 {
			vcLess = true
		}
		if v1 > v2 {
			otherLess = true
		}
	}

	if !vcLess && !otherLess {
		return 2 // Equal
	}
	if vcLess && !otherLess {
		return -1 // vc happened before other
	}
	if !vcLess && otherLess {
		return 1 // other happened before vc
	}
	return 0 // Concurrent
}

// DocumentState represents the shared document state.
type DocumentState struct {
	Content    string               `json:"content"`
	Version    VectorClock          `json:"version"`
	Cursors    map[string]Cursor    `json:"cursors"`
	Selections map[string]Selection `json:"selections"`
	mu         sync.RWMutex
}

// Cursor represents a user's cursor position.
type Cursor struct {
	ClientID string    `json:"client_id"`
	Position int       `json:"position"`
	Name     string    `json:"name"`
	Color    string    `json:"color"`
	LastSeen time.Time `json:"last_seen"`
}

// Selection represents a user's text selection.
type Selection struct {
	ClientID string `json:"client_id"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
}

// NewDocumentState creates a new shared document state.
func NewDocumentState() *DocumentState {
	return &DocumentState{
		Version:    make(VectorClock),
		Cursors:    make(map[string]Cursor),
		Selections: make(map[string]Selection),
	}
}

// ApplyOperation applies an operation to the document state.
func (ds *DocumentState) ApplyOperation(op Operation) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	switch op.Type {
	case OpInsert:
		if op.Position < 0 || op.Position > len(ds.Content) {
			return fmt.Errorf("invalid insert position: %d", op.Position)
		}
		ds.Content = ds.Content[:op.Position] + op.Content + ds.Content[op.Position:]
		ds.Version.Merge(op.Version)
		ds.Version.Increment(op.ClientID)

	case OpDelete:
		if op.Position < 0 || op.Position+op.Length > len(ds.Content) {
			return fmt.Errorf("invalid delete range: %d-%d", op.Position, op.Position+op.Length)
		}
		ds.Content = ds.Content[:op.Position] + ds.Content[op.Position+op.Length:]
		ds.Version.Merge(op.Version)
		ds.Version.Increment(op.ClientID)

	case OpUpdate:
		// Full content replacement (used for initial sync or conflict resolution)
		ds.Content = op.Content
		ds.Version = op.Version
		ds.Version.Increment(op.ClientID)

	case OpCursor:
		ds.Cursors[op.ClientID] = Cursor{
			ClientID: op.ClientID,
			Position: op.CursorPos,
			LastSeen: time.Now(),
		}

	case OpSelect:
		ds.Selections[op.ClientID] = Selection{
			ClientID: op.ClientID,
			Start:    op.SelectStart,
			End:      op.SelectEnd,
		}
	}

	return nil
}

// GetContent returns the current document content.
func (ds *DocumentState) GetContent() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.Content
}

// GetVersion returns the current version vector.
func (ds *DocumentState) GetVersion() VectorClock {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	version := make(VectorClock)
	for k, v := range ds.Version {
		version[k] = v
	}
	return version
}

// GetCursors returns all active cursors.
func (ds *DocumentState) GetCursors() map[string]Cursor {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make(map[string]Cursor)
	for k, v := range ds.Cursors {
		result[k] = v
	}
	return result
}

// MarshalJSON implements json.Marshaler.
func (ds *DocumentState) MarshalJSON() ([]byte, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	type Alias DocumentState
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(ds)})
}

// Message represents a WebSocket message between clients and server.
type Message struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	ClientID  string          `json:"client_id,omitempty"`
	SessionID string          `json:"session_id"`
}

// JoinMessage is sent when a client joins a session.
type JoinMessage struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	Color      string `json:"color"`
}

// LeaveMessage is sent when a client leaves a session.
type LeaveMessage struct {
	ClientID string `json:"client_id"`
}

// Client represents a connected collaborator.
type Client struct {
	ID       string
	Name     string
	Color    string
	Conn     *websocket.Conn
	Send     chan []byte
	LastSeen time.Time
	mu       sync.Mutex
}

// SendJSON sends a JSON message to the client.
func (c *Client) SendJSON(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteMessage(websocket.TextMessage, data)
}

// Close closes the client connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Conn != nil {
		c.Conn.Close()
	}
	close(c.Send)
}

// Room represents a collaboration session.
type Room struct {
	ID       string
	Clients  map[string]*Client
	Document *DocumentState
	History  []Operation
	mu       sync.RWMutex
}

// NewRoom creates a new collaboration room.
func NewRoom(id string) *Room {
	return &Room{
		ID:       id,
		Clients:  make(map[string]*Client),
		Document: NewDocumentState(),
		History:  make([]Operation, 0),
	}
}

// AddClient adds a client to the room.
func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[client.ID] = client
}

// RemoveClient removes a client from the room.
func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, clientID)
	delete(r.Document.Cursors, clientID)
	delete(r.Document.Selections, clientID)
}

// GetClients returns all clients in the room.
func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.Clients))
	for _, c := range r.Clients {
		clients = append(clients, c)
	}
	return clients
}

// Broadcast sends a message to all clients except the sender.
func (r *Room) Broadcast(senderID string, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, client := range r.Clients {
		if id == senderID {
			continue
		}
		select {
		case client.Send <- data:
		default:
			// Client send buffer full, skip
		}
	}
}

// ApplyAndBroadcast applies an operation and broadcasts it to all clients.
func (r *Room) ApplyAndBroadcast(op Operation) error {
	if err := r.Document.ApplyOperation(op); err != nil {
		return err
	}

	r.mu.Lock()
	r.History = append(r.History, op)
	r.mu.Unlock()

	r.Broadcast(op.ClientID, Message{
		Type:    "operation",
		Payload: mustMarshal(op),
	})

	return nil
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
