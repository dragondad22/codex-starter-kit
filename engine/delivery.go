package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

const deliveryStatePath = ".starter-kit/delivery/state.json"

type DeliveryDisposition string

const (
	DeliveryDispositionDraft            DeliveryDisposition = "draft"
	DeliveryDispositionChecksPending    DeliveryDisposition = "checks-pending"
	DeliveryDispositionReviewPending    DeliveryDisposition = "review-pending"
	DeliveryDispositionChangesRequested DeliveryDisposition = "changes-requested"
	DeliveryDispositionMergeReady       DeliveryDisposition = "merge-ready"
	DeliveryDispositionMerged           DeliveryDisposition = "merged"
	DeliveryDispositionClosedUnmerged   DeliveryDisposition = "closed-unmerged"
	DeliveryDispositionComplete         DeliveryDisposition = "complete"
	DeliveryDispositionNeedsReview      DeliveryDisposition = "needs-review"
	DeliveryEffectMarkReady                                 = "mark-ready"
	DeliveryEffectSquashMerge                               = "squash-merge"
	DeliveryEffectReconcileCompletion                       = "reconcile-completion"
)

type DeliveryIntent struct {
	SchemaVersion            int                   `json:"schema_version"`
	OperationID              string                `json:"operation_id"`
	SourceRevision           string                `json:"source_revision"`
	OperatingProfileRevision string                `json:"operating_profile_revision"`
	ManagedID                string                `json:"managed_id"`
	Target                   WorkTarget            `json:"target"`
	BaseBranch               string                `json:"base_branch"`
	HeadBranch               string                `json:"head_branch"`
	RequiredChecks           []string              `json:"required_checks"`
	Review                   WorkReviewRequirement `json:"review"`
	MergeMethod              string                `json:"merge_method"`
	Claim                    *WorkDeliveryClaim    `json:"delivery_claim"`
	EffectBoundary           WorkEffectBoundary    `json:"effect_boundary"`
}

type DeliveryRequest struct {
	Repository       string             `json:"repository"`
	Intent           DeliveryIntent     `json:"intent"`
	CompletionIntent *WorkDesiredIntent `json:"completion_intent,omitempty"`
}

type DeliveryCapability struct {
	SchemaVersion int       `json:"schema_version"`
	Online        bool      `json:"online"`
	Fresh         bool      `json:"fresh"`
	Actor         string    `json:"actor"`
	Mode          string    `json:"mode"`
	Permissions   []string  `json:"permissions"`
	ObservedAt    time.Time `json:"observed_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type DeliveryIssueObservation struct {
	ManagedID string `json:"managed_id"`
	State     string `json:"state"`
}

type DeliveryPullRequestObservation struct {
	Number           int    `json:"number"`
	State            string `json:"state"`
	Draft            bool   `json:"draft"`
	Base             string `json:"base"`
	Head             string `json:"head"`
	HeadRevision     string `json:"head_revision"`
	Merged           bool   `json:"merged"`
	MergeRevision    string `json:"merge_revision,omitempty"`
	MergeMethod      string `json:"merge_method,omitempty"`
	DefaultReachable bool   `json:"default_reachable"`
}

type DeliveryCheckObservation struct {
	Name         string `json:"name"`
	HeadRevision string `json:"head_revision"`
	State        string `json:"state"`
}

type DeliveryReviewObservation struct {
	Actor           string `json:"actor"`
	HeadRevision    string `json:"head_revision"`
	State           string `json:"state"`
	DistinctContext bool   `json:"distinct_context"`
	Capable         bool   `json:"capable"`
}

type DeliveryRulesObservation struct {
	Revision       string   `json:"revision"`
	RequiredChecks []string `json:"required_checks"`
	MergeMethods   []string `json:"merge_methods"`
}

type DeliveryObservation struct {
	SchemaVersion int                            `json:"schema_version"`
	Revision      string                         `json:"revision"`
	Issue         DeliveryIssueObservation       `json:"issue"`
	PullRequest   DeliveryPullRequestObservation `json:"pull_request"`
	Checks        []DeliveryCheckObservation     `json:"checks"`
	Reviews       []DeliveryReviewObservation    `json:"reviews"`
	Rules         DeliveryRulesObservation       `json:"rules"`
	Problems      []string                       `json:"problems,omitempty"`
}

type DeliveryInspection struct {
	SchemaVersion int                 `json:"schema_version"`
	ID            string              `json:"inspection_id"`
	Repository    string              `json:"repository"`
	Intent        DeliveryIntent      `json:"intent"`
	Capability    DeliveryCapability  `json:"capability"`
	Observation   DeliveryObservation `json:"observation"`
	Disposition   DeliveryDisposition `json:"disposition"`
	Problems      []string            `json:"problems"`
}

type DeliveryEffect struct {
	ID            string `json:"effect_id"`
	Kind          string `json:"kind"`
	PullRequest   int    `json:"pull_request"`
	HeadRevision  string `json:"head_revision"`
	MergeRevision string `json:"merge_revision,omitempty"`
}

type DeliveryPlan struct {
	SchemaVersion       int                `json:"schema_version"`
	ID                  string             `json:"plan_id"`
	Repository          string             `json:"repository"`
	Intent              DeliveryIntent     `json:"intent"`
	Capability          DeliveryCapability `json:"capability"`
	InspectionID        string             `json:"inspection_id"`
	ObservationRevision string             `json:"observation_revision"`
	Effects             []DeliveryEffect   `json:"effects"`
	NoChange            bool               `json:"no_change"`
}

type DeliveryAdapter interface {
	Capability(context.Context) (DeliveryCapability, error)
	ObserveDelivery(context.Context, DeliveryIntent) (DeliveryObservation, error)
	ApplyDelivery(context.Context, DeliveryEffect) (DeliveryEffectResult, error)
}

type DeliveryEffectResult struct {
	Outcome     string `json:"outcome"`
	Detail      string `json:"detail"`
	Recoverable bool   `json:"recoverable"`
}

type DeliveryApplyResult struct {
	SchemaVersion int                     `json:"schema_version"`
	PlanID        string                  `json:"plan_id"`
	Status        WorkApplyStatus         `json:"status"`
	Results       []DeliveryEffectResult  `json:"results"`
	Receipts      []DeliveryEffectReceipt `json:"receipts"`
}

type DeliveryEffectReceipt struct {
	SchemaVersion       int       `json:"schema_version"`
	PlanID              string    `json:"plan_id"`
	EffectID            string    `json:"effect_id"`
	EffectKind          string    `json:"effect_kind"`
	ManagedID           string    `json:"managed_id"`
	PullRequest         int       `json:"pull_request"`
	HeadRevision        string    `json:"head_revision"`
	MergeRevision       string    `json:"merge_revision,omitempty"`
	Actor               string    `json:"actor"`
	CredentialMode      string    `json:"credential_mode"`
	MandateID           string    `json:"mandate_id"`
	SourceRevision      string    `json:"source_revision"`
	ObservationRevision string    `json:"observation_revision"`
	Outcome             string    `json:"outcome"`
	Recoverable         bool      `json:"recoverable"`
	Detail              string    `json:"detail"`
	RecordedAt          time.Time `json:"recorded_at"`
}

type DeliveryVerification struct {
	SchemaVersion int                     `json:"schema_version"`
	OverallState  ControlState            `json:"overall_state"`
	Disposition   DeliveryDisposition     `json:"disposition"`
	Receipts      []DeliveryEffectReceipt `json:"receipts"`
	VerifiedAt    time.Time               `json:"verified_at"`
}

type DeliveryStatusResult struct {
	SchemaVersion int                     `json:"schema_version"`
	Disposition   DeliveryDisposition     `json:"disposition"`
	PlanID        string                  `json:"plan_id,omitempty"`
	Receipts      []DeliveryEffectReceipt `json:"receipts"`
}

type deliveryState struct {
	SchemaVersion int                     `json:"schema_version"`
	StateDigest   string                  `json:"state_digest"`
	Request       DeliveryRequest         `json:"request"`
	Inspection    DeliveryInspection      `json:"inspection"`
	Plan          *DeliveryPlan           `json:"plan,omitempty"`
	Receipts      []DeliveryEffectReceipt `json:"receipts"`
	Verification  *DeliveryVerification   `json:"verification,omitempty"`
	Completion    *DeliveryCompletion     `json:"completion,omitempty"`
	Disposition   DeliveryDisposition     `json:"disposition"`
}

type DeliveryCompletion struct {
	SchemaVersion  int       `json:"schema_version"`
	ManagedID      string    `json:"managed_id"`
	SourceRevision string    `json:"source_revision"`
	PullRequest    int       `json:"pull_request"`
	HeadRevision   string    `json:"head_revision"`
	MergeRevision  string    `json:"merge_revision"`
	MandateID      string    `json:"mandate_id"`
	RecordedAt     time.Time `json:"recorded_at"`
}

func (e *Engine) InspectDelivery(ctx context.Context, request DeliveryRequest) (DeliveryInspection, error) {
	root, err := cleanRepositoryRoot(request.Repository)
	if err != nil {
		return DeliveryInspection{}, err
	}
	if e.deliveryAdapter == nil {
		return DeliveryInspection{}, errors.New("delivery adapter is required")
	}
	capability, err := e.deliveryAdapter.Capability(ctx)
	if err != nil {
		return DeliveryInspection{}, err
	}
	observation, err := e.deliveryAdapter.ObserveDelivery(ctx, request.Intent)
	if err != nil {
		return DeliveryInspection{}, err
	}
	problems := deliveryProblems(request.Intent, capability, observation, e.clock.Now())
	if request.CompletionIntent != nil {
		completion := request.CompletionIntent
		if completion.OperationID != request.Intent.OperationID || completion.SourceRevision != request.Intent.SourceRevision || completion.OperatingProfileRevision != request.Intent.OperatingProfileRevision || completion.Task.ManagedID != request.Intent.ManagedID || !completion.Task.Closed || completion.Task.Status != "done" || !equalWorkTarget(completion.Target, request.Intent.Target) || completion.Governance == nil || request.Intent.Claim == nil || ExecutableIssueContractDigest(completion.Governance.Issue) != request.Intent.Claim.ContractDigest {
			problems = append(problems, "delivery completion intent does not match the governed outcome")
		}
	}
	disposition := DeliveryDispositionNeedsReview
	if len(problems) == 0 {
		disposition = deliveryDisposition(request.Intent, observation)
	}
	inspection := DeliveryInspection{SchemaVersion: 1, Repository: root, Intent: request.Intent, Capability: capability, Observation: observation, Disposition: disposition, Problems: problems}
	inspection.ID = digestJSON(struct {
		Repository  string
		Intent      DeliveryIntent
		Capability  DeliveryCapability
		Observation DeliveryObservation
	}{root, request.Intent, capability, observation})
	receipts := []DeliveryEffectReceipt{}
	var completion *DeliveryCompletion
	if prior, priorErr := readDeliveryState(root); priorErr == nil {
		receipts = slices.Clone(prior.Receipts)
		completion = prior.Completion
		if deliveryCompletionMatches(completion, request.Intent, observation) {
			disposition = DeliveryDispositionComplete
			inspection.Disposition = disposition
		}
	}
	if err := writeDeliveryState(root, deliveryState{SchemaVersion: 1, Request: request, Inspection: inspection, Receipts: receipts, Completion: completion, Disposition: disposition}); err != nil {
		return DeliveryInspection{}, err
	}
	return inspection, nil
}

func (e *Engine) PlanDelivery(_ context.Context, inspection DeliveryInspection) (DeliveryPlan, error) {
	if inspection.ID == "" || len(inspection.Problems) != 0 || inspection.Disposition == DeliveryDispositionNeedsReview {
		return DeliveryPlan{}, errors.New("delivery inspection is not plannable")
	}
	if slices.Contains([]DeliveryDisposition{DeliveryDispositionChecksPending, DeliveryDispositionReviewPending, DeliveryDispositionChangesRequested, DeliveryDispositionClosedUnmerged, DeliveryDispositionComplete}, inspection.Disposition) {
		plan := DeliveryPlan{SchemaVersion: 1, Repository: inspection.Repository, Intent: inspection.Intent, Capability: inspection.Capability, InspectionID: inspection.ID, ObservationRevision: inspection.Observation.Revision, NoChange: true}
		plan.ID = digestJSON(plan)
		if err := retainDeliveryPlan(inspection.Repository, inspection.ID, plan); err != nil {
			return DeliveryPlan{}, err
		}
		return plan, nil
	}
	kind := DeliveryEffectMarkReady
	if inspection.Disposition == DeliveryDispositionMergeReady {
		kind = DeliveryEffectSquashMerge
	} else if inspection.Disposition == DeliveryDispositionMerged {
		kind = DeliveryEffectReconcileCompletion
	}
	effect := DeliveryEffect{Kind: kind, PullRequest: inspection.Observation.PullRequest.Number, HeadRevision: inspection.Observation.PullRequest.HeadRevision, MergeRevision: inspection.Observation.PullRequest.MergeRevision}
	effect.ID = digestJSON(effect)
	plan := DeliveryPlan{SchemaVersion: 1, Repository: inspection.Repository, Intent: inspection.Intent, Capability: inspection.Capability, InspectionID: inspection.ID, ObservationRevision: inspection.Observation.Revision, Effects: []DeliveryEffect{effect}}
	plan.ID = digestJSON(plan)
	if err := retainDeliveryPlan(inspection.Repository, inspection.ID, plan); err != nil {
		return DeliveryPlan{}, err
	}
	return plan, nil
}

func (e *Engine) ApplyDelivery(ctx context.Context, expectedPlanID string, plan DeliveryPlan, mandate WorkExecutionMandate) (DeliveryApplyResult, error) {
	if expectedPlanID == "" || plan.ID != expectedPlanID || plan.ID != digestJSON(deliveryPlanWithoutID(plan)) {
		return DeliveryApplyResult{}, errors.New("delivery plan identity is invalid")
	}
	if plan.NoChange && len(plan.Effects) == 0 {
		state, err := readDeliveryState(plan.Repository)
		if err != nil {
			return DeliveryApplyResult{}, err
		}
		state.Disposition = state.Inspection.Disposition
		if err := writeDeliveryState(plan.Repository, state); err != nil {
			return DeliveryApplyResult{}, err
		}
		return DeliveryApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyNoChange}, nil
	}
	state, err := readDeliveryState(plan.Repository)
	if err != nil || state.Plan == nil || state.Plan.ID != plan.ID {
		return DeliveryApplyResult{}, errors.New("delivery plan is not the retained active plan")
	}
	current, observeErr := e.deliveryAdapter.ObserveDelivery(ctx, plan.Intent)
	if observeErr != nil {
		return DeliveryApplyResult{}, observeErr
	}
	if current.Revision != plan.ObservationRevision || current.PullRequest.HeadRevision != plan.Effects[0].HeadRevision {
		return DeliveryApplyResult{}, errors.New("delivery plan preconditions changed before apply")
	}
	if err := validateDeliveryMandate(mandate, plan, e.clock.Now()); err != nil {
		return DeliveryApplyResult{}, err
	}
	usage := map[string]int{}
	ledger, ledgerErr := readWorkMandateLedger(plan.Repository)
	if ledgerErr == nil {
		usage = cloneIntMap(ledger.Usage)
	} else if !errors.Is(ledgerErr, os.ErrNotExist) {
		return DeliveryApplyResult{}, ledgerErr
	}
	externalEffectCount := 0
	for _, effect := range plan.Effects {
		if effect.Kind != DeliveryEffectReconcileCompletion {
			externalEffectCount++
		}
	}
	if usage[mandate.ID]+externalEffectCount > mandate.MaxEffects {
		return DeliveryApplyResult{}, errors.New("delivery execution mandate effect ceiling is exhausted")
	}
	results := make([]DeliveryEffectResult, 0, len(plan.Effects))
	receipts := []DeliveryEffectReceipt{}
	for _, effect := range plan.Effects {
		var result DeliveryEffectResult
		var applyErr error
		if effect.Kind == DeliveryEffectReconcileCompletion {
			result, applyErr = e.applyDeliveryCompletion(ctx, state.Request, mandate)
		} else {
			usage[mandate.ID]++
			if err := writeWorkMandateLedger(plan.Repository, usage); err != nil {
				return DeliveryApplyResult{}, err
			}
			result, applyErr = e.deliveryAdapter.ApplyDelivery(ctx, effect)
		}
		if effect.Kind != DeliveryEffectReconcileCompletion && (applyErr != nil || result.Outcome == "ambiguous") {
			observed, observeErr := e.deliveryAdapter.ObserveDelivery(ctx, plan.Intent)
			if observeErr == nil && deliveryEffectObserved(effect, observed) {
				result = DeliveryEffectResult{Outcome: "applied", Detail: "recovered effect by exact postcondition observation"}
				applyErr = nil
			}
		}
		results = append(results, result)
		receipt := DeliveryEffectReceipt{
			SchemaVersion: 1, PlanID: plan.ID, EffectID: effect.ID, EffectKind: effect.Kind, ManagedID: plan.Intent.ManagedID,
			PullRequest: effect.PullRequest, HeadRevision: effect.HeadRevision, MergeRevision: effect.MergeRevision,
			Actor: plan.Capability.Actor, CredentialMode: plan.Capability.Mode, MandateID: mandate.ID, SourceRevision: plan.Intent.SourceRevision,
			ObservationRevision: plan.ObservationRevision, Outcome: result.Outcome, Recoverable: result.Recoverable, Detail: result.Detail, RecordedAt: e.clock.Now(),
		}
		state.Receipts = append(state.Receipts, receipt)
		receipts = append(receipts, receipt)
		if effect.Kind == DeliveryEffectReconcileCompletion && result.Outcome == "applied" {
			state.Completion = &DeliveryCompletion{SchemaVersion: 1, ManagedID: plan.Intent.ManagedID, SourceRevision: plan.Intent.SourceRevision, PullRequest: effect.PullRequest, HeadRevision: effect.HeadRevision, MergeRevision: effect.MergeRevision, MandateID: mandate.ID, RecordedAt: e.clock.Now()}
		}
		if applyErr != nil || result.Outcome != "applied" {
			state.Disposition = DeliveryDispositionNeedsReview
			if err := writeDeliveryState(plan.Repository, state); err != nil {
				return DeliveryApplyResult{}, err
			}
			return DeliveryApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyNonPass, Results: results, Receipts: receipts}, applyErr
		}
	}
	state.Disposition = DeliveryDispositionNeedsReview
	if err := writeDeliveryState(plan.Repository, state); err != nil {
		return DeliveryApplyResult{}, err
	}
	status := WorkApplyApplied
	if plan.NoChange {
		status = WorkApplyNoChange
	}
	return DeliveryApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: status, Results: results, Receipts: receipts}, nil
}

func deliveryCompletionMatches(completion *DeliveryCompletion, intent DeliveryIntent, observation DeliveryObservation) bool {
	if completion == nil {
		return false
	}
	pull := observation.PullRequest
	return completion.SchemaVersion == 1 && completion.ManagedID == intent.ManagedID && completion.SourceRevision == intent.SourceRevision && completion.PullRequest == pull.Number && completion.HeadRevision == pull.HeadRevision && completion.MergeRevision == pull.MergeRevision && observation.Issue.State == "closed" && pull.Merged && pull.DefaultReachable
}

func (e *Engine) applyDeliveryCompletion(ctx context.Context, request DeliveryRequest, mandate WorkExecutionMandate) (DeliveryEffectResult, error) {
	if request.CompletionIntent == nil {
		return DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery completion intent is absent", Recoverable: true}, errors.New("delivery completion intent is required")
	}
	managed := ManagedTaskRequest{Repository: request.Repository, Intent: *request.CompletionIntent, ExecutionMandate: &mandate}
	inspection, err := e.InspectManagedTask(ctx, managed)
	if err != nil {
		return DeliveryEffectResult{Outcome: "needs-review", Detail: "completion inspection did not pass", Recoverable: true}, err
	}
	plan, err := e.PlanManagedTask(ctx, inspection)
	if err != nil {
		return DeliveryEffectResult{Outcome: "needs-review", Detail: "completion plan did not pass", Recoverable: true}, err
	}
	apply, err := e.ApplyManagedTaskWithMandate(ctx, plan.ID, plan, mandate)
	if err != nil || apply.Status == WorkApplyNonPass {
		return DeliveryEffectResult{Outcome: "needs-review", Detail: "completion reconciliation did not apply", Recoverable: true}, err
	}
	verification, err := e.VerifyManagedTask(ctx, request.Repository)
	if err != nil || verification.OverallState != ControlPass {
		return DeliveryEffectResult{Outcome: "needs-review", Detail: "completion reconciliation did not verify", Recoverable: true}, err
	}
	return DeliveryEffectResult{Outcome: "applied", Detail: "reconciled qualifying merge through Work Manager"}, nil
}

func (e *Engine) VerifyDelivery(ctx context.Context, repository string) (DeliveryVerification, error) {
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return DeliveryVerification{}, err
	}
	state, err := readDeliveryState(root)
	if err != nil {
		return DeliveryVerification{}, err
	}
	observation, err := e.deliveryAdapter.ObserveDelivery(ctx, state.Request.Intent)
	if err != nil {
		return DeliveryVerification{}, err
	}
	problems := deliveryProblems(state.Request.Intent, state.Inspection.Capability, observation, e.clock.Now())
	disposition := DeliveryDispositionNeedsReview
	if len(problems) == 0 {
		disposition = deliveryDisposition(state.Request.Intent, observation)
	}
	overall := ControlNeedsReview
	if len(state.Receipts) != 0 && state.Receipts[len(state.Receipts)-1].Outcome == "applied" {
		last := state.Receipts[len(state.Receipts)-1]
		if last.EffectKind == DeliveryEffectMarkReady && disposition == DeliveryDispositionMergeReady || last.EffectKind == DeliveryEffectSquashMerge && observation.PullRequest.Merged && observation.PullRequest.DefaultReachable {
			overall = ControlPass
		}
	}
	verification := DeliveryVerification{SchemaVersion: 1, OverallState: overall, Disposition: disposition, Receipts: slices.Clone(state.Receipts), VerifiedAt: e.clock.Now()}
	state.Verification = &verification
	state.Disposition = disposition
	if err := writeDeliveryState(root, state); err != nil {
		return DeliveryVerification{}, err
	}
	return verification, nil
}

func (e *Engine) DeliveryStatus(_ context.Context, repository string) (DeliveryStatusResult, error) {
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return DeliveryStatusResult{}, err
	}
	state, err := readDeliveryState(root)
	if err != nil {
		return DeliveryStatusResult{}, err
	}
	planID := ""
	if state.Plan != nil {
		planID = state.Plan.ID
	}
	return DeliveryStatusResult{SchemaVersion: 1, Disposition: state.Disposition, PlanID: planID, Receipts: slices.Clone(state.Receipts)}, nil
}

func retainDeliveryPlan(repository, inspectionID string, plan DeliveryPlan) error {
	state, err := readDeliveryState(repository)
	if err != nil {
		return err
	}
	if state.Inspection.ID != inspectionID {
		return errors.New("delivery inspection is not the retained active inspection")
	}
	state.Plan = &plan
	return writeDeliveryState(repository, state)
}

func deliveryEffectObserved(effect DeliveryEffect, observation DeliveryObservation) bool {
	pull := observation.PullRequest
	if pull.Number != effect.PullRequest || pull.HeadRevision != effect.HeadRevision {
		return false
	}
	switch effect.Kind {
	case DeliveryEffectMarkReady:
		return pull.State == "open" && !pull.Draft && !pull.Merged
	case DeliveryEffectSquashMerge:
		return pull.State == "closed" && pull.Merged && pull.MergeMethod == "squash" && pull.DefaultReachable
	default:
		return false
	}
}

func DeliveryResourceDigest(intent DeliveryIntent) string {
	return digestJSON(intent)
}

func validateDeliveryMandate(mandate WorkExecutionMandate, plan DeliveryPlan, now time.Time) error {
	intent := plan.Intent
	if mandate.SchemaVersion != 1 || mandate.ID == "" || mandate.ID != BindWorkExecutionMandate(mandate).ID || mandate.ApprovedBy == "" || mandate.ApprovalID == "" || now.Before(mandate.ApprovedAt) || !now.Before(mandate.ExpiresAt) {
		return errors.New("delivery execution mandate is required or expired")
	}
	if !equalWorkTarget(mandate.Target, intent.Target) || mandate.OperationID != intent.OperationID || mandate.SelectedManagedID != intent.ManagedID || !slices.Contains(mandate.Actors, plan.Capability.Actor) || !slices.Contains(mandate.CredentialModes, plan.Capability.Mode) || !slices.Contains(mandate.SourceRevisions, intent.SourceRevision) || !slices.Contains(mandate.OperatingProfileRevisions, intent.OperatingProfileRevision) || !slices.Contains(mandate.ManagedIDs, intent.ManagedID) || !slices.Contains(mandate.ResourceDigests, DeliveryResourceDigest(intent)) {
		return errors.New("delivery plan is outside execution mandate identity")
	}
	if mandate.DataClass != intent.EffectBoundary.DataClass || mandate.CostCeiling != intent.EffectBoundary.CostCeiling || mandate.Destructive != intent.EffectBoundary.Destructive || mandate.Retention != intent.EffectBoundary.Retention || mandate.RecoveryOwner != intent.EffectBoundary.RecoveryOwner || mandate.MaxEffects < len(plan.Effects) {
		return errors.New("delivery plan exceeds execution mandate boundary")
	}
	for _, effect := range plan.Effects {
		if !slices.Contains(mandate.EffectKinds, effect.Kind) {
			return errors.New("delivery effect is outside execution mandate")
		}
	}
	return nil
}

func deliveryPlanWithoutID(plan DeliveryPlan) DeliveryPlan {
	plan.ID = ""
	return plan
}

func deliveryProblems(intent DeliveryIntent, capability DeliveryCapability, observation DeliveryObservation, now time.Time) []string {
	problems := slices.Clone(observation.Problems)
	claimValid := intent.Claim != nil && intent.Claim.ManagedID == intent.ManagedID && intent.Claim.SourceRevision == intent.SourceRevision
	if claimValid {
		_, claimErr := RenderWorkDeliveryClaim(*intent.Claim)
		claimValid = claimErr == nil
	}
	if intent.SchemaVersion != 1 || intent.OperationID == "" || intent.SourceRevision == "" || intent.OperatingProfileRevision == "" || intent.ManagedID == "" || intent.Target.RepositoryID == "" || intent.BaseBranch == "" || intent.HeadBranch == "" || intent.MergeMethod != "squash" || !claimValid {
		problems = append(problems, "delivery intent is invalid")
	}
	if capability.SchemaVersion != 1 || !capability.Online || !capability.Fresh || capability.Actor == "" || capability.Mode == "" || capability.ExpiresAt.IsZero() || !now.Before(capability.ExpiresAt) {
		problems = append(problems, "delivery capability is unavailable or stale")
	}
	pr := observation.PullRequest
	validPullState := pr.State == "open" && !pr.Merged || pr.State == "closed"
	validIssueState := observation.Issue.State == "open" || observation.Issue.State == "closed" && pr.Merged
	if observation.SchemaVersion != 1 || observation.Revision == "" || observation.Issue.ManagedID != intent.ManagedID || !validIssueState || pr.Number <= 0 || !validPullState || pr.Base != intent.BaseBranch || pr.Head != intent.HeadBranch || pr.HeadRevision == "" {
		problems = append(problems, "delivery linkage is incomplete or ambiguous")
	}
	if !slices.Contains(observation.Rules.MergeMethods, intent.MergeMethod) || !slices.Equal(observation.Rules.RequiredChecks, intent.RequiredChecks) {
		problems = append(problems, "effective rules do not match governed delivery intent")
	}
	return problems
}

func deliveryDisposition(intent DeliveryIntent, observation DeliveryObservation) DeliveryDisposition {
	pr := observation.PullRequest
	if pr.Merged {
		if pr.MergeMethod != intent.MergeMethod || pr.MergeRevision == "" || !pr.DefaultReachable {
			return DeliveryDispositionNeedsReview
		}
		return DeliveryDispositionMerged
	}
	if pr.State == "closed" {
		return DeliveryDispositionClosedUnmerged
	}
	for _, required := range intent.RequiredChecks {
		if !slices.ContainsFunc(observation.Checks, func(check DeliveryCheckObservation) bool {
			return check.Name == required && check.HeadRevision == pr.HeadRevision && check.State == "passed"
		}) {
			return DeliveryDispositionChecksPending
		}
	}
	if slices.ContainsFunc(observation.Reviews, func(review DeliveryReviewObservation) bool {
		return review.HeadRevision == pr.HeadRevision && review.State == "changes-requested"
	}) {
		return DeliveryDispositionChangesRequested
	}
	if intent.Review.Role != "" && !slices.ContainsFunc(observation.Reviews, func(review DeliveryReviewObservation) bool {
		return review.Actor == intent.Review.Role && review.HeadRevision == pr.HeadRevision && review.State == "approved" && (!intent.Review.DistinctContext || review.DistinctContext) && review.Capable
	}) {
		return DeliveryDispositionReviewPending
	}
	if pr.Draft {
		return DeliveryDispositionDraft
	}
	return DeliveryDispositionMergeReady
}

type InMemoryDeliveryAdapter struct {
	capability  DeliveryCapability
	observation DeliveryObservation
	queued      []queuedDeliveryResult
	applyCount  int
}

type queuedDeliveryResult struct {
	result        DeliveryEffectResult
	observeEffect bool
	err           error
}

func NewInMemoryDeliveryAdapter(capability DeliveryCapability, observation DeliveryObservation) *InMemoryDeliveryAdapter {
	return &InMemoryDeliveryAdapter{capability: capability, observation: observation}
}

func (adapter *InMemoryDeliveryAdapter) Capability(context.Context) (DeliveryCapability, error) {
	return adapter.capability, nil
}

func (adapter *InMemoryDeliveryAdapter) ObserveDelivery(context.Context, DeliveryIntent) (DeliveryObservation, error) {
	return adapter.observation, nil
}

func (adapter *InMemoryDeliveryAdapter) SetObservation(observation DeliveryObservation) {
	adapter.observation = observation
}

func (adapter *InMemoryDeliveryAdapter) ApplyDelivery(_ context.Context, effect DeliveryEffect) (DeliveryEffectResult, error) {
	adapter.applyCount++
	if len(adapter.queued) != 0 {
		queued := adapter.queued[0]
		adapter.queued = adapter.queued[1:]
		if queued.observeEffect {
			adapter.applyObservedEffect(effect)
		}
		return queued.result, queued.err
	}
	return adapter.applyObservedEffect(effect)
}

func (adapter *InMemoryDeliveryAdapter) applyObservedEffect(effect DeliveryEffect) (DeliveryEffectResult, error) {
	switch effect.Kind {
	case DeliveryEffectMarkReady:
		adapter.observation.PullRequest.Draft = false
		adapter.observation.Revision = digestJSON(adapter.observation)
		return DeliveryEffectResult{Outcome: "applied", Detail: "marked pull request ready"}, nil
	case DeliveryEffectSquashMerge:
		adapter.observation.PullRequest.State = "closed"
		adapter.observation.PullRequest.Merged = true
		adapter.observation.PullRequest.MergeRevision = "merge:" + effect.HeadRevision
		adapter.observation.PullRequest.MergeMethod = "squash"
		adapter.observation.PullRequest.DefaultReachable = true
		adapter.observation.Revision = digestJSON(adapter.observation)
		return DeliveryEffectResult{Outcome: "applied", Detail: "squash merged exact head"}, nil
	default:
		return DeliveryEffectResult{Outcome: "needs-review", Detail: "unsupported delivery effect", Recoverable: true}, errors.New("unsupported delivery effect")
	}
}

func (adapter *InMemoryDeliveryAdapter) QueueApplyResult(result DeliveryEffectResult, observeEffect bool, err error) {
	adapter.queued = append(adapter.queued, queuedDeliveryResult{result: result, observeEffect: observeEffect, err: err})
}

func (adapter *InMemoryDeliveryAdapter) ApplyCount() int {
	return adapter.applyCount
}

func writeDeliveryState(root string, state deliveryState) error {
	path := filepath.Join(root, filepath.FromSlash(deliveryStatePath))
	if err := ensureNoSymlinkParents(root, deliveryStatePath); err != nil {
		return fmt.Errorf("validate delivery state path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create delivery state directory: %w", err)
	}
	state.StateDigest = ""
	state.StateDigest = digestJSON(state)
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode delivery state: %w", err)
	}
	if containsSensitiveText(string(content)) {
		return errors.New("delivery state contains sensitive-looking material")
	}
	content = append(content, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".delivery-state-*.tmp")
	if err != nil {
		return fmt.Errorf("create delivery state staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write delivery state: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync delivery state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("commit delivery state: %w", err)
	}
	return nil
}

func readDeliveryState(root string) (deliveryState, error) {
	if err := ensureNoSymlinkComponents(root, deliveryStatePath); err != nil {
		return deliveryState{}, fmt.Errorf("validate delivery state path: %w", err)
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(deliveryStatePath)))
	if err != nil {
		return deliveryState{}, fmt.Errorf("read delivery state: %w", err)
	}
	var state deliveryState
	if err := json.Unmarshal(content, &state); err != nil {
		return deliveryState{}, fmt.Errorf("parse delivery state: %w", err)
	}
	if state.SchemaVersion != 1 {
		return deliveryState{}, errors.New("unsupported delivery state schema")
	}
	recorded := state.StateDigest
	state.StateDigest = ""
	if recorded == "" || recorded != digestJSON(state) {
		return deliveryState{}, errors.New("delivery state integrity is invalid")
	}
	state.StateDigest = recorded
	return state, nil
}
