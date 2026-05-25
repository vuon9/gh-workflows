# AGENTS.md - gh-workflows

This repository contains reusable GitHub Actions workflows and supporting tools.

## Guardrails

- Keep workflows reusable and project-neutral. App-specific values must come from `workflow_call` inputs or caller repository secrets.
- Do not commit Apple certificates, provisioning profiles, App Store Connect API keys, generated archives, IPAs, screenshots, logs containing secrets, or other release artifacts.
- Do not add organization-specific, customer-specific, or private project context to public docs, examples, prompts, or test fixtures.
- Keep release logic deterministic and testable. Put non-trivial behavior in Go code with unit tests instead of long shell scripts or large inline YAML blocks.
- Prefer small, domain-scoped folders such as `ios/testflight/`, `web/deploy/`, and `backend/ci/`.
- Reusable workflow YAML files must stay in `.github/workflows/` because GitHub requires that location.
- Use version tags such as `v1` for callers. Do not force app repositories to reference a moving branch for release automation.
- Keep workflow inputs stable. For breaking changes, publish a new major tag and update docs.
- Treat public compatibility as a design constraint: everything in this repository may be read, forked, and reused publicly.

## Verification

Before claiming a workflow/tool change is ready:

```bash
go test ./...
go build ./ios/testflight/cmd/ios-testflight
```

For workflow changes, also run at least one dry run from a caller app repository before removing legacy scripts.

## iOS TestFlight

The iOS TestFlight flow is split into:

- `.github/workflows/ios-testflight.yml`: reusable GitHub Actions entrypoint.
- `ios/testflight/cmd/ios-testflight`: Go CLI used by the workflow.
- `ios/testflight/internal/exportoptions`: `ExportOptions.plist` generation.
- `ios/testflight/internal/xcode`: xcodebuild command planning.

Caller repositories should keep only a thin workflow wrapper and app-specific configuration.
