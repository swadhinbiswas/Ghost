package chat

import (
	"encoding/json"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// PluginToolMessageItem renders plugin tool calls.
type PluginToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*PluginToolMessageItem)(nil)

// NewPluginToolMessageItem creates a new plugin tool message.
func NewPluginToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &PluginRenderContext{}, canceled)
}

// PluginRenderContext renders plugin tool messages.
type PluginRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *PluginRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Plugins"

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
	if pluginName, ok := params["name"].(string); ok && pluginName != "" {
		headerParams = append(headerParams, "plugin: "+pluginName)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "result.md", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
