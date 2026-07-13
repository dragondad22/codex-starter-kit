package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ApplyStatus is the explicit outcome of applying a plan.
type ApplyStatus string

const (
	// ApplyStatusApplied means all planned files were written and verified.
	ApplyStatusApplied ApplyStatus = "applied"
	// ApplyStatusNoChange means the repository already satisfied the operation.
	ApplyStatusNoChange ApplyStatus = "no_change"
	// ApplyStatusFailed means the operation did not commit and reports recovery state.
	ApplyStatusFailed ApplyStatus = "failed"
)

const (
	// LifecycleManaged indicates a valid managed-repository contract.
	LifecycleManaged = "managed"
	// LifecycleUnmanaged indicates the local managed-repository contract is absent.
	LifecycleUnmanaged = "unmanaged"
	// LifecycleManagedDegraded indicates a present but invalid local contract.
	LifecycleManagedDegraded = "managed_degraded"
)

// ApplyResult records the observable result of applying a plan.
type ApplyResult struct {
	SchemaVersion int         `json:"schema_version"`
	PlanID        string      `json:"plan_id"`
	Status        ApplyStatus `json:"status"`
	ChangedFiles  []string    `json:"changed_files"`
}

// ApplyFailure describes an explicit recoverable or non-recoverable apply failure.
type ApplyFailure struct {
	Stage        string   `json:"stage"`
	Recoverable  bool     `json:"recoverable"`
	ChangedFiles []string `json:"changed_files"`
	Cause        string   `json:"cause"`
}

func (f *ApplyFailure) Error() string {
	return fmt.Sprintf("apply failed during %s: %s", f.Stage, f.Cause)
}

// RepositoryStatus reports lifecycle state through the engine seam.
type RepositoryStatus struct {
	SchemaVersion int      `json:"schema_version"`
	Repository    string   `json:"repository"`
	Lifecycle     string   `json:"lifecycle"`
	Problems      []string `json:"problems"`
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
	root, err := cleanRepositoryRoot(plan.Repository)
	if err != nil {
		return ApplyResult{}, err
	}
	if root != plan.Repository {
		return ApplyResult{}, errors.New("plan repository path is not canonical")
	}
	if err := validatePlanFiles(root, plan.Files); err != nil {
		return ApplyResult{}, err
	}
	inspection, err := e.Inspect(ctx, plan.Repository)
	if err != nil {
		return ApplyResult{}, err
	}
	if inspection.PreconditionDigest != plan.RepositoryDigest {
		return ApplyResult{}, errors.New("repository changed after the plan was created")
	}
	if plan.NoChange {
		if !inspection.Managed || len(plan.Files) != 0 {
			return ApplyResult{}, errors.New("no-change plan requires a valid unchanged managed repository")
		}
		return ApplyResult{1, plan.ID, ApplyStatusNoChange, []string{}}, nil
	}
	lockPath := filepath.Join(root, ".starter-kit.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return ApplyResult{1, plan.ID, ApplyStatusFailed, []string{}}, &ApplyFailure{
			Stage: "lock", Recoverable: true, Cause: err.Error(), ChangedFiles: []string{},
		}
	}
	_ = lock.Close()
	defer os.Remove(lockPath)

	stageRoot, err := os.MkdirTemp(root, ".starter-kit-stage-")
	if err != nil {
		return ApplyResult{1, plan.ID, ApplyStatusFailed, []string{}}, &ApplyFailure{
			Stage: "stage", Recoverable: true, Cause: err.Error(), ChangedFiles: []string{},
		}
	}
	defer os.RemoveAll(stageRoot)
	for _, planned := range plan.Files {
		staged := filepath.Join(stageRoot, filepath.FromSlash(planned.Path))
		if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return failedApply(plan.ID, "stage", []string{}, err)
		}
		if err := os.WriteFile(staged, []byte(planned.Content), 0o644); err != nil {
			return failedApply(plan.ID, "stage", []string{}, err)
		}
		content, err := os.ReadFile(staged)
		if err != nil || digestBytes(content) != planned.Digest {
			if err == nil {
				err = errors.New("staged content digest mismatch")
			}
			return failedApply(plan.ID, "stage-verify", []string{}, err)
		}
	}

	changed := make([]string, 0, len(plan.Files))
	for _, planned := range stateLast(plan.Files) {
		if err := ensureNoSymlinkParents(root, planned.Path); err != nil {
			return failedCommittedApply(root, plan.ID, "commit", changed, err)
		}
		target := filepath.Join(root, filepath.FromSlash(planned.Path))
		if fileExists(target) {
			return failedCommittedApply(root, plan.ID, "commit", changed, fmt.Errorf("refusing to overwrite existing file: %s", planned.Path))
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return failedCommittedApply(root, plan.ID, "commit", changed, err)
		}
		staged := filepath.Join(stageRoot, filepath.FromSlash(planned.Path))
		if err := os.Rename(staged, target); err != nil {
			return failedCommittedApply(root, plan.ID, "commit", changed, err)
		}
		changed = append(changed, planned.Path)
	}
	contractPresent, problems := validateManagedContract(root)
	if !contractPresent || len(problems) != 0 {
		return failedCommittedApply(root, plan.ID, "postcondition", changed, fmt.Errorf("invalid managed contract: %v", problems))
	}
	return ApplyResult{1, plan.ID, ApplyStatusApplied, changed}, nil
}

func validatePlanFiles(root string, files []PlannedFile) error {
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		if err := validateRelativePath(root, file.Path); err != nil {
			return fmt.Errorf("invalid planned path %q: %w", file.Path, err)
		}
		if _, duplicate := seen[file.Path]; duplicate {
			return fmt.Errorf("duplicate planned path: %s", file.Path)
		}
		seen[file.Path] = struct{}{}
		if file.Ownership == "" || file.Source == "" || file.Digest == "" {
			return fmt.Errorf("incomplete planned provenance for %s", file.Path)
		}
		if digestBytes([]byte(file.Content)) != file.Digest {
			return fmt.Errorf("planned content digest mismatch: %s", file.Path)
		}
	}
	return nil
}

func ensureNoSymlinkParents(root, slashPath string) error {
	current := root
	parts := strings.Split(filepath.FromSlash(slashPath), string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("planned path traverses symlink: %s", slashPath)
		}
	}
	return nil
}

func rollbackFiles(root string, changed []string) error {
	var rollbackError error
	for index := len(changed) - 1; index >= 0; index-- {
		target := filepath.Join(root, filepath.FromSlash(changed[index]))
		if err := os.Remove(target); err != nil && rollbackError == nil {
			rollbackError = err
		}
		removeEmptyParents(root, filepath.Dir(target))
	}
	return rollbackError
}

func removeEmptyParents(root, directory string) {
	for directory != root {
		if err := os.Remove(directory); err != nil {
			return
		}
		directory = filepath.Dir(directory)
	}
}

func failedApply(planID, stage string, changed []string, err error) (ApplyResult, error) {
	result := ApplyResult{1, planID, ApplyStatusFailed, append([]string{}, changed...)}
	return result, &ApplyFailure{
		Stage: stage, Recoverable: true, Cause: err.Error(), ChangedFiles: append([]string{}, changed...),
	}
}

func failedCommittedApply(root, planID, stage string, changed []string, err error) (ApplyResult, error) {
	if rollbackErr := rollbackFiles(root, changed); rollbackErr != nil {
		result := ApplyResult{1, planID, ApplyStatusFailed, append([]string{}, changed...)}
		return result, &ApplyFailure{
			Stage: stage, Recoverable: false,
			Cause:        fmt.Sprintf("%v; rollback failed: %v", err, rollbackErr),
			ChangedFiles: append([]string{}, changed...),
		}
	}
	return failedApply(planID, stage, []string{}, err)
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
func (e *Engine) Status(ctx context.Context, repository string) (RepositoryStatus, error) {
	inspection, err := e.Inspect(ctx, repository)
	if err != nil {
		return RepositoryStatus{}, err
	}
	if inspection.Managed {
		return RepositoryStatus{1, inspection.Repository, LifecycleManaged, []string{}}, nil
	}
	if inspection.ContractPresent {
		return RepositoryStatus{1, inspection.Repository, LifecycleManagedDegraded, inspection.Problems}, nil
	}
	return RepositoryStatus{1, inspection.Repository, LifecycleUnmanaged, []string{}}, nil
}

func planWithBlankID(plan Plan) Plan {
	plan.ID = ""
	return plan
}
