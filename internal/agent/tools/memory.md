Manage project-specific memory and persistent facts across sessions.
Use this to store important project architectural choices, API guidelines, frequently used complex bash commands, specific code-style rules or anything else that would be helpful to recall on future runs. This memory is saved to `.ghost/memory.json`.
Usage examples:
- action 'set': Store a new guideline. (e.g. key="testing_command", value="npm run test")
- action 'get': Retrieve a specified memory key.
- action 'delete': Remove an obsolete fact.
- action 'list': Show all stored facts.
