package tools

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/filepathext"
	"github.com/swadhinbiswas/ghost/internal/fsext"
	"github.com/swadhinbiswas/ghost/internal/history"
	"github.com/swadhinbiswas/ghost/internal/permission"
)

type UndoParams struct {
	FilePath string `json:"file_path" description:"The path to the file to undo (optional, if omitted undoes the last modified file)"`
}

type UndoPermissionsParams struct {
	FilePath string `json:"file_path"`
}

type UndoResponseMetadata struct {
	FilePath        string `json:"file_path"`
	PreviousVersion int64  `json:"previous_version"`
	RestoredVersion int64  `json:"restored_version"`
}

const (
	UndoToolName = "undo"
)

//go:embed undo.md
var undoDescription []byte

func NewUndoTool(permissions permission.Service, files history.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		UndoToolName,
		string(undoDescription),
		func(ctx context.Context, params UndoParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for undo")
			}

			// If no file path given, find the most recently modified file in this session
			filePath := params.FilePath
			if filePath == "" {
				sessFiles, err := files.ListLatestSessionFiles(ctx, sessionID)
				if err != nil || len(sessFiles) == 0 {
					return fantasy.NewTextErrorResponse("no files tracked in this session to undo"), nil
				}
				filePath = sessFiles[0].Path
			}

			filePath = filepathext.SmartJoin(workingDir, filePath)

			// Get all versions of the file
			versions, err := files.ListVersions(ctx, filePath)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to list file versions: %w", err)
			}
			if len(versions) < 2 {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("file %s has no previous versions to undo to", filePath)), nil
			}

			// Versions are ordered DESC (newest first), so versions[0] is current, versions[1] is previous
			current := versions[0]
			previous := versions[1]

			absWorkingDir, _ := os.Getwd()
			relPath, _ := filepath.Rel(absWorkingDir, filePath)

			p, err := permissions.Request(ctx,
				permission.CreatePermissionRequest{
					SessionID:   sessionID,
					Path:        fsext.PathOrPrefix(filePath, workingDir),
					ToolCallID:  call.ID,
					ToolName:    UndoToolName,
					Action:      "write",
					Description: fmt.Sprintf("Undo last change to %s (restore version %d)", relPath, previous.Version),
					Params: UndoPermissionsParams{
						FilePath: filePath,
					},
				},
			)
			if err != nil {
				return fantasy.ToolResponse{}, err
			}
			if !p {
				return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
			}

			// Write previous content back to disk
			if err := os.WriteFile(filePath, []byte(previous.Content), 0o644); err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to restore file: %w", err)
			}

			// Create a new version entry for the undo
			_, err = files.CreateVersion(ctx, sessionID, filePath, previous.Content)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to record undo in history: %w", err)
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(fmt.Sprintf("Undid last change to %s (restored to version %d)", relPath, previous.Version)),
				UndoResponseMetadata{
					FilePath:        filePath,
					PreviousVersion: current.Version,
					RestoredVersion: previous.Version,
				},
			), nil
		},
	)
}
