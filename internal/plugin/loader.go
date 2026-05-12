package plugin

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	pluginsDirName = ".ghost/plugins"
)

// Loader discovers, validates, and manages plugins.
type Loader struct {
	pluginsDir string
	plugins    map[string]*Plugin
	mu         sync.RWMutex
}

// Plugin holds a loaded plugin with its runtime state.
type Plugin struct {
	Manifest *Manifest
	Runtime  *Runtime
}

// NewLoader creates a new plugin loader.
func NewLoader(workingDir string) *Loader {
	return &Loader{
		pluginsDir: filepath.Join(workingDir, pluginsDirName),
		plugins:    make(map[string]*Plugin),
	}
}

// NewLoaderWithDir creates a new plugin loader with a custom plugins directory.
func NewLoaderWithDir(pluginsDir string) *Loader {
	return &Loader{
		pluginsDir: pluginsDir,
		plugins:    make(map[string]*Plugin),
	}
}

// Discover scans the plugins directory and returns found plugins.
func (l *Loader) Discover() ([]*Manifest, error) {
	entries, err := os.ReadDir(l.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	var manifests []*Manifest

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		pluginDir := filepath.Join(l.pluginsDir, entry.Name())
		manifest, err := LoadManifest(pluginDir)
		if err != nil {
			slog.Warn("Failed to load plugin manifest", "plugin", entry.Name(), "error", err)
			continue
		}

		manifest.Enabled = true
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

// LoadAll discovers and loads all plugins.
func (l *Loader) LoadAll() ([]*Plugin, error) {
	manifests, err := l.Discover()
	if err != nil {
		return nil, err
	}

	var loaded []*Plugin

	for _, manifest := range manifests {
		plugin, err := l.LoadPlugin(manifest)
		if err != nil {
			slog.Warn("Failed to load plugin", "plugin", manifest.Name, "error", err)
			continue
		}
		loaded = append(loaded, plugin)
	}

	return loaded, nil
}

// LoadPlugin loads a single plugin from its manifest.
func (l *Loader) LoadPlugin(manifest *Manifest) (*Plugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if already loaded
	if existing, ok := l.plugins[manifest.Name]; ok {
		return existing, nil
	}

	// Create runtime
	runtime, err := NewRuntime(manifest)
	if err != nil {
		manifest.LoadError = err.Error()
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	plugin := &Plugin{
		Manifest: manifest,
		Runtime:  runtime,
	}

	plugin.Manifest.Loaded = true
	l.plugins[manifest.Name] = plugin

	slog.Info("Plugin loaded", "name", manifest.Name, "type", manifest.Type, "version", manifest.Version)

	return plugin, nil
}

// Get returns a loaded plugin by name.
func (l *Loader) Get(name string) (*Plugin, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	p, ok := l.plugins[name]
	return p, ok
}

// List returns all loaded plugins.
func (l *Loader) List() []*Plugin {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]*Plugin, 0, len(l.plugins))
	for _, p := range l.plugins {
		result = append(result, p)
	}
	return result
}

// GetTools returns all tool-type plugins.
func (l *Loader) GetTools() []*Plugin {
	return l.FilterByType(PluginTypeTool)
}

// GetHooks returns all hook-type plugins.
func (l *Loader) GetHooks() []*Plugin {
	return l.FilterByType(PluginTypeHook)
}

// GetAgents returns all agent-type plugins.
func (l *Loader) GetAgents() []*Plugin {
	return l.FilterByType(PluginTypeAgent)
}

// FilterByType returns plugins of a specific type.
func (l *Loader) FilterByType(pluginType PluginType) []*Plugin {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []*Plugin
	for _, p := range l.plugins {
		if p.Manifest.Type == pluginType {
			result = append(result, p)
		}
	}
	return result
}

// Unload unloads a plugin by name.
func (l *Loader) Unload(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	plugin, ok := l.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if plugin.Runtime != nil {
		plugin.Runtime.Stop()
	}

	delete(l.plugins, name)
	slog.Info("Plugin unloaded", "name", name)
	return nil
}

// UnloadAll unloads all plugins.
func (l *Loader) UnloadAll() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for name, plugin := range l.plugins {
		if plugin.Runtime != nil {
			plugin.Runtime.Stop()
		}
		delete(l.plugins, name)
	}
}

// Install installs a plugin from a directory.
func (l *Loader) Install(sourceDir string) (*Manifest, error) {
	// Load manifest from source
	manifest, err := LoadManifest(sourceDir)
	if err != nil {
		return nil, err
	}

	// Create target directory
	targetDir := filepath.Join(l.pluginsDir, manifest.Name)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Copy plugin files
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, 0644)
	})

	if err != nil {
		os.RemoveAll(targetDir)
		return nil, fmt.Errorf("failed to copy plugin files: %w", err)
	}

	manifest.InstallPath = targetDir
	return manifest, nil
}

// Uninstall removes a plugin by name.
func (l *Loader) Uninstall(name string) error {
	l.Unload(name)

	pluginDir := filepath.Join(l.pluginsDir, name)
	return os.RemoveAll(pluginDir)
}

// CreateDefaultPluginsDir creates the default plugins directory if it doesn't exist.
func CreateDefaultPluginsDir(workingDir string) error {
	dir := filepath.Join(workingDir, pluginsDirName)
	return os.MkdirAll(dir, 0755)
}
