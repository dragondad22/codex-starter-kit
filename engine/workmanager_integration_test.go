package engine_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestManagedTaskLifecycleConvergesThroughInMemoryAdapter(t *testing.T) {
	t.Parallel()

	repository := t.TempDir()
	initializeWorkGit(t, repository)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	adapter := engine.NewInMemoryWorkAdapter(engine.WorkCapability{
		SchemaVersion:         1,
		Online:                true,
		Fresh:                 true,
		Mode:                  "memory",
		Actor:                 "test:maintainer",
		Permissions:           []string{"issues:write", "projects:write", "pull_requests:read"},
		ConfigurationRevision: "project-config:v1",
		ObservedAt:            now,
		ExpiresAt:             now.Add(time.Hour),
	}, engine.WorkObservation{
		SchemaVersion:         1,
		Revision:              "observation:v1",
		ConfigurationRevision: "project-config:v1",
		Target: engine.WorkTarget{
			Host:         "memory.local",
			RepositoryID: "repository:fixture",
			ProjectID:    "project:fixture",
			FieldIDs:     map[string]string{"readiness": "field:readiness", "status": "field:status", "phase": "field:phase"},
			OptionIDs: map[string]string{
				"readiness:ready": "option:ready",
				"status:next":     "option:next",
				"phase:Phase 0":   "option:phase-0",
				"phase:Phase 1":   "option:phase-1",
				"phase:Phase 2":   "option:phase-2",
				"phase:Phase 3":   "option:phase-3",
				"phase:Phase 4":   "option:phase-4",
				"phase:Phase 5":   "option:phase-5",
				"phase:Phase 6":   "option:phase-6",
				"phase:Phase 7":   "option:phase-7",
				"phase:Phase 8":   "option:phase-8",
			},
		},
	})
	lifecycle := engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))
	request := engine.ManagedTaskRequest{
		Repository: repository,
		Intent: engine.WorkDesiredIntent{
			SchemaVersion:            1,
			OperationID:              "operation:issue-71",
			SourceRevision:           "issue-71:v1",
			OperatingProfileRevision: "operating-profile:v1",
			InputDigests:             map[string]string{"issue": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			Credential:               engine.WorkCredentialExpectation{Mode: "memory", Actor: "test:maintainer"},
			Target:                   adapter.Observation().Target,
			Task: engine.DesiredManagedTask{
				ManagedID:             "issue:71",
				IssueType:             "task",
				Title:                 "Manage one task deterministically through the lifecycle engine",
				Readiness:             "ready",
				Status:                "next",
				Phase:                 "Phase 3",
				PhaseAssignmentReason: "Cross-cutting Work Manager delivery is sequenced in Phase 3.",
				Review: []engine.WorkReviewRequirement{{
					Role:            "change-review",
					DistinctContext: true,
				}},
			},
		},
	}

	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatalf("inspect managed task: %v", err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatalf("plan managed task: %v", err)
	}
	if plan.ID == "" || len(plan.Effects) == 0 {
		t.Fatalf("expected immutable effect plan, got %#v", plan)
	}
	result, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatalf("apply managed task: %v", err)
	}
	if result.Status != engine.WorkApplyApplied || len(result.Receipts) == 0 {
		t.Fatalf("expected applied receipts, got %#v", result)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), repository)
	if err != nil {
		t.Fatalf("verify managed task: %v", err)
	}
	if verification.OverallState != engine.ControlPass {
		t.Fatalf("expected converged verification, got %#v", verification)
	}

	restarted := engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))
	status, err := restarted.ManagedTaskStatus(context.Background(), repository)
	if err != nil {
		t.Fatalf("read managed task status after restart: %v", err)
	}
	if status.Disposition != "converged" || len(status.Receipts) == 0 {
		t.Fatalf("expected durable converged status, got %#v", status)
	}
}

type fixedWorkClock struct{ now time.Time }

func (clock fixedWorkClock) Now() time.Time { return clock.now }

func TestManagedTaskReplayProducesNoEffectsAndPreservesReceipts(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	first, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatal(err)
	}

	replayedInspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	replayedPlan, err := lifecycle.PlanManagedTask(context.Background(), replayedInspection)
	if err != nil {
		t.Fatal(err)
	}
	if !replayedPlan.NoChange || len(replayedPlan.Effects) != 0 {
		t.Fatalf("expected semantic no-change replay, got %#v", replayedPlan)
	}
	replayed, err := lifecycle.ApplyManagedTask(context.Background(), replayedPlan.ID, replayedPlan)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Status != engine.WorkApplyNoChange || len(replayed.Receipts) != 0 {
		t.Fatalf("expected effect-free replay, got %#v", replayed)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Receipts) != len(first.Receipts) {
		t.Fatalf("expected prior receipts to survive replay: first=%d status=%d", len(first.Receipts), len(status.Receipts))
	}
}

func newManagedTaskFixture(t *testing.T) (*engine.Engine, *engine.InMemoryWorkAdapter, engine.ManagedTaskRequest, time.Time) {
	t.Helper()
	repository := t.TempDir()
	initializeWorkGit(t, repository)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	adapter := engine.NewInMemoryWorkAdapter(engine.WorkCapability{
		SchemaVersion: 1, Online: true, Fresh: true, Mode: "memory", Actor: "test:maintainer",
		Permissions:           []string{"issues:write", "projects:write", "pull_requests:read"},
		ConfigurationRevision: "project-config:v1", ObservedAt: now, ExpiresAt: now.Add(time.Hour),
	}, engine.WorkObservation{
		SchemaVersion: 1, Revision: "observation:v1", ConfigurationRevision: "project-config:v1",
		Target: engine.WorkTarget{
			Host: "memory.local", RepositoryID: "repository:fixture", ProjectID: "project:fixture",
			FieldIDs:  map[string]string{"readiness": "field:readiness", "status": "field:status", "phase": "field:phase"},
			OptionIDs: map[string]string{"readiness:ready": "option:ready", "readiness:blocked": "option:blocked", "status:next": "option:next", "status:done": "option:done", "phase:Phase 0": "option:phase-0", "phase:Phase 1": "option:phase-1", "phase:Phase 2": "option:phase-2", "phase:Phase 3": "option:phase-3", "phase:Phase 4": "option:phase-4", "phase:Phase 5": "option:phase-5", "phase:Phase 6": "option:phase-6", "phase:Phase 7": "option:phase-7", "phase:Phase 8": "option:phase-8"},
		},
	})
	request := engine.ManagedTaskRequest{
		Repository: repository,
		Intent: engine.WorkDesiredIntent{
			SchemaVersion: 1, OperationID: "operation:issue-71", SourceRevision: "issue-71:v1",
			OperatingProfileRevision: "operating-profile:v1",
			InputDigests:             map[string]string{"issue": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			Credential:               engine.WorkCredentialExpectation{Mode: "memory", Actor: "test:maintainer"},
			Target:                   adapter.Observation().Target,
			Task: engine.DesiredManagedTask{
				ManagedID: "issue:71", IssueType: "task", Title: "Manage one task deterministically through the lifecycle engine",
				Readiness: "ready", Status: "next", Phase: "Phase 3", PhaseAssignmentReason: "Cross-cutting Work Manager delivery is sequenced in Phase 3.",
				Review: []engine.WorkReviewRequirement{{Role: "change-review", DistinctContext: true}},
			},
		},
	}
	return engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter)), adapter, request, now
}

func initializeWorkGit(t *testing.T, repository string) {
	t.Helper()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize managed-task Git repository: %v: %s", err, output)
	}
}

func TestManagedTaskApplyRejectsChangedGovernedSource(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	request.Intent.SourceRevision = "issue-71:v2"
	request.Intent.InputDigests["issue"] = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err != nil {
		t.Fatal(err)
	}

	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("expected changed governed source to reject retained plan")
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "stale" {
		t.Fatalf("stale source must remain explicit, got %#v", status)
	}
}

func TestManagedTaskApplyRejectsChangedObservation(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	changed := adapter.Observation()
	changed.Revision = "observation:v2"
	adapter.SetObservation(changed)

	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("expected changed adapter observation to reject retained plan")
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "stale" {
		t.Fatalf("observation drift must remain explicit, got %#v", status)
	}
}

func TestManagedTaskApplyRejectsChangedCapability(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	capability, _ := adapter.Capability(context.Background())
	capability.Permissions = append(capability.Permissions, "metadata:read")
	adapter.SetCapability(capability)
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("expected changed capability manifest to reject retained plan")
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "stale" {
		t.Fatalf("changed capability must remain explicit stale state, got %#v", status)
	}
}

func TestManagedTaskApplyRejectsChangedOperatingProfile(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	request.Intent.OperatingProfileRevision = "operating-profile:v2"
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("expected operating-profile change to reject retained plan")
	}
}

func TestManagedTaskConfigurationMigrationRequiresNewGovernedInput(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	oldPlan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	observation := adapter.Observation()
	observation.Revision = "observation:v2"
	observation.ConfigurationRevision = "project-config:v2"
	observation.Target.OptionIDs["readiness:ready"] = "option:ready:v2"
	adapter.SetObservation(observation)
	capability, _ := adapter.Capability(context.Background())
	capability.ConfigurationRevision = "project-config:v2"
	capability.ObservedAt = now.Add(time.Minute)
	adapter.SetCapability(capability)
	if _, err := lifecycle.ApplyManagedTask(context.Background(), oldPlan.ID, oldPlan); err == nil {
		t.Fatal("configuration migration must invalidate the retained plan")
	}

	request.Intent.SourceRevision = "issue-71:v2"
	request.Intent.Target = observation.Target
	inspection, err = lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	newPlan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if newPlan.ID == oldPlan.ID || newPlan.ConfigurationRevision != "project-config:v2" {
		t.Fatalf("expected a new governed plan bound to migrated IDs, got %#v", newPlan)
	}
}

func TestManagedTaskDeniedAuthorityPersistsExplicitRecovery(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "denied", Attempt: 1, Recoverable: true, Detail: "minimum Project write authority was denied"}, false)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	result, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyNonPass || len(result.Receipts) != 1 || result.Receipts[0].Outcome != "denied" {
		t.Fatalf("expected explicit denied receipt, got %#v", result)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "denied" || len(status.Recovery) == 0 {
		t.Fatalf("expected durable denied recovery, got %#v", status)
	}
}

func TestManagedTaskAmbiguousCreateReconcilesWithoutDuplicate(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "ambiguous", Attempt: 1, Recoverable: true, Detail: "response lost after effect"}, true)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	result, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyNonPass || result.Receipts[0].Outcome != "ambiguous" {
		t.Fatalf("expected ambiguous receipt, got %#v", result)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "ambiguous" {
		t.Fatalf("expected explicit ambiguous state, got %#v", status)
	}

	recoveredInspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	recoveredPlan, err := lifecycle.PlanManagedTask(context.Background(), recoveredInspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveredPlan.Effects) != 1 || recoveredPlan.Effects[0].Kind != "reconcile-task" {
		t.Fatalf("stable-marker observation should prevent duplicate create and retain only reconciliation, got %#v", recoveredPlan)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), recoveredPlan.ID, recoveredPlan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState != engine.ControlPass {
		t.Fatalf("expected recovered task to verify, got %#v", verification)
	}
}

func TestManagedTaskUnresolvedAmbiguousCreateCannotPlanDuplicate(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "ambiguous", Attempt: 1, Recoverable: true, Detail: "response lost and marker lookup is not yet conclusive"}, false)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	unresolved, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if unresolved.Disposition != "ambiguous" {
		t.Fatalf("expected unresolved marker lookup to remain ambiguous, got %#v", unresolved)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), unresolved); err == nil {
		t.Fatal("unresolved ambiguous create must not produce a duplicate create plan")
	}
}

func TestManageTaskPreservesExplicitDeniedDisposition(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "denied", Attempt: 1, Recoverable: true, Detail: "Project write denied"}, false)
	journey, err := lifecycle.ManageTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if journey.Apply.Status != engine.WorkApplyNonPass || journey.Verification.OverallState == engine.ControlPass || journey.Status.Disposition != "denied" {
		t.Fatalf("composite request erased explicit denied state: %#v", journey)
	}
}

func TestManageTaskPreservesExplicitAdapterDispositions(t *testing.T) {
	for _, outcome := range []string{"unauthenticated", "not-found", "validation-failed", "offline", "failed"} {
		outcome := outcome
		t.Run(outcome, func(t *testing.T) {
			t.Parallel()
			lifecycle, adapter, request, _ := newManagedTaskFixture(t)
			adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: outcome, Attempt: 1, Recoverable: true, Detail: outcome + " effect"}, false)
			journey, err := lifecycle.ManageTask(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if journey.Status.Disposition != outcome || journey.Verification.OverallState == engine.ControlPass {
				t.Fatalf("composite request erased %s: %#v", outcome, journey)
			}
		})
	}
}

func TestManagedTaskPartialSuccessResumesOnlyRemainingEffect(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "applied", Attempt: 1, Detail: "issue created"}, true)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "denied", Attempt: 1, Recoverable: true, Detail: "Project write denied"}, false)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 2 {
		t.Fatalf("expected create and reconcile effects, got %#v", plan.Effects)
	}
	result, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyNonPass || len(result.Receipts) != 2 || result.Receipts[0].Outcome != "applied" || result.Receipts[1].Outcome != "denied" {
		t.Fatalf("expected retained partial receipts, got %#v", result)
	}

	lifecycle = engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))
	recoveryInspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	recoveryPlan, err := lifecycle.PlanManagedTask(context.Background(), recoveryInspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveryPlan.Effects) != 1 || recoveryPlan.Effects[0].Kind != "reconcile-task" {
		t.Fatalf("expected only the remaining reconciliation effect, got %#v", recoveryPlan.Effects)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), recoveryPlan.ID, recoveryPlan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState != engine.ControlPass {
		t.Fatalf("expected resumed task to converge, got %#v", verification)
	}
}

func TestManagedTaskOfflineIntentRequiresFreshReconnectHandshake(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	offline := engine.WorkCapability{
		SchemaVersion: 1, Online: false, Fresh: false, Mode: "memory", Actor: "test:maintainer",
		Permissions:           []string{"issues:write", "projects:write", "pull_requests:read"},
		ConfigurationRevision: "project-config:v1", ObservedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	adapter.SetCapability(offline)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != "queued-offline" {
		t.Fatalf("expected credential-free offline queue, got %#v", inspection)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
		t.Fatal("offline intent must not produce an effect plan")
	}

	reconnected := offline
	reconnected.Online = true
	adapter.SetCapability(reconnected)
	inspection, err = lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != "handshake-required" {
		t.Fatalf("reconnect without freshness must require handshake, got %#v", inspection)
	}
	reconnected.Fresh = true
	reconnected.ObservedAt = now.Add(time.Minute)
	adapter.SetCapability(reconnected)
	inspection, err = lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != "inspected" {
		t.Fatalf("fresh matching handshake should permit planning, got %#v", inspection)
	}
}

func TestManagedTaskExpiredPlanCannotApply(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	expiredEngine := engine.New(engine.WithClock(fixedWorkClock{now.Add(2 * time.Hour)}), engine.WithWorkAdapter(adapter))
	if _, err := expiredEngine.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("expected expired managed-task plan to reject effects")
	}
	status, err := expiredEngine.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "stale" || len(status.Recovery) == 0 {
		t.Fatalf("expected explicit expired-plan recovery, got %#v", status)
	}
}

func TestManagedTaskRateLimitPersistsBoundedRetryUntilReset(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{
		Outcome: "rate-limited", Attempt: 1, Recoverable: true, Detail: "secondary limit",
		Retry: &engine.WorkRetryState{Attempt: 1, MaxAttempts: 2, RetryAt: now.Add(time.Minute), ResetAt: now.Add(10 * time.Minute)},
	}, false)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "retry-pending" || status.Retry == nil || status.Retry.Attempt != 1 {
		t.Fatalf("expected retained first retry state, got %#v", status)
	}

	secondNow := now.Add(2 * time.Minute)
	capability, _ := adapter.Capability(context.Background())
	capability.ObservedAt = secondNow
	capability.ExpiresAt = secondNow.Add(time.Hour)
	adapter.SetCapability(capability)
	secondEngine := engine.New(engine.WithClock(fixedWorkClock{secondNow}), engine.WithWorkAdapter(adapter))
	inspection, err = secondEngine.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err = secondEngine.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	adapter.QueueApplyResult(engine.WorkEffectResult{
		Outcome: "rate-limited", Attempt: 2, Recoverable: true, Detail: "bounded retry exhausted",
		Retry: &engine.WorkRetryState{Attempt: 2, MaxAttempts: 2, RetryAt: secondNow.Add(time.Minute), ResetAt: now.Add(10 * time.Minute)},
	}, false)
	if _, err := secondEngine.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	status, err = secondEngine.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "retry-exhausted" || status.Retry == nil || status.Retry.Attempt != 2 {
		t.Fatalf("expected bounded retry exhaustion, got %#v", status)
	}
	blockedInspection, err := secondEngine.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if blockedInspection.Disposition != "retry-exhausted" {
		t.Fatalf("retry exhaustion must block until reset, got %#v", blockedInspection)
	}

	resetNow := now.Add(11 * time.Minute)
	capability.ObservedAt = resetNow
	capability.ExpiresAt = resetNow.Add(time.Hour)
	adapter.SetCapability(capability)
	resetEngine := engine.New(engine.WithClock(fixedWorkClock{resetNow}), engine.WithWorkAdapter(adapter))
	resetInspection, err := resetEngine.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if resetInspection.Disposition != "inspected" {
		t.Fatalf("recorded reset should permit a fresh plan, got %#v", resetInspection)
	}
}

func TestManagedTaskPolicyDerivesReadinessPhaseReviewAndCompletion(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "in-progress",
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:72", Status: "backlog", Closed: false}},
	}
	request.Intent.Task.Phase = ""
	request.Intent.Task.ParentPhase = "Phase 3"
	request.Intent.Task.Readiness = "blocked"
	request.Intent.Task.Blockers = []engine.WorkDependency{{ManagedID: "issue:64", Closed: true}}
	request.Intent.Task.Status = "next"
	request.Intent.Task.Closed = true
	request.Intent.Task.PromotionRecord = "docs/decisions/DEC-0013-question-and-research-work.md"
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) == 0 {
		t.Fatal("expected derived task effects")
	}
	derived := plan.Effects[len(plan.Effects)-1].Desired
	if derived.Readiness != "ready" || derived.Status != "done" || derived.Phase != "" {
		t.Fatalf("unexpected derived lifecycle facts: %#v", derived)
	}
	if plan.DerivedFacts.Phase != "Phase 3" || plan.DerivedFacts.PhaseSource != "parent" {
		t.Fatalf("expected parent-derived Phase context without a copied assignment, got %#v", plan.DerivedFacts)
	}
	if len(derived.Review) != 1 || !derived.Review[0].DistinctContext {
		t.Fatalf("review requirement must remain distinct: %#v", derived.Review)
	}
	if plan.DerivedFacts.ParentStatus != "in-progress" || plan.DerivedFacts.ParentClosed || plan.DerivedFacts.Completion != "complete" || plan.DerivedFacts.PromotionRecord != request.Intent.Task.PromotionRecord {
		t.Fatalf("expected incomplete parent and separate promoted completion facts, got %#v", plan.DerivedFacts)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState != engine.ControlPass {
		t.Fatalf("expected derived policy to verify, got %#v", verification)
	}
}

func TestManagedTaskPlansDirectPhaseAssignmentByImmutableOption(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready", StatusOption: "option:next",
	}
	observation.Revision = "observation:missing-phase"
	adapter.SetObservation(observation)
	lifecycle = engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))

	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || !slices.Contains(plan.Effects[0].Operations, "phase") {
		t.Fatalf("expected one immutable Phase correction, got %#v", plan.Effects)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("direct Phase assignment did not converge: %#v, %v", verification, err)
	}
}

func TestManagedTaskRequiresReasonForCrossCuttingDirectPhase(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentPhase = "Phase 3"
	request.Intent.Task.Phase = "Phase 5"
	request.Intent.Task.PhaseAssignmentReason = ""
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil || !strings.Contains(err.Error(), "cross-cutting") {
		t.Fatalf("expected unjustified cross-cutting Phase to stop, got %v", err)
	}
	request.Intent.Task.PhaseAssignmentReason = "Shared qualification work is sequenced with Phase 5."
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err != nil {
		t.Fatalf("expected justified cross-cutting Phase to inspect: %v", err)
	}
}

func TestManagedTaskRejectsInvalidDuplicatedAndStalePhaseInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(*engine.ManagedTaskRequest)
		want    string
	}{
		{name: "unsupported phase", prepare: func(request *engine.ManagedTaskRequest) {
			request.Intent.Task.Phase = "Phase 9"
		}, want: "unsupported roadmap Phase"},
		{name: "orphan parent phase", prepare: func(request *engine.ManagedTaskRequest) {
			request.Intent.Task.Phase = ""
			request.Intent.Task.PhaseAssignmentReason = ""
			request.Intent.Task.ParentPhase = "Phase 3"
		}, want: "native parent identity"},
		{name: "duplicated parent assignment", prepare: func(request *engine.ManagedTaskRequest) {
			request.Intent.Task.ParentManagedID = "issue:4"
			request.Intent.Task.ParentPhase = "Phase 3"
		}, want: "derive Phase from its parent"},
		{name: "stale option identity", prepare: func(request *engine.ManagedTaskRequest) {
			delete(request.Intent.Target.OptionIDs, "phase:Phase 3")
		}, want: "immutable Phase field or option identity"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lifecycle, _, request, _ := newManagedTaskFixture(t)
			test.prepare(&request)
			if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q rejection, got %v", test.want, err)
			}
		})
	}
}

func TestManagedTaskBindsInheritedPhaseToNativeParentObservation(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.Phase = ""
	request.Intent.Task.PhaseAssignmentReason = ""
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentPhase = "Phase 3"
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ParentManagedID: "issue:4",
		ReadinessOption: "option:ready", StatusOption: "option:next",
	}
	observation.Revision = "observation:unbound-parent-phase"
	adapter.SetObservation(observation)
	lifecycle = engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))

	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != "non-pass" || !strings.Contains(strings.Join(inspection.Problems, " "), "native parent") {
		t.Fatalf("caller-supplied parent Phase must not bypass native observation: %#v", inspection)
	}

	observation.Task.NativeParentManagedID = "issue:4"
	observation.Task.ParentPhaseOption = "option:phase-3"
	observation.Revision = "observation:bound-parent-phase"
	adapter.SetObservation(observation)
	lifecycle = engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))
	inspection, err = lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil || inspection.Disposition != "inspected" {
		t.Fatalf("native parent Phase observation should satisfy the binding: %#v, %v", inspection, err)
	}
}

func TestManagedTaskUnblockedReadinessDoesNotSelectStatus(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.Readiness = "blocked"
	request.Intent.Task.Status = "next"
	request.Intent.Task.Blockers = []engine.WorkDependency{{ManagedID: "issue:64", Closed: true}}
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	derived := plan.Effects[len(plan.Effects)-1].Desired
	if derived.Readiness != "ready" || derived.Status != "next" {
		t.Fatalf("unblocking may promote readiness but must preserve independently selected status: %#v", derived)
	}
}

func TestManagedQuestionCompletionRequiresPromotionResolution(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.IssueType = "question"
	request.Intent.Task.Closed = true
	request.Intent.Task.Status = "done"
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("closed question without promotion resolution must be rejected")
	}
	request.Intent.Task.PromotionRecord = "docs/decisions/DEC-0013-question-and-research-work.md"
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if plan.DerivedFacts.Completion != "complete" || plan.DerivedFacts.PromotionRecord != request.Intent.Task.PromotionRecord {
		t.Fatalf("question completion must retain its distinct promotion route: %#v", plan.DerivedFacts)
	}
}

func TestManagedTaskInspectionRejectsDifferentObservedManagedIdentity(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{ManagedID: "issue:other", IssueNodeID: "memory:issue:other", Title: "Other task", IssueType: "task"}
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition == "inspected" {
		t.Fatalf("different managed identity must be non-pass, got %#v", inspection)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
		t.Fatal("different observed managed identity must not produce a plan")
	}
}

func TestManagedTaskRejectsUnsupportedIntentSchemaWithoutState(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.SchemaVersion = 2
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("expected unsupported intent schema rejection")
	}
	if _, err := os.Stat(filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("rejected input must not create state, got %v", err)
	}
}

func TestManagedTaskRejectsSecretShapedProvenanceWithoutState(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.InputDigests["issue"] = "ghp_1234567890abcdefghijklmnopqrstuvwxyz"
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("expected secret-shaped provenance rejection")
	}
	if _, err := os.Stat(filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("rejected secret-shaped input must not create state, got %v", err)
	}
}

func TestManagedTaskStatusRejectsTamperedDurableState(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Replace(content, []byte(`"disposition": "planned"`), []byte(`"disposition": "applied"`), 1)
	if bytes.Equal(content, tampered) {
		t.Fatal("fixture did not locate durable disposition")
	}
	if err := os.WriteFile(path, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository); err == nil {
		t.Fatal("expected tampered durable state to fail closed")
	}
}

func TestManageTaskRequestReturnsCompleteLifecycleJourney(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	journey, err := lifecycle.ManageTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if journey.Inspection.ID == "" || journey.Plan.ID == "" || journey.Apply.Status != engine.WorkApplyApplied || journey.Verification.OverallState != engine.ControlPass || journey.Status.Disposition != "converged" {
		t.Fatalf("expected complete lifecycle journey, got %#v", journey)
	}
}

func TestManagedTaskStateRejectsReservedDirectorySymlink(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(request.Repository, ".starter-kit")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("expected symlinked managed-task state path rejection")
	}
	if _, err := os.Stat(filepath.Join(outside, "work-manager", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("managed-task state escaped repository through symlink: %v", err)
	}
}

func TestManagedTaskVerifyRequiresFreshMatchingCapability(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	capability, _ := adapter.Capability(context.Background())
	capability.Online = false
	capability.Fresh = false
	adapter.SetCapability(capability)
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState == engine.ControlPass {
		t.Fatalf("offline stale capability must not verify as pass: %#v", verification)
	}
}

func TestManagedTaskApplyHonorsRepositoryLifecycleLease(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, now := newManagedTaskFixture(t)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	gitDirectoryOutput, err := exec.Command("git", "-C", request.Repository, "rev-parse", "--absolute-git-dir").Output()
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(strings.TrimSpace(string(gitDirectoryOutput)), "starter-kit.lock")
	lease := fmt.Sprintf("{\"schema_version\":1,\"token\":\"%032x\",\"plan_id\":%q,\"pid\":%d,\"created_at\":%q}\n", 1, plan.ID, os.Getpid(), now.Format(time.RFC3339Nano))
	if err := os.WriteFile(lockPath, []byte(lease), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(lockPath) })
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("expected active repository lifecycle lease to serialize managed-task apply")
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Receipts) != 0 {
		t.Fatalf("locked apply must not attempt effects: %#v", status)
	}
}

func TestManagedTaskReceiptRedactsSecretShapedAdapterDetail(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	secret := "ghp_1234567890abcdefghijklmnopqrstuvwxyz"
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "denied", Attempt: 1, Recoverable: true, Detail: "denied token " + secret}, false)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	result, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Receipts[0].Detail, secret) {
		t.Fatalf("secret-shaped adapter detail escaped receipt redaction: %q", result.Receipts[0].Detail)
	}
	content, err := os.ReadFile(filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(content, []byte(secret)) {
		t.Fatal("secret-shaped adapter detail persisted in durable state")
	}
}

func TestManagedTaskRejectsSecretShapedObservationWithoutState(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	secret := "ghp_1234567890abcdefghijklmnopqrstuvwxyz"
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{ManagedID: "issue:71", Title: secret, IssueType: "task"}
	adapter.SetObservation(observation)
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("expected secret-shaped normalized observation rejection")
	} else if strings.Contains(err.Error(), secret) {
		t.Fatalf("rejection diagnostic echoed secret: %v", err)
	}
	if _, err := os.Stat(filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("rejected observation must not create state, got %v", err)
	}
}

func TestManagedTaskInvalidAdapterResultBecomesNeedsReview(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "mystery", Attempt: 0, Detail: "unversioned adapter result"}, false)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	result, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan)
	if err == nil {
		t.Fatal("invalid adapter result must fail the accepted apply")
	}
	if result.Status != engine.WorkApplyNonPass || len(result.Receipts) != 1 || result.Receipts[0].Outcome != "needs-review" {
		t.Fatalf("invalid result must persist an explicit needs-review receipt: %#v", result)
	}
	status, statusErr := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if statusErr != nil {
		t.Fatal(statusErr)
	}
	if status.Disposition != "needs-review" {
		t.Fatalf("invalid adapter result must remain needs-review, got %#v", status)
	}
}
