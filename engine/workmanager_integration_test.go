package engine_test

import (
	"bytes"
	"context"
	"crypto/sha256"
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
		Permissions:           []string{"issues:write", "projects:write", "pull_requests:read", "contents:read"},
		ConfigurationRevision: "project-config:v1",
		ObservedAt:            now,
		ExpiresAt:             now.Add(time.Hour),
	}, engine.WorkObservation{
		SchemaVersion:         1,
		Revision:              "observation:v1",
		ConfigurationRevision: "project-config:v1",
		Relationships:         engine.WorkRelationshipObservation{Observed: true},
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
		Permissions:           []string{"issues:write", "projects:write", "pull_requests:read", "contents:read"},
		ConfigurationRevision: "project-config:v1", ObservedAt: now, ExpiresAt: now.Add(time.Hour),
	}, engine.WorkObservation{
		SchemaVersion: 1, Revision: "observation:v1", ConfigurationRevision: "project-config:v1",
		Relationships: engine.WorkRelationshipObservation{Observed: true},
		Target: engine.WorkTarget{
			Host: "memory.local", RepositoryID: "repository:fixture", ProjectID: "project:fixture",
			FieldIDs:  map[string]string{"readiness": "field:readiness", "status": "field:status", "phase": "field:phase"},
			OptionIDs: map[string]string{"readiness:ready": "option:ready", "readiness:blocked": "option:blocked", "status:backlog": "option:backlog", "status:next": "option:next", "status:in-progress": "option:in-progress", "status:done": "option:done", "phase:Phase 0": "option:phase-0", "phase:Phase 1": "option:phase-1", "phase:Phase 2": "option:phase-2", "phase:Phase 3": "option:phase-3", "phase:Phase 4": "option:phase-4", "phase:Phase 5": "option:phase-5", "phase:Phase 6": "option:phase-6", "phase:Phase 7": "option:phase-7", "phase:Phase 8": "option:phase-8"},
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

func observeIntentRelationships(observation *engine.WorkObservation, task engine.DesiredManagedTask) {
	observation.Relationships = engine.WorkRelationshipObservation{Observed: true, Blockers: slices.Clone(task.Blockers)}
	if observation.Task != nil {
		observation.Task.Phase = task.Phase
		observation.Task.PhaseAssignmentReason = task.PhaseAssignmentReason
		if task.Phase != "" {
			observation.Task.PhaseOption = observation.Target.OptionIDs["phase:"+task.Phase]
		}
		if task.ParentPhase != "" {
			observation.Task.NativeParentManagedID = task.ParentManagedID
			observation.Task.ParentPhaseOption = observation.Target.OptionIDs["phase:"+task.ParentPhase]
		}
	}
	if task.ParentContext != nil {
		observation.Relationships.ParentManagedID = task.ParentContext.ManagedID
		observation.Relationships.OtherChildren = slices.Clone(task.ParentContext.OtherChildren)
	}
	for _, dependent := range task.Dependents {
		observation.Relationships.Dependents = append(observation.Relationships.Dependents, engine.WorkObservedDependent{
			ManagedID: dependent.ManagedID,
			Blockers:  slices.Clone(dependent.Blockers),
		})
	}
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

func TestManagedTaskApplyRejectsRelationshipOnlyObservationChange(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.Readiness = "blocked"
	request.Intent.Task.Blockers = []engine.WorkDependency{{ManagedID: "issue:64", Closed: false}}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", BlockedBy: []string{"issue:64"},
		ReadinessOption: "option:blocked", StatusOption: "option:next", Phase: "Phase 3",
		Review: request.Intent.Task.Review,
	}
	observation.Relationships = engine.WorkRelationshipObservation{Observed: true, Blockers: []engine.WorkDependency{{ManagedID: "issue:64", Closed: false}}}
	observation.Revision = "observation:blocker-open"
	adapter.SetObservation(observation)

	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	changed := adapter.Observation()
	changed.Relationships.Blockers[0].Closed = true
	changed.Revision = "observation:blocker-closed"
	adapter.SetObservation(changed)

	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("relationship-only observation drift must reject the retained plan")
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != "stale" || len(status.Receipts) != 0 {
		t.Fatalf("relationship-only drift must stop before effects: %#v", status)
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
		Permissions:           []string{"issues:write", "projects:write", "pull_requests:read", "contents:read"},
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

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
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
	observation := adapter.Observation()
	observation.RelatedTasks = []engine.WorkObservedTask{{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", StatusOption: "option:in-progress", PhaseOption: "option:phase-3"}}
	observation.Revision = "observation:parent-context"
	observeIntentRelationships(&observation, request.Intent.Task)
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

func TestManagedTaskReconcilesClosedItemParentAndFinalUnblockedDependent(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.Closed = true
	request.Intent.Task.Status = "next"
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "backlog", CompletionSatisfied: true,
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "backlog", Closed: false}},
	}
	request.Intent.Task.Dependents = []engine.WorkDependentContext{{
		ManagedID: "issue:74", Readiness: "blocked", Status: "backlog", ReadyEligible: true,
		Blockers: []engine.WorkDependency{{ManagedID: "issue:71", Closed: true}, {ManagedID: "issue:46", Closed: true}},
	}}
	request.Intent.Target.OptionIDs["status:backlog"] = "option:backlog"
	request.Intent.Target.OptionIDs["status:in-progress"] = "option:in-progress"

	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready",
		StatusOption: "option:next", Closed: true, Review: request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{
		{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", ReadinessOption: "option:ready", StatusOption: "option:backlog"},
		{ManagedID: "issue:74", IssueNodeID: "memory:issue:74", ProjectItemID: "memory:item:74", ReadinessOption: "option:blocked", StatusOption: "option:backlog"},
	}
	observation.Revision = "observation:closed-slice"
	observeIntentRelationships(&observation, request.Intent.Task)
	adapter.SetObservation(observation)

	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 3 {
		t.Fatalf("expected selected, parent, and dependent corrections, got %#v", plan.Effects)
	}
	if plan.Effects[0].ManagedID != "issue:71" || plan.Effects[0].Desired.Status != "done" {
		t.Fatalf("closed selected item did not plan Done: %#v", plan.Effects[0])
	}
	if plan.Effects[1].ManagedID != "issue:4" || plan.Effects[1].Desired.Status != "in-progress" || plan.Effects[1].Desired.Closed {
		t.Fatalf("incomplete parent did not plan In progress: %#v", plan.Effects[1])
	}
	if plan.Effects[2].ManagedID != "issue:74" || plan.Effects[2].Desired.Readiness != "ready" || plan.Effects[2].Desired.Status != "backlog" {
		t.Fatalf("final unblocked dependent changed the wrong fields: %#v", plan.Effects[2])
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState != engine.ControlPass {
		t.Fatalf("expected bounded reconciliation slice to converge, got %#v", verification)
	}
}

func TestManagedTaskClosesParentOnlyWhenEveryChildAndCompletionContractAreComplete(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.Closed = true
	request.Intent.Task.Status = "done"
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "in-progress", CompletionSatisfied: true,
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "done", Closed: true}},
	}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ParentManagedID: "issue:4",
		ReadinessOption: "option:ready", StatusOption: "option:done", Phase: "Phase 3", Closed: true, Review: request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", StatusOption: "option:in-progress"}}
	observation.Revision = "observation:parent-ready-to-close"
	observeIntentRelationships(&observation, request.Intent.Task)
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
	if len(plan.Effects) != 1 || plan.Effects[0].ManagedID != "issue:4" || !plan.Effects[0].Desired.Closed || plan.Effects[0].Desired.Status != "done" {
		t.Fatalf("expected one parent closure correction, got %#v", plan.Effects)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("parent closure did not verify: %#v, %v", verification, err)
	}
	replayInspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	replayPlan, err := lifecycle.PlanManagedTask(context.Background(), replayInspection)
	if err != nil {
		t.Fatal(err)
	}
	if !replayPlan.NoChange || len(replayPlan.Effects) != 0 {
		t.Fatalf("parent closure replay must be idempotent, got %#v", replayPlan)
	}
}

func TestManagedTaskRejectsUnexplainedOpenParentAfterEveryChildCloses(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.Closed = true
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "in-progress",
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "done", Closed: true}},
	}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready",
		StatusOption: "option:done", Closed: true, Review: request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", StatusOption: "option:in-progress"}}
	observeIntentRelationships(&observation, request.Intent.Task)
	observation.Revision = "observation:unexplained-complete-parent"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil || inspection.Disposition != "non-pass" || !slices.ContainsFunc(inspection.Problems, func(problem string) bool { return strings.Contains(problem, "completion contract") }) {
		t.Fatalf("all-children-complete parent must stop with an explicit completion result, got %#v, %v", inspection, err)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
		t.Fatal("non-pass parent completion observation must not produce a plan")
	}
}

func TestManagedTaskDoesNotPromoteDependentUntilEveryBlockerCloses(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.Dependents = []engine.WorkDependentContext{{
		ManagedID: "issue:74", Readiness: "blocked", Status: "backlog", ReadyEligible: true,
		Blockers: []engine.WorkDependency{{ManagedID: "issue:71", Closed: true}, {ManagedID: "issue:46", Closed: false}},
	}}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready",
		StatusOption: "option:next", Phase: "Phase 3", Review: request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{{ManagedID: "issue:74", IssueNodeID: "memory:issue:74", ProjectItemID: "memory:item:74", ReadinessOption: "option:blocked", StatusOption: "option:backlog"}}
	observation.Revision = "observation:dependent-still-blocked"
	observeIntentRelationships(&observation, request.Intent.Task)
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
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("dependent with an open blocker must remain blocked without effects, got %#v", plan)
	}
}

func TestManagedTaskUsesNativeRelationshipObservationInsteadOfCallerFacts(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.Closed = true
	request.Intent.Task.Status = "next"
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "in-progress", CompletionSatisfied: true,
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "done", Closed: true}},
	}
	request.Intent.Task.Dependents = []engine.WorkDependentContext{{
		ManagedID: "issue:74", Readiness: "blocked", Status: "backlog", ReadyEligible: true,
		Blockers: []engine.WorkDependency{{ManagedID: "issue:71", Closed: true}, {ManagedID: "issue:46", Closed: true}},
	}}
	request.Intent.Target.OptionIDs["status:backlog"] = "option:backlog"
	request.Intent.Target.OptionIDs["status:in-progress"] = "option:in-progress"

	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready",
		StatusOption: "option:next", Closed: true, Review: request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{
		{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", ReadinessOption: "option:ready", StatusOption: "option:in-progress"},
		{ManagedID: "issue:74", IssueNodeID: "memory:issue:74", ProjectItemID: "memory:item:74", ReadinessOption: "option:blocked", StatusOption: "option:backlog"},
	}
	observation.Relationships = engine.WorkRelationshipObservation{
		Observed: true, ParentManagedID: "issue:4",
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "backlog", Closed: false}},
		Dependents: []engine.WorkObservedDependent{{
			ManagedID: "issue:74",
			Blockers:  []engine.WorkDependency{{ManagedID: "issue:71", Closed: true}, {ManagedID: "issue:46", Closed: false}},
		}},
	}
	observation.Revision = "observation:native-relationships-win"
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
	if len(plan.Effects) != 1 || plan.Effects[0].ManagedID != "issue:71" || plan.Effects[0].Desired.Status != "done" {
		t.Fatalf("native open sibling/blocker must prevent parent closure and dependent promotion: %#v", plan.Effects)
	}
}

func TestManagedTaskUsesObservedParentLifecycleAsUnstartedBaseline(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "done", Closed: true,
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "done", Closed: true}},
	}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ParentManagedID: "issue:4",
		ReadinessOption: "option:ready", StatusOption: "option:next", Phase: "Phase 3", PhaseOption: "option:phase-3",
		PhaseAssignmentReason: request.Intent.Task.PhaseAssignmentReason,
		Review:                request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{{
		ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4",
		ReadinessOption: "option:ready", StatusOption: "option:backlog", Closed: false,
	}}
	observation.Relationships = engine.WorkRelationshipObservation{
		Observed: true, ParentManagedID: "issue:4",
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "backlog", Closed: false}},
	}
	observation.Revision = "observation:native-parent-baseline"
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
	if !plan.NoChange || len(plan.Effects) != 0 || plan.DerivedFacts.ParentStatus != "backlog" || plan.DerivedFacts.ParentClosed {
		t.Fatalf("stale caller Done/closed facts must not override the native unstarted parent: %#v", plan)
	}
}

func FuzzManagedTaskNativeRelationshipDerivation(f *testing.F) {
	for _, seed := range []struct{ siblingClosed, otherBlockerClosed bool }{{false, false}, {false, true}, {true, false}, {true, true}} {
		f.Add(seed.siblingClosed, seed.otherBlockerClosed)
	}
	f.Fuzz(func(t *testing.T, siblingClosed, otherBlockerClosed bool) {
		lifecycle, adapter, request, now := newManagedTaskFixture(t)
		request.Intent.Task.Closed = true
		request.Intent.Task.ParentManagedID = "issue:4"
		request.Intent.Task.ParentContext = &engine.WorkParentContext{
			ManagedID: "issue:4", Status: "backlog", CompletionSatisfied: true,
			OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "done", Closed: true}},
		}
		request.Intent.Task.Dependents = []engine.WorkDependentContext{{
			ManagedID: "issue:74", Readiness: "blocked", Status: "backlog", ReadyEligible: true,
			Blockers: []engine.WorkDependency{{ManagedID: "issue:71", Closed: true}, {ManagedID: "issue:46", Closed: true}},
		}}
		observation := adapter.Observation()
		observation.Task = &engine.WorkObservedTask{
			ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
			Title: request.Intent.Task.Title, IssueType: "task", ParentManagedID: "issue:4",
			ReadinessOption: "option:ready", StatusOption: "option:done", Closed: true,
			Review: request.Intent.Task.Review,
		}
		observation.RelatedTasks = []engine.WorkObservedTask{
			{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", ReadinessOption: "option:ready", StatusOption: "option:backlog"},
			{ManagedID: "issue:74", IssueNodeID: "memory:issue:74", ProjectItemID: "memory:item:74", ReadinessOption: "option:blocked", StatusOption: "option:backlog"},
		}
		observation.Relationships = engine.WorkRelationshipObservation{
			Observed: true, ParentManagedID: "issue:4",
			OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "backlog", Closed: siblingClosed}},
			Dependents: []engine.WorkObservedDependent{{ManagedID: "issue:74", Blockers: []engine.WorkDependency{
				{ManagedID: "issue:71", Closed: true}, {ManagedID: "issue:46", Closed: otherBlockerClosed},
			}}},
		}
		observation.Revision = fmt.Sprintf("observation:property:%t:%t", siblingClosed, otherBlockerClosed)
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
		parent := planEffectForManagedID(plan, "issue:4")
		dependent := planEffectForManagedID(plan, "issue:74")
		if parent == nil || parent.Desired.Closed != siblingClosed || (parent.Desired.Status == "done") != siblingClosed {
			t.Fatalf("parent property failed for siblingClosed=%t: %#v", siblingClosed, parent)
		}
		if !otherBlockerClosed && dependent != nil || otherBlockerClosed && (dependent == nil || dependent.Desired.Readiness != "ready" || dependent.Desired.Status != "backlog") {
			t.Fatalf("dependent property failed for otherBlockerClosed=%t: %#v", otherBlockerClosed, dependent)
		}
	})
}

func planEffectForManagedID(plan engine.WorkPlan, managedID string) *engine.WorkEffect {
	for index := range plan.Effects {
		if plan.Effects[index].ManagedID == managedID {
			return &plan.Effects[index]
		}
	}
	return nil
}

func TestManagedTaskRejectsDirectDependentCycle(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.Task.Dependents = []engine.WorkDependentContext{{
		ManagedID: "issue:74", Readiness: "blocked", Status: "backlog", ReadyEligible: true,
		Blockers: []engine.WorkDependency{
			{ManagedID: request.Intent.Task.ManagedID, Closed: true},
			{ManagedID: "issue:74", Closed: true},
		},
	}}

	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil || !strings.Contains(err.Error(), "dependent blocker context") {
		t.Fatalf("expected direct dependency cycle to stop inspection, got %v", err)
	}
	if _, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository); err == nil {
		t.Fatal("rejected dependency cycle must not create durable state")
	}
}

func TestManagedTaskRestoresLifecycleStateForNativelyReopenedIssue(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.Status = "backlog"
	request.Intent.Target.OptionIDs["status:backlog"] = "option:backlog"
	observation := adapter.Observation()
	observation.Target = request.Intent.Target
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready",
		StatusOption: "option:done", Phase: "Phase 3", Review: request.Intent.Task.Review, Closed: false,
	}
	observation.Revision = "observation:stale-closed"
	observeIntentRelationships(&observation, request.Intent.Task)
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
	if len(plan.Effects) != 1 || !slices.Equal(plan.Effects[0].Operations, []string{"status"}) {
		t.Fatalf("expected explicit Status restoration without rewriting native issue state, got %#v", plan.Effects)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("reopened task did not converge: %#v, %v", verification, err)
	}
}

func TestManagedTaskRelatedPartialFailureResumesOnlyUnconvergedCorrections(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	request.Intent.Task.Closed = true
	request.Intent.Task.Status = "next"
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{
		ManagedID: "issue:4", Status: "backlog",
		OtherChildren: []engine.WorkRelatedTask{{ManagedID: "issue:46", Status: "backlog"}},
	}
	request.Intent.Task.Dependents = []engine.WorkDependentContext{{
		ManagedID: "issue:74", Readiness: "blocked", Status: "backlog", ReadyEligible: true,
		Blockers: []engine.WorkDependency{{ManagedID: "issue:71", Closed: true}},
	}}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready",
		StatusOption: "option:next", Closed: true, Review: request.Intent.Task.Review,
	}
	observation.RelatedTasks = []engine.WorkObservedTask{
		{ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", ReadinessOption: "option:ready", StatusOption: "option:backlog"},
		{ManagedID: "issue:74", IssueNodeID: "memory:issue:74", ProjectItemID: "memory:item:74", ReadinessOption: "option:blocked", StatusOption: "option:backlog"},
	}
	observation.Revision = "observation:partial-related-slice"
	observeIntentRelationships(&observation, request.Intent.Task)
	adapter.SetObservation(observation)
	lifecycle = engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "applied", Attempt: 1, Detail: "selected item corrected"}, true)
	adapter.QueueApplyResult(engine.WorkEffectResult{Outcome: "denied", Attempt: 1, Recoverable: true, Detail: "parent Project write denied"}, false)

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
	if result.Status != engine.WorkApplyNonPass || len(result.Receipts) != 2 || result.Receipts[1].Outcome != "denied" {
		t.Fatalf("expected retained selected receipt and denied parent receipt, got %#v", result)
	}
	if result.Receipts[1].After != result.Receipts[1].Before {
		t.Fatalf("denied correction must not claim its desired after-state: %#v", result.Receipts[1])
	}

	restarted := engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(adapter))
	recoveryInspection, err := restarted.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	recoveryPlan, err := restarted.PlanManagedTask(context.Background(), recoveryInspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveryPlan.Effects) != 2 || recoveryPlan.Effects[0].ManagedID != "issue:4" || recoveryPlan.Effects[1].ManagedID != "issue:74" {
		t.Fatalf("recovery must omit the converged selected item and retain related deltas, got %#v", recoveryPlan.Effects)
	}
	if _, err := restarted.ApplyManagedTask(context.Background(), recoveryPlan.ID, recoveryPlan); err != nil {
		t.Fatal(err)
	}
	verification, err := restarted.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("related recovery did not converge: %#v, %v", verification, err)
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
		{name: "duplicate field identity", prepare: func(request *engine.ManagedTaskRequest) {
			request.Intent.Target.FieldIDs["phase"] = request.Intent.Target.FieldIDs["status"]
		}, want: "duplicate immutable field or option identities"},
		{name: "duplicate option identity", prepare: func(request *engine.ManagedTaskRequest) {
			request.Intent.Target.OptionIDs["phase:Phase 0"] = request.Intent.Target.OptionIDs["phase:Phase 1"]
		}, want: "duplicate immutable field or option identities"},
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
	request.Intent.Task.ParentContext = &engine.WorkParentContext{ManagedID: "issue:4", Status: "backlog", OtherChildren: []engine.WorkRelatedTask{}}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ParentManagedID: "issue:4",
		ReadinessOption: "option:ready", StatusOption: "option:next",
	}
	observation.RelatedTasks = []engine.WorkObservedTask{{
		ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4",
		ReadinessOption: "option:ready", StatusOption: "option:backlog",
	}}
	observation.Relationships = engine.WorkRelationshipObservation{Observed: true, ParentManagedID: "issue:4", OtherChildren: []engine.WorkRelatedTask{}}
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

	observation.RelatedTasks[0].PhaseOption = "option:phase-3"
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
	request.Intent.Task.NoPromotionRequired = true
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("question cannot claim both promotion and no-promotion resolution")
	}
	request.Intent.Task.NoPromotionRequired = false
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
	foundBacklink := false
	for _, effect := range plan.Effects {
		foundBacklink = foundBacklink || slices.Contains(effect.Operations, "promotion-link")
	}
	if !foundBacklink {
		t.Fatal("question completion omitted its issue-side promotion backlink")
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("question promotion backlink did not verify: %#v, %v", verification, err)
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

func TestManagedTaskInitializedStateDeletionOrCorruptionFailsClosed(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		mutate func(string) error
	}{
		{name: "deleted", mutate: os.Remove},
		{name: "corrupt", mutate: func(path string) error { return os.WriteFile(path, []byte("corrupt\n"), 0o600) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lifecycle, _, request, _ := newManagedTaskFixture(t)
			if _, err := lifecycle.InspectManagedTask(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			statePath := filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json")
			if err := test.mutate(statePath); err != nil {
				t.Fatal(err)
			}
			if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
				t.Fatal("initialized evidence loss reset managed-task authority state")
			}
		})
	}
}

func TestManagedTaskRejectsUnsupportedIntentSchemaWithoutState(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	request.Intent.SchemaVersion = 3
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("expected unsupported intent schema rejection")
	}
	if _, err := os.Stat(filepath.Join(request.Repository, ".starter-kit", "work-manager", "state.json")); !os.IsNotExist(err) {
		t.Fatalf("rejected input must not create state, got %v", err)
	}
}

func TestGovernedManagedTaskQualifiesFreshContractAndBindsPlan(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != "inspected" || inspection.Qualification == nil || inspection.Qualification.Assessment.Disposition != engine.WorkFreshnessFresh || inspection.Qualification.ID == "" {
		t.Fatalf("expected fresh governed-work qualification, got %#v", inspection)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if plan.QualificationID != inspection.Qualification.ID || plan.DerivedFacts.Freshness != engine.WorkFreshnessFresh || len(plan.Effects) == 0 {
		t.Fatalf("plan did not bind governed-work qualification: %#v", plan)
	}
	for _, effect := range plan.Effects {
		if effect.QualificationID != plan.QualificationID {
			t.Fatalf("effect did not bind governed-work qualification: %#v", effect)
		}
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	verification, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("governed work did not verify: %#v, %v", verification, err)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil || status.QualificationID != plan.QualificationID || status.Freshness != engine.WorkFreshnessFresh {
		t.Fatalf("status lost governed-work qualification: %#v, %v", status, err)
	}
}

func TestGovernedManagedTaskExternalEffectsRequireContainedMandate(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, now := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	capability, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability.Mode = "user-token"
	capability.Actor = "owner"
	adapter.SetCapability(capability)
	request.Intent.Credential = engine.WorkCredentialExpectation{Mode: "user-token", Actor: "owner"}
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err == nil {
		t.Fatal("external effects applied without a DEC-0022 mandate")
	}
	mandate := engine.BindWorkExecutionMandate(engine.WorkExecutionMandate{
		SchemaVersion: 1, ApprovedBy: "owner", ApprovalID: "issue-74-owner-approval", ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
		Target: plan.Target, OperationID: plan.OperationID, SelectedManagedID: request.Intent.Task.ManagedID, Actors: []string{"owner"}, CredentialModes: []string{"user-token"}, Permissions: capability.Permissions,
		OperatingProfileRevisions: []string{plan.OperatingProfileRevision}, ContractDigests: []string{engine.ExecutableIssueContractDigest(request.Intent.Governance.Issue)},
		GovernanceDigests: []string{engine.GovernedWorkContractDigest(*request.Intent.Governance)},
		InputDigests:      plan.InputDigests, GovernedSourceDigests: inspection.Qualification.Assessment.SourceDigests,
		SourceRevisions: []string{plan.SourceRevision}, ManagedIDs: []string{request.Intent.Task.ManagedID}, EffectKinds: []string{"create-task", "reconcile-task"},
		Operations: []string{"issue", "project", "readiness", "status", "phase"}, ResourceDigests: []string{engine.ManagedTaskResourceDigest(request.Intent.Task)}, MaxEffects: len(plan.Effects), DataClass: "public-project-metadata",
		CostCeiling: "zero-dollar", Destructive: "no-delete", Retention: "repository-evidence", RecoveryOwner: "owner",
	})
	wrongResource := mandate
	wrongResource.ID = ""
	wrongResource.ResourceDigests = []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	wrongResource = engine.BindWorkExecutionMandate(wrongResource)
	if _, err := lifecycle.ApplyManagedTaskWithMandate(context.Background(), plan.ID, plan, wrongResource); err == nil {
		t.Fatal("external effects applied under a mandate for different desired resources")
	}
	wrongGovernance := mandate
	wrongGovernance.ID = ""
	wrongGovernance.GovernanceDigests = []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	wrongGovernance = engine.BindWorkExecutionMandate(wrongGovernance)
	if _, err := lifecycle.ApplyManagedTaskWithMandate(context.Background(), plan.ID, plan, wrongGovernance); err == nil {
		t.Fatal("external effects applied under a mandate for different context-refresh authority")
	}
	wrongOperation := mandate
	wrongOperation.ID = ""
	wrongOperation.OperationID = "different-operation"
	wrongOperation = engine.BindWorkExecutionMandate(wrongOperation)
	if _, err := lifecycle.ApplyManagedTaskWithMandate(context.Background(), plan.ID, plan, wrongOperation); err == nil {
		t.Fatal("external effects applied under a mandate for a different selected operation")
	}
	result, err := lifecycle.ApplyManagedTaskWithMandate(context.Background(), plan.ID, plan, mandate)
	if err != nil || result.Status != engine.WorkApplyApplied {
		t.Fatalf("contained external effects did not apply: %#v, %v", result, err)
	}
	for _, receipt := range result.Receipts {
		if receipt.MandateID != mandate.ID {
			t.Fatalf("effect receipt lost mandate identity: %#v", receipt)
		}
	}
	if err := os.RemoveAll(filepath.Join(request.Repository, ".starter-kit", "work-manager")); err != nil {
		t.Fatal(err)
	}
	observation := adapter.Observation()
	observation.Task.StatusOption = "option:backlog"
	observation.Revision = "observation:new-drift-after-mandate-limit"
	adapter.SetObservation(observation)
	other := request
	other.Intent.OperationID = "intervening-operation"
	other.Intent.Task.ManagedID = "issue:intervening"
	other.ExecutionMandate = nil
	if _, err := lifecycle.InspectManagedTask(context.Background(), other); err != nil {
		t.Fatal(err)
	}
	inspection, err = lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err = lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyManagedTaskWithMandate(context.Background(), plan.ID, plan, mandate); err == nil {
		t.Fatal("reused mandate exceeded its cumulative effect ceiling")
	}
}

func TestGovernedManagedTaskEditedAcceptanceNeedsRefinementWithoutPlan(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	journey, err := lifecycle.ManageTask(context.Background(), request)
	if err != nil || journey.Verification.OverallState != engine.ControlPass {
		t.Fatalf("seed governed work: %#v, %v", journey, err)
	}
	observation := adapter.Observation()
	edited := *observation.Task.IssueContract
	edited.Acceptance = "- [ ] A materially different outcome is delivered."
	observation.Task.IssueContract = &edited
	observation.Task.IssueContractDigest = engine.ExecutableIssueContractDigest(edited)
	observation.Revision = "observation:edited-acceptance"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != string(engine.WorkFreshnessNeedsRefinement) || inspection.Qualification == nil || inspection.Qualification.Assessment.Disposition != engine.WorkFreshnessNeedsRefinement {
		t.Fatalf("edited acceptance was not returned to refinement: %#v", inspection)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
		t.Fatal("edited acceptance produced a plan")
	}
}

func TestGovernedManagedTaskRefreshesOnlyContainedCurrentContext(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	if _, err := lifecycle.ManageTask(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	observation := adapter.Observation()
	stale := *observation.Task.IssueContract
	stale.CurrentContext = "An older non-semantic fixture description."
	request.Intent.Governance.RefreshableContextDigests = []string{engine.ExecutableIssueContextDigest(stale.CurrentContext)}
	observation.Task.IssueContract = &stale
	observation.Task.IssueContractDigest = engine.ExecutableIssueContractDigest(stale)
	observation.Revision = "observation:stale-context"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != "inspected" || inspection.Qualification == nil || !slices.Contains(inspection.Qualification.Assessment.Repairs, "context") {
		t.Fatalf("contained context drift did not produce a bounded repair: %#v", inspection)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || !slices.Equal(plan.Effects[0].Operations, []string{"context"}) {
		t.Fatalf("context refresh broadened its effect: %#v", plan.Effects)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository); err != nil {
		t.Fatal(err)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil || status.Freshness != engine.WorkFreshnessContainedContextRefreshed {
		t.Fatalf("contained context refresh lacks a truthful final disposition: %#v, %v", status, err)
	}
}

func TestGovernedManagedTaskDoesNotOverwriteHumanOwnedContext(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	if _, err := lifecycle.ManageTask(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	observation := adapter.Observation()
	edited := *observation.Task.IssueContract
	edited.CurrentContext = "A human recorded a changed authority or risk fact."
	observation.Task.IssueContract = &edited
	observation.Task.IssueContractDigest = engine.ExecutableIssueContractDigest(edited)
	observation.Revision = "observation:human-context"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil || inspection.Qualification == nil || inspection.Qualification.Assessment.Disposition != engine.WorkFreshnessNeedsRefinement {
		t.Fatalf("human-owned context was not protected: %#v, %v", inspection, err)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
		t.Fatal("human-owned context change produced an overwrite plan")
	}
}

func TestGovernedManagedTaskReportsVerifiedMechanicalDriftRepair(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	if _, err := lifecycle.ManageTask(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	observation := adapter.Observation()
	observation.Task.StatusOption = "option:backlog"
	observation.Revision = "observation:mechanical-status-drift"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil || inspection.Disposition != "inspected" {
		t.Fatalf("mechanical drift did not remain repairable: %#v, %v", inspection, err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil || len(plan.Effects) != 1 || !slices.Contains(plan.Effects[0].Operations, "status") {
		t.Fatalf("mechanical status drift was not planned: %#v, %v", plan, err)
	}
	if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository); err != nil {
		t.Fatal(err)
	}
	status, err := lifecycle.ManagedTaskStatus(context.Background(), request.Repository)
	if err != nil || status.Freshness != engine.WorkFreshnessMechanicalDriftRepaired {
		t.Fatalf("mechanical repair lacks a truthful final disposition: %#v, %v", status, err)
	}
}

func TestGovernedManagedTaskMissingExecutableSchemaNeedsRefinement(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: "issue:71", IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: "task", ReadinessOption: "option:ready", StatusOption: "option:next",
		Phase: "Phase 3", PhaseOption: "option:phase-3", PhaseAssignmentReason: request.Intent.Task.PhaseAssignmentReason,
		Review: request.Intent.Task.Review, IssueContractProblems: []string{"executable issue contract is missing or invalid"},
	}
	observation.Revision = "observation:missing-contract"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != string(engine.WorkFreshnessNeedsRefinement) {
		t.Fatalf("missing executable schema did not stop at refinement: %#v", inspection)
	}
	if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
		t.Fatal("missing executable schema produced a plan")
	}
}

func TestGovernedManagedTaskStopsForDeliveredBlockedAndStaleSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*testing.T, *engine.InMemoryWorkAdapter, *engine.ManagedTaskRequest)
		want      engine.WorkFreshnessDisposition
	}{
		{
			name: "already delivered",
			configure: func(_ *testing.T, adapter *engine.InMemoryWorkAdapter, request *engine.ManagedTaskRequest) {
				observation := adapter.Observation()
				observation.Delivery = &engine.WorkDeliveryObservation{
					State: "complete", SourceRevision: request.Intent.SourceRevision,
					ContractDigest: engine.ExecutableIssueContractDigest(request.Intent.Governance.Issue), RepositoryRevision: "default-head", Evidence: []string{"pr:exact-outcome"},
				}
				observation.Revision = "observation:already-delivered"
				adapter.SetObservation(observation)
			},
			want: engine.WorkFreshnessAlreadyDelivered,
		},
		{
			name: "new native blocker",
			configure: func(_ *testing.T, adapter *engine.InMemoryWorkAdapter, _ *engine.ManagedTaskRequest) {
				observation := adapter.Observation()
				observation.Relationships.Blockers = []engine.WorkDependency{{ManagedID: "issue:blocker", Closed: false}}
				observation.Revision = "observation:new-blocker"
				adapter.SetObservation(observation)
			},
			want: engine.WorkFreshnessBlocked,
		},
		{
			name: "control or human action block",
			configure: func(_ *testing.T, adapter *engine.InMemoryWorkAdapter, request *engine.ManagedTaskRequest) {
				request.Intent.Task.Blockers = []engine.WorkDependency{{ManagedID: "issue:closed", Closed: true}}
				observation := adapter.Observation()
				observation.Task = observedGovernedTask(request, "option:blocked")
				observation.Relationships.Blockers = slices.Clone(request.Intent.Task.Blockers)
				observation.Revision = "observation:control-blocked"
				adapter.SetObservation(observation)
			},
			want: engine.WorkFreshnessBlocked,
		},
		{
			name: "stale governed source",
			configure: func(t *testing.T, _ *engine.InMemoryWorkAdapter, request *engine.ManagedTaskRequest) {
				if err := os.WriteFile(filepath.Join(request.Repository, "docs", "authority.md"), []byte("changed authority\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: engine.WorkFreshnessNeedsRefinement,
		},
		{
			name: "partial delivery",
			configure: func(_ *testing.T, adapter *engine.InMemoryWorkAdapter, _ *engine.ManagedTaskRequest) {
				observation := adapter.Observation()
				observation.Delivery = &engine.WorkDeliveryObservation{State: "partial", ResidualScope: "Acceptance remains incomplete.", Evidence: []string{"pr:partial"}}
				observation.Revision = "observation:partial-delivery"
				adapter.SetObservation(observation)
			},
			want: engine.WorkFreshnessNeedsRefinement,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lifecycle, adapter, request, _ := newManagedTaskFixture(t)
			configureGovernedWorkFixture(t, &request)
			test.configure(t, adapter, &request)
			inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if inspection.Qualification == nil || inspection.Qualification.Assessment.Disposition != test.want || inspection.Disposition != string(test.want) {
				t.Fatalf("freshness disposition = %#v, want %q", inspection, test.want)
			}
			if _, err := lifecycle.PlanManagedTask(context.Background(), inspection); err == nil {
				t.Fatalf("%s qualification produced a plan", test.want)
			}
		})
	}
}

func TestGovernedFreshnessDispositionPrecedenceIsExplicit(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		delivered bool
		blocked   bool
		stale     bool
		want      engine.WorkFreshnessDisposition
	}{
		{name: "exact delivery outranks semantic conflict", delivered: true, stale: true, want: engine.WorkFreshnessAlreadyDelivered},
		{name: "blocking outranks semantic conflict", blocked: true, stale: true, want: engine.WorkFreshnessBlocked},
		{name: "exact delivery outranks an obsolete block", delivered: true, blocked: true, want: engine.WorkFreshnessAlreadyDelivered},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lifecycle, adapter, request, _ := newManagedTaskFixture(t)
			configureGovernedWorkFixture(t, &request)
			observation := adapter.Observation()
			if test.delivered {
				observation.Delivery = &engine.WorkDeliveryObservation{
					State: "complete", SourceRevision: request.Intent.SourceRevision,
					ContractDigest: engine.ExecutableIssueContractDigest(request.Intent.Governance.Issue), RepositoryRevision: "default-head", Evidence: []string{"pr:exact-outcome"},
				}
			}
			if test.blocked {
				observation.Task = observedGovernedTask(&request, "option:blocked")
			}
			observation.Revision = "observation:combined-precedence"
			adapter.SetObservation(observation)
			if test.stale {
				if err := os.WriteFile(filepath.Join(request.Repository, "docs", "authority.md"), []byte("changed authority\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
			if err != nil || inspection.Qualification == nil || inspection.Qualification.Assessment.Disposition != test.want {
				t.Fatalf("combined disposition = %#v, %v; want %q", inspection, err, test.want)
			}
		})
	}
}

func TestExecutableIssueContractCanonicalRoundTrip(t *testing.T) {
	t.Parallel()

	contracts := []engine.ExecutableIssueContract{governedIssueContractFixture()}
	question := governedIssueContractFixture()
	question.Subtype = &engine.WorkSubtypeContract{Question: &engine.QuestionWorkContract{
		Question: "Which owner approves?", Impact: "Delivery depends on it.", Relationship: "blocking", AnswerAuthority: "product owner",
		EvidenceNeeds: "An owner record.", ResolutionCriteria: "One authoritative answer.", PromotionDestination: "docs/decisions/DEC-TEST.md",
	}}
	contracts = append(contracts, question)
	research := governedIssueContractFixture()
	research.Subtype = &engine.WorkSubtypeContract{Research: &engine.ResearchWorkContract{
		Objective: "Compare routes.", IntendedUse: "Select a route.", Scope: "Supported routes.", Exclusions: "Unrelated providers.",
		Provenance: "Primary sources.", DepthOrEffort: "Two hours.", Authority: "Read-only research.", StoppingConditions: "Routes are distinguishable.",
		Output: "docs/research/RESULT.md", Freshness: "Checked at execution.", ReviewNeeds: "Maintainer review.",
	}}
	contracts = append(contracts, research)
	for _, contract := range contracts {
		body, err := engine.RenderExecutableIssueContract(contract)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := engine.ParseExecutableIssueContract(body)
		if err != nil {
			t.Fatal(err)
		}
		if engine.ExecutableIssueContractDigest(parsed) != engine.ExecutableIssueContractDigest(contract) {
			t.Fatalf("canonical contract changed across render/parse: %#v", parsed)
		}
	}
	body, err := engine.RenderExecutableIssueContract(governedIssueContractFixture())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ParseExecutableIssueContract("changed visible preamble\n" + body); err == nil {
		t.Fatal("unexpected visible preamble was ignored")
	}
	formBody := strings.ReplaceAll(body, "\n## ", "\n### ")
	parsedForm, err := engine.ParseExecutableIssueContract(formBody)
	if err != nil || engine.ExecutableIssueContractDigest(parsedForm) != engine.ExecutableIssueContractDigest(governedIssueContractFixture()) {
		t.Fatalf("GitHub issue-form H3 body did not round-trip: %#v, %v", parsedForm, err)
	}
	questionBody, err := engine.RenderExecutableIssueContract(question)
	if err != nil {
		t.Fatal(err)
	}
	questionFormBody := strings.ReplaceAll(questionBody, "\n## ", "\n### ") + "\n\n### No-promotion resolution\n\n_No response_"
	parsedQuestionForm, err := engine.ParseExecutableIssueContract(questionFormBody)
	if err != nil || engine.ExecutableIssueContractDigest(parsedQuestionForm) != engine.ExecutableIssueContractDigest(question) {
		t.Fatalf("optional GitHub question-form response changed the contract: %#v, %v", parsedQuestionForm, err)
	}
}

func TestWorkDeliveryClaimRejectsUnboundedOrAmbiguousImplementedSources(t *testing.T) {
	t.Parallel()

	base := engine.WorkDeliveryClaim{
		SchemaVersion: 1, ManagedID: "issue:74", SourceRevision: "source:v1",
		ContractDigest:     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ImplementedSources: []engine.GovernedSourceBinding{{ID: "engine", Path: "engine/workmanager.go", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}},
	}
	if marker, err := engine.RenderWorkDeliveryClaim(base); err != nil {
		t.Fatal(err)
	} else if parsed, err := engine.ParseWorkDeliveryClaim(marker); err != nil || !slices.Equal(parsed.ImplementedSources, base.ImplementedSources) {
		t.Fatalf("delivery claim round trip = %#v, %v", parsed, err)
	}
	for _, mutate := range []func(*engine.WorkDeliveryClaim){
		func(claim *engine.WorkDeliveryClaim) { claim.ImplementedSources = nil },
		func(claim *engine.WorkDeliveryClaim) { claim.ImplementedSources[0].Path = "../escape" },
		func(claim *engine.WorkDeliveryClaim) {
			claim.ImplementedSources = append(claim.ImplementedSources, claim.ImplementedSources[0])
		},
		func(claim *engine.WorkDeliveryClaim) { claim.ImplementedSources[0].Digest = "not-a-digest" },
	} {
		claim := base
		claim.ImplementedSources = slices.Clone(base.ImplementedSources)
		mutate(&claim)
		if _, err := engine.RenderWorkDeliveryClaim(claim); err == nil {
			t.Fatalf("invalid implemented-source claim passed: %#v", claim)
		}
	}
}

func TestPromotedRecordBacklinkDoesNotAcceptManagedIDSubstringCollisions(t *testing.T) {
	t.Parallel()

	body, err := engine.RenderWorkPromotedRecordBacklink(engine.WorkPromotedRecordBacklink{
		SchemaVersion: 1, ManagedID: "issue:74", IssueURL: "https://github.com/example/repository/issues/74",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "issue:7") {
		t.Fatal("fixture no longer exercises the substring collision")
	}
	parsed, err := engine.ParseWorkPromotedRecordBacklink(body)
	if err != nil || parsed.ManagedID == "issue:7" || parsed.ManagedID != "issue:74" {
		t.Fatalf("promoted backlink identity = %#v, %v", parsed, err)
	}
}

func TestGovernedManagedTaskRequiresReferenceSourceCorrespondence(t *testing.T) {
	t.Parallel()

	lifecycle, _, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	request.Intent.Governance.Issue.GoverningReferences = "- DEC-OTHER — unrelated authority."
	if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
		t.Fatal("unbound visible governing reference was accepted")
	}
}

func TestGovernedQuestionAndResearchSubtypeContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*engine.ManagedTaskRequest)
		wantError bool
	}{
		{
			name: "question ready",
			configure: func(request *engine.ManagedTaskRequest) {
				request.Intent.Task.IssueType = "question"
				request.Intent.Task.Review = nil
				request.Intent.Governance.Issue.Subtype = &engine.WorkSubtypeContract{Question: &engine.QuestionWorkContract{
					Question: "Which owner approves the outcome?", Impact: "The answer controls delivery.", Relationship: "blocking",
					AnswerAuthority: "product owner", EvidenceNeeds: "an owner record", ResolutionCriteria: "one authoritative answer",
					PromotionDestination: "docs/decisions/DEC-TEST.md",
				}}
			},
		},
		{
			name: "question missing authority",
			configure: func(request *engine.ManagedTaskRequest) {
				request.Intent.Task.IssueType = "question"
				request.Intent.Task.Review = nil
				request.Intent.Governance.Issue.Subtype = &engine.WorkSubtypeContract{Question: &engine.QuestionWorkContract{
					Question: "Which owner approves the outcome?", Impact: "The answer controls delivery.", Relationship: "blocking",
					EvidenceNeeds: "an owner record", ResolutionCriteria: "one authoritative answer", PromotionDestination: "docs/decisions/DEC-TEST.md",
				}}
			},
			wantError: true,
		},
		{
			name: "closed question visible no-promotion resolution",
			configure: func(request *engine.ManagedTaskRequest) {
				request.Intent.Task.IssueType = "question"
				request.Intent.Task.Review = nil
				request.Intent.Task.Closed = true
				request.Intent.Task.Status = "done"
				request.Intent.Task.NoPromotionRequired = true
				request.Intent.Governance.Issue.Subtype = &engine.WorkSubtypeContract{Question: &engine.QuestionWorkContract{
					Question: "Does this answer require promotion?", Impact: "The issue must close truthfully.", Relationship: "related",
					AnswerAuthority: "product owner", EvidenceNeeds: "an owner resolution", ResolutionCriteria: "one explicit answer",
					PromotionDestination: "docs/decisions/DEC-TEST.md", NoPromotionResolution: "The answer only confirms that the governed destination is unchanged.",
				}}
			},
		},
		{
			name: "closed question hidden no-promotion resolution",
			configure: func(request *engine.ManagedTaskRequest) {
				request.Intent.Task.IssueType = "question"
				request.Intent.Task.Review = nil
				request.Intent.Task.Closed = true
				request.Intent.Task.Status = "done"
				request.Intent.Task.NoPromotionRequired = true
				request.Intent.Governance.Issue.Subtype = &engine.WorkSubtypeContract{Question: &engine.QuestionWorkContract{
					Question: "Does this answer require promotion?", Impact: "The issue must close truthfully.", Relationship: "related",
					AnswerAuthority: "product owner", EvidenceNeeds: "an owner resolution", ResolutionCriteria: "one explicit answer",
					PromotionDestination: "docs/decisions/DEC-TEST.md",
				}}
			},
			wantError: true,
		},
		{
			name: "closed research output",
			configure: func(request *engine.ManagedTaskRequest) {
				request.Intent.Task.IssueType = "research"
				request.Intent.Task.Review = nil
				request.Intent.Task.Closed = true
				request.Intent.Task.PromotionRecord = "docs/research/RESULT.md"
				request.Intent.Governance.Issue.Subtype = &engine.WorkSubtypeContract{Research: &engine.ResearchWorkContract{
					Objective: "Compare supported routes.", IntendedUse: "Select one route.", Scope: "Public GitHub.", Exclusions: "Paid features.",
					Provenance: "Official documentation.", DepthOrEffort: "Two hours.", Authority: "read-only web research.",
					StoppingConditions: "The routes are distinguishable.", Output: "docs/research/RESULT.md", Freshness: "Checked at execution.", ReviewNeeds: "Maintainer review.",
				}}
				bindPromotionOutput(t, request, "RESEARCH-RESULT", "docs/research/RESULT.md")
			},
		},
		{
			name: "closed research wrong output",
			configure: func(request *engine.ManagedTaskRequest) {
				request.Intent.Task.IssueType = "research"
				request.Intent.Task.Review = nil
				request.Intent.Task.Closed = true
				request.Intent.Task.PromotionRecord = "docs/research/OTHER.md"
				request.Intent.Governance.Issue.Subtype = &engine.WorkSubtypeContract{Research: &engine.ResearchWorkContract{
					Objective: "Compare supported routes.", IntendedUse: "Select one route.", Scope: "Public GitHub.", Exclusions: "Paid features.",
					Provenance: "Official documentation.", DepthOrEffort: "Two hours.", Authority: "read-only web research.",
					StoppingConditions: "The routes are distinguishable.", Output: "docs/research/RESULT.md", Freshness: "Checked at execution.", ReviewNeeds: "Maintainer review.",
				}}
			},
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lifecycle, _, request, _ := newManagedTaskFixture(t)
			configureGovernedWorkFixture(t, &request)
			test.configure(&request)
			inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
			if test.wantError {
				if err == nil {
					t.Fatalf("invalid subtype contract was accepted: %#v", inspection)
				}
				return
			}
			if err != nil || inspection.Qualification == nil || inspection.Qualification.Assessment.Disposition != engine.WorkFreshnessFresh {
				t.Fatalf("valid subtype contract did not qualify: %#v, %v", inspection, err)
			}
		})
	}
}

func TestGovernedFeatureHorizonIsIndependentAndCopiedChildValueIsCleared(t *testing.T) {
	t.Parallel()

	t.Run("feature direct Horizon", func(t *testing.T) {
		t.Parallel()
		lifecycle, adapter, request, _ := newManagedTaskFixture(t)
		configureGovernedWorkFixture(t, &request)
		configureHorizonTarget(adapter, &request)
		request.Intent.Task.IssueType = "feature"
		request.Intent.Task.Horizon = "now"
		journey, err := lifecycle.ManageTask(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		if journey.Plan.DerivedFacts.Horizon != "now" || journey.Plan.DerivedFacts.HorizonSource != "direct" || journey.Plan.DerivedFacts.Status != "next" {
			t.Fatalf("Horizon was conflated with execution state: %#v", journey.Plan.DerivedFacts)
		}
	})

	t.Run("ordinary child direct Horizon rejected", func(t *testing.T) {
		t.Parallel()
		lifecycle, adapter, request, _ := newManagedTaskFixture(t)
		configureGovernedWorkFixture(t, &request)
		configureHorizonTarget(adapter, &request)
		request.Intent.Task.Horizon = "now"
		if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
			t.Fatal("ordinary child received a copied direct Horizon")
		}
	})

	t.Run("Ready feature requires configured Horizon", func(t *testing.T) {
		t.Parallel()
		lifecycle, adapter, request, _ := newManagedTaskFixture(t)
		configureGovernedWorkFixture(t, &request)
		configureHorizonTarget(adapter, &request)
		request.Intent.Task.IssueType = "feature"
		if _, err := lifecycle.InspectManagedTask(context.Background(), request); err == nil {
			t.Fatal("Ready feature omitted configured Horizon")
		}
	})

	t.Run("absent Horizon capability remains explicit", func(t *testing.T) {
		t.Parallel()
		lifecycle, _, request, _ := newManagedTaskFixture(t)
		configureGovernedWorkFixture(t, &request)
		journey, err := lifecycle.ManageTask(context.Background(), request)
		if err != nil || journey.Plan.DerivedFacts.HorizonCapability != "not-configured" || journey.Plan.DerivedFacts.PhaseCapability != "configured" {
			t.Fatalf("optional roadmap capability state was not explicit: %#v, %v", journey.Plan.DerivedFacts, err)
		}
	})

	t.Run("copied child Horizon cleared", func(t *testing.T) {
		t.Parallel()
		lifecycle, adapter, request, _ := newManagedTaskFixture(t)
		configureGovernedWorkFixture(t, &request)
		configureHorizonTarget(adapter, &request)
		if _, err := lifecycle.ManageTask(context.Background(), request); err != nil {
			t.Fatal(err)
		}
		observation := adapter.Observation()
		observation.Task.HorizonOption = "option:horizon-now"
		observation.Revision = "observation:copied-horizon"
		adapter.SetObservation(observation)
		inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
		if err != nil || len(plan.Effects) != 1 || !slices.Contains(plan.Effects[0].Operations, "horizon") {
			t.Fatalf("copied child Horizon was not planned for clearing: %#v, %v", plan, err)
		}
		if _, err := lifecycle.ApplyManagedTask(context.Background(), plan.ID, plan); err != nil {
			t.Fatal(err)
		}
		if _, err := lifecycle.VerifyManagedTask(context.Background(), request.Repository); err != nil {
			t.Fatal(err)
		}
		if adapter.Observation().Task.HorizonOption != "" {
			t.Fatalf("copied child Horizon remained assigned: %#v", adapter.Observation().Task)
		}
	})
}

func TestGovernedChildDerivesHorizonFromNativeParent(t *testing.T) {
	t.Parallel()

	lifecycle, adapter, request, _ := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &request)
	configureHorizonTarget(adapter, &request)
	contract := request.Intent.Governance.Issue
	request.Intent.Task.ParentManagedID = "issue:4"
	request.Intent.Task.ParentContext = &engine.WorkParentContext{ManagedID: "issue:4", Status: "backlog"}
	observation := adapter.Observation()
	observation.Task = &engine.WorkObservedTask{
		ManagedID: request.Intent.Task.ManagedID, IssueNodeID: "memory:issue:71", ProjectItemID: "memory:item:71",
		Title: request.Intent.Task.Title, IssueType: request.Intent.Task.IssueType, ParentManagedID: "issue:4", NativeParentManagedID: "issue:4",
		ReadinessOption: "option:ready", StatusOption: "option:next", ParentHorizonOption: "option:horizon-next",
		Phase: "Phase 3", PhaseOption: "option:phase-3", PhaseAssignmentReason: request.Intent.Task.PhaseAssignmentReason,
		Review: request.Intent.Task.Review, IssueContract: &contract, IssueContractDigest: engine.ExecutableIssueContractDigest(contract),
	}
	observation.RelatedTasks = []engine.WorkObservedTask{{
		ManagedID: "issue:4", IssueNodeID: "memory:issue:4", ProjectItemID: "memory:item:4", Title: "Parent feature", IssueType: "feature",
		ReadinessOption: "option:ready", StatusOption: "option:backlog", HorizonOption: "option:horizon-next",
	}}
	observation.Relationships = engine.WorkRelationshipObservation{Observed: true, ParentManagedID: "issue:4"}
	observation.Revision = "observation:native-parent-horizon"
	adapter.SetObservation(observation)
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	copied := false
	for _, effect := range plan.Effects {
		copied = copied || slices.Contains(effect.Operations, "horizon")
	}
	if plan.DerivedFacts.Horizon != "next" || plan.DerivedFacts.HorizonSource != "parent" || copied {
		t.Fatalf("child did not derive parent Horizon without copying it: %#v", plan)
	}
}

func configureGovernedWorkFixture(t *testing.T, request *engine.ManagedTaskRequest) {
	t.Helper()
	content := []byte("# Governed authority\n\nStable source for the fixture.\n")
	path := filepath.Join(request.Repository, "docs", "authority.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	request.Intent.SchemaVersion = 2
	request.Intent.EffectBoundary = engine.WorkEffectBoundary{DataClass: "public-project-metadata", CostCeiling: "zero-dollar", Destructive: "no-delete", Retention: "repository-evidence", RecoveryOwner: "owner"}
	request.Intent.Governance = &engine.GovernedWorkContract{
		SchemaVersion: 1,
		Issue:         governedIssueContractFixture(),
		Sources: []engine.GovernedSourceBinding{{
			ID: "DEC-TEST", Path: "docs/authority.md", Digest: fmt.Sprintf("sha256:%x", digest),
		}},
	}
}

func observedGovernedTask(request *engine.ManagedTaskRequest, readinessOption string) *engine.WorkObservedTask {
	contract := request.Intent.Governance.Issue
	return &engine.WorkObservedTask{
		ManagedID: request.Intent.Task.ManagedID, IssueNodeID: "memory:" + request.Intent.Task.ManagedID, ProjectItemID: "memory:item:" + request.Intent.Task.ManagedID,
		Title: request.Intent.Task.Title, IssueType: request.Intent.Task.IssueType, ReadinessOption: readinessOption,
		StatusOption: request.Intent.Target.OptionIDs["status:"+request.Intent.Task.Status], Phase: request.Intent.Task.Phase,
		PhaseOption: request.Intent.Target.OptionIDs["phase:"+request.Intent.Task.Phase], PhaseAssignmentReason: request.Intent.Task.PhaseAssignmentReason,
		Review: request.Intent.Task.Review, IssueContract: &contract, IssueContractDigest: engine.ExecutableIssueContractDigest(contract),
	}
}

func bindPromotionOutput(t *testing.T, request *engine.ManagedTaskRequest, id, slashPath string) {
	t.Helper()
	backlink, err := engine.RenderWorkPromotedRecordBacklink(engine.WorkPromotedRecordBacklink{SchemaVersion: 1, ManagedID: request.Intent.Task.ManagedID, IssueURL: "https://github.com/example/repository/issues/" + strings.TrimPrefix(request.Intent.Task.ManagedID, "issue:")})
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("# Promoted result\n\n" + backlink + "\n")
	path := filepath.Join(request.Repository, filepath.FromSlash(slashPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	request.Intent.Governance.Sources = append(request.Intent.Governance.Sources, engine.GovernedSourceBinding{ID: id, Path: slashPath, Digest: fmt.Sprintf("sha256:%x", digest)})
	request.Intent.Governance.Issue.GoverningReferences += "\n- " + id + " — promoted output with reciprocal provenance."
}

func configureHorizonTarget(adapter *engine.InMemoryWorkAdapter, request *engine.ManagedTaskRequest) {
	observation := adapter.Observation()
	observation.Target.FieldIDs["horizon"] = "field:horizon"
	observation.Target.OptionIDs["horizon:now"] = "option:horizon-now"
	observation.Target.OptionIDs["horizon:next"] = "option:horizon-next"
	observation.Target.OptionIDs["horizon:later"] = "option:horizon-later"
	adapter.SetObservation(observation)
	request.Intent.Target = observation.Target
}

func governedIssueContractFixture() engine.ExecutableIssueContract {
	return engine.ExecutableIssueContract{
		SchemaVersion:       1,
		Parent:              "#4",
		HumanSummary:        "A maintainer can execute one governed task.\n\n**Done when:** the lifecycle binds current issue and source facts.",
		CurrentContext:      "The deterministic fixture supplies current normalized facts.",
		GoverningReferences: "- DEC-TEST — fixture authority.",
		Scope:               "Qualify one Ready managed task.",
		OutOfScope:          "External effects and release publication.",
		Acceptance:          "- [ ] A fresh issue produces a source-bound qualification.\n- [ ] Changed acceptance returns to refinement.",
		Verification:        "Exercise inspect, plan, apply, verify, and status through Work Manager.",
		Dependencies:        "No unresolved native blocker.",
		ReadinessAssertions: []string{
			"No unresolved product, architecture, policy, regulatory, or risk decision is hidden in this task.",
			"An authorized implementer can execute this without the originating conversation.",
		},
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
