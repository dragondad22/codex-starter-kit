package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

const workStatePath = ".starter-kit/work-manager/state.json"

// WorkCredentialExpectation identifies the credential-free actor contract expected at apply time.
type WorkCredentialExpectation struct {
	Mode  string `json:"mode"`
	Actor string `json:"actor"`
}

// WorkTarget contains immutable adapter target and Project configuration identities.
type WorkTarget struct {
	Host         string            `json:"host"`
	RepositoryID string            `json:"repository_id"`
	ProjectID    string            `json:"project_id"`
	FieldIDs     map[string]string `json:"field_ids"`
	OptionIDs    map[string]string `json:"option_ids"`
}

// WorkReviewRequirement keeps review policy distinct from implementation and checks.
type WorkReviewRequirement struct {
	Role                 string `json:"role"`
	DistinctContext      bool   `json:"distinct_context"`
	QualifiedIndependent bool   `json:"qualified_independent"`
}

// WorkDependency is one native blocker fact used to derive readiness.
type WorkDependency struct {
	ManagedID string `json:"managed_id"`
	Closed    bool   `json:"closed"`
}

// WorkRelatedTask is sibling context used to derive parent completion without managing that sibling.
type WorkRelatedTask struct {
	ManagedID string `json:"managed_id"`
	Status    string `json:"status"`
	Closed    bool   `json:"closed"`
}

// WorkParentContext supplies the current parent and other-child facts for one managed task.
type WorkParentContext struct {
	ManagedID     string            `json:"managed_id"`
	Status        string            `json:"status"`
	Closed        bool              `json:"closed"`
	OtherChildren []WorkRelatedTask `json:"other_children"`
}

// DesiredManagedTask is the Work Manager-owned desired state for one task.
type DesiredManagedTask struct {
	ManagedID             string                  `json:"managed_id"`
	IssueType             string                  `json:"issue_type"`
	Title                 string                  `json:"title"`
	ParentManagedID       string                  `json:"parent_managed_id,omitempty"`
	Blockers              []WorkDependency        `json:"blockers"`
	Readiness             string                  `json:"readiness"`
	Status                string                  `json:"status"`
	Phase                 string                  `json:"phase,omitempty"`
	ParentPhase           string                  `json:"parent_phase,omitempty"`
	PhaseAssignmentReason string                  `json:"phase_assignment_reason,omitempty"`
	PromotionRecord       string                  `json:"promotion_record,omitempty"`
	NoPromotionRequired   bool                    `json:"no_promotion_required"`
	Review                []WorkReviewRequirement `json:"review"`
	Closed                bool                    `json:"closed"`
	ParentContext         *WorkParentContext      `json:"parent_context,omitempty"`
}

// WorkDesiredIntent is credential-free, source-bound managed-task intent.
type WorkDesiredIntent struct {
	SchemaVersion            int                       `json:"schema_version"`
	OperationID              string                    `json:"operation_id"`
	SourceRevision           string                    `json:"source_revision"`
	OperatingProfileRevision string                    `json:"operating_profile_revision"`
	InputDigests             map[string]string         `json:"input_digests"`
	Credential               WorkCredentialExpectation `json:"credential_expectation"`
	Target                   WorkTarget                `json:"target"`
	Task                     DesiredManagedTask        `json:"task"`
}

// ManagedTaskRequest selects the local evidence repository and desired task intent.
type ManagedTaskRequest struct {
	Repository string            `json:"repository"`
	Intent     WorkDesiredIntent `json:"intent"`
}

// WorkCapability is the adapter-reported, expiring authority and availability snapshot.
type WorkCapability struct {
	SchemaVersion         int             `json:"schema_version"`
	Online                bool            `json:"online"`
	Fresh                 bool            `json:"fresh"`
	Mode                  string          `json:"mode"`
	Actor                 string          `json:"actor"`
	ActorKind             string          `json:"actor_kind,omitempty"`
	Account               string          `json:"account,omitempty"`
	InstallationID        string          `json:"installation_id,omitempty"`
	Host                  string          `json:"host,omitempty"`
	APIVersion            string          `json:"api_version,omitempty"`
	EvidenceMode          string          `json:"evidence_mode,omitempty"`
	Disposition           string          `json:"disposition,omitempty"`
	Problems              []string        `json:"problems,omitempty"`
	RepositoryID          string          `json:"repository_id,omitempty"`
	RepositoryOwner       string          `json:"repository_owner,omitempty"`
	ProjectID             string          `json:"project_id,omitempty"`
	ProjectOwner          string          `json:"project_owner,omitempty"`
	ProjectOwnerKind      string          `json:"project_owner_kind,omitempty"`
	Permissions           []string        `json:"permissions"`
	RequiredPermissions   []string        `json:"required_permissions,omitempty"`
	Limitations           []string        `json:"limitations,omitempty"`
	RESTRate              *WorkRateBudget `json:"rest_rate,omitempty"`
	GraphQLRate           *WorkRateBudget `json:"graphql_rate,omitempty"`
	ConfigurationRevision string          `json:"configuration_revision"`
	ObservedAt            time.Time       `json:"observed_at"`
	ExpiresAt             time.Time       `json:"expires_at"`
}

// WorkRateBudget is a credential-free snapshot of one GitHub API rate budget.
type WorkRateBudget struct {
	Resource  string    `json:"resource"`
	Limit     int       `json:"limit"`
	Used      int       `json:"used"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

// WorkObservedTask is normalized adapter state; raw transport requests are never retained.
type WorkObservedTask struct {
	ManagedID             string                  `json:"managed_id"`
	IssueNodeID           string                  `json:"issue_node_id"`
	ProjectItemID         string                  `json:"project_item_id"`
	Title                 string                  `json:"title"`
	IssueType             string                  `json:"issue_type"`
	ParentManagedID       string                  `json:"parent_managed_id,omitempty"`
	NativeParentManagedID string                  `json:"native_parent_managed_id,omitempty"`
	BlockedBy             []string                `json:"blocked_by"`
	ReadinessOption       string                  `json:"readiness_option_id"`
	StatusOption          string                  `json:"status_option_id"`
	Phase                 string                  `json:"phase,omitempty"`
	PhaseOption           string                  `json:"phase_option_id,omitempty"`
	ParentPhaseOption     string                  `json:"parent_phase_option_id,omitempty"`
	PhaseAssignmentReason string                  `json:"phase_assignment_reason,omitempty"`
	PromotionRecord       string                  `json:"promotion_record,omitempty"`
	Review                []WorkReviewRequirement `json:"review"`
	Closed                bool                    `json:"closed"`
}

// WorkObservation is a normalized, immutable-ID snapshot from a WorkAdapter.
type WorkObservation struct {
	SchemaVersion         int               `json:"schema_version"`
	Revision              string            `json:"revision"`
	ConfigurationRevision string            `json:"configuration_revision"`
	Target                WorkTarget        `json:"target"`
	Task                  *WorkObservedTask `json:"task,omitempty"`
	Disposition           string            `json:"disposition,omitempty"`
	Problems              []string          `json:"problems,omitempty"`
}

// WorkInspection binds desired policy, capability, and normalized observation.
type WorkInspection struct {
	SchemaVersion int               `json:"schema_version"`
	ID            string            `json:"inspection_id"`
	Repository    string            `json:"repository"`
	Intent        WorkDesiredIntent `json:"intent"`
	Capability    WorkCapability    `json:"capability"`
	Observation   WorkObservation   `json:"observation"`
	Disposition   string            `json:"disposition"`
	Problems      []string          `json:"problems"`
}

// WorkEffect is one semantic adapter effect derived by Work Manager policy.
type WorkEffect struct {
	ID         string             `json:"effect_id"`
	Kind       string             `json:"kind"`
	Operations []string           `json:"operations,omitempty"`
	Attempt    int                `json:"attempt"`
	ManagedID  string             `json:"managed_id"`
	Marker     string             `json:"marker"`
	Desired    DesiredManagedTask `json:"desired"`
}

// WorkPlan is immutable and bound to every source, target, actor, and observation precondition.
type WorkPlan struct {
	SchemaVersion            int                       `json:"schema_version"`
	ID                       string                    `json:"plan_id"`
	Repository               string                    `json:"repository"`
	OperationID              string                    `json:"operation_id"`
	SourceRevision           string                    `json:"source_revision"`
	OperatingProfileRevision string                    `json:"operating_profile_revision"`
	InputDigests             map[string]string         `json:"input_digests"`
	InspectionID             string                    `json:"inspection_id"`
	ObservationRevision      string                    `json:"observation_revision"`
	ConfigurationRevision    string                    `json:"configuration_revision"`
	CapabilityDigest         string                    `json:"capability_digest"`
	Target                   WorkTarget                `json:"target"`
	ExpectedCredential       WorkCredentialExpectation `json:"credential_expectation"`
	Preconditions            []string                  `json:"preconditions"`
	Impact                   []string                  `json:"impact"`
	Recovery                 []string                  `json:"recovery"`
	ExpiresAt                time.Time                 `json:"expires_at"`
	Effects                  []WorkEffect              `json:"effects"`
	NoChange                 bool                      `json:"no_change"`
	DerivedFacts             WorkDerivedFacts          `json:"derived_facts"`
}

// WorkDerivedFacts exposes policy results without adding a second managed item.
type WorkDerivedFacts struct {
	Readiness             string                  `json:"readiness"`
	Status                string                  `json:"status"`
	Phase                 string                  `json:"phase,omitempty"`
	PhaseSource           string                  `json:"phase_source,omitempty"`
	PhaseAssignmentReason string                  `json:"phase_assignment_reason,omitempty"`
	PromotionRecord       string                  `json:"promotion_record,omitempty"`
	Review                []WorkReviewRequirement `json:"review"`
	Completion            string                  `json:"completion"`
	ParentStatus          string                  `json:"parent_status,omitempty"`
	ParentClosed          bool                    `json:"parent_closed"`
}

// WorkEffectResult is the adapter's explicit result for one attempted semantic effect.
type WorkEffectResult struct {
	Outcome     string          `json:"outcome"`
	Attempt     int             `json:"attempt"`
	Detail      string          `json:"detail"`
	Recoverable bool            `json:"recoverable"`
	Retry       *WorkRetryState `json:"retry,omitempty"`
}

// WorkRetryState is bounded scheduling evidence for a rate-limited effect.
type WorkRetryState struct {
	Attempt     int       `json:"attempt"`
	MaxAttempts int       `json:"max_attempts"`
	RetryAt     time.Time `json:"retry_at"`
	ResetAt     time.Time `json:"reset_at"`
}

// WorkEffectReceipt preserves attributable effect and recovery evidence.
type WorkEffectReceipt struct {
	SchemaVersion       int             `json:"schema_version"`
	PlanID              string          `json:"plan_id"`
	OperationID         string          `json:"operation_id"`
	EffectID            string          `json:"effect_id"`
	EffectKind          string          `json:"effect_kind"`
	ManagedID           string          `json:"managed_id"`
	Actor               string          `json:"actor"`
	CredentialMode      string          `json:"credential_mode"`
	EvidenceMode        string          `json:"evidence_mode,omitempty"`
	Authority           []string        `json:"authority"`
	SourceRevision      string          `json:"source_revision"`
	ObservationRevision string          `json:"observation_revision"`
	RepositoryID        string          `json:"repository_id"`
	ProjectID           string          `json:"project_id"`
	Outcome             string          `json:"outcome"`
	Attempt             int             `json:"attempt"`
	Recoverable         bool            `json:"recoverable"`
	Retry               *WorkRetryState `json:"retry,omitempty"`
	Detail              string          `json:"detail"`
	RecordedAt          time.Time       `json:"recorded_at"`
}

// WorkApplyStatus is the aggregate result without erasing per-effect outcomes.
type WorkApplyStatus string

const (
	WorkApplyApplied  WorkApplyStatus = "applied"
	WorkApplyNoChange WorkApplyStatus = "no_change"
	WorkApplyNonPass  WorkApplyStatus = "non_pass"
)

// WorkApplyResult reports the current aggregate and every receipt created by this apply.
type WorkApplyResult struct {
	SchemaVersion int                 `json:"schema_version"`
	PlanID        string              `json:"plan_id"`
	Status        WorkApplyStatus     `json:"status"`
	Receipts      []WorkEffectReceipt `json:"receipts"`
	Recovery      []string            `json:"recovery"`
	Retry         *WorkRetryState     `json:"retry,omitempty"`
}

// WorkVerificationResult reports semantic convergence without claiming live evidence.
type WorkVerificationResult struct {
	SchemaVersion       int             `json:"schema_version"`
	VerificationID      string          `json:"verification_id"`
	OverallState        ControlState    `json:"overall_state"`
	Controls            []ControlResult `json:"controls"`
	EvidencePath        string          `json:"evidence_path"`
	VerifiedAt          time.Time       `json:"verified_at"`
	Capability          WorkCapability  `json:"capability"`
	ObservationRevision string          `json:"observation_revision"`
}

// ManagedTaskStatus is durable local state returned after interruption or restart.
type ManagedTaskStatus struct {
	SchemaVersion int                 `json:"schema_version"`
	Repository    string              `json:"repository"`
	Disposition   string              `json:"disposition"`
	PlanID        string              `json:"plan_id,omitempty"`
	Receipts      []WorkEffectReceipt `json:"receipts"`
	Problems      []string            `json:"problems"`
	Recovery      []string            `json:"recovery"`
	Retry         *WorkRetryState     `json:"retry,omitempty"`
}

// ManagedTaskLifecycleResult is the complete result of one managed-task lifecycle request.
type ManagedTaskLifecycleResult struct {
	SchemaVersion int                    `json:"schema_version"`
	Inspection    WorkInspection         `json:"inspection"`
	Plan          WorkPlan               `json:"plan"`
	Apply         WorkApplyResult        `json:"apply"`
	Verification  WorkVerificationResult `json:"verification"`
	Status        ManagedTaskStatus      `json:"status"`
}

// WorkAdapter is the transport seam. Policy and credential choice stay in the engine.
type WorkAdapter interface {
	Capability(context.Context) (WorkCapability, error)
	Observe(context.Context, WorkTarget, string) (WorkObservation, error)
	Apply(context.Context, WorkEffect) (WorkEffectResult, error)
}

type managedTaskState struct {
	SchemaVersion int                     `json:"schema_version"`
	StateDigest   string                  `json:"state_digest"`
	Request       ManagedTaskRequest      `json:"request"`
	Inspection    WorkInspection          `json:"inspection"`
	Plan          *WorkPlan               `json:"plan,omitempty"`
	Receipts      []WorkEffectReceipt     `json:"receipts"`
	Verification  *WorkVerificationResult `json:"verification,omitempty"`
	Disposition   string                  `json:"disposition"`
	Problems      []string                `json:"problems"`
	Recovery      []string                `json:"recovery"`
	Retry         *WorkRetryState         `json:"retry,omitempty"`
}

// ManageTask executes the complete credential-free lifecycle journey through the configured adapter.
func (e *Engine) ManageTask(ctx context.Context, request ManagedTaskRequest) (ManagedTaskLifecycleResult, error) {
	journey := ManagedTaskLifecycleResult{SchemaVersion: 1}
	inspection, err := e.InspectManagedTask(ctx, request)
	journey.Inspection = inspection
	if err != nil {
		return journey, err
	}
	plan, err := e.PlanManagedTask(ctx, inspection)
	journey.Plan = plan
	if err != nil {
		return journey, err
	}
	apply, err := e.ApplyManagedTask(ctx, plan.ID, plan)
	journey.Apply = apply
	if err != nil {
		return journey, err
	}
	verification, err := e.VerifyManagedTask(ctx, request.Repository)
	journey.Verification = verification
	if err != nil {
		return journey, err
	}
	status, err := e.ManagedTaskStatus(ctx, request.Repository)
	journey.Status = status
	if err != nil {
		return journey, err
	}
	return journey, nil
}

// InspectManagedTask reads adapter facts and persists normalized credential-free state.
func (e *Engine) InspectManagedTask(ctx context.Context, request ManagedTaskRequest) (WorkInspection, error) {
	if e.workAdapter == nil {
		return WorkInspection{}, errors.New("managed-task inspection requires a work adapter")
	}
	root, err := cleanRepositoryRoot(request.Repository)
	if err != nil {
		return WorkInspection{}, err
	}
	request.Repository = root
	if err := validateWorkIntent(request.Intent); err != nil {
		return WorkInspection{}, err
	}
	capability, err := e.workAdapter.Capability(ctx)
	if err != nil {
		return WorkInspection{}, fmt.Errorf("inspect work capability: %w", err)
	}
	var observation WorkObservation
	if capability.Disposition != "" && capability.Disposition != "available" {
		observation = WorkObservation{
			SchemaVersion: 1, ConfigurationRevision: capability.ConfigurationRevision,
			Target: cloneWorkTarget(request.Intent.Target), Disposition: capability.Disposition,
			Problems: slices.Clone(capability.Problems),
		}
		observation.Revision = digestJSON(struct {
			ManagedID   string
			Disposition string
			Problems    []string
		}{request.Intent.Task.ManagedID, observation.Disposition, observation.Problems})
	} else {
		observation, err = e.workAdapter.Observe(ctx, request.Intent.Target, request.Intent.Task.ManagedID)
		if err != nil {
			return WorkInspection{}, fmt.Errorf("inspect managed task observation: %w", err)
		}
	}
	now := e.clock.Now()
	problems := validateWorkHandshake(request.Intent, capability, observation, now)
	disposition := "inspected"
	if len(problems) != 0 {
		disposition = "non-pass"
		if !capability.Online {
			disposition = "queued-offline"
		} else if !capability.Fresh {
			disposition = "handshake-required"
		}
	}
	receipts := []WorkEffectReceipt{}
	var priorVerification *WorkVerificationResult
	var retry *WorkRetryState
	priorDisposition := ""
	priorRecovery := []string{}
	if prior, readErr := readManagedTaskState(root); readErr == nil && prior.Request.Intent.OperationID == request.Intent.OperationID && prior.Request.Intent.Task.ManagedID == request.Intent.Task.ManagedID {
		receipts = slices.Clone(prior.Receipts)
		priorVerification = prior.Verification
		retry = cloneWorkRetry(prior.Retry)
		priorDisposition = prior.Disposition
		priorRecovery = slices.Clone(prior.Recovery)
	}
	if priorDisposition == "ambiguous" && observation.Task == nil {
		disposition = "ambiguous"
		problems = append(problems, "ambiguous create remains unresolved by stable-marker observation")
	}
	if retry != nil {
		if !now.Before(retry.ResetAt) {
			retry = nil
		} else if retry.Attempt >= retry.MaxAttempts {
			disposition = "retry-exhausted"
			problems = append(problems, "bounded retry count is exhausted until the recorded reset")
		} else if now.Before(retry.RetryAt) {
			disposition = "retry-pending"
			problems = append(problems, "rate-limited effect is not eligible before the recorded retry time")
		}
	}
	sort.Strings(problems)
	inspection := WorkInspection{SchemaVersion: 1, Repository: root, Intent: request.Intent, Capability: capability, Observation: observation, Disposition: disposition, Problems: problems}
	inspection.ID = digestJSON(workInspectionWithoutID(inspection))
	state := managedTaskState{SchemaVersion: 1, Request: request, Inspection: inspection, Receipts: receipts, Verification: priorVerification, Disposition: disposition, Problems: problems, Recovery: priorRecovery, Retry: retry}
	if err := writeManagedTaskState(root, state); err != nil {
		return WorkInspection{}, err
	}
	return inspection, nil
}

// PlanManagedTask derives one immutable semantic delta without adapter effects.
func (e *Engine) PlanManagedTask(_ context.Context, inspection WorkInspection) (WorkPlan, error) {
	if inspection.ID == "" || inspection.ID != digestJSON(workInspectionWithoutID(inspection)) {
		return WorkPlan{}, errors.New("managed-task inspection identity is invalid")
	}
	if inspection.SchemaVersion != 1 || validateWorkIntent(inspection.Intent) != nil || len(validateWorkHandshake(inspection.Intent, inspection.Capability, inspection.Observation, e.clock.Now())) != 0 {
		return WorkPlan{}, errors.New("managed-task inspection schema or provenance is invalid")
	}
	if inspection.Disposition != "inspected" || len(inspection.Problems) != 0 {
		return WorkPlan{}, errors.New("managed-task inspection contains non-pass results")
	}
	desired := deriveManagedTask(inspection.Intent.Task)
	state, err := readManagedTaskState(inspection.Repository)
	if err != nil {
		return WorkPlan{}, err
	}
	effects := []WorkEffect{}
	newEffect := func(kind string, operations []string) WorkEffect {
		id := digestJSON(struct{ Kind, ManagedID, Source string }{kind, desired.ManagedID, inspection.Intent.SourceRevision})
		return WorkEffect{ID: id, Kind: kind, Operations: slices.Clone(operations), Attempt: nextWorkEffectAttempt(state.Receipts, id, e.clock.Now()), ManagedID: desired.ManagedID, Marker: "starter-kit-managed:" + desired.ManagedID, Desired: desired}
	}
	if inspection.Observation.Task == nil {
		effects = append(effects, newEffect("create-task", nil), newEffect("reconcile-task", remainingWorkOperations(desired, nil, inspection.Intent.Target)))
	} else if operations := remainingWorkOperations(desired, inspection.Observation.Task, inspection.Intent.Target); len(operations) != 0 {
		effects = append(effects, newEffect("reconcile-task", operations))
	}
	plan := WorkPlan{
		SchemaVersion: 1, Repository: inspection.Repository, OperationID: inspection.Intent.OperationID,
		SourceRevision: inspection.Intent.SourceRevision, InputDigests: cloneStringMap(inspection.Intent.InputDigests),
		OperatingProfileRevision: inspection.Intent.OperatingProfileRevision,
		InspectionID:             inspection.ID, ObservationRevision: inspection.Observation.Revision,
		ConfigurationRevision: inspection.Capability.ConfigurationRevision, Target: cloneWorkTarget(inspection.Intent.Target), ExpectedCredential: inspection.Intent.Credential,
		CapabilityDigest: digestJSON(inspection.Capability),
		Preconditions:    []string{"unchanged desired source", "fresh expected actor", "minimum declared permissions", "matching immutable target and configuration identities", "unexpired capability and plan"},
		Impact:           []string{"reconcile one managed task in the selected issue and Project target"},
		Recovery:         []string{"retain completed receipts", "refresh capability and observation", "create a new immutable plan for remaining semantic differences"},
		ExpiresAt:        inspection.Capability.ExpiresAt, Effects: effects, NoChange: len(effects) == 0,
		DerivedFacts: deriveManagedTaskFacts(desired),
	}
	plan.ID = digestJSON(workPlanWithoutID(plan))
	state.Plan = &plan
	state.Disposition = "planned"
	if plan.NoChange {
		state.Disposition = "converged"
	}
	if err := writeManagedTaskState(inspection.Repository, state); err != nil {
		return WorkPlan{}, err
	}
	return plan, nil
}

// ApplyManagedTask rechecks immutable preconditions and persists every effect receipt.
func (e *Engine) ApplyManagedTask(ctx context.Context, expectedPlanID string, plan WorkPlan) (WorkApplyResult, error) {
	if e.workAdapter == nil {
		return WorkApplyResult{}, errors.New("managed-task apply requires a work adapter")
	}
	if expectedPlanID == "" || expectedPlanID != plan.ID || plan.ID != digestJSON(workPlanWithoutID(plan)) {
		return WorkApplyResult{}, errors.New("managed-task plan identity is invalid")
	}
	if err := validateManagedTaskPlan(plan); err != nil {
		return WorkApplyResult{}, err
	}
	root, err := cleanRepositoryRoot(plan.Repository)
	if err != nil || root != plan.Repository {
		return WorkApplyResult{}, errors.New("managed-task plan repository is invalid")
	}
	lockPath, err := lifecycleLockPath(ctx, root)
	if err != nil {
		return WorkApplyResult{}, err
	}
	lease, leaseRecovery, leaseEvidence, err := acquireLifecycleLock(lockPath, plan.ID, e.clock.Now())
	if err != nil {
		return WorkApplyResult{}, fmt.Errorf("acquire managed-task lifecycle lease: %w", err)
	}
	defer releaseLifecycleLock(lockPath, lease)
	state, err := readManagedTaskState(plan.Repository)
	if err != nil {
		return WorkApplyResult{}, err
	}
	state.Recovery = append(state.Recovery, leaseRecovery...)
	state.Recovery = append(state.Recovery, leaseEvidence...)
	if state.Plan == nil || state.Plan.ID != plan.ID || state.Inspection.Intent.SourceRevision != plan.SourceRevision || state.Inspection.Intent.OperatingProfileRevision != plan.OperatingProfileRevision || state.Inspection.Observation.Revision != plan.ObservationRevision || digestJSON(plan.DerivedFacts) != digestJSON(deriveManagedTaskFacts(deriveManagedTask(state.Request.Intent.Task))) {
		state.Disposition = "stale"
		state.Problems = []string{"managed-task desired source, observation, or retained plan changed"}
		state.Recovery = slices.Clone(plan.Recovery)
		_ = writeManagedTaskState(plan.Repository, state)
		return WorkApplyResult{}, errors.New("managed-task plan is stale")
	}
	capability, err := e.workAdapter.Capability(ctx)
	if err != nil {
		return WorkApplyResult{}, fmt.Errorf("refresh work capability: %w", err)
	}
	observation, err := e.workAdapter.Observe(ctx, plan.Target, state.Request.Intent.Task.ManagedID)
	if err != nil {
		return WorkApplyResult{}, fmt.Errorf("refresh managed-task observation: %w", err)
	}
	problems := validateWorkHandshake(state.Request.Intent, capability, observation, e.clock.Now())
	if digestJSON(capability) != plan.CapabilityDigest || capability.ConfigurationRevision != plan.ConfigurationRevision {
		problems = append(problems, "adapter capability changed after planning")
	}
	if observation.Revision != plan.ObservationRevision {
		problems = append(problems, "adapter observation changed after planning")
	}
	if len(problems) != 0 || !e.clock.Now().Before(plan.ExpiresAt) {
		state.Disposition = "stale"
		state.Problems = append(problems, "managed-task plan expired or its preconditions changed")
		state.Recovery = slices.Clone(plan.Recovery)
		_ = writeManagedTaskState(plan.Repository, state)
		return WorkApplyResult{}, errors.New("managed-task plan preconditions are stale")
	}
	if plan.NoChange {
		state.Disposition = "converged"
		if err := writeManagedTaskState(plan.Repository, state); err != nil {
			return WorkApplyResult{}, err
		}
		return WorkApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyNoChange, Receipts: []WorkEffectReceipt{}, Recovery: []string{}}, nil
	}
	created := []WorkEffectReceipt{}
	for _, effect := range plan.Effects {
		result, applyErr := e.workAdapter.Apply(ctx, effect)
		if resultErr := validateWorkEffectResult(result); resultErr != nil {
			applyErr = resultErr
			result = WorkEffectResult{Outcome: "needs-review", Attempt: max(result.Attempt, 1), Detail: "adapter returned an invalid effect result"}
		}
		if result.Attempt != effect.Attempt {
			applyErr = errors.New("work adapter returned an attempt that does not match the immutable effect")
			result = WorkEffectResult{Outcome: "needs-review", Attempt: effect.Attempt, Detail: "adapter returned mismatched attempt evidence"}
		}
		outcome := result.Outcome
		if outcome == "" && applyErr != nil {
			outcome = "failed"
		}
		detail := result.Detail
		if detail != "" {
			detail = redactDiagnostics([]string{detail})[0]
		}
		receipt := WorkEffectReceipt{
			SchemaVersion: 1, PlanID: plan.ID, OperationID: plan.OperationID, EffectID: effect.ID, EffectKind: effect.Kind, ManagedID: effect.ManagedID,
			Actor: capability.Actor, CredentialMode: capability.Mode, EvidenceMode: capability.EvidenceMode, Authority: slices.Clone(capability.Permissions), SourceRevision: plan.SourceRevision,
			ObservationRevision: plan.ObservationRevision, RepositoryID: plan.Target.RepositoryID, ProjectID: plan.Target.ProjectID,
			Outcome: outcome, Attempt: result.Attempt, Recoverable: result.Recoverable, Retry: cloneWorkRetry(result.Retry), Detail: detail, RecordedAt: e.clock.Now(),
		}
		state.Receipts = append(state.Receipts, receipt)
		created = append(created, receipt)
		if applyErr != nil || outcome != "applied" {
			state.Disposition = outcome
			state.Retry = cloneWorkRetry(result.Retry)
			if outcome == "rate-limited" && result.Retry != nil {
				state.Disposition = "retry-pending"
				if result.Retry.Attempt >= result.Retry.MaxAttempts {
					state.Disposition = "retry-exhausted"
				}
			}
			state.Recovery = slices.Clone(plan.Recovery)
			if writeErr := writeManagedTaskState(plan.Repository, state); writeErr != nil {
				return WorkApplyResult{}, writeErr
			}
			return WorkApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyNonPass, Receipts: created, Recovery: slices.Clone(plan.Recovery)}, applyErr
		}
		if err := writeManagedTaskState(plan.Repository, state); err != nil {
			return WorkApplyResult{}, err
		}
	}
	state.Disposition = "applied"
	if err := writeManagedTaskState(plan.Repository, state); err != nil {
		return WorkApplyResult{}, err
	}
	return WorkApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: WorkApplyApplied, Receipts: created, Recovery: []string{}}, nil
}

// VerifyManagedTask re-observes the task and persists explicit convergence evidence.
func (e *Engine) VerifyManagedTask(ctx context.Context, repository string) (WorkVerificationResult, error) {
	if e.workAdapter == nil {
		return WorkVerificationResult{}, errors.New("managed-task verification requires a work adapter")
	}
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return WorkVerificationResult{}, err
	}
	state, err := readManagedTaskState(root)
	if err != nil {
		return WorkVerificationResult{}, err
	}
	capability, err := e.workAdapter.Capability(ctx)
	if err != nil {
		return WorkVerificationResult{}, fmt.Errorf("verify work capability: %w", err)
	}
	observation, err := e.workAdapter.Observe(ctx, state.Request.Intent.Target, state.Request.Intent.Task.ManagedID)
	if err != nil {
		return WorkVerificationResult{}, fmt.Errorf("verify managed task observation: %w", err)
	}
	desired := deriveManagedTask(state.Request.Intent.Task)
	control := ControlResult{ID: "WORK-MANAGER-001", State: ControlFail, Summary: "managed task differs from desired state", Rationale: "normalized adapter observation does not match Work Manager policy", Evidence: []EvidenceReference{}, Diagnostics: []string{}}
	capabilityProblems := validateWorkHandshake(state.Request.Intent, capability, observation, e.clock.Now())
	if observedTaskMatches(desired, observation.Task, state.Request.Intent.Target) && len(capabilityProblems) == 0 {
		control = ControlResult{ID: "WORK-MANAGER-001", State: ControlPass, Summary: "managed task matches desired state", Evidence: []EvidenceReference{{Kind: "machine-state", Target: workStatePath}}, Diagnostics: []string{}}
	} else if len(capabilityProblems) != 0 {
		control.Summary = "managed task capability evidence is non-pass"
		control.Rationale = "a fresh matching adapter identity, authority, target, and observation are required"
		control.Diagnostics = redactDiagnostics(capabilityProblems)
	}
	verification := WorkVerificationResult{SchemaVersion: 1, OverallState: control.State, Controls: []ControlResult{control}, EvidencePath: workStatePath, VerifiedAt: e.clock.Now(), Capability: capability, ObservationRevision: observation.Revision}
	verification.VerificationID = digestJSON(verification)
	state.Verification = &verification
	state.Inspection.Observation = observation
	priorDisposition := state.Disposition
	state.Disposition = "non-pass"
	if preserveManagedTaskNonPass(priorDisposition) {
		state.Disposition = priorDisposition
	}
	if verification.OverallState == ControlPass {
		state.Disposition = "converged"
		state.Problems = []string{}
		state.Recovery = []string{}
	}
	if err := writeManagedTaskState(root, state); err != nil {
		return WorkVerificationResult{}, err
	}
	return verification, nil
}

func preserveManagedTaskNonPass(disposition string) bool {
	return slices.Contains([]string{"queued-offline", "handshake-required", "unauthenticated", "denied", "not-found", "validation-failed", "ambiguous", "offline", "failed", "retry-pending", "retry-exhausted", "stale", "needs-review"}, disposition)
}

// ManagedTaskStatus returns durable state without contacting the adapter.
func (e *Engine) ManagedTaskStatus(_ context.Context, repository string) (ManagedTaskStatus, error) {
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return ManagedTaskStatus{}, err
	}
	state, err := readManagedTaskState(root)
	if err != nil {
		return ManagedTaskStatus{}, err
	}
	planID := ""
	if state.Plan != nil {
		planID = state.Plan.ID
	}
	return ManagedTaskStatus{SchemaVersion: 1, Repository: root, Disposition: state.Disposition, PlanID: planID, Receipts: slices.Clone(state.Receipts), Problems: slices.Clone(state.Problems), Recovery: slices.Clone(state.Recovery), Retry: cloneWorkRetry(state.Retry)}, nil
}

func validateWorkIntent(intent WorkDesiredIntent) error {
	if intent.SchemaVersion != 1 || intent.OperationID == "" || intent.SourceRevision == "" || intent.OperatingProfileRevision == "" || len(intent.InputDigests) == 0 || intent.Credential.Mode == "" || intent.Credential.Actor == "" {
		return errors.New("managed-task intent lacks required versioned provenance or actor expectation")
	}
	if intent.Task.ManagedID == "" || intent.Task.Title == "" || intent.Task.IssueType == "" || intent.Task.Readiness == "" || intent.Task.Status == "" {
		return errors.New("managed-task intent lacks required task fields")
	}
	if intent.Target.Host == "" || intent.Target.RepositoryID == "" || intent.Target.ProjectID == "" || len(intent.Target.FieldIDs) == 0 || len(intent.Target.OptionIDs) == 0 {
		return errors.New("managed-task intent lacks immutable target identities")
	}
	if !slices.Contains([]string{"task", "bug", "feature", "question", "research"}, intent.Task.IssueType) || !slices.Contains([]string{"intake", "needs-refinement", "ready", "blocked"}, intent.Task.Readiness) || !slices.Contains([]string{"backlog", "next", "in-progress", "done"}, intent.Task.Status) {
		return errors.New("managed-task intent contains an unsupported issue type or lifecycle value")
	}
	if intent.Task.Phase != "" && !validRoadmapPhase(intent.Task.Phase) || intent.Task.ParentPhase != "" && !validRoadmapPhase(intent.Task.ParentPhase) {
		return errors.New("managed-task intent contains an unsupported roadmap Phase")
	}
	if intent.Task.ParentPhase != "" && intent.Task.ParentManagedID == "" {
		return errors.New("parent-derived Phase requires a native parent identity")
	}
	if intent.Task.ParentManagedID != "" && intent.Task.Phase != "" {
		if intent.Task.Phase == intent.Task.ParentPhase {
			return errors.New("ordinary child work must derive Phase from its parent instead of duplicating the assignment")
		}
		if strings.TrimSpace(intent.Task.PhaseAssignmentReason) == "" {
			return errors.New("cross-cutting direct Phase assignment requires a reason")
		}
	} else if intent.Task.IssueType != "feature" && intent.Task.Phase != "" && strings.TrimSpace(intent.Task.PhaseAssignmentReason) == "" {
		return errors.New("cross-cutting direct Phase assignment requires a reason")
	}
	if slices.Contains([]string{"task", "bug", "feature"}, intent.Task.IssueType) {
		distinctReview := false
		for _, review := range intent.Task.Review {
			distinctReview = distinctReview || review.Role != "" && review.DistinctContext
		}
		if !distinctReview {
			return errors.New("managed implementation work requires a distinct review role")
		}
	}
	if intent.Task.Closed && intent.Task.IssueType == "question" && intent.Task.PromotionRecord == "" && !intent.Task.NoPromotionRequired {
		return errors.New("closed question requires a promotion record or explicit no-promotion resolution")
	}
	if intent.Task.Closed && intent.Task.IssueType == "research" && intent.Task.PromotionRecord == "" {
		return errors.New("closed research requires a durable promoted output")
	}
	if intent.Task.ParentContext != nil {
		if intent.Task.ParentContext.ManagedID == "" || intent.Task.ParentContext.ManagedID != intent.Task.ParentManagedID || !slices.Contains([]string{"backlog", "next", "in-progress", "done"}, intent.Task.ParentContext.Status) {
			return errors.New("managed-task intent contains invalid parent context")
		}
		for _, sibling := range intent.Task.ParentContext.OtherChildren {
			if sibling.ManagedID == "" || !slices.Contains([]string{"backlog", "next", "in-progress", "done"}, sibling.Status) {
				return errors.New("managed-task intent contains invalid sibling context")
			}
		}
	}
	derived := deriveManagedTask(intent.Task)
	if intent.Target.FieldIDs["readiness"] == "" || intent.Target.FieldIDs["status"] == "" || intent.Target.OptionIDs["readiness:"+derived.Readiness] == "" || intent.Target.OptionIDs["status:"+derived.Status] == "" {
		return errors.New("managed-task intent lacks required lifecycle field or option identities")
	}
	if derived.Phase != "" && (intent.Target.FieldIDs["phase"] == "" || intent.Target.OptionIDs["phase:"+derived.Phase] == "") {
		return errors.New("managed-task intent lacks immutable Phase field or option identity")
	}
	if intent.Task.Phase != "" || intent.Task.ParentPhase != "" {
		if intent.Target.FieldIDs["phase"] == "" {
			return errors.New("managed-task intent lacks immutable Phase field or option identity")
		}
		for _, phase := range RoadmapPhases() {
			if intent.Target.OptionIDs["phase:"+phase] == "" {
				return errors.New("managed-task intent lacks the complete immutable Phase option catalog")
			}
		}
	}
	values := []string{intent.OperationID, intent.SourceRevision, intent.OperatingProfileRevision, intent.Credential.Actor, intent.Task.ManagedID, intent.Task.Title, intent.Task.ParentManagedID, intent.Task.Phase, intent.Task.ParentPhase, intent.Task.PhaseAssignmentReason, intent.Task.PromotionRecord, intent.Target.Host, intent.Target.RepositoryID, intent.Target.ProjectID}
	for key, value := range intent.InputDigests {
		if key == "" || !validSHA256Digest(value) {
			return errors.New("managed-task intent contains an invalid input digest")
		}
		values = append(values, key, value)
	}
	for key, value := range intent.Target.FieldIDs {
		values = append(values, key, value)
	}
	for key, value := range intent.Target.OptionIDs {
		values = append(values, key, value)
	}
	for _, blocker := range intent.Task.Blockers {
		values = append(values, blocker.ManagedID)
	}
	for _, review := range intent.Task.Review {
		values = append(values, review.Role)
	}
	if intent.Task.ParentContext != nil {
		values = append(values, intent.Task.ParentContext.ManagedID, intent.Task.ParentContext.Status)
		for _, sibling := range intent.Task.ParentContext.OtherChildren {
			values = append(values, sibling.ManagedID, sibling.Status)
		}
	}
	if containsSensitiveText(strings.Join(values, "\n")) {
		return errors.New("managed-task intent contains sensitive-looking material")
	}
	return nil
}

func validateWorkHandshake(intent WorkDesiredIntent, capability WorkCapability, observation WorkObservation, now time.Time) []string {
	problems := []string{}
	if capability.SchemaVersion != 1 || observation.SchemaVersion != 1 {
		problems = append(problems, "unsupported capability or observation schema")
	}
	if capability.Mode == "" || capability.Actor == "" || capability.ConfigurationRevision == "" || capability.ObservedAt.IsZero() || capability.ExpiresAt.IsZero() {
		problems = append(problems, "adapter capability lacks valid identity, configuration, or freshness provenance")
	}
	if capability.Disposition != "" && capability.Disposition != "available" {
		problems = append(problems, capability.Problems...)
		if len(capability.Problems) == 0 {
			problems = append(problems, "adapter capability is "+capability.Disposition)
		}
	}
	if !capability.Online {
		problems = append(problems, "adapter is offline")
	}
	if !capability.Fresh || !now.Before(capability.ExpiresAt) {
		problems = append(problems, "adapter capability is stale or expired")
	}
	if capability.Mode != intent.Credential.Mode || capability.Actor != intent.Credential.Actor {
		problems = append(problems, "adapter identity does not match the expected actor")
	}
	for _, permission := range []string{"issues:write", "projects:write", "pull_requests:read"} {
		if !slices.Contains(capability.Permissions, permission) {
			problems = append(problems, "adapter lacks required permission: "+permission)
		}
	}
	seenPermissions := map[string]bool{}
	for _, permission := range capability.Permissions {
		if permission == "" || seenPermissions[permission] {
			problems = append(problems, "adapter capability contains an empty or duplicate permission")
			break
		}
		seenPermissions[permission] = true
	}
	if observation.Revision == "" || observation.ConfigurationRevision == "" || observation.Target.Host == "" || observation.Target.RepositoryID == "" || observation.Target.ProjectID == "" {
		problems = append(problems, "adapter observation lacks stable revision or target provenance")
	}
	if observation.Disposition != "" && observation.Disposition != "observed" {
		problems = append(problems, observation.Problems...)
		if len(observation.Problems) == 0 {
			problems = append(problems, "adapter observation is "+observation.Disposition)
		}
	}
	if observation.Task != nil && (observation.Task.ManagedID == "" || observation.Task.ManagedID != intent.Task.ManagedID || observation.Task.IssueNodeID == "" || observation.Task.Title == "" || observation.Task.IssueType == "") {
		problems = append(problems, "adapter observation contains an invalid task identity")
	}
	if observation.Task != nil {
		if observation.Task.NativeParentManagedID != intent.Task.ParentManagedID {
			problems = append(problems, "native parent observation does not match the governed parent identity")
		}
		if intent.Task.ParentManagedID != "" && observation.Task.ParentPhaseOption != intent.Target.OptionIDs["phase:"+intent.Task.ParentPhase] {
			problems = append(problems, "native parent Phase does not match the immutable parent Phase option")
		}
	}
	if capability.ConfigurationRevision != observation.ConfigurationRevision || !equalWorkTarget(intent.Target, observation.Target) {
		problems = append(problems, "adapter target or configuration identities changed")
	}
	sort.Strings(problems)
	return problems
}

func validateManagedTaskPlan(plan WorkPlan) error {
	if plan.SchemaVersion != 1 || plan.OperationID == "" || plan.SourceRevision == "" || plan.OperatingProfileRevision == "" || plan.InspectionID == "" || plan.ObservationRevision == "" || plan.ConfigurationRevision == "" || !validSHA256Digest(plan.CapabilityDigest) || plan.ExpiresAt.IsZero() {
		return errors.New("managed-task plan schema or provenance is invalid")
	}
	if plan.NoChange != (len(plan.Effects) == 0) {
		return errors.New("managed-task plan no-change state conflicts with effects")
	}
	for _, effect := range plan.Effects {
		if effect.Kind != "create-task" && effect.Kind != "reconcile-task" {
			return errors.New("managed-task plan contains an unsupported effect kind")
		}
		expectedID := digestJSON(struct{ Kind, ManagedID, Source string }{effect.Kind, effect.ManagedID, plan.SourceRevision})
		if effect.ID != expectedID || effect.Attempt <= 0 || effect.ManagedID == "" || effect.Marker != "starter-kit-managed:"+effect.ManagedID || effect.Desired.ManagedID != effect.ManagedID {
			return errors.New("managed-task plan contains invalid effect identity or marker provenance")
		}
		if effect.Kind == "create-task" && len(effect.Operations) != 0 || effect.Kind == "reconcile-task" && !validWorkOperations(effect.Operations) {
			return errors.New("managed-task plan contains invalid semantic operations")
		}
	}
	return nil
}

func validateWorkEffectResult(result WorkEffectResult) error {
	if !slices.Contains([]string{"applied", "ambiguous", "unauthenticated", "denied", "not-found", "validation-failed", "needs-review", "rate-limited", "failed", "offline"}, result.Outcome) || result.Attempt <= 0 {
		return errors.New("work adapter returned an invalid outcome or attempt")
	}
	if result.Outcome == "rate-limited" {
		if result.Retry == nil || result.Retry.Attempt != result.Attempt || result.Retry.MaxAttempts <= 0 || result.Retry.Attempt > result.Retry.MaxAttempts || result.Retry.RetryAt.IsZero() || result.Retry.ResetAt.IsZero() || result.Retry.ResetAt.Before(result.Retry.RetryAt) {
			return errors.New("work adapter returned invalid bounded retry evidence")
		}
	} else if result.Retry != nil {
		return errors.New("work adapter returned retry evidence for a non-rate outcome")
	}
	return nil
}

func deriveManagedTask(task DesiredManagedTask) DesiredManagedTask {
	derived := task
	derived.Blockers = slices.Clone(task.Blockers)
	derived.Review = slices.Clone(task.Review)
	if len(derived.Blockers) != 0 {
		allClosed := true
		for _, blocker := range derived.Blockers {
			allClosed = allClosed && blocker.Closed
		}
		if allClosed && derived.Readiness == "blocked" {
			derived.Readiness = "ready"
		} else if !allClosed {
			derived.Readiness = "blocked"
		}
	}
	if derived.Closed {
		derived.Status = "done"
	}
	return derived
}

func deriveManagedTaskFacts(task DesiredManagedTask) WorkDerivedFacts {
	phase, phaseSource := effectiveRoadmapPhase(task)
	facts := WorkDerivedFacts{Readiness: task.Readiness, Status: task.Status, Phase: phase, PhaseSource: phaseSource, PhaseAssignmentReason: task.PhaseAssignmentReason, PromotionRecord: task.PromotionRecord, Review: slices.Clone(task.Review), Completion: "incomplete"}
	if task.Closed {
		facts.Completion = "complete"
	}
	if task.ParentContext == nil {
		return facts
	}
	facts.ParentStatus = task.ParentContext.Status
	facts.ParentClosed = task.ParentContext.Closed
	allClosed := task.Closed
	anyStarted := task.Closed || task.Status == "in-progress" || task.Status == "done"
	for _, sibling := range task.ParentContext.OtherChildren {
		allClosed = allClosed && sibling.Closed
		anyStarted = anyStarted || sibling.Closed || sibling.Status == "in-progress" || sibling.Status == "done"
	}
	if allClosed {
		facts.ParentStatus = "done"
		facts.ParentClosed = true
	} else if anyStarted {
		facts.ParentStatus = "in-progress"
		facts.ParentClosed = false
	}
	return facts
}

func observedTaskMatches(desired DesiredManagedTask, observed *WorkObservedTask, target WorkTarget) bool {
	return len(remainingWorkOperations(desired, observed, target)) == 0
}

func remainingWorkOperations(desired DesiredManagedTask, observed *WorkObservedTask, target WorkTarget) []string {
	if observed == nil {
		operations := []string{"issue", "project", "readiness", "status"}
		if desired.Phase != "" {
			operations = append(operations, "phase")
		}
		return operations
	}
	blockedBy := make([]string, 0, len(desired.Blockers))
	for _, blocker := range desired.Blockers {
		blockedBy = append(blockedBy, blocker.ManagedID)
	}
	operations := []string{}
	if observed.ManagedID != desired.ManagedID || observed.Title != desired.Title || observed.IssueType != desired.IssueType || observed.ParentManagedID != desired.ParentManagedID || !slices.Equal(observed.BlockedBy, blockedBy) || observed.Phase != desired.Phase || observed.PhaseAssignmentReason != desired.PhaseAssignmentReason || observed.PromotionRecord != desired.PromotionRecord || !slices.Equal(observed.Review, desired.Review) || observed.Closed != desired.Closed {
		operations = append(operations, "issue")
	}
	if observed.ProjectItemID == "" {
		operations = append(operations, "project")
	}
	if observed.ReadinessOption != target.OptionIDs["readiness:"+desired.Readiness] {
		operations = append(operations, "readiness")
	}
	if observed.StatusOption != target.OptionIDs["status:"+desired.Status] {
		operations = append(operations, "status")
	}
	desiredPhaseOption := ""
	if desired.Phase != "" {
		desiredPhaseOption = target.OptionIDs["phase:"+desired.Phase]
	}
	if observed.PhaseOption != desiredPhaseOption {
		operations = append(operations, "phase")
	}
	return operations
}

func validWorkOperations(operations []string) bool {
	if len(operations) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, operation := range operations {
		if !slices.Contains([]string{"issue", "project", "readiness", "status", "phase"}, operation) || seen[operation] {
			return false
		}
		seen[operation] = true
	}
	return true
}

func validRoadmapPhase(value string) bool {
	return slices.Contains(RoadmapPhases(), value)
}

// RoadmapPhases returns the complete governed Phase option catalog in roadmap order.
func RoadmapPhases() []string {
	return []string{"Phase 0", "Phase 1", "Phase 2", "Phase 3", "Phase 4", "Phase 5", "Phase 6", "Phase 7", "Phase 8"}
}

func effectiveRoadmapPhase(task DesiredManagedTask) (string, string) {
	if task.Phase != "" {
		return task.Phase, "direct"
	}
	if task.ParentPhase != "" {
		return task.ParentPhase, "parent"
	}
	return "", ""
}

func nextWorkEffectAttempt(receipts []WorkEffectReceipt, effectID string, now time.Time) int {
	attempt := 1
	for _, receipt := range receipts {
		if receipt.EffectID == effectID && receipt.Outcome == "rate-limited" && receipt.Retry != nil && now.Before(receipt.Retry.ResetAt) && receipt.Attempt >= attempt {
			attempt = receipt.Attempt + 1
		}
	}
	return attempt
}

func workInspectionWithoutID(value WorkInspection) WorkInspection { value.ID = ""; return value }
func workPlanWithoutID(value WorkPlan) WorkPlan                   { value.ID = ""; return value }

func managedTaskStateFile(root string) string {
	return filepath.Join(root, filepath.FromSlash(workStatePath))
}

func writeManagedTaskState(root string, state managedTaskState) error {
	path := managedTaskStateFile(root)
	if err := ensureNoSymlinkParents(root, workStatePath); err != nil {
		return fmt.Errorf("validate managed-task state path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create managed-task state directory: %w", err)
	}
	if err := ensureNoSymlinkParents(root, workStatePath); err != nil {
		return fmt.Errorf("validate managed-task state directory: %w", err)
	}
	state.StateDigest = ""
	state.StateDigest = digestJSON(state)
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode managed-task state: %w", err)
	}
	if containsSensitiveText(string(content)) {
		return errors.New("managed-task state contains sensitive-looking material")
	}
	content = append(content, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".state-*.tmp")
	if err != nil {
		return fmt.Errorf("create managed-task state staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write managed-task state: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync managed-task state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("commit managed-task state: %w", err)
	}
	return nil
}

func readManagedTaskState(root string) (managedTaskState, error) {
	if err := ensureNoSymlinkComponents(root, workStatePath); err != nil {
		return managedTaskState{}, fmt.Errorf("validate managed-task state path: %w", err)
	}
	content, err := os.ReadFile(managedTaskStateFile(root))
	if err != nil {
		return managedTaskState{}, fmt.Errorf("read managed-task state: %w", err)
	}
	var state managedTaskState
	if err := json.Unmarshal(content, &state); err != nil {
		return managedTaskState{}, fmt.Errorf("parse managed-task state: %w", err)
	}
	if state.SchemaVersion != 1 {
		return managedTaskState{}, errors.New("unsupported managed-task state schema")
	}
	recordedDigest := state.StateDigest
	state.StateDigest = ""
	if recordedDigest == "" || recordedDigest != digestJSON(state) {
		return managedTaskState{}, errors.New("managed-task state integrity is invalid")
	}
	state.StateDigest = recordedDigest
	return state, nil
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneWorkTarget(target WorkTarget) WorkTarget {
	target.FieldIDs = cloneStringMap(target.FieldIDs)
	target.OptionIDs = cloneStringMap(target.OptionIDs)
	return target
}

func cloneWorkRetry(retry *WorkRetryState) *WorkRetryState {
	if retry == nil {
		return nil
	}
	copy := *retry
	return &copy
}

func equalWorkTarget(left, right WorkTarget) bool {
	return left.Host == right.Host && left.RepositoryID == right.RepositoryID && left.ProjectID == right.ProjectID && equalStringMap(left.FieldIDs, right.FieldIDs) && equalStringMap(left.OptionIDs, right.OptionIDs)
}

// InMemoryWorkAdapter is the credential-free production contract double for deterministic tests and offline development.
type InMemoryWorkAdapter struct {
	mu          sync.Mutex
	capability  WorkCapability
	observation WorkObservation
	results     []queuedWorkResult
}

type queuedWorkResult struct {
	result        WorkEffectResult
	observeEffect bool
}

// NewInMemoryWorkAdapter returns an adapter seeded with normalized capability and observation values.
func NewInMemoryWorkAdapter(capability WorkCapability, observation WorkObservation) *InMemoryWorkAdapter {
	return &InMemoryWorkAdapter{capability: capability, observation: cloneWorkObservation(observation)}
}

func (adapter *InMemoryWorkAdapter) Capability(context.Context) (WorkCapability, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	value := adapter.capability
	value.Permissions = slices.Clone(value.Permissions)
	value.RequiredPermissions = slices.Clone(value.RequiredPermissions)
	value.Limitations = slices.Clone(value.Limitations)
	value.Problems = slices.Clone(value.Problems)
	return value, nil
}

func (adapter *InMemoryWorkAdapter) Observe(_ context.Context, _ WorkTarget, _ string) (WorkObservation, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return cloneWorkObservation(adapter.observation), nil
}

func (adapter *InMemoryWorkAdapter) Apply(_ context.Context, effect WorkEffect) (WorkEffectResult, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	if len(adapter.results) != 0 {
		queued := adapter.results[0]
		adapter.results = adapter.results[1:]
		if queued.observeEffect {
			adapter.applyObservedEffect(effect)
		}
		return queued.result, nil
	}
	adapter.applyObservedEffect(effect)
	return WorkEffectResult{Outcome: "applied", Attempt: effect.Attempt, Detail: "in-memory semantic effect applied"}, nil
}

func (adapter *InMemoryWorkAdapter) applyObservedEffect(effect WorkEffect) {
	desired := deriveManagedTask(effect.Desired)
	if effect.Kind == "create-task" {
		adapter.observation.Task = &WorkObservedTask{ManagedID: desired.ManagedID, IssueNodeID: "memory:issue:" + desired.ManagedID, Title: desired.Title, IssueType: desired.IssueType}
		adapter.observation.Revision = digestJSON(adapter.observation.Task)
		return
	}
	blockedBy := make([]string, 0, len(desired.Blockers))
	for _, blocker := range desired.Blockers {
		blockedBy = append(blockedBy, blocker.ManagedID)
	}
	phaseOption := ""
	if desired.Phase != "" {
		phaseOption = adapter.observation.Target.OptionIDs["phase:"+desired.Phase]
	}
	parentPhaseOption := ""
	if desired.ParentPhase != "" {
		parentPhaseOption = adapter.observation.Target.OptionIDs["phase:"+desired.ParentPhase]
	}
	adapter.observation.Task = &WorkObservedTask{ManagedID: desired.ManagedID, IssueNodeID: "memory:issue:" + desired.ManagedID, ProjectItemID: "memory:project-item:" + desired.ManagedID, Title: desired.Title, IssueType: desired.IssueType, ParentManagedID: desired.ParentManagedID, NativeParentManagedID: desired.ParentManagedID, BlockedBy: blockedBy, ReadinessOption: adapter.observation.Target.OptionIDs["readiness:"+desired.Readiness], StatusOption: adapter.observation.Target.OptionIDs["status:"+desired.Status], Phase: desired.Phase, PhaseOption: phaseOption, ParentPhaseOption: parentPhaseOption, PhaseAssignmentReason: desired.PhaseAssignmentReason, PromotionRecord: desired.PromotionRecord, Review: slices.Clone(desired.Review), Closed: desired.Closed}
	adapter.observation.Revision = digestJSON(adapter.observation.Task)
}

// Observation returns a defensive snapshot for composing a credential-free request.
func (adapter *InMemoryWorkAdapter) Observation() WorkObservation {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return cloneWorkObservation(adapter.observation)
}

// SetObservation replaces normalized observed state for deterministic drift and recovery scenarios.
func (adapter *InMemoryWorkAdapter) SetObservation(observation WorkObservation) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	adapter.observation = cloneWorkObservation(observation)
}

// SetCapability replaces the deterministic identity, authority, freshness, and availability snapshot.
func (adapter *InMemoryWorkAdapter) SetCapability(capability WorkCapability) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	capability.Permissions = slices.Clone(capability.Permissions)
	capability.RequiredPermissions = slices.Clone(capability.RequiredPermissions)
	capability.Limitations = slices.Clone(capability.Limitations)
	capability.Problems = slices.Clone(capability.Problems)
	adapter.capability = capability
}

// QueueApplyResult injects one deterministic adapter outcome; observeEffect models a lost response after an effect.
func (adapter *InMemoryWorkAdapter) QueueApplyResult(result WorkEffectResult, observeEffect bool) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	adapter.results = append(adapter.results, queuedWorkResult{result: result, observeEffect: observeEffect})
}

func cloneWorkObservation(value WorkObservation) WorkObservation {
	value.Target = cloneWorkTarget(value.Target)
	if value.Task != nil {
		task := *value.Task
		task.BlockedBy = slices.Clone(task.BlockedBy)
		task.Review = slices.Clone(task.Review)
		value.Task = &task
	}
	return value
}
