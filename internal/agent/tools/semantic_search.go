// Package tools provides the semantic_search agent tool.
package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/semantic"
)

type SemanticSearchParams struct {
	Query string `json:"query" description:"Natural language description of what you're looking for (e.g., 'authentication middleware', 'database connection pool', 'error handling in API routes')"`
	Limit int    `json:"limit,omitempty" description:"Maximum number of results to return (default 10)"`
}

type SemanticSearchResponseMetadata struct {
	TotalResults int      `json:"total_results"`
	TopScore     float32  `json:"top_score"`
	Files        []string `json:"files"`
}

const SemanticSearchToolName = "semantic_search"

//go:embed semantic_search.md
var semanticSearchDescription []byte

func NewSemanticSearchTool(index *semantic.Index, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		SemanticSearchToolName,
		string(semanticSearchDescription),
		func(ctx context.Context, params SemanticSearchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" {
				return fantasy.NewTextErrorResponse("query is required. Describe what code you're looking for in natural language."), nil
			}

			if !index.IsBuilt() {
				return fantasy.NewTextErrorResponse("semantic index is not yet built. Please wait a moment and try again, or use grep/glob for keyword search."), nil
			}

			limit := params.Limit
			if limit <= 0 {
				limit = 10
			}

			results := index.Search(ctx, params.Query, limit)
			if len(results) == 0 {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("no results found for query: %q", params.Query)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Found %d relevant code snippets for: %q\n\n", len(results), params.Query))

			files := make(map[string]bool)
			topScore := float32(0)

			for i, r := range results {
				if r.Score > topScore {
					topScore = r.Score
				}
				files[r.FilePath] = true

				scorePct := int(r.Score * 100)
				sb.WriteString(fmt.Sprintf("--- Result %d (relevance: %d%%) ---\n", i+1, scorePct))
				sb.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n", r.FilePath, r.StartLine, r.EndLine))
				sb.WriteString(fmt.Sprintf("Language: %s\n\n", r.Language))
				sb.WriteString(r.Content)
				sb.WriteString("\n\n")
			}

			fileList := make([]string, 0, len(files))
			for f := range files {
				fileList = append(fileList, f)
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(sb.String()),
				SemanticSearchResponseMetadata{
					TotalResults: len(results),
					TopScore:     topScore,
					Files:        fileList,
				},
			), nil
		},
	)
}
