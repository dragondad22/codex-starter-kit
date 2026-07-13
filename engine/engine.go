// Package engine implements the lifecycle-engine interface used by the CLI, CI,
// plugins, and black-box tests.
package engine

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// Engine owns lifecycle orchestration behind the public operation seam.
type Engine struct{}

// Inspection reports read-only repository facts used by planning.
type Inspection struct {
	Repository    string `json:"repository"`
	Git           bool   `json:"git"`
	Managed       bool   `json:"managed"`
	UserFileCount int    `json:"user_file_count"`
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
	managed := fileExists(filepath.Join(root, ".starter-kit", "state.json"))
	count := 0
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
		if entry.IsDir() && (relative == ".git" || relative == ".starter-kit") {
			return filepath.SkipDir
		}
		if !entry.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect repository files: %w", err)
	}

	return Inspection{
		Repository:    root,
		Git:           git,
		Managed:       managed,
		UserFileCount: count,
	}, nil
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
	_, err := os.Stat(path)
	return err == nil
}
