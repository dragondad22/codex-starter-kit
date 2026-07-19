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
	"sync"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

const (
	SandboxRoleReconciler = "reconciler"
	SandboxRoleSeeder     = "seeder"
	SandboxRoleRules      = "rules"
	SandboxRoleReviewer   = "reviewer"
)

var sandboxRoles = []string{SandboxRoleReconciler, SandboxRoleSeeder, SandboxRoleRules}

// SandboxRoleExpectation binds one least-authority credential to its approved App installation.
type SandboxRoleExpectation struct {
	Mode                string   `json:"mode"`
	Actor               string   `json:"actor"`
	Account             string   `json:"account"`
	AccountID           string   `json:"account_id"`
	InstallationID      string   `json:"installation_id,omitempty"`
	RequiredPermissions []string `json:"required_permissions"`
	ClassicOAuthScopes  []string `json:"classic_oauth_scopes,omitempty"`
}

// SandboxConfig is the credential-free, immutable allowlist for one contract sandbox.
type SandboxConfig struct {
	Host                  string                            `json:"host"`
	RESTBaseURL           string                            `json:"rest_base_url"`
	GraphQLURL            string                            `json:"graphql_url"`
	APIVersion            string                            `json:"api_version"`
	ConfigurationRevision string                            `json:"configuration_revision"`
	Target                engine.SandboxTarget              `json:"target"`
	RepositoryOwner       string                            `json:"repository_owner"`
	RepositoryName        string                            `json:"repository_name"`
	ProjectNumber         int                               `json:"project_number"`
	ProjectOwnerKind      string                            `json:"project_owner_kind"`
	Resources             []engine.SandboxResourceSpec      `json:"resources"`
	Roles                 map[string]SandboxRoleExpectation `json:"roles"`
	EvidenceMode          string                            `json:"evidence_mode"`
	LiveTargetApproved    bool                              `json:"live_target_approved"`
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
	roles     []string
	proofMu   sync.Mutex
	proofs    map[string]engine.SandboxObservedResource
}

func NewSandbox(config SandboxConfig, providers map[string]CredentialProvider, client *http.Client, options ...SandboxOption) (*SandboxAdapter, error) {
	return newSandbox(config, providers, sandboxRolesForConfig(config), client, options...)
}

// NewSandboxRole builds one role-scoped adapter for secret-isolated workflow jobs.
func NewSandboxRole(config SandboxConfig, role string, provider CredentialProvider, client *http.Client, options ...SandboxOption) (*SandboxAdapter, error) {
	if !slices.Contains(append(slices.Clone(sandboxRoles), SandboxRoleReviewer), role) {
		return nil, errors.New("GitHub sandbox role is unsupported")
	}
	return newSandbox(config, map[string]CredentialProvider{role: provider}, []string{role}, client, options...)
}

func newSandbox(config SandboxConfig, providers map[string]CredentialProvider, requiredRoles []string, client *http.Client, options ...SandboxOption) (*SandboxAdapter, error) {
	if config.Host == "" || config.RESTBaseURL == "" || config.GraphQLURL == "" || config.APIVersion != "2026-03-10" || config.ConfigurationRevision == "" || config.RepositoryOwner == "" || config.RepositoryName == "" || config.ProjectNumber <= 0 || client == nil {
		return nil, errors.New("GitHub sandbox adapter configuration is incomplete or unsupported")
	}
	if config.ProjectOwnerKind == "" {
		config.ProjectOwnerKind = "organization"
	}
	if !slices.Contains([]string{"organization", "user"}, config.ProjectOwnerKind) {
		return nil, errors.New("GitHub sandbox Project owner kind is unsupported")
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
	providerCopy := make(map[string]CredentialProvider, len(requiredRoles))
	roleCopy := make(map[string]SandboxRoleExpectation, len(requiredRoles))
	for _, role := range requiredRoles {
		expectation, exists := config.Roles[role]
		provider := providers[role]
		accountMatchesTarget := role == SandboxRoleReviewer || expectation.AccountID == config.Target.OwnerID
		installationConfigured := expectation.Mode != "app-installation" || expectation.InstallationID != ""
		if !exists || expectation.Mode == "" || expectation.Actor == "" || expectation.Account == "" || expectation.AccountID == "" || !accountMatchesTarget || !installationConfigured || len(expectation.RequiredPermissions) == 0 || provider == nil {
			return nil, fmt.Errorf("GitHub sandbox role %s is not fully configured", role)
		}
		if expectation.Mode == "app-installation" {
			if id, err := strconv.ParseInt(expectation.InstallationID, 10, 64); err != nil || id <= 0 {
				return nil, fmt.Errorf("GitHub sandbox role %s installation is invalid", role)
			}
		}
		expectation.RequiredPermissions = slices.Clone(expectation.RequiredPermissions)
		expectation.ClassicOAuthScopes = normalizedHeaderTokens(strings.Join(expectation.ClassicOAuthScopes, ","))
		roleCopy[role] = expectation
		providerCopy[role] = provider
	}
	config.Roles = roleCopy
	if config.ProjectOwnerKind == "user" && slices.ContainsFunc(config.Resources, func(resource engine.SandboxResourceSpec) bool {
		return resource.Kind == engine.SandboxResourceProjectView
	}) {
		reconciler, configured := config.Roles[SandboxRoleReconciler]
		if !configured || reconciler.Mode != "user-token" {
			return nil, errors.New("user-owned Project view configuration requires the explicitly selected user-token route")
		}
	}
	config.Resources = cloneSandboxSpecs(config.Resources)
	adapter := &SandboxAdapter{config: config, providers: providerCopy, client: client, now: time.Now, roles: requiredRoles, proofs: map[string]engine.SandboxObservedResource{}}
	for _, option := range options {
		option(adapter)
	}
	return adapter, nil
}

func sandboxRolesForConfig(config SandboxConfig) []string {
	roles := []string{}
	for _, role := range append(slices.Clone(sandboxRoles), SandboxRoleReviewer) {
		if _, configured := config.Roles[role]; configured {
			roles = append(roles, role)
		}
	}
	return roles
}

func (adapter *SandboxAdapter) Capability(ctx context.Context) (engine.SandboxCapability, error) {
	now := adapter.now()
	capability := engine.SandboxCapability{SchemaVersion: 1, Available: true, Fresh: true, Actor: strings.Join(adapter.roles, "+"), EvidenceMode: adapter.config.EvidenceMode, Compatibility: "github.com:api.github.com:2026-03-10:native-rest-graphql", Target: adapter.config.Target, ConfigurationRevision: adapter.config.ConfigurationRevision, ObservedAt: now}
	for _, role := range adapter.roles {
		capability.CredentialIdentities = append(capability.CredentialIdentities, SandboxCredentialIdentity(role, adapter.config.Roles[role]))
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
		if credential.Mode == "user-token" {
			var actor struct {
				Login string `json:"login"`
				ID    int64  `json:"id"`
				Type  string `json:"type"`
			}
			response, identityErr := adapter.rest(ctx, credential, http.MethodGet, "/user", nil, &actor)
			expectation := adapter.config.Roles[role]
			observedScopes := responseHeaderTokens(response, "X-OAuth-Scopes")
			for _, scope := range observedScopes {
				capability.Permissions = append(capability.Permissions, role+":classic-scope:"+scope)
			}
			if identityErr != nil || actor.Login != expectation.Actor || strconv.FormatInt(actor.ID, 10) != expectation.AccountID || actor.Type != "User" || !slices.Contains(observedScopes, "project") || len(expectation.ClassicOAuthScopes) != 0 && !sameStringSet(observedScopes, expectation.ClassicOAuthScopes) {
				capability.Available = false
				capability.Fresh = false
				capability.Problems = append(capability.Problems, role+": user-token actor or exact classic OAuth scope set is unavailable")
				continue
			}
		}
	}
	sort.Strings(capability.Permissions)
	sort.Strings(capability.CredentialIdentities)
	sort.Strings(capability.Problems)
	return capability, nil
}

func responseHeaderTokens(response *http.Response, name string) []string {
	if response == nil {
		return nil
	}
	return normalizedHeaderTokens(response.Header.Get(name))
}

func normalizedHeaderTokens(header string) []string {
	values := []string{}
	seen := map[string]struct{}{}
	for _, value := range strings.Split(header, ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			if _, duplicate := seen[value]; !duplicate {
				values = append(values, value)
				seen[value] = struct{}{}
			}
		}
	}
	sort.Strings(values)
	return values
}

// SandboxCredentialIdentity returns the credential-free identity bound into plans and mandates.
func SandboxCredentialIdentity(role string, expectation SandboxRoleExpectation) string {
	return strings.Join([]string{role, expectation.Mode, expectation.Actor, expectation.Account, expectation.AccountID, expectation.InstallationID}, "|")
}

func (adapter *SandboxAdapter) Observe(ctx context.Context, target engine.SandboxTarget) (engine.SandboxObservation, error) {
	observation := engine.SandboxObservation{SchemaVersion: 1, Target: target, ConfigurationRevision: adapter.config.ConfigurationRevision, Resources: []engine.SandboxObservedResource{}}
	if target != adapter.config.Target {
		observation.Problems = []string{"sandbox observation target is outside the immutable allowlist"}
		observation.Revision = sandboxDigest(observation)
		return observation, nil
	}
	if _, configured := adapter.providers[SandboxRoleReconciler]; configured {
		credential, err := adapter.roleCredential(ctx, SandboxRoleReconciler)
		if err != nil {
			return observation, errors.New("sandbox reconciler credential is unavailable")
		}
		if adapter.hasResourceKind(engine.SandboxResourceLabel) {
			labels, err := adapter.observeLabels(ctx, credential)
			if err != nil {
				return observation, err
			}
			for _, desired := range adapter.config.Resources {
				if desired.Kind == engine.SandboxResourceLabel {
					if label, exists := labels[desired.Name]; exists {
						observation.Resources = append(observation.Resources, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: label.NodeID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"color": strings.ToUpper(label.Color), "description": label.Description})})
					}
				}
			}
		}
		projectResources, projectProblems := adapter.observeProject(ctx, credential)
		observation.Resources = append(observation.Resources, projectResources...)
		observation.Problems = append(observation.Problems, projectProblems...)
	}
	repositoryResources, repositoryProblems := adapter.observeRepositoryResources(ctx)
	observation.Resources = append(observation.Resources, repositoryResources...)
	observation.Problems = append(observation.Problems, repositoryProblems...)
	sort.Slice(observation.Resources, func(i, j int) bool { return observation.Resources[i].Key < observation.Resources[j].Key })
	observation.Revision = sandboxDigest(observation.Resources)
	return observation, nil
}

func (adapter *SandboxAdapter) Apply(ctx context.Context, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	if effect.Resource.Kind == engine.SandboxResourceProjectWorkflow {
		return engine.SandboxEffectResult{Outcome: "needs-review", Detail: "Project workflow configuration is human-owned and must be enabled in the approved Project UI"}, nil
	}
	if effect.Resource.Kind != engine.SandboxResourceLabel {
		return adapter.applyRepositoryResource(ctx, effect)
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
	path := "/repos/" + url.PathEscape(adapter.config.RepositoryOwner) + "/" + url.PathEscape(adapter.config.RepositoryName) + "/labels"
	labels, err := adapter.observeLabels(ctx, credential)
	if err != nil {
		return engine.SandboxEffectResult{}, err
	}
	method := http.MethodPost
	detail := "managed label created"
	if _, exists := labels[effect.Resource.Name]; exists {
		method = http.MethodPatch
		path += "/" + url.PathEscape(effect.Resource.Name)
		body["new_name"] = effect.Resource.Name
		detail = "managed label updated"
	}
	_, err = adapter.rest(ctx, credential, method, path, body, &label)
	if err != nil {
		return engine.SandboxEffectResult{}, err
	}
	return engine.SandboxEffectResult{Outcome: "applied", ResourceID: label.NodeID, Detail: detail}, nil
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
					Fields struct {
						Nodes []struct {
							ID string `json:"id"`
						} `json:"nodes"`
					} `json:"fields"`
					GroupByFields struct {
						Nodes []struct {
							ID string `json:"id"`
						} `json:"nodes"`
					} `json:"groupByFields"`
					SortByFields struct {
						Nodes []struct {
							Direction string `json:"direction"`
							Field     struct {
								ID string `json:"id"`
							} `json:"field"`
						} `json:"nodes"`
					} `json:"sortByFields"`
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
			Items struct {
				Nodes []struct {
					ID      string `json:"id"`
					Content struct {
						ID     string `json:"id"`
						Number int    `json:"number"`
						Title  string `json:"title"`
						Body   string `json:"body"`
						State  string `json:"state"`
					} `json:"content"`
					Status struct {
						Name string `json:"name"`
					} `json:"fieldValueByName"`
					FieldValues struct {
						Nodes []struct {
							OptionID string `json:"optionId"`
							Field    struct {
								ID string `json:"id"`
							} `json:"field"`
						} `json:"nodes"`
					} `json:"fieldValues"`
				} `json:"nodes"`
			} `json:"items"`
		} `json:"node"`
	} `json:"data"`
	Errors []graphQLError `json:"errors"`
}

func (adapter *SandboxAdapter) observeProject(ctx context.Context, credential Credential) ([]engine.SandboxObservedResource, []string) {
	var fields []projectField
	path := adapter.projectRESTPath() + "/fields"
	if _, err := adapter.rest(ctx, credential, http.MethodGet, path, nil, &fields); err != nil {
		return nil, []string{"Project field inventory is unavailable"}
	}
	var inventory projectGraphQLInventory
	query := `query($id:ID!){node(id:$id){... on ProjectV2{views(first:50){nodes{id name number layout filter fields(first:50){nodes{... on ProjectV2FieldCommon{id}}} groupByFields(first:10){nodes{... on ProjectV2FieldCommon{id}}} sortByFields(first:10){nodes{direction field{... on ProjectV2FieldCommon{id}}}}}} workflows(first:50){nodes{id name number enabled}} items(first:100){nodes{id content{... on Issue{id number title body state}} fieldValueByName(name:"Status"){... on ProjectV2ItemFieldSingleSelectValue{name}} fieldValues(first:50){nodes{... on ProjectV2ItemFieldSingleSelectValue{optionId field{... on ProjectV2FieldCommon{id}}}}}}}}}}`
	if err := adapter.graphql(ctx, credential, query, map[string]any{"id": adapter.config.Target.ProjectID}, &inventory); err != nil || len(inventory.Errors) != 0 {
		return nil, []string{"Project view or workflow inventory is unavailable"}
	}
	result := []engine.SandboxObservedResource{}
	problems := adapter.projectCatalogProblems(fields)
	for _, desired := range adapter.config.Resources {
		switch desired.Kind {
		case engine.SandboxResourceProjectField:
			for _, field := range fields {
				if field.Name == desired.Name {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: field.NodeID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"data_type": field.DataType, "node_id": field.NodeID})})
				}
			}
		case engine.SandboxResourceProjectOption:
			for _, field := range fields {
				if field.Name != desired.Attributes["field"] {
					continue
				}
				for _, option := range field.Options {
					if string(option.Name) == desired.Name {
						result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: option.ID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"field": field.Name, "color": option.Color, "description": string(option.Description), "option_id": option.ID})})
					}
				}
			}
		case engine.SandboxResourceProjectView:
			for _, view := range inventory.Data.Node.Views.Nodes {
				if view.Name == desired.Name {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: view.ID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"layout": normalizeProjectLayout(view.Layout), "filter": view.Filter, "number": strconv.Itoa(view.Number), "node_id": view.ID, "visible_fields": viewFieldIDs(view.Fields.Nodes), "group_by": viewFieldIDs(view.GroupByFields.Nodes), "sort_by": viewSortFields(view.SortByFields.Nodes)})})
				}
			}
		case engine.SandboxResourceProjectItemField:
			for _, item := range inventory.Data.Node.Items.Nodes {
				if item.Content.ID != desired.Attributes["content_id"] {
					continue
				}
				for _, value := range item.FieldValues.Nodes {
					if value.Field.ID == desired.Attributes["field_id"] && value.OptionID == desired.Attributes["option_id"] {
						result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: item.ID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"content_id": item.Content.ID, "field": desired.Attributes["field"], "field_id": value.Field.ID, "option_id": value.OptionID, "item_id": item.ID})})
					}
				}
			}
		case engine.SandboxResourceProjectWorkflow:
			for _, workflow := range inventory.Data.Node.Workflows.Nodes {
				if workflow.Name == desired.Name {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: workflow.ID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"enabled": strconv.FormatBool(workflow.Enabled), "number": strconv.Itoa(workflow.Number)})})
				}
			}
		case engine.SandboxResourceProjectItemProof:
			for _, item := range inventory.Data.Node.Items.Nodes {
				if strings.Contains(item.Content.Body, desired.Marker) {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: item.ID, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"number": strconv.Itoa(item.Content.Number), "state": strings.ToLower(item.Content.State), "status": item.Status.Name, "content_id": item.Content.ID})})
				}
			}
		}
	}
	return result, problems
}

func (adapter *SandboxAdapter) projectCatalogProblems(fields []projectField) []string {
	var desiredField *engine.SandboxResourceSpec
	desiredOptions := map[string]engine.SandboxResourceSpec{}
	for index := range adapter.config.Resources {
		resource := &adapter.config.Resources[index]
		if resource.Kind == engine.SandboxResourceProjectField && resource.Name == "Phase" {
			desiredField = resource
		}
		if resource.Kind == engine.SandboxResourceProjectOption && resource.Attributes["field"] == "Phase" {
			desiredOptions[resource.Name] = *resource
		}
	}
	if desiredField == nil {
		return nil
	}
	matching := []projectField{}
	for _, field := range fields {
		if field.Name == desiredField.Name {
			matching = append(matching, field)
		}
	}
	if len(matching) == 0 {
		return nil
	}
	if len(matching) != 1 {
		return []string{"Project must expose exactly one governed Phase field"}
	}
	field := matching[0]
	if normalizeProjectDataType(field.DataType) != normalizeProjectDataType(desiredField.Attributes["data_type"]) || desiredField.Attributes["node_id"] != "" && field.NodeID != desiredField.Attributes["node_id"] {
		return []string{"Project Phase field type or immutable identity conflicts with the governed catalog"}
	}
	if len(field.Options) != len(desiredOptions) {
		return []string{"Project must expose the complete Phase option catalog with no extras"}
	}
	seenNames := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	for _, option := range field.Options {
		name := string(option.Name)
		desired, exists := desiredOptions[name]
		_, duplicateName := seenNames[name]
		_, duplicateID := seenIDs[option.ID]
		if !exists || duplicateName || duplicateID || desired.Attributes["option_id"] != "" && option.ID != desired.Attributes["option_id"] {
			return []string{"Project must expose the complete Phase option catalog with stable names and immutable IDs"}
		}
		seenNames[name] = struct{}{}
		seenIDs[option.ID] = struct{}{}
	}
	return nil
}

func normalizeProjectDataType(value string) string {
	return strings.ReplaceAll(strings.ToLower(value), "_", "")
}

func normalizeProjectLayout(value string) string {
	value = strings.TrimSuffix(strings.ToLower(value), "_layout")
	return value
}

func viewFieldIDs(fields []struct {
	ID string `json:"id"`
}) string {
	ids := make([]string, 0, len(fields))
	for _, field := range fields {
		ids = append(ids, field.ID)
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func viewSortFields(fields []struct {
	Direction string `json:"direction"`
	Field     struct {
		ID string `json:"id"`
	} `json:"field"`
}) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		values = append(values, field.Field.ID+":"+strings.ToLower(field.Direction))
	}
	sort.Strings(values)
	return strings.Join(values, ",")
}

func (adapter *SandboxAdapter) projectRESTPath() string {
	if adapter.config.ProjectOwnerKind == "user" {
		return "/users/" + url.PathEscape(adapter.config.RepositoryOwner) + "/projectsV2/" + strconv.Itoa(adapter.config.ProjectNumber)
	}
	return "/orgs/" + url.PathEscape(adapter.config.RepositoryOwner) + "/projectsV2/" + strconv.Itoa(adapter.config.ProjectNumber)
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
	if err != nil || credential.Token == "" || credential.Mode != expectation.Mode || credential.Actor != expectation.Actor || credential.Account != expectation.Account || credential.AccountID != expectation.AccountID || credential.InstallationID != expectation.InstallationID || !sameStringSet(credential.Permissions, expectation.RequiredPermissions) || !adapter.now().Before(credential.ExpiresAt) {
		return Credential{}, errors.New("credential does not match approved sandbox role")
	}
	return credential, nil
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	left = slices.Clone(left)
	right = slices.Clone(right)
	sort.Strings(left)
	sort.Strings(right)
	return slices.Equal(left, right)
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
