package chat

import (
	"encoding/json"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// CollabToolMessageItem renders collaboration tool calls.
type CollabToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*CollabToolMessageItem)(nil)

// NewCollabToolMessageItem creates a new collaboration tool message.
func NewCollabToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &CollabRenderContext{}, canceled)
}

// CollabRenderContext renders collaboration tool messages.
type CollabRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *CollabRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Collaborate"

	if opts.IsPending() {
		return pendingTool(sty, name, opts.Anim, opts.Compact)
	}

	var params map[string]any
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	var headerParams []string
	if action, ok := params["action"].(string); ok && action != "" {
		headerParams = append(headerParams, action)
	}
	if room, ok := params["room"].(string); ok && room != "" {
		headerParams = append(headerParams, "room: "+room)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "result.md", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
