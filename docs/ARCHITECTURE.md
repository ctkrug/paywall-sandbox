# Architecture

A concise map of the codebase for anyone (human or model) picking this up
cold. See [`VISION.md`](VISION.md) for why, [`PROTOCOL.md`](PROTOCOL.md) for
the wire format, [`BACKLOG.md`](BACKLOG.md) for what's left.

## Modules

```
cmd/paywall-sandbox/    CLI entrypoint: subcommand dispatch + flag parsing
internal/paywall/       Wire types shared by both sides: Descriptor, Proof
internal/mockserver/    The mock 402 server: Rule matching, Server, logging
internal/client/        The CLI client side: Signer, Loop (challenge/pay/retry)
internal/scenario/      Declarative scenario files: load, validate, run, report
```

- **`internal/paywall`** — no dependencies beyond stdlib. Defines
  `Descriptor` (what's owed) and `Proof` (what satisfies it), their JSON
  encode/decode, and the two header names (`X-Payment-Required`,
  `X-Payment`) both sides key off. Anything that needs to speak the wire
  format imports this package; it imports nothing else in this repo.
- **`internal/mockserver`** — the server side. `Rule` describes one
  protected route (method/path/price), matching either an exact path or a
  `/*`-suffixed prefix; `Server` is an `http.Handler` that challenges
  unmatched-proof requests and forwards matched-proof ones to `Next`.
  Nonces are issued and consumed in an in-memory map guarded by a mutex —
  one-time use, hard TTL, no persistence (see `docs/PROTOCOL.md` for why).
  Expired-and-never-retried nonces are swept on the next challenge (and
  reclaimed immediately if a proof for one is presented) so a long-running
  `Server` doesn't accumulate them forever.
  `LogRequests` is a middleware wrapping any handler with structured
  request logging. `LoadRules`/`LoadRulesFile` (`config.go`) parse a JSON
  rule set (see `examples/rules.json`) into `[]Rule`, validating every
  entry. `Verifier` (`verifier.go`) is a pluggable interface for checking
  a `Proof.Signature`; `Server.Verifiers` maps scheme name to `Verifier`
  and defaults to `FakeScheme` only. `HMACVerifier` (`hmac_verifier.go`)
  is a second scheme checking a shared-secret HMAC of the nonce.
- **`internal/client`** — the CLI client side. `Signer` builds a `Proof`
  from a received `Descriptor`; `FakeSigner` and `HMACSigner` (mirroring
  `mockserver.HMACVerifier`) are the two implementations. `Loop` drives
  the exchange: send, detect a 402, decode the descriptor, sign, retry
  once — returning a `Result` with a step-by-step `Steps` trace used for
  `--verbose` output. Only depends on `internal/paywall` and stdlib
  `net/http`, so it can run against any target, not just
  `internal/mockserver`.
- **`internal/scenario`** — declarative scenario files (`docs/PROTOCOL.md`
  has the format). `Scenario`/`Step`/`Expect` are the JSON shape;
  `Load`/`LoadFile` parse and `Validate` it (delegating rule validation to
  `mockserver.LoadRules` by round-tripping `Rules` through JSON, so a
  scenario's rule set can never drift from what `serve --config` accepts).
  `Run` starts one `httptest.Server` per scenario (`newServer`, wired with a
  `Verifier` for `fake` always and `hmac-sha256` whenever `HMACKey` is set),
  drives every `Step` through a `client.Loop` (`signerFor` resolves the
  step's scheme to a `client.Signer`), and diffs the actual
  `Paid`/`FinalStatusCode` against `Expect` (`checkExpect`). A step's
  failure is recorded in its `StepResult`, not fatal — the rest of the
  scenario still runs. `Report.String()` renders the PASS/FAIL summary
  `test` prints.
- **`cmd/paywall-sandbox`** — thin CLI wrapper. `serve` stands up a
  `mockserver.Server` with rules either from flags (one rule) or
  `--config <file>` (`mockserver.LoadRulesFile`); `request` drives a
  `client.Loop` against `--url` (`resolveSigner` maps `--scheme`/
  `--hmac-key` to a `client.Signer`), bounded by `--timeout` (default
  `10s`) via `context.WithTimeout`; `test` loads a scenario file and runs
  it via `internal/scenario`, exiting non-zero on any step failure;
  `version` prints the build version.

## Data flow (server side)

```
request → Server.ServeHTTP
            → matchRule (Rule.Matches by method + path)
            → no match:            forward to Next
            → match, no/invalid proof:  issue Descriptor, 402
            → match, valid proof:  consume nonce, forward to Next
```

## Data flow (client side)

```
Loop.Do → send initial request
            → status != 402:  return result (Paid=false)
            → status == 402:  decode Descriptor (header, falls back to body)
                               → Signer.Sign(desc) -> Proof
                               → retry with X-Payment header set
                               → return result (Paid=true)
```

## Data flow (scenario)

```
scenario.Run → newServer (httptest.Server from Scenario.Rules/Verifiers)
                → for each Step:
                    signerFor(Step.Scheme) -> client.Signer
                    client.Loop.Do -> client.Result
                    checkExpect(Step.Expect, Result) -> "" (pass) or mismatch description
                → Report{Steps: []StepResult}
```

## Build / run / test

```
go build -o bin/paywall-sandbox ./cmd/paywall-sandbox
./bin/paywall-sandbox serve --path /paid --amount 100 --asset USDC
./bin/paywall-sandbox test examples/scenario.json
go test ./...
```

`make build|test|run|fmt|vet|lint` wrap the above; see `Makefile`.

`cmd/paywall-sandbox`'s subcommand handlers (`runServe`/`runRequest`/
`runTest`) call `os.Exit` directly on error, so `cli_test.go` covers their
exit-code behavior black-box: `TestMain` builds the binary once and tests
run it as a subprocess. `dispatch()`, `loadServeRules`, and `resolveSigner`
are pure enough to unit test in-process instead.

## Release

Pushing a `v*` tag runs `.github/workflows/release.yml`, which invokes
[GoReleaser](https://goreleaser.com) against `.goreleaser.yaml` to
cross-compile `linux`/`darwin`/`windows` × `amd64`/`arm64` binaries,
stamp `main.version` via ldflags, and publish them as a GitHub Release.
See `docs/RELEASING.md` for the operator steps.

## Where new work plugs in

- New proof schemes: implement `mockserver.Verifier` + `client.Signer` (see
  `docs/PROTOCOL.md`'s "Adding a proof scheme"); `HMACVerifier`/
  `HMACSigner` are the worked example. Neither `Server` nor `Loop` need to
  change.
- New rule config fields (e.g. per-rule TTL): extend `RuleConfig`/`Rule`
  and `validateRuleConfig` in `internal/mockserver/config.go`; `Server`
  itself doesn't need to change.
- New scenario assertions (e.g. asserting on a response header): extend
  `scenario.Expect` and `checkExpect` in `internal/scenario/run.go`;
  `Scenario`/`Step`/`Run` don't need to change.
