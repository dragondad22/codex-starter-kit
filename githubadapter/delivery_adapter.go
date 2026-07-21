package githubadapter

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

type DeliveryReviewerTrust struct {
	Actor                string
	Capable              bool
	DistinctContext      bool
	QualifiedIndependent bool
	ProductApprover      bool
}

type DeliveryAdapter struct {
	base       *Adapter
	effectBase *Adapter
	reviewers  map[string]DeliveryReviewerTrust
}

func (adapter *Adapter) repoPath() string {
	return "/repos/" + url.PathEscape(adapter.config.RepositoryOwner) + "/" + url.PathEscape(adapter.config.RepositoryName)
}

func NewDeliveryAdapter(base *Adapter, reviewers []DeliveryReviewerTrust, effectAdapters ...*Adapter) (*DeliveryAdapter, error) {
	if base == nil {
		return nil, errors.New("delivery adapter requires a GitHub transport")
	}
	if len(effectAdapters) > 1 || len(effectAdapters) == 1 && effectAdapters[0] == nil {
		return nil, errors.New("delivery adapter accepts at most one effect transport")
	}
	effectBase := base
	if len(effectAdapters) == 1 {
		effectBase = effectAdapters[0]
		if effectBase.config.Host != base.config.Host || effectBase.config.RESTBaseURL != base.config.RESTBaseURL || effectBase.config.RepositoryOwner != base.config.RepositoryOwner || effectBase.config.RepositoryName != base.config.RepositoryName || effectBase.config.RepositoryID != base.config.RepositoryID {
			return nil, errors.New("delivery observation and effect transports must bind the same repository")
		}
	}
	trusted := map[string]DeliveryReviewerTrust{}
	for _, reviewer := range reviewers {
		if reviewer.Actor == "" || trusted[reviewer.Actor].Actor != "" {
			return nil, errors.New("delivery reviewer trust is invalid or duplicated")
		}
		trusted[reviewer.Actor] = reviewer
	}
	return &DeliveryAdapter{base: base, effectBase: effectBase, reviewers: trusted}, nil
}

func (adapter *DeliveryAdapter) Capability(ctx context.Context) (engine.DeliveryCapability, error) {
	credential, err := adapter.effectBase.credential(ctx)
	if err != nil {
		return engine.DeliveryCapability{}, err
	}
	now := adapter.effectBase.now()
	return engine.DeliveryCapability{
		SchemaVersion: 1, Online: true, Fresh: now.Before(credential.ExpiresAt), Actor: credential.Actor, Mode: credential.Mode,
		Account: credential.Account, InstallationID: credential.InstallationID, RepositoryID: adapter.effectBase.config.RepositoryID,
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
	observation := engine.DeliveryObservation{SchemaVersion: 1, Problems: []string{}, Checks: []engine.DeliveryCheckObservation{}, Reviews: []engine.DeliveryReviewObservation{}, Approvals: []engine.DeliveryApprovalObservation{}}
	if len(issues) != 1 || issues[0].PullRequest != nil {
		observation.Problems = append(observation.Problems, "managed delivery issue identity is missing or ambiguous")
		observation.Revision = digest(observation)
		return observation, nil
	}
	issue := issues[0]
	observation.Issue = engine.DeliveryIssueObservation{ManagedID: intent.ManagedID, Number: issue.Number, State: strings.ToLower(issue.State)}
	rules, err := adapter.observeDeliveryRules(ctx, credential, intent.BaseBranch)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation.Rules = rules
	var branch struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	branchPath := adapter.base.repoPath() + "/git/ref/heads/" + escapePath(intent.HeadBranch)
	branchMissing := false
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, branchPath, nil, &branch); err != nil {
		if isResponseStatus(err, http.StatusNotFound) {
			branchMissing = true
		} else {
			return engine.DeliveryObservation{}, err
		}
	}
	if !branchMissing {
		observation.Branch = engine.DeliveryBranchObservation{Name: intent.HeadBranch, Revision: branch.Object.SHA, Present: true}
	}
	pull, err := adapter.findLinkedDeliveryPull(ctx, credential, issue, intent)
	if err != nil {
		observation.Problems = append(observation.Problems, err.Error())
		observation.Revision = digest(observation)
		return observation, nil
	}
	if pull.Number == 0 {
		observation.Revision = digest(observation)
		return observation, nil
	}
	if branchMissing {
		observation.Branch = engine.DeliveryBranchObservation{Name: pull.Head.Ref, Revision: pull.Head.SHA, Historical: true}
	}
	observation.PullRequest = engine.DeliveryPullRequestObservation{
		Number: pull.Number, State: strings.ToLower(pull.State), Draft: pull.Draft, Base: pull.Base.Ref, Head: pull.Head.Ref, HeadRevision: pull.Head.SHA,
		Merged: pull.Merged, MergeRevision: pull.MergeCommitSHA, RequestedReviewers: pull.requestedReviewerLogins(),
	}
	if !branchMissing && pull.Head.SHA != branch.Object.SHA {
		observation.Problems = append(observation.Problems, "pull request head does not match the observed delivery branch")
	}
	if intent.Claim == nil || !deliveryClaimMatches(pull.Body, *intent.Claim) {
		observation.Problems = append(observation.Problems, "pull request delivery claim does not match governed intent")
	} else if pull.Merged && pull.MergedAt != nil {
		current := githubPullRequest{Number: pull.Number, Body: pull.Body, Merged: pull.Merged, MergedAt: pull.MergedAt, MergeCommitSHA: pull.MergeCommitSHA}
		current.Base.Ref = pull.Base.Ref
		current.Base.Repository.NodeID = pull.Base.Repo.NodeID
		reachable, _, verifyErr := adapter.base.verifyCurrentDelivery(ctx, credential, current, *intent.Claim)
		if verifyErr != nil {
			return engine.DeliveryObservation{}, verifyErr
		}
		observation.PullRequest.DefaultReachable = reachable
	}
	checks, err := adapter.observeDeliveryChecks(ctx, credential, pull.Head.SHA)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation.Checks = checks
	reviews, approvals, err := adapter.observeDeliveryReviews(ctx, credential, pull.Number)
	if err != nil {
		return engine.DeliveryObservation{}, err
	}
	observation.Reviews = reviews
	observation.Approvals = approvals
	observation.Revision = digest(observation)
	return observation, nil
}

type deliveryPull struct {
	Number             int        `json:"number"`
	NodeID             string     `json:"node_id"`
	State              string     `json:"state"`
	Draft              bool       `json:"draft"`
	Body               string     `json:"body"`
	Merged             bool       `json:"merged"`
	MergeCommitSHA     string     `json:"merge_commit_sha"`
	MergedAt           *time.Time `json:"merged_at"`
	RequestedReviewers []struct {
		Login string `json:"login"`
	} `json:"requested_reviewers"`
	Head struct {
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

func (pull deliveryPull) requestedReviewerLogins() []string {
	logins := make([]string, 0, len(pull.RequestedReviewers))
	for _, reviewer := range pull.RequestedReviewers {
		if reviewer.Login != "" {
			logins = append(logins, reviewer.Login)
		}
	}
	slices.Sort(logins)
	return slices.Compact(logins)
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
	if next != "" {
		return deliveryPull{}, errors.New("GitHub delivery timeline pagination exceeded the configured bound")
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
	if len(matches) > 1 {
		return deliveryPull{}, errors.New("issue-linked delivery pull request is ambiguous")
	}
	if len(matches) == 0 {
		return deliveryPull{}, nil
	}
	return matches[0], nil
}

func (adapter *DeliveryAdapter) observeDeliveryChecks(ctx context.Context, credential Credential, head string) ([]engine.DeliveryCheckObservation, error) {
	path := adapter.base.repoPath() + "/commits/" + url.PathEscape(head)
	type checkRunsPage struct {
		CheckRuns []struct {
			ID          int64     `json:"id"`
			Name        string    `json:"name"`
			Status      string    `json:"status"`
			Conclusion  string    `json:"conclusion"`
			HeadSHA     string    `json:"head_sha"`
			CompletedAt time.Time `json:"completed_at"`
		} `json:"check_runs"`
	}
	checks := []engine.DeliveryCheckObservation{}
	next := path + "/check-runs?per_page=100"
	for page := 0; page < adapter.base.config.MaxPages && next != ""; page++ {
		var runs checkRunsPage
		response, err := adapter.base.rest(ctx, credential, http.MethodGet, next, nil, &runs)
		if err != nil {
			return nil, err
		}
		for _, run := range runs.CheckRuns {
			state := "pending"
			if run.Status == "completed" && run.Conclusion == "success" {
				state = "passed"
			} else if run.Status == "completed" {
				state = "failed"
			}
			checks = append(checks, engine.DeliveryCheckObservation{Name: run.Name, HeadRevision: run.HeadSHA, State: state, EvidenceID: "check-run:" + strconv.FormatInt(run.ID, 10), ObservedAt: run.CompletedAt})
		}
		next, err = adapter.base.nextRESTPath(response.Header.Get("Link"))
		if err != nil {
			return nil, err
		}
	}
	if next != "" {
		return nil, errors.New("GitHub delivery check pagination exceeded the configured bound")
	}
	var statuses struct {
		Statuses []struct {
			ID        int64     `json:"id"`
			Context   string    `json:"context"`
			State     string    `json:"state"`
			SHA       string    `json:"sha"`
			UpdatedAt time.Time `json:"updated_at"`
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
		checks = append(checks, engine.DeliveryCheckObservation{Name: status.Context, HeadRevision: status.SHA, State: state, EvidenceID: "status:" + strconv.FormatInt(status.ID, 10), ObservedAt: status.UpdatedAt})
	}
	return checks, nil
}

func (adapter *DeliveryAdapter) observeDeliveryReviews(ctx context.Context, credential Credential, number int) ([]engine.DeliveryReviewObservation, []engine.DeliveryApprovalObservation, error) {
	var reviews []struct {
		ID          int64     `json:"id"`
		State       string    `json:"state"`
		CommitID    string    `json:"commit_id"`
		SubmittedAt time.Time `json:"submitted_at"`
		User        struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	next := adapter.base.repoPath() + "/pulls/" + strconv.Itoa(number) + "/reviews?per_page=100"
	allReviews := []struct {
		ID          int64     `json:"id"`
		State       string    `json:"state"`
		CommitID    string    `json:"commit_id"`
		SubmittedAt time.Time `json:"submitted_at"`
		User        struct {
			Login string `json:"login"`
		} `json:"user"`
	}{}
	for page := 0; page < adapter.base.config.MaxPages && next != ""; page++ {
		reviews = nil
		response, err := adapter.base.rest(ctx, credential, http.MethodGet, next, nil, &reviews)
		if err != nil {
			return nil, nil, err
		}
		allReviews = append(allReviews, reviews...)
		next, err = adapter.base.nextRESTPath(response.Header.Get("Link"))
		if err != nil {
			return nil, nil, err
		}
	}
	if next != "" {
		return nil, nil, errors.New("GitHub delivery review pagination exceeded the configured bound")
	}
	result := make([]engine.DeliveryReviewObservation, 0, len(allReviews))
	approvals := []engine.DeliveryApprovalObservation{}
	for _, review := range allReviews {
		trust := adapter.reviewers[review.User.Login]
		state := strings.ToLower(strings.ReplaceAll(review.State, "_", "-"))
		evidenceID := "review:" + strconv.FormatInt(review.ID, 10)
		result = append(result, engine.DeliveryReviewObservation{Actor: review.User.Login, HeadRevision: review.CommitID, State: state, DistinctContext: trust.DistinctContext, Capable: trust.Capable, QualifiedIndependent: trust.QualifiedIndependent, EvidenceID: evidenceID, ObservedAt: review.SubmittedAt})
		if trust.ProductApprover {
			approvals = append(approvals, engine.DeliveryApprovalObservation{Actor: review.User.Login, HeadRevision: review.CommitID, State: state, DistinctContext: trust.DistinctContext, Capable: trust.Capable, QualifiedIndependent: trust.QualifiedIndependent, EvidenceID: evidenceID, ObservedAt: review.SubmittedAt})
		}
	}
	return result, approvals, nil
}

func (adapter *DeliveryAdapter) observeDeliveryRules(ctx context.Context, credential Credential, branch string) (engine.DeliveryRulesObservation, error) {
	var rules []struct {
		Type       string `json:"type"`
		Parameters struct {
			Required []struct {
				Context string `json:"context"`
			} `json:"required_status_checks"`
			RequiredApprovals             int  `json:"required_approving_review_count"`
			RequireCodeOwner              bool `json:"require_code_owner_review"`
			RequireConversationResolution bool `json:"required_review_thread_resolution"`
		} `json:"parameters"`
	}
	path := adapter.base.repoPath() + "/rules/branches/" + escapePath(branch)
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, path, nil, &rules); err != nil {
		return engine.DeliveryRulesObservation{}, err
	}
	required := []string{}
	problems := []string{}
	for _, rule := range rules {
		switch rule.Type {
		case "required_status_checks":
			for _, check := range rule.Parameters.Required {
				required = append(required, check.Context)
			}
		case "pull_request":
			if rule.Parameters.RequiredApprovals > 1 || rule.Parameters.RequireCodeOwner || rule.Parameters.RequireConversationResolution {
				problems = append(problems, "effective pull-request rules require unsupported stronger approval evidence")
			}
		case "merge_queue", "required_deployments", "required_code_scanning":
			problems = append(problems, "effective branch rules include unsupported merge gate: "+rule.Type)
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
	var base struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if _, err := adapter.base.rest(ctx, credential, http.MethodGet, adapter.base.repoPath()+"/branches/"+url.PathEscape(branch), nil, &base); err != nil {
		return engine.DeliveryRulesObservation{}, err
	}
	return engine.DeliveryRulesObservation{Revision: digest(struct {
		Rules      any
		Repository any
		Base       any
	}{rules, repository, base}), BaseRevision: base.Commit.SHA, RequiredChecks: required, MergeMethods: methods, Problems: problems}, nil
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
	credential, err := adapter.effectBase.credential(ctx)
	if err != nil {
		return engine.DeliveryEffectResult{Outcome: "denied", Detail: "delivery credential is unavailable", Recoverable: true}, err
	}
	if effect.Kind == engine.DeliveryEffectCreateBranch {
		body := map[string]string{"ref": "refs/heads/" + effect.Branch, "sha": effect.HeadRevision}
		if _, err := adapter.effectBase.rest(ctx, credential, http.MethodPost, adapter.effectBase.repoPath()+"/git/refs", body, nil); err != nil {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "create-branch result requires exact re-observation", Recoverable: true}, err
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "created exact delivery branch"}, nil
	}
	if effect.Kind == engine.DeliveryEffectCreatePullRequest {
		if effect.Claim == nil {
			return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery claim is missing", Recoverable: true}, errors.New("delivery claim is missing")
		}
		marker, err := engine.RenderWorkDeliveryClaim(*effect.Claim)
		if err != nil {
			return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery claim is invalid", Recoverable: true}, err
		}
		var branch struct {
			Object struct {
				SHA string `json:"sha"`
			} `json:"object"`
		}
		branchPath := adapter.effectBase.repoPath() + "/git/ref/heads/" + escapePath(effect.Branch)
		if _, err := adapter.effectBase.rest(ctx, credential, http.MethodGet, branchPath, nil, &branch); err != nil || branch.Object.SHA != effect.HeadRevision {
			return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery branch changed before pull request creation", Recoverable: true}, err
		}
		if effect.IssueNumber <= 0 {
			return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery issue number is missing", Recoverable: true}, errors.New("delivery issue number is missing")
		}
		body := map[string]any{"title": effect.Title, "body": "Closes #" + strconv.Itoa(effect.IssueNumber) + "\n\n" + marker, "head": effect.Branch, "base": effect.BaseBranch, "draft": true}
		if _, err := adapter.effectBase.rest(ctx, credential, http.MethodPost, adapter.effectBase.repoPath()+"/pulls", body, nil); err != nil {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "create-pull-request result requires exact re-observation", Recoverable: true}, err
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "created claimed draft pull request"}, nil
	}
	path := adapter.effectBase.repoPath() + "/pulls/" + strconv.Itoa(effect.PullRequest)
	var pull deliveryPull
	if _, err := adapter.effectBase.rest(ctx, credential, http.MethodGet, path, nil, &pull); err != nil {
		return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery pull request cannot be re-observed", Recoverable: true}, err
	}
	if pull.Head.SHA != effect.HeadRevision {
		return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "delivery head changed before effect", Recoverable: true}, nil
	}
	switch effect.Kind {
	case engine.DeliveryEffectRequestReview:
		body := map[string]any{"reviewers": []string{effect.Reviewer}}
		if _, err := adapter.effectBase.rest(ctx, credential, http.MethodPost, path+"/requested_reviewers", body, nil); err != nil {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "request-review result requires exact re-observation", Recoverable: true}, err
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "requested the declared delivery reviewer"}, nil
	case engine.DeliveryEffectMarkReady:
		var response struct {
			Errors []graphQLError `json:"errors"`
		}
		query := `mutation($id:ID!){markPullRequestReadyForReview(input:{pullRequestId:$id}){pullRequest{id}}}`
		if err := adapter.effectBase.mutateGraphQL(ctx, credential, query, map[string]any{"id": pull.NodeID}, &response); err != nil || len(response.Errors) != 0 {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "mark-ready result requires re-observation", Recoverable: true}, err
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "marked exact pull request ready"}, nil
	case engine.DeliveryEffectSquashMerge:
		body := map[string]any{"merge_method": "squash", "sha": effect.HeadRevision}
		var merged struct {
			Merged  bool   `json:"merged"`
			Message string `json:"message"`
			SHA     string `json:"sha"`
		}
		if _, err := adapter.effectBase.rest(ctx, credential, http.MethodPut, path+"/merge", body, &merged); err != nil {
			return engine.DeliveryEffectResult{Outcome: "ambiguous", Detail: "squash-merge result requires exact re-observation", Recoverable: true}, err
		}
		if !merged.Merged || merged.SHA == "" {
			return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "GitHub did not confirm the squash merge", Recoverable: true}, nil
		}
		return engine.DeliveryEffectResult{Outcome: "applied", Detail: "squash merged exact pull request head", ResourceRevision: merged.SHA}, nil
	default:
		return engine.DeliveryEffectResult{Outcome: "needs-review", Detail: "unsupported delivery effect", Recoverable: true}, errors.New("unsupported delivery effect")
	}
}
