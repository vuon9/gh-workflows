# gh-workflows

Shared GitHub Actions workflows and release tooling for Vuong's projects.

The goal is to keep product repositories clean: app repos keep small workflow wrappers and app-specific config, while common CI/release behavior lives here.

## Layout

```text
gh-workflows/
  .github/workflows/
    ios-testflight.yml
  ios/
    testflight/
      cmd/ios-testflight/
      internal/exportoptions/
      internal/xcode/
      README.md
```

GitHub requires reusable workflow files to live in `.github/workflows/`. Domain-specific implementation and docs live under folders such as `ios/`, `web/`, and `backend/`.

## iOS TestFlight Reusable Workflow

Caller repositories can trigger TestFlight upload with a thin wrapper:

```yaml
name: TestFlight

on:
  workflow_dispatch:

jobs:
  testflight:
    uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v1
    with:
      project-path: Spotto.xcodeproj
      scheme: Spotto
      team-id: 256XRVYZ9V
    secrets: inherit
```

For workspace-based apps, use `workspace-path` instead of `project-path`.

Required caller secrets:

- `APP_STORE_CONNECT_API_KEY_P8`: App Store Connect API private key content.
- `APP_STORE_CONNECT_API_KEY_ID`: App Store Connect API key ID.
- `APP_STORE_CONNECT_API_ISSUER_ID`: App Store Connect issuer ID.

Optional caller secret:

- `GH_WORKFLOWS_READ_TOKEN`: token with read access to this repository. This is only needed if GitHub's default token cannot check out a private `gh-workflows` repository from the caller workflow run.

Recommended caller controls:

- Protect the workflow with a GitHub Environment such as `testflight`.
- Run from `workflow_dispatch` first, then add branch/tag triggers after a successful pilot.
- Reference a version tag such as `@v1`, not `@main`.

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

Use major tags for stable workflow contracts:

```bash
git tag v1
git push origin v1
```

Breaking changes should use a new major tag. Additive inputs can stay under the current major tag after testing.
