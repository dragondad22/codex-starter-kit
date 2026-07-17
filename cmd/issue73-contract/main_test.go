package main

import (
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestQualificationPlansKeepRevocationLast(t *testing.T) {
	for _, role := range []string{githubadapter.SandboxRoleSeeder, githubadapter.SandboxRoleRules} {
		resources, _, _, _, err := rolePlan("qualification", role, "", "", "")
		if err != nil {
			t.Fatalf("%s qualification plan: %v", role, err)
		}
		if len(resources) < 2 || resources[len(resources)-1].Kind != engine.SandboxResourceTokenRevocation {
			t.Fatalf("%s resources do not revoke last: %#v", role, resources)
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
	for _, resource := range resources {
		states[resource.Kind] = resource.DesiredState
		if resource.Kind == engine.SandboxResourceFixtureIssue || resource.Kind == engine.SandboxResourceFixturePR {
			if resource.DesiredState == engine.SandboxResourceAbsent || resource.Attributes["state"] != "closed" {
				t.Fatalf("retained fixture is not closed: %#v", resource)
			}
		}
	}
	if states[engine.SandboxResourceFixtureBranch] != engine.SandboxResourceAbsent || states[engine.SandboxResourceFixtureWorkflow] != engine.SandboxResourceAbsent {
		t.Fatalf("ephemeral cleanup states = %#v", states)
	}
}
