package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestPhaseResourcesBindExactCatalogViewAndFeatureAssignments(t *testing.T) {
	resources := phaseResources()
	if len(resources) != 20 {
		t.Fatalf("resource count = %d", len(resources))
	}
	if resources[0].Kind != engine.SandboxResourceProjectField || resources[10].Kind != engine.SandboxResourceProjectView || resources[19].Attributes["option_id"] != "6d252c8e" {
		t.Fatalf("Phase resource contract = %#v", resources)
	}
	if resources[10].Attributes["group_by"] != "PVTSSF_lAHOASd_cc4BdI9qzhYRk9k" || resources[10].Attributes["sort_by"] != "PVTSSF_lAHOASd_cc4BdI9qzhYRk9k:asc" {
		t.Fatalf("Phases view contract = %#v", resources[10])
	}
}

func TestIssue46CommandRejectsAMandateThatDoesNotBindTheExactResourcesBeforeTransport(t *testing.T) {
	resources := phaseResources()
	target := engine.SandboxTarget{Host: "github.com", OwnerID: "19365745", RepositoryID: "R_kgDOTVs5Hg", ProjectID: "PVT_kwHOASd_cc4BdI9q", RepositoryName: "dragondad22/codex-starter-kit"}
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, time.UTC)
	authority := engine.SandboxAuthorityProfile{CredentialIdentities: []string{"reconciler|user-token|dragondad22|dragondad22|19365745|"}, Permissions: []string{"reconciler:classic-scope:project", "reconciler:projects:write"}, EvidenceMode: "live", Compatibility: "github.com:api.github.com:2026-03-10:native-rest-graphql", DataClass: "public-project-metadata", CostCeiling: "zero-dollar", Destructive: "no-delete-no-overwrite-human-view", Retention: "30-days"}
	mandate := engine.BindSandboxExecutionMandate(engine.SandboxExecutionMandate{SchemaVersion: 1, ApprovedBy: "owner", ApprovalID: "owner-record", ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Target: target, Actors: []string{"reconciler"}, MarkerPrefix: "starter-kit-contract:issue-46", UnmarkedKeys: resourceKeys(resources), ResourceKinds: []string{engine.SandboxResourceProjectField, engine.SandboxResourceProjectOption, engine.SandboxResourceProjectView, engine.SandboxResourceProjectItemField}, EffectKinds: []string{"reconcile-resource"}, MaxEffects: len(resources), DataClass: authority.DataClass, CostCeiling: authority.CostCeiling, Destructive: authority.Destructive, Retention: authority.Retention, RecoveryOwner: "owner", Authority: authority}, resources...)
	mandate.ResourceDigests[0] = "sha256:forged"
	encoded, err := json.Marshal(mandate)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "mandate.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	err = run(context.Background(), []string{"--stage", "plan", "--repository", t.TempDir(), "--source-revision", "source", "--observed-at", "2026-07-19T18:05:19Z", "--mandate", path})
	if err == nil || !strings.Contains(err.Error(), "exact Phase resource digests") {
		t.Fatalf("forged mandate error = %v", err)
	}
}

func TestIssue46CommandRequiresAnIndependentlyRetainedMandateArtifact(t *testing.T) {
	t.Setenv(tokenEnvironment, "unused")
	err := run(context.Background(), []string{"--stage", "apply", "--repository", t.TempDir(), "--source-revision", "source", "--observed-at", "2026-07-19T18:05:19Z", "--expected-plan-id", "plan"})
	if err == nil || !strings.Contains(err.Error(), "--mandate") {
		t.Fatalf("apply error = %v", err)
	}
	for _, manufactured := range []string{"--approval-id", "--approved-at", "--expires-at"} {
		err = run(context.Background(), []string{"--stage", "apply", "--repository", t.TempDir(), "--source-revision", "source", "--observed-at", "2026-07-19T18:05:19Z", manufactured, "caller-value"})
		if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
			t.Fatalf("manufactured approval flag %s error = %v", manufactured, err)
		}
	}
}
