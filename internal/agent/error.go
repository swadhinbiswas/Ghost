package agent

import "errors"

var (
	ErrRequestCancelled    = errors.New("request cancelled by user")
	ErrSessionBusy         = errors.New("session is currently processing another request please wait")
	ErrEmptyPrompt         = errors.New("prompt cannot be empty")
	ErrSessionMissing      = errors.New("session id missing from context")
	ErrAgentMessageMissing = errors.New("agent message id missing from context")
	ErrToolExecutionFailed = errors.New("tool execution failed")
)
