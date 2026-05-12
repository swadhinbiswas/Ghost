Run tests for the current project and return structured results.

This tool automatically detects the test framework being used and runs the appropriate test command. It parses the output to return structured pass/fail/skip counts.

Supported frameworks:
- Go: go test (go.mod detected)
- Node.js: jest, vitest, mocha, npm test (package.json detected)
- Python: pytest (pytest.ini, pyproject.toml, setup.py detected)
- Rust: cargo test (Cargo.toml detected)
- Java: mvn test (pom.xml detected), gradle test (build.gradle detected)

Parameters:
- path (optional): Path to the file or directory to test. Defaults to the working directory.
- pattern (optional): Regex pattern to filter test names by.

Returns:
- Structured test results with pass/fail/skip counts
- Full test output for failed tests
- Duration of the test run
- The command that was executed

Notes:
- Tests run with a 5-minute timeout
- For Go tests, uses -json flag for structured output
- For Jest/Vitest, uses --json flag
- For pytest, uses --tb=short for concise failure output
- Output is truncated to 4000 characters if too long
