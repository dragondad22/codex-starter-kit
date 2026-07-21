package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestQualificationRejectsInvalidInputsBeforeCredentials(t *testing.T) {
	for _, args := range [][]string{
		{"--revision", "main", "--path", "README.md", "--expected-digest", "sha256:" + strings.Repeat("0", 64)},
		{"--revision", strings.Repeat("a", 40), "--path", "../README.md", "--expected-digest", "sha256:" + strings.Repeat("0", 64)},
		{"--revision", strings.Repeat("a", 40), "--path", "README.md", "--expected-digest", "invalid"},
	} {
		if err := run(context.Background(), args, func(string) string { return "" }, http.DefaultClient, io.Discard); err == nil {
			t.Fatalf("invalid input was accepted: %v", args)
		}
	}
}

func TestQualificationEmitsRedactedDigestBoundReceipt(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	content := []byte("qualified\n")
	digest := "sha256:f1734a68232317c6dc71cbf33eb5858bf56b703bc1aafd29b7ba4cf893da3f70"
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "{}"
		switch {
		case request.URL.Path == "/app":
			body = `{"id":4319725,"slug":"codex-starter-kit-labs-reconciler","owner":{"login":"codex-starter-kit-labs","id":305967668}}`
		case request.URL.Path == "/app/installations/147093185":
			body = `{"id":147093185,"app_id":4319725,"app_slug":"codex-starter-kit-labs-reconciler","account":{"login":"codex-starter-kit-labs","id":305967668},"permissions":{"contents":"read"}}`
		case request.URL.Path == "/app/installations/147093185/access_tokens":
			body = `{"token":"installation-secret","expires_at":"2099-01-01T00:00:00Z","permissions":{"contents":"read"}}`
		case strings.Contains(request.URL.Path, "/contents/README.md"):
			if request.Header.Get("Authorization") != "Bearer installation-secret" {
				t.Fatal("content request lacked installation credential")
			}
			body = `{"type":"file","encoding":"base64","content":"` + base64.StdEncoding.EncodeToString(content) + `"}`
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	var output strings.Builder
	args := []string{"--revision", strings.Repeat("a", 40), "--path", "README.md", "--expected-digest", digest}
	if err := run(context.Background(), args, func(string) string { return privateKey }, client, &output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "installation-secret") || strings.Contains(output.String(), "PRIVATE KEY") {
		t.Fatal("receipt leaked credential material")
	}
	var result receipt
	if err := json.Unmarshal([]byte(output.String()), &result); err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "pass" || result.Digest != digest || !strings.Contains(strings.Join(result.Permissions, ","), "contents:read") {
		t.Fatalf("unexpected receipt: %#v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
