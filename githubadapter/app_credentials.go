package githubadapter

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AppInstallationConfig identifies one approved GitHub App installation token route.
type AppInstallationConfig struct {
	RESTBaseURL    string
	APIVersion     string
	AppID          string
	InstallationID string
	Actor          string
	Account        string
}

// PrivateKeyProvider supplies App private-key bytes at request time without durable state.
type PrivateKeyProvider interface {
	PrivateKey(context.Context) ([]byte, error)
}

type PrivateKeyProviderFunc func(context.Context) ([]byte, error)

func (provider PrivateKeyProviderFunc) PrivateKey(ctx context.Context) ([]byte, error) {
	return provider(ctx)
}

type AppCredentialOption func(*AppInstallationProvider)

func WithAppCredentialClock(clock func() time.Time) AppCredentialOption {
	return func(provider *AppInstallationProvider) {
		if clock != nil {
			provider.now = clock
		}
	}
}

// AppInstallationProvider mints short-lived installation credentials from an injected key.
type AppInstallationProvider struct {
	config AppInstallationConfig
	keys   PrivateKeyProvider
	client *http.Client
	now    func() time.Time
}

func NewAppInstallationProvider(config AppInstallationConfig, keys PrivateKeyProvider, client *http.Client, options ...AppCredentialOption) (*AppInstallationProvider, error) {
	if config.RESTBaseURL == "" || config.APIVersion != "2026-03-10" || config.AppID == "" || config.InstallationID == "" || config.Actor == "" || config.Account == "" || keys == nil || client == nil {
		return nil, errors.New("GitHub App installation provider configuration is incomplete")
	}
	parsed, err := url.Parse(config.RESTBaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("GitHub App installation provider endpoint is invalid")
	}
	for _, numeric := range []string{config.AppID, config.InstallationID} {
		if value, err := strconv.ParseInt(numeric, 10, 64); err != nil || value <= 0 {
			return nil, errors.New("GitHub App and installation IDs must be positive integers")
		}
	}
	provider := &AppInstallationProvider{config: config, keys: keys, client: client, now: time.Now}
	for _, option := range options {
		option(provider)
	}
	return provider, nil
}

func (provider *AppInstallationProvider) Credential(ctx context.Context) (Credential, error) {
	keyBytes, err := provider.keys.PrivateKey(ctx)
	if err != nil {
		return Credential{}, errors.New("GitHub App private key is unavailable")
	}
	key, err := parseRSAPrivateKey(keyBytes)
	if err != nil {
		return Credential{}, errors.New("GitHub App private key is invalid")
	}
	now := provider.now()
	jwt, err := signAppJWT(key, provider.config.AppID, now)
	if err != nil {
		return Credential{}, errors.New("GitHub App identity token could not be signed")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(provider.config.RESTBaseURL, "/")+"/app/installations/"+provider.config.InstallationID+"/access_tokens", bytes.NewReader([]byte("{}")))
	if err != nil {
		return Credential{}, errors.New("GitHub App token request could not be prepared")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+jwt)
	request.Header.Set("X-GitHub-Api-Version", provider.config.APIVersion)
	request.Header.Set("User-Agent", "codex-starter-kit")
	request.Header.Set("Content-Type", "application/json")
	response, err := provider.client.Do(request)
	if err != nil {
		return Credential{}, errors.New("GitHub App token endpoint is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Credential{}, errors.New("GitHub App token mint was denied")
	}
	var minted struct {
		Token       string            `json:"token"`
		ExpiresAt   time.Time         `json:"expires_at"`
		Permissions map[string]string `json:"permissions"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&minted); err != nil || minted.Token == "" || !now.Before(minted.ExpiresAt) {
		return Credential{}, errors.New("GitHub App token response is invalid")
	}
	permissions := make([]string, 0, len(minted.Permissions))
	for name, access := range minted.Permissions {
		permissions = append(permissions, strings.ReplaceAll(name, "_", "-")+":"+access)
	}
	sort.Strings(permissions)
	return Credential{Token: minted.Token, IdentityToken: jwt, Mode: "app-installation", Actor: provider.config.Actor, Account: provider.config.Account, InstallationID: provider.config.InstallationID, Permissions: permissions, PermissionSource: "installation-token-response", PermissionRevision: sandboxDigest(minted.Permissions), ExpiresAt: minted.ExpiresAt}, nil
}

func parseRSAPrivateKey(content []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, errors.New("PEM block is absent")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return key, nil
}

func signAppJWT(key *rsa.PrivateKey, appID string, now time.Time) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{"iat": now.Add(-60 * time.Second).Unix(), "exp": now.Add(9 * time.Minute).Unix(), "iss": appID})
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}
