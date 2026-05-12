package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/feedback"
)

//go:embed feedback_tool.md
var feedbackToolDescription []byte

const FeedbackToolName = "feedback"

// FeedbackParams defines the parameters for the feedback tool.
type FeedbackParams struct {
	Action    string `json:"action" description:"Action: 'rate', 'comment', 'stats', 'insights', 'list', 'clear'"`
	Rating    int    `json:"rating,omitempty" description:"Rating: 1 (positive), -1 (negative), 0 (neutral)"`
	MessageID string `json:"message_id,omitempty" description:"Message ID to rate"`
	Comment   string `json:"comment,omitempty" description:"Feedback comment"`
	Category  string `json:"category,omitempty" description:"Feedback category: accuracy, speed, tone, completeness, code_quality"`
	Limit     int    `json:"limit,omitempty" description:"Number of recent feedback to show (default: 10)"`
}

// NewFeedbackTool creates a tool for collecting and managing user feedback.
func NewFeedbackTool(store *feedback.Store) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		FeedbackToolName,
		string(feedbackToolDescription),
		func(ctx context.Context, params FeedbackParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Action == "" {
				return fantasy.NewTextErrorResponse("action is required"), nil
			}

			switch params.Action {
			case "rate":
				if params.MessageID == "" {
					return fantasy.NewTextErrorResponse("message_id is required for rate action"), nil
				}
				rating := feedback.Rating(params.Rating)
				if rating != feedback.RatingPositive && rating != feedback.RatingNegative && rating != feedback.RatingNeutral {
					return fantasy.NewTextErrorResponse("rating must be 1 (positive), -1 (negative), or 0 (neutral)"), nil
				}

				f := &feedback.Feedback{
					SessionID: getSessionID(ctx),
					MessageID: params.MessageID,
					Rating:    rating,
					Comment:   params.Comment,
					Category:  params.Category,
				}

				if err := store.Add(f); err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to save feedback: %s", err)), nil
				}

				return fantasy.NewTextResponse("Feedback recorded. Thank you!"), nil

			case "comment":
				if params.MessageID == "" {
					return fantasy.NewTextErrorResponse("message_id is required for comment action"), nil
				}

				// Check if feedback already exists
				existing := store.GetByMessage(params.MessageID)
				if existing != nil {
					existing.Comment = params.Comment
					if params.Category != "" {
						existing.Category = params.Category
					}
					if err := store.Save(); err != nil {
						return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to update feedback: %s", err)), nil
					}
					return fantasy.NewTextResponse("Feedback updated."), nil
				}

				// Create new feedback with neutral rating
				f := &feedback.Feedback{
					SessionID: getSessionID(ctx),
					MessageID: params.MessageID,
					Rating:    feedback.RatingNeutral,
					Comment:   params.Comment,
					Category:  params.Category,
				}

				if err := store.Add(f); err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to save feedback: %s", err)), nil
				}

				return fantasy.NewTextResponse("Comment saved."), nil

			case "stats":
				stats := store.GetStats()
				var sb strings.Builder
				sb.WriteString("## Feedback Statistics\n\n")
				sb.WriteString(fmt.Sprintf("**Total Feedback:** %d\n", stats.Total))
				sb.WriteString(fmt.Sprintf("**Positive:** %d (%.1f%%)\n", stats.Positive, float64(stats.Positive)/float64(max(stats.Total, 1))*100))
				sb.WriteString(fmt.Sprintf("**Negative:** %d (%.1f%%)\n", stats.Negative, float64(stats.Negative)/float64(max(stats.Total, 1))*100))
				sb.WriteString(fmt.Sprintf("**Neutral:** %d\n", stats.Neutral))
				sb.WriteString(fmt.Sprintf("**Satisfaction Rate:** %.1f%%\n", stats.SatisfactionRate))

				if len(stats.ByCategory) > 0 {
					sb.WriteString("\n**By Category:**\n")
					for cat, count := range stats.ByCategory {
						sb.WriteString(fmt.Sprintf("- %s: %d\n", cat, count))
					}
				}

				return fantasy.NewTextResponse(sb.String()), nil

			case "insights":
				insights := store.GetPatternInsights()
				if len(insights) == 0 {
					return fantasy.NewTextResponse("No patterns detected yet. More feedback needed for insights."), nil
				}

				var sb strings.Builder
				sb.WriteString("## Feedback Insights\n\n")
				for _, insight := range insights {
					sb.WriteString(fmt.Sprintf("- %s\n", insight))
				}
				return fantasy.NewTextResponse(sb.String()), nil

			case "list":
				limit := params.Limit
				if limit <= 0 {
					limit = 10
				}

				recent := store.GetRecent(limit)
				if len(recent) == 0 {
					return fantasy.NewTextResponse("No feedback recorded yet."), nil
				}

				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("## Recent Feedback (%d)\n\n", len(recent)))
				for _, f := range recent {
					ratingStr := "neutral"
					if f.Rating == feedback.RatingPositive {
						ratingStr = "positive"
					} else if f.Rating == feedback.RatingNegative {
						ratingStr = "negative"
					}

					sb.WriteString(fmt.Sprintf("- [%s] %s", ratingStr, f.Timestamp.Format("2006-01-02 15:04")))
					if f.Category != "" {
						sb.WriteString(fmt.Sprintf(" (%s)", f.Category))
					}
					sb.WriteString("\n")
					if f.Comment != "" {
						sb.WriteString(fmt.Sprintf("  Comment: %s\n", f.Comment))
					}
				}
				return fantasy.NewTextResponse(sb.String()), nil

			case "clear":
				if err := store.Clear(); err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to clear feedback: %s", err)), nil
				}
				return fantasy.NewTextResponse("All feedback cleared."), nil

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown action: %s", params.Action)), nil
			}
		},
	)
}

func getSessionID(ctx context.Context) string {
	if sid, ok := ctx.Value("session_id").(string); ok {
		return sid
	}
	return ""
}

// FormatFeedbackForPrompt formats feedback data for injection into the system prompt.
func FormatFeedbackForPrompt(store *feedback.Store) string {
	stats := store.GetStats()
	if stats.Total == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n<user_feedback_history>\n")
	sb.WriteString(fmt.Sprintf("Total feedback received: %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("Satisfaction rate: %.1f%%\n", stats.SatisfactionRate))

	insights := store.GetPatternInsights()
	if len(insights) > 0 {
		sb.WriteString("\nPatterns identified:\n")
		for _, insight := range insights {
			sb.WriteString(fmt.Sprintf("- %s\n", insight))
		}
	}

	sb.WriteString("\nUse this feedback to improve your responses. Pay attention to negative feedback patterns.\n")
	sb.WriteString("</user_feedback_history>\n")

	return sb.String()
}
