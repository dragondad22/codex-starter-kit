package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSandboxLifecycleReconcilesMissingManagedResourceAndReplays(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{
		Host:           "github.com",
		OwnerID:        "305967668",
		RepositoryID:   "1303189066",
		ProjectID:      "PVT_kwDOEjyyNM4Bdm9F",
		RepositoryName: "codex-starter-kit-labs/codex-starter-kit-sandbox",
	}
	manifest := SandboxManifest{
		SchemaVersion:         1,
		OperationID:           "issue-73-bootstrap-v1",
		SourceRevision:        "source-73",
		ConfigurationRevision: "configuration-73",
		ApprovedBy:            "dragondad22",
		ApprovedPlan:          "issue-73-bootstrap-v1",
		RecoveryOwner:         "sandbox-owner",
		MarkerPrefix:          "starter-kit-contract:",
		Target:                target,
		Resources: []SandboxResourceSpec{{
			Key:        "label:type-task",
			Kind:       SandboxResourceLabel,
			Name:       "type:task",
			Attributes: map[string]string{"color": "0075CA", "description": "Independently executable implementation work"},
		}},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{
			SchemaVersion:         1,
			Available:             true,
			Fresh:                 true,
			Actor:                 "codex-starter-kit-labs-reconciler",
			EvidenceMode:          "memory",
			Target:                target,
			Permissions:           []string{"issues:write", "organization-projects:write"},
			ConfigurationRevision: manifest.ConfigurationRevision,
			ObservedAt:            now,
			ExpiresAt:             now.Add(time.Hour),
		},
		SandboxObservation{
			SchemaVersion:         1,
			Target:                target,
			ConfigurationRevision: manifest.ConfigurationRevision,
			Resources:             []SandboxObservedResource{},
		},
	)
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	request := SandboxRequest{Repository: repository, Manifest: manifest}

	first, err := bootstrapApprovedSandbox(t, lifecycle, request, now)
	if err != nil {
		t.Fatalf("bootstrap sandbox: %v", err)
	}
	if first.Apply.Status != SandboxApplyApplied {
		t.Fatalf("apply status = %q, want applied", first.Apply.Status)
	}
	if len(first.Apply.Receipts) != 1 || first.Apply.Receipts[0].ResourceKey != "label:type-task" {
		t.Fatalf("unexpected receipts: %#v", first.Apply.Receipts)
	}
	if first.Verification.OverallState != ControlPass || first.Status.Disposition != "converged" {
		t.Fatalf("verification/status = %q/%q", first.Verification.OverallState, first.Status.Disposition)
	}

	second, err := bootstrapApprovedSandbox(t, lifecycle, request, now)
	if err != nil {
		t.Fatalf("replay sandbox: %v", err)
	}
	if !second.Plan.NoChange || second.Apply.Status != SandboxApplyNoChange {
		t.Fatalf("replay plan/apply = no_change:%v/%q", second.Plan.NoChange, second.Apply.Status)
	}
	if len(adapter.Effects()) != 1 {
		t.Fatalf("adapter effects after replay = %d, want 1", len(adapter.Effects()))
	}
}

func TestSandboxInspectionStopsOnUnrecognizedNameCollision(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{
		SchemaVersion: 1, OperationID: "approved-plan", SourceRevision: "source", ConfigurationRevision: "config",
		ApprovedBy: "owner", ApprovedPlan: "approved-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
		Resources: []SandboxResourceSpec{{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task", Attributes: map[string]string{"color": "0075CA"}}},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)},
		SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config", Resources: []SandboxObservedResource{{Key: "human:label:1", Kind: SandboxResourceLabel, Name: "type:task", ID: "human-label", Attributes: map[string]string{"color": "FFFFFF"}}}},
	)
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))

	inspection, err := lifecycle.InspectSandbox(context.Background(), SandboxRequest{Repository: repository, Manifest: manifest})
	if err != nil {
		t.Fatalf("inspect sandbox: %v", err)
	}
	if inspection.Disposition != "non-pass" || !strings.Contains(strings.Join(inspection.Problems, " "), "unrecognized") {
		t.Fatalf("inspection = %q %#v", inspection.Disposition, inspection.Problems)
	}
	if _, err := lifecycle.PlanSandbox(context.Background(), inspection); err == nil {
		t.Fatal("expected collision to prevent planning")
	}
	if len(adapter.Effects()) != 0 {
		t.Fatalf("collision produced %d effects", len(adapter.Effects()))
	}
}

func TestSandboxApplyRejectsChangedObservationBeforeEffects(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{
		SchemaVersion: 1, OperationID: "approved-plan", SourceRevision: "source", ConfigurationRevision: "config",
		ApprovedBy: "owner", ApprovedPlan: "approved-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
		Resources: []SandboxResourceSpec{{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task", Attributes: map[string]string{"color": "0075CA"}}},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)},
		SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config", Resources: []SandboxObservedResource{}},
	)
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	inspection, err := lifecycle.InspectSandbox(context.Background(), SandboxRequest{Repository: repository, Manifest: manifest})
	if err != nil {
		t.Fatalf("inspect sandbox: %v", err)
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	if err != nil {
		t.Fatalf("plan sandbox: %v", err)
	}
	adapter.SetObservation(SandboxObservation{
		SchemaVersion: 1, Target: target, ConfigurationRevision: "config",
		Resources: []SandboxObservedResource{{Key: "human:unrelated", Kind: SandboxResourceLabel, Name: "human-label", ID: "human-label", Attributes: map[string]string{"color": "FFFFFF"}}},
	})

	result, err := lifecycle.ApplySandbox(context.Background(), plan, approveSandbox(plan, now))
	if err != nil {
		t.Fatalf("apply sandbox: %v", err)
	}
	if result.Status != SandboxApplyNonPass || !strings.Contains(strings.Join(result.Problems, " "), "stale") {
		t.Fatalf("apply = %q %#v", result.Status, result.Problems)
	}
	if len(adapter.Effects()) != 0 {
		t.Fatalf("stale plan produced %d effects", len(adapter.Effects()))
	}
}

func TestSandboxPartialApplyPlansOnlyRemainingSemanticDelta(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{
		SchemaVersion: 1, OperationID: "approved-plan", SourceRevision: "source", ConfigurationRevision: "config",
		ApprovedBy: "owner", ApprovedPlan: "approved-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
		Resources: []SandboxResourceSpec{
			{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task", Attributes: map[string]string{"color": "0075CA"}},
			{Key: "label:contract-run", Kind: SandboxResourceLabel, Name: "contract-run", Attributes: map[string]string{"color": "5319E7"}},
		},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)},
		SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config", Resources: []SandboxObservedResource{}},
	)
	adapter.QueueApplyResult(SandboxEffectResult{Outcome: "applied", ResourceID: "label-1", Detail: "created"}, true)
	adapter.QueueApplyResult(SandboxEffectResult{Outcome: "denied", Detail: "permission denied"}, false)
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	request := SandboxRequest{Repository: repository, Manifest: manifest}

	first, err := bootstrapApprovedSandbox(t, lifecycle, request, now)
	if err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}
	if first.Apply.Status != SandboxApplyNonPass || len(first.Apply.Receipts) != 2 {
		t.Fatalf("first apply = %q %#v", first.Apply.Status, first.Apply.Receipts)
	}
	restarted := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	status, err := restarted.SandboxStatus(context.Background(), repository)
	if err != nil {
		t.Fatalf("restart status: %v", err)
	}
	if status.Disposition != "non-pass" || len(status.Receipts) != 2 || status.Receipts[0].ResourceKey != "label:type-task" {
		t.Fatalf("restart status = %#v", status)
	}
	inspection, err := lifecycle.InspectSandbox(context.Background(), request)
	if err != nil {
		t.Fatalf("reinspect: %v", err)
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	if err != nil {
		t.Fatalf("replan: %v", err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Resource.Key != "label:contract-run" {
		t.Fatalf("remaining effects: %#v", plan.Effects)
	}
}

func TestSandboxCleanupRemovesOnlyExactManagedFixture(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	managed := SandboxObservedResource{Key: "fixture:issue:run-1", Kind: SandboxResourceFixtureIssue, Name: "contract fixture", ID: "issue-1", Marker: "starter-kit-contract:run-1:issue"}
	human := SandboxObservedResource{Key: "human:issue:2", Kind: SandboxResourceFixtureIssue, Name: "human issue", ID: "issue-2"}
	manifest := SandboxManifest{
		SchemaVersion: 1, OperationID: "cleanup-plan", SourceRevision: "source", ConfigurationRevision: "config",
		ApprovedBy: "owner", ApprovedPlan: "cleanup-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
		Resources: []SandboxResourceSpec{{Key: managed.Key, Kind: managed.Kind, Name: managed.Name, Marker: managed.Marker, DesiredState: SandboxResourceAbsent}},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)},
		SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config", Resources: []SandboxObservedResource{managed, human}},
	)
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))

	result, err := bootstrapApprovedSandbox(t, lifecycle, SandboxRequest{Repository: repository, Manifest: manifest}, now)
	if err != nil {
		t.Fatalf("cleanup sandbox: %v", err)
	}
	if result.Apply.Status != SandboxApplyApplied || len(result.Plan.Effects) != 1 || result.Plan.Effects[0].Kind != "remove-resource" {
		t.Fatalf("cleanup plan/apply = %#v/%q", result.Plan.Effects, result.Apply.Status)
	}
	remaining := adapter.Observation().Resources
	if len(remaining) != 1 || remaining[0].Key != human.Key {
		t.Fatalf("remaining resources = %#v", remaining)
	}
}

func TestSandboxApplyDoesNotStealActiveLifecycleLease(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{
		SchemaVersion: 1, OperationID: "approved-plan", SourceRevision: "source", ConfigurationRevision: "config",
		ApprovedBy: "owner", ApprovedPlan: "approved-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
		Resources: []SandboxResourceSpec{{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task"}},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)},
		SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config"},
	)
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	inspection, err := lifecycle.InspectSandbox(context.Background(), SandboxRequest{Repository: repository, Manifest: manifest})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	lease := fmt.Sprintf("{\"schema_version\":1,\"token\":\"%032x\",\"plan_id\":%q,\"pid\":%d,\"created_at\":%q}\n", 1, "other-plan", os.Getpid(), now.Format(time.RFC3339Nano))
	lockPath := filepath.Join(repository, ".git", "starter-kit.lock")
	if err := os.WriteFile(lockPath, []byte(lease), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := lifecycle.ApplySandbox(context.Background(), plan, approveSandbox(plan, now)); err == nil {
		t.Fatal("expected active lifecycle lease to block sandbox apply")
	}
	if len(adapter.Effects()) != 0 {
		t.Fatalf("active lease allowed %d effects", len(adapter.Effects()))
	}
}

func TestSandboxManifestRejectsUnsupportedKindsAndSensitiveMaterial(t *testing.T) {
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	tests := []struct {
		name     string
		resource SandboxResourceSpec
		want     string
	}{
		{
			name:     "unsupported kind",
			resource: SandboxResourceSpec{Key: "repository:settings", Kind: "repository-settings", Name: "settings"},
			want:     "unsupported sandbox resource kind",
		},
		{
			name:     "credential-shaped attribute",
			resource: SandboxResourceSpec{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task", Attributes: map[string]string{"token": "ghp_1234567890abcdefghijklmnopqrstuvwxyz"}},
			want:     "sensitive-looking material",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := SandboxManifest{
				SchemaVersion: 1, OperationID: "approved-plan", SourceRevision: "source", ConfigurationRevision: "config",
				ApprovedBy: "owner", ApprovedPlan: "approved-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
				Resources: []SandboxResourceSpec{test.resource},
			}
			if err := validateSandboxManifest(manifest); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate manifest error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestSandboxApplyPersistsRedactedReceiptWhenAdapterFails(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{
		SchemaVersion: 1, OperationID: "approved-plan", SourceRevision: "source", ConfigurationRevision: "config",
		ApprovedBy: "owner", ApprovedPlan: "approved-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target,
		Resources: []SandboxResourceSpec{{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task"}},
	}
	adapter := NewInMemorySandboxAdapter(
		SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)},
		SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config"},
	)
	adapter.QueueApplyError(fmt.Errorf("transport failed with ghp_1234567890abcdefghijklmnopqrstuvwxyz"))
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	inspection, err := lifecycle.InspectSandbox(context.Background(), SandboxRequest{Repository: repository, Manifest: manifest})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}

	result, err := lifecycle.ApplySandbox(context.Background(), plan, approveSandbox(plan, now))
	if err != nil {
		t.Fatalf("apply should return durable non-pass result: %v", err)
	}
	if result.Status != SandboxApplyNonPass || len(result.Receipts) != 1 || result.Receipts[0].Outcome != "error" || strings.Contains(result.Receipts[0].Detail, "ghp_") {
		t.Fatalf("apply result = %#v", result)
	}
	status, err := lifecycle.SandboxStatus(context.Background(), repository)
	if err != nil || status.Disposition != "non-pass" || len(status.Receipts) != 1 {
		t.Fatalf("durable status = %#v, %v", status, err)
	}
}

func TestSandboxApplyRejectsApprovalForDifferentGeneratedPlan(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	repository := newSandboxRepository(t)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{SchemaVersion: 1, OperationID: "operation", SourceRevision: "source", ConfigurationRevision: "config", ApprovedBy: "owner", ApprovedPlan: "provisioning-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target, Resources: []SandboxResourceSpec{{Key: "label:type-task", Kind: SandboxResourceLabel, Name: "type:task"}}}
	adapter := NewInMemorySandboxAdapter(SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)}, SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config"})
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))
	inspection, err := lifecycle.InspectSandbox(context.Background(), SandboxRequest{Repository: repository, Manifest: manifest})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	approval := approveSandbox(plan, now)
	approval.PlanID = "different-plan"

	if _, err := lifecycle.ApplySandbox(context.Background(), plan, approval); err == nil || !strings.Contains(err.Error(), "separate approval") {
		t.Fatalf("apply error = %v", err)
	}
	if len(adapter.Effects()) != 0 {
		t.Fatalf("mismatched approval produced %d effects", len(adapter.Effects()))
	}
}

func TestSandboxVerificationCannotPassWithObservationProblems(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	target := SandboxTarget{Host: "github.com", OwnerID: "owner", RepositoryID: "repo", ProjectID: "project", RepositoryName: "owner/sandbox"}
	manifest := SandboxManifest{SchemaVersion: 1, OperationID: "operation", SourceRevision: "source", ConfigurationRevision: "config", ApprovedBy: "owner", ApprovedPlan: "provisioning-plan", RecoveryOwner: "sandbox-owner", MarkerPrefix: "starter-kit-contract:", Target: target}
	adapter := NewInMemorySandboxAdapter(SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: "app", EvidenceMode: "memory", Target: target, ConfigurationRevision: "config", ObservedAt: now, ExpiresAt: now.Add(time.Hour)}, SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: "config", Problems: []string{"Project inventory unavailable"}})
	lifecycle := New(WithClock(sandboxFixedClock{now}), WithSandboxAdapter(adapter))

	verification, err := lifecycle.VerifySandbox(context.Background(), manifest)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState == ControlPass || !strings.Contains(verification.Controls[0].Rationale, "inventory unavailable") {
		t.Fatalf("verification = %#v", verification)
	}
}

type sandboxFixedClock struct{ now time.Time }

func (clock sandboxFixedClock) Now() time.Time { return clock.now }

func approveSandbox(plan SandboxPlan, now time.Time) SandboxPlanApproval {
	return SandboxPlanApproval{SchemaVersion: 1, PlanID: plan.ID, ApprovedBy: "test-owner", ApprovalID: "test-approval:" + plan.ID, ApprovedAt: now}
}

func bootstrapApprovedSandbox(t *testing.T, lifecycle *Engine, request SandboxRequest, now time.Time) (SandboxLifecycleResult, error) {
	t.Helper()
	result := SandboxLifecycleResult{SchemaVersion: 1}
	inspection, err := lifecycle.InspectSandbox(context.Background(), request)
	result.Inspection = inspection
	if err != nil {
		return result, err
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	result.Plan = plan
	if err != nil {
		return result, err
	}
	result.Apply, err = lifecycle.ApplySandbox(context.Background(), plan, approveSandbox(plan, now))
	if err != nil {
		return result, err
	}
	result.Verification, err = lifecycle.VerifySandbox(context.Background(), request.Manifest)
	if err != nil {
		return result, err
	}
	if err := updateSandboxVerification(plan.Repository, result.Verification); err != nil {
		return result, err
	}
	result.Status, err = lifecycle.SandboxStatus(context.Background(), request.Repository)
	return result, err
}

func newSandboxRepository(t *testing.T) string {
	t.Helper()
	repository := t.TempDir()
	if output, err := exec.Command("git", "init", "--quiet", repository).CombinedOutput(); err != nil {
		t.Fatalf("initialize sandbox repository: %v: %s", err, output)
	}
	return repository
}
