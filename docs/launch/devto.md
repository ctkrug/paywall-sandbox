---
title: "Tollgate: testing HTTP 402 payment flows without real money"
published: false
tags: go, http, payments, testing
---

HTTP 402 Payment Required has been sitting in the HTTP spec, reserved and mostly
unused, for about thirty years. Lately it has started to matter: a wave of
pay-per-call APIs, led by the x402 pattern, uses it for a specific handshake. A
server answers a request with `402` and a payment descriptor. The client, or an
AI agent acting for it, settles payment out of band and retries with proof
attached. The server checks the proof and serves the resource.

I wanted to build against that flow, but the tooling is thin. There is no
dominant SDK, and the wire format is still being pieced together from scattered
specs. Testing my handling meant standing up real settlement: a wallet, a
network, live funds, and no clean way to script the edge cases I actually cared
about. So I built **Tollgate**, a local mock server and CLI that models the
challenge to settle to retry loop with no money involved. Here are two decisions
that shaped it.

## One client loop, two callers

The interesting logic is not the server, it is the client walk through the
handshake: send the request, and if you get a 402, parse the descriptor, build a
proof, retry once with the proof attached. I put that in a single reusable type,
`client.Loop`, that returns both the final response and a step-by-step trace of
what it did.

That paid off immediately, because two very different features are just callers
of the same loop. The `request` subcommand runs it once against any URL, mock or
real, and prints the trace when you pass `--verbose`. The scenario runner runs it
many times against an in-process server and asserts on the outcome of each step.
Neither one re-implements the handshake. When I later added a `--timeout` to
`request`, it was a `context.WithTimeout` at the call site; the loop already took
a context, so nothing inside it changed. If the loop had lived inside the CLI
command, both features would have grown their own copy and drifted.

## Making replay and memory leaks fail loudly

A mock is only useful if it fails the way the real thing should. Two details
mattered more than I expected.

The first is replay. When the server accepts a valid proof, it deletes that
nonce from its issued map in the same locked section that verified it. Present the
same proof twice and the second attempt finds no matching nonce and gets
challenged again. Getting single-use right in the mock means your client cannot
accidentally pass a test suite while depending on replay working, which a real
settlement layer would never allow.

The second is memory. Every challenge stores a nonce so a later proof can be
matched to it, but most challenges are never retried: abandoned clients, ignored
expiries, agents that give up. Left alone, that map grows forever. So the server
sweeps expired entries each time it stores a new one. It is a small thing, but it
is the difference between a toy that leaks under load and something you can leave
running in CI. I wrote a concurrency test that hammers the issued map from many
goroutines to keep that path honest.

## What I would do differently

The proof schemes today are `fake`, which accepts any signature once the nonce and
expiry check out, and `hmac-sha256`, a shared-secret signature that is one step
closer to real evidence. Neither is byte-compatible with a specific production
x402 profile yet, mostly because those profiles are still moving. If I picked this
back up, I would pin one real descriptor and proof format and ship a verifier that
matches it exactly, so the sandbox doubles as a conformance check. I would also
add a record mode that captures a real origin's 402 responses and replays them
locally, so you could develop your client against a captured real challenge with
the origin offline.

It is a single static Go binary with no runtime dependencies, MIT licensed.

- Repo: https://github.com/ctkrug/paywall-sandbox
- Live page: https://apps.charliekrug.com/paywall-sandbox/

If you are building on 402 or x402, I would genuinely like to know which parts of
the handshake you found hardest to test.
