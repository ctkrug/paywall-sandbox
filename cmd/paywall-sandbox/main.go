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
	os.Exit(dispatch(os.Args[1:]))
}

// dispatch resolves a subcommand and runs it, returning the process exit
// code. Split out from main so the routing logic (as opposed to the
// subcommands themselves, which call os.Exit directly) is testable.
func dispatch(args []string) int {
	if len(args) < 1 {
		usage()
		return 1
	}

	switch args[0] {
	case "version":
		fmt.Println("paywall-sandbox " + version)
		return 0
	case "serve":
		runServe(args[1:])
		return 0
	case "request":
		runRequest(args[1:])
		return 0
	case "test":
		runTest(args[1:])
		return 0
	default:
		usage()
		return 1
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: paywall-sandbox <version|serve|request|test> [flags]")
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8402", "address to listen on")
	path := fs.String("path", "/paid", "route that requires payment (ignored if --config is set)")
	amount := fs.Uint64("amount", 100, "price in the smallest unit of --asset (ignored if --config is set)")
	asset := fs.String("asset", "USDC", "asset/currency identifier (ignored if --config is set)")
	recipient := fs.String("recipient", "0xsandbox", "recipient address (ignored if --config is set)")
	config := fs.String("config", "", "path to a JSON rule config file (overrides --path/--amount/--asset/--recipient)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	rules, err := loadServeRules(*config, *path, *amount, *asset, *recipient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}

	srv := &mockserver.Server{Rules: rules}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("paywall-sandbox %s listening on %s (%d rule(s))", version, *addr, len(rules))
	log.Fatal(http.ListenAndServe(*addr, mockserver.LogRequests(logger, srv)))
}

func loadServeRules(configPath, path string, amount uint64, asset, recipient string) ([]mockserver.Rule, error) {
	if configPath == "" {
		return []mockserver.Rule{{Path: path, Amount: amount, Asset: asset, Recipient: recipient}}, nil
	}
	return mockserver.LoadRulesFile(configPath)
}
