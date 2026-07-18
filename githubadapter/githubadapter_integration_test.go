package githubadapter_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func TestUserTokenHandshakeReturnsBoundCapabilityWithoutExposingCredential(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer top-secret-token" {
			t.Fatalf("authorization header = %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("X-GitHub-Api-Version") != "2026-03-10" {
			t.Fatalf("API version = %q", request.Header.Get("X-GitHub-Api-Version"))
		}
		writer.Header().Set("X-RateLimit-Limit", "5000")
		writer.Header().Set("X-RateLimit-Remaining", "4990")
		writer.Header().Set("X-RateLimit-Used", "10")
		writer.Header().Set("X-RateLimit-Reset", "1784163600")
		writer.Header().Set("X-OAuth-Scopes", "repo, project")
		switch request.URL.Path {
		case "/user":
			writeJSON(t, writer, map[string]any{"login": "octocat", "type": "User"})
		case "/repos/octocat/example":
			writeJSON(t, writer, map[string]any{"node_id": "R_repo", "owner": map[string]any{"login": "octocat"}, "visibility": "public"})
		case "/graphql":
			writeJSON(t, writer, map[string]any{"data": map[string]any{"node": map[string]any{"id": "P_project", "owner": map[string]any{"login": "octocat", "__typename": "User"}, "fields": fixtureProjectFields()}, "rateLimit": map[string]any{"limit": 5000, "remaining": 4980, "resetAt": "2026-07-16T01:00:00Z"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	adapter, err := githubadapter.New(githubadapter.Config{
		Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10",
		Mode: "user-token", Actor: "octocat", ActorKind: "user",
		RepositoryOwner: "octocat", RepositoryName: "example", RepositoryID: "R_repo",
		ProjectOwner: "octocat", ProjectOwnerKind: "user", ProjectID: "P_project",
		FieldIDs:            map[string]string{"readiness": "F_readiness", "status": "F_status"},
		OptionIDs:           map[string]string{"readiness:ready": "O_ready", "status:next": "O_next"},
		RequiredPermissions: []string{"issues:write", "projects:write", "pull_requests:read"},
	}, githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		return githubadapter.Credential{Token: "top-secret-token", Mode: "user-token", Actor: "octocat", Permissions: []string{"issues:write", "projects:write", "pull_requests:read"}, ExpiresAt: now.Add(time.Hour)}, nil
	}), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}

	capability, err := adapter.Capability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !capability.Online || !capability.Fresh || capability.Mode != "user-token" || capability.Actor != "octocat" {
		t.Fatalf("unexpected capability: %#v", capability)
	}
	if capability.Host != "github.com" || capability.RepositoryID != "R_repo" || capability.ProjectID != "P_project" || capability.APIVersion != "2026-03-10" {
		t.Fatalf("capability target was not bound: %#v", capability)
	}
	encoded, err := json.Marshal(capability)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "top-secret-token") {
		t.Fatalf("capability exposed credential: %s", encoded)
	}
}

func TestAppHandshakeFollowsInstallationAndProjectFieldPagination(t *testing.T) {
	t.Parallel()
	var repositoryPages, fieldPages int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/app/installations/42" {
			writer.Header().Set("X-RateLimit-Limit", "5000")
			writer.Header().Set("X-RateLimit-Remaining", "4990")
			writer.Header().Set("X-RateLimit-Used", "10")
			writer.Header().Set("X-RateLimit-Reset", "1784163600")
		}
		switch {
		case request.URL.Path == "/app/installations/42":
			if request.Header.Get("Authorization") != "Bearer app-jwt" {
				t.Fatalf("App identity authorization = %q", request.Header.Get("Authorization"))
			}
			writeJSON(t, writer, map[string]any{"id": 42, "app_slug": "octo-work-manager", "account": map[string]any{"login": "acme"}, "target_type": "Organization"})
		case request.URL.Path == "/installation/repositories" && request.URL.Query().Get("page") == "":
			repositoryPages++
			writer.Header().Set("Link", "<"+"http://"+request.Host+"/installation/repositories?page=2>; rel=\"next\"")
			writeJSON(t, writer, map[string]any{"repositories": []any{map[string]any{"node_id": "R_other"}}})
		case request.URL.Path == "/installation/repositories" && request.URL.Query().Get("page") == "2":
			repositoryPages++
			writeJSON(t, writer, map[string]any{"repositories": []any{map[string]any{"node_id": "R_org"}}})
		case request.URL.Path == "/repos/acme/example":
			writeJSON(t, writer, map[string]any{"node_id": "R_org", "owner": map[string]any{"login": "acme"}})
		case request.URL.Path == "/graphql":
			fieldPages++
			fields := map[string]any{"nodes": []any{map[string]any{"id": "F_readiness", "options": []any{map[string]any{"id": "O_ready"}}}}, "pageInfo": map[string]any{"hasNextPage": true, "endCursor": "fields-2"}}
			if fieldPages == 2 {
				fields = map[string]any{"nodes": []any{map[string]any{"id": "F_status", "options": []any{map[string]any{"id": "O_next"}}}}, "pageInfo": map[string]any{"hasNextPage": false}}
			}
			writeJSON(t, writer, map[string]any{"data": map[string]any{"node": map[string]any{"id": "P_org", "owner": map[string]any{"login": "acme", "__typename": "Organization"}, "fields": fields}, "rateLimit": map[string]any{"limit": 5000, "remaining": 4980, "resetAt": "2026-07-16T01:00:00Z"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	config := adapterConfig(server, "app-installation", "octo-work-manager", "app", "acme", "example", "R_org", "acme", "organization", "P_org")
	config.InstallationID = "42"
	config.Account = "acme"
	adapter, err := githubadapter.New(config, credentialProvider(now, "app-installation", "octo-work-manager", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || capability.Disposition != "available" {
		t.Fatalf("paginated App capability = %#v, %v", capability, err)
	}
	if repositoryPages != 2 || fieldPages != 2 {
		t.Fatalf("pagination counts = repositories %d, fields %d", repositoryPages, fieldPages)
	}
	if capability.RESTRate == nil || capability.RESTRate.Limit != 5000 || capability.RESTRate.Used != 10 {
		t.Fatalf("REST budget did not come from installation-token authority: %#v", capability.RESTRate)
	}
}

func TestObserveFollowsRESTAndGraphQLPaginationUsingImmutableIDs(t *testing.T) {
	t.Parallel()
	graphqlPage := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/repos/octocat/example/issues" && request.URL.Query().Get("page") == "":
			writer.Header().Set("Link", "<"+"http://"+request.Host+"/repos/octocat/example/issues?page=2>; rel=\"next\"")
			writeJSON(t, writer, []any{})
		case request.URL.Path == "/repos/octocat/example/issues" && request.URL.Query().Get("page") == "2":
			writeJSON(t, writer, []any{map[string]any{
				"number": 17, "node_id": "I_issue", "title": "Managed task", "body": "<!-- starter-kit-managed:task-17 -->", "state": "open",
				"labels": []any{map[string]any{"name": "type:task"}},
			}})
		case request.URL.Path == "/graphql":
			graphqlPage++
			if graphqlPage == 1 {
				writeJSON(t, writer, map[string]any{"data": map[string]any{"node": map[string]any{"items": map[string]any{"nodes": []any{}, "pageInfo": map[string]any{"hasNextPage": true, "endCursor": "cursor-1"}}}}})
				return
			}
			writeJSON(t, writer, map[string]any{"data": map[string]any{"node": map[string]any{"items": map[string]any{"nodes": []any{map[string]any{
				"id": "PVTI_item", "content": map[string]any{"id": "I_issue"}, "fieldValues": map[string]any{"nodes": []any{
					map[string]any{"optionId": "O_ready", "field": map[string]any{"id": "F_readiness"}},
					map[string]any{"optionId": "O_next", "field": map[string]any{"id": "F_status"}},
				}},
			}}, "pageInfo": map[string]any{"hasNextPage": false, "endCursor": "cursor-2"}}}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	adapter := newUserAdapter(t, server, now)
	target := engine.WorkTarget{
		Host: "github.com", RepositoryID: "R_repo", ProjectID: "P_project",
		FieldIDs:  map[string]string{"readiness": "F_readiness", "status": "F_status"},
		OptionIDs: map[string]string{"readiness:ready": "O_ready", "status:next": "O_next"},
	}

	observation, err := adapter.Observe(context.Background(), target, "task-17")
	if err != nil {
		t.Fatal(err)
	}
	if observation.Disposition != "observed" || observation.Task == nil {
		t.Fatalf("unexpected observation: %#v", observation)
	}
	if observation.Task.IssueNodeID != "I_issue" || observation.Task.ProjectItemID != "PVTI_item" || observation.Task.ReadinessOption != "O_ready" || observation.Task.StatusOption != "O_next" {
		t.Fatalf("observation did not preserve immutable IDs: %#v", observation.Task)
	}
}

func TestLifecycleCreatesProjectsReconcilesVerifiesAndReplaysWithoutDuplicate(t *testing.T) {
	t.Parallel()
	fixture := &lifecycleFixture{fields: map[string]string{}}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	adapter := newPhaseUserAdapter(t, server, now)
	target := engine.WorkTarget{
		Host: "github.com", RepositoryID: "R_repo", ProjectID: "P_project",
		FieldIDs:  map[string]string{"readiness": "F_readiness", "status": "F_status", "phase": "F_phase"},
		OptionIDs: map[string]string{"readiness:ready": "O_ready", "status:next": "O_next", "phase:Phase 0": "O_phase_0", "phase:Phase 1": "O_phase_1", "phase:Phase 2": "O_phase_2", "phase:Phase 3": "O_phase_3", "phase:Phase 4": "O_phase_4", "phase:Phase 5": "O_phase_5", "phase:Phase 6": "O_phase_6", "phase:Phase 7": "O_phase_7", "phase:Phase 8": "O_phase_8"},
	}
	repository := t.TempDir()
	if output, err := exec.Command("git", "init", repository).CombinedOutput(); err != nil {
		t.Fatalf("initialize fixture repository: %v: %s", err, output)
	}
	request := engine.ManagedTaskRequest{Repository: repository, Intent: engine.WorkDesiredIntent{
		SchemaVersion: 1, OperationID: "operation-72", SourceRevision: "source-72", OperatingProfileRevision: "profile-1",
		InputDigests: map[string]string{"brief": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("issue-72")))},
		Credential:   engine.WorkCredentialExpectation{Mode: "user-token", Actor: "octocat"}, Target: target,
		Task: engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "feature", Title: "Reconcile one managed task", Readiness: "ready", Status: "next", Phase: "Phase 3", Review: []engine.WorkReviewRequirement{{Role: "reviewer", DistinctContext: true}}},
	}}
	lifecycle := engine.New(engine.WithClock(fixedClock{now}), engine.WithWorkAdapter(adapter))

	first, err := lifecycle.ManageTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Apply.Status != engine.WorkApplyApplied || first.Verification.OverallState != engine.ControlPass {
		t.Fatalf("first lifecycle did not converge: %#v", first)
	}
	for _, receipt := range first.Apply.Receipts {
		if receipt.EvidenceMode != "simulated" {
			t.Fatalf("deterministic adapter receipt was mislabeled: %#v", receipt)
		}
	}
	fixture.mu.Lock()
	createdAfterFirst := fixture.createCount
	mutationsAfterFirst := fixture.mutationCount
	fixture.mu.Unlock()
	if createdAfterFirst != 1 || mutationsAfterFirst == 0 {
		t.Fatalf("unexpected first effects: creates=%d mutations=%d", createdAfterFirst, mutationsAfterFirst)
	}
	if fixture.fields["F_phase"] != "O_phase_3" {
		t.Fatalf("direct Phase was not projected by immutable option ID: %v", fixture.fields)
	}

	second, err := lifecycle.ManageTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if second.Plan.NoChange != true || second.Apply.Status != engine.WorkApplyNoChange || second.Verification.OverallState != engine.ControlPass {
		t.Fatalf("replay was not a verified no-change: %#v", second)
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if fixture.createCount != createdAfterFirst || fixture.mutationCount != mutationsAfterFirst {
		t.Fatalf("replay duplicated effects: creates=%d mutations=%d", fixture.createCount, fixture.mutationCount)
	}
	state, err := os.ReadFile(repository + "/.starter-kit/work-manager/state.json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(state), "top-secret-token") {
		t.Fatal("durable lifecycle state contains the ephemeral credential")
	}
}

func TestObserveBindsInheritedPhaseToNativeParentProjectOption(t *testing.T) {
	t.Parallel()
	selected := engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "task", Title: "Child", ParentManagedID: "feature-4"}
	parent := engine.DesiredManagedTask{ManagedID: "feature-4", IssueType: "feature", Title: "Parent", Phase: "Phase 3"}
	fixture := &lifecycleFixture{
		fields: map[string]string{"F_readiness": "O_ready", "F_status": "O_next"}, projectItemID: "PVTI_item",
		parentFields: map[string]string{"F_phase": "O_phase_3"}, parentProjectItemID: "PVTI_parent",
		issue:  &githubFixtureIssue{Number: 17, NodeID: "I_issue", Title: "Child", Body: managedFixtureBody(t, selected), State: "open", Labels: []string{"type:task"}},
		parent: &githubFixtureIssue{Number: 4, NodeID: "I_parent", Title: "Parent", Body: managedFixtureBody(t, parent), State: "open", Labels: []string{"type:feature"}},
	}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	target := engine.WorkTarget{
		Host: "github.com", RepositoryID: "R_repo", ProjectID: "P_project",
		FieldIDs:  map[string]string{"readiness": "F_readiness", "status": "F_status", "phase": "F_phase"},
		OptionIDs: map[string]string{"readiness:ready": "O_ready", "status:next": "O_next", "phase:Phase 0": "O_phase_0", "phase:Phase 1": "O_phase_1", "phase:Phase 2": "O_phase_2", "phase:Phase 3": "O_phase_3", "phase:Phase 4": "O_phase_4", "phase:Phase 5": "O_phase_5", "phase:Phase 6": "O_phase_6", "phase:Phase 7": "O_phase_7", "phase:Phase 8": "O_phase_8"},
	}
	observation, err := newPhaseUserAdapter(t, server, now).Observe(context.Background(), target, "task-72")
	if err != nil || observation.Task == nil {
		t.Fatalf("native parent observation = %#v, %v", observation, err)
	}
	if observation.Task.NativeParentManagedID != "feature-4" || observation.Task.ParentPhaseOption != "O_phase_3" {
		t.Fatalf("native parent Phase was not bound by immutable identities: %#v", observation.Task)
	}
}

func TestOrganizationAppRunsTheSameManagedTaskLifecycle(t *testing.T) {
	t.Parallel()
	fixture := &lifecycleFixture{fields: map[string]string{}, app: true, repositoryOwner: "acme", repositoryID: "R_org", projectOwner: "acme", projectID: "P_org", actor: "octo-work-manager"}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	config := adapterConfig(server, "app-installation", "octo-work-manager", "app", "acme", "example", "R_org", "acme", "organization", "P_org")
	config.InstallationID = "42"
	config.Account = "acme"
	adapter, err := githubadapter.New(config, credentialProvider(now, "app-installation", "octo-work-manager", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	repository := t.TempDir()
	if output, err := exec.Command("git", "init", repository).CombinedOutput(); err != nil {
		t.Fatalf("initialize fixture repository: %v: %s", err, output)
	}
	target := managedTarget()
	target.RepositoryID = "R_org"
	target.ProjectID = "P_org"
	request := engine.ManagedTaskRequest{Repository: repository, Intent: engine.WorkDesiredIntent{
		SchemaVersion: 1, OperationID: "operation-app", SourceRevision: "source-app", OperatingProfileRevision: "profile-1",
		InputDigests: map[string]string{"brief": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("app-route")))},
		Credential:   engine.WorkCredentialExpectation{Mode: "app-installation", Actor: "octo-work-manager"}, Target: target,
		Task: engine.DesiredManagedTask{ManagedID: "task-app", IssueType: "task", Title: "App managed task", Readiness: "ready", Status: "next", Review: []engine.WorkReviewRequirement{{Role: "reviewer", DistinctContext: true}}},
	}}
	result, err := engine.New(engine.WithClock(fixedClock{now}), engine.WithWorkAdapter(adapter)).ManageTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verification.OverallState != engine.ControlPass || result.Apply.Status != engine.WorkApplyApplied {
		t.Fatalf("App route did not converge: %#v", result)
	}
}

func TestObservePreservesAmbiguousMarkerAndGraphQLPartialDataAsExplicitNonPass(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		issues      []any
		graphQL     map[string]any
		disposition string
	}{
		{
			name: "multiple markers", disposition: "ambiguous",
			issues: []any{
				map[string]any{"number": 1, "node_id": "I_1", "title": "one", "body": "<!-- starter-kit-managed:task-72 -->", "state": "open"},
				map[string]any{"number": 2, "node_id": "I_2", "title": "two", "body": "<!-- starter-kit-managed:task-72 -->", "state": "open"},
			},
		},
		{
			name: "GraphQL partial data", disposition: "needs-review",
			issues:  []any{map[string]any{"number": 1, "node_id": "I_1", "title": "one", "body": "<!-- starter-kit-managed:task-72 -->", "state": "open"}},
			graphQL: map[string]any{"data": map[string]any{"node": map[string]any{"items": map[string]any{"nodes": []any{}, "pageInfo": map[string]any{"hasNextPage": false}}}}, "errors": []any{map[string]any{"message": "field denied"}}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/repos/octocat/example/issues" {
					writeJSON(t, writer, test.issues)
					return
				}
				if request.URL.Path == "/graphql" && test.graphQL != nil {
					writeJSON(t, writer, test.graphQL)
					return
				}
				http.NotFound(writer, request)
			}))
			defer server.Close()
			now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
			observation, err := newUserAdapter(t, server, now).Observe(context.Background(), managedTarget(), "task-72")
			if err != nil {
				t.Fatal(err)
			}
			if observation.Disposition != test.disposition || len(observation.Problems) == 0 {
				t.Fatalf("unexpected explicit observation: %#v", observation)
			}
		})
	}
}

func TestApplyDistinguishesHiddenResourceAndRateLimit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		listStatus   int
		createStatus int
		headers      map[string]string
		outcome      string
	}{
		{name: "hidden resource", listStatus: http.StatusNotFound, outcome: "not-found"},
		{name: "expired authentication", listStatus: http.StatusUnauthorized, outcome: "unauthenticated"},
		{name: "insufficient permission", listStatus: http.StatusForbidden, outcome: "denied"},
		{name: "rate limited", listStatus: http.StatusOK, createStatus: http.StatusTooManyRequests, headers: map[string]string{"Retry-After": "60", "X-RateLimit-Remaining": "0"}, outcome: "rate-limited"},
		{name: "validation failure", listStatus: http.StatusOK, createStatus: http.StatusUnprocessableEntity, outcome: "validation-failed"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					writer.WriteHeader(test.listStatus)
					if test.listStatus == http.StatusOK {
						writeFixtureJSON(writer, []any{})
					}
					return
				}
				for key, value := range test.headers {
					writer.Header().Set(key, value)
				}
				writer.WriteHeader(test.createStatus)
			}))
			defer server.Close()
			now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
			result, _ := newUserAdapter(t, server, now).Apply(context.Background(), engine.WorkEffect{Kind: "create-task", ManagedID: "task-72", Marker: "starter-kit-managed:task-72", Desired: engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "task", Title: "task"}})
			if result.Outcome != test.outcome {
				t.Fatalf("outcome = %q, want %q (%#v)", result.Outcome, test.outcome, result)
			}
			if test.outcome == "rate-limited" && result.Retry == nil {
				t.Fatal("rate limit did not retain bounded retry evidence")
			}
		})
	}
}

func TestRateRetryUsesImmutableAttemptAndExponentialBound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writeFixtureJSON(writer, []any{})
			return
		}
		writer.Header().Set("X-RateLimit-Remaining", "0")
		writer.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	effect := engine.WorkEffect{Kind: "create-task", Attempt: 2, ManagedID: "task-72", Marker: "starter-kit-managed:task-72", Desired: engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "task", Title: "task"}}
	result, _ := newUserAdapter(t, server, now).Apply(context.Background(), effect)
	if result.Outcome != "rate-limited" || result.Attempt != 2 || result.Retry == nil || result.Retry.MaxAttempts != 3 || result.Retry.RetryAt.Sub(now) != 2*time.Minute {
		t.Fatalf("unexpected exponential retry: %#v", result)
	}
}

func TestReconcilePreservesHumanBodyAndUnmanagedLabels(t *testing.T) {
	t.Parallel()
	desired := engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "task", Title: "Updated title", Readiness: "ready", Status: "next"}
	fixture := &lifecycleFixture{fields: map[string]string{"F_readiness": "O_ready", "F_status": "O_next"}, projectItemID: "PVTI_item", issue: &githubFixtureIssue{
		Number: 17, NodeID: "I_issue", Title: "Old title", State: "open",
		Body:   "## Human summary\n\nKeep this brief.\n\n<!-- starter-kit-managed:task-72 -->",
		Labels: []string{"area:github", "ready-for-agent", "type:bug"},
	}}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	result, err := newUserAdapter(t, server, now).Apply(context.Background(), engine.WorkEffect{Kind: "reconcile-task", Operations: []string{"issue"}, Attempt: 1, ManagedID: "task-72", Marker: "starter-kit-managed:task-72", Desired: desired})
	if err != nil || result.Outcome != "applied" {
		t.Fatalf("reconcile = %#v, %v", result, err)
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if !strings.Contains(fixture.issue.Body, "Keep this brief.") || !strings.Contains(fixture.issue.Body, "starter-kit-managed-metadata") {
		t.Fatalf("human body was not preserved: %q", fixture.issue.Body)
	}
	joined := strings.Join(fixture.issue.Labels, ",")
	if !strings.Contains(joined, "area:github") || !strings.Contains(joined, "ready-for-agent") || !strings.Contains(joined, "type:task") || strings.Contains(joined, "type:bug") {
		t.Fatalf("labels were not reconciled safely: %v", fixture.issue.Labels)
	}
}

func TestReconcileClearsDuplicatedPhaseFromOrdinaryChild(t *testing.T) {
	t.Parallel()
	desired := engine.DesiredManagedTask{
		ManagedID: "task-72", IssueType: "task", Title: "Child task", ParentManagedID: "feature-4",
		ParentPhase: "Phase 3", Readiness: "ready", Status: "next",
	}
	fixture := &lifecycleFixture{fields: map[string]string{"F_readiness": "O_ready", "F_status": "O_next", "F_phase": "O_phase_3"}, projectItemID: "PVTI_item", issue: &githubFixtureIssue{
		Number: 17, NodeID: "I_issue", Title: "Child task", State: "open",
		Body: "<!-- starter-kit-managed:task-72 -->", Labels: []string{"type:task"},
	}}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	result, err := newPhaseUserAdapter(t, server, now).Apply(context.Background(), engine.WorkEffect{
		Kind: "reconcile-task", Operations: []string{"phase"}, Attempt: 1,
		ManagedID: "task-72", Marker: "starter-kit-managed:task-72", Desired: desired,
	})
	if err != nil || result.Outcome != "applied" {
		t.Fatalf("clear inherited Phase = %#v, %v", result, err)
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if _, exists := fixture.fields["F_phase"]; exists {
		t.Fatalf("ordinary child retained a duplicate direct Phase: %v", fixture.fields)
	}
}

func TestStaleProjectConfigurationStopsBeforeMutation(t *testing.T) {
	t.Parallel()
	server := handshakeServer(t, "octocat", "User", "octocat", "R_repo", "octocat", "P_project")
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	config := adapterConfig(server, "user-token", "octocat", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
	config.OptionIDs["status:next"] = "O_deleted"
	adapter, err := githubadapter.New(config, credentialProvider(now, "user-token", "octocat", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || capability.Disposition != "needs-review" || !strings.Contains(strings.Join(capability.Problems, " "), "stale") {
		t.Fatalf("stale configuration capability = %#v, %v", capability, err)
	}
}

func TestPhaseCatalogHandshakeRejectsDuplicateWrongTypeAndRenamedOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func([]any)
	}{
		{name: "duplicate Phase field"},
		{name: "wrong field type", mutate: func(nodes []any) {
			nodes[len(nodes)-1].(map[string]any)["dataType"] = "TEXT"
		}},
		{name: "renamed option", mutate: func(nodes []any) {
			options := nodes[len(nodes)-1].(map[string]any)["options"].([]any)
			options[0].(map[string]any)["name"] = "Phase Zero"
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fields := fixtureProjectFields()
			nodes := fields["nodes"].([]any)
			if test.name == "duplicate Phase field" {
				fields["nodes"] = append(nodes, map[string]any{"id": "F_phase_duplicate", "name": "Phase", "dataType": "SINGLE_SELECT", "options": []any{}})
			} else {
				test.mutate(nodes)
			}
			fixture := &lifecycleFixture{fields: map[string]string{}, projectFields: fields}
			server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
			defer server.Close()
			now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
			capability, err := newPhaseUserAdapter(t, server, now).Capability(context.Background())
			if err != nil || capability.Disposition != "needs-review" || !strings.Contains(strings.Join(capability.Problems, " "), "stale") {
				t.Fatalf("invalid Phase catalog must stop capability: %#v, %v", capability, err)
			}
		})
	}
}

func TestPartialProjectResultPlansOnlyRemainingSemanticOperations(t *testing.T) {
	t.Parallel()
	fixture := &lifecycleFixture{fields: map[string]string{}, failProjectAdd: true}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	adapter := newUserAdapter(t, server, now)
	repository := t.TempDir()
	if output, err := exec.Command("git", "init", repository).CombinedOutput(); err != nil {
		t.Fatalf("initialize fixture repository: %v: %s", err, output)
	}
	request := engine.ManagedTaskRequest{Repository: repository, Intent: engine.WorkDesiredIntent{
		SchemaVersion: 1, OperationID: "partial-72", SourceRevision: "source-72", OperatingProfileRevision: "profile-1",
		InputDigests: map[string]string{"brief": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("partial")))},
		Credential:   engine.WorkCredentialExpectation{Mode: "user-token", Actor: "octocat"}, Target: managedTarget(),
		Task: engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "task", Title: "Partial task", Readiness: "ready", Status: "next", Review: []engine.WorkReviewRequirement{{Role: "reviewer", DistinctContext: true}}},
	}}
	lifecycle := engine.New(engine.WithClock(fixedClock{now}), engine.WithWorkAdapter(adapter))
	first, err := lifecycle.ManageTask(context.Background(), request)
	if err != nil || first.Apply.Status != engine.WorkApplyNonPass {
		t.Fatalf("partial lifecycle = %#v, %v", first, err)
	}
	inspection, err := lifecycle.InspectManagedTask(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := lifecycle.PlanManagedTask(context.Background(), inspection)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Effects) != 1 || slices.Contains(plan.Effects[0].Operations, "issue") || slices.Contains(plan.Effects[0].Operations, "project") || !slices.Equal(plan.Effects[0].Operations, []string{"readiness", "status"}) {
		t.Fatalf("partial recovery did not contain only remaining operations: %#v", plan.Effects)
	}
}

func TestCredentialProviderDiagnosticsAreRedactedAtTheAdapterBoundary(t *testing.T) {
	t.Parallel()
	server := handshakeServer(t, "octocat", "User", "octocat", "R_repo", "octocat", "P_project")
	defer server.Close()
	config := adapterConfig(server, "user-token", "octocat", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
	adapter, err := githubadapter.New(config, githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		return githubadapter.Credential{}, fmt.Errorf("provider leaked ghp_1234567890abcdefghijklmnopqrstuvwxyz")
	}), server.Client())
	if err != nil {
		t.Fatal(err)
	}
	capability, err := adapter.Capability(context.Background())
	if err != nil || strings.Contains(strings.Join(capability.Problems, " "), "ghp_") {
		t.Fatalf("capability leaked provider diagnostics: %#v, %v", capability, err)
	}
	result, applyErr := adapter.Apply(context.Background(), engine.WorkEffect{Kind: "create-task", Attempt: 1, ManagedID: "task-72", Marker: "starter-kit-managed:task-72", Desired: engine.DesiredManagedTask{ManagedID: "task-72"}})
	if applyErr == nil || strings.Contains(applyErr.Error(), "ghp_") || strings.Contains(result.Detail, "ghp_") {
		t.Fatalf("apply leaked provider diagnostics: %#v, %v", result, applyErr)
	}
}

func TestLostCreateResponseRecoversByStableMarkerWithoutDuplicate(t *testing.T) {
	t.Parallel()
	fixture := &lifecycleFixture{fields: map[string]string{}, loseCreateResponse: true}
	server := httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	defer server.Close()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	adapter := newUserAdapter(t, server, now)
	effect := engine.WorkEffect{Kind: "create-task", ManagedID: "task-72", Marker: "starter-kit-managed:task-72", Desired: engine.DesiredManagedTask{ManagedID: "task-72", IssueType: "task", Title: "task"}}

	first, err := adapter.Apply(context.Background(), effect)
	if err == nil || first.Outcome != "offline" {
		t.Fatalf("lost response = %#v, %v", first, err)
	}
	second, err := adapter.Apply(context.Background(), effect)
	if err != nil || second.Outcome != "applied" || !strings.Contains(second.Detail, "recovered") {
		t.Fatalf("recovery = %#v, %v", second, err)
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if fixture.createCount != 1 {
		t.Fatalf("create count = %d", fixture.createCount)
	}
}

func TestIdentityModesPermissionExpiryAndUnsupportedCombinationsRemainDistinct(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC)
	t.Run("organization App installation", func(t *testing.T) {
		server := handshakeServer(t, "octo-work-manager", "App", "acme", "R_org", "acme", "P_org")
		defer server.Close()
		config := adapterConfig(server, "app-installation", "octo-work-manager", "app", "acme", "example", "R_org", "acme", "organization", "P_org")
		config.InstallationID = "42"
		config.Account = "acme"
		adapter, err := githubadapter.New(config, credentialProvider(now, "app-installation", "octo-work-manager", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
		if err != nil {
			t.Fatal(err)
		}
		capability, err := adapter.Capability(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if capability.ActorKind != "app" || capability.InstallationID != "42" || capability.ProjectOwnerKind != "organization" || capability.EvidenceMode != "simulated" {
			t.Fatalf("unexpected App capability: %#v", capability)
		}
	})

	t.Run("API installation identity mismatch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("X-RateLimit-Limit", "5000")
			writer.Header().Set("X-RateLimit-Remaining", "4990")
			writer.Header().Set("X-RateLimit-Used", "10")
			writer.Header().Set("X-RateLimit-Reset", "1784163600")
			if request.URL.Path != "/app/installations/42" {
				http.NotFound(writer, request)
				return
			}
			writeJSON(t, writer, map[string]any{"id": 99, "app_slug": "other-app", "account": map[string]any{"login": "other-account"}, "target_type": "Organization"})
		}))
		defer server.Close()
		config := adapterConfig(server, "app-installation", "octo-work-manager", "app", "acme", "example", "R_org", "acme", "organization", "P_org")
		config.InstallationID = "42"
		config.Account = "acme"
		adapter, err := githubadapter.New(config, credentialProvider(now, "app-installation", "octo-work-manager", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
		if err != nil {
			t.Fatal(err)
		}
		capability, err := adapter.Capability(context.Background())
		if err != nil || capability.Disposition != "needs-review" || !strings.Contains(strings.Join(capability.Problems, " "), "identity") {
			t.Fatalf("wrong API installation identity = %#v, %v", capability, err)
		}
	})

	t.Run("Actions token cannot become Project authority", func(t *testing.T) {
		server := handshakeServer(t, "github-actions[bot]", "Bot", "octocat", "R_repo", "octocat", "P_project")
		defer server.Close()
		config := adapterConfig(server, "actions-job", "github-actions[bot]", "bot", "octocat", "example", "R_repo", "octocat", "user", "P_project")
		adapter, err := githubadapter.New(config, credentialProvider(now, "actions-job", "github-actions[bot]", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
		if err != nil {
			t.Fatal(err)
		}
		capability, err := adapter.Capability(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if capability.Disposition != "unsupported" || len(capability.Limitations) == 0 {
			t.Fatalf("Actions limitation was not explicit: %#v", capability)
		}
	})

	t.Run("one-less permission is denied before transport", func(t *testing.T) {
		server := handshakeServer(t, "octocat", "User", "octocat", "R_repo", "octocat", "P_project")
		defer server.Close()
		config := adapterConfig(server, "user-token", "octocat", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
		adapter, err := githubadapter.New(config, credentialProvider(now, "user-token", "octocat", []string{"issues:write", "pull_requests:read"}), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
		if err != nil {
			t.Fatal(err)
		}
		capability, err := adapter.Capability(context.Background())
		if err != nil || capability.Disposition != "denied" || len(capability.Problems) == 0 {
			t.Fatalf("one-less permission capability = %#v, %v", capability, err)
		}
	})

	t.Run("expired credential requires reconnect", func(t *testing.T) {
		server := handshakeServer(t, "octocat", "User", "octocat", "R_repo", "octocat", "P_project")
		defer server.Close()
		config := adapterConfig(server, "user-token", "octocat", "user", "octocat", "example", "R_repo", "octocat", "user", "P_project")
		expiresAt := now.Add(-time.Minute)
		adapter, err := githubadapter.New(config, githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
			return githubadapter.Credential{Token: "token", Mode: "user-token", Actor: "octocat", Permissions: allPermissions(), ExpiresAt: expiresAt}, nil
		}), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
		if err != nil {
			t.Fatal(err)
		}
		stale, err := adapter.Capability(context.Background())
		if err != nil || stale.Fresh {
			t.Fatalf("expired capability = %#v, %v", stale, err)
		}
		expiresAt = now.Add(time.Hour)
		fresh, err := adapter.Capability(context.Background())
		if err != nil || !fresh.Fresh {
			t.Fatalf("reconnected capability = %#v, %v", fresh, err)
		}
	})

	t.Run("App cannot silently target user-owned Project", func(t *testing.T) {
		server := handshakeServer(t, "octo-work-manager", "App", "acme", "R_org", "octocat", "P_user")
		defer server.Close()
		config := adapterConfig(server, "app-installation", "octo-work-manager", "app", "acme", "example", "R_org", "octocat", "user", "P_user")
		config.InstallationID = "42"
		config.Account = "acme"
		if _, err := githubadapter.New(config, credentialProvider(now, "app-installation", "octo-work-manager", allPermissions()), server.Client()); err == nil || !strings.Contains(err.Error(), "organization-owned") {
			t.Fatalf("unsupported combination error = %v", err)
		}
	})
}

func allPermissions() []string {
	return []string{"issues:write", "projects:write", "pull_requests:read"}
}

func credentialProvider(now time.Time, mode, actor string, permissions []string) githubadapter.CredentialProviderFunc {
	return func(context.Context) (githubadapter.Credential, error) {
		credential := githubadapter.Credential{Token: "token", Mode: mode, Actor: actor, Permissions: permissions, ExpiresAt: now.Add(time.Hour)}
		if mode == "app-installation" {
			credential.Account = "acme"
			credential.InstallationID = "42"
			credential.IdentityToken = "app-jwt"
			credential.PermissionSource = "installation-token-response"
			credential.PermissionRevision = "sha256:fixture-installation-permissions"
		}
		return credential, nil
	}
}

func adapterConfig(server *httptest.Server, mode, actor, actorKind, repositoryOwner, repositoryName, repositoryID, projectOwner, projectOwnerKind, projectID string) githubadapter.Config {
	return githubadapter.Config{
		Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10",
		Mode: mode, Actor: actor, ActorKind: actorKind, RepositoryOwner: repositoryOwner, RepositoryName: repositoryName, RepositoryID: repositoryID,
		ProjectOwner: projectOwner, ProjectOwnerKind: projectOwnerKind, ProjectID: projectID,
		FieldIDs: map[string]string{"readiness": "F_readiness", "status": "F_status"}, OptionIDs: map[string]string{"readiness:ready": "O_ready", "status:next": "O_next"},
		RequiredPermissions: allPermissions(), EvidenceMode: "simulated",
	}
}

func handshakeServer(t *testing.T, actor, actorType, repositoryOwner, repositoryID, projectOwner, projectID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-RateLimit-Limit", "5000")
		writer.Header().Set("X-RateLimit-Remaining", "4990")
		writer.Header().Set("X-RateLimit-Used", "10")
		writer.Header().Set("X-RateLimit-Reset", "1784163600")
		writer.Header().Set("X-OAuth-Scopes", "repo, project")
		switch request.URL.Path {
		case "/user":
			writeJSON(t, writer, map[string]any{"login": actor, "type": actorType})
		case "/app/installations/42":
			writeJSON(t, writer, map[string]any{"id": 42, "app_slug": actor, "account": map[string]any{"login": projectOwner}, "target_type": "Organization"})
		case "/installation/repositories":
			writeJSON(t, writer, map[string]any{"repositories": []any{map[string]any{"node_id": repositoryID}}})
		case "/repos/" + repositoryOwner + "/example":
			writeJSON(t, writer, map[string]any{"node_id": repositoryID, "owner": map[string]any{"login": repositoryOwner}})
		case "/graphql":
			ownerKind := "User"
			if actorType == "App" {
				ownerKind = "Organization"
			}
			writeJSON(t, writer, map[string]any{"data": map[string]any{"node": map[string]any{"id": projectID, "owner": map[string]any{"login": projectOwner, "__typename": ownerKind}, "fields": fixtureProjectFields()}, "rateLimit": map[string]any{"limit": 5000, "remaining": 4980, "resetAt": "2026-07-16T01:00:00Z"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
}

func managedTarget() engine.WorkTarget {
	return engine.WorkTarget{Host: "github.com", RepositoryID: "R_repo", ProjectID: "P_project", FieldIDs: map[string]string{"readiness": "F_readiness", "status": "F_status"}, OptionIDs: map[string]string{"readiness:ready": "O_ready", "status:next": "O_next"}}
}

type lifecycleFixture struct {
	mu                  sync.Mutex
	issue               *githubFixtureIssue
	parent              *githubFixtureIssue
	projectItemID       string
	parentProjectItemID string
	fields              map[string]string
	parentFields        map[string]string
	projectFields       map[string]any
	createCount         int
	mutationCount       int
	loseCreateResponse  bool
	failProjectAdd      bool
	app                 bool
	actor               string
	repositoryOwner     string
	repositoryID        string
	projectOwner        string
	projectID           string
}

type githubFixtureIssue struct {
	Number int
	NodeID string
	Title  string
	Body   string
	State  string
	Labels []string
}

func (fixture *lifecycleFixture) serveHTTP(writer http.ResponseWriter, request *http.Request) {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	actor := fixture.actor
	if actor == "" {
		actor = "octocat"
	}
	repositoryOwner := fixture.repositoryOwner
	if repositoryOwner == "" {
		repositoryOwner = "octocat"
	}
	repositoryID := fixture.repositoryID
	if repositoryID == "" {
		repositoryID = "R_repo"
	}
	projectOwner := fixture.projectOwner
	if projectOwner == "" {
		projectOwner = "octocat"
	}
	projectID := fixture.projectID
	if projectID == "" {
		projectID = "P_project"
	}
	issuesPath := "/repos/" + repositoryOwner + "/example/issues"
	writer.Header().Set("X-RateLimit-Limit", "5000")
	writer.Header().Set("X-RateLimit-Remaining", "4990")
	writer.Header().Set("X-RateLimit-Used", "10")
	writer.Header().Set("X-RateLimit-Reset", "1784163600")
	writer.Header().Set("X-OAuth-Scopes", "repo, project")
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/user" && !fixture.app:
		writeFixtureJSON(writer, map[string]any{"login": actor, "type": "User"})
	case request.Method == http.MethodGet && request.URL.Path == "/app/installations/42" && fixture.app:
		writeFixtureJSON(writer, map[string]any{"id": 42, "app_slug": actor, "account": map[string]any{"login": projectOwner}, "target_type": "Organization"})
	case request.Method == http.MethodGet && request.URL.Path == "/installation/repositories" && fixture.app:
		writeFixtureJSON(writer, map[string]any{"repositories": []any{map[string]any{"node_id": repositoryID}}})
	case request.Method == http.MethodGet && request.URL.Path == "/repos/"+repositoryOwner+"/example":
		writeFixtureJSON(writer, map[string]any{"node_id": repositoryID, "owner": map[string]any{"login": repositoryOwner}})
	case request.Method == http.MethodGet && request.URL.Path == issuesPath:
		issues := []any{}
		if fixture.issue != nil {
			labels := []any{}
			for _, label := range fixture.issue.Labels {
				labels = append(labels, map[string]any{"name": label})
			}
			issues = append(issues, map[string]any{"number": fixture.issue.Number, "node_id": fixture.issue.NodeID, "title": fixture.issue.Title, "body": fixture.issue.Body, "state": fixture.issue.State, "labels": labels})
		}
		writeFixtureJSON(writer, issues)
	case request.Method == http.MethodGet && request.URL.Path == issuesPath+"/17/parent":
		if fixture.parent == nil {
			http.NotFound(writer, request)
			return
		}
		labels := []any{}
		for _, label := range fixture.parent.Labels {
			labels = append(labels, map[string]any{"name": label})
		}
		writeFixtureJSON(writer, map[string]any{"number": fixture.parent.Number, "node_id": fixture.parent.NodeID, "title": fixture.parent.Title, "body": fixture.parent.Body, "state": fixture.parent.State, "labels": labels})
	case request.Method == http.MethodPost && request.URL.Path == issuesPath:
		var input struct {
			Title  string   `json:"title"`
			Body   string   `json:"body"`
			Labels []string `json:"labels"`
		}
		_ = json.NewDecoder(request.Body).Decode(&input)
		fixture.issue = &githubFixtureIssue{Number: 17, NodeID: "I_issue", Title: input.Title, Body: input.Body, State: "open", Labels: input.Labels}
		fixture.createCount++
		fixture.mutationCount++
		if fixture.loseCreateResponse {
			fixture.loseCreateResponse = false
			connection, _, err := writer.(http.Hijacker).Hijack()
			if err == nil {
				_ = connection.Close()
			}
			return
		}
		writer.WriteHeader(http.StatusCreated)
		writeFixtureJSON(writer, map[string]any{"number": 17, "node_id": "I_issue", "title": input.Title, "body": input.Body, "state": "open"})
	case request.Method == http.MethodPatch && request.URL.Path == issuesPath+"/17":
		var input struct {
			Title  string   `json:"title"`
			Body   string   `json:"body"`
			State  string   `json:"state"`
			Labels []string `json:"labels"`
		}
		_ = json.NewDecoder(request.Body).Decode(&input)
		fixture.issue.Title, fixture.issue.Body, fixture.issue.State, fixture.issue.Labels = input.Title, input.Body, input.State, input.Labels
		fixture.mutationCount++
		writeFixtureJSON(writer, map[string]any{"number": 17, "node_id": "I_issue"})
	case request.Method == http.MethodPost && request.URL.Path == "/graphql":
		var input struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(request.Body).Decode(&input)
		switch {
		case strings.Contains(input.Query, "ManagedTaskProject"):
			ownerKind := "User"
			if fixture.app {
				ownerKind = "Organization"
			}
			fields := fixture.projectFields
			if fields == nil {
				fields = fixtureProjectFields()
			}
			writeFixtureJSON(writer, map[string]any{"data": map[string]any{"node": map[string]any{"id": projectID, "owner": map[string]any{"login": projectOwner, "__typename": ownerKind}, "fields": fields}, "rateLimit": map[string]any{"limit": 5000, "remaining": 4980, "resetAt": "2026-07-16T01:00:00Z"}}})
		case strings.Contains(input.Query, "ManagedTaskObservation"):
			nodes := []any{}
			if fixture.projectItemID != "" {
				fieldNodes := []any{}
				for fieldID, optionID := range fixture.fields {
					fieldNodes = append(fieldNodes, map[string]any{"optionId": optionID, "field": map[string]any{"id": fieldID}})
				}
				nodes = append(nodes, map[string]any{"id": fixture.projectItemID, "content": map[string]any{"id": "I_issue"}, "fieldValues": map[string]any{"nodes": fieldNodes}})
			}
			if fixture.parent != nil && fixture.parentProjectItemID != "" {
				fieldNodes := []any{}
				for fieldID, optionID := range fixture.parentFields {
					fieldNodes = append(fieldNodes, map[string]any{"optionId": optionID, "field": map[string]any{"id": fieldID}})
				}
				nodes = append(nodes, map[string]any{"id": fixture.parentProjectItemID, "content": map[string]any{"id": fixture.parent.NodeID}, "fieldValues": map[string]any{"nodes": fieldNodes}})
			}
			writeFixtureJSON(writer, map[string]any{"data": map[string]any{"node": map[string]any{"items": map[string]any{"nodes": nodes, "pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil}}}}})
		case strings.Contains(input.Query, "addProjectV2ItemById"):
			fixture.projectItemID = "PVTI_item"
			fixture.mutationCount++
			if fixture.failProjectAdd {
				fixture.failProjectAdd = false
				writeFixtureJSON(writer, map[string]any{"data": map[string]any{"addProjectV2ItemById": map[string]any{"item": map[string]any{"id": fixture.projectItemID}}}, "errors": []any{map[string]any{"message": "partial Project response"}}})
				return
			}
			writeFixtureJSON(writer, map[string]any{"data": map[string]any{"addProjectV2ItemById": map[string]any{"item": map[string]any{"id": fixture.projectItemID}}}})
		case strings.Contains(input.Query, "updateProjectV2ItemFieldValue"):
			fieldID, _ := input.Variables["field"].(string)
			optionID, _ := input.Variables["option"].(string)
			fixture.fields[fieldID] = optionID
			fixture.mutationCount++
			writeFixtureJSON(writer, map[string]any{"data": map[string]any{"update": map[string]any{"projectV2Item": map[string]any{"id": fixture.projectItemID}}}})
		case strings.Contains(input.Query, "clearProjectV2ItemFieldValue"):
			fieldID, _ := input.Variables["field"].(string)
			delete(fixture.fields, fieldID)
			fixture.mutationCount++
			writeFixtureJSON(writer, map[string]any{"data": map[string]any{"update": map[string]any{"projectV2Item": map[string]any{"id": fixture.projectItemID}}}})
		default:
			http.Error(writer, "unknown GraphQL operation", http.StatusBadRequest)
		}
	default:
		http.NotFound(writer, request)
	}
}

func writeFixtureJSON(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(value)
}

func managedFixtureBody(t *testing.T, desired engine.DesiredManagedTask) string {
	t.Helper()
	encoded, err := json.Marshal(desired)
	if err != nil {
		t.Fatal(err)
	}
	return "<!-- starter-kit-managed:" + desired.ManagedID + " -->\n<!-- starter-kit-managed-metadata:" + base64.RawURLEncoding.EncodeToString(encoded) + " -->"
}

func fixtureProjectFields() map[string]any {
	phaseOptions := []any{}
	for index := 0; index <= 8; index++ {
		phaseOptions = append(phaseOptions, map[string]any{"id": fmt.Sprintf("O_phase_%d", index), "name": fmt.Sprintf("Phase %d", index)})
	}
	return map[string]any{"nodes": []any{
		map[string]any{"id": "F_readiness", "options": []any{map[string]any{"id": "O_ready"}}},
		map[string]any{"id": "F_status", "options": []any{map[string]any{"id": "O_next"}}},
		map[string]any{"id": "F_phase", "name": "Phase", "dataType": "SINGLE_SELECT", "options": phaseOptions},
	}}
}

func newUserAdapter(t *testing.T, server *httptest.Server, now time.Time) *githubadapter.Adapter {
	t.Helper()
	adapter, err := githubadapter.New(githubadapter.Config{
		Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10",
		Mode: "user-token", Actor: "octocat", ActorKind: "user",
		RepositoryOwner: "octocat", RepositoryName: "example", RepositoryID: "R_repo",
		ProjectOwner: "octocat", ProjectOwnerKind: "user", ProjectID: "P_project",
		FieldIDs:            map[string]string{"readiness": "F_readiness", "status": "F_status"},
		OptionIDs:           map[string]string{"readiness:ready": "O_ready", "status:next": "O_next"},
		RequiredPermissions: []string{"issues:write", "projects:write", "pull_requests:read"},
	}, githubadapter.CredentialProviderFunc(func(context.Context) (githubadapter.Credential, error) {
		return githubadapter.Credential{Token: "top-secret-token", Mode: "user-token", Actor: "octocat", Permissions: []string{"issues:write", "projects:write", "pull_requests:read"}, ExpiresAt: now.Add(time.Hour)}, nil
	}), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func newPhaseUserAdapter(t *testing.T, server *httptest.Server, now time.Time) *githubadapter.Adapter {
	t.Helper()
	config := githubadapter.Config{
		Host: "github.com", RESTBaseURL: server.URL, GraphQLURL: server.URL + "/graphql", APIVersion: "2026-03-10",
		Mode: "user-token", Actor: "octocat", ActorKind: "user",
		RepositoryOwner: "octocat", RepositoryName: "example", RepositoryID: "R_repo",
		ProjectOwner: "octocat", ProjectOwnerKind: "user", ProjectID: "P_project",
		FieldIDs:            map[string]string{"readiness": "F_readiness", "status": "F_status", "phase": "F_phase"},
		OptionIDs:           map[string]string{"readiness:ready": "O_ready", "status:next": "O_next", "phase:Phase 0": "O_phase_0", "phase:Phase 1": "O_phase_1", "phase:Phase 2": "O_phase_2", "phase:Phase 3": "O_phase_3", "phase:Phase 4": "O_phase_4", "phase:Phase 5": "O_phase_5", "phase:Phase 6": "O_phase_6", "phase:Phase 7": "O_phase_7", "phase:Phase 8": "O_phase_8"},
		RequiredPermissions: []string{"issues:write", "projects:write", "pull_requests:read"},
	}
	adapter, err := githubadapter.New(config, credentialProvider(now, "user-token", "octocat", allPermissions()), server.Client(), githubadapter.WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value any) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatal(err)
	}
}
