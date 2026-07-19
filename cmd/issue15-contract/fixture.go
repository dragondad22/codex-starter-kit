package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const fixtureTombstone = "Contract fixture retired. Immutable evidence is retained in the 30-day workflow artifact for issue #15."

type fixtureIssue struct {
	ID     int64  `json:"id"`
	NodeID string `json:"node_id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
}

type fixtureEvidence struct {
	SchemaVersion int                     `json:"schema_version"`
	Mandate       contractMandate         `json:"mandate"`
	Stage         string                  `json:"stage"`
	Marker        string                  `json:"marker"`
	Role          string                  `json:"role"`
	Actor         string                  `json:"actor"`
	Issues        map[string]fixtureIssue `json:"issues"`
	Relationships []string                `json:"relationships"`
	ProjectStates map[string]string       `json:"project_states,omitempty"`
	Disposition   string                  `json:"disposition"`
	Problems      []string                `json:"problems"`
}

type fixtureAPI struct {
	client     *http.Client
	token      string
	restBase   string
	graphQLURL string
}

func verifyContractLease(ctx context.Context, role string, mandate contractMandate, lease fixtureEvidence) error {
	configuration, err := roleConfiguration(role)
	if err != nil {
		return err
	}
	provider, err := appProvider(configuration.App)
	if err != nil {
		return err
	}
	credential, err := provider.Credential(ctx)
	if err != nil {
		return err
	}
	if credential.Actor != configuration.App.Actor || !exactPermissions(credential.Permissions, configuration.RequiredPermissions) {
		return errors.New("lease verification credential differs from the reviewed role contract")
	}
	api := &fixtureAPI{client: http.DefaultClient, token: credential.Token, restBase: restBaseURL, graphQLURL: graphQLURL}
	_, err = api.verifyLease(ctx, mandate, []fixtureEvidence{lease}, true)
	return err
}

func runFixtureStage(ctx context.Context, stage, role string, mandate contractMandate, lease ...fixtureEvidence) (fixtureEvidence, error) {
	configuration, err := roleConfiguration(role)
	if err != nil {
		return fixtureEvidence{}, err
	}
	provider, err := appProvider(configuration.App)
	if err != nil {
		return fixtureEvidence{}, err
	}
	credential, err := provider.Credential(ctx)
	if err != nil {
		return fixtureEvidence{}, err
	}
	if credential.Actor != configuration.App.Actor || !exactPermissions(credential.Permissions, configuration.RequiredPermissions) {
		return fixtureEvidence{}, errors.New("fixture credential differs from its reviewed role contract")
	}
	api := &fixtureAPI{client: http.DefaultClient, token: credential.Token, restBase: restBaseURL, graphQLURL: graphQLURL}
	issues := map[string]fixtureIssue{}
	disposition := "configured"
	problems := []string{}
	switch stage {
	case "setup":
		issues, err = api.findFixtures(ctx)
		if err == nil && len(issues) != 0 {
			err = errors.New("marked fixture state already exists without this execution lease")
		}
		if err != nil {
			return fixtureEvidence{}, err
		}
		issues, err = api.setup(ctx, issues)
		if err != nil {
			setupErr := err
			recoveryIssues, discoveryErr := api.findFixtures(ctx)
			if discoveryErr == nil {
				issues = recoveryIssues
			}
			cleanupErr := api.cleanupPartial(ctx, issues)
			if cleanupErr != nil {
				disposition = "needs-recovery"
				problems = []string{setupErr.Error(), cleanupErr.Error()}
			} else if discoveryErr != nil {
				disposition = "needs-recovery"
				problems = []string{setupErr.Error(), discoveryErr.Error()}
			} else {
				disposition = "recovered-non-pass"
				problems = []string{setupErr.Error()}
			}
			err = nil
		}
	case "project-setup":
		issues, err = api.verifyLease(ctx, mandate, lease, true)
		if err == nil {
			err = api.setProjectBaseline(ctx, issues)
			if err == nil {
				err = api.verifyProjectBaseline(ctx, issues)
			}
		}
	case "cleanup":
		issues, err = api.verifyLease(ctx, mandate, lease, false)
		if err == nil {
			err = api.cleanup(ctx, issues)
		}
	default:
		err = fmt.Errorf("unsupported fixture stage %q", stage)
	}
	if err != nil {
		return fixtureEvidence{}, err
	}
	evidence := fixtureEvidence{SchemaVersion: 1, Mandate: mandate, Stage: stage, Marker: runMarker, Role: role, Actor: credential.Actor, Issues: issues, Relationships: []string{}, Disposition: disposition, Problems: problems}
	if disposition == "configured" && (stage == "setup" || stage == "project-setup") {
		evidence.Relationships = []string{parentManagedID + "->" + selectedManagedID, parentManagedID + "->" + siblingManagedID, selectedManagedID + "->" + dependentManagedID, blockerManagedID + "->" + dependentManagedID}
	}
	if stage == "project-setup" {
		evidence.ProjectStates = baselineStates()
	}
	return evidence, nil
}

func (api *fixtureAPI) setup(ctx context.Context, existing map[string]fixtureIssue) (map[string]fixtureIssue, error) {
	desired := fixtureTasks()
	for _, managedID := range fixtureOrder() {
		task := desired[managedID]
		body, err := managedBody(task)
		if err != nil {
			return nil, err
		}
		issue, ok := existing[managedID]
		state := fixtureIssueStates()[managedID]
		payload := map[string]any{"title": task.Title, "body": body, "labels": []string{"type:task"}}
		if ok {
			payload["state"] = state
			if err := api.rest(ctx, http.MethodPatch, issuePath()+"/"+strconv.Itoa(issue.Number), payload, &issue); err != nil {
				return nil, err
			}
		} else {
			if err := api.rest(ctx, http.MethodPost, issuePath(), payload, &issue); err != nil {
				return nil, err
			}
			if state == "closed" {
				if err := api.rest(ctx, http.MethodPatch, issuePath()+"/"+strconv.Itoa(issue.Number), map[string]any{"state": state, "state_reason": "completed"}, &issue); err != nil {
					return nil, err
				}
			}
		}
		if issue.ID == 0 || issue.NodeID == "" || issue.Number == 0 || !strings.Contains(issue.Body, "starter-kit-managed:"+managedID) {
			return nil, fmt.Errorf("fixture %s lacks its immutable GitHub or managed identity", managedID)
		}
		existing[managedID] = issue
	}
	if err := api.ensureSubIssue(ctx, existing[parentManagedID], existing[selectedManagedID]); err != nil {
		return nil, err
	}
	if err := api.ensureSubIssue(ctx, existing[parentManagedID], existing[siblingManagedID]); err != nil {
		return nil, err
	}
	if err := api.ensureDependency(ctx, existing[dependentManagedID], existing[selectedManagedID]); err != nil {
		return nil, err
	}
	if err := api.ensureDependency(ctx, existing[dependentManagedID], existing[blockerManagedID]); err != nil {
		return nil, err
	}
	if err := api.verifySetupRelationships(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (api *fixtureAPI) verifySetupRelationships(ctx context.Context, issues map[string]fixtureIssue) error {
	var children []fixtureIssue
	if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(issues[parentManagedID].Number)+"/sub_issues?per_page=100", nil, &children); err != nil {
		return err
	}
	childIDs := map[int64]bool{}
	for _, child := range children {
		childIDs[child.ID] = true
	}
	var blockers []fixtureIssue
	if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(issues[dependentManagedID].Number)+"/dependencies/blocked_by?per_page=100", nil, &blockers); err != nil {
		return err
	}
	blockerIDs := map[int64]bool{}
	for _, blocker := range blockers {
		blockerIDs[blocker.ID] = true
	}
	if !childIDs[issues[selectedManagedID].ID] || !childIDs[issues[siblingManagedID].ID] || !blockerIDs[issues[selectedManagedID].ID] || !blockerIDs[issues[blockerManagedID].ID] {
		return errors.New("native fixture relationships did not converge")
	}
	return nil
}

func (api *fixtureAPI) cleanup(ctx context.Context, issues map[string]fixtureIssue) error {
	if err := requireFixtureSet(issues); err != nil {
		return err
	}
	return api.cleanupPartial(ctx, issues)
}

func (api *fixtureAPI) cleanupPartial(ctx context.Context, issues map[string]fixtureIssue) error {
	for _, relationship := range []struct{ parent, child string }{{parentManagedID, selectedManagedID}, {parentManagedID, siblingManagedID}} {
		if issues[relationship.parent].ID == 0 || issues[relationship.child].ID == 0 {
			continue
		}
		if err := api.removeSubIssue(ctx, issues[relationship.parent], issues[relationship.child]); err != nil {
			return err
		}
	}
	for _, blocker := range []string{selectedManagedID, blockerManagedID} {
		if issues[dependentManagedID].ID == 0 || issues[blocker].ID == 0 {
			continue
		}
		if err := api.removeDependency(ctx, issues[dependentManagedID], issues[blocker]); err != nil {
			return err
		}
	}
	for _, managedID := range fixtureOrder() {
		issue := issues[managedID]
		if issue.ID == 0 {
			continue
		}
		if err := api.rest(ctx, http.MethodPatch, issuePath()+"/"+strconv.Itoa(issue.Number), map[string]any{"state": "closed", "state_reason": "completed", "body": fixtureTombstone, "labels": []string{}}, &issue); err != nil {
			return err
		}
	}
	if len(issues) == len(fixtureOrder()) {
		return api.verifyCleanup(ctx, issues)
	}
	return nil
}

func (api *fixtureAPI) verifyLease(ctx context.Context, mandate contractMandate, leases []fixtureEvidence, requireComplete bool) (map[string]fixtureIssue, error) {
	if len(leases) != 1 || leases[0].SchemaVersion != 1 || leases[0].Stage != "setup" || leases[0].Mandate.Digest != mandate.Digest || leases[0].Marker != runMarker {
		return nil, errors.New("exact setup fixture lease is absent or outside the approved mandate")
	}
	if len(leases[0].Issues) == 0 {
		return nil, errors.New("setup fixture lease contains no immutable resources")
	}
	if requireComplete {
		if err := requireFixtureSet(leases[0].Issues); err != nil {
			return nil, err
		}
	}
	if len(leases[0].Issues) > len(fixtureOrder()) {
		return nil, errors.New("setup fixture lease exceeds the reviewed resource count")
	}
	allowed := map[string]bool{}
	for _, managedID := range fixtureOrder() {
		allowed[managedID] = true
	}
	for managedID := range leases[0].Issues {
		if !allowed[managedID] {
			return nil, errors.New("setup fixture lease contains an unreviewed managed identity")
		}
	}
	observed := map[string]fixtureIssue{}
	for managedID, expected := range leases[0].Issues {
		var issue fixtureIssue
		if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(expected.Number), nil, &issue); err != nil {
			return nil, err
		}
		identity, active := managedIDFromBody(issue.Body)
		active = active && identity == managedID && strings.Contains(issue.Body, "<!-- "+runMarker+" -->")
		retired := !requireComplete && issue.Body == fixtureTombstone && strings.EqualFold(issue.State, "closed")
		if (!active && !retired) || issue.ID != expected.ID || issue.NodeID != expected.NodeID {
			return nil, fmt.Errorf("fixture lease identity changed for %s", managedID)
		}
		observed[managedID] = issue
	}
	return observed, nil
}

func (api *fixtureAPI) findFixtures(ctx context.Context) (map[string]fixtureIssue, error) {
	issues := []fixtureIssue{}
	for page := 1; page <= 10; page++ {
		var observed []fixtureIssue
		path := issuePath() + "?state=all&per_page=100&page=" + strconv.Itoa(page)
		if err := api.rest(ctx, http.MethodGet, path, nil, &observed); err != nil {
			return nil, err
		}
		issues = append(issues, observed...)
		if len(observed) < 100 {
			break
		}
		if page == 10 {
			return nil, errors.New("fixture issue discovery exceeded its reviewed pagination bound")
		}
	}
	found := map[string]fixtureIssue{}
	for _, issue := range issues {
		if !strings.Contains(issue.Body, "<!-- "+runMarker+" -->") {
			continue
		}
		managedID, ok := managedIDFromBody(issue.Body)
		if !ok {
			return nil, fmt.Errorf("marked fixture #%d lacks one managed identity", issue.Number)
		}
		if _, duplicate := found[managedID]; duplicate {
			return nil, fmt.Errorf("managed fixture %s is ambiguous", managedID)
		}
		found[managedID] = issue
	}
	return found, nil
}

func (api *fixtureAPI) ensureSubIssue(ctx context.Context, parent, child fixtureIssue) error {
	var children []fixtureIssue
	if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(parent.Number)+"/sub_issues?per_page=100", nil, &children); err != nil {
		return err
	}
	for _, candidate := range children {
		if candidate.ID == child.ID {
			return nil
		}
	}
	return api.rest(ctx, http.MethodPost, issuePath()+"/"+strconv.Itoa(parent.Number)+"/sub_issues", map[string]any{"sub_issue_id": child.ID}, &fixtureIssue{})
}

func (api *fixtureAPI) removeSubIssue(ctx context.Context, parent, child fixtureIssue) error {
	var children []fixtureIssue
	if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(parent.Number)+"/sub_issues?per_page=100", nil, &children); err != nil {
		return err
	}
	for _, candidate := range children {
		if candidate.ID == child.ID {
			return api.rest(ctx, http.MethodDelete, issuePath()+"/"+strconv.Itoa(parent.Number)+"/sub_issue", map[string]any{"sub_issue_id": child.ID}, &fixtureIssue{})
		}
	}
	return nil
}

func (api *fixtureAPI) ensureDependency(ctx context.Context, dependent, blocker fixtureIssue) error {
	var blockers []fixtureIssue
	path := issuePath() + "/" + strconv.Itoa(dependent.Number) + "/dependencies/blocked_by"
	if err := api.rest(ctx, http.MethodGet, path+"?per_page=100", nil, &blockers); err != nil {
		return err
	}
	for _, candidate := range blockers {
		if candidate.ID == blocker.ID {
			return nil
		}
	}
	return api.rest(ctx, http.MethodPost, path, map[string]any{"issue_id": blocker.ID}, &fixtureIssue{})
}

func (api *fixtureAPI) removeDependency(ctx context.Context, dependent, blocker fixtureIssue) error {
	var blockers []fixtureIssue
	path := issuePath() + "/" + strconv.Itoa(dependent.Number) + "/dependencies/blocked_by"
	if err := api.rest(ctx, http.MethodGet, path+"?per_page=100", nil, &blockers); err != nil {
		return err
	}
	for _, candidate := range blockers {
		if candidate.ID == blocker.ID {
			return api.rest(ctx, http.MethodDelete, path+"/"+strconv.FormatInt(blocker.ID, 10), nil, nil)
		}
	}
	return nil
}

func (api *fixtureAPI) setProjectBaseline(ctx context.Context, issues map[string]fixtureIssue) error {
	for _, managedID := range fixtureOrder() {
		issue := issues[managedID]
		itemID, err := api.projectItem(ctx, issue.NodeID)
		if err != nil {
			return err
		}
		readiness, status := baselineOptions(managedID)
		if err := api.setProjectOption(ctx, itemID, fieldReadiness, readiness); err != nil {
			return err
		}
		if err := api.setProjectOption(ctx, itemID, fieldStatus, status); err != nil {
			return err
		}
	}
	return nil
}

func (api *fixtureAPI) verifyProjectBaseline(ctx context.Context, issues map[string]fixtureIssue) error {
	var observed struct {
		Data struct {
			Node struct {
				Items struct {
					Nodes []struct {
						Content struct {
							ID string `json:"id"`
						} `json:"content"`
						Values struct {
							Nodes []struct {
								OptionID string `json:"optionId"`
								Field    struct {
									ID string `json:"id"`
								} `json:"field"`
							} `json:"nodes"`
						} `json:"fieldValues"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"items"`
			} `json:"node"`
		} `json:"data"`
	}
	query := `query($project:ID!){node(id:$project){... on ProjectV2{items(first:100){nodes{content{... on Issue{id}} fieldValues(first:20){nodes{... on ProjectV2ItemFieldSingleSelectValue{optionId field{... on ProjectV2FieldCommon{id}}}}}} pageInfo{hasNextPage}}}}}`
	if err := api.graphql(ctx, query, map[string]any{"project": projectID}, &observed); err != nil {
		return err
	}
	if observed.Data.Node.Items.PageInfo.HasNextPage {
		return errors.New("fixture Project verification exceeded its reviewed bound")
	}
	states := map[string]map[string]string{}
	for _, item := range observed.Data.Node.Items.Nodes {
		fields := map[string]string{}
		for _, value := range item.Values.Nodes {
			fields[value.Field.ID] = value.OptionID
		}
		states[item.Content.ID] = fields
	}
	for managedID, issue := range issues {
		readiness, status := baselineOptions(managedID)
		if states[issue.NodeID][fieldReadiness] != readiness || states[issue.NodeID][fieldStatus] != status {
			return fmt.Errorf("fixture %s Project baseline did not converge", managedID)
		}
	}
	return nil
}

func (api *fixtureAPI) verifyCleanup(ctx context.Context, issues map[string]fixtureIssue) error {
	var children []fixtureIssue
	if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(issues[parentManagedID].Number)+"/sub_issues?per_page=100", nil, &children); err != nil {
		return err
	}
	for _, child := range children {
		if child.ID == issues[selectedManagedID].ID || child.ID == issues[siblingManagedID].ID {
			return errors.New("marked sub-issue relationship remains after cleanup")
		}
	}
	var blockers []fixtureIssue
	if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(issues[dependentManagedID].Number)+"/dependencies/blocked_by?per_page=100", nil, &blockers); err != nil {
		return err
	}
	for _, blocker := range blockers {
		if blocker.ID == issues[selectedManagedID].ID || blocker.ID == issues[blockerManagedID].ID {
			return errors.New("marked dependency relationship remains after cleanup")
		}
	}
	for managedID, expected := range issues {
		var issue fixtureIssue
		if err := api.rest(ctx, http.MethodGet, issuePath()+"/"+strconv.Itoa(expected.Number), nil, &issue); err != nil {
			return err
		}
		if !strings.EqualFold(issue.State, "closed") {
			return fmt.Errorf("fixture %s remains open after cleanup", managedID)
		}
		if strings.Contains(issue.Body, runMarker) || strings.Contains(issue.Body, "starter-kit-managed:") {
			return fmt.Errorf("fixture %s retains active ownership markers after cleanup", managedID)
		}
	}
	return nil
}

func (api *fixtureAPI) projectItem(ctx context.Context, issueNodeID string) (string, error) {
	var observed struct {
		Data struct {
			Node struct {
				Items struct {
					Nodes []struct {
						ID      string `json:"id"`
						Content struct {
							ID string `json:"id"`
						} `json:"content"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"items"`
			} `json:"node"`
		} `json:"data"`
	}
	query := `query($project:ID!){node(id:$project){... on ProjectV2{items(first:100){nodes{id content{... on Issue{id}}} pageInfo{hasNextPage}}}}}`
	if err := api.graphql(ctx, query, map[string]any{"project": projectID}, &observed); err != nil {
		return "", err
	}
	if observed.Data.Node.Items.PageInfo.HasNextPage {
		return "", errors.New("fixture Project item lookup exceeded its reviewed bound")
	}
	for _, item := range observed.Data.Node.Items.Nodes {
		if item.Content.ID == issueNodeID {
			return item.ID, nil
		}
	}
	var added struct {
		Data struct {
			Add struct {
				Item struct {
					ID string `json:"id"`
				} `json:"item"`
			} `json:"addProjectV2ItemById"`
		} `json:"data"`
	}
	mutation := `mutation($project:ID!,$content:ID!){addProjectV2ItemById(input:{projectId:$project,contentId:$content}){item{id}}}`
	if err := api.graphql(ctx, mutation, map[string]any{"project": projectID, "content": issueNodeID}, &added); err != nil {
		return "", err
	}
	if added.Data.Add.Item.ID == "" {
		return "", errors.New("fixture Project add returned no immutable item identity")
	}
	return added.Data.Add.Item.ID, nil
}

func (api *fixtureAPI) setProjectOption(ctx context.Context, itemID, fieldID, optionID string) error {
	var result struct {
		Data struct {
			Update struct {
				Item struct {
					ID string `json:"id"`
				} `json:"projectV2Item"`
			} `json:"updateProjectV2ItemFieldValue"`
		} `json:"data"`
	}
	mutation := `mutation($project:ID!,$item:ID!,$field:ID!,$option:String!){updateProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field,value:{singleSelectOptionId:$option}}){projectV2Item{id}}}`
	if err := api.graphql(ctx, mutation, map[string]any{"project": projectID, "item": itemID, "field": fieldID, "option": optionID}, &result); err != nil {
		return err
	}
	if result.Data.Update.Item.ID != itemID {
		return errors.New("fixture Project field mutation lacked its exact postcondition identity")
	}
	return nil
}

func (api *fixtureAPI) rest(ctx context.Context, method, path string, body, output any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(api.restBase, "/")+path, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+api.token)
	request.Header.Set("X-GitHub-Api-Version", apiVersion)
	request.Header.Set("User-Agent", "codex-starter-kit")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := api.client.Do(request)
	if err != nil {
		return errors.New("fixture GitHub REST endpoint is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("fixture GitHub REST request %s %s returned %d", method, path, response.StatusCode)
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(output); err != nil {
		return errors.New("fixture GitHub REST response is invalid")
	}
	return nil
}

func (api *fixtureAPI) graphql(ctx context.Context, query string, variables map[string]any, output any) error {
	var raw struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	payload, _ := json.Marshal(map[string]any{"query": query, "variables": variables})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, api.graphQLURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+api.token)
	request.Header.Set("X-GitHub-Api-Version", apiVersion)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "codex-starter-kit")
	response, err := api.client.Do(request)
	if err != nil {
		return errors.New("fixture GitHub GraphQL endpoint is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&raw) != nil || len(raw.Errors) != 0 || len(raw.Data) == 0 {
		return errors.New("fixture GitHub GraphQL request returned non-pass evidence")
	}
	// raw.Data contains the value of the top-level data property; callers expect the full envelope.
	envelope, _ := json.Marshal(map[string]json.RawMessage{"data": raw.Data})
	return json.Unmarshal(envelope, output)
}

func fixtureTasks() map[string]engine.DesiredManagedTask {
	tasks := map[string]engine.DesiredManagedTask{selectedManagedID: contractTask()}
	for managedID, title := range map[string]string{
		parentManagedID: "Contract fixture: parent", siblingManagedID: "Contract fixture: sibling",
		blockerManagedID: "Contract fixture: blocker", dependentManagedID: "Contract fixture: dependent",
	} {
		tasks[managedID] = engine.DesiredManagedTask{ManagedID: managedID, IssueType: "task", Title: title, Blockers: []engine.WorkDependency{}, Readiness: "ready", Status: "backlog", Review: []engine.WorkReviewRequirement{{Role: "independent-reviewer", DistinctContext: true}}}
	}
	return tasks
}

func fixtureIssueStates() map[string]string {
	return map[string]string{parentManagedID: "open", selectedManagedID: "closed", siblingManagedID: "open", blockerManagedID: "closed", dependentManagedID: "open"}
}

func fixtureOrder() []string {
	return []string{parentManagedID, selectedManagedID, siblingManagedID, blockerManagedID, dependentManagedID}
}

func baselineOptions(managedID string) (string, string) {
	switch managedID {
	case selectedManagedID:
		return readinessReady, statusNext
	case dependentManagedID:
		return readinessBlocked, statusBacklog
	case blockerManagedID:
		return readinessReady, statusDone
	default:
		return readinessReady, statusBacklog
	}
}

func baselineStates() map[string]string {
	result := map[string]string{}
	for _, managedID := range fixtureOrder() {
		readiness, status := baselineOptions(managedID)
		result[managedID] = readiness + ":" + status
	}
	return result
}

func requireFixtureSet(issues map[string]fixtureIssue) error {
	missing := []string{}
	if len(issues) != len(fixtureOrder()) {
		return errors.New("fixture lease does not contain the exact reviewed resource count")
	}
	for _, managedID := range fixtureOrder() {
		if issues[managedID].ID == 0 {
			missing = append(missing, managedID)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return fmt.Errorf("marked fixture set is incomplete: %s", strings.Join(missing, ", "))
	}
	return nil
}

func managedIDFromBody(body string) (string, bool) {
	const prefix = "<!-- starter-kit-managed:"
	start := strings.Index(body, prefix)
	if start < 0 {
		return "", false
	}
	start += len(prefix)
	end := strings.Index(body[start:], " -->")
	if end < 0 {
		return "", false
	}
	value := strings.TrimSpace(body[start : start+end])
	return value, value != "" && !strings.ContainsAny(value, "\r\n<>")
}

func issuePath() string {
	return "/repos/" + url.PathEscape(repositoryOwner) + "/" + url.PathEscape(repositoryName) + "/issues"
}

var _ githubadapter.CredentialProvider = (*githubadapter.AppInstallationProvider)(nil)
