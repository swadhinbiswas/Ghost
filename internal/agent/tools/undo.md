Undo the last change made to a file, restoring it to its previous version.

This tool reverts a file to the state it was in before the most recent edit, create, or delete operation performed by the agent in the current session. It uses the file history tracking system to find the previous version.

Use this tool when:
- An edit introduced a bug or unwanted change
- You want to explore an alternative approach after making changes
- You accidentally modified the wrong file

Behavior:
- If file_path is provided, undoes changes to that specific file
- If file_path is omitted, undoes the last modified file in the current session
- Only undoes the most recent change (one step back)
- The undo action itself is recorded in history, so you can undo the undo
- Requires the file to have at least 2 versions in history

Parameters:
- file_path (optional): Path to the file to undo. If omitted, the last modified file is used.

Returns:
- Confirmation message with the file path and version numbers
- Error if the file has no previous versions or doesn't exist in history
