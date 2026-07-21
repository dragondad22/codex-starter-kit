// Command issue75-delivery emits the credential-free, content-addressed delivery
// request and execution mandate for the exact Issue #75 synthetic sandbox fixture.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

const (
	sandboxRepositoryID = "R_kgDOTa0WSg"
	sandboxProjectID    = "PVT_kwDOEjyyNM4Bdm9F"
	sandboxAccount      = "codex-starter-kit-labs"
	deliveryBranch      = "contract/issue-75-20260721-01"
	implementedPath     = ".github/workflows/issue-75-fixture-check.yml"
	requiredCheck       = "contract-delivery"
	reviewer            = "american-dragon-designs"
	operatingProfile    = "delegated-v1"
	operationID         = "issue-75-live-delivery-v1"
)

var (
	commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type issueIdentity struct {
	Number     int64  `json:"number"`
	DatabaseID int64  `json:"database_id"`
	NodeID     string `json:"node_id"`
}

type artifactContract struct {
	SchemaVersion       int                          `json:"schema_version"`
	EvidenceMode        string                       `json:"evidence_mode"`
	RequestPointer      string                       `json:"request_pointer"`
	MandatePointer      string                       `json:"mandate_pointer"`
	ExecutableIssueBody string                       `json:"executable_issue_body"`
	Parent              issueIdentity                `json:"parent"`
	Delivery            issueIdentity                `json:"delivery"`
	Dependent           issueIdentity                `json:"dependent"`
	ImplementedSource   engine.GovernedSourceBinding `json:"implemented_source"`
	HeadBranch          string                       `json:"head_branch"`
	RequiredCheck       string                       `json:"required_check"`
	Reviewer            string                       `json:"reviewer"`
	MergeMethod         string                       `json:"merge_method"`
	NativeVerification  []string                     `json:"native_verification"`
}

type outputEnvelope struct {
	Request          engine.DeliveryRequest      `json:"request"`
	Mandate          engine.WorkExecutionMandate `json:"mandate"`
	ArtifactContract artifactContract            `json:"artifact_contract"`
}

type options struct {
	sourceRevision    string
	implementedDigest string
	approvedBy        string
	approvalID        string
	approvedAt        time.Time
	expiresAt         time.Time
	parent            issueIdentity
	delivery          issueIdentity
	dependent         issueIdentity
}

func main() {
	if err := run(os.Args[1:], time.Now().UTC(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string, now time.Time, output io.Writer) error {
	flags := flag.NewFlagSet("issue75-delivery", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	sourceRevision := flags.String("source-revision", "", "exact reviewed Starter Kit source revision")
	implementedDigest := flags.String("implemented-source-digest", "", "sha256 digest of the final fixture workflow")
	approvedBy := flags.String("approved-by", "", "approving human identity")
	approvalID := flags.String("approval-id", "", "durable approval record identity")
	approvedAt := flags.String("approved-at", "", "approval time in RFC3339 format")
	expiresAt := flags.String("expires-at", "", "authority expiry in RFC3339 format")
	parentNumber := flags.String("parent-number", "", "exact parent issue number")
	parentID := flags.String("parent-id", "", "exact parent issue database ID")
	parentNodeID := flags.String("parent-node-id", "", "exact parent issue node ID")
	deliveryNumber := flags.String("delivery-number", "", "exact delivery issue number")
	deliveryID := flags.String("delivery-id", "", "exact delivery issue database ID")
	deliveryNodeID := flags.String("delivery-node-id", "", "exact delivery issue node ID")
	dependentNumber := flags.String("dependent-number", "", "exact dependent issue number")
	dependentID := flags.String("dependent-id", "", "exact dependent issue database ID")
	dependentNodeID := flags.String("dependent-node-id", "", "exact dependent issue node ID")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("valid named flags are required; positional arguments are unsupported")
	}
	approved, expires, err := validateApproval(*approvedBy, *approvalID, *approvedAt, *expiresAt, now)
	if err != nil {
		return err
	}
	parent, err := parseIssueIdentity(*parentNumber, *parentID, *parentNodeID)
	if err != nil {
		return fmt.Errorf("parent identity: %w", err)
	}
	delivery, err := parseIssueIdentity(*deliveryNumber, *deliveryID, *deliveryNodeID)
	if err != nil {
		return fmt.Errorf("delivery identity: %w", err)
	}
	dependent, err := parseIssueIdentity(*dependentNumber, *dependentID, *dependentNodeID)
	if err != nil {
		return fmt.Errorf("dependent identity: %w", err)
	}
	value := options{
		sourceRevision: *sourceRevision, implementedDigest: *implementedDigest,
		approvedBy: *approvedBy, approvalID: *approvalID, approvedAt: approved, expiresAt: expires,
		parent: parent, delivery: delivery, dependent: dependent,
	}
	result, err := build(value)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func parseIssueIdentity(number, databaseID, nodeID string) (issueIdentity, error) {
	parsedNumber, numberErr := parsePositiveDecimal(number)
	parsedDatabaseID, databaseErr := parsePositiveDecimal(databaseID)
	if numberErr != nil || databaseErr != nil || strings.TrimSpace(nodeID) == "" || nodeID != strings.TrimSpace(nodeID) {
		return issueIdentity{}, errors.New("number and database ID must be canonical positive decimals and node ID must be non-empty")
	}
	return issueIdentity{Number: parsedNumber, DatabaseID: parsedDatabaseID, NodeID: nodeID}, nil
}

func parsePositiveDecimal(value string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 || strconv.FormatInt(parsed, 10) != value {
		return 0, errors.New("not a canonical positive decimal")
	}
	return parsed, nil
}

func validateApproval(approvedBy, approvalID, approvedAt, expiresAt string, now time.Time) (time.Time, time.Time, error) {
	if strings.TrimSpace(approvedBy) == "" || approvedBy != strings.TrimSpace(approvedBy) || strings.TrimSpace(approvalID) == "" || approvalID != strings.TrimSpace(approvalID) {
		return time.Time{}, time.Time{}, errors.New("approved-by and approval-id are required without surrounding whitespace")
	}
	approved, err := time.Parse(time.RFC3339, approvedAt)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("approved-at must be RFC3339")
	}
	expires, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("expires-at must be RFC3339")
	}
	if now.Before(approved) || !now.Before(expires) || !approved.Before(expires) {
		return time.Time{}, time.Time{}, errors.New("approval must be active at generation time")
	}
	return approved, expires, nil
}

func build(value options) (outputEnvelope, error) {
	if !commitPattern.MatchString(value.sourceRevision) {
		return outputEnvelope{}, errors.New("source-revision must be an exact lowercase 40-character commit")
	}
	if !digestPattern.MatchString(value.implementedDigest) {
		return outputEnvelope{}, errors.New("implemented-source-digest must be canonical sha256:hex")
	}
	if !distinctIssues(value.parent, value.delivery, value.dependent) {
		return outputEnvelope{}, errors.New("parent, delivery, and dependent issue identities must be pairwise distinct")
	}
	if containsCredentialMarker(value.approvedBy, value.approvalID, value.parent.NodeID, value.delivery.NodeID, value.dependent.NodeID) {
		return outputEnvelope{}, errors.New("credential-like material is forbidden in credential-free inputs")
	}

	target := liveTarget()
	managedID := fmt.Sprintf("issue:%d", value.delivery.Number)
	parentManagedID := fmt.Sprintf("issue:%d", value.parent.Number)
	dependentManagedID := fmt.Sprintf("issue:%d", value.dependent.Number)
	issue := engine.ExecutableIssueContract{
		SchemaVersion:       1,
		Parent:              fmt.Sprintf("#%d", value.parent.Number),
		HumanSummary:        "Exercise one synthetic task through the governed delivery lifecycle and reconcile completion.",
		CurrentContext:      "The approved Issue #75 sandbox fixture supplies exact issue, relationship, Project, branch, check, review, and merge evidence.",
		GoverningReferences: "- fixture-check — exact implemented workflow used by the synthetic required check",
		Scope:               "Deliver only the marker-owned Issue #75 public-synthetic sandbox fixture.",
		OutOfScope:          "Production repositories, paid effects, unrelated sandbox resources, and human-implementation verification.",
		Acceptance:          "- [ ] The exact head passes contract-delivery.\n- [ ] A distinct capable reviewer approves the exact head.\n- [ ] The pull request is squash merged and the governed issue, sole-child parent, and final unblocked dependent are reconciled from native evidence.",
		Verification:        "Retain inspect, plan, apply, verify, status, exact GitHub identity, check, review, rules, merge, and Project reconciliation evidence for 30 days.",
		Dependencies:        "Use only the approved role-separated GitHub App authorities and the exact Issue #75 execution mandate.",
		ReadinessAssertions: []string{
			"No unresolved product, architecture, policy, regulatory, or risk decision is hidden in this task.",
			"An authorized implementer can execute this without the originating conversation.",
		},
	}
	issueBody, err := engine.RenderExecutableIssueContract(issue)
	if err != nil {
		return outputEnvelope{}, fmt.Errorf("synthetic executable issue contract: %w", err)
	}
	contractDigest := engine.ExecutableIssueContractDigest(issue)
	implemented := engine.GovernedSourceBinding{ID: "fixture-check", Path: implementedPath, Digest: value.implementedDigest}
	governance := &engine.GovernedWorkContract{SchemaVersion: 1, Issue: issue, Sources: []engine.GovernedSourceBinding{implemented}}
	boundary := engine.WorkEffectBoundary{
		DataClass: "public-synthetic", CostCeiling: "zero-dollar", Destructive: "issue-75-sandbox-delivery-only",
		Retention: "30-day-raw-evidence", RecoveryOwner: value.approvedBy,
	}
	claim := &engine.WorkDeliveryClaim{
		SchemaVersion: 1, ManagedID: managedID, SourceRevision: value.sourceRevision,
		ContractDigest: contractDigest, ImplementedSources: []engine.GovernedSourceBinding{implemented},
	}
	if _, err := engine.RenderWorkDeliveryClaim(*claim); err != nil {
		return outputEnvelope{}, fmt.Errorf("delivery claim: %w", err)
	}
	intent := engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: operationID, SourceRevision: value.sourceRevision,
		OperatingProfileRevision: operatingProfile, ManagedID: managedID,
		Title: "Issue 75 contract fixture: governed delivery", Target: target, BaseBranch: "main", HeadBranch: deliveryBranch,
		RequiredChecks: []string{requiredCheck},
		Review:         engine.WorkReviewRequirement{Role: reviewer, DistinctContext: true},
		MergeMethod:    "squash", Claim: claim, EffectBoundary: boundary,
	}
	completion := engine.WorkDesiredIntent{
		SchemaVersion: 2, OperationID: operationID, SourceRevision: value.sourceRevision,
		OperatingProfileRevision: operatingProfile, InputDigests: map[string]string{"issue": contractDigest},
		Credential: engine.WorkCredentialExpectation{Mode: "app-installation", Actor: "codex-starter-kit-labs-reconciler"},
		Target:     target,
		Task: engine.DesiredManagedTask{
			ManagedID: managedID, IssueType: "task", Title: intent.Title, ParentManagedID: parentManagedID,
			Readiness: "ready", Status: "done", Closed: true,
			Review:        []engine.WorkReviewRequirement{{Role: reviewer, DistinctContext: true}},
			ParentContext: &engine.WorkParentContext{ManagedID: parentManagedID, Status: "in-progress", CompletionSatisfied: true, OtherChildren: []engine.WorkRelatedTask{}},
			Dependents: []engine.WorkDependentContext{{
				ManagedID: dependentManagedID, Readiness: "blocked", Status: "backlog", ReadyEligible: true,
				Blockers: []engine.WorkDependency{{ManagedID: managedID, Closed: true}},
			}},
		},
		Governance: governance, EffectBoundary: boundary,
	}
	request := engine.DeliveryRequest{Repository: ".", Intent: intent, CompletionIntent: &completion}
	authorities := approvedAuthorities()
	permissions := append(slices.Clone(authorities[0].Permissions), authorities[1].Permissions...)
	governedSourceDigests := map[string]string{implemented.ID: implemented.Digest}
	mandate := engine.BindWorkExecutionMandate(engine.WorkExecutionMandate{
		SchemaVersion: 1, ApprovedBy: value.approvedBy, ApprovalID: value.approvalID,
		ApprovedAt: value.approvedAt, ExpiresAt: value.expiresAt,
		Target: target, OperationID: operationID, SelectedManagedID: managedID,
		Actors:          []string{"codex-starter-kit-labs-reconciler", "codex-starter-kit-labs-seeder"},
		CredentialModes: []string{"app-installation"}, Permissions: permissions, Authorities: authorities,
		OperatingProfileRevisions: []string{operatingProfile}, ContractDigests: []string{contractDigest},
		GovernanceDigests: []string{engine.GovernedWorkContractDigest(*governance)}, InputDigests: completion.InputDigests,
		GovernedSourceDigests: governedSourceDigests, SourceRevisions: []string{value.sourceRevision}, ManagedIDs: []string{managedID},
		EffectKinds: []string{
			engine.DeliveryEffectCreateBranch, engine.DeliveryEffectCreatePullRequest, engine.DeliveryEffectMarkReady,
			engine.DeliveryEffectRequestReview, engine.DeliveryEffectSquashMerge, engine.DeliveryEffectReconcileCompletion, "reconcile-task",
		},
		Operations:      []string{"closure", "context", "issue", "project", "readiness", "status"},
		ResourceDigests: []string{engine.DeliveryResourceDigest(intent), engine.ManagedTaskResourceDigest(completion.Task)}, MaxEffects: 8,
		DataClass: boundary.DataClass, CostCeiling: boundary.CostCeiling, Destructive: boundary.Destructive,
		Retention: boundary.Retention, RecoveryOwner: boundary.RecoveryOwner,
	})
	artifact := artifactContract{
		SchemaVersion: 1, EvidenceMode: "credential-free-input", RequestPointer: "/request", MandatePointer: "/mandate", ExecutableIssueBody: issueBody,
		Parent: value.parent, Delivery: value.delivery, Dependent: value.dependent, ImplementedSource: implemented,
		HeadBranch: deliveryBranch, RequiredCheck: requiredCheck, Reviewer: reviewer, MergeMethod: "squash",
		NativeVerification: []string{
			"observe exact parent, delivery, and dependent issue number, database ID, and node ID",
			"observe the delivery issue as Ready, Done, and closed after squash merge",
			"observe the sole-child parent as Done and closed",
			"observe the final unblocked dependent promoted from Blocked/Backlog to Ready",
		},
	}
	return outputEnvelope{Request: request, Mandate: mandate, ArtifactContract: artifact}, nil
}

func containsCredentialMarker(values ...string) bool {
	joined := strings.ToLower(strings.Join(values, "\n"))
	for _, marker := range []string{"github_pat_", "ghp_", "gho_", "ghs_", "bearer ", "private key", "access_token", "private_key"} {
		if strings.Contains(joined, marker) {
			return true
		}
	}
	return false
}

func distinctIssues(values ...issueIdentity) bool {
	numbers, databaseIDs, nodeIDs := map[int64]bool{}, map[int64]bool{}, map[string]bool{}
	for _, value := range values {
		if numbers[value.Number] || databaseIDs[value.DatabaseID] || nodeIDs[value.NodeID] {
			return false
		}
		numbers[value.Number], databaseIDs[value.DatabaseID], nodeIDs[value.NodeID] = true, true, true
	}
	return true
}

func approvedAuthorities() []engine.WorkExecutionAuthority {
	return []engine.WorkExecutionAuthority{
		{Actor: "codex-starter-kit-labs-seeder", CredentialMode: "app-installation", Account: sandboxAccount, InstallationID: "147094309", RepositoryID: sandboxRepositoryID, Permissions: []string{"contents:write", "metadata:read", "pull-requests:write"}},
		{Actor: "codex-starter-kit-labs-reconciler", CredentialMode: "app-installation", Account: sandboxAccount, InstallationID: "147093185", RepositoryID: sandboxRepositoryID, Permissions: []string{"actions:read", "checks:read", "contents:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"}},
	}
}

func liveTarget() engine.WorkTarget {
	return engine.WorkTarget{
		Host: "github.com", RepositoryID: sandboxRepositoryID, ProjectID: sandboxProjectID,
		FieldIDs: map[string]string{
			"status": "PVTSSF_lADOEjyyNM4Bdm9FzhYHTIk", "readiness": "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZA",
			"horizon": "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZE", "phase": "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZI",
		},
		OptionIDs: map[string]string{
			"status:backlog": "f75ad846", "status:next": "c9b40fc5", "status:in-progress": "47fc9ee4", "status:done": "98236657",
			"readiness:intake": "8d6f41b6", "readiness:needs-refinement": "26a4c98a", "readiness:ready": "2323ce77", "readiness:blocked": "983e3745",
			"horizon:now": "b1f7820f", "horizon:next": "8920dc74", "horizon:later": "965eb3dd",
			"phase:Phase 0": "7fcb7c26", "phase:Phase 1": "e6cbdc17", "phase:Phase 2": "db48cb41", "phase:Phase 3": "3a97d4af", "phase:Phase 4": "e8eef021",
			"phase:Phase 5": "358327da", "phase:Phase 6": "e3063f78", "phase:Phase 7": "3c19af01", "phase:Phase 8": "865934cf",
		},
	}
}
