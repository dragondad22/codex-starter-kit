// Command issue75-contract validates and emits the credential-free execution envelope
// for the approved issue #75 sandbox qualification. Its default mode never contacts
// GitHub. Explicit --execute-step mode uses separately injected, down-scoped GitHub App
// credentials for one exact inspect, plan, conditional apply, verify, and status cycle.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const (
	sandboxHost       = "github.com"
	sandboxRepository = "R_kgDOTa0WSg"
	sandboxProject    = "PVT_kwDOEjyyNM4Bdm9F"
	sandboxRESTID     = int64(1303189066)
	sandboxOwner      = "codex-starter-kit-labs"
	sandboxName       = "codex-starter-kit-sandbox"
	sandboxAccountID  = "305967668"
)

var (
	commitPattern  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	managedPattern = regexp.MustCompile(`^issue:[1-9][0-9]*$`)
	digestPattern  = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type workflowStep struct {
	Operation       string `json:"operation"`
	Authority       string `json:"authority"`
	Effectful       bool   `json:"effectful"`
	RequiredBinding string `json:"required_binding"`
}

type envelope struct {
	SchemaVersion            int                             `json:"schema_version"`
	EvidenceMode             string                          `json:"evidence_mode"`
	Outcome                  string                          `json:"outcome"`
	Target                   engine.WorkTarget               `json:"target"`
	ManagedID                string                          `json:"managed_id"`
	OperationID              string                          `json:"operation_id"`
	SourceRevision           string                          `json:"source_revision"`
	OperatingProfileRevision string                          `json:"operating_profile_revision"`
	BaseBranch               string                          `json:"base_branch"`
	HeadBranch               string                          `json:"head_branch"`
	RequiredChecks           []string                        `json:"required_checks"`
	Review                   engine.WorkReviewRequirement    `json:"review"`
	ProductApproval          engine.WorkReviewRequirement    `json:"product_approval,omitempty"`
	DeliveryResourceDigest   string                          `json:"delivery_resource_digest"`
	MandateID                string                          `json:"mandate_id"`
	ApprovedEffects          []string                        `json:"approved_effects"`
	Authorities              []engine.WorkExecutionAuthority `json:"authorities"`
	Workflow                 []workflowStep                  `json:"workflow"`
	Limitations              []string                        `json:"limitations"`
}

type transitionEvidence struct {
	SchemaVersion int                         `json:"schema_version"`
	EvidenceMode  string                      `json:"evidence_mode"`
	Outcome       string                      `json:"outcome"`
	Inspection    engine.DeliveryInspection   `json:"inspection"`
	Plan          engine.DeliveryPlan         `json:"plan"`
	Apply         *engine.DeliveryApplyResult `json:"apply,omitempty"`
	Verification  engine.DeliveryVerification `json:"verification"`
	Status        engine.DeliveryStatusResult `json:"status"`
}

type executionRuntime struct {
	restBaseURL string
	graphqlURL  string
	evidence    string
	live        bool
	now         func() time.Time
}

func main() {
	if err := runWithDependencies(context.Background(), os.Args[1:], os.Getenv, http.DefaultClient, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	return runWithDependencies(context.Background(), args, os.Getenv, http.DefaultClient, output)
}

func runWithDependencies(ctx context.Context, args []string, getenv func(string) string, client *http.Client, output io.Writer) error {
	flags := flag.NewFlagSet("issue75-contract", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	requestPath := flags.String("request-file", "", "credential-free DeliveryRequest JSON")
	mandatePath := flags.String("mandate-file", "", "content-addressed WorkExecutionMandate JSON")
	executeStep := flags.Bool("execute-step", false, "execute one exact lifecycle transition with scoped GitHub App credentials")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *requestPath == "" || *mandatePath == "" {
		return errors.New("request-file and mandate-file flags are required; positional arguments are unsupported")
	}
	var request engine.DeliveryRequest
	if err := readStrictJSON(*requestPath, &request); err != nil {
		return fmt.Errorf("delivery request: %w", err)
	}
	var mandate engine.WorkExecutionMandate
	if err := readStrictJSON(*mandatePath, &mandate); err != nil {
		return fmt.Errorf("execution mandate: %w", err)
	}
	if err := validateEnvelope(request, mandate); err != nil {
		return err
	}
	if *executeStep {
		runtime := executionRuntime{restBaseURL: "https://api.github.com", graphqlURL: "https://api.github.com/graphql", evidence: "live", live: true, now: func() time.Time { return time.Now().UTC() }}
		return executeTransition(ctx, request, mandate, getenv, client, runtime, output)
	}

	intent := request.Intent
	result := envelope{
		SchemaVersion: 1, EvidenceMode: "credential-free-plan", Outcome: "planned",
		Target: intent.Target, ManagedID: intent.ManagedID, OperationID: intent.OperationID,
		SourceRevision: intent.SourceRevision, OperatingProfileRevision: intent.OperatingProfileRevision,
		BaseBranch: intent.BaseBranch, HeadBranch: intent.HeadBranch,
		RequiredChecks: slices.Clone(intent.RequiredChecks), Review: intent.Review,
		ProductApproval: intent.ProductApproval, DeliveryResourceDigest: engine.DeliveryResourceDigest(intent),
		MandateID: mandate.ID, ApprovedEffects: slices.Clone(mandate.EffectKinds),
		Authorities: slices.Clone(mandate.Authorities),
		Workflow: []workflowStep{
			{Operation: "inspect", Authority: "configured read-only observer", RequiredBinding: "exact request and governed source revision"},
			{Operation: "plan", Authority: "lifecycle engine", RequiredBinding: "inspection identity and observation revision"},
			{Operation: "apply", Authority: "configured scoped effect actor", Effectful: true, RequiredBinding: "exact plan ID and mandate ID"},
			{Operation: "verify", Authority: "configured read-only observer", RequiredBinding: "exact postcondition observation"},
			{Operation: "status", Authority: "lifecycle engine", RequiredBinding: "integrity-protected retained state"},
		},
		Limitations: []string{
			"this command performs no GitHub observation or mutation",
			"credentials are supplied only to the separately configured adapters at execution time",
			"each changed observation requires a new inspect and plan cycle",
			"failed gates, stale identity, or authority mismatch remain explicit non-passes",
		},
	}
	slices.Sort(result.RequiredChecks)
	slices.Sort(result.ApprovedEffects)
	slices.SortFunc(result.Authorities, func(left, right engine.WorkExecutionAuthority) int {
		return strings.Compare(left.Actor+"\x00"+left.CredentialMode, right.Actor+"\x00"+right.CredentialMode)
	})
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func executeTransition(ctx context.Context, request engine.DeliveryRequest, mandate engine.WorkExecutionMandate, getenv func(string) string, client *http.Client, runtime executionRuntime, output io.Writer) error {
	if err := validateLiveTarget(request); err != nil {
		return err
	}
	now := runtime.now()
	if now.Before(mandate.ApprovedAt) || !now.Before(mandate.ExpiresAt) {
		return errors.New("execution mandate is not active")
	}
	reconcilerKey := getenv("CSK_RECONCILER_APP_PRIVATE_KEY")
	seederKey := getenv("CSK_SEEDER_APP_PRIVATE_KEY")
	if reconcilerKey == "" || seederKey == "" {
		return errors.New("approved GitHub App private keys are unavailable")
	}
	observer, effect, err := executionAdapters(client, runtime, reconcilerKey, seederKey)
	if err != nil {
		return err
	}
	delivery, err := githubadapter.NewDeliveryAdapter(observer, []githubadapter.DeliveryReviewerTrust{{Actor: "american-dragon-designs", Capable: true, DistinctContext: true}}, effect)
	if err != nil {
		return err
	}
	lifecycle := engine.New(engine.WithWorkAdapter(observer), engine.WithDeliveryAdapter(delivery))
	inspection, err := lifecycle.InspectDelivery(ctx, request)
	if err != nil {
		return err
	}
	plan, err := lifecycle.PlanDelivery(ctx, inspection)
	if err != nil {
		return err
	}
	var apply *engine.DeliveryApplyResult
	if len(plan.Effects) != 0 {
		result, err := lifecycle.ApplyDelivery(ctx, plan.ID, plan, mandate)
		if err != nil {
			return err
		}
		apply = &result
	}
	verification, err := lifecycle.VerifyDelivery(ctx, request.Repository)
	if err != nil {
		return err
	}
	status, err := lifecycle.DeliveryStatus(ctx, request.Repository)
	if err != nil {
		return err
	}
	result := transitionEvidence{SchemaVersion: 1, EvidenceMode: "live", Outcome: string(verification.OverallState), Inspection: inspection, Plan: plan, Apply: apply, Verification: verification, Status: status}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func executionAdapters(client *http.Client, runtime executionRuntime, reconcilerKey, seederKey string) (*githubadapter.Adapter, *githubadapter.Adapter, error) {
	provider := func(config githubadapter.AppInstallationConfig, key string) (*githubadapter.AppInstallationProvider, error) {
		return githubadapter.NewAppInstallationProvider(config, githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) { return []byte(key), nil }), client, githubadapter.WithAppCredentialClock(runtime.now))
	}
	reconcilerPermissions := map[string]string{"actions": "read", "checks": "read", "contents": "read", "issues": "write", "metadata": "read", "organization_projects": "write", "pull_requests": "read", "statuses": "read"}
	seederPermissions := map[string]string{"contents": "write", "metadata": "read", "pull_requests": "write"}
	reconcilerProvider, err := provider(appConfig(runtime.restBaseURL, "4319725", "147093185", "codex-starter-kit-labs-reconciler", reconcilerPermissions), reconcilerKey)
	if err != nil {
		return nil, nil, err
	}
	seederProvider, err := provider(appConfig(runtime.restBaseURL, "4319763", "147094309", "codex-starter-kit-labs-seeder", seederPermissions), seederKey)
	if err != nil {
		return nil, nil, err
	}
	observerConfig := adapterConfig(runtime, "codex-starter-kit-labs-reconciler", "147093185", []string{"actions:read", "checks:read", "contents:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"})
	effectConfig := adapterConfig(runtime, "codex-starter-kit-labs-seeder", "147094309", []string{"contents:write", "metadata:read", "pull-requests:write"})
	observer, err := githubadapter.New(observerConfig, reconcilerProvider, client, githubadapter.WithClock(runtime.now))
	if err != nil {
		return nil, nil, err
	}
	effect, err := githubadapter.New(effectConfig, seederProvider, client, githubadapter.WithClock(runtime.now))
	if err != nil {
		return nil, nil, err
	}
	return observer, effect, nil
}

func appConfig(rest, appID, installationID, actor string, permissions map[string]string) githubadapter.AppInstallationConfig {
	return githubadapter.AppInstallationConfig{RESTBaseURL: rest, APIVersion: "2026-03-10", AppID: appID, InstallationID: installationID, Actor: actor, Account: sandboxOwner, AccountID: sandboxAccountID, RepositoryIDs: []int64{sandboxRESTID}, TokenPermissions: permissions}
}

func adapterConfig(runtime executionRuntime, actor, installationID string, permissions []string) githubadapter.Config {
	return githubadapter.Config{Host: sandboxHost, RESTBaseURL: runtime.restBaseURL, GraphQLURL: runtime.graphqlURL, APIVersion: "2026-03-10", Mode: "app-installation", Actor: actor, ActorKind: "github-app", Account: sandboxOwner, InstallationID: installationID, RepositoryOwner: sandboxOwner, RepositoryName: sandboxName, RepositoryID: sandboxRepository, ProjectOwner: sandboxOwner, ProjectOwnerKind: "organization", ProjectID: sandboxProject, FieldIDs: liveFieldIDs(), OptionIDs: liveOptionIDs(), RequiredPermissions: permissions, MaxPages: 10, EvidenceMode: runtime.evidence, LiveTargetApproved: runtime.live, MutationInterval: time.Second}
}

func validateLiveTarget(request engine.DeliveryRequest) error {
	if !equalMap(request.Intent.Target.FieldIDs, liveFieldIDs()) || !equalMap(request.Intent.Target.OptionIDs, liveOptionIDs()) || request.Intent.ProductApproval.Role != "" || !equalMap(request.CompletionIntent.Target.FieldIDs, liveFieldIDs()) || !equalMap(request.CompletionIntent.Target.OptionIDs, liveOptionIDs()) {
		return errors.New("execution target fields, options, or approval policy differ from the approved live manifest")
	}
	return nil
}

func liveFieldIDs() map[string]string {
	return map[string]string{
		"status":    "PVTSSF_lADOEjyyNM4Bdm9FzhYHTIk",
		"readiness": "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZA",
		"horizon":   "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZE",
		"phase":     "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZI",
	}
}

func liveOptionIDs() map[string]string {
	return map[string]string{
		"status:backlog":             "f75ad846",
		"status:next":                "c9b40fc5",
		"status:in-progress":         "47fc9ee4",
		"status:done":                "98236657",
		"readiness:intake":           "8d6f41b6",
		"readiness:needs-refinement": "26a4c98a",
		"readiness:ready":            "2323ce77",
		"readiness:blocked":          "983e3745",
		"horizon:now":                "b1f7820f",
		"horizon:next":               "8920dc74",
		"horizon:later":              "965eb3dd",
		"phase:Phase 0":              "7fcb7c26",
		"phase:Phase 1":              "e6cbdc17",
		"phase:Phase 2":              "db48cb41",
		"phase:Phase 3":              "3a97d4af",
		"phase:Phase 4":              "e8eef021",
		"phase:Phase 5":              "358327da",
		"phase:Phase 6":              "e3063f78",
		"phase:Phase 7":              "3c19af01",
		"phase:Phase 8":              "865934cf",
	}
}

func readStrictJSON(path string, destination any) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("input is unavailable")
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("input is not valid canonical JSON")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("input contains trailing JSON")
	}
	return nil
}

func validateEnvelope(request engine.DeliveryRequest, mandate engine.WorkExecutionMandate) error {
	intent := request.Intent
	if intent.SchemaVersion != 1 || intent.Target.Host != sandboxHost || intent.Target.RepositoryID != sandboxRepository || intent.Target.ProjectID != sandboxProject || !managedPattern.MatchString(intent.ManagedID) || intent.Title == "" || intent.OperationID == "" || !commitPattern.MatchString(intent.SourceRevision) || intent.OperatingProfileRevision == "" {
		return errors.New("delivery request is not bound to the approved immutable issue #75 sandbox target")
	}
	if request.Repository == "" || intent.BaseBranch != "main" || !strings.HasPrefix(intent.HeadBranch, "contract/issue-75-") || intent.MergeMethod != "squash" || len(intent.RequiredChecks) == 0 || hasDuplicateOrEmpty(intent.RequiredChecks) {
		return errors.New("delivery request branch, checks, or squash policy is invalid")
	}
	if intent.Review.Role != "american-dragon-designs" || !intent.Review.DistinctContext || intent.Claim == nil || intent.Claim.ManagedID != intent.ManagedID || intent.Claim.SourceRevision != intent.SourceRevision || !digestPattern.MatchString(intent.Claim.ContractDigest) {
		return errors.New("delivery request lacks the approved distinct review or governed claim identity")
	}
	if _, err := engine.RenderWorkDeliveryClaim(*intent.Claim); err != nil {
		return errors.New("delivery request claim is invalid")
	}
	if intent.ProductApproval.Role != "" && intent.ProductApproval.Role == intent.Review.Role {
		return errors.New("product approval cannot reuse the distinct review identity")
	}
	if intent.EffectBoundary.DataClass != "public-synthetic" || intent.EffectBoundary.CostCeiling != "zero-dollar" || intent.EffectBoundary.Destructive == "" || intent.EffectBoundary.Retention == "" || intent.EffectBoundary.RecoveryOwner == "" {
		return errors.New("delivery request exceeds the approved sandbox operating boundary")
	}
	completion := request.CompletionIntent
	if completion == nil || completion.SchemaVersion != 2 || completion.OperationID != intent.OperationID || completion.SourceRevision != intent.SourceRevision || completion.OperatingProfileRevision != intent.OperatingProfileRevision || completion.Task.ManagedID != intent.ManagedID || completion.Task.Status != "done" || !completion.Task.Closed || completion.Governance == nil || !sameTarget(completion.Target, intent.Target) {
		return errors.New("delivery request lacks an exact governed completion intent")
	}
	if _, err := engine.RenderExecutableIssueContract(completion.Governance.Issue); err != nil || engine.ExecutableIssueContractDigest(completion.Governance.Issue) != intent.Claim.ContractDigest {
		return errors.New("completion governance does not match the delivery claim")
	}
	if containsSensitive(request) || containsSensitive(mandate) {
		return errors.New("credential-like material is forbidden in credential-free contract inputs")
	}
	if mandate.SchemaVersion != 1 || mandate.ID == "" || engine.BindWorkExecutionMandate(mandate).ID != mandate.ID || !sameTarget(mandate.Target, intent.Target) || mandate.OperationID != intent.OperationID || mandate.SelectedManagedID != intent.ManagedID || !slices.Contains(mandate.SourceRevisions, intent.SourceRevision) || !slices.Contains(mandate.OperatingProfileRevisions, intent.OperatingProfileRevision) || !slices.Contains(mandate.ManagedIDs, intent.ManagedID) || !slices.Contains(mandate.ResourceDigests, engine.DeliveryResourceDigest(intent)) {
		return errors.New("execution mandate does not exactly bind the delivery request")
	}
	governedSourceDigests := map[string]string{}
	for _, source := range completion.Governance.Sources {
		governedSourceDigests[source.ID] = source.Digest
	}
	approvedPermissions := []string{"actions:read", "checks:read", "contents:read", "contents:write", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "pull-requests:write", "statuses:read"}
	if !sameSet(mandate.Actors, []string{"codex-starter-kit-labs-reconciler", "codex-starter-kit-labs-seeder"}) || !sameSet(mandate.CredentialModes, []string{"app-installation"}) || !sameSet(mandate.Permissions, approvedPermissions) || !equalMap(mandate.InputDigests, completion.InputDigests) || !equalMap(mandate.GovernedSourceDigests, governedSourceDigests) {
		return errors.New("execution mandate authority or governed inputs exceed the approved envelope")
	}
	requiredEffects := []string{engine.DeliveryEffectCreateBranch, engine.DeliveryEffectCreatePullRequest, engine.DeliveryEffectMarkReady, engine.DeliveryEffectRequestReview, engine.DeliveryEffectSquashMerge, engine.DeliveryEffectReconcileCompletion, "reconcile-task"}
	if !sameSet(mandate.EffectKinds, requiredEffects) || !validOperations(mandate.Operations) || !slices.Contains(mandate.Operations, "closure") || !slices.Contains(mandate.ResourceDigests, engine.ManagedTaskResourceDigest(completion.Task)) || !slices.Contains(mandate.ContractDigests, intent.Claim.ContractDigest) || !slices.Contains(mandate.GovernanceDigests, engine.GovernedWorkContractDigest(*completion.Governance)) || mandate.MaxEffects < 5 || mandate.MaxEffects > 16 || mandate.DataClass != intent.EffectBoundary.DataClass || mandate.CostCeiling != intent.EffectBoundary.CostCeiling || mandate.Destructive != intent.EffectBoundary.Destructive || mandate.Retention != intent.EffectBoundary.Retention || mandate.RecoveryOwner != intent.EffectBoundary.RecoveryOwner || mandate.ApprovedBy == "" || mandate.ApprovalID == "" || mandate.ApprovedAt.IsZero() || !mandate.ExpiresAt.After(mandate.ApprovedAt) {
		return errors.New("execution mandate scope or operating ceiling is invalid")
	}
	if len(mandate.Authorities) != 2 || !slices.ContainsFunc(mandate.Authorities, approvedSeederAuthority) || !slices.ContainsFunc(mandate.Authorities, approvedReconcilerAuthority) {
		return errors.New("execution mandate lacks the approved separated effect and reconciliation authorities")
	}
	return nil
}

func hasDuplicateOrEmpty(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func sameSet(left, right []string) bool {
	left = slices.Clone(left)
	right = slices.Clone(right)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}

func validOperations(values []string) bool {
	if hasDuplicateOrEmpty(values) {
		return false
	}
	for _, value := range values {
		if !slices.Contains([]string{"issue", "project", "readiness", "status", "horizon", "phase", "closure", "context", "promotion-link"}, value) {
			return false
		}
	}
	return true
}

func approvedSeederAuthority(value engine.WorkExecutionAuthority) bool {
	return value.Actor == "codex-starter-kit-labs-seeder" && value.CredentialMode == "app-installation" && value.Account == "codex-starter-kit-labs" && value.InstallationID == "147094309" && value.RepositoryID == sandboxRepository && sameSet(value.Permissions, []string{"contents:write", "metadata:read", "pull-requests:write"})
}

func approvedReconcilerAuthority(value engine.WorkExecutionAuthority) bool {
	return value.Actor == "codex-starter-kit-labs-reconciler" && value.CredentialMode == "app-installation" && value.Account == "codex-starter-kit-labs" && value.InstallationID == "147093185" && value.RepositoryID == sandboxRepository && sameSet(value.Permissions, []string{"actions:read", "checks:read", "contents:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"})
}

func sameTarget(left, right engine.WorkTarget) bool {
	return left.Host == right.Host && left.RepositoryID == right.RepositoryID && left.ProjectID == right.ProjectID && equalMap(left.FieldIDs, right.FieldIDs) && equalMap(left.OptionIDs, right.OptionIDs)
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

func containsSensitive(value any) bool {
	encoded, _ := json.Marshal(value)
	lower := strings.ToLower(string(encoded))
	for _, marker := range []string{"github_pat_", "ghp_", "gho_", "ghs_", "bearer ", "private key", "access_token", "private_key"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
