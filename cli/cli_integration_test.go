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

	exitCode := cli.Run([]string{"create", "--repository", repository}, &stdout, &stderr)
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

func TestApplyAndStatusCommandsUsePlanIdentifier(t *testing.T) {
	repository := t.TempDir()
	command := exec.Command("git", "init", "--quiet", repository)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("initialize Git repository: %v: %s", err, output)
	}
	plan, err := engine.New().Create(t.Context(), repository)
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
	exitCode = cli.Run([]string{
		"plan", "--operation", "create", "--repository", repository,
	}, &stdout, &stderr)
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
