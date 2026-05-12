package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"charm.land/fantasy"
)

//go:embed swarm.md
var swarmDescription []byte

const SwarmToolName = "swarm"

// SwarmParams defines the parameters for spawning a swarm of worker agents.
type SwarmParams struct {
	Task     string   `json:"task" description:"The main task to be decomposed and executed by the swarm"`
	Workers  int      `json:"workers,omitempty" description:"Number of parallel workers (default: 3, max: 5)"`
	Strategy string   `json:"strategy,omitempty" description:"Decomposition strategy: 'parallel' (default), 'divide_conquer', or 'compare'"`
	Subtasks []string `json:"subtasks,omitempty" description:"Optional: pre-defined subtasks. If empty, the task will be decomposed automatically."`
}

// NewSwarmTool creates a swarm tool that spawns parallel worker agents.
func NewSwarmTool(runSubAgent SubAgentRunner, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		SwarmToolName,
		string(swarmDescription),
		func(ctx context.Context, params SwarmParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Task == "" {
				return fantasy.NewTextErrorResponse("task is required"), nil
			}

			workers := params.Workers
			if workers <= 0 {
				workers = 3
			}
			if workers > 5 {
				workers = 5
			}

			strategy := params.Strategy
			if strategy == "" {
				strategy = "parallel"
			}

			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session id missing from context")
			}

			agentMessageID := GetMessageFromContext(ctx)
			if agentMessageID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("agent message id missing from context")
			}

			// Decompose task into subtasks if not provided
			subtasks := params.Subtasks
			if len(subtasks) == 0 {
				subtasks = decomposeTask(params.Task, workers, strategy)
			}

			// Limit to worker count
			if len(subtasks) > workers {
				subtasks = subtasks[:workers]
			}

			// Execute workers in parallel
			results := executeSwarm(ctx, runSubAgent, sessionID, agentMessageID, call.ID, subtasks, strategy)

			// Merge results based on strategy
			merged := mergeResults(results, strategy)

			return fantasy.NewTextResponse(merged), nil
		},
	)
}

// SubAgentRunner is the interface for running sub-agents.
type SubAgentRunner func(ctx context.Context, params SubAgentRunParams) (fantasy.ToolResponse, error)

// SubAgentRunParams holds the parameters for running a sub-agent.
type SubAgentRunParams struct {
	SessionID      string
	AgentMessageID string
	ToolCallID     string
	Prompt         string
	SessionTitle   string
}

// swarmWorkerResult holds the result from a single swarm worker.
type swarmWorkerResult struct {
	Index    int
	Task     string
	Response string
	Error    string
}

// decomposeTask breaks down a complex task into subtasks.
func decomposeTask(task string, numWorkers int, strategy string) []string {
	switch strategy {
	case "divide_conquer":
		return divideConquerDecompose(task, numWorkers)
	case "compare":
		return compareDecompose(task, numWorkers)
	default:
		return parallelDecompose(task, numWorkers)
	}
}

// parallelDecompose splits a task into parallel aspects.
func parallelDecompose(task string, n int) []string {
	aspects := []string{
		fmt.Sprintf("Analyze and research: %s", task),
		fmt.Sprintf("Implementation planning for: %s", task),
		fmt.Sprintf("Edge cases and error handling for: %s", task),
		fmt.Sprintf("Testing strategy for: %s", task),
		fmt.Sprintf("Documentation and examples for: %s", task),
	}

	if n > len(aspects) {
		n = len(aspects)
	}
	return aspects[:n]
}

// divideConquerDecompose splits a task into sequential phases.
func divideConquerDecompose(task string, n int) []string {
	phases := []string{
		fmt.Sprintf("Phase 1 - Research & Analysis: Understand requirements and constraints for: %s", task),
		fmt.Sprintf("Phase 2 - Design: Create architecture and design for: %s", task),
		fmt.Sprintf("Phase 3 - Implementation: Build the core solution for: %s", task),
		fmt.Sprintf("Phase 4 - Testing: Verify and validate the solution for: %s", task),
	}

	if n > len(phases) {
		n = len(phases)
	}
	return phases[:n]
}

// compareDecompose creates variations for comparison.
func compareDecompose(task string, n int) []string {
	approaches := []string{
		fmt.Sprintf("Approach A (conservative): Solve using established patterns - %s", task),
		fmt.Sprintf("Approach B (balanced): Solve using a mix of proven and modern techniques - %s", task),
		fmt.Sprintf("Approach C (innovative): Solve using cutting-edge or creative approaches - %s", task),
		fmt.Sprintf("Approach D (minimal): Solve with the simplest possible solution - %s", task),
		fmt.Sprintf("Approach E (comprehensive): Solve with full-featured, production-ready solution - %s", task),
	}

	if n > len(approaches) {
		n = len(approaches)
	}
	return approaches[:n]
}

// executeSwarm runs all workers in parallel and collects results.
func executeSwarm(ctx context.Context, runner SubAgentRunner, sessionID, agentMessageID, toolCallID string, subtasks []string, strategy string) []swarmWorkerResult {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []swarmWorkerResult
	)

	for i, subtask := range subtasks {
		wg.Add(1)
		go func(idx int, task string) {
			defer wg.Done()

			resp, err := runner(ctx, SubAgentRunParams{
				SessionID:      sessionID,
				AgentMessageID: agentMessageID,
				ToolCallID:     fmt.Sprintf("%s-worker-%d", toolCallID, idx),
				Prompt:         task,
				SessionTitle:   fmt.Sprintf("Swarm Worker %d", idx+1),
			})

			result := swarmWorkerResult{
				Index: idx,
				Task:  task,
			}

			if err != nil {
				result.Error = err.Error()
			} else {
				result.Response = resp.Content
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i, subtask)
	}

	wg.Wait()
	return results
}

// mergeResults combines worker results based on the strategy.
func mergeResults(results []swarmWorkerResult, strategy string) string {
	var sb strings.Builder

	switch strategy {
	case "compare":
		sb.WriteString("## Swarm Comparison Results\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("### %s\n\n", r.Task))
			if r.Error != "" {
				sb.WriteString(fmt.Sprintf("**Error:** %s\n\n", r.Error))
			} else {
				sb.WriteString(r.Response + "\n\n")
			}
		}
		sb.WriteString("---\n\n**Next Steps:** Review the approaches above and select the best one, or combine elements from multiple approaches.\n")

	case "divide_conquer":
		sb.WriteString("## Swarm Phased Results\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("### %s\n\n", r.Task))
			if r.Error != "" {
				sb.WriteString(fmt.Sprintf("**Error:** %s\n\n", r.Error))
			} else {
				sb.WriteString(r.Response + "\n\n")
			}
		}

	default:
		sb.WriteString("## Swarm Parallel Results\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("### Worker %d: %s\n\n", r.Index+1, r.Task))
			if r.Error != "" {
				sb.WriteString(fmt.Sprintf("**Error:** %s\n\n", r.Error))
			} else {
				sb.WriteString(r.Response + "\n\n")
			}
		}
		sb.WriteString("---\n\n**Summary:** All workers completed. Review each section above for comprehensive coverage.\n")
	}

	return sb.String()
}
