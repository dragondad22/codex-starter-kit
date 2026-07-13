package engine_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestInspectEmptyGitRepository(t *testing.T) {
	repository := newGitRepository(t)

	result, err := engine.New().Inspect(context.Background(), repository)
	if err != nil {
		t.Fatalf("inspect empty Git repository: %v", err)
	}

	if result.Repository != filepath.Clean(repository) {
		t.Fatalf("repository = %q, want %q", result.Repository, filepath.Clean(repository))
	}
	if !result.Git {
		t.Fatal("expected repository to be detected as Git")
	}
	if result.Managed {
		t.Fatal("empty repository must not be reported as managed")
	}
	if result.UserFileCount != 0 {
		t.Fatalf("user file count = %d, want 0", result.UserFileCount)
	}
}

func TestInspectSnapshotChangesWhenContentChangesWithoutCountChange(t *testing.T) {
	repository := newGitRepository(t)
	path := filepath.Join(repository, "README.md")
	if err := os.WriteFile(path, []byte("first\n"), 0o644); err != nil {
		t.Fatalf("write first content: %v", err)
	}
	first, err := engine.New().Inspect(t.Context(), repository)
	if err != nil {
		t.Fatalf("inspect first content: %v", err)
	}
	if err := os.WriteFile(path, []byte("other\n"), 0o644); err != nil {
		t.Fatalf("write replacement content: %v", err)
	}
	second, err := engine.New().Inspect(t.Context(), repository)
	if err != nil {
		t.Fatalf("inspect replacement content: %v", err)
	}
	if first.UserFileCount != second.UserFileCount {
		t.Fatal("fixture must preserve the file count")
	}
	if first.PreconditionDigest == second.PreconditionDigest {
		t.Fatal("content replacement must change the precondition digest")
	}
}

func TestApplyCreatePlanProducesManagedRepository(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}

	result, err := lifecycle.Apply(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	if result.PlanID != plan.ID || result.Status != engine.ApplyStatusApplied {
		t.Fatalf("unexpected apply result: %#v", result)
	}
	for _, planned := range plan.Files {
		content, readErr := os.ReadFile(filepath.Join(repository, filepath.FromSlash(planned.Path)))
		if readErr != nil {
			t.Fatalf("read applied file %s: %v", planned.Path, readErr)
		}
		if string(content) != planned.Content {
			t.Fatalf("applied content differs for %s", planned.Path)
		}
	}
	manifestContent, err := os.ReadFile(filepath.Join(repository, ".starter-kit", "managed-files.json"))
	if err != nil {
		t.Fatalf("read managed-file manifest: %v", err)
	}
	var manifest struct {
		Self struct {
			Path      string `json:"path"`
			Ownership string `json:"ownership"`
			Source    string `json:"source"`
		} `json:"self"`
	}
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		t.Fatalf("decode managed-file manifest: %v", err)
	}
	if manifest.Self.Path != ".starter-kit/managed-files.json" || manifest.Self.Ownership != "managed" || manifest.Self.Source == "" {
		t.Fatalf("manifest does not classify itself: %#v", manifest.Self)
	}
	eventContent, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(plan.ResultPath)))
	if err != nil {
		t.Fatalf("read apply event: %v", err)
	}
	var event struct {
		PlanID string `json:"plan_id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(eventContent, &event); err != nil {
		t.Fatalf("decode apply event: %v", err)
	}
	if event.PlanID != plan.ID || event.Status != string(engine.ApplyStatusApplied) {
		t.Fatalf("unexpected apply event: %#v", event)
	}

	status, err := lifecycle.Status(context.Background(), repository)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManaged || status.SchemaVersion != 1 {
		t.Fatalf("unexpected managed status: %#v", status)
	}
}

func TestCreateReturnsStableReviewablePlan(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()

	first, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	second, err := lifecycle.Plan(context.Background(), engine.PlanRequest{
		Operation: engine.CreateOperation,
		Create:    approvedCreate(repository),
	})
	if err != nil {
		t.Fatalf("plan create: %v", err)
	}

	if first.ID == "" {
		t.Fatal("plan ID must not be empty")
	}
	if first.ID != second.ID {
		t.Fatalf("unchanged plan IDs differ: %q != %q", first.ID, second.ID)
	}
	if first.Operation != engine.CreateOperation {
		t.Fatalf("operation = %q, want %q", first.Operation, engine.CreateOperation)
	}
	wantPaths := []string{
		".starter-kit/layout.json",
		".starter-kit/managed-files.json",
		".starter-kit/policy-lock.json",
		".starter-kit/project.json",
		".starter-kit/routes.json",
		".starter-kit/state.json",
		"AGENTS.md",
		"docs/decisions/INDEX.md",
		"docs/evidence/CONFORMANCE.md",
		"docs/product/BRIEF.md",
		"docs/product/PERSONAS.md",
	}
	gotPaths := make([]string, 0, len(first.Files))
	for _, file := range first.Files {
		gotPaths = append(gotPaths, file.Path)
		if file.Ownership == "" || file.Source == "" || file.Digest == "" {
			t.Fatalf("planned file lacks ownership, provenance, or digest: %#v", file)
		}
	}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("planned paths = %#v, want %#v", gotPaths, wantPaths)
	}
}

func TestCreateRequiresExplicitHumanOwnedApprovals(t *testing.T) {
	repository := newGitRepository(t)
	_, err := engine.New().Create(t.Context(), engine.CreateRequest{Repository: repository})
	if err == nil {
		t.Fatal("create must require an approved brief and confirmed owner persona")
	}
}

func TestCreateAfterApplyReturnsExplicitNoChange(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(context.Background(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}

	unchanged, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create unchanged plan: %v", err)
	}
	if !unchanged.NoChange || len(unchanged.Files) != 0 {
		t.Fatalf("expected no-change plan, got %#v", unchanged)
	}
	result, err := lifecycle.Apply(context.Background(), unchanged.ID, unchanged)
	if err != nil {
		t.Fatalf("apply no-change plan: %v", err)
	}
	if result.Status != engine.ApplyStatusNoChange || len(result.ChangedFiles) != 0 {
		t.Fatalf("unexpected no-change result: %#v", result)
	}
}

func TestCreateDoesNotReportNoChangeForDifferentApprovedBrief(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	request := approvedCreate(repository)
	plan, err := lifecycle.Create(t.Context(), request)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	request.Brief = "A different approved project outcome."

	if _, err := lifecycle.Create(t.Context(), request); err == nil {
		t.Fatal("different create inputs must require reconciliation, not no-change")
	}
}

func TestDriftedManagedRepositoryIsDegradedInsteadOfNoChange(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	if err := os.Remove(filepath.Join(repository, "AGENTS.md")); err != nil {
		t.Fatalf("remove managed file: %v", err)
	}

	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status drifted repository: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManagedDegraded || len(status.Problems) == 0 {
		t.Fatalf("unexpected drift status: %#v", status)
	}
	if _, err := lifecycle.Create(t.Context(), approvedCreate(repository)); err == nil {
		t.Fatal("drifted managed repository must not produce no-change")
	}
}

func TestManifestCannotHideADeletedRequiredArtifact(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	manifestPath := filepath.Join(repository, ".starter-kit", "managed-files.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	entries := manifest["files"].([]interface{})
	filtered := make([]interface{}, 0, len(entries)-1)
	for _, raw := range entries {
		entry := raw.(map[string]interface{})
		if entry["path"] != "AGENTS.md" {
			filtered = append(filtered, entry)
		}
	}
	manifest["files"] = filtered
	content, err = json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, content, 0o644); err != nil {
		t.Fatalf("write altered manifest: %v", err)
	}
	if err := os.Remove(filepath.Join(repository, "AGENTS.md")); err != nil {
		t.Fatalf("remove required artifact: %v", err)
	}

	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status altered contract: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManagedDegraded {
		t.Fatalf("altered manifest hid required artifact: %#v", status)
	}
}

func TestApplyRejectsRepositoryChangedAfterPlanning(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatalf("change repository: %v", err)
	}

	if _, err := lifecycle.Apply(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("apply must reject changed repository preconditions")
	}
	if _, err := os.Stat(filepath.Join(repository, ".starter-kit", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("state file exists after rejected apply: %v", err)
	}
}

func TestApplyAcquiresLifecycleLockBeforePreconditionRecheck(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".starter-kit.lock"), []byte("held\n"), 0o600); err != nil {
		t.Fatalf("create held lock: %v", err)
	}

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "lock" || !failure.Recoverable {
		t.Fatalf("unexpected lock failure: %#v, %v", failure, err)
	}
	if result.Status != engine.ApplyStatusFailed {
		t.Fatalf("lock failure result = %#v", result)
	}
}

func TestApplyRejectsPlanWhoseContentDoesNotMatchIdentifier(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	plan.Files[0].Content = "tampered\n"

	if _, err := lifecycle.Apply(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("apply must reject plan content that does not match its identifier")
	}
	if _, err := os.Stat(filepath.Join(repository, ".starter-kit", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("state file exists after rejected plan: %v", err)
	}
}

func TestApplyRejectsSelfConsistentPlanPathOutsideRepository(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	plan.Files[0].Path = "../escape.txt"
	plan.ID = identifyPlan(t, plan)

	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err == nil {
		t.Fatal("apply must reject a self-consistent plan that escapes the repository")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(repository), "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("escape path exists after rejected plan: %v", err)
	}
}

func TestApplyRollsBackWhenManagedContractPostconditionFails(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	for index := range plan.Files {
		if plan.Files[index].Path == ".starter-kit/managed-files.json" {
			plan.Files[index].Content = "{\"schema_version\":1}\n"
			digest := sha256.Sum256([]byte(plan.Files[index].Content))
			plan.Files[index].Digest = "sha256:" + hex.EncodeToString(digest[:])
		}
	}
	plan.ID = identifyPlan(t, plan)

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	if err == nil {
		t.Fatal("invalid managed contract must fail postcondition verification")
	}
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || !failure.Recoverable || failure.Stage != "postcondition" {
		t.Fatalf("unexpected apply failure: %#v, %v", failure, err)
	}
	if result.Status != engine.ApplyStatusFailed || len(result.ChangedFiles) != 0 {
		t.Fatalf("rollback result must report no retained changes: %#v", result)
	}
	for _, path := range []string{"AGENTS.md", ".starter-kit/state.json"} {
		if _, statErr := os.Stat(filepath.Join(repository, filepath.FromSlash(path))); !os.IsNotExist(statErr) {
			t.Fatalf("%s exists after rollback: %v", path, statErr)
		}
	}
}

func newGitRepository(t *testing.T) string {
	t.Helper()
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	return repository
}

func approvedCreate(repository string) engine.CreateRequest {
	return engine.CreateRequest{
		Repository:            repository,
		Brief:                 "Create a minimal managed repository for lifecycle-engine development.",
		BriefApproved:         true,
		OwnerPersonaConfirmed: true,
	}
}

func identifyPlan(t *testing.T, plan engine.Plan) string {
	t.Helper()
	plan.ID = ""
	content, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("encode plan identity fixture: %v", err)
	}
	digest := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(digest[:])
}
