# Contributing to WebScan

Thank you for considering contributing to WebScan!

## Getting Started

1. **Fork** the repository and **clone** your fork.
2. Create a branch: `git checkout -b feature/my-feature`.
3. Make your changes following the code style below.
4. Run tests: `go test ./... -v`
5. Run linter: `golangci-lint run ./...`
6. Commit with a clear message and open a **Pull Request** against `main`.

## Code Style

- Follow Go conventions (`gofmt`, `go vet`).
- Keep functions short and focused. Prefer composition over deep nesting.
- Add unit tests for new packages and functions.
- Use `#nosec G304` (and similar) only where user-supplied paths are intentional; document why.

## Adding Templates

YAML detection templates live in `templates/`. Each template must have:
- Unique `id` (kebab-case, e.g., `cve-2024-12345-myapp`)
- A `severity` field: `info`, `low`, `medium`, `high`, or `critical`
- At least one `request` with meaningful `matchers`

Test a new template before submitting:

```bash
./webscan.exe vuln -t http://your-test-target -T templates/your-template.yaml
```

## Adding Signatures

HTTP fingerprint signatures live in `test/signatures/popular_signatures.json`.
Each entry must include:
- `name`: technology/vendor name
- `match_headers` or `match_body`: detection criteria

## Reporting Issues

- Search existing issues before opening a new one.
- Include Go version, OS, CLI command and full output.
- For security issues, please **do not** open a public issue — email maintainers privately.

## Disclaimer

WebScan is intended for authorized security testing only. Contributors must ensure
their changes cannot be trivially weaponized for unauthorized access.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
