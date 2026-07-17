package githubadapter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

const (
	SandboxRoleReconciler = "reconciler"
	SandboxRoleSeeder     = "seeder"
	SandboxRoleRules      = "rules"
)

var sandboxRoles = []string{SandboxRoleReconciler, SandboxRoleSeeder, SandboxRoleRules}

// SandboxRoleExpectation binds one least-authority credential to its approved App installation.
type SandboxRoleExpectation struct {
	Mode                string
	Actor               string
	Account             string
	InstallationID      string
	RequiredPermissions []string
}

// SandboxConfig is the credential-free, immutable allowlist for one contract sandbox.
type SandboxConfig struct {
	Host                  string
	RESTBaseURL           string
	GraphQLURL            string
	APIVersion            string
	ConfigurationRevision string
	Target                engine.SandboxTarget
	RepositoryOwner       string
	RepositoryName        string
	ProjectNumber         int
	Resources             []engine.SandboxResourceSpec
	Roles                 map[string]SandboxRoleExpectation
	EvidenceMode          string
	LiveTargetApproved    bool
}

type SandboxOption func(*SandboxAdapter)

func WithSandboxClock(clock func() time.Time) SandboxOption {
	return func(adapter *SandboxAdapter) {
		if clock != nil {
			adapter.now = clock
		}
	}
}

// SandboxAdapter implements the engine sandbox seam with role-separated native HTTP clients.
type SandboxAdapter struct {
	config    SandboxConfig
	providers map[string]CredentialProvider
	client    *http.Client
	now       func() time.Time
}

func NewSandbox(config SandboxConfig, providers map[string]CredentialProvider, client *http.Client, options ...SandboxOption) (*SandboxAdapter, error) {
	if config.Host == "" || config.RESTBaseURL == "" || config.GraphQLURL == "" || config.APIVersion != "2026-03-10" || config.ConfigurationRevision == "" || config.RepositoryOwner == "" || config.RepositoryName == "" || config.ProjectNumber <= 0 || client == nil {
		return nil, errors.New("GitHub sandbox adapter configuration is incomplete or unsupported")
	}
	if config.Target.Host != config.Host || config.Target.OwnerID == "" || config.Target.RepositoryID == "" || config.Target.ProjectID == "" || config.Target.RepositoryName != config.RepositoryOwner+"/"+config.RepositoryName {
		return nil, errors.New("GitHub sandbox adapter target identity is inconsistent")
	}
	if config.EvidenceMode == "" {
		config.EvidenceMode = "simulated"
	}
	if !slices.Contains([]string{"simulated", "live"}, config.EvidenceMode) || config.EvidenceMode == "live" && !config.LiveTargetApproved {
		return nil, errors.New("GitHub sandbox adapter evidence mode is unsupported or unapproved")
	}
	for _, endpoint := range []string{config.RESTBaseURL, config.GraphQLURL} {
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, errors.New("GitHub sandbox adapter endpoint is invalid")
		}
		if config.EvidenceMode == "live" && (config.Host != "github.com" || parsed.Scheme != "https" || parsed.Host != "api.github.com") {
			return nil, errors.New("live GitHub sandbox endpoints must use the approved HTTPS API host")
		}
	}
	providerCopy := make(map[string]CredentialProvider, len(sandboxRoles))
	roleCopy := make(map[string]SandboxRoleExpectation, len(sandboxRoles))
	for _, role := range sandboxRoles {
		expectation, exists := config.Roles[role]
		provider := providers[role]
		if !exists || expectation.Mode == "" || expectation.Actor == "" || expectation.Account == "" || expectation.InstallationID == "" || len(expectation.RequiredPermissions) == 0 || provider == nil {
			return nil, fmt.Errorf("GitHub sandbox role %s is not fully configured", role)
		}
		if id, err := strconv.ParseInt(expectation.InstallationID, 10, 64); err != nil || id <= 0 {
			return nil, fmt.Errorf("GitHub sandbox role %s installation is invalid", role)
		}
		expectation.RequiredPermissions = slices.Clone(expectation.RequiredPermissions)
		roleCopy[role] = expectation
		providerCopy[role] = provider
	}
	config.Roles = roleCopy
	config.Resources = cloneSandboxSpecs(config.Resources)
	adapter := &SandboxAdapter{config: config, providers: providerCopy, client: client, now: time.Now}
	for _, option := range options {
		option(adapter)
	}
	return adapter, nil
}

func (adapter *SandboxAdapter) Capability(ctx context.Context) (engine.SandboxCapability, error) {
	now := adapter.now()
	capability := engine.SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: strings.Join(sandboxRoles, "+"), EvidenceMode: adapter.config.EvidenceMode, Target: adapter.config.Target, ConfigurationRevision: adapter.config.ConfigurationRevision, ObservedAt: now}
	for _, role := range sandboxRoles {
		credential, err := adapter.roleCredential(ctx, role)
		if err != nil {
			capability.Available = false
			capability.Fresh = false
			capability.Problems = append(capability.Problems, role+": credential unavailable or mismatched")
			continue
		}
		for _, permission := range credential.Permissions {
			capability.Permissions = append(capability.Permissions, role+":"+permission)
		}
		if capability.ExpiresAt.IsZero() || credential.ExpiresAt.Before(capability.ExpiresAt) {
			capability.ExpiresAt = credential.ExpiresAt
		}
	}
	sort.Strings(capability.Permissions)
	sort.Strings(capability.Problems)
	return capability, nil
}

func (adapter *SandboxAdapter) Observe(ctx context.Context, target engine.SandboxTarget) (engine.SandboxObservation, error) {
	observation := engine.SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: adapter.config.ConfigurationRevision, Resources: []engine.SandboxObservedResource{}}
	if target != adapter.config.Target {
		observation.Problems = []string{"sandbox observation target is outside the immutable allowlist"}
		observation.Revision = sandboxDigest(observation)
		return observation, nil
	}
	credential, err := adapter.roleCredential(ctx, SandboxRoleReconciler)
	if err != nil {
		return observation, errors.New("sandbox reconciler credential is unavailable")
	}
	labels, err := adapter.observeLabels(ctx, credential)
	if err != nil {
		return observation, err
	}
	for _, desired := range adapter.config.Resources {
		if desired.Kind == engine.SandboxResourceLabel {
			if label, exists := labels[desired.Name]; exists {
				observation.Resources = append(observation.Resources, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: label.NodeID, Marker: desired.Marker, Attributes: map[string]string{"color": strings.ToUpper(label.Color), "description": label.Description}})
			}
		}
	}
	projectResources, projectProblems := adapter.observeProject(ctx, credential)
	observation.Resources = append(observation.Resources, projectResources...)
	observation.Problems = append(observation.Problems, projectProblems...)
	sort.Slice(observation.Resources, func(i, j int) bool { return observation.Resources[i].Key < observation.Resources[j].Key })
	observation.Revision = sandboxDigest(observation.Resources)
	return observation, nil
}

func (adapter *SandboxAdapter) Apply(ctx context.Context, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	if effect.Resource.Kind == engine.SandboxResourceProjectWorkflow {
		return engine.SandboxEffectResult{Outcome: "needs-review", Detail: "Project workflow configuration is human-owned and must be enabled in the approved Project UI"}, nil
	}
	if effect.Resource.Kind != engine.SandboxResourceLabel {
		return engine.SandboxEffectResult{Outcome: "not-configured", Detail: "sandbox resource kind has no production effect handler"}, nil
	}
	credential, err := adapter.roleCredential(ctx, SandboxRoleReconciler)
	if err != nil {
		return engine.SandboxEffectResult{}, errors.New("sandbox reconciler credential is unavailable")
	}
	if effect.Kind == "remove-resource" {
		return engine.SandboxEffectResult{Outcome: "needs-review", Detail: "baseline labels are not removed automatically"}, nil
	}
	body := map[string]string{"name": effect.Resource.Name, "color": effect.Resource.Attributes["color"], "description": effect.Resource.Attributes["description"]}
	var label sandboxLabel
	_, err = adapter.rest(ctx, credential, http.MethodPost, "/repos/"+url.PathEscape(adapter.config.RepositoryOwner)+"/"+url.PathEscape(adapter.config.RepositoryName)+"/labels", body, &label)
	if err != nil {
		return engine.SandboxEffectResult{}, err
	}
	return engine.SandboxEffectResult{Outcome: "applied", ResourceID: label.NodeID, Detail: "managed label created"}, nil
}

type projectField struct {
	ID       int             `json:"id"`
	NodeID   string          `json:"node_id"`
	Name     string          `json:"name"`
	DataType string          `json:"data_type"`
	Options  []projectOption `json:"options"`
}

type projectOption struct {
	ID          string      `json:"id"`
	Name        projectText `json:"name"`
	Color       string      `json:"color"`
	Description projectText `json:"description"`
}

type projectText string

func (value *projectText) UnmarshalJSON(data []byte) error {
	var direct string
	if json.Unmarshal(data, &direct) == nil {
		*value = projectText(direct)
		return nil
	}
	var wrapped struct {
		Raw string `json:"raw"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	*value = projectText(wrapped.Raw)
	return nil
}

type projectGraphQLInventory struct {
	Data struct {
		Node struct {
			Views struct {
				Nodes []struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					Number int    `json:"number"`
					Layout string `json:"layout"`
					Filter string `json:"filter"`
				} `json:"nodes"`
			} `json:"views"`
			Workflows struct {
				Nodes []struct {
					ID      string `json:"id"`
					Name    string `json:"name"`
					Number  int    `json:"number"`
					Enabled bool   `json:"enabled"`
				} `json:"nodes"`
			} `json:"workflows"`
		} `json:"node"`
	} `json:"data"`
	Errors []graphQLError `json:"errors"`
}

func (adapter *SandboxAdapter) observeProject(ctx context.Context, credential Credential) ([]engine.SandboxObservedResource, []string) {
	var fields []projectField
	path := "/orgs/" + url.PathEscape(adapter.config.RepositoryOwner) + "/projectsV2/" + strconv.Itoa(adapter.config.ProjectNumber) + "/fields"
	if _, err := adapter.rest(ctx, credential, http.MethodGet, path, nil, &fields); err != nil {
		return nil, []string{"Project field inventory is unavailable"}
	}
	var inventory projectGraphQLInventory
	query := `query($id:ID!){node(id:$id){... on ProjectV2{views(first:50){nodes{id name number layout filter}} workflows(first:50){nodes{id name number enabled}}}}}`
	if err := adapter.graphql(ctx, credential, query, map[string]any{"id": adapter.config.Target.ProjectID}, &inventory); err != nil || len(inventory.Errors) != 0 {
		return nil, []string{"Project view or workflow inventory is unavailable"}
	}
	result := []engine.SandboxObservedResource{}
	for _, desired := range adapter.config.Resources {
		switch desired.Kind {
		case engine.SandboxResourceProjectField:
			for _, field := range fields {
				if field.Name == desired.Name {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: field.NodeID, Marker: desired.Marker, Attributes: map[string]string{"data_type": field.DataType}})
				}
			}
		case engine.SandboxResourceProjectOption:
			for _, field := range fields {
				if field.Name != desired.Attributes["field"] {
					continue
				}
				for _, option := range field.Options {
					if string(option.Name) == desired.Name {
						result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: option.ID, Marker: desired.Marker, Attributes: map[string]string{"field": field.Name, "color": option.Color, "description": string(option.Description)}})
					}
				}
			}
		case engine.SandboxResourceProjectView:
			for _, view := range inventory.Data.Node.Views.Nodes {
				if view.Name == desired.Name {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: view.ID, Marker: desired.Marker, Attributes: map[string]string{"layout": normalizeProjectLayout(view.Layout), "filter": view.Filter, "number": strconv.Itoa(view.Number)}})
				}
			}
		case engine.SandboxResourceProjectWorkflow:
			for _, workflow := range inventory.Data.Node.Workflows.Nodes {
				if workflow.Name == desired.Name {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: workflow.ID, Marker: desired.Marker, Attributes: map[string]string{"enabled": strconv.FormatBool(workflow.Enabled), "number": strconv.Itoa(workflow.Number)}})
				}
			}
		}
	}
	return result, nil
}

func normalizeProjectLayout(value string) string {
	value = strings.TrimSuffix(strings.ToLower(value), "_layout")
	return value
}

type sandboxLabel struct {
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func (adapter *SandboxAdapter) observeLabels(ctx context.Context, credential Credential) (map[string]sandboxLabel, error) {
	var labels []sandboxLabel
	_, err := adapter.rest(ctx, credential, http.MethodGet, "/repos/"+url.PathEscape(adapter.config.RepositoryOwner)+"/"+url.PathEscape(adapter.config.RepositoryName)+"/labels", nil, &labels)
	if err != nil {
		return nil, err
	}
	result := make(map[string]sandboxLabel, len(labels))
	for _, label := range labels {
		result[label.Name] = label
	}
	return result, nil
}

func (adapter *SandboxAdapter) roleCredential(ctx context.Context, role string) (Credential, error) {
	expectation := adapter.config.Roles[role]
	credential, err := adapter.providers[role].Credential(ctx)
	if err != nil || credential.Token == "" || credential.Mode != expectation.Mode || credential.Actor != expectation.Actor || credential.Account != expectation.Account || credential.InstallationID != expectation.InstallationID || !containsAll(credential.Permissions, expectation.RequiredPermissions) || !adapter.now().Before(credential.ExpiresAt) {
		return Credential{}, errors.New("credential does not match approved sandbox role")
	}
	return credential, nil
}

func (adapter *SandboxAdapter) rest(ctx context.Context, credential Credential, method, path string, body, output any) (*http.Response, error) {
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
		return nil, errors.New("GitHub sandbox REST transport is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response, &responseError{StatusCode: response.StatusCode, Header: response.Header.Clone()}
	}
	if output != nil {
		if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(output); err != nil {
			return response, errors.New("decode GitHub sandbox REST response")
		}
	}
	return response, nil
}

func (adapter *SandboxAdapter) graphql(ctx context.Context, credential Credential, query string, variables map[string]any, output any) error {
	encoded, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, adapter.config.GraphQLURL, bytes.NewReader(encoded))
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
		return errors.New("GitHub sandbox GraphQL transport is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &responseError{StatusCode: response.StatusCode, Header: response.Header.Clone()}
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(output); err != nil {
		return errors.New("decode GitHub sandbox GraphQL response")
	}
	return nil
}

func cloneSandboxSpecs(values []engine.SandboxResourceSpec) []engine.SandboxResourceSpec {
	result := make([]engine.SandboxResourceSpec, len(values))
	for index, value := range values {
		attributes := make(map[string]string, len(value.Attributes))
		for key, item := range value.Attributes {
			attributes[key] = item
		}
		value.Attributes = attributes
		result[index] = value
	}
	return result
}

func sandboxDigest(value any) string {
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
