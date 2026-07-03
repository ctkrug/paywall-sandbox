# Vision

## The problem

HTTP 402 Payment Required is being repurposed by a new wave of "pay-per-call"
API proposals (x402 and similar), where a server challenges a request with
`402` plus a machine-readable payment descriptor, and a client (often an
autonomous agent) settles payment and retries. There's no single ratified
spec yet — just a handful of reference implementations and blog posts
describing roughly the same shape.

That's a problem for anyone building against it today: you can't develop or
test your integration without either (a) standing up a real settlement rail
just to exercise error paths and retries, or (b) hand-rolling throwaway mocks
that drift from what real servers actually send.

## Who it's for

Developers building either side of a 402 pay-per-call integration:

- **API builders** who want to add a paywall to a route and need to verify
  their challenge responses are well-formed before wiring up real settlement.
- **Client/agent builders** who need to test the challenge → pay → retry loop
  against a target that behaves like a real 402 server, without spending real
  money per test run.

## The core idea

A local, dependency-free mock that speaks the 402 challenge/response
protocol correctly and predictably:

- A **mock server** that challenges configured routes exactly like a real
  pay-per-call API would — same headers, same descriptor shape, same nonce
  and expiry semantics — but settles nothing for real.
- A **CLI** that can drive that server (or, eventually, any real target)
  through the full exchange, so both sides of an integration can be
  developed and tested independently.

Correctness of the protocol shape is the whole value proposition: a sandbox
that doesn't match how real challenges are issued and validated is worse
than no sandbox at all, because it teaches the wrong lesson.

## Key design decisions

- **Go, single static binary.** No runtime, no dependency install step —
  `go build` and you have something you can hand to a teammate or drop into
  CI.
- **Stdlib-first.** `net/http` and `encoding/json` are enough to model this
  protocol correctly; pulling in a framework would obscure the wire format
  this project exists to make legible.
- **One-time nonces, hard expiry.** The mock enforces the same replay and
  freshness rules a real settlement-backed server would need, so tests
  written against it catch real integration bugs (stale proofs, replayed
  headers) instead of only the happy path.
- **`fake` is explicitly fake.** The one proof scheme shipped in v1 exists to
  exercise the protocol shape, not to simulate settlement security. Real
  schemes are additive, not a rewrite — see [`BACKLOG.md`](BACKLOG.md).
- **Wire format is documented, not just implemented.** See
  [`PROTOCOL.md`](PROTOCOL.md) — the spec is a first-class artifact here,
  not an afterthought to the code.

## What "v1 done" looks like

- `paywall-sandbox serve` stands up a configurable mock 402 server with
  one or more protected routes.
- A CLI client can run the full challenge → pay → retry loop against any
  target (mock or real) and report what happened at each step.
- Rules are configurable from a file, not just flags, so a test suite can
  check in its mock server's shape alongside its tests.
- A declarative scenario format lets a user script an expected
  challenge/response sequence and run it in CI as an assertion, not just a
  manual `curl`.
- CI is green on build, vet, race-enabled tests, and lint for every push.
