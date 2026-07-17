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
	configuration = "issue-73-live-config-v2"
	markerPrefix  = "starter-kit-contract:"
	runMarker     = "starter-kit-contract:issue-73-20260717-01"
)

type planInput struct {
	Role    string                              `json:"role"`
	Request engine.SandboxRequest               `json:"request"`
	Config  githubadapter.SandboxConfig         `json:"config"`
	App     githubadapter.AppInstallationConfig `json:"app"`
}

func main() {
	flags := flag.NewFlagSet("issue73-contract", flag.ExitOnError)
	role := flags.String("role", "", "reconciler, seeder, or rules")
	repositoryPath := flags.String("repository", ".", "local evidence repository")
	source := flags.String("source-revision", "", "exact starter-kit source revision")
	baseSHA := flags.String("base-sha", "", "sandbox main revision used for fixture branches")
	flags.Parse(os.Args[1:])
	if *source == "" || (*role == githubadapter.SandboxRoleSeeder && *baseSHA == "") || flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "--role, --source-revision, and seeder --base-sha are required")
		os.Exit(2)
	}
	resources, expectation, app, err := setupRole(*role, *baseSHA)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	target := engine.SandboxTarget{Host: "github.com", OwnerID: ownerID, RepositoryID: repositoryID, ProjectID: projectID, RepositoryName: repository}
	manifest := engine.SandboxManifest{SchemaVersion: 1, OperationID: "issue-73-live-setup-" + *role + "-v2", SourceRevision: *source, ConfigurationRevision: configuration, ApprovedBy: "dragondad22", ApprovedPlan: "issue-73-bootstrap-v1", RecoveryOwner: "dragondad22", MarkerPrefix: markerPrefix, Target: target, Resources: resources}
	config := githubadapter.SandboxConfig{Host: "github.com", RESTBaseURL: "https://api.github.com", GraphQLURL: "https://api.github.com/graphql", APIVersion: "2026-03-10", ConfigurationRevision: configuration, Target: target, RepositoryOwner: "codex-starter-kit-labs", RepositoryName: "codex-starter-kit-sandbox", ProjectNumber: 1, Resources: resources, Roles: map[string]githubadapter.SandboxRoleExpectation{*role: expectation}, EvidenceMode: "live", LiveTargetApproved: true}
	json.NewEncoder(os.Stdout).Encode(planInput{Role: *role, Request: engine.SandboxRequest{Repository: *repositoryPath, Manifest: manifest}, Config: config, App: app})
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
	workflow := `name: Contract fixture checks
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

func rulesResources() []engine.SandboxResourceSpec {
	definition := `{"enforcement":"disabled","target":"branch","conditions":{"ref_name":{"include":["refs/heads/contract/issue-73-20260717-01/**"],"exclude":[]}},"rules":[{"type":"deletion"}]}`
	return []engine.SandboxResourceSpec{resource("ruleset:fixture", engine.SandboxResourceRuleset, runMarker+":ruleset", runMarker, map[string]string{"enforcement": "disabled", "target": "branch", "input:definition": definition})}
}

func resource(key, kind, name, marker string, attributes map[string]string) engine.SandboxResourceSpec {
	return engine.SandboxResourceSpec{Key: key, Kind: kind, Name: name, Marker: marker, Attributes: attributes}
}
