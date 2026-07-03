# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Initial repo scaffold: `paywall` descriptor/proof types, `mockserver`
  402 challenge/response handler, and a `serve` CLI subcommand.
- `internal/client` package implementing the challenge → pay → retry
  loop, and a `request` CLI subcommand (with `--verbose` tracing) that
  drives it against any target, mock or real.
- Configurable rule sets: `serve --config <file>` loads protected routes
  from a JSON file, and `Rule.Path` supports `/*` prefix matching in
  addition to exact paths.
- Pluggable proof schemes: `mockserver.Verifier` / `client.Signer`
  interfaces, plus a second `hmac-sha256` scheme alongside `fake`.
- Scenario scripting: a declarative JSON scenario format (`internal/scenario`)
  and a `test <scenario>` CLI subcommand that runs a sequence of requests
  against an in-process server and asserts on the outcome, exiting non-zero
  on any failure so it doubles as a CI check.
