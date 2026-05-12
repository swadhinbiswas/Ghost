Discover and search for code symbols (functions, classes, types, etc.) in the codebase.
Use this to find function definitions, class declarations, struct definitions, interfaces, and more.

**Supported Languages:** Go, Python, JavaScript, TypeScript, Rust, Java, C, C++

**Parameters:**
- `path`: File or directory to search (defaults to working directory)
- `language`: Filter by language (go|python|javascript|typescript|rust|java|c|cpp)
- `symbol`: Symbol name to search for (partial match)
- `type`: Symbol type (function|class|method|variable|interface|struct|type)

**Examples:**
```json
{"symbol": "getUser", "type": "function"}
{"language": "go", "type": "struct"}
{"path": "src/components", "type": "class"}
{"language": "typescript", "symbol": "Handler"}
```

**When to use:**
- Finding where a function or class is defined
- Exploring codebase structure
- Looking for all implementations of an interface
- Discovering available types in a package/module
