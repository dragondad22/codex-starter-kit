package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type manifestSelf struct {
	Path      string `json:"path"`
	Ownership string `json:"ownership"`
	Source    string `json:"source"`
}

type manifestEntry struct {
	Path      string `json:"path"`
	Ownership string `json:"ownership"`
	Source    string `json:"source"`
	Digest    string `json:"digest"`
}

type managedManifest struct {
	SchemaVersion int             `json:"schema_version"`
	Self          manifestSelf    `json:"self"`
	Files         []manifestEntry `json:"files"`
}

var requiredManagedPathsV1 = []string{
	".starter-kit/layout.json",
	".starter-kit/policy-lock.json",
	".starter-kit/project.json",
	".starter-kit/routes.json",
	".starter-kit/state.json",
	"AGENTS.md",
	"docs/decisions/INDEX.md",
	"docs/evidence/CONFORMANCE.md",
	"docs/product/BRIEF.md",
	"docs/product/PERSONAS.md",
}

func validateManagedContract(root string) (bool, []string) {
	starterPath := filepath.Join(root, ".starter-kit")
	if !fileExists(starterPath) {
		return false, []string{}
	}
	problems := make([]string, 0)
	manifestPath := filepath.Join(starterPath, "managed-files.json")
	if !fileExists(manifestPath) {
		contractMarker := false
		for _, marker := range []string{"state.json", "project.json", "layout.json", "policy-lock.json", "routes.json"} {
			if fileExists(filepath.Join(starterPath, marker)) {
				contractMarker = true
				break
			}
		}
		if !contractMarker {
			return false, []string{}
		}
	}
	if err := ensureNoSymlinkComponents(root, ".starter-kit/managed-files.json"); err != nil {
		return true, []string{"managed-file manifest is unavailable or traverses a symlink"}
	}
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return true, []string{fmt.Sprintf("read managed-file manifest: %v", err)}
	}
	var manifest managedManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return true, []string{fmt.Sprintf("parse managed-file manifest: %v", err)}
	}
	if manifest.SchemaVersion != 1 {
		problems = append(problems, fmt.Sprintf("unsupported managed-file schema: %d", manifest.SchemaVersion))
	}
	if manifest.Self.Path != ".starter-kit/managed-files.json" || manifest.Self.Ownership != "managed" || manifest.Self.Source != "engine:create:v1" {
		problems = append(problems, "managed-file manifest does not classify itself")
	}
	seen := make(map[string]struct{}, len(manifest.Files))
	for _, entry := range manifest.Files {
		if err := validateRelativePath(root, entry.Path); err != nil {
			problems = append(problems, fmt.Sprintf("invalid managed path %q: %v", entry.Path, err))
			continue
		}
		if _, duplicate := seen[entry.Path]; duplicate {
			problems = append(problems, fmt.Sprintf("duplicate managed path: %s", entry.Path))
			continue
		}
		seen[entry.Path] = struct{}{}
		if entry.Ownership == "" || entry.Source == "" || entry.Digest == "" {
			problems = append(problems, fmt.Sprintf("incomplete provenance for %s", entry.Path))
			continue
		}
		if !supportedManagedProvenance(entry.Path, entry.Source) {
			problems = append(problems, fmt.Sprintf("managed path has unsupported provenance: %s", entry.Path))
		}
		if err := ensureNoSymlinkComponents(root, entry.Path); err != nil {
			problems = append(problems, fmt.Sprintf("managed path is unavailable or traverses a symlink: %s", entry.Path))
			continue
		}
		if expectedOwnership, expected := createFileOwnershipV1[entry.Path]; !expected || entry.Ownership != expectedOwnership {
			problems = append(problems, fmt.Sprintf("managed path has unsupported ownership: %s", entry.Path))
		}
		fileContent, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.Path)))
		if readErr != nil {
			problems = append(problems, fmt.Sprintf("read managed file %s: %v", entry.Path, readErr))
			continue
		}
		if digestBytes(fileContent) != entry.Digest {
			problems = append(problems, fmt.Sprintf("managed file drift: %s", entry.Path))
		}
	}
	for _, required := range requiredManagedPathsV1 {
		if _, ok := seen[required]; !ok {
			problems = append(problems, fmt.Sprintf("managed-file manifest omits required artifact: %s", required))
		}
	}
	problems = append(problems, validateSeedStructuredState(root)...)
	eventDirectory := ".starter-kit/events"
	eventPaths := []string{}
	var eventGlobErr error
	if fileExists(filepath.Join(root, filepath.FromSlash(eventDirectory))) {
		if err := ensureNoSymlinkComponents(root, eventDirectory); err != nil {
			problems = append(problems, "operation-event directory is unavailable or traverses a symlink")
		} else {
			eventPaths, eventGlobErr = filepath.Glob(filepath.Join(starterPath, "events", "*.json"))
		}
	}
	if eventGlobErr != nil {
		problems = append(problems, fmt.Sprintf("enumerate operation events: %v", eventGlobErr))
	}
	for _, eventPath := range eventPaths {
		relative, relErr := filepath.Rel(root, eventPath)
		if relErr != nil || ensureNoSymlinkComponents(root, filepath.ToSlash(relative)) != nil {
			problems = append(problems, "operation event is unavailable or traverses a symlink")
			continue
		}
		eventContent, readErr := os.ReadFile(eventPath)
		if readErr != nil {
			problems = append(problems, fmt.Sprintf("read operation event: %v", readErr))
			continue
		}
		var event operationEvent
		if err := json.Unmarshal(eventContent, &event); err != nil {
			problems = append(problems, fmt.Sprintf("parse operation event: %v", err))
			continue
		}
		recordedDigest := event.EventDigest
		event.EventDigest = ""
		if event.SchemaVersion != 1 || event.Ownership != "machine-evidence" || event.Source == "" || event.PlanID == "" || event.Operation == "" || event.Status == "" || event.Actor == "" || event.Authority == "" || event.ExternalEffects == nil || event.Diagnostics == nil || recordedDigest == "" || digestJSON(event) != recordedDigest {
			problems = append(problems, fmt.Sprintf("operation event lacks required provenance: %s", filepath.Base(eventPath)))
		}
	}
	evidenceDirectory := ".starter-kit/evidence"
	evidencePaths := []string{}
	var evidenceGlobErr error
	if fileExists(filepath.Join(root, filepath.FromSlash(evidenceDirectory))) {
		if err := ensureNoSymlinkComponents(root, evidenceDirectory); err != nil {
			problems = append(problems, "verification-evidence directory is unavailable or traverses a symlink")
		} else {
			evidencePaths, evidenceGlobErr = filepath.Glob(filepath.Join(starterPath, "evidence", "verify-*.json"))
		}
	}
	if evidenceGlobErr != nil {
		problems = append(problems, fmt.Sprintf("enumerate verification evidence: %v", evidenceGlobErr))
	}
	for _, evidencePath := range evidencePaths {
		relative, relErr := filepath.Rel(root, evidencePath)
		if relErr != nil || ensureNoSymlinkComponents(root, filepath.ToSlash(relative)) != nil {
			problems = append(problems, "verification evidence is unavailable or traverses a symlink")
			continue
		}
		evidenceContent, readErr := os.ReadFile(evidencePath)
		if readErr != nil {
			problems = append(problems, fmt.Sprintf("read verification evidence: %v", readErr))
			continue
		}
		var evidence VerificationResult
		if err := json.Unmarshal(evidenceContent, &evidence); err != nil {
			problems = append(problems, fmt.Sprintf("parse verification evidence: %v", err))
			continue
		}
		if evidence.SchemaVersion != 1 || evidence.Ownership != "machine-evidence" || evidence.Source != "engine:verify:v1" || evidence.VerificationID == "" || evidence.EvidenceDigest == "" || verificationDigest(evidence) != evidence.EvidenceDigest {
			problems = append(problems, fmt.Sprintf("verification evidence lacks valid provenance: %s", filepath.Base(evidencePath)))
			continue
		}
		knownControls := map[string]bool{
			"CORE-TRUTH-001": true, "CORE-SECRETS-001": true, "CORE-OWNERSHIP-001": true,
			"CORE-COVERAGE-001": true, "CORE-RECOVERY-001": true, "CORE-ROUTES-001": true,
		}
		seenControls := map[string]bool{}
		for _, control := range evidence.Controls {
			if !knownControls[control.ID] || seenControls[control.ID] {
				problems = append(problems, fmt.Sprintf("verification evidence has unknown or duplicate control identity: %s", control.ID))
			}
			seenControls[control.ID] = true
			if !validControlState(control.State) {
				problems = append(problems, fmt.Sprintf("verification evidence has invalid state for %s", control.ID))
			}
			if control.State == ControlPass && len(control.Evidence) == 0 {
				problems = append(problems, fmt.Sprintf("passing control lacks evidence: %s", control.ID))
			}
			if control.State != ControlPass && control.Rationale == "" {
				problems = append(problems, fmt.Sprintf("non-passing control lacks rationale: %s", control.ID))
			}
			if control.State == ControlAcceptedException && (control.UnderlyingState == "" || control.UnderlyingState == ControlPass || control.UnderlyingState == ControlAcceptedException) {
				problems = append(problems, fmt.Sprintf("accepted exception lacks underlying state: %s", control.ID))
			}
			if control.State != ControlAcceptedException && control.UnderlyingState != "" {
				problems = append(problems, fmt.Sprintf("ordinary control has an underlying state: %s", control.ID))
			}
			if control.State == ControlNotApplicable && len(control.Evidence) == 0 {
				problems = append(problems, fmt.Sprintf("not-applicable control lacks supporting facts: %s", control.ID))
			}
		}
		if len(seenControls) != len(knownControls) {
			problems = append(problems, "verification evidence does not contain exactly the seed control set")
		}
		if !validControlState(evidence.OverallState) || evidence.OverallState != OverallControlState(evidence.Controls) {
			problems = append(problems, "verification evidence overall state is invalid")
		}
	}
	sort.Strings(problems)
	return true, redactDiagnostics(problems)
}

func validateSeedStructuredState(root string) []string {
	problems := []string{}
	readJSON := func(path string, value interface{}) bool {
		if err := ensureNoSymlinkComponents(root, path); err != nil {
			problems = append(problems, "structured managed state is unavailable or traverses a symlink")
			return false
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			problems = append(problems, "read structured managed state: "+err.Error())
			return false
		}
		if err := json.Unmarshal(content, value); err != nil {
			problems = append(problems, "parse structured managed state: "+err.Error())
			return false
		}
		return true
	}

	var state struct {
		SchemaVersion int    `json:"schema_version"`
		Lifecycle     string `json:"lifecycle"`
		EngineVersion string `json:"engine_version"`
	}
	if readJSON(".starter-kit/state.json", &state) && (state.SchemaVersion != 1 || state.Lifecycle != LifecycleManaged || state.EngineVersion != "0.1.0-dev") {
		problems = append(problems, "state does not describe the supported managed lifecycle and engine version")
	}

	var layout struct {
		SchemaVersion int               `json:"schema_version"`
		Roles         map[string]string `json:"roles"`
	}
	expectedRoles := map[string]string{"decisions": "docs/decisions", "evidence": "docs/evidence", "product": "docs/product"}
	if readJSON(".starter-kit/layout.json", &layout) {
		if layout.SchemaVersion != 1 || !equalStringMap(layout.Roles, expectedRoles) {
			problems = append(problems, "layout does not contain the supported logical role mapping")
		}
		for _, path := range layout.Roles {
			if err := validateRelativePath(root, path); err != nil {
				problems = append(problems, "layout contains an unsafe logical role path")
			}
		}
	}

	var routes struct {
		SchemaVersion int               `json:"schema_version"`
		Routes        map[string]string `json:"routes"`
	}
	expectedRoutes := map[string]string{
		"artifact:conformance": "docs/evidence/CONFORMANCE.md", "artifact:decision-index": "docs/decisions/INDEX.md",
		"artifact:project-brief": "docs/product/BRIEF.md", "artifact:personas": "docs/product/PERSONAS.md",
	}
	if readJSON(".starter-kit/routes.json", &routes) && (routes.SchemaVersion != 1 || !equalStringMap(routes.Routes, expectedRoutes)) {
		problems = append(problems, "routes do not contain the supported stable breadcrumb mapping")
	}

	var project struct {
		SchemaVersion         int    `json:"schema_version"`
		Lifecycle             string `json:"lifecycle"`
		ProjectType           string `json:"project_type"`
		BriefApproved         bool   `json:"brief_approved"`
		OwnerPersonaConfirmed bool   `json:"owner_persona_confirmed"`
	}
	if readJSON(".starter-kit/project.json", &project) && (project.SchemaVersion != 1 || project.Lifecycle != LifecycleManaged || project.ProjectType != "unspecified" || !project.BriefApproved || !project.OwnerPersonaConfirmed) {
		problems = append(problems, "project state lacks the supported lifecycle or required approvals")
	}

	var policy struct {
		SchemaVersion int      `json:"schema_version"`
		Status        string   `json:"status"`
		Packs         []string `json:"packs"`
		Reason        string   `json:"reason"`
	}
	if readJSON(".starter-kit/policy-lock.json", &policy) && (policy.SchemaVersion != 1 || policy.Status != "not_configured" || len(policy.Packs) != 0 || policy.Reason == "") {
		problems = append(problems, "policy lock does not truthfully describe the unconfigured seed policy")
	}
	return problems
}

func equalStringMap(actual, expected map[string]string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for key, value := range expected {
		if actual[key] != value {
			return false
		}
	}
	return true
}

func supportedManagedProvenance(path, source string) bool {
	if source == "engine:create:v1" {
		return true
	}
	return path == "docs/evidence/CONFORMANCE.md" && source == "engine:verify:v1"
}

func validControlState(state ControlState) bool {
	switch state {
	case ControlPass, ControlFail, ControlNotApplicable, ControlNotConfigured, ControlNeedsReview, ControlAcceptedException:
		return true
	default:
		return false
	}
}

func validateRelativePath(root, slashPath string) error {
	if slashPath == "" || strings.ContainsRune(slashPath, '\x00') {
		return fmt.Errorf("path is empty or contains NUL")
	}
	if strings.Contains(slashPath, "\\") {
		return fmt.Errorf("path must use forward slashes")
	}
	for _, character := range slashPath {
		if character < 0x20 || character > 0x7e {
			return fmt.Errorf("path must use printable ASCII to avoid normalization ambiguity")
		}
	}
	for _, segment := range strings.Split(slashPath, "/") {
		if err := validatePortablePathSegment(segment); err != nil {
			return err
		}
	}
	native := filepath.FromSlash(slashPath)
	if filepath.IsAbs(native) || filepath.Clean(native) == "." || filepath.Clean(native) != native {
		return fmt.Errorf("path is not a clean relative path")
	}
	target := filepath.Join(root, native)
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes repository root")
	}
	return nil
}

func validatePortablePathSegment(segment string) error {
	if segment == "" || segment == "." || segment == ".." {
		return fmt.Errorf("path contains an empty or relative segment")
	}
	if strings.HasSuffix(segment, ".") || strings.HasSuffix(segment, " ") {
		return fmt.Errorf("path segment has a trailing dot or space")
	}
	if strings.ContainsAny(segment, `<>:"|?*`) {
		return fmt.Errorf("path segment contains a reserved character")
	}
	base := strings.ToUpper(strings.SplitN(segment, ".", 2)[0])
	reserved := base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || base == "CLOCK$"
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9' {
		reserved = true
	}
	if reserved {
		return fmt.Errorf("path segment uses a reserved device name")
	}
	return nil
}
