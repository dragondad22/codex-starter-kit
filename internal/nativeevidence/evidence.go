// Package nativeevidence exercises the Phase 1 lifecycle seam and records the
// portable semantics separately from host-specific capabilities.
package nativeevidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	SourceRevision string       `json:"source_revision"`
	Platform       Platform     `json:"platform"`
	Capabilities   []Capability `json:"capabilities"`
	Semantics      Semantics    `json:"semantics"`
	SemanticDigest string       `json:"semantic_digest"`
	EvidenceDigest string       `json:"evidence_digest"`
}

type ReportReference struct {
	Platform       string `json:"platform"`
	SourceRevision string `json:"source_revision"`
	SemanticDigest string `json:"semantic_digest"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Comparison struct {
	SchemaVersion    int               `json:"schema_version"`
	Ownership        string            `json:"ownership"`
	Source           string            `json:"source"`
	SourceRevision   string            `json:"source_revision"`
	Equivalent       bool              `json:"equivalent"`
	Platforms        []string          `json:"platforms"`
	SemanticDigest   string            `json:"semantic_digest"`
	ReportReferences []ReportReference `json:"report_references"`
	EvidenceDigest   string            `json:"evidence_digest"`
}

type fixedClock struct{ value time.Time }

func (clock fixedClock) Now() time.Time { return clock.value }

func Capture(ctx context.Context) (Report, error) {
	sourceRevision, err := testedSourceRevision(ctx)
	if err != nil {
		return Report{}, err
	}
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
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "phase1-evidence:capture:v1", SourceRevision: sourceRevision,
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
	if err := validateReport(report, false); err != nil {
		return Report{}, fmt.Errorf("captured native evidence is invalid: %w", err)
	}
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
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return Report{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Report{}, errors.New("native evidence report contains trailing JSON")
	}
	if err := validateReport(report, true); err != nil {
		return Report{}, fmt.Errorf("native evidence report is invalid: %w", err)
	}
	return report, nil
}

func Valid(report Report) bool {
	return validateReport(report, false) == nil
}

func validateReport(report Report, requireHostedProvenance bool) error {
	if report.SchemaVersion != 1 || report.Ownership != "machine-evidence" || report.Source != "phase1-evidence:capture:v1" {
		return errors.New("report envelope does not match the native evidence schema")
	}
	if !validSourceRevision(report.SourceRevision) {
		return errors.New("tested source revision is missing or invalid")
	}
	expectedArchitectures := map[string]string{"darwin": "arm64", "linux": "amd64", "windows": "amd64"}
	if expectedArchitectures[report.Platform.GOOS] != report.Platform.GOARCH {
		return fmt.Errorf("unsupported native platform %s/%s", report.Platform.GOOS, report.Platform.GOARCH)
	}
	if report.Platform.GoVersion == "" || report.Platform.GitVersion == "" || report.Platform.FilesystemAssumption == "" {
		return errors.New("tool or filesystem provenance is incomplete")
	}
	if requireHostedProvenance && (report.Platform.RunnerOS == "" || report.Platform.RunnerArch == "" || report.Platform.ImageOS == "" || report.Platform.ImageVersion == "") {
		return errors.New("hosted runner image or architecture provenance is incomplete")
	}
	if err := validateCapabilities(report.Capabilities); err != nil {
		return err
	}
	if err := validateSemantics(report.Semantics); err != nil {
		return err
	}
	if report.SemanticDigest != semanticDigest(report.Semantics) {
		return errors.New("semantic digest does not match the semantic contract")
	}
	if report.EvidenceDigest != evidenceDigest(report) {
		return errors.New("evidence digest does not match the complete report")
	}
	return nil
}

func validateCapabilities(capabilities []Capability) error {
	required := map[string]bool{
		"case-behavior": true, "directory-junction": true, "file-symlink": true,
		"native-path-separator": true, "owner-only-mode": true,
		"portable-lf-managed-content": true, "same-directory-atomic-replacement": true,
	}
	allowedStates := map[string]bool{
		"supported": true, "not-configured": true, "not-applicable": true,
		"needs-review": true, "degraded": true, "unsupported": true,
	}
	seen := map[string]bool{}
	for _, capability := range capabilities {
		if !required[capability.ID] || seen[capability.ID] {
			return fmt.Errorf("unexpected or duplicate native capability %q", capability.ID)
		}
		if !allowedStates[capability.State] || capability.Details == "" {
			return fmt.Errorf("native capability %s has an invalid state or empty details", capability.ID)
		}
		seen[capability.ID] = true
	}
	if len(seen) != len(required) {
		return fmt.Errorf("native capability set is incomplete: found %d, want %d", len(seen), len(required))
	}
	return nil
}

func validateSemantics(semantics Semantics) error {
	if !equalStrings(semantics.Operations, []string{"inspect", "create", "plan", "apply", "status", "verify"}) ||
		semantics.PlanSchemaVersion < 1 || semantics.PlanOperation != "create" || !semantics.PlanStable ||
		semantics.ApplyStatus != "applied" || semantics.ReplayStatus != "applied" || semantics.NoChangeStatus != "no_change" ||
		semantics.Lifecycle != "managed" || !semantics.Managed || semantics.LineEndings != "lf" {
		return errors.New("lifecycle operation or state semantics are incomplete")
	}
	if semantics.EvidenceSchema < 1 || semantics.EvidenceOwnership != "machine-evidence" || semantics.EvidenceSource == "" ||
		semantics.EngineVersion == "" || semantics.RepositorySchema < 1 || semantics.PolicyVersion == "" ||
		semantics.SourceRevisionKind == "missing" || semantics.OverallControlState != "needs-review" || len(semantics.CoverageLimitations) == 0 {
		return errors.New("verification evidence semantics are incomplete")
	}
	seenArtifacts := map[string]bool{}
	for _, artifact := range semantics.Artifacts {
		if artifact.Path == "" || artifact.Ownership == "" || artifact.Source == "" || seenArtifacts[artifact.Path] {
			return errors.New("managed artifact contract is incomplete or duplicated")
		}
		seenArtifacts[artifact.Path] = true
	}
	if len(seenArtifacts) == 0 {
		return errors.New("managed artifact contract is empty")
	}
	expectedStates := map[string]string{
		"CORE-COVERAGE-001": "pass", "CORE-OWNERSHIP-001": "pass",
		"CORE-RECOVERY-001": "needs-review", "CORE-ROUTES-001": "pass",
		"CORE-SECRETS-001": "not-configured", "CORE-TRUTH-001": "pass",
	}
	seenControls := map[string]bool{}
	for _, control := range semantics.Controls {
		if expectedStates[control.ID] == "" || seenControls[control.ID] || control.State != expectedStates[control.ID] {
			return fmt.Errorf("control %s is unexpected, duplicated, or has the wrong state", control.ID)
		}
		if control.UnderlyingState != "" {
			return fmt.Errorf("control %s has an unexpected underlying state", control.ID)
		}
		if control.State == "pass" && len(control.Evidence) == 0 {
			return fmt.Errorf("passing control %s lacks evidence", control.ID)
		}
		for _, reference := range control.Evidence {
			if reference.Kind == "" || reference.Target == "" {
				return fmt.Errorf("control %s has an incomplete evidence reference", control.ID)
			}
		}
		seenControls[control.ID] = true
	}
	if len(seenControls) != len(expectedStates) {
		return fmt.Errorf("seed control set is incomplete: found %d, want %d", len(seenControls), len(expectedStates))
	}
	for _, limitation := range semantics.CoverageLimitations {
		if limitation == "" {
			return errors.New("coverage limitations contain an empty entry")
		}
	}
	return nil
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
	references := []ReportReference{}
	seen := map[string]bool{}
	semanticDigest := ""
	sourceRevision := ""
	for _, path := range paths {
		report, err := Read(path)
		if err != nil {
			return Comparison{}, fmt.Errorf("read %s: %w", path, err)
		}
		if seen[report.Platform.GOOS] {
			return Comparison{}, fmt.Errorf("duplicate native evidence for %s", report.Platform.GOOS)
		}
		seen[report.Platform.GOOS] = true
		platform := report.Platform.GOOS + "/" + report.Platform.GOARCH
		platforms = append(platforms, platform)
		references = append(references, ReportReference{
			Platform: platform, SourceRevision: report.SourceRevision,
			SemanticDigest: report.SemanticDigest, EvidenceDigest: report.EvidenceDigest,
		})
		if sourceRevision == "" {
			sourceRevision = report.SourceRevision
		} else if report.SourceRevision != sourceRevision {
			return Comparison{}, fmt.Errorf("native source revision drift: %s has %s, expected %s", report.Platform.GOOS, report.SourceRevision, sourceRevision)
		}
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
	sort.Slice(references, func(i, j int) bool { return references[i].Platform < references[j].Platform })
	summary := Comparison{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "phase1-evidence:compare:v1",
		SourceRevision: sourceRevision, Equivalent: true, Platforms: platforms,
		SemanticDigest: semanticDigest, ReportReferences: references,
	}
	summary.EvidenceDigest = comparisonDigest(summary)
	return summary, nil
}

func nativeCapabilities(root string) []Capability {
	capabilities := []Capability{
		probeCaseBehavior(root),
		probeSymlink(root),
		probeOwnerOnlyMode(root),
		probeAtomicReplacement(root),
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

func probeAtomicReplacement(root string) Capability {
	directory := filepath.Join(root, ".git", "phase1-replacement-probe")
	_ = os.MkdirAll(directory, 0o700)
	source := filepath.Join(directory, "source")
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(source, []byte("replacement"), 0o600); err != nil {
		return Capability{ID: "same-directory-atomic-replacement", State: "needs-review", Details: "replacement source could not be created"}
	}
	if err := os.WriteFile(target, []byte("original"), 0o600); err != nil {
		return Capability{ID: "same-directory-atomic-replacement", State: "needs-review", Details: "existing replacement target could not be created"}
	}
	if err := os.Rename(source, target); err != nil {
		return Capability{ID: "same-directory-atomic-replacement", State: "unsupported", Details: "native rename could not replace an existing same-directory destination"}
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "replacement" {
		return Capability{ID: "same-directory-atomic-replacement", State: "degraded", Details: "existing destination replacement did not expose the complete staged content"}
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		return Capability{ID: "same-directory-atomic-replacement", State: "degraded", Details: "replacement left the staged source visible"}
	}
	return Capability{ID: "same-directory-atomic-replacement", State: "supported", Details: "native rename replaced an existing same-directory destination with complete staged content"}
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

func testedSourceRevision(ctx context.Context) (string, error) {
	if revision := strings.TrimSpace(os.Getenv("GITHUB_SHA")); revision != "" {
		if !validSourceRevision(revision) {
			return "", errors.New("GITHUB_SHA is not a complete hexadecimal source revision")
		}
		return strings.ToLower(revision), nil
	}
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "", errors.New("GitHub-hosted native evidence lacks GITHUB_SHA source provenance")
	}
	revision, err := exec.CommandContext(ctx, "git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("resolve tested source revision: %w", err)
	}
	value := strings.TrimSpace(string(revision))
	if !validSourceRevision(value) {
		return "", errors.New("resolved tested source revision is not a complete hexadecimal revision")
	}
	return strings.ToLower(value), nil
}

func validSourceRevision(revision string) bool {
	if len(revision) != 40 && len(revision) != 64 {
		return false
	}
	for _, character := range revision {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
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
