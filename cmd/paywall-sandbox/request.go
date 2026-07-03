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

	fmt.Printf("%s %s -> %d\n", *method, *url, result.FinalStatusCode)
	if len(result.FinalBody) > 0 {
		fmt.Println(string(result.FinalBody))
	}

	if result.FinalStatusCode >= http.StatusBadRequest {
		os.Exit(1)
	}
}
