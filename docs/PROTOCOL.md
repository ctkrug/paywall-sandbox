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
consumed. `scheme` identifies how `signature` should be interpreted —
this sandbox ships exactly one scheme, `fake`, which the server accepts
unconditionally once the nonce checks out. It exists to exercise the
protocol shape end to end; it is **not** a real payment mechanism and
proves nothing about actual settlement. Real schemes (on-chain transfer
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
