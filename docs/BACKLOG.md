# Backlog

High-level epic/story breakdown for the v1 build. See
[`VISION.md`](VISION.md) for the why and [`PROTOCOL.md`](PROTOCOL.md) for
the wire format these stories implement against.

## Epic: CLI client (challenge/pay/retry)

- [x] Add a `fake` payment proof signer that builds a `Proof` from a
      received `Descriptor`
- [x] Implement the client-side challenge → pay → retry loop as a reusable
      package (initial request, detect 402, build proof, retry once)
- [x] Wire a `request` subcommand into the CLI so the loop can be driven
      against any URL, mock or real
- [x] Add `--verbose` inspection output to the request loop (print every
      header/descriptor/proof exchanged)

## Epic: configurable rule sets

- [x] Load rule sets from a JSON config file instead of only CLI flags
- [x] Support path-prefix/wildcard route matching (today: exact path only)
- [x] Add an example config file and docs for authoring rule sets

## Epic: additional proof schemes

- [x] Define a pluggable `Signer`/`Verifier` interface so schemes beyond
      `fake` can be registered without touching `mockserver.Server`
- [x] Add a second, HMAC-based local proof scheme as a real (if still
      offline) settlement simulation
- [x] Document how to add a new scheme in `PROTOCOL.md`

## Epic: scenario scripting & polish

- [x] Design a declarative scenario file format describing an expected
      challenge/response sequence
- [x] Add a `paywall-sandbox test <scenario>` command that runs a scenario
      and exits non-zero on any assertion failure
- [x] Check in an example scenario file runnable from CI
- [x] Expand the README with full CLI usage examples once `request` and
      `test` subcommands exist
- [x] Add a release workflow that builds and publishes versioned binaries
