# gh-workflows

Shared GitHub Actions workflows and release tooling for Vuong's projects.

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
    uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v0.1.1
    with:
      project-path: Spotto.xcodeproj
      scheme: Spotto
      team-id: 256XRVYZ9V
      dry-run: true
    secrets: inherit
```

For workspace-based apps, use `workspace-path` instead of `project-path`.

Required caller secrets when `dry-run` is `false`:

- `APP_STORE_CONNECT_API_KEY_P8`: App Store Connect API private key content.
- `APP_STORE_CONNECT_API_KEY_ID`: App Store Connect API key ID.
- `APP_STORE_CONNECT_API_ISSUER_ID`: App Store Connect issuer ID.

Recommended caller controls:

- Protect the workflow with a GitHub Environment such as `testflight`.
- Run from `workflow_dispatch` first, then add branch/tag triggers after a successful pilot.
- Reference a version tag such as `@v0.1.1` for pilots or `@v1` after the workflow is proven, not `@main`.

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
  --project Spotto.xcodeproj \
  --scheme Spotto \
  --team-id 256XRVYZ9V \
  --skip-tests \
  --dry-run
```

## Versioning

Use the pilot tag while validating the first app migration:

```yaml
uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v0.1.1
```

After one real TestFlight upload succeeds, create a major tag for the stable workflow contract:

```bash
git tag v1
git push origin v1
```

Breaking changes should use a new major tag. Additive inputs can stay under the current major tag after testing.
