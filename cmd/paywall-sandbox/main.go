// Command paywall-sandbox runs a local mock server for testing HTTP 402
// micropayment flows.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("paywall-sandbox " + version)
	case "serve":
		runServe(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: paywall-sandbox <version|serve> [flags]")
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8402", "address to listen on")
	path := fs.String("path", "/paid", "route that requires payment")
	amount := fs.Uint64("amount", 100, "price in the smallest unit of --asset")
	asset := fs.String("asset", "USDC", "asset/currency identifier")
	recipient := fs.String("recipient", "0xsandbox", "recipient address")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	srv := &mockserver.Server{
		Rules: []mockserver.Rule{
			{Path: *path, Amount: *amount, Asset: *asset, Recipient: *recipient},
		},
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("paywall-sandbox %s listening on %s (paid route: %s)", version, *addr, *path)
	log.Fatal(http.ListenAndServe(*addr, mockserver.LogRequests(logger, srv)))
}
