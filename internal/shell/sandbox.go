package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type SandboxType string

const (
	SandboxNone   SandboxType = "none"
	SandboxDocker SandboxType = "docker"
)

var (
	sandboxMu      sync.Mutex
	containerCache = make(map[string]string) // map working dir to container ID
)

// SandboxManager manages sandboxed execution of commands
type SandboxManager struct {
	Type       SandboxType
	Image      string
	WorkingDir string
}

// NewSandboxManager creates a new SandboxManager
func NewSandboxManager(sandboxType SandboxType, image string, workingDir string) *SandboxManager {
	if image == "" {
		image = "golang:latest"
	}
	return &SandboxManager{
		Type:       sandboxType,
		Image:      image,
		WorkingDir: workingDir,
	}
}

// getOrCreateContainer ensures a docker container is running for the current working directory
func (s *SandboxManager) getOrCreateContainer(ctx context.Context) (string, error) {
	sandboxMu.Lock()
	defer sandboxMu.Unlock()

	containerID, exists := containerCache[s.WorkingDir]
	if exists {
		// check if it's still running
		cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", containerID)
		out, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(out)) == "true" {
			return containerID, nil
		}
		// container stopped or doesn't exist, we'll create a new one
	}

	// Create a new container
	// Mount the workspace into /workspace and run an idle loop
	containerName := fmt.Sprintf("ghost-sandbox-%x", len(s.WorkingDir))

	// Try to remove old container if it exists
	exec.Command("docker", "rm", "-f", containerName).Run()

	cmd := exec.CommandContext(ctx, "docker", "run", "-d",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", s.WorkingDir),
		"-w", "/workspace",
		s.Image,
		"tail", "-f", "/dev/null", // keep alive
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to start sandbox container: %w, output: %s", err, string(out))
	}

	newID := strings.TrimSpace(string(out))
	containerCache[s.WorkingDir] = newID
	return newID, nil
}

// ExecStream executes a command inside the sandbox
func (s *SandboxManager) ExecStream(ctx context.Context, command string, stdout io.Writer, stderr io.Writer) error {
	if s.Type != SandboxDocker {
		return fmt.Errorf("sandbox type %s not supported", s.Type)
	}

	containerID, err := s.getOrCreateContainer(ctx)
	if err != nil {
		return err
	}

	// Execute command using docker exec
	cmd := exec.CommandContext(ctx, "docker", "exec", containerID, "bash", "-c", command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

// Exec executes a command inside the sandbox and returns its output
func (s *SandboxManager) Exec(ctx context.Context, command string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	err := s.ExecStream(ctx, command, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}
