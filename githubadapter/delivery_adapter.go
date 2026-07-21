package githubadapter

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/dragondad22/codex-starter-kit/engine"
)

type DeliveryReviewerTrust struct {
	Actor           string
	Capable         bool
	DistinctContext bool
}

type DeliveryAdapter struct {
	base      *Adapter
	reviewers map[string]DeliveryReviewerTrust
}

func (adapter *Adapter) repoPath() string {
	return "/repos/" + url.PathEscape(adapter.config.RepositoryOwner) + "/" + url.PathEscape(adapter.config.RepositoryName)
}

func NewDeliveryAdapter(base *Adapter, reviewers []DeliveryReviewerTrust) (*DeliveryAdapter, error) {
	if base == nil {
		return nil, errors.New("delivery adapter requires a GitHub transport")
	}
	trusted := map[string]DeliveryReviewerTrust{}
	for _, reviewer := range reviewers {
		if reviewer.Actor == "" || trusted[reviewer.Actor].Actor != "" {
			return nil, errors.New("delivery reviewer trust is invalid or duplicated")
		}
		trusted[reviewer.Actor] = reviewer
	}
	return &DeliveryAdapter{base: base, reviewers: trusted}, nil
}

func (adapter *DeliveryAdapter) Capability(ctx context.Context) (engine.DeliveryCapability, error) {
	credential, err := adapter.base.credential(ctx)
	if err != nil {
		return engine.DeliveryCapability{}, err
	}
	now := adapter.base.now()
	return engine.DeliveryCapability{
		SchemaVersion: 1, Online: true, Fresh: now.Before(credential.ExpiresAt), Actor: credential.Actor, Mode: credential.Mode,
		Permissions: slices.Clone(credential.Permissions), ObservedAt: now, ExpiresAt: credential.ExpiresAt,
	}, nil
}

func (adapter *DeliveryAdapter) ObserveDelivery(ctx context.Context, intent engine.DeliveryIntent) (engine.DeliveryObservation, error) {
	credential, err := adapter.base.credential(ctx)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	if intent.ManagedID == "" || intent.Target.Host != adapter.base.config.Host || intent.Target.RepositoryID != adapter.base.config.RepositoryID || intent.Target.ProjectID != adapter.base.config.ProjectID {
		return engine.DeliveryObservation{}, errors.New("delivery target is outside the allowlisted GitHub manifest")
	}
	issues, err := adapter.base.findManagedIssues(ctx, credential, intent.ManagedID)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation := engine.DeliveryObservation{SchemaVersion: 1, Problems: []string{}, Checks: []engine.DeliveryCheckObservation{}, Reviews: []engine.DeliveryReviewObservation{}}
	if len(issues) != 1 || issues[0].PullRequest != nil {
		observation.Problems = append(observation.Problems, "managed delivery issue identity is missing or ambiguous")
		observation.Revision = digest(observation)
		return observation, nil
	}
	issue := issues[0]
	observation.Issue = engine.DeliveryIssueObservation{ManagedID: intent.ManagedID, State: strings.ToLower(issue.State)}
	pull, err := adapter.findLinkedDeliveryPull(ctx, credential, issue, intent)
	if err != nil {
		observation.Problems = append(observation.Problems, err.Error())
		observation.Revision = digest(observation)
		return observation, nil
	}
	observation.PullRequest = engine.DeliveryPullRequestObservation{
		Number: pull.Number, State: strings.ToLower(pull.State), Draft: pull.Draft, Base: pull.Base.Ref, Head: pull.Head.Ref, HeadRevision: pull.Head.SHA,
		Merged: pull.Merged, MergeRevision: pull.MergeCommitSHA,
	}
	if intent.Claim == nil || !deliveryClaimMatches(pull.Body, *intent.Claim) {
		observation.Problems = append(observation.Problems, "pull request delivery claim does not match governed intent")
	}
	checks, err := adapter.observeDeliveryChecks(ctx, credential, pull.Head.SHA)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation.Checks = checks
	reviews, err := adapter.observeDeliveryReviews(ctx, credential, pull.Number)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation.Reviews = reviews
	rules, err := adapter.observeDeliveryRules(ctx, credential, intent.BaseBranch)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation.Rules = rules
	observation.Revision = digest(observation)
	return observation, nil
}

type deliveryPull struct {
	Number         int    `json:"number"`
	NodeID         string `json:"node_id"`
	State          string `json:"state"`
	Draft          bool   `json:"draft"`
	Body           string `json:"body"`
	Merged         bool   `json:"merged"`
	MergeCommitSHA string `json:"merge_commit_sha"`
	Head           struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref  string `json:"ref"`
		Repo struct {
			NodeID string `json:"node_id"`
		} `json:"repo"`
	} `json:"base"`
}

func (adapter *DeliveryAdapter) findLinkedDeliveryPull(ctx context.Context, credential Credential, issue githubIssue, intent engine.DeliveryIntent) (deliveryPull, error) {
	next := adapter.base.issuePath() + "/" + strconv.Itoa(issue.Number) + "/timeline?per_page=100"
	numbers := []int{}
	for page := 0; page < adapter.base.config.MaxPages && next != ""; page++ {
		events := []githubTimelineEvent{}
		response, err := adapter.base.rest(ctx, credential, http.MethodGet, next, nil, &events)
		if err != nil {
			return deliveryPull{}, err
		}
		for _, event := range events {
			if event.Event == "cross-referenced" && event.Source.Issue.PullRequest != nil && adapter.base.sameRepositoryURL(event.Source.Issue.RepositoryURL) {
				numbers = append(numbers, event.Source.Issue.Number)
			}
		}
		next, err = adapter.base.nextRESTPath(response.Header.Get("Link"))
		if err != nil {
			return deliveryPull{}, err
		}
	}
	slices.Sort(numbers)
	numbers = slices.Compact(numbers)
	matches := []deliveryPull{}
	for _, number := range numbers {
		var pull deliveryPull
		path := adapter.base.repoPath() + "/pulls/" + strconv.Itoa(number)
		if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path, nil, &pull); err != nil {
			return deliveryPull{}, err
		}
		if pull.Head.Ref == intent.HeadBranch && pull.Base.Ref == intent.BaseBranch && pull.Base.Repo.NodeID == adapter.base.config.RepositoryID {
			matches = append(matches, pull)
		}
	}
	if len(matches) != 1 {
		return deliveryPull{}, errors.New("issue-linked delivery pull request is missing or ambiguous")
	}
	return matches[0], nil
}

func (adapter *DeliveryAdapter) observeDeliveryChecks(ctx context.Context, credential Credential, head string) ([]engine.DeliveryCheckObservation, error) {
	path := adapter.base.repoPath() + "/commits/" + url.PathEscape(head)
	var runs struct {
		CheckRuns []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
			HeadSHA    string `json:"head_sha"`
		} `json:"check_runs"`
	}
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path+"/check-runs?per_page=100", nil, &runs); err != nil {
		return nil, err
	}
	checks := []engine.DeliveryCheckObservation{}
	for _, run := range runs.CheckRuns {
		state := "pending"
		if run.Status == "completed" && run.Conclusion == "success" {
			state = "passed"
		} else if run.Status == "completed" {
			state = "failed"
		}
		checks = append(checks, engine.DeliveryCheckObservation{Name: run.Name, HeadRevision: run.HeadSHA, State: state})
	}
	var statuses struct {
		Statuses []struct {
			Context string `json:"context"`
			State   string `json:"state"`
			SHA     string `json:"sha"`
		} `json:"statuses"`
	}
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path+"/status", nil, &statuses); err != nil {
		return nil, err
	}
	for _, status := range statuses.Statuses {
		state := "pending"
		if status.State == "success" {
			state = "passed"
		} else if slices.Contains([]string{"failure", "error"}, status.State) {
			state = "failed"
		}
		checks = append(checks, engine.DeliveryCheckObservation{Name: status.Context, HeadRevision: status.SHA, State: state})
	}
	return checks, nil
}

func (adapter *DeliveryAdapter) observeDeliveryReviews(ctx context.Context, credential Credential, number int) ([]engine.DeliveryReviewObservation, error) {
	var reviews []struct {
		State    string `json:"state"`
		CommitID string `json:"commit_id"`
		User     struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	path := adapter.base.repoPath() + "/pulls/" + strconv.Itoa(number) + "/reviews?per_page=100"
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path, nil, &reviews); err != nil {
		return nil, err
	}
	result := make([]engine.DeliveryReviewObservation, 0, len(reviews))
	for _, review := range reviews {
		trust := adapter.reviewers[review.User.Login]
		state := strings.ToLower(strings.ReplaceAll(review.State, "_", "-"))
		result = append(result, engine.DeliveryReviewObservation{Actor: review.User.Login, HeadRevision: review.CommitID, State: state, DistinctContext: trust.DistinctContext, Capable: trust.Capable})
	}
	return result, nil
}

func (adapter *DeliveryAdapter) observeDeliveryRules(ctx context.Context, credential Credential, branch string) (engine.DeliveryRulesObservation, error) {
	var rules []struct {
		Type       string `json:"type"`
		Parameters struct {
			Required []struct {
				Context string `json:"context"`
			} `json:"required_status_checks"`
		} `json:"parameters"`
	}
	path := adapter.base.repoPath() + "/rules/branches/" + escapePath(branch)
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path, nil, &rules); err != nil {
		return engine.DeliveryRulesObservation{}, err
	}
	required := []string{}
	for _, rule := range rules {
		if rule.Type == "required_status_checks" {
			for _, check := range rule.Parameters.Required {
				required = append(required, check.Context)
			}
		}
	}
	slices.Sort(required)
	required = slices.Compact(required)
	var repository struct {
		NodeID           string `json:"node_id"`
		DefaultBranch    string `json:"default_branch"`
		AllowSquashMerge bool   `json:"allow_squash_merge"`
	}
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, adapter.base.repoPath(), nil, &repository); err != nil {
		return engine.DeliveryRulesObservation{}, err
	}
	methods := []string{}
	if repository.NodeID == adapter.base.config.RepositoryID && repository.DefaultBranch == branch && repository.AllowSquashMerge {
		methods = append(methods, "squash")
	}
	return engine.DeliveryRulesObservation{Revision: digest(struct {
		Rules      any
		Repository any
	}{rules, repository}), RequiredChecks: required, MergeMethods: methods}, nil
}

func deliveryClaimMatches(body string, expected engine.WorkDeliveryClaim) bool {
	observed, err := engine.ParseWorkDeliveryClaim(body)
	if err != nil {
		return false
	}
	left, leftErr := engine.RenderWorkDeliveryClaim(observed)
	right, rightErr := engine.RenderWorkDeliveryClaim(expected)
	return leftErr == nil && rightErr == nil && left == right
}

func (adapter *DeliveryAdapter) ApplyDelivery(ctx context.Context, effect engine.DeliveryEffect) (engine.DeliveryEffectResult, error) {
	credential, err := adapter.base.credential(ctx)
	if err != nil {
		return engine.DeliveryEffectResult{Outcome: "denied", Detail: "delivery credential is unavailable", Recoverable: true}, err
	}
	path := adapter.base.repoPath() + "/pulls/" + strconv.Itoa(effect.PullRequest)
	var pull deliveryPull
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path, nil, &pull); err != nil {
		return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery pull request cannot be re-observed", Recoverable: true}, err
	}
	if pull.Head.SHA != effect.HeadRevision {
		return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery head changed before effect", Recoverable: true}, nil
	}
	switch effect.Kind {
	case engine.DeliveryEffectMarkReady:
		var response struct {
			Errors []graphQLError `json:"errors"`
		}
		query := `mutation($id:ID!){markPullRequestReadyForReview(input:{pullRequestId:$id}){pullRequest{id}}}`
		if err := adapter.base.mutateGraphQL(ctx, credential, query, map[string]any{"id": pull.NodeID}, &response); err != nil || len(response.Errors) != 0 {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "mark-ready result requires re-observation", Recoverable: true}, err
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "marked exact pull request ready"}, nil
	case engine.DeliveryEffectSquashMerge:
		body := map[string]any{"merge_method": "squash", "sha": effect.HeadRevision}
		var merged struct {
			Merged  bool   `json:"merged"`
			Message string `json:"message"`
		}
		if _, err := adapter.base.rest(ctx, credential, http.MethodPut, path+"/merge", body, &merged); err != nil {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "squash-merge result requires exact re-observation", Recoverable: true}, err
		}
		if !merged.Merged {
			return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "GitHub did not confirm the squash merge", Recoverable: true}, nil
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "squash merged exact pull request head"}, nil
	default:
		return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "unsupported delivery effect", Recoverable: true}, errors.New("unsupported delivery effect")
	}
}
