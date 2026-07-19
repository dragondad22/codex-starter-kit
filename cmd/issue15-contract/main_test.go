package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestContractIntentBindsReviewedSourceAndNativeGraph(t *testing.T) {
	intent, err := contractIntent(strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if intent.SourceRevision != strings.Repeat("a", 40) || intent.Credential.Actor != reconcilerActor {
		t.Fatalf("intent provenance = %#v", intent)
	}
	if intent.Task.ManagedID != selectedManagedID || intent.Task.ParentContext == nil || intent.Task.ParentContext.ManagedID != parentManagedID {
		t.Fatalf("parent contract = %#v", intent.Task)
	}
	if len(intent.Task.ParentContext.OtherChildren) != 1 || intent.Task.ParentContext.OtherChildren[0].ManagedID != siblingManagedID {
		t.Fatalf("sibling contract = %#v", intent.Task.ParentContext.OtherChildren)
	}
	if len(intent.Task.Dependents) != 1 || len(intent.Task.Dependents[0].Blockers) != 2 {
		t.Fatalf("dependent contract = %#v", intent.Task.Dependents)
	}
	if !intent.Task.Closed || intent.Task.Status != "next" || intent.Task.Readiness != "ready" {
		t.Fatalf("selected policy must exercise closure-derived Done: %#v", intent.Task)
	}
	if fixtureIssueStates()[selectedManagedID] != "closed" || baselineStates()[selectedManagedID] != readinessReady+":"+statusNext {
		t.Fatal("fixture must begin as an already-closed issue with stale non-Done Project status")
	}
}

func TestManagedBodyRoundTripsExactTaskMetadata(t *testing.T) {
	desired := engine.DesiredManagedTask{ManagedID: selectedManagedID, IssueType: "task", Title: "Contract fixture: selected", Readiness: "ready", Status: "next"}
	body, err := managedBody(desired)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "<!-- starter-kit-managed:"+selectedManagedID+" -->") || !strings.Contains(body, runMarker) {
		t.Fatalf("managed marker missing from %q", body)
	}
	const prefix = "<!-- starter-kit-managed-metadata:"
	start := strings.Index(body, prefix) + len(prefix)
	end := strings.Index(body[start:], " -->")
	decoded, err := base64.RawURLEncoding.DecodeString(body[start : start+end])
	if err != nil {
		t.Fatal(err)
	}
	var observed engine.DesiredManagedTask
	if err := json.Unmarshal(decoded, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.ManagedID != desired.ManagedID || observed.Status != desired.Status {
		t.Fatalf("round trip = %#v", observed)
	}
}

func TestRoleConfigurationsAreLeastPurposeSeparated(t *testing.T) {
	seeder, err := roleConfiguration("seeder")
	if err != nil {
		t.Fatal(err)
	}
	reconciler, err := roleConfiguration("reconciler")
	if err != nil {
		t.Fatal(err)
	}
	if seeder.App.Actor != seederActor || reconciler.App.Actor != reconcilerActor {
		t.Fatalf("role actors = %#v / %#v", seeder.App, reconciler.App)
	}
	if strings.Join(seeder.RequiredPermissions, ",") != "contents:write,issues:write,metadata:read,pull-requests:write,workflows:write" {
		t.Fatalf("seeder permissions = %v", seeder.RequiredPermissions)
	}
	if strings.Join(reconciler.RequiredPermissions, ",") != "actions:read,checks:read,issues:write,metadata:read,organization-projects:write,pull-requests:read,statuses:read" {
		t.Fatalf("reconciler permissions = %v", reconciler.RequiredPermissions)
	}
}

func TestContractIntentRejectsNonCommitRevision(t *testing.T) {
	if _, err := contractIntent("main"); err == nil {
		t.Fatal("expected source revision rejection")
	}
}

func TestContractMandateBindsSourceWorkflowResourcesAndLease(t *testing.T) {
	source := strings.Repeat("a", 40)
	workflow := strings.Repeat("b", 64)
	mandate, err := bindContractMandate(source, "issue-comment:123", "2026-07-19T12:00:00Z", "2026-07-21T12:00:00Z", workflow)
	if err != nil {
		t.Fatal(err)
	}
	if mandate.Digest == "" || mandate.SourceRevision != source || mandate.WorkflowDigest != workflow || mandate.ResourceDigest != contractResourceDigest() {
		t.Fatalf("mandate = %#v", mandate)
	}
	changed, err := bindContractMandate(source, "issue-comment:123", "2026-07-19T12:00:00Z", "2026-07-21T12:00:00Z", strings.Repeat("c", 64))
	if err != nil {
		t.Fatal(err)
	}
	if changed.Digest == mandate.Digest {
		t.Fatal("workflow change must change mandate identity")
	}
}

func TestExactPermissionsRejectBroadenedFixtureAuthority(t *testing.T) {
	if !exactPermissions([]string{"issues:write", "metadata:read"}, []string{"metadata:read", "issues:write"}) {
		t.Fatal("same permission set should pass")
	}
	if exactPermissions([]string{"contents:write", "issues:write", "metadata:read"}, []string{"issues:write", "metadata:read"}) {
		t.Fatal("broadened permission set must not pass")
	}
}

func TestOwnerApprovalRequiresExactOwnerAuthoredMandateFacts(t *testing.T) {
	t.Parallel()
	mandate, err := bindContractMandate(strings.Repeat("a", 40), "123", "2026-07-19T12:00:00Z", "2026-07-20T12:00:00Z", strings.Repeat("b", 64))
	if err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"starter-kit-mandate: issue-15", "decision: approved", "source_revision: " + mandate.SourceRevision,
		"workflow_digest: " + mandate.WorkflowDigest, "resource_digest: " + mandate.ResourceDigest,
		"expires_at: " + mandate.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, "\n")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/dragondad22/codex-starter-kit/issues/comments/123" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"body": body, "created_at": "2026-07-19T12:00:00Z", "author_association": "OWNER",
			"user": map[string]string{"login": "dragondad22", "type": "User"},
		})
	}))
	defer server.Close()
	if err := verifyOwnerApprovalAt(context.Background(), mandate, server.URL, server.Client()); err != nil {
		t.Fatal(err)
	}
	mandate.WorkflowDigest = strings.Repeat("c", 64)
	if err := verifyOwnerApprovalAt(context.Background(), mandate, server.URL, server.Client()); err == nil {
		t.Fatal("changed workflow digest must not reuse approval")
	}
}

func TestFixtureRelationshipsUsePinnedNativeRequestShapes(t *testing.T) {
	t.Parallel()
	requests := []struct {
		method string
		path   string
		body   string
	}{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		content, _ := io.ReadAll(request.Body)
		requests = append(requests, struct {
			method string
			path   string
			body   string
		}{request.Method, request.URL.Path, string(content)})
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet:
			_, _ = writer.Write([]byte(`[]`))
		case request.Method == http.MethodPost:
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"id":2}`))
		default:
			http.Error(writer, "unexpected", http.StatusBadRequest)
		}
	}))
	defer server.Close()
	api := fixtureAPI{client: server.Client(), token: "test", restBase: server.URL, graphQLURL: server.URL + "/graphql"}
	parent := fixtureIssue{ID: 1, Number: 11}
	child := fixtureIssue{ID: 2, Number: 12}
	dependent := fixtureIssue{ID: 3, Number: 13}
	if err := api.ensureSubIssue(context.Background(), parent, child); err != nil {
		t.Fatal(err)
	}
	if err := api.ensureDependency(context.Background(), dependent, child); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 4 {
		t.Fatalf("requests = %#v", requests)
	}
	if requests[1].method != http.MethodPost || requests[1].path != issuePath()+"/11/sub_issues" || requests[1].body != `{"sub_issue_id":2}` {
		t.Fatalf("sub-issue mutation = %#v", requests[1])
	}
	if requests[3].method != http.MethodPost || requests[3].path != issuePath()+"/13/dependencies/blocked_by" || requests[3].body != `{"issue_id":2}` {
		t.Fatalf("dependency mutation = %#v", requests[3])
	}
}

func TestFixtureGraphQLFailsClosedOnPartialResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"data":{"node":null},"errors":[{"message":"denied"}]}`))
	}))
	defer server.Close()
	api := fixtureAPI{client: server.Client(), token: "test", restBase: server.URL, graphQLURL: server.URL}
	var output any
	if err := api.graphql(context.Background(), "query{viewer{login}}", nil, &output); err == nil {
		t.Fatal("partial GraphQL response must be non-pass")
	}
}

func TestPartialRecoveryVerifiesRetiredImmutableFixture(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != issuePath()+"/12" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":2,"node_id":"I_2","number":12,"state":"closed","body":"` + fixtureTombstone + `"}`))
	}))
	defer server.Close()
	api := fixtureAPI{client: server.Client(), token: "test", restBase: server.URL, graphQLURL: server.URL + "/graphql"}
	issues := map[string]fixtureIssue{selectedManagedID: {ID: 2, NodeID: "I_2", Number: 12}}
	if err := api.verifyCleanup(context.Background(), issues); err != nil {
		t.Fatal(err)
	}
}
