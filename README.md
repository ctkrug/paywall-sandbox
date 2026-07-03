# Paywall Sandbox

[![CI](https://github.com/ctkrug/paywall-sandbox/actions/workflows/ci.yml/badge.svg)](https://github.com/ctkrug/paywall-sandbox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A local mock server and CLI for testing HTTP 402 micropayment flows — the
`x402`-style "pay-per-request" pattern — against your app before you wire up
real settlement.

## Why

HTTP 402 Payment Required has sat unused in the spec for 30 years. A new wave
of "pay-per-call" APIs (x402 and friends) is finally giving it a real shape:
a server challenges a request with `402` plus a payment descriptor, the
client (or an agent acting for it) settles payment out of band, then retries
the request with proof of payment attached. There's no dominant SDK yet, and
the wire format is still being ferreted out from scattered specs and
reference implementations.

Paywall Sandbox lets you develop and test against that flow **without**
touching a real settlement network, a real wallet, or real money:

- Stand up a mock origin server that challenges arbitrary routes with `402`
  and a configurable payment descriptor.
- Drive it with a CLI that plays both sides — issue the initial request,
  inspect the challenge, construct a payment proof, and retry.
- Script assertions against the exchange (status codes, headers, descriptor
  shape, retry behavior) so you can catch integration bugs before they hit a
  real payment rail.

## Features

- **Mock 402 server** — configurable per-route challenge rules (price,
  currency/asset, recipient, expiry, nonce) served as a structured payment
  descriptor in the `402` response body/headers. Routes load from flags or a
  JSON rule set file, with exact-path or `/*`-prefix matching.
- **CLI client** — `paywall-sandbox request --url <url>` to run the full
  challenge → pay → retry loop against any target, mock or real.
- **Pluggable proof schemes** — `fake` (unconditional accept, for exercising
  the protocol shape) and `hmac-sha256` (a shared-secret signature, one step
  closer to real settlement evidence) ship today; see
  [`docs/PROTOCOL.md`](docs/PROTOCOL.md) for how to add another.
- **Scenario scripting** — `paywall-sandbox test <scenario.json>` runs a
  declarative JSON scenario describing an expected challenge/response
  sequence against an in-process server and exits non-zero on any assertion
  failure, so it's usable as a CI check.
- **Inspection mode** — `--verbose` on both `request` and `test` traces every
  header, descriptor field, and proof exchanged, so you can see exactly what
  a real client would need to do.

## Quick start

```console
$ go build -o bin/paywall-sandbox ./cmd/paywall-sandbox
$ ./bin/paywall-sandbox serve --path /paid --amount 100 --asset USDC
paywall-sandbox dev listening on :8402 (1 rule(s))

# in another shell
$ curl -i http://localhost:8402/paid
HTTP/1.1 402 Payment Required
X-Payment-Required: {"amount":100,"asset":"USDC","recipient":"0xsandbox","nonce":"...","expiresAt":"..."}

# or drive the whole challenge -> pay -> retry loop with the CLI client
$ ./bin/paywall-sandbox request --url http://localhost:8402/paid --verbose
--- initial request ---
GET http://localhost:8402/paid -> 402
--- 402 challenge received ---
GET http://localhost:8402/paid -> 402
descriptor: {Amount:100 Asset:USDC Recipient:0xsandbox Nonce:... ExpiresAt:...}
--- retry with proof ---
GET http://localhost:8402/paid -> 200
proof: {Nonce:... Scheme:fake Signature:...}
GET http://localhost:8402/paid -> 200
```

See [`docs/PROTOCOL.md`](docs/PROTOCOL.md) for the full challenge/response
wire format. `request` works against any target, mock or real — it only
assumes the target speaks the protocol documented there.

Multiple routes, or a rule set you want to check into your own repo, load
from a JSON file instead of flags:

```console
$ ./bin/paywall-sandbox serve --config examples/rules.json
paywall-sandbox dev listening on :8402 (2 rule(s))
```

`request` defaults to the `fake` scheme; pass `--scheme hmac-sha256
--hmac-key <secret>` to settle against a server configured for that scheme
instead (see [`docs/PROTOCOL.md`](docs/PROTOCOL.md) for how schemes work).

### Scenario scripting

`test` runs a declarative scenario — its own rule set, hmac key (if any),
and a sequence of requests with expected outcomes — against a server it
starts and tears down itself, so it needs nothing already running:

```console
$ ./bin/paywall-sandbox test examples/scenario.json
scenario: paid route settles, free route does not
  [PASS] GET /paid is challenged and settles with the fake scheme
  [PASS] GET /free is never challenged
```

A failing step is reported and the process exits `1`, so `test` doubles as a
CI assertion:

```console
$ ./bin/paywall-sandbox test examples/scenario.json && echo "scenario ok"
```

See [`examples/scenario.json`](examples/scenario.json) and
[`examples/scenario-hmac.json`](examples/scenario-hmac.json) for runnable
examples, and [`docs/PROTOCOL.md`](docs/PROTOCOL.md) for the full scenario
file format.

## Stack

- **Go** (1.22+) — single static binary, no runtime dependencies.
- Standard library `net/http` for the mock server; `cobra`-style flag
  parsing kept minimal (stdlib `flag` to start, evaluated in the backlog).
- `go test` for unit/integration tests; GitHub Actions for CI.

## Status

The core challenge/response loop, configurable rule sets, pluggable proof
schemes, and scenario scripting are implemented. See
[`docs/VISION.md`](docs/VISION.md) for the design and
[`docs/BACKLOG.md`](docs/BACKLOG.md) for what's left.

## License

MIT — see [`LICENSE`](LICENSE).
