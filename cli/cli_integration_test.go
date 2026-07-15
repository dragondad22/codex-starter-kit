package cli_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dragondad22/codex-starter-kit/cli"
	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestCreateCommandEmitsLanguageNeutralPlan(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run(createArguments("create", repository), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var plan engine.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan JSON: %v: %s", err, stdout.String())
	}
	if plan.Operation != engine.CreateOperation || plan.ID == "" {
		t.Fatalf("unexpected create plan: %#v", plan)
	}
}

func TestCapabilitiesCommandReportsNonMutatingCompatibilityFacts(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"capabilities"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var report engine.CapabilityReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode capability report: %v: %s", err, stdout.String())
	}
	if report.SchemaVersion != 1 || report.Engine.Name != "starter-kit" || report.Engine.Version != "0.3.0" {
		t.Fatalf("unexpected engine identity: %#v", report)
	}
	if report.Protocol.Name != "starter-kit.lifecycle" || report.Protocol.Version != 1 {
		t.Fatalf("unexpected protocol: %#v", report.Protocol)
	}
	if len(report.Operations) == 0 || report.Operations[0] != "apply" {
		t.Fatalf("operations are absent or unstable: %#v", report.Operations)
	}
	if report.Engine.Provenance != engine.ProvenanceUnverified {
		t.Fatalf("engine self-asserted provenance trust: %#v", report.Engine)
	}
}

func TestVersionCommandReportsCanonicalProductIdentity(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"version"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var identity struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &identity); err != nil {
		t.Fatalf("decode product identity: %v: %s", err, stdout.String())
	}
	if identity.Name != "codex-starter-kit" || identity.Version != "0.3.0" {
		t.Fatalf("unexpected product identity: %#v", identity)
	}
}

func TestChangesRenderProducesUnreleasedHumanSummary(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Generate human release summaries from structured change records.",
  "category": "added",
  "audiences": ["users", "developers"],
  "components": ["cli", "release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "render", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, expected := range []string{
		"# Changelog", "## [Unreleased]", "### Added",
		"Generate human release summaries from structured change records. (#78)",
	} {
		if !bytes.Contains(stdout.Bytes(), []byte(expected)) {
			t.Fatalf("rendered changelog missing %q:\n%s", expected, stdout.String())
		}
	}
}

func TestChangesRenderFiltersAudienceViews(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Show users a generated changelog.",
  "category": "added",
  "audiences": ["users"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	writeChangeRecord(t, repository, "issue-79-operator-note", `{
  "schema_version": 1,
  "id": "issue-79-operator-note",
  "summary": "Require operators to rotate a release credential.",
  "category": "changed",
  "audiences": ["operators"],
  "components": ["release"],
  "issues": [79],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "render", "--repository", repository, "--audience", "users"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Show users a generated changelog.") {
		t.Fatalf("user change missing:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "rotate a release credential") {
		t.Fatalf("operator-only change leaked into user view:\n%s", stdout.String())
	}
}

func TestChangesRenderRejectsComponentVersionSkew(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Keep product versions synchronized.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	pluginManifest := filepath.Join(repository, "plugins", "codex-starter-kit", ".codex-plugin", "plugin.json")
	if err := os.WriteFile(pluginManifest, []byte(`{"version":"0.2.0"}`), 0o644); err != nil {
		t.Fatalf("write skewed plugin manifest: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "render", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q, stderr = %q", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "component version mismatch") {
		t.Fatalf("missing actionable version-skew diagnostic: %q", stderr.String())
	}
}

func TestChangesValidateReportsVersionAndPendingRecordCounts(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Validate release records in CI.",
  "category": "added",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var result struct {
		Version           string `json:"version"`
		UnreleasedRecords int    `json:"unreleased_records"`
		ExternalRecords   int    `json:"external_records"`
		InternalRecords   int    `json:"internal_records"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode validation result: %v: %s", err, stdout.String())
	}
	if result.Version != "0.3.0" || result.UnreleasedRecords != 1 || result.ExternalRecords != 1 || result.InternalRecords != 0 {
		t.Fatalf("unexpected validation result: %#v", result)
	}
}

func TestChangesValidateRejectsUnsafeRecordIdentity(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Keep one valid record in the fixture.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	writeChangeRecord(t, repository, "unsafe record", `{
  "schema_version": 1,
  "id": "unsafe record",
  "summary": "This identity is ambiguous across tooling.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q, stderr = %q", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "safe lowercase identifier") {
		t.Fatalf("missing unsafe-identity diagnostic: %q", stderr.String())
	}
}

func TestReleasePrepareArchivesRecordsAndNeverClaimsPublication(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Generate release summaries from durable records.",
  "category": "added",
  "audiences": ["users", "developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	admission := writeReleaseAdmission(t, repository, "0.4.0", "issue-78-release-tracking")

	exitCode := cli.Run([]string{
		"release", "prepare", "--repository", repository,
		"--version", "0.4.0", "--date", "2026-07-15", "--admission", admission,
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var result struct {
		Version   string `json:"version"`
		State     string `json:"state"`
		Published bool   `json:"published"`
		Records   int    `json:"records"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode preparation result: %v: %s", err, stdout.String())
	}
	if result.Version != "0.4.0" || result.State != "prepared" || result.Published || result.Records != 1 {
		t.Fatalf("unexpected preparation result: %#v", result)
	}
	for path, expected := range map[string]string{
		"product-version.json":                                  `"version": "0.4.0"`,
		"changes/releases/0.4.0/release.json":                   `"state": "prepared"`,
		"changes/releases/0.4.0/issue-78-release-tracking.json": `"id": "issue-78-release-tracking"`,
		"CHANGELOG.md": "## [0.4.0] - 2026-07-15",
	} {
		content, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read prepared %s: %v", path, err)
		}
		if !strings.Contains(string(content), expected) {
			t.Fatalf("%s missing %q:\n%s", path, expected, content)
		}
	}
	if _, err := os.Stat(filepath.Join(repository, "changes", "unreleased", "issue-78-release-tracking.json")); !os.IsNotExist(err) {
		t.Fatalf("unreleased record was not archived: %v", err)
	}
	changelog, err := os.ReadFile(filepath.Join(repository, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(changelog), "## [Unreleased]") || !strings.Contains(string(changelog), "Generate release summaries") {
		t.Fatalf("prepared changelog lost fresh or released view:\n%s", changelog)
	}
}

func TestChangesValidateRejectsTrailingJSONContent(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Reject concealed trailing documents.",
  "category": "security",
  "audiences": ["security"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}
{"concealed": true}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q, stderr = %q", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "trailing JSON content") {
		t.Fatalf("missing trailing-content diagnostic: %q", stderr.String())
	}
}

func TestChangesValidateRejectsUnknownAudience(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Reject unreachable audience views.",
  "category": "changed",
  "audiences": ["mystery-audience"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q, stderr = %q", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "unsupported audience") {
		t.Fatalf("missing audience diagnostic: %q", stderr.String())
	}
}

func TestChangesValidateRejectsOmittedRequiredBoolean(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Do not infer required release facts.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "internal_only": false
}`)
	var stdout, stderr bytes.Buffer
	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 || !strings.Contains(stderr.String(), `missing required field "breaking"`) {
		t.Fatalf("missing required-field rejection: exit=%d stderr=%q", exitCode, stderr.String())
	}
}

func TestChangesCheckRejectsStaleGeneratedChangelog(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Keep generated communication current.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	if err := os.WriteFile(filepath.Join(repository, "CHANGELOG.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "check", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1; stdout = %q, stderr = %q", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "generated CHANGELOG.md is stale") {
		t.Fatalf("missing stale-view diagnostic: %q", stderr.String())
	}
}

func TestChangesValidateRejectsDuplicateRecordIdentity(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "First use of the stable record identity.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	writeChangeRecord(t, repository, "issue-79-duplicate", `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Second use of the same stable record identity.",
  "category": "changed",
  "audiences": ["developers"],
  "components": ["release"],
  "issues": [79],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr.String(), "duplicate change record id") {
		t.Fatalf("missing duplicate-record diagnostic: %q", stderr.String())
	}
}

func TestChangesValidateRejectsInternalOnlyRecordWithoutDisposition(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Internal work still needs a durable disposition.",
  "category": "changed",
  "audiences": [],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": true
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr)
	if exitCode != 1 || !strings.Contains(stderr.String(), "internal_disposition") {
		t.Fatalf("missing internal-only disposition failure: exit=%d stderr=%q", exitCode, stderr.String())
	}
}

func TestReleasePrepareRejectsNonIncrementingVersionWithoutMutation(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Preserve pending work after refused preparation.",
  "category": "fixed",
  "audiences": ["users"],
  "components": ["release"],
  "issues": [78],
  "breaking": false,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	admission := writeReleaseAdmission(t, repository, "0.3.0", "issue-78-release-tracking")

	exitCode := cli.Run([]string{
		"release", "prepare", "--repository", repository,
		"--version", "0.3.0", "--date", "2026-07-15", "--admission", admission,
	}, &stdout, &stderr)
	if exitCode != 1 || !strings.Contains(stderr.String(), "must be greater") {
		t.Fatalf("missing non-incrementing failure: exit=%d stderr=%q", exitCode, stderr.String())
	}
	version, err := os.ReadFile(filepath.Join(repository, "product-version.json"))
	if err != nil || !strings.Contains(string(version), `"version":"0.3.0"`) {
		t.Fatalf("refused preparation mutated product version: content=%q err=%v", version, err)
	}
	if _, err := os.Stat(filepath.Join(repository, "changes", "unreleased", "issue-78-release-tracking.json")); err != nil {
		t.Fatalf("refused preparation lost pending record: %v", err)
	}
}

func TestChangesRenderCallsOutBreakingChanges(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{
  "schema_version": 1,
  "id": "issue-78-release-tracking",
  "summary": "Remove the legacy release contract.",
  "category": "removed",
  "audiences": ["users", "developers"],
  "components": ["release"],
  "issues": [78],
  "breaking": true,
  "internal_only": false
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"changes", "render", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "**BREAKING:** Remove the legacy release contract.") {
		t.Fatalf("breaking change was not called out:\n%s", stdout.String())
	}
}

func TestReleasePrepareAdmitsOnlyExplicitRecordsAndSynchronizesContracts(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{"schema_version":1,"id":"issue-78-release-tracking","summary":"Selected.","category":"added","audiences":["users"],"components":["release"],"issues":[78],"breaking":false,"internal_only":false}`)
	writeChangeRecord(t, repository, "issue-79-later", `{"schema_version":1,"id":"issue-79-later","summary":"Later.","category":"changed","audiences":["users"],"components":["release"],"issues":[79],"breaking":false,"internal_only":false}`)
	admission := writeReleaseAdmission(t, repository, "0.4.0", "issue-78-release-tracking")
	var stdout, stderr bytes.Buffer
	if exit := cli.Run([]string{"release", "prepare", "--repository", repository, "--version", "0.4.0", "--date", "2026-07-15", "--admission", admission}, &stdout, &stderr); exit != 0 {
		t.Fatalf("prepare failed: %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repository, "changes", "unreleased", "issue-79-later.json")); err != nil {
		t.Fatalf("unselected record was not retained: %v", err)
	}
	for _, path := range []string{"plugins/codex-starter-kit/.codex-plugin/plugin.json", "plugins/codex-starter-kit/contracts/capability-model-v1.json", "plugins/codex-starter-kit/contracts/approval-boundaries-v1.json"} {
		content, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(path)))
		if err != nil || !strings.Contains(string(content), "0.4.0") {
			t.Fatalf("version surface %s not synchronized: %s %v", path, content, err)
		}
	}
}

func TestChangesRenderCanBoundSummaryToOneReleaseAndAudience(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{"schema_version":1,"id":"issue-78-release-tracking","summary":"User-visible release item.","category":"added","audiences":["users"],"components":["release"],"issues":[78],"breaking":false,"internal_only":false}`)
	admission := writeReleaseAdmission(t, repository, "0.4.0", "issue-78-release-tracking")
	var stdout, stderr bytes.Buffer
	if exit := cli.Run([]string{"release", "prepare", "--repository", repository, "--version", "0.4.0", "--date", "2026-07-15", "--admission", admission}, &stdout, &stderr); exit != 0 {
		t.Fatalf("prepare failed: %s", stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if exit := cli.Run([]string{"changes", "render", "--repository", repository, "--release", "0.4.0", "--audience", "users"}, &stdout, &stderr); exit != 0 {
		t.Fatalf("render failed: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "User-visible release item") || strings.Contains(stdout.String(), "Later.") || !strings.Contains(stdout.String(), "source-digest") {
		t.Fatalf("unexpected bounded summary:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Prepared for 2026-07-15.") || strings.Contains(stdout.String(), "Released 2026-07-15.") {
		t.Fatalf("prepared summary claimed release:\n%s", stdout.String())
	}
}

func TestChangesValidateRejectsIdentityReusedFromReleaseHistory(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{"schema_version":1,"id":"issue-78-release-tracking","summary":"Original.","category":"added","audiences":["users"],"components":["release"],"issues":[78],"breaking":false,"internal_only":false}`)
	admission := writeReleaseAdmission(t, repository, "0.4.0", "issue-78-release-tracking")
	var stdout, stderr bytes.Buffer
	if exit := cli.Run([]string{"release", "prepare", "--repository", repository, "--version", "0.4.0", "--date", "2026-07-15", "--admission", admission}, &stdout, &stderr); exit != 0 {
		t.Fatalf("prepare failed: %s", stderr.String())
	}
	writeChangeRecord(t, repository, "issue-78-release-tracking", `{"schema_version":1,"id":"issue-78-release-tracking","summary":"Reused.","category":"fixed","audiences":["users"],"components":["release"],"issues":[78],"breaking":false,"internal_only":false}`)
	stdout.Reset()
	stderr.Reset()
	if exit := cli.Run([]string{"changes", "validate", "--repository", repository}, &stdout, &stderr); exit != 1 || !strings.Contains(stderr.String(), "across release history") {
		t.Fatalf("missing historical duplicate rejection: exit=%d stderr=%s", exit, stderr.String())
	}
}

func TestReleaseRecoverRestoresDurableJournal(t *testing.T) {
	repository := t.TempDir()
	writeReleaseFixture(t, repository, `{"schema_version":1,"id":"issue-78-release-tracking","summary":"Recoverable.","category":"fixed","audiences":["users"],"components":["release"],"issues":[78],"breaking":false,"internal_only":false}`)
	versionPath := filepath.Join(repository, "product-version.json")
	original, _ := os.ReadFile(versionPath)
	if err := os.WriteFile(versionPath, []byte(`{"schema_version":1,"product":"codex-starter-kit","version":"0.4.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(repository, "changes", "releases", "0.4.0")
	if err := os.MkdirAll(archive, 0o755); err != nil {
		t.Fatal(err)
	}
	journal, _ := json.Marshal(map[string]any{"schema_version": 1, "state": "preparing", "archive_directory": "changes/releases/0.4.0", "originals": map[string]string{"product-version.json": base64.StdEncoding.EncodeToString(original)}, "absent": []string{"CHANGELOG.md"}})
	if err := os.WriteFile(filepath.Join(repository, "changes", "release-transaction.json"), journal, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exit := cli.Run([]string{"release", "recover", "--repository", repository}, &stdout, &stderr); exit != 0 || !strings.Contains(stdout.String(), "recovered") {
		t.Fatalf("recover failed: exit=%d stdout=%s stderr=%s", exit, stdout.String(), stderr.String())
	}
	restored, _ := os.ReadFile(versionPath)
	if string(restored) != string(original) {
		t.Fatalf("version was not restored: %s", restored)
	}
}

func writeReleaseFixture(t *testing.T, repository, record string) {
	t.Helper()
	for path, content := range map[string]string{
		"product-version.json":                                            `{"schema_version":1,"product":"codex-starter-kit","version":"0.3.0"}`,
		"plugins/codex-starter-kit/.codex-plugin/plugin.json":             `{"version":"0.3.0"}`,
		"plugins/codex-starter-kit/contracts/capability-model-v1.json":    `{"plugin_version":"0.3.0"}`,
		"plugins/codex-starter-kit/contracts/approval-boundaries-v1.json": `{"plugin_version":"0.3.0"}`,
		"changes/unreleased/issue-78-release-tracking.json":               record,
	} {
		fullPath := filepath.Join(repository, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("create fixture directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", path, err)
		}
	}
}

func writeReleaseAdmission(t *testing.T, repository, version string, ids ...string) string {
	t.Helper()
	content, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"version":        version,
		"milestone":      version,
		"release_issue":  900,
		"approved_by":    "product-owner",
		"records":        ids,
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repository, "changes", "admissions", version+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeChangeRecord(t *testing.T, repository, id, record string) {
	t.Helper()
	path := filepath.Join(repository, "changes", "unreleased", id+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create change-record directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(record), 0o644); err != nil {
		t.Fatalf("write change record: %v", err)
	}
}

func TestCreateCommandEmitsStructuredReconciliation(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("human work\n"), 0o644); err != nil {
		t.Fatalf("write human-owned conflict: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run(createArguments("create", repository), &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var reconciliation engine.ReconciliationRequired
	if err := json.Unmarshal(stderr.Bytes(), &reconciliation); err != nil {
		t.Fatalf("decode reconciliation JSON: %v: %s", err, stderr.String())
	}
	if len(reconciliation.Conflicts) != 1 || reconciliation.Conflicts[0].Path != "README.md" || len(reconciliation.Recovery) == 0 {
		t.Fatalf("unexpected reconciliation result: %#v", reconciliation)
	}
}

func TestApplyAndStatusCommandsUsePlanIdentifier(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	planDocument, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("encode plan: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(planPath, planDocument, 0o600); err != nil {
		t.Fatalf("write plan fixture: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{
		"apply", "--plan", planPath, "--plan-id", plan.ID,
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("apply exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var result engine.ApplyResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode apply result: %v", err)
	}
	if result.Status != engine.ApplyStatusApplied {
		t.Fatalf("apply status = %q", result.Status)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = cli.Run([]string{"status", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("status exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var status engine.RepositoryStatus
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		t.Fatalf("decode status result: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManaged {
		t.Fatalf("lifecycle = %q, want managed", status.Lifecycle)
	}
}

func TestApplyCommandPreservesStructuredConflictAndRecoveryDetails(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	document, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("encode plan: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(planPath, document, 0o600); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("new human work\n"), 0o644); err != nil {
		t.Fatalf("write post-plan conflict: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"apply", "--plan", planPath, "--plan-id", plan.ID}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var envelope struct {
		Result  engine.ApplyResult   `json:"result"`
		Failure *engine.ApplyFailure `json:"failure"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &envelope); err != nil {
		t.Fatalf("decode apply failure JSON: %v: %s", err, stderr.String())
	}
	if envelope.Failure == nil || envelope.Failure.Stage != "reconcile" || len(envelope.Failure.Conflicts) != 1 || len(envelope.Failure.Recovery) == 0 {
		t.Fatalf("structured apply failure lost reconciliation facts: %#v", envelope)
	}
}

func TestInspectAndPlanCommandsExposeEngineOperations(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"inspect", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("inspect exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var inspection engine.Inspection
	if err := json.Unmarshal(stdout.Bytes(), &inspection); err != nil {
		t.Fatalf("decode inspection: %v", err)
	}
	if !inspection.Git || inspection.Managed {
		t.Fatalf("unexpected inspection: %#v", inspection)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = cli.Run(createArguments("plan", repository), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("plan exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var plan engine.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan: %v", err)
	}
	if plan.Operation != engine.CreateOperation || plan.ID == "" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestVerifyCommandEmitsMachineReadableControlResults(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{
		"verify-plan", "--repository", repository, "--scope", "repository", "--gate", "development",
		"--actor", "integration-test", "--authority", "approved issue #27 fixture",
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("verify-plan exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var verifyPlan engine.VerifyPlan
	if err := json.Unmarshal(stdout.Bytes(), &verifyPlan); err != nil {
		t.Fatalf("decode verification plan: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "verify-plan.json")
	if err := os.WriteFile(planPath, stdout.Bytes(), 0o600); err != nil {
		t.Fatalf("write verification plan: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	exitCode = cli.Run([]string{"verify", "--plan", planPath, "--plan-id", verifyPlan.ID}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("verify exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var result engine.VerificationResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode verification result: %v", err)
	}
	if result.OverallState == engine.ControlPass || len(result.Controls) == 0 {
		t.Fatalf("unexpected verification result: %#v", result)
	}
	if result.EngineVersion != "0.3.0" {
		t.Fatalf("verification engine version = %q, want canonical product version", result.EngineVersion)
	}
}

func approvedCreate(repository string) engine.CreateRequest {
	return engine.CreateRequest{
		Repository:            repository,
		Brief:                 "Create a managed repository for CLI testing.",
		BriefApproved:         true,
		OwnerPersonaConfirmed: true,
	}
}

func createArguments(operation, repository string) []string {
	args := []string{
		operation,
		"--repository", repository,
		"--brief", "Create a managed repository for CLI testing.",
		"--approve-brief",
		"--confirm-owner-persona",
	}
	if operation == "plan" {
		args = append(args, "--operation", "create")
	}
	return args
}
