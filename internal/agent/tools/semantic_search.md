Search the codebase using natural language to find relevant code snippets.

This tool uses semantic search (TF-IDF embeddings) to find code that matches your intent, not just keyword matches. It understands concepts like "authentication", "error handling", "database connection", etc.

Use this tool when:
- You want to find code related to a concept, not a specific function name
- grep/glob searches return too many or too few results
- You're exploring an unfamiliar codebase
- You need to understand how a feature is implemented across multiple files

Parameters:
- query (required): Natural language description of what you're looking for
- limit (optional): Max results to return (default 10)

Examples:
- "authentication middleware that validates JWT tokens"
- "error handling for database connections"
- "API rate limiting implementation"
- "how user permissions are checked"
- "database migration setup"

Returns:
- Ranked list of relevant code snippets with file paths, line numbers, and relevance scores
- Each result shows the full content of the matching code chunk

Note: The semantic index is built in the background when Ghost starts. If it's not ready yet, use grep/glob as fallback.
