package githubadapter_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestDeliveryAdapterObservesExactLinkedHeadChecksReviewAndRules(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	claim := engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: "issue:75", SourceRevision: "source-1", ContractDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ImplementedSources: []engine.GovernedSourceBinding{{ID: "source", Path: "docs/implementation.md", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}
	marker, err := engine.RenderWorkDeliveryClaim(claim)
	if err != nil {
		t.Fatal(err)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/repos/octocat/example/issues":
			json.NewEncoder(writer).Encode([]any{map[string]any{"number": 75, "node_id": "I_75", "state": "open", "body": "<!-- starter-kit-managed:issue:75 -->"}})
		case "/repos/octocat/example/issues/75/timeline":
			json.NewEncoder(writer).Encode([]any{map[string]any{"event": "cross-referenced", "source": map[string]any{"issue": map[string]any{"number": 101, "repository_url": server.URL + "/repos/octocat/example", "pull_request": map[string]any{}}}}})
		case "/repos/octocat/example/pulls/101":
			json.NewEncoder(writer).Encode(map[string]any{"number": 101, "node_id": "PR_101", "state": "open", "draft": false, "body": marker, "head": map[string]any{"ref": "task/75-delivery-squash-completion", "sha": "head-1"}, "base": map[string]any{"ref": "main", "repo": map[string]any{"node_id": "R_repo"}}})
		case "/repos/octocat/example/commits/head-1/check-runs":
			json.NewEncoder(writer).Encode(map[string]any{"check_runs": []any{map[string]any{"name": "foundation", "status": "completed", "conclusion": "success", "head_sha": "head-1"}}})
		case "/repos/octocat/example/commits/head-1/status":
			json.NewEncoder(writer).Encode(map[string]any{"statuses": []any{}})
		case "/repos/octocat/example/pulls/101/reviews":
			json.NewEncoder(writer).Encode([]any{map[string]any{"id": 501, "state": "APPROVED", "commit_id": "head-1", "user": map[string]any{"login": "reviewer"}}})
		case "/repos/octocat/example/rules/branches/main":
			json.NewEncoder(writer).Encode([]any{map[string]any{"type": "required_status_checks", "parameters": map[string]any{"required_status_checks": []any{map[string]any{"context": "foundation"}}}}})
		case "/repos/octocat/example":
			json.NewEncoder(writer).Encode(map[string]any{"node_id": "R_repo", "default_branch": "main", "allow_squash_merge": true})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	base := newUserAdapter(t, server, now)
	adapter, err := githubadapter.NewDeliveryAdapter(base, []githubadapter.DeliveryReviewerTrust{{Actor: "reviewer", Capable: true, DistinctContext: true}})
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.ObserveDelivery(context.Background(), engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: "deliver-75", SourceRevision: "source-1", OperatingProfileRevision: "profile-1", ManagedID: "issue:75", Target: managedTarget(),
		BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", RequiredChecks: []string{"foundation"}, Review: engine.WorkReviewRequirement{Role: "reviewer", DistinctContext: true}, MergeMethod: "squash", Claim: &claim,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(observation.Problems) != 0 || observation.PullRequest.Number != 101 || observation.PullRequest.HeadRevision != "head-1" || len(observation.Checks) != 1 || observation.Checks[0].State != "passed" || len(observation.Reviews) != 1 || !observation.Reviews[0].Capable || observation.Rules.Revision == "" {
		t.Fatalf("delivery observation = %#v", observation)
	}
}
