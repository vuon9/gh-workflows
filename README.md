# gh-workflows

Reusable GitHub Actions workflows and release tooling.

The goal is to keep product repositories clean: app repos keep small workflow wrappers and app-specific config, while common CI/release behavior lives here.

## Workflows

### `ios-testflight.yml`

Caller repositories can trigger TestFlight upload with a thin wrapper:

```yaml
name: TestFlight

on:
  workflow_dispatch:

jobs:
  testflight:
    uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v0.1.5
    with:
      project-path: MyApp.xcodeproj
      scheme: MyApp
      team-id: ABCDE12345
      dry-run: true
      skip-cert-check: false
      runner-label: macos-26
    secrets: inherit
```

For workspace-based apps, use `workspace-path` instead of `project-path`.

Required caller secrets when `dry-run` is `false`:

- `APP_STORE_CONNECT_API_KEY_P8`: App Store Connect API private key content.
- `APP_STORE_CONNECT_API_KEY_ID`: App Store Connect API key ID.
- `APP_STORE_CONNECT_API_ISSUER_ID`: App Store Connect issuer ID.

Useful inputs:

- `dry-run`: print commands without running archive/export.
- `skip-tests`: skip simulator tests before archive.
- `skip-cert-check`: skip the local Apple Distribution identity preflight check. This is useful when testing whether `xcodebuild -allowProvisioningUpdates` can handle signing on a clean GitHub-hosted macOS runner.
- `runner-label`: runner used for the release job. Defaults to `macos-26` because App Store Connect requires the iOS 26 SDK or newer for uploads.

Recommended caller controls:

- Protect the workflow with a GitHub Environment such as `testflight`.
- Run from `workflow_dispatch` first, then add branch/tag triggers after a successful pilot.
- Reference a version tag such as `@v0.1.3` for pilots or `@v1` after the workflow is proven, not `@main`.

## Local Development

Run the Go tests:

```bash
go test ./...
```

Build the iOS TestFlight CLI:

```bash
go build ./ios/testflight/cmd/ios-testflight
```

Example dry run from an iOS app repository:

```bash
go run /path/to/gh-workflows/ios/testflight/cmd/ios-testflight \
  --project MyApp.xcodeproj \
  --scheme MyApp \
  --team-id ABCDE12345 \
  --skip-tests \
  --dry-run
```

## Local Preflight With `act`

[`act`](https://github.com/nektos/act) can catch basic workflow wiring problems before pushing:

```bash
act workflow_dispatch \
  --validate \
  -W .github/workflows/testflight.yml \
  --input dry-run=true \
  --input skip-tests=true
```

It can also dry-run the job graph:

```bash
act workflow_dispatch \
  --dryrun \
  -W .github/workflows/testflight.yml \
  --input dry-run=true \
  --input skip-tests=true
```

Use `act` as a preflight only. iOS release workflows still need a real GitHub-hosted macOS runner to verify Xcode, signing, Keychain behavior, and App Store Connect upload.

## Versioning

Use the pilot tag while validating the first app migration:

```yaml
uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v0.1.5
```

After one real TestFlight upload succeeds, create a major tag for the stable workflow contract:

```bash
git tag v1
git push origin v1
```

Breaking changes should use a new major tag. Additive inputs can stay under the current major tag after testing.
