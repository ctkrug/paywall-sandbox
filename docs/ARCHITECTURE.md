# Architecture

A concise map of the codebase for anyone (human or model) picking this up
cold. See [`VISION.md`](VISION.md) for why, [`PROTOCOL.md`](PROTOCOL.md) for
the wire format, [`BACKLOG.md`](BACKLOG.md) for what's left.

## Modules

```
cmd/paywall-sandbox/    CLI entrypoint: subcommand dispatch + flag parsing
internal/paywall/       Wire types shared by both sides: Descriptor, Proof
internal/mockserver/    The mock 402 server: Rule matching, Server, logging
```

- **`internal/paywall`** — no dependencies beyond stdlib. Defines
  `Descriptor` (what's owed) and `Proof` (what satisfies it), their JSON
  encode/decode, and the two header names (`X-Payment-Required`,
  `X-Payment`) both sides key off. Anything that needs to speak the wire
  format imports this package; it imports nothing else in this repo.
- **`internal/mockserver`** — the server side. `Rule` describes one
  protected route (method/path/price); `Server` is an `http.Handler` that
  challenges unmatched-proof requests and forwards matched-proof ones to
  `Next`. Nonces are issued and consumed in an in-memory map guarded by a
  mutex — one-time use, hard TTL, no persistence (see `docs/PROTOCOL.md` for
  why). `LogRequests` is a middleware wrapping any handler with structured
  request logging.
- **`cmd/paywall-sandbox`** — thin CLI wrapper. `serve` stands up a
  `mockserver.Server` with a single rule from flags; `version` prints the
  build version.

## Data flow (server side)

```
request → Server.ServeHTTP
            → matchRule (Rule.Matches by method + path)
            → no match:            forward to Next
            → match, no/invalid proof:  issue Descriptor, 402
            → match, valid proof:  consume nonce, forward to Next
```

## Build / run / test

```
go build -o bin/paywall-sandbox ./cmd/paywall-sandbox
./bin/paywall-sandbox serve --path /paid --amount 100 --asset USDC
go test ./...
```

`make build|test|run|fmt|vet|lint` wrap the above; see `Makefile`.

## Where new work plugs in

- New proof schemes: keep `paywall.Proof.Scheme` as the discriminator;
  `mockserver.Server.acceptProof` is the only place that currently
  hardcodes `FakeScheme` — see `docs/BACKLOG.md`'s pluggable-signer story.
- New rule sources (config file, wildcard paths): `internal/mockserver`,
  extending `Rule`/`Rule.Matches` and adding a loader, not touching
  `Server`.
- A CLI client that drives the loop from the other side: a new
  `internal/client`-shaped package plus a `request` subcommand — see
  `docs/BACKLOG.md`'s CLI client epic.
