# gh-workflows

Reusable GitHub Actions workflows and release tooling.

The goal is to keep product repositories clean: app repos keep small workflow wrappers and app-specific config, while the shared CI/release behavior lives here.

## Workflows

| Workflow | Purpose | Docs |
| --- | --- | --- |
| [`homebrew-cask-update.yml`](.github/workflows/homebrew-cask-update.yml) | Update a cask in a separate Homebrew tap after a release asset exists. | [`homebrew/README.md`](homebrew/README.md) |
| [`ai-code-review.yml`](.github/workflows/ai-code-review.yml) | AI code review for pull requests with OpenCode, no GitHub App required. | [`ai-code-review/README.md`](ai-code-review/README.md) |
| [`macos-release.yml`](.github/workflows/macos-release.yml) | Sign, notarize, staple, package, and publish a Developer ID macOS DMG. | [`mac/README.md`](mac/README.md) |
| [`ios-testflight.yml`](.github/workflows/ios-testflight.yml) | Archive an iOS app and upload it to TestFlight. | [`ios/testflight/README.md`](ios/testflight/README.md) |

## Local development

```bash
go test ./...
go build ./ios/testflight/cmd/ios-testflight
go build ./mac/cmd/macos-release
```

Each workflow's README includes a thin caller example plus its inputs and secrets. `act` can catch basic workflow wiring problems (`act --validate`, `act --dryrun`), but iOS and macOS release flows still need a real GitHub-hosted macOS runner to verify Xcode, signing, Keychain, and App Store Connect uploads.

## Versioning

Pin callers to a tag or major version, not `main`. Additive inputs can stay under the current major tag after testing. Breaking changes use a new major tag.
