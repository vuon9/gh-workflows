# Homebrew Cask Update

Reusable workflow [`homebrew-cask-update.yml`](../.github/workflows/homebrew-cask-update.yml) that updates a Homebrew cask in a separate tap repository after a release asset exists.

## What it does

1. Checks out the tap repository.
2. Resolves the release asset (by regex or explicit URL) and downloads it.
3. Computes the real SHA256.
4. Updates the cask's `version`, `sha256`, and `url` lines.
5. Validates with `ruby -c`, `brew style`, `brew audit`, `brew fetch`, and `brew install --cask --dry-run`.
6. Commits the cask update and opens (or updates) a pull request to the tap repo, unless `open-pull-request` is false.

## Usage

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

## Secrets

- `TAP_REPO_TOKEN` (required): fine-grained token or GitHub App token with write access to the tap repository. The caller `GITHUB_TOKEN` cannot write to another repository.

## Inputs

- `tap-repository` (required): tap repo such as `owner/homebrew-tap`.
- `cask-token` (required): cask token such as `myapp`.
- `cask-path`: cask path inside the tap repo. Defaults to `Casks/<cask-token>.rb`.
- `source-repository`: release repo. Defaults to the caller repository.
- `release-tag`: release tag. Defaults to the caller ref name.
- `artifact-name-regex`: must match exactly one release asset when `artifact-url` is omitted. Defaults to `\\.dmg$`.
- `artifact-url`: explicit release artifact URL. When set, release asset lookup is skipped.
- `version`: cask version. Defaults to the release tag basename with a leading `v` removed.
- `base-branch`: tap repository base branch. Defaults to `main`.
- `update-branch`: branch for PR updates. Defaults to `homebrew/<cask-token>-<release-tag>`.
- `open-pull-request`: defaults to `true`. Set to `false` only when direct tap pushes are acceptable.
- `dry-run`: updates and validates locally without pushing.
- `runner-label`: default `macos-26`. Use macOS so Homebrew cask validation is available.

## Notes

- The workflow runs on a macOS runner so Homebrew cask validation is available.
- Never commit release artifacts or secret files to the repository.
