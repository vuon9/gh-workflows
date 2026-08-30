# iOS TestFlight

Reusable TestFlight upload tooling for iOS apps.

## What It Does

The reusable workflow:

1. Checks out the caller app repository.
2. Checks out this `gh-workflows` repository at the same commit as the called workflow.
3. Builds the Go CLI.
4. Writes the App Store Connect API key from caller secrets into a temporary file.
5. Runs optional simulator tests.
6. Archives the app with `xcodebuild archive`.
7. Exports the archive as either a reusable release artifact or an App Store Connect/TestFlight upload with `xcodebuild -exportArchive`.

The Go CLI owns command construction and `ExportOptions.plist` generation so the workflow does not become a large shell script.

## Inputs

Use either `project-path` or `workspace-path`.

Core inputs:

- `project-path`: path to the `.xcodeproj` in the caller repository.
- `workspace-path`: path to the `.xcworkspace` in the caller repository.
- `scheme`: Xcode scheme.
- `team-id`: Apple Developer Team ID.
- `configuration`: defaults to `Release`.
- `test-destination`: defaults to `platform=iOS Simulator,name=iPhone 17`.
- `archive-destination`: defaults to `generic/platform=iOS`.
- `archive-path`: defaults to `build/TestFlight/<scheme>.xcarchive`.
- `export-path`: defaults to `build/TestFlight/export`.
- `export-destination`: defaults to `upload`. Use `export` for a tag-triggered Release artifact.
- `upload-artifact-name`: upload `build/TestFlight` as a GitHub Actions artifact after archive/export.
- `download-artifact-name`: download a prior `build/TestFlight` artifact before export.
- `signing-style`: defaults to `automatic`; use `manual` on clean hosted runners with imported signing assets.
- `bundle-id`: required for manual signing.
- `provisioning-profile`: required for manual signing.
- `skip-tests`: defaults to `false`.
- `skip-archive`: defaults to `false`. Use `true` when reusing an archive downloaded from a Release artifact.
- `skip-cert-check`: defaults to `false`.
- `clean`: defaults to `true`.
- `dry-run`: defaults to `false`.
- `runner-label`: defaults to `macos-26`. Use an Xcode 26 or newer runner for App Store Connect uploads.

Signing inputs:

- `export-method`: defaults to `app-store-connect`.
- `signing-style`: defaults to `automatic`.
- `bundle-id`: required only when using manual signing.
- `provisioning-profile`: required only when using manual signing.

## Secrets

The caller repository provides these secrets:

- `APP_STORE_CONNECT_API_KEY_P8`
- `APP_STORE_CONNECT_API_KEY_ID`
- `APP_STORE_CONNECT_API_ISSUER_ID`

For manual signing, the caller also provides:

- `APPLE_DISTRIBUTION_CERTIFICATE_P12_BASE64`
- `APPLE_DISTRIBUTION_CERTIFICATE_PASSWORD`
- `APPLE_PROVISIONING_PROFILE_BASE64`

Do not commit `.p8` files, certificates, provisioning profiles, archives, or IPAs to any repository.

## Manual Signing Secret Setup

Manual signing is intended for GitHub-hosted macOS runners that should not create Apple signing assets during CI.

Export the Apple Distribution identity from a trusted Mac. The export can require interactive macOS Keychain approval:

```bash
security find-identity -v -p codesigning

P12_PASSWORD="$(openssl rand -base64 24)"
security export \
  -k "$HOME/Library/Keychains/login.keychain-db" \
  -t identities \
  -f pkcs12 \
  -o /tmp/apple-distribution.p12 \
  -P "$P12_PASSWORD"
base64 -i /tmp/apple-distribution.p12 -o /tmp/apple-distribution.p12.b64
```

Encode the app's App Store provisioning profile:

```bash
base64 -i "$HOME/Library/Developer/Xcode/UserData/Provisioning Profiles/<UUID>.mobileprovision" \
  -o /tmp/app-store-profile.mobileprovision.b64
```

Store the values as repository secrets:

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

Clean up local temporary files:

```bash
rm -f /tmp/apple-distribution.p12 /tmp/apple-distribution.p12.b64 /tmp/app-store-profile.mobileprovision.b64
unset P12_PASSWORD
```

## CLI

Build:

```bash
go build ./ios/testflight/cmd/ios-testflight
```

Dry run:

```bash
go run ./ios/testflight/cmd/ios-testflight \
  --project MyApp.xcodeproj \
  --scheme MyApp \
  --team-id ABCDE12345 \
  --skip-tests \
  --export-destination export \
  --dry-run
```

The dry run writes `build/TestFlight/ExportOptions.plist` and prints the `xcodebuild` commands without executing them.

Upload a previously archived release artifact:

```bash
go run ./ios/testflight/cmd/ios-testflight \
  --project MyApp.xcodeproj \
  --scheme MyApp \
  --team-id ABCDE12345 \
  --skip-tests \
  --skip-archive \
  --archive-path build/TestFlight/MyApp.xcarchive \
  --export-destination upload \
  --dry-run
```
