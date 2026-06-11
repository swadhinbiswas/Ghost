package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInSkillsPath(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "ghost-skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	skillDir1 := filepath.Join(tempDir, "skill1")
	skillDir2 := filepath.Join(tempDir, "skill2")
	outsideDir := filepath.Join(tempDir, "outside")

	if err := os.MkdirAll(skillDir1, 0755); err != nil {
		t.Fatalf("Failed to create skill1 dir: %v", err)
	}
	if err := os.MkdirAll(skillDir2, 0755); err != nil {
		t.Fatalf("Failed to create skill2 dir: %v", err)
	}
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("Failed to create outside dir: %v", err)
	}

	skillsPaths := []string{skillDir1, skillDir2}

	// Test case 1: File inside skillDir1
	file1 := filepath.Join(skillDir1, "SKILL.md")
	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	if !IsInSkillsPath(file1, skillsPaths) {
		t.Errorf("Expected %s to be in skills path", file1)
	}

	// Test case 2: File inside subfolder of skillDir1
	subfolder := filepath.Join(skillDir1, "scripts")
	if err := os.MkdirAll(subfolder, 0755); err != nil {
		t.Fatalf("Failed to create subfolder: %v", err)
	}
	scriptFile := filepath.Join(subfolder, "run.sh")
	if err := os.WriteFile(scriptFile, []byte("echo test"), 0755); err != nil {
		t.Fatalf("Failed to write scriptFile: %v", err)
	}
	if !IsInSkillsPath(scriptFile, skillsPaths) {
		t.Errorf("Expected %s to be in skills path", scriptFile)
	}

	// Test case 3: File in skillDir2
	file2 := filepath.Join(skillDir2, "SKILL.md")
	if err := os.WriteFile(file2, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}
	if !IsInSkillsPath(file2, skillsPaths) {
		t.Errorf("Expected %s to be in skills path", file2)
	}

	// Test case 4: File outside skills paths
	outsideFile := filepath.Join(outsideDir, "other.txt")
	if err := os.WriteFile(outsideFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write outsideFile: %v", err)
	}
	if IsInSkillsPath(outsideFile, skillsPaths) {
		t.Errorf("Expected %s NOT to be in skills path", outsideFile)
	}

	// Test case 5: Empty skills paths
	if IsInSkillsPath(file1, nil) {
		t.Errorf("Expected file1 NOT to be in skills path when skillsPaths is nil")
	}
}
