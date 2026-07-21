package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestDeliveryPlansReadyTransitionOnlyAfterExactHeadGatesPass(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(engine.DeliveryObservation) engine.DeliveryObservation { return readyDraftObservation() })

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionDraft {
		t.Fatalf("disposition = %q, want draft", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectMarkReady {
		t.Fatalf("effects = %#v, want one mark-ready transition", plan.Effects)
	}
}

func TestDeliveryKeepsPendingChecksDistinctAndPlansNoEffect(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Checks[0].State = "pending"
		observation.Revision = "observation:checks-pending"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionChecksPending {
		t.Fatalf("disposition = %q, want checks-pending", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("plan = %#v, want explicit no-change wait", plan)
	}
}

func TestDeliveryKeepsPendingDistinctReviewSeparateFromChecks(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Reviews = nil
		observation.PullRequest.Draft = false
		observation.Revision = "observation:review-pending"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionReviewPending {
		t.Fatalf("disposition = %q, want review-pending", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("plan = %#v, want explicit no-change wait", plan)
	}
}

func TestDeliveryReturnsChangesRequestedToImplementation(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Reviews[0].State = "changes-requested"
		observation.PullRequest.Draft = false
		observation.Revision = "observation:changes-requested"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionChangesRequested {
		t.Fatalf("disposition = %q, want changes-requested", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("plan = %#v, want implementation wait", plan)
	}
}

func TestDeliveryPlansOnlySquashMergeForExactReadyHead(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.Revision = "observation:merge-ready"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionMergeReady {
		t.Fatalf("disposition = %q, want merge-ready", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectSquashMerge || plan.Effects[0].HeadRevision != "head-1" {
		t.Fatalf("effects = %#v, want exact-head squash merge", plan.Effects)
	}
}

func TestDeliveryPlansCompletionReconciliationAfterQualifyingMerge(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.PullRequest.State = "closed"
		observation.PullRequest.Merged = true
		observation.PullRequest.MergeRevision = "merge-1"
		observation.PullRequest.MergeMethod = "squash"
		observation.PullRequest.DefaultReachable = true
		observation.Revision = "observation:merged"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionMerged {
		t.Fatalf("disposition = %q, want merged", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectReconcileCompletion {
		t.Fatalf("effects = %#v, want completion reconciliation", plan.Effects)
	}
}

func TestDeliveryPreservesClosedUnmergedAsTerminalNonPass(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.PullRequest.State = "closed"
		observation.Revision = "observation:closed-unmerged"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionClosedUnmerged {
		t.Fatalf("disposition = %q, want closed-unmerged", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("plan = %#v, want terminal no-change", plan)
	}
}

func TestDeliveryRefusesExternalEffectWithoutExecutionMandate(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, engine.WorkExecutionMandate{}); err == nil {
		t.Fatal("external delivery effect applied without a current mandate")
	}
	observed, err := adapter.ObserveDelivery(context.Background(), request.Intent)
	if err != nil {
		t.Fatal(err)
	}
	if !observed.PullRequest.Draft {
		t.Fatal("denied mark-ready effect mutated adapter state")
	}
}

func TestDeliveryMandateMustContainExactActorTargetAndEffect(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	mandate := deliveryMandate(request, []string{"different-actor"}, plan.Effects[0].Kind)

	if _, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate); err == nil {
		t.Fatal("delivery applied under a mandate for a different actor")
	}
	observed, _ := adapter.ObserveDelivery(context.Background(), request.Intent)
	if !observed.PullRequest.Draft {
		t.Fatal("out-of-mandate effect mutated adapter state")
	}
}

func TestDeliveryAppliesContainedReadyTransitionAndReobservesIt(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	mandate := deliveryMandate(request, []string{"merger"}, plan.Effects[0].Kind)

	result, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyApplied || len(result.Results) != 1 || result.Results[0].Outcome != "applied" {
		t.Fatalf("apply = %#v, want one applied result", result)
	}
	observed, _ := adapter.ObserveDelivery(context.Background(), request.Intent)
	if observed.PullRequest.Draft {
		t.Fatal("contained mark-ready effect was not observable")
	}
}

func TestDeliveryMandateEffectCeilingIsCumulativeAcrossPlans(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	readyPlan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	mandate := deliveryMandate(request, []string{"merger"}, engine.DeliveryEffectMarkReady)
	mandate.EffectKinds = []string{engine.DeliveryEffectMarkReady, engine.DeliveryEffectSquashMerge}
	mandate.MaxEffects = 1
	mandate = engine.BindWorkExecutionMandate(mandate)
	if _, err := lifecycle.ApplyDelivery(context.Background(), readyPlan.ID, readyPlan, mandate); err != nil {
		t.Fatal(err)
	}

	inspection, _ = lifecycle.InspectDelivery(context.Background(), request)
	mergePlan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.ApplyDelivery(context.Background(), mergePlan.ID, mergePlan, mandate); err == nil {
		t.Fatal("second effect exceeded cumulative mandate ceiling")
	}
}

func TestDeliveryNoChangeWaitRequiresNoExternalAuthority(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Checks[0].State = "pending"
		return observation
	})
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)

	result, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, engine.WorkExecutionMandate{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyNoChange || len(result.Results) != 0 {
		t.Fatalf("apply = %#v, want authority-free no-change", result)
	}
}

func TestDeliveryRejectsPlanWhenExactHeadChangesBeforeApply(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	changed := readyDraftObservation()
	changed.PullRequest.HeadRevision = "head-2"
	changed.Checks[0].HeadRevision = "head-2"
	changed.Reviews[0].HeadRevision = "head-2"
	changed.Revision = "observation:head-2"
	adapter.SetObservation(changed)
	mandate := deliveryMandate(request, []string{"merger"}, plan.Effects[0].Kind)

	if _, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate); err == nil {
		t.Fatal("stale exact-head plan applied after branch changed")
	}
	observed, _ := adapter.ObserveDelivery(context.Background(), request.Intent)
	if !observed.PullRequest.Draft || observed.PullRequest.HeadRevision != "head-2" {
		t.Fatalf("stale plan mutated changed observation: %#v", observed.PullRequest)
	}
}

func deliveryFixture(t *testing.T, mutate func(engine.DeliveryObservation) engine.DeliveryObservation) (*engine.Engine, *engine.InMemoryDeliveryAdapter, engine.DeliveryRequest) {
	t.Helper()
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	observation := mutate(readyDraftObservation())
	adapter := engine.NewInMemoryDeliveryAdapter(engine.DeliveryCapability{
		SchemaVersion: 1, Online: true, Fresh: true, Actor: "merger", Mode: "github-app", Permissions: []string{"pull_requests:write"}, ObservedAt: now, ExpiresAt: now.Add(time.Hour),
	}, observation)
	lifecycle := engine.New(engine.WithClock(deliveryClock{now}), engine.WithDeliveryAdapter(adapter))
	request := engine.DeliveryRequest{Repository: t.TempDir(), Intent: engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: "deliver-75", SourceRevision: "source-1", ManagedID: "issue:75",
		BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", RequiredChecks: []string{"foundation"},
		Review: engine.WorkReviewRequirement{Role: "reviewer", DistinctContext: true}, MergeMethod: "squash", OperatingProfileRevision: "profile-1",
		Target:         engine.WorkTarget{Host: "github.com", RepositoryID: "R_repo", ProjectID: "P_project"},
		EffectBoundary: engine.WorkEffectBoundary{DataClass: "public-project-metadata", CostCeiling: "zero-dollar", Destructive: "no-delete", Retention: "repository-evidence", RecoveryOwner: "owner"},
	}}
	return lifecycle, adapter, request
}

func deliveryMandate(request engine.DeliveryRequest, actors []string, effectKind string) engine.WorkExecutionMandate {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	return engine.BindWorkExecutionMandate(engine.WorkExecutionMandate{
		SchemaVersion: 1, ApprovedBy: "owner", ApprovalID: "approval-75", ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
		Target: request.Intent.Target, OperationID: request.Intent.OperationID, SelectedManagedID: request.Intent.ManagedID,
		Actors: actors, CredentialModes: []string{"github-app"}, Permissions: []string{"pull_requests:write"},
		OperatingProfileRevisions: []string{request.Intent.OperatingProfileRevision}, SourceRevisions: []string{request.Intent.SourceRevision}, ManagedIDs: []string{request.Intent.ManagedID},
		EffectKinds: []string{effectKind}, ResourceDigests: []string{engine.DeliveryResourceDigest(request.Intent)}, MaxEffects: 2,
		DataClass: request.Intent.EffectBoundary.DataClass, CostCeiling: request.Intent.EffectBoundary.CostCeiling, Destructive: request.Intent.EffectBoundary.Destructive,
		Retention: request.Intent.EffectBoundary.Retention, RecoveryOwner: request.Intent.EffectBoundary.RecoveryOwner,
	})
}

func readyDraftObservation() engine.DeliveryObservation {
	return engine.DeliveryObservation{
		SchemaVersion: 1,
		Revision:      "observation:draft",
		Issue:         engine.DeliveryIssueObservation{ManagedID: "issue:75", State: "open"},
		PullRequest: engine.DeliveryPullRequestObservation{
			Number: 101, State: "open", Draft: true, Base: "main", Head: "task/75-delivery-squash-completion", HeadRevision: "head-1",
		},
		Checks:  []engine.DeliveryCheckObservation{{Name: "foundation", HeadRevision: "head-1", State: "passed"}},
		Reviews: []engine.DeliveryReviewObservation{{Actor: "reviewer", HeadRevision: "head-1", State: "approved", DistinctContext: true, Capable: true}},
		Rules:   engine.DeliveryRulesObservation{Revision: "rules-1", RequiredChecks: []string{"foundation"}, MergeMethods: []string{"squash"}},
	}
}

type deliveryClock struct{ now time.Time }

func (clock deliveryClock) Now() time.Time { return clock.now }
