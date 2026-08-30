# AI Code Review

Reusable workflow [`ai-code-review.yml`](../.github/workflows/ai-code-review.yml) that runs an AI code review on pull requests using OpenCode (`opencode github run`) with the runner `GITHUB_TOKEN`, so no GitHub App install is required.

## What it does

- Auto-reviews every non-draft PR by default (`mode: pr`).
- Supports on-demand review via a `/oc` or `/opencode` comment (`mode: comment`).
- Converges to a single OpenCode comment per issue/PR: it keeps the oldest comment id, rewrites its body, and deletes duplicates from earlier runs, so re-runs upsert instead of stacking.

## Usage

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
    uses: vuon9/gh-workflows/.github/workflows/ai-code-review.yml@v0.4.0
    with:
      mode: pr
    secrets:
      OPENCODE_API_KEY: ${{ secrets.OPENCODE_API_KEY }}
```

## Secrets

- `OPENCODE_API_KEY` (required): OpenCode API key (from `opencode auth login` on your machine, or the OpenCode dashboard).

## Inputs

- `model`: defaults to `opencode/muse-spark-1.2-contributor-free`. The non-free `opencode/muse-spark-1.2-contributor` requires workspace credits.
- `prompt`: custom review prompt. Defaults to a thorough review checklist.
- `mode`: `pr` (auto-review non-draft PRs) or `comment` (on-demand `/oc`).
- `comment-marker`: hidden HTML comment tag used to track the workflow's comment. Defaults to `opencode-ai-review`.

## Notes

- `permissions` MUST live on the caller job, not inside the reusable job: a `permissions` block inside a `workflow_call` job causes `startup_failure`.
- `if:` conditions are not allowed on caller jobs that `uses:` a reusable workflow; the `mode` input handles gating inside the reusable workflow.
- `issues: write` is required so the workflow can create, update, and delete the PR conversation comment; `issues: read` is not enough.
- The workflow is at `.github/workflows/ai-code-review.yml` and can also be called directly by other repositories.
