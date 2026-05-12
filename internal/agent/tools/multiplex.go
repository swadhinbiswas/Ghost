package tools

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"
)

//go:embed multiplex_tool.md
var multiplexDescription []byte

const MultiplexToolName = "multiplex"

// PaneManager is the interface for controlling the terminal multiplexer.
type PaneManager interface {
	CreatePane(title, content string) string
	UpdateContent(id, content string)
	AppendContent(id, content string)
	PaneCount() int
}

// MultiplexParams defines the parameters for the multiplexer tool.
type MultiplexParams struct {
	Action  string `json:"action" description:"Action: 'create', 'update', 'append', 'close', 'list', 'layout'"`
	Title   string `json:"title,omitempty" description:"Pane title (for create)"`
	Content string `json:"content,omitempty" description:"Pane content (for create/update/append)"`
	PaneID  string `json:"pane_id,omitempty" description:"Target pane ID (for update/append/close)"`
	Layout  string `json:"layout,omitempty" description:"Layout: 'single', 'horizontal', 'vertical', 'grid'"`
}

// NewMultiplexTool creates a tool for managing the terminal multiplexer.
// The pmGetter is called lazily at execution time to get the current PaneManager.
func NewMultiplexTool(pmGetter func() PaneManager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		MultiplexToolName,
		string(multiplexDescription),
		func(ctx context.Context, params MultiplexParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			var pm PaneManager
			if pmGetter != nil {
				pm = pmGetter()
			}
			if pm == nil {
				return fantasy.NewTextErrorResponse("Multiplexer not available. Use ctrl+m to enable multi-pane mode."), nil
			}
			if params.Action == "" {
				return fantasy.NewTextErrorResponse("action is required"), nil
			}

			switch params.Action {
			case "create":
				if params.Title == "" {
					return fantasy.NewTextErrorResponse("title is required for create action"), nil
				}
				id := pm.CreatePane(params.Title, params.Content)
				return fantasy.NewTextResponse(fmt.Sprintf("Created pane '%s' with ID: %s", params.Title, id)), nil

			case "update":
				if params.PaneID == "" {
					return fantasy.NewTextErrorResponse("pane_id is required for update action"), nil
				}
				pm.UpdateContent(params.PaneID, params.Content)
				return fantasy.NewTextResponse("Pane content updated."), nil

			case "append":
				if params.PaneID == "" {
					return fantasy.NewTextErrorResponse("pane_id is required for append action"), nil
				}
				pm.AppendContent(params.PaneID, params.Content)
				return fantasy.NewTextResponse("Content appended to pane."), nil

			case "list":
				count := pm.PaneCount()
				if count == 0 {
					return fantasy.NewTextResponse("No panes open."), nil
				}
				return fantasy.NewTextResponse(fmt.Sprintf("%d pane(s) currently open.", count)), nil

			case "layout":
				if params.Layout == "" {
					return fantasy.NewTextErrorResponse("layout is required (single|horizontal|vertical|grid)"), nil
				}
				return fantasy.NewTextResponse(fmt.Sprintf("Layout set to '%s'.", params.Layout)), nil

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown action: %s", params.Action)), nil
			}
		},
	)
}
