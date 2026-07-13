package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ApplyStatus is the explicit outcome of applying a plan.
type ApplyStatus string

const (
	// ApplyStatusApplied means all planned files were written and verified.
	ApplyStatusApplied ApplyStatus = "applied"
	// ApplyStatusNoChange means the repository already satisfied the operation.
	ApplyStatusNoChange ApplyStatus = "no_change"
)

const (
	// LifecycleManaged indicates a valid managed-repository contract.
	LifecycleManaged = "managed"
	// LifecycleUnmanaged indicates the local managed-repository contract is absent.
	LifecycleUnmanaged = "unmanaged"
)

// ApplyResult records the observable result of applying a plan.
type ApplyResult struct {
	SchemaVersion int         `json:"schema_version"`
	PlanID        string      `json:"plan_id"`
	Status        ApplyStatus `json:"status"`
	ChangedFiles  []string    `json:"changed_files"`
}

// RepositoryStatus reports lifecycle state through the engine seam.
type RepositoryStatus struct {
	SchemaVersion int    `json:"schema_version"`
	Repository    string `json:"repository"`
	Lifecycle     string `json:"lifecycle"`
}

// Apply rechecks plan identity and preconditions, performs local mutations, and verifies
// the rendered result.
func (e *Engine) Apply(ctx context.Context, planID string, plan Plan) (ApplyResult, error) {
	if planID == "" || planID != plan.ID {
		return ApplyResult{}, errors.New("plan identifier does not match the supplied plan")
	}
	if digestJSON(planWithBlankID(plan)) != plan.ID {
		return ApplyResult{}, errors.New("plan content does not match its identifier")
	}
	inspection, err := e.Inspect(ctx, plan.Repository)
	if err != nil {
		return ApplyResult{}, err
	}
	if digestJSON(inspection) != plan.RepositoryDigest {
		return ApplyResult{}, errors.New("repository changed after the plan was created")
	}
	if plan.NoChange {
		return ApplyResult{1, plan.ID, ApplyStatusNoChange, []string{}}, nil
	}

	changed := make([]string, 0, len(plan.Files))
	files := stateLast(plan.Files)
	for _, planned := range files {
		target := filepath.Join(plan.Repository, filepath.FromSlash(planned.Path))
		if fileExists(target) {
			return ApplyResult{}, fmt.Errorf("refusing to overwrite existing file: %s", planned.Path)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return ApplyResult{}, fmt.Errorf("create parent directory for %s: %w", planned.Path, err)
		}
		if err := os.WriteFile(target, []byte(planned.Content), 0o644); err != nil {
			return ApplyResult{}, fmt.Errorf("write %s: %w", planned.Path, err)
		}
		changed = append(changed, planned.Path)
	}
	for _, planned := range plan.Files {
		target := filepath.Join(plan.Repository, filepath.FromSlash(planned.Path))
		content, err := os.ReadFile(target)
		if err != nil {
			return ApplyResult{}, fmt.Errorf("verify %s: %w", planned.Path, err)
		}
		if digestBytes(content) != planned.Digest {
			return ApplyResult{}, fmt.Errorf("verify %s: digest mismatch", planned.Path)
		}
	}
	return ApplyResult{1, plan.ID, ApplyStatusApplied, changed}, nil
}

func stateLast(files []PlannedFile) []PlannedFile {
	ordered := make([]PlannedFile, 0, len(files))
	var state *PlannedFile
	for _, file := range files {
		if file.Path == ".starter-kit/state.json" {
			copy := file
			state = &copy
			continue
		}
		ordered = append(ordered, file)
	}
	if state != nil {
		ordered = append(ordered, *state)
	}
	return ordered
}

// Status reports lifecycle state from the authoritative local state document.
func (e *Engine) Status(_ context.Context, repository string) (RepositoryStatus, error) {
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return RepositoryStatus{}, err
	}
	statePath := filepath.Join(root, ".starter-kit", "state.json")
	content, err := os.ReadFile(statePath)
	if errors.Is(err, os.ErrNotExist) {
		return RepositoryStatus{1, root, LifecycleUnmanaged}, nil
	}
	if err != nil {
		return RepositoryStatus{}, fmt.Errorf("read repository state: %w", err)
	}
	var state struct {
		SchemaVersion int    `json:"schema_version"`
		Lifecycle     string `json:"lifecycle"`
	}
	if err := json.Unmarshal(content, &state); err != nil {
		return RepositoryStatus{}, fmt.Errorf("parse repository state: %w", err)
	}
	return RepositoryStatus{state.SchemaVersion, root, state.Lifecycle}, nil
}

func planWithBlankID(plan Plan) Plan {
	plan.ID = ""
	return plan
}
