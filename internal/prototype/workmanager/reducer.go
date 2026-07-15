package workmanager

import (
	"fmt"
	"slices"
)

var requiredPermissions = []string{"issues:write", "projects:write", "pull_requests:read"}

func InitialState() State {
	desired := DesiredIntent{
		SchemaVersion: 1, SourceRevision: "issue-64:v1", OperationID: "work-manager-prototype-64",
		InputDigests: map[string]string{"issue-brief": "sha256:issue-64-v1", "governing-decisions": "sha256:decisions-v1"},
		Credential:   CredentialExpectation{Mode: "user-token", Actor: "dragondad22"},
		Target: Target{
			Host: "github.com", RepositoryID: "R_codex-starter-kit", ProjectID: "PVT_project-8",
			ConfigurationRevision: "project-config:v1",
			FieldIDs:              map[string]string{"readiness": "field-readiness", "status": "field-status", "phase": "field-phase"},
			OptionIDs:             map[string]string{"readiness:ready": "ready-v1", "readiness:blocked": "blocked-v1", "readiness:needs-refinement": "refine-v1", "status:backlog": "backlog-v1", "status:in-progress": "progress-v1", "status:done": "done-v1"},
		},
		WorkItems: []DesiredWorkItem{
			{ManagedID: "issue:4", IssueType: "feature", Title: "Build the GitHub executable-work system", Readiness: "needs-refinement", Status: "in-progress", Phase: "Phase 3"},
			{ManagedID: "issue:15", IssueType: "bug", Title: "Closed issues do not transition Project Status to Done", ParentManagedID: "issue:4", Readiness: "ready", Status: "done", Closed: true},
			{ManagedID: "issue:16", IssueType: "question", Title: "Govern question and research work items", ParentManagedID: "issue:4", Readiness: "ready", Status: "done", PromotionRecord: "docs/decisions/DEC-0013-question-and-research-work.md", Closed: true},
			{ManagedID: "issue:46", IssueType: "task", Title: "Expose roadmap phase membership in the GitHub Project", ParentManagedID: "issue:4", Readiness: "ready", Status: "backlog"},
			{ManagedID: "issue:64", IssueType: "task", Title: "Prototype deterministic Work Manager reconciliation", ParentManagedID: "issue:4", Readiness: "ready", Status: "in-progress", Review: []ReviewRequirement{{Role: "change-review", DistinctContext: true}}},
			{ManagedID: "future:sandbox-matrix", IssueType: "research", Title: "Qualify live GitHub behavior safely", ParentManagedID: "issue:4", BlockedBy: []string{"issue:64"}, Readiness: "blocked", Status: "backlog"},
		},
	}
	observed := map[string]ObservedWorkItem{}
	for _, item := range desired.WorkItems[:5] {
		observed[item.ManagedID] = observedFromDesired(item, desired.Target)
	}
	// Seed the historical #15 defect: GitHub closed the issue while Project Status drifted.
	drift := observed["issue:15"]
	drift.StatusOption = desired.Target.OptionIDs["status:in-progress"]
	observed["issue:15"] = drift
	return State{
		SchemaVersion: 1, Desired: desired,
		Capability: Capability{
			Online: true, Fresh: true, Mode: "user-token", Actor: "dragondad22",
			Permissions: slices.Clone(requiredPermissions), RESTRemaining: 4990, GraphQLRemaining: 4990,
			ConfigurationRevision: "project-config:v1",
			ObservedAt:            "2026-07-15T00:00:00Z",
			Rate:                  RateState{Resource: "graphql", Limit: 5000, Used: 10, Remaining: 4990, ResetAt: "2026-07-15T01:00:00Z", MaxAttempts: 2, Disposition: "available"},
		},
		Observation: Observation{Revision: "github-observation:v1", ConfigurationRevision: "project-config:v1", Host: desired.Target.Host, RepositoryID: desired.Target.RepositoryID, ProjectID: desired.Target.ProjectID, FieldIDs: cloneMap(desired.Target.FieldIDs), OptionIDs: cloneMap(desired.Target.OptionIDs), WorkItems: observed},
		Receipts:    []EffectReceipt{}, Disposition: "unplanned", Message: "Seeded #15 status drift, governed #16 completion, inherited #46 phase context, and a #64-blocked sandbox ticket.",
	}
}

func Reduce(input State, action Action) State {
	state := cloneState(input)
	switch action {
	case PlanReconciliation:
		return plan(state)
	case ApplyNextSuccess:
		return applyNext(state, false)
	case LoseCreateResponse:
		return applyNext(state, true)
	case ObserveAmbiguous:
		return observeAmbiguous(state)
	case HitRateLimit:
		return rateLimited(state)
	case GoOffline:
		state.Capability.Online = false
		state.Capability.Fresh = false
		state.Plan = nil
		state.QueuedIntent = cloneIntentPointer(state.Desired)
		state.Disposition = "queued-offline"
		state.Message = "Stored credential-free desired intent; no HTTP request or secret was queued."
		return state
	case Reconnect:
		state.Capability.Online = true
		state.Capability.Fresh = false
		state.Plan = nil
		state.Disposition = "handshake-required"
		state.Message = "Connectivity returned, but identity, capability, and preconditions must be refreshed."
		return state
	case RefreshHandshake:
		state.Capability.Online = true
		state.Capability.Fresh = true
		state.Capability.Mode = state.Desired.Credential.Mode
		state.Capability.Actor = state.Desired.Credential.Actor
		state.Capability.Permissions = slices.Clone(requiredPermissions)
		state.Capability.ConfigurationRevision = state.Observation.ConfigurationRevision
		state.QueuedIntent = nil
		state.Plan = nil
		state.Disposition = "unplanned"
		state.Message = "Fresh adapter handshake matches the expected actor and minimum capability set."
		return state
	case MigrateFieldOption:
		state.Observation.Revision = nextObservationRevision(state.Observation.Revision)
		state.Observation.ConfigurationRevision = "project-config:v2"
		state.Observation.OptionIDs["readiness:ready"] = "ready-v2"
		state.Capability.ConfigurationRevision = "project-config:v2"
		state.Plan = nil
		state.Disposition = "stale"
		state.Message = "Observed Readiness option identity changed; the prior desired contract cannot be applied."
		return state
	case AcceptMigration:
		state.Desired.Target.ConfigurationRevision = state.Observation.ConfigurationRevision
		state.Desired.Target.OptionIDs = cloneMap(state.Observation.OptionIDs)
		state.Desired.SourceRevision = "issue-64:v2"
		state.Plan = nil
		state.Disposition = "unplanned"
		state.Message = "Work Manager accepted the observed identities as a new governed input; re-planning is required."
		return state
	case CompleteBlocker:
		for index := range state.Desired.WorkItems {
			if state.Desired.WorkItems[index].ManagedID == "issue:64" {
				state.Desired.WorkItems[index].Closed = true
				state.Desired.WorkItems[index].Status = "done"
			}
		}
		deriveLifecycle(&state.Desired)
		state.Desired.SourceRevision = "issue-64:completed"
		state.Plan = nil
		state.Disposition = "unplanned"
		state.Message = "Completing #64 promoted its dependent to Ready while leaving Status Backlog until explicitly selected."
		return state
	case ChangeSource:
		state.Desired.SourceRevision = "issue-64:changed-after-plan"
		state.Disposition = "stale"
		state.Message = "The governed desired source changed; any retained plan must fail apply preconditions."
		return state
	case ResetRate:
		state.Capability.Rate.Used = 10
		state.Capability.Rate.Remaining = 4990
		state.Capability.Rate.RetryAfterSeconds = 0
		state.Capability.Rate.Attempt = 0
		state.Capability.Rate.Disposition = "available"
		state.Capability.GraphQLRemaining = 4990
		state.Capability.Fresh = false
		state.Plan = nil
		state.Disposition = "handshake-required"
		state.Message = "The observed rate window reset; identity and capabilities still require a fresh handshake."
		return state
	default:
		state.Disposition = "needs-review"
		state.Message = "Unknown prototype action."
		return state
	}
}

func plan(state State) State {
	state.Plan = nil
	if !state.Capability.Online {
		state.QueuedIntent = cloneIntentPointer(state.Desired)
		state.Disposition = "queued-offline"
		state.Message = "Offline: retained desired intent only."
		return state
	}
	if !state.Capability.Fresh || state.Capability.Mode != state.Desired.Credential.Mode || state.Capability.Actor != state.Desired.Credential.Actor || !containsAll(state.Capability.Permissions, requiredPermissions) {
		state.Disposition = "handshake-required"
		state.Message = "Identity or minimum capabilities are stale or mismatched."
		return state
	}
	if state.Capability.Rate.Disposition == "exhausted" {
		state.Disposition = "rate-limited"
		state.Message = "The bounded retry count is exhausted until the recorded rate window resets."
		return state
	}
	if state.Observation.Host != state.Desired.Target.Host || state.Observation.RepositoryID != state.Desired.Target.RepositoryID || state.Observation.ProjectID != state.Desired.Target.ProjectID || state.Observation.ConfigurationRevision != state.Desired.Target.ConfigurationRevision || !equalMap(state.Observation.FieldIDs, state.Desired.Target.FieldIDs) || !equalMap(state.Observation.OptionIDs, state.Desired.Target.OptionIDs) {
		state.Disposition = "stale"
		state.Message = "Project, field, or option identity changed; discard the plan and refresh governed inputs."
		return state
	}
	effects := []Effect{}
	for _, desired := range state.Desired.WorkItems {
		observed, exists := state.Observation.WorkItems[desired.ManagedID]
		if !exists {
			effects = append(effects, Effect{ID: effectID("create-issue", desired.ManagedID), Kind: "create-issue", ManagedID: desired.ManagedID, Marker: "starter-kit-managed:" + desired.ManagedID, Title: desired.Title})
			continue
		}
		readiness := state.Desired.Target.OptionIDs["readiness:"+desired.Readiness]
		status := state.Desired.Target.OptionIDs["status:"+desired.Status]
		if observed.Title != desired.Title || observed.ReadinessOption != readiness || observed.StatusOption != status || observed.ParentManagedID != desired.ParentManagedID || !slices.Equal(observed.BlockedBy, desired.BlockedBy) || observed.Phase != desired.Phase || observed.PromotionRecord != desired.PromotionRecord || !slices.Equal(observed.Review, desired.Review) || observed.Closed != desired.Closed {
			closed := desired.Closed
			effects = append(effects, Effect{ID: effectID("reconcile-work-item", desired.ManagedID), Kind: "reconcile-work-item", ManagedID: desired.ManagedID, Title: desired.Title, ReadinessOption: readiness, StatusOption: status, ParentManagedID: desired.ParentManagedID, BlockedBy: slices.Clone(desired.BlockedBy), Phase: desired.Phase, PromotionRecord: desired.PromotionRecord, Review: slices.Clone(desired.Review), Closed: &closed})
		}
	}
	state.Plan = &Plan{
		SchemaVersion: 1, ID: "plan:" + state.Desired.SourceRevision + ":" + state.Observation.Revision,
		OperationID: state.Desired.OperationID, SourceRevision: state.Desired.SourceRevision,
		InputDigests:        cloneMap(state.Desired.InputDigests),
		ObservationRevision: state.Observation.Revision, ConfigurationRevision: state.Desired.Target.ConfigurationRevision,
		Host: state.Desired.Target.Host, RepositoryID: state.Desired.Target.RepositoryID, ProjectID: state.Desired.Target.ProjectID,
		FieldIDs: cloneMap(state.Desired.Target.FieldIDs), OptionIDs: cloneMap(state.Desired.Target.OptionIDs),
		Preconditions:     []string{"fresh expected actor", "minimum declared permissions", "immutable repository and Project IDs", "matching field and option identities", "unchanged desired source"},
		RequiredApprovals: []string{"approved desired source", "effect authority"},
		Impact:            []string{"GitHub issue and Project desired-state reconciliation"},
		Recovery:          "retain completed receipts and re-observe before planning the remaining semantic delta",
		ExpiresAt:         "2026-07-15T02:00:00Z",
		Effects:           effects,
	}
	state.Disposition = "planned"
	if len(effects) == 0 {
		state.Disposition = "converged"
	}
	state.Message = fmt.Sprintf("Immutable plan contains %d remaining semantic effect(s).", len(effects))
	return state
}

func applyNext(state State, ambiguous bool) State {
	if state.Plan == nil || len(state.Plan.Effects) == 0 {
		state.Disposition = "needs-review"
		state.Message = "No planned effect is available."
		return state
	}
	outcome := "applied"
	if ambiguous {
		outcome = "ambiguous"
	}
	return Apply(state, state.Plan.ID, outcome)
}

// Apply consumes an exact plan identity and rechecks every modeled precondition before
// representing one adapter effect result.
func Apply(input State, planID, outcome string) State {
	state := cloneState(input)
	if state.Plan == nil || state.Plan.ID != planID || len(state.Plan.Effects) == 0 {
		state.Plan = nil
		state.Disposition = "stale"
		state.Message = "The supplied immutable plan identity is absent or mismatched."
		return state
	}
	planSnapshot := *state.Plan
	if !planStillCurrent(state, planSnapshot) {
		state.Plan = nil
		state.Disposition = "stale"
		state.Message = "Apply rejected a stale source, observation, identity, configuration, or authority precondition before effects."
		return state
	}
	effect := state.Plan.Effects[0]
	if outcome == "ambiguous" && effect.Kind != "create-issue" {
		state.Disposition = "needs-review"
		state.Message = "Lost-response simulation applies only to a create effect."
		return state
	}
	if outcome != "ambiguous" && outcome != "applied" {
		state.Disposition = "needs-review"
		state.Message = "Unknown adapter effect outcome."
		return state
	}
	if outcome == "ambiguous" {
		state.Receipts = append(state.Receipts, newReceipt(state, planSnapshot, effect, "ambiguous", 1, "Transport lost the response; do not retry create until marker lookup resolves it."))
		state.AmbiguousManagedID = effect.ManagedID
		state.Plan = nil
		state.Disposition = "needs-review"
		state.Message = "Create outcome is ambiguous; stable managed identity is the recovery seam."
		return state
	}
	applyEffect(&state, effect)
	state.Receipts = append(state.Receipts, newReceipt(state, planSnapshot, effect, "applied", 1, "Postcondition represented by refreshed authoritative observation."))
	state.Plan = nil
	state.Disposition = "unplanned"
	state.Message = "One effect completed; re-observe and plan only the remaining semantic delta."
	return state
}

func planStillCurrent(state State, plan Plan) bool {
	return state.Capability.Online && state.Capability.Fresh &&
		state.Capability.Rate.Disposition != "exhausted" &&
		state.Capability.Mode == state.Desired.Credential.Mode && state.Capability.Actor == state.Desired.Credential.Actor &&
		containsAll(state.Capability.Permissions, requiredPermissions) &&
		plan.SourceRevision == state.Desired.SourceRevision && plan.ObservationRevision == state.Observation.Revision &&
		equalMap(plan.InputDigests, state.Desired.InputDigests) && state.Capability.ObservedAt < plan.ExpiresAt &&
		plan.ConfigurationRevision == state.Observation.ConfigurationRevision && plan.ConfigurationRevision == state.Desired.Target.ConfigurationRevision &&
		plan.Host == state.Observation.Host && plan.RepositoryID == state.Observation.RepositoryID && plan.ProjectID == state.Observation.ProjectID &&
		plan.Host == state.Desired.Target.Host && plan.RepositoryID == state.Desired.Target.RepositoryID && plan.ProjectID == state.Desired.Target.ProjectID &&
		equalMap(plan.FieldIDs, state.Observation.FieldIDs) && equalMap(plan.OptionIDs, state.Observation.OptionIDs) &&
		equalMap(plan.FieldIDs, state.Desired.Target.FieldIDs) && equalMap(plan.OptionIDs, state.Desired.Target.OptionIDs)
}

func applyEffect(state *State, effect Effect) {
	observed := state.Observation.WorkItems[effect.ManagedID]
	if effect.Kind == "create-issue" {
		observed = ObservedWorkItem{ManagedID: effect.ManagedID, IssueNodeID: "I_" + effect.ManagedID, ProjectItemID: "PVTI_" + effect.ManagedID, Title: effect.Title}
	}
	if effect.Title != "" {
		observed.Title = effect.Title
	}
	if effect.ReadinessOption != "" {
		observed.ReadinessOption = effect.ReadinessOption
	}
	if effect.StatusOption != "" {
		observed.StatusOption = effect.StatusOption
	}
	if effect.Closed != nil {
		observed.Closed = *effect.Closed
	}
	if effect.Kind == "reconcile-work-item" {
		observed.ParentManagedID = effect.ParentManagedID
		observed.BlockedBy = slices.Clone(effect.BlockedBy)
		observed.Phase = effect.Phase
		observed.PromotionRecord = effect.PromotionRecord
		observed.Review = slices.Clone(effect.Review)
		if observed.ProjectItemID == "" {
			observed.ProjectItemID = "PVTI_" + effect.ManagedID
		}
	}
	state.Observation.WorkItems[effect.ManagedID] = observed
	state.Observation.Revision = nextObservationRevision(state.Observation.Revision)
}

func observeAmbiguous(state State) State {
	if state.AmbiguousManagedID == "" {
		state.Disposition = "needs-review"
		state.Message = "There is no ambiguous managed create to look up."
		return state
	}
	managedID := state.AmbiguousManagedID
	var desired DesiredWorkItem
	for _, item := range state.Desired.WorkItems {
		if item.ManagedID == managedID {
			desired = item
		}
	}
	state.Observation.WorkItems[managedID] = ObservedWorkItem{ManagedID: managedID, IssueNodeID: "I_recovered", Title: desired.Title}
	state.Observation.Revision = nextObservationRevision(state.Observation.Revision)
	state.Receipts = append(state.Receipts, recoveryReceipt(state, managedID, "Found exactly one issue by stable non-secret managed marker; Project and relationship effects remain unobserved."))
	state.AmbiguousManagedID = ""
	state.Plan = nil
	state.Disposition = "unplanned"
	state.Message = "Ambiguous create resolved without duplication; re-plan against the discovered immutable IDs."
	return state
}

func rateLimited(state State) State {
	if state.Plan == nil || len(state.Plan.Effects) == 0 {
		state.Disposition = "needs-review"
		state.Message = "No planned mutation exists to receive a rate-limit result."
		return state
	}
	planSnapshot := *state.Plan
	effect := state.Plan.Effects[0]
	state.Capability.Rate.Attempt++
	state.Capability.Rate.Used = state.Capability.Rate.Limit
	state.Capability.Rate.Remaining = 0
	state.Capability.GraphQLRemaining = 0
	state.Capability.Rate.RetryAfterSeconds = 60
	state.Capability.Rate.Disposition = "retry-pending"
	state.Receipts = append(state.Receipts, newReceipt(state, planSnapshot, effect, "rate-limited", state.Capability.Rate.Attempt, "Retained rate resource, limit, use, reset, retry-after, attempt, and desired intent."))
	state.Capability.Fresh = false
	state.QueuedIntent = cloneIntentPointer(state.Desired)
	state.Plan = nil
	state.Disposition = "retry-pending"
	state.Message = "Queued desired intent, not the failed transport request; a fresh handshake and plan are required."
	if state.Capability.Rate.Attempt >= state.Capability.Rate.MaxAttempts {
		state.Capability.Rate.Disposition = "exhausted"
		state.Disposition = "rate-limited"
		state.Message = "Bounded retry count is exhausted; stop until the recorded reset and a fresh handshake."
	}
	return state
}

func observedFromDesired(item DesiredWorkItem, target Target) ObservedWorkItem {
	return ObservedWorkItem{ManagedID: item.ManagedID, IssueNodeID: "I_" + item.ManagedID, ProjectItemID: "PVTI_" + item.ManagedID, Title: item.Title, ReadinessOption: target.OptionIDs["readiness:"+item.Readiness], StatusOption: target.OptionIDs["status:"+item.Status], ParentManagedID: item.ParentManagedID, BlockedBy: slices.Clone(item.BlockedBy), Phase: item.Phase, PromotionRecord: item.PromotionRecord, Review: slices.Clone(item.Review), Closed: item.Closed}
}

func deriveLifecycle(intent *DesiredIntent) {
	closed := make(map[string]bool, len(intent.WorkItems))
	for _, item := range intent.WorkItems {
		closed[item.ManagedID] = item.Closed
	}
	for index := range intent.WorkItems {
		item := &intent.WorkItems[index]
		if item.Readiness != "blocked" || len(item.BlockedBy) == 0 {
			continue
		}
		allClosed := true
		for _, blocker := range item.BlockedBy {
			if !closed[blocker] {
				allClosed = false
				break
			}
		}
		if allClosed {
			item.Readiness = "ready"
		}
	}
	for parentIndex := range intent.WorkItems {
		parent := &intent.WorkItems[parentIndex]
		hasChild := false
		allChildrenClosed := true
		anyChildStarted := false
		for _, child := range intent.WorkItems {
			if child.ParentManagedID != parent.ManagedID {
				continue
			}
			hasChild = true
			allChildrenClosed = allChildrenClosed && child.Closed
			anyChildStarted = anyChildStarted || child.Closed || child.Status == "in-progress" || child.Status == "done"
		}
		if hasChild && allChildrenClosed {
			parent.Closed = true
			parent.Status = "done"
		} else if hasChild && anyChildStarted {
			parent.Status = "in-progress"
		}
	}
}

func newReceipt(state State, plan Plan, effect Effect, outcome string, attempt int, detail string) EffectReceipt {
	return EffectReceipt{
		SchemaVersion: 1, PlanID: plan.ID, OperationID: plan.OperationID, EffectID: effect.ID, EffectKind: effect.Kind, ManagedID: effect.ManagedID,
		Actor: state.Capability.Actor, CredentialMode: state.Capability.Mode, Authority: slices.Clone(state.Capability.Permissions),
		SourceRevision: plan.SourceRevision, ObservationRevision: plan.ObservationRevision,
		RepositoryID: plan.RepositoryID, ProjectID: plan.ProjectID,
		Outcome: outcome, Attempt: attempt, Detail: detail,
	}
}

func recoveryReceipt(state State, managedID, detail string) EffectReceipt {
	receipt := EffectReceipt{SchemaVersion: 1, EffectID: "lookup:" + managedID, EffectKind: "lookup-managed-marker", ManagedID: managedID, Outcome: "reconciled", Attempt: 1, Detail: detail}
	for index := len(state.Receipts) - 1; index >= 0; index-- {
		prior := state.Receipts[index]
		if prior.ManagedID != managedID {
			continue
		}
		receipt.PlanID = prior.PlanID
		receipt.OperationID = prior.OperationID
		receipt.Actor = prior.Actor
		receipt.CredentialMode = prior.CredentialMode
		receipt.Authority = slices.Clone(prior.Authority)
		receipt.SourceRevision = prior.SourceRevision
		receipt.ObservationRevision = state.Observation.Revision
		receipt.RepositoryID = prior.RepositoryID
		receipt.ProjectID = prior.ProjectID
		break
	}
	return receipt
}

func effectID(kind, managedID string) string        { return "effect:" + kind + ":" + managedID }
func nextObservationRevision(current string) string { return current + "+refresh" }

func containsAll(actual, required []string) bool {
	for _, want := range required {
		if !slices.Contains(actual, want) {
			return false
		}
	}
	return true
}

func equalMap(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func cloneMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func cloneIntentPointer(intent DesiredIntent) *DesiredIntent {
	cloned := cloneIntent(intent)
	return &cloned
}

func cloneIntent(intent DesiredIntent) DesiredIntent {
	result := intent
	result.InputDigests = cloneMap(intent.InputDigests)
	result.Target.FieldIDs = cloneMap(intent.Target.FieldIDs)
	result.Target.OptionIDs = cloneMap(intent.Target.OptionIDs)
	result.WorkItems = slices.Clone(intent.WorkItems)
	for index := range result.WorkItems {
		result.WorkItems[index].BlockedBy = slices.Clone(intent.WorkItems[index].BlockedBy)
		result.WorkItems[index].Review = slices.Clone(intent.WorkItems[index].Review)
	}
	return result
}

func cloneState(input State) State {
	result := input
	result.Desired = cloneIntent(input.Desired)
	result.Capability.Permissions = slices.Clone(input.Capability.Permissions)
	result.Observation.FieldIDs = cloneMap(input.Observation.FieldIDs)
	result.Observation.OptionIDs = cloneMap(input.Observation.OptionIDs)
	result.Observation.WorkItems = make(map[string]ObservedWorkItem, len(input.Observation.WorkItems))
	for key, value := range input.Observation.WorkItems {
		value.BlockedBy = slices.Clone(value.BlockedBy)
		value.Review = slices.Clone(value.Review)
		result.Observation.WorkItems[key] = value
	}
	result.Receipts = slices.Clone(input.Receipts)
	for index := range result.Receipts {
		result.Receipts[index].Authority = slices.Clone(input.Receipts[index].Authority)
	}
	if input.QueuedIntent != nil {
		result.QueuedIntent = cloneIntentPointer(*input.QueuedIntent)
	}
	if input.Plan != nil {
		planCopy := *input.Plan
		planCopy.Preconditions = slices.Clone(input.Plan.Preconditions)
		planCopy.InputDigests = cloneMap(input.Plan.InputDigests)
		planCopy.RequiredApprovals = slices.Clone(input.Plan.RequiredApprovals)
		planCopy.Impact = slices.Clone(input.Plan.Impact)
		planCopy.FieldIDs = cloneMap(input.Plan.FieldIDs)
		planCopy.OptionIDs = cloneMap(input.Plan.OptionIDs)
		planCopy.Effects = slices.Clone(input.Plan.Effects)
		for index := range planCopy.Effects {
			planCopy.Effects[index].BlockedBy = slices.Clone(input.Plan.Effects[index].BlockedBy)
			planCopy.Effects[index].Review = slices.Clone(input.Plan.Effects[index].Review)
		}
		result.Plan = &planCopy
	}
	return result
}
