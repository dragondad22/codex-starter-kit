// Package engine implements the lifecycle-engine interface used by the CLI, CI,
// plugins, and black-box tests.
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// Engine owns lifecycle orchestration behind the public operation seam.
type Engine struct{}

// Inspection reports read-only repository facts used by planning.
type Inspection struct {
	SchemaVersion      int      `json:"schema_version"`
	Repository         string   `json:"repository"`
	Git                bool     `json:"git"`
	GitHead            string   `json:"git_head,omitempty"`
	GitStatusDigest    string   `json:"git_status_digest,omitempty"`
	Managed            bool     `json:"managed"`
	ContractPresent    bool     `json:"contract_present"`
	Problems           []string `json:"problems"`
	UserFileCount      int      `json:"user_file_count"`
	SnapshotDigest     string   `json:"snapshot_digest"`
	PreconditionDigest string   `json:"precondition_digest"`
}

// New returns a lifecycle engine with native filesystem and Git adapters.
func New() *Engine {
	return &Engine{}
}

// Inspect gathers repository facts without modifying the repository.
func (e *Engine) Inspect(ctx context.Context, repository string) (Inspection, error) {
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return Inspection{}, err
	}

	gitCommand := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--git-dir")
	git := gitCommand.Run() == nil
	gitHead := ""
	gitStatusDigest := ""
	if git {
		headCommand := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "HEAD")
		if output, headErr := headCommand.Output(); headErr == nil {
			gitHead = string(bytes.TrimSpace(output))
		}
		statusCommand := exec.CommandContext(ctx, "git", "-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
		output, statusErr := statusCommand.Output()
		if statusErr != nil {
			return Inspection{}, fmt.Errorf("inspect Git status: %w", statusErr)
		}
		gitStatusDigest = digestBytes(output)
	}

	count := 0
	entries := make([]snapshotEntry, 0)
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if entry.IsDir() && relative == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		slashPath := filepath.ToSlash(relative)
		if relative != ".starter-kit" && !isWithinStarterKit(slashPath) {
			count++
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, linkErr := os.Readlink(path)
			if linkErr != nil {
				return linkErr
			}
			entries = append(entries, snapshotEntry{slashPath, "symlink", digestBytes([]byte(target))})
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		entries = append(entries, snapshotEntry{slashPath, "file", digestBytes(content)})
		return nil
	})
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect repository files: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	snapshotDigest := digestJSON(entries)
	contractPresent, problems := validateManagedContract(root)
	managed := contractPresent && len(problems) == 0
	inspection := Inspection{
		SchemaVersion:   1,
		Repository:      root,
		Git:             git,
		GitHead:         gitHead,
		GitStatusDigest: gitStatusDigest,
		Managed:         managed,
		ContractPresent: contractPresent,
		Problems:        problems,
		UserFileCount:   count,
		SnapshotDigest:  snapshotDigest,
	}
	inspection.PreconditionDigest = digestJSON(struct {
		Repository      string `json:"repository"`
		Git             bool   `json:"git"`
		GitHead         string `json:"git_head"`
		GitStatusDigest string `json:"git_status_digest"`
		SnapshotDigest  string `json:"snapshot_digest"`
	}{root, git, gitHead, gitStatusDigest, snapshotDigest})
	return inspection, nil
}

type snapshotEntry struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Digest string `json:"digest"`
}

func isWithinStarterKit(path string) bool {
	return path == ".starter-kit" || len(path) > len(".starter-kit/") && path[:len(".starter-kit/")] == ".starter-kit/"
}

func cleanRepositoryRoot(repository string) (string, error) {
	if repository == "" {
		return "", errors.New("repository path is required")
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		return "", fmt.Errorf("resolve repository path: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("stat repository: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository is not a directory: %s", root)
	}
	return filepath.Clean(root), nil
}

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
