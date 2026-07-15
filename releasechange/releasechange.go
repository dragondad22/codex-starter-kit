// Package releasechange validates the human-owned release records and renders generated
// communication views. It does not publish, tag, or approve a release.
package releasechange

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Record is one authoritative, human-owned description of a material change.
type Record struct {
	SchemaVersion       int      `json:"schema_version"`
	ID                  string   `json:"id"`
	Summary             string   `json:"summary"`
	Category            string   `json:"category"`
	Audiences           []string `json:"audiences"`
	Components          []string `json:"components"`
	Issues              []int    `json:"issues"`
	PullRequests        []int    `json:"pull_requests,omitempty"`
	Breaking            bool     `json:"breaking"`
	InternalOnly        bool     `json:"internal_only"`
	InternalDisposition string   `json:"internal_disposition,omitempty"`
}

// ValidationResult summarizes the current product identity and pending change records.
type ValidationResult struct {
	Version           string `json:"version"`
	UnreleasedRecords int    `json:"unreleased_records"`
	ExternalRecords   int    `json:"external_records"`
	InternalRecords   int    `json:"internal_records"`
}

// PreparationResult reports a local release-preparation transaction. Published is always
// false because publication requires a separate approved adapter and exact merged source.
type PreparationResult struct {
	Version   string `json:"version"`
	State     string `json:"state"`
	Published bool   `json:"published"`
	Records   int    `json:"records"`
}

type archivedRecord struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}

type releaseManifest struct {
	SchemaVersion   int              `json:"schema_version"`
	Version         string           `json:"version"`
	PreviousVersion string           `json:"previous_version"`
	Date            string           `json:"date"`
	Milestone       string           `json:"milestone"`
	ReleaseIssue    int              `json:"release_issue"`
	ApprovedBy      string           `json:"approved_by"`
	AdmissionSHA256 string           `json:"admission_sha256"`
	State           string           `json:"state"`
	Published       bool             `json:"published"`
	Records         []archivedRecord `json:"records"`
}

type releaseAdmission struct {
	SchemaVersion int      `json:"schema_version"`
	Version       string   `json:"version"`
	Milestone     string   `json:"milestone"`
	ReleaseIssue  int      `json:"release_issue"`
	ApprovedBy    string   `json:"approved_by"`
	Records       []string `json:"records"`
}

type productVersion struct {
	SchemaVersion int    `json:"schema_version"`
	Product       string `json:"product"`
	Version       string `json:"version"`
}

type transactionJournal struct {
	SchemaVersion    int               `json:"schema_version"`
	State            string            `json:"state"`
	ArchiveDirectory string            `json:"archive_directory"`
	Originals        map[string]string `json:"originals"`
	Absent           []string          `json:"absent"`
}

type releasedRecords struct {
	Manifest releaseManifest
	Records  []Record
}

var categoryOrder = []string{"added", "changed", "deprecated", "removed", "fixed", "security"}

var semanticVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

var safeIdentifier = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var stableReleaseVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

var categoryHeadings = map[string]string{
	"added": "Added", "changed": "Changed", "deprecated": "Deprecated",
	"removed": "Removed", "fixed": "Fixed", "security": "Security",
}

var supportedAudiences = map[string]bool{
	"developers":   true,
	"operators":    true,
	"security":     true,
	"stakeholders": true,
	"users":        true,
}

// Validate checks product/component version synchronization and every unreleased record
// without writing generated views or release state.
func Validate(repository string) (ValidationResult, error) {
	version, err := validateProductVersion(repository)
	if err != nil {
		return ValidationResult{}, err
	}
	records, err := loadUnreleased(repository)
	if err != nil {
		return ValidationResult{}, err
	}
	releases, err := loadReleases(repository)
	if err != nil {
		return ValidationResult{}, err
	}
	seen := map[string]bool{}
	for _, release := range releases {
		for _, record := range release.Records {
			seen[record.ID] = true
		}
	}
	for _, record := range records {
		if seen[record.ID] {
			return ValidationResult{}, fmt.Errorf("duplicate change record id across release history: %q", record.ID)
		}
		seen[record.ID] = true
	}
	result := ValidationResult{Version: version, UnreleasedRecords: len(records)}
	for _, record := range records {
		if record.InternalOnly {
			result.InternalRecords++
		} else {
			result.ExternalRecords++
		}
	}
	return result, nil
}

// Check validates records and requires the checked-in human changelog to equal the
// deterministic generated view.
func Check(repository string) (ValidationResult, error) {
	result, err := Validate(repository)
	if err != nil {
		return ValidationResult{}, err
	}
	expected, err := Render(repository, "")
	if err != nil {
		return ValidationResult{}, err
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("resolve repository: %w", err)
	}
	changelogPath := filepath.Join(root, "CHANGELOG.md")
	if err := ensureNoSymlinkComponents(root, changelogPath); err != nil {
		return ValidationResult{}, err
	}
	actual, err := os.ReadFile(changelogPath)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("read generated CHANGELOG.md: %w", err)
	}
	if normalizeLineEndings(string(actual)) != normalizeLineEndings(expected) {
		return ValidationResult{}, errors.New("generated CHANGELOG.md is stale; regenerate it with `starter-kit changes render --repository .`")
	}
	return result, nil
}

func normalizeLineEndings(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}

// Render returns a deterministic Markdown view of unreleased external changes. An empty
// audience includes every external audience; a named audience filters the view.
func Render(repository, audience string) (string, error) {
	if audience != "" && !supportedAudiences[audience] {
		return "", fmt.Errorf("unsupported audience %q", audience)
	}
	if _, err := validateProductVersion(repository); err != nil {
		return "", err
	}
	records, err := loadUnreleased(repository)
	if err != nil {
		return "", err
	}
	releases, err := loadReleases(repository)
	if err != nil {
		return "", err
	}
	return renderDocument(records, releases, audience), nil
}

// RenderRelease returns a bounded audience summary for one prepared or published release.
func RenderRelease(repository, version, audience string) (string, error) {
	if audience != "" && !supportedAudiences[audience] {
		return "", fmt.Errorf("unsupported audience %q", audience)
	}
	if _, err := validateProductVersion(repository); err != nil {
		return "", err
	}
	releases, err := loadReleases(repository)
	if err != nil {
		return "", err
	}
	for _, release := range releases {
		if release.Manifest.Version == version {
			return renderReleaseDocument(release, audience), nil
		}
	}
	return "", fmt.Errorf("release %q is not prepared", version)
}

func renderReleaseDocument(release releasedRecords, audience string) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Codex Starter Kit %s release notes\n\n", release.Manifest.Version)
	digestInput, _ := json.Marshal(struct {
		Release  releasedRecords `json:"release"`
		Audience string          `json:"audience"`
	}{release, audience})
	fmt.Fprintf(&output, "<!-- source-digest: sha256:%x -->\n\n", sha256.Sum256(digestInput))
	if release.Manifest.Published {
		fmt.Fprintf(&output, "Released %s.\n", release.Manifest.Date)
	} else {
		fmt.Fprintf(&output, "Prepared for %s.\n", release.Manifest.Date)
		output.WriteString("\n> Prepared locally; Git tag and GitHub Release publication are not recorded.\n")
	}
	renderRecords(&output, release.Records, audience)
	return output.String()
}

func renderDocument(records []Record, releases []releasedRecords, audience string) string {
	var output strings.Builder
	output.WriteString("# Changelog\n\n")
	output.WriteString("All notable Codex Starter Kit changes are generated from structured change records.\n\n")
	digestInput, _ := json.Marshal(struct {
		Records  []Record          `json:"records"`
		Releases []releasedRecords `json:"releases"`
		Audience string            `json:"audience"`
	}{records, releases, audience})
	fmt.Fprintf(&output, "<!-- source-digest: sha256:%x -->\n\n", sha256.Sum256(digestInput))
	output.WriteString("## [Unreleased]\n")
	renderRecords(&output, records, audience)
	for _, release := range releases {
		fmt.Fprintf(&output, "\n## [%s] - %s\n", release.Manifest.Version, release.Manifest.Date)
		if !release.Manifest.Published {
			output.WriteString("\n> Prepared locally; Git tag and GitHub Release publication are not recorded.\n")
		}
		renderRecords(&output, release.Records, audience)
	}
	return output.String()
}

func renderRecords(output *strings.Builder, records []Record, audience string) {
	for _, category := range categoryOrder {
		selected := make([]Record, 0)
		for _, record := range records {
			if record.InternalOnly || record.Category != category || audience != "" && !contains(record.Audiences, audience) {
				continue
			}
			selected = append(selected, record)
		}
		if len(selected) == 0 {
			continue
		}
		fmt.Fprintf(output, "\n### %s\n", categoryHeadings[category])
		for _, record := range selected {
			prefix := ""
			if record.Breaking {
				prefix = "**BREAKING:** "
			}
			fmt.Fprintf(output, "- %s%s%s\n", prefix, record.Summary, references(record))
		}
	}
}

// Prepare archives the exact pending records, synchronizes release-version surfaces, and
// regenerates the changelog. It performs no Git, GitHub, network, signing, or publication
// effect and never marks the release as published.
func Prepare(repository, nextVersion, releaseDate, admissionPath string) (result PreparationResult, err error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return PreparationResult{}, fmt.Errorf("resolve repository: %w", err)
	}
	journalPath := filepath.Join(root, "changes", "release-transaction.json")
	if _, statErr := os.Stat(journalPath); statErr == nil {
		return PreparationResult{}, errors.New("unfinished release transaction exists; run `starter-kit release recover --repository .`")
	} else if !os.IsNotExist(statErr) {
		return PreparationResult{}, fmt.Errorf("inspect release transaction: %w", statErr)
	}
	currentVersion, err := validateProductVersion(repository)
	if err != nil {
		return PreparationResult{}, err
	}
	if !stableReleaseVersion.MatchString(nextVersion) {
		return PreparationResult{}, fmt.Errorf("release version must be stable SemVer: %q", nextVersion)
	}
	if !versionGreater(nextVersion, currentVersion) {
		return PreparationResult{}, fmt.Errorf("release version %s must be greater than current version %s", nextVersion, currentVersion)
	}
	if _, err := time.Parse("2006-01-02", releaseDate); err != nil {
		return PreparationResult{}, fmt.Errorf("release date must use YYYY-MM-DD: %w", err)
	}
	records, err := loadUnreleased(repository)
	if err != nil {
		return PreparationResult{}, err
	}
	if len(records) == 0 {
		return PreparationResult{}, errors.New("release preparation requires at least one unreleased change record")
	}
	admission, admissionContent, err := loadAdmission(root, admissionPath, nextVersion)
	if err != nil {
		return PreparationResult{}, err
	}
	recordByID := make(map[string]Record, len(records))
	for _, record := range records {
		recordByID[record.ID] = record
	}
	selectedRecords := make([]Record, 0, len(admission.Records))
	selectedIDs := map[string]bool{}
	for _, id := range admission.Records {
		if selectedIDs[id] {
			return PreparationResult{}, fmt.Errorf("release admission contains duplicate record id %q", id)
		}
		selectedIDs[id] = true
		record, ok := recordByID[id]
		if !ok {
			return PreparationResult{}, fmt.Errorf("release admission references unavailable record %q", id)
		}
		selectedRecords = append(selectedRecords, record)
	}
	remainingRecords := make([]Record, 0, len(records)-len(selectedRecords))
	for _, record := range records {
		if !selectedIDs[record.ID] {
			remainingRecords = append(remainingRecords, record)
		}
	}
	archiveDirectory := filepath.Join(root, "changes", "releases", nextVersion)
	if err := ensureNoSymlinkComponents(root, filepath.Dir(archiveDirectory)); err != nil {
		return PreparationResult{}, err
	}
	if _, statErr := os.Stat(archiveDirectory); !os.IsNotExist(statErr) {
		if statErr != nil {
			return PreparationResult{}, fmt.Errorf("inspect release archive: %w", statErr)
		}
		return PreparationResult{}, fmt.Errorf("release archive already exists: %s", nextVersion)
	}

	originals := map[string][]byte{}
	paths, _ := filepath.Glob(filepath.Join(root, "changes", "unreleased", "*.json"))
	for _, path := range paths {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return PreparationResult{}, fmt.Errorf("read pending record for archive: %w", readErr)
		}
		originals[path] = content
	}
	versionPath := filepath.Join(root, "product-version.json")
	pluginPath := filepath.Join(root, "plugins", "codex-starter-kit", ".codex-plugin", "plugin.json")
	capabilityPath := filepath.Join(root, "plugins", "codex-starter-kit", "contracts", "capability-model-v1.json")
	approvalPath := filepath.Join(root, "plugins", "codex-starter-kit", "contracts", "approval-boundaries-v1.json")
	changelogPath := filepath.Join(root, "CHANGELOG.md")
	for _, path := range []string{versionPath, pluginPath, capabilityPath, approvalPath} {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return PreparationResult{}, fmt.Errorf("read transaction input %s: %w", filepath.Base(path), readErr)
		}
		originals[path] = content
	}
	changelogExisted := true
	if content, readErr := os.ReadFile(changelogPath); readErr == nil {
		originals[changelogPath] = content
	} else if os.IsNotExist(readErr) {
		changelogExisted = false
	} else {
		return PreparationResult{}, fmt.Errorf("read changelog transaction input: %w", readErr)
	}

	manifest := releaseManifest{
		SchemaVersion: 1, Version: nextVersion, PreviousVersion: currentVersion,
		Date: releaseDate, Milestone: admission.Milestone, ReleaseIssue: admission.ReleaseIssue,
		ApprovedBy: admission.ApprovedBy, AdmissionSHA256: fmt.Sprintf("%x", sha256.Sum256(admissionContent)),
		State: "prepared", Published: false,
		Records: make([]archivedRecord, 0, len(selectedRecords)),
	}
	for _, record := range selectedRecords {
		content := originals[filepath.Join(root, "changes", "unreleased", record.ID+".json")]
		manifest.Records = append(manifest.Records, archivedRecord{ID: record.ID, SHA256: fmt.Sprintf("%x", sha256.Sum256(content))})
	}
	manifestContent, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return PreparationResult{}, fmt.Errorf("encode release manifest: %w", err)
	}
	manifestContent = append(manifestContent, '\n')

	updatedVersions := map[string][]byte{}
	for _, path := range []string{versionPath, pluginPath, capabilityPath, approvalPath} {
		var document map[string]any
		if err := decodeOneJSON(originals[path], &document); err != nil {
			return PreparationResult{}, fmt.Errorf("decode %s for version transaction: %w", filepath.Base(path), err)
		}
		field := "plugin_version"
		if path == versionPath || path == pluginPath {
			field = "version"
		}
		document[field] = nextVersion
		content, marshalErr := json.MarshalIndent(document, "", "  ")
		if marshalErr != nil {
			return PreparationResult{}, fmt.Errorf("encode %s for version transaction: %w", filepath.Base(path), marshalErr)
		}
		updatedVersions[path] = append(content, '\n')
	}
	existingReleases, err := loadReleases(repository)
	if err != nil {
		return PreparationResult{}, err
	}
	allReleases := append([]releasedRecords{{Manifest: manifest, Records: selectedRecords}}, existingReleases...)
	changelogContent := []byte(renderDocument(remainingRecords, allReleases, ""))

	journal := transactionJournal{SchemaVersion: 1, State: "preparing", ArchiveDirectory: relativePath(root, archiveDirectory), Originals: map[string]string{}}
	for path, content := range originals {
		journal.Originals[relativePath(root, path)] = base64.StdEncoding.EncodeToString(content)
	}
	if !changelogExisted {
		journal.Absent = append(journal.Absent, relativePath(root, changelogPath))
	}
	journalContent, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return PreparationResult{}, fmt.Errorf("encode release transaction: %w", err)
	}
	if err := atomicWrite(journalPath, append(journalContent, '\n'), 0o644); err != nil {
		return PreparationResult{}, fmt.Errorf("write release transaction: %w", err)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_, _ = Recover(root)
	}()
	if err := os.MkdirAll(archiveDirectory, 0o755); err != nil {
		return PreparationResult{}, fmt.Errorf("create release archive: %w", err)
	}
	for _, record := range selectedRecords {
		sourcePath := filepath.Join(root, "changes", "unreleased", record.ID+".json")
		if err := atomicWrite(filepath.Join(archiveDirectory, record.ID+".json"), originals[sourcePath], 0o644); err != nil {
			return PreparationResult{}, fmt.Errorf("archive change record: %w", err)
		}
	}
	if err := atomicWrite(filepath.Join(archiveDirectory, "release.json"), manifestContent, 0o644); err != nil {
		return PreparationResult{}, fmt.Errorf("write release manifest: %w", err)
	}
	if err := atomicWrite(filepath.Join(archiveDirectory, "admission.json"), admissionContent, 0o644); err != nil {
		return PreparationResult{}, fmt.Errorf("archive release admission: %w", err)
	}
	for _, path := range []string{versionPath, pluginPath, capabilityPath, approvalPath} {
		if err := atomicWrite(path, updatedVersions[path], 0o644); err != nil {
			return PreparationResult{}, fmt.Errorf("write version surface %s: %w", filepath.Base(path), err)
		}
	}
	if err := atomicWrite(changelogPath, changelogContent, 0o644); err != nil {
		return PreparationResult{}, fmt.Errorf("write changelog: %w", err)
	}
	for path := range originals {
		id := strings.TrimSuffix(filepath.Base(path), ".json")
		if strings.Contains(filepath.ToSlash(path), "/changes/unreleased/") && selectedIDs[id] {
			if err := os.Remove(path); err != nil {
				return PreparationResult{}, fmt.Errorf("remove archived pending record: %w", err)
			}
		}
	}
	if err := os.Remove(journalPath); err != nil {
		return PreparationResult{}, fmt.Errorf("complete release transaction: %w", err)
	}
	committed = true
	return PreparationResult{Version: nextVersion, State: "prepared", Published: false, Records: len(selectedRecords)}, nil
}

// Recover compensates an interrupted local release-preparation transaction.
func Recover(repository string) (PreparationResult, error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return PreparationResult{}, fmt.Errorf("resolve repository: %w", err)
	}
	journalPath := filepath.Join(root, "changes", "release-transaction.json")
	content, err := os.ReadFile(journalPath)
	if err != nil {
		return PreparationResult{}, fmt.Errorf("read release transaction: %w", err)
	}
	var journal transactionJournal
	if err := decodeOneJSON(content, &journal); err != nil || journal.SchemaVersion != 1 || journal.State != "preparing" {
		return PreparationResult{}, errors.New("invalid release transaction journal; manual review required")
	}
	if !safeRelativePath(journal.ArchiveDirectory) || !strings.HasPrefix(journal.ArchiveDirectory, "changes/releases/") {
		return PreparationResult{}, errors.New("invalid release transaction archive path; manual review required")
	}
	for relative := range journal.Originals {
		if !safeRelativePath(relative) {
			return PreparationResult{}, errors.New("invalid release transaction original path; manual review required")
		}
	}
	for _, relative := range journal.Absent {
		if !safeRelativePath(relative) {
			return PreparationResult{}, errors.New("invalid release transaction absent path; manual review required")
		}
	}
	for relative, encoded := range journal.Originals {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := ensureNoSymlinkComponents(root, path); err != nil {
			return PreparationResult{}, err
		}
		original, decodeErr := base64.StdEncoding.DecodeString(encoded)
		if decodeErr != nil {
			return PreparationResult{}, fmt.Errorf("decode transaction original %s: %w", relative, decodeErr)
		}
		if err := atomicWrite(path, original, 0o644); err != nil {
			return PreparationResult{}, fmt.Errorf("restore transaction original %s: %w", relative, err)
		}
	}
	for _, relative := range journal.Absent {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := ensureNoSymlinkComponents(root, path); err != nil {
			return PreparationResult{}, err
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return PreparationResult{}, fmt.Errorf("remove transaction-created path %s: %w", relative, err)
		}
	}
	archive := filepath.Join(root, filepath.FromSlash(journal.ArchiveDirectory))
	if err := ensureNoSymlinkComponents(root, archive); err != nil {
		return PreparationResult{}, err
	}
	if err := os.RemoveAll(archive); err != nil {
		return PreparationResult{}, fmt.Errorf("remove partial release archive: %w", err)
	}
	if err := os.Remove(journalPath); err != nil {
		return PreparationResult{}, fmt.Errorf("remove recovered release transaction: %w", err)
	}
	return PreparationResult{State: "recovered", Published: false}, nil
}

func relativePath(root, path string) string {
	relative, _ := filepath.Rel(root, path)
	return filepath.ToSlash(relative)
}

func safeRelativePath(path string) bool {
	if path == "" || filepath.IsAbs(path) {
		return false
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func loadAdmission(root, admissionPath, nextVersion string) (releaseAdmission, []byte, error) {
	if admissionPath == "" {
		return releaseAdmission{}, nil, errors.New("release preparation requires an explicit admission record")
	}
	if !filepath.IsAbs(admissionPath) {
		admissionPath = filepath.Join(root, admissionPath)
	}
	if err := ensureNoSymlinkComponents(root, admissionPath); err != nil {
		return releaseAdmission{}, nil, err
	}
	content, err := os.ReadFile(admissionPath)
	if err != nil {
		return releaseAdmission{}, nil, fmt.Errorf("read release admission: %w", err)
	}
	var admission releaseAdmission
	if err := decodeOneJSON(content, &admission); err != nil {
		return releaseAdmission{}, nil, fmt.Errorf("decode release admission: %w", err)
	}
	if admission.SchemaVersion != 1 || admission.Version != nextVersion || admission.Milestone != nextVersion || admission.ReleaseIssue <= 0 || strings.TrimSpace(admission.ApprovedBy) == "" || len(admission.Records) == 0 {
		return releaseAdmission{}, nil, errors.New("release admission must bind schema v1, version/milestone, release issue, approver, and at least one record")
	}
	return admission, content, nil
}

func validateProductVersion(repository string) (string, error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return "", fmt.Errorf("resolve repository: %w", err)
	}
	versionPath := filepath.Join(root, "product-version.json")
	if err := ensureNoSymlinkComponents(root, versionPath); err != nil {
		return "", err
	}
	content, err := os.ReadFile(versionPath)
	if err != nil {
		return "", fmt.Errorf("read authoritative product-version.json: %w", err)
	}
	var product productVersion
	if err := decodeOneJSON(content, &product); err != nil {
		return "", fmt.Errorf("decode authoritative product-version.json: %w", err)
	}
	if product.SchemaVersion != 1 || product.Product != "codex-starter-kit" || !semanticVersion.MatchString(product.Version) {
		return "", errors.New("product-version.json must identify codex-starter-kit with schema v1 and valid SemVer")
	}
	surfaces := map[string]string{
		filepath.Join(root, "plugins", "codex-starter-kit", ".codex-plugin", "plugin.json"):             "version",
		filepath.Join(root, "plugins", "codex-starter-kit", "contracts", "capability-model-v1.json"):    "plugin_version",
		filepath.Join(root, "plugins", "codex-starter-kit", "contracts", "approval-boundaries-v1.json"): "plugin_version",
	}
	for path, field := range surfaces {
		if err := ensureNoSymlinkComponents(root, path); err != nil {
			return "", err
		}
		surfaceContent, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", fmt.Errorf("read version surface %s: %w", filepath.Base(path), readErr)
		}
		var surface map[string]any
		if err := decodeOneJSON(surfaceContent, &surface); err != nil {
			return "", fmt.Errorf("decode version surface %s: %w", filepath.Base(path), err)
		}
		if surface[field] != product.Version {
			return "", fmt.Errorf("component version mismatch: %s=%q, product=%q", filepath.Base(path), surface[field], product.Version)
		}
	}
	return product.Version, nil
}

func loadUnreleased(repository string) ([]Record, error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return nil, fmt.Errorf("resolve repository: %w", err)
	}
	unreleasedDirectory := filepath.Join(root, "changes", "unreleased")
	if err := ensureNoSymlinkComponents(root, unreleasedDirectory); err != nil {
		return nil, err
	}
	paths, err := filepath.Glob(filepath.Join(unreleasedDirectory, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("discover change records: %w", err)
	}
	sort.Strings(paths)
	records := make([]Record, 0, len(paths))
	seenIDs := map[string]bool{}
	for _, path := range paths {
		if err := ensureNoSymlinkComponents(root, path); err != nil {
			return nil, err
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", filepath.Base(path), readErr)
		}
		var record Record
		if decodeErr := decodeRecord(content, &record); decodeErr != nil {
			return nil, fmt.Errorf("decode %s: %w", filepath.Base(path), decodeErr)
		}
		if seenIDs[record.ID] {
			return nil, fmt.Errorf("%s: duplicate change record id %q", filepath.Base(path), record.ID)
		}
		seenIDs[record.ID] = true
		if err := validateRecord(record, strings.TrimSuffix(filepath.Base(path), ".json")); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		records = append(records, record)
	}
	return records, nil
}

func loadReleases(repository string) ([]releasedRecords, error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return nil, fmt.Errorf("resolve repository: %w", err)
	}
	releasesDirectory := filepath.Join(root, "changes", "releases")
	if err := ensureNoSymlinkComponents(root, releasesDirectory); err != nil {
		return nil, err
	}
	manifestPaths, err := filepath.Glob(filepath.Join(releasesDirectory, "*", "release.json"))
	if err != nil {
		return nil, fmt.Errorf("discover release manifests: %w", err)
	}
	releases := make([]releasedRecords, 0, len(manifestPaths))
	globalIDs := map[string]bool{}
	for _, manifestPath := range manifestPaths {
		if err := ensureNoSymlinkComponents(root, manifestPath); err != nil {
			return nil, err
		}
		content, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			return nil, fmt.Errorf("read release manifest: %w", readErr)
		}
		var manifest releaseManifest
		if decodeErr := decodeOneJSON(content, &manifest); decodeErr != nil {
			return nil, fmt.Errorf("decode %s: %w", filepath.ToSlash(manifestPath), decodeErr)
		}
		if manifest.SchemaVersion != 1 || manifest.State != "prepared" && manifest.State != "published" {
			return nil, fmt.Errorf("%s: unsupported release manifest state", filepath.ToSlash(manifestPath))
		}
		if !stableReleaseVersion.MatchString(manifest.Version) || !stableReleaseVersion.MatchString(manifest.PreviousVersion) {
			return nil, fmt.Errorf("%s: release and previous versions must be stable SemVer", filepath.ToSlash(manifestPath))
		}
		if _, dateErr := time.Parse("2006-01-02", manifest.Date); dateErr != nil {
			return nil, fmt.Errorf("%s: invalid release date: %w", filepath.ToSlash(manifestPath), dateErr)
		}
		if manifest.Published != (manifest.State == "published") {
			return nil, fmt.Errorf("%s: state and published flag disagree", filepath.ToSlash(manifestPath))
		}
		if manifest.Version != filepath.Base(filepath.Dir(manifestPath)) {
			return nil, fmt.Errorf("%s: release version must match archive directory", filepath.ToSlash(manifestPath))
		}
		admissionPath := filepath.Join(filepath.Dir(manifestPath), "admission.json")
		admissionContent, admissionErr := os.ReadFile(admissionPath)
		if admissionErr != nil {
			return nil, fmt.Errorf("read archived release admission: %w", admissionErr)
		}
		if fmt.Sprintf("%x", sha256.Sum256(admissionContent)) != manifest.AdmissionSHA256 {
			return nil, fmt.Errorf("%s: release admission digest mismatch", filepath.ToSlash(manifestPath))
		}
		var admission releaseAdmission
		if decodeErr := decodeOneJSON(admissionContent, &admission); decodeErr != nil {
			return nil, fmt.Errorf("decode archived release admission: %w", decodeErr)
		}
		if admission.SchemaVersion != 1 || admission.Version != manifest.Version || admission.Milestone != manifest.Milestone || admission.ReleaseIssue != manifest.ReleaseIssue || admission.ApprovedBy != manifest.ApprovedBy {
			return nil, fmt.Errorf("%s: admission authority does not match release manifest", filepath.ToSlash(manifestPath))
		}
		records := make([]Record, 0, len(manifest.Records))
		seen := map[string]bool{}
		for _, archived := range manifest.Records {
			if seen[archived.ID] {
				return nil, fmt.Errorf("%s: duplicate archived record id %q", filepath.ToSlash(manifestPath), archived.ID)
			}
			seen[archived.ID] = true
			if globalIDs[archived.ID] {
				return nil, fmt.Errorf("duplicate change record id across release history: %q", archived.ID)
			}
			globalIDs[archived.ID] = true
			recordPath := filepath.Join(filepath.Dir(manifestPath), archived.ID+".json")
			if err := ensureNoSymlinkComponents(root, recordPath); err != nil {
				return nil, err
			}
			recordContent, recordErr := os.ReadFile(recordPath)
			if recordErr != nil {
				return nil, fmt.Errorf("read archived record %s: %w", archived.ID, recordErr)
			}
			if digest := fmt.Sprintf("%x", sha256.Sum256(recordContent)); digest != archived.SHA256 {
				return nil, fmt.Errorf("archived record digest mismatch: %s", archived.ID)
			}
			var record Record
			if decodeErr := decodeRecord(recordContent, &record); decodeErr != nil {
				return nil, fmt.Errorf("decode archived record %s: %w", archived.ID, decodeErr)
			}
			if err := validateRecord(record, archived.ID); err != nil {
				return nil, fmt.Errorf("archived record %s: %w", archived.ID, err)
			}
			records = append(records, record)
		}
		if len(admission.Records) != len(manifest.Records) {
			return nil, fmt.Errorf("%s: admission and manifest record counts disagree", filepath.ToSlash(manifestPath))
		}
		for index, id := range admission.Records {
			if id != manifest.Records[index].ID {
				return nil, fmt.Errorf("%s: admission and manifest record identities disagree", filepath.ToSlash(manifestPath))
			}
		}
		sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
		releases = append(releases, releasedRecords{Manifest: manifest, Records: records})
	}
	sort.Slice(releases, func(i, j int) bool {
		return versionGreater(releases[i].Manifest.Version, releases[j].Manifest.Version)
	})
	return releases, nil
}

func validateRecord(record Record, filenameID string) error {
	if record.SchemaVersion != 1 {
		return errors.New("schema_version must be 1")
	}
	if !safeIdentifier.MatchString(record.ID) {
		return errors.New("id must be a safe lowercase identifier using letters, digits, and single hyphens")
	}
	if record.ID != filenameID {
		return errors.New("id must be non-empty and match the filename")
	}
	if strings.TrimSpace(record.Summary) == "" {
		return errors.New("summary must be non-empty")
	}
	if _, ok := categoryHeadings[record.Category]; !ok {
		return fmt.Errorf("unsupported category %q", record.Category)
	}
	if len(record.Components) == 0 {
		return errors.New("components must be non-empty")
	}
	seenComponents := map[string]bool{}
	for _, component := range record.Components {
		if !safeIdentifier.MatchString(component) {
			return fmt.Errorf("component must be a safe lowercase identifier: %q", component)
		}
		if seenComponents[component] {
			return fmt.Errorf("duplicate component: %q", component)
		}
		seenComponents[component] = true
	}
	seenAudiences := map[string]bool{}
	for _, audience := range record.Audiences {
		if !supportedAudiences[audience] {
			return fmt.Errorf("unsupported audience %q", audience)
		}
		if seenAudiences[audience] {
			return fmt.Errorf("duplicate audience: %q", audience)
		}
		seenAudiences[audience] = true
	}
	if len(record.Issues)+len(record.PullRequests) == 0 {
		return errors.New("at least one issue or pull request reference is required")
	}
	for _, reference := range append(append([]int{}, record.Issues...), record.PullRequests...) {
		if reference <= 0 {
			return errors.New("issue and pull request references must be positive integers")
		}
	}
	if record.InternalOnly {
		if len(record.Audiences) != 0 || strings.TrimSpace(record.InternalDisposition) == "" {
			return errors.New("internal-only records require an internal_disposition and no external audiences")
		}
	} else {
		if len(record.Audiences) == 0 {
			return errors.New("external records require at least one audience")
		}
		if strings.TrimSpace(record.InternalDisposition) != "" {
			return errors.New("external records must not declare internal_disposition")
		}
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func references(record Record) string {
	values := make([]string, 0, len(record.Issues)+len(record.PullRequests))
	for _, issue := range record.Issues {
		values = append(values, fmt.Sprintf("#%d", issue))
	}
	for _, pullRequest := range record.PullRequests {
		values = append(values, fmt.Sprintf("PR #%d", pullRequest))
	}
	if len(values) == 0 {
		return ""
	}
	return " (" + strings.Join(values, ", ") + ")"
}

func versionGreater(candidate, current string) bool {
	candidateParts := stableReleaseVersion.FindStringSubmatch(candidate)
	currentParts := stableReleaseVersion.FindStringSubmatch(current)
	if candidateParts == nil || currentParts == nil {
		return false
	}
	for index := 1; index <= 3; index++ {
		candidateNumber, _ := strconv.Atoi(candidateParts[index])
		currentNumber, _ := strconv.Atoi(currentParts[index])
		if candidateNumber != currentNumber {
			return candidateNumber > currentNumber
		}
	}
	return false
}

func atomicWrite(path string, content []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".starter-kit-release-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return os.Rename(temporaryPath, path)
	} else if err != nil {
		return err
	}
	backupFile, err := os.CreateTemp(filepath.Dir(path), ".starter-kit-release-backup-*")
	if err != nil {
		return err
	}
	backupPath := backupFile.Name()
	if err := backupFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(backupPath); err != nil {
		return err
	}
	if err := os.Rename(path, backupPath); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Rename(backupPath, path)
		return err
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("remove replacement backup: %w", err)
	}
	return nil
}

func decodeOneJSON(content []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON content")
		}
		return fmt.Errorf("trailing JSON content: %w", err)
	}
	return nil
}

func decodeRecord(content []byte, record *Record) error {
	var fields map[string]json.RawMessage
	if err := decodeOneJSON(content, &fields); err != nil {
		return err
	}
	for _, required := range []string{"schema_version", "id", "summary", "category", "audiences", "components", "issues", "breaking", "internal_only"} {
		if _, ok := fields[required]; !ok {
			return fmt.Errorf("missing required field %q", required)
		}
	}
	return decodeOneJSON(content, record)
}

func ensureNoSymlinkComponents(root, path string) error {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("release path escapes repository: %s", path)
	}
	current := root
	components := []string{"."}
	if relative != "." {
		components = strings.Split(relative, string(filepath.Separator))
	}
	for _, component := range components {
		if component != "." {
			current = filepath.Join(current, component)
		}
		info, statErr := os.Lstat(current)
		if os.IsNotExist(statErr) {
			return nil
		}
		if statErr != nil {
			return fmt.Errorf("inspect release path %s: %w", current, statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("release path traverses symbolic link: %s", current)
		}
	}
	return nil
}
