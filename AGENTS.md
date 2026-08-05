# Repository Guidelines

## Project Structure & Module Organization

Evolution Go is a Go API service. The entry point is `cmd/evolution-go/main.go`. Core code lives in `pkg/`, organized by feature and layer, such as `pkg/message/{handler,service,repository,model}`, `pkg/instance/...`, and `pkg/events/{webhook,rabbitmq,nats,websocket}`. Shared packages include `pkg/config`, `pkg/middleware`, `pkg/routes`, `pkg/storage`, and `pkg/utils`. Tests are colocated as `*_test.go`. Swagger files are in `docs/`; guides are under `docs/wiki/`. Docker assets are in `Dockerfile`, `docker/`, and `docker/examples/`. Static and browser helper assets are in `public/` and `passkey-helper/`.

## Build, Test, and Development Commands

Use `make help` to list all targets.

- `make deps`: download and verify modules.
- `cp .env.example .env`: create local configuration.
- `make dev`: run the server in development mode with `go run`.
- `make run`: run the server in normal mode.
- `make build`: compile `build/evolution-go`.
- `make test`: run all Go tests.
- `make test-coverage`: generate `coverage.out` and `coverage.html`.
- `make test-race`: run tests with the race detector.
- `make swagger`: regenerate Swagger docs with `swag`.
- `make docker-build` / `make docker-run`: build and run the image.

## Coding Style & Naming Conventions

Format Go code with `make fmt` or `go fmt ./...` before committing. Run `make vet` and `make lint` when changing behavior; `make lint` requires `golangci-lint`. Use idiomatic Go naming: exported identifiers use `CamelCase`, unexported identifiers use `camelCase`, and packages use short lowercase names. Keep the existing Handler -> Service -> Repository structure for HTTP features, and add new feature packages under `pkg/<feature>/`.

## Testing Guidelines

The project uses Go's standard testing package. Place tests next to the code they cover and name functions `TestXxx`. Prefer focused unit tests for services, repositories, and utilities; use table-driven tests where inputs vary. Run `make test` before opening a PR, and use `make test-race` for concurrency-sensitive changes.

## Commit & Pull Request Guidelines

Project docs request Conventional Commits such as `feat: adiciona suporte para audio`, `fix: corrige erro ao deletar instancia`, `docs: atualiza guia`, and `chore: atualiza dependencias`. Recent history also includes sync commits; for contributor work, prefer descriptive Conventional Commit messages.

Pull requests should describe the change, link issues, include logs or screenshots when behavior is visible, and note configuration or migration impacts. Update docs and run `make swagger` when API endpoints change. Confirm formatting, linting, and tests in the checklist.

## Security & Configuration Tips

Do not commit `.env`, API keys, database URLs, license data, or storage credentials. Start from `.env.example` or `docker/examples/.env.example`.
