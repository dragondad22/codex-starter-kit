package engine

import (
	"context"
	"errors"
	"os"
	"slices"
	"time"
)

type DeliveryDisposition string

const (
	DeliveryDispositionDraft            DeliveryDisposition = "draft"
	DeliveryDispositionChecksPending    DeliveryDisposition = "checks-pending"
	DeliveryDispositionReviewPending    DeliveryDisposition = "review-pending"
	DeliveryDispositionChangesRequested DeliveryDisposition = "changes-requested"
	DeliveryDispositionMergeReady       DeliveryDisposition = "merge-ready"
	DeliveryDispositionMerged           DeliveryDisposition = "merged"
	DeliveryDispositionClosedUnmerged   DeliveryDisposition = "closed-unmerged"
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
	EffectBoundary           WorkEffectBoundary    `json:"effect_boundary"`
}

type DeliveryRequest struct {
	Repository string         `json:"repository"`
	Intent     DeliveryIntent `json:"intent"`
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
	SchemaVersion int                    `json:"schema_version"`
	PlanID        string                 `json:"plan_id"`
	Status        WorkApplyStatus        `json:"status"`
	Results       []DeliveryEffectResult `json:"results"`
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
	return inspection, nil
}

func (e *Engine) PlanDelivery(_ context.Context, inspection DeliveryInspection) (DeliveryPlan, error) {
	if inspection.ID == "" || len(inspection.Problems) != 0 || inspection.Disposition == DeliveryDispositionNeedsReview {
		return DeliveryPlan{}, errors.New("delivery inspection is not plannable")
	}
	if slices.Contains([]DeliveryDisposition{DeliveryDispositionChecksPending, DeliveryDispositionReviewPending, DeliveryDispositionChangesRequested, DeliveryDispositionClosedUnmerged}, inspection.Disposition) {
		plan := DeliveryPlan{SchemaVersion: 1, Repository: inspection.Repository, Intent: inspection.Intent, Capability: inspection.Capability, InspectionID: inspection.ID, ObservationRevision: inspection.Observation.Revision, NoChange: true}
		plan.ID = digestJSON(plan)
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
	return plan, nil
}

func (e *Engine) ApplyDelivery(ctx context.Context, expectedPlanID string, plan DeliveryPlan, mandate WorkExecutionMandate) (DeliveryApplyResult, error) {
	if expectedPlanID == "" || plan.ID != expectedPlanID || plan.ID != digestJSON(deliveryPlanWithoutID(plan)) {
		return DeliveryApplyResult{}, errors.New("delivery plan identity is invalid")
	}
	if plan.NoChange && len(plan.Effects) == 0 {
		return DeliveryApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyNoChange}, nil
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
	if usage[mandate.ID]+len(plan.Effects) > mandate.MaxEffects {
		return DeliveryApplyResult{}, errors.New("delivery execution mandate effect ceiling is exhausted")
	}
	results := make([]DeliveryEffectResult, 0, len(plan.Effects))
	for _, effect := range plan.Effects {
		usage[mandate.ID]++
		if err := writeWorkMandateLedger(plan.Repository, usage); err != nil {
			return DeliveryApplyResult{}, err
		}
		result, err := e.deliveryAdapter.ApplyDelivery(ctx, effect)
		if err != nil {
			return DeliveryApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyNonPass, Results: results}, err
		}
		results = append(results, result)
	}
	status := WorkApplyApplied
	if plan.NoChange {
		status = WorkApplyNoChange
	}
	return DeliveryApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: status, Results: results}, nil
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
	if intent.SchemaVersion != 1 || intent.OperationID == "" || intent.SourceRevision == "" || intent.OperatingProfileRevision == "" || intent.ManagedID == "" || intent.Target.RepositoryID == "" || intent.BaseBranch == "" || intent.HeadBranch == "" || intent.MergeMethod != "squash" {
		problems = append(problems, "delivery intent is invalid")
	}
	if capability.SchemaVersion != 1 || !capability.Online || !capability.Fresh || capability.Actor == "" || capability.Mode == "" || capability.ExpiresAt.IsZero() || !now.Before(capability.ExpiresAt) {
		problems = append(problems, "delivery capability is unavailable or stale")
	}
	pr := observation.PullRequest
	validPullState := pr.State == "open" && !pr.Merged || pr.State == "closed"
	if observation.SchemaVersion != 1 || observation.Revision == "" || observation.Issue.ManagedID != intent.ManagedID || observation.Issue.State != "open" || pr.Number <= 0 || !validPullState || pr.Base != intent.BaseBranch || pr.Head != intent.HeadBranch || pr.HeadRevision == "" {
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
