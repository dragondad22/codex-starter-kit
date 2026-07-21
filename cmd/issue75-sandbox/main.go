// Command issue75-sandbox emits credential-free, content-addressed sandbox plan inputs
// for the approved Issue #75 live delivery qualification.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const (
	sandboxOwnerID      = "305967668"
	sandboxRepositoryID = "R_kgDOTa0WSg"
	sandboxRESTID       = int64(1303189066)
	sandboxProjectID    = "PVT_kwDOEjyyNM4Bdm9F"
	sandboxRepository   = "codex-starter-kit-labs/codex-starter-kit-sandbox"
	sandboxOwner        = "codex-starter-kit-labs"
	sandboxName         = "codex-starter-kit-sandbox"
	configuration       = "issue-75-sandbox-config-v1"
	runMarker           = "starter-kit-contract:issue-75-20260721-01"
	deliveryHeadBranch  = "contract/issue-75-20260721-01"
	workflowPath        = ".github/workflows/issue-75-fixture-check.yml"
)

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type planInput struct {
	Role          string                              `json:"role"`
	StageContract stageContract                       `json:"stage_contract"`
	Request       engine.SandboxRequest               `json:"request"`
	Config        githubadapter.SandboxConfig         `json:"config"`
	App           githubadapter.AppInstallationConfig `json:"app"`
	Mandate       engine.SandboxExecutionMandate      `json:"mandate"`
}

type stageContract struct {
	Stage                string   `json:"stage"`
	IdentityRequirements []string `json:"identity_requirements"`
	IdentityOutputs      []string `json:"identity_outputs"`
}

type issueIdentity struct {
	Number string
	ID     string
	NodeID string
}

type options struct {
	stage          string
	repository     string
	sourceRevision string
	approvedBy     string
	approvalID     string
	approvedAt     time.Time
	expiresAt      time.Time
	parent         issueIdentity
	delivery       issueIdentity
	dependent      issueIdentity
	pullNumber     string
	pullID         string
	pullNodeID     string
	branchHeadSHA  string
}

func main() {
	if err := run(os.Args[1:], time.Now().UTC(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string, now time.Time, output io.Writer) error {
	flags := flag.NewFlagSet("issue75-sandbox", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	stage := flags.String("stage", "", "issues-setup, relationships-setup, file-initial, file-stale, cleanup-relationships, cleanup-file, cleanup-delivery, or cleanup-issues")
	repository := flags.String("repository", ".", "local evidence repository")
	source := flags.String("source-revision", "", "exact starter-kit source revision")
	approvedBy := flags.String("approved-by", "", "approving human identity")
	approvalID := flags.String("approval-id", "", "durable approval record identity")
	approvedAt := flags.String("approved-at", "", "approval time in RFC3339 format")
	expiresAt := flags.String("expires-at", "", "authority expiry in RFC3339 format")
	parentNumber := flags.String("parent-number", "", "exact fixture parent issue number")
	parentID := flags.String("parent-id", "", "exact fixture parent database ID")
	parentNodeID := flags.String("parent-node-id", "", "exact fixture parent node ID")
	deliveryNumber := flags.String("delivery-number", "", "exact fixture delivery issue number")
	deliveryID := flags.String("delivery-id", "", "exact fixture delivery database ID")
	deliveryNodeID := flags.String("delivery-node-id", "", "exact fixture delivery node ID")
	dependentNumber := flags.String("dependent-number", "", "exact fixture dependent issue number")
	dependentID := flags.String("dependent-id", "", "exact fixture dependent database ID")
	dependentNodeID := flags.String("dependent-node-id", "", "exact fixture dependent node ID")
	pullNumber := flags.String("pull-number", "", "exact delivery pull request number")
	pullID := flags.String("pull-id", "", "exact delivery pull request database ID")
	pullNodeID := flags.String("pull-node-id", "", "exact delivery pull request node ID")
	branchHeadSHA := flags.String("branch-head-sha", "", "exact delivery branch head revision")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("valid named flags are required; positional arguments are unsupported")
	}
	approvedTime, expiryTime, err := validateApproval(*approvedBy, *approvalID, *approvedAt, *expiresAt, now)
	if err != nil {
		return err
	}
	value := options{
		stage: *stage, repository: *repository, sourceRevision: *source,
		approvedBy: *approvedBy, approvalID: *approvalID, approvedAt: approvedTime, expiresAt: expiryTime,
		parent:     issueIdentity{Number: *parentNumber, ID: *parentID, NodeID: *parentNodeID},
		delivery:   issueIdentity{Number: *deliveryNumber, ID: *deliveryID, NodeID: *deliveryNodeID},
		dependent:  issueIdentity{Number: *dependentNumber, ID: *dependentID, NodeID: *dependentNodeID},
		pullNumber: *pullNumber, pullID: *pullID, pullNodeID: *pullNodeID, branchHeadSHA: *branchHeadSHA,
	}
	input, err := buildPlanInput(value)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(input)
}

func validateApproval(approvedBy, approvalID, approvedAt, expiresAt string, now time.Time) (time.Time, time.Time, error) {
	if strings.TrimSpace(approvedBy) == "" || strings.TrimSpace(approvalID) == "" || approvedAt == "" || expiresAt == "" {
		return time.Time{}, time.Time{}, errors.New("approved-by, approval-id, approved-at, and expires-at are required")
	}
	approved, err := time.Parse(time.RFC3339, approvedAt)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("approved-at must be RFC3339")
	}
	expires, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("expires-at must be RFC3339")
	}
	if !approved.Before(expires) || now.Before(approved) || !now.Before(expires) {
		return time.Time{}, time.Time{}, errors.New("approval must already be active and expire after the current time")
	}
	return approved, expires, nil
}

func buildPlanInput(value options) (planInput, error) {
	if strings.TrimSpace(value.repository) == "" || !commitPattern.MatchString(value.sourceRevision) {
		return planInput{}, errors.New("repository and an exact lowercase 40-character source-revision are required")
	}
	role, resources, err := stageResources(value)
	if err != nil {
		return planInput{}, err
	}
	expectation, app := roleConfiguration(role, value.stage)
	target := engine.SandboxTarget{Host: "github.com", OwnerID: sandboxOwnerID, RepositoryID: sandboxRepositoryID, ProjectID: sandboxProjectID, RepositoryName: sandboxRepository}
	authority := authorityProfile(role, expectation, cleanupStage(value.stage))
	markerPrefix := runMarker
	if value.stage == "cleanup-delivery" {
		markerPrefix = "Closes #" + value.delivery.Number
	}
	manifest := engine.SandboxManifest{
		SchemaVersion: 1, OperationID: "issue-75-live-" + value.stage + "-v1", SourceRevision: value.sourceRevision,
		ConfigurationRevision: configuration, ApprovedBy: value.approvedBy, ApprovedPlan: value.approvalID,
		RecoveryOwner: value.approvedBy, MarkerPrefix: markerPrefix, Target: target, Authority: authority, Resources: resources,
	}
	config := githubadapter.SandboxConfig{
		Host: "github.com", RESTBaseURL: "https://api.github.com", GraphQLURL: "https://api.github.com/graphql", APIVersion: "2026-03-10",
		ConfigurationRevision: configuration, Target: target, RepositoryOwner: sandboxOwner, RepositoryName: sandboxName, ProjectNumber: 1, ProjectOwnerKind: "organization",
		Resources: resources, Roles: map[string]githubadapter.SandboxRoleExpectation{role: expectation}, EvidenceMode: "live", LiveTargetApproved: true,
	}
	effects := []string{"reconcile-resource"}
	if cleanupStage(value.stage) {
		effects = []string{"remove-resource"}
	}
	kinds := make([]string, 0, len(resources))
	for _, resource := range resources {
		if !slices.Contains(kinds, resource.Kind) {
			kinds = append(kinds, resource.Kind)
		}
	}
	sort.Strings(kinds)
	mandate := engine.BindSandboxExecutionMandate(engine.SandboxExecutionMandate{
		SchemaVersion: 1, ApprovedBy: value.approvedBy, ApprovalID: value.approvalID, ApprovedAt: value.approvedAt, ExpiresAt: value.expiresAt,
		Target: target, Actors: []string{expectation.Actor}, MarkerPrefix: markerPrefix, UnmarkedKeys: []string{}, ResourceKinds: kinds, EffectKinds: effects, MaxEffects: len(resources),
		DataClass: authority.DataClass, CostCeiling: authority.CostCeiling, Destructive: authority.Destructive, Retention: authority.Retention,
		RecoveryOwner: value.approvedBy, Authority: authority,
	}, resources...)
	return planInput{Role: role, StageContract: contractForStage(value.stage), Request: engine.SandboxRequest{Repository: value.repository, Manifest: manifest}, Config: config, App: app, Mandate: mandate}, nil
}

func stageResources(value options) (string, []engine.SandboxResourceSpec, error) {
	switch value.stage {
	case "issues-setup":
		return githubadapter.SandboxRoleSeeder, fixtureIssues(false, value), nil
	case "relationships-setup", "cleanup-relationships":
		if err := validateIssueIdentities(value.parent, value.delivery, value.dependent); err != nil {
			return "", nil, err
		}
		resources := relationshipResources(value)
		if value.stage == "cleanup-relationships" {
			makeAbsent(resources)
		}
		return githubadapter.SandboxRoleReconciler, resources, nil
	case "file-initial":
		return githubadapter.SandboxRoleSeeder, []engine.SandboxResourceSpec{workflowResource(deliveryHeadBranch, initialWorkflow(), false)}, nil
	case "file-stale":
		return githubadapter.SandboxRoleSeeder, []engine.SandboxResourceSpec{workflowResource(deliveryHeadBranch, finalWorkflow(), false)}, nil
	case "cleanup-file":
		return githubadapter.SandboxRoleSeeder, []engine.SandboxResourceSpec{workflowResource("main", finalWorkflow(), true)}, nil
	case "cleanup-delivery":
		if !positiveDecimal(value.delivery.Number) || !positiveDecimal(value.pullNumber) || !positiveDecimal(value.pullID) || strings.TrimSpace(value.pullNodeID) == "" || !commitPattern.MatchString(value.branchHeadSHA) {
			return "", nil, errors.New("cleanup-delivery requires exact delivery and pull request identities plus a lowercase 40-character branch-head-sha")
		}
		return githubadapter.SandboxRoleSeeder, cleanupDeliveryResources(value), nil
	case "cleanup-issues":
		if err := validateIssueIdentities(value.parent, value.delivery, value.dependent); err != nil {
			return "", nil, err
		}
		return githubadapter.SandboxRoleSeeder, fixtureIssues(true, value), nil
	default:
		return "", nil, fmt.Errorf("unsupported stage %q", value.stage)
	}
}

func contractForStage(stage string) stageContract {
	issueIdentities := []string{"parent_number", "parent_id", "parent_node_id", "delivery_number", "delivery_id", "delivery_node_id", "dependent_number", "dependent_id", "dependent_node_id"}
	contract := stageContract{Stage: stage, IdentityRequirements: []string{}, IdentityOutputs: []string{}}
	switch stage {
	case "issues-setup":
		contract.IdentityOutputs = issueIdentities
	case "relationships-setup", "cleanup-relationships", "cleanup-issues":
		contract.IdentityRequirements = issueIdentities
	case "cleanup-delivery":
		contract.IdentityRequirements = []string{"delivery_number", "pull_number", "pull_id", "pull_node_id", "branch_head_sha"}
	}
	return contract
}

func validateIssueIdentities(values ...issueIdentity) error {
	seenNumbers := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	seenNodes := map[string]struct{}{}
	for _, value := range values {
		if !positiveDecimal(value.Number) || !positiveDecimal(value.ID) || strings.TrimSpace(value.NodeID) == "" {
			return errors.New("relationship and issue cleanup stages require exact positive issue numbers, database IDs, and node IDs for parent, delivery, and dependent")
		}
		if _, exists := seenNumbers[value.Number]; exists {
			return errors.New("fixture issue identities must be distinct")
		}
		if _, exists := seenIDs[value.ID]; exists {
			return errors.New("fixture issue identities must be distinct")
		}
		if _, exists := seenNodes[value.NodeID]; exists {
			return errors.New("fixture issue identities must be distinct")
		}
		seenNumbers[value.Number], seenIDs[value.ID], seenNodes[value.NodeID] = struct{}{}, struct{}{}, struct{}{}
	}
	return nil
}

func positiveDecimal(value string) bool {
	parsed, err := strconv.ParseInt(value, 10, 64)
	return err == nil && parsed > 0 && strconv.FormatInt(parsed, 10) == value
}

func fixtureIssues(absent bool, value options) []engine.SandboxResourceSpec {
	fixtures := []struct {
		key, name, title, suffix string
		identity                 issueIdentity
	}{
		{"fixture:issue:parent", "parent", "Issue 75 contract fixture: parent", "issue:parent", value.parent},
		{"fixture:issue:delivery", "delivery", "Issue 75 contract fixture: delivery", "issue:delivery", value.delivery},
		{"fixture:issue:dependent", "dependent", "Issue 75 contract fixture: dependent", "issue:dependent", value.dependent},
	}
	resources := make([]engine.SandboxResourceSpec, 0, len(fixtures))
	for _, fixture := range fixtures {
		attributes := map[string]string{"title": fixture.title, "state": "open", "input:labels": "contract-run,type:task"}
		if absent {
			attributes["state"] = "closed"
			attributes["number"] = fixture.identity.Number
			attributes["id"] = fixture.identity.ID
			attributes["node_id"] = fixture.identity.NodeID
		}
		resources = append(resources, resource(fixture.key, engine.SandboxResourceFixtureIssue, fixture.name, runMarker+":"+fixture.suffix, attributes, absent))
	}
	return resources
}

func relationshipResources(value options) []engine.SandboxResourceSpec {
	return []engine.SandboxResourceSpec{
		relationship("relationship:parent-delivery", "parent delivery", "parent-sub-issue", value.parent, value.delivery),
		relationship("relationship:delivery-dependent", "delivery blocks dependent", "blocker-dependent", value.delivery, value.dependent),
	}
}

func cleanupDeliveryResources(value options) []engine.SandboxResourceSpec {
	claimMarker := "Closes #" + value.delivery.Number
	pull := resource("fixture:pr:delivery", engine.SandboxResourceFixturePR, "delivery", claimMarker, map[string]string{
		"title": "Contract fixture: governed delivery", "state": "closed", "draft": "false", "head": deliveryHeadBranch, "base": "main",
		"number": value.pullNumber, "id": value.pullID, "node_id": value.pullNodeID, "head_sha": value.branchHeadSHA,
	}, true)
	branch := resource("fixture:branch:delivery", engine.SandboxResourceFixtureBranch, deliveryHeadBranch, claimMarker+":branch:delivery", map[string]string{
		"sha": value.branchHeadSHA,
	}, true)
	return []engine.SandboxResourceSpec{pull, branch}
}

func relationship(key, name, kind string, source, target issueIdentity) engine.SandboxResourceSpec {
	return resource(key, engine.SandboxResourceIssueRelationship, name, runMarker, map[string]string{
		"relationship":  kind,
		"source_number": source.Number, "source_id": source.ID, "source_node_id": source.NodeID,
		"target_number": target.Number, "target_id": target.ID, "target_node_id": target.NodeID,
	}, false)
}

func workflowResource(branch, content string, absent bool) engine.SandboxResourceSpec {
	digest := sha256.Sum256([]byte(content))
	return resource("file:delivery-check", engine.SandboxResourceRepositoryFile, "issue-75-fixture-check.yml", runMarker, map[string]string{
		"path": workflowPath, "branch": branch, "content_sha256": "sha256:" + hex.EncodeToString(digest[:]), "input:content": content,
	}, absent)
}

func initialWorkflow() string {
	return "# " + runMarker + "\nname: Issue 75 contract\non:\n  pull_request:\njobs:\n  contract-delivery:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo initial-head\n"
}

func finalWorkflow() string {
	return "# " + runMarker + "\nname: Issue 75 contract\non:\n  pull_request:\njobs:\n  contract-delivery:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo stale-head-qualified\n"
}

func resource(key, kind, name, marker string, attributes map[string]string, absent bool) engine.SandboxResourceSpec {
	value := engine.SandboxResourceSpec{Key: key, Kind: kind, Name: name, Marker: marker, Attributes: attributes}
	if absent {
		value.DesiredState = engine.SandboxResourceAbsent
	}
	return value
}

func makeAbsent(resources []engine.SandboxResourceSpec) {
	for index := range resources {
		resources[index].DesiredState = engine.SandboxResourceAbsent
	}
}

func cleanupStage(stage string) bool {
	return strings.HasPrefix(stage, "cleanup-")
}

func roleConfiguration(role, stage string) (githubadapter.SandboxRoleExpectation, githubadapter.AppInstallationConfig) {
	actor, appID, installationID := "codex-starter-kit-labs-seeder", "4319763", "147094309"
	permissions := []string{"issues:write", "metadata:read"}
	tokenPermissions := map[string]string{"issues": "write", "metadata": "read"}
	if role == githubadapter.SandboxRoleReconciler {
		actor, appID, installationID = "codex-starter-kit-labs-reconciler", "4319725", "147093185"
	} else if slices.Contains([]string{"file-initial", "file-stale", "cleanup-file"}, stage) {
		permissions = []string{"contents:write", "metadata:read"}
		tokenPermissions = map[string]string{"contents": "write", "metadata": "read"}
	} else if stage == "cleanup-delivery" {
		permissions = []string{"contents:write", "metadata:read", "pull-requests:write"}
		tokenPermissions = map[string]string{"contents": "write", "metadata": "read", "pull_requests": "write"}
	}
	return githubadapter.SandboxRoleExpectation{Mode: "app-installation", Actor: actor, Account: sandboxOwner, AccountID: sandboxOwnerID, InstallationID: installationID, RequiredPermissions: permissions},
		githubadapter.AppInstallationConfig{RESTBaseURL: "https://api.github.com", APIVersion: "2026-03-10", AppID: appID, InstallationID: installationID, Actor: actor, Account: sandboxOwner, AccountID: sandboxOwnerID, RepositoryIDs: []int64{sandboxRESTID}, TokenPermissions: tokenPermissions}
}

func authorityProfile(role string, expectation githubadapter.SandboxRoleExpectation, cleanup bool) engine.SandboxAuthorityProfile {
	permissions := make([]string, 0, len(expectation.RequiredPermissions))
	for _, permission := range expectation.RequiredPermissions {
		permissions = append(permissions, role+":"+permission)
	}
	destructive := "no-delete"
	if cleanup {
		destructive = "marker-scoped-fixture-cleanup-only"
	}
	return engine.SandboxAuthorityProfile{
		CredentialIdentities: []string{githubadapter.SandboxCredentialIdentity(role, expectation)}, Permissions: permissions,
		EvidenceMode: "live", Compatibility: "github.com:api.github.com:2026-03-10:native-rest-graphql",
		DataClass: "public-synthetic", CostCeiling: "zero-dollar", Destructive: destructive, Retention: "30-day-raw-evidence",
	}
}
