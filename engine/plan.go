package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// Operation identifies a lifecycle-engine operation.
type Operation string

const (
	// CreateOperation initializes an empty or new managed repository.
	CreateOperation Operation = "create"
)

// Plan is an immutable, reviewable set of proposed repository operations.
type Plan struct {
	SchemaVersion    int           `json:"schema_version"`
	ID               string        `json:"plan_id"`
	Operation        Operation     `json:"operation"`
	Repository       string        `json:"repository"`
	RepositoryDigest string        `json:"repository_digest"`
	Files            []PlannedFile `json:"files"`
	NoChange         bool          `json:"no_change"`
}

// PlannedFile describes one file mutation without executing it.
type PlannedFile struct {
	Path      string `json:"path"`
	Ownership string `json:"ownership"`
	Source    string `json:"source"`
	Digest    string `json:"digest"`
	Content   string `json:"content"`
}

// Create composes the create operation into a reviewable plan.
func (e *Engine) Create(ctx context.Context, repository string) (Plan, error) {
	return e.Plan(ctx, repository, CreateOperation)
}

// Plan composes an immutable operation plan without modifying the repository.
func (e *Engine) Plan(ctx context.Context, repository string, operation Operation) (Plan, error) {
	if operation != CreateOperation {
		return Plan{}, fmt.Errorf("unsupported plan operation: %s", operation)
	}
	inspection, err := e.Inspect(ctx, repository)
	if err != nil {
		return Plan{}, err
	}
	if inspection.Managed {
		return noChangePlan(inspection, operation)
	}
	if inspection.UserFileCount != 0 {
		return Plan{}, fmt.Errorf("create requires an empty repository; found %d user files", inspection.UserFileCount)
	}

	files, err := createFiles()
	if err != nil {
		return Plan{}, fmt.Errorf("render create plan: %w", err)
	}
	plan := Plan{
		SchemaVersion:    1,
		Operation:        operation,
		Repository:       inspection.Repository,
		RepositoryDigest: digestJSON(inspection),
		Files:            files,
	}
	plan.ID = digestJSON(plan)
	return plan, nil
}

func noChangePlan(inspection Inspection, operation Operation) (Plan, error) {
	plan := Plan{
		SchemaVersion:    1,
		Operation:        operation,
		Repository:       inspection.Repository,
		RepositoryDigest: digestJSON(inspection),
		NoChange:         true,
		Files:            []PlannedFile{},
	}
	plan.ID = digestJSON(plan)
	return plan, nil
}

func createFiles() ([]PlannedFile, error) {
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
			}{1, "not_configured", []string{}, "seed core-trust pack is implemented by issue #27"}),
		},
		".starter-kit/project.json": {
			ownership: "managed",
			content: jsonDocument(struct {
				SchemaVersion int    `json:"schema_version"`
				Lifecycle     string `json:"lifecycle"`
				ProjectType   string `json:"project_type"`
			}{1, "managed", "unspecified"}),
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
			content:   "# Managed repository\n\nStart with the project brief, then follow stable routes in `.starter-kit/routes.json`.\n",
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
			content:   "# Project brief\n\nStatus: approved seed\n\nDescribe the project outcome without replacing this record automatically.\n",
		},
		"docs/product/PERSONAS.md": {
			ownership: "human-owned",
			content:   "# Persona registry\n\n## PER-OWNER — Project owner\n\nStatus: confirmed\n\nOwns product direction and ordinary implementation approval.\n",
		},
	}

	manifestEntries := make([]struct {
		Path      string `json:"path"`
		Ownership string `json:"ownership"`
		Source    string `json:"source"`
		Digest    string `json:"digest"`
	}, 0, len(base))
	for path, file := range base {
		manifestEntries = append(manifestEntries, struct {
			Path      string `json:"path"`
			Ownership string `json:"ownership"`
			Source    string `json:"source"`
			Digest    string `json:"digest"`
		}{path, file.ownership, "engine:create:v1", digestBytes([]byte(file.content))})
	}
	sort.Slice(manifestEntries, func(i, j int) bool { return manifestEntries[i].Path < manifestEntries[j].Path })
	base[".starter-kit/managed-files.json"] = struct {
		ownership string
		content   string
	}{"managed", jsonDocument(struct {
		SchemaVersion int         `json:"schema_version"`
		Files         interface{} `json:"files"`
	}{1, manifestEntries})}

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
