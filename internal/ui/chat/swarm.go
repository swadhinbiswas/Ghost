package chat

import (
	"encoding/json"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// SwarmToolMessageItem renders swarm tool calls.
type SwarmToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*SwarmToolMessageItem)(nil)

// NewSwarmToolMessageItem creates a new swarm tool message.
func NewSwarmToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &SwarmRenderContext{}, canceled)
}

// SwarmRenderContext renders swarm tool messages.
type SwarmRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *SwarmRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Swarm"

	if opts.IsPending() {
		return pendingTool(sty, name, opts.Anim, opts.Compact)
	}

	var params map[string]any
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	task := ""
	if t, ok := params["task"].(string); ok {
		task = t
	}
	strategy := ""
	if s, ok := params["strategy"].(string); ok {
		strategy = s
	}
	workers := ""
	if w, ok := params["workers"].(float64); ok {
		workers = string(rune(int(w) + '0'))
	}

	var headerParams []string
	if task != "" {
		if len(task) > 60 {
			task = task[:57] + "..."
		}
		headerParams = append(headerParams, "task: "+task)
	}
	if strategy != "" {
		headerParams = append(headerParams, "strategy: "+strategy)
	}
	if workers != "" {
		headerParams = append(headerParams, workers+" workers")
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "result.md", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
