package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Operation identifies a lifecycle-engine operation.
type Operation string

const (
	// CreateOperation initializes an empty or new managed repository.
	CreateOperation Operation = "create"
	VerifyOperation Operation = "verify"
)

// Plan is an immutable, reviewable set of proposed repository operations.
type Plan struct {
	SchemaVersion    int            `json:"schema_version"`
	ID               string         `json:"plan_id"`
	Operation        Operation      `json:"operation"`
	Repository       string         `json:"repository"`
	RepositoryDigest string         `json:"repository_digest"`
	Files            []PlannedFile  `json:"files"`
	NoChange         bool           `json:"no_change"`
	Approval         CreateApproval `json:"approval"`
	Result           PlannedResult  `json:"result"`
}

// CreateRequest contains the human-owned inputs and confirmations needed to plan create.
type CreateRequest struct {
	Repository            string `json:"repository"`
	Brief                 string `json:"brief"`
	BriefApproved         bool   `json:"brief_approved"`
	OwnerPersonaConfirmed bool   `json:"owner_persona_confirmed"`
}

// PlanRequest identifies an operation and its approved inputs.
type PlanRequest struct {
	Operation Operation     `json:"operation"`
	Create    CreateRequest `json:"create"`
}

// CreateApproval retains the approval facts that shaped a create plan.
type CreateApproval struct {
	BriefDigest           string `json:"brief_digest"`
	BriefApproved         bool   `json:"brief_approved"`
	OwnerPersonaConfirmed bool   `json:"owner_persona_confirmed"`
}

// PlannedResult declares the ownership and destination of operation-result evidence.
type PlannedResult struct {
	Path      string `json:"path"`
	Ownership string `json:"ownership"`
	Source    string `json:"source"`
}

// PlannedFile describes one file mutation without executing it.
type PlannedFile struct {
	Path      string `json:"path"`
	Ownership string `json:"ownership"`
	Source    string `json:"source"`
	Digest    string `json:"digest"`
	Content   string `json:"content"`
}

// ReconciliationConflict identifies existing material that create cannot replace.
type ReconciliationConflict struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Ownership string `json:"ownership"`
	Reason    string `json:"reason"`
}

// ReconciliationRequired is a reviewable stop result for preserved repository content.
type ReconciliationRequired struct {
	SchemaVersion int                      `json:"schema_version"`
	Repository    string                   `json:"repository"`
	Conflicts     []ReconciliationConflict `json:"conflicts"`
	Problems      []string                 `json:"problems"`
	Recovery      []string                 `json:"recovery"`
}

func (r *ReconciliationRequired) Error() string {
	return fmt.Sprintf("create requires reconciliation for %d existing conflicts", len(r.Conflicts))
}

var createFileOwnershipV1 = map[string]string{
	".starter-kit/layout.json":        "managed",
	".starter-kit/managed-files.json": "managed",
	".starter-kit/policy-lock.json":   "managed",
	".starter-kit/project.json":       "managed",
	".starter-kit/routes.json":        "generated",
	".starter-kit/state.json":         "managed",
	"AGENTS.md":                       "generated",
	"docs/decisions/INDEX.md":         "human-owned",
	"docs/evidence/CONFORMANCE.md":    "generated",
	"docs/product/BRIEF.md":           "human-owned",
	"docs/product/PERSONAS.md":        "human-owned",
}

// Create composes the create operation into a reviewable plan.
func (e *Engine) Create(ctx context.Context, request CreateRequest) (Plan, error) {
	return e.Plan(ctx, PlanRequest{Operation: CreateOperation, Create: request})
}

// Plan composes an immutable operation plan without modifying the repository.
func (e *Engine) Plan(ctx context.Context, request PlanRequest) (Plan, error) {
	if request.Operation != CreateOperation {
		return Plan{}, fmt.Errorf("unsupported plan operation: %s", request.Operation)
	}
	if request.Create.Brief == "" || !request.Create.BriefApproved || !request.Create.OwnerPersonaConfirmed {
		return Plan{}, errors.New("create requires an approved brief and confirmed owner persona")
	}
	if containsSensitiveText(request.Create.Brief) {
		return Plan{}, errors.New("create brief contains sensitive-looking content that cannot enter a plan")
	}
	inspection, err := e.Inspect(ctx, request.Create.Repository)
	if err != nil {
		return Plan{}, err
	}
	files, err := createFiles(request.Create)
	if err != nil {
		return Plan{}, fmt.Errorf("render create plan: %w", err)
	}
	if inspection.Managed {
		if !plannedFilesMatch(inspection.Repository, files) {
			return Plan{}, plannedReconciliation(
				inspection.Repository, files,
				[]string{"managed repository content differs from the newly approved create inputs"},
			)
		}
		return noChangePlan(inspection, request)
	}
	if inspection.ContractPresent {
		return Plan{}, plannedReconciliation(inspection.Repository, files, inspection.Problems)
	}
	if inspection.UserFileCount != 0 || inspection.UserDirectoryCount != 0 {
		conflicts, conflictErr := userOwnedConflicts(inspection.Repository)
		if conflictErr != nil {
			return Plan{}, conflictErr
		}
		return Plan{}, &ReconciliationRequired{
			SchemaVersion: 1,
			Repository:    inspection.Repository,
			Conflicts:     conflicts,
			Problems:      []string{"create has no authority to replace existing repository content"},
			Recovery: []string{
				"review and preserve each listed user-owned path",
				"use a future authorized retrofit/reconciliation operation or select an empty repository",
				"re-run create only after the repository facts intentionally change",
			},
		}
	}

	plan := Plan{
		SchemaVersion:    1,
		Operation:        request.Operation,
		Repository:       inspection.Repository,
		RepositoryDigest: inspection.PreconditionDigest,
		Files:            files,
		Approval: CreateApproval{
			BriefDigest:           digestBytes([]byte(request.Create.Brief)),
			BriefApproved:         request.Create.BriefApproved,
			OwnerPersonaConfirmed: request.Create.OwnerPersonaConfirmed,
		},
		Result: PlannedResult{
			Path:      operationEventPath(request.Operation, inspection.PreconditionDigest),
			Ownership: "machine-evidence",
			Source:    "engine:apply:v1",
		},
	}
	plan.ID = digestJSON(plan)
	return plan, nil
}

func plannedReconciliation(root string, files []PlannedFile, problems []string) *ReconciliationRequired {
	conflicts := []ReconciliationConflict{}
	for _, file := range files {
		target := filepath.Join(root, filepath.FromSlash(file.Path))
		info, err := os.Lstat(target)
		if errors.Is(err, os.ErrNotExist) {
			conflicts = append(conflicts, ReconciliationConflict{
				Path: file.Path, Kind: "missing", Ownership: file.Ownership,
				Reason: "approved create input expects an artifact that is not present",
			})
			continue
		}
		if err != nil {
			conflicts = append(conflicts, ReconciliationConflict{
				Path: file.Path, Kind: "unavailable", Ownership: file.Ownership,
				Reason: "existing artifact could not be inspected safely",
			})
			continue
		}
		kind := "file"
		if info.Mode()&os.ModeSymlink != 0 {
			kind = "symlink"
		}
		content, readErr := os.ReadFile(target)
		if kind != "file" || readErr != nil || digestBytes(content) != file.Digest {
			conflicts = append(conflicts, ReconciliationConflict{
				Path: file.Path, Kind: kind, Ownership: file.Ownership,
				Reason: "existing artifact differs from the newly approved create input",
			})
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Path < conflicts[j].Path })
	return &ReconciliationRequired{
		SchemaVersion: 1, Repository: root, Conflicts: conflicts,
		Problems: append([]string{}, problems...),
		Recovery: []string{
			"review every changed, missing, or unavailable artifact without overwriting human-owned content",
			"use a future authorized retrofit/reconciliation operation for accepted changes",
			"re-run create only when the approved inputs and repository facts intentionally agree",
		},
	}
}

func userOwnedConflicts(root string) ([]ReconciliationConflict, error) {
	conflicts := []ReconciliationConflict{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() && relative == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() && strings.HasPrefix(relative, ".starter-kit-stage-") {
			return filepath.SkipDir
		}
		kind := "file"
		if entry.IsDir() {
			children, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			if len(children) != 0 {
				return nil
			}
			kind = "directory"
		} else if entry.Type()&os.ModeSymlink != 0 {
			kind = "symlink"
		}
		conflicts = append(conflicts, ReconciliationConflict{
			Path: reviewableConflictPath(filepath.ToSlash(relative)), Kind: kind, Ownership: "user-owned",
			Reason: "existing content cannot be replaced without explicit reconciliation authority",
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("enumerate reconciliation conflicts: %w", err)
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Path < conflicts[j].Path })
	return conflicts, nil
}

func reviewableConflictPath(path string) string {
	if !containsSensitiveText(path) {
		return path
	}
	digest := strings.TrimPrefix(digestBytes([]byte(path)), "sha256:")
	return "[REDACTED]-sha256:" + digest[:16]
}

func plannedFilesMatch(root string, files []PlannedFile) bool {
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil || digestBytes(content) != file.Digest {
			return false
		}
	}
	return true
}

func noChangePlan(inspection Inspection, request PlanRequest) (Plan, error) {
	plan := Plan{
		SchemaVersion:    1,
		Operation:        request.Operation,
		Repository:       inspection.Repository,
		RepositoryDigest: inspection.PreconditionDigest,
		NoChange:         true,
		Files:            []PlannedFile{},
		Approval: CreateApproval{
			BriefDigest:           digestBytes([]byte(request.Create.Brief)),
			BriefApproved:         request.Create.BriefApproved,
			OwnerPersonaConfirmed: request.Create.OwnerPersonaConfirmed,
		},
		Result: PlannedResult{
			Path:      operationEventPath(request.Operation, inspection.PreconditionDigest),
			Ownership: "machine-evidence",
			Source:    "engine:apply:v1",
		},
	}
	plan.ID = digestJSON(plan)
	return plan, nil
}

func operationEventPath(operation Operation, preconditionDigest string) string {
	digest := strings.TrimPrefix(preconditionDigest, "sha256:")
	if len(digest) > 16 {
		digest = digest[:16]
	}
	return fmt.Sprintf(".starter-kit/events/%s-%s.json", operation, digest)
}

func createFiles(request CreateRequest) ([]PlannedFile, error) {
	base := map[string]struct {
		ownership string
		content   string
	}{
		".starter-kit/layout.json": {
			ownership: "managed",
			content: jsonDocument(struct {
				SchemaVersion int               `json:"schema_version"`
				Roles         map[string]string `json:"roles"`
			}{1, map[string]string{
				"decisions": "docs/decisions",
				"evidence":  "docs/evidence",
				"product":   "docs/product",
			}}),
		},
		".starter-kit/policy-lock.json": {
			ownership: "managed",
			content: jsonDocument(struct {
				SchemaVersion int      `json:"schema_version"`
				Status        string   `json:"status"`
				Packs         []string `json:"packs"`
				Reason        string   `json:"reason"`
			}{1, "not_configured", []string{}, "signed core-trust policy pack is not configured"}),
		},
		".starter-kit/project.json": {
			ownership: "managed",
			content: jsonDocument(struct {
				SchemaVersion         int    `json:"schema_version"`
				Lifecycle             string `json:"lifecycle"`
				ProjectType           string `json:"project_type"`
				BriefApproved         bool   `json:"brief_approved"`
				OwnerPersonaConfirmed bool   `json:"owner_persona_confirmed"`
			}{1, "managed", "unspecified", request.BriefApproved, request.OwnerPersonaConfirmed}),
		},
		".starter-kit/routes.json": {
			ownership: "generated",
			content: jsonDocument(struct {
				SchemaVersion int               `json:"schema_version"`
				Routes        map[string]string `json:"routes"`
			}{1, map[string]string{
				"artifact:conformance":    "docs/evidence/CONFORMANCE.md",
				"artifact:decision-index": "docs/decisions/INDEX.md",
				"artifact:project-brief":  "docs/product/BRIEF.md",
				"artifact:personas":       "docs/product/PERSONAS.md",
			}}),
		},
		".starter-kit/state.json": {
			ownership: "managed",
			content: jsonDocument(struct {
				SchemaVersion int    `json:"schema_version"`
				Lifecycle     string `json:"lifecycle"`
				EngineVersion string `json:"engine_version"`
			}{1, "managed", "0.1.0-dev"}),
		},
		"AGENTS.md": {
			ownership: "generated",
			content: "# Managed repository\n\n" +
				"Start with the project brief, then follow stable routes in `.starter-kit/routes.json`.\n\n" +
				"When conversation surfaces durable untracked work, a consequential question, or a decision that must be promoted, " +
				"search open and closed GitHub Issues first. Update a duplicate or contained issue, suggest a lifecycle-specific issue for genuinely new work, " +
				"and route an approved material decision to its authoritative record. " +
				"Prompt at a natural checkpoint; ordinary clarification stays in the conversation. " +
				"Do not begin material implementation until the applicable issue's Readiness is `Ready`, and reference its `#N` while working.\n",
		},
		"docs/decisions/INDEX.md": {
			ownership: "human-owned",
			content:   "# Decision Index\n\nHuman-owned decisions are added here and are never silently regenerated.\n",
		},
		"docs/evidence/CONFORMANCE.md": {
			ownership: "generated",
			content:   "# Conformance summary\n\nInitial verification has not run. No controls are reported as passing.\n",
		},
		"docs/product/BRIEF.md": {
			ownership: "human-owned",
			content:   "# Project brief\n\nStatus: approved\n\n" + request.Brief + "\n",
		},
		"docs/product/PERSONAS.md": {
			ownership: "human-owned",
			content:   "# Persona registry\n\n## PER-OWNER — Project owner\n\nStatus: confirmed\n\nOwns product direction and ordinary implementation approval.\n",
		},
	}

	manifestEntries := make([]manifestEntry, 0, len(base))
	for path, file := range base {
		manifestEntries = append(manifestEntries, manifestEntry{
			Path:      path,
			Ownership: file.ownership,
			Source:    "engine:create:v1",
			Digest:    digestBytes([]byte(file.content)),
		})
	}
	sort.Slice(manifestEntries, func(i, j int) bool { return manifestEntries[i].Path < manifestEntries[j].Path })
	base[".starter-kit/managed-files.json"] = struct {
		ownership string
		content   string
	}{"managed", jsonDocument(managedManifest{
		SchemaVersion: 1,
		Self: manifestSelf{
			Path:      ".starter-kit/managed-files.json",
			Ownership: "managed",
			Source:    "engine:create:v1",
		},
		Files: manifestEntries,
	})}

	paths := make([]string, 0, len(base))
	for path := range base {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]PlannedFile, 0, len(paths))
	for _, path := range paths {
		file := base[path]
		files = append(files, PlannedFile{
			Path:      path,
			Ownership: file.ownership,
			Source:    "engine:create:v1",
			Digest:    digestBytes([]byte(file.content)),
			Content:   file.content,
		})
	}
	return files, nil
}

func jsonDocument(value interface{}) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(encoded) + "\n"
}

func digestJSON(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func validSHA256Digest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}
