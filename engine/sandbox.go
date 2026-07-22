package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SandboxResourceLabel             = "label"
	SandboxResourceProjectField      = "project-field"
	SandboxResourceProjectOption     = "project-option"
	SandboxResourceProjectView       = "project-view"
	SandboxResourceProjectItemField  = "project-item-field"
	SandboxResourceProjectWorkflow   = "project-workflow"
	SandboxResourceProjectItemProof  = "project-item-proof"
	SandboxResourceRuleset           = "ruleset"
	SandboxResourceFixtureIssue      = "fixture-issue"
	SandboxResourceFixtureBranch     = "fixture-branch"
	SandboxResourceFixturePR         = "fixture-pr"
	SandboxResourceFixtureWorkflow   = "fixture-workflow"
	SandboxResourceFixtureReview     = "fixture-review"
	SandboxResourceIssueRelationship = "issue-relationship"
	SandboxResourceRepositoryFile    = "repository-file"
	SandboxResourceFixtureDenial     = "fixture-denial-proof"
	SandboxResourceTokenRevocation   = "token-revocation"
	SandboxResourcePresent           = "present"
	SandboxResourceAbsent            = "absent"
)

const sandboxStatePath = ".starter-kit/sandbox/state.json"

var supportedSandboxResourceKinds = map[string]struct{}{
	SandboxResourceLabel:             {},
	SandboxResourceProjectField:      {},
	SandboxResourceProjectOption:     {},
	SandboxResourceProjectView:       {},
	SandboxResourceProjectItemField:  {},
	SandboxResourceProjectWorkflow:   {},
	SandboxResourceProjectItemProof:  {},
	SandboxResourceRuleset:           {},
	SandboxResourceFixtureIssue:      {},
	SandboxResourceFixtureBranch:     {},
	SandboxResourceFixturePR:         {},
	SandboxResourceFixtureWorkflow:   {},
	SandboxResourceFixtureReview:     {},
	SandboxResourceIssueRelationship: {},
	SandboxResourceRepositoryFile:    {},
	SandboxResourceFixtureDenial:     {},
	SandboxResourceTokenRevocation:   {},
}

// SandboxTarget binds bootstrap work to one approved owner, repository, and Project.
type SandboxTarget struct {
	Host           string `json:"host"`
	OwnerID        string `json:"owner_id"`
	RepositoryID   string `json:"repository_id"`
	ProjectID      string `json:"project_id"`
	RepositoryName string `json:"repository_name"`
}

// SandboxResourceSpec describes one managed external resource without credentials.
type SandboxResourceSpec struct {
	Key          string            `json:"key"`
	Kind         string            `json:"kind"`
	Name         string            `json:"name"`
	Marker       string            `json:"marker,omitempty"`
	DesiredState string            `json:"desired_state,omitempty"`
	Attributes   map[string]string `json:"attributes"`
}

// SandboxAuthorityProfile makes the non-resource authority route part of every
// immutable plan and execution mandate.
type SandboxAuthorityProfile struct {
	CredentialIdentities []string `json:"credential_identities"`
	Permissions          []string `json:"permissions"`
	EvidenceMode         string   `json:"evidence_mode"`
	Compatibility        string   `json:"compatibility"`
	DataClass            string   `json:"data_classification"`
	CostCeiling          string   `json:"cost_ceiling"`
	Destructive          string   `json:"destructive_ceiling"`
	Retention            string   `json:"retention"`
}

// SandboxManifest is the approved desired state for one isolated contract sandbox.
type SandboxManifest struct {
	SchemaVersion         int                     `json:"schema_version"`
	OperationID           string                  `json:"operation_id"`
	SourceRevision        string                  `json:"source_revision"`
	ConfigurationRevision string                  `json:"configuration_revision"`
	ApprovedBy            string                  `json:"approved_by"`
	ApprovedPlan          string                  `json:"approved_plan"`
	RecoveryOwner         string                  `json:"recovery_owner"`
	MarkerPrefix          string                  `json:"marker_prefix"`
	Target                SandboxTarget           `json:"target"`
	Authority             SandboxAuthorityProfile `json:"authority"`
	Resources             []SandboxResourceSpec   `json:"resources"`
}

// SandboxRequest selects the local evidence repository and approved manifest.
type SandboxRequest struct {
	Repository string          `json:"repository"`
	Manifest   SandboxManifest `json:"manifest"`
}

// SandboxCapability is the adapter-reported, expiring bootstrap authority snapshot.
type SandboxCapability struct {
	SchemaVersion         int           `json:"schema_version"`
	Available             bool          `json:"available"`
	Fresh                 bool          `json:"fresh"`
	Actor                 string        `json:"actor"`
	EvidenceMode          string        `json:"evidence_mode"`
	Target                SandboxTarget `json:"target"`
	Permissions           []string      `json:"permissions"`
	CredentialIdentities  []string      `json:"credential_identities"`
	Compatibility         string        `json:"compatibility"`
	ConfigurationRevision string        `json:"configuration_revision"`
	Problems              []string      `json:"problems,omitempty"`
	ObservedAt            time.Time     `json:"observed_at"`
	ExpiresAt             time.Time     `json:"expires_at"`
}

// SandboxObservedResource is one normalized external resource observation.
type SandboxObservedResource struct {
	Key        string            `json:"key"`
	Kind       string            `json:"kind"`
	Name       string            `json:"name"`
	ID         string            `json:"id"`
	Marker     string            `json:"marker,omitempty"`
	Attributes map[string]string `json:"attributes"`
}

// SandboxObservation is a credential-free snapshot of the allowlisted target.
type SandboxObservation struct {
	SchemaVersion         int                       `json:"schema_version"`
	Revision              string                    `json:"revision"`
	Target                SandboxTarget             `json:"target"`
	ConfigurationRevision string                    `json:"configuration_revision"`
	Resources             []SandboxObservedResource `json:"resources"`
	Problems              []string                  `json:"problems,omitempty"`
}

// SandboxInspection binds the approved manifest to current authority and observations.
type SandboxInspection struct {
	SchemaVersion int                `json:"schema_version"`
	ID            string             `json:"inspection_id"`
	Repository    string             `json:"repository"`
	Manifest      SandboxManifest    `json:"manifest"`
	Capability    SandboxCapability  `json:"capability"`
	Observation   SandboxObservation `json:"observation"`
	Disposition   string             `json:"disposition"`
	Problems      []string           `json:"problems"`
}

// SandboxEffect is one semantic resource correction selected by the engine.
type SandboxEffect struct {
	ID       string              `json:"effect_id"`
	Kind     string              `json:"kind"`
	Attempt  int                 `json:"attempt"`
	Resource SandboxResourceSpec `json:"resource"`
}

// SandboxPlan is immutable and bound to the reviewed source and observation.
type SandboxPlan struct {
	SchemaVersion         int                     `json:"schema_version"`
	ID                    string                  `json:"plan_id"`
	Repository            string                  `json:"repository"`
	OperationID           string                  `json:"operation_id"`
	SourceRevision        string                  `json:"source_revision"`
	ConfigurationRevision string                  `json:"configuration_revision"`
	InspectionID          string                  `json:"inspection_id"`
	ObservationRevision   string                  `json:"observation_revision"`
	Target                SandboxTarget           `json:"target"`
	ProvisioningPlan      string                  `json:"provisioning_plan"`
	RecoveryOwner         string                  `json:"recovery_owner"`
	Authority             SandboxAuthorityProfile `json:"authority"`
	Effects               []SandboxEffect         `json:"effects"`
	NoChange              bool                    `json:"no_change"`
}

// SandboxExecutionMandate authorizes a bounded family of semantic sandbox effects.
// Exact plans remain content-addressed evidence and must be contained by this mandate.
type SandboxExecutionMandate struct {
	SchemaVersion   int                     `json:"schema_version"`
	ID              string                  `json:"mandate_id"`
	ApprovedBy      string                  `json:"approved_by"`
	ApprovalID      string                  `json:"approval_id"`
	ApprovedAt      time.Time               `json:"approved_at"`
	ExpiresAt       time.Time               `json:"expires_at"`
	Target          SandboxTarget           `json:"target"`
	Actors          []string                `json:"actors"`
	MarkerPrefix    string                  `json:"marker_prefix"`
	UnmarkedKeys    []string                `json:"unmarked_resource_keys"`
	ResourceKinds   []string                `json:"resource_kinds"`
	EffectKinds     []string                `json:"effect_kinds"`
	ResourceDigests []string                `json:"resource_digests"`
	MaxEffects      int                     `json:"max_effects"`
	DataClass       string                  `json:"data_classification"`
	CostCeiling     string                  `json:"cost_ceiling"`
	Destructive     string                  `json:"destructive_ceiling"`
	Retention       string                  `json:"retention"`
	RecoveryOwner   string                  `json:"recovery_owner"`
	Authority       SandboxAuthorityProfile `json:"authority"`
}

// SandboxPlanApproval accepts legacy exact-plan approval (schema 1) or one approved
// execution mandate (schema 2). The legacy name remains wire-compatible with v1 evidence.
type SandboxPlanApproval struct {
	SchemaVersion int                      `json:"schema_version"`
	PlanID        string                   `json:"plan_id,omitempty"`
	ApprovedBy    string                   `json:"approved_by,omitempty"`
	ApprovalID    string                   `json:"approval_id,omitempty"`
	ApprovedAt    time.Time                `json:"approved_at,omitempty"`
	Mandate       *SandboxExecutionMandate `json:"mandate,omitempty"`
}

// SandboxEffectResult is the adapter's explicit result for one semantic effect.
type SandboxEffectResult struct {
	Outcome    string `json:"outcome"`
	ResourceID string `json:"resource_id,omitempty"`
	Detail     string `json:"detail"`
}

// SandboxEffectReceipt records one attributable external-resource attempt.
type SandboxEffectReceipt struct {
	SchemaVersion int       `json:"schema_version"`
	PlanID        string    `json:"plan_id"`
	EffectID      string    `json:"effect_id"`
	ResourceKey   string    `json:"resource_key"`
	ResourceKind  string    `json:"resource_kind"`
	ResourceID    string    `json:"resource_id,omitempty"`
	Actor         string    `json:"actor"`
	EvidenceMode  string    `json:"evidence_mode"`
	Outcome       string    `json:"outcome"`
	Detail        string    `json:"detail"`
	RecoveryOwner string    `json:"recovery_owner"`
	Authorization string    `json:"authorization_id,omitempty"`
	RecordedAt    time.Time `json:"recorded_at"`
}

type SandboxApplyStatus string

const (
	SandboxApplyApplied  SandboxApplyStatus = "applied"
	SandboxApplyNoChange SandboxApplyStatus = "no_change"
	SandboxApplyNonPass  SandboxApplyStatus = "non_pass"
)

type SandboxApplyResult struct {
	SchemaVersion int                    `json:"schema_version"`
	PlanID        string                 `json:"plan_id"`
	Status        SandboxApplyStatus     `json:"status"`
	Receipts      []SandboxEffectReceipt `json:"receipts"`
	Problems      []string               `json:"problems"`
}

type SandboxVerificationResult struct {
	SchemaVersion int             `json:"schema_version"`
	OverallState  ControlState    `json:"overall_state"`
	Controls      []ControlResult `json:"controls"`
	VerifiedAt    time.Time       `json:"verified_at"`
}

type SandboxStatus struct {
	SchemaVersion int                    `json:"schema_version"`
	Disposition   string                 `json:"disposition"`
	PlanID        string                 `json:"plan_id,omitempty"`
	Receipts      []SandboxEffectReceipt `json:"receipts"`
	Problems      []string               `json:"problems"`
}

type sandboxState struct {
	SchemaVersion int                        `json:"schema_version"`
	StateDigest   string                     `json:"state_digest"`
	OperationID   string                     `json:"operation_id"`
	Plan          SandboxPlan                `json:"plan"`
	Receipts      []SandboxEffectReceipt     `json:"receipts"`
	Verification  *SandboxVerificationResult `json:"verification,omitempty"`
	Disposition   string                     `json:"disposition"`
	Problems      []string                   `json:"problems"`
}

type SandboxLifecycleResult struct {
	SchemaVersion int                       `json:"schema_version"`
	Inspection    SandboxInspection         `json:"inspection"`
	Plan          SandboxPlan               `json:"plan"`
	Apply         SandboxApplyResult        `json:"apply"`
	Verification  SandboxVerificationResult `json:"verification"`
	Status        SandboxStatus             `json:"status"`
}

type SandboxPlanningResult struct {
	SchemaVersion int               `json:"schema_version"`
	Inspection    SandboxInspection `json:"inspection"`
	Plan          SandboxPlan       `json:"plan"`
}

// SandboxAdapter is the external-resource seam; approval and desired policy stay in the engine.
type SandboxAdapter interface {
	Capability(context.Context) (SandboxCapability, error)
	Observe(context.Context, SandboxTarget) (SandboxObservation, error)
	Apply(context.Context, SandboxEffect) (SandboxEffectResult, error)
}

func (e *Engine) InspectSandbox(ctx context.Context, request SandboxRequest) (SandboxInspection, error) {
	if e.sandboxAdapter == nil {
		return SandboxInspection{}, errors.New("sandbox inspection requires a sandbox adapter")
	}
	root, err := cleanRepositoryRoot(request.Repository)
	if err != nil {
		return SandboxInspection{}, err
	}
	if err := validateSandboxManifest(request.Manifest); err != nil {
		return SandboxInspection{}, err
	}
	capability, err := e.sandboxAdapter.Capability(ctx)
	if err != nil {
		return SandboxInspection{}, fmt.Errorf("inspect sandbox capability: %w", err)
	}
	observation, err := e.sandboxAdapter.Observe(ctx, request.Manifest.Target)
	if err != nil {
		return SandboxInspection{}, fmt.Errorf("inspect sandbox observation: %w", err)
	}
	if observation.Revision == "" {
		observation.Revision = digestJSON(observation.Resources)
	}
	problems := sandboxHandshakeProblems(request.Manifest, capability, observation, e.clock.Now())
	disposition := "inspected"
	if len(problems) != 0 {
		disposition = "non-pass"
	}
	inspection := SandboxInspection{SchemaVersion: 1, Repository: root, Manifest: cloneSandboxManifest(request.Manifest), Capability: cloneSandboxCapability(capability), Observation: cloneSandboxObservation(observation), Disposition: disposition, Problems: problems}
	inspection.ID = digestJSON(struct {
		Repository  string
		Manifest    SandboxManifest
		Capability  SandboxCapability
		Observation SandboxObservation
	}{root, inspection.Manifest, capability, observation})
	return inspection, nil
}

func (e *Engine) PlanSandbox(_ context.Context, inspection SandboxInspection) (SandboxPlan, error) {
	if inspection.Disposition != "inspected" || len(inspection.Problems) != 0 {
		return SandboxPlan{}, fmt.Errorf("sandbox inspection is non-pass: %s", strings.Join(inspection.Problems, "; "))
	}
	observed := make(map[string]SandboxObservedResource, len(inspection.Observation.Resources))
	for _, resource := range inspection.Observation.Resources {
		observed[resource.Key] = resource
	}
	effects := make([]SandboxEffect, 0)
	for _, resource := range inspection.Manifest.Resources {
		current, exists := observed[resource.Key]
		desiredState := resource.DesiredState
		if desiredState == "" {
			desiredState = SandboxResourcePresent
		}
		if desiredState == SandboxResourceAbsent && !exists {
			continue
		}
		if desiredState == SandboxResourcePresent && exists && sandboxResourceMatches(resource, current) {
			continue
		}
		effectKind := "reconcile-resource"
		if desiredState == SandboxResourceAbsent {
			effectKind = "remove-resource"
		}
		effect := SandboxEffect{Kind: effectKind, Attempt: 1, Resource: cloneSandboxResourceSpec(resource)}
		effect.ID = digestJSON(struct {
			OperationID string
			Resource    SandboxResourceSpec
		}{inspection.Manifest.OperationID, effect.Resource})
		effects = append(effects, effect)
	}
	plan := SandboxPlan{SchemaVersion: 1, Repository: inspection.Repository, OperationID: inspection.Manifest.OperationID, SourceRevision: inspection.Manifest.SourceRevision, ConfigurationRevision: inspection.Manifest.ConfigurationRevision, InspectionID: inspection.ID, ObservationRevision: inspection.Observation.Revision, Target: inspection.Manifest.Target, ProvisioningPlan: inspection.Manifest.ApprovedPlan, RecoveryOwner: inspection.Manifest.RecoveryOwner, Authority: cloneSandboxAuthority(inspection.Manifest.Authority), Effects: effects, NoChange: len(effects) == 0}
	plan.ID = digestJSON(plan)
	return plan, nil
}

func (e *Engine) ApplySandbox(ctx context.Context, plan SandboxPlan, approval SandboxPlanApproval) (SandboxApplyResult, error) {
	result := SandboxApplyResult{SchemaVersion: 1, PlanID: plan.ID, Status: SandboxApplyNoChange, Receipts: []SandboxEffectReceipt{}, Problems: []string{}}
	if e.sandboxAdapter == nil {
		return result, errors.New("sandbox apply requires a sandbox adapter")
	}
	if plan.ID == "" || plan.ID != digestJSON(sandboxPlanWithoutID(plan)) {
		return result, errors.New("sandbox apply requires the exact plan identifier")
	}
	authorizationID, err := validateSandboxAuthorization(plan, approval, e.clock.Now())
	if err != nil {
		return result, err
	}
	root, err := cleanRepositoryRoot(plan.Repository)
	if err != nil || root != plan.Repository {
		return result, errors.New("sandbox plan repository is invalid")
	}
	lockPath, err := lifecycleLockPath(ctx, root)
	if err != nil {
		return result, err
	}
	lease, _, _, err := acquireLifecycleLock(lockPath, plan.ID, e.clock.Now())
	if err != nil {
		return result, fmt.Errorf("acquire sandbox lifecycle lease: %w", err)
	}
	defer releaseLifecycleLock(lockPath, lease)
	if prior, readErr := readSandboxState(plan.Repository); readErr == nil && prior.OperationID == plan.OperationID {
		result.Receipts = slices.Clone(prior.Receipts)
	}
	if plan.NoChange {
		return result, writeSandboxApplyState(plan, result)
	}
	capability, err := e.sandboxAdapter.Capability(ctx)
	if err != nil {
		return result, fmt.Errorf("refresh sandbox capability: %w", err)
	}
	if !capability.Available || !capability.Fresh || capability.ConfigurationRevision != plan.ConfigurationRevision || !equalSandboxTarget(capability.Target, plan.Target) || !e.clock.Now().Before(capability.ExpiresAt) {
		result.Status = SandboxApplyNonPass
		result.Problems = []string{"sandbox plan is stale or authority is unavailable"}
		return result, writeSandboxApplyState(plan, result)
	}
	if approval.SchemaVersion == 2 && !slices.Contains(approval.Mandate.Actors, capability.Actor) {
		result.Status = SandboxApplyNonPass
		result.Problems = []string{"sandbox actor is outside approved mandate"}
		return result, writeSandboxApplyState(plan, result)
	}
	if approval.SchemaVersion == 2 && !sandboxCapabilityMatchesAuthority(capability, approval.Mandate.Authority) {
		result.Status = SandboxApplyNonPass
		result.Problems = []string{"sandbox credential identity, permissions, or compatibility are outside approved mandate"}
		return result, writeSandboxApplyState(plan, result)
	}
	observation, err := e.sandboxAdapter.Observe(ctx, plan.Target)
	if err != nil {
		return result, fmt.Errorf("refresh sandbox observation: %w", err)
	}
	if observation.Revision == "" {
		observation.Revision = digestJSON(observation.Resources)
	}
	if observation.Revision != plan.ObservationRevision || observation.ConfigurationRevision != plan.ConfigurationRevision || !equalSandboxTarget(observation.Target, plan.Target) {
		result.Status = SandboxApplyNonPass
		result.Problems = []string{"sandbox plan is stale because the observation changed"}
		return result, writeSandboxApplyState(plan, result)
	}
	result.Status = SandboxApplyApplied
	for _, effect := range plan.Effects {
		applied, applyErr := e.sandboxAdapter.Apply(ctx, effect)
		if applyErr != nil {
			result.Status = SandboxApplyNonPass
			result.Problems = append(result.Problems, effect.Resource.Key+": adapter effect failed")
			result.Receipts = append(result.Receipts, SandboxEffectReceipt{SchemaVersion: 1, PlanID: plan.ID, EffectID: effect.ID, ResourceKey: effect.Resource.Key, ResourceKind: effect.Resource.Kind, Actor: capability.Actor, EvidenceMode: capability.EvidenceMode, Outcome: "error", Detail: "sandbox adapter effect failed; inspect provider diagnostics outside retained evidence", RecoveryOwner: plan.RecoveryOwner, Authorization: authorizationID, RecordedAt: e.clock.Now()})
			break
		}
		receipt := SandboxEffectReceipt{SchemaVersion: 1, PlanID: plan.ID, EffectID: effect.ID, ResourceKey: effect.Resource.Key, ResourceKind: effect.Resource.Kind, ResourceID: applied.ResourceID, Actor: capability.Actor, EvidenceMode: capability.EvidenceMode, Outcome: applied.Outcome, Detail: applied.Detail, RecoveryOwner: plan.RecoveryOwner, Authorization: authorizationID, RecordedAt: e.clock.Now()}
		result.Receipts = append(result.Receipts, receipt)
		if applied.Outcome != "applied" && applied.Outcome != "no-change" {
			result.Status = SandboxApplyNonPass
			result.Problems = append(result.Problems, fmt.Sprintf("%s: %s", effect.Resource.Key, applied.Outcome))
			break
		}
	}
	return result, writeSandboxApplyState(plan, result)
}

func validateSandboxAuthorization(plan SandboxPlan, approval SandboxPlanApproval, now time.Time) (string, error) {
	if approval.SchemaVersion == 1 {
		if approval.PlanID != plan.ID || approval.ApprovedBy == "" || approval.ApprovalID == "" || approval.ApprovedAt.IsZero() {
			return "", errors.New("sandbox apply requires separate approval of the exact generated plan")
		}
		return approval.ApprovalID, nil
	}
	if approval.SchemaVersion != 2 || approval.Mandate == nil {
		return "", errors.New("sandbox apply requires an exact-plan approval or execution mandate")
	}
	mandate := *approval.Mandate
	if mandate.SchemaVersion != 1 || mandate.ID == "" || mandate.ID != digestJSON(sandboxExecutionMandateWithoutID(mandate)) || mandate.ApprovedBy == "" || mandate.ApprovalID == "" || mandate.ApprovedAt.IsZero() || mandate.ExpiresAt.IsZero() || mandate.ExpiresAt.Before(mandate.ApprovedAt) || len(mandate.Actors) == 0 || len(mandate.ResourceDigests) == 0 || !validSandboxAuthority(mandate.Authority) || mandate.DataClass != mandate.Authority.DataClass || mandate.CostCeiling != mandate.Authority.CostCeiling || mandate.Destructive != mandate.Authority.Destructive || mandate.Retention != mandate.Authority.Retention || mandate.MaxEffects < 0 {
		return "", errors.New("sandbox execution mandate is invalid")
	}
	if now.Before(mandate.ApprovedAt) || !now.Before(mandate.ExpiresAt) {
		return "", errors.New("sandbox execution mandate is not currently valid")
	}
	if !equalSandboxTarget(plan.Target, mandate.Target) || plan.RecoveryOwner != mandate.RecoveryOwner || len(plan.Effects) > mandate.MaxEffects || !equalSandboxAuthority(plan.Authority, mandate.Authority) {
		return "", errors.New("sandbox plan is outside approved mandate")
	}
	resourceKinds := make(map[string]struct{}, len(mandate.ResourceKinds))
	for _, kind := range mandate.ResourceKinds {
		resourceKinds[kind] = struct{}{}
	}
	effectKinds := make(map[string]struct{}, len(mandate.EffectKinds))
	for _, kind := range mandate.EffectKinds {
		effectKinds[kind] = struct{}{}
	}
	resourceDigests := make(map[string]struct{}, len(mandate.ResourceDigests))
	for _, digest := range mandate.ResourceDigests {
		resourceDigests[digest] = struct{}{}
	}
	for _, effect := range plan.Effects {
		if _, ok := resourceKinds[effect.Resource.Kind]; !ok {
			return "", errors.New("sandbox plan is outside approved mandate")
		}
		if _, ok := effectKinds[effect.Kind]; !ok {
			return "", errors.New("sandbox plan is outside approved mandate")
		}
		if _, ok := resourceDigests[sandboxResourceAuthorizationDigest(effect.Resource)]; !ok {
			return "", errors.New("sandbox plan is outside approved mandate")
		}
	}
	return mandate.ID, nil
}

func sandboxExecutionMandateWithoutID(value SandboxExecutionMandate) SandboxExecutionMandate {
	value.ID = ""
	return value
}

func validSandboxAuthority(value SandboxAuthorityProfile) bool {
	return len(value.CredentialIdentities) != 0 && len(value.Permissions) != 0 && value.EvidenceMode != "" && value.Compatibility != "" && value.DataClass != "" && value.CostCeiling != "" && value.Destructive != "" && value.Retention != ""
}

func emptySandboxAuthority(value SandboxAuthorityProfile) bool {
	return len(value.CredentialIdentities) == 0 && len(value.Permissions) == 0 && value.EvidenceMode == "" && value.Compatibility == "" && value.DataClass == "" && value.CostCeiling == "" && value.Destructive == "" && value.Retention == ""
}

func equalSandboxAuthority(left, right SandboxAuthorityProfile) bool {
	return sameSandboxStrings(left.CredentialIdentities, right.CredentialIdentities) && sameSandboxStrings(left.Permissions, right.Permissions) && left.EvidenceMode == right.EvidenceMode && left.Compatibility == right.Compatibility && left.DataClass == right.DataClass && left.CostCeiling == right.CostCeiling && left.Destructive == right.Destructive && left.Retention == right.Retention
}

func sandboxCapabilityMatchesAuthority(capability SandboxCapability, authority SandboxAuthorityProfile) bool {
	return capability.EvidenceMode == authority.EvidenceMode && capability.Compatibility == authority.Compatibility && sameSandboxStrings(capability.CredentialIdentities, authority.CredentialIdentities) && sameSandboxStrings(capability.Permissions, authority.Permissions)
}

func sameSandboxStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	left = slices.Clone(left)
	right = slices.Clone(right)
	sort.Strings(left)
	sort.Strings(right)
	return slices.Equal(left, right)
}

// BindSandboxExecutionMandate returns the content-addressed authority envelope retained
// by callers and receipts. It does not grant authority; ApprovedBy and ApprovalID must
// identify the separate human-owned decision record.

func BindSandboxExecutionMandate(value SandboxExecutionMandate, resources ...SandboxResourceSpec) SandboxExecutionMandate {
	value.ResourceDigests = value.ResourceDigests[:0]
	for _, resource := range resources {
		value.ResourceDigests = append(value.ResourceDigests, sandboxResourceAuthorizationDigest(resource))
	}
	sort.Strings(value.ResourceDigests)
	value.ID = digestJSON(sandboxExecutionMandateWithoutID(value))
	return value
}

func sandboxResourceAuthorizationDigest(resource SandboxResourceSpec) string {
	return digestJSON(cloneSandboxResourceSpec(resource))
}

func sandboxPlanWithoutID(value SandboxPlan) SandboxPlan {
	value.ID = ""
	return value
}

// SandboxStatus returns integrity-checked durable bootstrap state after interruption or restart.
func (e *Engine) SandboxStatus(_ context.Context, repository string) (SandboxStatus, error) {
	root, err := cleanRepositoryRoot(repository)
	if err != nil {
		return SandboxStatus{}, err
	}
	state, err := readSandboxState(root)
	if err != nil {
		return SandboxStatus{}, err
	}
	return SandboxStatus{SchemaVersion: 1, Disposition: state.Disposition, PlanID: state.Plan.ID, Receipts: slices.Clone(state.Receipts), Problems: slices.Clone(state.Problems)}, nil
}

func (e *Engine) VerifySandbox(ctx context.Context, manifest SandboxManifest) (SandboxVerificationResult, error) {
	if e.sandboxAdapter == nil {
		return SandboxVerificationResult{}, errors.New("sandbox verification requires a sandbox adapter")
	}
	capability, err := e.sandboxAdapter.Capability(ctx)
	if err != nil {
		return SandboxVerificationResult{}, fmt.Errorf("verify sandbox capability: %w", err)
	}
	observation, err := e.sandboxAdapter.Observe(ctx, manifest.Target)
	if err != nil {
		return SandboxVerificationResult{}, fmt.Errorf("verify sandbox observation: %w", err)
	}
	missing := sandboxHandshakeProblems(manifest, capability, observation, e.clock.Now())
	missing = append(missing, sandboxResourceProblems(manifest.Resources, observation.Resources)...)
	sort.Strings(missing)
	control := ControlResult{ID: "GITHUB-SANDBOX-001", State: ControlPass, Summary: "approved sandbox resources match the manifest", Evidence: []EvidenceReference{{Kind: "external-state", Target: manifest.Target.RepositoryName}}, Diagnostics: []string{}}
	if len(missing) != 0 {
		control.State = ControlFail
		control.Summary = "sandbox resources do not match the approved manifest"
		control.Rationale = strings.Join(missing, "; ")
		control.Diagnostics = slices.Clone(missing)
	}
	return SandboxVerificationResult{SchemaVersion: 1, OverallState: control.State, Controls: []ControlResult{control}, VerifiedAt: e.clock.Now()}, nil
}

func validateSandboxManifest(manifest SandboxManifest) error {
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return errors.New("sandbox manifest cannot be encoded")
	}
	if containsSensitiveText(string(encoded)) {
		return errors.New("sandbox manifest contains sensitive-looking material")
	}
	if manifest.SchemaVersion != 1 || manifest.OperationID == "" || manifest.SourceRevision == "" || manifest.ConfigurationRevision == "" || manifest.ApprovedBy == "" || manifest.ApprovedPlan == "" || manifest.RecoveryOwner == "" {
		return errors.New("sandbox manifest requires schema, operation, source, configuration, and approval identities")
	}
	if manifest.MarkerPrefix == "" || manifest.Target.Host == "" || manifest.Target.OwnerID == "" || manifest.Target.RepositoryID == "" || manifest.Target.ProjectID == "" || manifest.Target.RepositoryName == "" {
		return errors.New("sandbox manifest requires a marker prefix and immutable target identities")
	}
	if !emptySandboxAuthority(manifest.Authority) && !validSandboxAuthority(manifest.Authority) {
		return errors.New("sandbox manifest authority profile is incomplete")
	}
	seen := map[string]struct{}{}
	for _, resource := range manifest.Resources {
		if resource.Key == "" || resource.Kind == "" || resource.Name == "" {
			return errors.New("sandbox resources require key, kind, and name")
		}
		if _, supported := supportedSandboxResourceKinds[resource.Kind]; !supported {
			return fmt.Errorf("unsupported sandbox resource kind: %s", resource.Kind)
		}
		if resource.DesiredState != "" && resource.DesiredState != SandboxResourcePresent && resource.DesiredState != SandboxResourceAbsent {
			return fmt.Errorf("unsupported sandbox resource desired state: %s", resource.DesiredState)
		}
		if resource.DesiredState == SandboxResourceAbsent && (resource.Marker == "" || !strings.HasPrefix(resource.Marker, manifest.MarkerPrefix)) {
			return fmt.Errorf("sandbox cleanup resource %s requires an exact approved marker", resource.Key)
		}
		if resource.Kind == SandboxResourceProjectItemField && (resource.Attributes["content_id"] == "" || resource.Attributes["field"] == "" || resource.Attributes["field_id"] == "" || resource.Attributes["option_id"] == "") {
			return fmt.Errorf("Project item field resource %s requires immutable content, field, and option identities", resource.Key)
		}
		if resource.Kind == SandboxResourceProjectView && !slices.Contains([]string{"table", "board", "roadmap"}, resource.Attributes["layout"]) {
			return fmt.Errorf("Project view resource %s requires a supported layout", resource.Key)
		}
		if resource.Kind == SandboxResourceRuleset {
			definition, err := canonicalSandboxJSON(resource.Attributes["input:definition"])
			if err != nil || resource.Attributes["definition"] != definition || resource.Attributes["definition_sha256"] != digestBytes([]byte(definition)) || resource.Marker == "" || !strings.HasPrefix(resource.Marker, manifest.MarkerPrefix) {
				return fmt.Errorf("ruleset resource %s requires an exact canonical definition, digest, and approved marker", resource.Key)
			}
		}
		if resource.Kind == SandboxResourceIssueRelationship {
			relationship := resource.Attributes["relationship"]
			if !slices.Contains([]string{"parent-sub-issue", "blocker-dependent"}, relationship) || !validPositiveDecimal(resource.Attributes["source_number"]) || !validPositiveDecimal(resource.Attributes["source_id"]) || resource.Attributes["source_node_id"] == "" || !validPositiveDecimal(resource.Attributes["target_number"]) || !validPositiveDecimal(resource.Attributes["target_id"]) || resource.Attributes["target_node_id"] == "" || resource.Attributes["source_id"] == resource.Attributes["target_id"] {
				return fmt.Errorf("issue relationship resource %s requires a supported relationship and exact source and target issue identities", resource.Key)
			}
			if resource.Marker == "" || !strings.HasPrefix(resource.Marker, manifest.MarkerPrefix) {
				return fmt.Errorf("issue relationship resource %s requires an exact approved marker", resource.Key)
			}
		}
		if resource.Kind == SandboxResourceRepositoryFile {
			content := resource.Attributes["input:content"]
			if !validRepositoryFilePath(resource.Attributes["path"]) || resource.Attributes["branch"] == "" || resource.Attributes["content_sha256"] == "" || content == "" || resource.Marker == "" || !strings.HasPrefix(resource.Marker, manifest.MarkerPrefix) || !strings.Contains(content, resource.Marker) {
				return fmt.Errorf("repository file resource %s requires an exact path, branch, content digest, and marker-owned content", resource.Key)
			}
			if digestBytes([]byte(content)) != resource.Attributes["content_sha256"] {
				return fmt.Errorf("repository file resource %s content digest does not match approved content", resource.Key)
			}
		}
		if resource.Kind == SandboxResourceFixtureIssue && resource.Attributes["body_sha256"] != "" {
			body := resource.Attributes["input:body"]
			if body == "" || !strings.Contains(body, resource.Marker) || digestBytes([]byte(body)) != resource.Attributes["body_sha256"] {
				return fmt.Errorf("fixture issue resource %s body digest or marker ownership is invalid", resource.Key)
			}
		}
		if _, duplicate := seen[resource.Key]; duplicate {
			return fmt.Errorf("duplicate sandbox resource key: %s", resource.Key)
		}
		seen[resource.Key] = struct{}{}
	}
	return nil
}

func canonicalSandboxJSON(raw string) (string, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func validPositiveDecimal(value string) bool {
	parsed, err := strconv.ParseInt(value, 10, 64)
	return err == nil && parsed > 0 && strconv.FormatInt(parsed, 10) == value
}

func validRepositoryFilePath(value string) bool {
	if value == "" || value == "." || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || pathpkg.Clean(value) != value {
		return false
	}
	return !slices.Contains(strings.Split(value, "/"), "..")
}

func sandboxHandshakeProblems(manifest SandboxManifest, capability SandboxCapability, observation SandboxObservation, now time.Time) []string {
	problems := slices.Clone(capability.Problems)
	problems = append(problems, observation.Problems...)
	if capability.SchemaVersion != 1 || !capability.Available || !capability.Fresh {
		problems = append(problems, "sandbox capability is unavailable or stale")
	}
	if !now.Before(capability.ExpiresAt) {
		problems = append(problems, "sandbox capability is expired")
	}
	if capability.ConfigurationRevision != manifest.ConfigurationRevision || observation.ConfigurationRevision != manifest.ConfigurationRevision {
		problems = append(problems, "sandbox configuration revision does not match the manifest")
	}
	if !equalSandboxTarget(capability.Target, manifest.Target) || !equalSandboxTarget(observation.Target, manifest.Target) {
		problems = append(problems, "sandbox target identity does not match the manifest")
	}
	if !emptySandboxAuthority(manifest.Authority) && !sandboxCapabilityMatchesAuthority(capability, manifest.Authority) {
		problems = append(problems, "sandbox credential identity, permissions, evidence mode, or compatibility do not match the manifest")
	}
	keys := map[string]struct{}{}
	desiredNames := make(map[string]string, len(manifest.Resources))
	for _, resource := range manifest.Resources {
		desiredNames[sandboxResourceIdentity(resource.Kind, resource.Name, resource.Attributes)] = resource.Key
	}
	for _, resource := range observation.Resources {
		if _, duplicate := keys[resource.Key]; duplicate {
			problems = append(problems, "duplicate observed sandbox resource: "+resource.Key)
		}
		keys[resource.Key] = struct{}{}
		if desiredKey, collision := desiredNames[sandboxResourceIdentity(resource.Kind, resource.Name, resource.Attributes)]; collision && desiredKey != resource.Key {
			problems = append(problems, "unrecognized resource collides with managed kind/name: "+resource.Kind+"/"+resource.Name)
		}
	}
	sort.Strings(problems)
	return problems
}

func sandboxResourceIdentity(kind, name string, attributes map[string]string) string {
	identity := kind + "\x00" + name
	if kind == SandboxResourceProjectOption {
		identity += "\x00" + attributes["field"]
	}
	return identity
}

func sandboxResourceProblems(desired []SandboxResourceSpec, observed []SandboxObservedResource) []string {
	byKey := make(map[string]SandboxObservedResource, len(observed))
	for _, resource := range observed {
		byKey[resource.Key] = resource
	}
	problems := []string{}
	for _, resource := range desired {
		current, exists := byKey[resource.Key]
		desiredState := resource.DesiredState
		if desiredState == "" {
			desiredState = SandboxResourcePresent
		}
		if desiredState == SandboxResourceAbsent {
			if exists {
				problems = append(problems, "residual resource "+resource.Key)
			}
			continue
		}
		if !exists {
			problems = append(problems, "missing resource "+resource.Key)
			continue
		}
		if !sandboxResourceMatches(resource, current) {
			problems = append(problems, "drifted resource "+resource.Key)
		}
	}
	sort.Strings(problems)
	return problems
}

func sandboxResourceMatches(desired SandboxResourceSpec, observed SandboxObservedResource) bool {
	return desired.Key == observed.Key && desired.Kind == observed.Kind && desired.Name == observed.Name && desired.Marker == observed.Marker && equalStringMap(sandboxEvidenceAttributes(desired.Attributes), observed.Attributes)
}

func sandboxEvidenceAttributes(attributes map[string]string) map[string]string {
	result := make(map[string]string, len(attributes))
	for key, value := range attributes {
		if !strings.HasPrefix(key, "input:") {
			result[key] = value
		}
	}
	return result
}

func equalSandboxTarget(left, right SandboxTarget) bool { return left == right }

func cloneSandboxManifest(value SandboxManifest) SandboxManifest {
	value.Authority = cloneSandboxAuthority(value.Authority)
	resources := make([]SandboxResourceSpec, len(value.Resources))
	for index, resource := range value.Resources {
		resources[index] = cloneSandboxResourceSpec(resource)
	}
	value.Resources = resources
	return value
}

func cloneSandboxAuthority(value SandboxAuthorityProfile) SandboxAuthorityProfile {
	value.CredentialIdentities = slices.Clone(value.CredentialIdentities)
	value.Permissions = slices.Clone(value.Permissions)
	return value
}

func cloneSandboxResourceSpec(value SandboxResourceSpec) SandboxResourceSpec {
	value.Attributes = cloneStringMap(value.Attributes)
	return value
}

func cloneSandboxCapability(value SandboxCapability) SandboxCapability {
	value.Permissions = slices.Clone(value.Permissions)
	value.CredentialIdentities = slices.Clone(value.CredentialIdentities)
	value.Problems = slices.Clone(value.Problems)
	return value
}

func cloneSandboxObservation(value SandboxObservation) SandboxObservation {
	value.Problems = slices.Clone(value.Problems)
	resources := make([]SandboxObservedResource, len(value.Resources))
	for index, resource := range value.Resources {
		resource.Attributes = cloneStringMap(resource.Attributes)
		resources[index] = resource
	}
	value.Resources = resources
	return value
}

// InMemorySandboxAdapter is the credential-free contract double for deterministic bootstrap behavior.
type InMemorySandboxAdapter struct {
	mu          sync.Mutex
	capability  SandboxCapability
	observation SandboxObservation
	effects     []SandboxEffect
	results     []queuedSandboxResult
}

type queuedSandboxResult struct {
	result        SandboxEffectResult
	err           error
	observeEffect bool
}

func NewInMemorySandboxAdapter(capability SandboxCapability, observation SandboxObservation) *InMemorySandboxAdapter {
	if observation.Revision == "" {
		observation.Revision = digestJSON(observation.Resources)
	}
	return &InMemorySandboxAdapter{capability: cloneSandboxCapability(capability), observation: cloneSandboxObservation(observation)}
}

func (adapter *InMemorySandboxAdapter) Capability(context.Context) (SandboxCapability, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return cloneSandboxCapability(adapter.capability), nil
}

func (adapter *InMemorySandboxAdapter) Observe(context.Context, SandboxTarget) (SandboxObservation, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return cloneSandboxObservation(adapter.observation), nil
}

func (adapter *InMemorySandboxAdapter) Apply(_ context.Context, effect SandboxEffect) (SandboxEffectResult, error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	adapter.effects = append(adapter.effects, effect)
	if len(adapter.results) != 0 {
		queued := adapter.results[0]
		adapter.results = adapter.results[1:]
		if queued.err != nil {
			return SandboxEffectResult{}, queued.err
		}
		if queued.observeEffect {
			adapter.applyObservedResource(effect, queued.result.ResourceID)
		}
		return queued.result, nil
	}
	resourceID := "memory:" + effect.Resource.Key
	adapter.applyObservedResource(effect, resourceID)
	return SandboxEffectResult{Outcome: "applied", ResourceID: resourceID, Detail: "in-memory sandbox resource reconciled"}, nil
}

func (adapter *InMemorySandboxAdapter) applyObservedResource(effect SandboxEffect, resourceID string) {
	if effect.Kind == "remove-resource" {
		for index, current := range adapter.observation.Resources {
			if current.Key == effect.Resource.Key {
				adapter.observation.Resources = append(adapter.observation.Resources[:index], adapter.observation.Resources[index+1:]...)
				break
			}
		}
		adapter.observation.Revision = digestJSON(adapter.observation.Resources)
		return
	}
	if resourceID == "" {
		resourceID = "memory:" + effect.Resource.Key
	}
	observed := SandboxObservedResource{Key: effect.Resource.Key, Kind: effect.Resource.Kind, Name: effect.Resource.Name, ID: resourceID, Marker: effect.Resource.Marker, Attributes: cloneStringMap(effect.Resource.Attributes)}
	replaced := false
	for index, current := range adapter.observation.Resources {
		if current.Key == observed.Key {
			adapter.observation.Resources[index] = observed
			replaced = true
			break
		}
	}
	if !replaced {
		adapter.observation.Resources = append(adapter.observation.Resources, observed)
	}
	sort.Slice(adapter.observation.Resources, func(left, right int) bool {
		return adapter.observation.Resources[left].Key < adapter.observation.Resources[right].Key
	})
	adapter.observation.Revision = digestJSON(adapter.observation.Resources)
}

func (adapter *InMemorySandboxAdapter) Effects() []SandboxEffect {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return slices.Clone(adapter.effects)
}

func (adapter *InMemorySandboxAdapter) Observation() SandboxObservation {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	return cloneSandboxObservation(adapter.observation)
}

func (adapter *InMemorySandboxAdapter) SetObservation(observation SandboxObservation) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	if observation.Revision == "" {
		observation.Revision = digestJSON(observation.Resources)
	}
	adapter.observation = cloneSandboxObservation(observation)
}

func (adapter *InMemorySandboxAdapter) QueueApplyResult(result SandboxEffectResult, observeEffect bool) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	adapter.results = append(adapter.results, queuedSandboxResult{result: result, observeEffect: observeEffect})
}

func (adapter *InMemorySandboxAdapter) QueueApplyError(err error) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	adapter.results = append(adapter.results, queuedSandboxResult{err: err})
}

func sandboxStateFile(root string) string {
	return filepath.Join(root, filepath.FromSlash(sandboxStatePath))
}

func writeSandboxApplyState(plan SandboxPlan, result SandboxApplyResult) error {
	disposition := "converged"
	if result.Status == SandboxApplyNonPass {
		disposition = "non-pass"
	}
	state := sandboxState{SchemaVersion: 1, OperationID: plan.OperationID, Plan: plan, Receipts: slices.Clone(result.Receipts), Disposition: disposition, Problems: slices.Clone(result.Problems)}
	return writeSandboxState(plan.Repository, state)
}

func updateSandboxVerification(root string, verification SandboxVerificationResult) error {
	state, err := readSandboxState(root)
	if err != nil {
		return err
	}
	state.Verification = &verification
	if verification.OverallState != ControlPass {
		state.Disposition = "non-pass"
		state.Problems = append(state.Problems, "sandbox verification did not pass")
	}
	return writeSandboxState(root, state)
}

func writeSandboxState(root string, state sandboxState) error {
	path := sandboxStateFile(root)
	if err := ensureNoSymlinkParents(root, sandboxStatePath); err != nil {
		return fmt.Errorf("validate sandbox state path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create sandbox state directory: %w", err)
	}
	if err := ensureNoSymlinkParents(root, sandboxStatePath); err != nil {
		return fmt.Errorf("validate sandbox state directory: %w", err)
	}
	state.StateDigest = ""
	state.StateDigest = digestJSON(state)
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sandbox state: %w", err)
	}
	if containsSensitiveText(string(content)) {
		return errors.New("sandbox state contains sensitive-looking material")
	}
	content = append(content, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".state-*.tmp")
	if err != nil {
		return fmt.Errorf("create sandbox state staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write sandbox state: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync sandbox state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("commit sandbox state: %w", err)
	}
	return nil
}

func readSandboxState(root string) (sandboxState, error) {
	if err := ensureNoSymlinkComponents(root, sandboxStatePath); err != nil {
		return sandboxState{}, fmt.Errorf("validate sandbox state path: %w", err)
	}
	content, err := os.ReadFile(sandboxStateFile(root))
	if err != nil {
		return sandboxState{}, fmt.Errorf("read sandbox state: %w", err)
	}
	var state sandboxState
	if err := json.Unmarshal(content, &state); err != nil {
		return sandboxState{}, fmt.Errorf("parse sandbox state: %w", err)
	}
	if state.SchemaVersion != 1 {
		return sandboxState{}, errors.New("unsupported sandbox state schema")
	}
	recordedDigest := state.StateDigest
	state.StateDigest = ""
	if recordedDigest == "" || recordedDigest != digestJSON(state) {
		return sandboxState{}, errors.New("sandbox state integrity is invalid")
	}
	state.StateDigest = recordedDigest
	return state, nil
}
