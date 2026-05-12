package chat

import (
	"encoding/json"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// FeedbackToolMessageItem renders feedback tool calls.
type FeedbackToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*FeedbackToolMessageItem)(nil)

// NewFeedbackToolMessageItem creates a new feedback tool message.
func NewFeedbackToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &FeedbackRenderContext{}, canceled)
}

// FeedbackRenderContext renders feedback tool messages.
type FeedbackRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *FeedbackRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Feedback"

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
	if rating, ok := params["rating"].(float64); ok {
		if rating > 0 {
			headerParams = append(headerParams, "positive")
		} else if rating < 0 {
			headerParams = append(headerParams, "negative")
		}
	}
	if cat, ok := params["category"].(string); ok && cat != "" {
		headerParams = append(headerParams, "category: "+cat)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "result.md", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
