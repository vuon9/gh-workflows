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
7. Uploads the archive to App Store Connect/TestFlight with `xcodebuild -exportArchive`.

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
- `skip-tests`: defaults to `false`.
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

Do not commit `.p8` files, certificates, provisioning profiles, archives, or IPAs to any repository.

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
  --dry-run
```

The dry run writes `build/TestFlight/ExportOptions.plist` and prints the `xcodebuild` commands without executing them.
