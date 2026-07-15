package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	starterkit "github.com/dragondad22/codex-starter-kit"
)

// ControlState is one explicit control-evaluation state.
type ControlState string

const (
	ControlPass              ControlState = "pass"
	ControlFail              ControlState = "fail"
	ControlNotApplicable     ControlState = "not-applicable"
	ControlNotConfigured     ControlState = "not-configured"
	ControlNeedsReview       ControlState = "needs-review"
	ControlAcceptedException ControlState = "accepted-exception"
)

// EvidenceReference points from a control result to current supporting evidence.
type EvidenceReference struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
}

// ControlResult records one stable control identity and its truthful state.
type ControlResult struct {
	ID              string              `json:"id"`
	State           ControlState        `json:"state"`
	UnderlyingState ControlState        `json:"underlying_state,omitempty"`
	Summary         string              `json:"summary"`
	Rationale       string              `json:"rationale,omitempty"`
	Evidence        []EvidenceReference `json:"evidence"`
	Diagnostics     []string            `json:"diagnostics"`
}

// VerifyRequest identifies the explicit verification scope and lifecycle gate.
type VerifyRequest struct {
	Repository string `json:"repository"`
	Scope      string `json:"scope"`
	Gate       string `json:"gate"`
	Actor      string `json:"actor"`
	Authority  string `json:"authority"`
}

// VerifyPlan is an immutable, reviewable verification transaction.
type VerifyPlan struct {
	SchemaVersion    int    `json:"schema_version"`
	ID               string `json:"plan_id"`
	Repository       string `json:"repository"`
	RepositoryDigest string `json:"repository_digest"`
	Scope            string `json:"scope"`
	Gate             string `json:"gate"`
	Actor            string `json:"actor"`
	Authority        string `json:"authority"`
}

// VerificationResult is the machine-readable conformance evidence manifest.
type VerificationResult struct {
	SchemaVersion           int             `json:"schema_version"`
	VerificationID          string          `json:"verification_id"`
	EvidenceDigest          string          `json:"evidence_digest"`
	Ownership               string          `json:"ownership"`
	Source                  string          `json:"source"`
	Scope                   string          `json:"scope"`
	Gate                    string          `json:"gate"`
	SourceRevision          string          `json:"source_revision"`
	SourceSnapshotDigest    string          `json:"source_snapshot_digest"`
	EngineVersion           string          `json:"engine_version"`
	RepositorySchemaVersion int             `json:"repository_schema_version"`
	PolicyVersion           string          `json:"policy_version"`
	VerifiedAt              time.Time       `json:"verified_at"`
	OverallState            ControlState    `json:"overall_state"`
	Controls                []ControlResult `json:"controls"`
	CoverageLimitations     []string        `json:"coverage_limitations"`
	EvidencePath            string          `json:"evidence_path"`
	EventPath               string          `json:"event_path"`
	Actor                   string          `json:"actor"`
	Authority               string          `json:"authority"`
}

// Verify evaluates seed controls, persists machine evidence, and regenerates the human
// conformance summary without converting missing capability into pass.
// PrepareVerify captures verification inputs and repository preconditions without mutation.
func (e *Engine) PrepareVerify(ctx context.Context, request VerifyRequest) (VerifyPlan, error) {
	if request.Repository == "" || request.Scope == "" || request.Gate == "" || request.Actor == "" || request.Authority == "" {
		return VerifyPlan{}, errors.New("verify requires repository, scope, gate, actor, and authority")
	}
	if containsSensitiveText(strings.Join([]string{request.Scope, request.Gate, request.Actor, request.Authority}, "\n")) {
		return VerifyPlan{}, errors.New("verification metadata contains sensitive-looking material")
	}
	root, err := cleanRepositoryRoot(request.Repository)
	if err != nil {
		return VerifyPlan{}, err
	}
	inspection, err := e.Inspect(ctx, root)
	if err != nil {
		return VerifyPlan{}, err
	}
	plan := VerifyPlan{SchemaVersion: 1, Repository: root, RepositoryDigest: inspection.SnapshotDigest, Scope: request.Scope, Gate: request.Gate, Actor: request.Actor, Authority: request.Authority}
	plan.ID = digestJSON(plan)
	return plan, nil

}

// Verify evaluates an immutable verification plan and persists its evidence transaction.
func (e *Engine) Verify(ctx context.Context, expectedPlanID string, plan VerifyPlan) (verification VerificationResult, verifyErr error) {
	recordedID := plan.ID
	plan.ID = ""
	if plan.SchemaVersion != 1 || recordedID == "" || expectedPlanID != recordedID || digestJSON(plan) != recordedID {
		return VerificationResult{}, errors.New("verification plan identity is invalid")
	}
	plan.ID = recordedID
	if !validSHA256Digest(plan.RepositoryDigest) {
		return VerificationResult{}, errors.New("verification repository digest is invalid")
	}
	if containsSensitiveText(strings.Join([]string{plan.Scope, plan.Gate, plan.Actor, plan.Authority}, "\n")) {
		return VerificationResult{}, errors.New("verification plan metadata contains sensitive-looking material")
	}
	root, err := cleanRepositoryRoot(plan.Repository)
	if err != nil {
		return VerificationResult{}, err
	}
	if root != plan.Repository {
		return VerificationResult{}, errors.New("verification repository path is not canonical")
	}
	lockPath, err := lifecycleLockPath(ctx, root)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() {
		if verifyErr != nil {
			if recordErr := recordVerificationFailure(lockPath, plan, verifyErr); recordErr != nil {
				verifyErr = fmt.Errorf("%w; recording verification failure: %v", verifyErr, recordErr)
			}
		}
	}()
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return VerificationResult{}, fmt.Errorf("acquire verification lock: %w", err)
	}
	_ = lock.Close()
	defer os.Remove(lockPath)

	inspection, err := e.Inspect(ctx, root)
	if err != nil {
		return VerificationResult{}, err
	}
	if inspection.SnapshotDigest != plan.RepositoryDigest {
		return VerificationResult{}, errors.New("verification plan precondition no longer matches repository content")
	}
	controls := evaluateSeedControls(root, inspection)
	verifiedAt := e.clock.Now().UTC()
	sourceRevision := inspection.GitHead
	if sourceRevision == "" {
		sourceRevision = inspection.SnapshotDigest
	}
	result := VerificationResult{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "engine:verify:v1",
		Scope: plan.Scope, Gate: plan.Gate, SourceRevision: sourceRevision,
		SourceSnapshotDigest: inspection.SnapshotDigest, EngineVersion: starterkit.Version(),
		RepositorySchemaVersion: 1, PolicyVersion: "not-configured", VerifiedAt: verifiedAt,
		Actor: plan.Actor, Authority: plan.Authority,
		Controls: controls, OverallState: OverallControlState(controls),
		CoverageLimitations: []string{
			"No approved secret-scanning capability is configured.",
			"Create recovery compensates multi-path interruption; it does not claim atomic external effects.",
			"Runtime support is limited to source-built Phase 1 create/verify on the published native matrix; no packaged release or broader lifecycle operation is covered.",
		},
	}
	identity := struct {
		Scope, Gate, Revision, Snapshot string
		Actor, Authority                string
		VerifiedAt                      time.Time
		Controls                        []ControlResult
	}{result.Scope, result.Gate, result.SourceRevision, result.SourceSnapshotDigest, result.Actor, result.Authority, result.VerifiedAt, result.Controls}
	result.VerificationID = digestJSON(identity)
	result.EvidencePath = ".starter-kit/evidence/verify-" + strings.TrimPrefix(result.VerificationID, "sha256:") + ".json"
	result.EventPath = ".starter-kit/events/verify-" + strings.TrimPrefix(result.VerificationID, "sha256:") + ".json"
	result.EvidenceDigest = verificationDigest(result)
	if err := persistVerification(root, plan.ID, result); err != nil {
		return VerificationResult{}, err
	}
	return result, nil
}

func recordVerificationFailure(lockPath string, plan VerifyPlan, failure error) error {
	event := operationEvent{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "engine:verify:v1",
		PlanID: plan.ID, Operation: VerifyOperation, Status: ApplyStatusFailed,
		Actor: plan.Actor, Authority: plan.Authority, RepositoryDigest: plan.RepositoryDigest,
		ChangedFiles: []string{}, ExternalEffects: []string{},
		Diagnostics: redactDiagnostics([]string{failure.Error()}),
		Conflicts:   []ReconciliationConflict{}, Recovery: []string{}, Evidence: []string{}, Recoverable: true,
	}
	event.EventDigest = digestJSON(event)
	directory := filepath.Join(filepath.Dir(lockPath), "starter-kit-attempts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	path := filepath.Join(directory, "verify-"+strings.TrimPrefix(plan.ID, "sha256:")+".json")
	content := []byte(jsonDocument(event))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err == nil {
		if _, writeErr := file.Write(content); writeErr != nil {
			_ = file.Close()
			return writeErr
		}
		return file.Close()
	}
	if !errors.Is(err, os.ErrExist) {
		return err
	}
	existing, readErr := os.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	if string(existing) != string(content) {
		return errors.New("verification attempt path already contains different evidence")
	}
	return nil
}

// OverallControlState aggregates without converting any non-pass state into pass.
func OverallControlState(results []ControlResult) ControlState {
	priority := []ControlState{
		ControlFail, ControlNeedsReview, ControlNotConfigured,
		ControlAcceptedException, ControlNotApplicable,
	}
	for _, state := range priority {
		for _, result := range results {
			if result.State == state {
				return state
			}
		}
	}
	if len(results) == 0 {
		return ControlNotConfigured
	}
	return ControlPass
}

func evaluateSeedControls(root string, inspection Inspection) []ControlResult {
	results := []ControlResult{
		{
			ID: "CORE-TRUTH-001", State: ControlPass,
			Summary:  "Every seed control has an explicit state and passing states cite evidence.",
			Evidence: []EvidenceReference{}, Diagnostics: []string{},
		},
		{
			ID: "CORE-SECRETS-001", State: ControlNotConfigured,
			Summary:   "Secret protection cannot be verified because no approved scanner is configured.",
			Rationale: "Issue #27 does not authorize a partial scanner to claim repository coverage.",
			Evidence:  []EvidenceReference{}, Diagnostics: []string{},
		},
		{
			ID: "CORE-OWNERSHIP-001", State: ControlPass,
			Summary:  "Required artifacts match the ownership and provenance manifest.",
			Evidence: []EvidenceReference{{Kind: "repository", Target: ".starter-kit/managed-files.json"}}, Diagnostics: []string{},
		},
		{
			ID: "CORE-COVERAGE-001", State: ControlPass,
			Summary:  "Verification discloses evaluated controls and explicit coverage limitations.",
			Evidence: []EvidenceReference{{Kind: "inline", Target: "coverage_limitations"}}, Diagnostics: []string{},
		},
		{
			ID: "CORE-RECOVERY-001", State: ControlNeedsReview,
			Summary: "The create-v1 recovery protocol is implemented, but this verification run cannot bind " +
				"the executing binary to retained native test provenance.",
			Rationale: "An unversioned source build cannot bind itself to retained native-equivalence evidence; a future versioned release must provide that provenance before recovery can pass.",
			Evidence:  []EvidenceReference{}, Diagnostics: []string{},
		},
		evaluateRoutes(root),
	}
	if !inspection.Managed {
		for index := range results {
			if results[index].ID == "CORE-OWNERSHIP-001" {
				results[index].State = ControlFail
				results[index].Summary = "The managed-repository ownership contract is invalid."
				results[index].Rationale = "Inspection found that the managed artifact contract does not match current repository content."
				results[index].Evidence = []EvidenceReference{}
				results[index].Diagnostics = redactDiagnostics(inspection.Problems)
			}
		}
	}
	truthInputs := struct {
		Supported []ControlState
		Controls  []ControlResult
	}{
		Supported: []ControlState{ControlPass, ControlFail, ControlNotApplicable, ControlNotConfigured, ControlNeedsReview, ControlAcceptedException},
	}
	for _, control := range results {
		if control.ID != "CORE-TRUTH-001" {
			truthInputs.Controls = append(truthInputs.Controls, control)
		}
	}
	for index := range results {
		if results[index].ID == "CORE-TRUTH-001" {
			results[index].Evidence = []EvidenceReference{{Kind: "control-set-digest", Target: digestJSON(truthInputs)}}
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results
}

var sensitiveDiagnosticPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)gh[pousr]_[a-z0-9]{20,}`),
	regexp.MustCompile(`(?i)(api[_-]?key|token|password)[=: ]+[^ ,;]+`),
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
}

func redactDiagnostics(diagnostics []string) []string {
	redacted := append([]string{}, diagnostics...)
	for index := range redacted {
		for _, pattern := range sensitiveDiagnosticPatterns {
			redacted[index] = pattern.ReplaceAllString(redacted[index], "[REDACTED]")
		}
	}
	return redacted
}

func containsSensitiveText(value string) bool {
	for _, pattern := range sensitiveDiagnosticPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func evaluateRoutes(root string) ControlResult {
	result := ControlResult{
		ID: "CORE-ROUTES-001", State: ControlPass,
		Summary:  "Every stable seed route resolves to an existing repository artifact.",
		Evidence: []EvidenceReference{{Kind: "repository", Target: ".starter-kit/routes.json"}}, Diagnostics: []string{},
	}
	content, err := os.ReadFile(filepath.Join(root, ".starter-kit", "routes.json"))
	if err != nil {
		result.State, result.Evidence = ControlFail, []EvidenceReference{}
		result.Rationale = "The stable-route index could not be read."
		result.Diagnostics = []string{"routes file is unavailable"}
		return result
	}
	var routes struct {
		Routes map[string]string `json:"routes"`
	}
	if err := json.Unmarshal(content, &routes); err != nil {
		result.State, result.Evidence = ControlFail, []EvidenceReference{}
		result.Rationale = "The stable-route index is not valid JSON."
		result.Diagnostics = []string{"routes file is invalid JSON"}
		return result
	}
	required := map[string]string{
		"artifact:conformance":    "docs/evidence/CONFORMANCE.md",
		"artifact:decision-index": "docs/decisions/INDEX.md",
		"artifact:project-brief":  "docs/product/BRIEF.md",
		"artifact:personas":       "docs/product/PERSONAS.md",
	}
	for id, expected := range required {
		target, exists := routes.Routes[id]
		if !exists || target != expected {
			result.State, result.Evidence = ControlFail, []EvidenceReference{}
			result.Rationale = "A required stable route is missing or resolves to an unexpected target."
			result.Diagnostics = []string{"required route is missing or changed: " + id}
			return result
		}
		if err := validateRelativePath(root, target); err != nil || !fileExists(filepath.Join(root, filepath.FromSlash(target))) {
			result.State, result.Evidence = ControlFail, []EvidenceReference{}
			result.Rationale = "A required stable route does not resolve to an existing repository artifact."
			result.Diagnostics = []string{"one or more required routes do not resolve"}
			return result
		}
	}
	return result
}

func verificationDigest(result VerificationResult) string {
	result.EvidenceDigest = ""
	return digestJSON(result)
}

func persistVerification(root, planID string, result VerificationResult) error {
	evidenceContent := jsonDocument(result)
	summaryContent := renderConformanceSummary(result)
	manifestPath := filepath.Join(root, ".starter-kit", "managed-files.json")
	manifestContent, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read managed-file manifest: %w", err)
	}
	var manifest managedManifest
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		return fmt.Errorf("parse managed-file manifest: %w", err)
	}
	foundSummary := false
	for index := range manifest.Files {
		if manifest.Files[index].Path == "docs/evidence/CONFORMANCE.md" {
			manifest.Files[index].Digest = digestBytes([]byte(summaryContent))
			manifest.Files[index].Source = "engine:verify:v1"
			foundSummary = true
		}
	}
	if !foundSummary {
		return errors.New("managed-file manifest omits conformance summary")
	}
	updatedManifest := jsonDocument(manifest)
	event := operationEvent{
		SchemaVersion: 1, Ownership: "machine-evidence", Source: "engine:verify:v1",
		PlanID: planID, Operation: VerifyOperation, Status: ApplyStatusApplied,
		Actor: result.Actor, Authority: result.Authority, RepositoryDigest: result.SourceSnapshotDigest,
		ChangedFiles:    []string{result.EvidencePath, "docs/evidence/CONFORMANCE.md", ".starter-kit/managed-files.json"},
		ExternalEffects: []string{}, Diagnostics: []string{},
		Conflicts: []ReconciliationConflict{}, Recovery: []string{}, Evidence: []string{},
	}
	event.EventDigest = digestJSON(event)
	eventContent := jsonDocument(event)

	stageRoot, err := os.MkdirTemp(root, ".starter-kit-stage-verify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stageRoot)
	writes := []struct {
		path    string
		content string
		replace bool
	}{
		{result.EvidencePath, evidenceContent, false},
		{result.EventPath, eventContent, false},
		{"docs/evidence/CONFORMANCE.md", summaryContent, true},
		{".starter-kit/managed-files.json", updatedManifest, true},
	}
	for _, write := range writes {
		staged := filepath.Join(stageRoot, filepath.FromSlash(write.path))
		if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(staged, []byte(write.content), 0o644); err != nil {
			return err
		}
	}
	committed := make([]verificationCommit, 0, len(writes))
	for index, write := range writes {
		target := filepath.Join(root, filepath.FromSlash(write.path))
		staged := filepath.Join(stageRoot, filepath.FromSlash(write.path))
		backup := filepath.Join(stageRoot, fmt.Sprintf("backup-%d", index))
		commit := verificationCommit{target: target, backup: backup, replacement: write.replace}
		if write.replace {
			if err := os.Rename(target, backup); err != nil {
				rollbackVerification(committed)
				return err
			}
		} else if existing, err := os.ReadFile(target); err == nil {
			if string(existing) == write.content {
				continue
			}
			rollbackVerification(committed)
			return fmt.Errorf("verification evidence path already exists with different content: %s", write.path)
		} else if !errors.Is(err, os.ErrNotExist) {
			rollbackVerification(committed)
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			rollbackVerification(committed)
			return err
		}
		if err := os.Rename(staged, target); err != nil {
			if write.replace {
				_ = os.Rename(backup, target)
			}
			rollbackVerification(committed)
			return err
		}
		committed = append(committed, commit)
	}
	if problems := validateVerificationCommit(root, result); len(problems) != 0 {
		rollbackVerification(committed)
		return fmt.Errorf("verification invalidated managed contract: %v", problems)
	}
	return nil
}

func validateVerificationCommit(root string, result VerificationResult) []string {
	problems := []string{}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(result.EvidencePath)))
	if err != nil {
		return []string{"verification evidence is unavailable"}
	}
	var persisted VerificationResult
	if json.Unmarshal(content, &persisted) != nil || persisted.EvidenceDigest != verificationDigest(persisted) {
		problems = append(problems, "verification evidence digest is invalid")
	}
	summary, err := os.ReadFile(filepath.Join(root, "docs", "evidence", "CONFORMANCE.md"))
	if err != nil || string(summary) != renderConformanceSummary(result) {
		problems = append(problems, "conformance summary does not match authoritative evidence")
	}
	return problems
}

type verificationCommit struct {
	target      string
	backup      string
	replacement bool
}

func rollbackVerification(commits []verificationCommit) {
	for index := len(commits) - 1; index >= 0; index-- {
		commit := commits[index]
		_ = os.Remove(commit.target)
		if commit.replacement {
			_ = os.Rename(commit.backup, commit.target)
		}
	}
}

func renderConformanceSummary(result VerificationResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Conformance summary\n\nOverall state: `%s`\n\n", result.OverallState)
	fmt.Fprintf(&builder, "Scope: `%s`  \nGate: `%s`  \nVerification: `%s`  \nEvidence digest: `%s`  \nSource snapshot: `%s`\n\n", result.Scope, result.Gate, result.VerificationID, result.EvidenceDigest, result.SourceSnapshotDigest)
	builder.WriteString("## Seed controls\n\n| Control | State | Summary |\n|---|---|---|\n")
	for _, control := range result.Controls {
		fmt.Fprintf(&builder, "| `%s` | `%s` | %s |\n", control.ID, control.State, control.Summary)
	}
	builder.WriteString("\n## Coverage limitations\n\n")
	for _, limitation := range result.CoverageLimitations {
		fmt.Fprintf(&builder, "- %s\n", limitation)
	}
	return builder.String()
}
