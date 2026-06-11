package tools

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"charm.land/fantasy"
)

//go:embed screenshot.md
var screenshotDescription []byte

const ScreenshotToolName = "screenshot"

// ScreenshotParams defines the parameters for the screenshot tool.
type ScreenshotParams struct {
	Mode   string `json:"mode,omitempty" description:"Capture mode: 'full' (default), 'window', 'area', or 'file'"`
	Path   string `json:"path,omitempty" description:"Path to an existing image file to read (when mode='file')"`
	Region string `json:"region,omitempty" description:"Region to capture: 'WxH+X+Y' (e.g. '800x600+100+50')"`
	Window string `json:"window,omitempty" description:"Window name/title to capture (when mode='window')"`
}

// NewScreenshotTool creates a tool for capturing and analyzing screenshots.
func NewScreenshotTool(workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ScreenshotToolName,
		string(screenshotDescription),
		func(ctx context.Context, params ScreenshotParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			var imageData []byte
			var mimeType string
			var caption string

			switch params.Mode {
			case "file":
				if params.Path == "" {
					return fantasy.NewTextErrorResponse("path is required when mode='file'"), nil
				}
				path := params.Path
				if !filepath.IsAbs(path) {
					path = filepath.Join(workingDir, path)
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to read image: %s", err)), nil
				}
				imageData = data
				mimeType = detectImageMIMEType(path)
				caption = fmt.Sprintf("Loaded image: %s", filepath.Base(params.Path))

			case "area":
				data, path, err := captureArea(params.Region)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to capture area: %s", err)), nil
				}
				imageData = data
				mimeType = "image/png"
				caption = fmt.Sprintf("Captured area, saved to %s", path)

			case "window":
				if params.Window == "" {
					return fantasy.NewTextErrorResponse("window is required when mode='window'"), nil
				}
				data, path, err := captureWindow(params.Window)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to capture window: %s", err)), nil
				}
				imageData = data
				mimeType = "image/png"
				caption = fmt.Sprintf("Captured window '%s', saved to %s", params.Window, path)

			case "full", "":
				data, path, err := captureFull()
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to capture screen: %s", err)), nil
				}
				imageData = data
				mimeType = "image/png"
				caption = fmt.Sprintf("Full screen capture, saved to %s", path)

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown mode: %s. Use 'full', 'window', 'area', or 'file'", params.Mode)), nil
			}

			// Encode as base64 for the response
			encoded := base64.StdEncoding.EncodeToString(imageData)

			return fantasy.ToolResponse{
				Content:   caption,
				Data:      imageData,
				MediaType: mimeType,
				Metadata:  fmt.Sprintf("base64:%s", encoded[:40]+"..."),
			}, nil
		},
	)
}

// captureFull captures the entire screen.
func captureFull() ([]byte, string, error) {
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("ghost-screenshot-%d.png", time.Now().UnixNano()))

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("screencapture", "-x", tmpPath)
		if err := cmd.Run(); err != nil {
			return nil, "", fmt.Errorf("screencapture failed: %w", err)
		}
	case "linux":
		// Try grim (Wayland), then scrot (X11), then import (ImageMagick)
		if err := tryCapture("grim", tmpPath); err == nil {
			break
		}
		if err := tryCapture("scrot", "-s", tmpPath); err == nil {
			break
		}
		if err := tryCapture("import", "-window", "root", tmpPath); err == nil {
			break
		}
		return nil, "", fmt.Errorf("no screenshot tool found. Install grim, scrot, or ImageMagick")
	case "windows":
		return nil, "", fmt.Errorf("windows screenshot requires PowerShell; use 'file' mode with an existing screenshot")
	default:
		return nil, "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	data, err := os.ReadFile(tmpPath)
	return data, tmpPath, err
}

// captureArea captures a specific region of the screen.
func captureArea(region string) ([]byte, string, error) {
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("ghost-screenshot-area-%d.png", time.Now().UnixNano()))

	switch runtime.GOOS {
	case "darwin":
		if region == "" {
			return nil, "", fmt.Errorf("region is required on macOS (format: WxH+X+Y)")
		}
		parts := parseRegion(region)
		cmd := exec.Command("screencapture", "-x", "-R",
			fmt.Sprintf("%d,%d,%d,%d", parts.x, parts.y, parts.w, parts.h), tmpPath)
		if err := cmd.Run(); err != nil {
			return nil, "", fmt.Errorf("screencapture failed: %w", err)
		}
	case "linux":
		if region == "" {
			return nil, "", fmt.Errorf("region is required (format: WxH+X+Y)")
		}
		parts := parseRegion(region)
		if err := tryCapture("import", "-window", "root", "-geometry",
			fmt.Sprintf("%dx%d+%d+%d", parts.w, parts.h, parts.x, parts.y), tmpPath); err != nil {
			return nil, "", fmt.Errorf("import failed: %w", err)
		}
	default:
		return nil, "", fmt.Errorf("area capture not supported on %s", runtime.GOOS)
	}

	data, err := os.ReadFile(tmpPath)
	return data, tmpPath, err
}

// captureWindow captures a specific window by title.
func captureWindow(title string) ([]byte, string, error) {
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("ghost-screenshot-win-%d.png", time.Now().UnixNano()))

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("screencapture", "-x", "-l", title, tmpPath)
		if err := cmd.Run(); err != nil {
			return nil, "", fmt.Errorf("screencapture failed: %w", err)
		}
	case "linux":
		cmd := exec.Command("xdotool", "search", "--name", title)
		output, err := cmd.Output()
		if err != nil {
			return nil, "", fmt.Errorf("window '%s' not found: %w", title, err)
		}
		windowID := strings.TrimSpace(string(output))
		if err := tryCapture("import", "-window", windowID, tmpPath); err != nil {
			return nil, "", fmt.Errorf("import failed: %w", err)
		}
	default:
		return nil, "", fmt.Errorf("window capture not supported on %s", runtime.GOOS)
	}

	data, err := os.ReadFile(tmpPath)
	return data, tmpPath, err
}

type region struct {
	w, h, x, y int
}

// parseRegion parses a region string like "800x600+100+50".
func parseRegion(s string) region {
	var r region
	_, _ = fmt.Sscanf(s, "%dx%d+%d+%d", &r.w, &r.h, &r.x, &r.y)
	return r
}

// tryCapture attempts to run a capture command.
func tryCapture(cmd string, args ...string) error {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}
	c := exec.Command(cmd, args...)
	return c.Run()
}

// detectImageMIMEType detects the MIME type of an image file.
func detectImageMIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "image/png"
	}
}

// AnalyzeScreenshot is a helper to format a screenshot analysis prompt.
func AnalyzeScreenshot(description string) string {
	var sb strings.Builder
	sb.WriteString("Analyze the provided screenshot and provide the following:\n\n")
	sb.WriteString("1. **UI Elements**: Describe the main UI components visible\n")
	sb.WriteString("2. **Content**: Extract any visible text or data\n")
	sb.WriteString("3. **Errors/Issues**: Identify any error messages, warnings, or visual issues\n")
	sb.WriteString("4. **Suggestions**: Recommend actions based on what you see\n\n")
	if description != "" {
		fmt.Fprintf(&sb, "Context: %s\n\n", description)
	}
	sb.WriteString("Be specific and actionable in your analysis.")
	return sb.String()
}

// ScreenshotAsBase64 converts image data to a base64 data URI.
func ScreenshotAsBase64(data []byte, mimeType string) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// CleanupScreenshots removes temporary screenshot files older than the given duration.
func CleanupScreenshots(tempDir string, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "ghost-screenshot-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(tempDir, entry.Name()))
		}
	}

	return nil
}
