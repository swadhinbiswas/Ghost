package chat

import (
	"encoding/json"
	"strings"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// ScreenshotToolMessageItem renders screenshot tool calls.
type ScreenshotToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*ScreenshotToolMessageItem)(nil)

// NewScreenshotToolMessageItem creates a new screenshot tool message.
func NewScreenshotToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &ScreenshotRenderContext{}, canceled)
}

// ScreenshotRenderContext renders screenshot tool messages.
type ScreenshotRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *ScreenshotRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Screenshot"

	if opts.IsPending() {
		return pendingTool(sty, name, opts.Anim, opts.Compact)
	}

	var params map[string]any
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	var headerParams []string
	if mode, ok := params["mode"].(string); ok && mode != "" {
		headerParams = append(headerParams, "mode: "+mode)
	}
	if path, ok := params["path"].(string); ok && path != "" {
		headerParams = append(headerParams, "path: "+path)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	// Handle image data if present
	if opts.Result.Data != "" && strings.HasPrefix(opts.Result.MIMEType, "image/") {
		body := sty.Tool.Body.Render(toolOutputImageContent(sty, opts.Result.Data, opts.Result.MIMEType))
		return joinToolParts(header, body)
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "result.md", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
