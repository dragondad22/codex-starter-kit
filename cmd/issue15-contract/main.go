// Command issue15-contract runs the source-bound live proof for issue #15.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const (
	ownerID             = "305967668"
	repositoryOwner     = "codex-starter-kit-labs"
	repositoryName      = "codex-starter-kit-sandbox"
	repositoryID        = "R_kgDOTa0WSg"
	projectID           = "PVT_kwDOEjyyNM4Bdm9F"
	reconcilerActor     = "codex-starter-kit-labs-reconciler"
	seederActor         = "codex-starter-kit-labs-seeder"
	runMarker           = "starter-kit-contract:issue-15-20260719-01"
	parentManagedID     = "issue15:parent"
	selectedManagedID   = "issue15:selected"
	siblingManagedID    = "issue15:sibling"
	blockerManagedID    = "issue15:blocker"
	dependentManagedID  = "issue15:dependent"
	operatingProfile    = "issue-15-live-contract-v1"
	apiVersion          = "2026-03-10"
	restBaseURL         = "https://api.github.com"
	graphQLURL          = "https://api.github.com/graphql"
	fieldReadiness      = "PVTSSF_lADOEjyyNM4Bdm9FzhYHTZA"
	fieldStatus         = "PVTSSF_lADOEjyyNM4Bdm9FzhYHTIk"
	readinessIntake     = "8d6f41b6"
	readinessRefinement = "26a4c98a"
	readinessReady      = "2323ce77"
	readinessBlocked    = "983e3745"
	statusBacklog       = "f75ad846"
	statusNext          = "c9b40fc5"
	statusInProgress    = "47fc9ee4"
	statusDone          = "98236657"
)

type roleConfig struct {
	App                 githubadapter.AppInstallationConfig
	RequiredPermissions []string
}

type planningEvidence struct {
	SchemaVersion int                   `json:"schema_version"`
	Mandate       contractMandate       `json:"mandate"`
	Inspection    engine.WorkInspection `json:"inspection"`
	Plan          engine.WorkPlan       `json:"plan"`
}

type applyEvidence struct {
	SchemaVersion int                           `json:"schema_version"`
	Mandate       contractMandate               `json:"mandate"`
	Apply         engine.WorkApplyResult        `json:"apply"`
	Verification  engine.WorkVerificationResult `json:"verification"`
	Status        engine.ManagedTaskStatus      `json:"status"`
}

type contractMandate struct {
	SchemaVersion  int                 `json:"schema_version"`
	Digest         string              `json:"digest"`
	ApprovalID     string              `json:"approval_id"`
	ApprovedBy     string              `json:"approved_by"`
	ApprovedAt     time.Time           `json:"approved_at"`
	ExpiresAt      time.Time           `json:"expires_at"`
	SourceRevision string              `json:"source_revision"`
	WorkflowDigest string              `json:"workflow_digest"`
	ResourceDigest string              `json:"resource_digest"`
	Target         engine.WorkTarget   `json:"target"`
	Actors         []string            `json:"actors"`
	Permissions    map[string][]string `json:"permissions"`
	Effects        []string            `json:"effects"`
	Marker         string              `json:"marker"`
	DataClass      string              `json:"data_class"`
	CostCeiling    string              `json:"cost_ceiling"`
	Destructive    string              `json:"destructive"`
	Retention      string              `json:"retention"`
	CleanupWithin  string              `json:"cleanup_within"`
	RecoveryOwner  string              `json:"recovery_owner"`
}

func main() {
	flags := flag.NewFlagSet("issue15-contract", flag.ExitOnError)
	stage := flags.String("stage", "", "mandate, setup, project-setup, plan, apply, or cleanup")
	repository := flags.String("repository", ".", "local evidence repository")
	source := flags.String("source-revision", "", "exact reviewed Starter Kit commit")
	planPath := flags.String("plan", "", "planning evidence JSON for apply")
	fixturePath := flags.String("fixture", "", "exact setup fixture lease for project setup or cleanup")
	planID := flags.String("plan-id", "", "exact approved plan identity")
	approvalID := flags.String("approval-id", "", "durable owner approval record")
	approvedAt := flags.String("approved-at", "", "approval timestamp in RFC3339")
	expiresAt := flags.String("expires-at", "", "mandate expiry in RFC3339")
	workflowDigest := flags.String("workflow-digest", "", "SHA-256 of the installed reviewed workflow")
	workflowPath := flags.String("workflow", "docs/evidence/issue-15-contract.yml", "reviewed workflow file in the source checkout")
	mandateDigest := flags.String("mandate-digest", "", "exact approved mandate digest")
	flags.Parse(os.Args[1:])
	if flags.NArg() != 0 || *stage == "" {
		fatal("--stage is required and positional arguments are unsupported")
	}
	ctx := context.Background()
	var result any
	var err error
	switch *stage {
	case "mandate", "setup", "project-setup", "plan", "apply", "cleanup":
	default:
		fatal("unsupported stage %q", *stage)
	}
	if _, intentErr := contractIntent(*source); intentErr != nil {
		fatal("%v", intentErr)
	}
	if err := verifyReviewedSource(*source, *workflowPath, *workflowDigest); err != nil {
		fatal("%v", err)
	}
	mandate, mandateErr := bindContractMandate(*source, *approvalID, *approvedAt, *expiresAt, *workflowDigest)
	if mandateErr != nil {
		fatal("%v", mandateErr)
	}
	now := time.Now().UTC()
	if *stage != "mandate" && (*mandateDigest == "" || mandate.Digest != *mandateDigest || now.Before(mandate.ApprovedAt) || !now.Before(mandate.ExpiresAt)) {
		fatal("execution mandate is absent, mismatched, or expired")
	}
	lease := []fixtureEvidence{}
	if *stage == "project-setup" || *stage == "cleanup" {
		var fixture fixtureEvidence
		if *fixturePath == "" || readStrictJSON(*fixturePath, &fixture) != nil {
			fatal("--fixture must name the exact valid setup evidence for %s", *stage)
		}
		lease = append(lease, fixture)
	}
	switch *stage {
	case "mandate":
		result = mandate
	case "setup":
		result, err = runFixtureStage(ctx, "setup", "seeder", mandate)
	case "project-setup":
		result, err = runFixtureStage(ctx, "project-setup", "reconciler", mandate, lease...)
	case "cleanup":
		result, err = runFixtureStage(ctx, "cleanup", "seeder", mandate, lease...)
	case "plan":
		result, err = runPlan(ctx, *repository, *source, mandate)
	case "apply":
		if *planPath == "" || *planID == "" {
			fatal("--plan and --plan-id are required for apply")
		}
		result, err = runApply(ctx, *repository, *source, *planPath, *planID, mandate)
	}
	if err != nil {
		fatal("%v", err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fatal("encode evidence: %v", err)
	}
}

func runPlan(ctx context.Context, repository, source string, mandate contractMandate) (planningEvidence, error) {
	adapter, err := newWorkAdapter("reconciler")
	if err != nil {
		return planningEvidence{}, err
	}
	intent, err := contractIntent(source)
	if err != nil {
		return planningEvidence{}, err
	}
	lifecycle := engine.New(engine.WithWorkAdapter(adapter))
	inspection, err := lifecycle.InspectManagedTask(ctx, engine.ManagedTaskRequest{Repository: repository, Intent: intent})
	if err != nil {
		return planningEvidence{}, err
	}
	plan, err := lifecycle.PlanManagedTask(ctx, inspection)
	if err != nil {
		return planningEvidence{}, err
	}
	return planningEvidence{SchemaVersion: 1, Mandate: mandate, Inspection: inspection, Plan: plan}, nil
}

func runApply(ctx context.Context, repository, source, path, expectedPlanID string, mandate contractMandate) (applyEvidence, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return applyEvidence{}, fmt.Errorf("read planning evidence: %w", err)
	}
	var planning planningEvidence
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&planning); err != nil {
		return applyEvidence{}, fmt.Errorf("decode planning evidence: %w", err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		return applyEvidence{}, err
	}
	if planning.SchemaVersion != 1 || planning.Mandate.Digest != mandate.Digest || planning.Plan.ID != expectedPlanID || planning.Plan.SourceRevision != source || planning.Plan.Repository != filepath.Clean(root) {
		return applyEvidence{}, errors.New("planning evidence differs from the approved plan, source, or repository")
	}
	adapter, err := newWorkAdapter("reconciler")
	if err != nil {
		return applyEvidence{}, err
	}
	lifecycle := engine.New(engine.WithWorkAdapter(adapter))
	apply, err := lifecycle.ApplyManagedTask(ctx, expectedPlanID, planning.Plan)
	if err != nil {
		return applyEvidence{}, err
	}
	verification, err := lifecycle.VerifyManagedTask(ctx, repository)
	if err != nil {
		return applyEvidence{}, err
	}
	status, err := lifecycle.ManagedTaskStatus(ctx, repository)
	if err != nil {
		return applyEvidence{}, err
	}
	if verification.OverallState != engine.ControlPass || status.Disposition != "converged" {
		return applyEvidence{}, errors.New("managed-task live proof did not converge")
	}
	return applyEvidence{SchemaVersion: 1, Mandate: mandate, Apply: apply, Verification: verification, Status: status}, nil
}

func contractIntent(source string) (engine.WorkDesiredIntent, error) {
	if !isCommit(source) {
		return engine.WorkDesiredIntent{}, errors.New("--source-revision must be an exact 40- or 64-character hexadecimal commit")
	}
	target := workTarget()
	task := contractTask()
	return engine.WorkDesiredIntent{
		SchemaVersion: 1, OperationID: "issue-15-live-native-reconciliation-v1", SourceRevision: source,
		OperatingProfileRevision: operatingProfile, InputDigests: map[string]string{"fixture-spec": contractResourceDigest()},
		Credential: engine.WorkCredentialExpectation{Mode: "app-installation", Actor: reconcilerActor}, Target: target, Task: task,
	}, nil
}

func contractTask() engine.DesiredManagedTask {
	return engine.DesiredManagedTask{
		ManagedID: selectedManagedID, IssueType: "task", Title: "Contract fixture: selected",
		ParentManagedID: parentManagedID,
		Blockers:        []engine.WorkDependency{},
		Readiness:       "ready", Status: "next", Closed: true,
		Review: []engine.WorkReviewRequirement{{Role: "independent-reviewer", DistinctContext: true, QualifiedIndependent: true}},
		ParentContext: &engine.WorkParentContext{
			ManagedID: parentManagedID, Status: "backlog", Closed: false, CompletionSatisfied: false,
			OtherChildren: []engine.WorkRelatedTask{{ManagedID: siblingManagedID, Status: "backlog", Closed: false}},
		},
		Dependents: []engine.WorkDependentContext{{
			ManagedID: dependentManagedID, Readiness: "blocked", Status: "backlog", Closed: false, ReadyEligible: true,
			Blockers: []engine.WorkDependency{{ManagedID: selectedManagedID, Closed: false}, {ManagedID: blockerManagedID, Closed: true}},
		}},
	}
}

func bindContractMandate(source, approvalID, approvedAtValue, expiresAtValue, workflowDigest string) (contractMandate, error) {
	approvedAt, approvedErr := time.Parse(time.RFC3339, approvedAtValue)
	expiresAt, expiresErr := time.Parse(time.RFC3339, expiresAtValue)
	if !isCommit(source) || approvalID == "" || !isSHA256(workflowDigest) || approvedErr != nil || expiresErr != nil || expiresAt.Before(approvedAt) || expiresAt.Sub(approvedAt) > 72*time.Hour {
		return contractMandate{}, errors.New("mandate requires an approval record, exact source/workflow digest, and a valid bounded RFC3339 lease")
	}
	mandate := contractMandate{
		SchemaVersion: 1, ApprovalID: approvalID, ApprovedBy: "dragondad22", ApprovedAt: approvedAt.UTC(), ExpiresAt: expiresAt.UTC(),
		SourceRevision: source, WorkflowDigest: workflowDigest, ResourceDigest: contractResourceDigest(), Target: workTarget(),
		Actors: []string{reconcilerActor, seederActor}, Permissions: map[string][]string{
			"reconciler": {"actions:read", "checks:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"},
			"seeder":     {"contents:write", "issues:write", "metadata:read", "pull-requests:write", "workflows:write"},
		}, Effects: []string{"create-five-absent-marked-issues", "add-two-native-sub-issue-links", "add-two-native-dependencies", "set-ten-project-field-values", "apply-one-work-manager-plan", "remove-native-links-and-close-leased-fixtures"},
		Marker: runMarker, DataClass: "public-synthetic", CostCeiling: "zero-dollar", Destructive: "marker-scoped-fixture-cleanup-only", Retention: "30-day-raw-evidence", CleanupWithin: "24h", RecoveryOwner: "dragondad22",
	}
	mandate.Digest = digestJSON(mandate)
	return mandate, nil
}

func contractResourceDigest() string {
	resources := struct {
		Marker        string
		Target        engine.WorkTarget
		ManagedIDs    []string
		Tasks         map[string]engine.DesiredManagedTask
		IssueStates   map[string]string
		Relationships []string
		Baseline      map[string]string
	}{runMarker, workTarget(), fixtureOrder(), fixtureTasks(), fixtureIssueStates(), []string{parentManagedID + "->" + selectedManagedID, parentManagedID + "->" + siblingManagedID, selectedManagedID + "->" + dependentManagedID, blockerManagedID + "->" + dependentManagedID}, baselineStates()}
	return digestJSON(resources)
}

func digestJSON(value any) string {
	if mandate, ok := value.(contractMandate); ok {
		mandate.Digest = ""
		value = mandate
	}
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func readStrictJSON(path string, output any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("JSON input contains a trailing value")
	}
	return nil
}

func workTarget() engine.WorkTarget {
	return engine.WorkTarget{
		Host: "github.com", RepositoryID: repositoryID, ProjectID: projectID,
		FieldIDs: map[string]string{"readiness": fieldReadiness, "status": fieldStatus},
		OptionIDs: map[string]string{
			"readiness:intake": readinessIntake, "readiness:needs-refinement": readinessRefinement,
			"readiness:ready": readinessReady, "readiness:blocked": readinessBlocked,
			"status:backlog": statusBacklog, "status:next": statusNext,
			"status:in-progress": statusInProgress, "status:done": statusDone,
		},
	}
}

func newWorkAdapter(role string) (*githubadapter.Adapter, error) {
	configuration, err := roleConfiguration(role)
	if err != nil {
		return nil, err
	}
	provider, err := appProvider(configuration.App)
	if err != nil {
		return nil, err
	}
	target := workTarget()
	return githubadapter.New(githubadapter.Config{
		Host: "github.com", RESTBaseURL: restBaseURL, GraphQLURL: graphQLURL, APIVersion: apiVersion,
		Mode: "app-installation", Actor: configuration.App.Actor, ActorKind: "app", Account: repositoryOwner,
		InstallationID: configuration.App.InstallationID, RepositoryOwner: repositoryOwner, RepositoryName: repositoryName,
		RepositoryID: repositoryID, ProjectOwner: repositoryOwner, ProjectOwnerKind: "organization", ProjectID: projectID,
		FieldIDs: target.FieldIDs, OptionIDs: target.OptionIDs, RequiredPermissions: configuration.RequiredPermissions,
		MaxPages: 10, EvidenceMode: "live", LiveTargetApproved: true,
	}, provider, http.DefaultClient)
}

func roleConfiguration(role string) (roleConfig, error) {
	switch role {
	case "seeder":
		return roleConfig{App: appConfig("4319763", "147094309", seederActor), RequiredPermissions: []string{"contents:write", "issues:write", "metadata:read", "pull-requests:write", "workflows:write"}}, nil
	case "reconciler":
		return roleConfig{App: appConfig("4319725", "147093185", reconcilerActor), RequiredPermissions: []string{"actions:read", "checks:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"}}, nil
	default:
		return roleConfig{}, fmt.Errorf("unsupported contract role %q", role)
	}
}

func appConfig(appID, installationID, actor string) githubadapter.AppInstallationConfig {
	return githubadapter.AppInstallationConfig{RESTBaseURL: restBaseURL, APIVersion: apiVersion, AppID: appID, InstallationID: installationID, Actor: actor, Account: repositoryOwner, AccountID: ownerID}
}

func appProvider(config githubadapter.AppInstallationConfig) (*githubadapter.AppInstallationProvider, error) {
	privateKey := os.Getenv("CSK_APP_PRIVATE_KEY")
	if privateKey == "" {
		return nil, errors.New("CSK_APP_PRIVATE_KEY is unavailable")
	}
	return githubadapter.NewAppInstallationProvider(config, githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) {
		return []byte(privateKey), nil
	}), http.DefaultClient)
}

func managedBody(desired engine.DesiredManagedTask) (string, error) {
	encoded, err := json.Marshal(desired)
	if err != nil {
		return "", err
	}
	return "<!-- " + runMarker + " -->\n<!-- starter-kit-managed:" + desired.ManagedID + " -->\n<!-- starter-kit-managed-metadata:" + base64.RawURLEncoding.EncodeToString(encoded) + " -->", nil
}

func isCommit(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func isSHA256(value string) bool {
	return len(value) == 64 && isCommit(value)
}

func verifyReviewedSource(source, workflowPath, expectedWorkflowDigest string) error {
	command := exec.Command("git", "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil || strings.TrimSpace(string(output)) != source {
		return errors.New("executing checkout does not match the approved source revision")
	}
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		return errors.New("reviewed workflow is unavailable in the executing source")
	}
	digest := sha256.Sum256(content)
	if hex.EncodeToString(digest[:]) != expectedWorkflowDigest {
		return errors.New("executing workflow source does not match its approved digest")
	}
	return nil
}

func exactPermissions(observed, expected []string) bool {
	left := append([]string{}, observed...)
	right := append([]string{}, expected...)
	sort.Strings(left)
	sort.Strings(right)
	return strings.Join(left, "\x00") == strings.Join(right, "\x00")
}

func fatal(format string, values ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}
