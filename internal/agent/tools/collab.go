package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/collab"
)

//go:embed collab_tool.md
var collabToolDescription []byte

const CollabToolName = "collaborate"

// CollabParams defines the parameters for the collaboration tool.
type CollabParams struct {
	Action  string `json:"action" description:"Action: 'start', 'join', 'leave', 'list', 'status', 'share', 'message'"`
	Room    string `json:"room,omitempty" description:"Room/session ID"`
	Name    string `json:"name,omitempty" description:"Your display name"`
	Message string `json:"message,omitempty" description:"Chat message to broadcast"`
	Content string `json:"content,omitempty" description:"Content to share or insert"`
}

// NewCollabTool creates a tool for real-time collaboration.
func NewCollabTool(hub *collab.Hub) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		CollabToolName,
		string(collabToolDescription),
		func(ctx context.Context, params CollabParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Action == "" {
				return fantasy.NewTextErrorResponse("action is required"), nil
			}

			switch params.Action {
			case "start":
				roomID := params.Room
				if roomID == "" {
					roomID = fmt.Sprintf("session-%s", call.ID)
				}
				hub.GetOrCreateRoom(roomID)
				return fantasy.NewTextResponse(fmt.Sprintf("Collaboration room '%s' created. Share this ID with others to join.", roomID)), nil

			case "join":
				if params.Room == "" {
					return fantasy.NewTextErrorResponse("room is required to join"), nil
				}
				room, exists := hub.GetRoom(params.Room)
				if !exists {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("room '%s' not found", params.Room)), nil
				}
				clientCount := len(room.GetClients())
				return fantasy.NewTextResponse(fmt.Sprintf("Connected to room '%s'. %d participant(s) already present.", params.Room, clientCount)), nil

			case "leave":
				if params.Room == "" {
					return fantasy.NewTextErrorResponse("room is required to leave"), nil
				}
				room, exists := hub.GetRoom(params.Room)
				if !exists {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("room '%s' not found", params.Room)), nil
				}
				room.RemoveClient(params.Name)
				return fantasy.NewTextResponse(fmt.Sprintf("Left room '%s'.", params.Room)), nil

			case "list":
				rooms := hub.ListRooms()
				if len(rooms) == 0 {
					return fantasy.NewTextResponse("No active collaboration rooms."), nil
				}
				var sb strings.Builder
				sb.WriteString("Active Collaboration Rooms:\n\n")
				for _, r := range rooms {
					clients := r.GetClients()
					sb.WriteString(fmt.Sprintf("- **%s** (%d participant(s))\n", r.ID, len(clients)))
					for _, c := range clients {
						sb.WriteString(fmt.Sprintf("  - %s", c.Name))
						if c.Color != "" {
							sb.WriteString(fmt.Sprintf(" [%s]", c.Color))
						}
						sb.WriteString("\n")
					}
				}
				return fantasy.NewTextResponse(sb.String()), nil

			case "status":
				if params.Room == "" {
					return fantasy.NewTextErrorResponse("room is required for status"), nil
				}
				room, exists := hub.GetRoom(params.Room)
				if !exists {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("room '%s' not found", params.Room)), nil
				}
				clients := room.GetClients()
				content := room.Document.GetContent()
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Room: %s\n", params.Room))
				sb.WriteString(fmt.Sprintf("Participants: %d\n", len(clients)))
				sb.WriteString(fmt.Sprintf("Content length: %d characters\n", len(content)))
				sb.WriteString(fmt.Sprintf("Version: %v\n", room.Document.GetVersion()))
				return fantasy.NewTextResponse(sb.String()), nil

			case "share":
				if params.Room == "" {
					return fantasy.NewTextErrorResponse("room is required to share content"), nil
				}
				room, exists := hub.GetRoom(params.Room)
				if !exists {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("room '%s' not found", params.Room)), nil
				}
				op := collab.Operation{
					Type:     collab.OpUpdate,
					Content:  params.Content,
					ClientID: params.Name,
					Version:  room.Document.GetVersion(),
				}
				if err := room.ApplyAndBroadcast(op); err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to share content: %s", err)), nil
				}
				return fantasy.NewTextResponse("Content shared with all participants."), nil

			case "message":
				if params.Room == "" {
					return fantasy.NewTextErrorResponse("room is required to send message"), nil
				}
				room, exists := hub.GetRoom(params.Room)
				if !exists {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("room '%s' not found", params.Room)), nil
				}
				room.Broadcast(params.Name, collab.Message{
					Type: "chat",
					Payload: mustMarshal(map[string]interface{}{
						"client_id":   params.Name,
						"client_name": params.Name,
						"message":     params.Message,
					}),
				})
				return fantasy.NewTextResponse("Message broadcast to room."), nil

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown action: %s", params.Action)), nil
			}
		},
	)
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
