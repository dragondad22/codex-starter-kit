package githubadapter_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

func TestAppInstallationProviderMintsEphemeralBoundCredential(t *testing.T) {
	now := time.Date(2026, 7, 16, 20, 0, 0, 0, time.UTC)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") || len(strings.Split(strings.TrimPrefix(authorization, "Bearer "), ".")) != 3 {
			t.Fatalf("authorization is not an App JWT")
		}
		permissions := map[string]string{"issues": "write", "organization_projects": "write", "metadata": "read"}
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/app":
			json.NewEncoder(response).Encode(map[string]any{"id": 4319725, "slug": "codex-starter-kit-labs-reconciler", "owner": map[string]any{"login": "codex-starter-kit-labs", "id": 305967668}})
		case request.Method == http.MethodGet && request.URL.Path == "/app/installations/147093185":
			json.NewEncoder(response).Encode(map[string]any{"id": 147093185, "app_id": 4319725, "app_slug": "codex-starter-kit-labs-reconciler", "account": map[string]any{"login": "codex-starter-kit-labs", "id": 305967668}, "permissions": permissions})
		case request.Method == http.MethodPost && request.URL.Path == "/app/installations/147093185/access_tokens":
			json.NewEncoder(response).Encode(map[string]any{"token": "installation-secret", "expires_at": now.Add(time.Hour).Format(time.RFC3339), "permissions": permissions})
		default:
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	provider, err := githubadapter.NewAppInstallationProvider(githubadapter.AppInstallationConfig{
		RESTBaseURL: server.URL, APIVersion: "2026-03-10", AppID: "4319725", InstallationID: "147093185",
		Actor: "codex-starter-kit-labs-reconciler", Account: "codex-starter-kit-labs", AccountID: "305967668",
	}, githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) { return privateKey, nil }), server.Client(), githubadapter.WithAppCredentialClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	credential, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	if credential.Token != "installation-secret" || credential.IdentityToken == "" || credential.InstallationID != "147093185" || credential.AccountID != "305967668" || !strings.Contains(strings.Join(credential.Permissions, " "), "organization-projects:write") {
		t.Fatalf("credential = %#v", credential)
	}
	encoded, err := json.Marshal(credential)
	if err != nil || strings.Contains(string(encoded), "installation-secret") || strings.Contains(string(encoded), credential.IdentityToken) {
		t.Fatalf("credential JSON exposed secret: %s, %v", encoded, err)
	}
}

func TestAppInstallationProviderMintsExactRepositoryAndPermissionSubset(t *testing.T) {
	now := time.Date(2026, 7, 21, 20, 0, 0, 0, time.UTC)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	requested := map[string]string{"contents": "write", "metadata": "read", "pull_requests": "write"}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/app":
			json.NewEncoder(response).Encode(map[string]any{"id": 4319735, "slug": "codex-starter-kit-labs-seeder", "owner": map[string]any{"login": "codex-starter-kit-labs", "id": 305967668}})
		case request.Method == http.MethodGet && request.URL.Path == "/app/installations/147094309":
			json.NewEncoder(response).Encode(map[string]any{"id": 147094309, "app_id": 4319735, "app_slug": "codex-starter-kit-labs-seeder", "account": map[string]any{"login": "codex-starter-kit-labs", "id": 305967668}, "permissions": map[string]string{"contents": "write", "issues": "write", "metadata": "read", "pull_requests": "write", "workflows": "write"}})
		case request.Method == http.MethodPost && request.URL.Path == "/app/installations/147094309/access_tokens":
			var body struct {
				RepositoryIDs []int64           `json:"repository_ids"`
				Permissions   map[string]string `json:"permissions"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil || len(body.RepositoryIDs) != 1 || body.RepositoryIDs[0] != 1303189066 || len(body.Permissions) != len(requested) {
				t.Fatalf("mint request = %#v, err = %v", body, err)
			}
			for name, access := range requested {
				if body.Permissions[name] != access {
					t.Fatalf("mint permissions = %#v", body.Permissions)
				}
			}
			json.NewEncoder(response).Encode(map[string]any{"token": "down-scoped-secret", "expires_at": now.Add(time.Hour).Format(time.RFC3339), "permissions": requested, "repositories": []any{map[string]any{"id": 1303189066}}})
		default:
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	provider, err := githubadapter.NewAppInstallationProvider(githubadapter.AppInstallationConfig{
		RESTBaseURL: server.URL, APIVersion: "2026-03-10", AppID: "4319735", InstallationID: "147094309", Actor: "codex-starter-kit-labs-seeder", Account: "codex-starter-kit-labs", AccountID: "305967668",
		RepositoryIDs: []int64{1303189066}, TokenPermissions: requested,
	}, githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) { return privateKey, nil }), server.Client(), githubadapter.WithAppCredentialClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	credential, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(credential.Permissions, " ")
	if credential.Token != "down-scoped-secret" || !strings.Contains(joined, "contents:write") || !strings.Contains(joined, "pull-requests:write") || strings.Contains(joined, "issues:write") || strings.Contains(joined, "workflows:write") {
		t.Fatalf("credential = %#v", credential)
	}
}
