# Security

## Scope

Paywall Sandbox is a **development and testing tool**. It is not a payment
processor and settles nothing for real. The only proof scheme it ships,
`fake` (see [`docs/PROTOCOL.md`](docs/PROTOCOL.md)), is accepted once its
nonce checks out — it does not verify that any value actually changed hands.

**Do not run `paywall-sandbox serve` on a publicly reachable address and
treat a `200` response as evidence of real payment.** It is meant to run
locally or in CI, standing in for a real settlement-backed 402 server while
you develop the rest of your integration.

## Reporting a vulnerability

If you find a bug in the challenge/response logic itself (e.g. a way to
bypass nonce/expiry checks that would also apply to a real proof scheme
built on this project's interfaces), please open an issue describing it.
