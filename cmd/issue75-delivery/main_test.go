package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

var testNow = time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)

func TestRunEmitsDeterministicCredentialFreeBoundArtifacts(t *testing.T) {
	t.Parallel()
	args := validArgs()
	var first, second bytes.Buffer
	if err := run(args, testNow, &first); err != nil {
		t.Fatal(err)
	}
	if err := run(args, testNow, &second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("identical inputs must emit byte-identical JSON")
	}

	var result outputEnvelope
	decoder := json.NewDecoder(bytes.NewReader(first.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	request, mandate := result.Request, result.Mandate
	if request.Repository != "." || request.Intent.ManagedID != "issue:102" || request.Intent.HeadBranch != deliveryBranch || request.Intent.MergeMethod != "squash" {
		t.Fatalf("unexpected exact delivery identity: %#v", request)
	}
	if !slices.Equal(request.Intent.RequiredChecks, []string{requiredCheck}) || request.Intent.Review.Role != reviewer || !request.Intent.Review.DistinctContext || request.Intent.Review.QualifiedIndependent {
		t.Fatalf("delivery assurance is not exact: %#v", request.Intent)
	}
	if request.Intent.ProductApproval.Role != "" {
		t.Fatalf("synthetic fixture must not invent product approval: %#v", request.Intent.ProductApproval)
	}
	if request.Intent.Claim == nil || request.Intent.Claim.ContractDigest == "" || len(request.Intent.Claim.ImplementedSources) != 1 || request.Intent.Claim.ImplementedSources[0].Path != implementedPath {
		t.Fatalf("claim does not bind the exact implemented source: %#v", request.Intent.Claim)
	}
	if _, err := engine.RenderWorkDeliveryClaim(*request.Intent.Claim); err != nil {
		t.Fatalf("claim must be engine-valid: %v", err)
	}

	completion := request.CompletionIntent
	if completion == nil || completion.SchemaVersion != 2 || completion.Task.Readiness != "ready" || completion.Task.Status != "done" || !completion.Task.Closed {
		t.Fatalf("completion intent is incomplete: %#v", completion)
	}
	if completion.Task.ParentManagedID != "issue:101" || completion.Task.ParentContext == nil || !completion.Task.ParentContext.CompletionSatisfied || len(completion.Task.ParentContext.OtherChildren) != 0 {
		t.Fatalf("sole-child parent completion is not exact: %#v", completion.Task.ParentContext)
	}
	if len(completion.Task.Review) != 1 || completion.Task.Review[0].Role != reviewer || !completion.Task.Review[0].DistinctContext || completion.Task.Review[0].QualifiedIndependent {
		t.Fatalf("completion review must match the distinct-capable baseline: %#v", completion.Task.Review)
	}
	if len(completion.Task.Dependents) != 1 || completion.Task.Dependents[0].ManagedID != "issue:103" || completion.Task.Dependents[0].Readiness != "blocked" || completion.Task.Dependents[0].Status != "backlog" || !completion.Task.Dependents[0].ReadyEligible || !slices.Equal(completion.Task.Dependents[0].Blockers, []engine.WorkDependency{{ManagedID: "issue:102", Closed: true}}) {
		t.Fatalf("final-blocker dependent promotion is not exact: %#v", completion.Task.Dependents)
	}
	if completion.Governance == nil {
		t.Fatal("completion governance is required")
	}
	if _, err := engine.RenderExecutableIssueContract(completion.Governance.Issue); err != nil {
		t.Fatalf("executable issue contract must be complete: %v", err)
	}
	if engine.ExecutableIssueContractDigest(completion.Governance.Issue) != request.Intent.Claim.ContractDigest {
		t.Fatal("claim and completion contract digests differ")
	}
	if mandate.ID == "" || engine.BindWorkExecutionMandate(mandate).ID != mandate.ID {
		t.Fatalf("mandate is not content-addressed: %q", mandate.ID)
	}
	if mandate.MaxEffects != 8 || mandate.SelectedManagedID != request.Intent.ManagedID || !slices.Contains(mandate.ResourceDigests, engine.DeliveryResourceDigest(request.Intent)) || !slices.Contains(mandate.ResourceDigests, engine.ManagedTaskResourceDigest(completion.Task)) {
		t.Fatalf("mandate is not exactly resource-bound: %#v", mandate)
	}
	if len(mandate.Authorities) != 2 || mandate.Authorities[0].Actor == mandate.Authorities[1].Actor {
		t.Fatalf("effect and reconciliation authorities are not separated: %#v", mandate.Authorities)
	}
	if len(request.Intent.Target.FieldIDs) != 4 || len(request.Intent.Target.OptionIDs) != 20 {
		t.Fatalf("live Project target is incomplete: %#v", request.Intent.Target)
	}
	if result.ArtifactContract.Parent.Number != 101 || result.ArtifactContract.Delivery.DatabaseID != 202 || result.ArtifactContract.Dependent.NodeID != "I_dependent" || result.ArtifactContract.RequestPointer != "/request" || result.ArtifactContract.MandatePointer != "/mandate" {
		t.Fatalf("artifact split/native identity contract is incomplete: %#v", result.ArtifactContract)
	}
	parsedIssue, err := engine.ParseExecutableIssueContract(result.ArtifactContract.ExecutableIssueBody)
	if err != nil || engine.ExecutableIssueContractDigest(parsedIssue) != request.Intent.Claim.ContractDigest {
		t.Fatalf("artifact issue body does not preserve the claimed executable contract: %v", err)
	}
	if !strings.Contains(result.ArtifactContract.ExecutableIssueBody, deliveryIssueMarker) {
		t.Fatal("artifact issue body lacks the marker-owned fixture identity")
	}
	lower := strings.ToLower(first.String())
	for _, forbidden := range []string{"github_pat_", "ghp_", "private key", "access_token"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("credential marker %q leaked", forbidden)
		}
	}
}

func TestRunRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func([]string) []string
		want   string
	}{
		{name: "source revision", change: replaceFlag("--source-revision", strings.Repeat("A", 40)), want: "source-revision"},
		{name: "implemented digest", change: replaceFlag("--implemented-source-digest", strings.Repeat("b", 64)), want: "implemented-source-digest"},
		{name: "noncanonical number", change: replaceFlag("--parent-number", "01"), want: "parent identity"},
		{name: "missing node", change: replaceFlag("--delivery-node-id", ""), want: "delivery identity"},
		{name: "duplicate number", change: replaceFlag("--dependent-number", "102"), want: "pairwise distinct"},
		{name: "duplicate database id", change: replaceFlag("--dependent-id", "202"), want: "pairwise distinct"},
		{name: "duplicate node id", change: replaceFlag("--dependent-node-id", "I_delivery"), want: "pairwise distinct"},
		{name: "future approval", change: replaceFlag("--approved-at", "2026-07-22T00:00:00Z"), want: "active"},
		{name: "expired approval", change: replaceFlag("--expires-at", "2026-07-21T11:59:59Z"), want: "active"},
		{name: "bad approval time", change: replaceFlag("--approved-at", "yesterday"), want: "RFC3339"},
		{name: "secret marker", change: replaceFlag("--approval-id", "github_pat_forbidden"), want: "credential-like"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := run(test.change(validArgs()), testNow, &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q error, got %v", test.want, err)
			}
			if output.Len() != 0 {
				t.Fatalf("invalid input emitted partial output: %q", output.String())
			}
		})
	}
}

func TestRunRejectsPositionalAndMissingFlags(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"unexpected"}, {"--source-revision", strings.Repeat("a", 40)}} {
		var output bytes.Buffer
		if err := run(args, testNow, &output); err == nil {
			t.Fatalf("expected incomplete arguments to fail: %#v", args)
		}
	}
}

func validArgs() []string {
	return []string{
		"--source-revision", strings.Repeat("a", 40),
		"--implemented-source-digest", "sha256:" + strings.Repeat("b", 64),
		"--approved-by", "dragondad22", "--approval-id", "issue-comment-5039686787",
		"--approved-at", "2026-07-21T00:00:00Z", "--expires-at", "2026-08-20T00:00:00Z",
		"--parent-number", "101", "--parent-id", "201", "--parent-node-id", "I_parent",
		"--delivery-number", "102", "--delivery-id", "202", "--delivery-node-id", "I_delivery",
		"--dependent-number", "103", "--dependent-id", "203", "--dependent-node-id", "I_dependent",
	}
}

func replaceFlag(name, value string) func([]string) []string {
	return func(args []string) []string {
		args = slices.Clone(args)
		for index := range args {
			if args[index] == name {
				args[index+1] = value
				return args
			}
		}
		panic("test flag not found: " + name)
	}
}
