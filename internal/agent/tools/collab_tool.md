Manage real-time collaboration sessions with other users or AI agents.
Use this to start collaboration rooms, share content, broadcast messages, and monitor participants.

**Actions:**
- `start`: Create a new collaboration room (optional `room` ID)
- `join`: Join an existing room (requires `room`)
- `leave`: Leave a room (requires `room`)
- `list`: List all active collaboration rooms
- `status`: Get room status including participants and content (requires `room`)
- `share`: Share content with all room participants (requires `room`, `content`)
- `message`: Send a chat message to the room (requires `room`, `message`)

**Parameters:**
- `action`: The action to perform
- `room`: Room/session ID
- `name`: Your display name
- `content`: Content to share
- `message`: Chat message text

**Examples:**
```json
{"action": "start", "room": "my-session"}
{"action": "join", "room": "my-session", "name": "Alice"}
{"action": "list"}
{"action": "share", "room": "my-session", "content": "Here is the code..."}
{"action": "message", "room": "my-session", "message": "Ready to review?"}
```
