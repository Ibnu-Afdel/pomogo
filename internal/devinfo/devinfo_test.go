package devinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pomogo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Case 1: No git directory
	branch := FindGitBranch(tempDir)
	if branch != "" {
		t.Errorf("expected empty branch for non-git dir, got %q", branch)
	}

	// Case 2: Standard git branch
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/feature/timer\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD file: %v", err)
	}

	subDir := filepath.Join(tempDir, "src", "ui")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	branch = FindGitBranch(subDir)
	if branch != "feature/timer" {
		t.Errorf("expected branch 'feature/timer', got %q", branch)
	}

	// Case 3: Detached HEAD (short SHA)
	if err := os.WriteFile(headPath, []byte("a1c2e3f4b5d6e7f8a1c2e3f4b5d6e7f8a1c2e3f4\n"), 0644); err != nil {
		t.Fatalf("failed to write detached HEAD file: %v", err)
	}

	branch = FindGitBranch(subDir)
	if branch != "a1c2e3f" {
		t.Errorf("expected short SHA 'a1c2e3f', got %q", branch)
	}
}
