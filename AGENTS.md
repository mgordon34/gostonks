# Repository Guidelines

## Project Structure & Module Organization
Target a conventional Go layout:
- `cmd/<service>/main.go` holds entrypoints for each Dockerized service.
- Shared Go code lives under `internal/<domain>` with reusable packages in `pkg/`. Avoid leaking internal APIs across services unless absolutely necessary.
- Container assets (`Dockerfile`, healthcheck scripts) stay in `deploy/containers/<service>`, and any Compose-level configuration (env files, example secrets) belongs in `deploy/compose`.
- Store infrastructure manifests (migrations, fixtures, mock data) in `assets/` and keep configuration templates in `config/`.
- Tests mirror the Go packages inside `tests/`, e.g., `internal/portfolio/service.go` => `tests/internal/portfolio/service_test.go`.

## Build, Test, and Development Commands
Standard workflow assumes Go 1.21+ and Docker Desktop:
- `go mod tidy` keeps dependencies in sync; never edit `go.sum` manually.
- Use `docker compose up --build` for the full multi-service stack (API, worker, db, etc.). For incremental development, `docker compose up <service>` isolates just the container you need.
- Add helper make targets (`make dev`, `make test`, `make lint`, `make down`) so contributors can avoid memorizing long Compose invocations; document any new target in `README.md`.
- Before tagging a release, run `docker compose build --pull` plus `go clean -testcache && go test ./... -race` to ensure reproducible images and clean test runs.

## Coding Style & Naming Conventions
Follow standard Go idioms:
- Always run `gofmt`, `goimports`, and `golangci-lint run` before committing. Use 2-space indentation in YAML/Compose files but keep default tabs in `.go` sources.
- Package names should be short, all-lowercase, and meaningful; exported identifiers use PascalCase while unexported ones remain camelCase.
- Limit files to a single primary type or responsibility. Prefer functional options for configuration-heavy constructors to keep container wiring clean.
- Document cross-container interfaces (e.g., gRPC, REST, message schemas) in `docs/` so service teams can evolve independently.

## Testing Guidelines
- Unit tests live next to their packages (`*_test.go`), with higher-level integration and contract tests under `tests/`.
- Use `go test ./... -race -coverprofile=coverage.out` locally; CI should enforce ≥80% branch coverage.
- For Compose-based integration tests, spin up the minimal set of services via `docker compose -f deploy/compose/docker-compose.test.yml up --build --exit-code-from tests`.
- When fixing regressions, first add or extend a failing test (unit or integration) that reproduces the bug, then patch the code. Capture logs from all relevant containers for debugging.

## Commit & Pull Request Guidelines
Continue with imperative, ≤72-character commit subjects and concise bodies explaining why a change is needed. Reference issues with `Refs #<id>`. Pull requests must outline:
1. Problem statement (include which service(s) in the Compose stack are affected).
2. Solution overview, including any new containers or Compose services.
3. Testing evidence: command output for `go test`, `golangci-lint`, and relevant Compose runs.
4. Follow-up work or rollout notes (migrations, env changes). Request review before merging and ensure the Compose stack builds cleanly in CI.
