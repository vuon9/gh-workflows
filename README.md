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

Optional caller secrets when `signing-style: manual`:

- `APPLE_DISTRIBUTION_CERTIFICATE_P12_BASE64`: base64-encoded Apple Distribution `.p12`.
- `APPLE_DISTRIBUTION_CERTIFICATE_PASSWORD`: `.p12` password.
- `APPLE_PROVISIONING_PROFILE_BASE64`: base64-encoded App Store provisioning profile.

Useful inputs:

- `dry-run`: print commands without running archive/export.
- `skip-tests`: skip simulator tests before archive.
- `skip-archive`: skip archive and export an archive downloaded from a previous release artifact.
- `export-destination`: `upload` for TestFlight/App Store Connect upload, `export` for a reusable IPA/archive artifact.
- `upload-artifact-name`: upload `build/TestFlight` as a GitHub Actions artifact after archive/export.
- `download-artifact-name`: download a prior `build/TestFlight` artifact before export.
- `skip-cert-check`: skip the local Apple Distribution identity preflight check. This is useful when testing whether `xcodebuild -allowProvisioningUpdates` can handle signing on a clean GitHub-hosted macOS runner.
- `runner-label`: runner used for the release job. Defaults to `macos-26` because App Store Connect requires the iOS 26 SDK or newer for uploads.
- `signing-style`: defaults to `automatic`; use `manual` with `bundle-id` and `provisioning-profile` when running on clean hosted runners.
- `bundle-id` / `provisioning-profile`: required with manual signing so archive/export can use an installed profile instead of creating signing assets.

Recommended caller controls:

- Protect the workflow with a GitHub Environment such as `testflight`.
- Use a tag-triggered workflow named `Release` to build and export the release artifact.
- Gate the dependent TestFlight upload through a regular approval job with a GitHub Environment such as `testflight`; configure required reviewers on that environment so the upload is actually manual.
- Reference a version tag such as `@v0.1.3` for pilots or `@v1` after the workflow is proven, not `@main`.

Recommended release wrapper shape:

```yaml
name: Release

on:
  push:
    tags:
      - 'ios/myapp/v*'

jobs:
  release:
    uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v0.1.6
    with:
      project-path: MyApp.xcodeproj
      scheme: MyApp
      team-id: ABCDE12345
      export-destination: export
      upload-artifact-name: myapp-ios-release-${{ github.run_id }}
      runner-label: macos-26
    secrets: inherit

  approve-testflight:
    needs: release
    runs-on: ubuntu-latest
    environment: testflight
    steps:
      - run: echo "TestFlight upload approved for $GITHUB_REF_NAME"

  testflight:
    needs: approve-testflight
    uses: vuon9/gh-workflows/.github/workflows/ios-testflight.yml@v0.1.6
    with:
      project-path: MyApp.xcodeproj
      scheme: MyApp
      team-id: ABCDE12345
      skip-tests: true
      skip-archive: true
      archive-path: build/TestFlight/MyApp.xcarchive
      download-artifact-name: myapp-ios-release-${{ github.run_id }}
      runner-label: macos-26
    secrets: inherit
```

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
