package githubadapter_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestSandboxAdapterAggregatesRoleAuthorityAndObservesManagedLabels(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer reconciler-token" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		switch request.URL.Path {
		case "/orgs/labs/projectsV2/1":
			json.NewEncoder(response).Encode(map[string]any{"node_id": "project-id", "number": 1, "owner": map[string]any{"login": "labs", "id": "owner-id", "type": "Organization"}})
		case "/repos/labs/sandbox/labels":
			json.NewEncoder(response).Encode([]map[string]any{{"id": 7, "node_id": "LA_label", "name": "type:task", "color": "0075CA", "description": "Task"}})
		case "/orgs/labs/projectsV2/1/fields":
			json.NewEncoder(response).Encode([]any{})
		case "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}}}})
		default:
			t.Fatalf("path = %q", request.URL.Path)
		}
	}))
	defer server.Close()

	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	adapter, err := githubadapter.NewSandbox(githubadapter.SandboxConfig{
		Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10",
		ConfigurationRevision: "config-1", Target: target, RepositoryOwner: "labs", RepositoryName: "sandbox", ProjectNumber: 1,
		EvidenceMode: "simulated", Resources: []engine.SandboxResourceSpec{{Key: "label:type-task", Kind: engine.SandboxResourceLabel, Name: "type:task"}},
		Roles: map[string]githubadapter.SandboxRoleExpectation{
			githubadapter.SandboxRoleReconciler: {Mode: "app-installation", Actor: "reconciler", Account: "labs", AccountID: "owner-id", InstallationID: "1", RequiredPermissions: []string{"issues:write", "organization-projects:write"}},
			githubadapter.SandboxRoleSeeder:     {Mode: "app-installation", Actor: "seeder", Account: "labs", AccountID: "owner-id", InstallationID: "2", RequiredPermissions: []string{"contents:write", "workflows:write"}},
			githubadapter.SandboxRoleRules:      {Mode: "app-installation", Actor: "rules", Account: "labs", AccountID: "owner-id", InstallationID: "3", RequiredPermissions: []string{"administration:write"}},
		},
	}, sandboxProviders(now), server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("new sandbox adapter: %v", err)
	}

	capability, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatalf("capability: %v", err)
	}
	if !capability.Available || capability.Actor != "reconciler+seeder+rules" || capability.ConfigurationRevision != "config-1" {
		t.Fatalf("capability = %#v", capability)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if len(observation.Resources) != 1 || observation.Resources[0].Key != "label:type-task" || observation.Resources[0].ID != "LA_label" {
		t.Fatalf("resources = %#v", observation.Resources)
	}
}

func TestSandboxAdapterRejectsBroaderThanApprovedRolePermission(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	providers := sandboxProviders(now)
	providers[githubadapter.SandboxRoleSeeder] = githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		return githubadapter.Credential{Token: "token", Mode: "app-installation", Actor: "seeder", Account: "labs", AccountID: "owner-id", InstallationID: "2", Permissions: []string{"contents:write", "workflows:write", "administration:write"}, ExpiresAt: now.Add(time.Hour)}, nil
	})
	adapter, err := githubadapter.NewSandbox(sandboxConfig(server, target), providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	capability, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if capability.Available || !strings.Contains(strings.Join(capability.Problems, " "), "seeder") {
		t.Fatalf("capability = %#v", capability)
	}
}

func TestSandboxAdapterUpdatesExistingManagedLabelInsteadOfCreatingDuplicate(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	var patched bool
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/labels":
			json.NewEncoder(response).Encode([]map[string]any{{"node_id": "LA_label", "name": "type:task", "color": "FFFFFF", "description": "old"}})
		case request.Method == http.MethodPatch && request.URL.Path == "/repos/labs/sandbox/labels/type:task":
			patched = true
			json.NewEncoder(response).Encode(map[string]any{"node_id": "LA_label", "name": "type:task", "color": "0075CA", "description": "Task"})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	adapter, err := githubadapter.NewSandbox(config, sandboxProviders(now), server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: engine.SandboxResourceSpec{Key: "label:type-task", Kind: engine.SandboxResourceLabel, Name: "type:task", Attributes: map[string]string{"color": "0075CA", "description": "Task"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !patched || result.ResourceID != "LA_label" {
		t.Fatalf("result = %#v, patched = %v", result, patched)
	}
}

func TestSandboxAdapterRoutesRulesAndFixturesToSeparateRoles(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/rulesets":
			if request.Header.Get("Authorization") != "Bearer rules-token" {
				t.Fatalf("rules authorization = %q", request.Header.Get("Authorization"))
			}
			json.NewEncoder(response).Encode([]any{})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/labs/sandbox/rulesets":
			if request.Header.Get("Authorization") != "Bearer rules-token" {
				t.Fatalf("rules authorization = %q", request.Header.Get("Authorization"))
			}
			json.NewEncoder(response).Encode(map[string]any{"id": 44, "name": "starter-kit-contract:run:rules", "enforcement": "disabled", "target": "branch"})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/labs/sandbox/issues":
			if request.Header.Get("Authorization") != "Bearer seeder-token" {
				t.Fatalf("seeder authorization = %q", request.Header.Get("Authorization"))
			}
			json.NewEncoder(response).Encode(map[string]any{"number": 9, "node_id": "I_issue", "title": "fixture", "body": "starter-kit-contract:run:issue", "state": "open"})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/issues":
			json.NewEncoder(response).Encode([]any{})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	config.Resources = []engine.SandboxResourceSpec{
		{Key: "ruleset:run", Kind: engine.SandboxResourceRuleset, Name: "starter-kit-contract:run:rules", Marker: "starter-kit-contract:run", Attributes: map[string]string{"enforcement": "disabled", "target": "branch", "input:definition": `{"enforcement":"disabled","target":"branch","conditions":{"ref_name":{"include":["refs/heads/contract/run/**"],"exclude":[]}},"rules":[]}`}},
		{Key: "fixture:issue", Kind: engine.SandboxResourceFixtureIssue, Name: "fixture", Marker: "starter-kit-contract:run:issue", Attributes: map[string]string{"title": "fixture", "state": "open"}},
	}
	adapter, err := githubadapter.NewSandbox(config, sandboxProviders(now), server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	rules, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: config.Resources[0]})
	if err != nil || rules.ResourceID != "44" {
		t.Fatalf("rules result = %#v, %v", rules, err)
	}
	issue, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: config.Resources[1]})
	if err != nil || issue.ResourceID != "9" {
		t.Fatalf("issue result = %#v, %v", issue, err)
	}
}

func TestSandboxRulesetObservationDriftForcesSetupUpdate(t *testing.T) {
	now := time.Date(2026, 7, 21, 22, 0, 0, 0, time.UTC)
	resource := deliveryRulesetResource(false)
	puts := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/rulesets":
			json.NewEncoder(response).Encode([]any{map[string]any{"id": 44, "name": resource.Name, "enforcement": "active", "target": "branch"}})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/rulesets/44":
			json.NewEncoder(response).Encode(rulesetHTTPDefinition(44, resource.Name, false, 15368, false))
		case request.Method == http.MethodPut && request.URL.Path == "/repos/labs/sandbox/rulesets/44":
			puts++
			json.NewEncoder(response).Encode(map[string]any{"id": 44, "name": resource.Name})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	config.Resources = []engine.SandboxResourceSpec{resource}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleRules, sandboxProviders(now)[githubadapter.SandboxRoleRules], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Resources) != 1 || observation.Resources[0].Attributes["definition"] == resource.Attributes["definition"] || observation.Resources[0].Attributes["definition_sha256"] == resource.Attributes["definition_sha256"] {
		t.Fatalf("drifted live definition was not retained: %#v, %v", observation, err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: resource})
	if err != nil || result.Outcome != "applied" || puts != 1 {
		t.Fatalf("drifted setup was not updated: %#v, puts=%d, err=%v", result, puts, err)
	}
}

func TestSandboxRulesetCleanupRefusesDefinitionDrift(t *testing.T) {
	now := time.Date(2026, 7, 21, 22, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name          string
		strict        bool
		integrationID int64
		detailName    string
		detailID      int64
		bypassActor   bool
		wantOutcome   string
		wantDeletes   int
	}{
		{name: "strictness drift", strict: false, integrationID: 15368, detailID: 44, wantOutcome: "needs-review"},
		{name: "integration drift", strict: true, integrationID: 999, detailID: 44, wantOutcome: "needs-review"},
		{name: "bypass actor drift", strict: true, integrationID: 15368, detailID: 44, bypassActor: true, wantOutcome: "needs-review"},
		{name: "identity drift", strict: true, integrationID: 15368, detailName: "different-ruleset", detailID: 45, wantOutcome: "needs-review"},
		{name: "exact definition", strict: true, integrationID: 15368, detailID: 44, wantOutcome: "applied", wantDeletes: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := deliveryRulesetResource(true)
			detailName := test.detailName
			if detailName == "" {
				detailName = resource.Name
			}
			deletes := 0
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.Header().Set("Content-Type", "application/json")
				switch {
				case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/rulesets":
					json.NewEncoder(response).Encode([]any{map[string]any{"id": 44, "name": resource.Name, "enforcement": "active", "target": "branch"}})
				case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/rulesets/44":
					json.NewEncoder(response).Encode(rulesetHTTPDefinition(test.detailID, detailName, test.strict, test.integrationID, test.bypassActor))
				case request.Method == http.MethodDelete && request.URL.Path == "/repos/labs/sandbox/rulesets/44":
					deletes++
					response.WriteHeader(http.StatusNoContent)
				default:
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
			}))
			defer server.Close()
			target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
			config := sandboxConfig(server, target)
			config.Resources = []engine.SandboxResourceSpec{resource}
			adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleRules, sandboxProviders(now)[githubadapter.SandboxRoleRules], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
			if err != nil {
				t.Fatal(err)
			}
			result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: resource})
			if err != nil || result.Outcome != test.wantOutcome || deletes != test.wantDeletes {
				t.Fatalf("cleanup = %#v, deletes=%d, err=%v", result, deletes, err)
			}
		})
	}
}

func deliveryRulesetResource(absent bool) engine.SandboxResourceSpec {
	definition := `{"bypass_actors":[],"conditions":{"ref_name":{"exclude":[],"include":["refs/heads/main"]}},"enforcement":"active","rules":[{"parameters":{"required_status_checks":[{"context":"contract-delivery","integration_id":15368}],"strict_required_status_checks_policy":true},"type":"required_status_checks"}],"target":"branch"}`
	resource := engine.SandboxResourceSpec{Key: "ruleset:delivery", Kind: engine.SandboxResourceRuleset, Name: "starter-kit-contract:issue-75:rules", Marker: "starter-kit-contract:issue-75", Attributes: map[string]string{"enforcement": "active", "target": "branch", "definition": definition, "definition_sha256": testSandboxSHA256(definition), "input:definition": definition}}
	if absent {
		resource.DesiredState = engine.SandboxResourceAbsent
	}
	return resource
}

func rulesetHTTPDefinition(id int64, name string, strict bool, integrationID int64, bypassActor bool) map[string]any {
	bypassActors := []any{}
	if bypassActor {
		bypassActors = append(bypassActors, map[string]any{"actor_id": 4, "actor_type": "Integration", "bypass_mode": "always"})
	}
	return map[string]any{"id": id, "name": name, "enforcement": "active", "target": "branch", "bypass_actors": bypassActors, "conditions": map[string]any{"ref_name": map[string]any{"exclude": []any{}, "include": []string{"refs/heads/main"}}}, "rules": []any{map[string]any{"type": "required_status_checks", "parameters": map[string]any{"required_status_checks": []any{map[string]any{"context": "contract-delivery", "integration_id": integrationID}}, "strict_required_status_checks_policy": strict}}}}
}

func TestSandboxAdapterClaimsFixtureWorkflowOnlyWhenContentExactlyMatches(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	const approved = "name: Contract fixture checks\non:\n  pull_request:\n"
	for _, test := range []struct {
		name       string
		content    string
		wantClaims int
	}{
		{name: "exact approved content", content: approved, wantClaims: 1},
		{name: "human-modified content", content: approved + "# modified\n", wantClaims: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodGet || request.URL.Path != "/repos/labs/sandbox/contents/.github/workflows/contract-fixture.yml" {
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				json.NewEncoder(response).Encode(map[string]any{"sha": "workflow-sha", "content": base64.StdEncoding.EncodeToString([]byte(test.content))})
			}))
			defer server.Close()

			target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
			config := sandboxConfig(server, target)
			config.Resources = []engine.SandboxResourceSpec{{
				Key: "fixture:workflow", Kind: engine.SandboxResourceFixtureWorkflow, Name: "contract-fixture.yml", Marker: "starter-kit-contract:run:workflow",
				Attributes: map[string]string{"path": ".github/workflows/contract-fixture.yml", "input:content": approved},
			}}
			adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
			if err != nil {
				t.Fatal(err)
			}

			observation, err := adapter.Observe(context.Background(), target)
			if err != nil {
				t.Fatal(err)
			}
			if len(observation.Resources) != test.wantClaims {
				t.Fatalf("resources = %#v, want %d claimed workflows", observation.Resources, test.wantClaims)
			}
		})
	}
}

func TestSandboxAdapterRetainsExpectedFixtureDenialProof(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusUnprocessableEntity} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodDelete && request.URL.Path == "/repos/labs/sandbox/git/refs/heads/contract/run/cleanup":
					response.WriteHeader(status)
				case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/rules/branches/contract/run/cleanup":
					json.NewEncoder(response).Encode([]map[string]string{{"type": "deletion"}})
				case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/git/ref/heads/contract/run/cleanup":
					json.NewEncoder(response).Encode(map[string]any{"ref": "refs/heads/contract/run/cleanup"})
				default:
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
			}))
			defer server.Close()
			target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
			config := sandboxConfig(server, target)
			proof := engine.SandboxResourceSpec{Key: "proof:rules-denial", Kind: engine.SandboxResourceFixtureDenial, Name: "active rules denial", Marker: "starter-kit-contract:run", Attributes: map[string]string{"branch": "contract/run/cleanup", "status": "denied"}}
			config.Resources = []engine.SandboxResourceSpec{proof}
			adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
			if err != nil {
				t.Fatal(err)
			}

			result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: proof})
			wantID := "http-" + strconv.Itoa(status)
			if err != nil || result.Outcome != "applied" || result.ResourceID != wantID {
				t.Fatalf("apply = %#v, %v", result, err)
			}
			observation, err := adapter.Observe(context.Background(), target)
			if err != nil || len(observation.Resources) != 1 || observation.Resources[0].Key != proof.Key || observation.Resources[0].ID != wantID {
				t.Fatalf("observation = %#v, %v", observation, err)
			}
		})
	}
}

func TestSandboxAdapterRejectsUnattributedFixtureDenial(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodDelete:
			response.WriteHeader(http.StatusUnprocessableEntity)
		case request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/rules/branches/"):
			json.NewEncoder(response).Encode([]map[string]string{{"type": "required_status_checks"}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	proof := engine.SandboxResourceSpec{Key: "proof:rules-denial", Kind: engine.SandboxResourceFixtureDenial, Name: "active rules denial", Marker: "starter-kit-contract:run", Attributes: map[string]string{"branch": "contract/run/cleanup", "status": "denied"}}
	config.Resources = []engine.SandboxResourceSpec{proof}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	if result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: proof}); err == nil || result.Outcome == "applied" {
		t.Fatalf("unattributed denial = %#v, %v", result, err)
	}
}

func TestSandboxAdapterRevokesAppTokenAndRetainsRejectionProof(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodDelete && request.URL.Path == "/installation/token":
			response.WriteHeader(http.StatusNoContent)
		case request.Method == http.MethodGet && request.URL.Path == "/installation/repositories":
			response.WriteHeader(http.StatusUnauthorized)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	proof := engine.SandboxResourceSpec{Key: "proof:seeder-revocation", Kind: engine.SandboxResourceTokenRevocation, Name: "seeder token revocation", Marker: "starter-kit-contract:run", Attributes: map[string]string{"role": githubadapter.SandboxRoleSeeder, "state": "revoked", "status": "401"}}
	config.Resources = []engine.SandboxResourceSpec{proof}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: proof})
	if err != nil || result.Outcome != "applied" || result.ResourceID != "http-401" || result.Detail != "App installation credential was revoked and rejected" {
		t.Fatalf("apply = %#v, %v", result, err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Resources) != 1 || observation.Resources[0].Key != proof.Key {
		t.Fatalf("observation = %#v, %v", observation, err)
	}
}

func TestSandboxAdapterRejectsGraphQLDraftTransitionErrors(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/pulls":
			json.NewEncoder(response).Encode([]map[string]any{{
				"node_id": "PR_fixture", "number": 13, "title": "Contract fixture: success",
				"body": "starter-kit-contract:run", "state": "open", "draft": true,
				"head": map[string]any{"ref": "contract/success"}, "base": map[string]any{"ref": "main"},
			}})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/pulls/13":
			json.NewEncoder(response).Encode(map[string]any{"node_id": "PR_fixture", "number": 13, "draft": true})
		case request.Method == http.MethodPatch && request.URL.Path == "/repos/labs/sandbox/pulls/13":
			json.NewEncoder(response).Encode(map[string]any{"number": 13})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"errors": []map[string]any{{"message": "transition rejected"}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	fixture := engine.SandboxResourceSpec{Key: "fixture:pr:success", Kind: engine.SandboxResourceFixturePR, Name: "success", Marker: "starter-kit-contract:run", Attributes: map[string]string{"title": "Contract fixture: success", "state": "open", "draft": "false", "head": "contract/success", "base": "main"}}
	config.Resources = []engine.SandboxResourceSpec{fixture}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: fixture}); err == nil || !strings.Contains(err.Error(), "draft transition") {
		t.Fatalf("apply error = %v", err)
	}
}

func TestSandboxAdapterObservesUserOwnedPhaseViewAndAssignment(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/dragondad22/codex-starter-kit/labels":
			json.NewEncoder(response).Encode([]any{})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			json.NewEncoder(response).Encode([]map[string]any{{"id": 50, "node_id": "F_phase", "name": "Phase", "data_type": "single_select", "options": []map[string]any{{"id": "O_phase_0", "name": "Phase 0", "color": "GRAY", "description": ""}}}})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{
				"views": map[string]any{"nodes": []map[string]any{{
					"id": "V_phases", "name": "Phases", "number": 6, "layout": "TABLE_LAYOUT", "filter": "",
					"fields":        map[string]any{"nodes": []map[string]any{{"id": "F_title", "name": "Title"}, {"id": "F_status", "name": "Status"}, {"id": "F_progress", "name": "Sub-issues progress"}, {"id": "F_readiness", "name": "Readiness"}}},
					"groupByFields": map[string]any{"nodes": []map[string]any{{"id": "F_phase", "name": "Phase"}}},
					"sortByFields":  map[string]any{"nodes": []map[string]any{{"direction": "ASC", "field": map[string]any{"id": "F_phase", "name": "Phase"}}}},
				}}},
				"workflows": map[string]any{"nodes": []any{}},
				"items":     map[string]any{"nodes": []map[string]any{{"id": "ITEM_1", "content": map[string]any{"id": "I_feature_1", "number": 1, "title": "Feature 1", "body": "", "state": "OPEN"}, "fieldValues": map[string]any{"nodes": []map[string]any{{"optionId": "O_phase_0", "field": map[string]any{"id": "F_phase"}}}}}}},
			}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	view := engine.SandboxResourceSpec{Key: "project-view:phases", Kind: engine.SandboxResourceProjectView, Name: "Phases", Attributes: map[string]string{"layout": "table", "filter": "", "visible_fields": "F_progress,F_readiness,F_status,F_title", "group_by": "F_phase", "sort_by": "F_phase:asc"}}
	assignment := engine.SandboxResourceSpec{Key: "project-item-field:feature-1-phase", Kind: engine.SandboxResourceProjectItemField, Name: "Feature 1 Phase", Attributes: map[string]string{"content_id": "I_feature_1", "field": "Phase", "field_id": "F_phase", "option_id": "O_phase_0"}}
	config, providers := userProjectSandboxConfig(server, target, now, view, assignment)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Problems) != 0 || len(observation.Resources) != 2 {
		t.Fatalf("observation = %#v, %v", observation, err)
	}
	if observation.Resources[0].ID != "ITEM_1" || observation.Resources[1].ID != "V_phases" {
		t.Fatalf("immutable Phase resources = %#v", observation.Resources)
	}
}

func TestSandboxAdapterReportsPersistentTransientProviderFailureWithoutEffects(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	projectReads := 0
	effectRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			effectRequests++
		}
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8":
			projectReads++
			response.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(response).Encode(map[string]any{"message": "provider detail must stay private"})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error { return nil }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	problems := strings.Join(capability.Problems, ";")
	if err != nil || capability.Available || capability.Fresh || projectReads != 3 || effectRequests != 0 || !strings.Contains(problems, "provider is transiently unavailable after bounded read retries") || strings.Contains(problems, "provider detail") || strings.Contains(problems, "identity") {
		t.Fatalf("persistent transient capability = %#v, reads=%d, effects=%d, err=%v", capability, projectReads, effectRequests, err)
	}
}

func TestSandboxAdapterReportsPersistentTransientUserIdentityRead(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	identityReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/user" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		identityReads++
		response.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(response).Encode(map[string]any{"message": "provider detail must stay private"})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error { return nil }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	problems := strings.Join(capability.Problems, ";")
	if err != nil || capability.Available || capability.Fresh || identityReads != 3 || !strings.Contains(problems, "provider is transiently unavailable after bounded read retries") || strings.Contains(problems, "actor") || strings.Contains(problems, "scope") || strings.Contains(problems, "provider detail") {
		t.Fatalf("persistent transient user identity = %#v, reads=%d, err=%v", capability, identityReads, err)
	}
}

func TestSandboxAdapterRecoversProjectFieldInventoryAfterTransientProviderFailure(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	fieldReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			fieldReads++
			if fieldReads == 1 {
				response.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			json.NewEncoder(response).Encode([]map[string]any{{"node_id": "F_phase", "name": "Phase", "data_type": "single_select", "options": []any{}}})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	field := engine.SandboxResourceSpec{Key: "project-field:phase", Kind: engine.SandboxResourceProjectField, Name: "Phase", Attributes: map[string]string{"data_type": "single_select", "node_id": "F_phase"}}
	config, providers := userProjectSandboxConfig(server, target, now, field)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error { return nil }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Problems) != 0 || len(observation.Resources) != 1 || observation.Resources[0].ID != "F_phase" || fieldReads != 2 {
		t.Fatalf("transient field recovery = %#v, reads=%d, err=%v", observation, fieldReads, err)
	}
}

func TestSandboxAdapterReportsPersistentTransientProjectInventory(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	fieldReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/users/dragondad22/projectsV2/8/fields" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		fieldReads++
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	field := engine.SandboxResourceSpec{Key: "project-field:phase", Kind: engine.SandboxResourceProjectField, Name: "Phase", Attributes: map[string]string{"data_type": "single_select", "node_id": "F_phase"}}
	config, providers := userProjectSandboxConfig(server, target, now, field)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error { return nil }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	problems := strings.Join(observation.Problems, ";")
	if err != nil || len(observation.Resources) != 0 || fieldReads != 3 || !strings.Contains(problems, "provider is transiently unavailable after bounded read retries") || strings.Contains(problems, "field inventory") {
		t.Fatalf("persistent transient inventory = %#v, reads=%d, err=%v", observation, fieldReads, err)
	}
}

func TestSandboxAdapterInventoriesTheCompletePhaseOptionCatalog(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	options := []map[string]any{}
	resources := []engine.SandboxResourceSpec{{Key: "project-field:phase", Kind: engine.SandboxResourceProjectField, Name: "Phase", Attributes: map[string]string{"data_type": "single_select", "node_id": "F_phase"}}}
	for index := 0; index <= 8; index++ {
		name := fmt.Sprintf("Phase %d", index)
		id := fmt.Sprintf("O_phase_%d", index)
		options = append(options, map[string]any{"id": id, "name": name, "color": "GRAY", "description": ""})
		resources = append(resources, engine.SandboxResourceSpec{Key: fmt.Sprintf("project-option:phase-%d", index), Kind: engine.SandboxResourceProjectOption, Name: name, Attributes: map[string]string{"field": "Phase", "color": "GRAY", "description": "", "option_id": id}})
	}
	options = append(options, map[string]any{"id": "O_phase_extra", "name": "Phase 9", "color": "GRAY", "description": ""})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			json.NewEncoder(response).Encode([]map[string]any{{"id": 50, "node_id": "F_phase", "name": "Phase", "data_type": "single_select", "options": options}})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now, resources...)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || !strings.Contains(strings.Join(observation.Problems, ";"), "complete Phase option catalog") {
		t.Fatalf("Phase catalog observation = %#v, %v", observation, err)
	}
}

func TestSandboxAdapterFollowsProjectItemCursorsForPhaseAssignments(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields" {
			json.NewEncoder(response).Encode([]any{})
			return
		}
		var input struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Variables["after"] == nil {
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []map[string]any{{"id": "ITEM_other", "content": map[string]any{"id": "I_other"}}}, "pageInfo": map[string]any{"hasNextPage": true, "endCursor": "cursor-1"}}}}})
			return
		}
		if input.Variables["after"] != "cursor-1" {
			t.Fatalf("cursor = %#v", input.Variables["after"])
		}
		json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"items": map[string]any{"nodes": []map[string]any{{"id": "ITEM_1", "content": map[string]any{"id": "I_feature_1"}, "fieldValues": map[string]any{"nodes": []map[string]any{{"optionId": "O_phase_0", "field": map[string]any{"id": "F_phase"}}}}}}, "pageInfo": map[string]any{"hasNextPage": false}}}}})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	assignment := engine.SandboxResourceSpec{Key: "project-item-field:feature-1-phase", Kind: engine.SandboxResourceProjectItemField, Name: "Feature 1 Phase", Attributes: map[string]string{"content_id": "I_feature_1", "field": "Phase", "field_id": "F_phase", "option_id": "O_phase_0"}}
	config, providers := userProjectSandboxConfig(server, target, now, assignment)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Problems) != 0 || len(observation.Resources) != 1 || observation.Resources[0].ID != "ITEM_1" {
		t.Fatalf("paginated Project observation = %#v, %v", observation, err)
	}
}

func TestSandboxAdapterReportsProjectItemPaginationExhaustion(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			json.NewEncoder(response).Encode([]any{})
			return
		}
		page++
		node := map[string]any{"items": map[string]any{"nodes": []any{}, "pageInfo": map[string]any{"hasNextPage": true, "endCursor": fmt.Sprintf("cursor-%d", page)}}}
		if page == 1 {
			node["views"] = map[string]any{"nodes": []any{}}
			node["workflows"] = map[string]any{"nodes": []any{}}
		}
		json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": node}})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	assignment := engine.SandboxResourceSpec{Key: "project-item-field:feature-1-phase", Kind: engine.SandboxResourceProjectItemField, Name: "Feature 1 Phase", Attributes: map[string]string{"content_id": "I_feature_1", "field": "Phase", "field_id": "F_phase", "option_id": "O_phase_0"}}
	config, providers := userProjectSandboxConfig(server, target, now, assignment)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || !strings.Contains(strings.Join(observation.Problems, ";"), "pagination exhausted") {
		t.Fatalf("pagination exhaustion = %#v, %v", observation, err)
	}
}

func TestSandboxAdapterVerifiesUserProjectActorAndClassicScope(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		switch request.URL.Path {
		case "/user":
			response.Header().Set("X-OAuth-Scopes", "gist, project, repo")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case "/users/dragondad22/projectsV2/8":
			json.NewEncoder(response).Encode(map[string]any{"node_id": "P_project", "number": 8, "owner": map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	expectation := config.Roles[githubadapter.SandboxRoleReconciler]
	expectation.ClassicOAuthScopes = []string{"gist", "project", "repo"}
	config.Roles[githubadapter.SandboxRoleReconciler] = expectation
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || !capability.Available || capability.Actor != githubadapter.SandboxRoleReconciler || !slices.Equal(capability.Permissions, []string{"reconciler:classic-scope:gist", "reconciler:classic-scope:project", "reconciler:classic-scope:repo", "reconciler:projects:write"}) {
		t.Fatalf("user Project capability = %#v, %v", capability, err)
	}
}

func TestSandboxAdapterRejectsProjectIdentityThatDoesNotMatchGraphQLTarget(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	projectReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case "/users/dragondad22/projectsV2/8":
			projectReads++
			json.NewEncoder(response).Encode(map[string]any{"node_id": "P_other", "number": 8, "owner": map[string]any{"login": "someone-else", "id": 42, "type": "User"}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	capability, err := adapter.Capability(context.Background())
	if err != nil || capability.Available || projectReads != 1 || !strings.Contains(strings.Join(capability.Problems, ";"), "Project immutable identity or owner") {
		t.Fatalf("mismatched Project capability = %#v, %v", capability, err)
	}
}

func TestSandboxAdapterRecoversProjectIdentityAfterGatewayTransient(t *testing.T) {
	for _, status := range []int{http.StatusBadGateway, http.StatusGatewayTimeout} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
			projectReads := 0
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/user":
					response.Header().Set("X-OAuth-Scopes", "project")
					json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
				case "/users/dragondad22/projectsV2/8":
					projectReads++
					if projectReads == 1 {
						response.WriteHeader(status)
						return
					}
					json.NewEncoder(response).Encode(map[string]any{"node_id": "P_project", "number": 8, "owner": map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"}})
				default:
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
			}))
			defer server.Close()
			target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
			config, providers := userProjectSandboxConfig(server, target, now)
			adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error { return nil }))
			if err != nil {
				t.Fatal(err)
			}
			capability, err := adapter.Capability(context.Background())
			if err != nil || !capability.Available || projectReads != 2 {
				t.Fatalf("gateway transient recovery = %#v, reads=%d, err=%v", capability, projectReads, err)
			}
		})
	}
}

func TestSandboxAdapterHonorsBoundedRetryAfterForProjectReads(t *testing.T) {
	for _, test := range []struct {
		name          string
		retryAfter    []string
		wantAvailable bool
		wantReads     int
		wantWaits     []time.Duration
		wantProblem   string
	}{
		{name: "within budget", retryAfter: []string{"1"}, wantAvailable: true, wantReads: 2, wantWaits: []time.Duration{time.Second}},
		{name: "beyond budget", retryAfter: []string{"3"}, wantAvailable: false, wantReads: 1, wantWaits: []time.Duration{}, wantProblem: "rate limit exceeds the bounded read retry budget"},
		{name: "beyond cumulative budget", retryAfter: []string{"1", "2"}, wantAvailable: false, wantReads: 2, wantWaits: []time.Duration{time.Second}, wantProblem: "rate limit exceeds the bounded read retry budget"},
	} {
		t.Run(test.name, func(t *testing.T) {
			now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
			projectReads := 0
			waits := []time.Duration{}
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/user":
					response.Header().Set("X-OAuth-Scopes", "project")
					json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
				case "/users/dragondad22/projectsV2/8":
					projectReads++
					if projectReads <= len(test.retryAfter) {
						response.Header().Set("Retry-After", test.retryAfter[projectReads-1])
						response.WriteHeader(http.StatusTooManyRequests)
						return
					}
					json.NewEncoder(response).Encode(map[string]any{"node_id": "P_project", "number": 8, "owner": map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"}})
				default:
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
			}))
			defer server.Close()
			target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
			config, providers := userProjectSandboxConfig(server, target, now)
			adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(_ context.Context, delay time.Duration) error {
				waits = append(waits, delay)
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			capability, err := adapter.Capability(context.Background())
			if err != nil || capability.Available != test.wantAvailable || projectReads != test.wantReads || !slices.Equal(waits, test.wantWaits) {
				t.Fatalf("Retry-After capability = %#v, reads=%d, waits=%v, err=%v", capability, projectReads, waits, err)
			}
			problems := strings.Join(capability.Problems, ";")
			if test.wantProblem != "" && (!strings.Contains(problems, test.wantProblem) || strings.Contains(problems, "transiently unavailable")) {
				t.Fatalf("bounded Retry-After diagnostic = %#v", capability.Problems)
			}
		})
	}
}

func TestSandboxAdapterReportsTransientAfterRateWaitConsumesRetryBudget(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	projectReads := 0
	waits := []time.Duration{}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case "/users/dragondad22/projectsV2/8":
			projectReads++
			if projectReads == 1 {
				response.Header().Set("Retry-After", "2")
				response.WriteHeader(http.StatusTooManyRequests)
				return
			}
			response.WriteHeader(http.StatusServiceUnavailable)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(_ context.Context, delay time.Duration) error {
		waits = append(waits, delay)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	problems := strings.Join(capability.Problems, ";")
	if err != nil || capability.Available || projectReads != 2 || !slices.Equal(waits, []time.Duration{2 * time.Second}) || !strings.Contains(problems, "provider is transiently unavailable") || strings.Contains(problems, "identity") {
		t.Fatalf("mixed transient capability = %#v, reads=%d, waits=%v, err=%v", capability, projectReads, waits, err)
	}
}

func TestSandboxAdapterDoesNotRetryNonTransientProjectResponses(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusTooManyRequests} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
			projectReads := 0
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/user":
					response.Header().Set("X-OAuth-Scopes", "project")
					json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
				case "/users/dragondad22/projectsV2/8":
					projectReads++
					if status == http.StatusTooManyRequests {
						response.Header().Set("Retry-After", "+1")
					}
					response.WriteHeader(status)
				default:
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
			}))
			defer server.Close()
			target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
			config, providers := userProjectSandboxConfig(server, target, now)
			adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error {
				t.Fatal("non-transient response must not wait for a retry")
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			capability, err := adapter.Capability(context.Background())
			if err != nil || capability.Available || projectReads != 1 || strings.Contains(strings.Join(capability.Problems, ";"), "transiently unavailable") {
				t.Fatalf("non-transient capability = %#v, reads=%d, err=%v", capability, projectReads, err)
			}
		})
	}
}

func TestSandboxAdapterDoesNotRetryMalformedSuccessfulProjectResponse(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	projectReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case "/users/dragondad22/projectsV2/8":
			projectReads++
			response.Write([]byte("{"))
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error {
		t.Fatal("decode failure must not wait for a retry")
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || capability.Available || projectReads != 1 || strings.Contains(strings.Join(capability.Problems, ";"), "transiently unavailable") {
		t.Fatalf("malformed response capability = %#v, reads=%d, err=%v", capability, projectReads, err)
	}
}

func TestSandboxAdapterDoesNotRetryRESTEffect(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	viewCreates := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			json.NewEncoder(response).Encode([]any{})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
		case request.Method == http.MethodPost && request.URL.Path == "/users/dragondad22/projectsV2/8/views":
			viewCreates++
			response.WriteHeader(http.StatusServiceUnavailable)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	view := engine.SandboxResourceSpec{Key: "project-view:phases", Kind: engine.SandboxResourceProjectView, Name: "Phases", Attributes: map[string]string{"layout": "table", "filter": "", "visible_fields": "", "group_by": "", "sort_by": ""}}
	config, providers := userProjectSandboxConfig(server, target, now, view)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error {
		t.Fatal("effect request must not wait for a retry")
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: view})
	if err == nil || result.Outcome == "applied" || viewCreates != 1 {
		t.Fatalf("transient REST effect = %#v, creates=%d, err=%v", result, viewCreates, err)
	}
}

func TestSandboxAdapterStopsTransientRetryWhenContextIsCanceled(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	projectReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case "/users/dragondad22/projectsV2/8":
			projectReads++
			response.WriteHeader(http.StatusServiceUnavailable)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	ctx, cancel := context.WithCancel(context.Background())
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error {
		cancel()
		return ctx.Err()
	}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = adapter.Capability(ctx)
	if !errors.Is(err, context.Canceled) || projectReads != 1 {
		t.Fatalf("canceled retry reads=%d, err=%v", projectReads, err)
	}
}

func TestSandboxAdapterRecoversProjectIdentityAfterTransientProviderFailure(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	projectReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8":
			projectReads++
			if projectReads == 1 {
				response.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(response).Encode(map[string]any{"message": "provider detail must stay private"})
				return
			}
			json.NewEncoder(response).Encode(map[string]any{"node_id": "P_project", "number": 8, "owner": map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }), githubadapter.WithSandboxRetryWait(func(context.Context, time.Duration) error { return nil }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || !capability.Available || !capability.Fresh || len(capability.Problems) != 0 || projectReads != 2 {
		t.Fatalf("transient identity recovery = %#v, reads=%d, err=%v", capability, projectReads, err)
	}
}

func TestSandboxAdapterRejectsUnexpressibleRequiredProjectViewConfigurationWithoutCreating(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/dragondad22/codex-starter-kit/labels":
			json.NewEncoder(response).Encode([]any{})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			json.NewEncoder(response).Encode([]any{})
		case request.Method == http.MethodPost && request.URL.Path == "/users/dragondad22/projectsV2/8/views":
			created = true
			json.NewEncoder(response).Encode(map[string]any{"value": map[string]any{"node_id": "V_phases"}})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			views := []any{}
			if created {
				views = append(views, map[string]any{"id": "V_phases", "name": "Phases", "number": 6, "layout": "TABLE_LAYOUT", "filter": "", "fields": map[string]any{"nodes": []map[string]any{{"id": "F_phase"}, {"id": "F_status"}}}, "groupByFields": map[string]any{"nodes": []any{}}, "sortByFields": map[string]any{"nodes": []any{}}})
			}
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": views}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	view := engine.SandboxResourceSpec{Key: "project-view:phases", Kind: engine.SandboxResourceProjectView, Name: "Phases", Attributes: map[string]string{"layout": "table", "filter": "", "visible_fields": "F_phase,F_status", "group_by": "F_phase", "sort_by": "F_phase:asc", "input:visible_fields": "50,51"}}
	config, providers := userProjectSandboxConfig(server, target, now, view)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: view})
	if err != nil || result.Outcome != "not-configured" || created {
		t.Fatalf("view result = %#v, created=%v, err=%v", result, created, err)
	}
}

func TestSandboxAdapterAdoptsTheProviderAssignedIdentityForAnExpressibleCreatedView(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			json.NewEncoder(response).Encode([]any{})
		case request.Method == http.MethodPost && request.URL.Path == "/users/dragondad22/projectsV2/8/views":
			created = true
			json.NewEncoder(response).Encode(map[string]any{"value": map[string]any{"node_id": "V_provider"}})
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			views := []any{}
			if created {
				views = append(views, map[string]any{"id": "V_provider", "name": "All work", "number": 7, "layout": "TABLE_LAYOUT", "filter": "", "fields": map[string]any{"nodes": []map[string]any{{"id": "F_status"}}}, "groupByFields": map[string]any{"nodes": []any{}}, "sortByFields": map[string]any{"nodes": []any{}}})
			}
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": views}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	view := engine.SandboxResourceSpec{Key: "project-view:all-work", Kind: engine.SandboxResourceProjectView, Name: "All work", Attributes: map[string]string{"layout": "table", "filter": "", "visible_fields": "F_status", "group_by": "", "sort_by": "", "input:visible_fields": "50"}}
	config, providers := userProjectSandboxConfig(server, target, now, view)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: view})
	if err != nil || result.Outcome != "applied" || result.ResourceID != "V_provider" {
		t.Fatalf("view identity adoption = %#v, %v", result, err)
	}
}

func TestSandboxAdapterRejectsUnexpectedClassicOAuthScope(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-OAuth-Scopes", "gist, project, repo")
		json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now)
	expectation := config.Roles[githubadapter.SandboxRoleReconciler]
	expectation.ClassicOAuthScopes = []string{"project"}
	config.Roles[githubadapter.SandboxRoleReconciler] = expectation
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || capability.Available || !slices.Equal(capability.Permissions, []string{"reconciler:classic-scope:gist", "reconciler:classic-scope:project", "reconciler:classic-scope:repo", "reconciler:projects:write"}) || !strings.Contains(strings.Join(capability.Problems, ";"), "scope set") {
		t.Fatalf("expanded scope capability = %#v, %v", capability, err)
	}
}

func TestSandboxAdapterReportsUnavailableUserViewRoute(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields" {
			json.NewEncoder(response).Encode([]any{})
			return
		}
		if request.Method == http.MethodPost && request.URL.Path == "/graphql" {
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
			return
		}
		response.WriteHeader(http.StatusNotFound)
		json.NewEncoder(response).Encode(map[string]any{"message": "Not Found"})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	view := engine.SandboxResourceSpec{Key: "project-view:phases", Kind: engine.SandboxResourceProjectView, Name: "Phases", Attributes: map[string]string{"layout": "table", "filter": "", "visible_fields": "", "group_by": "", "sort_by": ""}}
	config, providers := userProjectSandboxConfig(server, target, now, view)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: view})
	if err != nil || result.Outcome != "not-configured" {
		t.Fatalf("unavailable view route = %#v, %v", result, err)
	}
}

func TestSandboxAdapterReconcilesProjectOptionsThroughConfiguredOwnerRoute(t *testing.T) {
	for _, test := range []struct {
		name      string
		ownerKind string
		fieldPath string
	}{
		{name: "user", ownerKind: "user", fieldPath: "/users/dragondad22/projectsV2/8/fields"},
		{name: "organization", ownerKind: "organization", fieldPath: "/orgs/dragondad22/projectsV2/8/fields"},
	} {
		t.Run(test.name, func(t *testing.T) {
			now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodGet && request.URL.Path == test.fieldPath:
					json.NewEncoder(response).Encode([]map[string]any{{"node_id": "F_phase", "name": "Phase", "data_type": "single_select", "options": []any{}}})
				case request.Method == http.MethodPost && request.URL.Path == "/graphql":
					json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"updateProjectV2Field": map[string]any{"projectV2Field": map[string]any{"id": "F_phase"}}}})
				default:
					t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
			}))
			defer server.Close()
			target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
			option := engine.SandboxResourceSpec{Key: "project-option:phase-0", Kind: engine.SandboxResourceProjectOption, Name: "Phase 0", Attributes: map[string]string{"field": "Phase", "color": "GRAY", "description": ""}}
			config, providers := userProjectSandboxConfig(server, target, now, option)
			config.ProjectOwnerKind = test.ownerKind
			adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
			if err != nil {
				t.Fatal(err)
			}
			result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: option})
			if err != nil || result.Outcome != "applied" {
				t.Fatalf("option reconciliation = %#v, %v", result, err)
			}
		})
	}
}

func TestSandboxAdapterAdoptsProviderAssignedPhaseIdentitiesOnCleanCreate(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	resources := []engine.SandboxResourceSpec{{Key: "project-field:phase", Kind: engine.SandboxResourceProjectField, Name: "Phase", Attributes: map[string]string{"data_type": "single_select"}}}
	for index := 0; index <= 8; index++ {
		resources = append(resources, engine.SandboxResourceSpec{Key: fmt.Sprintf("project-option:phase-%d", index), Kind: engine.SandboxResourceProjectOption, Name: fmt.Sprintf("Phase %d", index), Attributes: map[string]string{"field": "Phase", "color": "GRAY", "description": ""}})
	}
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/user":
			response.Header().Set("X-OAuth-Scopes", "project")
			json.NewEncoder(response).Encode(map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8":
			json.NewEncoder(response).Encode(map[string]any{"node_id": "P_project", "number": 8, "owner": map[string]any{"login": "dragondad22", "id": 19365745, "type": "User"}})
		case request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields":
			fields := []any{}
			if created {
				options := []map[string]any{}
				for index := 0; index <= 8; index++ {
					options = append(options, map[string]any{"id": fmt.Sprintf("O_provider_%d", index), "name": fmt.Sprintf("Phase %d", index), "color": "GRAY", "description": ""})
				}
				fields = append(fields, map[string]any{"id": 50, "node_id": "F_provider", "name": "Phase", "data_type": "single_select", "options": options})
			}
			json.NewEncoder(response).Encode(fields)
		case request.Method == http.MethodPost && request.URL.Path == "/graphql":
			var input struct {
				Query string `json:"query"`
			}
			if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(input.Query, "createProjectV2Field") {
				created = true
				json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"createProjectV2Field": map[string]any{"projectV2Field": map[string]any{"id": "F_provider"}}}})
				return
			}
			if strings.Contains(input.Query, "updateProjectV2Field") {
				t.Fatal("already-converged provider options must not be rewritten")
			}
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}}}}})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	config, providers := userProjectSandboxConfig(server, target, now, resources...)
	expectation := config.Roles[githubadapter.SandboxRoleReconciler]
	expectation.ClassicOAuthScopes = []string{"project"}
	config.Roles[githubadapter.SandboxRoleReconciler] = expectation
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	authority := engine.SandboxAuthorityProfile{
		CredentialIdentities: []string{githubadapter.SandboxCredentialIdentity(githubadapter.SandboxRoleReconciler, expectation)},
		Permissions:          []string{"reconciler:classic-scope:project", "reconciler:projects:write"},
		EvidenceMode:         "simulated", Compatibility: "github.com:api.github.com:2026-03-10:native-rest-graphql",
		DataClass: "public-project-metadata", CostCeiling: "zero-dollar", Destructive: "no-delete", Retention: "30-days",
	}
	manifest := engine.SandboxManifest{SchemaVersion: 1, OperationID: "clean-phase-create", SourceRevision: "source", ConfigurationRevision: config.ConfigurationRevision, ApprovedBy: "owner", ApprovedPlan: "approval-record", RecoveryOwner: "owner", MarkerPrefix: "starter-kit-contract:phase", Target: target, Authority: authority, Resources: resources}
	lifecycle := engine.New(engine.WithClock(adapterFixedClock{now}), engine.WithSandboxAdapter(adapter))
	repository := t.TempDir()
	if output, err := exec.Command("git", "init", "--quiet", repository).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	inspection, err := lifecycle.InspectSandbox(context.Background(), engine.SandboxRequest{Repository: repository, Manifest: manifest})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanSandbox(context.Background(), inspection)
	if err != nil || len(plan.Effects) != 10 {
		t.Fatalf("clean-create plan = %#v, %v", plan, err)
	}
	mandate := engine.BindSandboxExecutionMandate(engine.SandboxExecutionMandate{
		SchemaVersion: 1, ApprovedBy: "owner", ApprovalID: "approval-record", ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Target: target,
		Actors: []string{githubadapter.SandboxRoleReconciler}, MarkerPrefix: manifest.MarkerPrefix, ResourceKinds: []string{engine.SandboxResourceProjectField, engine.SandboxResourceProjectOption}, EffectKinds: []string{"reconcile-resource"}, MaxEffects: 10,
		DataClass: authority.DataClass, CostCeiling: authority.CostCeiling, Destructive: authority.Destructive, Retention: authority.Retention, RecoveryOwner: manifest.RecoveryOwner, Authority: authority,
	}, resources...)
	apply, err := lifecycle.ApplySandbox(context.Background(), plan, engine.SandboxPlanApproval{SchemaVersion: 2, Mandate: &mandate})
	if err != nil || apply.Status != engine.SandboxApplyApplied || len(apply.Receipts) != 10 || apply.Receipts[0].ResourceID != "F_provider" || apply.Receipts[1].ResourceID != "O_provider_0" {
		t.Fatalf("clean-create apply = %#v, %v", apply, err)
	}
	verification, err := lifecycle.VerifySandbox(context.Background(), manifest)
	if err != nil || verification.OverallState != engine.ControlPass {
		t.Fatalf("clean-create verification = %#v, %v", verification, err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Problems) != 0 || len(observation.Resources) != 10 || observation.Resources[0].ID != "F_provider" {
		t.Fatalf("provider-bound observation = %#v, %v", observation, err)
	}
	config.Resources[0].Attributes["node_id"] = "F_stale"
	config.Resources[1].Attributes["option_id"] = "O_stale"
	staleAdapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	staleObservation, err := staleAdapter.Observe(context.Background(), target)
	if err != nil || !strings.Contains(strings.Join(staleObservation.Problems, ";"), "immutable identity") {
		t.Fatalf("stale provider identities = %#v, %v", staleObservation, err)
	}
}

type adapterFixedClock struct{ now time.Time }

func (clock adapterFixedClock) Now() time.Time { return clock.now }

func TestSandboxAdapterReconcilesProjectItemFieldByImmutableIdentity(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	option := "O_old"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/users/dragondad22/projectsV2/8/fields" {
			json.NewEncoder(response).Encode([]any{})
			return
		}
		if request.Method != http.MethodPost || request.URL.Path != "/graphql" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		var input struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.Contains(input.Query, "updateProjectV2ItemFieldValue"):
			if input.Variables["project"] != "P_project" || input.Variables["item"] != "ITEM_1" || input.Variables["field"] != "F_phase" || input.Variables["option"] != "O_phase_0" {
				t.Fatalf("immutable update variables = %#v", input.Variables)
			}
			option = "O_phase_0"
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"update": map[string]any{"projectV2Item": map[string]any{"id": "ITEM_1"}}}})
		case strings.Contains(input.Query, "fieldValues"):
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []map[string]any{{"id": "ITEM_1", "content": map[string]any{"id": "I_feature_1", "number": 1}, "fieldValues": map[string]any{"nodes": []map[string]any{{"optionId": option, "field": map[string]any{"id": "F_phase"}}}}}}}}}})
		default:
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"items": map[string]any{"nodes": []map[string]any{{"id": "ITEM_1", "content": map[string]any{"id": "I_feature_1"}}}}}}})
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_repo", ProjectID: "P_project", RepositoryName: "dragondad22/codex-starter-kit"}
	assignment := engine.SandboxResourceSpec{Key: "project-item-field:feature-1-phase", Kind: engine.SandboxResourceProjectItemField, Name: "Feature 1 Phase", Attributes: map[string]string{"content_id": "I_feature_1", "field": "Phase", "field_id": "F_phase", "option_id": "O_phase_0"}}
	config, providers := userProjectSandboxConfig(server, target, now, assignment)
	adapter, err := githubadapter.NewSandbox(config, providers, server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: assignment})
	if err != nil || result.Outcome != "applied" || result.ResourceID != "ITEM_1" || option != "O_phase_0" {
		t.Fatalf("assignment result = %#v, option=%s, err=%v", result, option, err)
	}
	replay, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: assignment})
	if err != nil || replay.Outcome != "no-change" {
		t.Fatalf("assignment replay = %#v, %v", replay, err)
	}
}

func TestSandboxAdapterObservesExactNativeRelationshipsAndMarkerOwnedFile(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75"
	content := marker + "\ndelivery fixture\n"
	issue := func(id int, number int, node string) map[string]any {
		return map[string]any{"id": id, "number": number, "node_id": node, "title": "fixture", "body": marker, "state": "open"}
	}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/orgs/labs/projectsV2/1/fields":
			json.NewEncoder(response).Encode([]any{})
		case "/graphql":
			json.NewEncoder(response).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"views": map[string]any{"nodes": []any{}}, "workflows": map[string]any{"nodes": []any{}}, "items": map[string]any{"nodes": []any{}, "pageInfo": map[string]any{"hasNextPage": false}}}}})
		case "/repos/labs/sandbox/issues/10":
			if request.Header.Get("Authorization") != "Bearer reconciler-token" {
				t.Fatalf("relationship authorization = %q", request.Header.Get("Authorization"))
			}
			json.NewEncoder(response).Encode(issue(100, 10, "I_parent"))
		case "/repos/labs/sandbox/issues/10/sub_issues":
			json.NewEncoder(response).Encode([]any{issue(101, 11, "I_child")})
		case "/repos/labs/sandbox/issues/12":
			json.NewEncoder(response).Encode(issue(102, 12, "I_blocker"))
		case "/repos/labs/sandbox/issues/13":
			json.NewEncoder(response).Encode(issue(103, 13, "I_dependent"))
		case "/repos/labs/sandbox/issues/13/dependencies/blocked_by":
			json.NewEncoder(response).Encode([]any{issue(102, 12, "I_blocker")})
		case "/repos/labs/sandbox/contents/.starter-kit/delivery-claim.txt":
			if request.Header.Get("Authorization") != "Bearer seeder-token" {
				t.Fatalf("file authorization = %q", request.Header.Get("Authorization"))
			}
			if request.URL.Query().Get("ref") != "main" {
				t.Fatalf("file ref = %q", request.URL.Query().Get("ref"))
			}
			json.NewEncoder(response).Encode(map[string]any{"sha": "blob-sha", "content": base64.StdEncoding.EncodeToString([]byte(content))})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	relationships := []engine.SandboxResourceSpec{
		relationshipResource("parent-sub-issue", marker, "10", "100", "I_parent", "11", "101", "I_child"),
		relationshipResource("blocker-dependent", marker, "12", "102", "I_blocker", "13", "103", "I_dependent"),
	}
	config.Resources = relationships
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleReconciler, sandboxProviders(now)[githubadapter.SandboxRoleReconciler], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	relationshipObservation, err := adapter.Observe(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	config.Resources = []engine.SandboxResourceSpec{{Key: "file:claim", Kind: engine.SandboxResourceRepositoryFile, Name: "delivery-claim.txt", Marker: marker, Attributes: map[string]string{"path": ".starter-kit/delivery-claim.txt", "branch": "main", "content_sha256": testSandboxSHA256(content), "input:content": content}}}
	adapter, err = githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	fileObservation, err := adapter.Observe(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	resources := append(fileObservation.Resources, relationshipObservation.Resources...)
	if len(resources) != 3 {
		t.Fatalf("resources = %#v", resources)
	}
	if resources[0].ID != "blob-sha" || resources[1].ID != "102:103" || resources[2].ID != "100:101" {
		t.Fatalf("stable native resource IDs = %#v", resources)
	}
}

func TestSandboxAdapterExactFixtureCleanupConvergesAfterCloseAndDelete(t *testing.T) {
	now := time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75"
	issueOpen, pullOpen, branchPresent := true, true, true
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/issues/20":
			state := "closed"
			if issueOpen {
				state = "open"
			}
			json.NewEncoder(response).Encode(map[string]any{"id": 200, "number": 20, "node_id": "I_issue", "body": marker, "state": state})
		case request.Method == http.MethodPatch && request.URL.Path == "/repos/labs/sandbox/issues/20":
			issueOpen = false
			response.WriteHeader(http.StatusOK)
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/pulls/21":
			state := "closed"
			if pullOpen {
				state = "open"
			}
			json.NewEncoder(response).Encode(map[string]any{"id": 201, "number": 21, "node_id": "PR_pull", "body": marker, "state": state, "head": map[string]any{"ref": "contract/run-75", "sha": "head-sha"}, "base": map[string]any{"ref": "main"}})
		case request.Method == http.MethodPatch && request.URL.Path == "/repos/labs/sandbox/pulls/21":
			pullOpen = false
			response.WriteHeader(http.StatusOK)
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/git/ref/heads/contract/run-75":
			if !branchPresent {
				http.NotFound(response, request)
				return
			}
			json.NewEncoder(response).Encode(map[string]any{"object": map[string]any{"sha": "head-sha"}})
		case request.Method == http.MethodDelete && request.URL.Path == "/repos/labs/sandbox/git/refs/heads/contract/run-75":
			branchPresent = false
			response.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected cleanup request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	config.Resources = []engine.SandboxResourceSpec{
		{Key: "cleanup:issue", Kind: engine.SandboxResourceFixtureIssue, Name: "issue", Marker: marker, DesiredState: engine.SandboxResourceAbsent, Attributes: map[string]string{"number": "20", "id": "200", "node_id": "I_issue", "state": "closed"}},
		{Key: "cleanup:pr", Kind: engine.SandboxResourceFixturePR, Name: "pull", Marker: marker, DesiredState: engine.SandboxResourceAbsent, Attributes: map[string]string{"number": "21", "id": "201", "node_id": "PR_pull", "state": "closed", "head": "contract/run-75", "base": "main", "head_sha": "head-sha"}},
		{Key: "cleanup:branch", Kind: engine.SandboxResourceFixtureBranch, Name: "contract/run-75", Marker: marker, DesiredState: engine.SandboxResourceAbsent, Attributes: map[string]string{"sha": "head-sha"}},
	}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	for _, resource := range config.Resources {
		result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: resource})
		if err != nil || result.Outcome != "applied" {
			t.Fatalf("cleanup %s = %#v, %v", resource.Key, result, err)
		}
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	if len(observation.Resources) != 0 {
		t.Fatalf("closed/deleted cleanup resources remained observed: %#v", observation.Resources)
	}
	for _, resource := range config.Resources {
		result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: resource})
		if err != nil || result.Outcome != "no-change" {
			t.Fatalf("cleanup replay %s = %#v, %v", resource.Key, result, err)
		}
	}
}

func TestSandboxAdapterFixtureIssueObservationEmitsExactIdentityHandoff(t *testing.T) {
	now := time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75:issue:parent"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/repos/labs/sandbox/issues" {
			t.Fatalf("unexpected identity request: %s %s", request.Method, request.URL.Path)
		}
		json.NewEncoder(response).Encode([]any{map[string]any{"id": 200, "number": 20, "node_id": "I_parent", "title": "parent", "body": marker, "state": "open"}})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	config.Resources = []engine.SandboxResourceSpec{{Key: "fixture:issue:parent", Kind: engine.SandboxResourceFixtureIssue, Name: "parent", Marker: marker, Attributes: map[string]string{"title": "parent", "state": "open"}}}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Resources) != 1 {
		t.Fatalf("observation = %#v, %v", observation, err)
	}
	attributes := observation.Resources[0].Attributes
	if attributes["number"] != "20" || attributes["id"] != "200" || attributes["node_id"] != "I_parent" {
		t.Fatalf("identity handoff = %#v", attributes)
	}
}

func TestSandboxAdapterReconcilesExactGovernedFixtureIssueBody(t *testing.T) {
	now := time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75:issue:delivery"
	body := "governed body\n\n" + marker
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/issues/20":
			json.NewEncoder(response).Encode(map[string]any{"id": 200, "number": 20, "node_id": "I_issue", "body": marker, "state": "open"})
		case request.Method == http.MethodPatch && request.URL.Path == "/repos/labs/sandbox/issues/20":
			var payload map[string]any
			json.NewDecoder(request.Body).Decode(&payload)
			if payload["body"] != body {
				t.Fatalf("governed body payload = %#v", payload)
			}
			json.NewEncoder(response).Encode(map[string]any{"id": 200, "number": 20, "node_id": "I_issue", "body": body, "state": "open"})
		default:
			t.Fatalf("unexpected governed issue request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	resource := engine.SandboxResourceSpec{Key: "fixture:issue:delivery", Kind: engine.SandboxResourceFixtureIssue, Name: "delivery", Marker: marker, Attributes: map[string]string{
		"title": "delivery", "state": "open", "number": "20", "id": "200", "node_id": "I_issue", "body_sha256": testSandboxSHA256(body), "input:body": body,
	}}
	config.Resources = []engine.SandboxResourceSpec{resource}
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: resource})
	if err != nil || result.Outcome != "applied" {
		t.Fatalf("governed issue result = %#v, %v", result, err)
	}
}

func TestSandboxAdapterExactFixtureCleanupRefusesIdentityAndHeadDrift(t *testing.T) {
	now := time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("cleanup drift triggered mutation: %s %s", request.Method, request.URL.Path)
		}
		switch request.URL.Path {
		case "/repos/labs/sandbox/issues/20":
			json.NewEncoder(response).Encode(map[string]any{"id": 200, "number": 20, "node_id": "I_replaced", "body": marker, "state": "open"})
		case "/repos/labs/sandbox/pulls/21":
			json.NewEncoder(response).Encode(map[string]any{"id": 201, "number": 21, "node_id": "PR_pull", "body": marker, "state": "open", "head": map[string]any{"ref": "contract/run-75", "sha": "changed-sha"}, "base": map[string]any{"ref": "main"}})
		case "/repos/labs/sandbox/git/ref/heads/contract/run-75":
			json.NewEncoder(response).Encode(map[string]any{"object": map[string]any{"sha": "changed-sha"}})
		default:
			t.Fatalf("unexpected drift request: %s", request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	resources := []engine.SandboxResourceSpec{
		{Key: "cleanup:issue", Kind: engine.SandboxResourceFixtureIssue, Name: "issue", Marker: marker, DesiredState: engine.SandboxResourceAbsent, Attributes: map[string]string{"number": "20", "id": "200", "node_id": "I_issue", "state": "closed"}},
		{Key: "cleanup:pr", Kind: engine.SandboxResourceFixturePR, Name: "pull", Marker: marker, DesiredState: engine.SandboxResourceAbsent, Attributes: map[string]string{"number": "21", "id": "201", "node_id": "PR_pull", "state": "closed", "head": "contract/run-75", "base": "main", "head_sha": "head-sha"}},
		{Key: "cleanup:branch", Kind: engine.SandboxResourceFixtureBranch, Name: "contract/run-75", Marker: marker, DesiredState: engine.SandboxResourceAbsent, Attributes: map[string]string{"sha": "head-sha"}},
	}
	for _, resource := range resources {
		result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: resource})
		if err != nil || result.Outcome != "needs-review" {
			t.Fatalf("drift cleanup %s = %#v, %v", resource.Key, result, err)
		}
	}
}

func TestSandboxAdapterAppliesAndSafelyRemovesExactNativeResources(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75"
	oldContent := marker + "\nold\n"
	newContent := marker + "\nnew\n"
	var relationshipCreated, relationshipDeleted, fileUpdated, fileDeleted bool
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/issues/10":
			json.NewEncoder(response).Encode(map[string]any{"id": 100, "number": 10, "node_id": "I_parent", "body": marker})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/issues/11":
			json.NewEncoder(response).Encode(map[string]any{"id": 101, "number": 11, "node_id": "I_child", "body": marker})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/labs/sandbox/issues/10/sub_issues":
			relationshipCreated = true
			response.WriteHeader(http.StatusCreated)
		case request.Method == http.MethodDelete && request.URL.Path == "/repos/labs/sandbox/issues/10/sub_issue":
			var body map[string]any
			json.NewDecoder(request.Body).Decode(&body)
			if body["sub_issue_id"] != float64(101) {
				t.Fatalf("relationship delete body = %#v", body)
			}
			relationshipDeleted = true
			response.WriteHeader(http.StatusNoContent)
		case request.Method == http.MethodGet && request.URL.Path == "/repos/labs/sandbox/contents/.starter-kit/delivery-claim.txt":
			json.NewEncoder(response).Encode(map[string]any{"sha": "old-sha", "content": base64.StdEncoding.EncodeToString([]byte(oldContent))})
		case request.Method == http.MethodPut && request.URL.Path == "/repos/labs/sandbox/contents/.starter-kit/delivery-claim.txt":
			var body map[string]any
			json.NewDecoder(request.Body).Decode(&body)
			if body["sha"] != "old-sha" || body["branch"] != "contract/run-75" {
				t.Fatalf("file update body = %#v", body)
			}
			fileUpdated = true
			json.NewEncoder(response).Encode(map[string]any{"content": map[string]any{"sha": "new-sha"}})
		case request.Method == http.MethodDelete && request.URL.Path == "/repos/labs/sandbox/contents/.starter-kit/delivery-claim.txt":
			fileDeleted = true
			response.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleReconciler, sandboxProviders(now)[githubadapter.SandboxRoleReconciler], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	relationship := relationshipResource("parent-sub-issue", marker, "10", "100", "I_parent", "11", "101", "I_child")
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: relationship})
	if err != nil || result.Outcome != "applied" || !relationshipCreated {
		t.Fatalf("relationship create result = %#v, created=%v, err=%v", result, relationshipCreated, err)
	}
	result, err = adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: relationship})
	if err != nil || result.Outcome != "applied" || !relationshipDeleted {
		t.Fatalf("relationship result = %#v, deleted=%v, err=%v", result, relationshipDeleted, err)
	}
	file := engine.SandboxResourceSpec{Key: "file:claim", Kind: engine.SandboxResourceRepositoryFile, Name: "delivery-claim.txt", Marker: marker, Attributes: map[string]string{"path": ".starter-kit/delivery-claim.txt", "branch": "contract/run-75", "content_sha256": testSandboxSHA256(newContent), "input:content": newContent}}
	adapter, err = githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(now)[githubadapter.SandboxRoleSeeder], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	result, err = adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "reconcile-resource", Resource: file})
	if err != nil || result.ResourceID != "new-sha" || !fileUpdated {
		t.Fatalf("file result = %#v, updated=%v, err=%v", result, fileUpdated, err)
	}
	deleteFile := file
	deleteFile.Attributes = maps.Clone(file.Attributes)
	deleteFile.Attributes["content_sha256"] = testSandboxSHA256(oldContent)
	result, err = adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: deleteFile})
	if err != nil || result.Outcome != "applied" || !fileDeleted {
		t.Fatalf("file delete result = %#v, deleted=%v, err=%v", result, fileDeleted, err)
	}
}

func TestSandboxAdapterDoesNotDeleteUnownedRepositoryFile(t *testing.T) {
	marker := "starter-kit-contract:run-75"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unowned file triggered mutation: %s", request.Method)
		}
		json.NewEncoder(response).Encode(map[string]any{"sha": "human-sha", "content": base64.StdEncoding.EncodeToString([]byte(marker + "\nhuman edit\n"))})
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleSeeder, sandboxProviders(time.Now())[githubadapter.SandboxRoleSeeder], server.Client())
	if err != nil {
		t.Fatal(err)
	}
	resource := engine.SandboxResourceSpec{Key: "file:claim", Kind: engine.SandboxResourceRepositoryFile, Name: "claim", Marker: marker, Attributes: map[string]string{"path": "claim.txt", "branch": "main", "content_sha256": testSandboxSHA256(marker), "input:content": marker}}
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: resource})
	if err != nil || result.Outcome != "needs-review" {
		t.Fatalf("result = %#v, err=%v", result, err)
	}
}

func TestSandboxAdapterDoesNotMutateUnownedIssueRelationship(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	marker := "starter-kit-contract:run-75"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unowned relationship triggered mutation: %s", request.Method)
		}
		switch request.URL.Path {
		case "/repos/labs/sandbox/issues/10":
			json.NewEncoder(response).Encode(map[string]any{"id": 100, "number": 10, "node_id": "I_parent", "body": "human-owned"})
		case "/repos/labs/sandbox/issues/11":
			json.NewEncoder(response).Encode(map[string]any{"id": 101, "number": 11, "node_id": "I_child", "body": marker})
		default:
			t.Fatalf("unexpected request: %s", request.URL.Path)
		}
	}))
	defer server.Close()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "owner-id", RepositoryID: "repo-id", ProjectID: "project-id", RepositoryName: "labs/sandbox"}
	config := sandboxConfig(server, target)
	adapter, err := githubadapter.NewSandboxRole(config, githubadapter.SandboxRoleReconciler, sandboxProviders(now)[githubadapter.SandboxRoleReconciler], server.Client(), githubadapter.WithSandboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	resource := relationshipResource("parent-sub-issue", marker, "10", "100", "I_parent", "11", "101", "I_child")
	result, err := adapter.Apply(context.Background(), engine.SandboxEffect{Kind: "remove-resource", Resource: resource})
	if err != nil || result.Outcome != "needs-review" {
		t.Fatalf("result = %#v, err=%v", result, err)
	}
}

func relationshipResource(relationship, marker, sourceNumber, sourceID, sourceNodeID, targetNumber, targetID, targetNodeID string) engine.SandboxResourceSpec {
	return engine.SandboxResourceSpec{Key: "relationship:" + relationship, Kind: engine.SandboxResourceIssueRelationship, Name: relationship, Marker: marker, Attributes: map[string]string{
		"relationship": relationship, "source_number": sourceNumber, "source_id": sourceID, "source_node_id": sourceNodeID, "target_number": targetNumber, "target_id": targetID, "target_node_id": targetNodeID,
	}}
}

func testSandboxSHA256(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func userProjectSandboxConfig(server *httptest.Server, target engine.SandboxTarget, now time.Time, resources ...engine.SandboxResourceSpec) (githubadapter.SandboxConfig, map[string]githubadapter.CredentialProvider) {
	expectation := githubadapter.SandboxRoleExpectation{Mode: "user-token", Actor: "dragondad22", Account: "dragondad22", AccountID: "19365745", RequiredPermissions: []string{"projects:write"}}
	config := githubadapter.SandboxConfig{Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10", ConfigurationRevision: "phase-config-v1", Target: target, RepositoryOwner: "dragondad22", RepositoryName: "codex-starter-kit", ProjectNumber: 8, ProjectOwnerKind: "user", EvidenceMode: "simulated", Resources: resources, Roles: map[string]githubadapter.SandboxRoleExpectation{githubadapter.SandboxRoleReconciler: expectation}}
	provider := githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		return githubadapter.Credential{Token: "token", Mode: "user-token", Actor: "dragondad22", Account: "dragondad22", AccountID: "19365745", Permissions: []string{"projects:write"}, ExpiresAt: now.Add(time.Hour)}, nil
	})
	return config, map[string]githubadapter.CredentialProvider{githubadapter.SandboxRoleReconciler: provider}
}

func sandboxConfig(server *httptest.Server, target engine.SandboxTarget) githubadapter.SandboxConfig {
	return githubadapter.SandboxConfig{
		Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10",
		ConfigurationRevision: "config-1", Target: target, RepositoryOwner: "labs", RepositoryName: "sandbox", ProjectNumber: 1,
		EvidenceMode: "simulated", Resources: []engine.SandboxResourceSpec{{Key: "label:type-task", Kind: engine.SandboxResourceLabel, Name: "type:task"}},
		Roles: map[string]githubadapter.SandboxRoleExpectation{
			githubadapter.SandboxRoleReconciler: {Mode: "app-installation", Actor: "reconciler", Account: "labs", AccountID: "owner-id", InstallationID: "1", RequiredPermissions: []string{"issues:write", "organization-projects:write"}},
			githubadapter.SandboxRoleSeeder:     {Mode: "app-installation", Actor: "seeder", Account: "labs", AccountID: "owner-id", InstallationID: "2", RequiredPermissions: []string{"contents:write", "workflows:write"}},
			githubadapter.SandboxRoleRules:      {Mode: "app-installation", Actor: "rules", Account: "labs", AccountID: "owner-id", InstallationID: "3", RequiredPermissions: []string{"administration:write"}},
		},
	}
}

func sandboxProviders(now time.Time) map[string]githubadapter.CredentialProvider {
	permissions := map[string][]string{
		githubadapter.SandboxRoleReconciler: {"issues:write", "organization-projects:write"},
		githubadapter.SandboxRoleSeeder:     {"contents:write", "workflows:write"},
		githubadapter.SandboxRoleRules:      {"administration:write"},
	}
	providers := map[string]githubadapter.CredentialProvider{}
	for role, rolePermissions := range permissions {
		role := role
		rolePermissions := rolePermissions
		providers[role] = githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
			installation := map[string]string{githubadapter.SandboxRoleReconciler: "1", githubadapter.SandboxRoleSeeder: "2", githubadapter.SandboxRoleRules: "3"}[role]
			return githubadapter.Credential{Token: role + "-token", Mode: "app-installation", Actor: role, Account: "labs", AccountID: "owner-id", InstallationID: installation, Permissions: rolePermissions, PermissionSource: "test", PermissionRevision: "permissions-1", ExpiresAt: now.Add(time.Hour)}, nil
		})
	}
	return providers
}
