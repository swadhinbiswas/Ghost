package tools

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/diff"
	"github.com/swadhinbiswas/ghost/internal/filepathext"
	"github.com/swadhinbiswas/ghost/internal/fsext"
	"github.com/swadhinbiswas/ghost/internal/history"
	"github.com/swadhinbiswas/ghost/internal/permission"
)

// editPlan holds the computed state for a single file edit.
type editPlan struct {
	absPath     string
	relPath     string
	oldContent  string
	newContent  string
	isCreate    bool
	isCrlf      bool
	diffPreview string
	additions   int
	removals    int
}

type AtomicEditFile struct {
	FilePath   string `json:"file_path" description:"Path to the file (relative to working directory)"`
	OldString  string `json:"old_string" description:"Text to find (exact match including whitespace). Empty means create new file."`
	NewString  string `json:"new_string" description:"Text to replace with"`
	ReplaceAll bool   `json:"replace_all,omitempty" description:"Replace all occurrences (default false)"`
}

type AtomicEditParams struct {
	Edits []AtomicEditFile `json:"edits" description:"List of file edits to apply atomically. All succeed or none are applied."`
}

type AtomicEditPermissionsParams struct {
	FilesEdited int    `json:"files_edited"`
	Summary     string `json:"summary"`
}

type AtomicEditResponseMetadata struct {
	FilesEdited  int      `json:"files_edited"`
	Additions    int      `json:"additions"`
	Removals     int      `json:"removals"`
	DiffPreviews []string `json:"diff_previews"`
}

const AtomicEditToolName = "atomic_edit"

//go:embed atomic_edit.md
var atomicEditDescription []byte

func NewAtomicEditTool(
	permissions permission.Service,
	files history.Service,
	workingDir string,
) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		AtomicEditToolName,
		string(atomicEditDescription),
		func(ctx context.Context, params AtomicEditParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if len(params.Edits) == 0 {
				return fantasy.NewTextErrorResponse("at least one edit is required"), nil
			}

			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required")
			}

			// Phase 1: Plan all edits - read files, compute changes, validate
			plans := make([]editPlan, 0, len(params.Edits))
			backups := make(map[string]string) // absPath -> original content for rollback

			for _, edit := range params.Edits {
				absPath := filepathext.SmartJoin(workingDir, edit.FilePath)
				relPath := strings.TrimPrefix(absPath, workingDir)
				relPath = strings.TrimPrefix(relPath, "/")

				if edit.OldString == "" {
					// Create new file
					if _, err := os.Stat(absPath); err == nil {
						return fantasy.NewTextErrorResponse(fmt.Sprintf("file already exists: %s (use edit tool instead)", edit.FilePath)), nil
					}
					plans = append(plans, editPlan{
						absPath:    absPath,
						relPath:    relPath,
						newContent: edit.NewString,
						isCreate:   true,
					})
					continue
				}

				// Read existing file
				content, err := os.ReadFile(absPath)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("cannot read %s: %v", edit.FilePath, err)), nil
				}

				oldContent, isCrlf := fsext.ToUnixLineEndings(string(content))

				// Verify oldString exists
				if !strings.Contains(oldContent, edit.OldString) {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("old_string not found in %s", edit.FilePath)), nil
				}

				// Check for multiple matches
				if !edit.ReplaceAll && strings.Count(oldContent, edit.OldString) > 1 {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("old_string appears multiple times in %s. Use replace_all=true or provide more context", edit.FilePath)), nil
				}

				// Compute new content
				newContent := strings.ReplaceAll(oldContent, edit.OldString, edit.NewString)
				if !edit.ReplaceAll {
					newContent = strings.Replace(oldContent, edit.OldString, edit.NewString, 1)
				}

				if oldContent == newContent {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("no changes would be made to %s (new content is identical)", edit.FilePath)), nil
				}

				// Generate diff preview
				diffPreview := diff.GenerateDiffPreview(oldContent, newContent, 5)
				_, additions, removals := diff.GenerateDiff(oldContent, newContent, relPath)

				plans = append(plans, editPlan{
					absPath:     absPath,
					relPath:     relPath,
					oldContent:  oldContent,
					newContent:  newContent,
					isCrlf:      isCrlf,
					diffPreview: diffPreview,
					additions:   additions,
					removals:    removals,
				})
				backups[absPath] = oldContent
			}

			// Phase 2: Single permission request for all edits
			summary := buildEditSummary(plans)
			p, err := permissions.Request(ctx,
				permission.CreatePermissionRequest{
					SessionID:   sessionID,
					Path:        fsext.PathOrPrefix(workingDir, workingDir),
					ToolCallID:  call.ID,
					ToolName:    AtomicEditToolName,
					Action:      "write",
					Description: fmt.Sprintf("Atomically edit %d file(s): %s", len(plans), summary),
					Params: AtomicEditPermissionsParams{
						FilesEdited: len(plans),
						Summary:     summary,
					},
				},
			)
			if err != nil {
				return fantasy.ToolResponse{}, err
			}
			if !p {
				return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
			}

			// Phase 3: Apply all edits with rollback on any failure
			appliedPaths := []string{}
			totalAdditions := 0
			totalRemovals := 0
			diffPreviews := []string{}

			for _, plan := range plans {
				content := plan.newContent
				if plan.isCrlf {
					content, _ = fsext.ToWindowsLineEndings(content)
				}

				// Ensure parent directory exists
				if plan.isCreate {
					dir := filepath.Dir(plan.absPath)
					if err := os.MkdirAll(dir, 0o755); err != nil {
						rollback(appliedPaths, backups)
						return fantasy.ToolResponse{}, fmt.Errorf("failed to create directory for %s: %w", plan.relPath, err)
					}
				}

				// Write file
				if err := os.WriteFile(plan.absPath, []byte(content), 0o644); err != nil {
					rollback(appliedPaths, backups)
					return fantasy.ToolResponse{}, fmt.Errorf("failed to write %s: %w; all changes have been rolled back", plan.relPath, err)
				}

				appliedPaths = append(appliedPaths, plan.absPath)
				totalAdditions += plan.additions
				totalRemovals += plan.removals
				diffPreviews = append(diffPreviews, fmt.Sprintf("--- %s ---\n%s", plan.relPath, plan.diffPreview))

				// Update history
				if plan.isCreate {
					_, _ = files.Create(ctx, sessionID, plan.absPath, "")
				}
				_, _ = files.CreateVersion(ctx, sessionID, plan.absPath, plan.newContent)
			}

			// Phase 4: Build response
			var sb strings.Builder
			fmt.Fprintf(&sb, "Successfully edited %d file(s) atomically\n\n", len(plans))
			fmt.Fprintf(&sb, "Summary: +%d added, -%d removed\n\n", totalAdditions, totalRemovals)

			for _, dp := range diffPreviews {
				sb.WriteString(dp)
				sb.WriteString("\n\n")
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(sb.String()),
				AtomicEditResponseMetadata{
					FilesEdited:  len(plans),
					Additions:    totalAdditions,
					Removals:     totalRemovals,
					DiffPreviews: diffPreviews,
				},
			), nil
		},
	)
}

func buildEditSummary(plans []editPlan) string {
	creates := 0
	edits := 0
	for _, p := range plans {
		if p.isCreate {
			creates++
		} else {
			edits++
		}
	}
	parts := []string{}
	if creates > 0 {
		parts = append(parts, fmt.Sprintf("%d new file(s)", creates))
	}
	if edits > 0 {
		parts = append(parts, fmt.Sprintf("%d existing file(s)", edits))
	}
	return strings.Join(parts, ", ")
}

// rollback restores all applied files to their original content.
func rollback(appliedPaths []string, backups map[string]string) {
	for _, path := range appliedPaths {
		if content, ok := backups[path]; ok {
			_ = os.WriteFile(path, []byte(content), 0o644)
		} else {
			// Was a new file - delete it
			os.Remove(path)
		}
	}
}
