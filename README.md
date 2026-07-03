# Paywall Sandbox

A local mock server and CLI for testing HTTP 402 micropayment flows — the
`x402`-style "pay-per-request" pattern — against your app before you wire up
real settlement.

## Why

HTTP 402 Payment Required has sat unused in the spec for 30 years. A new wave
of "pay-per-call" APIs (x402 and friends) is finally giving it a real shape:
a server challenges a request with `402` plus a payment descriptor, the
client (or an agent acting for it) settles payment out of band, then retries
the request with proof of payment attached. There's no dominant SDK yet, and
the wire format is still being feretted out from scattered specs and reference
implementations.

Paywall Sandbox lets you develop and test against that flow **without**
touching a real settlement network, a real wallet, or real money:

- Stand up a mock origin server that challenges arbitrary routes with `402`
  and a configurable payment descriptor.
- Drive it with a CLI that plays both sides — issue the initial request,
  inspect the challenge, construct a payment proof, and retry.
- Script assertions against the exchange (status codes, headers, descriptor
  shape, retry behavior) so you can catch integration bugs before they hit a
  real payment rail.

## Planned features

- **Mock 402 server** — configurable per-route challenge rules (price,
  currency/asset, recipient, expiry, nonce) served as a structured payment
  descriptor in the `402` response body/headers.
- **CLI client** — `paywall-sandbox request <url>` to run the full
  challenge → pay → retry loop against any target, mock or real.
- **Proof construction** — pluggable payment-proof builders, starting with a
  fake/local scheme, so real settlement integrations can be swapped in later.
- **Scenario scripting** — declarative YAML/JSON scenarios describing
  expected challenge/response sequences, runnable in CI.
- **Inspection mode** — verbose logging of every header, descriptor field,
  and retry so you can see exactly what a real client would need to do.

## Stack

- **Go** (1.22+) — single static binary, no runtime dependencies.
- Standard library `net/http` for the mock server; `cobra`-style flag
  parsing kept minimal (stdlib `flag` to start, evaluated in the backlog).
- `go test` for unit/integration tests; GitHub Actions for CI.

## Status

Early scaffold — see [`docs/VISION.md`](docs/VISION.md) for the design and
[`docs/BACKLOG.md`](docs/BACKLOG.md) for the build plan.

## License

MIT — see [`LICENSE`](LICENSE).
