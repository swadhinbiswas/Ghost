package tools

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/permission"
)

type TestParams struct {
	Path    string `json:"path,omitempty" description:"Path to the file or directory to test (defaults to working directory)"`
	Pattern string `json:"pattern,omitempty" description:"Regex pattern to filter test names"`
}

type TestPermissionsParams struct {
	Path    string `json:"path"`
	Pattern string `json:"pattern"`
}

type TestResult struct {
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
	Skipped int    `json:"skipped"`
	Errors  string `json:"errors,omitempty"`
}

type TestResponseMetadata struct {
	Framework string     `json:"framework"`
	Command   string     `json:"command"`
	Duration  string     `json:"duration"`
	Result    TestResult `json:"result"`
}

const (
	TestToolName = "test"
	testTimeout  = 5 * time.Minute
)

//go:embed test.md
var testDescription []byte

func NewTestTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		TestToolName,
		string(testDescription),
		func(ctx context.Context, params TestParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for test")
			}

			testPath := params.Path
			if testPath == "" {
				testPath = workingDir
			}

			framework := detectFramework(workingDir)
			if framework == "" {
				return fantasy.NewTextErrorResponse("could not detect a test framework. Supported: Go (go test), Node.js (jest/vitest), Python (pytest), Rust (cargo test), Java (mvn test/gradle test)"), nil
			}

			cmd := buildTestCommand(framework, testPath, params.Pattern)

			// Request permission
			p, err := permissions.Request(ctx,
				permission.CreatePermissionRequest{
					SessionID:   sessionID,
					Path:        workingDir,
					ToolCallID:  call.ID,
					ToolName:    TestToolName,
					Action:      "execute",
					Description: fmt.Sprintf("Run tests using %s", framework),
					Params: TestPermissionsParams{
						Path:    testPath,
						Pattern: params.Pattern,
					},
				},
			)
			if err != nil {
				return fantasy.ToolResponse{}, err
			}
			if !p {
				return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
			}

			start := time.Now()
			ctx, cancel := context.WithTimeout(ctx, testTimeout)
			defer cancel()

			c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
			c.Dir = workingDir
			c.Env = os.Environ()

			output, err := c.CombinedOutput()
			duration := time.Since(start)

			result := parseTestOutput(framework, string(output))

			meta := TestResponseMetadata{
				Framework: framework,
				Command:   strings.Join(cmd, " "),
				Duration:  duration.Round(time.Millisecond).String(),
				Result:    result,
			}

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return fantasy.WithResponseMetadata(
						fantasy.NewTextResponse(fmt.Sprintf("Test timed out after %s\n\n%s", testTimeout, truncateTestOutput(string(output)))),
						meta,
					), nil
				}
				return fantasy.WithResponseMetadata(
					fantasy.NewTextResponse(fmt.Sprintf("Tests failed:\n\n%s", formatTestResult(result, string(output)))),
					meta,
				), nil
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(fmt.Sprintf("All tests passed!\n%s", formatTestResult(result, ""))),
				meta,
			), nil
		},
	)
}

func detectFramework(dir string) string {
	// Go
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return "go"
	}
	// Rust
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		return "rust"
	}
	// Node.js
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
		content := string(data)
		if strings.Contains(content, `"jest"`) || strings.Contains(content, `"jest":`) {
			return "jest"
		}
		if strings.Contains(content, `"vitest"`) || strings.Contains(content, `"vitest":`) {
			return "vitest"
		}
		if strings.Contains(content, `"mocha"`) || strings.Contains(content, `"mocha":`) {
			return "mocha"
		}
		return "node"
	}
	// Python
	if _, err := os.Stat(filepath.Join(dir, "pytest.ini")); err == nil {
		return "pytest"
	}
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		return "pytest"
	}
	if _, err := os.Stat(filepath.Join(dir, "setup.py")); err == nil {
		return "pytest"
	}
	// Java/Maven
	if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
		return "maven"
	}
	// Java/Gradle
	if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
		return "gradle"
	}
	if _, err := os.Stat(filepath.Join(dir, "build.gradle.kts")); err == nil {
		return "gradle"
	}
	return ""
}

func buildTestCommand(framework, path, pattern string) []string {
	switch framework {
	case "go":
		cmd := []string{"go", "test", "-v", "-json", "./..."}
		if path != "" && path != "." {
			cmd = []string{"go", "test", "-v", "-json", path}
		}
		if pattern != "" {
			cmd = append(cmd, "-run", pattern)
		}
		return cmd
	case "jest":
		cmd := []string{"npx", "jest", "--json", "--verbose"}
		if path != "" && path != "." {
			cmd = append(cmd, path)
		}
		if pattern != "" {
			cmd = append(cmd, "-t", pattern)
		}
		return cmd
	case "vitest":
		cmd := []string{"npx", "vitest", "run", "--reporter=verbose", "--json"}
		if path != "" && path != "." {
			cmd = append(cmd, path)
		}
		if pattern != "" {
			cmd = append(cmd, "-t", pattern)
		}
		return cmd
	case "mocha":
		cmd := []string{"npx", "mocha", "--reporter", "json"}
		if path != "" && path != "." {
			cmd = append(cmd, path)
		}
		return cmd
	case "node":
		return []string{"npm", "test"}
	case "pytest":
		cmd := []string{"pytest", "-v", "--tb=short"}
		if path != "" && path != "." {
			cmd = append(cmd, path)
		}
		if pattern != "" {
			cmd = append(cmd, "-k", pattern)
		}
		return cmd
	case "rust":
		cmd := []string{"cargo", "test"}
		if path != "" && path != "." {
			cmd = append(cmd, "--manifest-path", filepath.Join(path, "Cargo.toml"))
		}
		if pattern != "" {
			cmd = append(cmd, "--", pattern)
		}
		return cmd
	case "maven":
		cmd := []string{"mvn", "test"}
		if pattern != "" {
			cmd = append(cmd, "-Dtest="+pattern)
		}
		return cmd
	case "gradle":
		cmd := []string{"./gradlew", "test"}
		if pattern != "" {
			cmd = append(cmd, "--tests", pattern)
		}
		return cmd
	default:
		return []string{"npm", "test"}
	}
}

func parseTestOutput(framework, output string) TestResult {
	result := TestResult{}

	switch framework {
	case "go":
		// Parse go test -json output
		passRe := regexp.MustCompile(`"Action":"pass"`)
		failRe := regexp.MustCompile(`"Action":"fail"`)
		skipRe := regexp.MustCompile(`"Action":"skip"`)
		result.Passed = len(passRe.FindAllString(output, -1))
		result.Failed = len(failRe.FindAllString(output, -1))
		result.Skipped = len(skipRe.FindAllString(output, -1))
	case "jest", "vitest":
		passedRe := regexp.MustCompile(`"numPassedTests":(\d+)`)
		failedRe := regexp.MustCompile(`"numFailedTests":(\d+)`)
		if m := passedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Passed)
		}
		if m := failedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Failed)
		}
	case "pytest":
		passedRe := regexp.MustCompile(`(\d+) passed`)
		failedRe := regexp.MustCompile(`(\d+) failed`)
		errorRe := regexp.MustCompile(`(\d+) error`)
		skippedRe := regexp.MustCompile(`(\d+) skipped`)
		if m := passedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Passed)
		}
		if m := failedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Failed)
		}
		if m := errorRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Failed)
		}
		if m := skippedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Skipped)
		}
	case "rust":
		passedRe := regexp.MustCompile(`(\d+) passed`)
		failedRe := regexp.MustCompile(`(\d+) failed`)
		ignoredRe := regexp.MustCompile(`(\d+) ignored`)
		if m := passedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Passed)
		}
		if m := failedRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Failed)
		}
		if m := ignoredRe.FindStringSubmatch(output); len(m) > 1 {
			fmt.Sscanf(m[1], "%d", &result.Skipped)
		}
	default:
		// Generic fallback: count "PASS" and "FAIL" lines
		passRe := regexp.MustCompile(`(?m)^\s*✓|PASS|\d+\s+passing`)
		failRe := regexp.MustCompile(`(?m)^\s*✗|FAIL|\d+\s+failing`)
		result.Passed = len(passRe.FindAllString(output, -1))
		result.Failed = len(failRe.FindAllString(output, -1))
	}

	return result
}

func formatTestResult(result TestResult, rawOutput string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Passed: %d | Failed: %d | Skipped: %d\n", result.Passed, result.Failed, result.Skipped))
	if result.Failed > 0 && rawOutput != "" {
		sb.WriteString("\n--- Test Output ---\n")
		sb.WriteString(truncateTestOutput(rawOutput))
	}
	return sb.String()
}

func truncateTestOutput(output string) string {
	const maxLen = 4000
	if len(output) <= maxLen {
		return output
	}
	return output[:maxLen] + "\n\n... (output truncated)"
}
