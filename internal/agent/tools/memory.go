package tools

import (
"context"
_ "embed"
"fmt"

"charm.land/fantasy"
"github.com/swadhinbiswas/ghost/internal/memory"
"github.com/swadhinbiswas/ghost/internal/config"
)

//go:embed memory.md
var memoryDescription []byte

const MemoryToolName = "memory"

type MemoryParams struct {
	Action string `json:"action" description:"The action to perform: 'get', 'set', 'delete', or 'list'"`
	Key    string `json:"key,omitempty" description:"The memory key to operate on (required for get/set/delete)"`
	Value  string `json:"value,omitempty" description:"The value to store (required for set)"`
}

func NewMemoryTool(store *config.ConfigStore) fantasy.AgentTool {
	return fantasy.NewAgentTool(
MemoryToolName,
string(memoryDescription),
func(ctx context.Context, params MemoryParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
memStore, err := memory.NewMemoryStore(store.WorkingDir())
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to init memory store: %w", err)
			}

			switch params.Action {
			case "get":
				if params.Key == "" {
					return fantasy.ToolResponse{}, fmt.Errorf("key is required for get action")
				}
				val, ok := memStore.Get(params.Key)
				if !ok {
					return fantasy.NewTextResponse(fmt.Sprintf("Memory key '%s' not found", params.Key)), nil
				}
				return fantasy.NewTextResponse(fmt.Sprintf("%s: %s", params.Key, val)), nil

			case "set":
				if params.Key == "" || params.Value == "" {
					return fantasy.ToolResponse{}, fmt.Errorf("key and value are required for set action")
				}
				if err := memStore.Set(params.Key, params.Value); err != nil {
					return fantasy.ToolResponse{}, fmt.Errorf("failed to save memory: %w", err)
				}
				return fantasy.NewTextResponse(fmt.Sprintf("Successfully saved '%s' to memory.", params.Key)), nil

			case "delete":
				if params.Key == "" {
					return fantasy.ToolResponse{}, fmt.Errorf("key is required for delete action")
				}
				if err := memStore.Delete(params.Key); err != nil {
					return fantasy.ToolResponse{}, fmt.Errorf("failed to delete memory: %w", err)
				}
				return fantasy.NewTextResponse(fmt.Sprintf("Successfully deleted '%s' from memory.", params.Key)), nil

			case "list":
				facts := memStore.GetAll()
				if len(facts) == 0 {
					return fantasy.NewTextResponse("Project memory is currently empty."), nil
				}
				out := "Project Memory:\n"
				for k, v := range facts {
					out += fmt.Sprintf("- %s: %s\n", k, v)
				}
				return fantasy.NewTextResponse(out), nil

			default:
				return fantasy.ToolResponse{}, fmt.Errorf("invalid action: %s. Use get, set, delete, or list", params.Action)
			}
		},
	)
}
