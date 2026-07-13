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
	if manifest.Self.Path != ".starter-kit/managed-files.json" || manifest.Self.Ownership != "managed" || manifest.Self.Source == "" {
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
	stateContent, stateErr := os.ReadFile(filepath.Join(starterPath, "state.json"))
	if stateErr == nil {
		var state struct {
			SchemaVersion int    `json:"schema_version"`
			Lifecycle     string `json:"lifecycle"`
		}
		if err := json.Unmarshal(stateContent, &state); err != nil {
			problems = append(problems, fmt.Sprintf("parse state: %v", err))
		} else if state.SchemaVersion != 1 || state.Lifecycle != LifecycleManaged {
			problems = append(problems, "state does not describe a supported managed lifecycle")
		}
	}
	eventPaths, eventGlobErr := filepath.Glob(filepath.Join(starterPath, "events", "*.json"))
	if eventGlobErr != nil {
		problems = append(problems, fmt.Sprintf("enumerate operation events: %v", eventGlobErr))
	}
	for _, eventPath := range eventPaths {
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
	evidencePaths, evidenceGlobErr := filepath.Glob(filepath.Join(starterPath, "evidence", "verify-*.json"))
	if evidenceGlobErr != nil {
		problems = append(problems, fmt.Sprintf("enumerate verification evidence: %v", evidenceGlobErr))
	}
	for _, evidencePath := range evidencePaths {
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
	return true, problems
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
