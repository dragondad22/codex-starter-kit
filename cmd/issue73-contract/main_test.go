package main

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestIssue73MandateNamesAdapterRolesAndImmutableTarget(t *testing.T) {
	target := engine.SandboxTarget{Host: "github.com", OwnerID: ownerID, RepositoryID: repositoryID, ProjectID: projectID, RepositoryName: repository}
	resources, expectation, _, _, err := rolePlan("rules-proof", githubadapter.SandboxRoleSeeder, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	mandate := issue73Mandate(target, resources, issue73Authority(githubadapter.SandboxRoleSeeder, expectation))
	for _, actor := range []string{"reconciler", "seeder", "rules", "reviewer", "american-dragon-designs"} {
		if !slices.Contains(mandate.Actors, actor) {
			t.Fatalf("mandate actors %v omit %q", mandate.Actors, actor)
		}
	}
	if mandate.ID == "" || mandate.Target != target || mandate.DataClass != "public-synthetic" || mandate.CostCeiling != "zero-dollar" || mandate.MarkerPrefix != runMarker || len(mandate.ResourceDigests) != len(resources) || !slices.Equal(mandate.Authority.Permissions, []string{"seeder:contents:write", "seeder:issues:write", "seeder:metadata:read", "seeder:pull-requests:write", "seeder:workflows:write"}) {
		t.Fatalf("mandate = %#v", mandate)
	}
}

func TestQualificationPlansKeepRevocationLast(t *testing.T) {
	for _, role := range []string{githubadapter.SandboxRoleSeeder, githubadapter.SandboxRoleRules} {
		resources, _, _, _, err := rolePlan("qualification", role, "", "", "")
		if err != nil {
			t.Fatalf("%s qualification plan: %v", role, err)
		}
		if len(resources) < 2 || resources[len(resources)-1].Kind != engine.SandboxResourceTokenRevocation {
			t.Fatalf("%s resources do not revoke last: %#v", role, resources)
		}
		target := engine.SandboxTarget{Host: "github.com", OwnerID: ownerID, RepositoryID: repositoryID, ProjectID: projectID, RepositoryName: repository}
		manifest := engine.SandboxManifest{SchemaVersion: 1, OperationID: "test", SourceRevision: "source", ConfigurationRevision: configuration, ApprovedBy: "owner", ApprovedPlan: "approved", RecoveryOwner: "owner", MarkerPrefix: markerPrefix, Target: target, Resources: resources}
		adapter := engine.NewInMemorySandboxAdapter(engine.SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Target: target}, engine.SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: configuration})
		repositoryPath := t.TempDir()
		if err := os.MkdirAll(repositoryPath+"/.git", 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := engine.New(engine.WithSandboxAdapter(adapter)).InspectSandbox(context.Background(), engine.SandboxRequest{Repository: repositoryPath, Manifest: manifest}); err != nil {
			t.Fatalf("%s qualification manifest: %v", role, err)
		}
	}
}

func TestProjectProofIncludesPositiveNegativeAndCloseAutomationCases(t *testing.T) {
	resources, _, _, _, err := rolePlan("project-proof", githubadapter.SandboxRoleReconciler, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 3 || resources[0].DesiredState == engine.SandboxResourceAbsent || resources[1].DesiredState != engine.SandboxResourceAbsent || resources[2].Attributes["status"] != "Done" {
		t.Fatalf("project proofs = %#v", resources)
	}
}

func TestCleanupClosesRetainedRecordsAndDeletesEphemeralResources(t *testing.T) {
	resources := cleanupSeederResources()
	states := map[string]string{}
	lastPullRequest := -1
	firstBranch := len(resources)
	for index, resource := range resources {
		states[resource.Kind] = resource.DesiredState
		if resource.Kind == engine.SandboxResourceFixtureIssue || resource.Kind == engine.SandboxResourceFixturePR {
			if resource.DesiredState == engine.SandboxResourceAbsent || resource.Attributes["state"] != "closed" {
				t.Fatalf("retained fixture is not closed: %#v", resource)
			}
		}
		if resource.Kind == engine.SandboxResourceFixturePR {
			lastPullRequest = index
		}
		if resource.Kind == engine.SandboxResourceFixtureBranch && firstBranch == len(resources) {
			firstBranch = index
		}
	}
	if states[engine.SandboxResourceFixtureBranch] != engine.SandboxResourceAbsent || states[engine.SandboxResourceFixtureWorkflow] != engine.SandboxResourceAbsent {
		t.Fatalf("ephemeral cleanup states = %#v", states)
	}
	if lastPullRequest < 0 || firstBranch <= lastPullRequest {
		t.Fatalf("cleanup must close pull requests before deleting their branches: %#v", resources)
	}
}
