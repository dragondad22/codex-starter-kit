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

const (
	workStatePath         = ".starter-kit/work-manager/state.json"
	workMandateLedgerPath = ".starter-kit/work-mandates.json"
)

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
	ManagedID           string            `json:"managed_id"`
	Status              string            `json:"status"`
	Closed              bool              `json:"closed"`
	CompletionSatisfied bool              `json:"completion_satisfied"`
	OtherChildren       []WorkRelatedTask `json:"other_children"`
}

// WorkDependentContext supplies one direct dependent and its complete blocker slice.
type WorkDependentContext struct {
	ManagedID     string           `json:"managed_id"`
	Readiness     string           `json:"readiness"`
	Status        string           `json:"status"`
	Closed        bool             `json:"closed"`
	ReadyEligible bool             `json:"ready_eligible"`
	Blockers      []WorkDependency `json:"blockers"`
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
	Horizon               string                  `json:"horizon,omitempty"`
	ParentHorizon         string                  `json:"parent_horizon,omitempty"`
	Phase                 string                  `json:"phase,omitempty"`
	ParentPhase           string                  `json:"parent_phase,omitempty"`
	PhaseAssignmentReason string                  `json:"phase_assignment_reason,omitempty"`
	PromotionRecord       string                  `json:"promotion_record,omitempty"`
	NoPromotionRequired   bool                    `json:"no_promotion_required"`
	Review                []WorkReviewRequirement `json:"review"`
	Closed                bool                    `json:"closed"`
	ParentContext         *WorkParentContext      `json:"parent_context,omitempty"`
	Dependents            []WorkDependentContext  `json:"dependents,omitempty"`
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
	Governance               *GovernedWorkContract     `json:"governance,omitempty"`
	EffectBoundary           WorkEffectBoundary        `json:"effect_boundary,omitempty"`
}

// WorkEffectBoundary states the data, cost, destructive, retention, and recovery facts a plan must preserve.
type WorkEffectBoundary struct {
	DataClass     string `json:"data_class"`
	CostCeiling   string `json:"cost_ceiling"`
	Destructive   string `json:"destructive"`
	Retention     string `json:"retention"`
	RecoveryOwner string `json:"recovery_owner"`
}

// ManagedTaskRequest selects the local evidence repository and desired task intent.
type ManagedTaskRequest struct {
	Repository       string                `json:"repository"`
	Intent           WorkDesiredIntent     `json:"intent"`
	ExecutionMandate *WorkExecutionMandate `json:"execution_mandate,omitempty"`
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
	ManagedID             string                   `json:"managed_id"`
	IssueNodeID           string                   `json:"issue_node_id"`
	IssueURL              string                   `json:"issue_url,omitempty"`
	ProjectItemID         string                   `json:"project_item_id"`
	Title                 string                   `json:"title"`
	IssueType             string                   `json:"issue_type"`
	ParentManagedID       string                   `json:"parent_managed_id,omitempty"`
	NativeParentManagedID string                   `json:"native_parent_managed_id,omitempty"`
	BlockedBy             []string                 `json:"blocked_by"`
	ReadinessOption       string                   `json:"readiness_option_id"`
	StatusOption          string                   `json:"status_option_id"`
	HorizonOption         string                   `json:"horizon_option_id,omitempty"`
	ParentHorizonOption   string                   `json:"parent_horizon_option_id,omitempty"`
	Phase                 string                   `json:"phase,omitempty"`
	PhaseOption           string                   `json:"phase_option_id,omitempty"`
	ParentPhaseOption     string                   `json:"parent_phase_option_id,omitempty"`
	PhaseAssignmentReason string                   `json:"phase_assignment_reason,omitempty"`
	PromotionRecord       string                   `json:"promotion_record,omitempty"`
	PromotionBacklink     bool                     `json:"promotion_backlink"`
	Review                []WorkReviewRequirement  `json:"review"`
	Closed                bool                     `json:"closed"`
	IssueContract         *ExecutableIssueContract `json:"issue_contract,omitempty"`
	IssueContractDigest   string                   `json:"issue_contract_digest,omitempty"`
	IssueContractProblems []string                 `json:"issue_contract_problems,omitempty"`
}

// WorkObservedDependent is one natively observed direct dependent and its complete blocker slice.
type WorkObservedDependent struct {
	ManagedID string           `json:"managed_id"`
	Blockers  []WorkDependency `json:"blockers"`
}

// WorkRelationshipObservation contains bounded native hierarchy and dependency facts.
// Completion satisfaction and Ready eligibility remain governed intent rather than adapter facts.
type WorkRelationshipObservation struct {
	Observed        bool                    `json:"observed"`
	ParentManagedID string                  `json:"parent_managed_id,omitempty"`
	OtherChildren   []WorkRelatedTask       `json:"other_children"`
	Blockers        []WorkDependency        `json:"blockers"`
	Dependents      []WorkObservedDependent `json:"dependents"`
}

// WorkObservation is a normalized, immutable-ID snapshot from a WorkAdapter.
type WorkObservation struct {
	SchemaVersion         int                         `json:"schema_version"`
	Revision              string                      `json:"revision"`
	ConfigurationRevision string                      `json:"configuration_revision"`
	Target                WorkTarget                  `json:"target"`
	Task                  *WorkObservedTask           `json:"task,omitempty"`
	RelatedTasks          []WorkObservedTask          `json:"related_tasks,omitempty"`
	Relationships         WorkRelationshipObservation `json:"relationships"`
	Delivery              *WorkDeliveryObservation    `json:"delivery,omitempty"`
	Disposition           string                      `json:"disposition,omitempty"`
	Problems              []string                    `json:"problems,omitempty"`
}

// WorkInspection binds desired policy, capability, and normalized observation.
type WorkInspection struct {
	SchemaVersion int                       `json:"schema_version"`
	ID            string                    `json:"inspection_id"`
	Repository    string                    `json:"repository"`
	Intent        WorkDesiredIntent         `json:"intent"`
	Capability    WorkCapability            `json:"capability"`
	Observation   WorkObservation           `json:"observation"`
	Qualification *ManagedWorkQualification `json:"qualification,omitempty"`
	Disposition   string                    `json:"disposition"`
	Problems      []string                  `json:"problems"`
}

// WorkEffect is one semantic adapter effect derived by Work Manager policy.
type WorkEffect struct {
	ID              string                   `json:"effect_id"`
	Kind            string                   `json:"kind"`
	Operations      []string                 `json:"operations,omitempty"`
	Before          WorkLifecycleState       `json:"before"`
	After           WorkLifecycleState       `json:"after"`
	Attempt         int                      `json:"attempt"`
	ManagedID       string                   `json:"managed_id"`
	Marker          string                   `json:"marker"`
	Desired         DesiredManagedTask       `json:"desired"`
	QualificationID string                   `json:"qualification_id,omitempty"`
	IssueContract   *ExecutableIssueContract `json:"issue_contract,omitempty"`
}

// WorkLifecycleState is the semantic before/after lifecycle evidence for one correction.
type WorkLifecycleState struct {
	Readiness string `json:"readiness,omitempty"`
	Status    string `json:"status,omitempty"`
	Closed    bool   `json:"closed"`
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
	QualificationID          string                    `json:"qualification_id,omitempty"`
	EffectBoundary           WorkEffectBoundary        `json:"effect_boundary,omitempty"`
}

type WorkExecutionAuthority struct {
	Actor          string   `json:"actor"`
	CredentialMode string   `json:"credential_mode"`
	Account        string   `json:"account,omitempty"`
	InstallationID string   `json:"installation_id,omitempty"`
	RepositoryID   string   `json:"repository_id"`
	Permissions    []string `json:"permissions"`
}

// WorkExecutionMandate is DEC-0022 authority for a bounded family of external Work Manager effects.
type WorkExecutionMandate struct {
	SchemaVersion             int                      `json:"schema_version"`
	ID                        string                   `json:"mandate_id"`
	ApprovedBy                string                   `json:"approved_by"`
	ApprovalID                string                   `json:"approval_id"`
	ApprovedAt                time.Time                `json:"approved_at"`
	ExpiresAt                 time.Time                `json:"expires_at"`
	Target                    WorkTarget               `json:"target"`
	OperationID               string                   `json:"operation_id"`
	SelectedManagedID         string                   `json:"selected_managed_id"`
	Actors                    []string                 `json:"actors"`
	CredentialModes           []string                 `json:"credential_modes"`
	Permissions               []string                 `json:"permissions"`
	Authorities               []WorkExecutionAuthority `json:"authorities,omitempty"`
	OperatingProfileRevisions []string                 `json:"operating_profile_revisions"`
	ContractDigests           []string                 `json:"contract_digests"`
	GovernanceDigests         []string                 `json:"governance_digests"`
	InputDigests              map[string]string        `json:"input_digests"`
	GovernedSourceDigests     map[string]string        `json:"governed_source_digests"`
	SourceRevisions           []string                 `json:"source_revisions"`
	ManagedIDs                []string                 `json:"managed_ids"`
	EffectKinds               []string                 `json:"effect_kinds"`
	Operations                []string                 `json:"operations"`
	ResourceDigests           []string                 `json:"resource_digests"`
	MaxEffects                int                      `json:"max_effects"`
	DataClass                 string                   `json:"data_class"`
	CostCeiling               string                   `json:"cost_ceiling"`
	Destructive               string                   `json:"destructive"`
	Retention                 string                   `json:"retention"`
	RecoveryOwner             string                   `json:"recovery_owner"`
}

// WorkDerivedFacts exposes policy results without adding a second managed item.
type WorkDerivedFacts struct {
	Readiness             string                   `json:"readiness"`
	Status                string                   `json:"status"`
	Horizon               string                   `json:"horizon,omitempty"`
	HorizonSource         string                   `json:"horizon_source,omitempty"`
	HorizonCapability     string                   `json:"horizon_capability"`
	Phase                 string                   `json:"phase,omitempty"`
	PhaseSource           string                   `json:"phase_source,omitempty"`
	PhaseCapability       string                   `json:"phase_capability"`
	PhaseAssignmentReason string                   `json:"phase_assignment_reason,omitempty"`
	PromotionRecord       string                   `json:"promotion_record,omitempty"`
	Review                []WorkReviewRequirement  `json:"review"`
	Completion            string                   `json:"completion"`
	ParentStatus          string                   `json:"parent_status,omitempty"`
	ParentClosed          bool                     `json:"parent_closed"`
	Freshness             WorkFreshnessDisposition `json:"freshness,omitempty"`
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
	SchemaVersion       int                `json:"schema_version"`
	PlanID              string             `json:"plan_id"`
	OperationID         string             `json:"operation_id"`
	EffectID            string             `json:"effect_id"`
	EffectKind          string             `json:"effect_kind"`
	Operations          []string           `json:"operations,omitempty"`
	Before              WorkLifecycleState `json:"before"`
	After               WorkLifecycleState `json:"after"`
	ManagedID           string             `json:"managed_id"`
	Actor               string             `json:"actor"`
	CredentialMode      string             `json:"credential_mode"`
	MandateID           string             `json:"mandate_id,omitempty"`
	EvidenceMode        string             `json:"evidence_mode,omitempty"`
	Authority           []string           `json:"authority"`
	SourceRevision      string             `json:"source_revision"`
	ObservationRevision string             `json:"observation_revision"`
	RepositoryID        string             `json:"repository_id"`
	ProjectID           string             `json:"project_id"`
	Outcome             string             `json:"outcome"`
	Attempt             int                `json:"attempt"`
	Recoverable         bool               `json:"recoverable"`
	Retry               *WorkRetryState    `json:"retry,omitempty"`
	Detail              string             `json:"detail"`
	RecordedAt          time.Time          `json:"recorded_at"`
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
	SchemaVersion   int                      `json:"schema_version"`
	Repository      string                   `json:"repository"`
	Disposition     string                   `json:"disposition"`
	PlanID          string                   `json:"plan_id,omitempty"`
	Receipts        []WorkEffectReceipt      `json:"receipts"`
	Problems        []string                 `json:"problems"`
	Recovery        []string                 `json:"recovery"`
	Retry           *WorkRetryState          `json:"retry,omitempty"`
	QualificationID string                   `json:"qualification_id,omitempty"`
	Freshness       WorkFreshnessDisposition `json:"freshness,omitempty"`
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
	Observe(context.Context, WorkTarget, string, ...string) (WorkObservation, error)
	Apply(context.Context, WorkEffect) (WorkEffectResult, error)
}

// GovernedWorkObservationRequest asks an adapter for delivery evidence bound to one exact outcome.
type GovernedWorkObservationRequest struct {
	ManagedID         string   `json:"managed_id"`
	RelatedManagedIDs []string `json:"related_managed_ids"`
	SourceRevision    string   `json:"source_revision"`
	ContractDigest    string   `json:"contract_digest"`
}

// GovernedWorkAdapter extends observation only where related delivery evidence is supported.
type GovernedWorkAdapter interface {
	ObserveGovernedWork(context.Context, WorkTarget, GovernedWorkObservationRequest) (WorkObservation, error)
}

func (e *Engine) observeManagedWork(ctx context.Context, intent WorkDesiredIntent) (WorkObservation, error) {
	if intent.SchemaVersion == 2 && intent.Governance != nil {
		adapter, ok := e.workAdapter.(GovernedWorkAdapter)
		if !ok {
			return WorkObservation{}, errors.New("schema-v2 managed work requires governed delivery observation support")
		}
		return adapter.ObserveGovernedWork(ctx, intent.Target, GovernedWorkObservationRequest{
			ManagedID: intent.Task.ManagedID, RelatedManagedIDs: relatedManagedIDs(intent.Task), SourceRevision: intent.SourceRevision,
			ContractDigest: ExecutableIssueContractDigest(intent.Governance.Issue),
		})
	}
	return e.workAdapter.Observe(ctx, intent.Target, intent.Task.ManagedID, relatedManagedIDs(intent.Task)...)
}

type managedTaskState struct {
	SchemaVersion int                     `json:"schema_version"`
	StateDigest   string                  `json:"state_digest"`
	Request       ManagedTaskRequest      `json:"request"`
	Inspection    WorkInspection          `json:"inspection"`
	Plan          *WorkPlan               `json:"plan,omitempty"`
	Receipts      []WorkEffectReceipt     `json:"receipts"`
	MandateUsage  map[string]int          `json:"mandate_usage,omitempty"`
	Verification  *WorkVerificationResult `json:"verification,omitempty"`
	Disposition   string                  `json:"disposition"`
	Problems      []string                `json:"problems"`
	Recovery      []string                `json:"recovery"`
	Retry         *WorkRetryState         `json:"retry,omitempty"`
}

type workMandateLedger struct {
	SchemaVersion int            `json:"schema_version"`
	StateDigest   string         `json:"state_digest"`
	Usage         map[string]int `json:"usage"`
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
	var apply WorkApplyResult
	if request.ExecutionMandate != nil {
		apply, err = e.ApplyManagedTaskWithMandate(ctx, plan.ID, plan, *request.ExecutionMandate)
	} else {
		apply, err = e.ApplyManagedTask(ctx, plan.ID, plan)
	}
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
	if request.ExecutionMandate != nil {
		if err := validateWorkExecutionMandateInput(*request.ExecutionMandate); err != nil {
			return WorkInspection{}, err
		}
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
		observation, err = e.observeManagedWork(ctx, request.Intent)
		if err != nil {
			return WorkInspection{}, fmt.Errorf("inspect managed task observation: %w", err)
		}
	}
	now := e.clock.Now()
	problems := validateWorkHandshake(request.Intent, capability, observation, now)
	disposition := "inspected"
	var qualification *ManagedWorkQualification
	if len(problems) == 0 && request.Intent.SchemaVersion == 2 {
		qualified, qualifyErr := qualifyGovernedWork(root, request.Intent, observation)
		if qualifyErr != nil {
			return WorkInspection{}, qualifyErr
		}
		qualification = &qualified
		if qualified.Assessment.Disposition == WorkFreshnessAlreadyDelivered && slices.Contains([]string{"task", "bug", "feature"}, request.Intent.Task.IssueType) {
			disposition = string(qualified.Assessment.Disposition)
		} else if qualified.Assessment.Disposition != WorkFreshnessFresh {
			disposition = string(qualified.Assessment.Disposition)
			problems = append(problems, qualified.Assessment.Reasons...)
		}
	}
	if len(problems) != 0 {
		if disposition == "inspected" {
			disposition = "non-pass"
		}
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
	mandateUsage := map[string]int{}
	ledgerMissing := false
	if ledger, ledgerErr := readWorkMandateLedger(root); ledgerErr == nil {
		mandateUsage = cloneIntMap(ledger.Usage)
	} else {
		ledgerMissing = errors.Is(ledgerErr, os.ErrNotExist)
		if !ledgerMissing {
			return WorkInspection{}, ledgerErr
		}
	}
	prior, readErr := readManagedTaskState(root)
	if readErr != nil {
		if !errors.Is(readErr, os.ErrNotExist) {
			return WorkInspection{}, readErr
		}
		if _, directoryErr := os.Stat(filepath.Dir(managedTaskStateFile(root))); directoryErr == nil {
			return WorkInspection{}, errors.New("managed-task state is missing from an initialized evidence directory")
		} else if !errors.Is(directoryErr, os.ErrNotExist) {
			return WorkInspection{}, fmt.Errorf("inspect managed-task evidence directory: %w", directoryErr)
		}
	} else {
		if ledgerMissing && len(prior.MandateUsage) != 0 {
			return WorkInspection{}, errors.New("Work Manager mandate ledger is missing after external usage")
		}
		if prior.Request.Intent.OperationID == request.Intent.OperationID && prior.Request.Intent.Task.ManagedID == request.Intent.Task.ManagedID {
			receipts = slices.Clone(prior.Receipts)
			priorVerification = prior.Verification
			retry = cloneWorkRetry(prior.Retry)
			priorDisposition = prior.Disposition
			priorRecovery = slices.Clone(prior.Recovery)
		}
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
	inspection := WorkInspection{SchemaVersion: 1, Repository: root, Intent: request.Intent, Capability: capability, Observation: observation, Qualification: qualification, Disposition: disposition, Problems: problems}
	inspection.ID = digestJSON(workInspectionWithoutID(inspection))
	state := managedTaskState{SchemaVersion: 1, Request: request, Inspection: inspection, Receipts: receipts, MandateUsage: mandateUsage, Verification: priorVerification, Disposition: disposition, Problems: problems, Recovery: priorRecovery, Retry: retry}
	if err := writeManagedTaskState(root, state); err != nil {
		return WorkInspection{}, err
	}
	return inspection, nil
}

func workQualificationAllowsPlan(inspection WorkInspection) bool {
	if inspection.Intent.SchemaVersion == 1 {
		return inspection.Disposition == "inspected"
	}
	if inspection.Qualification == nil {
		return false
	}
	switch inspection.Qualification.Assessment.Disposition {
	case WorkFreshnessFresh:
		return inspection.Disposition == "inspected"
	case WorkFreshnessAlreadyDelivered:
		return inspection.Disposition == string(WorkFreshnessAlreadyDelivered) && slices.Contains([]string{"task", "bug", "feature"}, inspection.Intent.Task.IssueType)
	default:
		return false
	}
}

func effectiveManagedTaskForQualification(intent WorkDesiredIntent, observation WorkObservation, qualification *ManagedWorkQualification) (DesiredManagedTask, error) {
	task := intent.Task
	if qualification != nil && qualification.Assessment.Disposition == WorkFreshnessAlreadyDelivered {
		task.Closed = true
		task.Status = "done"
	}
	return effectiveManagedTask(task, observation, intent.Target)
}

// PlanManagedTask derives one immutable semantic delta without adapter effects.
func (e *Engine) PlanManagedTask(_ context.Context, inspection WorkInspection) (WorkPlan, error) {
	if inspection.ID == "" || inspection.ID != digestJSON(workInspectionWithoutID(inspection)) {
		return WorkPlan{}, errors.New("managed-task inspection identity is invalid")
	}
	if inspection.SchemaVersion != 1 || validateWorkIntent(inspection.Intent) != nil || len(validateWorkHandshake(inspection.Intent, inspection.Capability, inspection.Observation, e.clock.Now())) != 0 {
		return WorkPlan{}, errors.New("managed-task inspection schema or provenance is invalid")
	}
	if !workQualificationAllowsPlan(inspection) || len(inspection.Problems) != 0 {
		return WorkPlan{}, errors.New("managed-task inspection contains non-pass results")
	}
	if inspection.Intent.SchemaVersion == 2 && (inspection.Qualification == nil || !validSHA256Digest(inspection.Qualification.ID) || !slices.Contains([]WorkFreshnessDisposition{WorkFreshnessFresh, WorkFreshnessAlreadyDelivered}, inspection.Qualification.Assessment.Disposition)) {
		return WorkPlan{}, errors.New("managed-task inspection lacks an executable governed-work qualification")
	}
	desired, err := effectiveManagedTaskForQualification(inspection.Intent, inspection.Observation, inspection.Qualification)
	if err != nil {
		return WorkPlan{}, err
	}
	state, err := readManagedTaskState(inspection.Repository)
	if err != nil {
		return WorkPlan{}, err
	}
	effects := []WorkEffect{}
	qualificationID := ""
	if inspection.Qualification != nil {
		qualificationID = inspection.Qualification.ID
	}
	newEffect := func(kind string, operations []string, effectDesired DesiredManagedTask, observed *WorkObservedTask) WorkEffect {
		id := managedWorkEffectID(kind, effectDesired, operations, inspection.Intent.SourceRevision, qualificationID)
		before := observedLifecycleState(observed, inspection.Intent.Target)
		effect := WorkEffect{ID: id, Kind: kind, Operations: slices.Clone(operations), Before: before, After: mergedLifecycleState(before, effectDesired, operations), Attempt: nextWorkEffectAttempt(state.Receipts, id, e.clock.Now()), ManagedID: effectDesired.ManagedID, Marker: "starter-kit-managed:" + effectDesired.ManagedID, Desired: effectDesired, QualificationID: qualificationID}
		if effectDesired.ManagedID == inspection.Intent.Task.ManagedID && inspection.Intent.Governance != nil {
			contract := inspection.Intent.Governance.Issue
			effect.IssueContract = &contract
		}
		return effect
	}
	if inspection.Observation.Task == nil {
		effects = append(effects, newEffect("create-task", nil, desired, nil), newEffect("reconcile-task", remainingWorkOperations(desired, nil, inspection.Intent.Target), desired, nil))
	} else if operations := remainingWorkOperations(desired, inspection.Observation.Task, inspection.Intent.Target); len(operations) != 0 {
		effects = append(effects, newEffect("reconcile-task", operations, desired, inspection.Observation.Task))
	}
	if inspection.Qualification != nil && slices.Contains(inspection.Qualification.Assessment.Repairs, "context") {
		found := false
		for index := range effects {
			if effects[index].ManagedID == desired.ManagedID && effects[index].Kind == "reconcile-task" {
				effects[index].Operations = append(effects[index].Operations, "context")
				effects[index].After = mergedLifecycleState(effects[index].Before, effects[index].Desired, effects[index].Operations)
				found = true
				break
			}
		}
		if !found {
			effects = append(effects, newEffect("reconcile-task", []string{"context"}, desired, inspection.Observation.Task))
		}
	}
	for _, related := range deriveRelatedManagedTasks(desired) {
		observed := findObservedRelatedTask(inspection.Observation.RelatedTasks, related.ManagedID)
		if operations := remainingRelatedWorkOperations(related, observed, inspection.Intent.Target); len(operations) != 0 {
			effects = append(effects, newEffect("reconcile-task", operations, related, observed))
		}
	}
	for index := range effects {
		effects[index].ID = managedWorkEffectID(effects[index].Kind, effects[index].Desired, effects[index].Operations, inspection.Intent.SourceRevision, qualificationID)
		effects[index].Attempt = nextWorkEffectAttempt(state.Receipts, effects[index].ID, e.clock.Now())
	}
	plan := WorkPlan{
		SchemaVersion: 1, Repository: inspection.Repository, OperationID: inspection.Intent.OperationID,
		SourceRevision: inspection.Intent.SourceRevision, InputDigests: cloneStringMap(inspection.Intent.InputDigests),
		OperatingProfileRevision: inspection.Intent.OperatingProfileRevision,
		InspectionID:             inspection.ID, ObservationRevision: inspection.Observation.Revision,
		ConfigurationRevision: inspection.Capability.ConfigurationRevision, Target: cloneWorkTarget(inspection.Intent.Target), ExpectedCredential: inspection.Intent.Credential,
		CapabilityDigest: digestJSON(inspection.Capability),
		Preconditions:    []string{"unchanged desired source", "fresh expected actor", "minimum declared permissions", "matching immutable target and configuration identities", "unexpired capability and plan"},
		Impact:           []string{"reconcile the selected managed task and its bounded parent/direct-dependent Project slice"},
		Recovery:         []string{"retain completed receipts", "refresh capability and observation", "create a new immutable plan for remaining semantic differences"},
		ExpiresAt:        inspection.Capability.ExpiresAt, Effects: effects, NoChange: len(effects) == 0,
		DerivedFacts: deriveManagedTaskFacts(desired, inspection.Intent.Target), QualificationID: qualificationID,
		EffectBoundary: inspection.Intent.EffectBoundary,
	}
	if inspection.Qualification != nil {
		plan.DerivedFacts.Freshness = inspection.Qualification.Assessment.Disposition
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

// ApplyManagedTask rechecks immutable preconditions and permits only effect-free or memory effects.
func (e *Engine) ApplyManagedTask(ctx context.Context, expectedPlanID string, plan WorkPlan) (WorkApplyResult, error) {
	return e.applyManagedTask(ctx, expectedPlanID, plan, nil)
}

// ApplyManagedTaskWithMandate applies external effects only when contained by DEC-0022 authority.
func (e *Engine) ApplyManagedTaskWithMandate(ctx context.Context, expectedPlanID string, plan WorkPlan, mandate WorkExecutionMandate) (WorkApplyResult, error) {
	return e.applyManagedTask(ctx, expectedPlanID, plan, &mandate)
}

func (e *Engine) applyManagedTask(ctx context.Context, expectedPlanID string, plan WorkPlan, mandate *WorkExecutionMandate) (WorkApplyResult, error) {
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
	effective, effectiveErr := effectiveManagedTaskForQualification(state.Request.Intent, state.Inspection.Observation, state.Inspection.Qualification)
	expectedFacts := deriveManagedTaskFacts(effective, state.Request.Intent.Target)
	if state.Inspection.Qualification != nil {
		expectedFacts.Freshness = state.Inspection.Qualification.Assessment.Disposition
	}
	if state.Plan == nil || state.Plan.ID != plan.ID || state.Inspection.Intent.SourceRevision != plan.SourceRevision || state.Inspection.Intent.OperatingProfileRevision != plan.OperatingProfileRevision || state.Inspection.Observation.Revision != plan.ObservationRevision || effectiveErr != nil || digestJSON(plan.DerivedFacts) != digestJSON(expectedFacts) {
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
	observation, err := e.observeManagedWork(ctx, state.Request.Intent)
	if err != nil {
		return WorkApplyResult{}, fmt.Errorf("refresh managed-task observation: %w", err)
	}
	problems := validateWorkHandshake(state.Request.Intent, capability, observation, e.clock.Now())
	if state.Request.Intent.SchemaVersion == 2 {
		qualification, qualifyErr := qualifyGovernedWork(root, state.Request.Intent, observation)
		if qualifyErr != nil || qualification.ID != plan.QualificationID || state.Inspection.Qualification == nil || qualification.Assessment.Disposition != state.Inspection.Qualification.Assessment.Disposition || !slices.Contains([]WorkFreshnessDisposition{WorkFreshnessFresh, WorkFreshnessAlreadyDelivered}, qualification.Assessment.Disposition) {
			problems = append(problems, "governed-work qualification changed after planning")
		}
	}
	if digestJSON(capability) != plan.CapabilityDigest || capability.ConfigurationRevision != plan.ConfigurationRevision {
		problems = append(problems, "adapter capability changed after planning")
	}
	if observation.Revision != plan.ObservationRevision {
		problems = append(problems, "adapter observation changed after planning")
	}
	mandateID := ""
	if !plan.NoChange && capability.Mode != "memory" {
		if mandate == nil {
			problems = append(problems, "external Work Manager effects require a DEC-0022 execution mandate")
		} else if mandateErr := validateWorkExecutionMandate(*mandate, plan, state, capability, e.clock.Now()); mandateErr != nil {
			problems = append(problems, mandateErr.Error())
		} else {
			mandateID = mandate.ID
		}
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
		if mandateID != "" {
			if state.MandateUsage == nil {
				state.MandateUsage = map[string]int{}
			}
			state.MandateUsage[mandateID]++
			if err := writeWorkMandateLedger(plan.Repository, state.MandateUsage); err != nil {
				return WorkApplyResult{}, err
			}
		}
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
		receiptAfter := effect.Before
		if outcome == "applied" {
			receiptAfter = effect.After
		}
		receipt := WorkEffectReceipt{
			SchemaVersion: 1, PlanID: plan.ID, OperationID: plan.OperationID, EffectID: effect.ID, EffectKind: effect.Kind, ManagedID: effect.ManagedID,
			Operations: slices.Clone(effect.Operations), Before: effect.Before, After: receiptAfter,
			Actor: capability.Actor, CredentialMode: capability.Mode, MandateID: mandateID, EvidenceMode: capability.EvidenceMode, Authority: slices.Clone(capability.Permissions), SourceRevision: plan.SourceRevision,
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
	observation, err := e.observeManagedWork(ctx, state.Request.Intent)
	if err != nil {
		return WorkVerificationResult{}, fmt.Errorf("verify managed task observation: %w", err)
	}
	desired, desiredErr := effectiveManagedTaskForQualification(state.Request.Intent, observation, state.Inspection.Qualification)
	control := ControlResult{ID: "WORK-MANAGER-001", State: ControlFail, Summary: "managed task differs from desired state", Rationale: "normalized adapter observation does not match Work Manager policy", Evidence: []EvidenceReference{}, Diagnostics: []string{}}
	capabilityProblems := validateWorkHandshake(state.Request.Intent, capability, observation, e.clock.Now())
	if state.Request.Intent.SchemaVersion == 2 {
		qualification, qualifyErr := qualifyGovernedWork(root, state.Request.Intent, observation)
		if qualifyErr != nil || state.Inspection.Qualification == nil || qualification.Assessment.Disposition != state.Inspection.Qualification.Assessment.Disposition || !slices.Contains([]WorkFreshnessDisposition{WorkFreshnessFresh, WorkFreshnessAlreadyDelivered}, qualification.Assessment.Disposition) || qualification.Assessment.ContractDigest != state.Inspection.Qualification.Assessment.ContractDigest || !equalStringMap(qualification.Assessment.SourceDigests, state.Inspection.Qualification.Assessment.SourceDigests) {
			capabilityProblems = append(capabilityProblems, "governed-work qualification is stale or non-pass")
		}
	}
	relatedMatch := true
	if desiredErr == nil {
		for _, related := range deriveRelatedManagedTasks(desired) {
			relatedMatch = relatedMatch && len(remainingRelatedWorkOperations(related, findObservedRelatedTask(observation.RelatedTasks, related.ManagedID), state.Request.Intent.Target)) == 0
		}
	} else {
		relatedMatch = false
	}
	if observedTaskMatches(desired, observation.Task, state.Request.Intent.Target) && relatedMatch && len(capabilityProblems) == 0 {
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
	return slices.Contains([]string{"queued-offline", "handshake-required", "unauthenticated", "denied", "not-found", "validation-failed", "ambiguous", "offline", "failed", "retry-pending", "retry-exhausted", "stale", "needs-review", "needs-refinement", "already-delivered", "blocked"}, disposition)
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
	status := ManagedTaskStatus{SchemaVersion: 1, Repository: root, Disposition: state.Disposition, PlanID: planID, Receipts: slices.Clone(state.Receipts), Problems: slices.Clone(state.Problems), Recovery: slices.Clone(state.Recovery), Retry: cloneWorkRetry(state.Retry)}
	if state.Inspection.Qualification != nil {
		status.QualificationID = state.Inspection.Qualification.ID
		status.Freshness = completedWorkFreshness(state)
	}
	return status, nil
}

func completedWorkFreshness(state managedTaskState) WorkFreshnessDisposition {
	if state.Inspection.Qualification == nil {
		return ""
	}
	freshness := state.Inspection.Qualification.Assessment.Disposition
	if freshness != WorkFreshnessFresh || state.Verification == nil || state.Verification.OverallState != ControlPass {
		return freshness
	}
	created := false
	repaired := false
	contextRefreshed := false
	for _, receipt := range state.Receipts {
		if receipt.Outcome != "applied" || state.Plan == nil || receipt.PlanID != state.Plan.ID {
			continue
		}
		created = created || receipt.EffectKind == "create-task"
		repaired = repaired || receipt.EffectKind == "reconcile-task" && mechanicalWorkOperations(receipt.Operations)
		contextRefreshed = contextRefreshed || slices.Contains(receipt.Operations, "context")
	}
	if created {
		return WorkFreshnessFresh
	}
	if contextRefreshed {
		return WorkFreshnessContainedContextRefreshed
	}
	if repaired {
		return WorkFreshnessMechanicalDriftRepaired
	}
	return WorkFreshnessFresh
}

func mechanicalWorkOperations(operations []string) bool {
	if len(operations) == 0 {
		return false
	}
	for _, operation := range operations {
		if !slices.Contains([]string{"project", "readiness", "status", "horizon", "phase"}, operation) {
			return false
		}
	}
	return true
}

// BindWorkExecutionMandate returns a canonical content-addressed DEC-0022 authority envelope.
func BindWorkExecutionMandate(value WorkExecutionMandate) WorkExecutionMandate {
	value.ID = ""
	value.Target = cloneWorkTarget(value.Target)
	value.InputDigests = cloneStringMap(value.InputDigests)
	value.GovernedSourceDigests = cloneStringMap(value.GovernedSourceDigests)
	for _, values := range []*[]string{&value.Actors, &value.CredentialModes, &value.Permissions, &value.OperatingProfileRevisions, &value.ContractDigests, &value.GovernanceDigests, &value.SourceRevisions, &value.ManagedIDs, &value.EffectKinds, &value.Operations, &value.ResourceDigests} {
		*values = slices.Clone(*values)
		slices.Sort(*values)
		*values = slices.Compact(*values)
	}
	value.Authorities = slices.Clone(value.Authorities)
	for index := range value.Authorities {
		value.Authorities[index].Permissions = slices.Clone(value.Authorities[index].Permissions)
		slices.Sort(value.Authorities[index].Permissions)
		value.Authorities[index].Permissions = slices.Compact(value.Authorities[index].Permissions)
	}
	slices.SortFunc(value.Authorities, func(left, right WorkExecutionAuthority) int {
		if compared := strings.Compare(left.Actor, right.Actor); compared != 0 {
			return compared
		}
		if compared := strings.Compare(left.CredentialMode, right.CredentialMode); compared != 0 {
			return compared
		}
		return strings.Compare(left.InstallationID, right.InstallationID)
	})
	value.ID = digestJSON(workExecutionMandateWithoutID(value))
	return value
}

// ManagedTaskResourceDigest returns the exact desired-resource identity used by execution mandates.
func ManagedTaskResourceDigest(desired DesiredManagedTask) string {
	return digestJSON(desired)
}

func workExecutionMandateWithoutID(value WorkExecutionMandate) WorkExecutionMandate {
	value.ID = ""
	return value
}

func validateWorkExecutionMandate(mandate WorkExecutionMandate, plan WorkPlan, state managedTaskState, capability WorkCapability, now time.Time) error {
	if mandate.SchemaVersion != 1 || mandate.ID == "" || mandate.ID != BindWorkExecutionMandate(mandate).ID || mandate.ApprovedBy == "" || mandate.ApprovalID == "" || mandate.ApprovedAt.IsZero() || mandate.ExpiresAt.IsZero() || now.Before(mandate.ApprovedAt) || !now.Before(mandate.ExpiresAt) || mandate.ExpiresAt.Before(mandate.ApprovedAt) {
		return errors.New("Work Manager execution mandate is invalid or expired")
	}
	priorEffects := state.MandateUsage[mandate.ID]
	if !equalWorkTarget(mandate.Target, plan.Target) || mandate.OperationID != plan.OperationID || mandate.SelectedManagedID != state.Request.Intent.Task.ManagedID || !slices.Contains(mandate.Actors, capability.Actor) || !slices.Contains(mandate.CredentialModes, capability.Mode) || !slices.Contains(mandate.OperatingProfileRevisions, plan.OperatingProfileRevision) || !slices.Contains(mandate.SourceRevisions, plan.SourceRevision) || priorEffects+len(plan.Effects) > mandate.MaxEffects || !equalStringMap(mandate.InputDigests, plan.InputDigests) {
		return errors.New("Work Manager plan is outside the approved execution mandate")
	}
	actualPermissions := slices.Clone(capability.Permissions)
	slices.Sort(actualPermissions)
	boundary := plan.EffectBoundary
	if !workAuthorityMatches(mandate, capability.Actor, capability.Mode, capability.Account, capability.InstallationID, capability.RepositoryID, actualPermissions) || boundary.DataClass == "" || boundary.CostCeiling == "" || boundary.Destructive == "" || boundary.Retention == "" || boundary.RecoveryOwner == "" || mandate.DataClass != boundary.DataClass || mandate.CostCeiling != boundary.CostCeiling || mandate.Destructive != boundary.Destructive || mandate.Retention != boundary.Retention || mandate.RecoveryOwner != boundary.RecoveryOwner {
		return errors.New("Work Manager authority or operating ceilings are outside the approved execution mandate")
	}
	if state.Request.Intent.SchemaVersion == 2 {
		if state.Request.Intent.Governance == nil || state.Inspection.Qualification == nil || !slices.Contains(mandate.ContractDigests, ExecutableIssueContractDigest(state.Request.Intent.Governance.Issue)) || !slices.Contains(mandate.GovernanceDigests, GovernedWorkContractDigest(*state.Request.Intent.Governance)) || !equalStringMap(mandate.GovernedSourceDigests, state.Inspection.Qualification.Assessment.SourceDigests) {
			return errors.New("Work Manager governed outcome is outside the approved execution mandate")
		}
	}
	for _, effect := range plan.Effects {
		if !slices.Contains(mandate.ManagedIDs, effect.ManagedID) || !slices.Contains(mandate.EffectKinds, effect.Kind) || !slices.Contains(mandate.ResourceDigests, ManagedTaskResourceDigest(effect.Desired)) {
			return errors.New("Work Manager semantic effect is outside the approved execution mandate")
		}
		for _, operation := range effect.Operations {
			if !slices.Contains(mandate.Operations, operation) {
				return errors.New("Work Manager operation is outside the approved execution mandate")
			}
		}
	}
	return nil
}

func workAuthorityMatches(mandate WorkExecutionMandate, actor, mode, account, installationID, repositoryID string, permissions []string) bool {
	actual := slices.Clone(permissions)
	slices.Sort(actual)
	if len(mandate.Authorities) == 0 {
		return slices.Equal(actual, mandate.Permissions)
	}
	return slices.ContainsFunc(mandate.Authorities, func(authority WorkExecutionAuthority) bool {
		expected := slices.Clone(authority.Permissions)
		slices.Sort(expected)
		return authority.Actor == actor && authority.CredentialMode == mode && (authority.Account == "" || authority.Account == account) && (authority.InstallationID == "" || authority.InstallationID == installationID) && authority.RepositoryID == repositoryID && slices.Equal(expected, actual)
	})
}

func validateWorkExecutionMandateInput(mandate WorkExecutionMandate) error {
	if mandate.SchemaVersion != 1 || mandate.ID == "" || mandate.ID != BindWorkExecutionMandate(mandate).ID || mandate.ApprovedBy == "" || mandate.ApprovalID == "" || mandate.OperationID == "" || mandate.SelectedManagedID == "" || mandate.ApprovedAt.IsZero() || mandate.ExpiresAt.IsZero() || mandate.ExpiresAt.Before(mandate.ApprovedAt) {
		return errors.New("Work Manager execution mandate input is invalid")
	}
	values := []string{mandate.ApprovedBy, mandate.ApprovalID, mandate.OperationID, mandate.SelectedManagedID, mandate.DataClass, mandate.CostCeiling, mandate.Destructive, mandate.Retention, mandate.RecoveryOwner, mandate.Target.Host, mandate.Target.RepositoryID, mandate.Target.ProjectID}
	for _, list := range [][]string{mandate.Actors, mandate.CredentialModes, mandate.Permissions, mandate.OperatingProfileRevisions, mandate.ContractDigests, mandate.GovernanceDigests, mandate.SourceRevisions, mandate.ManagedIDs, mandate.EffectKinds, mandate.Operations, mandate.ResourceDigests} {
		values = append(values, list...)
	}
	for key, value := range mandate.InputDigests {
		values = append(values, key, value)
	}
	for key, value := range mandate.GovernedSourceDigests {
		values = append(values, key, value)
	}
	seenAuthorities := map[string]bool{}
	for _, authority := range mandate.Authorities {
		if authority.Actor == "" || authority.CredentialMode == "" || authority.RepositoryID != mandate.Target.RepositoryID || len(authority.Permissions) == 0 || !slices.Contains(mandate.Actors, authority.Actor) || !slices.Contains(mandate.CredentialModes, authority.CredentialMode) {
			return errors.New("Work Manager execution mandate contains an invalid actor-scoped authority")
		}
		for _, permission := range authority.Permissions {
			if !slices.Contains(mandate.Permissions, permission) {
				return errors.New("Work Manager actor-scoped authority exceeds the mandate permission envelope")
			}
		}
		key := digestJSON(authority)
		if seenAuthorities[key] {
			return errors.New("Work Manager execution mandate contains a duplicate actor-scoped authority")
		}
		seenAuthorities[key] = true
		values = append(values, authority.Actor, authority.CredentialMode, authority.Account, authority.InstallationID, authority.RepositoryID)
		values = append(values, authority.Permissions...)
	}
	if containsSensitiveText(strings.Join(values, "\n")) {
		return errors.New("Work Manager execution mandate contains sensitive-looking material")
	}
	return nil
}

func validateWorkIntent(intent WorkDesiredIntent) error {
	if !slices.Contains([]int{1, 2}, intent.SchemaVersion) || intent.OperationID == "" || intent.SourceRevision == "" || intent.OperatingProfileRevision == "" || len(intent.InputDigests) == 0 || intent.Credential.Mode == "" || intent.Credential.Actor == "" {
		return errors.New("managed-task intent lacks required versioned provenance or actor expectation")
	}
	if intent.SchemaVersion == 1 && intent.Governance != nil {
		return errors.New("schema-v1 managed-task intent cannot claim governed-work qualification")
	}
	if intent.SchemaVersion == 2 {
		if err := validateGovernedWorkContract(intent.Governance); err != nil {
			return err
		}
		if err := validateWorkSubtypeContract(intent.Task, intent.Governance); err != nil {
			return err
		}
		boundary := intent.EffectBoundary
		if boundary.DataClass == "" || boundary.CostCeiling == "" || boundary.Destructive == "" || boundary.Retention == "" || boundary.RecoveryOwner == "" {
			return errors.New("schema-v2 managed work requires an explicit effect boundary")
		}
	}
	if intent.Task.ManagedID == "" || intent.Task.Title == "" || intent.Task.IssueType == "" || intent.Task.Readiness == "" || intent.Task.Status == "" {
		return errors.New("managed-task intent lacks required task fields")
	}
	if intent.Target.Host == "" || intent.Target.RepositoryID == "" || intent.Target.ProjectID == "" || len(intent.Target.FieldIDs) == 0 || len(intent.Target.OptionIDs) == 0 {
		return errors.New("managed-task intent lacks immutable target identities")
	}
	if duplicateMapValue(intent.Target.FieldIDs) || duplicateMapValue(intent.Target.OptionIDs) {
		return errors.New("managed-task intent contains duplicate immutable field or option identities")
	}
	if !slices.Contains([]string{"task", "bug", "feature", "question", "research"}, intent.Task.IssueType) || !slices.Contains([]string{"intake", "needs-refinement", "ready", "blocked"}, intent.Task.Readiness) || !slices.Contains([]string{"backlog", "next", "in-progress", "done"}, intent.Task.Status) {
		return errors.New("managed-task intent contains an unsupported issue type or lifecycle value")
	}
	if intent.Task.Phase != "" && !validRoadmapPhase(intent.Task.Phase) || intent.Task.ParentPhase != "" && !validRoadmapPhase(intent.Task.ParentPhase) {
		return errors.New("managed-task intent contains an unsupported roadmap Phase")
	}
	if intent.Task.Horizon != "" && !validRoadmapHorizon(intent.Task.Horizon) || intent.Task.ParentHorizon != "" && !validRoadmapHorizon(intent.Task.ParentHorizon) {
		return errors.New("managed-task intent contains an unsupported roadmap Horizon")
	}
	if intent.Task.Horizon != "" && intent.Task.IssueType != "feature" {
		return errors.New("ordinary child work must derive Horizon from its parent instead of receiving a direct assignment")
	}
	if intent.SchemaVersion == 2 && intent.Task.IssueType == "feature" && intent.Task.Readiness == "ready" && intent.Target.FieldIDs["horizon"] != "" && intent.Task.Horizon == "" {
		return errors.New("Ready feature work requires an explicit Horizon when the capability is configured")
	}
	if intent.Task.ParentHorizon != "" && intent.Task.ParentManagedID == "" {
		return errors.New("parent-derived Horizon requires a native parent identity")
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
	if intent.Task.NoPromotionRequired && (intent.SchemaVersion != 2 || intent.Task.IssueType != "question" || !intent.Task.Closed || intent.Task.PromotionRecord != "") {
		return errors.New("no-promotion resolution is valid only for a closed question without a promotion record")
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
		seenChildren := map[string]bool{intent.Task.ManagedID: true}
		for _, sibling := range intent.Task.ParentContext.OtherChildren {
			if sibling.ManagedID == "" || seenChildren[sibling.ManagedID] || !slices.Contains([]string{"backlog", "next", "in-progress", "done"}, sibling.Status) {
				return errors.New("managed-task intent contains invalid sibling context")
			}
			seenChildren[sibling.ManagedID] = true
		}
	}
	seenRelated := map[string]bool{intent.Task.ManagedID: true}
	if intent.Task.ParentManagedID != "" {
		seenRelated[intent.Task.ParentManagedID] = true
	}
	for _, dependent := range intent.Task.Dependents {
		if dependent.ManagedID == "" || seenRelated[dependent.ManagedID] || !slices.Contains([]string{"intake", "needs-refinement", "ready", "blocked"}, dependent.Readiness) || !slices.Contains([]string{"backlog", "next", "in-progress", "done"}, dependent.Status) {
			return errors.New("managed-task intent contains invalid direct dependent context")
		}
		seenRelated[dependent.ManagedID] = true
		selectedIsBlocker := false
		seenBlockers := map[string]bool{}
		for _, blocker := range dependent.Blockers {
			if blocker.ManagedID == "" || blocker.ManagedID == dependent.ManagedID || seenBlockers[blocker.ManagedID] {
				return errors.New("managed-task intent contains invalid dependent blocker context")
			}
			seenBlockers[blocker.ManagedID] = true
			selectedIsBlocker = selectedIsBlocker || blocker.ManagedID == intent.Task.ManagedID
		}
		if !selectedIsBlocker {
			return errors.New("direct dependent context does not name the selected task as a blocker")
		}
	}
	derived := deriveManagedTask(intent.Task)
	if intent.Target.FieldIDs["readiness"] == "" || intent.Target.FieldIDs["status"] == "" || intent.Target.OptionIDs["readiness:"+derived.Readiness] == "" || intent.Target.OptionIDs["status:"+derived.Status] == "" {
		return errors.New("managed-task intent lacks required lifecycle field or option identities")
	}
	if derived.Phase != "" && (intent.Target.FieldIDs["phase"] == "" || intent.Target.OptionIDs["phase:"+derived.Phase] == "") {
		return errors.New("managed-task intent lacks immutable Phase field or option identity")
	}
	if intent.Task.Horizon != "" || intent.Task.ParentHorizon != "" {
		if intent.Target.FieldIDs["horizon"] == "" {
			return errors.New("managed-task intent lacks immutable Horizon field or option identity")
		}
		for _, horizon := range RoadmapHorizons() {
			if intent.Target.OptionIDs["horizon:"+horizon] == "" {
				return errors.New("managed-task intent lacks the complete immutable Horizon option catalog")
			}
		}
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
	for _, related := range deriveRelatedManagedTasks(derived) {
		if related.Readiness != "" && intent.Target.OptionIDs["readiness:"+related.Readiness] == "" || related.Status != "" && intent.Target.OptionIDs["status:"+related.Status] == "" {
			return errors.New("managed-task intent lacks a related lifecycle option identity")
		}
	}
	values := []string{intent.OperationID, intent.SourceRevision, intent.OperatingProfileRevision, intent.Credential.Actor, intent.Task.ManagedID, intent.Task.Title, intent.Task.ParentManagedID, intent.Task.Horizon, intent.Task.ParentHorizon, intent.Task.Phase, intent.Task.ParentPhase, intent.Task.PhaseAssignmentReason, intent.Task.PromotionRecord, intent.Target.Host, intent.Target.RepositoryID, intent.Target.ProjectID}
	if intent.Governance != nil {
		values = append(values, intent.Governance.Issue.HumanSummary, intent.Governance.Issue.CurrentContext, intent.Governance.Issue.GoverningReferences, intent.Governance.Issue.Scope, intent.Governance.Issue.OutOfScope, intent.Governance.Issue.Acceptance, intent.Governance.Issue.Verification, intent.Governance.Issue.Dependencies)
		for _, source := range intent.Governance.Sources {
			values = append(values, source.ID, source.Path, source.Digest)
		}
	}
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
	for _, dependent := range intent.Task.Dependents {
		values = append(values, dependent.ManagedID, dependent.Readiness, dependent.Status)
		for _, blocker := range dependent.Blockers {
			values = append(values, blocker.ManagedID)
		}
	}
	if containsSensitiveText(strings.Join(values, "\n")) {
		return errors.New("managed-task intent contains sensitive-looking material")
	}
	return nil
}

func validateWorkHandshake(intent WorkDesiredIntent, capability WorkCapability, observation WorkObservation, now time.Time) []string {
	problems := []string{}
	if capability.SchemaVersion != 1 || !slices.Contains([]int{1, 2}, observation.SchemaVersion) {
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
	requiredPermissions := []string{"issues:write", "projects:write"}
	if intent.SchemaVersion == 2 {
		requiredPermissions = append(requiredPermissions, "pull_requests:read", "contents:read")
	}
	for _, permission := range requiredPermissions {
		if !workPermissionAvailable(capability.Permissions, permission) {
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
	expectedRelated := relatedManagedIDs(intent.Task)
	seenRelated := map[string]bool{}
	for _, related := range observation.RelatedTasks {
		if related.ManagedID == "" || related.IssueNodeID == "" || related.ProjectItemID == "" || seenRelated[related.ManagedID] || !slices.Contains(expectedRelated, related.ManagedID) {
			problems = append(problems, "adapter observation contains an invalid or unexpected related task identity")
			continue
		}
		seenRelated[related.ManagedID] = true
	}
	for _, managedID := range expectedRelated {
		if !seenRelated[managedID] {
			problems = append(problems, "adapter observation is missing required related task: "+managedID)
		}
	}
	if observation.Task != nil {
		if _, err := effectiveManagedTask(intent.Task, observation, intent.Target); err != nil {
			problems = append(problems, err.Error())
		}
	}
	if capability.ConfigurationRevision != observation.ConfigurationRevision || !equalWorkTarget(intent.Target, observation.Target) {
		problems = append(problems, "adapter target or configuration identities changed")
	}
	sort.Strings(problems)
	return problems
}

func workPermissionAvailable(permissions []string, required string) bool {
	aliases := []string{required, strings.ReplaceAll(required, "_", "-")}
	if required == "projects:write" {
		aliases = append(aliases, "organization-projects:write")
	}
	return slices.ContainsFunc(aliases, func(alias string) bool { return slices.Contains(permissions, alias) })
}

func duplicateMapValue(values map[string]string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value != "" && seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func validateManagedTaskPlan(plan WorkPlan) error {
	if plan.SchemaVersion != 1 || plan.OperationID == "" || plan.SourceRevision == "" || plan.OperatingProfileRevision == "" || plan.InspectionID == "" || plan.ObservationRevision == "" || plan.ConfigurationRevision == "" || !validSHA256Digest(plan.CapabilityDigest) || plan.ExpiresAt.IsZero() {
		return errors.New("managed-task plan schema or provenance is invalid")
	}
	if plan.NoChange != (len(plan.Effects) == 0) {
		return errors.New("managed-task plan no-change state conflicts with effects")
	}
	if plan.QualificationID != "" && !validSHA256Digest(plan.QualificationID) {
		return errors.New("managed-task plan contains an invalid governed-work qualification")
	}
	for _, effect := range plan.Effects {
		if effect.Kind != "create-task" && effect.Kind != "reconcile-task" {
			return errors.New("managed-task plan contains an unsupported effect kind")
		}
		expectedID := managedWorkEffectID(effect.Kind, effect.Desired, effect.Operations, plan.SourceRevision, plan.QualificationID)
		if effect.ID != expectedID || effect.QualificationID != plan.QualificationID || effect.Attempt <= 0 || effect.ManagedID == "" || effect.Marker != "starter-kit-managed:"+effect.ManagedID || effect.Desired.ManagedID != effect.ManagedID {
			return errors.New("managed-task plan contains invalid effect identity or marker provenance")
		}
		if effect.After != mergedLifecycleState(effect.Before, effect.Desired, effect.Operations) {
			return errors.New("managed-task plan contains invalid lifecycle after-state evidence")
		}
		if effect.Kind == "create-task" && len(effect.Operations) != 0 || effect.Kind == "reconcile-task" && !validWorkOperations(effect.Operations) {
			return errors.New("managed-task plan contains invalid semantic operations")
		}
	}
	return nil
}

func managedWorkEffectID(kind string, desired DesiredManagedTask, operations []string, source, qualificationID string) string {
	return digestJSON(struct {
		Kind, ManagedID, Source, QualificationID, ResourceDigest string
		Operations                                               []string
	}{kind, desired.ManagedID, source, qualificationID, ManagedTaskResourceDigest(desired), slices.Clone(operations)})
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
	derived.Dependents = slices.Clone(task.Dependents)
	for index := range derived.Dependents {
		derived.Dependents[index].Blockers = slices.Clone(task.Dependents[index].Blockers)
	}
	if task.ParentContext != nil {
		parent := *task.ParentContext
		parent.OtherChildren = slices.Clone(task.ParentContext.OtherChildren)
		derived.ParentContext = &parent
	}
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

func effectiveManagedTask(task DesiredManagedTask, observation WorkObservation, target WorkTarget) (DesiredManagedTask, error) {
	if observation.Task == nil {
		return deriveManagedTask(task), nil
	}
	if !observation.Relationships.Observed {
		return DesiredManagedTask{}, errors.New("adapter observation lacks bounded native relationship facts")
	}
	effective := task
	effective.Closed = task.Closed || observation.Task.Closed
	effective.Blockers = slices.Clone(observation.Relationships.Blockers)

	expectedParent := ""
	if task.ParentContext != nil {
		expectedParent = task.ParentContext.ManagedID
	}
	if observation.Relationships.ParentManagedID != expectedParent {
		return DesiredManagedTask{}, errors.New("native parent relationship differs from governed intent")
	}
	if task.ParentManagedID != expectedParent {
		return DesiredManagedTask{}, errors.New("governed parent metadata differs from parent reconciliation policy")
	}
	if task.ParentManagedID != "" && target.FieldIDs["phase"] != "" {
		parentObserved := findObservedRelatedTask(observation.RelatedTasks, expectedParent)
		if parentObserved == nil {
			return DesiredManagedTask{}, errors.New("native parent observation does not match the governed parent identity")
		}
		observedParentPhase := semanticWorkOption(target, "phase", parentObserved.PhaseOption)
		if parentObserved.PhaseOption != "" && observedParentPhase == "" || task.ParentPhase != "" && observedParentPhase != task.ParentPhase {
			return DesiredManagedTask{}, errors.New("native parent Phase does not match the immutable parent Phase option")
		}
		effective.ParentPhase = observedParentPhase
	}
	if task.ParentManagedID != "" && target.FieldIDs["horizon"] != "" {
		parentObserved := findObservedRelatedTask(observation.RelatedTasks, expectedParent)
		if parentObserved == nil {
			return DesiredManagedTask{}, errors.New("native parent observation does not match the governed parent identity")
		}
		observedParentHorizon := semanticWorkOption(target, "horizon", parentObserved.HorizonOption)
		if parentObserved.HorizonOption != "" && observedParentHorizon == "" || task.ParentHorizon != "" && observedParentHorizon != task.ParentHorizon {
			return DesiredManagedTask{}, errors.New("native parent Horizon does not match the immutable parent Horizon option")
		}
		effective.ParentHorizon = observedParentHorizon
	}
	if !sameManagedIDsFromDependencies(task.Blockers, observation.Relationships.Blockers) {
		return DesiredManagedTask{}, errors.New("native blocker relationships differ from governed intent")
	}

	if task.ParentContext != nil {
		parentObserved := findObservedRelatedTask(observation.RelatedTasks, expectedParent)
		if parentObserved == nil {
			return DesiredManagedTask{}, errors.New("native parent is missing its managed Project observation")
		}
		parent := *task.ParentContext
		parent.Status = observedStatus(*parentObserved, target)
		parent.Closed = parentObserved.Closed
		parent.OtherChildren = slices.Clone(observation.Relationships.OtherChildren)
		if parent.Status == "" || !sameManagedIDsFromRelated(task.ParentContext.OtherChildren, parent.OtherChildren) {
			return DesiredManagedTask{}, errors.New("native parent child slice differs from governed intent")
		}
		effective.ParentContext = &parent
	}

	if !sameManagedIDsFromDependents(task.Dependents, observation.Relationships.Dependents) {
		return DesiredManagedTask{}, errors.New("native direct-dependent relationships differ from governed intent")
	}
	effective.Dependents = make([]WorkDependentContext, 0, len(task.Dependents))
	for _, policy := range task.Dependents {
		observedTask := findObservedRelatedTask(observation.RelatedTasks, policy.ManagedID)
		observedDependent := findObservedDependent(observation.Relationships.Dependents, policy.ManagedID)
		if observedTask == nil || observedDependent == nil {
			return DesiredManagedTask{}, errors.New("native dependent is missing its managed Project or blocker observation")
		}
		if !sameManagedIDsFromDependencies(policy.Blockers, observedDependent.Blockers) {
			return DesiredManagedTask{}, errors.New("native dependent blocker relationships differ from governed intent")
		}
		dependent := policy
		dependent.Closed = policy.Closed || observedTask.Closed
		dependent.Blockers = slices.Clone(observedDependent.Blockers)
		if observedReadiness(*observedTask, target) == "" || observedStatus(*observedTask, target) == "" {
			return DesiredManagedTask{}, errors.New("native dependent lacks semantic lifecycle observations")
		}
		effective.Dependents = append(effective.Dependents, dependent)
	}

	effective = deriveManagedTask(effective)
	if effective.ParentContext != nil {
		allClosed := effective.Closed
		for _, sibling := range effective.ParentContext.OtherChildren {
			allClosed = allClosed && sibling.Closed
		}
		if allClosed && !effective.ParentContext.CompletionSatisfied {
			return DesiredManagedTask{}, errors.New("all native parent children are closed without a satisfied parent completion contract")
		}
	}
	if target.OptionIDs["readiness:"+effective.Readiness] == "" || target.OptionIDs["status:"+effective.Status] == "" {
		return DesiredManagedTask{}, errors.New("native task state requires an unavailable lifecycle option identity")
	}
	for _, related := range deriveRelatedManagedTasks(effective) {
		if related.Readiness != "" && target.OptionIDs["readiness:"+related.Readiness] == "" || related.Status != "" && target.OptionIDs["status:"+related.Status] == "" {
			return DesiredManagedTask{}, errors.New("native related state requires an unavailable lifecycle option identity")
		}
	}
	return effective, nil
}

func semanticWorkOption(target WorkTarget, field, optionID string) string {
	for key, candidate := range target.OptionIDs {
		name, value, ok := strings.Cut(key, ":")
		if ok && name == field && candidate == optionID {
			return value
		}
	}
	return ""
}

func observedReadiness(task WorkObservedTask, target WorkTarget) string {
	return observedOptionValue("readiness", task.ReadinessOption, target)
}

func observedStatus(task WorkObservedTask, target WorkTarget) string {
	return observedOptionValue("status", task.StatusOption, target)
}

func observedOptionValue(field, optionID string, target WorkTarget) string {
	for key, candidate := range target.OptionIDs {
		name, value, ok := strings.Cut(key, ":")
		if ok && name == field && candidate == optionID {
			return value
		}
	}
	return ""
}

func sameManagedIDsFromDependencies(expected, observed []WorkDependency) bool {
	left := make([]string, 0, len(expected))
	right := make([]string, 0, len(observed))
	for _, item := range expected {
		left = append(left, item.ManagedID)
	}
	for _, item := range observed {
		right = append(right, item.ManagedID)
	}
	sort.Strings(left)
	sort.Strings(right)
	return slices.Equal(left, right)
}

func sameManagedIDsFromRelated(expected, observed []WorkRelatedTask) bool {
	left := make([]string, 0, len(expected))
	right := make([]string, 0, len(observed))
	for _, item := range expected {
		left = append(left, item.ManagedID)
	}
	for _, item := range observed {
		right = append(right, item.ManagedID)
	}
	sort.Strings(left)
	sort.Strings(right)
	return slices.Equal(left, right)
}

func sameManagedIDsFromDependents(expected []WorkDependentContext, observed []WorkObservedDependent) bool {
	left := make([]string, 0, len(expected))
	right := make([]string, 0, len(observed))
	for _, item := range expected {
		left = append(left, item.ManagedID)
	}
	for _, item := range observed {
		right = append(right, item.ManagedID)
	}
	sort.Strings(left)
	sort.Strings(right)
	return slices.Equal(left, right)
}

func findObservedDependent(dependents []WorkObservedDependent, managedID string) *WorkObservedDependent {
	for index := range dependents {
		if dependents[index].ManagedID == managedID {
			return &dependents[index]
		}
	}
	return nil
}

func deriveManagedTaskFacts(task DesiredManagedTask, target WorkTarget) WorkDerivedFacts {
	phase, phaseSource := effectiveRoadmapPhase(task)
	horizon, horizonSource := effectiveRoadmapHorizon(task)
	facts := WorkDerivedFacts{Readiness: task.Readiness, Status: task.Status, Horizon: horizon, HorizonSource: horizonSource, HorizonCapability: "not-configured", Phase: phase, PhaseSource: phaseSource, PhaseCapability: "not-configured", PhaseAssignmentReason: task.PhaseAssignmentReason, PromotionRecord: task.PromotionRecord, Review: slices.Clone(task.Review), Completion: "incomplete"}
	if target.FieldIDs["horizon"] != "" {
		facts.HorizonCapability = "configured"
	}
	if target.FieldIDs["phase"] != "" {
		facts.PhaseCapability = "configured"
	}
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
	} else {
		facts.ParentClosed = false
		if facts.ParentStatus == "in-progress" || facts.ParentStatus == "done" {
			facts.ParentStatus = "backlog"
		}
	}
	return facts
}

func relatedManagedIDs(task DesiredManagedTask) []string {
	ids := make([]string, 0, 1+len(task.Dependents))
	if task.ParentContext != nil {
		ids = append(ids, task.ParentContext.ManagedID)
	}
	dependentIDs := make([]string, 0, len(task.Dependents))
	for _, dependent := range task.Dependents {
		dependentIDs = append(dependentIDs, dependent.ManagedID)
	}
	sort.Strings(dependentIDs)
	return append(ids, dependentIDs...)
}

func deriveRelatedManagedTasks(task DesiredManagedTask) []DesiredManagedTask {
	related := []DesiredManagedTask{}
	if task.ParentContext != nil {
		parent := DesiredManagedTask{
			ManagedID: task.ParentContext.ManagedID,
			Status:    task.ParentContext.Status,
			Closed:    task.ParentContext.Closed,
		}
		allClosed := task.Closed
		anyStarted := task.Closed || task.Status == "in-progress" || task.Status == "done"
		for _, sibling := range task.ParentContext.OtherChildren {
			allClosed = allClosed && sibling.Closed
			anyStarted = anyStarted || sibling.Closed || sibling.Status == "in-progress" || sibling.Status == "done"
		}
		if allClosed && task.ParentContext.CompletionSatisfied {
			parent.Status = "done"
			parent.Closed = true
		} else if anyStarted {
			parent.Status = "in-progress"
			parent.Closed = false
		} else {
			parent.Closed = false
			if parent.Status == "in-progress" || parent.Status == "done" {
				parent.Status = "backlog"
			}
		}
		related = append(related, parent)
	}
	dependents := slices.Clone(task.Dependents)
	sort.Slice(dependents, func(left, right int) bool { return dependents[left].ManagedID < dependents[right].ManagedID })
	for _, dependent := range dependents {
		desired := DesiredManagedTask{ManagedID: dependent.ManagedID, Readiness: dependent.Readiness, Status: dependent.Status, Closed: dependent.Closed}
		allClosed := len(dependent.Blockers) != 0
		for _, blocker := range dependent.Blockers {
			allClosed = allClosed && blocker.Closed
		}
		if !allClosed {
			desired.Readiness = "blocked"
		} else if dependent.ReadyEligible && desired.Readiness == "blocked" {
			desired.Readiness = "ready"
		}
		if desired.Closed {
			desired.Status = "done"
		}
		related = append(related, desired)
	}
	return related
}

func findObservedRelatedTask(tasks []WorkObservedTask, managedID string) *WorkObservedTask {
	for index := range tasks {
		if tasks[index].ManagedID == managedID {
			return &tasks[index]
		}
	}
	return nil
}

func remainingRelatedWorkOperations(desired DesiredManagedTask, observed *WorkObservedTask, target WorkTarget) []string {
	if observed == nil {
		return []string{"project"}
	}
	operations := []string{}
	if desired.Readiness != "" && observed.ReadinessOption != target.OptionIDs["readiness:"+desired.Readiness] {
		operations = append(operations, "readiness")
	}
	if desired.Status != "" && observed.StatusOption != target.OptionIDs["status:"+desired.Status] {
		operations = append(operations, "status")
	}
	if desired.Readiness == "" && observed.Closed != desired.Closed {
		operations = append(operations, "closure")
	}
	return operations
}

func observedLifecycleState(observed *WorkObservedTask, target WorkTarget) WorkLifecycleState {
	if observed == nil {
		return WorkLifecycleState{}
	}
	state := WorkLifecycleState{Closed: observed.Closed}
	for key, optionID := range target.OptionIDs {
		field, value, ok := strings.Cut(key, ":")
		if !ok {
			continue
		}
		if field == "readiness" && optionID == observed.ReadinessOption {
			state.Readiness = value
		}
		if field == "status" && optionID == observed.StatusOption {
			state.Status = value
		}
	}
	return state
}

func mergedLifecycleState(before WorkLifecycleState, desired DesiredManagedTask, operations []string) WorkLifecycleState {
	after := before
	if desired.Readiness != "" {
		after.Readiness = desired.Readiness
	}
	if desired.Status != "" {
		after.Status = desired.Status
	}
	if slices.Contains(operations, "issue") || slices.Contains(operations, "closure") {
		after.Closed = desired.Closed
	}
	return after
}

func observedTaskMatches(desired DesiredManagedTask, observed *WorkObservedTask, target WorkTarget) bool {
	return len(remainingWorkOperations(desired, observed, target)) == 0
}

func remainingWorkOperations(desired DesiredManagedTask, observed *WorkObservedTask, target WorkTarget) []string {
	if observed == nil {
		operations := []string{"issue", "project", "readiness", "status"}
		if desired.IssueType == "question" && desired.Closed && desired.PromotionRecord != "" && !desired.NoPromotionRequired {
			operations = append(operations, "promotion-link")
		}
		if desired.Horizon != "" {
			operations = append(operations, "horizon")
		}
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
	if observed.ManagedID != desired.ManagedID || observed.Title != desired.Title || observed.IssueType != desired.IssueType || observed.ParentManagedID != desired.ParentManagedID || !slices.Equal(observed.BlockedBy, blockedBy) || observed.Phase != desired.Phase || observed.PhaseAssignmentReason != desired.PhaseAssignmentReason || observed.PromotionRecord != desired.PromotionRecord || !slices.Equal(observed.Review, desired.Review) {
		operations = append(operations, "issue")
	}
	if observed.Closed != desired.Closed {
		operations = append(operations, "closure")
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
	if desired.IssueType == "question" && desired.Closed && desired.PromotionRecord != "" && !desired.NoPromotionRequired && !observed.PromotionBacklink {
		operations = append(operations, "promotion-link")
	}
	desiredHorizonOption := ""
	if desired.Horizon != "" {
		desiredHorizonOption = target.OptionIDs["horizon:"+desired.Horizon]
	}
	if target.FieldIDs["horizon"] != "" && observed.HorizonOption != desiredHorizonOption {
		operations = append(operations, "horizon")
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
		if !slices.Contains([]string{"issue", "project", "readiness", "status", "horizon", "phase", "closure", "context", "promotion-link"}, operation) || seen[operation] {
			return false
		}
		seen[operation] = true
	}
	return true
}

func validRoadmapPhase(value string) bool {
	return slices.Contains(RoadmapPhases(), value)
}

func validRoadmapHorizon(value string) bool {
	return slices.Contains(RoadmapHorizons(), value)
}

// RoadmapHorizons returns the governed rolling-intent options without execution meaning.
func RoadmapHorizons() []string {
	return []string{"now", "next", "later"}
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

func effectiveRoadmapHorizon(task DesiredManagedTask) (string, string) {
	if task.Horizon != "" {
		return task.Horizon, "direct"
	}
	if task.ParentHorizon != "" {
		return task.ParentHorizon, "parent"
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

func workMandateLedgerFile(root string) string {
	return filepath.Join(root, filepath.FromSlash(workMandateLedgerPath))
}

func readWorkMandateLedger(root string) (workMandateLedger, error) {
	if err := ensureNoSymlinkComponents(root, workMandateLedgerPath); err != nil {
		return workMandateLedger{}, fmt.Errorf("validate Work Manager mandate ledger path: %w", err)
	}
	content, err := os.ReadFile(workMandateLedgerFile(root))
	if err != nil {
		return workMandateLedger{}, fmt.Errorf("read Work Manager mandate ledger: %w", err)
	}
	var ledger workMandateLedger
	if json.Unmarshal(content, &ledger) != nil || ledger.SchemaVersion != 1 || ledger.Usage == nil {
		return workMandateLedger{}, errors.New("Work Manager mandate ledger is invalid")
	}
	recordedDigest := ledger.StateDigest
	ledger.StateDigest = ""
	if recordedDigest == "" || recordedDigest != digestJSON(ledger) {
		return workMandateLedger{}, errors.New("Work Manager mandate ledger integrity is invalid")
	}
	ledger.StateDigest = recordedDigest
	return ledger, nil
}

func writeWorkMandateLedger(root string, usage map[string]int) error {
	path := workMandateLedgerFile(root)
	if err := ensureNoSymlinkParents(root, workMandateLedgerPath); err != nil {
		return fmt.Errorf("validate Work Manager mandate ledger path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create Work Manager mandate ledger directory: %w", err)
	}
	ledger := workMandateLedger{SchemaVersion: 1, Usage: cloneIntMap(usage)}
	ledger.StateDigest = digestJSON(ledger)
	content, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Work Manager mandate ledger: %w", err)
	}
	content = append(content, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".work-mandates-*.tmp")
	if err != nil {
		return fmt.Errorf("create Work Manager mandate ledger staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write Work Manager mandate ledger: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync Work Manager mandate ledger: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("commit Work Manager mandate ledger: %w", err)
	}
	return nil
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

func cloneIntMap(input map[string]int) map[string]int {
	result := make(map[string]int, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
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

func (adapter *InMemoryWorkAdapter) Observe(_ context.Context, _ WorkTarget, _ string, _ ...string) (WorkObservation, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return cloneWorkObservation(adapter.observation), nil
}

// ObserveGovernedWork returns the deterministic delivery observation supplied by the fixture.
func (adapter *InMemoryWorkAdapter) ObserveGovernedWork(ctx context.Context, target WorkTarget, request GovernedWorkObservationRequest) (WorkObservation, error) {
	return adapter.Observe(ctx, target, request.ManagedID, request.RelatedManagedIDs...)
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
		if effect.IssueContract != nil {
			contract := *effect.IssueContract
			contract.ReadinessAssertions = slices.Clone(effect.IssueContract.ReadinessAssertions)
			adapter.observation.Task.IssueContract = &contract
			adapter.observation.Task.IssueContractDigest = ExecutableIssueContractDigest(contract)
		}
		adapter.observation.Revision = digestJSON(adapter.observation.Task)
		return
	}
	if adapter.observation.Task == nil || effect.ManagedID != adapter.observation.Task.ManagedID {
		for index := range adapter.observation.RelatedTasks {
			if adapter.observation.RelatedTasks[index].ManagedID != effect.ManagedID {
				continue
			}
			related := &adapter.observation.RelatedTasks[index]
			if slices.Contains(effect.Operations, "readiness") {
				related.ReadinessOption = adapter.observation.Target.OptionIDs["readiness:"+desired.Readiness]
			}
			if slices.Contains(effect.Operations, "status") {
				related.StatusOption = adapter.observation.Target.OptionIDs["status:"+desired.Status]
			}
			if slices.Contains(effect.Operations, "closure") {
				related.Closed = desired.Closed
			}
			adapter.observation.Revision = digestJSON(adapter.observation)
			return
		}
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
	horizonOption := ""
	if desired.Horizon != "" {
		horizonOption = adapter.observation.Target.OptionIDs["horizon:"+desired.Horizon]
	}
	parentHorizonOption := ""
	if desired.ParentHorizon != "" {
		parentHorizonOption = adapter.observation.Target.OptionIDs["horizon:"+desired.ParentHorizon]
	}
	priorContract := adapter.observation.Task.IssueContract
	priorContractDigest := adapter.observation.Task.IssueContractDigest
	priorContractProblems := slices.Clone(adapter.observation.Task.IssueContractProblems)
	priorIssueURL := adapter.observation.Task.IssueURL
	if slices.Contains(effect.Operations, "context") && effect.IssueContract != nil {
		contract := *effect.IssueContract
		contract.ReadinessAssertions = slices.Clone(effect.IssueContract.ReadinessAssertions)
		priorContract = &contract
		priorContractDigest = ExecutableIssueContractDigest(contract)
		priorContractProblems = nil
	}
	promotionBacklink := adapter.observation.Task.PromotionBacklink
	if slices.Contains(effect.Operations, "promotion-link") {
		promotionBacklink = true
	}
	adapter.observation.Task = &WorkObservedTask{ManagedID: desired.ManagedID, IssueNodeID: "memory:issue:" + desired.ManagedID, IssueURL: priorIssueURL, ProjectItemID: "memory:project-item:" + desired.ManagedID, Title: desired.Title, IssueType: desired.IssueType, ParentManagedID: desired.ParentManagedID, NativeParentManagedID: desired.ParentManagedID, BlockedBy: blockedBy, ReadinessOption: adapter.observation.Target.OptionIDs["readiness:"+desired.Readiness], StatusOption: adapter.observation.Target.OptionIDs["status:"+desired.Status], HorizonOption: horizonOption, ParentHorizonOption: parentHorizonOption, Phase: desired.Phase, PhaseOption: phaseOption, ParentPhaseOption: parentPhaseOption, PhaseAssignmentReason: desired.PhaseAssignmentReason, PromotionRecord: desired.PromotionRecord, PromotionBacklink: promotionBacklink, Review: slices.Clone(desired.Review), Closed: desired.Closed, IssueContract: priorContract, IssueContractDigest: priorContractDigest, IssueContractProblems: priorContractProblems}
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
		task.IssueContractProblems = slices.Clone(task.IssueContractProblems)
		if task.IssueContract != nil {
			contract := *task.IssueContract
			contract.ReadinessAssertions = slices.Clone(task.IssueContract.ReadinessAssertions)
			task.IssueContract = &contract
		}
		value.Task = &task
	}
	if value.Delivery != nil {
		delivery := *value.Delivery
		delivery.Evidence = slices.Clone(value.Delivery.Evidence)
		value.Delivery = &delivery
	}
	value.RelatedTasks = slices.Clone(value.RelatedTasks)
	value.Relationships.OtherChildren = slices.Clone(value.Relationships.OtherChildren)
	value.Relationships.Blockers = slices.Clone(value.Relationships.Blockers)
	value.Relationships.Dependents = slices.Clone(value.Relationships.Dependents)
	for index := range value.Relationships.Dependents {
		value.Relationships.Dependents[index].Blockers = slices.Clone(value.Relationships.Dependents[index].Blockers)
	}
	for index := range value.RelatedTasks {
		value.RelatedTasks[index].BlockedBy = slices.Clone(value.RelatedTasks[index].BlockedBy)
		value.RelatedTasks[index].Review = slices.Clone(value.RelatedTasks[index].Review)
	}
	return value
}
