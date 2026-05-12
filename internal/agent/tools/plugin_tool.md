Manage plugins that extend Ghost with new tools, hooks, and agents.

**Actions:**
- `list`: Show all installed plugins
- `info`: Get details about a specific plugin (requires `name`)
- `execute`: Run a plugin's entrypoint or script (requires `name`, optional `input`/`script`/`args`)
- `install`: Install a plugin from a directory (requires `source_dir`)
- `uninstall`: Remove a plugin (requires `name`)

**Plugin Types:**
- `tool`: Provides new AI tools
- `hook`: Registers event hooks (pre/post processing)
- `theme`: Custom UI themes
- `script`: Executable scripts
- `agent`: Standalone AI agents

**Examples:**
```json
{"action": "list"}
{"action": "info", "name": "my-plugin"}
{"action": "execute", "name": "my-plugin", "input": "hello world"}
{"action": "execute", "name": "my-plugin", "script": "scripts/analyze.sh", "args": ["--verbose"]}
{"action": "install", "source_dir": "/path/to/plugin"}
```

**Plugin Structure:**
```
.ghost/plugins/<name>/
├── plugin.json    # Manifest (required)
├── main.py        # Entrypoint (or .sh, .js, .go, binary)
└── scripts/       # Additional scripts
```
