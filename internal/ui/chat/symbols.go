package chat

import (
	"encoding/json"

	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// SymbolsToolMessageItem renders symbols tool calls.
type SymbolsToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*SymbolsToolMessageItem)(nil)

// NewSymbolsToolMessageItem creates a new symbols tool message.
func NewSymbolsToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &SymbolsRenderContext{}, canceled)
}

// SymbolsRenderContext renders symbols tool messages.
type SymbolsRenderContext struct{}

// RenderTool implements the ToolRenderer interface.
func (r *SymbolsRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := "Symbols"

	if opts.IsPending() {
		return pendingTool(sty, name, opts.Anim, opts.Compact)
	}

	var params map[string]any
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	var headerParams []string
	if sym, ok := params["symbol"].(string); ok && sym != "" {
		headerParams = append(headerParams, "symbol: "+sym)
	}
	if lang, ok := params["language"].(string); ok && lang != "" {
		headerParams = append(headerParams, "lang: "+lang)
	}
	if typ, ok := params["type"].(string); ok && typ != "" {
		headerParams = append(headerParams, "type: "+typ)
	}
	if path, ok := params["path"].(string); ok && path != "" {
		headerParams = append(headerParams, "path: "+path)
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, headerParams...)
	if opts.Compact || !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputCodeContent(sty, "result.md", opts.Result.Content, 0, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
