# The 402 challenge/response wire format

This document is the sandbox's own spec for the flow it mocks. HTTP 402
Payment Required has no single ratified body format — the ecosystem (x402
and similar "pay-per-call" proposals) is converging on a shape rather than a
standard. This is the shape `paywall-sandbox` implements and tests against.

## The exchange

1. **Client** sends a normal request to a protected route.
2. **Server** has no proof of payment for it, so it responds:
   - Status: `402 Payment Required`
   - Header `X-Payment-Required`: the JSON-encoded payment descriptor
     (mirrored into the body for clients that don't want to parse headers).
3. **Client** settles the described payment out of band (a real rail, or —
   in this sandbox — the `fake` scheme) and constructs a proof.
4. **Client** retries the *same* request with header `X-Payment: <proof>`.
5. **Server** validates the proof against the descriptor it issued and, if
   valid, forwards the request as if payment had never been required.

## Payment descriptor

```json
{
  "amount": 100,
  "asset": "USDC",
  "recipient": "0xsandbox",
  "nonce": "6e8a301d7cfe2c717d3458955d73c044",
  "expiresAt": "2026-07-03T08:39:10.785509644Z"
}
```

| Field       | Meaning                                                        |
|-------------|-----------------------------------------------------------------|
| `amount`    | Price in the smallest unit of `asset` (cents, base token unit). |
| `asset`     | Currency or token identifier.                                   |
| `recipient` | Address/account the payment must settle to.                     |
| `nonce`     | Unique per challenge; binds a proof to the exchange that issued it and prevents replay. |
| `expiresAt` | RFC 3339 timestamp; a proof presented after this is rejected.   |

## Proof of payment

```json
{
  "nonce": "6e8a301d7cfe2c717d3458955d73c044",
  "scheme": "fake",
  "signature": "..."
}
```

`nonce` must match a descriptor the server actually issued and not yet
consumed. `scheme` identifies how `signature` should be interpreted and is
checked by a pluggable `Verifier` (see "Adding a proof scheme" below).
This sandbox ships two: `fake`, which the server accepts unconditionally
once the nonce checks out, and `hmac-sha256`, which checks `signature`
against a shared secret. Both are explicitly **not** real payment
mechanisms and prove nothing about actual settlement — they exist to
exercise the protocol shape end to end. Real schemes (on-chain transfer
receipts, signed settlement attestations, etc.) are future backlog work —
see [`BACKLOG.md`](BACKLOG.md).

## Rule config file

`serve --config <file>` loads the server's protected routes from a JSON
file instead of the single `--path`/`--amount`/`--asset`/`--recipient`
rule, so a project can check a mock server's route shape into its own
repo alongside its tests:

```json
{
  "rules": [
    {"path": "/paid", "amount": 100, "asset": "USDC", "recipient": "0xsandbox"},
    {"method": "POST", "path": "/api/premium/*", "amount": 500, "asset": "USD", "recipient": "acct-premium"}
  ]
}
```

| Field       | Meaning                                                          |
|-------------|-------------------------------------------------------------------|
| `method`    | Optional; restricts the rule to one HTTP method. Omitted matches any. |
| `path`      | Exact request path, or a prefix ending in `/*` (matches the prefix itself and anything nested under it). |
| `amount`    | Price in the smallest unit of `asset`. Must be greater than zero. |
| `asset`     | Currency or token identifier. Must not be empty.                 |
| `recipient` | Address/account the payment must settle to. Must not be empty.   |

See [`examples/rules.json`](../examples/rules.json) for a runnable example.

## Scenario files

`test <scenario.json>` (`internal/scenario`) runs a declarative script
against an in-process server it starts and tears down itself — no
already-running `serve` required, so a scenario is a single self-contained
CI assertion:

```json
{
  "name": "paid route settles, free route does not",
  "hmacKey": "shared-secret",
  "rules": [
    {"path": "/paid", "amount": 100, "asset": "USDC", "recipient": "0xsandbox"}
  ],
  "steps": [
    {
      "name": "GET /paid settles with the fake scheme",
      "method": "GET",
      "path": "/paid",
      "scheme": "fake",
      "expect": {"paid": true, "finalStatus": 200}
    }
  ]
}
```

| Field           | Meaning                                                        |
|-----------------|-------------------------------------------------------------------|
| `name`          | Identifies the scenario in `test` output.                        |
| `rules`         | Same shape as a `serve --config` rule set (see above); the server the scenario runs against is configured from these. |
| `hmacKey`       | Optional shared secret; required if any step uses `hmac-sha256`.  |
| `steps`         | The requests to make, in order.                                  |
| `steps[].name`      | Identifies the step in `test` output.                          |
| `steps[].method`    | HTTP method. Defaults to `GET`.                                |
| `steps[].path`      | Request path, relative to the scenario server's root.          |
| `steps[].scheme`    | Proof scheme to settle a challenge with, if one is issued. Defaults to `fake`. |
| `steps[].expect.paid`        | Must match whether the client attempted to settle a challenge (`client.Result.Paid`) — true whenever a 402 was issued, regardless of whether the proof was ultimately accepted. |
| `steps[].expect.finalStatus` | Must match the status code of the step's last response (`client.Result.FinalStatusCode`). Required. |

Because `paid` and `finalStatus` are independent, a scenario can assert on
rejection paths too — e.g. a step using the wrong `hmac-sha256` key expects
`paid: true` (a retry was attempted) but `finalStatus: 402` (the server
rejected the mismatched signature).

`test` exits `0` only if every step's actual outcome matches its `expect`;
otherwise it prints a `[FAIL]` line with the mismatch per step and exits
`1`. See [`examples/scenario.json`](../examples/scenario.json) and
[`examples/scenario-hmac.json`](../examples/scenario-hmac.json) for
runnable examples.

## Adding a proof scheme

`nonce`/expiry replay protection is scheme-agnostic; a scheme only needs to
define how `signature` is produced and checked. To add one:

1. **Server side** (`internal/mockserver`): implement `Verifier` —
   `Verify(paywall.Descriptor, paywall.Proof) error` — and register it
   under your scheme name in `Server.Verifiers`. `HMACVerifier`
   (`hmac_verifier.go`) is a worked example: it checks `Signature` against
   `hex(HMAC-SHA256(Key, Descriptor.Nonce))`.
2. **Client side** (`internal/client`): implement `Signer` —
   `Sign(paywall.Descriptor) (paywall.Proof, error)` — that builds a
   matching proof. `HMACSigner` mirrors `HMACVerifier`'s computation.
3. Wire the new `Signer`/scheme into the `request` subcommand's
   `resolveSigner` (`cmd/paywall-sandbox/request.go`) if it should be
   selectable via `--scheme`, and into `internal/scenario`'s `signerFor` if
   scenario steps should be able to select it too.

`FakeScheme` and `HMACScheme` are each defined independently in both
packages (not shared via import) so `internal/client` stays decoupled from
`internal/mockserver` — it can drive any target that speaks a given
scheme, mock or real.

## Design choices and open questions

- **One-time nonces.** A nonce is consumed on first successful use, so a
  captured `X-Payment` header can't be replayed against a second request.
  This mirrors how most real settlement proofs work (a receipt is tied to
  one transfer) but means a client must request a fresh challenge per call
  — no descriptor reuse across requests.
- **Header vs. body.** Real deployments disagree on whether the descriptor
  belongs in a header, the body, or both. This sandbox does both so it can
  be a useful test target either way a client chooses to read it.
- **No expiry grace period.** `expiresAt` is a hard cutoff. Real rails may
  want clock-skew tolerance; that's left to a future scheme rather than
  baked into the core challenge/response contract.
