Manage terminal panes in the multiplexer view.
Use this to create side-by-side panels for viewing files, monitoring output, running tests, etc.

**Actions:**
- `create`: Create a new pane (requires `title`, optional `content`)
- `update`: Replace pane content (requires `pane_id`, `content`)
- `append`: Append to pane content (requires `pane_id`, `content`)
- `close`: Close a specific pane (optional `pane_id`) or the active pane
- `list`: List all open panes
- `focus`: Focus a specific pane (requires `pane_id`)
- `layout`: Change layout (requires `layout`: single|horizontal|vertical|grid)

**When to use:**
- Viewing a file while editing another
- Running tests in one pane while editing code in another
- Comparing multiple files side-by-side
- Monitoring logs or output while working

**Example:**
```json
{"action": "create", "title": "Test Output", "content": "Running tests..."}
{"action": "update", "pane_id": "pane-1", "content": "New content here"}
{"action": "layout", "layout": "horizontal"}
```
