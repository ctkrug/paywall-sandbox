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

- [ ] Load rule sets from a JSON config file instead of only CLI flags
- [ ] Support path-prefix/wildcard route matching (today: exact path only)
- [ ] Add an example config file and docs for authoring rule sets

## Epic: additional proof schemes

- [ ] Define a pluggable `Signer`/`Verifier` interface so schemes beyond
      `fake` can be registered without touching `mockserver.Server`
- [ ] Add a second, HMAC-based local proof scheme as a real (if still
      offline) settlement simulation
- [ ] Document how to add a new scheme in `PROTOCOL.md`

## Epic: scenario scripting & polish

- [ ] Design a declarative scenario file format describing an expected
      challenge/response sequence
- [ ] Add a `paywall-sandbox test <scenario>` command that runs a scenario
      and exits non-zero on any assertion failure
- [ ] Check in an example scenario file runnable from CI
- [ ] Expand the README with full CLI usage examples once `request` and
      `test` subcommands exist
- [ ] Add a release workflow that builds and publishes versioned binaries
