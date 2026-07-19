package main

import (
	"testing"

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
