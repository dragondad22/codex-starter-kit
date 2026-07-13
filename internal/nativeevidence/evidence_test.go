package nativeevidence

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestCaptureExercisesCompletePhase1Seam(t *testing.T) {
	setCaptureProvenance(t)
	report, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture native evidence: %v", err)
	}
	if report.SchemaVersion != 1 || !validSourceRevision(report.SourceRevision) || report.Platform.GOOS == "" || report.Platform.GOARCH == "" || report.Platform.GitVersion == "" {
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
	if len(report.Capabilities) != 7 {
		t.Fatalf("platform capabilities were not disclosed: %#v", report.Capabilities)
	}
	atomicReplacement := capabilityByID(report.Capabilities, "same-directory-atomic-replacement")
	if atomicReplacement == nil || atomicReplacement.State == "" || !strings.Contains(atomicReplacement.Details, "existing") {
		t.Fatalf("atomic replacement capability is not explicit: %#v", report.Capabilities)
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
	setCaptureProvenance(t)
	report, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture fixture: %v", err)
	}
	directory := t.TempDir()
	for index, goos := range []string{"linux", "darwin", "windows"} {
		copy := hostedReport(report, goos)
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
	if !summary.Equivalent || len(summary.Platforms) != 3 || len(summary.ReportReferences) != 3 || summary.SourceRevision != report.SourceRevision || summary.SemanticDigest != report.SemanticDigest || summary.EvidenceDigest != comparisonDigest(summary) {
		t.Fatalf("unexpected comparison summary: %#v", summary)
	}
	for _, reference := range summary.ReportReferences {
		if reference.SourceRevision != report.SourceRevision || reference.SemanticDigest != report.SemanticDigest || reference.EvidenceDigest == "" {
			t.Fatalf("comparison does not bind its native reports: %#v", reference)
		}
	}

	windowsPath := filepath.Join(directory, "windows", EvidenceFilename)
	drifted := hostedReport(report, "windows")
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

func TestReadRejectsIncompleteEvidenceContracts(t *testing.T) {
	setCaptureProvenance(t)
	report, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture fixture: %v", err)
	}
	base := hostedReport(report, "linux")
	tests := map[string]func(Report) Report{
		"missing tested revision": func(candidate Report) Report {
			candidate.SourceRevision = ""
			return candidate
		},
		"missing tool provenance": func(candidate Report) Report {
			candidate.Platform.GitVersion = ""
			return candidate
		},
		"missing capability": func(candidate Report) Report {
			candidate.Capabilities = candidate.Capabilities[1:]
			return candidate
		},
		"invalid capability state": func(candidate Report) Report {
			candidate.Capabilities[0].State = "pass"
			return candidate
		},
		"duplicate control": func(candidate Report) Report {
			candidate.Semantics.Controls[len(candidate.Semantics.Controls)-1] = candidate.Semantics.Controls[0]
			candidate.SemanticDigest = semanticDigest(candidate.Semantics)
			return candidate
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := mutate(base)
			candidate.EvidenceDigest = evidenceDigest(candidate)
			path := filepath.Join(t.TempDir(), EvidenceFilename)
			if err := Write(path, candidate); err != nil {
				t.Fatalf("write malformed report: %v", err)
			}
			if _, err := Read(path); err == nil {
				t.Fatal("read accepted a self-consistent but incomplete report")
			}
		})
	}
}

func TestCompareRejectsSourceRevisionDrift(t *testing.T) {
	setCaptureProvenance(t)
	report, err := Capture(context.Background())
	if err != nil {
		t.Fatalf("capture fixture: %v", err)
	}
	directory := t.TempDir()
	for _, goos := range []string{"linux", "darwin", "windows"} {
		candidate := hostedReport(report, goos)
		if goos == "windows" {
			candidate.SourceRevision = strings.Repeat("a", 40)
		}
		candidate.EvidenceDigest = evidenceDigest(candidate)
		path := filepath.Join(directory, goos, EvidenceFilename)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create report directory: %v", err)
		}
		if err := Write(path, candidate); err != nil {
			t.Fatalf("write report: %v", err)
		}
	}
	if _, err := Compare(directory); err == nil {
		t.Fatal("comparison accepted source revision drift")
	}
}

func TestTestedSourceRevisionRejectsDirtyLocalWorktree(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITHUB_SHA", "")
	directory := t.TempDir()
	runGit(t, directory, "init", "--quiet")
	tracked := filepath.Join(directory, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("clean\n"), 0o600); err != nil {
		t.Fatalf("write tracked fixture: %v", err)
	}
	runGit(t, directory, "add", "tracked.txt")
	runGit(t, directory, "-c", "user.name=Phase 1 Evidence", "-c", "user.email=evidence@example.invalid", "commit", "--quiet", "-m", "fixture")
	revision, err := testedSourceRevision(context.Background(), directory)
	if err != nil || !validSourceRevision(revision) {
		t.Fatalf("resolve clean local revision: %q, %v", revision, err)
	}
	if err := os.WriteFile(tracked, []byte("dirty\n"), 0o600); err != nil {
		t.Fatalf("dirty tracked fixture: %v", err)
	}
	if _, err := testedSourceRevision(context.Background(), directory); err == nil {
		t.Fatal("local source provenance accepted a dirty tracked worktree")
	}
	runGit(t, directory, "checkout", "--quiet", "--", "tracked.txt")
	if err := os.WriteFile(filepath.Join(directory, "untracked.txt"), []byte("dirty\n"), 0o600); err != nil {
		t.Fatalf("write untracked fixture: %v", err)
	}
	if _, err := testedSourceRevision(context.Background(), directory); err == nil {
		t.Fatal("local source provenance accepted an untracked file")
	}
}

func hostedReport(report Report, goos string) Report {
	copy := report
	copy.Capabilities = append([]Capability{}, report.Capabilities...)
	copy.Semantics.Operations = append([]string{}, report.Semantics.Operations...)
	copy.Semantics.Artifacts = append([]ArtifactContract{}, report.Semantics.Artifacts...)
	copy.Semantics.CoverageLimitations = append([]string{}, report.Semantics.CoverageLimitations...)
	copy.Semantics.Controls = append([]ControlSemantic{}, report.Semantics.Controls...)
	for index := range copy.Semantics.Controls {
		copy.Semantics.Controls[index].Evidence = append([]engine.EvidenceReference{}, report.Semantics.Controls[index].Evidence...)
	}
	copy.Platform.GOOS = goos
	copy.Platform.GOARCH = map[string]string{"darwin": "arm64", "linux": "amd64", "windows": "amd64"}[goos]
	copy.Platform.RunnerOS = map[string]string{"darwin": "macOS", "linux": "Linux", "windows": "Windows"}[goos]
	copy.Platform.RunnerArch = map[string]string{"darwin": "ARM64", "linux": "X64", "windows": "X64"}[goos]
	copy.Platform.ImageOS = map[string]string{"darwin": "macos26", "linux": "ubuntu24", "windows": "win25-vs2026"}[goos]
	copy.Platform.ImageVersion = "fixture"
	return copy
}

func capabilityByID(capabilities []Capability, id string) *Capability {
	for index := range capabilities {
		if capabilities[index].ID == id {
			return &capabilities[index]
		}
	}
	return nil
}

func setCaptureProvenance(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_SHA", strings.Repeat("f", 40))
	t.Setenv("RUNNER_OS", map[string]string{"darwin": "macOS", "linux": "Linux", "windows": "Windows"}[runtime.GOOS])
	t.Setenv("RUNNER_ARCH", map[string]string{"amd64": "X64", "arm64": "ARM64"}[runtime.GOARCH])
	t.Setenv("ImageOS", "test-image")
	t.Setenv("ImageVersion", "test-version")
}

func runGit(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
