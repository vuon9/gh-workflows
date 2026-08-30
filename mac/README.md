# macOS Release

Reusable workflow [`macos-release.yml`](../.github/workflows/macos-release.yml) that signs, notarizes, staples, packages, and publishes a Developer ID macOS DMG after an app-specific build job uploads a `.app` bundle archive.

## What it does

- Downloads an app bundle archive uploaded by a caller-owned build job.
- Installs Go and `create-dmg`.
- Imports the Developer ID Application certificate into a temporary keychain.
- Signs the app with hardened runtime and timestamping.
- Notarizes and staples the app.
- Creates, signs, notarizes, staples, and verifies the DMG.
- Uploads the DMG as a GitHub Actions artifact and, on tags, a GitHub Release asset.

## Usage

```yaml
name: macOS Release

on:
  push:
    tags:
      - 'macos/myapp/v*'
  workflow_dispatch:

jobs:
  build:
    runs-on: macos-26
    steps:
      - uses: actions/checkout@v6
      - run: task darwin:package:universal
      - run: tar -czf "$RUNNER_TEMP/macos-app.tar.gz" -C bin MyApp.app
      - uses: actions/upload-artifact@v7.0.1
        with:
          name: myapp-macos-app-${{ github.run_id }}
          path: ${{ runner.temp }}/macos-app.tar.gz

  release:
    needs: build
    uses: vuon9/gh-workflows/.github/workflows/macos-release.yml@v0.2.0
    with:
      app-name: MyApp
      team-id: ABCDE12345
      app-path: bin/MyApp.app
      app-artifact-name: myapp-macos-app-${{ github.run_id }}
      dmg-name: MyApp-macos-universal.dmg
      artifact-name: myapp-macos-release-${{ github.run_id }}
      runner-label: macos-26
      github-release-prerelease: ${{ contains(github.ref_name, '-') }}
    secrets: inherit
```

## Secrets

Required:

- `APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_P12_BASE64`: base64-encoded Developer ID Application `.p12`.
- `APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_PASSWORD`: `.p12` password.
- `APP_STORE_CONNECT_API_KEY_P8`: App Store Connect API private key content.
- `APP_STORE_CONNECT_API_KEY_ID`: App Store Connect API key ID.
- `APP_STORE_CONNECT_API_ISSUER_ID`: App Store Connect issuer ID.

Optional:

- `MACOS_CODESIGN_IDENTITY`: full `Developer ID Application: ... (TEAMID)` identity. If omitted, the workflow finds the imported Developer ID Application identity for `team-id`.

## Inputs

- `app-name` (required): display name of the macOS app bundle.
- `team-id` (required): Apple Developer Team ID.
- `app-path` (required): path where the app bundle should be extracted, relative to `working-directory`.
- `app-artifact-name` (required): name of the uploaded `.app` archive artifact from the caller build job.
- `dmg-name` (required): DMG file name to create.
- `working-directory`: caller repository working directory. Defaults to `.`.
- `app-archive-name`: file name of the `.tar.gz` archive inside `app-artifact-name`. Defaults to `macos-app.tar.gz`.
- `artifact-name`: GitHub Actions artifact name for the signed DMG. Defaults to `macos-release`.
- `runner-label`: default `macos-26`.
- `upload-github-release`: upload the signed DMG to a GitHub Release on tag builds. Defaults to `true`.
- `github-release-prerelease`: mark the GitHub Release as a prerelease on tag builds. Defaults to `false`. Set `true` for prerelease tags such as `v1.0.0-rc.1`.

## Recommended caller controls

- Use a product/platform tag namespace such as `macos/myapp/v1.0.0`.
- Reference a version tag after the workflow is released, not `@main`.
- Keep Apple secrets in the app repository or a protected GitHub Environment.
- Keep app-specific build systems such as Wails, Xcode, Electron, or custom scripts in the caller repository.
- Run the first release manually with `workflow_dispatch` before cutting the public tag.
- Do final Gatekeeper verification by downloading the uploaded DMG on a clean macOS machine.
