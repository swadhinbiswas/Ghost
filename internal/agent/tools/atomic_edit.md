Apply edits to multiple files atomically - all edits succeed or none are applied.

This tool is designed for refactors that touch multiple files. Unlike the edit tool which modifies one file at a time, atomic_edit ensures consistency: if any edit fails, ALL changes are rolled back.

Use this tool when:
- Renaming a function/type across multiple files
- Updating an API signature that affects callers in different files
- Applying the same pattern change across many files
- Any change where partial application would leave the codebase broken

Parameters:
- edits (required): List of file edits, each containing:
  - file_path: Path to the file (relative to working directory)
  - old_string: Text to find (exact match including whitespace). Leave empty to create a new file.
  - new_string: Text to replace with
  - replace_all: Replace all occurrences (default false)

Behavior:
- Phase 1: Read all files, validate all edits, compute diffs
- Phase 2: Single permission request for all changes
- Phase 3: Apply all edits - if ANY fails, ALL are rolled back
- Phase 4: Return combined diff summary

Important:
- All files must exist (unless creating new)
- All old_string values must be found exactly
- File modification time must not change between read and write
- Rollback restores original content on any failure

Example:
{
  "edits": [
    {
      "file_path": "src/auth/middleware.go",
      "old_string": "func Authenticate(r *http.Request) error",
      "new_string": "func Authenticate(r *http.Request, db *Database) error"
    },
    {
      "file_path": "src/handlers/user.go",
      "old_string": "if err := auth.Authenticate(r); err != nil",
      "new_string": "if err := auth.Authenticate(r, db); err != nil"
    },
    {
      "file_path": "src/handlers/admin.go",
      "old_string": "if err := auth.Authenticate(r); err != nil",
      "new_string": "if err := auth.Authenticate(r, db); err != nil"
    }
  ]
}
