package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestContractEmitsRedactedDeterministicWorkflow(t *testing.T) {
	request, mandate := contractFixture(t)
	requestPath := writeJSON(t, "request.json", request)
	mandatePath := writeJSON(t, "mandate.json", mandate)
	var first strings.Builder
	if err := run([]string{"--request-file", requestPath, "--mandate-file", mandatePath}, &first); err != nil {
		t.Fatal(err)
	}
	var second strings.Builder
	if err := run([]string{"--mandate-file", mandatePath, "--request-file", requestPath}, &second); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("credential-free envelope is not deterministic")
	}
	if strings.Contains(first.String(), request.Repository) || strings.Contains(first.String(), "token") || strings.Contains(first.String(), "private_key") {
		t.Fatalf("envelope leaked local or credential material: %s", first.String())
	}
	var result envelope
	if err := json.Unmarshal([]byte(first.String()), &result); err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "planned" || result.MandateID != mandate.ID || result.DeliveryResourceDigest != engine.DeliveryResourceDigest(request.Intent) || len(result.Workflow) != 5 || !result.Workflow[2].Effectful {
		t.Fatalf("unexpected envelope: %#v", result)
	}
}

func TestContractFailsClosedBeforeWorkflowOutput(t *testing.T) {
	request, mandate := contractFixture(t)
	tests := []struct {
		name   string
		mutate func(*engine.DeliveryRequest, *engine.WorkExecutionMandate)
	}{
		{"mutable source", func(request *engine.DeliveryRequest, _ *engine.WorkExecutionMandate) {
			request.Intent.SourceRevision = "main"
		}},
		{"wrong repository", func(request *engine.DeliveryRequest, _ *engine.WorkExecutionMandate) {
			request.Intent.Target.RepositoryID = "999"
		}},
		{"self review", func(request *engine.DeliveryRequest, _ *engine.WorkExecutionMandate) {
			request.Intent.Review.Role = "codex-starter-kit-labs-seeder"
		}},
		{"missing completion", func(request *engine.DeliveryRequest, _ *engine.WorkExecutionMandate) { request.CompletionIntent = nil }},
		{"broadened mandate", func(_ *engine.DeliveryRequest, mandate *engine.WorkExecutionMandate) {
			mandate.EffectKinds = append(mandate.EffectKinds, "bypass-rules")
			mandate.ID = engine.BindWorkExecutionMandate(*mandate).ID
		}},
		{"credential material", func(request *engine.DeliveryRequest, _ *engine.WorkExecutionMandate) {
			request.Intent.Title = "github_pat_forbidden"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidateRequest := request
			candidateMandate := mandate
			test.mutate(&candidateRequest, &candidateMandate)
			requestPath := writeJSON(t, "request.json", candidateRequest)
			mandatePath := writeJSON(t, "mandate.json", candidateMandate)
			var output strings.Builder
			if err := run([]string{"--request-file", requestPath, "--mandate-file", mandatePath}, &output); err == nil || output.Len() != 0 {
				t.Fatalf("unsafe input produced output: %q, %v", output.String(), err)
			}
		})
	}
}

func TestContractRejectsUnknownJSONFields(t *testing.T) {
	request, mandate := contractFixture(t)
	requestPath := writeJSON(t, "request.json", request)
	content, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	content = []byte(strings.TrimSuffix(string(content), "\n}") + ",\n\"access_token\":\"secret\"\n}\n")
	if err := os.WriteFile(requestPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	if err := run([]string{"--request-file", requestPath, "--mandate-file", writeJSON(t, "mandate.json", mandate)}, &output); err == nil || output.Len() != 0 {
		t.Fatalf("unknown field was not rejected: %q, %v", output.String(), err)
	}
}

func TestExecuteStepRejectsNonLiveManifestBeforeCredentialsOrOutput(t *testing.T) {
	request, mandate := contractFixture(t)
	requestPath := writeJSON(t, "request.json", request)
	mandatePath := writeJSON(t, "mandate.json", mandate)
	called := false
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, io.EOF
	})}
	var output strings.Builder
	getenv := func(string) string { return "-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----" }
	err := runWithDependencies(context.Background(), []string{"--request-file", requestPath, "--mandate-file", mandatePath, "--execute-step"}, getenv, client, &output)
	if err == nil || output.Len() != 0 || called || strings.Contains(err.Error(), "secret") {
		t.Fatalf("live execution did not fail closed: output=%q called=%t err=%v", output.String(), called, err)
	}
}

func TestSeederCredentialMintUsesExactRepositoryAndPermissionScope(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	secret := "installation-secret"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/app":
			io.WriteString(writer, `{"id":4319763,"slug":"codex-starter-kit-labs-seeder","owner":{"login":"codex-starter-kit-labs","id":305967668}}`)
		case "/app/installations/147094309":
			io.WriteString(writer, `{"id":147094309,"app_id":4319763,"app_slug":"codex-starter-kit-labs-seeder","account":{"login":"codex-starter-kit-labs","id":305967668},"permissions":{"contents":"write","metadata":"read","pull_requests":"write","issues":"write","workflows":"write"}}`)
		case "/app/installations/147094309/access_tokens":
			var body struct {
				RepositoryIDs []int64           `json:"repository_ids"`
				Permissions   map[string]string `json:"permissions"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			expected := map[string]string{"contents": "write", "metadata": "read", "pull_requests": "write"}
			if !slices.Equal(body.RepositoryIDs, []int64{sandboxRESTID}) || !equalMap(body.Permissions, expected) {
				t.Fatalf("mint scope = %#v / %#v", body.RepositoryIDs, body.Permissions)
			}
			io.WriteString(writer, `{"token":"`+secret+`","expires_at":"2099-01-01T00:00:00Z","permissions":{"contents":"write","metadata":"read","pull_requests":"write"},"repositories":[{"id":1303189066}]}`)
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	provider, err := githubadapter.NewAppInstallationProvider(
		appConfig(server.URL, "4319763", "147094309", "codex-starter-kit-labs-seeder", map[string]string{"contents": "write", "metadata": "read", "pull_requests": "write"}),
		githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) { return []byte(privateKey), nil }), server.Client(),
		githubadapter.WithAppCredentialClock(func() time.Time { return time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(credential)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) || strings.Contains(string(encoded), "PRIVATE KEY") || !slices.Equal(credential.Permissions, []string{"contents:write", "metadata:read", "pull-requests:write"}) {
		t.Fatalf("credential evidence leaked or broadened: %s / %#v", encoded, credential.Permissions)
	}
}

func contractFixture(t *testing.T) (engine.DeliveryRequest, engine.WorkExecutionMandate) {
	t.Helper()
	target := engine.WorkTarget{Host: sandboxHost, RepositoryID: sandboxRepository, ProjectID: sandboxProject, FieldIDs: map[string]string{"status": "field-status", "readiness": "field-readiness"}, OptionIDs: map[string]string{"status:done": "done", "readiness:ready": "ready"}}
	issue := engine.ExecutableIssueContract{
		SchemaVersion: 1, Parent: "#4",
		HumanSummary: "A fixture issue traverses governed delivery.", CurrentContext: "The sandbox provides synthetic facts.",
		GoverningReferences: "- DEC-0008 — governed delivery.", Scope: "Qualify one sandbox delivery.",
		OutOfScope: "Production effects.", Acceptance: "- [ ] The exact fixture completes.", Verification: "Inspect, plan, apply, verify, and status.",
		ReadinessAssertions: []string{"No unresolved product, architecture, policy, regulatory, or risk decision is hidden in this task.", "An authorized implementer can execute this without the originating conversation."},
	}
	source := strings.Repeat("a", 40)
	contractDigest := engine.ExecutableIssueContractDigest(issue)
	boundary := engine.WorkEffectBoundary{DataClass: "public-synthetic", CostCeiling: "zero-dollar", Destructive: "issue-75-sandbox-delivery-only", Retention: "30-day-raw-evidence", RecoveryOwner: "dragondad22"}
	intent := engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: "issue-75-live-v1", SourceRevision: source, OperatingProfileRevision: "delegated-v1",
		ManagedID: "issue:15", Title: "Contract fixture: governed delivery", Target: target, BaseBranch: "main", HeadBranch: "contract/issue-75-20260721-01",
		RequiredChecks: []string{"contract-fixture"}, Review: engine.WorkReviewRequirement{Role: "american-dragon-designs", DistinctContext: true}, MergeMethod: "squash",
		Claim:          &engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: "issue:15", SourceRevision: source, ContractDigest: contractDigest, ImplementedSources: []engine.GovernedSourceBinding{{ID: "issue", Path: "fixtures/issue-75.md", Digest: "sha256:" + strings.Repeat("b", 64)}}},
		EffectBoundary: boundary,
	}
	completion := engine.WorkDesiredIntent{
		SchemaVersion: 2, OperationID: intent.OperationID, SourceRevision: source, OperatingProfileRevision: intent.OperatingProfileRevision,
		InputDigests: map[string]string{"issue": contractDigest}, Credential: engine.WorkCredentialExpectation{Mode: "app-installation", Actor: "codex-starter-kit-labs-reconciler"}, Target: target,
		Task:       engine.DesiredManagedTask{ManagedID: intent.ManagedID, IssueType: "task", Title: intent.Title, Readiness: "ready", Status: "done", Closed: true, NoPromotionRequired: true},
		Governance: &engine.GovernedWorkContract{SchemaVersion: 1, Issue: issue, Sources: slices.Clone(intent.Claim.ImplementedSources)}, EffectBoundary: boundary,
	}
	request := engine.DeliveryRequest{Repository: t.TempDir(), Intent: intent, CompletionIntent: &completion}
	authorized := []engine.WorkExecutionAuthority{
		{Actor: "codex-starter-kit-labs-seeder", CredentialMode: "app-installation", Account: "codex-starter-kit-labs", InstallationID: "147094309", RepositoryID: sandboxRepository, Permissions: []string{"contents:write", "metadata:read", "pull-requests:write"}},
		{Actor: "codex-starter-kit-labs-reconciler", CredentialMode: "app-installation", Account: "codex-starter-kit-labs", InstallationID: "147093185", RepositoryID: sandboxRepository, Permissions: []string{"actions:read", "checks:read", "contents:read", "issues:write", "metadata:read", "organization-projects:write", "pull-requests:read", "statuses:read"}},
	}
	permissions := append(slices.Clone(authorized[0].Permissions), authorized[1].Permissions...)
	approved := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	mandate := engine.BindWorkExecutionMandate(engine.WorkExecutionMandate{
		SchemaVersion: 1, ApprovedBy: "dragondad22", ApprovalID: "issue-comment-test", ApprovedAt: approved, ExpiresAt: approved.Add(14 * 24 * time.Hour),
		Target: target, OperationID: intent.OperationID, SelectedManagedID: intent.ManagedID,
		Actors: []string{"codex-starter-kit-labs-reconciler", "codex-starter-kit-labs-seeder"}, CredentialModes: []string{"app-installation"}, Permissions: permissions, Authorities: authorized,
		OperatingProfileRevisions: []string{intent.OperatingProfileRevision}, ContractDigests: []string{contractDigest}, GovernanceDigests: []string{engine.GovernedWorkContractDigest(*completion.Governance)}, InputDigests: completion.InputDigests,
		GovernedSourceDigests: map[string]string{"issue": intent.Claim.ImplementedSources[0].Digest}, SourceRevisions: []string{source}, ManagedIDs: []string{intent.ManagedID},
		EffectKinds: []string{engine.DeliveryEffectCreateBranch, engine.DeliveryEffectCreatePullRequest, engine.DeliveryEffectMarkReady, engine.DeliveryEffectRequestReview, engine.DeliveryEffectSquashMerge, engine.DeliveryEffectReconcileCompletion, "reconcile-task"},
		Operations:  []string{"closure", "project", "readiness", "status"}, ResourceDigests: []string{engine.DeliveryResourceDigest(intent), engine.ManagedTaskResourceDigest(completion.Task)}, MaxEffects: 8,
		DataClass: boundary.DataClass, CostCeiling: boundary.CostCeiling, Destructive: boundary.Destructive, Retention: boundary.Retention, RecoveryOwner: boundary.RecoveryOwner,
	})
	return request, mandate
}

func writeJSON(t *testing.T, name string, value any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
