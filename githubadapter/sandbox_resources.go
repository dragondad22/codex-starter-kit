package githubadapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func (adapter *SandboxAdapter) observeRepositoryResources(ctx context.Context) ([]engine.SandboxObservedResource, []string) {
	result := []engine.SandboxObservedResource{}
	problems := []string{}
	if adapter.hasResourceKind(engine.SandboxResourceRuleset) {
		credential, err := adapter.roleCredential(ctx, SandboxRoleRules)
		if err != nil {
			problems = append(problems, "rules credential is unavailable")
		} else if resources, err := adapter.observeRulesets(ctx, credential); err != nil {
			problems = append(problems, "ruleset inventory is unavailable")
		} else {
			result = append(result, resources...)
		}
	}
	if adapter.hasAnyResourceKind(engine.SandboxResourceFixtureIssue, engine.SandboxResourceFixtureBranch, engine.SandboxResourceFixturePR, engine.SandboxResourceFixtureWorkflow) {
		credential, err := adapter.roleCredential(ctx, SandboxRoleSeeder)
		if err != nil {
			problems = append(problems, "seeder credential is unavailable")
		} else if resources, err := adapter.observeFixtures(ctx, credential); err != nil {
			problems = append(problems, "fixture inventory is unavailable")
		} else {
			result = append(result, resources...)
		}
	}
	if adapter.hasResourceKind(engine.SandboxResourceFixtureReview) {
		credential, err := adapter.roleCredential(ctx, SandboxRoleReviewer)
		if err != nil {
			problems = append(problems, "reviewer credential is unavailable")
		} else if resources, err := adapter.observeReviews(ctx, credential); err != nil {
			problems = append(problems, "review inventory is unavailable")
		} else {
			result = append(result, resources...)
		}
	}
	result = append(result, adapter.observeEphemeralProofs()...)
	return result, problems
}

func (adapter *SandboxAdapter) applyRepositoryResource(ctx context.Context, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	switch effect.Resource.Kind {
	case engine.SandboxResourceProjectField, engine.SandboxResourceProjectOption, engine.SandboxResourceProjectView:
		credential, err := adapter.roleCredential(ctx, SandboxRoleReconciler)
		if err != nil {
			return engine.SandboxEffectResult{}, errors.New("sandbox reconciler credential is unavailable")
		}
		return adapter.applyProjectResource(ctx, credential, effect)
	case engine.SandboxResourceProjectItemProof:
		return engine.SandboxEffectResult{Outcome: "fail", Detail: "built-in Project workflow proof is observation-only; missing or drifted state is not repaired"}, nil
	case engine.SandboxResourceRuleset:
		credential, err := adapter.roleCredential(ctx, SandboxRoleRules)
		if err != nil {
			return engine.SandboxEffectResult{}, errors.New("sandbox rules credential is unavailable")
		}
		return adapter.applyRuleset(ctx, credential, effect)
	case engine.SandboxResourceFixtureIssue, engine.SandboxResourceFixtureBranch, engine.SandboxResourceFixturePR, engine.SandboxResourceFixtureWorkflow:
		credential, err := adapter.roleCredential(ctx, SandboxRoleSeeder)
		if err != nil {
			return engine.SandboxEffectResult{}, errors.New("sandbox seeder credential is unavailable")
		}
		return adapter.applyFixture(ctx, credential, effect)
	case engine.SandboxResourceFixtureReview:
		credential, err := adapter.roleCredential(ctx, SandboxRoleReviewer)
		if err != nil {
			return engine.SandboxEffectResult{}, errors.New("sandbox reviewer credential is unavailable")
		}
		return adapter.applyReview(ctx, credential, effect)
	case engine.SandboxResourceFixtureDenial:
		return adapter.applyFixtureDenial(ctx, effect)
	case engine.SandboxResourceTokenRevocation:
		return adapter.applyTokenRevocation(ctx, effect)
	default:
		return engine.SandboxEffectResult{Outcome: "not-configured", Detail: "sandbox resource kind has no production effect handler"}, nil
	}
}

func (adapter *SandboxAdapter) observeEphemeralProofs() []engine.SandboxObservedResource {
	adapter.proofMu.Lock()
	defer adapter.proofMu.Unlock()
	result := []engine.SandboxObservedResource{}
	for _, desired := range adapter.config.Resources {
		if desired.Kind != engine.SandboxResourceFixtureDenial && desired.Kind != engine.SandboxResourceTokenRevocation {
			continue
		}
		if proof, exists := adapter.proofs[desired.Key]; exists {
			result = append(result, proof)
		}
	}
	return result
}

func (adapter *SandboxAdapter) retainEphemeralProof(resource engine.SandboxResourceSpec, id string) {
	adapter.proofMu.Lock()
	defer adapter.proofMu.Unlock()
	adapter.proofs[resource.Key] = engine.SandboxObservedResource{
		Key: resource.Key, Kind: resource.Kind, Name: resource.Name, ID: id, Marker: resource.Marker,
		Attributes: desiredAttributes(resource, resource.Attributes),
	}
}

func (adapter *SandboxAdapter) applyFixtureDenial(ctx context.Context, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	if effect.Kind == "remove-resource" {
		return engine.SandboxEffectResult{Outcome: "not-applicable", Detail: "denial proofs are ephemeral run evidence"}, nil
	}
	credential, err := adapter.roleCredential(ctx, SandboxRoleSeeder)
	if err != nil {
		return engine.SandboxEffectResult{}, errors.New("sandbox seeder credential is unavailable")
	}
	path := adapter.repoPath() + "/git/refs/heads/" + escapePath(effect.Resource.Attributes["branch"])
	_, err = adapter.rest(ctx, credential, http.MethodDelete, path, nil, nil)
	if err == nil {
		return engine.SandboxEffectResult{Outcome: "fail", Detail: "fixture branch deletion was unexpectedly allowed"}, nil
	}
	status := http.StatusForbidden
	if isResponseStatus(err, http.StatusUnprocessableEntity) {
		status = http.StatusUnprocessableEntity
	} else if !isResponseStatus(err, http.StatusForbidden) {
		return engine.SandboxEffectResult{}, err
	}
	proofID := "http-" + strconv.Itoa(status)
	adapter.retainEphemeralProof(effect.Resource, proofID)
	return engine.SandboxEffectResult{Outcome: "applied", ResourceID: proofID, Detail: "active fixture ruleset denied branch deletion"}, nil
}

func (adapter *SandboxAdapter) applyTokenRevocation(ctx context.Context, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	if effect.Kind == "remove-resource" {
		return engine.SandboxEffectResult{Outcome: "not-applicable", Detail: "revocation proofs are ephemeral run evidence"}, nil
	}
	role := effect.Resource.Attributes["role"]
	if role == SandboxRoleReviewer {
		return engine.SandboxEffectResult{Outcome: "not-applicable", Detail: "reviewer credential revocation is human-owned"}, nil
	}
	credential, err := adapter.roleCredential(ctx, role)
	if err != nil {
		return engine.SandboxEffectResult{}, errors.New("sandbox App credential is unavailable for revocation")
	}
	if _, err := adapter.rest(ctx, credential, http.MethodDelete, "/installation/token", nil, nil); err != nil {
		return engine.SandboxEffectResult{}, err
	}
	if _, err := adapter.rest(ctx, credential, http.MethodGet, "/installation/repositories", nil, nil); !isResponseStatus(err, http.StatusUnauthorized) {
		return engine.SandboxEffectResult{Outcome: "fail", Detail: "revoked App credential remained usable or returned an unexpected state"}, nil
	}
	adapter.retainEphemeralProof(effect.Resource, "http-401")
	return engine.SandboxEffectResult{Outcome: "applied", ResourceID: "http-401", Detail: "App installation credential was revoked and rejected"}, nil
}

func (adapter *SandboxAdapter) applyProjectResource(ctx context.Context, credential Credential, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	path := "/orgs/" + url.PathEscape(adapter.config.RepositoryOwner) + "/projectsV2/" + strconv.Itoa(adapter.config.ProjectNumber)
	switch effect.Resource.Kind {
	case engine.SandboxResourceProjectView:
		if effect.Kind == "remove-resource" {
			return engine.SandboxEffectResult{Outcome: "needs-review", Detail: "Project view removal is not available in the approved public API contract"}, nil
		}
		existing, problems := adapter.observeProject(ctx, credential)
		for _, problem := range problems {
			return engine.SandboxEffectResult{}, errors.New(problem)
		}
		for _, resource := range existing {
			if resource.Key == effect.Resource.Key {
				return engine.SandboxEffectResult{Outcome: "needs-review", ResourceID: resource.ID, Detail: "existing Project view drift requires human-owned view reconciliation"}, nil
			}
		}
		body := map[string]any{"name": effect.Resource.Name, "layout": effect.Resource.Attributes["layout"]}
		if filter := effect.Resource.Attributes["filter"]; filter != "" {
			body["filter"] = filter
		}
		var response struct {
			Value struct {
				NodeID string `json:"node_id"`
			} `json:"value"`
		}
		if _, err := adapter.rest(ctx, credential, http.MethodPost, path+"/views", body, &response); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: response.Value.NodeID, Detail: "Project view created"}, nil
	case engine.SandboxResourceProjectField:
		if effect.Kind == "remove-resource" {
			return engine.SandboxEffectResult{Outcome: "needs-review", Detail: "baseline Project fields are not removed automatically"}, nil
		}
		input := map[string]any{"projectId": adapter.config.Target.ProjectID, "name": effect.Resource.Name, "dataType": strings.ToUpper(effect.Resource.Attributes["data_type"])}
		options := adapter.desiredProjectOptions(effect.Resource.Name)
		if len(options) != 0 {
			input["singleSelectOptions"] = options
		}
		var response struct {
			Data struct {
				Create struct {
					Field struct {
						ID string `json:"id"`
					} `json:"projectV2Field"`
				} `json:"createProjectV2Field"`
			} `json:"data"`
			Errors []graphQLError `json:"errors"`
		}
		query := `mutation($input:CreateProjectV2FieldInput!){createProjectV2Field(input:$input){projectV2Field{... on ProjectV2FieldCommon{id}}}}`
		if err := adapter.graphql(ctx, credential, query, map[string]any{"input": input}, &response); err != nil || len(response.Errors) != 0 {
			return engine.SandboxEffectResult{}, errors.New("Project field creation failed")
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: response.Data.Create.Field.ID, Detail: "Project field and approved options created"}, nil
	case engine.SandboxResourceProjectOption:
		return adapter.applyProjectOption(ctx, credential, effect)
	}
	return engine.SandboxEffectResult{Outcome: "not-configured", Detail: "Project resource kind is unsupported"}, nil
}

func (adapter *SandboxAdapter) applyProjectOption(ctx context.Context, credential Credential, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	fields, err := adapter.projectFields(ctx, credential)
	if err != nil {
		return engine.SandboxEffectResult{}, err
	}
	fieldName := effect.Resource.Attributes["field"]
	for _, field := range fields {
		if field.Name != fieldName {
			continue
		}
		options := adapter.desiredProjectOptions(fieldName)
		input := map[string]any{"fieldId": field.NodeID, "singleSelectOptions": options}
		var response struct {
			Data struct {
				Update struct {
					Field struct {
						ID string `json:"id"`
					} `json:"projectV2Field"`
				} `json:"updateProjectV2Field"`
			} `json:"data"`
			Errors []graphQLError `json:"errors"`
		}
		query := `mutation($input:UpdateProjectV2FieldInput!){updateProjectV2Field(input:$input){projectV2Field{... on ProjectV2FieldCommon{id}}}}`
		if err := adapter.graphql(ctx, credential, query, map[string]any{"input": input}, &response); err != nil || len(response.Errors) != 0 {
			return engine.SandboxEffectResult{}, errors.New("Project option reconciliation failed")
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: response.Data.Update.Field.ID, Detail: "Project single-select options reconciled atomically"}, nil
	}
	return engine.SandboxEffectResult{Outcome: "needs-review", Detail: "Project option cannot be reconciled before its field exists"}, nil
}

func (adapter *SandboxAdapter) desiredProjectOptions(field string) []map[string]string {
	options := []map[string]string{}
	for _, resource := range adapter.config.Resources {
		if resource.Kind == engine.SandboxResourceProjectOption && resource.Attributes["field"] == field && resource.DesiredState != engine.SandboxResourceAbsent {
			options = append(options, map[string]string{"id": resource.Attributes["input:id"], "name": resource.Name, "color": resource.Attributes["color"], "description": resource.Attributes["description"]})
			if options[len(options)-1]["id"] == "" {
				delete(options[len(options)-1], "id")
			}
		}
	}
	return options
}

func (adapter *SandboxAdapter) projectFields(ctx context.Context, credential Credential) ([]projectField, error) {
	var fields []projectField
	path := "/orgs/" + url.PathEscape(adapter.config.RepositoryOwner) + "/projectsV2/" + strconv.Itoa(adapter.config.ProjectNumber) + "/fields"
	_, err := adapter.rest(ctx, credential, http.MethodGet, path, nil, &fields)
	return fields, err
}

type sandboxRuleset struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Enforcement string `json:"enforcement"`
	Target      string `json:"target"`
}

func (adapter *SandboxAdapter) observeRulesets(ctx context.Context, credential Credential) ([]engine.SandboxObservedResource, error) {
	var rulesets []sandboxRuleset
	if _, err := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/rulesets?includes_parents=false", nil, &rulesets); err != nil {
		return nil, err
	}
	result := []engine.SandboxObservedResource{}
	for _, desired := range adapter.config.Resources {
		if desired.Kind != engine.SandboxResourceRuleset {
			continue
		}
		for _, ruleset := range rulesets {
			if ruleset.Name == desired.Name && (desired.Marker == "" || strings.Contains(ruleset.Name, desired.Marker)) {
				result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: strconv.FormatInt(ruleset.ID, 10), Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"enforcement": ruleset.Enforcement, "target": ruleset.Target})})
			}
		}
	}
	return result, nil
}

func (adapter *SandboxAdapter) applyRuleset(ctx context.Context, credential Credential, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	resources, err := adapter.observeRulesets(ctx, credential)
	if err != nil {
		return engine.SandboxEffectResult{}, err
	}
	var existingID string
	for _, resource := range resources {
		if resource.Key == effect.Resource.Key {
			existingID = resource.ID
		}
	}
	if effect.Kind == "remove-resource" {
		if existingID == "" {
			return engine.SandboxEffectResult{Outcome: "no-change", Detail: "marked fixture ruleset is absent"}, nil
		}
		if _, err := adapter.rest(ctx, credential, http.MethodDelete, adapter.repoPath()+"/rulesets/"+existingID, nil, nil); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: existingID, Detail: "marked fixture ruleset deleted"}, nil
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(effect.Resource.Attributes["input:definition"]), &body); err != nil {
		return engine.SandboxEffectResult{}, errors.New("ruleset definition is invalid")
	}
	body["name"] = effect.Resource.Name
	method := http.MethodPost
	path := adapter.repoPath() + "/rulesets"
	if existingID != "" {
		method = http.MethodPut
		path += "/" + existingID
	}
	var ruleset sandboxRuleset
	if _, err := adapter.rest(ctx, credential, method, path, body, &ruleset); err != nil {
		return engine.SandboxEffectResult{}, err
	}
	return engine.SandboxEffectResult{Outcome: "applied", ResourceID: strconv.FormatInt(ruleset.ID, 10), Detail: "marked fixture ruleset reconciled"}, nil
}

type sandboxIssue struct {
	ID          int64     `json:"id"`
	NodeID      string    `json:"node_id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	PullRequest *struct{} `json:"pull_request"`
}

type sandboxPullRequest struct {
	ID     int64  `json:"id"`
	NodeID string `json:"node_id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Draft  bool   `json:"draft"`
	Head   struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

func (adapter *SandboxAdapter) observeFixtures(ctx context.Context, credential Credential) ([]engine.SandboxObservedResource, error) {
	result := []engine.SandboxObservedResource{}
	var issues []sandboxIssue
	if adapter.hasResourceKind(engine.SandboxResourceFixtureIssue) {
		if _, err := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/issues?state=all&per_page=100", nil, &issues); err != nil {
			return nil, err
		}
	}
	var pulls []sandboxPullRequest
	if adapter.hasResourceKind(engine.SandboxResourceFixturePR) {
		if _, err := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/pulls?state=all&per_page=100", nil, &pulls); err != nil {
			return nil, err
		}
	}
	for _, desired := range adapter.config.Resources {
		switch desired.Kind {
		case engine.SandboxResourceFixtureIssue:
			for _, issue := range issues {
				if issue.PullRequest == nil && strings.Contains(issue.Body, desired.Marker) {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: strconv.Itoa(issue.Number), Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"title": issue.Title, "state": issue.State, "node_id": issue.NodeID})})
				}
			}
		case engine.SandboxResourceFixtureBranch:
			var ref struct {
				Ref    string `json:"ref"`
				Object struct {
					SHA string `json:"sha"`
				} `json:"object"`
			}
			_, err := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/git/ref/heads/"+escapePath(desired.Name), nil, &ref)
			if isResponseStatus(err, http.StatusNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: ref.Object.SHA, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"sha": ref.Object.SHA})})
		case engine.SandboxResourceFixturePR:
			for _, pull := range pulls {
				if strings.Contains(pull.Body, desired.Marker) {
					result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: strconv.Itoa(pull.Number), Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"title": pull.Title, "state": pull.State, "draft": strconv.FormatBool(pull.Draft), "head": pull.Head.Ref, "base": pull.Base.Ref, "node_id": pull.NodeID})})
				}
			}
		case engine.SandboxResourceFixtureWorkflow:
			path := desired.Attributes["path"]
			var content struct {
				SHA     string `json:"sha"`
				Content string `json:"content"`
			}
			_, err := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/contents/"+escapePath(path), nil, &content)
			if isResponseStatus(err, http.StatusNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
			if err != nil || string(decoded) != desired.Attributes["input:content"] {
				continue
			}
			result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: content.SHA, Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"path": path, "content_sha256": sandboxDigest(string(decoded))})})
		}
	}
	return result, nil
}

func (adapter *SandboxAdapter) applyFixture(ctx context.Context, credential Credential, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	switch effect.Resource.Kind {
	case engine.SandboxResourceFixtureIssue:
		if effect.Kind == "remove-resource" {
			return adapter.closeFixture(ctx, credential, "issues", effect)
		}
		body := map[string]any{"title": effect.Resource.Attributes["title"], "body": markerBody(effect.Resource)}
		if raw := effect.Resource.Attributes["input:labels"]; raw != "" {
			body["labels"] = strings.Split(raw, ",")
		}
		method := http.MethodPost
		path := adapter.repoPath() + "/issues"
		resources, err := adapter.observeFixtures(ctx, credential)
		if err != nil {
			return engine.SandboxEffectResult{}, err
		}
		for _, resource := range resources {
			if resource.Key == effect.Resource.Key {
				method = http.MethodPatch
				path += "/" + resource.ID
				body["state"] = effect.Resource.Attributes["state"]
			}
		}
		var issue sandboxIssue
		if _, err := adapter.rest(ctx, credential, method, path, body, &issue); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: strconv.Itoa(issue.Number), Detail: "marked fixture issue created"}, nil
	case engine.SandboxResourceFixtureBranch:
		path := adapter.repoPath() + "/git/refs"
		if effect.Kind == "remove-resource" {
			path = adapter.repoPath() + "/git/refs/heads/" + escapePath(effect.Resource.Name)
			if _, err := adapter.rest(ctx, credential, http.MethodDelete, path, nil, nil); err != nil && !isResponseStatus(err, http.StatusNotFound) {
				return engine.SandboxEffectResult{}, err
			}
			return engine.SandboxEffectResult{Outcome: "applied", ResourceID: effect.Resource.Name, Detail: "marked fixture branch deleted"}, nil
		}
		var ref struct {
			Object struct {
				SHA string `json:"sha"`
			} `json:"object"`
		}
		if _, err := adapter.rest(ctx, credential, http.MethodPost, path, map[string]string{"ref": "refs/heads/" + effect.Resource.Name, "sha": effect.Resource.Attributes["input:base_sha"]}, &ref); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		if contentPath := effect.Resource.Attributes["input:path"]; contentPath != "" {
			body := map[string]any{"message": effect.Resource.Marker, "content": base64.StdEncoding.EncodeToString([]byte(effect.Resource.Attributes["input:content"])), "branch": effect.Resource.Name}
			var contentResponse struct {
				Commit struct {
					SHA string `json:"sha"`
				} `json:"commit"`
			}
			if _, err := adapter.rest(ctx, credential, http.MethodPut, adapter.repoPath()+"/contents/"+escapePath(contentPath), body, &contentResponse); err != nil {
				return engine.SandboxEffectResult{}, err
			}
			ref.Object.SHA = contentResponse.Commit.SHA
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: ref.Object.SHA, Detail: "marked fixture branch created"}, nil
	case engine.SandboxResourceFixturePR:
		if effect.Kind == "remove-resource" {
			return adapter.closeFixture(ctx, credential, "pulls", effect)
		}
		body := map[string]any{"title": effect.Resource.Attributes["title"], "body": markerBody(effect.Resource), "head": effect.Resource.Attributes["head"], "base": effect.Resource.Attributes["base"], "draft": effect.Resource.Attributes["draft"] == "true"}
		method := http.MethodPost
		path := adapter.repoPath() + "/pulls"
		resources, err := adapter.observeFixtures(ctx, credential)
		if err != nil {
			return engine.SandboxEffectResult{}, err
		}
		var existingNodeID string
		var existingDraft bool
		for _, resource := range resources {
			if resource.Key == effect.Resource.Key {
				method = http.MethodPatch
				path += "/" + resource.ID
				var existing sandboxPullRequest
				if _, err := adapter.rest(ctx, credential, http.MethodGet, path, nil, &existing); err != nil {
					return engine.SandboxEffectResult{}, err
				}
				existingNodeID = existing.NodeID
				existingDraft = existing.Draft
				delete(body, "head")
				delete(body, "draft")
				body["state"] = effect.Resource.Attributes["state"]
			}
		}
		var pull sandboxPullRequest
		if _, err := adapter.rest(ctx, credential, method, path, body, &pull); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		desiredDraft := effect.Resource.Attributes["draft"] == "true"
		if existingNodeID != "" && existingDraft != desiredDraft {
			mutation := `mutation($id:ID!){markPullRequestReadyForReview(input:{pullRequestId:$id}){pullRequest{id}}}`
			if desiredDraft {
				mutation = `mutation($id:ID!){convertPullRequestToDraft(input:{pullRequestId:$id}){pullRequest{id}}}`
			}
			var response struct {
				Errors []graphQLError `json:"errors"`
			}
			if err := adapter.graphql(ctx, credential, mutation, map[string]any{"id": existingNodeID}, &response); err != nil {
				return engine.SandboxEffectResult{}, err
			}
			if len(response.Errors) != 0 {
				return engine.SandboxEffectResult{}, errors.New("GitHub rejected the fixture pull request draft transition")
			}
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: strconv.Itoa(pull.Number), Detail: "marked fixture pull request created"}, nil
	case engine.SandboxResourceFixtureWorkflow:
		path := effect.Resource.Attributes["path"]
		var existing struct {
			SHA string `json:"sha"`
		}
		_, getErr := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/contents/"+escapePath(path), nil, &existing)
		if getErr != nil && !isResponseStatus(getErr, http.StatusNotFound) {
			return engine.SandboxEffectResult{}, getErr
		}
		body := map[string]any{"message": effect.Resource.Marker, "content": base64.StdEncoding.EncodeToString([]byte(effect.Resource.Attributes["input:content"]))}
		if existing.SHA != "" {
			body["sha"] = existing.SHA
		}
		if effect.Kind == "remove-resource" {
			if existing.SHA == "" {
				return engine.SandboxEffectResult{Outcome: "no-change", Detail: "marked fixture workflow is absent"}, nil
			}
			body = map[string]any{"message": effect.Resource.Marker, "sha": existing.SHA}
			if _, err := adapter.rest(ctx, credential, http.MethodDelete, adapter.repoPath()+"/contents/"+escapePath(path), body, nil); err != nil {
				return engine.SandboxEffectResult{}, err
			}
			return engine.SandboxEffectResult{Outcome: "applied", ResourceID: existing.SHA, Detail: "marked fixture workflow deleted"}, nil
		}
		var response struct {
			Content struct {
				SHA string `json:"sha"`
			} `json:"content"`
		}
		if _, err := adapter.rest(ctx, credential, http.MethodPut, adapter.repoPath()+"/contents/"+escapePath(path), body, &response); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: response.Content.SHA, Detail: "marked fixture workflow reconciled"}, nil
	}
	return engine.SandboxEffectResult{Outcome: "not-configured", Detail: "fixture resource kind is unsupported"}, nil
}

func (adapter *SandboxAdapter) closeFixture(ctx context.Context, credential Credential, family string, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	resources, err := adapter.observeFixtures(ctx, credential)
	if err != nil {
		return engine.SandboxEffectResult{}, err
	}
	for _, resource := range resources {
		if resource.Key != effect.Resource.Key {
			continue
		}
		if _, err := adapter.rest(ctx, credential, http.MethodPatch, adapter.repoPath()+"/"+family+"/"+resource.ID, map[string]string{"state": "closed"}, nil); err != nil {
			return engine.SandboxEffectResult{}, err
		}
		return engine.SandboxEffectResult{Outcome: "applied", ResourceID: resource.ID, Detail: "marked fixture closed"}, nil
	}
	return engine.SandboxEffectResult{Outcome: "no-change", Detail: "marked fixture is absent"}, nil
}

func (adapter *SandboxAdapter) observeReviews(ctx context.Context, credential Credential) ([]engine.SandboxObservedResource, error) {
	result := []engine.SandboxObservedResource{}
	for _, desired := range adapter.config.Resources {
		if desired.Kind != engine.SandboxResourceFixtureReview {
			continue
		}
		var reviews []struct {
			ID     int64  `json:"id"`
			Body   string `json:"body"`
			State  string `json:"state"`
			Commit string `json:"commit_id"`
			User   struct {
				ID int64 `json:"id"`
			} `json:"user"`
		}
		if _, err := adapter.rest(ctx, credential, http.MethodGet, adapter.repoPath()+"/pulls/"+desired.Attributes["pull_number"]+"/reviews", nil, &reviews); err != nil {
			return nil, err
		}
		for _, review := range reviews {
			if strings.Contains(review.Body, desired.Marker) && strconv.FormatInt(review.User.ID, 10) == desired.Attributes["reviewer_id"] {
				result = append(result, engine.SandboxObservedResource{Key: desired.Key, Kind: desired.Kind, Name: desired.Name, ID: strconv.FormatInt(review.ID, 10), Marker: desired.Marker, Attributes: desiredAttributes(desired, map[string]string{"pull_number": desired.Attributes["pull_number"], "reviewer_id": desired.Attributes["reviewer_id"], "state": review.State, "commit_id": review.Commit})})
			}
		}
	}
	return result, nil
}

func (adapter *SandboxAdapter) applyReview(ctx context.Context, credential Credential, effect engine.SandboxEffect) (engine.SandboxEffectResult, error) {
	if effect.Kind == "remove-resource" {
		return engine.SandboxEffectResult{Outcome: "not-applicable", Detail: "submitted GitHub reviews are retained evidence and cannot be deleted"}, nil
	}
	body := map[string]string{"body": effect.Resource.Marker, "event": effect.Resource.Attributes["input:event"], "commit_id": effect.Resource.Attributes["commit_id"]}
	var review struct {
		ID int64 `json:"id"`
	}
	if _, err := adapter.rest(ctx, credential, http.MethodPost, adapter.repoPath()+"/pulls/"+effect.Resource.Attributes["pull_number"]+"/reviews", body, &review); err != nil {
		return engine.SandboxEffectResult{}, err
	}
	return engine.SandboxEffectResult{Outcome: "applied", ResourceID: strconv.FormatInt(review.ID, 10), Detail: "distinct fixture review submitted"}, nil
}

func (adapter *SandboxAdapter) hasResourceKind(kind string) bool {
	for _, resource := range adapter.config.Resources {
		if resource.Kind == kind {
			return true
		}
	}
	return false
}

func (adapter *SandboxAdapter) hasAnyResourceKind(kinds ...string) bool {
	for _, kind := range kinds {
		if adapter.hasResourceKind(kind) {
			return true
		}
	}
	return false
}

func (adapter *SandboxAdapter) repoPath() string {
	return "/repos/" + url.PathEscape(adapter.config.RepositoryOwner) + "/" + url.PathEscape(adapter.config.RepositoryName)
}

func desiredAttributes(desired engine.SandboxResourceSpec, available map[string]string) map[string]string {
	result := make(map[string]string, len(desired.Attributes))
	for key := range desired.Attributes {
		if strings.HasPrefix(key, "input:") {
			continue
		}
		if value, exists := available[key]; exists {
			result[key] = value
		}
	}
	return result
}

func markerBody(resource engine.SandboxResourceSpec) string {
	body := resource.Attributes["input:body"]
	if strings.Contains(body, resource.Marker) {
		return body
	}
	if body == "" {
		return resource.Marker
	}
	return body + "\n\n" + resource.Marker
}

func escapePath(value string) string {
	segments := strings.Split(value, "/")
	for index := range segments {
		segments[index] = url.PathEscape(segments[index])
	}
	return strings.Join(segments, "/")
}

func isResponseStatus(err error, status int) bool {
	var failure *responseError
	return errors.As(err, &failure) && failure.StatusCode == status
}
