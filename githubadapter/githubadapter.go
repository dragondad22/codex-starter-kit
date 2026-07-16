// Package githubadapter implements the lifecycle engine WorkAdapter seam with native HTTP.
package githubadapter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

var errGraphQLPartial = errors.New("GitHub GraphQL observation returned partial errors")

// Config is the credential-free, allowlisted GitHub target manifest.
type Config struct {
	Host                string
	RESTBaseURL         string
	GraphQLURL          string
	APIVersion          string
	Mode                string
	Actor               string
	ActorKind           string
	Account             string
	InstallationID      string
	RepositoryOwner     string
	RepositoryName      string
	RepositoryID        string
	ProjectOwner        string
	ProjectOwnerKind    string
	ProjectID           string
	FieldIDs            map[string]string
	OptionIDs           map[string]string
	RequiredPermissions []string
	MaxPages            int
	EvidenceMode        string
	LiveTargetApproved  bool
}

// Credential is an ephemeral authority value supplied at request time. Token is never returned by the adapter.
type Credential struct {
	Token          string    `json:"-"`
	Mode           string    `json:"mode"`
	Actor          string    `json:"actor"`
	Account        string    `json:"account,omitempty"`
	InstallationID string    `json:"installation_id,omitempty"`
	Permissions    []string  `json:"permissions"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// CredentialProvider supplies one ephemeral credential without placing it in desired state.
type CredentialProvider interface {
	Credential(context.Context) (Credential, error)
}

// CredentialProviderFunc adapts a function to CredentialProvider.
type CredentialProviderFunc func(context.Context) (Credential, error)

func (provider CredentialProviderFunc) Credential(ctx context.Context) (Credential, error) {
	return provider(ctx)
}

// Option configures internal deterministic dependencies.
type Option func(*Adapter)

// WithClock supplies the adapter observation clock.
func WithClock(clock func() time.Time) Option {
	return func(adapter *Adapter) {
		if clock != nil {
			adapter.now = clock
		}
	}
}

// Adapter implements engine.WorkAdapter with native REST and GraphQL transport.
type Adapter struct {
	config   Config
	provider CredentialProvider
	client   *http.Client
	now      func() time.Time
}

// New validates a fixed target and returns a GitHub adapter.
func New(config Config, provider CredentialProvider, client *http.Client, options ...Option) (*Adapter, error) {
	if config.Host == "" || config.RESTBaseURL == "" || config.GraphQLURL == "" || config.APIVersion == "" || config.Mode == "" || config.Actor == "" || config.ActorKind == "" || config.RepositoryOwner == "" || config.RepositoryName == "" || config.RepositoryID == "" || config.ProjectOwner == "" || config.ProjectOwnerKind == "" || config.ProjectID == "" || len(config.RequiredPermissions) == 0 || config.FieldIDs["readiness"] == "" || config.FieldIDs["status"] == "" || len(config.OptionIDs) == 0 {
		return nil, errors.New("GitHub adapter configuration lacks required target, identity, or permission facts")
	}
	if !slices.Contains([]string{"app-installation", "user-token", "actions-job"}, config.Mode) {
		return nil, errors.New("GitHub adapter credential mode is unsupported")
	}
	if config.APIVersion != "2026-03-10" {
		return nil, errors.New("GitHub adapter REST API version is unsupported")
	}
	if config.Mode == "app-installation" && (config.InstallationID == "" || config.Account == "" || config.ProjectOwnerKind != "organization") {
		return nil, errors.New("GitHub App installation mode requires an installation and organization-owned Project")
	}
	if config.EvidenceMode == "" {
		config.EvidenceMode = "simulated"
	}
	if !slices.Contains([]string{"simulated", "live"}, config.EvidenceMode) {
		return nil, errors.New("GitHub adapter evidence mode is unsupported")
	}
	if config.EvidenceMode == "live" && !config.LiveTargetApproved {
		return nil, errors.New("live GitHub adapter requires an approved target manifest")
	}
	if provider == nil || client == nil {
		return nil, errors.New("GitHub adapter requires credential provider and HTTP client")
	}
	for _, raw := range []string{config.RESTBaseURL, config.GraphQLURL} {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, errors.New("GitHub adapter endpoint is invalid")
		}
		if config.EvidenceMode == "live" && (config.Host != "github.com" || parsed.Scheme != "https" || parsed.Host != "api.github.com") {
			return nil, errors.New("live GitHub.com adapter endpoints must use the approved HTTPS API host")
		}
	}
	config.RequiredPermissions = slices.Clone(config.RequiredPermissions)
	config.FieldIDs = cloneMap(config.FieldIDs)
	config.OptionIDs = cloneMap(config.OptionIDs)
	if config.MaxPages == 0 {
		config.MaxPages = 10
	}
	if config.MaxPages < 1 || config.MaxPages > 100 {
		return nil, errors.New("GitHub adapter pagination bound must be between 1 and 100")
	}
	adapter := &Adapter{config: config, provider: provider, client: client, now: func() time.Time { return time.Now().UTC() }}
	for _, option := range options {
		option(adapter)
	}
	return adapter, nil
}

// Observe reads one stable-marker issue and its Project item through bounded pagination.
func (adapter *Adapter) Observe(ctx context.Context, target engine.WorkTarget, managedID string) (engine.WorkObservation, error) {
	credential, err := adapter.credential(ctx)
	if err != nil {
		return engine.WorkObservation{}, err
	}
	if managedID == "" || target.Host != adapter.config.Host || target.RepositoryID != adapter.config.RepositoryID || target.ProjectID != adapter.config.ProjectID || !equalMap(target.FieldIDs, adapter.config.FieldIDs) || !equalMap(target.OptionIDs, adapter.config.OptionIDs) {
		return engine.WorkObservation{}, errors.New("GitHub observation target is outside the allowlisted manifest")
	}
	issues, err := adapter.findManagedIssues(ctx, credential, managedID)
	if err != nil {
		return engine.WorkObservation{}, err
	}
	observation := engine.WorkObservation{
		SchemaVersion: 1, ConfigurationRevision: adapter.configurationRevision(credential.Permissions),
		Target: cloneTarget(target), Disposition: "observed", Problems: []string{},
	}
	if len(issues) > 1 {
		observation.Disposition = "ambiguous"
		observation.Problems = []string{"multiple issues contain the stable managed marker"}
		observation.Revision = digest(struct {
			ManagedID string
			Matches   []githubIssue
		}{managedID, issues})
		return observation, nil
	}
	if len(issues) == 0 {
		observation.Revision = digest(struct{ ManagedID string }{managedID})
		return observation, nil
	}

	issue := issues[0]
	projectItem, err := adapter.findProjectItem(ctx, credential, issue.NodeID, target)
	if err != nil {
		if errors.Is(err, errGraphQLPartial) {
			observation.Disposition = "needs-review"
			observation.Problems = []string{errGraphQLPartial.Error()}
			observation.Revision = digest(struct {
				ManagedID string
				Problem   string
			}{managedID, errGraphQLPartial.Error()})
			return observation, nil
		}
		return engine.WorkObservation{}, err
	}
	issueType := "task"
	for _, label := range issue.Labels {
		if strings.HasPrefix(label.Name, "type:") {
			issueType = strings.TrimPrefix(label.Name, "type:")
			break
		}
	}
	observed := &engine.WorkObservedTask{
		ManagedID: managedID, IssueNodeID: issue.NodeID, Title: issue.Title, IssueType: issueType,
		Closed: strings.EqualFold(issue.State, "closed"),
	}
	if projectItem != nil {
		observed.ProjectItemID = projectItem.ID
		for _, field := range projectItem.FieldValues.Nodes {
			switch field.Field.ID {
			case target.FieldIDs["readiness"]:
				observed.ReadinessOption = field.OptionID
			case target.FieldIDs["status"]:
				observed.StatusOption = field.OptionID
			case target.FieldIDs["phase"]:
				observed.Phase = field.OptionID
			}
		}
	}
	if metadata, ok := parseManagedMetadata(issue.Body); ok && metadata.ManagedID == managedID {
		observed.IssueType = metadata.IssueType
		observed.ParentManagedID = metadata.ParentManagedID
		observed.BlockedBy = make([]string, 0, len(metadata.Blockers))
		for _, blocker := range metadata.Blockers {
			observed.BlockedBy = append(observed.BlockedBy, blocker.ManagedID)
		}
		observed.Phase = metadata.Phase
		observed.PromotionRecord = metadata.PromotionRecord
		observed.Review = slices.Clone(metadata.Review)
	}
	observation.Task = observed
	observation.Revision = digest(observed)
	return observation, nil
}

// Apply performs one allowlisted semantic effect with stable-marker recovery.
func (adapter *Adapter) Apply(ctx context.Context, effect engine.WorkEffect) (engine.WorkEffectResult, error) {
	credential, err := adapter.credential(ctx)
	if err != nil {
		return engine.WorkEffectResult{Outcome: "denied", Attempt: 1, Detail: "selected GitHub authority is unavailable"}, err
	}
	if effect.ManagedID == "" || effect.Marker != "starter-kit-managed:"+effect.ManagedID || effect.Desired.ManagedID != effect.ManagedID {
		return engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "effect identity is outside the managed marker contract"}, errors.New("invalid GitHub managed-task effect")
	}
	issues, err := adapter.findManagedIssues(ctx, credential, effect.ManagedID)
	if err != nil {
		return adapter.transportResult(err, nil)
	}
	if len(issues) > 1 {
		return engine.WorkEffectResult{Outcome: "ambiguous", Attempt: 1, Detail: "multiple issues contain the stable managed marker", Recoverable: true}, nil
	}
	if effect.Kind == "create-task" {
		if len(issues) == 1 {
			return engine.WorkEffectResult{Outcome: "applied", Attempt: 1, Detail: "recovered the existing stable-marker issue after create"}, nil
		}
		var created githubIssue
		response, createErr := adapter.rest(ctx, credential, http.MethodPost, adapter.issuePath(), map[string]any{
			"title":  effect.Desired.Title,
			"body":   managedBody(effect.Desired),
			"labels": []string{"type:" + effect.Desired.IssueType},
		}, &created)
		if createErr != nil {
			return adapter.transportResult(createErr, response)
		}
		if created.NodeID == "" {
			return engine.WorkEffectResult{Outcome: "ambiguous", Attempt: 1, Detail: "issue create response lacked immutable identity", Recoverable: true}, nil
		}
		return engine.WorkEffectResult{Outcome: "applied", Attempt: 1, Detail: "created the stable-marker issue"}, nil
	}
	if effect.Kind != "reconcile-task" {
		return engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "unsupported GitHub semantic effect"}, errors.New("unsupported GitHub effect")
	}
	if len(issues) == 0 {
		return engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "stable-marker issue is absent", Recoverable: true}, nil
	}
	issue := issues[0]
	state := "open"
	if effect.Desired.Closed {
		state = "closed"
	}
	response, updateErr := adapter.rest(ctx, credential, http.MethodPatch, adapter.issuePath()+"/"+strconv.Itoa(issue.Number), map[string]any{
		"title": effect.Desired.Title, "body": managedBody(effect.Desired), "state": state, "labels": []string{"type:" + effect.Desired.IssueType},
	}, &struct{}{})
	if updateErr != nil {
		return adapter.transportResult(updateErr, response)
	}
	target := engine.WorkTarget{Host: adapter.config.Host, RepositoryID: adapter.config.RepositoryID, ProjectID: adapter.config.ProjectID, FieldIDs: cloneMap(adapter.config.FieldIDs), OptionIDs: cloneMap(adapter.config.OptionIDs)}
	// The immutable field and option IDs are fixed by the adapter target manifest, not chosen by transport.
	item, itemErr := adapter.findProjectItem(ctx, credential, issue.NodeID, target)
	if itemErr != nil {
		return adapter.transportResult(itemErr, nil)
	}
	itemID := ""
	if item != nil {
		itemID = item.ID
	} else {
		var added struct {
			Data struct {
				Add struct {
					Item struct {
						ID string `json:"id"`
					} `json:"item"`
				} `json:"addProjectV2ItemById"`
			} `json:"data"`
			Errors []graphQLError `json:"errors"`
		}
		query := `mutation AddManagedTaskToProject($project: ID!, $content: ID!) { addProjectV2ItemById(input: {projectId: $project, contentId: $content}) { item { id } } }`
		if graphErr := adapter.graphql(ctx, credential, query, map[string]any{"project": adapter.config.ProjectID, "content": issue.NodeID}, &added); graphErr != nil {
			return adapter.transportResult(graphErr, nil)
		}
		if len(added.Errors) != 0 || added.Data.Add.Item.ID == "" {
			return engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "GitHub Project add returned partial or missing data", Recoverable: true}, nil
		}
		itemID = added.Data.Add.Item.ID
	}
	for _, field := range []struct {
		Field  string
		Option string
	}{
		{adapter.configFieldID("readiness"), adapter.configOptionID("readiness", effect.Desired.Readiness)},
		{adapter.configFieldID("status"), adapter.configOptionID("status", effect.Desired.Status)},
	} {
		if field.Field == "" || field.Option == "" {
			return engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "immutable lifecycle field or option identity is unavailable", Recoverable: true}, nil
		}
		var updated struct {
			Data struct {
				Update struct {
					Item struct {
						ID string `json:"id"`
					} `json:"projectV2Item"`
				} `json:"updateProjectV2ItemFieldValue"`
			} `json:"data"`
			Errors []graphQLError `json:"errors"`
		}
		query := `mutation UpdateManagedTaskField($project: ID!, $item: ID!, $field: ID!, $option: String!) { updateProjectV2ItemFieldValue(input: {projectId: $project, itemId: $item, fieldId: $field, value: {singleSelectOptionId: $option}}) { projectV2Item { id } } }`
		if graphErr := adapter.graphql(ctx, credential, query, map[string]any{"project": adapter.config.ProjectID, "item": itemID, "field": field.Field, "option": field.Option}, &updated); graphErr != nil {
			return adapter.transportResult(graphErr, nil)
		}
		if len(updated.Errors) != 0 || updated.Data.Update.Item.ID == "" {
			return engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "GitHub Project field update returned partial or missing data", Recoverable: true}, nil
		}
	}
	return engine.WorkEffectResult{Outcome: "applied", Attempt: 1, Detail: "reconciled the stable-marker issue and Project item"}, nil
}

type githubIssue struct {
	Number int    `json:"number"`
	NodeID string `json:"node_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func (adapter *Adapter) findManagedIssues(ctx context.Context, credential Credential, managedID string) ([]githubIssue, error) {
	marker := "<!-- starter-kit-managed:" + managedID + " -->"
	path := "/repos/" + url.PathEscape(adapter.config.RepositoryOwner) + "/" + url.PathEscape(adapter.config.RepositoryName) + "/issues?state=all&per_page=100"
	matches := []githubIssue{}
	for page := 0; page < adapter.config.MaxPages && path != ""; page++ {
		var issues []githubIssue
		response, err := adapter.rest(ctx, credential, http.MethodGet, path, nil, &issues)
		if err != nil {
			return nil, err
		}
		for _, issue := range issues {
			if strings.Contains(issue.Body, marker) {
				matches = append(matches, issue)
			}
		}
		path, err = adapter.nextRESTPath(response.Header.Get("Link"))
		if err != nil {
			return nil, err
		}
	}
	if path != "" {
		return nil, errors.New("GitHub REST pagination exceeded the configured bound")
	}
	return matches, nil
}

type projectItem struct {
	ID      string `json:"id"`
	Content struct {
		ID string `json:"id"`
	} `json:"content"`
	FieldValues struct {
		Nodes []struct {
			OptionID string `json:"optionId"`
			Field    struct {
				ID string `json:"id"`
			} `json:"field"`
		} `json:"nodes"`
	} `json:"fieldValues"`
}

func (adapter *Adapter) findProjectItem(ctx context.Context, credential Credential, issueNodeID string, target engine.WorkTarget) (*projectItem, error) {
	cursor := any(nil)
	for page := 0; page < adapter.config.MaxPages; page++ {
		var response struct {
			Data struct {
				Node struct {
					Items struct {
						Nodes    []projectItem `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"items"`
				} `json:"node"`
			} `json:"data"`
			Errors []graphQLError `json:"errors"`
		}
		query := `query ManagedTaskObservation($project: ID!, $after: String) { node(id: $project) { ... on ProjectV2 { items(first: 100, after: $after) { nodes { id content { ... on Issue { id } } fieldValues(first: 100) { nodes { ... on ProjectV2ItemFieldSingleSelectValue { optionId field { ... on ProjectV2FieldCommon { id } } } } } } pageInfo { hasNextPage endCursor } } } } }`
		if err := adapter.graphql(ctx, credential, query, map[string]any{"project": target.ProjectID, "after": cursor}, &response); err != nil {
			return nil, err
		}
		if len(response.Errors) != 0 {
			return nil, errGraphQLPartial
		}
		for _, item := range response.Data.Node.Items.Nodes {
			if item.Content.ID == issueNodeID {
				copy := item
				return &copy, nil
			}
		}
		if !response.Data.Node.Items.PageInfo.HasNextPage {
			return nil, nil
		}
		cursor = response.Data.Node.Items.PageInfo.EndCursor
	}
	return nil, errors.New("GitHub GraphQL pagination exceeded the configured bound")
}

func (adapter *Adapter) credential(ctx context.Context) (Credential, error) {
	credential, err := adapter.provider.Credential(ctx)
	if err != nil {
		return Credential{}, fmt.Errorf("acquire GitHub credential: %w", err)
	}
	if credential.Token == "" || credential.Mode != adapter.config.Mode || credential.Actor != adapter.config.Actor || !containsAll(credential.Permissions, adapter.config.RequiredPermissions) {
		return Credential{}, errors.New("GitHub credential does not match the selected minimum authority")
	}
	if credential.Mode == "app-installation" && (credential.InstallationID != adapter.config.InstallationID || credential.Account != adapter.config.Account) {
		return Credential{}, errors.New("GitHub App credential does not match the selected installation account")
	}
	return credential, nil
}

func (adapter *Adapter) configurationRevision(permissions []string) string {
	return digest(struct {
		Config      Config
		Permissions []string
	}{adapter.config, permissions})
}

func (adapter *Adapter) nextRESTPath(link string) (string, error) {
	if link == "" {
		return "", nil
	}
	for _, part := range strings.Split(link, ",") {
		if !strings.Contains(part, `rel="next"`) {
			continue
		}
		left := strings.Index(part, "<")
		right := strings.Index(part, ">")
		if left < 0 || right <= left {
			return "", errors.New("GitHub REST pagination link is invalid")
		}
		next, err := url.Parse(part[left+1 : right])
		base, baseErr := url.Parse(adapter.config.RESTBaseURL)
		if err != nil || baseErr != nil || next.Scheme != base.Scheme || next.Host != base.Host {
			return "", errors.New("GitHub REST pagination escaped the configured host")
		}
		return next.RequestURI(), nil
	}
	return "", nil
}

func cloneTarget(target engine.WorkTarget) engine.WorkTarget {
	target.FieldIDs = cloneMap(target.FieldIDs)
	target.OptionIDs = cloneMap(target.OptionIDs)
	return target
}

func cloneMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func equalMap(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func (adapter *Adapter) configFieldID(name string) string { return adapter.config.FieldIDs[name] }

func (adapter *Adapter) configOptionID(field, value string) string {
	return adapter.config.OptionIDs[field+":"+value]
}

func (adapter *Adapter) issuePath() string {
	return "/repos/" + url.PathEscape(adapter.config.RepositoryOwner) + "/" + url.PathEscape(adapter.config.RepositoryName) + "/issues"
}

func managedBody(desired engine.DesiredManagedTask) string {
	encoded, _ := json.Marshal(desired)
	metadata := base64.RawURLEncoding.EncodeToString(encoded)
	return "<!-- starter-kit-managed:" + desired.ManagedID + " -->\n<!-- starter-kit-managed-metadata:" + metadata + " -->"
}

func parseManagedMetadata(body string) (engine.DesiredManagedTask, bool) {
	const prefix = "<!-- starter-kit-managed-metadata:"
	start := strings.Index(body, prefix)
	if start < 0 {
		return engine.DesiredManagedTask{}, false
	}
	start += len(prefix)
	end := strings.Index(body[start:], " -->")
	if end < 0 {
		return engine.DesiredManagedTask{}, false
	}
	encoded := body[start : start+end]
	content, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(content) > 64<<10 {
		return engine.DesiredManagedTask{}, false
	}
	var desired engine.DesiredManagedTask
	if err := json.Unmarshal(content, &desired); err != nil {
		return engine.DesiredManagedTask{}, false
	}
	return desired, true
}

func (adapter *Adapter) transportResult(err error, response *http.Response) (engine.WorkEffectResult, error) {
	result := engine.WorkEffectResult{Outcome: "failed", Attempt: 1, Detail: "GitHub transport effect failed", Recoverable: true}
	var requestFailure *responseError
	if response == nil && errors.As(err, &requestFailure) {
		response = &http.Response{StatusCode: requestFailure.StatusCode, Header: requestFailure.Header}
	}
	if response == nil {
		result.Outcome = "offline"
		result.Detail = "GitHub transport is offline"
		return result, err
	}
	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		if response.Header.Get("Retry-After") == "" && response.Header.Get("X-RateLimit-Remaining") != "0" {
			result.Outcome = "denied"
			result.Detail = "GitHub authority was denied or the selected resource is hidden"
			return result, err
		}
	}
	if response.StatusCode == http.StatusTooManyRequests || response.Header.Get("Retry-After") != "" || response.Header.Get("X-RateLimit-Remaining") == "0" {
		now := adapter.now().UTC()
		retrySeconds, _ := strconv.Atoi(response.Header.Get("Retry-After"))
		if retrySeconds <= 0 {
			retrySeconds = 60
		}
		retryAt := now.Add(time.Duration(retrySeconds) * time.Second)
		resetUnix, _ := strconv.ParseInt(response.Header.Get("X-RateLimit-Reset"), 10, 64)
		resetAt := time.Unix(resetUnix, 0).UTC()
		if resetUnix == 0 || resetAt.Before(retryAt) {
			resetAt = retryAt
		}
		result.Outcome = "rate-limited"
		result.Detail = "GitHub rate limit requires bounded delayed retry"
		result.Retry = &engine.WorkRetryState{Attempt: 1, MaxAttempts: 3, RetryAt: retryAt, ResetAt: resetAt}
	}
	return result, err
}

// Capability performs a non-mutating identity, target, permission, version, and budget handshake.
func (adapter *Adapter) Capability(ctx context.Context) (engine.WorkCapability, error) {
	credential, providerErr := adapter.provider.Credential(ctx)
	now := adapter.now().UTC()
	capability := adapter.baseCapability(credential, now)
	if providerErr != nil {
		capability.Online = false
		capability.Fresh = false
		capability.Disposition = "not-configured"
		capability.Problems = []string{"selected GitHub credential is unavailable"}
		return capability, nil
	}
	if credential.Token == "" {
		capability.Online = false
		capability.Fresh = false
		capability.Disposition = "not-configured"
		capability.Problems = []string{"selected GitHub credential is empty"}
		return capability, nil
	}
	if credential.Mode != adapter.config.Mode || credential.Actor != adapter.config.Actor {
		capability.Disposition = "needs-review"
		capability.Problems = []string{"GitHub credential mode or actor does not match the selected identity"}
		return capability, nil
	}
	if !containsAll(credential.Permissions, adapter.config.RequiredPermissions) {
		capability.Disposition = "denied"
		capability.Problems = []string{"GitHub credential lacks the declared minimum permissions"}
		return capability, nil
	}
	if credential.Mode == "app-installation" && (credential.InstallationID != adapter.config.InstallationID || credential.Account != adapter.config.Account) {
		capability.Disposition = "needs-review"
		capability.Problems = []string{"GitHub App credential does not match the selected installation account"}
		return capability, nil
	}

	var actor struct {
		Login string `json:"login"`
		Slug  string `json:"slug"`
		Type  string `json:"type"`
	}
	actorPath := "/user"
	if credential.Mode == "app-installation" {
		actorPath = "/app"
	}
	actorResponse, err := adapter.rest(ctx, credential, http.MethodGet, actorPath, nil, &actor)
	if err != nil {
		return adapter.failedCapability(capability, err), nil
	}
	observedActor := actor.Login
	observedKind := actor.Type
	if credential.Mode == "app-installation" {
		observedActor = actor.Slug
		observedKind = "app"
	}
	if observedActor != adapter.config.Actor || !strings.EqualFold(observedKind, adapter.config.ActorKind) {
		capability.Disposition = "needs-review"
		capability.Problems = []string{"GitHub API actor does not match the selected identity"}
		return capability, nil
	}

	var repository struct {
		NodeID string `json:"node_id"`
		Owner  struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	_, err = adapter.rest(ctx, credential, http.MethodGet, "/repos/"+url.PathEscape(adapter.config.RepositoryOwner)+"/"+url.PathEscape(adapter.config.RepositoryName), nil, &repository)
	if err != nil {
		return adapter.failedCapability(capability, err), nil
	}
	if repository.NodeID != adapter.config.RepositoryID || repository.Owner.Login != adapter.config.RepositoryOwner {
		capability.Disposition = "needs-review"
		capability.Problems = []string{"GitHub repository identity does not match the allowlisted target"}
		return capability, nil
	}

	var projectResponse struct {
		Data struct {
			Node struct {
				ID    string `json:"id"`
				Owner struct {
					Login    string `json:"login"`
					TypeName string `json:"__typename"`
				} `json:"owner"`
			} `json:"node"`
			RateLimit struct {
				Limit     int       `json:"limit"`
				Remaining int       `json:"remaining"`
				ResetAt   time.Time `json:"resetAt"`
			} `json:"rateLimit"`
		} `json:"data"`
		Errors []graphQLError `json:"errors"`
	}
	if err := adapter.graphql(ctx, credential, `query ManagedTaskProject($id: ID!) { node(id: $id) { ... on ProjectV2 { id owner { __typename ... on User { login } ... on Organization { login } } } } rateLimit { limit remaining resetAt } }`, map[string]any{"id": adapter.config.ProjectID}, &projectResponse); err != nil {
		return adapter.failedCapability(capability, err), nil
	}
	if len(projectResponse.Errors) != 0 || projectResponse.Data.Node.ID != adapter.config.ProjectID || projectResponse.Data.Node.Owner.Login != adapter.config.ProjectOwner || !strings.EqualFold(projectResponse.Data.Node.Owner.TypeName, adapter.config.ProjectOwnerKind) {
		capability.Disposition = "needs-review"
		capability.Problems = []string{"GitHub Project identity is unavailable or does not match the allowlisted target"}
		return capability, nil
	}

	capability.RepositoryID = repository.NodeID
	capability.RepositoryOwner = repository.Owner.Login
	capability.ProjectID = projectResponse.Data.Node.ID
	capability.ProjectOwner = projectResponse.Data.Node.Owner.Login
	capability.ProjectOwnerKind = adapter.config.ProjectOwnerKind
	capability.RESTRate = rateBudget(actorResponse.Header, "rest")
	capability.GraphQLRate = &engine.WorkRateBudget{Resource: "graphql", Limit: projectResponse.Data.RateLimit.Limit, Remaining: projectResponse.Data.RateLimit.Remaining, ResetAt: projectResponse.Data.RateLimit.ResetAt}
	if credential.Mode == "actions-job" {
		capability.Disposition = "unsupported"
		capability.Problems = []string{"Actions GITHUB_TOKEN is repository-local and cannot provide Project mutation authority"}
		capability.Limitations = slices.Clone(capability.Problems)
	}
	capability.ConfigurationRevision = adapter.configurationRevision(capability.Permissions)
	return capability, nil
}

func (adapter *Adapter) baseCapability(credential Credential, now time.Time) engine.WorkCapability {
	mode := credential.Mode
	if mode == "" {
		mode = adapter.config.Mode
	}
	actor := credential.Actor
	if actor == "" {
		actor = adapter.config.Actor
	}
	expiresAt := credential.ExpiresAt.UTC()
	if expiresAt.IsZero() {
		expiresAt = now
	}
	capability := engine.WorkCapability{
		SchemaVersion: 1, Online: true, Fresh: expiresAt.After(now), Mode: mode, Actor: actor,
		ActorKind: adapter.config.ActorKind, Account: adapter.config.Account, InstallationID: adapter.config.InstallationID,
		Host: adapter.config.Host, APIVersion: adapter.config.APIVersion, EvidenceMode: adapter.config.EvidenceMode,
		Disposition: "available", Problems: []string{}, Permissions: slices.Clone(credential.Permissions), RequiredPermissions: slices.Clone(adapter.config.RequiredPermissions),
		RepositoryID: adapter.config.RepositoryID, RepositoryOwner: adapter.config.RepositoryOwner,
		ProjectID: adapter.config.ProjectID, ProjectOwner: adapter.config.ProjectOwner, ProjectOwnerKind: adapter.config.ProjectOwnerKind,
		ObservedAt: now, ExpiresAt: expiresAt,
	}
	capability.ConfigurationRevision = adapter.configurationRevision(capability.Permissions)
	return capability
}

func (adapter *Adapter) failedCapability(capability engine.WorkCapability, err error) engine.WorkCapability {
	var failure *responseError
	capability.Disposition = "needs-review"
	capability.Problems = []string{"GitHub capability handshake failed"}
	if !errors.As(err, &failure) {
		capability.Online = false
		capability.Fresh = false
		capability.Disposition = "offline"
		capability.Problems = []string{"GitHub capability transport is offline"}
		return capability
	}
	if failure.StatusCode == http.StatusUnauthorized || failure.StatusCode == http.StatusForbidden || failure.StatusCode == http.StatusNotFound {
		capability.Disposition = "denied"
		capability.Problems = []string{"GitHub capability authority was denied or the selected resource is hidden"}
	}
	return capability
}

func (adapter *Adapter) rest(ctx context.Context, credential Credential, method, path string, body any, output any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(adapter.config.RESTBaseURL, "/")+path, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+credential.Token)
	request.Header.Set("X-GitHub-Api-Version", adapter.config.APIVersion)
	request.Header.Set("User-Agent", "codex-starter-kit")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := adapter.client.Do(request)
	if err != nil {
		return nil, errors.New("GitHub REST transport is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response, &responseError{StatusCode: response.StatusCode, Header: response.Header.Clone()}
	}
	if output != nil {
		decoder := json.NewDecoder(io.LimitReader(response.Body, 4<<20))
		if err := decoder.Decode(output); err != nil {
			return response, errors.New("decode GitHub REST response")
		}
	}
	return response, nil
}

type responseError struct {
	StatusCode int
	Header     http.Header
}

func (failure *responseError) Error() string {
	return fmt.Sprintf("GitHub request returned status %d", failure.StatusCode)
}

type graphQLError struct {
	Message string `json:"message"`
}

func (adapter *Adapter) graphql(ctx context.Context, credential Credential, query string, variables map[string]any, output any) error {
	requestBody, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, adapter.config.GraphQLURL, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+credential.Token)
	request.Header.Set("X-GitHub-Api-Version", adapter.config.APIVersion)
	request.Header.Set("User-Agent", "codex-starter-kit")
	request.Header.Set("Content-Type", "application/json")
	response, err := adapter.client.Do(request)
	if err != nil {
		return errors.New("GitHub GraphQL transport is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &responseError{StatusCode: response.StatusCode, Header: response.Header.Clone()}
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(output); err != nil {
		return errors.New("decode GitHub GraphQL response")
	}
	return nil
}

func containsAll(values, required []string) bool {
	for _, value := range required {
		if !slices.Contains(values, value) {
			return false
		}
	}
	return true
}

func rateBudget(header http.Header, resource string) *engine.WorkRateBudget {
	limit, _ := strconv.Atoi(header.Get("X-RateLimit-Limit"))
	remaining, _ := strconv.Atoi(header.Get("X-RateLimit-Remaining"))
	reset, _ := strconv.ParseInt(header.Get("X-RateLimit-Reset"), 10, 64)
	return &engine.WorkRateBudget{Resource: resource, Limit: limit, Remaining: remaining, ResetAt: time.Unix(reset, 0).UTC()}
}

func digest(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
