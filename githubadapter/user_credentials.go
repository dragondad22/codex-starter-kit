package githubadapter

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"
)

// UserTokenConfig binds a selected-repository reviewer token to its independently observed actor.
type UserTokenConfig struct {
	RESTBaseURL         string   `json:"rest_base_url"`
	APIVersion          string   `json:"api_version"`
	Actor               string   `json:"actor"`
	ActorID             string   `json:"actor_id"`
	RepositoryOwner     string   `json:"repository_owner"`
	RepositoryName      string   `json:"repository_name"`
	ApprovedPermissions []string `json:"approved_permissions"`
}

// NewReviewerTokenProvider validates actor and repository role on each short capability lease.
func NewReviewerTokenProvider(config UserTokenConfig, token string, client *http.Client, now func() time.Time) (CredentialProvider, error) {
	if config.RESTBaseURL == "" || config.APIVersion != "2026-03-10" || config.Actor == "" || config.ActorID == "" || config.RepositoryOwner == "" || config.RepositoryName == "" || len(config.ApprovedPermissions) == 0 || token == "" || client == nil {
		return nil, errors.New("reviewer token provider configuration is incomplete")
	}
	parsed, err := url.Parse(config.RESTBaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("reviewer token provider endpoint is invalid")
	}
	if now == nil {
		now = time.Now
	}
	return CredentialProviderFunc(func(ctx context.Context) (Credential, error) {
		transport := &Adapter{config: Config{RESTBaseURL: config.RESTBaseURL, APIVersion: config.APIVersion}, client: client}
		base := Credential{Token: token}
		var user struct {
			Login string `json:"login"`
			ID    int64  `json:"id"`
		}
		if _, err := transport.rest(ctx, base, http.MethodGet, "/user", nil, &user); err != nil || user.Login != config.Actor || strconv.FormatInt(user.ID, 10) != config.ActorID {
			return Credential{}, errors.New("reviewer token actor does not match the approved identity")
		}
		var repository struct {
			Permissions struct {
				Admin bool `json:"admin"`
				Push  bool `json:"push"`
				Pull  bool `json:"pull"`
			} `json:"permissions"`
		}
		path := "/repos/" + url.PathEscape(config.RepositoryOwner) + "/" + url.PathEscape(config.RepositoryName)
		if _, err := transport.rest(ctx, base, http.MethodGet, path, nil, &repository); err != nil || !repository.Permissions.Pull || !repository.Permissions.Push || repository.Permissions.Admin {
			return Credential{}, errors.New("reviewer token repository role does not match the approved non-admin Write role")
		}
		observedAt := now()
		return Credential{Token: token, Mode: "user-token", Actor: user.Login, Account: user.Login, AccountID: strconv.FormatInt(user.ID, 10), Permissions: slices.Clone(config.ApprovedPermissions), PermissionSource: "approved-fine-grained-token-manifest-plus-live-effect", PermissionRevision: sandboxDigest(config.ApprovedPermissions), ExpiresAt: observedAt.Add(5 * time.Minute)}, nil
	}), nil
}
