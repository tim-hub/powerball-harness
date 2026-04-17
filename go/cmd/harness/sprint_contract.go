package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tim-hub/powerball-harness/go/internal/hookhandler"
)

// runSprintContract implements:
//
//	harness sprint-contract <task-id> [plans-file] [output-file]
//
// Generates a sprint-contract JSON from Plans.md for the given task ID.
func runSprintContract(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "Usage: harness sprint-contract <task-id> [plans-file] [output-file]")
		os.Exit(1)
	}

	taskID := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness sprint-contract: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}

	var plansFile string
	if len(args) >= 2 && args[1] != "" {
		plansFile, err = filepath.Abs(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "harness sprint-contract: invalid plans file %q: %v\n", args[1], err)
			os.Exit(1)
		}
	}

	var outputFile string
	if len(args) >= 3 && args[2] != "" {
		outputFile, err = filepath.Abs(args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "harness sprint-contract: invalid output file %q: %v\n", args[2], err)
			os.Exit(1)
		}
	}

	projectRoot := cwd
	if plansFile != "" {
		projectRoot = filepath.Dir(plansFile)
	}

	generator := &hookhandler.SprintContractGenerator{
		ProjectRoot: projectRoot,
		PlansFile:   plansFile,
		OutputFile:  outputFile,
	}

	written, err := generator.Write(taskID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(written)
}
