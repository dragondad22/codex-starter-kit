package githubadapter

import (
	"slices"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestFinalizeDeliveryObservationCanonicalizesEvidenceOrder(t *testing.T) {
	now := time.Date(2026, 7, 26, 23, 30, 0, 0, time.UTC)
	checks := []engine.DeliveryCheckObservation{
		{Name: "second", EvidenceID: "check-run:2", ObservedAt: now.Add(time.Second)},
		{Name: "first", EvidenceID: "check-run:1", ObservedAt: now},
	}
	reviews := []engine.DeliveryReviewObservation{
		{Actor: "reviewer", EvidenceID: "review:2", ObservedAt: now.Add(time.Second)},
		{Actor: "reviewer", EvidenceID: "review:1", ObservedAt: now},
	}
	approvals := []engine.DeliveryApprovalObservation{
		{Actor: "approver", EvidenceID: "review:4", ObservedAt: now.Add(time.Second)},
		{Actor: "approver", EvidenceID: "review:3", ObservedAt: now},
	}
	left := engine.DeliveryObservation{
		SchemaVersion: 1,
		Problems:      []string{"second problem", "first problem"},
		Checks:        slices.Clone(checks),
		Reviews:       slices.Clone(reviews),
		Approvals:     slices.Clone(approvals),
		Rules:         engine.DeliveryRulesObservation{Problems: []string{"second rule problem", "first rule problem"}},
	}
	reversedChecks := slices.Clone(checks)
	reversedReviews := slices.Clone(reviews)
	reversedApprovals := slices.Clone(approvals)
	slices.Reverse(reversedChecks)
	slices.Reverse(reversedReviews)
	slices.Reverse(reversedApprovals)
	right := engine.DeliveryObservation{
		SchemaVersion: 1,
		Problems:      []string{"first problem", "second problem"},
		Checks:        reversedChecks,
		Reviews:       reversedReviews,
		Approvals:     reversedApprovals,
		Rules:         engine.DeliveryRulesObservation{Problems: []string{"first rule problem", "second rule problem"}},
	}

	left = finalizeDeliveryObservation(left)
	right = finalizeDeliveryObservation(right)

	if left.Revision == "" || left.Revision != right.Revision {
		t.Fatalf("revisions differ for equivalent evidence: %q != %q", left.Revision, right.Revision)
	}
	if left.Checks[0].EvidenceID != "check-run:1" || left.Reviews[0].EvidenceID != "review:1" || left.Approvals[0].EvidenceID != "review:3" {
		t.Fatalf("evidence was not canonicalized: checks=%#v reviews=%#v approvals=%#v", left.Checks, left.Reviews, left.Approvals)
	}
}
