package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ctkrug/paywall-sandbox/internal/client"
)

func runRequest(args []string) {
	fs := flag.NewFlagSet("request", flag.ExitOnError)
	url := fs.String("url", "", "URL to request (required)")
	method := fs.String("method", http.MethodGet, "HTTP method to use")
	verbose := fs.Bool("verbose", false, "print every header/descriptor/proof exchanged")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if *url == "" {
		fmt.Fprintln(os.Stderr, "request: --url is required")
		os.Exit(1)
	}

	loop := &client.Loop{Signer: client.FakeSigner{}}
	result, err := loop.Do(context.Background(), *method, *url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		for _, step := range result.Steps {
			printStep(step)
		}
	}

	fmt.Printf("%s %s -> %d\n", *method, *url, result.FinalStatusCode)
	if len(result.FinalBody) > 0 {
		fmt.Println(string(result.FinalBody))
	}

	if result.FinalStatusCode >= http.StatusBadRequest {
		os.Exit(1)
	}
}

func printStep(s client.Step) {
	fmt.Printf("--- %s ---\n%s %s -> %d\n", s.Label, s.Method, s.URL, s.StatusCode)
	if s.Descriptor != nil {
		fmt.Printf("descriptor: %+v\n", *s.Descriptor)
	}
	if s.Proof != nil {
		fmt.Printf("proof: %+v\n", *s.Proof)
	}
}
