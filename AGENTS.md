# Repository Guidelines

## Project Structure & Module Organization

This is a Go module for the Evolution Go WhatsApp API. The executable entry point is `cmd/evolution-go/main.go`. Application code lives under `pkg/`, organized by domain: handlers, services, repositories, models, middleware, events, storage, config, and routes. Tests are colocated as `*_test.go` files. Swagger output is in `docs/`, longer documentation is in `docs/wiki/`, Docker assets are in `Dockerfile` and `docker/`, and static or built manager assets live in `public/` and `manager/dist/`.

## Build, Test, and Development Commands

- `make deps`: download and verify Go modules.
- `make dev`: run the API in development mode; `make watch` adds hot reload with `air`.
- `make build`: compile `build/evolution-go`.
- `make test`, `make test-coverage`, `make test-race`: run tests, coverage, or race detection.
- `make fmt`, `make vet`, `make lint`, `make check`: format, analyze, lint, and run the full pre-submit suite.
- `make swagger`: regenerate Swagger docs with `swag`.
- `make docker-build` and `make docker-run`: build and run the Docker image.

## Coding Style & Naming Conventions

Use standard Go formatting via `gofmt`/`go fmt`; keep tabs as produced by the formatter. Follow existing package naming patterns such as `pkg/message`, `pkg/instance`, and `pkg/sendMessage`. Keep HTTP handlers, business logic, persistence, and models in their established `handler`, `service`, `repository`, and `model` layers. Export names only when they are part of a package boundary.

## Testing Guidelines

Use Go’s built-in `testing` package and helper libraries such as `sqlmock` where appropriate. Name test files `*_test.go` and test functions `TestXxx`. Add or update tests near the changed package, especially for services, repositories, utilities, and protocol behavior. Run `make test` normally and `make test-race` when touching concurrency, event producers, or WebSocket code.

## Commit & Pull Request Guidelines

Recent history uses short Portuguese messages, `sync:` commits, and conventional-style entries such as `chore(docs): ...`. Prefer concise imperative messages with an optional scope, for example `fix(message): validate media payload`. Pull requests should describe the change, list validation performed, link related issues, and include screenshots only for Manager UI or documentation visual changes. Mention configuration, migration, or deployment impacts explicitly.

## Security & Configuration Tips

Do not commit `.env` files, API keys, license data, database URLs, or MinIO credentials. Use `.env.example` when available and document new environment variables in `README.md` or `docs/wiki/referencia/environment-variables.md`.
