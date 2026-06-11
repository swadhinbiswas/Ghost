package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"charm.land/fantasy"
)

const SelfHealingPrompt = `

# Self-Healing Verification Loops
When you modify files using the "edit", "write", "multiedit", or "atomic_edit" tools, Ghost will run a verification command (e.g. compilation/test suites) in the background.
If the verification command fails, you will receive a tool error response structured like this:
<verification_failure>
Your changes broke the build. Here is the compiler/test error:
[error output]
Please use the edit/write tools to fix the errors and run the compilation/verification again.
</verification_failure>

If you receive this error, do NOT ask the user for help or stop. You must analyze the error output, figure out the cause of the failure (e.g., compile error, import issue, syntax mistake, type mismatch, or failed unit test), and call the appropriate edit/write tools to fix it.
`

// VerifyBuild runs the verification command and returns build/test output if it fails.
func (a *sessionAgent) VerifyBuild(ctx context.Context) (bool, string, error) {
	if a.cfg == nil {
		return true, "", nil
	}

	cmdStr := a.cfg.Config().Options.VerificationCommand
	if cmdStr == "" {
		return true, "", nil
	}

	slog.Info("Running verification build/test command", "command", cmdStr)

	// Create context with a timeout so a hanging test/build command doesn't block forever
	verifyCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	// Use shell to run the command to support complex pipelines (e.g. "npm run build && npm run test")
	if strings.Contains(cmdStr, " ") || strings.Contains(cmdStr, "&&") || strings.Contains(cmdStr, "||") || strings.Contains(cmdStr, "|") {
		cmd = exec.CommandContext(verifyCtx, "bash", "-c", cmdStr)
	} else {
		cmd = exec.CommandContext(verifyCtx, cmdStr)
	}

	cmd.Dir = a.cfg.WorkingDir()

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("Verification command failed", "error", err, "output", string(out))
		return false, string(out), nil
	}

	slog.Info("Verification command passed successfully")
	return true, "", nil
}

type verificationWrappedTool struct {
	fantasy.AgentTool
	a         *sessionAgent
	sessionID string
}

func (w *verificationWrappedTool) Run(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error) {
	resp, err := w.AgentTool.Run(ctx, params)
	if err != nil || resp.IsError {
		return resp, err
	}

	if w.a.cfg == nil {
		return resp, nil
	}

	cmdStr := w.a.cfg.Config().Options.VerificationCommand
	if cmdStr == "" {
		return resp, nil
	}

	maxRetries := w.a.cfg.Config().Options.MaxVerificationRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	w.a.verificationMu.Lock()
	retries := w.a.verificationRetries[w.sessionID]
	w.a.verificationMu.Unlock()

	if retries >= maxRetries {
		slog.Warn("Max verification retries reached, returning original result", "retries", retries, "sessionID", w.sessionID)
		return resp, nil
	}

	// Run verification command
	ok, out, err := w.a.VerifyBuild(ctx)
	if err != nil {
		return resp, err
	}

	if ok {
		return resp, nil
	}

	// Increment retry counter
	w.a.verificationMu.Lock()
	w.a.verificationRetries[w.sessionID] = retries + 1
	w.a.verificationMu.Unlock()

	// Return verification failure XML response block
	failMsg := fmt.Sprintf("<verification_failure>\nYour changes broke the build. Here is the compiler/test error:\n%s\nPlease use the edit/write tools to fix the errors and run the compilation/verification again.\n</verification_failure>", out)

	return fantasy.NewTextErrorResponse(failMsg), nil
}

func (a *sessionAgent) wrapToolsWithVerification(ctx context.Context, sessionID string, tools []fantasy.AgentTool) []fantasy.AgentTool {
	if a.cfg == nil {
		return tools
	}
	cmdStr := a.cfg.Config().Options.VerificationCommand
	if cmdStr == "" {
		return tools
	}

	wrapped := make([]fantasy.AgentTool, len(tools))
	for i, t := range tools {
		name := t.Info().Name
		if name == "edit" || name == "write" || name == "multiedit" || name == "atomic_edit" {
			wrapped[i] = &verificationWrappedTool{
				AgentTool: t,
				a:         a,
				sessionID: sessionID,
			}
		} else {
			wrapped[i] = t
		}
	}
	return wrapped
}
