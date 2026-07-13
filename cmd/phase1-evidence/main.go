// Command phase1-evidence captures and compares native Phase 1 lifecycle semantics.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dragondad22/codex-starter-kit/internal/nativeevidence"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "capture or compare operation is required")
		return 2
	}
	switch args[0] {
	case "capture":
		flags := flag.NewFlagSet("capture", flag.ContinueOnError)
		output := flags.String("output", "", "native evidence JSON destination")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if *output == "" {
			fmt.Fprintln(os.Stderr, "--output is required")
			return 2
		}
		report, err := nativeevidence.Capture(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := nativeevidence.Write(*output, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	case "compare":
		flags := flag.NewFlagSet("compare", flag.ContinueOnError)
		directory := flags.String("directory", "", "directory containing three native evidence reports")
		output := flags.String("output", "", "optional comparison JSON destination")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if *directory == "" {
			fmt.Fprintln(os.Stderr, "--directory is required")
			return 2
		}
		summary, err := nativeevidence.Compare(*directory)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if *output != "" {
			if err := writeJSONFile(*output, summary); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(summary); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unsupported operation: %s\n", args[0])
		return 2
	}
}

func writeJSONFile(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}
