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
	if strings.Join(seeder.RequiredPermissions, ",") != "issues:write,metadata:read" {
		t.Fatalf("seeder permissions = %v", seeder.RequiredPermissions)
	}
	if strings.Join(reconciler.RequiredPermissions, ",") != "issues:write,metadata:read,organization-projects:write" {
		t.Fatalf("reconciler permissions = %v", reconciler.RequiredPermissions)
	}
}

func TestContractIntentRejectsNonCommitRevision(t *testing.T) {
	if _, err := contractIntent("main"); err == nil {
		t.Fatal("expected source revision rejection")
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
