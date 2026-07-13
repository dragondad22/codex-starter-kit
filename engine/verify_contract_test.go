package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequiredBreadcrumbCannotPassWhenMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".starter-kit"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"routes":{"artifact:conformance":"docs/evidence/CONFORMANCE.md"}}`
	if err := os.WriteFile(filepath.Join(root, ".starter-kit", "routes.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result := evaluateRoutes(root)
	if result.State == ControlPass {
		t.Fatal("missing required breadcrumbs produced pass")
	}
}

func TestDiagnosticsAreRedactedBeforeEvidence(t *testing.T) {
	diagnostics := redactDiagnostics([]string{
		"token=super-secret-value",
		"-----BEGIN PRIVATE KEY-----",
	})
	joined := strings.Join(diagnostics, " ")
	if strings.Contains(joined, "super-secret-value") || strings.Contains(joined, "PRIVATE KEY") {
		t.Fatalf("sensitive diagnostic survived redaction: %q", joined)
	}
}

func TestOverallStateFixturesNeverProducePass(t *testing.T) {
	fixtures := []ControlResult{
		{ID: "FAIL", State: ControlFail, Rationale: "fixture"},
		{ID: "NA", State: ControlNotApplicable, Rationale: "fixture", Evidence: []EvidenceReference{{Kind: "fact", Target: "fixture"}}},
		{ID: "MISSING", State: ControlNotConfigured, Rationale: "fixture"},
		{ID: "REVIEW", State: ControlNeedsReview, Rationale: "fixture"},
		{ID: "EXCEPTION", State: ControlAcceptedException, UnderlyingState: ControlFail, Rationale: "fixture"},
	}
	for _, fixture := range fixtures {
		if state := OverallControlState([]ControlResult{fixture}); state == ControlPass {
			t.Fatalf("%s fixture produced pass", fixture.State)
		}
	}
}
