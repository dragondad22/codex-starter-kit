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
		if request.URL.Path != "/app/installations/147093185/access_tokens" || request.Method != http.MethodPost {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		authorization := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") || len(strings.Split(strings.TrimPrefix(authorization, "Bearer "), ".")) != 3 {
			t.Fatalf("authorization is not an App JWT")
		}
		json.NewEncoder(response).Encode(map[string]any{
			"token": "installation-secret", "expires_at": now.Add(time.Hour).Format(time.RFC3339),
			"permissions": map[string]string{"issues": "write", "organization_projects": "write", "metadata": "read"},
		})
	}))
	defer server.Close()

	provider, err := githubadapter.NewAppInstallationProvider(githubadapter.AppInstallationConfig{
		RESTBaseURL: server.URL, APIVersion: "2026-03-10", AppID: "4319725", InstallationID: "147093185",
		Actor: "codex-starter-kit-labs-reconciler", Account: "codex-starter-kit-labs",
	}, githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) { return privateKey, nil }), server.Client(), githubadapter.WithAppCredentialClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	credential, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	if credential.Token != "installation-secret" || credential.IdentityToken == "" || credential.InstallationID != "147093185" || !strings.Contains(strings.Join(credential.Permissions, " "), "organization-projects:write") {
		t.Fatalf("credential = %#v", credential)
	}
	encoded, err := json.Marshal(credential)
	if err != nil || strings.Contains(string(encoded), "installation-secret") || strings.Contains(string(encoded), credential.IdentityToken) {
		t.Fatalf("credential JSON exposed secret: %s, %v", encoded, err)
	}
}
