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

func TestDeliveryPlansIssueNamedBranchWhenItIsAbsent(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Branch = engine.DeliveryBranchObservation{}
		observation.PullRequest = engine.DeliveryPullRequestObservation{}
		observation.Checks = nil
		observation.Reviews = nil
		observation.Rules.BaseRevision = "base-1"
		observation.Revision = "observation:branch-absent"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionBranchAbsent {
		t.Fatalf("disposition = %q, want branch-absent", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectCreateBranch || plan.Effects[0].HeadRevision != "base-1" {
		t.Fatalf("effects = %#v, want exact-base branch creation", plan.Effects)
	}
}

func TestDeliveryPlansDraftPullRequestWhenBranchExists(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest = engine.DeliveryPullRequestObservation{}
		observation.Checks = nil
		observation.Reviews = nil
		observation.Revision = "observation:pull-request-absent"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionPullRequestAbsent {
		t.Fatalf("disposition = %q, want pull-request-absent", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectCreatePullRequest || plan.Effects[0].Claim == nil {
		t.Fatalf("effects = %#v, want claimed draft pull request", plan.Effects)
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

func TestDeliveryKeepsFailedChecksDistinctAndPlansNoEffect(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.Checks[0].State = "failed"
		observation.Revision = "observation:checks-failed"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionChecksFailed {
		t.Fatalf("disposition = %q, want checks-failed", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("plan = %#v, want explicit no-change stop", plan)
	}
}

func TestDeliveryKeepsPendingDistinctReviewSeparateFromChecks(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Reviews = nil
		observation.PullRequest.Draft = false
		observation.PullRequest.RequestedReviewers = []string{"reviewer"}
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

func TestDeliveryRequiresQualifiedIndependenceWhenPolicyAddsIt(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.Reviews[0].QualifiedIndependent = false
		return observation
	})
	request.Intent.Review.QualifiedIndependent = true
	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionReviewPending {
		t.Fatalf("disposition = %q, want qualified review pending", inspection.Disposition)
	}
}

func TestDeliveryUsesLatestEffectiveReviewForEachActor(t *testing.T) {
	now := time.Date(2026, 7, 21, 20, 0, 0, 0, time.UTC)
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.Reviews = []engine.DeliveryReviewObservation{
			{Actor: "reviewer", HeadRevision: "head-1", State: "changes-requested", DistinctContext: true, Capable: true, EvidenceID: "review:1", ObservedAt: now},
			{Actor: "reviewer", HeadRevision: "head-1", State: "approved", DistinctContext: true, Capable: true, EvidenceID: "review:2", ObservedAt: now.Add(time.Minute)},
		}
		return observation
	})
	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionMergeReady {
		t.Fatalf("disposition = %q, want latest approval", inspection.Disposition)
	}
}

func TestDeliveryPlansReviewerRoutingSeparatelyFromApproval(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.Reviews = nil
		observation.PullRequest.Draft = false
		observation.PullRequest.RequestedReviewers = nil
		observation.Revision = "observation:review-unrequested"
		return observation
	})

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionReviewUnrequested {
		t.Fatalf("disposition = %q, want review-unrequested", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectRequestReview || plan.Effects[0].Reviewer != "reviewer" {
		t.Fatalf("effects = %#v, want reviewer routing", plan.Effects)
	}
}

func TestDeliveryKeepsOptionalProductApprovalSeparateFromReview(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.PullRequest.RequestedReviewers = []string{"reviewer"}
		observation.Revision = "observation:approval-unrequested"
		return observation
	})
	request.Intent.ProductApproval = engine.WorkReviewRequirement{Role: "product-owner", DistinctContext: true}

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionApprovalUnrequested {
		t.Fatalf("disposition = %q, want approval-unrequested", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || plan.Effects[0].Kind != engine.DeliveryEffectRequestReview || plan.Effects[0].Reviewer != "product-owner" {
		t.Fatalf("effects = %#v, want product approval routing", plan.Effects)
	}
}

func TestDeliveryWaitsForRequestedProductApprovalOnExactHead(t *testing.T) {
	lifecycle, _, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		observation.PullRequest.RequestedReviewers = []string{"reviewer", "product-owner"}
		observation.Revision = "observation:approval-pending"
		return observation
	})
	request.Intent.ProductApproval = engine.WorkReviewRequirement{Role: "product-owner", DistinctContext: true}

	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Disposition != engine.DeliveryDispositionApprovalPending {
		t.Fatalf("disposition = %q, want approval-pending", inspection.Disposition)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoChange || len(plan.Effects) != 0 {
		t.Fatalf("plan = %#v, want explicit approval wait", plan)
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

func TestDeliveryRecoversLostEffectResponseByExactReobservationWithoutRetry(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	adapter.QueueApplyResult(engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "response lost", Recoverable: true}, true, context.DeadlineExceeded)
	mandate := deliveryMandate(request, []string{"merger"}, plan.Effects[0].Kind)

	result, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyApplied || len(result.Results) != 1 || result.Results[0].Outcome != "applied" || adapter.ApplyCount() != 1 {
		t.Fatalf("lost-response recovery = %#v, calls=%d", result, adapter.ApplyCount())
	}
}

func TestDeliveryKeepsUnresolvedLostResponseAsZeroRetryNonPass(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	adapter.QueueApplyResult(engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "response lost", Recoverable: true}, false, context.DeadlineExceeded)
	mandate := deliveryMandate(request, []string{"merger"}, plan.Effects[0].Kind)

	result, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate)
	if err == nil || result.Status != engine.WorkApplyNonPass || adapter.ApplyCount() != 1 {
		t.Fatalf("unresolved response = %#v, err=%v, calls=%d", result, err, adapter.ApplyCount())
	}
}

func TestDeliveryNeverInfersSquashFromAmbiguousConcurrentMerge(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation {
		observation.PullRequest.Draft = false
		return observation
	})
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	if plan.Effects[0].Kind != engine.DeliveryEffectSquashMerge {
		t.Fatalf("effect = %#v", plan.Effects[0])
	}
	adapter.QueueApplyResult(engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "response lost", Recoverable: true}, true, context.DeadlineExceeded)
	mandate := deliveryMandate(request, []string{"merger"}, engine.DeliveryEffectSquashMerge)

	result, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate)
	if err == nil || result.Status != engine.WorkApplyNonPass || result.Receipts[0].MergeRevision != "" {
		t.Fatalf("ambiguous merge = %#v, err = %v", result, err)
	}
}

func TestDeliveryVerificationCannotPassWhenObservationHasProblems(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	mandate := deliveryMandate(request, []string{"merger"}, plan.Effects[0].Kind)
	if _, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate); err != nil {
		t.Fatal(err)
	}
	broken, _ := adapter.ObserveDelivery(context.Background(), request.Intent)
	broken.Problems = []string{"claim became ambiguous"}
	broken.Revision = "observation:ambiguous"
	adapter.SetObservation(broken)
	verification, err := lifecycle.VerifyDelivery(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState != engine.ControlNeedsReview {
		t.Fatalf("verification = %#v", verification)
	}
}

func TestDeliveryVerificationAndStatusSurviveRestart(t *testing.T) {
	lifecycle, adapter, request := deliveryFixture(t, func(observation engine.DeliveryObservation) engine.DeliveryObservation { return observation })
	inspection, _ := lifecycle.InspectDelivery(context.Background(), request)
	plan, _ := lifecycle.PlanDelivery(context.Background(), inspection)
	mandate := deliveryMandate(request, []string{"merger"}, plan.Effects[0].Kind)
	if _, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	restarted := engine.New(engine.WithClock(deliveryClock{now}), engine.WithDeliveryAdapter(adapter))

	verification, err := restarted.VerifyDelivery(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if verification.OverallState != engine.ControlPass || len(verification.Receipts) != 1 || verification.Receipts[0].MandateID != mandate.ID {
		t.Fatalf("verification = %#v", verification)
	}
	status, err := restarted.DeliveryStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Disposition != engine.DeliveryDispositionMergeReady || len(status.Receipts) != 1 {
		t.Fatalf("status = %#v", status)
	}
}

func TestDeliveryQualifyingMergeComposesWorkManagerCompletion(t *testing.T) {
	_, workAdapter, completion, now := newManagedTaskFixture(t)
	configureGovernedWorkFixture(t, &completion)
	completion.Intent.Task.Status = "done"
	completion.Intent.Task.Closed = true
	workObservation := workAdapter.Observation()
	workObservation.Task = observedGovernedTask(&completion, completion.Intent.Target.OptionIDs["readiness:ready"])
	workObservation.Task.StatusOption = completion.Intent.Target.OptionIDs["status:next"]
	workObservation.Revision = "observation:delivery-completion"
	workAdapter.SetObservation(workObservation)

	claim := engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: completion.Intent.Task.ManagedID, SourceRevision: completion.Intent.SourceRevision, ContractDigest: engine.ExecutableIssueContractDigest(completion.Intent.Governance.Issue), ImplementedSources: []engine.GovernedSourceBinding{{ID: "implementation", Path: "docs/implementation.md", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}
	deliveryObservation := readyDraftObservation()
	deliveryObservation.Issue.ManagedID = completion.Intent.Task.ManagedID
	deliveryObservation.PullRequest.Draft = false
	deliveryObservation.PullRequest.State = "closed"
	deliveryObservation.PullRequest.Merged = true
	deliveryObservation.PullRequest.MergeRevision = "merge-1"
	deliveryObservation.PullRequest.MergeMethod = "squash"
	deliveryObservation.PullRequest.DefaultReachable = true
	deliveryObservation.Checks[0].EvidenceID = "check-run:75"
	deliveryObservation.Reviews[0].EvidenceID = "review:75"
	deliveryObservation.Revision = "observation:qualifying-merge"
	deliveryAdapter := engine.NewInMemoryDeliveryAdapter(engine.DeliveryCapability{SchemaVersion: 1, Online: true, Fresh: true, Actor: "test:maintainer", Mode: "memory", RepositoryID: completion.Intent.Target.RepositoryID, Permissions: []string{"issues:write", "projects:write"}, ObservedAt: now, ExpiresAt: now.Add(time.Hour)}, deliveryObservation)
	lifecycle := engine.New(engine.WithClock(fixedWorkClock{now}), engine.WithWorkAdapter(workAdapter), engine.WithDeliveryAdapter(deliveryAdapter))
	request := engine.DeliveryRequest{Repository: completion.Repository, CompletionIntent: &completion.Intent, Intent: engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: completion.Intent.OperationID, SourceRevision: completion.Intent.SourceRevision, OperatingProfileRevision: completion.Intent.OperatingProfileRevision, Title: completion.Intent.Task.Title,
		ManagedID: completion.Intent.Task.ManagedID, Target: completion.Intent.Target, BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", RequiredChecks: []string{"foundation"},
		Review: engine.WorkReviewRequirement{Role: "reviewer", DistinctContext: true}, MergeMethod: "squash", Claim: &claim, EffectBoundary: completion.Intent.EffectBoundary,
	}}
	inspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanDelivery(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	workInspection, err := lifecycle.InspectManagedTask(context.Background(), completion)
	if err != nil {
		t.Fatal(err)
	}
	workPlan, err := lifecycle.PlanManagedTask(context.Background(), workInspection)
	if err != nil {
		t.Fatal(err)
	}
	mandate := completionMandate(request, completion, plan, workPlan, now)

	result, err := lifecycle.ApplyDelivery(context.Background(), plan.ID, plan, mandate)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.WorkApplyApplied {
		t.Fatalf("apply = %#v", result)
	}
	observed := workAdapter.Observation()
	if observed.Task == nil || !observed.Task.Closed || observed.Task.StatusOption != completion.Intent.Target.OptionIDs["status:done"] {
		t.Fatalf("completion observation = %#v", observed.Task)
	}
	completedDelivery := deliveryObservation
	completedDelivery.Issue.State = "closed"
	completedDelivery.Revision = "observation:completed"
	deliveryAdapter.SetObservation(completedDelivery)
	replayedInspection, err := lifecycle.InspectDelivery(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if replayedInspection.Disposition != engine.DeliveryDispositionComplete {
		t.Fatalf("replay disposition = %q, want complete", replayedInspection.Disposition)
	}
	status, err := lifecycle.DeliveryStatus(context.Background(), request.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if status.Completion == nil || status.Completion.IntentDigest != engine.DeliveryResourceDigest(request.Intent) || status.Completion.Checks[0].EvidenceID != "check-run:75" || status.Completion.Reviews[0].EvidenceID != "review:75" || len(status.Completion.ReconciliationReceipts) == 0 {
		t.Fatalf("completion evidence = %#v", status.Completion)
	}
	replayedPlan, err := lifecycle.PlanDelivery(context.Background(), replayedInspection)
	if err != nil {
		t.Fatal(err)
	}
	if !replayedPlan.NoChange || len(replayedPlan.Effects) != 0 {
		t.Fatalf("replay plan = %#v, want no-change", replayedPlan)
	}
}

func deliveryFixture(t *testing.T, mutate func(engine.DeliveryObservation) engine.DeliveryObservation) (*engine.Engine, *engine.InMemoryDeliveryAdapter, engine.DeliveryRequest) {
	t.Helper()
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	observation := mutate(readyDraftObservation())
	adapter := engine.NewInMemoryDeliveryAdapter(engine.DeliveryCapability{
		SchemaVersion: 1, Online: true, Fresh: true, Actor: "merger", Mode: "github-app", RepositoryID: "R_repo", Permissions: []string{"pull_requests:write"}, ObservedAt: now, ExpiresAt: now.Add(time.Hour),
	}, observation)
	lifecycle := engine.New(engine.WithClock(deliveryClock{now}), engine.WithDeliveryAdapter(adapter))
	request := engine.DeliveryRequest{Repository: t.TempDir(), Intent: engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: "deliver-75", SourceRevision: "source-1", ManagedID: "issue:75", Title: "Deliver issue 75",
		BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", RequiredChecks: []string{"foundation"},
		Review: engine.WorkReviewRequirement{Role: "reviewer", DistinctContext: true}, MergeMethod: "squash", OperatingProfileRevision: "profile-1",
		Target:         engine.WorkTarget{Host: "github.com", RepositoryID: "R_repo", ProjectID: "P_project"},
		Claim:          &engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: "issue:75", SourceRevision: "source-1", ContractDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ImplementedSources: []engine.GovernedSourceBinding{{ID: "source", Path: "docs/implementation.md", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}},
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

func completionMandate(request engine.DeliveryRequest, completion engine.ManagedTaskRequest, deliveryPlan engine.DeliveryPlan, workPlan engine.WorkPlan, now time.Time) engine.WorkExecutionMandate {
	effectKinds := []string{engine.DeliveryEffectReconcileCompletion}
	operations := []string{}
	managedIDs := []string{request.Intent.ManagedID}
	for _, effect := range workPlan.Effects {
		effectKinds = append(effectKinds, effect.Kind)
		operations = append(operations, effect.Operations...)
		managedIDs = append(managedIDs, effect.ManagedID)
	}
	return engine.BindWorkExecutionMandate(engine.WorkExecutionMandate{
		SchemaVersion: 1, ApprovedBy: "owner", ApprovalID: "approval-completion", ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
		Target: request.Intent.Target, OperationID: request.Intent.OperationID, SelectedManagedID: request.Intent.ManagedID,
		Actors: []string{"test:maintainer"}, CredentialModes: []string{"memory"}, Permissions: []string{"issues:write", "projects:write", "pull_requests:read", "contents:read"},
		Authorities: []engine.WorkExecutionAuthority{
			{Actor: "test:maintainer", CredentialMode: "memory", RepositoryID: request.Intent.Target.RepositoryID, Permissions: []string{"issues:write", "projects:write"}},
			{Actor: "test:maintainer", CredentialMode: "memory", RepositoryID: request.Intent.Target.RepositoryID, Permissions: []string{"issues:write", "projects:write", "pull_requests:read", "contents:read"}},
		},
		OperatingProfileRevisions: []string{request.Intent.OperatingProfileRevision}, ContractDigests: []string{engine.ExecutableIssueContractDigest(completion.Intent.Governance.Issue)},
		GovernanceDigests: []string{engine.GovernedWorkContractDigest(*completion.Intent.Governance)}, InputDigests: completion.Intent.InputDigests,
		GovernedSourceDigests: map[string]string{completion.Intent.Governance.Sources[0].ID: completion.Intent.Governance.Sources[0].Digest}, SourceRevisions: []string{request.Intent.SourceRevision}, ManagedIDs: managedIDs,
		EffectKinds: effectKinds, Operations: operations, ResourceDigests: []string{engine.DeliveryResourceDigest(request.Intent), engine.ManagedTaskResourceDigest(completion.Intent.Task)}, MaxEffects: 10,
		DataClass: request.Intent.EffectBoundary.DataClass, CostCeiling: request.Intent.EffectBoundary.CostCeiling, Destructive: request.Intent.EffectBoundary.Destructive,
		Retention: request.Intent.EffectBoundary.Retention, RecoveryOwner: request.Intent.EffectBoundary.RecoveryOwner,
	})
}

func readyDraftObservation() engine.DeliveryObservation {
	return engine.DeliveryObservation{
		SchemaVersion: 1,
		Revision:      "observation:draft",
		Issue:         engine.DeliveryIssueObservation{ManagedID: "issue:75", State: "open"},
		Branch:        engine.DeliveryBranchObservation{Name: "task/75-delivery-squash-completion", Revision: "head-1", Present: true},
		PullRequest: engine.DeliveryPullRequestObservation{
			Number: 101, State: "open", Draft: true, Base: "main", Head: "task/75-delivery-squash-completion", HeadRevision: "head-1",
		},
		Checks:  []engine.DeliveryCheckObservation{{Name: "foundation", HeadRevision: "head-1", State: "passed"}},
		Reviews: []engine.DeliveryReviewObservation{{Actor: "reviewer", HeadRevision: "head-1", State: "approved", DistinctContext: true, Capable: true}},
		Rules:   engine.DeliveryRulesObservation{Revision: "rules-1", BaseRevision: "base-1", RequiredChecks: []string{"foundation"}, MergeMethods: []string{"squash"}},
	}
}

type deliveryClock struct{ now time.Time }

func (clock deliveryClock) Now() time.Time { return clock.now }
