package agent

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/config"
)

type mockAgentTool struct {
	name    string
	runFunc func(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error)
	opts    fantasy.ProviderOptions
}

func (m *mockAgentTool) Info() fantasy.ToolInfo {
	return fantasy.ToolInfo{
		Name: m.name,
	}
}

func (m *mockAgentTool) Run(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, params)
	}
	return fantasy.NewTextResponse("success"), nil
}

func (m *mockAgentTool) ProviderOptions() fantasy.ProviderOptions {
	return m.opts
}

func (m *mockAgentTool) SetProviderOptions(opts fantasy.ProviderOptions) {
	m.opts = opts
}

func TestWrapToolsWithVerification(t *testing.T) {
	tempDir := t.TempDir()
	store, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	store.Config().Options = &config.Options{
		VerificationCommand:    "echo 'verified'",
		MaxVerificationRetries: 3,
	}

	a := &sessionAgent{
		cfg:                 store,
		verificationRetries: make(map[string]int),
	}

	tools := []fantasy.AgentTool{
		&mockAgentTool{name: "edit"},
		&mockAgentTool{name: "write"},
		&mockAgentTool{name: "ls"},
	}

	wrapped := a.wrapToolsWithVerification(context.Background(), "session-1", tools)
	if len(wrapped) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(wrapped))
	}

	// Only edit and write should be wrapped
	if _, ok := wrapped[0].(*verificationWrappedTool); !ok {
		t.Errorf("expected tool 0 (edit) to be wrapped, got %T", wrapped[0])
	}
	if _, ok := wrapped[1].(*verificationWrappedTool); !ok {
		t.Errorf("expected tool 1 (write) to be wrapped, got %T", wrapped[1])
	}
	if _, ok := wrapped[2].(*verificationWrappedTool); ok {
		t.Errorf("expected tool 2 (ls) not to be wrapped, got %T", wrapped[2])
	}
}

func TestVerificationWrappedTool_Success(t *testing.T) {
	tempDir := t.TempDir()
	store, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Always passes
	store.Config().Options = &config.Options{
		VerificationCommand:    "echo 'passed'",
		MaxVerificationRetries: 3,
	}

	a := &sessionAgent{
		cfg:                 store,
		verificationRetries: make(map[string]int),
	}

	mt := &mockAgentTool{
		name: "edit",
		runFunc: func(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return fantasy.NewTextResponse("tool-success"), nil
		},
	}

	wt := &verificationWrappedTool{
		AgentTool: mt,
		a:         a,
		sessionID: "session-1",
	}

	resp, err := wt.Run(context.Background(), fantasy.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IsError {
		t.Errorf("expected success response, got error: %s", resp.Content)
	}

	if resp.Content != "tool-success" {
		t.Errorf("expected content 'tool-success', got '%s'", resp.Content)
	}

	if a.verificationRetries["session-1"] != 0 {
		t.Errorf("expected 0 retries, got %d", a.verificationRetries["session-1"])
	}
}

func TestVerificationWrappedTool_FailureAndRetry(t *testing.T) {
	tempDir := t.TempDir()
	store, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Always fails (false returns exit code 1)
	store.Config().Options = &config.Options{
		VerificationCommand:    "false",
		MaxVerificationRetries: 2,
	}

	a := &sessionAgent{
		cfg:                 store,
		verificationRetries: make(map[string]int),
	}

	mt := &mockAgentTool{
		name: "edit",
		runFunc: func(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return fantasy.NewTextResponse("tool-success"), nil
		},
	}

	wt := &verificationWrappedTool{
		AgentTool: mt,
		a:         a,
		sessionID: "session-1",
	}

	// 1st retry
	resp, err := wt.Run(context.Background(), fantasy.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsError {
		t.Errorf("expected error response because verification fails")
	}
	if a.verificationRetries["session-1"] != 1 {
		t.Errorf("expected 1 retry, got %d", a.verificationRetries["session-1"])
	}

	// 2nd retry
	resp, err = wt.Run(context.Background(), fantasy.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsError {
		t.Errorf("expected error response because verification fails")
	}
	if a.verificationRetries["session-1"] != 2 {
		t.Errorf("expected 2 retries, got %d", a.verificationRetries["session-1"])
	}

	// 3rd run: max retries reached, should bypass and return original success response
	resp, err = wt.Run(context.Background(), fantasy.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsError {
		t.Errorf("expected success response because max retries reached")
	}
	if resp.Content != "tool-success" {
		t.Errorf("expected 'tool-success', got '%s'", resp.Content)
	}
}
