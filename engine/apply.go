package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	// LifecycleSetupIncomplete indicates a recoverable create transaction did not finish.
	LifecycleSetupIncomplete = "setup_incomplete"
)

// ApplyResult records the observable result of applying a plan.
type ApplyResult struct {
	SchemaVersion int         `json:"schema_version"`
	PlanID        string      `json:"plan_id"`
	Status        ApplyStatus `json:"status"`
	ChangedFiles  []string    `json:"changed_files"`
	Recovery      []string    `json:"recovery"`
	Evidence      []string    `json:"evidence"`
}

// ApplyFailure describes an explicit recoverable or non-recoverable apply failure.
type ApplyFailure struct {
	Stage        string                   `json:"stage"`
	Recoverable  bool                     `json:"recoverable"`
	ChangedFiles []string                 `json:"changed_files"`
	Conflicts    []ReconciliationConflict `json:"conflicts"`
	Recovery     []string                 `json:"recovery"`
	Evidence     []string                 `json:"evidence"`
	Cause        string                   `json:"cause"`
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
	Recovery      []string `json:"recovery"`
	Evidence      []string `json:"evidence"`
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
	if !plan.Approval.BriefApproved || !plan.Approval.OwnerPersonaConfirmed || !validSHA256Digest(plan.Approval.BriefDigest) {
		return ApplyResult{}, errors.New("plan lacks required brief or persona approval evidence")
	}
	if !validSHA256Digest(plan.RepositoryDigest) {
		return ApplyResult{}, errors.New("plan repository digest is invalid")
	}
	root, err := cleanRepositoryRoot(plan.Repository)
	if err != nil {
		return ApplyResult{}, err
	}
	if root != plan.Repository {
		return ApplyResult{}, errors.New("plan repository path is not canonical")
	}
	if err := validatePlanFiles(root, plan.Files); err != nil {
		return rejectPlanBeforeTransaction(ctx, root, plan, err)
	}
	if !plan.NoChange {
		if err := validateCreatePlanContract(plan.Files); err != nil {
			return rejectPlanBeforeTransaction(ctx, root, plan, err)
		}
	}
	expectedResultPath := operationEventPath(plan.Operation, plan.RepositoryDigest)
	if err := validateRelativePath(root, plan.Result.Path); err != nil || plan.Result.Path != expectedResultPath || !strings.HasPrefix(plan.Result.Path, ".starter-kit/events/") || plan.Result.Ownership != "machine-evidence" || plan.Result.Source != "engine:apply:v1" {
		return rejectPlanBeforeTransaction(ctx, root, plan, errors.New("plan result path is not the approved engine event path"))
	}
	lockPath, err := lifecycleLockPath(ctx, root)
	if err != nil {
		return ApplyResult{}, err
	}
	leaseContent, recovery, evidence, err := acquireLifecycleLock(lockPath, plan.ID, e.clock.Now().UTC())
	if err != nil {
		result := ApplyResult{
			SchemaVersion: 1, PlanID: plan.ID, Status: ApplyStatusFailed,
			ChangedFiles: []string{}, Recovery: []string{}, Evidence: []string{},
		}
		failure := &ApplyFailure{
			Stage: "lock", Recoverable: true, Cause: err.Error(), ChangedFiles: []string{},
			Conflicts: []ReconciliationConflict{},
			Recovery: []string{
				"wait for the active lifecycle operation to finish and retry the same immutable plan",
				"if the recorded process is no longer active, preserve the lease and retry for automatic stale recovery",
			},
			Evidence: []string{},
		}
		attemptReference := "git:starter-kit-attempts/" + strings.TrimPrefix(plan.ID, "sha256:") + ".json"
		failure.Evidence = []string{attemptReference}
		result.Evidence = []string{attemptReference}
		if eventErr := recordGitAttempt(lockPath, plan, result, failure); eventErr != nil {
			failure.Evidence = []string{}
			result.Evidence = []string{}
			failure.Recoverable = false
			failure.Cause = fmt.Sprintf("%s; recording lock attempt failed: %v", failure.Cause, eventErr)
		}
		return result, failure
	}
	defer releaseLifecycleLock(lockPath, leaseContent)
	fail := func(stage string, changed []string, recoverable bool, err error) (ApplyResult, error) {
		return failAndRecord(root, plan, stage, changed, recoverable, recovery, evidence, err)
	}
	stageRecovery, stageEvidence, err := recoverAbandonedStages(root, lockPath, plan.ID, evidence, e.clock.Now().UTC())
	recovery = append(recovery, stageRecovery...)
	evidence = append(evidence, stageEvidence...)
	if err != nil {
		recovery = append(recovery,
			"preserve the unrecognized staging tree and review its ownership",
			"retry only after explicit reconciliation removes or reclassifies the conflict",
		)
		return fail("recover-stage", []string{}, true, err)
	}

	inspection, err := e.Inspect(ctx, plan.Repository)
	if err != nil {
		return fail("precondition", []string{}, true, err)
	}
	if result, priorEventDigest, applied := stableAppliedCreateResult(root, inspection, plan); applied {
		reference, eventErr := recordReplayAttempt(lockPath, plan, priorEventDigest, e.clock.Now().UTC())
		if eventErr != nil {
			return fail("record-replay", []string{}, true, eventErr)
		}
		result.Recovery = append([]string{}, result.Recovery...)
		result.Evidence = append(result.Evidence, reference)
		return result, nil
	}
	if inspection.PreconditionDigest != plan.RepositoryDigest {
		if resumable, conflicts := interruptedCreateMatchesPlan(root, plan); resumable {
			recovery = append(recovery, "resumed the same immutable create plan without replacing its matching committed prefix")
		} else {
			if len(conflicts) != 0 {
				return reconcileApplyFailure(root, plan, recovery, evidence, conflicts)
			}
			return fail("precondition", []string{}, true, errors.New("repository changed after the plan was created"))
		}
	}
	if plan.NoChange {
		if !inspection.Managed || len(plan.Files) != 0 {
			return fail("precondition", []string{}, true, errors.New("no-change plan requires a valid unchanged managed repository"))
		}
		result := ApplyResult{
			SchemaVersion: 1, PlanID: plan.ID, Status: ApplyStatusNoChange,
			ChangedFiles: []string{plan.Result.Path}, Recovery: recovery, Evidence: evidence,
		}
		if err := recordApplyEvent(root, plan, result, nil); err != nil {
			return fail("record-result", []string{}, true, err)
		}
		return result, nil
	}
	var currentLease lifecycleLease
	if err := json.Unmarshal(leaseContent, &currentLease); err != nil {
		return fail("stage", []string{}, false, errors.New("active lifecycle lease is unreadable"))
	}
	stageRoot := filepath.Join(root, ".starter-kit-stage-"+currentLease.Token)
	if err := os.Mkdir(stageRoot, 0o700); err != nil {
		return fail("stage", []string{}, true, err)
	}
	marker := stageTransactionMarker{
		SchemaVersion: 1, Ownership: "machine-state", Source: "engine:apply:v1",
		LeaseToken: currentLease.Token, PlanID: plan.ID, CreatedAt: e.clock.Now().UTC(),
	}
	if err := os.WriteFile(filepath.Join(stageRoot, ".starter-kit-transaction.json"), []byte(jsonDocument(marker)), 0o600); err != nil {
		_ = os.RemoveAll(stageRoot)
		return fail("stage", []string{}, true, err)
	}
	defer os.RemoveAll(stageRoot)
	for _, planned := range plan.Files {
		staged := filepath.Join(stageRoot, filepath.FromSlash(planned.Path))
		if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return fail("stage", []string{}, true, err)
		}
		if err := os.WriteFile(staged, []byte(planned.Content), 0o644); err != nil {
			return fail("stage", []string{}, true, err)
		}
		content, err := os.ReadFile(staged)
		if err != nil || digestBytes(content) != planned.Digest {
			if err == nil {
				err = errors.New("staged content digest mismatch")
			}
			return fail("stage-verify", []string{}, true, err)
		}
	}

	result := ApplyResult{
		SchemaVersion: 1, PlanID: plan.ID, Status: ApplyStatusApplied,
		ChangedFiles: commitPaths(plan), Recovery: recovery, Evidence: evidence,
	}
	eventFile := plannedEventFile(plan, result, nil)
	stagedEvent := filepath.Join(stageRoot, filepath.FromSlash(eventFile.Path))
	if err := os.MkdirAll(filepath.Dir(stagedEvent), 0o755); err != nil {
		return fail("stage-result", []string{}, true, err)
	}
	if err := os.WriteFile(stagedEvent, []byte(eventFile.Content), 0o644); err != nil {
		return fail("stage-result", []string{}, true, err)
	}

	changed := make([]string, 0, len(result.ChangedFiles))
	for _, planned := range commitFiles(plan.Files, eventFile) {
		if err := ensureNoSymlinkParents(root, planned.Path); err != nil {
			return failedCommittedApply(root, plan, "commit", changed, recovery, evidence, err)
		}
		target := filepath.Join(root, filepath.FromSlash(planned.Path))
		if fileExists(target) {
			if existingPlannedFileMatches(target, planned, plan) {
				continue
			}
			return failedCommittedApply(root, plan, "commit", changed, recovery, evidence, fmt.Errorf("refusing to overwrite existing file: %s", planned.Path))
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return failedCommittedApply(root, plan, "commit", changed, recovery, evidence, err)
		}
		staged := filepath.Join(stageRoot, filepath.FromSlash(planned.Path))
		if err := os.Rename(staged, target); err != nil {
			return failedCommittedApply(root, plan, "commit", changed, recovery, evidence, err)
		}
		changed = append(changed, planned.Path)
	}
	contractPresent, problems := validateManagedContract(root)
	if !contractPresent || len(problems) != 0 {
		return failedCommittedApply(root, plan, "postcondition", changed, recovery, evidence, fmt.Errorf("invalid managed contract: %v", problems))
	}
	return result, nil
}

func interruptedCreateMatchesPlan(root string, plan Plan) (bool, []ReconciliationConflict) {
	if plan.NoChange || fileExists(filepath.Join(root, ".starter-kit", "state.json")) {
		return false, nil
	}
	allowed := make(map[string]PlannedFile, len(plan.Files))
	for _, file := range plan.Files {
		allowed[file.Path] = file
	}
	matched := 0
	conflicts := []ReconciliationConflict{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() && relative == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() && strings.HasPrefix(relative, ".starter-kit-stage-") {
			return filepath.SkipDir
		}
		slashPath := filepath.ToSlash(relative)
		if entry.IsDir() {
			if approvedDirectoryPrefix(slashPath, allowed) {
				return nil
			}
			conflicts = append(conflicts, reconciliationConflict(slashPath, "directory", "unplanned directory appeared after planning"))
			return filepath.SkipDir
		}
		planned, approved := allowed[slashPath]
		kind := "file"
		if entry.Type()&os.ModeSymlink != 0 {
			kind = "symlink"
		}
		if !approved || kind == "symlink" || !existingPlannedFileMatches(path, planned, plan) {
			conflicts = append(conflicts, reconciliationConflict(slashPath, kind, "content does not match the immutable create plan"))
			return nil
		}
		matched++
		return nil
	})
	if err != nil {
		return false, []ReconciliationConflict{reconciliationConflict(".", "repository", "repository progress could not be inspected safely")}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Path < conflicts[j].Path })
	return matched != 0 && len(conflicts) == 0, conflicts
}

func approvedDirectoryPrefix(directory string, allowed map[string]PlannedFile) bool {
	prefix := directory + "/"
	for path := range allowed {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func existingPlannedFileMatches(target string, planned PlannedFile, plan Plan) bool {
	if planned.Path == plan.Result.Path {
		return false
	}
	content, err := os.ReadFile(target)
	return err == nil && digestBytes(content) == planned.Digest
}

func reconciliationConflict(path, kind, reason string) ReconciliationConflict {
	return ReconciliationConflict{Path: reviewableConflictPath(path), Kind: kind, Ownership: "user-owned", Reason: reason}
}

func reconcileApplyFailure(root string, plan Plan, recovery, evidence []string, conflicts []ReconciliationConflict) (ApplyResult, error) {
	recovery = append(recovery,
		"review and preserve every listed conflict",
		"replan only after the repository facts or explicit reconciliation authority change",
	)
	originalEvidence := append([]string{}, evidence...)
	evidence = append(evidence, plan.Result.Path)
	result := ApplyResult{
		SchemaVersion: 1, PlanID: plan.ID, Status: ApplyStatusFailed,
		ChangedFiles: []string{plan.Result.Path}, Recovery: recovery, Evidence: evidence,
	}
	failure := &ApplyFailure{
		Stage: "reconcile", Recoverable: true, ChangedFiles: []string{plan.Result.Path},
		Conflicts: conflicts, Recovery: recovery, Evidence: evidence,
		Cause: "repository content conflicts with the immutable create plan",
	}
	if err := recordApplyEvent(root, plan, result, failure); err != nil {
		result.ChangedFiles = []string{}
		result.Evidence = originalEvidence
		failure.ChangedFiles = []string{}
		failure.Evidence = originalEvidence
		failure.Recoverable = false
		failure.Cause += "; recording reconciliation evidence failed: " + err.Error()
	}
	return result, failure
}

func stableAppliedCreateResult(root string, inspection Inspection, plan Plan) (ApplyResult, string, bool) {
	if !inspection.Managed {
		return ApplyResult{}, "", false
	}
	if !plan.NoChange {
		if !plannedFilesMatch(root, plan.Files) {
			return ApplyResult{}, "", false
		}
	} else if len(plan.Files) != 0 {
		return ApplyResult{}, "", false
	}
	event, valid := readCompletedCreateEvent(root, plan)
	if !valid {
		return ApplyResult{}, "", false
	}
	return ApplyResult{
		SchemaVersion: 1, PlanID: plan.ID, Status: event.Status,
		ChangedFiles: append([]string{}, event.ChangedFiles...),
		Recovery:     append([]string{}, event.Recovery...), Evidence: append([]string{}, event.Evidence...),
	}, event.EventDigest, true
}

func readCompletedCreateEvent(root string, plan Plan) (operationEvent, bool) {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(plan.Result.Path)))
	if err != nil {
		return operationEvent{}, false
	}
	var event operationEvent
	if json.Unmarshal(content, &event) != nil {
		return operationEvent{}, false
	}
	recordedDigest := event.EventDigest
	event.EventDigest = ""
	expectedStatus := ApplyStatusApplied
	expectedFiles := commitPaths(plan)
	if plan.NoChange {
		expectedStatus = ApplyStatusNoChange
		expectedFiles = []string{plan.Result.Path}
	}
	valid := event.SchemaVersion == 1 && event.Ownership == "machine-evidence" && event.Source == "engine:apply:v1" &&
		event.PlanID == plan.ID && event.Operation == plan.Operation && event.Status == expectedStatus &&
		event.Actor == "not-configured" && event.Authority == "not-configured" && event.RepositoryDigest == plan.RepositoryDigest &&
		reflectStringSlices(event.ChangedFiles, expectedFiles) && len(event.ExternalEffects) == 0 && len(event.Diagnostics) == 0 &&
		len(event.Conflicts) == 0 && event.Error == "" && !event.Recoverable && recordedDigest != "" && digestJSON(event) == recordedDigest &&
		validSuccessfulRecoveryMetadata(root, plan, event.Recovery, event.Evidence)
	event.EventDigest = recordedDigest
	return event, valid
}

func reflectStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

type replayAttemptEvidence struct {
	SchemaVersion    int         `json:"schema_version"`
	Ownership        string      `json:"ownership"`
	Source           string      `json:"source"`
	PlanID           string      `json:"plan_id"`
	Operation        Operation   `json:"operation"`
	Status           ApplyStatus `json:"status"`
	RepositoryDigest string      `json:"repository_digest"`
	PriorEventDigest string      `json:"prior_event_digest"`
	ChangedFiles     []string    `json:"changed_files"`
	ObservedAt       time.Time   `json:"observed_at"`
	AttemptToken     string      `json:"attempt_token"`
	EvidenceDigest   string      `json:"evidence_digest"`
}

func recordReplayAttempt(lockPath string, plan Plan, priorEventDigest string, observedAt time.Time) (string, error) {
	token, err := randomLeaseToken()
	if err != nil {
		return "", err
	}
	record := replayAttemptEvidence{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "engine:apply-replay:v1",
		PlanID: plan.ID, Operation: plan.Operation, Status: ApplyStatusNoChange,
		RepositoryDigest: plan.RepositoryDigest, PriorEventDigest: priorEventDigest,
		ChangedFiles: []string{}, ObservedAt: observedAt, AttemptToken: token,
	}
	record.EvidenceDigest = digestJSON(record)
	filename := "replay-" + strings.TrimPrefix(record.EvidenceDigest, "sha256:") + ".json"
	directory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", err
	}
	if err := writeEvidenceOnce(filepath.Join(directory, filename), []byte(jsonDocument(record))); err != nil {
		return "", err
	}
	return "git:starter-kit-attempts/" + filename, nil
}

func rejectPlanBeforeTransaction(ctx context.Context, root string, plan Plan, validationErr error) (ApplyResult, error) {
	diagnostics := redactDiagnostics([]string{validationErr.Error()})
	failure := &ApplyFailure{
		Stage: "validate-plan", Recoverable: true, ChangedFiles: []string{}, Cause: diagnostics[0],
	}
	result := ApplyResult{
		SchemaVersion: 1, PlanID: plan.ID, Status: ApplyStatusFailed,
		ChangedFiles: []string{}, Recovery: []string{}, Evidence: []string{},
	}
	lockPath, err := lifecycleLockPath(ctx, root)
	if err != nil {
		failure.Recoverable = false
		failure.Cause = fmt.Sprintf("%s; locating rejected-plan evidence failed: %v", failure.Cause, err)
		return result, failure
	}
	evidencePlan := plan
	evidencePlan.Result = PlannedResult{
		Path:      operationEventPath(plan.Operation, plan.RepositoryDigest),
		Ownership: "machine-evidence",
		Source:    "engine:apply:v1",
	}
	if err := recordGitAttempt(lockPath, evidencePlan, result, failure); err != nil {
		failure.Recoverable = false
		failure.Cause = fmt.Sprintf("%s; recording rejected-plan evidence failed: %v", failure.Cause, err)
	}
	return result, failure
}

func validateCreatePlanContract(files []PlannedFile) error {
	if len(files) != len(createFileOwnershipV1) {
		return errors.New("create plan does not contain the exact approved artifact set")
	}
	for _, file := range files {
		ownership, expected := createFileOwnershipV1[file.Path]
		if !expected || file.Ownership != ownership || file.Source != "engine:create:v1" {
			return errors.New("create plan artifact ownership or provenance is not approved")
		}
	}
	return nil
}

func recordGitAttempt(lockPath string, plan Plan, result ApplyResult, applyErr error) error {
	directory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	filename := strings.TrimPrefix(plan.ID, "sha256:") + ".json"
	target := filepath.Join(directory, filename)
	content := []byte(plannedEventFile(plan, result, applyErr).Content)
	return writeEvidenceOnce(target, content)
}

func lifecycleLockPath(ctx context.Context, root string) (string, error) {
	command := structuredGitCommand(ctx, root, "rev-parse", "--absolute-git-dir")
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
	SchemaVersion    int                      `json:"schema_version"`
	Ownership        string                   `json:"ownership"`
	Source           string                   `json:"source"`
	PlanID           string                   `json:"plan_id"`
	Operation        Operation                `json:"operation"`
	Status           ApplyStatus              `json:"status"`
	Actor            string                   `json:"actor"`
	Authority        string                   `json:"authority"`
	RepositoryDigest string                   `json:"repository_digest"`
	ChangedFiles     []string                 `json:"changed_files"`
	ExternalEffects  []string                 `json:"external_effects"`
	Diagnostics      []string                 `json:"diagnostics"`
	Conflicts        []ReconciliationConflict `json:"conflicts"`
	Recovery         []string                 `json:"recovery"`
	Evidence         []string                 `json:"evidence"`
	Error            string                   `json:"error,omitempty"`
	Recoverable      bool                     `json:"recoverable"`
	EventDigest      string                   `json:"event_digest"`
}

func plannedEventFile(plan Plan, result ApplyResult, applyErr error) PlannedFile {
	event := operationEvent{
		SchemaVersion: 1, Ownership: plan.Result.Ownership, Source: plan.Result.Source,
		PlanID: plan.ID, Operation: plan.Operation, Status: result.Status,
		Actor: "not-configured", Authority: "not-configured",
		RepositoryDigest: plan.RepositoryDigest, ChangedFiles: result.ChangedFiles,
		ExternalEffects: []string{}, Diagnostics: []string{},
		Conflicts: []ReconciliationConflict{},
		Recovery:  append([]string{}, result.Recovery...), Evidence: append([]string{}, result.Evidence...),
	}
	var failure *ApplyFailure
	if errors.As(applyErr, &failure) {
		event.Error = failure.Cause
		event.Recoverable = failure.Recoverable
		event.Conflicts = append([]ReconciliationConflict{}, failure.Conflicts...)
		event.Recovery = append([]string{}, failure.Recovery...)
		event.Evidence = append([]string{}, failure.Evidence...)
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

type lifecycleLease struct {
	SchemaVersion int       `json:"schema_version"`
	Token         string    `json:"token"`
	PlanID        string    `json:"plan_id"`
	PID           int       `json:"pid"`
	CreatedAt     time.Time `json:"created_at"`
}

func acquireLifecycleLock(lockPath, planID string, now time.Time) ([]byte, []string, []string, error) {
	token, err := randomLeaseToken()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate lifecycle lease token: %w", err)
	}
	lease := lifecycleLease{
		SchemaVersion: 1, Token: token, PlanID: planID,
		PID: os.Getpid(), CreatedAt: now,
	}
	content := []byte(jsonDocument(lease))
	recovery := []string{}
	evidence := []string{}
	for attempt := 0; attempt < 2; attempt++ {
		lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			if _, writeErr := lock.Write(content); writeErr != nil {
				_ = lock.Close()
				_ = os.Remove(lockPath)
				return nil, nil, nil, fmt.Errorf("write lifecycle lease: %w", writeErr)
			}
			if syncErr := lock.Sync(); syncErr != nil {
				_ = lock.Close()
				_ = os.Remove(lockPath)
				return nil, nil, nil, fmt.Errorf("sync lifecycle lease: %w", syncErr)
			}
			if closeErr := lock.Close(); closeErr != nil {
				_ = os.Remove(lockPath)
				return nil, nil, nil, fmt.Errorf("close lifecycle lease: %w", closeErr)
			}
			return content, recovery, evidence, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, nil, nil, fmt.Errorf("acquire lifecycle lease: %w", err)
		}
		existing, readErr := os.ReadFile(lockPath)
		if readErr != nil {
			return nil, nil, nil, fmt.Errorf("read existing lifecycle lease: %w", readErr)
		}
		var stale lifecycleLease
		if json.Unmarshal(existing, &stale) != nil || stale.SchemaVersion != 1 || !validLeaseToken(stale.Token) || stale.PlanID == "" || stale.PID <= 0 || stale.CreatedAt.IsZero() {
			return nil, nil, nil, errors.New("lifecycle lease is active or malformed and requires review")
		}
		if processAlive(stale.PID) || now.Before(stale.CreatedAt.Add(5*time.Minute)) {
			return nil, nil, nil, errors.New("lifecycle lease is held by an active or recently started process")
		}
		reference, archiveErr := archiveStaleLifecycleLease(lockPath, existing, stale, now, lease.Token)
		if archiveErr != nil {
			return nil, nil, nil, archiveErr
		}
		recovery = append(recovery, "recovered a stale lifecycle lease after confirming its recorded process was not active")
		evidence = append(evidence, reference)
	}
	return nil, nil, nil, errors.New("lifecycle lease changed during stale recovery")
}

func randomLeaseToken() (string, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func validLeaseToken(token string) bool {
	decoded, err := hex.DecodeString(token)
	return err == nil && len(decoded) == 16 && token == strings.ToLower(token)
}

func releaseLifecycleLock(lockPath string, leaseContent []byte) {
	current, err := os.ReadFile(lockPath)
	if err == nil && string(current) == string(leaseContent) {
		_ = os.Remove(lockPath)
	}
}

type abandonedStageEvidence struct {
	SchemaVersion  int       `json:"schema_version"`
	Ownership      string    `json:"ownership"`
	Source         string    `json:"source"`
	PlanID         string    `json:"plan_id"`
	LeaseToken     string    `json:"lease_token"`
	StageDigest    string    `json:"stage_digest"`
	RecoveredAt    time.Time `json:"recovered_at"`
	EvidenceDigest string    `json:"evidence_digest"`
}

type staleLeaseEvidence struct {
	SchemaVersion  int            `json:"schema_version"`
	Ownership      string         `json:"ownership"`
	Source         string         `json:"source"`
	Lease          lifecycleLease `json:"lease"`
	LeaseDigest    string         `json:"lease_digest"`
	RecoveredAt    time.Time      `json:"recovered_at"`
	EvidenceDigest string         `json:"evidence_digest"`
}

func archiveStaleLifecycleLease(lockPath string, expected []byte, stale lifecycleLease, recoveredAt time.Time, recoveryToken string) (string, error) {
	directory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("create stale-lease evidence directory: %w", err)
	}
	quarantine := filepath.Join(directory, ".stale-recovery-"+recoveryToken)
	if err := os.Mkdir(quarantine, 0o700); err != nil {
		return "", fmt.Errorf("create stale-lease quarantine: %w", err)
	}
	quarantinedLease := filepath.Join(quarantine, "lease.json")
	if err := os.Rename(lockPath, quarantinedLease); err != nil {
		_ = os.Remove(quarantine)
		return "", fmt.Errorf("quarantine stale lifecycle lease: %w", err)
	}
	moved, err := os.ReadFile(quarantinedLease)
	if err != nil || string(moved) != string(expected) {
		return "", errors.New("lifecycle lease changed during stale recovery; quarantined content was preserved")
	}
	record := staleLeaseEvidence{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "engine:apply:v1",
		Lease: stale, LeaseDigest: digestBytes(moved), RecoveredAt: recoveredAt,
	}
	record.EvidenceDigest = digestJSON(record)
	filename := "stale-lock-" + strings.TrimPrefix(record.EvidenceDigest, "sha256:") + ".json"
	if err := writeEvidenceOnce(filepath.Join(directory, filename), []byte(jsonDocument(record))); err != nil {
		return "", fmt.Errorf("record stale lifecycle lease evidence: %w", err)
	}
	if err := os.Remove(quarantinedLease); err != nil {
		return "", fmt.Errorf("remove evidenced stale lease quarantine: %w", err)
	}
	if err := os.Remove(quarantine); err != nil {
		return "", fmt.Errorf("remove stale lease quarantine directory: %w", err)
	}
	return "git:starter-kit-attempts/" + filename, nil
}

type stageTransactionMarker struct {
	SchemaVersion int       `json:"schema_version"`
	Ownership     string    `json:"ownership"`
	Source        string    `json:"source"`
	LeaseToken    string    `json:"lease_token"`
	PlanID        string    `json:"plan_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func recoverAbandonedStages(root, lockPath, planID string, leaseEvidence []string, now time.Time) ([]string, []string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}
	recovery := []string{}
	evidence := []string{}
	references := append([]string{}, leaseEvidence...)
	attemptDirectory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
	if attemptEntries, readErr := os.ReadDir(attemptDirectory); readErr == nil {
		for _, entry := range attemptEntries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "stale-lock-") && strings.HasSuffix(entry.Name(), ".json") {
				references = append(references, "git:starter-kit-attempts/"+entry.Name())
			}
		}
	}
	staleTokens := map[string]string{}
	for _, reference := range references {
		record, valid := readStaleLeaseEvidence(filepath.Dir(lockPath), reference)
		if valid && record.Lease.PlanID == planID {
			staleTokens[record.Lease.Token] = reference
		}
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".starter-kit-stage-") {
			continue
		}
		stagePath := filepath.Join(root, entry.Name())
		token := strings.TrimPrefix(entry.Name(), ".starter-kit-stage-")
		staleReference, authorized := staleTokens[token]
		if !authorized {
			return recovery, evidence, errors.New("unrecognized staging tree is preserved and requires reconciliation")
		}
		markerContent, err := os.ReadFile(filepath.Join(stagePath, ".starter-kit-transaction.json"))
		if err != nil {
			return recovery, evidence, errors.New("staging tree lacks a readable engine transaction marker")
		}
		var marker stageTransactionMarker
		if json.Unmarshal(markerContent, &marker) != nil || marker.SchemaVersion != 1 || marker.Ownership != "machine-state" || marker.Source != "engine:apply:v1" || marker.LeaseToken != token || marker.PlanID != planID || marker.CreatedAt.IsZero() {
			return recovery, evidence, errors.New("staging tree transaction marker does not match the stale lease and plan")
		}
		stageDigest, err := pathTreeDigest(stagePath)
		if err != nil {
			return recovery, evidence, fmt.Errorf("digest abandoned stage: %w", err)
		}
		record := abandonedStageEvidence{
			SchemaVersion: 1, Ownership: "machine-evidence", Source: "engine:apply:v1",
			PlanID: planID, LeaseToken: token, StageDigest: stageDigest, RecoveredAt: now,
		}
		record.EvidenceDigest = digestJSON(record)
		nameDigest := strings.TrimPrefix(record.EvidenceDigest, "sha256:")
		filename := "abandoned-stage-" + nameDigest + ".json"
		directory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return recovery, evidence, err
		}
		if err := writeEvidenceOnce(filepath.Join(directory, filename), []byte(jsonDocument(record))); err != nil {
			return recovery, evidence, err
		}
		if err := os.RemoveAll(stagePath); err != nil {
			return recovery, evidence, fmt.Errorf("remove abandoned stage: %w", err)
		}
		recovery = append(recovery, "removed an abandoned staging tree after preserving its content digest")
		if !containsString(leaseEvidence, staleReference) && !containsString(evidence, staleReference) {
			evidence = append(evidence, staleReference)
		}
		evidence = append(evidence, "git:starter-kit-attempts/"+filename)
	}
	return recovery, evidence, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func readStaleLeaseEvidence(gitDirectory, reference string) (staleLeaseEvidence, bool) {
	const prefix = "git:starter-kit-attempts/stale-lock-"
	if !strings.HasPrefix(reference, prefix) || !strings.HasSuffix(reference, ".json") {
		return staleLeaseEvidence{}, false
	}
	filename := strings.TrimPrefix(reference, "git:starter-kit-attempts/")
	content, err := os.ReadFile(filepath.Join(gitDirectory, "starter-kit-attempts", filename))
	if err != nil {
		return staleLeaseEvidence{}, false
	}
	var record staleLeaseEvidence
	if json.Unmarshal(content, &record) != nil {
		return staleLeaseEvidence{}, false
	}
	recordedDigest := record.EvidenceDigest
	record.EvidenceDigest = ""
	digest := digestJSON(record)
	expectedFilename := "stale-lock-" + strings.TrimPrefix(digest, "sha256:") + ".json"
	valid := record.SchemaVersion == 1 && record.Ownership == "machine-evidence" && record.Source == "engine:apply:v1" &&
		record.Lease.SchemaVersion == 1 && validLeaseToken(record.Lease.Token) && validSHA256Digest(record.Lease.PlanID) &&
		record.Lease.PID > 0 && !record.Lease.CreatedAt.IsZero() && validSHA256Digest(record.LeaseDigest) && !record.RecoveredAt.IsZero() &&
		recordedDigest == digest && filename == expectedFilename
	record.EvidenceDigest = recordedDigest
	return record, valid
}

func readAbandonedStageEvidence(gitDirectory, reference string) (abandonedStageEvidence, bool) {
	const prefix = "git:starter-kit-attempts/abandoned-stage-"
	if !strings.HasPrefix(reference, prefix) || !strings.HasSuffix(reference, ".json") {
		return abandonedStageEvidence{}, false
	}
	filename := strings.TrimPrefix(reference, "git:starter-kit-attempts/")
	content, err := os.ReadFile(filepath.Join(gitDirectory, "starter-kit-attempts", filename))
	if err != nil {
		return abandonedStageEvidence{}, false
	}
	var record abandonedStageEvidence
	if json.Unmarshal(content, &record) != nil {
		return abandonedStageEvidence{}, false
	}
	recordedDigest := record.EvidenceDigest
	record.EvidenceDigest = ""
	digest := digestJSON(record)
	expectedFilename := "abandoned-stage-" + strings.TrimPrefix(digest, "sha256:") + ".json"
	valid := record.SchemaVersion == 1 && record.Ownership == "machine-evidence" && record.Source == "engine:apply:v1" &&
		record.PlanID != "" && validLeaseToken(record.LeaseToken) && validSHA256Digest(record.StageDigest) &&
		!record.RecoveredAt.IsZero() && recordedDigest == digest && filename == expectedFilename
	record.EvidenceDigest = recordedDigest
	return record, valid
}

func validSuccessfulRecoveryMetadata(root string, plan Plan, recovery, evidence []string) bool {
	allowedRecovery := map[string]bool{
		"recovered a stale lifecycle lease after confirming its recorded process was not active": true,
		"removed an abandoned staging tree after preserving its content digest":                  true,
		"resumed the same immutable create plan without replacing its matching committed prefix": true,
	}
	for _, action := range recovery {
		if !allowedRecovery[action] {
			return false
		}
	}
	gitDirectory := filepath.Join(root, ".git")
	for _, reference := range evidence {
		if stale, valid := readStaleLeaseEvidence(gitDirectory, reference); valid {
			if stale.Lease.PlanID != plan.ID {
				return false
			}
			continue
		}
		if stage, valid := readAbandonedStageEvidence(gitDirectory, reference); valid {
			if stage.PlanID != plan.ID {
				return false
			}
			continue
		}
		return false
	}
	return true
}

func pathTreeDigest(root string) (string, error) {
	type entryDigest struct {
		Path   string `json:"path"`
		Kind   string `json:"kind"`
		Digest string `json:"digest"`
	}
	entries := []entryDigest{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			entries = append(entries, entryDigest{Path: filepath.ToSlash(relative), Kind: "directory"})
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			entries = append(entries, entryDigest{Path: filepath.ToSlash(relative), Kind: "symlink", Digest: digestBytes([]byte(target))})
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entries = append(entries, entryDigest{Path: filepath.ToSlash(relative), Kind: "file", Digest: digestBytes(content)})
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return digestJSON(entries), nil
}

func writeEvidenceOnce(target string, content []byte) error {
	if existing, err := os.ReadFile(target); err == nil {
		if string(existing) == string(content) {
			return nil
		}
		return errors.New("recovery evidence path already contains different content")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			existing, readErr := os.ReadFile(target)
			if readErr == nil && string(existing) == string(content) {
				return nil
			}
			return errors.New("evidence path already contains different content")
		}
		return err
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
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
			return fmt.Errorf("invalid planned path: %w", err)
		}
		pathKey := strings.ToLower(file.Path)
		if _, duplicate := seen[pathKey]; duplicate {
			return errors.New("planned paths collide under case-insensitive filesystem semantics")
		}
		seen[pathKey] = struct{}{}
		if file.Ownership == "" || file.Source == "" || file.Digest == "" {
			return errors.New("planned file has incomplete provenance")
		}
		if digestBytes([]byte(file.Content)) != file.Digest {
			return errors.New("planned content digest mismatch")
		}
		if containsSensitiveText(file.Content) {
			return errors.New("planned content contains sensitive-looking material")
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

func ensureNoSymlinkComponents(root, slashPath string) error {
	if err := ensureNoSymlinkParents(root, slashPath); err != nil {
		return err
	}
	info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(slashPath)))
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("managed path is a symlink")
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

func failAndRecord(root string, plan Plan, stage string, changed []string, recoverable bool, recovery, evidence []string, err error) (ApplyResult, error) {
	retained := append([]string{}, changed...)
	retained = append(retained, plan.Result.Path)
	result := ApplyResult{
		SchemaVersion: 1, PlanID: plan.ID, Status: ApplyStatusFailed,
		ChangedFiles: retained, Recovery: append([]string{}, recovery...), Evidence: append([]string{}, evidence...),
	}
	failure := &ApplyFailure{
		Stage: stage, Recoverable: recoverable, Cause: err.Error(),
		ChangedFiles: append([]string{}, retained...), Conflicts: []ReconciliationConflict{},
		Recovery: append([]string{}, recovery...), Evidence: append([]string{}, evidence...),
	}
	if eventErr := recordApplyEvent(root, plan, result, failure); eventErr != nil {
		result.ChangedFiles = append([]string{}, changed...)
		failure.ChangedFiles = append([]string{}, changed...)
		failure.Recoverable = false
		failure.Cause = fmt.Sprintf("%s; recording failure event failed: %v", failure.Cause, eventErr)
	}
	return result, failure
}

func failedCommittedApply(root string, plan Plan, stage string, changed, recovery, evidence []string, err error) (ApplyResult, error) {
	if rollbackErr := rollbackFiles(root, changed); rollbackErr != nil {
		return failAndRecord(root, plan, stage, changed, false, recovery, evidence, fmt.Errorf("%v; rollback failed: %v", err, rollbackErr))
	}
	return failAndRecord(root, plan, stage, []string{}, true, recovery, evidence, err)
}

// Status reports lifecycle state from the authoritative local state document.
func (e *Engine) Status(ctx context.Context, repository string) (RepositoryStatus, error) {
	inspection, err := e.Inspect(ctx, repository)
	if err != nil {
		return RepositoryStatus{}, err
	}
	if inspection.Managed {
		return RepositoryStatus{
			SchemaVersion: 1, Repository: inspection.Repository, Lifecycle: LifecycleManaged,
			Problems: []string{}, Recovery: []string{}, Evidence: []string{},
		}, nil
	}
	if incompleteCreateState(inspection.Repository) {
		problems := append([]string{}, inspection.Problems...)
		problems = append(problems, "create transaction is incomplete and does not establish managed conformance")
		sort.Strings(problems)
		return RepositoryStatus{
			SchemaVersion: 1, Repository: inspection.Repository, Lifecycle: LifecycleSetupIncomplete,
			Problems: problems,
			Recovery: []string{
				"retry the same immutable create plan so matching committed artifacts are preserved and missing artifacts are completed",
				"if the original plan is unavailable, preserve the repository and use an explicitly authorized reconciliation workflow",
			},
			Evidence: recoveryEvidenceReferences(inspection.Repository),
		}, nil
	}
	if inspection.ContractPresent {
		return RepositoryStatus{
			SchemaVersion: 1, Repository: inspection.Repository, Lifecycle: LifecycleManagedDegraded,
			Problems: inspection.Problems, Recovery: []string{}, Evidence: recoveryEvidenceReferences(inspection.Repository),
		}, nil
	}
	return RepositoryStatus{
		SchemaVersion: 1, Repository: inspection.Repository, Lifecycle: LifecycleUnmanaged,
		Problems: []string{}, Recovery: []string{}, Evidence: recoveryEvidenceReferences(inspection.Repository),
	}, nil
}

func incompleteCreateState(root string) bool {
	if fileExists(filepath.Join(root, ".starter-kit", "state.json")) {
		return false
	}
	for _, marker := range []string{
		".starter-kit/managed-files.json", ".starter-kit/layout.json", ".starter-kit/project.json",
	} {
		if fileExists(filepath.Join(root, filepath.FromSlash(marker))) {
			return true
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".starter-kit-stage-") {
			return true
		}
	}
	return false
}

func recoveryEvidenceReferences(root string) []string {
	directory := filepath.Join(root, ".git", "starter-kit-attempts")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return []string{}
	}
	references := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && validRecoveryEvidenceFilename(entry.Name()) {
			references = append(references, "git:starter-kit-attempts/"+entry.Name())
		}
	}
	sort.Strings(references)
	return references
}

func validRecoveryEvidenceFilename(name string) bool {
	stem := strings.TrimSuffix(name, ".json")
	if stem == name {
		return false
	}
	for _, prefix := range []string{"stale-lock-", "abandoned-stage-"} {
		if strings.HasPrefix(stem, prefix) {
			stem = strings.TrimPrefix(stem, prefix)
			break
		}
	}
	decoded, err := hex.DecodeString(stem)
	return err == nil && (len(decoded) == 16 || len(decoded) == 32)
}

func planWithBlankID(plan Plan) Plan {
	plan.ID = ""
	return plan
}
