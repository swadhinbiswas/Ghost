package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/plugin"
)

//go:embed plugin_tool.md
var pluginToolDescription []byte

const PluginToolName = "plugins"

// PluginParams defines the parameters for the plugin tool.
type PluginParams struct {
	Action    string            `json:"action" description:"Action: 'list', 'info', 'execute', 'install', 'uninstall', 'enable', 'disable'"`
	Name      string            `json:"name,omitempty" description:"Plugin name (required for info/execute/uninstall/enable/disable)"`
	Input     string            `json:"input,omitempty" description:"Input data to pass to the plugin"`
	Script    string            `json:"script,omitempty" description:"Script to run within the plugin"`
	Args      []string          `json:"args,omitempty" description:"Arguments for the script"`
	SourceDir string            `json:"source_dir,omitempty" description:"Source directory for install action"`
	Config    map[string]string `json:"config,omitempty" description:"Configuration overrides"`
}

// NewPluginTool creates a tool for managing plugins.
func NewPluginTool(loader *plugin.Loader) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		PluginToolName,
		string(pluginToolDescription),
		func(ctx context.Context, params PluginParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Action == "" {
				return fantasy.NewTextErrorResponse("action is required"), nil
			}

			switch params.Action {
			case "list":
				return handleList(loader)
			case "info":
				if params.Name == "" {
					return fantasy.NewTextErrorResponse("name is required for info action"), nil
				}
				return handleInfo(loader, params.Name)
			case "execute":
				if params.Name == "" {
					return fantasy.NewTextErrorResponse("name is required for execute action"), nil
				}
				return handleExecute(ctx, loader, params)
			case "install":
				if params.SourceDir == "" {
					return fantasy.NewTextErrorResponse("source_dir is required for install action"), nil
				}
				return handleInstall(loader, params.SourceDir)
			case "uninstall":
				if params.Name == "" {
					return fantasy.NewTextErrorResponse("name is required for uninstall action"), nil
				}
				return handleUninstall(loader, params.Name)
			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown action: %s", params.Action)), nil
			}
		},
	)
}

func handleList(loader *plugin.Loader) (fantasy.ToolResponse, error) {
	plugins := loader.List()
	if len(plugins) == 0 {
		return fantasy.NewTextResponse("No plugins installed.\n\nTo install a plugin, create a `.ghost/plugins/<name>/plugin.json` manifest file."), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Installed Plugins (%d):\n\n", len(plugins))

	for _, p := range plugins {
		m := p.Manifest
		status := "loaded"
		if !m.Loaded {
			status = "failed"
		}
		fmt.Fprintf(&sb, "- **%s** v%s (%s) - %s [%s]\n", m.Name, m.Version, m.Type, m.Description, status)
		if m.LoadError != "" {
			fmt.Fprintf(&sb, "  Error: %s\n", m.LoadError)
		}
	}

	return fantasy.NewTextResponse(sb.String()), nil
}

func handleInfo(loader *plugin.Loader, name string) (fantasy.ToolResponse, error) {
	p, ok := loader.Get(name)
	if !ok {
		// Try to find in filesystem
		manifests, err := loader.Discover()
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to discover plugins: %s", err)), nil
		}
		for _, m := range manifests {
			if m.Name == name {
				return formatPluginInfo(m), nil
			}
		}
		return fantasy.NewTextErrorResponse(fmt.Sprintf("plugin not found: %s", name)), nil
	}

	return formatPluginInfo(p.Manifest), nil
}

func formatPluginInfo(m *plugin.Manifest) fantasy.ToolResponse {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## %s v%s\n\n", m.Name, m.Version)
	fmt.Fprintf(&sb, "**Type:** %s\n", m.Type)
	fmt.Fprintf(&sb, "**Description:** %s\n", m.Description)
	fmt.Fprintf(&sb, "**Author:** %s\n", m.Author)
	fmt.Fprintf(&sb, "**License:** %s\n", m.License)
	fmt.Fprintf(&sb, "**Homepage:** %s\n", m.Homepage)
	fmt.Fprintf(&sb, "**Status:** %s\n", func() string {
		if m.LoadError != "" {
			return "error: " + m.LoadError
		}
		if m.Loaded {
			return "loaded"
		}
		return "discovered"
	}())

	if len(m.Permissions.Commands) > 0 {
		fmt.Fprintf(&sb, "\n**Allowed Commands:** %s\n", strings.Join(m.Permissions.Commands, ", "))
	}
	if len(m.Permissions.AllowedPaths) > 0 {
		fmt.Fprintf(&sb, "**Allowed Paths:** %s\n", strings.Join(m.Permissions.AllowedPaths, ", "))
	}

	return fantasy.NewTextResponse(sb.String())
}

func handleExecute(ctx context.Context, loader *plugin.Loader, params PluginParams) (fantasy.ToolResponse, error) {
	p, ok := loader.Get(params.Name)
	if !ok {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("plugin not found or not loaded: %s", params.Name)), nil
	}

	var output string
	var err error

	if params.Script != "" {
		output, err = p.Runtime.ExecuteScript(ctx, params.Script, params.Args...)
	} else {
		output, err = p.Runtime.Execute(ctx, params.Input)
	}

	if err != nil {
		return fantasy.ToolResponse{
			Content: fmt.Sprintf("Plugin execution failed: %s\n\nOutput:\n%s", err, output),
			IsError: true,
		}, nil
	}

	return fantasy.NewTextResponse(output), nil
}

func handleInstall(loader *plugin.Loader, sourceDir string) (fantasy.ToolResponse, error) {
	manifest, err := loader.Install(sourceDir)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to install plugin: %s", err)), nil
	}

	// Load the installed plugin
	_, err = loader.LoadPlugin(manifest)
	if err != nil {
		return fantasy.NewTextResponse(fmt.Sprintf("Plugin installed but failed to load: %s\n\nManifest:\n%s", err, formatManifestJSON(manifest))), nil
	}

	return fantasy.NewTextResponse(fmt.Sprintf("Plugin '%s' v%s installed and loaded successfully.", manifest.Name, manifest.Version)), nil
}

func handleUninstall(loader *plugin.Loader, name string) (fantasy.ToolResponse, error) {
	err := loader.Uninstall(name)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to uninstall plugin: %s", err)), nil
	}

	return fantasy.NewTextResponse(fmt.Sprintf("Plugin '%s' uninstalled successfully.", name)), nil
}

func formatManifestJSON(m *plugin.Manifest) string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}
