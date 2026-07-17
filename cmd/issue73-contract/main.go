// Command issue73-contract emits the approved, credential-free live qualification inputs.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const (
	ownerID       = "305967668"
	repositoryID  = "1303189066"
	projectID     = "PVT_kwDOEjyyNM4Bdm9F"
	repository    = "codex-starter-kit-labs/codex-starter-kit-sandbox"
	configuration = "issue-73-live-config-v3"
	markerPrefix  = "starter-kit-contract:"
	runMarker     = "starter-kit-contract:issue-73-20260717-01"
)

type planInput struct {
	Role     string                              `json:"role"`
	Request  engine.SandboxRequest               `json:"request"`
	Config   githubadapter.SandboxConfig         `json:"config"`
	App      githubadapter.AppInstallationConfig `json:"app"`
	Reviewer githubadapter.UserTokenConfig       `json:"reviewer"`
}

func main() {
	flags := flag.NewFlagSet("issue73-contract", flag.ExitOnError)
	role := flags.String("role", "", "reconciler, seeder, or rules")
	stage := flags.String("stage", "setup", "setup, qualification, rules-proof, project-proof, or cleanup")
	repositoryPath := flags.String("repository", ".", "local evidence repository")
	source := flags.String("source-revision", "", "exact starter-kit source revision")
	baseSHA := flags.String("base-sha", "", "sandbox main revision used for fixture branches")
	successHead := flags.String("success-head", "", "exact success fixture PR head revision")
	failingHead := flags.String("failing-head", "", "exact failing fixture PR head revision")
	flags.Parse(os.Args[1:])
	if *source == "" || (*stage == "setup" && *role == githubadapter.SandboxRoleSeeder && *baseSHA == "") || (*stage == "qualification" && *role == githubadapter.SandboxRoleReviewer && (*successHead == "" || *failingHead == "")) || flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "--role, --stage, --source-revision, and stage-specific revisions are required")
		os.Exit(2)
	}
	resources, expectation, app, reviewer, err := rolePlan(*stage, *role, *baseSHA, *successHead, *failingHead)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	target := engine.SandboxTarget{Host: "github.com", OwnerID: ownerID, RepositoryID: repositoryID, ProjectID: projectID, RepositoryName: repository}
	manifest := engine.SandboxManifest{SchemaVersion: 1, OperationID: "issue-73-live-" + *stage + "-" + *role + "-v3", SourceRevision: *source, ConfigurationRevision: configuration, ApprovedBy: "dragondad22", ApprovedPlan: "issue-73-bootstrap-v1", RecoveryOwner: "dragondad22", MarkerPrefix: markerPrefix, Target: target, Resources: resources}
	config := githubadapter.SandboxConfig{Host: "github.com", RESTBaseURL: "https://api.github.com", GraphQLURL: "https://api.github.com/graphql", APIVersion: "2026-03-10", ConfigurationRevision: configuration, Target: target, RepositoryOwner: "codex-starter-kit-labs", RepositoryName: "codex-starter-kit-sandbox", ProjectNumber: 1, Resources: resources, Roles: map[string]githubadapter.SandboxRoleExpectation{*role: expectation}, EvidenceMode: "live", LiveTargetApproved: true}
	json.NewEncoder(os.Stdout).Encode(planInput{Role: *role, Request: engine.SandboxRequest{Repository: *repositoryPath, Manifest: manifest}, Config: config, App: app, Reviewer: reviewer})
}

func rolePlan(stage, role, baseSHA, successHead, failingHead string) ([]engine.SandboxResourceSpec, githubadapter.SandboxRoleExpectation, githubadapter.AppInstallationConfig, githubadapter.UserTokenConfig, error) {
	if stage == "setup" {
		resources, expectation, appConfig, err := setupRole(role, baseSHA)
		return resources, expectation, appConfig, githubadapter.UserTokenConfig{}, err
	}
	switch stage + ":" + role {
	case "qualification:seeder":
		resources := []engine.SandboxResourceSpec{
			resource("fixture:pr:success", engine.SandboxResourceFixturePR, "success", runMarker+":pr:success", map[string]string{"title": "Contract fixture: success", "state": "open", "draft": "false", "head": "contract/issue-73-20260717-01/success", "base": "main"}),
			resource("fixture:issue:child", engine.SandboxResourceFixtureIssue, "child", runMarker+":issue:child", map[string]string{"title": "Contract fixture: child", "state": "closed", "input:labels": "contract-run,type:task"}),
			revocationResource(role),
		}
		permissions := []string{"contents:write", "issues:write", "metadata:read", "pull-requests:write", "workflows:write"}
		return resources, expectation(role, "codex-starter-kit-labs-seeder", "147094309", permissions), app("4319763", "147094309", "codex-starter-kit-labs-seeder"), githubadapter.UserTokenConfig{}, nil
	case "qualification:rules":
		resources := append(activeRulesResources(), revocationResource(role))
		permissions := []string{"administration:write", "metadata:read"}
		return resources, expectation(role, "codex-starter-kit-labs-rules", "147094473", permissions), app("4319800", "147094473", "codex-starter-kit-labs-rules"), githubadapter.UserTokenConfig{}, nil
	case "qualification:reviewer":
		permissions := []string{"contents:read", "pull-requests:write"}
		expectation := githubadapter.SandboxRoleExpectation{Mode: "user-token", Actor: "american-dragon-designs", Account: "american-dragon-designs", AccountID: "305973890", RequiredPermissions: permissions}
		reviewer := githubadapter.UserTokenConfig{RESTBaseURL: "https://api.github.com", APIVersion: "2026-03-10", Actor: "american-dragon-designs", ActorID: "305973890", RepositoryOwner: "codex-starter-kit-labs", RepositoryName: "codex-starter-kit-sandbox", ApprovedPermissions: permissions}
		resources := []engine.SandboxResourceSpec{
			resource("fixture:review:success", engine.SandboxResourceFixtureReview, "success approval", runMarker+":review:success", map[string]string{"pull_number": "13", "reviewer_id": "305973890", "state": "APPROVED", "commit_id": successHead, "input:event": "APPROVE"}),
			resource("fixture:review:failing", engine.SandboxResourceFixtureReview, "failing changes request", runMarker+":review:failing", map[string]string{"pull_number": "14", "reviewer_id": "305973890", "state": "CHANGES_REQUESTED", "commit_id": failingHead, "input:event": "REQUEST_CHANGES"}),
		}
		return resources, expectation, githubadapter.AppInstallationConfig{}, reviewer, nil
	case "rules-proof:seeder":
		resources := []engine.SandboxResourceSpec{
			resource("proof:rules-denial", engine.SandboxResourceFixtureDenial, "active fixture rules denial", runMarker+":proof:rules-denial", map[string]string{"branch": "contract/issue-73-20260717-01/cleanup", "status": "denied"}),
			revocationResource(role),
		}
		permissions := []string{"contents:write", "issues:write", "metadata:read", "pull-requests:write", "workflows:write"}
		return resources, expectation(role, "codex-starter-kit-labs-seeder", "147094309", permissions), app("4319763", "147094309", "codex-starter-kit-labs-seeder"), githubadapter.UserTokenConfig{}, nil
	case "project-proof:reconciler":
		resources := []engine.SandboxResourceSpec{
			resource("proof:auto-add-positive", engine.SandboxResourceProjectItemProof, "marked issue auto-added", runMarker+":issue:parent", map[string]string{"number": "4", "state": "open", "status": "Backlog"}),
			absentResource("proof:auto-add-negative", engine.SandboxResourceProjectItemProof, "unmarked issue not auto-added", runMarker+":issue:pagination-c", map[string]string{"number": "12"}),
			resource("proof:close-to-done", engine.SandboxResourceProjectItemProof, "closed issue moved to Done", runMarker+":issue:child", map[string]string{"number": "5", "state": "closed", "status": "Done"}),
		}
		permissions := []string{"actions:read", "checks:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"}
		return resources, expectation(role, "codex-starter-kit-labs-reconciler", "147093185", permissions), app("4319725", "147093185", "codex-starter-kit-labs-reconciler"), githubadapter.UserTokenConfig{}, nil
	case "cleanup:seeder":
		permissions := []string{"contents:write", "issues:write", "metadata:read", "pull-requests:write", "workflows:write"}
		return cleanupSeederResources(), expectation(role, "codex-starter-kit-labs-seeder", "147094309", permissions), app("4319763", "147094309", "codex-starter-kit-labs-seeder"), githubadapter.UserTokenConfig{}, nil
	case "cleanup:rules":
		permissions := []string{"administration:write", "metadata:read"}
		return []engine.SandboxResourceSpec{absentResource("ruleset:fixture", engine.SandboxResourceRuleset, runMarker+":ruleset", runMarker, map[string]string{"enforcement": "active", "target": "branch", "input:definition": activeRulesDefinition()})}, expectation(role, "codex-starter-kit-labs-rules", "147094473", permissions), app("4319800", "147094473", "codex-starter-kit-labs-rules"), githubadapter.UserTokenConfig{}, nil
	default:
		return nil, githubadapter.SandboxRoleExpectation{}, githubadapter.AppInstallationConfig{}, githubadapter.UserTokenConfig{}, fmt.Errorf("unsupported stage and role %q/%q", stage, role)
	}
}

func setupRole(role, baseSHA string) ([]engine.SandboxResourceSpec, githubadapter.SandboxRoleExpectation, githubadapter.AppInstallationConfig, error) {
	switch role {
	case githubadapter.SandboxRoleReconciler:
		permissions := []string{"actions:read", "checks:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"}
		return reconcilerResources(), expectation(role, "codex-starter-kit-labs-reconciler", "147093185", permissions), app("4319725", "147093185", "codex-starter-kit-labs-reconciler"), nil
	case githubadapter.SandboxRoleSeeder:
		permissions := []string{"contents:write", "issues:write", "metadata:read", "pull-requests:write", "workflows:write"}
		return seederResources(baseSHA), expectation(role, "codex-starter-kit-labs-seeder", "147094309", permissions), app("4319763", "147094309", "codex-starter-kit-labs-seeder"), nil
	case githubadapter.SandboxRoleRules:
		permissions := []string{"administration:write", "metadata:read"}
		return rulesResources(), expectation(role, "codex-starter-kit-labs-rules", "147094473", permissions), app("4319800", "147094473", "codex-starter-kit-labs-rules"), nil
	default:
		return nil, githubadapter.SandboxRoleExpectation{}, githubadapter.AppInstallationConfig{}, fmt.Errorf("unsupported setup role %q", role)
	}
}

func expectation(_ string, actor, installation string, permissions []string) githubadapter.SandboxRoleExpectation {
	return githubadapter.SandboxRoleExpectation{Mode: "app-installation", Actor: actor, Account: "codex-starter-kit-labs", AccountID: ownerID, InstallationID: installation, RequiredPermissions: permissions}
}

func app(id, installation, actor string) githubadapter.AppInstallationConfig {
	return githubadapter.AppInstallationConfig{RESTBaseURL: "https://api.github.com", APIVersion: "2026-03-10", AppID: id, InstallationID: installation, Actor: actor, Account: "codex-starter-kit-labs", AccountID: ownerID}
}

func reconcilerResources() []engine.SandboxResourceSpec {
	resources := []engine.SandboxResourceSpec{
		resource("label:type-task", engine.SandboxResourceLabel, "type:task", "", map[string]string{"color": "0075CA", "description": "Independently executable implementation work"}),
		resource("label:ready-for-agent", engine.SandboxResourceLabel, "ready-for-agent", "", map[string]string{"color": "0E8A16", "description": "Complete agent brief and intended agent executor"}),
		resource("label:contract-run", engine.SandboxResourceLabel, "contract-run", "", map[string]string{"color": "5319E7", "description": "Synthetic Codex Starter Kit contract fixture"}),
	}
	fields := []struct {
		name    string
		options []struct{ name, color, description string }
	}{
		{"Status", options([]string{"Backlog", "Next", "In progress", "Done"}, []string{"GRAY", "BLUE", "YELLOW", "PURPLE"}, []string{"Tracked but not selected for immediate execution", "Explicitly selected as the immediate queue", "Delivery has started", "Completed"})},
		{"Readiness", options([]string{"Intake", "Needs refinement", "Ready", "Blocked"}, nil, nil)},
		{"Horizon", options([]string{"Now", "Next", "Later"}, nil, nil)},
		{"Phase", options([]string{"Phase 0", "Phase 1", "Phase 2", "Phase 3", "Phase 4", "Phase 5", "Phase 6", "Phase 7", "Phase 8"}, nil, nil)},
	}
	for _, field := range fields {
		resources = append(resources, resource("project-field:"+strings.ToLower(field.name), engine.SandboxResourceProjectField, field.name, "", map[string]string{"data_type": "single_select"}))
		for _, option := range field.options {
			resources = append(resources, resource("project-option:"+strings.ToLower(field.name)+":"+strings.ToLower(strings.ReplaceAll(option.name, " ", "-")), engine.SandboxResourceProjectOption, option.name, "", map[string]string{"field": field.name, "color": option.color, "description": option.description}))
		}
	}
	resources = append(resources,
		resource("project-view:execution", engine.SandboxResourceProjectView, "Execution", "", map[string]string{"layout": "table"}),
		resource("project-view:horizon", engine.SandboxResourceProjectView, "Horizon", "", map[string]string{"layout": "roadmap"}),
		resource("project-workflow:auto-add", engine.SandboxResourceProjectWorkflow, "Auto-add to project", "", map[string]string{"enabled": "true", "number": "7"}),
		resource("project-workflow:item-closed", engine.SandboxResourceProjectWorkflow, "Item closed", "", map[string]string{"enabled": "true", "number": "1"}),
	)
	return resources
}

func options(names, colors, descriptions []string) []struct{ name, color, description string } {
	result := make([]struct{ name, color, description string }, len(names))
	for index, name := range names {
		color := "GRAY"
		if len(colors) != 0 {
			color = colors[index]
		}
		description := ""
		if len(descriptions) != 0 {
			description = descriptions[index]
		}
		result[index] = struct{ name, color, description string }{name, color, description}
	}
	return result
}

func seederResources(baseSHA string) []engine.SandboxResourceSpec {
	workflow := fixtureWorkflowContent()
	resources := []engine.SandboxResourceSpec{resource("fixture:workflow", engine.SandboxResourceFixtureWorkflow, "contract-fixture.yml", runMarker+":workflow", map[string]string{"path": ".github/workflows/contract-fixture.yml", "input:content": workflow})}
	names := []string{"parent", "child", "blocker", "dependent", "question", "research", "pagination-a", "pagination-b", "pagination-c"}
	for _, name := range names {
		labels := "contract-run,type:task"
		if name == "pagination-c" {
			labels = "type:task"
		}
		resources = append(resources, resource("fixture:issue:"+name, engine.SandboxResourceFixtureIssue, name, runMarker+":issue:"+name, map[string]string{"title": "Contract fixture: " + name, "state": "open", "input:labels": labels}))
	}
	branches := []string{"success", "failing", "cleanup"}
	for _, name := range branches {
		branch := "contract/issue-73-20260717-01/" + name
		resources = append(resources, resource("fixture:branch:"+name, engine.SandboxResourceFixtureBranch, branch, runMarker+":branch:"+name, map[string]string{"input:base_sha": baseSHA, "input:path": "fixtures/" + name + ".txt", "input:content": runMarker + ":branch:" + name + "\n"}))
	}
	resources = append(resources,
		resource("fixture:pr:success", engine.SandboxResourceFixturePR, "success", runMarker+":pr:success", map[string]string{"title": "Contract fixture: success", "state": "open", "draft": "true", "head": "contract/issue-73-20260717-01/success", "base": "main"}),
		resource("fixture:pr:failing", engine.SandboxResourceFixturePR, "failing", runMarker+":pr:failing", map[string]string{"title": "Contract fixture: failing", "state": "open", "draft": "false", "head": "contract/issue-73-20260717-01/failing", "base": "main"}),
	)
	return resources
}

func fixtureWorkflowContent() string {
	return `name: Contract fixture checks
on:
  pull_request:
permissions:
  contents: read
jobs:
  deterministic:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/github-script@ed597411d8f924073f98dfc5c65a23a2325f34cd
        with:
          script: |
            if (context.payload.pull_request.head.ref.includes('/failing')) {
              core.setFailed('intentional fixture failure')
            }
`
}

func rulesResources() []engine.SandboxResourceSpec {
	definition := `{"enforcement":"disabled","target":"branch","conditions":{"ref_name":{"include":["refs/heads/contract/issue-73-20260717-01/**"],"exclude":[]}},"rules":[{"type":"deletion"}]}`
	return []engine.SandboxResourceSpec{resource("ruleset:fixture", engine.SandboxResourceRuleset, runMarker+":ruleset", runMarker, map[string]string{"enforcement": "disabled", "target": "branch", "input:definition": definition})}
}

func activeRulesDefinition() string {
	return `{"enforcement":"active","target":"branch","conditions":{"ref_name":{"include":["refs/heads/contract/issue-73-20260717-01/**"],"exclude":[]}},"rules":[{"type":"deletion"}]}`
}

func activeRulesResources() []engine.SandboxResourceSpec {
	return []engine.SandboxResourceSpec{resource("ruleset:fixture", engine.SandboxResourceRuleset, runMarker+":ruleset", runMarker, map[string]string{"enforcement": "active", "target": "branch", "input:definition": activeRulesDefinition()})}
}

func revocationResource(role string) engine.SandboxResourceSpec {
	return resource("proof:token-revocation:"+role, engine.SandboxResourceTokenRevocation, role+" credential revocation", runMarker+":proof:token-revocation:"+role, map[string]string{"role": role, "state": "revoked", "status": "401"})
}

func cleanupSeederResources() []engine.SandboxResourceSpec {
	resources := []engine.SandboxResourceSpec{absentResource("fixture:workflow", engine.SandboxResourceFixtureWorkflow, "contract-fixture.yml", runMarker+":workflow", map[string]string{"path": ".github/workflows/contract-fixture.yml", "input:content": fixtureWorkflowContent()})}
	for _, name := range []string{"parent", "child", "blocker", "dependent", "question", "research", "pagination-a", "pagination-b", "pagination-c"} {
		labels := "contract-run,type:task"
		if name == "pagination-c" {
			labels = "type:task"
		}
		resources = append(resources, resource("fixture:issue:"+name, engine.SandboxResourceFixtureIssue, name, runMarker+":issue:"+name, map[string]string{"title": "Contract fixture: " + name, "state": "closed", "input:labels": labels}))
	}
	for _, name := range []string{"success", "failing", "cleanup"} {
		branch := "contract/issue-73-20260717-01/" + name
		resources = append(resources, absentResource("fixture:branch:"+name, engine.SandboxResourceFixtureBranch, branch, runMarker+":branch:"+name, map[string]string{}))
	}
	resources = append(resources,
		resource("fixture:pr:success", engine.SandboxResourceFixturePR, "success", runMarker+":pr:success", map[string]string{"title": "Contract fixture: success", "state": "closed", "draft": "false", "head": "contract/issue-73-20260717-01/success", "base": "main"}),
		resource("fixture:pr:failing", engine.SandboxResourceFixturePR, "failing", runMarker+":pr:failing", map[string]string{"title": "Contract fixture: failing", "state": "closed", "draft": "false", "head": "contract/issue-73-20260717-01/failing", "base": "main"}),
	)
	return resources
}

func absentResource(key, kind, name, marker string, attributes map[string]string) engine.SandboxResourceSpec {
	value := resource(key, kind, name, marker, attributes)
	value.DesiredState = engine.SandboxResourceAbsent
	return value
}

func resource(key, kind, name, marker string, attributes map[string]string) engine.SandboxResourceSpec {
	return engine.SandboxResourceSpec{Key: key, Kind: kind, Name: name, Marker: marker, Attributes: attributes}
}
