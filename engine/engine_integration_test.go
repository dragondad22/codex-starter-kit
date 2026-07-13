package engine_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestInspectEmptyGitRepository(t *testing.T) {
	repository := newGitRepository(t)

	result, err := engine.New().Inspect(context.Background(), repository)
	if err != nil {
		t.Fatalf("inspect empty Git repository: %v", err)
	}

	canonicalRepository, err := filepath.EvalSymlinks(repository)
	if err != nil {
		t.Fatalf("canonicalize repository fixture: %v", err)
	}
	if result.Repository != filepath.Clean(canonicalRepository) {
		t.Fatalf("repository = %q, want canonical %q", result.Repository, filepath.Clean(canonicalRepository))
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

func TestLifecycleGitExecutionTreatsRepositoryMetacharactersAsOneArgument(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "injected.txt")
	repository := filepath.Join(root, "repo;touch injected.txt")
	if err := os.Mkdir(repository, 0o755); err != nil {
		t.Fatalf("create repository path: %v", err)
	}
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}

	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan for metacharacter path: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan for metacharacter path: %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("repository path was interpreted as a shell command: %v", err)
	}
}

func TestLifecycleGitExecutionIgnoresHostileGitEnvironmentOverrides(t *testing.T) {
	repository := newGitRepository(t)
	outside := newGitRepository(t)
	if err := os.WriteFile(filepath.Join(outside, ".git", "starter-kit.lock"), []byte("outside lock\n"), 0o600); err != nil {
		t.Fatalf("create hostile external lock: %v", err)
	}
	t.Setenv("GIT_DIR", filepath.Join(outside, ".git"))
	t.Setenv("GIT_WORK_TREE", outside)

	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan with hostile Git environment: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("hostile Git environment redirected lifecycle execution: %v", err)
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
	eventContent, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(plan.Result.Path)))
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

func TestCreateRejectsFixtureSecretWithoutEchoingItIntoDiagnostics(t *testing.T) {
	repository := newGitRepository(t)
	secret := "ghp_12345678901234567890"
	request := approvedCreate(repository)
	request.Brief = "Create a repository with token " + secret

	_, err := engine.New().Create(t.Context(), request)
	if err == nil {
		t.Fatal("create accepted fixture secret into a reviewable plan")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("create diagnostic exposed fixture secret: %v", err)
	}
}

func TestCreateRejectsSecretBearingRepositoryPathWithoutEchoingIt(t *testing.T) {
	secret := "ghp_12345678901234567890"
	repository := filepath.Join(t.TempDir(), "repository-"+secret)
	if err := os.Mkdir(repository, 0o755); err != nil {
		t.Fatalf("create secret-bearing repository path: %v", err)
	}
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}

	if _, err := engine.New().Create(t.Context(), approvedCreate(repository)); err == nil {
		t.Fatal("create accepted a secret-bearing repository path into a plan")
	} else if strings.Contains(err.Error(), secret) {
		t.Fatalf("repository-path diagnostic exposed fixture secret: %v", err)
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
	if result.Status != engine.ApplyStatusNoChange || len(result.ChangedFiles) != 1 || result.ChangedFiles[0] != unchanged.Result.Path {
		t.Fatalf("unexpected no-change result: %#v", result)
	}
	replayed, err := lifecycle.Apply(context.Background(), unchanged.ID, unchanged)
	if err != nil {
		t.Fatalf("replay identical no-change plan: %v", err)
	}
	if !reflect.DeepEqual(replayed, result) {
		t.Fatalf("replayed no-change result changed semantics:\nfirst:  %#v\nsecond: %#v", result, replayed)
	}
}

func TestIdenticalAppliedCreatePlanReturnsStableIdempotentResult(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	first, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	beforeReplay, err := lifecycle.Inspect(t.Context(), repository)
	if err != nil {
		t.Fatalf("inspect before replay: %v", err)
	}

	second, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	if err != nil {
		t.Fatalf("replay identical applied plan: %v", err)
	}
	if !reflect.DeepEqual(second, first) {
		t.Fatalf("replayed result changed semantics:\nfirst:  %#v\nsecond: %#v", first, second)
	}
	afterReplay, err := lifecycle.Inspect(t.Context(), repository)
	if err != nil {
		t.Fatalf("inspect after replay: %v", err)
	}
	if afterReplay.PreconditionDigest != beforeReplay.PreconditionDigest {
		t.Fatalf("idempotent replay mutated repository: before %s, after %s", beforeReplay.PreconditionDigest, afterReplay.PreconditionDigest)
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

	_, err = lifecycle.Create(t.Context(), request)
	if err == nil {
		t.Fatal("different create inputs must require reconciliation, not no-change")
	}
	var reconciliation *engine.ReconciliationRequired
	if !errors.As(err, &reconciliation) || len(reconciliation.Conflicts) == 0 {
		t.Fatalf("different approved inputs lack reviewable reconciliation: %#v, %v", reconciliation, err)
	}
	foundBrief := false
	for _, conflict := range reconciliation.Conflicts {
		foundBrief = foundBrief || conflict.Path == "docs/product/BRIEF.md"
	}
	if !foundBrief {
		t.Fatalf("reconciliation did not identify the changed human-owned brief: %#v", reconciliation.Conflicts)
	}
}

func TestCreateReturnsReviewableReconciliationForExistingUserContent(t *testing.T) {
	repository := newGitRepository(t)
	path := filepath.Join(repository, "README.md")
	original := []byte("human work\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("write human-owned conflict: %v", err)
	}

	_, err := engine.New().Create(t.Context(), approvedCreate(repository))
	var reconciliation *engine.ReconciliationRequired
	if !errors.As(err, &reconciliation) {
		t.Fatalf("create did not return reviewable reconciliation: %v", err)
	}
	if len(reconciliation.Conflicts) != 1 || reconciliation.Conflicts[0].Path != "README.md" || reconciliation.Conflicts[0].Ownership != "user-owned" {
		t.Fatalf("unexpected reconciliation conflicts: %#v", reconciliation.Conflicts)
	}
	if len(reconciliation.Recovery) == 0 {
		t.Fatalf("reconciliation omitted safe next actions: %#v", reconciliation)
	}
	content, readErr := os.ReadFile(path)
	if readErr != nil || !reflect.DeepEqual(content, original) {
		t.Fatalf("reconciliation did not preserve human work: %q, %v", content, readErr)
	}
}

func TestReconciliationRedactsSecretBearingConflictPaths(t *testing.T) {
	repository := newGitRepository(t)
	secret := "ghp_12345678901234567890"
	if err := os.WriteFile(filepath.Join(repository, "conflict-"+secret), []byte("human work\n"), 0o644); err != nil {
		t.Fatalf("write secret-bearing conflict: %v", err)
	}

	_, err := engine.New().Create(t.Context(), approvedCreate(repository))
	var reconciliation *engine.ReconciliationRequired
	if !errors.As(err, &reconciliation) {
		t.Fatalf("create did not return reconciliation: %v", err)
	}
	document, marshalErr := json.Marshal(reconciliation)
	if marshalErr != nil {
		t.Fatalf("encode reconciliation: %v", marshalErr)
	}
	if strings.Contains(string(document), secret) || !strings.Contains(string(document), "[REDACTED]-sha256:") {
		t.Fatalf("reconciliation exposed secret-bearing conflict path: %s", document)
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

func TestStatusFailsClosedForSelfConsistentAdversarialManagedState(t *testing.T) {
	tests := map[string]func(*testing.T, string){
		"layout escape": func(t *testing.T, repository string) {
			rewriteManagedFile(t, repository, ".starter-kit/layout.json", `{"schema_version":1,"roles":{"decisions":"../outside","evidence":"docs/evidence","product":"docs/product"}}`+"\n")
		},
		"missing routes": func(t *testing.T, repository string) {
			rewriteManagedFile(t, repository, ".starter-kit/routes.json", `{"schema_version":1,"routes":{}}`+"\n")
		},
		"unsupported engine state": func(t *testing.T, repository string) {
			rewriteManagedFile(t, repository, ".starter-kit/state.json", `{"schema_version":1,"lifecycle":"managed","engine_version":"untrusted"}`+"\n")
		},
		"forged ownership": func(t *testing.T, repository string) {
			manifestPath := filepath.Join(repository, ".starter-kit", "managed-files.json")
			content, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			var manifest map[string]interface{}
			if err := json.Unmarshal(content, &manifest); err != nil {
				t.Fatalf("decode manifest: %v", err)
			}
			for _, raw := range manifest["files"].([]interface{}) {
				entry := raw.(map[string]interface{})
				if entry["path"] == "AGENTS.md" {
					entry["ownership"] = "human-owned"
				}
			}
			encoded, err := json.Marshal(manifest)
			if err != nil {
				t.Fatalf("encode manifest: %v", err)
			}
			if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o644); err != nil {
				t.Fatalf("write manifest: %v", err)
			}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repository := newGitRepository(t)
			lifecycle := engine.New()
			plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
			if err != nil {
				t.Fatalf("create plan: %v", err)
			}
			if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
				t.Fatalf("apply plan: %v", err)
			}
			mutate(t, repository)

			status, err := lifecycle.Status(t.Context(), repository)
			if err != nil {
				t.Fatalf("status adversarial repository: %v", err)
			}
			if status.Lifecycle != engine.LifecycleManagedDegraded || len(status.Problems) == 0 {
				t.Fatalf("adversarial managed state did not fail closed: %#v", status)
			}
		})
	}
}

func TestStatusRedactsFixtureSecretFromAdversarialOwnershipData(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	secret := "ghp_12345678901234567890"
	manifestPath := filepath.Join(repository, ".starter-kit", "managed-files.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	manifest["files"].([]interface{})[0].(map[string]interface{})["path"] = "../" + secret
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, encoded, 0o644); err != nil {
		t.Fatalf("write adversarial manifest: %v", err)
	}

	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status adversarial repository: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManagedDegraded {
		t.Fatalf("adversarial manifest did not degrade status: %#v", status)
	}
	if strings.Contains(strings.Join(status.Problems, " "), secret) {
		t.Fatalf("status exposed fixture secret: %#v", status.Problems)
	}
}

func TestStatusFailsClosedForSelfConsistentAdversarialProvenance(t *testing.T) {
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
	manifest["files"].([]interface{})[0].(map[string]interface{})["source"] = "attacker:forged:v1"
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o644); err != nil {
		t.Fatalf("write adversarial manifest: %v", err)
	}

	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status adversarial repository: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManagedDegraded || len(status.Problems) == 0 {
		t.Fatalf("forged persisted provenance did not fail closed: %#v", status)
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

func TestApplyPersistsReviewableReconciliationWithoutReplacingNewHumanWork(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	path := filepath.Join(repository, "README.md")
	original := []byte("new human work\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("write post-plan human work: %v", err)
	}

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "reconcile" || !failure.Recoverable {
		t.Fatalf("apply did not return reviewable reconciliation: %#v, %v", failure, err)
	}
	if len(failure.Conflicts) != 1 || failure.Conflicts[0].Path != "README.md" || len(failure.Recovery) == 0 {
		t.Fatalf("reconciliation omitted conflict facts or recovery: %#v", failure)
	}
	if len(result.ChangedFiles) != 1 || result.ChangedFiles[0] != plan.Result.Path {
		t.Fatalf("reconciliation result did not report its evidence effect: %#v", result)
	}
	content, readErr := os.ReadFile(path)
	if readErr != nil || !reflect.DeepEqual(content, original) {
		t.Fatalf("reconciliation replaced human work: %q, %v", content, readErr)
	}
	event, readErr := os.ReadFile(filepath.Join(repository, filepath.FromSlash(plan.Result.Path)))
	if readErr != nil || !strings.Contains(string(event), `"path": "README.md"`) {
		t.Fatalf("reconciliation evidence omitted conflict: %s, %v", event, readErr)
	}
}

func TestFailureEvidenceDoesNotPreventARecoveredCreateRetry(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	fixturePath := filepath.Join(repository, "temporary.txt")
	if err := os.WriteFile(fixturePath, []byte("change\n"), 0o644); err != nil {
		t.Fatalf("change repository: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err == nil {
		t.Fatal("changed repository must fail apply")
	}
	if err := os.Remove(fixturePath); err != nil {
		t.Fatalf("restore repository: %v", err)
	}

	retry, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("plan recovered create: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), retry.ID, retry); err != nil {
		t.Fatalf("apply recovered create: %v", err)
	}
}

func TestApplyRejectsIndexOnlyGitChangeWithSameFilesystemSnapshot(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	path := filepath.Join(repository, "README.md")
	if err := os.WriteFile(path, []byte("staged only\n"), 0o644); err != nil {
		t.Fatalf("write index fixture: %v", err)
	}
	command := exec.Command("git", "-C", repository, "add", "README.md")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("stage index fixture: %v: %s", err, output)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("restore empty worktree: %v", err)
	}

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "precondition" {
		t.Fatalf("unexpected Git precondition failure: %#v, %v", failure, err)
	}
	if result.Status != engine.ApplyStatusFailed {
		t.Fatalf("Git precondition result = %#v", result)
	}
}

func TestApplyAcquiresLifecycleLockBeforePreconditionRecheck(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".git", "starter-kit.lock"), []byte("held\n"), 0o600); err != nil {
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
	attemptPath := filepath.Join(repository, ".git", "starter-kit-attempts", strings.TrimPrefix(plan.ID, "sha256:")+".json")
	if _, err := os.Stat(attemptPath); err != nil {
		t.Fatalf("lock failure attempt evidence missing: %v", err)
	}
}

func TestApplyRecoversDeadStaleLifecycleLeaseWithEvidence(t *testing.T) {
	repository := newGitRepository(t)
	now := time.Date(2026, 7, 13, 17, 0, 0, 0, time.UTC)
	lifecycle := engine.New(engine.WithClock(fixedClock{now}))
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	token := "0123456789abcdef0123456789abcdef"
	lease := fmt.Sprintf("{\"schema_version\":1,\"token\":%q,\"plan_id\":%q,\"pid\":2147483647,\"created_at\":%q}\n", token, plan.ID, now.Add(-time.Hour).Format(time.RFC3339Nano))
	lockPath := filepath.Join(repository, ".git", "starter-kit.lock")
	if err := os.WriteFile(lockPath, []byte(lease), 0o600); err != nil {
		t.Fatalf("write stale lifecycle lease: %v", err)
	}

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	if err != nil {
		t.Fatalf("recover stale lease and apply: %v", err)
	}
	if result.Status != engine.ApplyStatusApplied || len(result.Recovery) == 0 || len(result.Evidence) == 0 {
		t.Fatalf("stale lease recovery was not disclosed: %#v", result)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("recovered lease remained at lock path: %v", err)
	}
	archive := filepath.Join(repository, ".git", "starter-kit-attempts", "stale-lock-"+token+".json")
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("stale lease evidence missing: %v", err)
	}
}

func TestApplyDoesNotStealLiveLifecycleLease(t *testing.T) {
	repository := newGitRepository(t)
	now := time.Date(2026, 7, 13, 17, 30, 0, 0, time.UTC)
	lifecycle := engine.New(engine.WithClock(fixedClock{now}))
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	token := "11111111111111111111111111111111"
	lease := fmt.Sprintf("{\"schema_version\":1,\"token\":%q,\"plan_id\":%q,\"pid\":%d,\"created_at\":%q}\n", token, plan.ID, os.Getpid(), now.Add(-time.Hour).Format(time.RFC3339Nano))
	lockPath := filepath.Join(repository, ".git", "starter-kit.lock")
	if err := os.WriteFile(lockPath, []byte(lease), 0o600); err != nil {
		t.Fatalf("write live lifecycle lease: %v", err)
	}

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "lock" || !failure.Recoverable || len(failure.Recovery) == 0 {
		t.Fatalf("live lease did not block with recovery guidance: %#v, %v", failure, err)
	}
	if result.Status != engine.ApplyStatusFailed {
		t.Fatalf("live lease result is not failed: %#v", result)
	}
	content, readErr := os.ReadFile(lockPath)
	if readErr != nil || string(content) != lease {
		t.Fatalf("apply stole or changed a live lease: %q, %v", content, readErr)
	}
	if _, err := os.Stat(filepath.Join(repository, ".git", "starter-kit-attempts", "stale-lock-"+token+".json")); !os.IsNotExist(err) {
		t.Fatalf("live lease was archived as stale: %v", err)
	}
}

func TestApplyResumesInterruptedMatchingCreateWithoutReplacingCommittedPrefix(t *testing.T) {
	repository := newGitRepository(t)
	now := time.Date(2026, 7, 13, 18, 0, 0, 0, time.UTC)
	lifecycle := engine.New(engine.WithClock(fixedClock{now}))
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	committed := plan.Files[:4]
	for _, file := range committed {
		target := filepath.Join(repository, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("create interrupted parent: %v", err)
		}
		if err := os.WriteFile(target, []byte(file.Content), 0o644); err != nil {
			t.Fatalf("write interrupted committed prefix: %v", err)
		}
	}
	preservedPath := filepath.Join(repository, filepath.FromSlash(committed[0].Path))
	preservedInfo, err := os.Stat(preservedPath)
	if err != nil {
		t.Fatalf("stat committed prefix: %v", err)
	}
	token := "abcdef0123456789abcdef0123456789"
	stagePath := filepath.Join(repository, ".starter-kit-stage-"+token)
	if err := os.Mkdir(stagePath, 0o700); err != nil {
		t.Fatalf("create abandoned stage: %v", err)
	}
	marker := fmt.Sprintf("{\"schema_version\":1,\"ownership\":\"machine-state\",\"source\":\"engine:apply:v1\",\"lease_token\":%q,\"plan_id\":%q,\"created_at\":%q}\n", token, plan.ID, now.Add(-time.Hour).Format(time.RFC3339Nano))
	if err := os.WriteFile(filepath.Join(stagePath, ".starter-kit-transaction.json"), []byte(marker), 0o600); err != nil {
		t.Fatalf("write abandoned stage marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stagePath, "partial"), []byte("staged\n"), 0o600); err != nil {
		t.Fatalf("write abandoned stage: %v", err)
	}
	lease := fmt.Sprintf("{\"schema_version\":1,\"token\":%q,\"plan_id\":%q,\"pid\":2147483647,\"created_at\":%q}\n", token, plan.ID, now.Add(-time.Hour).Format(time.RFC3339Nano))
	if err := os.WriteFile(filepath.Join(repository, ".git", "starter-kit.lock"), []byte(lease), 0o600); err != nil {
		t.Fatalf("write interrupted lifecycle lease: %v", err)
	}

	result, err := lifecycle.Apply(t.Context(), plan.ID, plan)
	if err != nil {
		t.Fatalf("resume interrupted create: %v", err)
	}
	if result.Status != engine.ApplyStatusApplied || len(result.Recovery) < 2 {
		t.Fatalf("interrupted create recovery was not disclosed: %#v", result)
	}
	if _, err := os.Stat(stagePath); !os.IsNotExist(err) {
		t.Fatalf("abandoned stage remained after recovery: %v", err)
	}
	afterInfo, err := os.Stat(preservedPath)
	if err != nil {
		t.Fatalf("stat preserved prefix after recovery: %v", err)
	}
	if !afterInfo.ModTime().Equal(preservedInfo.ModTime()) {
		t.Fatalf("resume replaced an already committed matching artifact")
	}
	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil || status.Lifecycle != engine.LifecycleManaged {
		t.Fatalf("resumed repository is not managed: %#v, %v", status, err)
	}
}

func TestStatusExplainsIncompleteCreateAndSafeRecovery(t *testing.T) {
	repository := newGitRepository(t)
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	for _, file := range plan.Files[:4] {
		target := filepath.Join(repository, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("create interrupted parent: %v", err)
		}
		if err := os.WriteFile(target, []byte(file.Content), 0o644); err != nil {
			t.Fatalf("write interrupted artifact: %v", err)
		}
	}
	if err := os.Mkdir(filepath.Join(repository, ".starter-kit-stage-abandoned"), 0o700); err != nil {
		t.Fatalf("create abandoned stage: %v", err)
	}

	status, err := engine.New().Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status incomplete create: %v", err)
	}
	if status.Lifecycle != engine.LifecycleSetupIncomplete || len(status.Problems) == 0 || len(status.Recovery) == 0 {
		t.Fatalf("status did not explain incomplete recoverable setup: %#v", status)
	}
	if _, err := os.Stat(filepath.Join(repository, ".starter-kit-stage-abandoned")); err != nil {
		t.Fatalf("read-only status mutated abandoned stage: %v", err)
	}
}

func TestApplyPreservesUnrecognizedStagingLookalikeForReconciliation(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	stagePath := filepath.Join(repository, ".starter-kit-stage-user-content")
	if err := os.Mkdir(stagePath, 0o700); err != nil {
		t.Fatalf("create staging lookalike: %v", err)
	}
	original := []byte("preserve me\n")
	if err := os.WriteFile(filepath.Join(stagePath, "human.txt"), original, 0o600); err != nil {
		t.Fatalf("write staging lookalike content: %v", err)
	}

	_, err = lifecycle.Apply(t.Context(), plan.ID, plan)
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "recover-stage" {
		t.Fatalf("unrecognized stage did not stop safely: %#v, %v", failure, err)
	}
	content, readErr := os.ReadFile(filepath.Join(stagePath, "human.txt"))
	if readErr != nil || !reflect.DeepEqual(content, original) {
		t.Fatalf("unrecognized stage content was removed: %q, %v", content, readErr)
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

func TestApplyDoesNotEchoFixtureSecretFromHostilePlanPath(t *testing.T) {
	repository := newGitRepository(t)
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	secret := "ghp_12345678901234567890"
	plan.Files[0].Path = "../" + secret + "/escape.txt"
	plan.ID = identifyPlan(t, plan)

	_, err = engine.New().Apply(t.Context(), plan.ID, plan)
	if err == nil {
		t.Fatal("apply accepted hostile secret-bearing path")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("apply diagnostic exposed fixture secret: %v", err)
	}
}

func TestApplyRejectsSecretBearingForgedResultPathBeforeManagedEffects(t *testing.T) {
	repository := newGitRepository(t)
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	secret := "ghp_12345678901234567890"
	plan.Result.Path = ".starter-kit/events/create-" + secret + ".json"
	plan.ID = identifyPlan(t, plan)

	_, err = engine.New().Apply(t.Context(), plan.ID, plan)
	if err == nil {
		t.Fatal("apply accepted a forged secret-bearing result path")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("apply diagnostic exposed fixture secret: %v", err)
	}
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "validate-plan" || len(failure.ChangedFiles) != 0 {
		t.Fatalf("forged result path lacks structured pre-transaction failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repository, ".starter-kit")); !os.IsNotExist(err) {
		t.Fatalf("rejected result path produced repository effects: %v", err)
	}
	evidencePath := filepath.Join(repository, ".git", "starter-kit-attempts", strings.TrimPrefix(plan.ID, "sha256:")+".json")
	evidence, readErr := os.ReadFile(evidencePath)
	if readErr != nil {
		t.Fatalf("read rejected-plan evidence: %v", readErr)
	}
	if strings.Contains(string(evidence), secret) || !strings.Contains(string(evidence), `"status": "failed"`) {
		t.Fatalf("rejected-plan evidence is unsafe or untruthful: %s", evidence)
	}
}

func TestApplyRejectsSecretBearingRepositoryDigestBeforeGeneratingEvidence(t *testing.T) {
	repository := newGitRepository(t)
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	secret := "ghp_12345678901234567890"
	plan.RepositoryDigest = secret
	plan.Result.Path = ".starter-kit/events/create-" + secret[:16] + ".json"
	plan.ID = identifyPlan(t, plan)

	_, err = engine.New().Apply(t.Context(), plan.ID, plan)
	if err == nil {
		t.Fatal("apply accepted a secret-bearing repository digest")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("apply diagnostic exposed fixture secret: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repository, ".starter-kit")); !os.IsNotExist(err) {
		t.Fatalf("invalid repository digest produced managed effects: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repository, ".git", "starter-kit-attempts")); !os.IsNotExist(err) {
		t.Fatalf("invalid repository digest entered ordinary evidence: %v", err)
	}
}

func TestApplyRejectsFixtureSecretInSelfConsistentPlanBeforeStaging(t *testing.T) {
	repository := newGitRepository(t)
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	secret := "ghp_12345678901234567890"
	plan.Files[0].Content = "token=" + secret + "\n"
	digest := sha256.Sum256([]byte(plan.Files[0].Content))
	plan.Files[0].Digest = "sha256:" + hex.EncodeToString(digest[:])
	plan.ID = identifyPlan(t, plan)

	_, err = engine.New().Apply(t.Context(), plan.ID, plan)
	if err == nil {
		t.Fatal("apply accepted fixture secret in a self-consistent plan")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("apply diagnostic exposed fixture secret: %v", err)
	}
	var failure *engine.ApplyFailure
	if !errors.As(err, &failure) || failure.Stage != "validate-plan" || len(failure.ChangedFiles) != 0 {
		t.Fatalf("fixture secret lacks structured pre-transaction failure: %v", err)
	}
}

func TestApplyRejectsUnsafeCrossPlatformPathNamespaceBeforeManagedEffects(t *testing.T) {
	tests := map[string]func(*engine.Plan){
		"absolute":            func(plan *engine.Plan) { plan.Files[0].Path = "/escape.txt" },
		"windows absolute":    func(plan *engine.Plan) { plan.Files[0].Path = "C:/escape.txt" },
		"unclean":             func(plan *engine.Plan) { plan.Files[0].Path = "docs/../escape.txt" },
		"empty segment":       func(plan *engine.Plan) { plan.Files[0].Path = "docs//escape.txt" },
		"reserved name":       func(plan *engine.Plan) { plan.Files[0].Path = "docs/CON.txt" },
		"trailing dot":        func(plan *engine.Plan) { plan.Files[0].Path = "docs/escape." },
		"ambiguous unicode":   func(plan *engine.Plan) { plan.Files[0].Path = "docs/café.txt" },
		"case-fold collision": func(plan *engine.Plan) { plan.Files[0].Path, plan.Files[1].Path = "docs/Case.txt", "docs/case.txt" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repository := newGitRepository(t)
			plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
			if err != nil {
				t.Fatalf("create plan: %v", err)
			}
			mutate(&plan)
			plan.ID = identifyPlan(t, plan)

			_, err = engine.New().Apply(t.Context(), plan.ID, plan)
			if err == nil {
				t.Fatal("apply accepted unsafe cross-platform path namespace")
			}
			var failure *engine.ApplyFailure
			if !errors.As(err, &failure) || failure.Stage != "validate-plan" || len(failure.ChangedFiles) != 0 {
				t.Fatalf("unsafe path lacks structured pre-transaction failure: %v", err)
			}
			if _, err := os.Stat(filepath.Join(repository, ".starter-kit", "state.json")); !os.IsNotExist(err) {
				t.Fatalf("state exists after rejected plan: %v", err)
			}
		})
	}
}

func TestCreateRejectsReservedDirectorySymlinkEscapeDuringPlanning(t *testing.T) {
	repository := newGitRepository(t)
	outside := t.TempDir()
	link := filepath.Join(repository, ".starter-kit")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("native filesystem cannot create symlink fixture: %v", err)
	}

	if _, err := engine.New().Create(t.Context(), approvedCreate(repository)); err == nil {
		t.Fatal("create planned through reserved directory symlink")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatalf("read outside directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("planning wrote outside repository through symlink: %#v", entries)
	}
}

func TestCreateRejectsSymlinkRepositoryRoot(t *testing.T) {
	actual := newGitRepository(t)
	link := filepath.Join(t.TempDir(), "repository-link")
	if err := os.Symlink(actual, link); err != nil {
		t.Skipf("native filesystem cannot create symlink fixture: %v", err)
	}

	if _, err := engine.New().Create(t.Context(), approvedCreate(link)); err == nil {
		t.Fatal("create accepted a symlink as the authorized repository root")
	}
}

func TestCreateCanonicalizesRepositoryRootBelowSymlinkedAncestor(t *testing.T) {
	actual := newGitRepository(t)
	parentLink := filepath.Join(t.TempDir(), "parent-link")
	if err := os.Symlink(filepath.Dir(actual), parentLink); err != nil {
		t.Skipf("native filesystem cannot create symlink fixture: %v", err)
	}
	linkedRoot := filepath.Join(parentLink, filepath.Base(actual))
	canonicalRoot, err := filepath.EvalSymlinks(actual)
	if err != nil {
		t.Fatalf("canonicalize actual repository: %v", err)
	}

	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(linkedRoot))
	if err != nil {
		t.Fatalf("create through canonicalizable ancestor alias: %v", err)
	}
	if plan.Repository != canonicalRoot {
		t.Fatalf("plan repository = %q, want canonical root %q", plan.Repository, canonicalRoot)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply canonical plan: %v", err)
	}
	if _, err := os.Stat(filepath.Join(canonicalRoot, ".starter-kit", "state.json")); err != nil {
		t.Fatalf("canonical repository did not receive managed state: %v", err)
	}
}

func TestStatusRejectsManagedArtifactSymlinkEvenWhenContentDigestMatches(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	managedPath := filepath.Join(repository, "AGENTS.md")
	content, err := os.ReadFile(managedPath)
	if err != nil {
		t.Fatalf("read managed artifact: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, content, 0o644); err != nil {
		t.Fatalf("write external target: %v", err)
	}
	if err := os.Remove(managedPath); err != nil {
		t.Fatalf("remove managed artifact: %v", err)
	}
	if err := os.Symlink(outside, managedPath); err != nil {
		t.Skipf("native filesystem cannot create symlink fixture: %v", err)
	}

	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status symlinked repository: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManagedDegraded || len(status.Problems) == 0 {
		t.Fatalf("matching external content concealed managed symlink: %#v", status)
	}
}

func TestCreatePreservesExistingDirectoryThatCaseCollidesWithManagedPath(t *testing.T) {
	repository := newGitRepository(t)
	existing := filepath.Join(repository, "DOCS")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatalf("create user-owned directory: %v", err)
	}

	if _, err := engine.New().Create(t.Context(), approvedCreate(repository)); err == nil {
		t.Fatal("create planned over case-colliding user-owned directory")
	}
	info, err := os.Stat(existing)
	if err != nil || !info.IsDir() {
		t.Fatalf("user-owned directory was not preserved: %v", err)
	}
	caseMode := "case-sensitive"
	if alternate, err := os.Stat(filepath.Join(repository, "docs")); err == nil && os.SameFile(info, alternate) {
		caseMode = "case-insensitive"
	}
	t.Logf("native repository filesystem is %s for DOCS/docs", caseMode)
}

func TestApplyRejectsSelfConsistentUnsupportedPlanContracts(t *testing.T) {
	tests := map[string]func(*engine.Plan){
		"schema":    func(plan *engine.Plan) { plan.SchemaVersion = 2 },
		"operation": func(plan *engine.Plan) { plan.Operation = engine.Operation("upgrade") },
		"approval":  func(plan *engine.Plan) { plan.Approval.BriefApproved = false },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repository := newGitRepository(t)
			plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
			if err != nil {
				t.Fatalf("create plan: %v", err)
			}
			mutate(&plan)
			plan.ID = identifyPlan(t, plan)
			if _, err := engine.New().Apply(t.Context(), plan.ID, plan); err == nil {
				t.Fatal("apply accepted unsupported self-consistent plan contract")
			}
		})
	}
}

func TestApplyRejectsCreatePlanThatExpandsOrReclassifiesManagedWrites(t *testing.T) {
	tests := map[string]func(*engine.Plan){
		"ownership": func(plan *engine.Plan) { plan.Files[0].Ownership = "human-owned" },
		"source":    func(plan *engine.Plan) { plan.Files[0].Source = "repository:untrusted" },
		"extra path": func(plan *engine.Plan) {
			content := "unapproved\n"
			digest := sha256.Sum256([]byte(content))
			plan.Files = append(plan.Files, engine.PlannedFile{
				Path: "unapproved.txt", Ownership: "managed", Source: "engine:create:v1",
				Digest: "sha256:" + hex.EncodeToString(digest[:]), Content: content,
			})
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repository := newGitRepository(t)
			plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
			if err != nil {
				t.Fatalf("create plan: %v", err)
			}
			mutate(&plan)
			plan.ID = identifyPlan(t, plan)

			_, err = engine.New().Apply(t.Context(), plan.ID, plan)
			if err == nil {
				t.Fatal("apply accepted expanded or reclassified create write")
			}
			var failure *engine.ApplyFailure
			if !errors.As(err, &failure) || failure.Stage != "validate-plan" || len(failure.ChangedFiles) != 0 {
				t.Fatalf("invalid create contract lacks structured pre-transaction failure: %v", err)
			}
		})
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
	if result.Status != engine.ApplyStatusFailed || len(result.ChangedFiles) != 1 || result.ChangedFiles[0] != plan.Result.Path {
		t.Fatalf("rollback result must report only retained failure evidence: %#v", result)
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

func rewriteManagedFile(t *testing.T, repository, slashPath, content string) {
	t.Helper()
	target := filepath.Join(repository, filepath.FromSlash(slashPath))
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write adversarial managed file: %v", err)
	}
	manifestPath := filepath.Join(repository, ".starter-kit", "managed-files.json")
	manifestContent, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	digest := sha256.Sum256([]byte(content))
	for _, raw := range manifest["files"].([]interface{}) {
		entry := raw.(map[string]interface{})
		if entry["path"] == slashPath {
			entry["digest"] = "sha256:" + hex.EncodeToString(digest[:])
		}
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}
