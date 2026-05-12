package agent

import (
	"fmt"
	"regexp"
	"strings"

	"charm.land/fantasy"
)

// SelfHealingPrompt is injected into the system prompt to enable self-correction behavior.
const SelfHealingPrompt = `
<self_healing>
When a tool call fails, follow these steps before asking the user for help:

1. **Analyze the error**: Read the error message carefully to understand what went wrong.
2. **Identify the root cause**: Determine if it's a file path issue, syntax error, missing dependency, permission problem, etc.
3. **Attempt automatic fix**:
   - File not found: Use 'glob' to search for the correct path, then retry
   - Syntax error: Review and fix the command syntax, then retry
   - Compilation error: Check imports, fix type errors, then retry
   - Test failure: Read the test output, identify failing tests, fix the code, then retry
   - Command failed: Check arguments, verify prerequisites, then retry
4. **Limit retries**: Try at most 2 automatic fixes. If still failing, explain the issue to the user.

Common error patterns and fixes:
- "no such file or directory" → File path is wrong, use glob to find it
- "command not found" → Missing tool, try alternative approach
- "permission denied" → Need different approach or user intervention
- "syntax error" → Check quotes, brackets, escaping
- "undefined" → Missing import or variable scope issue
- "exit status 1" → Command failed, check stderr for details
</self_healing>
`

// AnalyzeError provides a detailed analysis of a tool execution error.
// This is used to enhance error responses with actionable suggestions.
func AnalyzeError(err error, response fantasy.ToolResponse, toolName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[Tool '%s' failed]\n", toolName))

	errStr := ""
	if err != nil {
		errStr = err.Error()
	} else if response.IsError {
		errStr = response.Content
	}

	if errStr == "" {
		return sb.String()
	}

	// Categorize and provide suggestions
	switch {
	case strings.Contains(errStr, "file not found"), strings.Contains(errStr, "no such file"):
		sb.WriteString("Category: File not found\n")
		sb.WriteString("Auto-fix: Use 'glob' tool to search for the correct file path, then retry with the correct path.\n")

	case strings.Contains(errStr, "permission denied"):
		sb.WriteString("Category: Permission denied\n")
		sb.WriteString("Auto-fix: This typically requires user intervention. Consider asking the user to run with elevated permissions.\n")

	case strings.Contains(errStr, "syntax error"), strings.Contains(errStr, "unexpected"):
		sb.WriteString("Category: Syntax error\n")
		sb.WriteString("Auto-fix: Check for unclosed quotes, brackets, or escaping issues. Review command syntax and retry.\n")

	case strings.Contains(errStr, "command not found"):
		sb.WriteString("Category: Command not found\n")
		sb.WriteString("Auto-fix: The command may not be installed. Try an alternative approach or check PATH.\n")

	case strings.Contains(errStr, "exit status"):
		sb.WriteString("Category: Command failed\n")
		sb.WriteString("Auto-fix: Check the command arguments and stderr output. Verify prerequisites and retry.\n")

	case strings.Contains(errStr, "compilation failed"), strings.Contains(errStr, "build failed"):
		sb.WriteString("Category: Build/compilation failed\n")
		sb.WriteString("Auto-fix: Review compiler errors, check imports and types, fix issues, then retry build.\n")

	case strings.Contains(errStr, "test failed"), strings.Contains(errStr, "FAIL"):
		sb.WriteString("Category: Test failure\n")
		sb.WriteString("Auto-fix: Read test output to identify failing tests, fix the code, then re-run tests.\n")

	case strings.Contains(errStr, "undefined"), strings.Contains(errStr, "cannot find"):
		sb.WriteString("Category: Undefined reference\n")
		sb.WriteString("Auto-fix: Check for missing imports, typos in variable/function names, or scope issues.\n")

	default:
		sb.WriteString(fmt.Sprintf("Error: %s\n", truncateString(errStr, 200)))
		sb.WriteString("Auto-fix: Analyze the error above and attempt a targeted fix based on the error type.\n")
	}

	return sb.String()
}

// EnhanceToolResult wraps a tool result with error analysis if it failed.
// This helps the LLM understand what went wrong and how to fix it.
func EnhanceToolResult(response fantasy.ToolResponse, toolName string) fantasy.ToolResponse {
	if !response.IsError {
		return response
	}

	analysis := AnalyzeError(nil, response, toolName)
	response.Content = analysis + "\nOriginal error:\n" + response.Content
	return response
}

// ExtractErrorType categorizes an error string into a known error type.
func ExtractErrorType(errStr string) string {
	switch {
	case strings.Contains(errStr, "file not found"), strings.Contains(errStr, "no such file"):
		return "file_not_found"
	case strings.Contains(errStr, "permission denied"):
		return "permission_denied"
	case strings.Contains(errStr, "syntax error"):
		return "syntax_error"
	case strings.Contains(errStr, "command not found"):
		return "command_not_found"
	case strings.Contains(errStr, "exit status"):
		return "command_failed"
	case strings.Contains(errStr, "compilation failed"), strings.Contains(errStr, "build failed"):
		return "compilation_failed"
	case strings.Contains(errStr, "test failed"), strings.Contains(errStr, "FAIL"):
		return "test_failed"
	case strings.Contains(errStr, "undefined"), strings.Contains(errStr, "cannot find"):
		return "undefined_reference"
	default:
		return "unknown"
	}
}

// IsRecoverableError determines if an error can potentially be auto-fixed.
func IsRecoverableError(errStr string) bool {
	unrecoverable := []string{
		"permission denied",
		"segmentation fault",
		"out of memory",
		"disk full",
		"network unreachable",
		"connection refused",
	}

	for _, pattern := range unrecoverable {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	return true
}

// ExtractFilePaths extracts file paths from an error message.
func ExtractFilePaths(errStr string) []string {
	// Match common file path patterns
	pathRegex := regexp.MustCompile(`(?:cannot open|file|path)["']?\s*[:"]?([^"'\s,]+(?:\.\w+)?)["']?`)
	matches := pathRegex.FindAllStringSubmatch(errStr, -1)

	var paths []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			path := match[1]
			if !seen[path] && len(path) > 2 {
				seen[path] = true
				paths = append(paths, path)
			}
		}
	}

	return paths
}

// ExtractMissingImports extracts missing import/package names from compilation errors.
func ExtractMissingImports(errStr string) []string {
	// Go: cannot find package "..."
	goRegex := regexp.MustCompile(`cannot find package "([^"]+)"`)
	// TypeScript/JS: Cannot find module '...'
	tsRegex := regexp.MustCompile(`Cannot find module ['"]([^'"]+)['"]`)
	// Python: No module named '...'
	pyRegex := regexp.MustCompile(`No module named ['"]([^'"]+)['"]`)

	var imports []string
	seen := make(map[string]bool)

	for _, regex := range []*regexp.Regexp{goRegex, tsRegex, pyRegex} {
		for _, match := range regex.FindAllStringSubmatch(errStr, -1) {
			if len(match) > 1 {
				imp := match[1]
				if !seen[imp] {
					seen[imp] = true
					imports = append(imports, imp)
				}
			}
		}
	}

	return imports
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
