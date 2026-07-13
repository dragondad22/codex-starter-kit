// Package nativeevidence exercises the Phase 1 lifecycle seam and records the
// portable semantics separately from host-specific capabilities.
package nativeevidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

const EvidenceFilename = "phase1-native-evidence.json"

type Platform struct {
	GOOS                 string `json:"goos"`
	GOARCH               string `json:"goarch"`
	GoVersion            string `json:"go_version"`
	GitVersion           string `json:"git_version"`
	RunnerOS             string `json:"runner_os"`
	RunnerArch           string `json:"runner_arch"`
	ImageOS              string `json:"image_os"`
	ImageVersion         string `json:"image_version"`
	FilesystemAssumption string `json:"filesystem_assumption"`
}

type Capability struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Details string `json:"details"`
}

type ArtifactContract struct {
	Path      string `json:"path"`
	Ownership string `json:"ownership"`
	Source    string `json:"source"`
}

type ControlSemantic struct {
	ID              string                     `json:"id"`
	State           string                     `json:"state"`
	UnderlyingState string                     `json:"underlying_state,omitempty"`
	Evidence        []engine.EvidenceReference `json:"evidence"`
}

type Semantics struct {
	Operations          []string           `json:"operations"`
	PlanSchemaVersion   int                `json:"plan_schema_version"`
	PlanOperation       string             `json:"plan_operation"`
	PlanStable          bool               `json:"plan_stable"`
	Artifacts           []ArtifactContract `json:"artifacts"`
	ApplyStatus         string             `json:"apply_status"`
	ReplayStatus        string             `json:"replay_status"`
	NoChangeStatus      string             `json:"no_change_status"`
	Lifecycle           string             `json:"lifecycle"`
	Managed             bool               `json:"managed"`
	LineEndings         string             `json:"line_endings"`
	EvidenceSchema      int                `json:"evidence_schema_version"`
	EvidenceOwnership   string             `json:"evidence_ownership"`
	EvidenceSource      string             `json:"evidence_source"`
	EngineVersion       string             `json:"engine_version"`
	RepositorySchema    int                `json:"repository_schema_version"`
	PolicyVersion       string             `json:"policy_version"`
	SourceRevisionKind  string             `json:"source_revision_kind"`
	OverallControlState string             `json:"overall_control_state"`
	Controls            []ControlSemantic  `json:"controls"`
	CoverageLimitations []string           `json:"coverage_limitations"`
}

type Report struct {
	SchemaVersion  int          `json:"schema_version"`
	Ownership      string       `json:"ownership"`
	Source         string       `json:"source"`
	Platform       Platform     `json:"platform"`
	Capabilities   []Capability `json:"capabilities"`
	Semantics      Semantics    `json:"semantics"`
	SemanticDigest string       `json:"semantic_digest"`
	EvidenceDigest string       `json:"evidence_digest"`
}

type Comparison struct {
	SchemaVersion  int      `json:"schema_version"`
	Ownership      string   `json:"ownership"`
	Source         string   `json:"source"`
	Equivalent     bool     `json:"equivalent"`
	Platforms      []string `json:"platforms"`
	SemanticDigest string   `json:"semantic_digest"`
	EvidenceFiles  []string `json:"evidence_files"`
	EvidenceDigest string   `json:"evidence_digest"`
}

type fixedClock struct{ value time.Time }

func (clock fixedClock) Now() time.Time { return clock.value }

func Capture(ctx context.Context) (Report, error) {
	root, err := os.MkdirTemp("", "starter-kit-phase1-evidence-")
	if err != nil {
		return Report{}, err
	}
	defer os.RemoveAll(root)
	if output, err := exec.CommandContext(ctx, "git", "init", "--quiet", root).CombinedOutput(); err != nil {
		return Report{}, fmt.Errorf("initialize evidence repository: %w: %s", err, output)
	}
	gitVersion, err := exec.CommandContext(ctx, "git", "--version").Output()
	if err != nil {
		return Report{}, fmt.Errorf("read Git version: %w", err)
	}

	lifecycle := engine.New(engine.WithClock(fixedClock{time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)}))
	initial, err := lifecycle.Inspect(ctx, root)
	if err != nil {
		return Report{}, fmt.Errorf("inspect empty native repository: %w", err)
	}
	if !initial.Git || initial.Managed {
		return Report{}, errors.New("empty native evidence repository has unexpected Git or lifecycle state")
	}
	request := engine.CreateRequest{
		Repository: root, Brief: "Native Phase 1 semantic-equivalence evidence fixture.",
		BriefApproved: true, OwnerPersonaConfirmed: true,
	}
	createPlan, err := lifecycle.Create(ctx, request)
	if err != nil {
		return Report{}, fmt.Errorf("create native plan: %w", err)
	}
	explicitPlan, err := lifecycle.Plan(ctx, engine.PlanRequest{Operation: engine.CreateOperation, Create: request})
	if err != nil {
		return Report{}, fmt.Errorf("plan native create: %w", err)
	}
	applyResult, err := lifecycle.Apply(ctx, createPlan.ID, createPlan)
	if err != nil {
		return Report{}, fmt.Errorf("apply native create: %w", err)
	}
	replayResult, err := lifecycle.Apply(ctx, createPlan.ID, createPlan)
	if err != nil {
		return Report{}, fmt.Errorf("replay native create: %w", err)
	}
	status, err := lifecycle.Status(ctx, root)
	if err != nil {
		return Report{}, fmt.Errorf("status native repository: %w", err)
	}
	noChangePlan, err := lifecycle.Create(ctx, request)
	if err != nil {
		return Report{}, fmt.Errorf("plan native no-change create: %w", err)
	}
	noChangeResult, err := lifecycle.Apply(ctx, noChangePlan.ID, noChangePlan)
	if err != nil {
		return Report{}, fmt.Errorf("apply native no-change create: %w", err)
	}
	verifyPlan, err := lifecycle.PrepareVerify(ctx, engine.VerifyRequest{
		Repository: root, Scope: "repository", Gate: "development",
		Actor: "phase1-native-evidence", Authority: "issue-30-native-equivalence",
	})
	if err != nil {
		return Report{}, fmt.Errorf("plan native verification: %w", err)
	}
	verification, err := lifecycle.Verify(ctx, verifyPlan.ID, verifyPlan)
	if err != nil {
		return Report{}, fmt.Errorf("verify native repository: %w", err)
	}
	finalInspection, err := lifecycle.Inspect(ctx, root)
	if err != nil {
		return Report{}, fmt.Errorf("inspect verified native repository: %w", err)
	}

	brief, err := os.ReadFile(filepath.Join(root, "docs", "product", "BRIEF.md"))
	if err != nil {
		return Report{}, fmt.Errorf("read generated brief: %w", err)
	}
	lineEndings := "lf"
	if strings.Contains(string(brief), "\r\n") {
		lineEndings = "crlf"
	}
	artifacts := make([]ArtifactContract, 0, len(createPlan.Files))
	for _, file := range createPlan.Files {
		artifacts = append(artifacts, ArtifactContract{file.Path, file.Ownership, file.Source})
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	controls := make([]ControlSemantic, 0, len(verification.Controls))
	for _, control := range verification.Controls {
		evidence := append([]engine.EvidenceReference{}, control.Evidence...)
		sort.Slice(evidence, func(i, j int) bool {
			if evidence[i].Kind == evidence[j].Kind {
				return evidence[i].Target < evidence[j].Target
			}
			return evidence[i].Kind < evidence[j].Kind
		})
		controls = append(controls, ControlSemantic{control.ID, string(control.State), string(control.UnderlyingState), evidence})
	}
	sort.Slice(controls, func(i, j int) bool { return controls[i].ID < controls[j].ID })
	semantics := Semantics{
		Operations:        []string{"inspect", "create", "plan", "apply", "status", "verify"},
		PlanSchemaVersion: createPlan.SchemaVersion, PlanOperation: string(createPlan.Operation),
		PlanStable: createPlan.ID == explicitPlan.ID, Artifacts: artifacts,
		ApplyStatus: string(applyResult.Status), ReplayStatus: string(replayResult.Status),
		NoChangeStatus: string(noChangeResult.Status), Lifecycle: status.Lifecycle,
		Managed: finalInspection.Managed, LineEndings: lineEndings,
		EvidenceSchema: verification.SchemaVersion, EvidenceOwnership: verification.Ownership,
		EvidenceSource: verification.Source, EngineVersion: verification.EngineVersion,
		RepositorySchema: verification.RepositorySchemaVersion, PolicyVersion: verification.PolicyVersion,
		SourceRevisionKind:  sourceRevisionKind(verification.SourceRevision),
		OverallControlState: string(verification.OverallState), Controls: controls,
		CoverageLimitations: append([]string{}, verification.CoverageLimitations...),
	}
	report := Report{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "phase1-evidence:capture:v1",
		Platform: Platform{
			GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(),
			GitVersion: strings.TrimSpace(string(gitVersion)), RunnerOS: os.Getenv("RUNNER_OS"),
			RunnerArch: os.Getenv("RUNNER_ARCH"), ImageOS: os.Getenv("ImageOS"),
			ImageVersion:         os.Getenv("ImageVersion"),
			FilesystemAssumption: "native filesystem backing os.TempDir; behavior is probed, filesystem brand is not inferred",
		},
		Capabilities: nativeCapabilities(root), Semantics: semantics,
	}
	if os.Getenv("GITHUB_ACTIONS") == "true" && (report.Platform.RunnerOS == "" || report.Platform.RunnerArch == "" || report.Platform.ImageOS == "" || report.Platform.ImageVersion == "") {
		return Report{}, errors.New("GitHub-hosted native evidence lacks resolved runner image or architecture provenance")
	}
	report.SemanticDigest = semanticDigest(report.Semantics)
	report.EvidenceDigest = evidenceDigest(report)
	return report, nil
}

func Write(path string, report Report) error {
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func Read(path string) (Report, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	var report Report
	if err := json.Unmarshal(content, &report); err != nil {
		return Report{}, err
	}
	if !Valid(report) {
		return Report{}, errors.New("native evidence report digest or schema is invalid")
	}
	return report, nil
}

func Valid(report Report) bool {
	return report.SchemaVersion == 1 && report.Ownership == "machine-evidence" && report.Source == "phase1-evidence:capture:v1" &&
		report.SemanticDigest == semanticDigest(report.Semantics) && report.EvidenceDigest == evidenceDigest(report)
}

func Compare(directory string) (Comparison, error) {
	paths := []string{}
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && entry.Name() == EvidenceFilename {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return Comparison{}, err
	}
	sort.Strings(paths)
	if len(paths) != 3 {
		return Comparison{}, fmt.Errorf("expected three native evidence reports, found %d", len(paths))
	}
	platforms := []string{}
	seen := map[string]bool{}
	semanticDigest := ""
	for _, path := range paths {
		report, err := Read(path)
		if err != nil {
			return Comparison{}, fmt.Errorf("read %s: %w", path, err)
		}
		if seen[report.Platform.GOOS] {
			return Comparison{}, fmt.Errorf("duplicate native evidence for %s", report.Platform.GOOS)
		}
		seen[report.Platform.GOOS] = true
		platforms = append(platforms, report.Platform.GOOS+"/"+report.Platform.GOARCH)
		if semanticDigest == "" {
			semanticDigest = report.SemanticDigest
		} else if report.SemanticDigest != semanticDigest {
			return Comparison{}, fmt.Errorf("native semantic drift: %s has %s, expected %s", report.Platform.GOOS, report.SemanticDigest, semanticDigest)
		}
	}
	for _, required := range []string{"darwin", "linux", "windows"} {
		if !seen[required] {
			return Comparison{}, fmt.Errorf("native evidence is missing %s", required)
		}
	}
	sort.Strings(platforms)
	summary := Comparison{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "phase1-evidence:compare:v1",
		Equivalent: true, Platforms: platforms, SemanticDigest: semanticDigest, EvidenceFiles: paths,
	}
	summary.EvidenceDigest = comparisonDigest(summary)
	return summary, nil
}

func nativeCapabilities(root string) []Capability {
	capabilities := []Capability{
		probeCaseBehavior(root),
		probeSymlink(root),
		probeOwnerOnlyMode(root),
		probeRename(root),
		{ID: "native-path-separator", State: "supported", Details: fmt.Sprintf("separator byte %d", os.PathSeparator)},
		{ID: "portable-lf-managed-content", State: "supported", Details: "managed text is emitted with LF and semantic readers do not translate it"},
		probeJunction(root),
	}
	sort.Slice(capabilities, func(i, j int) bool { return capabilities[i].ID < capabilities[j].ID })
	return capabilities
}

func probeCaseBehavior(root string) Capability {
	directory := filepath.Join(root, ".git", "phase1-case-probe")
	_ = os.MkdirAll(directory, 0o700)
	lower := filepath.Join(directory, "case-sensitive")
	upper := filepath.Join(directory, "CASE-SENSITIVE")
	if err := os.WriteFile(lower, []byte("probe"), 0o600); err != nil {
		return Capability{ID: "case-behavior", State: "needs-review", Details: "case probe could not be written"}
	}
	if _, err := os.Stat(upper); err == nil {
		return Capability{ID: "case-behavior", State: "supported", Details: "case-insensitive lookup observed; portable collision policy remains authoritative"}
	}
	return Capability{ID: "case-behavior", State: "supported", Details: "case-sensitive lookup observed; portable collision policy remains authoritative"}
}

func probeSymlink(root string) Capability {
	directory := filepath.Join(root, ".git", "phase1-symlink-probe")
	_ = os.MkdirAll(directory, 0o700)
	target := filepath.Join(directory, "target")
	link := filepath.Join(directory, "link")
	if err := os.WriteFile(target, []byte("probe"), 0o600); err != nil {
		return Capability{ID: "file-symlink", State: "needs-review", Details: "symlink target could not be created"}
	}
	if err := os.Symlink(target, link); err != nil {
		return Capability{ID: "file-symlink", State: "not-configured", Details: "native runner did not grant file-symlink creation"}
	}
	return Capability{ID: "file-symlink", State: "supported", Details: "native file-symlink creation and rejection fixtures are available"}
}

func probeOwnerOnlyMode(root string) Capability {
	path := filepath.Join(root, ".git", "phase1-mode-probe")
	if err := os.WriteFile(path, []byte("probe"), 0o600); err != nil {
		return Capability{ID: "owner-only-mode", State: "needs-review", Details: "mode probe could not be created"}
	}
	info, err := os.Stat(path)
	if err != nil {
		return Capability{ID: "owner-only-mode", State: "needs-review", Details: "mode probe could not be inspected"}
	}
	if runtime.GOOS == "windows" {
		return Capability{ID: "owner-only-mode", State: "not-applicable", Details: "POSIX mode bits do not establish Windows ACL assurance"}
	}
	if info.Mode().Perm()&0o077 != 0 {
		return Capability{ID: "owner-only-mode", State: "degraded", Details: "requested owner-only mode was broadened by the native filesystem"}
	}
	return Capability{ID: "owner-only-mode", State: "supported", Details: "requested 0600 mode remained owner-only"}
}

func probeRename(root string) Capability {
	directory := filepath.Join(root, ".git", "phase1-rename-probe")
	_ = os.MkdirAll(directory, 0o700)
	source := filepath.Join(directory, "source")
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(source, []byte("probe"), 0o600); err != nil {
		return Capability{ID: "same-directory-rename", State: "needs-review", Details: "rename source could not be created"}
	}
	if err := os.Rename(source, target); err != nil {
		return Capability{ID: "same-directory-rename", State: "unsupported", Details: "same-directory staged rename failed"}
	}
	return Capability{ID: "same-directory-rename", State: "supported", Details: "same-directory staged rename succeeded"}
}

func semanticDigest(semantics Semantics) string { return digestJSON(semantics) }

func sourceRevisionKind(revision string) string {
	if strings.HasPrefix(revision, "sha256:") {
		return "working-tree-snapshot-digest"
	}
	if revision != "" {
		return "git-revision"
	}
	return "missing"
}

func evidenceDigest(report Report) string {
	report.EvidenceDigest = ""
	return digestJSON(report)
}

func comparisonDigest(summary Comparison) string {
	summary.EvidenceDigest = ""
	return digestJSON(summary)
}

func digestJSON(value any) string {
	content, _ := json.Marshal(value)
	digest := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
