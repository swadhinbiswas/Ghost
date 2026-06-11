package collab

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestHubCreateAndGetRoom(t *testing.T) {
	h := NewHub()

	// Create a room and verify it exists.
	roomID := "test-room-1"
	room := h.GetOrCreateRoom(roomID)
	if room == nil {
		t.Fatal("expected non-nil room")
	}
	if room.ID != roomID {
		t.Fatalf("expected room ID %q, got %q", roomID, room.ID)
	}

	// Getting the same room again should return the same instance.
	room2 := h.GetOrCreateRoom(roomID)
	if room != room2 {
		t.Error("expected same room instance on second GetOrCreateRoom call")
	}

	// ListRooms should contain the room.
	rooms := h.ListRooms()
	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}
	if rooms[0].ID != roomID {
		t.Errorf("expected room ID %q in list, got %q", roomID, rooms[0].ID)
	}
}

func TestHubGetRoomNotFound(t *testing.T) {
	h := NewHub()
	room, ok := h.GetRoom("nonexistent")
	if ok {
		t.Error("expected room not to exist")
	}
	if room != nil {
		t.Error("expected nil room for nonexistent ID")
	}
}

func TestHubDeleteRoom(t *testing.T) {
	h := NewHub()
	roomID := "delete-me"
	h.GetOrCreateRoom(roomID)

	// Verify room exists.
	if _, ok := h.GetRoom(roomID); !ok {
		t.Fatal("room should exist before deletion")
	}

	h.DeleteRoom(roomID)

	// Verify room is gone.
	if _, ok := h.GetRoom(roomID); ok {
		t.Error("room should not exist after deletion")
	}
}

func TestHubDeleteRoomNotFound(t *testing.T) {
	h := NewHub()
	// Should not panic when deleting a nonexistent room.
	h.DeleteRoom("nonexistent")
}

func TestHubListRoomsEmpty(t *testing.T) {
	h := NewHub()
	rooms := h.ListRooms()
	if len(rooms) != 0 {
		t.Fatalf("expected 0 rooms, got %d", len(rooms))
	}
}

func TestHubMultipleRooms(t *testing.T) {
	h := NewHub()
	ids := []string{"room-a", "room-b", "room-c"}
	for _, id := range ids {
		h.GetOrCreateRoom(id)
	}

	rooms := h.ListRooms()
	if len(rooms) != len(ids) {
		t.Fatalf("expected %d rooms, got %d", len(ids), len(rooms))
	}

	// Delete one and verify the others remain.
	h.DeleteRoom("room-b")
	rooms = h.ListRooms()
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms after deletion, got %d", len(rooms))
	}
}

func TestHubConnectionCount(t *testing.T) {
	h := NewHub()
	if c := h.ConnectionCount(); c != 0 {
		t.Fatalf("expected 0 connections on new hub, got %d", c)
	}
}

func TestHubClientCount(t *testing.T) {
	h := NewHub()
	roomID := "clients-test"
	room := h.GetOrCreateRoom(roomID)

	if c := h.ClientCount(); c != 0 {
		t.Fatalf("expected 0 clients on new room, got %d", c)
	}

	_ = room // room exists, no clients added yet
	if c := h.ClientCount(); c != 0 {
		t.Fatalf("expected 0 clients with no clients added, got %d", c)
	}
}

func TestHubShutdown(t *testing.T) {
	h := NewHub()
	h.GetOrCreateRoom("room-1")
	h.GetOrCreateRoom("room-2")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}

	// After shutdown, no rooms should remain.
	rooms := h.ListRooms()
	if len(rooms) != 0 {
		t.Fatalf("expected 0 rooms after shutdown, got %d", len(rooms))
	}
}

func TestHubShutdownTimeout(t *testing.T) {
	h := NewHub()
	// Add a room with a client that has a full send channel to simulate stuck goroutines.
	roomID := "stuck-room"
	h.GetOrCreateRoom(roomID)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond) // Ensure context expires

	err := h.Shutdown(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded on timed out shutdown, got: %v", err)
	}
}

func TestIsValidRoomID(t *testing.T) {
	tests := []struct {
		roomID string
		want   bool
	}{
		{"abc123", true},
		{"my-room", true},
		{"room_123", true},
		{"a", true},
		{"", false},
		{"room with spaces", false},
		{"room/../traversal", false},
		{"room<script>", false},
		{"a-b_c123", true},
	}

	for _, tt := range tests {
		t.Run(tt.roomID, func(t *testing.T) {
			got := isValidRoomID(tt.roomID)
			if got != tt.want {
				t.Errorf("isValidRoomID(%q) = %v, want %v", tt.roomID, got, tt.want)
			}
		})
	}
}

func TestRoomAddAndRemoveClient(t *testing.T) {
	room := NewRoom("test")
	client := &Client{ID: "user1", Name: "User One"}

	room.AddClient(client)

	clients := room.GetClients()
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if clients[0].ID != "user1" {
		t.Errorf("expected client ID user1, got %s", clients[0].ID)
	}

	room.RemoveClient("user1")
	clients = room.GetClients()
	if len(clients) != 0 {
		t.Fatalf("expected 0 clients after removal, got %d", len(clients))
	}
}

func TestRoomRemoveClientCleansCursorsAndSelections(t *testing.T) {
	room := NewRoom("test")
	client := &Client{ID: "user1", Name: "User One"}
	room.AddClient(client)

	// Manually add cursor and selection state.
	room.Document.Cursors["user1"] = Cursor{ClientID: "user1", Position: 5}
	room.Document.Selections["user1"] = Selection{ClientID: "user1", Start: 0, End: 10}

	room.RemoveClient("user1")

	if _, ok := room.Document.Cursors["user1"]; ok {
		t.Error("cursor should be removed when client leaves")
	}
	if _, ok := room.Document.Selections["user1"]; ok {
		t.Error("selection should be removed when client leaves")
	}
}

func TestRoomGetClientsEmpty(t *testing.T) {
	room := NewRoom("empty")
	clients := room.GetClients()
	if len(clients) != 0 {
		t.Fatalf("expected 0 clients, got %d", len(clients))
	}
}

func TestRoomBroadcast(t *testing.T) {
	room := NewRoom("test")
	c1 := &Client{ID: "user1", Name: "User One"}
	c2 := &Client{ID: "user2", Name: "User Two"}
	c2.Send = make(chan []byte, 1)
	room.AddClient(c1)
	room.AddClient(c2)

	// c1 sends, so only c2 should receive (sender is excluded).
	msg := Message{Type: "chat", Payload: []byte(`{"message":"hello"}`)}
	room.Broadcast("user1", msg)

	select {
	case data := <-c2.Send:
		if len(data) == 0 {
			t.Error("expected non-empty message data")
		}
	default:
		t.Error("expected message to be sent to c2")
	}
}

func TestRoomBroadcastDropsFullBuffer(t *testing.T) {
	room := NewRoom("test")
	c1 := &Client{ID: "user1", Name: "User One"}
	// Client with a send buffer of 1, already full.
	c2 := &Client{ID: "user2", Name: "User Two"}
	c2.Send = make(chan []byte, 1)
	c2.Send <- []byte("existing") // fill the buffer
	room.AddClient(c1)
	room.AddClient(c2)

	// This should not block or panic; it should drop the message and log a warning.
	msg := Message{Type: "chat", Payload: []byte(`{"message":"hello"}`)}
	room.Broadcast("user1", msg)
	// If we get here without hanging, the test passes.
}

func TestRoomApplyAndBroadcast(t *testing.T) {
	room := NewRoom("test")
	c1 := &Client{ID: "user1", Name: "User One"}
	c2 := &Client{ID: "user2", Name: "User Two"}
	c2.Send = make(chan []byte, 16)
	room.AddClient(c1)
	room.AddClient(c2)

	// Clear send channel
	c2.Send = make(chan []byte, 16)

	op := Operation{
		Type:      OpInsert,
		ClientID:  "user1",
		Position:  0,
		Content:   "Hello",
		Timestamp: time.Now(),
		Version:   VectorClock{"user1": 1},
	}

	if err := room.ApplyAndBroadcast(op); err != nil {
		t.Fatalf("ApplyAndBroadcast returned error: %v", err)
	}

	// Content should be applied.
	content := room.Document.GetContent()
	if content != "Hello" {
		t.Errorf("expected content %q, got %q", "Hello", content)
	}

	// Operation should be in history.
	if len(room.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(room.History))
	}
}

func TestMustMarshalWithValidInput(t *testing.T) {
	data := mustMarshal(map[string]string{"key": "value"})
	if data == nil {
		t.Fatal("expected non-nil result for valid input")
	}
	if string(data) == "" {
		t.Error("expected non-empty marshaled data")
	}
}

func TestMustMarshalWithInvalidInput(t *testing.T) {
	// Channels can't be marshaled to JSON; mustMarshal should return nil and log.
	ch := make(chan int)
	data := mustMarshal(ch)
	if data != nil {
		t.Error("expected nil result for unmarshalable input")
	}
}

func TestShareURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		roomID  string
		want    string
	}{
		{
			name:    "default base URL",
			baseURL: "",
			roomID:  "session-abc123",
			want:    "ws://localhost:8787/?room=session-abc123",
		},
		{
			name:    "custom base URL",
			baseURL: "ws://ghost.example.com:8787",
			roomID:  "session-xyz",
			want:    "ws://ghost.example.com:8787/?room=session-xyz",
		},
		{
			name:    "base URL with existing path",
			baseURL: "ws://ghost.example.com/collab",
			roomID:  "session-foo",
			want:    "ws://ghost.example.com/collab?room=session-foo",
		},
		{
			name:    "empty room ID",
			baseURL: "",
			roomID:  "",
			want:    "ws://localhost:8787/?room=",
		},
		{
			name:    "base URL without scheme gets ws:// prefix",
			baseURL: "ghost.example.com:8787",
			roomID:  "session-bar",
			want:    "ws://ghost.example.com:8787/?room=session-bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShareURL(tt.baseURL, tt.roomID)
			if got != tt.want {
				t.Errorf("ShareURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVectorClock(t *testing.T) {
	vc := make(VectorClock)
	vc.Increment("user1")
	vc.Increment("user1")
	vc.Increment("user2")

	if vc["user1"] != 2 {
		t.Errorf("expected user1 count 2, got %d", vc["user1"])
	}
	if vc["user2"] != 1 {
		t.Errorf("expected user2 count 1, got %d", vc["user2"])
	}

	other := VectorClock{"user1": 1, "user3": 3}
	vc.Merge(other)

	if vc["user1"] != 2 { // max(2, 1)
		t.Errorf("expected user1 count 2 after merge, got %d", vc["user1"])
	}
	if vc["user3"] != 3 {
		t.Errorf("expected user3 count 3 after merge, got %d", vc["user3"])
	}
}

func TestVectorClockCompare(t *testing.T) {
	tests := []struct {
		name     string
		vc       VectorClock
		other    VectorClock
		expected int
	}{
		{
			name:     "equal clocks",
			vc:       VectorClock{"a": 1, "b": 2},
			other:    VectorClock{"a": 1, "b": 2},
			expected: 2,
		},
		{
			name:     "vc is less",
			vc:       VectorClock{"a": 1},
			other:    VectorClock{"a": 2},
			expected: -1,
		},
		{
			name:     "vc is greater",
			vc:       VectorClock{"a": 3},
			other:    VectorClock{"a": 1},
			expected: 1,
		},
		{
			name:     "concurrent clocks",
			vc:       VectorClock{"a": 2, "b": 1},
			other:    VectorClock{"a": 1, "b": 2},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vc.Compare(tt.other)
			if got != tt.expected {
				t.Errorf("Compare() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDocumentState(t *testing.T) {
	ds := NewDocumentState()

	// Test initial state.
	if ds.GetContent() != "" {
		t.Errorf("expected empty content, got %q", ds.GetContent())
	}

	// Test insert operation.
	op := Operation{
		Type:      OpInsert,
		ClientID:  "user1",
		Position:  0,
		Content:   "Hello World",
		Timestamp: time.Now(),
		Version:   VectorClock{"user1": 1},
	}
	if err := ds.ApplyOperation(op); err != nil {
		t.Fatalf("ApplyOperation failed: %v", err)
	}
	if ds.GetContent() != "Hello World" {
		t.Errorf("expected content %q, got %q", "Hello World", ds.GetContent())
	}

	// Test delete operation.
	delOp := Operation{
		Type:      OpDelete,
		ClientID:  "user1",
		Position:  6,
		Length:    5,
		Timestamp: time.Now(),
		Version:   VectorClock{"user1": 2},
	}
	if err := ds.ApplyOperation(delOp); err != nil {
		t.Fatalf("ApplyOperation (delete) failed: %v", err)
	}
	if ds.GetContent() != "Hello " {
		t.Errorf("expected content %q, got %q", "Hello ", ds.GetContent())
	}

	// Test invalid insert position.
	badOp := Operation{
		Type:      OpInsert,
		ClientID:  "user1",
		Position:  100,
		Content:   "bad",
		Timestamp: time.Now(),
		Version:   VectorClock{"user1": 3},
	}
	if err := ds.ApplyOperation(badOp); err == nil {
		t.Error("expected error for invalid insert position")
	}
}

func TestDocumentStateCursors(t *testing.T) {
	ds := NewDocumentState()

	// Set a cursor via operation.
	op := Operation{
		Type:      OpCursor,
		ClientID:  "user1",
		CursorPos: 5,
		Timestamp: time.Now(),
	}
	if err := ds.ApplyOperation(op); err != nil {
		t.Fatalf("ApplyOperation (cursor) failed: %v", err)
	}

	cursors := ds.GetCursors()
	if len(cursors) != 1 {
		t.Fatalf("expected 1 cursor, got %d", len(cursors))
	}
	if cursors["user1"].Position != 5 {
		t.Errorf("expected cursor position 5, got %d", cursors["user1"].Position)
	}
}

func TestDocumentStateConcurrentAccess(t *testing.T) {
	ds := NewDocumentState()
	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent inserts from two goroutines.
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			op := Operation{
				Type:      OpInsert,
				ClientID:  "user1",
				Position:  0,
				Content:   "a",
				Timestamp: time.Now(),
				Version:   VectorClock{"user1": 1},
			}
			_ = ds.ApplyOperation(op)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			op := Operation{
				Type:      OpInsert,
				ClientID:  "user2",
				Position:  0,
				Content:   "b",
				Timestamp: time.Now(),
				Version:   VectorClock{"user2": 1},
			}
			_ = ds.ApplyOperation(op)
		}
	}()

	// Wait for both goroutines.
	wg.Wait()

	// Should have 200 characters total.
	content := ds.GetContent()
	if len(content) != 200 {
		t.Errorf("expected 200 characters, got %d", len(content))
	}
}

func TestDocumentStateMarshalJSON(t *testing.T) {
	ds := NewDocumentState()
	op := Operation{
		Type:      OpInsert,
		ClientID:  "user1",
		Position:  0,
		Content:   "test",
		Timestamp: time.Now(),
		Version:   VectorClock{"user1": 1},
	}
	_ = ds.ApplyOperation(op)

	data, err := ds.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestClientClose(t *testing.T) {
	// Create a mock client and close it; ensure the Send channel is closed.
	sendCh := make(chan []byte, 1)
	c := &Client{
		ID:   "test",
		Conn: nil, // nil is OK for this test
		Send: sendCh,
	}

	c.Close()

	select {
	case _, ok := <-sendCh:
		if ok {
			t.Error("Send channel should be closed after Close()")
		}
	default:
		t.Error("Send channel should be closed and readable")
	}
}
