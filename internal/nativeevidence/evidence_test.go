package nativeevidence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureExercisesCompletePhase1Seam(t *testing.T) {
	report, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture native evidence: %v", err)
	}
	if report.SchemaVersion != 1 || report.Platform.GOOS == "" || report.Platform.GOARCH == "" || report.Platform.GitVersion == "" {
		t.Fatalf("platform provenance incomplete: %#v", report.Platform)
	}
	wantOperations := []string{"inspect", "create", "plan", "apply", "status", "verify"}
	if !equalStrings(report.Semantics.Operations, wantOperations) {
		t.Fatalf("operations = %#v, want %#v", report.Semantics.Operations, wantOperations)
	}
	if report.Semantics.ApplyStatus != "applied" || report.Semantics.ReplayStatus != "applied" || report.Semantics.NoChangeStatus != "no_change" || report.Semantics.Lifecycle != "managed" {
		t.Fatalf("lifecycle semantics incomplete: %#v", report.Semantics)
	}
	if report.Semantics.OverallControlState != "needs-review" || len(report.Semantics.Controls) != 6 {
		t.Fatalf("verification semantics incomplete: %#v", report.Semantics)
	}
	if report.SemanticDigest == "" || report.EvidenceDigest == "" || !Valid(report) {
		t.Fatalf("report lacks valid content identity: %#v", report)
	}
	if len(report.Capabilities) < 6 {
		t.Fatalf("platform capabilities were not disclosed: %#v", report.Capabilities)
	}
	second, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture equivalent repository: %v", err)
	}
	if second.SemanticDigest != report.SemanticDigest {
		t.Fatalf("temporary repository identity leaked into portable semantics: %s != %s", second.SemanticDigest, report.SemanticDigest)
	}
}

func TestCompareRequiresThreeEquivalentNativeReports(t *testing.T) {
	report, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture fixture: %v", err)
	}
	directory := t.TempDir()
	for index, goos := range []string{"linux", "darwin", "windows"} {
		copy := report
		copy.Platform.GOOS = goos
		copy.Platform.GOARCH = "amd64"
		copy.EvidenceDigest = evidenceDigest(copy)
		path := filepath.Join(directory, goos, EvidenceFilename)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create report directory: %v", err)
		}
		if err := Write(path, copy); err != nil {
			t.Fatalf("write report %d: %v", index, err)
		}
	}

	summary, err := Compare(directory)
	if err != nil {
		t.Fatalf("compare equivalent reports: %v", err)
	}
	if !summary.Equivalent || len(summary.Platforms) != 3 || summary.SemanticDigest != report.SemanticDigest || summary.EvidenceDigest != comparisonDigest(summary) {
		t.Fatalf("unexpected comparison summary: %#v", summary)
	}

	windowsPath := filepath.Join(directory, "windows", EvidenceFilename)
	drifted := report
	drifted.Platform.GOOS = "windows"
	drifted.Semantics.Lifecycle = "managed_degraded"
	drifted.SemanticDigest = semanticDigest(drifted.Semantics)
	drifted.EvidenceDigest = evidenceDigest(drifted)
	if err := Write(windowsPath, drifted); err != nil {
		t.Fatalf("write drifted report: %v", err)
	}
	if _, err := Compare(directory); err == nil {
		t.Fatal("comparison accepted platform semantic drift")
	}
}
