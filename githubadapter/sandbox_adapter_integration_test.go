package githubadapter_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete || request.URL.Path != "/repos/labs/sandbox/git/refs/heads/contract/run/cleanup" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		response.WriteHeader(http.StatusForbidden)
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
	if err != nil || result.Outcome != "applied" || result.ResourceID != "http-403" {
		t.Fatalf("apply = %#v, %v", result, err)
	}
	observation, err := adapter.Observe(context.Background(), target)
	if err != nil || len(observation.Resources) != 1 || observation.Resources[0].Key != proof.Key {
		t.Fatalf("observation = %#v, %v", observation, err)
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
