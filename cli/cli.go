// Package cli adapts the language-neutral command interface to the lifecycle engine.
package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dragondad22/codex-starter-kit/engine"
)

// Run executes one CLI request and returns a process-style exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "operation is required")
		return 2
	}
	switch args[0] {
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
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		plan, err := engine.New().Create(context.Background(), *repository)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeJSON(stdout, stderr, plan)
	case "plan":
		flags := flag.NewFlagSet("plan", flag.ContinueOnError)
		flags.SetOutput(stderr)
		repository := flags.String("repository", "", "repository root")
		operation := flags.String("operation", "", "operation to plan")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		plan, err := engine.New().Plan(context.Background(), *repository, engine.Operation(*operation))
		if err != nil {
			fmt.Fprintln(stderr, err)
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
			fmt.Fprintln(stderr, err)
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
	default:
		fmt.Fprintf(stderr, "unsupported operation: %s\n", args[0])
		return 2
	}
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
