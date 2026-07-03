# Releasing

Releases are cut from a git tag, not a branch or a manual build step.

1. Make sure `main` is green (CI passing) and `CHANGELOG.md` has an entry
   for the version you're about to cut.
2. Tag the release commit and push the tag:

   ```console
   $ git tag v0.1.0
   $ git push origin v0.1.0
   ```

3. The [`release` workflow](../.github/workflows/release.yml) picks up the
   `v*` tag, runs [GoReleaser](https://goreleaser.com) against
   [`.goreleaser.yaml`](../.goreleaser.yaml), and publishes a GitHub Release
   with:
   - `linux`/`darwin`/`windows` binaries for `amd64`/`arm64`
   - a `checksums.txt` for the archives
   - release notes generated from the commits since the previous tag

No manual cross-compilation or upload step is needed — pushing the tag is
the entire release.
