package githubadapter_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			githubadapter.SandboxRoleReconciler: {Mode: "app-installation", Actor: "reconciler", Account: "labs", InstallationID: "1", RequiredPermissions: []string{"issues:write", "organization-projects:write"}},
			githubadapter.SandboxRoleSeeder:     {Mode: "app-installation", Actor: "seeder", Account: "labs", InstallationID: "2", RequiredPermissions: []string{"contents:write", "workflows:write"}},
			githubadapter.SandboxRoleRules:      {Mode: "app-installation", Actor: "rules", Account: "labs", InstallationID: "3", RequiredPermissions: []string{"administration:write"}},
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
			return githubadapter.Credential{Token: role + "-token", Mode: "app-installation", Actor: role, Account: "labs", InstallationID: installation, Permissions: rolePermissions, PermissionSource: "test", PermissionRevision: "permissions-1", ExpiresAt: now.Add(time.Hour)}, nil
		})
	}
	return providers
}
