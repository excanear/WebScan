Release automation (goreleaser)

This repository includes a `release.yml` workflow that uses goreleaser to publish
GitHub releases when a tag `v*.*.*` is pushed.

Required repository secrets / configuration:

- `GITHUB_TOKEN` — automatically provided by GitHub Actions for the workflow; no manual action required for basic releases.
- Optional: `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` — if you want goreleaser to sign artifacts.
- Optional: `HOMEBREW_GITHUB_API_TOKEN` — if publishing Homebrew taps via goreleaser.

Recommendations:

- Add any required secrets at: Settings → Secrets → Actions in your repository.
- Test goreleaser locally with `goreleaser release --snapshot --skip-publish` before enabling automatic releases.

Troubleshooting:

- If a release job fails due to permissions, verify `GITHUB_TOKEN` and repo permissions for Actions.
- To run goreleaser in dry-run from CI, set the action args to `release --snapshot --skip-publish`.

Security:

- Avoid committing signing keys to the repository. Use GitHub Secrets for sensitive data.
