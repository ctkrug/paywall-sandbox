package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ctkrug/paywall-sandbox/internal/scenario"
)

func runTest(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "print every header/descriptor/proof exchanged per step")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "test: exactly one scenario file is required")
		os.Exit(1)
	}

	sc, err := scenario.LoadFile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "test: %v\n", err)
		os.Exit(1)
	}

	report, err := scenario.Run(context.Background(), sc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		for _, sr := range report.Steps {
			if sr.Result == nil {
				continue
			}
			fmt.Printf("=== %s ===\n", sr.Name)
			for _, step := range sr.Result.Steps {
				printStep(step)
			}
		}
	}

	fmt.Print(report.String())
	if !report.Passed() {
		os.Exit(1)
	}
}
