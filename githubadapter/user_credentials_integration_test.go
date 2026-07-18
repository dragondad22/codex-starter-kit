package githubadapter_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestReviewerTokenProviderObservesDistinctNonAdminIdentity(t *testing.T) {
	now := time.Date(2026, 7, 17, 2, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer reviewer-secret" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		switch request.URL.Path {
		case "/user":
			json.NewEncoder(response).Encode(map[string]any{"login": "american-dragon-designs", "id": 305973890})
		case "/repos/labs/sandbox":
			json.NewEncoder(response).Encode(map[string]any{"permissions": map[string]bool{"pull": true, "push": true, "admin": false}})
		default:
			t.Fatalf("path = %q", request.URL.Path)
		}
	}))
	defer server.Close()
	provider, err := githubadapter.NewReviewerTokenProvider(githubadapter.UserTokenConfig{RESTBaseURL: server.URL, APIVersion: "2026-03-10", Actor: "american-dragon-designs", ActorID: "305973890", RepositoryOwner: "labs", RepositoryName: "sandbox", ApprovedPermissions: []string{"contents:read", "pull-requests:write"}}, "reviewer-secret", server.Client(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}

	credential, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if credential.Actor != "american-dragon-designs" || credential.AccountID != "305973890" || !credential.ExpiresAt.Equal(now.Add(5*time.Minute)) {
		t.Fatalf("credential = %#v", credential)
	}
}
