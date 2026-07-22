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
	DeliveryDispositionDraft               DeliveryDisposition = "draft"
	DeliveryDispositionBranchAbsent        DeliveryDisposition = "branch-absent"
	DeliveryDispositionPullRequestAbsent   DeliveryDisposition = "pull-request-absent"
	DeliveryDispositionChecksPending       DeliveryDisposition = "checks-pending"
	DeliveryDispositionChecksFailed        DeliveryDisposition = "checks-failed"
	DeliveryDispositionReviewPending       DeliveryDisposition = "review-pending"
	DeliveryDispositionReviewUnrequested   DeliveryDisposition = "review-unrequested"
	DeliveryDispositionApprovalPending     DeliveryDisposition = "approval-pending"
	DeliveryDispositionApprovalUnrequested DeliveryDisposition = "approval-unrequested"
	DeliveryDispositionChangesRequested    DeliveryDisposition = "changes-requested"
	DeliveryDispositionMergeReady          DeliveryDisposition = "merge-ready"
	DeliveryDispositionMerged              DeliveryDisposition = "merged"
	DeliveryDispositionClosedUnmerged      DeliveryDisposition = "closed-unmerged"
	DeliveryDispositionComplete            DeliveryDisposition = "complete"
	DeliveryDispositionNeedsReview         DeliveryDisposition = "needs-review"
	DeliveryEffectMarkReady                                    = "mark-ready"
	DeliveryEffectCreateBranch                                 = "create-branch"
	DeliveryEffectCreatePullRequest                            = "create-pull-request"
	DeliveryEffectRequestReview                                = "request-review"
	DeliveryEffectSquashMerge                                  = "squash-merge"
	DeliveryEffectReconcileCompletion                          = "reconcile-completion"
)

type DeliveryIntent struct {
	SchemaVersion            int                       `json:"schema_version"`
	OperationID              string                    `json:"operation_id"`
	SourceRevision           string                    `json:"source_revision"`
	OperatingProfileRevision string                    `json:"operating_profile_revision"`
	ManagedID                string                    `json:"managed_id"`
	Title                    string                    `json:"title"`
	Target                   WorkTarget                `json:"target"`
	BaseBranch               string                    `json:"base_branch"`
	HeadBranch               string                    `json:"head_branch"`
	RequiredChecks           []DeliveryCheckIdentity   `json:"required_checks"`
	Review                   DeliveryReviewDeclaration `json:"review"`
	ProductApproval          WorkReviewRequirement     `json:"product_approval,omitempty"`
	MergeMethod              string                    `json:"merge_method"`
	Claim                    *WorkDeliveryClaim        `json:"delivery_claim"`
	EffectBoundary           WorkEffectBoundary        `json:"effect_boundary"`
}

// DeliveryCheckIdentity preserves the provider identity required by branch rules.
// IntegrationID zero denotes an explicitly legacy, unbound status context.
type DeliveryCheckIdentity struct {
	Name          string `json:"name"`
	IntegrationID int64  `json:"integration_id,omitempty"`
}

// DeliveryReviewDeclaration is the governed review route for one exact implementation.
// A distinct context is expressed by unequal named contexts, not a self-attested boolean.
type DeliveryReviewDeclaration struct {
	Actor                  string   `json:"actor"`
	Role                   string   `json:"role"`
	Capability             string   `json:"capability"`
	ReviewedSourceRevision string   `json:"reviewed_source_revision"`
	ImplementationContext  string   `json:"implementation_context"`
	ReviewContext          string   `json:"review_context"`
	ApprovalRoute          string   `json:"approval_route"`
	FindingsRoute          string   `json:"findings_route"`
	Limitations            []string `json:"limitations"`
	StrongerPolicyRequired bool     `json:"stronger_policy_required"`
}

type DeliveryRequest struct {
	Repository       string             `json:"repository"`
	Intent           DeliveryIntent     `json:"intent"`
	CompletionIntent *WorkDesiredIntent `json:"completion_intent,omitempty"`
}

type DeliveryCapability struct {
	SchemaVersion  int       `json:"schema_version"`
	Online         bool      `json:"online"`
	Fresh          bool      `json:"fresh"`
	Actor          string    `json:"actor"`
	Mode           string    `json:"mode"`
	Account        string    `json:"account,omitempty"`
	InstallationID string    `json:"installation_id,omitempty"`
	RepositoryID   string    `json:"repository_id"`
	Permissions    []string  `json:"permissions"`
	ObservedAt     time.Time `json:"observed_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type DeliveryIssueObservation struct {
	ManagedID string `json:"managed_id"`
	Number    int    `json:"number"`
	State     string `json:"state"`
}

type DeliveryBranchObservation struct {
	Name       string `json:"name"`
	Revision   string `json:"revision"`
	Present    bool   `json:"present"`
	Historical bool   `json:"historical"`
}

type DeliveryPullRequestObservation struct {
	ID                 int64    `json:"id"`
	NodeID             string   `json:"node_id"`
	Number             int      `json:"number"`
	State              string   `json:"state"`
	Draft              bool     `json:"draft"`
	Base               string   `json:"base"`
	Head               string   `json:"head"`
	HeadRevision       string   `json:"head_revision"`
	Merged             bool     `json:"merged"`
	MergeRevision      string   `json:"merge_revision,omitempty"`
	MergeMethod        string   `json:"merge_method,omitempty"`
	DefaultReachable   bool     `json:"default_reachable"`
	RequestedReviewers []string `json:"requested_reviewers"`
	ClosesIssueNumber  int      `json:"closes_issue_number,omitempty"`
}

type DeliveryCheckObservation struct {
	Name          string    `json:"name"`
	IntegrationID int64     `json:"integration_id,omitempty"`
	HeadRevision  string    `json:"head_revision"`
	State         string    `json:"state"`
	EvidenceID    string    `json:"evidence_id,omitempty"`
	ObservedAt    time.Time `json:"observed_at,omitempty"`
}

type DeliveryReviewObservation struct {
	Actor                   string    `json:"actor"`
	HeadRevision            string    `json:"head_revision"`
	State                   string    `json:"state"`
	Role                    string    `json:"role"`
	Capability              string    `json:"capability"`
	ReviewedSourceRevision  string    `json:"reviewed_source_revision"`
	ImplementationContext   string    `json:"implementation_context"`
	ReviewContext           string    `json:"review_context"`
	ApprovalRoute           string    `json:"approval_route"`
	FindingsRoute           string    `json:"findings_route"`
	Limitations             []string  `json:"limitations"`
	StrongerPolicySatisfied bool      `json:"stronger_policy_satisfied"`
	EvidenceID              string    `json:"evidence_id,omitempty"`
	ObservedAt              time.Time `json:"observed_at,omitempty"`
}

type DeliveryApprovalObservation struct {
	Actor                string    `json:"actor"`
	HeadRevision         string    `json:"head_revision"`
	State                string    `json:"state"`
	DistinctContext      bool      `json:"distinct_context"`
	Capable              bool      `json:"capable"`
	QualifiedIndependent bool      `json:"qualified_independent"`
	EvidenceID           string    `json:"evidence_id,omitempty"`
	ObservedAt           time.Time `json:"observed_at,omitempty"`
}

type DeliveryRulesObservation struct {
	Revision       string                  `json:"revision"`
	BaseRevision   string                  `json:"base_revision"`
	RequiredChecks []DeliveryCheckIdentity `json:"required_checks"`
	MergeMethods   []string                `json:"merge_methods"`
	Problems       []string                `json:"problems,omitempty"`
}

type DeliveryObservation struct {
	SchemaVersion int                            `json:"schema_version"`
	Revision      string                         `json:"revision"`
	Issue         DeliveryIssueObservation       `json:"issue"`
	Branch        DeliveryBranchObservation      `json:"branch"`
	PullRequest   DeliveryPullRequestObservation `json:"pull_request"`
	Checks        []DeliveryCheckObservation     `json:"checks"`
	Reviews       []DeliveryReviewObservation    `json:"reviews"`
	Approvals     []DeliveryApprovalObservation  `json:"approvals"`
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
	ID            string             `json:"effect_id"`
	Kind          string             `json:"kind"`
	PullRequest   int                `json:"pull_request"`
	HeadRevision  string             `json:"head_revision"`
	MergeRevision string             `json:"merge_revision,omitempty"`
	Branch        string             `json:"branch,omitempty"`
	BaseBranch    string             `json:"base_branch,omitempty"`
	Claim         *WorkDeliveryClaim `json:"delivery_claim,omitempty"`
	Reviewer      string             `json:"reviewer,omitempty"`
	Title         string             `json:"title,omitempty"`
	IssueNumber   int                `json:"issue_number,omitempty"`
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
	ApplyDelivery(context.Context, DeliveryEffect, DeliveryCapability) (DeliveryEffectResult, error)
}

type DeliveryEffectResult struct {
	Outcome          string              `json:"outcome"`
	Detail           string              `json:"detail"`
	Recoverable      bool                `json:"recoverable"`
	ResourceRevision string              `json:"resource_revision,omitempty"`
	WorkReceipts     []WorkEffectReceipt `json:"work_receipts,omitempty"`
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
	SchemaVersion      int                     `json:"schema_version"`
	Disposition        DeliveryDisposition     `json:"disposition"`
	PlanID             string                  `json:"plan_id,omitempty"`
	Receipts           []DeliveryEffectReceipt `json:"receipts"`
	HistoricalReceipts []DeliveryEffectReceipt `json:"historical_receipts,omitempty"`
	Completion         *DeliveryCompletion     `json:"completion,omitempty"`
}

type deliveryState struct {
	SchemaVersion      int                     `json:"schema_version"`
	StateDigest        string                  `json:"state_digest"`
	Request            DeliveryRequest         `json:"request"`
	Inspection         DeliveryInspection      `json:"inspection"`
	Plan               *DeliveryPlan           `json:"plan,omitempty"`
	Receipts           []DeliveryEffectReceipt `json:"receipts"`
	HistoricalReceipts []DeliveryEffectReceipt `json:"historical_receipts,omitempty"`
	Verification       *DeliveryVerification   `json:"verification,omitempty"`
	Completion         *DeliveryCompletion     `json:"completion,omitempty"`
	Disposition        DeliveryDisposition     `json:"disposition"`
}

type DeliveryCompletion struct {
	SchemaVersion          int                           `json:"schema_version"`
	ManagedID              string                        `json:"managed_id"`
	SourceRevision         string                        `json:"source_revision"`
	PullRequest            int                           `json:"pull_request"`
	HeadRevision           string                        `json:"head_revision"`
	MergeRevision          string                        `json:"merge_revision"`
	MandateID              string                        `json:"mandate_id"`
	IntentDigest           string                        `json:"intent_digest"`
	Checks                 []DeliveryCheckObservation    `json:"checks"`
	Reviews                []DeliveryReviewObservation   `json:"reviews"`
	Approvals              []DeliveryApprovalObservation `json:"approvals"`
	Rules                  DeliveryRulesObservation      `json:"rules"`
	ReconciliationReceipts []WorkEffectReceipt           `json:"reconciliation_receipts"`
	RecordedAt             time.Time                     `json:"recorded_at"`
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
	var prior deliveryState
	havePrior := false
	historicalReceipts := []DeliveryEffectReceipt{}
	if retained, priorErr := readDeliveryState(root); priorErr == nil {
		prior = retained
		historicalReceipts = slices.Clone(prior.HistoricalReceipts)
		if DeliveryResourceDigest(prior.Request.Intent) == DeliveryResourceDigest(request.Intent) {
			havePrior = true
			if deliverySquashReceiptMatches(prior.Receipts, request.Intent, observation) {
				observation.PullRequest.MergeMethod = request.Intent.MergeMethod
				observation.Revision = digestJSON(observation)
			}
		} else {
			historicalReceipts = append(historicalReceipts, prior.Receipts...)
		}
	} else if !errors.Is(priorErr, os.ErrNotExist) {
		return DeliveryInspection{}, priorErr
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
	if havePrior {
		receipts = slices.Clone(prior.Receipts)
		completion = prior.Completion
		if deliveryCompletionMatches(completion, request.Intent, observation) {
			disposition = DeliveryDispositionComplete
			inspection.Disposition = disposition
		}
	}
	if err := writeDeliveryState(root, deliveryState{SchemaVersion: 1, Request: request, Inspection: inspection, Receipts: receipts, HistoricalReceipts: historicalReceipts, Completion: completion, Disposition: disposition}); err != nil {
		return DeliveryInspection{}, err
	}
	return inspection, nil
}

func (e *Engine) PlanDelivery(_ context.Context, inspection DeliveryInspection) (DeliveryPlan, error) {
	if inspection.ID == "" || len(inspection.Problems) != 0 || inspection.Disposition == DeliveryDispositionNeedsReview {
		return DeliveryPlan{}, errors.New("delivery inspection is not plannable")
	}
	if slices.Contains([]DeliveryDisposition{DeliveryDispositionChecksPending, DeliveryDispositionChecksFailed, DeliveryDispositionReviewPending, DeliveryDispositionApprovalPending, DeliveryDispositionChangesRequested, DeliveryDispositionClosedUnmerged, DeliveryDispositionComplete}, inspection.Disposition) {
		plan := DeliveryPlan{SchemaVersion: 1, Repository: inspection.Repository, Intent: inspection.Intent, Capability: inspection.Capability, InspectionID: inspection.ID, ObservationRevision: inspection.Observation.Revision, NoChange: true}
		plan.ID = digestJSON(plan)
		if err := retainDeliveryPlan(inspection.Repository, inspection.ID, plan); err != nil {
			return DeliveryPlan{}, err
		}
		return plan, nil
	}
	kind := DeliveryEffectMarkReady
	effect := DeliveryEffect{PullRequest: inspection.Observation.PullRequest.Number, HeadRevision: inspection.Observation.PullRequest.HeadRevision, MergeRevision: inspection.Observation.PullRequest.MergeRevision}
	if inspection.Disposition == DeliveryDispositionBranchAbsent {
		kind = DeliveryEffectCreateBranch
		effect.Branch = inspection.Intent.HeadBranch
		effect.BaseBranch = inspection.Intent.BaseBranch
		effect.HeadRevision = inspection.Observation.Rules.BaseRevision
	} else if inspection.Disposition == DeliveryDispositionPullRequestAbsent {
		kind = DeliveryEffectCreatePullRequest
		effect.Branch = inspection.Intent.HeadBranch
		effect.BaseBranch = inspection.Intent.BaseBranch
		effect.HeadRevision = inspection.Observation.Branch.Revision
		effect.Claim = inspection.Intent.Claim
		effect.Title = inspection.Intent.Title
		effect.IssueNumber = inspection.Observation.Issue.Number
	} else if inspection.Disposition == DeliveryDispositionReviewUnrequested {
		kind = DeliveryEffectRequestReview
		effect.Reviewer = inspection.Intent.Review.Actor
	} else if inspection.Disposition == DeliveryDispositionApprovalUnrequested {
		kind = DeliveryEffectRequestReview
		effect.Reviewer = inspection.Intent.ProductApproval.Role
	} else if inspection.Disposition == DeliveryDispositionMergeReady {
		kind = DeliveryEffectSquashMerge
	} else if inspection.Disposition == DeliveryDispositionMerged {
		kind = DeliveryEffectReconcileCompletion
	}
	effect.Kind = kind
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
	currentCapability, capabilityErr := e.deliveryAdapter.Capability(ctx)
	if capabilityErr != nil {
		return DeliveryApplyResult{}, capabilityErr
	}
	if !sameDeliveryCapabilityAuthority(plan.Capability, currentCapability, e.clock.Now()) {
		return DeliveryApplyResult{}, errors.New("delivery capability changed before apply")
	}
	current, observeErr := e.deliveryAdapter.ObserveDelivery(ctx, plan.Intent)
	if observeErr != nil {
		return DeliveryApplyResult{}, observeErr
	}
	if current.Revision != plan.ObservationRevision {
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
			result, applyErr = e.deliveryAdapter.ApplyDelivery(ctx, effect, currentCapability)
		}
		if effect.Kind != DeliveryEffectReconcileCompletion && (applyErr != nil || result.Outcome == "ambiguous") {
			observed, observeErr := e.deliveryAdapter.ObserveDelivery(ctx, plan.Intent)
			if observeErr == nil && deliveryEffectObserved(effect, observed) {
				result = DeliveryEffectResult{Outcome: "applied", Detail: "recovered effect by exact postcondition observation"}
				applyErr = nil
			}
		}
		results = append(results, result)
		if effect.Kind == DeliveryEffectSquashMerge && result.ResourceRevision != "" {
			effect.MergeRevision = result.ResourceRevision
		}
		receipt := DeliveryEffectReceipt{
			SchemaVersion: 1, PlanID: plan.ID, EffectID: effect.ID, EffectKind: effect.Kind, ManagedID: plan.Intent.ManagedID,
			PullRequest: effect.PullRequest, HeadRevision: effect.HeadRevision, MergeRevision: effect.MergeRevision,
			Actor: plan.Capability.Actor, CredentialMode: plan.Capability.Mode, MandateID: mandate.ID, SourceRevision: plan.Intent.SourceRevision,
			ObservationRevision: plan.ObservationRevision, Outcome: result.Outcome, Recoverable: result.Recoverable, Detail: result.Detail, RecordedAt: e.clock.Now(),
		}
		state.Receipts = append(state.Receipts, receipt)
		receipts = append(receipts, receipt)
		if effect.Kind == DeliveryEffectReconcileCompletion && result.Outcome == "applied" {
			state.Completion = &DeliveryCompletion{SchemaVersion: 1, ManagedID: plan.Intent.ManagedID, SourceRevision: plan.Intent.SourceRevision, PullRequest: effect.PullRequest, HeadRevision: effect.HeadRevision, MergeRevision: effect.MergeRevision, MandateID: mandate.ID, IntentDigest: DeliveryResourceDigest(plan.Intent), Checks: slices.Clone(state.Inspection.Observation.Checks), Reviews: slices.Clone(state.Inspection.Observation.Reviews), Approvals: slices.Clone(state.Inspection.Observation.Approvals), Rules: state.Inspection.Observation.Rules, ReconciliationReceipts: slices.Clone(result.WorkReceipts), RecordedAt: e.clock.Now()}
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
	return completion.SchemaVersion == 1 && completion.IntentDigest == DeliveryResourceDigest(intent) && completion.ManagedID == intent.ManagedID && completion.SourceRevision == intent.SourceRevision && completion.PullRequest == pull.Number && completion.HeadRevision == pull.HeadRevision && completion.MergeRevision == pull.MergeRevision && observation.Issue.State == "closed" && pull.Merged && pull.DefaultReachable
}

func deliverySquashReceiptMatches(receipts []DeliveryEffectReceipt, intent DeliveryIntent, observation DeliveryObservation) bool {
	pull := observation.PullRequest
	return pull.Merged && pull.DefaultReachable && pull.MergeRevision != "" && slices.ContainsFunc(receipts, func(receipt DeliveryEffectReceipt) bool {
		return receipt.EffectKind == DeliveryEffectSquashMerge && receipt.Outcome == "applied" && receipt.ManagedID == intent.ManagedID && receipt.SourceRevision == intent.SourceRevision && receipt.PullRequest == pull.Number && receipt.HeadRevision == pull.HeadRevision && receipt.MergeRevision == pull.MergeRevision
	})
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
	return DeliveryEffectResult{Outcome: "applied", Detail: "reconciled qualifying merge through Work Manager", WorkReceipts: slices.Clone(apply.Receipts)}, nil
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
	if deliverySquashReceiptMatches(state.Receipts, state.Request.Intent, observation) {
		observation.PullRequest.MergeMethod = state.Request.Intent.MergeMethod
	}
	problems := deliveryProblems(state.Request.Intent, state.Inspection.Capability, observation, e.clock.Now())
	disposition := DeliveryDispositionNeedsReview
	if len(problems) == 0 {
		disposition = deliveryDisposition(state.Request.Intent, observation)
	}
	overall := ControlNeedsReview
	if len(problems) == 0 && len(state.Receipts) != 0 && state.Receipts[len(state.Receipts)-1].Outcome == "applied" {
		last := state.Receipts[len(state.Receipts)-1]
		if last.EffectKind == DeliveryEffectReconcileCompletion && deliveryCompletionMatches(state.Completion, state.Request.Intent, observation) {
			disposition = DeliveryDispositionComplete
			overall = ControlPass
		} else if last.EffectKind == DeliveryEffectSquashMerge && disposition == DeliveryDispositionMerged {
			overall = ControlPass
		} else if state.Plan != nil {
			for _, effect := range state.Plan.Effects {
				if effect.ID == last.EffectID && deliveryEffectObserved(effect, observation) {
					overall = ControlPass
				}
			}
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
	return DeliveryStatusResult{SchemaVersion: 1, Disposition: state.Disposition, PlanID: planID, Receipts: slices.Clone(state.Receipts), HistoricalReceipts: slices.Clone(state.HistoricalReceipts), Completion: state.Completion}, nil
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
	if len(observation.Problems) != 0 {
		return false
	}
	pull := observation.PullRequest
	switch effect.Kind {
	case DeliveryEffectCreateBranch:
		return observation.Branch.Name == effect.Branch && observation.Branch.Revision == effect.HeadRevision
	case DeliveryEffectCreatePullRequest:
		return effect.IssueNumber > 0 && pull.ClosesIssueNumber == effect.IssueNumber && pull.Number > 0 && pull.State == "open" && pull.Draft && pull.Head == effect.Branch && pull.Base == effect.BaseBranch && pull.HeadRevision == effect.HeadRevision
	case DeliveryEffectRequestReview:
		return pull.Number == effect.PullRequest && pull.HeadRevision == effect.HeadRevision && (slices.Contains(pull.RequestedReviewers, effect.Reviewer) || slices.ContainsFunc(observation.Reviews, func(review DeliveryReviewObservation) bool {
			return review.Actor == effect.Reviewer && review.HeadRevision == effect.HeadRevision
		}))
	case DeliveryEffectMarkReady:
		return pull.Number == effect.PullRequest && pull.HeadRevision == effect.HeadRevision && pull.State == "open" && !pull.Draft && !pull.Merged
	case DeliveryEffectSquashMerge:
		return false
	default:
		return false
	}
}

func sameDeliveryCapabilityAuthority(planned, current DeliveryCapability, now time.Time) bool {
	leftPermissions := slices.Clone(planned.Permissions)
	rightPermissions := slices.Clone(current.Permissions)
	slices.Sort(leftPermissions)
	slices.Sort(rightPermissions)
	return current.SchemaVersion == 1 && current.Online && current.Fresh && now.Before(current.ExpiresAt) &&
		planned.Actor == current.Actor && planned.Mode == current.Mode && planned.Account == current.Account &&
		planned.InstallationID == current.InstallationID && planned.RepositoryID == current.RepositoryID &&
		planned.ExpiresAt.Equal(current.ExpiresAt) && slices.Equal(leftPermissions, rightPermissions)
}

func DeliveryResourceDigest(intent DeliveryIntent) string {
	return digestJSON(intent)
}

func validDeliveryReviewDeclaration(review DeliveryReviewDeclaration, intent DeliveryIntent) bool {
	if review.Actor == "" || review.Role == "" || review.Capability == "" || review.ReviewedSourceRevision != intent.SourceRevision || review.ImplementationContext == "" || review.ReviewContext == "" || review.ReviewContext == review.ImplementationContext || review.ApprovalRoute == "" || review.FindingsRoute == "" || len(review.Limitations) == 0 {
		return false
	}
	for _, limitation := range review.Limitations {
		if limitation == "" {
			return false
		}
	}
	return true
}

func reviewEvidenceMatchesDeclaration(evidence DeliveryReviewObservation, declaration DeliveryReviewDeclaration) bool {
	return evidence.Actor == declaration.Actor && evidence.Role == declaration.Role && evidence.Capability == declaration.Capability &&
		evidence.ReviewedSourceRevision == declaration.ReviewedSourceRevision && evidence.ImplementationContext == declaration.ImplementationContext &&
		evidence.ReviewContext == declaration.ReviewContext && evidence.ReviewContext != evidence.ImplementationContext &&
		evidence.ApprovalRoute == declaration.ApprovalRoute && evidence.FindingsRoute == declaration.FindingsRoute &&
		slices.Equal(evidence.Limitations, declaration.Limitations) && (!declaration.StrongerPolicyRequired || evidence.StrongerPolicySatisfied)
}

func deliveryCheckKey(identity DeliveryCheckIdentity) string {
	return fmt.Sprintf("%s#%d", identity.Name, identity.IntegrationID)
}

func normalizedCheckIdentities(checks []DeliveryCheckIdentity) []DeliveryCheckIdentity {
	result := slices.Clone(checks)
	slices.SortFunc(result, func(left, right DeliveryCheckIdentity) int {
		if left.Name < right.Name {
			return -1
		}
		if left.Name > right.Name {
			return 1
		}
		return int(left.IntegrationID - right.IntegrationID)
	})
	return result
}

func validRequiredCheckIdentities(checks []DeliveryCheckIdentity) bool {
	if len(checks) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, check := range checks {
		key := deliveryCheckKey(check)
		if check.Name == "" || check.IntegrationID < 0 || seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

func validateDeliveryMandate(mandate WorkExecutionMandate, plan DeliveryPlan, now time.Time) error {
	intent := plan.Intent
	if validateWorkExecutionMandateInput(mandate) != nil || now.Before(mandate.ApprovedAt) || !now.Before(mandate.ExpiresAt) {
		return errors.New("delivery execution mandate is required or expired")
	}
	if !equalWorkTarget(mandate.Target, intent.Target) || mandate.OperationID != intent.OperationID || mandate.SelectedManagedID != intent.ManagedID || !slices.Contains(mandate.Actors, plan.Capability.Actor) || !slices.Contains(mandate.CredentialModes, plan.Capability.Mode) || !slices.Contains(mandate.SourceRevisions, intent.SourceRevision) || !slices.Contains(mandate.OperatingProfileRevisions, intent.OperatingProfileRevision) || !slices.Contains(mandate.ManagedIDs, intent.ManagedID) || !slices.Contains(mandate.ResourceDigests, DeliveryResourceDigest(intent)) {
		return errors.New("delivery plan is outside execution mandate identity")
	}
	if !workAuthorityMatches(mandate, plan.Capability.Actor, plan.Capability.Mode, plan.Capability.Account, plan.Capability.InstallationID, plan.Capability.RepositoryID, plan.Capability.Permissions) {
		return errors.New("delivery capability does not match its execution mandate authority")
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
	if intent.SchemaVersion != 1 || intent.OperationID == "" || intent.SourceRevision == "" || intent.OperatingProfileRevision == "" || intent.ManagedID == "" || intent.Title == "" || intent.Target.RepositoryID == "" || intent.BaseBranch == "" || intent.HeadBranch == "" || intent.MergeMethod != "squash" || !claimValid || !validDeliveryReviewDeclaration(intent.Review, intent) || !validRequiredCheckIdentities(intent.RequiredChecks) {
		problems = append(problems, "delivery intent is invalid")
	}
	if capability.SchemaVersion != 1 || !capability.Online || !capability.Fresh || capability.Actor == "" || capability.Mode == "" || capability.RepositoryID != intent.Target.RepositoryID || capability.ExpiresAt.IsZero() || !now.Before(capability.ExpiresAt) {
		problems = append(problems, "delivery capability is unavailable or stale")
	}
	requiredPermission := "pull-requests:write"
	if observation.Branch.Name == "" {
		requiredPermission = "contents:write"
	}
	if !slices.Contains(capability.Permissions, requiredPermission) {
		problems = append(problems, "delivery capability lacks required permission: "+requiredPermission)
	}
	pr := observation.PullRequest
	validIssueState := observation.Issue.State == "open" || observation.Issue.State == "closed" && pr.Merged
	if observation.SchemaVersion != 1 || observation.Revision == "" || observation.Issue.ManagedID != intent.ManagedID || observation.Issue.Number <= 0 || !validIssueState || observation.Rules.BaseRevision == "" {
		problems = append(problems, "delivery linkage is incomplete or ambiguous")
	}
	if observation.Branch.Name == "" {
		if pr.Number != 0 {
			problems = append(problems, "delivery pull request exists without the issue branch")
		}
		return problems
	}
	if observation.Branch.Name != intent.HeadBranch || observation.Branch.Revision == "" {
		problems = append(problems, "delivery branch identity does not match governed intent")
	}
	if pr.Number == 0 {
		return problems
	}
	if pr.ClosesIssueNumber != observation.Issue.Number {
		problems = append(problems, "delivery pull request does not reciprocally close the exact managed issue")
	}
	validPullState := pr.State == "open" && !pr.Merged || pr.State == "closed"
	if !validPullState || pr.Base != intent.BaseBranch || pr.Head != intent.HeadBranch || pr.HeadRevision == "" || pr.HeadRevision != observation.Branch.Revision {
		problems = append(problems, "delivery pull request identity does not match the exact branch")
	}
	if !pr.Merged && !observation.Branch.Present {
		problems = append(problems, "open delivery pull request head branch is missing")
	}
	problems = append(problems, observation.Rules.Problems...)
	expectedChecks := normalizedCheckIdentities(intent.RequiredChecks)
	observedChecks := normalizedCheckIdentities(observation.Rules.RequiredChecks)
	if !slices.Contains(observation.Rules.MergeMethods, intent.MergeMethod) || !slices.Equal(observedChecks, expectedChecks) {
		problems = append(problems, "effective rules do not match governed delivery intent")
	}
	if _, ambiguous := effectiveDeliveryChecks(observation.Checks, observation.PullRequest.HeadRevision, intent.RequiredChecks); ambiguous {
		problems = append(problems, "effective delivery check evidence is ambiguous")
	}
	if _, ambiguous := effectiveDeliveryReviews(observation.Reviews, observation.PullRequest.HeadRevision); ambiguous {
		problems = append(problems, "effective delivery review evidence is ambiguous")
	}
	return problems
}

func deliveryDisposition(intent DeliveryIntent, observation DeliveryObservation) DeliveryDisposition {
	pr := observation.PullRequest
	effectiveChecks, _ := effectiveDeliveryChecks(observation.Checks, pr.HeadRevision, intent.RequiredChecks)
	effectiveReviews, _ := effectiveDeliveryReviews(observation.Reviews, pr.HeadRevision)
	if observation.Branch.Name == "" {
		return DeliveryDispositionBranchAbsent
	}
	if pr.Number == 0 {
		return DeliveryDispositionPullRequestAbsent
	}
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
		if effectiveChecks[deliveryCheckKey(required)].State == "failed" {
			return DeliveryDispositionChecksFailed
		}
		if effectiveChecks[deliveryCheckKey(required)].State != "passed" {
			return DeliveryDispositionChecksPending
		}
	}
	if pr.Draft {
		return DeliveryDispositionDraft
	}
	for _, review := range effectiveReviews {
		if review.State == "changes-requested" {
			return DeliveryDispositionChangesRequested
		}
	}
	if !slices.Contains(pr.RequestedReviewers, intent.Review.Actor) && effectiveReviews[intent.Review.Actor].Actor == "" {
		return DeliveryDispositionReviewUnrequested
	}
	reviewEvidence := effectiveReviews[intent.Review.Actor]
	if reviewEvidence.State != "approved" || !reviewEvidenceMatchesDeclaration(reviewEvidence, intent.Review) {
		return DeliveryDispositionReviewPending
	}
	approvalEvidence := effectiveDeliveryApproval(observation.Approvals, pr.HeadRevision, intent.ProductApproval.Role)
	if intent.ProductApproval.Role != "" && !slices.Contains(pr.RequestedReviewers, intent.ProductApproval.Role) && approvalEvidence.Actor == "" {
		return DeliveryDispositionApprovalUnrequested
	}
	if intent.ProductApproval.Role != "" && (approvalEvidence.State != "approved" || !approvalEvidence.Capable || intent.ProductApproval.DistinctContext && !approvalEvidence.DistinctContext || intent.ProductApproval.QualifiedIndependent && !approvalEvidence.QualifiedIndependent || approvalEvidence.EvidenceID != "" && approvalEvidence.EvidenceID == reviewEvidence.EvidenceID) {
		return DeliveryDispositionApprovalPending
	}
	return DeliveryDispositionMergeReady
}

func effectiveDeliveryChecks(checks []DeliveryCheckObservation, head string, required []DeliveryCheckIdentity) (map[string]DeliveryCheckObservation, bool) {
	result := map[string]DeliveryCheckObservation{}
	ambiguous := false
	byName := map[string]map[int64]DeliveryCheckObservation{}
	for _, check := range checks {
		if check.HeadRevision != head || check.Name == "" {
			continue
		}
		if byName[check.Name] == nil {
			byName[check.Name] = map[int64]DeliveryCheckObservation{}
		}
		key := deliveryCheckKey(DeliveryCheckIdentity{Name: check.Name, IntegrationID: check.IntegrationID})
		prior, exists := result[key]
		if !exists || check.ObservedAt.After(prior.ObservedAt) || check.ObservedAt.Equal(prior.ObservedAt) && (check.ObservedAt.IsZero() || check.EvidenceID > prior.EvidenceID) {
			result[key] = check
			byName[check.Name][check.IntegrationID] = check
			continue
		}
		if check.ObservedAt.Equal(prior.ObservedAt) && !check.ObservedAt.IsZero() && check.State != prior.State {
			ambiguous = true
		}
	}
	for _, identity := range required {
		if identity.IntegrationID != 0 {
			continue
		}
		identities := byName[identity.Name]
		if len(identities) > 1 {
			ambiguous = true
			delete(result, deliveryCheckKey(identity))
		} else if len(identities) == 1 {
			for _, check := range identities {
				result[deliveryCheckKey(identity)] = check
			}
		}
	}
	return result, ambiguous
}

func effectiveDeliveryReviews(reviews []DeliveryReviewObservation, head string) (map[string]DeliveryReviewObservation, bool) {
	result := map[string]DeliveryReviewObservation{}
	ambiguous := false
	for _, review := range reviews {
		if review.HeadRevision != head || review.Actor == "" {
			continue
		}
		prior, exists := result[review.Actor]
		if !exists || review.ObservedAt.After(prior.ObservedAt) || review.ObservedAt.Equal(prior.ObservedAt) && (review.ObservedAt.IsZero() || review.EvidenceID > prior.EvidenceID) {
			result[review.Actor] = review
			continue
		}
		if review.ObservedAt.Equal(prior.ObservedAt) && !review.ObservedAt.IsZero() && review.State != prior.State {
			ambiguous = true
		}
	}
	return result, ambiguous
}

func effectiveDeliveryApproval(approvals []DeliveryApprovalObservation, head, actor string) DeliveryApprovalObservation {
	var result DeliveryApprovalObservation
	for _, approval := range approvals {
		if approval.HeadRevision != head || approval.Actor != actor {
			continue
		}
		if result.Actor == "" || approval.ObservedAt.After(result.ObservedAt) || approval.ObservedAt.Equal(result.ObservedAt) && (approval.ObservedAt.IsZero() || approval.EvidenceID > result.EvidenceID) {
			result = approval
		}
	}
	return result
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

func (adapter *InMemoryDeliveryAdapter) SetCapability(capability DeliveryCapability) {
	adapter.capability = capability
}

func (adapter *InMemoryDeliveryAdapter) ApplyDelivery(_ context.Context, effect DeliveryEffect, expected DeliveryCapability) (DeliveryEffectResult, error) {
	if !sameDeliveryCapabilityAuthority(expected, adapter.capability, expected.ObservedAt) {
		return DeliveryEffectResult{Outcome: "denied", Detail: "delivery effect capability changed", Recoverable: true}, errors.New("delivery effect capability changed")
	}
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
	case DeliveryEffectCreateBranch:
		adapter.observation.Branch = DeliveryBranchObservation{Name: effect.Branch, Revision: effect.HeadRevision, Present: true}
		adapter.observation.Revision = digestJSON(adapter.observation)
		return DeliveryEffectResult{Outcome: "applied", Detail: "created issue-named branch"}, nil
	case DeliveryEffectCreatePullRequest:
		adapter.observation.PullRequest = DeliveryPullRequestObservation{Number: 101, State: "open", Draft: true, Base: effect.BaseBranch, Head: effect.Branch, HeadRevision: effect.HeadRevision, ClosesIssueNumber: effect.IssueNumber}
		adapter.observation.Revision = digestJSON(adapter.observation)
		return DeliveryEffectResult{Outcome: "applied", Detail: "created draft delivery pull request"}, nil
	case DeliveryEffectRequestReview:
		adapter.observation.PullRequest.RequestedReviewers = append(adapter.observation.PullRequest.RequestedReviewers, effect.Reviewer)
		adapter.observation.Revision = digestJSON(adapter.observation)
		return DeliveryEffectResult{Outcome: "applied", Detail: "requested distinct reviewer"}, nil
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
		return DeliveryEffectResult{Outcome: "applied", Detail: "squash merged exact head", ResourceRevision: adapter.observation.PullRequest.MergeRevision}, nil
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
