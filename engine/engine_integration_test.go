package engine_test

import (
	"context"
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

func TestApplyCreatePlanProducesManagedRepository(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), repository)
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

	first, err := lifecycle.Create(context.Background(), repository)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	second, err := lifecycle.Plan(context.Background(), repository, engine.CreateOperation)
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

func TestCreateAfterApplyReturnsExplicitNoChange(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), repository)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(context.Background(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}

	unchanged, err := lifecycle.Create(context.Background(), repository)
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

func TestApplyRejectsRepositoryChangedAfterPlanning(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), repository)
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

func TestApplyRejectsPlanWhoseContentDoesNotMatchIdentifier(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New()
	plan, err := lifecycle.Create(context.Background(), repository)
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

func newGitRepository(t *testing.T) string {
	t.Helper()
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	return repository
}
