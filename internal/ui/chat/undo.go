package chat

import (
	"encoding/json"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// UndoToolMessageItem renders undo tool calls.
type UndoToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*UndoToolMessageItem)(nil)

// NewUndoToolMessageItem creates a new undo tool message.
func NewUndoToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &UndoRenderContext{}, canceled)
}

// UndoRenderContext renders undo tool messages.
type UndoRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *UndoRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Undo"

	if opts.IsPending() {
		return pendingTool(sty, name, opts.Anim, opts.Compact)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputPlainContent(sty, opts.Result.Content, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}

// TestToolMessageItem renders test runner tool calls.
type TestToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*TestToolMessageItem)(nil)

// NewTestToolMessageItem creates a new test tool message.
func NewTestToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &TestRenderContext{}, canceled)
}

// TestRenderContext renders test tool messages.
type TestRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *TestRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Test"

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
	if framework, ok := params["framework"].(string); ok && framework != "" {
		headerParams = append(headerParams, "framework: "+framework)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "test-output", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
