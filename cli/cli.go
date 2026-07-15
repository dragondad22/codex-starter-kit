// Package cli adapts the language-neutral command interface to the lifecycle engine.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	starterkit "github.com/dragondad22/codex-starter-kit"
	"github.com/dragondad22/codex-starter-kit/engine"
	"github.com/dragondad22/codex-starter-kit/releasechange"
)

// Run executes one CLI request and returns a process-style exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "operation is required")
		return 2
	}
	switch args[0] {
	case "release":
		if len(args) < 2 || args[1] != "prepare" && args[1] != "recover" {
			fmt.Fprintln(stderr, "release prepare or recover operation is required")
			return 2
		}
		flags := flag.NewFlagSet("release "+args[1], flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root containing release records")
		if args[1] == "recover" {
			if err := flags.Parse(args[2:]); err != nil {
				return 2
			}
			if *repository == "" || flags.NArg() != 0 {
				fmt.Fprintln(stderr, "--repository is required and positional arguments are unsupported")
				return 2
			}
			result, err := releasechange.Recover(*repository)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			return writeJSON(stdout, stderr, result)
		}
		version := flags.String("version", "", "explicit next stable semantic version")
		date := flags.String("date", "", "explicit release date in YYYY-MM-DD form")
		admission := flags.String("admission", "", "approved release-admission JSON path")
		if err := flags.Parse(args[2:]); err != nil {
			return 2
		}
		if *repository == "" || *version == "" || *date == "" || *admission == "" || flags.NArg() != 0 {
			fmt.Fprintln(stderr, "--repository, --version, --date, and --admission are required; positional arguments are unsupported")
			return 2
		}
		result, err := releasechange.Prepare(*repository, *version, *date, *admission)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, result)
	case "changes":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "changes check, validate, or render operation is required")
			return 2
		}
		flags := flag.NewFlagSet("changes "+args[1], flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root containing change records")
		var audience *string
		var release *string
		if args[1] == "render" {
			audience = flags.String("audience", "", "optional audience filter")
			release = flags.String("release", "", "optional prepared release version")
		} else if args[1] != "validate" && args[1] != "check" {
			fmt.Fprintln(stderr, "changes check, validate, or render operation is required")
			return 2
		}
		if err := flags.Parse(args[2:]); err != nil {
			return 2
		}
		if *repository == "" || flags.NArg() != 0 {
			fmt.Fprintln(stderr, "--repository is required and positional arguments are unsupported")
			return 2
		}
		if args[1] == "validate" || args[1] == "check" {
			var result releasechange.ValidationResult
			var err error
			if args[1] == "check" {
				result, err = releasechange.Check(*repository)
			} else {
				result, err = releasechange.Validate(*repository)
			}
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			return writeJSON(stdout, stderr, result)
		}
		var document string
		var err error
		if *release == "" {
			document, err = releasechange.Render(*repository, *audience)
		} else {
			document, err = releasechange.RenderRelease(*repository, *release, *audience)
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, err := io.WriteString(stdout, document); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "version":
		flags := flag.NewFlagSet("version", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "version accepts no arguments")
			return 2
		}
		return writeJSON(stdout, stderr, struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{Name: "codex-starter-kit", Version: starterkit.Version()})
	case "capabilities":
		flags := flag.NewFlagSet("capabilities", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "capabilities accepts no arguments")
			return 2
		}
		return writeJSON(stdout, stderr, engine.New().Capabilities())
	case "manage-task":
		flags := flag.NewFlagSet("manage-task", flag.ContinueOnError)
		flags.SetOutput(stderr)
		inputPath := flags.String("input", "", "versioned managed-task request, capability, and observation JSON")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if *inputPath == "" || flags.NArg() != 0 {
			fmt.Fprintln(stderr, "--input is required and positional arguments are unsupported")
			return 2
		}
		content, err := os.ReadFile(*inputPath)
		if err != nil {
			fmt.Fprintf(stderr, "read managed-task input: %v\n", err)
			return 1
		}
		var input struct {
			Request     engine.ManagedTaskRequest `json:"request"`
			Capability  engine.WorkCapability     `json:"capability"`
			Observation engine.WorkObservation    `json:"observation"`
		}
		if err := decodeOneCLIJSON(content, &input); err != nil {
			fmt.Fprintf(stderr, "decode managed-task input: %v\n", err)
			return 1
		}
		adapter := engine.NewInMemoryWorkAdapter(input.Capability, input.Observation)
		journey, err := engine.New(engine.WithWorkAdapter(adapter)).ManageTask(context.Background(), input.Request)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, journey)
	case "inspect":
		flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		inspection, err := engine.New().Inspect(context.Background(), *repository)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, inspection)
	case "create":
		flags := flag.NewFlagSet("create", flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root")
		brief := flags.String("brief", "", "approved project brief")
		briefApproved := flags.Bool("approve-brief", false, "confirm the supplied brief is approved")
		ownerConfirmed := flags.Bool("confirm-owner-persona", false, "confirm the seed owner persona")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		plan, err := engine.New().Create(context.Background(), engine.CreateRequest{
			Repository:            *repository,
			Brief:                 *brief,
			BriefApproved:         *briefApproved,
			OwnerPersonaConfirmed: *ownerConfirmed,
		})
		if err != nil {
			writeEngineError(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, plan)
	case "plan":
		flags := flag.NewFlagSet("plan", flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root")
		operation := flags.String("operation", "", "operation to plan")
		brief := flags.String("brief", "", "approved project brief")
		briefApproved := flags.Bool("approve-brief", false, "confirm the supplied brief is approved")
		ownerConfirmed := flags.Bool("confirm-owner-persona", false, "confirm the seed owner persona")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		plan, err := engine.New().Plan(context.Background(), engine.PlanRequest{
			Operation: engine.Operation(*operation),
			Create: engine.CreateRequest{
				Repository:            *repository,
				Brief:                 *brief,
				BriefApproved:         *briefApproved,
				OwnerPersonaConfirmed: *ownerConfirmed,
			},
		})
		if err != nil {
			writeEngineError(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, plan)
	case "apply":
		flags := flag.NewFlagSet("apply", flag.ContinueOnError)
		flags.SetOutput(stderr)
		planPath := flags.String("plan", "", "path to plan JSON")
		planID := flags.String("plan-id", "", "expected plan identifier")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		content, err := os.ReadFile(*planPath)
		if err != nil {
			fmt.Fprintf(stderr, "read plan: %v\n", err)
			return 1
		}
		var plan engine.Plan
		if err := json.Unmarshal(content, &plan); err != nil {
			fmt.Fprintf(stderr, "decode plan: %v\n", err)
			return 1
		}
		result, err := engine.New().Apply(context.Background(), *planID, plan)
		if err != nil {
			writeApplyFailure(stderr, result, err)
			return 1
		}
		return writeJSON(stdout, stderr, result)
	case "status":
		flags := flag.NewFlagSet("status", flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		status, err := engine.New().Status(context.Background(), *repository)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, status)
	case "verify":
		flags := flag.NewFlagSet("verify", flag.ContinueOnError)
		flags.SetOutput(stderr)
		planPath := flags.String("plan", "", "path to verification plan JSON")
		planID := flags.String("plan-id", "", "expected verification plan identifier")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		content, err := os.ReadFile(*planPath)
		if err != nil {
			fmt.Fprintf(stderr, "read verification plan: %v\n", err)
			return 1
		}
		var plan engine.VerifyPlan
		if err := json.Unmarshal(content, &plan); err != nil {
			fmt.Fprintf(stderr, "decode verification plan: %v\n", err)
			return 1
		}
		result, err := engine.New().Verify(context.Background(), *planID, plan)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, result)
	case "verify-plan":
		flags := flag.NewFlagSet("verify-plan", flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root")
		scope := flags.String("scope", "", "verification scope")
		gate := flags.String("gate", "", "lifecycle gate")
		actor := flags.String("actor", "", "requesting actor")
		authority := flags.String("authority", "", "authority for evidence regeneration")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		plan, err := engine.New().PrepareVerify(context.Background(), engine.VerifyRequest{
			Repository: *repository, Scope: *scope, Gate: *gate, Actor: *actor, Authority: *authority,
		})
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, plan)
	default:
		fmt.Fprintf(stderr, "unsupported operation: %s\n", args[0])
		return 2
	}
}

func decodeOneCLIJSON(content []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}

func writeApplyFailure(stderr io.Writer, result engine.ApplyResult, err error) {
	failure := struct {
		SchemaVersion int                  `json:"schema_version"`
		Result        engine.ApplyResult   `json:"result"`
		Failure       *engine.ApplyFailure `json:"failure,omitempty"`
		Error         string               `json:"error"`
		Recoverable   bool                 `json:"recoverable"`
	}{SchemaVersion: 1, Result: result, Error: err.Error()}
	var applyFailure *engine.ApplyFailure
	if errors.As(err, &applyFailure) {
		failure.Recoverable = applyFailure.Recoverable
		failure.Failure = applyFailure
	}
	encoder := json.NewEncoder(stderr)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(failure)
}

func writeEngineError(stderr io.Writer, err error) {
	var reconciliation *engine.ReconciliationRequired
	if errors.As(err, &reconciliation) {
		encoder := json.NewEncoder(stderr)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(reconciliation)
		return
	}
	fmt.Fprintln(stderr, err)
}

func writeJSON(stdout, stderr io.Writer, value interface{}) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
