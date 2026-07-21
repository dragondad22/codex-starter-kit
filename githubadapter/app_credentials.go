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
	RESTBaseURL      string            `json:"rest_base_url"`
	APIVersion       string            `json:"api_version"`
	AppID            string            `json:"app_id"`
	InstallationID   string            `json:"installation_id"`
	Actor            string            `json:"actor"`
	Account          string            `json:"account"`
	AccountID        string            `json:"account_id"`
	RepositoryIDs    []int64           `json:"repository_ids,omitempty"`
	TokenPermissions map[string]string `json:"token_permissions,omitempty"`
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
	if config.RESTBaseURL == "" || config.APIVersion != "2026-03-10" || config.AppID == "" || config.InstallationID == "" || config.Actor == "" || config.Account == "" || config.AccountID == "" || keys == nil || client == nil {
		return nil, errors.New("GitHub App installation provider configuration is incomplete")
	}
	parsed, err := url.Parse(config.RESTBaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("GitHub App installation provider endpoint is invalid")
	}
	for _, numeric := range []string{config.AppID, config.InstallationID, config.AccountID} {
		if value, err := strconv.ParseInt(numeric, 10, 64); err != nil || value <= 0 {
			return nil, errors.New("GitHub App and installation IDs must be positive integers")
		}
	}
	if len(config.RepositoryIDs) > 500 {
		return nil, errors.New("GitHub App token repository scope exceeds 500 repositories")
	}
	config.RepositoryIDs = append([]int64(nil), config.RepositoryIDs...)
	sort.Slice(config.RepositoryIDs, func(left, right int) bool { return config.RepositoryIDs[left] < config.RepositoryIDs[right] })
	for index, id := range config.RepositoryIDs {
		if id <= 0 || index > 0 && id == config.RepositoryIDs[index-1] {
			return nil, errors.New("GitHub App token repository scope is invalid or duplicated")
		}
	}
	config.TokenPermissions = clonePermissionMap(config.TokenPermissions)
	for name, access := range config.TokenPermissions {
		if name == "" || strings.Contains(name, "-") || !strings.Contains(" read write ", " "+access+" ") {
			return nil, errors.New("GitHub App token permission scope is invalid")
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
	var app struct {
		ID    int64  `json:"id"`
		Slug  string `json:"slug"`
		Owner struct {
			Login string `json:"login"`
			ID    int64  `json:"id"`
		} `json:"owner"`
	}
	if err := provider.appJSON(ctx, jwt, http.MethodGet, "/app", nil, &app); err != nil || strconv.FormatInt(app.ID, 10) != provider.config.AppID || app.Slug != provider.config.Actor || app.Owner.Login != provider.config.Account || strconv.FormatInt(app.Owner.ID, 10) != provider.config.AccountID {
		return Credential{}, errors.New("GitHub App API identity does not match the approved configuration")
	}
	var installation struct {
		ID      int64  `json:"id"`
		AppID   int64  `json:"app_id"`
		AppSlug string `json:"app_slug"`
		Account struct {
			Login string `json:"login"`
			ID    int64  `json:"id"`
		} `json:"account"`
		Permissions map[string]string `json:"permissions"`
	}
	if err := provider.appJSON(ctx, jwt, http.MethodGet, "/app/installations/"+provider.config.InstallationID, nil, &installation); err != nil || strconv.FormatInt(installation.ID, 10) != provider.config.InstallationID || strconv.FormatInt(installation.AppID, 10) != provider.config.AppID || installation.AppSlug != app.Slug || installation.Account.Login != provider.config.Account || strconv.FormatInt(installation.Account.ID, 10) != provider.config.AccountID {
		return Credential{}, errors.New("GitHub App installation identity does not match the approved configuration")
	}
	for name, requested := range provider.config.TokenPermissions {
		if !permissionContains(installation.Permissions[name], requested) {
			return Credential{}, errors.New("GitHub App installation does not contain the requested token permission scope")
		}
	}
	var minted struct {
		Token        string            `json:"token"`
		ExpiresAt    time.Time         `json:"expires_at"`
		Permissions  map[string]string `json:"permissions"`
		Repositories []struct {
			ID int64 `json:"id"`
		} `json:"repositories"`
	}
	body, err := json.Marshal(struct {
		RepositoryIDs []int64           `json:"repository_ids,omitempty"`
		Permissions   map[string]string `json:"permissions,omitempty"`
	}{RepositoryIDs: provider.config.RepositoryIDs, Permissions: provider.config.TokenPermissions})
	if err != nil {
		return Credential{}, errors.New("GitHub App token scope could not be encoded")
	}
	if err := provider.appJSON(ctx, jwt, http.MethodPost, "/app/installations/"+provider.config.InstallationID+"/access_tokens", body, &minted); err != nil {
		return Credential{}, err
	}
	if minted.Token == "" || !now.Before(minted.ExpiresAt) {
		return Credential{}, errors.New("GitHub App token response is invalid")
	}
	expectedPermissions := installation.Permissions
	if len(provider.config.TokenPermissions) != 0 {
		expectedPermissions = provider.config.TokenPermissions
	}
	if sandboxDigest(minted.Permissions) != sandboxDigest(expectedPermissions) {
		return Credential{}, errors.New("GitHub App token permissions differ from the observed installation")
	}
	if len(provider.config.RepositoryIDs) != 0 {
		observedIDs := make([]int64, 0, len(minted.Repositories))
		for _, repository := range minted.Repositories {
			observedIDs = append(observedIDs, repository.ID)
		}
		sort.Slice(observedIDs, func(left, right int) bool { return observedIDs[left] < observedIDs[right] })
		if len(observedIDs) != len(provider.config.RepositoryIDs) {
			return Credential{}, errors.New("GitHub App token repository scope differs from the request")
		}
		for index := range observedIDs {
			if observedIDs[index] != provider.config.RepositoryIDs[index] || index > 0 && observedIDs[index] == observedIDs[index-1] {
				return Credential{}, errors.New("GitHub App token repository scope differs from the request")
			}
		}
	}
	permissions := make([]string, 0, len(minted.Permissions))
	for name, access := range minted.Permissions {
		permissions = append(permissions, strings.ReplaceAll(name, "_", "-")+":"+access)
	}
	sort.Strings(permissions)
	return Credential{Token: minted.Token, IdentityToken: jwt, Mode: "app-installation", Actor: app.Slug, Account: installation.Account.Login, AccountID: strconv.FormatInt(installation.Account.ID, 10), InstallationID: provider.config.InstallationID, Permissions: permissions, PermissionSource: "installation-token-response", PermissionRevision: sandboxDigest(minted.Permissions), ExpiresAt: minted.ExpiresAt}, nil
}

func clonePermissionMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	result := make(map[string]string, len(input))
	for name, access := range input {
		result[name] = access
	}
	return result
}

func permissionContains(actual, requested string) bool {
	level := map[string]int{"read": 1, "write": 2}
	return level[actual] >= level[requested] && level[requested] != 0
}

func (provider *AppInstallationProvider) appJSON(ctx context.Context, jwt, method, path string, body []byte, output any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(provider.config.RESTBaseURL, "/")+path, reader)
	if err != nil {
		return errors.New("GitHub App request could not be prepared")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+jwt)
	request.Header.Set("X-GitHub-Api-Version", provider.config.APIVersion)
	request.Header.Set("User-Agent", "codex-starter-kit")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := provider.client.Do(request)
	if err != nil {
		return errors.New("GitHub App endpoint is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errors.New("GitHub App request was denied")
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(output); err != nil {
		return errors.New("GitHub App response is invalid")
	}
	return nil
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
