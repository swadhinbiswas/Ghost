package diff

import (
	"strings"

	"github.com/aymanbagabas/go-udiff"
)

// GenerateDiff creates a unified diff from two file contents
func GenerateDiff(beforeContent, afterContent, fileName string) (string, int, int) {
	fileName = strings.TrimPrefix(fileName, "/")

	var (
		unified   = udiff.Unified("a/"+fileName, "b/"+fileName, beforeContent, afterContent)
		additions = 0
		removals  = 0
	)

	lines := strings.SplitSeq(unified, "\n")
	for line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			additions++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removals++
		}
	}

	return unified, additions, removals
}

// GenerateDiffPreview returns a truncated unified diff showing only changed
// blocks with the given number of context lines on each side.
// It is designed to be concise enough for LLM tool responses.
func GenerateDiffPreview(beforeContent, afterContent string, contextLines int) string {
	fileName := "file"
	unified := udiff.Unified("a/"+fileName, "b/"+fileName, beforeContent, afterContent)

	lines := strings.Split(unified, "\n")
	if len(lines) <= 0 {
		return "(no changes)"
	}

	// Collect hunks: groups of lines that include +/- changes plus context.
	var preview []string
	inHunk := false
	hunkBuffer := []string{}

	for i, line := range lines {
		isChange := strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")
		isContext := strings.HasPrefix(line, " ")
		isHeader := strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "diff") || line == ""

		if isChange {
			if !inHunk {
				// Look back for context lines
				start := max(0, i-contextLines)
				// Find the nearest header or context boundary
				for j := i - 1; j >= start; j-- {
					if strings.HasPrefix(lines[j], "@@") {
						start = j
						break
					}
				}
				hunkBuffer = append(hunkBuffer, lines[start:i]...)
				inHunk = true
			}
			hunkBuffer = append(hunkBuffer, line)
		} else if isContext && inHunk {
			hunkBuffer = append(hunkBuffer, line)
			// Check if we've gone past context window since last change
			lastChangeIdx := -1
			for j := len(hunkBuffer) - 1; j >= 0; j-- {
				if strings.HasPrefix(hunkBuffer[j], "+") || strings.HasPrefix(hunkBuffer[j], "-") {
					lastChangeIdx = j
					break
				}
			}
			if lastChangeIdx >= 0 && len(hunkBuffer)-1-lastChangeIdx >= contextLines {
				// End hunk
				preview = append(preview, strings.Join(hunkBuffer, "\n"))
				hunkBuffer = nil
				inHunk = false
			}
		} else if isHeader {
			if inHunk && len(hunkBuffer) > 0 {
				preview = append(preview, strings.Join(hunkBuffer, "\n"))
				hunkBuffer = nil
			}
			preview = append(preview, line)
			inHunk = false
		}
	}

	// Flush remaining hunk
	if len(hunkBuffer) > 0 {
		preview = append(preview, strings.Join(hunkBuffer, "\n"))
	}

	result := strings.Join(preview, "\n")
	if result == "" {
		return "(no changes)"
	}
	return result
}
