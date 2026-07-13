package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestVerifyCreatedRepositoryEmitsTruthfulSeedResults(t *testing.T) {
	repository := newGitRepository(t)
	clock := fixedClock{time.Date(2026, 7, 12, 20, 0, 0, 0, time.UTC)}
	lifecycle := engine.New(engine.WithClock(clock))
	plan, err := lifecycle.Create(context.Background(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(context.Background(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}

	verifyPlan, err := lifecycle.PrepareVerify(context.Background(), engine.VerifyRequest{
		Repository: repository,
		Scope:      "repository",
		Gate:       "development",
		Actor:      "integration-test",
		Authority:  "approved issue #27 fixture",
	})
	if err != nil {
		t.Fatalf("prepare verify: %v", err)
	}
	result, err := lifecycle.Verify(context.Background(), verifyPlan.ID, verifyPlan)
	if err != nil {
		t.Fatalf("verify repository: %v", err)
	}
	if result.OverallState != engine.ControlNotConfigured {
		t.Fatalf("overall state = %q, want not-configured", result.OverallState)
	}
	wantStates := map[string]engine.ControlState{
		"CORE-TRUTH-001":     engine.ControlPass,
		"CORE-SECRETS-001":   engine.ControlNotConfigured,
		"CORE-OWNERSHIP-001": engine.ControlPass,
		"CORE-COVERAGE-001":  engine.ControlPass,
		"CORE-RECOVERY-001":  engine.ControlNotConfigured,
		"CORE-ROUTES-001":    engine.ControlPass,
	}
	if len(result.Controls) != len(wantStates) {
		t.Fatalf("control count = %d, want %d", len(result.Controls), len(wantStates))
	}
	for _, control := range result.Controls {
		if control.State != wantStates[control.ID] {
			t.Fatalf("control %s state = %q, want %q", control.ID, control.State, wantStates[control.ID])
		}
		if control.State == engine.ControlPass && len(control.Evidence) == 0 {
			t.Fatalf("passing control %s lacks evidence", control.ID)
		}
	}
	if result.VerifiedAt != clock.Now() || result.EvidencePath == "" {
		t.Fatalf("verification provenance incomplete: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(repository, filepath.FromSlash(result.EvidencePath))); err != nil {
		t.Fatalf("machine evidence missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repository, "docs", "evidence", "CONFORMANCE.md")); err != nil {
		t.Fatalf("human conformance summary missing: %v", err)
	}
	status, err := lifecycle.Status(context.Background(), repository)
	if err != nil {
		t.Fatalf("status after verify: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManaged {
		t.Fatalf("verification left contract invalid: %#v", status)
	}
}

func TestAggregateNeverConvertsExplicitNonPassStateIntoPass(t *testing.T) {
	states := []engine.ControlState{
		engine.ControlFail,
		engine.ControlNotApplicable,
		engine.ControlNotConfigured,
		engine.ControlNeedsReview,
		engine.ControlAcceptedException,
	}
	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			results := []engine.ControlResult{
				{ID: "CORE-PASS-001", State: engine.ControlPass},
				{ID: "CORE-FIXTURE-001", State: state, UnderlyingState: engine.ControlFail},
			}
			if got := engine.OverallControlState(results); got == engine.ControlPass {
				t.Fatalf("aggregate converted %q into pass", state)
			}
		})
	}
}

func TestVerifyPersistsFailureEvidenceForDegradedRepository(t *testing.T) {
	repository := newGitRepository(t)
	lifecycle := engine.New(engine.WithClock(fixedClock{time.Date(2026, 7, 12, 22, 0, 0, 0, time.UTC)}))
	createPlan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), createPlan.ID, createPlan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".starter-kit", "routes.json"), []byte("{\"schema_version\":1,\"routes\":{}}\n"), 0o644); err != nil {
		t.Fatalf("damage route index: %v", err)
	}
	verifyPlan, err := lifecycle.PrepareVerify(t.Context(), engine.VerifyRequest{
		Repository: repository, Scope: "repository", Gate: "development",
		Actor: "integration-test", Authority: "approved degraded fixture",
	})
	if err != nil {
		t.Fatalf("prepare verify: %v", err)
	}
	result, err := lifecycle.Verify(t.Context(), verifyPlan.ID, verifyPlan)
	if err != nil {
		t.Fatalf("verify degraded repository: %v", err)
	}
	if result.OverallState != engine.ControlFail {
		t.Fatalf("degraded repository state = %q, want fail", result.OverallState)
	}
	if _, err := os.Stat(filepath.Join(repository, filepath.FromSlash(result.EvidencePath))); err != nil {
		t.Fatalf("failure evidence missing: %v", err)
	}
	status, err := lifecycle.Status(t.Context(), repository)
	if err != nil {
		t.Fatalf("status degraded repository: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManagedDegraded {
		t.Fatalf("failure evidence concealed degraded state: %#v", status)
	}
}

func TestVerifyEquivalentControlledRepositoriesProducesEquivalentSemantics(t *testing.T) {
	clock := fixedClock{time.Date(2026, 7, 12, 21, 0, 0, 0, time.UTC)}
	verify := func(t *testing.T) engine.VerificationResult {
		t.Helper()
		repository := newGitRepository(t)
		lifecycle := engine.New(engine.WithClock(clock))
		plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
		if err != nil {
			t.Fatalf("create plan: %v", err)
		}
		if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
			t.Fatalf("apply plan: %v", err)
		}
		verifyPlan, err := lifecycle.PrepareVerify(t.Context(), engine.VerifyRequest{
			Repository: repository, Scope: "repository", Gate: "development",
			Actor: "integration-test", Authority: "approved issue #27 fixture",
		})
		if err != nil {
			t.Fatalf("prepare verify: %v", err)
		}
		result, err := lifecycle.Verify(t.Context(), verifyPlan.ID, verifyPlan)
		if err != nil {
			t.Fatalf("verify repository: %v", err)
		}
		return result
	}

	first := verify(t)
	second := verify(t)
	if first.OverallState != second.OverallState || !reflect.DeepEqual(first.Controls, second.Controls) || !reflect.DeepEqual(first.CoverageLimitations, second.CoverageLimitations) {
		t.Fatalf("equivalent inputs produced different semantics:\n%#v\n%#v", first, second)
	}
}

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }
