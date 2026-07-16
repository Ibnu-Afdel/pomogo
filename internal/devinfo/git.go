package devinfo

import (
	"os"
	"path/filepath"
	"strings"
)

// FindGitBranch searches upward from startPath for a .git directory (or file)
// and returns the current branch name or short SHA.
func FindGitBranch(startPath string) string {
	current, err := filepath.Abs(startPath)
	if err != nil {
		current = startPath
	}
	for {
		gitPath := filepath.Join(current, ".git")
		fi, err := os.Stat(gitPath)
		if err == nil {
			var headPath string
			if fi.IsDir() {
				headPath = filepath.Join(gitPath, "HEAD")
			} else {
				// git worktree or submodule, .git is a file containing "gitdir: ..."
				bytes, err := os.ReadFile(gitPath)
				if err == nil {
					line := strings.TrimSpace(string(bytes))
					if strings.HasPrefix(line, "gitdir: ") {
						realGitDir := strings.TrimPrefix(line, "gitdir: ")
						if !filepath.IsAbs(realGitDir) {
							realGitDir = filepath.Clean(filepath.Join(current, realGitDir))
						}
						headPath = filepath.Join(realGitDir, "HEAD")
					}
				}
			}

			if headPath != "" {
				bytes, err := os.ReadFile(headPath)
				if err == nil {
					line := strings.TrimSpace(string(bytes))
					if strings.HasPrefix(line, "ref: ") {
						ref := strings.TrimPrefix(line, "ref: ")
						if strings.HasPrefix(ref, "refs/heads/") {
							return strings.TrimPrefix(ref, "refs/heads/")
						}
						return ref
					}
					// Detached HEAD, return short SHA
					if len(line) >= 7 {
						return line[:7]
					}
					return line
				}
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}
