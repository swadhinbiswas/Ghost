package plugin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Runtime manages the execution environment for a plugin.
type Runtime struct {
	manifest *Manifest
	cmd      *exec.Cmd
	stdout   bytes.Buffer
	stderr   bytes.Buffer
	mu       sync.Mutex
	stopped  bool
	done     chan struct{}
}

// NewRuntime creates a new runtime for a plugin.
func NewRuntime(manifest *Manifest) (*Runtime, error) {
	r := &Runtime{
		manifest: manifest,
		done:     make(chan struct{}),
	}
	return r, nil
}

// Execute runs the plugin's entrypoint with the given input.
func (r *Runtime) Execute(ctx context.Context, input string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return "", fmt.Errorf("plugin runtime is stopped")
	}

	entryPath := r.manifest.EntryPath()
	if entryPath == "" {
		return "", fmt.Errorf("no entrypoint defined for plugin")
	}

	if _, err := os.Stat(entryPath); err != nil {
		return "", fmt.Errorf("entrypoint not found: %s", entryPath)
	}

	// Build command
	cmd, err := r.buildCommand(ctx, entryPath, input)
	if err != nil {
		return "", err
	}

	r.cmd = cmd
	cmd.Stdout = &r.stdout
	cmd.Stderr = &r.stderr

	// Set up environment
	cmd.Env = r.buildEnvironment()

	// Set working directory
	cmd.Dir = r.manifest.InstallPath

	// Apply resource limits
	r.applyResourceLimits(cmd)

	// Run with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, r.manifest.TimeoutDuration())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		output := r.stdout.String()
		if err != nil {
			stderr := r.stderr.String()
			if stderr != "" {
				return output, fmt.Errorf("plugin execution failed: %s\nstderr: %s", err, stderr)
			}
			return output, fmt.Errorf("plugin execution failed: %s", err)
		}
		return strings.TrimSpace(output), nil

	case <-timeoutCtx.Done():
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return "", fmt.Errorf("plugin execution timed out after %s", r.manifest.TimeoutDuration())
	}
}

// ExecuteScript runs an arbitrary script from the plugin directory.
func (r *Runtime) ExecuteScript(ctx context.Context, scriptPath string, args ...string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return "", fmt.Errorf("plugin runtime is stopped")
	}

	// Ensure script is within plugin directory
	fullPath := filepath.Join(r.manifest.InstallPath, scriptPath)
	if !strings.HasPrefix(fullPath, r.manifest.InstallPath) {
		return "", fmt.Errorf("script path escapes plugin directory")
	}

	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("script not found: %s", fullPath)
	}

	cmd := exec.CommandContext(ctx, fullPath, args...)
	cmd.Dir = r.manifest.InstallPath
	cmd.Env = r.buildEnvironment()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	timeoutCtx, cancel := context.WithTimeout(ctx, r.manifest.TimeoutDuration())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		output := stdout.String()
		if err != nil {
			return output, fmt.Errorf("script failed: %s\nstderr: %s", err, stderr.String())
		}
		return strings.TrimSpace(output), nil

	case <-timeoutCtx.Done():
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return "", fmt.Errorf("script timed out after %s", r.manifest.TimeoutDuration())
	}
}

// buildCommand creates an exec.Cmd for the plugin entrypoint.
func (r *Runtime) buildCommand(ctx context.Context, entryPath, input string) (*exec.Cmd, error) {
	runtime := r.manifest.Runtime
	args := r.manifest.Args

	// Determine how to run the entrypoint
	switch runtime {
	case "":
		// Auto-detect from extension
		ext := filepath.Ext(entryPath)
		switch ext {
		case ".sh", ".bash":
			return exec.CommandContext(ctx, "bash", append([]string{entryPath}, args...)...), nil
		case ".py":
			return exec.CommandContext(ctx, "python3", append([]string{entryPath}, args...)...), nil
		case ".rb":
			return exec.CommandContext(ctx, "ruby", append([]string{entryPath}, args...)...), nil
		case ".js":
			return exec.CommandContext(ctx, "node", append([]string{entryPath}, args...)...), nil
		case ".ts":
			return exec.CommandContext(ctx, "npx", append([]string{"ts-node", entryPath}, args...)...), nil
		case ".go":
			return exec.CommandContext(ctx, "go", append([]string{"run", entryPath}, args...)...), nil
		default:
			// Try to run directly
			return exec.CommandContext(ctx, entryPath, args...), nil
		}

	case "python", "python3":
		return exec.CommandContext(ctx, "python3", append([]string{entryPath}, args...)...), nil
	case "node":
		return exec.CommandContext(ctx, "node", append([]string{entryPath}, args...)...), nil
	case "bash":
		return exec.CommandContext(ctx, "bash", append([]string{entryPath}, args...)...), nil
	case "ruby":
		return exec.CommandContext(ctx, "ruby", append([]string{entryPath}, args...)...), nil
	case "go":
		return exec.CommandContext(ctx, "go", append([]string{"run", entryPath}, args...)...), nil
	case "binary":
		return exec.CommandContext(ctx, entryPath, args...), nil
	default:
		return exec.CommandContext(ctx, runtime, append([]string{entryPath}, args...)...), nil
	}
}

// buildEnvironment builds the environment variables for the plugin.
func (r *Runtime) buildEnvironment() []string {
	env := os.Environ()

	// Add plugin-specific variables
	env = append(env,
		fmt.Sprintf("GHOST_PLUGIN_NAME=%s", r.manifest.Name),
		fmt.Sprintf("GHOST_PLUGIN_VERSION=%s", r.manifest.Version),
		fmt.Sprintf("GHOST_PLUGIN_DIR=%s", r.manifest.InstallPath),
	)

	// Add custom env from manifest
	for key, value := range r.manifest.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Filter allowed env vars
	if len(r.manifest.Permissions.EnvVars) > 0 {
		var filtered []string
		allowedMap := make(map[string]bool)
		for _, allowed := range r.manifest.Permissions.EnvVars {
			allowedMap[allowed] = true
			allowedMap[strings.ToLower(allowed)] = true
		}
		for _, e := range env {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) > 0 && allowedMap[parts[0]] {
				filtered = append(filtered, e)
			}
		}
		env = filtered
	}

	return env
}

// applyResourceLimits applies resource constraints to the command.
func (r *Runtime) applyResourceLimits(cmd *exec.Cmd) {
	// Note: Full resource limiting (cgroups, ulimit) requires platform-specific code.
	// For now, we rely on context timeout and manual process monitoring.

	// On Linux, we could use prlimit or setrlimit, but that's complex.
	// A simpler approach is to monitor memory usage in a goroutine.
	if r.manifest.Permissions.MaxMemoryMB > 0 {
		go r.monitorMemory(cmd)
	}
}

// monitorMemory watches the process memory usage and kills it if it exceeds the limit.
func (r *Runtime) monitorMemory(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	limitBytes := r.manifest.MaxMemoryBytes()

	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				return
			}
			// Memory monitoring is platform-specific and complex.
			// This is a placeholder for future implementation.
			_ = limitBytes
		}
	}
}

// Stop gracefully stops the plugin runtime.
func (r *Runtime) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return
	}
	r.stopped = true

	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
	}

	close(r.done)
}

// IsStopped returns whether the runtime is stopped.
func (r *Runtime) IsStopped() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopped
}
