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
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	Inspection    engine.WorkInspection `json:"inspection"`
	Plan          engine.WorkPlan       `json:"plan"`
}

type applyEvidence struct {
	SchemaVersion int                           `json:"schema_version"`
	Apply         engine.WorkApplyResult        `json:"apply"`
	Verification  engine.WorkVerificationResult `json:"verification"`
	Status        engine.ManagedTaskStatus      `json:"status"`
}

func main() {
	flags := flag.NewFlagSet("issue15-contract", flag.ExitOnError)
	stage := flags.String("stage", "", "setup, project-setup, plan, apply, or cleanup")
	repository := flags.String("repository", ".", "local evidence repository")
	source := flags.String("source-revision", "", "exact reviewed Starter Kit commit")
	planPath := flags.String("plan", "", "planning evidence JSON for apply")
	planID := flags.String("plan-id", "", "exact approved plan identity")
	flags.Parse(os.Args[1:])
	if flags.NArg() != 0 || *stage == "" {
		fatal("--stage is required and positional arguments are unsupported")
	}
	ctx := context.Background()
	var result any
	var err error
	switch *stage {
	case "setup", "project-setup", "plan", "apply", "cleanup":
	default:
		fatal("unsupported stage %q", *stage)
	}
	if *stage == "plan" || *stage == "apply" {
		if _, intentErr := contractIntent(*source); intentErr != nil {
			fatal("%v", intentErr)
		}
	}
	switch *stage {
	case "setup":
		result, err = runFixtureStage(ctx, "setup", "seeder")
	case "project-setup":
		result, err = runFixtureStage(ctx, "project-setup", "reconciler")
	case "cleanup":
		result, err = runFixtureStage(ctx, "cleanup", "seeder")
	case "plan":
		result, err = runPlan(ctx, *repository, *source)
	case "apply":
		if *planPath == "" || *planID == "" {
			fatal("--plan and --plan-id are required for apply")
		}
		result, err = runApply(ctx, *repository, *source, *planPath, *planID)
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

func runPlan(ctx context.Context, repository, source string) (planningEvidence, error) {
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
	return planningEvidence{SchemaVersion: 1, Inspection: inspection, Plan: plan}, nil
}

func runApply(ctx context.Context, repository, source, path, expectedPlanID string) (applyEvidence, error) {
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
	if planning.SchemaVersion != 1 || planning.Plan.ID != expectedPlanID || planning.Plan.SourceRevision != source || planning.Plan.Repository != filepath.Clean(root) {
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
	return applyEvidence{SchemaVersion: 1, Apply: apply, Verification: verification, Status: status}, nil
}

func contractIntent(source string) (engine.WorkDesiredIntent, error) {
	if !isCommit(source) {
		return engine.WorkDesiredIntent{}, errors.New("--source-revision must be an exact 40- or 64-character hexadecimal commit")
	}
	target := workTarget()
	task := engine.DesiredManagedTask{
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
	specDigest := sha256.Sum256([]byte(runMarker + ":native-relationship-contract-v1"))
	return engine.WorkDesiredIntent{
		SchemaVersion: 1, OperationID: "issue-15-live-native-reconciliation-v1", SourceRevision: source,
		OperatingProfileRevision: operatingProfile, InputDigests: map[string]string{"fixture-spec": hex.EncodeToString(specDigest[:])},
		Credential: engine.WorkCredentialExpectation{Mode: "app-installation", Actor: reconcilerActor}, Target: target, Task: task,
	}, nil
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
		return roleConfig{App: appConfig("4319763", "147094309", seederActor), RequiredPermissions: []string{"issues:write", "metadata:read"}}, nil
	case "reconciler":
		return roleConfig{App: appConfig("4319725", "147093185", reconcilerActor), RequiredPermissions: []string{"issues:write", "metadata:read", "organization-projects:write"}}, nil
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

func fatal(format string, values ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}
