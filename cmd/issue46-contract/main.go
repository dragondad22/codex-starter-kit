// Command issue46-contract plans and verifies the bounded operational Phase configuration.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const tokenEnvironment = "STARTER_KIT_ISSUE46_TOKEN"

type result struct {
	SchemaVersion int                               `json:"schema_version"`
	Planning      engine.SandboxPlanningResult      `json:"planning"`
	Mandate       *engine.SandboxExecutionMandate   `json:"mandate,omitempty"`
	Apply         *engine.SandboxApplyResult        `json:"apply,omitempty"`
	Verification  *engine.SandboxVerificationResult `json:"verification,omitempty"`
	ReplayPlan    *engine.SandboxPlan               `json:"replay_plan,omitempty"`
	ReplayApply   *engine.SandboxApplyResult        `json:"replay_apply,omitempty"`
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, arguments []string) error {
	flags := flag.NewFlagSet("issue46-contract", flag.ContinueOnError)
	stage := flags.String("stage", "", "plan or apply")
	repository := flags.String("repository", "", "local evidence-state repository")
	source := flags.String("source-revision", "", "reviewed source revision")
	expectedPlan := flags.String("expected-plan-id", "", "exact reviewed plan identity")
	approvalID := flags.String("approval-id", "", "durable owner approval record")
	approvedAtText := flags.String("approved-at", "", "owner approval time in RFC3339")
	expiresAtText := flags.String("expires-at", "", "mandate expiry in RFC3339")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if (*stage != "plan" && *stage != "apply") || *repository == "" || *source == "" || flags.NArg() != 0 {
		return errors.New("--stage plan|apply, --repository, and --source-revision are required; positional arguments are unsupported")
	}
	token := os.Getenv(tokenEnvironment)
	if token == "" {
		return fmt.Errorf("%s is required", tokenEnvironment)
	}
	now := time.Now().UTC()
	resources := phaseResources()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_kgDOTVs5Hg", ProjectID: "PVT_kwHOASd_cc4BdI9q", RepositoryName: "dragondad22/codex-starter-kit"}
	expectation := githubadapter.SandboxRoleExpectation{Mode: "user-token", Actor: "dragondad22", Account: "dragondad22", AccountID: "19365745", RequiredPermissions: []string{"projects:write"}}
	config := githubadapter.SandboxConfig{
		Host: "github.com", RESTBaseURL: "https://api.github.com", GraphQLURL: "https://api.github.com/graphql", APIVersion: "2026-03-10",
		ConfigurationRevision: "issue-46-phase-configuration-v1", Target: target, RepositoryOwner: "dragondad22", RepositoryName: "codex-starter-kit", ProjectNumber: 8, ProjectOwnerKind: "user",
		EvidenceMode: "live", LiveTargetApproved: true, Resources: resources,
		Roles: map[string]githubadapter.SandboxRoleExpectation{githubadapter.SandboxRoleReconciler: expectation},
	}
	provider := githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		return githubadapter.Credential{Token: token, Mode: "user-token", Actor: "dragondad22", Account: "dragondad22", AccountID: "19365745", Permissions: []string{"projects:write"}, ExpiresAt: now.Add(time.Hour)}, nil
	})
	adapter, err := githubadapter.NewSandbox(config, map[string]githubadapter.CredentialProvider{githubadapter.SandboxRoleReconciler: provider}, http.DefaultClient)
	if err != nil {
		return err
	}
	lifecycle := engine.New(engine.WithSandboxAdapter(adapter))
	authority := engine.SandboxAuthorityProfile{
		CredentialIdentities: []string{githubadapter.SandboxCredentialIdentity(githubadapter.SandboxRoleReconciler, expectation)},
		Permissions:          []string{"reconciler:projects:write"}, EvidenceMode: "live", Compatibility: "github.com:api.github.com:2026-03-10:native-rest-graphql",
		DataClass: "public-project-metadata", CostCeiling: "zero-dollar", Destructive: "no-delete-no-overwrite-human-view", Retention: "30-days",
	}
	manifest := engine.SandboxManifest{SchemaVersion: 1, OperationID: "issue-46-phase-configuration", SourceRevision: *source, ConfigurationRevision: config.ConfigurationRevision, ApprovedBy: "dragondad22", ApprovedPlan: "dec-0022:issue-46", RecoveryOwner: "dragondad22", MarkerPrefix: "starter-kit-contract:issue-46", Target: target, Authority: authority, Resources: resources}
	inspection, err := lifecycle.InspectSandbox(ctx, engine.SandboxRequest{Repository: *repository, Manifest: manifest})
	if err != nil {
		return err
	}
	plan, err := lifecycle.PlanSandbox(ctx, inspection)
	if err != nil {
		return err
	}
	output := result{SchemaVersion: 1, Planning: engine.SandboxPlanningResult{SchemaVersion: 1, Inspection: inspection, Plan: plan}}
	if *stage == "plan" {
		return json.NewEncoder(os.Stdout).Encode(output)
	}
	if *expectedPlan == "" || plan.ID != *expectedPlan || *approvalID == "" {
		return errors.New("apply requires the exact --expected-plan-id and --approval-id")
	}
	approvedAt, err := time.Parse(time.RFC3339, *approvedAtText)
	if err != nil {
		return errors.New("--approved-at must be RFC3339")
	}
	expiresAt, err := time.Parse(time.RFC3339, *expiresAtText)
	if err != nil {
		return errors.New("--expires-at must be RFC3339")
	}
	mandate := engine.BindSandboxExecutionMandate(engine.SandboxExecutionMandate{
		SchemaVersion: 1, ApprovedBy: "dragondad22", ApprovalID: *approvalID, ApprovedAt: approvedAt, ExpiresAt: expiresAt, Target: target,
		Actors: []string{githubadapter.SandboxRoleReconciler}, MarkerPrefix: manifest.MarkerPrefix, UnmarkedKeys: resourceKeys(resources),
		ResourceKinds: []string{engine.SandboxResourceProjectField, engine.SandboxResourceProjectOption, engine.SandboxResourceProjectView, engine.SandboxResourceProjectItemField}, EffectKinds: []string{"reconcile-resource"}, MaxEffects: len(resources),
		DataClass: authority.DataClass, CostCeiling: authority.CostCeiling, Destructive: authority.Destructive, Retention: authority.Retention, RecoveryOwner: manifest.RecoveryOwner, Authority: authority,
	}, resources...)
	approval := engine.SandboxPlanApproval{SchemaVersion: 2, Mandate: &mandate}
	apply, err := lifecycle.ApplySandbox(ctx, plan, approval)
	if err != nil {
		return err
	}
	verification, err := lifecycle.VerifySandbox(ctx, manifest)
	if err != nil {
		return err
	}
	replayInspection, err := lifecycle.InspectSandbox(ctx, engine.SandboxRequest{Repository: *repository, Manifest: manifest})
	if err != nil {
		return err
	}
	replayPlan, err := lifecycle.PlanSandbox(ctx, replayInspection)
	if err != nil {
		return err
	}
	replayApply, err := lifecycle.ApplySandbox(ctx, replayPlan, approval)
	if err != nil {
		return err
	}
	output.Mandate, output.Apply, output.Verification, output.ReplayPlan, output.ReplayApply = &mandate, &apply, &verification, &replayPlan, &replayApply
	return json.NewEncoder(os.Stdout).Encode(output)
}

func phaseResources() []engine.SandboxResourceSpec {
	resources := []engine.SandboxResourceSpec{{Key: "project-field:phase", Kind: engine.SandboxResourceProjectField, Name: "Phase", Attributes: map[string]string{"data_type": "single_select", "node_id": "PVTSSF_lAHOASd_cc4BdI9qzhYRk9k"}}}
	optionIDs := []string{"221d176d", "f817c01d", "8188d955", "6b779f39", "a7bbab56", "2880879a", "d4e86930", "85d21677", "6d252c8e"}
	contentIDs := []string{"I_kwDOTVs5Hs8AAAABIfFdIg", "I_kwDOTVs5Hs8AAAABIfFd4w", "I_kwDOTVs5Hs8AAAABIfFejQ", "I_kwDOTVs5Hs8AAAABIfFfSg", "I_kwDOTVs5Hs8AAAABIfFf8g", "I_kwDOTVs5Hs8AAAABIfFgrA", "I_kwDOTVs5Hs8AAAABIfFhTg", "I_kwDOTVs5Hs8AAAABIfFiKA", "I_kwDOTVs5Hs8AAAABIfFi4w"}
	itemIDs := []string{"PVTI_lAHOASd_cc4BdI9qzgyhGAM", "PVTI_lAHOASd_cc4BdI9qzgyhGAs", "PVTI_lAHOASd_cc4BdI9qzgyhGBM", "PVTI_lAHOASd_cc4BdI9qzgyhGCU", "PVTI_lAHOASd_cc4BdI9qzgyhGDs", "PVTI_lAHOASd_cc4BdI9qzgyhGEQ", "PVTI_lAHOASd_cc4BdI9qzgyhGE0", "PVTI_lAHOASd_cc4BdI9qzgyhGF0", "PVTI_lAHOASd_cc4BdI9qzgyhGGQ"}
	for index, optionID := range optionIDs {
		phase := fmt.Sprintf("Phase %d", index)
		resources = append(resources, engine.SandboxResourceSpec{Key: fmt.Sprintf("project-option:phase-%d", index), Kind: engine.SandboxResourceProjectOption, Name: phase, Attributes: map[string]string{"field": "Phase", "color": "GRAY", "description": "", "option_id": optionID, "input:id": optionID}})
	}
	resources = append(resources, engine.SandboxResourceSpec{Key: "project-view:phases", Kind: engine.SandboxResourceProjectView, Name: "Phases", Attributes: map[string]string{
		"layout": "table", "filter": "", "number": "6", "node_id": "PVTV_lAHOASd_cc4BdI9qzgLBdLU", "visible_fields": "PVTF_lAHOASd_cc4BdI9qzhXspNk,PVTF_lAHOASd_cc4BdI9qzhXspOI,PVTSSF_lAHOASd_cc4BdI9qzhXspNs,PVTSSF_lAHOASd_cc4BdI9qzhXspPQ", "group_by": "PVTSSF_lAHOASd_cc4BdI9qzhYRk9k", "sort_by": "PVTSSF_lAHOASd_cc4BdI9qzhYRk9k:asc", "input:visible_fields": "367830233,367830235,367830242,367830260,370250713",
	}})
	for index, optionID := range optionIDs {
		resources = append(resources, engine.SandboxResourceSpec{Key: fmt.Sprintf("project-item-field:feature-%d-phase", index+1), Kind: engine.SandboxResourceProjectItemField, Name: fmt.Sprintf("Feature #%d Phase", index+1), Attributes: map[string]string{"content_id": contentIDs[index], "item_id": itemIDs[index], "field": "Phase", "field_id": "PVTSSF_lAHOASd_cc4BdI9qzhYRk9k", "option_id": optionID}})
	}
	return resources
}

func resourceKeys(resources []engine.SandboxResourceSpec) []string {
	keys := make([]string, 0, len(resources))
	for _, resource := range resources {
		keys = append(keys, resource.Key)
	}
	return keys
}
