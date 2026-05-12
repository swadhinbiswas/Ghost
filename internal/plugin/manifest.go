package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	manifestFile = "plugin.json"
)

// PluginType defines the type of plugin.
type PluginType string

const (
	PluginTypeTool   PluginType = "tool"
	PluginTypeHook   PluginType = "hook"
	PluginTypeTheme  PluginType = "theme"
	PluginTypeScript PluginType = "script"
	PluginTypeAgent  PluginType = "agent"
)

// Manifest defines the plugin metadata and configuration.
type Manifest struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description,omitempty"`
	Author      string     `json:"author,omitempty"`
	License     string     `json:"license,omitempty"`
	Homepage    string     `json:"homepage,omitempty"`
	Type        PluginType `json:"type"`
	Icon        string     `json:"icon,omitempty"`

	// Entry points
	Main       string `json:"main,omitempty"`
	Entrypoint string `json:"entrypoint,omitempty"`

	// Runtime configuration
	Runtime string            `json:"runtime,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`

	// Permissions
	Permissions Permissions `json:"permissions,omitempty"`

	// Hooks (for hook-type plugins)
	Hooks []HookDef `json:"hooks,omitempty"`

	// Tool definition (for tool-type plugins)
	ToolDef *ToolDef `json:"tool,omitempty"`

	// Agent definition (for agent-type plugins)
	AgentDef *AgentDef `json:"agent,omitempty"`

	// Dependencies
	Dependencies map[string]string `json:"dependencies,omitempty"`

	// Compatibility
	MinGhostVersion string   `json:"min_ghost_version,omitempty"`
	MaxGhostVersion string   `json:"max_ghost_version,omitempty"`
	OS              []string `json:"os,omitempty"`
	Arch            []string `json:"arch,omitempty"`

	// Internal state
	InstallPath string `json:"-"`
	Enabled     bool   `json:"-"`
	Loaded      bool   `json:"-"`
	LoadError   string `json:"-"`
}

// Permissions defines what a plugin is allowed to do.
type Permissions struct {
	Network      bool          `json:"network,omitempty"`
	FileSystem   bool          `json:"filesystem,omitempty"`
	AllowedPaths []string      `json:"allowed_paths,omitempty"`
	EnvVars      []string      `json:"allowed_env,omitempty"`
	Commands     []string      `json:"allowed_commands,omitempty"`
	MaxMemoryMB  int           `json:"max_memory_mb,omitempty"`
	MaxTimeout   time.Duration `json:"max_timeout,omitempty"`
}

// HookDef defines a hook that a plugin can register.
type HookDef struct {
	Name    string `json:"name"`
	Trigger string `json:"trigger"`
	Script  string `json:"script,omitempty"`
}

// ToolDef defines a tool provided by a plugin.
type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  string `json:"parameters,omitempty"`
	Script      string `json:"script,omitempty"`
}

// AgentDef defines an agent provided by a plugin.
type AgentDef struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Model        string            `json:"model,omitempty"`
	Tools        []string          `json:"tools,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
}

// LoadManifest reads and validates a plugin manifest from the given directory.
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, manifestFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Set install path
	manifest.InstallPath = dir

	// Validate required fields
	if err := manifest.Validate(); err != nil {
		return nil, err
	}

	// Check compatibility
	if err := manifest.CheckCompatibility(); err != nil {
		manifest.LoadError = err.Error()
	}

	return &manifest, nil
}

// Validate checks that the manifest has all required fields.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if m.Type == "" {
		return fmt.Errorf("plugin type is required (tool|hook|theme|script|agent)")
	}

	// Check type-specific requirements
	switch m.Type {
	case PluginTypeTool:
		if m.ToolDef == nil || m.ToolDef.Name == "" {
			return fmt.Errorf("tool plugins must define 'tool.name'")
		}
	case PluginTypeHook:
		if len(m.Hooks) == 0 {
			return fmt.Errorf("hook plugins must define at least one hook")
		}
	case PluginTypeAgent:
		if m.AgentDef == nil || m.AgentDef.Name == "" {
			return fmt.Errorf("agent plugins must define 'agent.name'")
		}
	}

	// Validate version format
	if !isValidSemver(m.Version) {
		return fmt.Errorf("invalid version format: %s (expected semver)", m.Version)
	}

	return nil
}

// CheckCompatibility checks if the plugin is compatible with the current environment.
func (m *Manifest) CheckCompatibility() error {
	// Check OS
	if len(m.OS) > 0 {
		match := false
		for _, os := range m.OS {
			if strings.EqualFold(os, runtime.GOOS) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("plugin not compatible with OS: %s", runtime.GOOS)
		}
	}

	// Check architecture
	if len(m.Arch) > 0 {
		match := false
		for _, arch := range m.Arch {
			if strings.EqualFold(arch, runtime.GOARCH) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("plugin not compatible with architecture: %s", runtime.GOARCH)
		}
	}

	return nil
}

// EntryPath returns the path to the plugin's entrypoint script.
func (m *Manifest) EntryPath() string {
	if m.Entrypoint != "" {
		return filepath.Join(m.InstallPath, m.Entrypoint)
	}
	if m.Main != "" {
		return filepath.Join(m.InstallPath, m.Main)
	}
	return ""
}

// HasPermission checks if the plugin has a specific permission.
func (m *Manifest) HasPermission(perm string) bool {
	switch perm {
	case "network":
		return m.Permissions.Network
	case "filesystem":
		return m.Permissions.FileSystem
	default:
		return false
	}
}

// IsPathAllowed checks if a file path is within the plugin's allowed paths.
func (m *Manifest) IsPathAllowed(path string) bool {
	if !m.Permissions.FileSystem {
		return false
	}
	if len(m.Permissions.AllowedPaths) == 0 {
		return true
	}
	for _, allowed := range m.Permissions.AllowedPaths {
		rel, err := filepath.Rel(allowed, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return true
		}
	}
	return false
}

// IsCommandAllowed checks if a command is in the plugin's allowed commands.
func (m *Manifest) IsCommandAllowed(cmd string) bool {
	if len(m.Permissions.Commands) == 0 {
		return true
	}
	for _, allowed := range m.Permissions.Commands {
		if allowed == cmd || allowed == "*" {
			return true
		}
	}
	return false
}

// Timeout returns the plugin's execution timeout with a default fallback.
func (m *Manifest) TimeoutDuration() time.Duration {
	if m.Timeout > 0 {
		return m.Timeout
	}
	return 30 * time.Second
}

// MaxMemory returns the plugin's memory limit in bytes.
func (m *Manifest) MaxMemoryBytes() int64 {
	if m.Permissions.MaxMemoryMB > 0 {
		return int64(m.Permissions.MaxMemoryMB) * 1024 * 1024
	}
	return 256 * 1024 * 1024 // 256MB default
}

// isValidSemver checks if a string is a valid semantic version.
func isValidSemver(s string) bool {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
	}
	return true
}
