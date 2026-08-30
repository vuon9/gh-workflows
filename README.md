# gh-workflows

Reusable GitHub Actions workflows and release tooling.

The goal is to keep product repositories clean: app repos keep small workflow wrappers and app-specific config, while common CI/release behavior lives here.

## Workflows

### `homebrew-cask-update.yml`

Caller repositories can update a cask in a separate Homebrew tap after a release
asset exists. The workflow resolves the release asset, downloads it, calculates
the real SHA256, updates the cask file, runs Homebrew validation, and opens a PR
to the tap repository by default.

```yaml
name: Update Homebrew Tap

on:
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      release-tag:
        type: string
        required: true

jobs:
  update-homebrew:
    uses: vuon9/gh-workflows/.github/workflows/homebrew-cask-update.yml@v0.2.0
    with:
      tap-repository: owner/homebrew-tap
      cask-token: myapp
      release-tag: ${{ github.event.release.tag_name || inputs.release-tag }}
      artifact-name-regex: "MyApp-.*\\.dmg$"
      open-pull-request: true
    secrets:
      TAP_REPO_TOKEN: ${{ secrets.TAP_REPO_TOKEN }}
```

Required caller secret:

- `TAP_REPO_TOKEN`: fine-grained token or GitHub App token with write access to
  the tap repository. The caller repository `GITHUB_TOKEN` is not enough when
  the tap lives in a different repository.

Useful inputs:

- `tap-repository`: tap repo such as `owner/homebrew-tap`.
- `cask-token`: cask token such as `myapp`.
- `cask-path`: cask path inside the tap repo. Defaults to `Casks/<cask-token>.rb`.
- `source-repository`: release repo. Defaults to the caller repository.
- `release-tag`: release tag. Defaults to the caller ref name.
- `artifact-name-regex`: must match exactly one release asset when `artifact-url`
  is omitted.
- `artifact-url`: explicit release artifact URL when the caller already knows it.
- `version`: cask version. Defaults to the release tag basename with a leading
  `v` removed.
- `open-pull-request`: defaults to `true`. Set to `false` only when direct tap
  pushes are acceptable.
- `dry-run`: updates and validates locally without pushing.

### `ai-code-review.yml`

Reusable AI code review for pull requests using OpenCode (`opencode github run`)
with the runner `GITHUB_TOKEN`, so no GitHub App install is required. Every
non-draft PR gets a review comment automatically; on-demand review can be
triggered by commenting `/oc` or `/opencode` on a PR.

```yaml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]

concurrency:
  group: ai-code-review-${{ github.event.pull_request.number || github.event.issue.number || github.ref }}
  cancel-in-progress: true

jobs:
  review:
    permissions:
      contents: read
      pull-requests: write
      issues: write
    uses: vuon9/gh-workflows/.github/workflows/ai-code-review-reusable.yml@v0.3.0
    with:
      mode: pr
    secrets:
      OPENCODE_API_KEY: ${{ secrets.OPENCODE_API_KEY }}
```

Required caller secret:

- `OPENCODE_API_KEY`: OpenCode API key (from `opencode auth login` on your
  machine, or the OpenCode dashboard).

Useful inputs:

- `model`: defaults to `opencode/muse-spark-1.2-contributor-free`. The
  non-free `opencode/muse-spark-1.2-contributor` requires workspace credits.
- `prompt`: custom review prompt. Defaults to a thorough review checklist.
- `mode`: `pr` (auto-review non-draft PRs) or `comment` (on-demand `/oc`).

Notes:

- `permissions` MUST live on the caller job, not inside the reusable job:
  a `permissions` block inside a `workflow_call` job causes `startup_failure`.
- `if:` conditions are not allowed on caller jobs that `uses:` a reusable
  workflow; the `mode` input handles gating inside the reusable workflow.
- The reusable workflow is at `.github/workflows/ai-code-review-reusable.yml`
  and can also be called directly by other repositories.

### `macos-release.yml`

Caller repositories can sign, notarize, staple, package, and publish a Developer ID
macOS DMG after an app-specific build job uploads a `.app` bundle archive.

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

Required caller secrets:

- `APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_P12_BASE64`: base64-encoded Developer ID Application `.p12`.
- `APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_PASSWORD`: `.p12` password.
- `APP_STORE_CONNECT_API_KEY_P8`: App Store Connect API private key content.
- `APP_STORE_CONNECT_API_KEY_ID`: App Store Connect API key ID.
- `APP_STORE_CONNECT_API_ISSUER_ID`: App Store Connect issuer ID.

Optional caller secret:

- `MACOS_CODESIGN_IDENTITY`: full `Developer ID Application: ... (TEAMID)` identity. If omitted, the workflow finds the imported Developer ID Application identity for `team-id`.

What the workflow does:

- Downloads an app bundle archive uploaded by a caller-owned build job.
- Installs Go and `create-dmg`.
- Imports the Developer ID Application certificate into a temporary keychain.
- Signs the app with hardened runtime and timestamping.
- Notarizes and staples the app.
- Creates, signs, notarizes, staples, and verifies the DMG.
- Uploads the DMG as a GitHub Actions artifact and, on tags, a GitHub Release asset.
  Set `github-release-prerelease: true` for prerelease tags such as `v1.0.0-rc.1`.

Recommended caller controls:

- Use a product/platform tag namespace such as `macos/myapp/v1.0.0`.
- Reference a version tag after the workflow is released, not `@main`.
- Keep Apple secrets in the app repository or a protected GitHub Environment.
- Keep app-specific build systems such as Wails, Xcode, Electron, or custom scripts in the caller repository.
- Run the first release manually with `workflow_dispatch` before cutting the public tag.
- Do final Gatekeeper verification by downloading the uploaded DMG on a clean macOS machine.

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

### Preparing Manual Signing Secrets

Use manual signing when a clean GitHub-hosted macOS runner should sign without asking Xcode to create certificates or profiles.

1. Export an Apple Distribution certificate from a trusted Mac:

```bash
security find-identity -v -p codesigning
```

Pick the `Apple Distribution: ... (TEAMID)` identity, then export it to a temporary `.p12`. This may show a macOS Keychain approval prompt:

```bash
P12_PASSWORD="$(openssl rand -base64 24)"
security export \
  -k "$HOME/Library/Keychains/login.keychain-db" \
  -t identities \
  -f pkcs12 \
  -o /tmp/apple-distribution.p12 \
  -P "$P12_PASSWORD"
base64 -i /tmp/apple-distribution.p12 -o /tmp/apple-distribution.p12.b64
```

2. Base64-encode the App Store provisioning profile for the app:

```bash
base64 -i "$HOME/Library/Developer/Xcode/UserData/Provisioning Profiles/<UUID>.mobileprovision" \
  -o /tmp/app-store-profile.mobileprovision.b64
```

3. Set repository secrets:

```bash
gh secret set APPLE_DISTRIBUTION_CERTIFICATE_P12_BASE64 \
  --repo owner/app-repo \
  < /tmp/apple-distribution.p12.b64

printf '%s' "$P12_PASSWORD" | gh secret set APPLE_DISTRIBUTION_CERTIFICATE_PASSWORD \
  --repo owner/app-repo

gh secret set APPLE_PROVISIONING_PROFILE_BASE64 \
  --repo owner/app-repo \
  < /tmp/app-store-profile.mobileprovision.b64
```

4. Delete local temporary signing files:

```bash
rm -f /tmp/apple-distribution.p12 /tmp/apple-distribution.p12.b64 /tmp/app-store-profile.mobileprovision.b64
unset P12_PASSWORD
```

Never commit `.p12`, `.mobileprovision`, `.p8`, archives, IPAs, or base64 secret files.

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
