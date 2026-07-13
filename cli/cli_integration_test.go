package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dragondad22/codex-starter-kit/cli"
	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestCreateCommandEmitsLanguageNeutralPlan(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run(createArguments("create", repository), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var plan engine.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan JSON: %v: %s", err, stdout.String())
	}
	if plan.Operation != engine.CreateOperation || plan.ID == "" {
		t.Fatalf("unexpected create plan: %#v", plan)
	}
}

func TestCreateCommandEmitsStructuredReconciliation(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("human work\n"), 0o644); err != nil {
		t.Fatalf("write human-owned conflict: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run(createArguments("create", repository), &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var reconciliation engine.ReconciliationRequired
	if err := json.Unmarshal(stderr.Bytes(), &reconciliation); err != nil {
		t.Fatalf("decode reconciliation JSON: %v: %s", err, stderr.String())
	}
	if len(reconciliation.Conflicts) != 1 || reconciliation.Conflicts[0].Path != "README.md" || len(reconciliation.Recovery) == 0 {
		t.Fatalf("unexpected reconciliation result: %#v", reconciliation)
	}
}

func TestApplyAndStatusCommandsUsePlanIdentifier(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	planDocument, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("encode plan: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(planPath, planDocument, 0o600); err != nil {
		t.Fatalf("write plan fixture: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{
		"apply", "--plan", planPath, "--plan-id", plan.ID,
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("apply exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var result engine.ApplyResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode apply result: %v", err)
	}
	if result.Status != engine.ApplyStatusApplied {
		t.Fatalf("apply status = %q", result.Status)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = cli.Run([]string{"status", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("status exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var status engine.RepositoryStatus
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		t.Fatalf("decode status result: %v", err)
	}
	if status.Lifecycle != engine.LifecycleManaged {
		t.Fatalf("lifecycle = %q, want managed", status.Lifecycle)
	}
}

func TestApplyCommandPreservesStructuredConflictAndRecoveryDetails(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	plan, err := engine.New().Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	document, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("encode plan: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(planPath, document, 0o600); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("new human work\n"), 0o644); err != nil {
		t.Fatalf("write post-plan conflict: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"apply", "--plan", planPath, "--plan-id", plan.ID}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var envelope struct {
		Result  engine.ApplyResult   `json:"result"`
		Failure *engine.ApplyFailure `json:"failure"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &envelope); err != nil {
		t.Fatalf("decode apply failure JSON: %v: %s", err, stderr.String())
	}
	if envelope.Failure == nil || envelope.Failure.Stage != "reconcile" || len(envelope.Failure.Conflicts) != 1 || len(envelope.Failure.Recovery) == 0 {
		t.Fatalf("structured apply failure lost reconciliation facts: %#v", envelope)
	}
}

func TestInspectAndPlanCommandsExposeEngineOperations(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{"inspect", "--repository", repository}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("inspect exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var inspection engine.Inspection
	if err := json.Unmarshal(stdout.Bytes(), &inspection); err != nil {
		t.Fatalf("decode inspection: %v", err)
	}
	if !inspection.Git || inspection.Managed {
		t.Fatalf("unexpected inspection: %#v", inspection)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = cli.Run(createArguments("plan", repository), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("plan exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var plan engine.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan: %v", err)
	}
	if plan.Operation != engine.CreateOperation || plan.ID == "" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestVerifyCommandEmitsMachineReadableControlResults(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	lifecycle := engine.New()
	plan, err := lifecycle.Create(t.Context(), approvedCreate(repository))
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := lifecycle.Apply(t.Context(), plan.ID, plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cli.Run([]string{
		"verify-plan", "--repository", repository, "--scope", "repository", "--gate", "development",
		"--actor", "integration-test", "--authority", "approved issue #27 fixture",
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("verify-plan exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var verifyPlan engine.VerifyPlan
	if err := json.Unmarshal(stdout.Bytes(), &verifyPlan); err != nil {
		t.Fatalf("decode verification plan: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "verify-plan.json")
	if err := os.WriteFile(planPath, stdout.Bytes(), 0o600); err != nil {
		t.Fatalf("write verification plan: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	exitCode = cli.Run([]string{"verify", "--plan", planPath, "--plan-id", verifyPlan.ID}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("verify exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	var result engine.VerificationResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode verification result: %v", err)
	}
	if result.OverallState == engine.ControlPass || len(result.Controls) == 0 {
		t.Fatalf("unexpected verification result: %#v", result)
	}
}

func approvedCreate(repository string) engine.CreateRequest {
	return engine.CreateRequest{
		Repository:            repository,
		Brief:                 "Create a managed repository for CLI testing.",
		BriefApproved:         true,
		OwnerPersonaConfirmed: true,
	}
}

func createArguments(operation, repository string) []string {
	args := []string{
		operation,
		"--repository", repository,
		"--brief", "Create a managed repository for CLI testing.",
		"--approve-brief",
		"--confirm-owner-persona",
	}
	if operation == "plan" {
		args = append(args, "--operation", "create")
	}
	return args
}
