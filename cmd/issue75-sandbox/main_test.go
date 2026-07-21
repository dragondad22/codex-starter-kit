package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

var fixedNow = time.Date(2026, 7, 21, 18, 0, 0, 0, time.UTC)

func TestStagesEmitExactRoleScopedSandboxInputs(t *testing.T) {
	tests := []struct {
		stage       string
		role        string
		kind        string
		count       int
		permissions []string
		cleanup     bool
	}{
		{"issues-setup", githubadapter.SandboxRoleSeeder, engine.SandboxResourceFixtureIssue, 3, []string{"issues:write", "metadata:read"}, false},
		{"issues-governed", githubadapter.SandboxRoleSeeder, engine.SandboxResourceFixtureIssue, 3, []string{"issues:write", "metadata:read"}, false},
		{"project-setup", githubadapter.SandboxRoleReconciler, engine.SandboxResourceProjectItemField, 6, []string{"metadata:read", "organization-projects:write"}, false},
		{"relationships-setup", githubadapter.SandboxRoleReconciler, engine.SandboxResourceIssueRelationship, 2, []string{"issues:write", "metadata:read"}, false},
		{"rules-setup", githubadapter.SandboxRoleRules, engine.SandboxResourceRuleset, 1, []string{"administration:write", "metadata:read"}, false},
		{"file-initial", githubadapter.SandboxRoleSeeder, engine.SandboxResourceRepositoryFile, 1, []string{"contents:write", "metadata:read"}, false},
		{"file-stale", githubadapter.SandboxRoleSeeder, engine.SandboxResourceRepositoryFile, 1, []string{"contents:write", "metadata:read"}, false},
		{"cleanup-relationships", githubadapter.SandboxRoleReconciler, engine.SandboxResourceIssueRelationship, 2, []string{"issues:write", "metadata:read"}, true},
		{"cleanup-rules", githubadapter.SandboxRoleRules, engine.SandboxResourceRuleset, 1, []string{"administration:write", "metadata:read"}, true},
		{"cleanup-file", githubadapter.SandboxRoleSeeder, engine.SandboxResourceRepositoryFile, 1, []string{"contents:write", "metadata:read"}, true},
		{"cleanup-delivery", githubadapter.SandboxRoleSeeder, "", 2, []string{"contents:write", "metadata:read", "pull-requests:write"}, true},
		{"cleanup-issues", githubadapter.SandboxRoleSeeder, engine.SandboxResourceFixtureIssue, 3, []string{"issues:write", "metadata:read"}, true},
	}
	for _, test := range tests {
		t.Run(test.stage, func(t *testing.T) {
			input := mustBuild(t, test.stage)
			if input.Role != test.role || len(input.Request.Manifest.Resources) != test.count {
				t.Fatalf("role/resources = %q/%#v", input.Role, input.Request.Manifest.Resources)
			}
			expectation := input.Config.Roles[test.role]
			if !slices.Equal(expectation.RequiredPermissions, test.permissions) || len(input.Config.Roles) != 1 {
				t.Fatalf("permissions = %#v", input.Config.Roles)
			}
			if len(input.App.RepositoryIDs) != 1 || input.App.RepositoryIDs[0] != sandboxRESTID || input.App.AccountID != sandboxOwnerID {
				t.Fatalf("app is not exact repository scoped: %#v", input.App)
			}
			if len(input.App.TokenPermissions) != len(test.permissions) {
				t.Fatalf("token permissions are broader than the role: %#v", input.App.TokenPermissions)
			}
			for _, permission := range test.permissions {
				parts := strings.Split(permission, ":")
				if input.App.TokenPermissions[strings.ReplaceAll(parts[0], "-", "_")] != parts[1] {
					t.Fatalf("token permissions do not match %q: %#v", permission, input.App.TokenPermissions)
				}
			}
			for _, resource := range input.Request.Manifest.Resources {
				if test.kind != "" && resource.Kind != test.kind || (resource.DesiredState == engine.SandboxResourceAbsent) != test.cleanup {
					t.Fatalf("resource = %#v", resource)
				}
			}
			expectedEffect := []string{"reconcile-resource"}
			expectedDestructive := "no-delete"
			if test.cleanup {
				expectedEffect = []string{"remove-resource"}
				expectedDestructive = "marker-scoped-fixture-cleanup-only"
			}
			if input.Mandate.ID == "" || len(input.Mandate.ResourceDigests) != test.count || input.Mandate.MaxEffects != test.count || !slices.Equal(input.Mandate.EffectKinds, expectedEffect) || input.Mandate.Destructive != expectedDestructive || !slices.Equal(input.Mandate.Actors, []string{expectation.Actor}) {
				t.Fatalf("mandate = %#v", input.Mandate)
			}
			validateManifest(t, input)
		})
	}
}

func TestGovernedIssueStageInstallsManagedBodiesAndExactDeliveryContract(t *testing.T) {
	resources := mustBuild(t, "issues-governed").Request.Manifest.Resources
	if len(resources) != 3 {
		t.Fatalf("governed resources = %#v", resources)
	}
	for index, managedID := range []string{"issue:11", "issue:12", "issue:13"} {
		body := resources[index].Attributes["input:body"]
		if !strings.Contains(body, "<!-- starter-kit-managed:"+managedID+" -->") || !strings.Contains(body, resources[index].Marker) || resources[index].Attributes["body_sha256"] != contentDigest(body) {
			t.Fatalf("governed issue %d = %#v", index, resources[index])
		}
		if _, err := engine.ParseExecutableIssueContract(body); err != nil {
			t.Fatalf("governed issue contract %d: %v", index, err)
		}
	}
}

func TestProjectSetupSelectsReadyWorkAndBlockedDependent(t *testing.T) {
	resources := mustBuild(t, "project-setup").Request.Manifest.Resources
	if len(resources) != 6 {
		t.Fatalf("project resources = %#v", resources)
	}
	want := []struct{ content, field, option string }{
		{"I_parent", statusFieldID, statusInProgressID}, {"I_parent", readinessFieldID, readinessReadyID},
		{"I_delivery", statusFieldID, statusInProgressID}, {"I_delivery", readinessFieldID, readinessReadyID},
		{"I_dependent", statusFieldID, statusBacklogID}, {"I_dependent", readinessFieldID, readinessBlockedID},
	}
	for index, expected := range want {
		attributes := resources[index].Attributes
		if attributes["content_id"] != expected.content || attributes["field_id"] != expected.field || attributes["option_id"] != expected.option {
			t.Fatalf("project resource %d = %#v", index, resources[index])
		}
	}
}

func TestStageContractDeclaresIdentityHandoffAndDeliveryCleanup(t *testing.T) {
	setup := mustBuild(t, "issues-setup")
	if len(setup.StageContract.IdentityRequirements) != 0 || !slices.Equal(setup.StageContract.IdentityOutputs, []string{"parent_number", "parent_id", "parent_node_id", "delivery_number", "delivery_id", "delivery_node_id", "dependent_number", "dependent_id", "dependent_node_id"}) {
		t.Fatalf("issue identity outputs = %#v", setup.StageContract)
	}
	for _, stage := range []string{"relationships-setup", "cleanup-relationships", "cleanup-issues"} {
		contract := mustBuild(t, stage).StageContract
		if !slices.Equal(contract.IdentityRequirements, setup.StageContract.IdentityOutputs) {
			t.Fatalf("%s identity requirements = %#v", stage, contract)
		}
	}
	cleanup := mustBuild(t, "cleanup-delivery")
	if !slices.Equal(cleanup.StageContract.IdentityRequirements, []string{"delivery_number", "pull_number", "pull_id", "pull_node_id", "branch_head_sha"}) {
		t.Fatalf("delivery cleanup contract = %#v", cleanup.StageContract)
	}
	resources := cleanup.Request.Manifest.Resources
	if len(resources) != 2 || resources[0].Kind != engine.SandboxResourceFixturePR || resources[1].Kind != engine.SandboxResourceFixtureBranch {
		t.Fatalf("delivery cleanup ordering = %#v", resources)
	}
	if resources[0].Marker != "Closes #12" || resources[0].Attributes["number"] != "17" || resources[0].Attributes["id"] != "117" || resources[0].Attributes["node_id"] != "PR_delivery" || resources[0].Attributes["head_sha"] != strings.Repeat("b", 40) || resources[0].Attributes["head"] != deliveryHeadBranch || resources[1].Attributes["sha"] != strings.Repeat("b", 40) || resources[1].Name != deliveryHeadBranch {
		t.Fatalf("delivery cleanup identities = %#v", resources)
	}
}

func TestIssueFixturesAndRelationshipsCarryExactOrganicTopology(t *testing.T) {
	setup := mustBuild(t, "issues-setup")
	if got := setup.Request.Manifest.Resources; len(got) != 3 || got[0].Name != "parent" || got[1].Name != "delivery" || got[2].Name != "dependent" {
		t.Fatalf("fixture issues = %#v", got)
	}
	relationships := mustBuild(t, "relationships-setup").Request.Manifest.Resources
	parent := relationships[0]
	blocked := relationships[1]
	if parent.Attributes["relationship"] != "parent-sub-issue" || parent.Attributes["source_number"] != "11" || parent.Attributes["target_number"] != "12" {
		t.Fatalf("parent relationship = %#v", parent)
	}
	if blocked.Attributes["relationship"] != "blocker-dependent" || blocked.Attributes["source_number"] != "12" || blocked.Attributes["target_number"] != "13" {
		t.Fatalf("blocker relationship = %#v", blocked)
	}
	cleanup := mustBuild(t, "cleanup-issues").Request.Manifest.Resources
	for index, identity := range []issueIdentity{{"11", "101", "I_parent"}, {"12", "102", "I_delivery"}, {"13", "103", "I_dependent"}} {
		resource := cleanup[index]
		if resource.DesiredState != engine.SandboxResourceAbsent || resource.Attributes["state"] != "closed" || resource.Attributes["number"] != identity.Number || resource.Attributes["id"] != identity.ID || resource.Attributes["node_id"] != identity.NodeID {
			t.Fatalf("cleanup issue = %#v", resource)
		}
	}
}

func TestWorkflowStagesBindChangedHeadContentAndExactFinalCleanup(t *testing.T) {
	initial := mustBuild(t, "file-initial").Request.Manifest.Resources[0]
	stale := mustBuild(t, "file-stale").Request.Manifest.Resources[0]
	cleanup := mustBuild(t, "cleanup-file").Request.Manifest.Resources[0]
	if initial.Attributes["branch"] != "main" || stale.Attributes["branch"] != deliveryHeadBranch || cleanup.Attributes["branch"] != "main" {
		t.Fatalf("workflow branches = %q/%q/%q", initial.Attributes["branch"], stale.Attributes["branch"], cleanup.Attributes["branch"])
	}
	if initial.Attributes["path"] != ".github/workflows/issue-75-fixture-check.yml" {
		t.Fatalf("fixture check would overwrite a control workflow: %#v", initial.Attributes)
	}
	if initial.Attributes["input:content"] == stale.Attributes["input:content"] || initial.Attributes["content_sha256"] == stale.Attributes["content_sha256"] {
		t.Fatal("stale stage must create a new approved head revision")
	}
	if cleanup.Attributes["input:content"] != stale.Attributes["input:content"] || cleanup.Attributes["content_sha256"] != stale.Attributes["content_sha256"] || cleanup.DesiredState != engine.SandboxResourceAbsent {
		t.Fatalf("cleanup is not bound to final approved content: %#v", cleanup)
	}
	for _, content := range []string{initial.Attributes["input:content"], stale.Attributes["input:content"]} {
		if !strings.Contains(content, runMarker) || !strings.Contains(content, "pull_request:") || !strings.Contains(content, "contract-delivery:") {
			t.Fatalf("workflow content = %q", content)
		}
	}
}

func TestRulesStagesBindExactActiveMainCheckAndMarkerScopedCleanup(t *testing.T) {
	setup := mustBuild(t, "rules-setup").Request.Manifest.Resources[0]
	cleanup := mustBuild(t, "cleanup-rules").Request.Manifest.Resources[0]
	if setup.Name != runMarker+":ruleset:delivery-check" || setup.Marker != runMarker || setup.Attributes["enforcement"] != "active" || setup.Attributes["target"] != "branch" {
		t.Fatalf("rules setup identity = %#v", setup)
	}
	var definition struct {
		Enforcement string `json:"enforcement"`
		Conditions  struct {
			RefName struct {
				Include []string `json:"include"`
			} `json:"ref_name"`
		} `json:"conditions"`
		Rules []struct {
			Type       string `json:"type"`
			Parameters struct {
				Required []struct {
					Context       string `json:"context"`
					IntegrationID int64  `json:"integration_id"`
				} `json:"required_status_checks"`
			} `json:"parameters"`
		} `json:"rules"`
	}
	if err := json.Unmarshal([]byte(setup.Attributes["input:definition"]), &definition); err != nil {
		t.Fatal(err)
	}
	if definition.Enforcement != "active" || !slices.Equal(definition.Conditions.RefName.Include, []string{"refs/heads/main"}) || len(definition.Rules) != 1 || definition.Rules[0].Type != "required_status_checks" || len(definition.Rules[0].Parameters.Required) != 1 || definition.Rules[0].Parameters.Required[0].Context != "contract-delivery" || definition.Rules[0].Parameters.Required[0].IntegrationID != githubActionsIntegrationID {
		t.Fatalf("rules definition = %#v", definition)
	}
	if cleanup.DesiredState != engine.SandboxResourceAbsent || cleanup.Name != setup.Name || cleanup.Attributes["input:definition"] != setup.Attributes["input:definition"] {
		t.Fatalf("rules cleanup is not exact: %#v", cleanup)
	}
}

func TestRunIsDeterministicAndCredentialFree(t *testing.T) {
	args := validArgs("relationships-setup")
	var first, second bytes.Buffer
	if err := run(args, fixedNow, &first); err != nil {
		t.Fatal(err)
	}
	if err := run(args, fixedNow, &second); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("same approved input did not produce deterministic JSON")
	}
	lower := strings.ToLower(first.String())
	for _, forbidden := range []string{"private_key", "access_token", "client_secret"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("output contains credential field %q", forbidden)
		}
	}
	var decoded planInput
	if err := json.Unmarshal(first.Bytes(), &decoded); err != nil || decoded.Mandate.ID == "" {
		t.Fatalf("decode = %#v, %v", decoded, err)
	}
}

func TestRunRejectsUnapprovedOrAmbiguousInputs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		now  time.Time
	}{
		{"missing approval", withoutFlag(validArgs("issues-setup"), "--approved-by"), fixedNow},
		{"bad approval time", replaceFlag(validArgs("issues-setup"), "--approved-at", "yesterday"), fixedNow},
		{"future approval", replaceFlag(validArgs("issues-setup"), "--approved-at", "2026-07-22T00:00:00Z"), fixedNow},
		{"expired approval", replaceFlag(validArgs("issues-setup"), "--expires-at", "2026-07-21T17:00:00Z"), fixedNow},
		{"bad source", replaceFlag(validArgs("issues-setup"), "--source-revision", "main"), fixedNow},
		{"unknown stage", replaceFlag(validArgs("issues-setup"), "--stage", "cleanup"), fixedNow},
		{"missing relationship identity", withoutFlag(validArgs("relationships-setup"), "--parent-id"), fixedNow},
		{"governed delivery input required", validArgs("issues-governed"), fixedNow},
		{"leading zero identity", replaceFlag(validArgs("relationships-setup"), "--delivery-number", "012"), fixedNow},
		{"duplicate identity", replaceFlag(validArgs("relationships-setup"), "--dependent-id", "102"), fixedNow},
		{"cleanup identities required", identitiesOmitted(validArgs("cleanup-issues")), fixedNow},
		{"cleanup delivery pull required", withoutFlag(validArgs("cleanup-delivery"), "--pull-number"), fixedNow},
		{"cleanup delivery pull id required", withoutFlag(validArgs("cleanup-delivery"), "--pull-id"), fixedNow},
		{"cleanup delivery pull node required", withoutFlag(validArgs("cleanup-delivery"), "--pull-node-id"), fixedNow},
		{"cleanup delivery sha required", withoutFlag(validArgs("cleanup-delivery"), "--branch-head-sha"), fixedNow},
		{"cleanup delivery sha exact", replaceFlag(validArgs("cleanup-delivery"), "--branch-head-sha", "main"), fixedNow},
		{"positional argument", append(validArgs("issues-setup"), "unexpected"), fixedNow},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := run(test.args, test.now, &bytes.Buffer{}); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

func mustBuild(t *testing.T, stage string) planInput {
	t.Helper()
	args := validArgs(stage)
	if stage == "issues-governed" {
		args = append(args, "--delivery-input-file", governedDeliveryInput(t))
	}
	var output bytes.Buffer
	if err := run(args, fixedNow, &output); err != nil {
		t.Fatal(err)
	}
	var input planInput
	if err := json.Unmarshal(output.Bytes(), &input); err != nil {
		t.Fatal(err)
	}
	return input
}

func governedDeliveryInput(t *testing.T) string {
	t.Helper()
	contract := relatedIssueContract("Exact delivery contract.", runMarker+":issue:delivery")
	request := engine.DeliveryRequest{
		Intent: engine.DeliveryIntent{SourceRevision: strings.Repeat("a", 40), ManagedID: "issue:12", HeadBranch: deliveryHeadBranch},
		CompletionIntent: &engine.WorkDesiredIntent{
			Task:       engine.DesiredManagedTask{ManagedID: "issue:12", IssueType: "task", Title: "Issue 75 contract fixture: governed delivery", ParentManagedID: "issue:11", Readiness: "ready", Status: "done", Closed: true, Dependents: []engine.WorkDependentContext{{ManagedID: "issue:13"}}},
			Governance: &engine.GovernedWorkContract{SchemaVersion: 1, Issue: contract},
		},
	}
	path := filepath.Join(t.TempDir(), "delivery.json")
	content, err := json.Marshal(map[string]any{"request": request})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func validArgs(stage string) []string {
	return []string{
		"--stage", stage, "--repository", ".", "--source-revision", strings.Repeat("a", 40),
		"--approved-by", "owner", "--approval-id", "issue-comment-123", "--approved-at", "2026-07-21T17:00:00Z", "--expires-at", "2026-07-22T18:00:00Z",
		"--parent-number", "11", "--parent-id", "101", "--parent-node-id", "I_parent",
		"--delivery-number", "12", "--delivery-id", "102", "--delivery-node-id", "I_delivery",
		"--dependent-number", "13", "--dependent-id", "103", "--dependent-node-id", "I_dependent",
		"--pull-number", "17", "--pull-id", "117", "--pull-node-id", "PR_delivery", "--branch-head-sha", strings.Repeat("b", 40),
	}
}

func withoutFlag(args []string, name string) []string {
	result := slices.Clone(args)
	for index := 0; index < len(result); index++ {
		if result[index] == name {
			return append(result[:index], result[index+2:]...)
		}
	}
	return result
}

func replaceFlag(args []string, name, value string) []string {
	result := slices.Clone(args)
	for index := 0; index+1 < len(result); index++ {
		if result[index] == name {
			result[index+1] = value
			return result
		}
	}
	return result
}

func identitiesOmitted(args []string) []string {
	for _, name := range []string{"--parent-number", "--parent-id", "--parent-node-id", "--delivery-number", "--delivery-id", "--delivery-node-id", "--dependent-number", "--dependent-id", "--dependent-node-id"} {
		args = withoutFlag(args, name)
	}
	return args
}

func validateManifest(t *testing.T, input planInput) {
	t.Helper()
	repository := t.TempDir()
	if err := os.Mkdir(repository+"/.git", 0o755); err != nil {
		t.Fatal(err)
	}
	request := input.Request
	request.Repository = repository
	capability := engine.SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: input.Config.Roles[input.Role].Actor, EvidenceMode: "live", Target: input.Config.Target, Permissions: input.Mandate.Authority.Permissions, CredentialIdentities: input.Mandate.Authority.CredentialIdentities, Compatibility: input.Mandate.Authority.Compatibility, ConfigurationRevision: configuration, ObservedAt: fixedNow, ExpiresAt: fixedNow.Add(time.Hour)}
	observation := engine.SandboxObservation{SchemaVersion: 1, Target: input.Config.Target, ConfigurationRevision: configuration}
	adapter := engine.NewInMemorySandboxAdapter(capability, observation)
	if _, err := engine.New(engine.WithSandboxAdapter(adapter)).InspectSandbox(context.Background(), request); err != nil {
		t.Fatal(err)
	}
}
