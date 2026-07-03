package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/client"
)

func runRequest(args []string) {
	fs := flag.NewFlagSet("request", flag.ExitOnError)
	url := fs.String("url", "", "URL to request (required)")
	method := fs.String("method", http.MethodGet, "HTTP method to use")
	scheme := fs.String("scheme", client.FakeScheme, "proof scheme to settle a challenge with (fake, hmac-sha256)")
	hmacKey := fs.String("hmac-key", "", "shared secret for --scheme hmac-sha256")
	verbose := fs.Bool("verbose", false, "print every header/descriptor/proof exchanged")
	timeout := fs.Duration("timeout", 10*time.Second, "give up if the target doesn't respond within this long")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if *url == "" {
		fmt.Fprintln(os.Stderr, "request: --url is required")
		os.Exit(1)
	}

	signer, err := resolveSigner(*scheme, *hmacKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	loop := &client.Loop{Signer: signer}
	result, err := loop.Do(ctx, *method, *url)
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

// resolveSigner maps a --scheme flag value to the client.Signer that
// produces it, matching the schemes internal/scenario resolves scenario
// steps against.
func resolveSigner(scheme, hmacKey string) (client.Signer, error) {
	switch scheme {
	case client.FakeScheme:
		return client.FakeSigner{}, nil
	case client.HMACScheme:
		if hmacKey == "" {
			return nil, fmt.Errorf("--scheme %s requires --hmac-key", client.HMACScheme)
		}
		return client.HMACSigner{Key: []byte(hmacKey)}, nil
	default:
		return nil, fmt.Errorf("unknown --scheme %q", scheme)
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
