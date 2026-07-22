package githubadapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestDeliveryAdapterObservesBranchAndPullRequestAbsenceSeparately(t *testing.T) {
	for _, test := range []struct {
		name         string
		branchExists bool
	}{
		{name: "branch absent"},
		{name: "branch exists without pull request", branchExists: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				switch request.URL.Path {
				case "/repos/octocat/example/issues":
					json.NewEncoder(writer).Encode([]any{map[string]any{"number": 75, "node_id": "I_75", "state": "open", "body": "<!-- starter-kit-managed:issue:75 -->"}})
				case "/repos/octocat/example/rules/branches/main":
					json.NewEncoder(writer).Encode([]any{})
				case "/repos/octocat/example":
					json.NewEncoder(writer).Encode(map[string]any{"node_id": "R_repo", "default_branch": "main", "allow_squash_merge": true})
				case "/repos/octocat/example/branches/main":
					json.NewEncoder(writer).Encode(map[string]any{"commit": map[string]any{"sha": "base-1"}})
				case "/repos/octocat/example/git/ref/heads/task/75-delivery-squash-completion":
					if !test.branchExists {
						http.NotFound(writer, request)
						return
					}
					json.NewEncoder(writer).Encode(map[string]any{"object": map[string]any{"sha": "head-1"}})
				case "/repos/octocat/example/issues/75/timeline":
					json.NewEncoder(writer).Encode([]any{})
				default:
					http.NotFound(writer, request)
				}
			}))
			defer server.Close()
			adapter, err := githubadapter.NewDeliveryAdapter(newUserAdapter(t, server, now), nil)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := adapter.ObserveDelivery(context.Background(), deliveryIntent(nil))
			if err != nil {
				t.Fatal(err)
			}
			if observation.Revision == "" || len(observation.Problems) != 0 {
				t.Fatalf("observation = %#v", observation)
			}
			if test.branchExists && (observation.Branch.Revision != "head-1" || observation.PullRequest.Number != 0) {
				t.Fatalf("branch-present observation = %#v", observation)
			}
			if !test.branchExists && observation.Branch.Revision != "" {
				t.Fatalf("branch-absent observation = %#v", observation)
			}
		})
	}
}

func TestDeliveryAdapterAppliesOrganicProgressionEffects(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	claim := deliveryIntent(nil).Claim
	requests := []string{}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		requests = append(requests, request.Method+" "+request.URL.Path)
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/repos/octocat/example/git/refs":
			json.NewEncoder(writer).Encode(map[string]any{"object": map[string]any{"sha": "base-1"}})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/octocat/example/git/ref/heads/task/75-delivery-squash-completion":
			json.NewEncoder(writer).Encode(map[string]any{"object": map[string]any{"sha": "head-1"}})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/octocat/example/pulls":
			body := bytes.Buffer{}
			body.ReadFrom(request.Body)
			if !bytes.Contains(body.Bytes(), []byte(`"draft":true`)) || !bytes.Contains(body.Bytes(), []byte("starter-kit-delivery:")) || !bytes.Contains(body.Bytes(), []byte("Closes #75")) {
				t.Errorf("create pull body = %s", body.String())
			}
			json.NewEncoder(writer).Encode(map[string]any{"number": 101})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/octocat/example/pulls/101":
			json.NewEncoder(writer).Encode(map[string]any{"number": 101, "node_id": "PR_101", "head": map[string]any{"sha": "head-1"}})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/octocat/example/pulls/101/requested_reviewers":
			json.NewEncoder(writer).Encode(map[string]any{})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	adapter, err := githubadapter.NewDeliveryAdapter(newUserAdapter(t, server, now), nil)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	effects := []engine.DeliveryEffect{
		{Kind: engine.DeliveryEffectCreateBranch, Branch: "task/75-delivery-squash-completion", HeadRevision: "base-1"},
		{Kind: engine.DeliveryEffectCreatePullRequest, Branch: "task/75-delivery-squash-completion", BaseBranch: "main", HeadRevision: "head-1", Title: "Deliver issue 75", IssueNumber: 75, Claim: claim},
		{Kind: engine.DeliveryEffectRequestReview, PullRequest: 101, HeadRevision: "head-1", Reviewer: "reviewer"},
	}
	for _, effect := range effects {
		result, err := adapter.ApplyDelivery(context.Background(), effect, expected)
		if err != nil || result.Outcome != "applied" {
			t.Fatalf("%s result = %#v, err = %v", effect.Kind, result, err)
		}
	}
	if len(requests) != 5 {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestDeliveryAdapterUsesASeparateLeastAuthorityEffectCredential(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("effect authorization = %q", request.Header.Get("Authorization"))
		}
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/octocat/example/pulls/101":
			json.NewEncoder(writer).Encode(map[string]any{"number": 101, "node_id": "PR_101", "head": map[string]any{"sha": "head-1"}})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/octocat/example/pulls/101/requested_reviewers":
			json.NewEncoder(writer).Encode(map[string]any{})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	readBase := newUserAdapter(t, server, now)
	effectConfig := adapterConfig(server, "user-token", "merger", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
	effectConfig.RequiredPermissions = []string{"pull-requests:write"}
	effectBase, err := githubadapter.New(effectConfig, credentialProvider(now, "user-token", "merger", []string{"pull-requests:write"}), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := githubadapter.NewDeliveryAdapter(readBase, nil, effectBase)
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || capability.Actor != "merger" || len(capability.Permissions) != 1 || capability.Permissions[0] != "pull-requests:write" {
		t.Fatalf("effect capability = %#v, err = %v", capability, err)
	}
	result, err := adapter.ApplyDelivery(context.Background(), engine.DeliveryEffect{Kind: engine.DeliveryEffectRequestReview, PullRequest: 101, HeadRevision: "head-1", Reviewer: "reviewer"}, capability)
	if err != nil || result.Outcome != "applied" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestDeliveryAdapterSerializesRESTMutations(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	requestedReview := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/repos/octocat/example/git/refs":
			json.NewEncoder(writer).Encode(map[string]any{})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/octocat/example/pulls/101":
			json.NewEncoder(writer).Encode(map[string]any{"number": 101, "node_id": "PR_101", "head": map[string]any{"sha": "head-1"}})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/octocat/example/pulls/101/requested_reviewers":
			requestedReview = true
			json.NewEncoder(writer).Encode(map[string]any{})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	config := adapterConfig(server, "user-token", "merger", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
	config.MutationInterval = time.Hour
	base, err := githubadapter.New(config, credentialProvider(now, "user-token", "merger", config.RequiredPermissions), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := githubadapter.NewDeliveryAdapter(base, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result, err := adapter.ApplyDelivery(context.Background(), engine.DeliveryEffect{Kind: engine.DeliveryEffectCreateBranch, Branch: "task/75", HeadRevision: "base-1"}, expected); err != nil || result.Outcome != "applied" {
		t.Fatalf("first mutation = %#v, %v", result, err)
	}
	timed, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := adapter.ApplyDelivery(timed, engine.DeliveryEffect{Kind: engine.DeliveryEffectRequestReview, PullRequest: 101, HeadRevision: "head-1", Reviewer: "reviewer"}, expected); err == nil {
		t.Fatal("second REST mutation bypassed serialized pacing")
	}
	if requestedReview {
		t.Fatal("paced review mutation reached GitHub after cancellation")
	}
}

func TestDeliveryAdapterRejectsCredentialChangedAfterCapabilityRefresh(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	defer server.Close()
	config := adapterConfig(server, "user-token", "merger", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
	calls := 0
	provider := githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		calls++
		actor := "merger"
		permissions := slices.Clone(config.RequiredPermissions)
		if calls > 1 {
			actor = "different-actor"
			permissions = append(permissions, "administration:write")
		}
		return githubadapter.Credential{Token: "token", Mode: "user-token", Actor: actor, Permissions: permissions, ExpiresAt: now.Add(time.Hour)}, nil
	})
	base, err := githubadapter.New(config, provider, server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := githubadapter.NewDeliveryAdapter(base, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.ApplyDelivery(context.Background(), engine.DeliveryEffect{Kind: engine.DeliveryEffectCreateBranch, Branch: "task/75", HeadRevision: "base-1"}, expected)
	if err == nil || result.Outcome != "denied" || requests != 0 {
		t.Fatalf("changed effect credential result = %#v, requests=%d, err=%v", result, requests, err)
	}
}

func deliveryIntent(claim *engine.WorkDeliveryClaim) engine.DeliveryIntent {
	if claim == nil {
		value := engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: "issue:75", SourceRevision: "source-1", ContractDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ImplementedSources: []engine.GovernedSourceBinding{{ID: "source", Path: "docs/implementation.md", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}
		claim = &value
	}
	return engine.DeliveryIntent{SchemaVersion: 1, OperationID: "deliver-75", SourceRevision: "source-1", OperatingProfileRevision: "profile-1", ManagedID: "issue:75", Title: "Deliver issue 75", Target: managedTarget(), BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", MergeMethod: "squash", Claim: claim}
}

func TestDeliveryAdapterObservesExactLinkedHeadChecksReviewAndRules(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	claim := engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: "issue:75", SourceRevision: "source-1", ContractDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ImplementedSources: []engine.GovernedSourceBinding{{ID: "source", Path: "docs/implementation.md", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}
	marker, err := engine.RenderWorkDeliveryClaim(claim)
	if err != nil {
		t.Fatal(err)
	}
	pullBody := "Closes #75\n\n" + marker
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/repos/octocat/example/issues":
			json.NewEncoder(writer).Encode([]any{map[string]any{"number": 75, "node_id": "I_75", "state": "open", "body": "<!-- starter-kit-managed:issue:75 -->"}})
		case "/repos/octocat/example/issues/75/timeline":
			json.NewEncoder(writer).Encode([]any{map[string]any{"event": "cross-referenced", "source": map[string]any{"issue": map[string]any{"number": 101, "repository_url": server.URL + "/repos/octocat/example", "pull_request": map[string]any{}}}}})
		case "/repos/octocat/example/git/ref/heads/task/75-delivery-squash-completion":
			json.NewEncoder(writer).Encode(map[string]any{"object": map[string]any{"sha": "head-1"}})
		case "/repos/octocat/example/pulls/101":
			json.NewEncoder(writer).Encode(map[string]any{"id": 1001, "number": 101, "node_id": "PR_101", "state": "open", "draft": false, "body": pullBody, "requested_reviewers": []any{map[string]any{"login": "reviewer"}}, "head": map[string]any{"ref": "task/75-delivery-squash-completion", "sha": "head-1"}, "base": map[string]any{"ref": "main", "repo": map[string]any{"node_id": "R_repo"}}})
		case "/repos/octocat/example/commits/head-1/check-runs":
			json.NewEncoder(writer).Encode(map[string]any{"check_runs": []any{map[string]any{"name": "foundation", "status": "completed", "conclusion": "success", "head_sha": "head-1", "app": map[string]any{"id": 15368}}}})
		case "/repos/octocat/example/commits/head-1/status":
			json.NewEncoder(writer).Encode(map[string]any{"statuses": []any{}})
		case "/repos/octocat/example/pulls/101/reviews":
			json.NewEncoder(writer).Encode([]any{map[string]any{"id": 501, "state": "APPROVED", "commit_id": "head-1", "user": map[string]any{"login": "reviewer"}}})
		case "/repos/octocat/example/rules/branches/main":
			json.NewEncoder(writer).Encode([]any{map[string]any{"type": "required_status_checks", "parameters": map[string]any{"required_status_checks": []any{map[string]any{"context": "foundation", "integration_id": 15368}}}}})
		case "/repos/octocat/example":
			json.NewEncoder(writer).Encode(map[string]any{"node_id": "R_repo", "default_branch": "main", "allow_squash_merge": true})
		case "/repos/octocat/example/branches/main":
			json.NewEncoder(writer).Encode(map[string]any{"commit": map[string]any{"sha": "base-1"}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	base := newUserAdapter(t, server, now)
	review := engine.DeliveryReviewDeclaration{Actor: "reviewer", Role: "delivery-reviewer", Capability: "governed-delivery-review", ReviewedSourceRevision: "source-1", ImplementationContext: "implementation-context", ReviewContext: "github-pull-request-review", ApprovalRoute: "github-pull-request-review", FindingsRoute: "github-pull-request-review-comments", Limitations: []string{"exact head only"}}
	adapter, err := githubadapter.NewDeliveryAdapter(base, []githubadapter.DeliveryReviewerTrust{{Declaration: review}})
	if err != nil {
		t.Fatal(err)
	}
	observation, err := adapter.ObserveDelivery(context.Background(), engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: "deliver-75", SourceRevision: "source-1", OperatingProfileRevision: "profile-1", ManagedID: "issue:75", Title: "Deliver issue 75", Target: managedTarget(),
		BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", RequiredChecks: []engine.DeliveryCheckIdentity{{Name: "foundation", IntegrationID: 15368}}, Review: review, MergeMethod: "squash", Claim: &claim,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(observation.Problems) != 0 || observation.PullRequest.ID != 1001 || observation.PullRequest.NodeID != "PR_101" || observation.PullRequest.Number != 101 || observation.PullRequest.HeadRevision != "head-1" || observation.PullRequest.ClosesIssueNumber != 75 || len(observation.Checks) != 1 || observation.Checks[0].State != "passed" || observation.Checks[0].IntegrationID != 15368 || len(observation.Reviews) != 1 || observation.Reviews[0].Capability != review.Capability || observation.Rules.Revision == "" {
		t.Fatalf("delivery observation = %#v", observation)
	}
	pullBody = marker
	unlinked, err := adapter.ObserveDelivery(context.Background(), engine.DeliveryIntent{
		SchemaVersion: 1, OperationID: "deliver-75", SourceRevision: "source-1", OperatingProfileRevision: "profile-1", ManagedID: "issue:75", Title: "Deliver issue 75", Target: managedTarget(),
		BaseBranch: "main", HeadBranch: "task/75-delivery-squash-completion", RequiredChecks: []engine.DeliveryCheckIdentity{{Name: "foundation", IntegrationID: 15368}}, Review: review, MergeMethod: "squash", Claim: &claim,
	})
	if err != nil || unlinked.PullRequest.ClosesIssueNumber != 0 || len(unlinked.Problems) == 0 {
		t.Fatalf("non-reciprocal PR linkage was accepted: %#v, %v", unlinked, err)
	}
}

func TestDeliveryAdapterRejectsObserverEffectAPIRouteMismatch(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	observer := newUserAdapter(t, server, now)
	config := adapterConfig(server, "user-token", "merger", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
	config.GraphQLURL = server.URL + "/different-graphql"
	effect, err := githubadapter.New(config, credentialProvider(now, "user-token", "merger", config.RequiredPermissions), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := githubadapter.NewDeliveryAdapter(observer, nil, effect); err == nil {
		t.Fatal("different GraphQL API route was accepted for the effect transport")
	}
}

func TestDeliveryAdapterVerifiesMergedDeliveryAfterHeadBranchDeletion(t *testing.T) {
	now := time.Date(2026, 7, 21, 21, 0, 0, 0, time.UTC)
	claim := engine.WorkDeliveryClaim{SchemaVersion: 1, ManagedID: "issue:75", SourceRevision: "source-1", ContractDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ImplementedSources: []engine.GovernedSourceBinding{{ID: "source", Path: "docs/implementation.md", Digest: "sha256:c00f126946018c4244ea7766b1087b63bd73085dc482e645981911efec70612a"}}}
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
		case "/repos/octocat/example/git/ref/heads/task/75-delivery-squash-completion":
			http.NotFound(writer, request)
		case "/repos/octocat/example/pulls/101":
			json.NewEncoder(writer).Encode(map[string]any{"number": 101, "node_id": "PR_101", "state": "closed", "merged": true, "merged_at": now, "merge_commit_sha": "merge-1", "body": "Closes #75\n\n" + marker, "head": map[string]any{"ref": "task/75-delivery-squash-completion", "sha": "head-1"}, "base": map[string]any{"ref": "main", "repo": map[string]any{"node_id": "R_repo"}}})
		case "/repos/octocat/example/pulls/101/files":
			json.NewEncoder(writer).Encode([]any{map[string]any{"filename": "docs/implementation.md", "status": "modified"}})
		case "/repos/octocat/example/commits/head-1/check-runs":
			json.NewEncoder(writer).Encode(map[string]any{"check_runs": []any{}})
		case "/repos/octocat/example/commits/head-1/status":
			json.NewEncoder(writer).Encode(map[string]any{"statuses": []any{}})
		case "/repos/octocat/example/pulls/101/reviews":
			json.NewEncoder(writer).Encode([]any{})
		case "/repos/octocat/example/rules/branches/main":
			json.NewEncoder(writer).Encode([]any{})
		case "/repos/octocat/example":
			json.NewEncoder(writer).Encode(map[string]any{"node_id": "R_repo", "default_branch": "main", "allow_squash_merge": true})
		case "/repos/octocat/example/branches/main":
			json.NewEncoder(writer).Encode(map[string]any{"commit": map[string]any{"sha": "base-2"}})
		case "/repos/octocat/example/compare/merge-1...base-2":
			json.NewEncoder(writer).Encode(map[string]any{"status": "ahead", "merge_base_commit": map[string]any{"sha": "merge-1"}})
		case "/repos/octocat/example/contents/docs/implementation.md":
			json.NewEncoder(writer).Encode(map[string]any{"type": "file", "encoding": "base64", "content": "ZGVsaXZlcmVkCg=="})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	adapter, err := githubadapter.NewDeliveryAdapter(newUserAdapter(t, server, now), nil)
	if err != nil {
		t.Fatal(err)
	}
	intent := deliveryIntent(&claim)
	intent.RequiredChecks = nil
	intent.Review = engine.DeliveryReviewDeclaration{}
	observation, err := adapter.ObserveDelivery(context.Background(), intent)
	if err != nil {
		t.Fatal(err)
	}
	if len(observation.Problems) != 0 || !observation.PullRequest.Merged || !observation.PullRequest.DefaultReachable || observation.Branch.Revision != "head-1" || observation.PullRequest.MergeMethod != "" {
		t.Fatalf("merged observation = %#v", observation)
	}
}
