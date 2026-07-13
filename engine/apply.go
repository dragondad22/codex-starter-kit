package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	if plan.SchemaVersion != 1 || plan.Operation != CreateOperation {
		return ApplyResult{}, errors.New("plan schema version or operation is unsupported")
	}
	if !plan.Approval.BriefApproved || !plan.Approval.OwnerPersonaConfirmed || plan.Approval.BriefDigest == "" {
		return ApplyResult{}, errors.New("plan lacks required brief or persona approval evidence")
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
	if err := validateRelativePath(root, plan.Result.Path); err != nil || !strings.HasPrefix(plan.Result.Path, ".starter-kit/events/") || plan.Result.Ownership != "machine-evidence" || plan.Result.Source == "" {
		return ApplyResult{}, errors.New("plan result path is not a valid engine event path")
	}
	lockPath, err := lifecycleLockPath(ctx, root)
	if err != nil {
		return ApplyResult{}, err
	}
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		result := ApplyResult{1, plan.ID, ApplyStatusFailed, []string{}}
		failure := &ApplyFailure{
			Stage: "lock", Recoverable: true, Cause: err.Error(), ChangedFiles: []string{},
		}
		if eventErr := recordGitAttempt(lockPath, plan, result, failure); eventErr != nil {
			failure.Recoverable = false
			failure.Cause = fmt.Sprintf("%s; recording lock attempt failed: %v", failure.Cause, eventErr)
		}
		return result, failure
	}
	_ = lock.Close()
	defer os.Remove(lockPath)

	inspection, err := e.Inspect(ctx, plan.Repository)
	if err != nil {
		return failAndRecord(root, plan, "precondition", []string{}, true, err)
	}
	if inspection.PreconditionDigest != plan.RepositoryDigest {
		return failAndRecord(root, plan, "precondition", []string{}, true, errors.New("repository changed after the plan was created"))
	}
	if plan.NoChange {
		if !inspection.Managed || len(plan.Files) != 0 {
			return failAndRecord(root, plan, "precondition", []string{}, true, errors.New("no-change plan requires a valid unchanged managed repository"))
		}
		result := ApplyResult{1, plan.ID, ApplyStatusNoChange, []string{plan.Result.Path}}
		if err := recordApplyEvent(root, plan, result, nil); err != nil {
			return failAndRecord(root, plan, "record-result", []string{}, true, err)
		}
		return result, nil
	}
	stageRoot, err := os.MkdirTemp(root, ".starter-kit-stage-")
	if err != nil {
		return failAndRecord(root, plan, "stage", []string{}, true, err)
	}
	defer os.RemoveAll(stageRoot)
	for _, planned := range plan.Files {
		staged := filepath.Join(stageRoot, filepath.FromSlash(planned.Path))
		if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return failAndRecord(root, plan, "stage", []string{}, true, err)
		}
		if err := os.WriteFile(staged, []byte(planned.Content), 0o644); err != nil {
			return failAndRecord(root, plan, "stage", []string{}, true, err)
		}
		content, err := os.ReadFile(staged)
		if err != nil || digestBytes(content) != planned.Digest {
			if err == nil {
				err = errors.New("staged content digest mismatch")
			}
			return failAndRecord(root, plan, "stage-verify", []string{}, true, err)
		}
	}

	result := ApplyResult{1, plan.ID, ApplyStatusApplied, commitPaths(plan)}
	eventFile := plannedEventFile(plan, result, nil)
	stagedEvent := filepath.Join(stageRoot, filepath.FromSlash(eventFile.Path))
	if err := os.MkdirAll(filepath.Dir(stagedEvent), 0o755); err != nil {
		return failAndRecord(root, plan, "stage-result", []string{}, true, err)
	}
	if err := os.WriteFile(stagedEvent, []byte(eventFile.Content), 0o644); err != nil {
		return failAndRecord(root, plan, "stage-result", []string{}, true, err)
	}

	changed := make([]string, 0, len(result.ChangedFiles))
	for _, planned := range commitFiles(plan.Files, eventFile) {
		if err := ensureNoSymlinkParents(root, planned.Path); err != nil {
			return failedCommittedApply(root, plan, "commit", changed, err)
		}
		target := filepath.Join(root, filepath.FromSlash(planned.Path))
		if fileExists(target) {
			return failedCommittedApply(root, plan, "commit", changed, fmt.Errorf("refusing to overwrite existing file: %s", planned.Path))
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return failedCommittedApply(root, plan, "commit", changed, err)
		}
		staged := filepath.Join(stageRoot, filepath.FromSlash(planned.Path))
		if err := os.Rename(staged, target); err != nil {
			return failedCommittedApply(root, plan, "commit", changed, err)
		}
		changed = append(changed, planned.Path)
	}
	contractPresent, problems := validateManagedContract(root)
	if !contractPresent || len(problems) != 0 {
		return failedCommittedApply(root, plan, "postcondition", changed, fmt.Errorf("invalid managed contract: %v", problems))
	}
	return result, nil
}

func recordGitAttempt(lockPath string, plan Plan, result ApplyResult, applyErr error) error {
	directory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	filename := strings.TrimPrefix(plan.ID, "sha256:") + ".json"
	target := filepath.Join(directory, filename)
	content := []byte(plannedEventFile(plan, result, applyErr).Content)
	if existing, err := os.ReadFile(target); err == nil {
		if string(existing) == string(content) {
			return nil
		}
		return errors.New("Git attempt path already contains different evidence")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".attempt-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

func lifecycleLockPath(ctx context.Context, root string) (string, error) {
	command := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--absolute-git-dir")
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("resolve Git directory for lifecycle lock: %w", err)
	}
	gitDirectory := strings.TrimSpace(string(output))
	if gitDirectory == "" || !filepath.IsAbs(gitDirectory) {
		return "", errors.New("Git returned a non-absolute lifecycle lock directory")
	}
	return filepath.Join(gitDirectory, "starter-kit.lock"), nil
}

type operationEvent struct {
	SchemaVersion    int         `json:"schema_version"`
	Ownership        string      `json:"ownership"`
	Source           string      `json:"source"`
	PlanID           string      `json:"plan_id"`
	Operation        Operation   `json:"operation"`
	Status           ApplyStatus `json:"status"`
	RepositoryDigest string      `json:"repository_digest"`
	ChangedFiles     []string    `json:"changed_files"`
	Error            string      `json:"error,omitempty"`
	Recoverable      bool        `json:"recoverable"`
	EventDigest      string      `json:"event_digest"`
}

func plannedEventFile(plan Plan, result ApplyResult, applyErr error) PlannedFile {
	event := operationEvent{
		SchemaVersion: 1, Ownership: plan.Result.Ownership, Source: plan.Result.Source,
		PlanID: plan.ID, Operation: plan.Operation, Status: result.Status,
		RepositoryDigest: plan.RepositoryDigest, ChangedFiles: result.ChangedFiles,
	}
	var failure *ApplyFailure
	if errors.As(applyErr, &failure) {
		event.Error = failure.Cause
		event.Recoverable = failure.Recoverable
	} else if applyErr != nil {
		event.Error = applyErr.Error()
	}
	event.EventDigest = digestJSON(event)
	content := jsonDocument(event)
	return PlannedFile{
		Path: plan.Result.Path, Ownership: plan.Result.Ownership, Source: plan.Result.Source,
		Digest: digestBytes([]byte(content)), Content: content,
	}
}

func recordApplyEvent(root string, plan Plan, result ApplyResult, applyErr error) error {
	eventFile := plannedEventFile(plan, result, applyErr)
	content := []byte(eventFile.Content)
	target := filepath.Join(root, filepath.FromSlash(eventFile.Path))
	if existing, err := os.ReadFile(target); err == nil {
		if string(existing) == string(content) {
			return nil
		}
		return errors.New("operation event path already contains different evidence")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := ensureNoSymlinkParents(root, plan.Result.Path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	defer removeEmptyParents(root, filepath.Dir(target))
	temporary, err := os.CreateTemp(filepath.Dir(target), ".event-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

func commitFiles(files []PlannedFile, event PlannedFile) []PlannedFile {
	ordered := make([]PlannedFile, 0, len(files)+1)
	var state *PlannedFile
	for _, file := range files {
		if file.Path == ".starter-kit/state.json" {
			copy := file
			state = &copy
			continue
		}
		ordered = append(ordered, file)
	}
	ordered = append(ordered, event)
	if state != nil {
		ordered = append(ordered, *state)
	}
	return ordered
}

func commitPaths(plan Plan) []string {
	event := PlannedFile{Path: plan.Result.Path}
	files := commitFiles(plan.Files, event)
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths
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

func failAndRecord(root string, plan Plan, stage string, changed []string, recoverable bool, err error) (ApplyResult, error) {
	retained := append([]string{}, changed...)
	retained = append(retained, plan.Result.Path)
	result := ApplyResult{1, plan.ID, ApplyStatusFailed, retained}
	failure := &ApplyFailure{
		Stage: stage, Recoverable: recoverable, Cause: err.Error(),
		ChangedFiles: append([]string{}, retained...),
	}
	if eventErr := recordApplyEvent(root, plan, result, failure); eventErr != nil {
		result.ChangedFiles = append([]string{}, changed...)
		failure.ChangedFiles = append([]string{}, changed...)
		failure.Recoverable = false
		failure.Cause = fmt.Sprintf("%s; recording failure event failed: %v", failure.Cause, eventErr)
	}
	return result, failure
}

func failedCommittedApply(root string, plan Plan, stage string, changed []string, err error) (ApplyResult, error) {
	if rollbackErr := rollbackFiles(root, changed); rollbackErr != nil {
		return failAndRecord(root, plan, stage, changed, false, fmt.Errorf("%v; rollback failed: %v", err, rollbackErr))
	}
	return failAndRecord(root, plan, stage, []string{}, true, err)
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
