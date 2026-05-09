# CI/CD

The repository uses GitHub Actions for both CI and CD because the project already contains `.github/workflows`, is structured as a single repository, and ships two Dockerized services that are natural deployment units.

## What CI checks

- Backend unit tests: `ya tool go test ./...` (локально с `ya`; в CI — `go test`, см. `scripts/ci/backend-checks.sh`)
- Backend lint: `golangci-lint run ./... --config=.golangci.yml`
- Backend build: `ya tool go build ./cmd/api` (или `go build` без `ya`)
- Frontend install, OpenAPI client generation, lint and production build
- Docker Compose syntax validation
- Backend and frontend Docker image build
- Migration smoke check by starting `postgres` and `backend` through Compose and waiting for `/health/ready`

There is no dedicated frontend test suite in the current repository, so CI does not invent a fake `npm test` stage.

## Workflows

- `.github/workflows/ci.yml`
  - `backend`
  - `frontend`
  - `docker`
- `.github/workflows/cd.yml`
  - `resolve-target`
  - `build-and-push`
  - `deploy`

## Local helper scripts

- `./scripts/ci/backend-checks.sh`
- `./scripts/ci/frontend-checks.sh`
- `./scripts/ci/docker-checks.sh`

These scripts are the same entry points used by CI to keep the workflows readable and to avoid duplicating command sequences.

## Deployment model

CD deploys Docker images, not source code.

1. The workflow resolves whether the current event should deploy:
   - manual `workflow_dispatch` with an explicit target environment
   - automatic `push` deployment when the pushed branch matches one of the configured branch variables
2. The workflow builds backend and frontend images and pushes them to the configured registry
3. The workflow uploads:
   - `deploy/docker-compose.deploy.yml`
   - a generated `.deploy.env`
4. The workflow logs into the target registry on the remote host
5. The workflow runs `docker compose config -q`, `pull`, and `up -d --remove-orphans`

The deployment scaffold intentionally assumes an external PostgreSQL instance through `DB_DSN`. This avoids coupling CD to a single server topology and works for managed databases or separately provisioned self-hosted databases.

## Repository variables

### Repo-level variables

- `STAGING_DEPLOY_ENABLED`
- `STAGING_DEPLOY_BRANCH`
- `PRODUCTION_DEPLOY_ENABLED`
- `PRODUCTION_DEPLOY_BRANCH`

These control whether automatic deployments from `push` are active and which branches map to `staging` or `production`.

### Environment variables for `staging` and `production`

- `DOCKER_REGISTRY`
- `DOCKER_BACKEND_IMAGE`
- `DOCKER_FRONTEND_IMAGE`
- `VITE_API_BASE_URL`
- `DEPLOY_HOST`
- `DEPLOY_PORT`
- `DEPLOY_USER`
- `DEPLOY_PATH`
- `DEPLOY_PROJECT_NAME`
- `DEPLOY_COMPOSE_FILENAME`
- `BACKEND_PORT`
- `FRONTEND_PORT`
- `FRONTEND_PUBLIC_URL`

Use GitHub Environment-scoped variables so staging and production can differ cleanly.

## Environment secrets

For each GitHub Environment (`staging`, `production`), configure:

- `DOCKER_REGISTRY_USERNAME`
- `DOCKER_REGISTRY_PASSWORD`
- `DEPLOY_SSH_PRIVATE_KEY`
- `DEPLOY_SSH_KNOWN_HOSTS`
- `DEPLOY_RUNTIME_ENV_FILE`

`DEPLOY_RUNTIME_ENV_FILE` should contain newline-separated `KEY=value` pairs for the application runtime. Start from `deploy/runtime.env.example`.

Example shape:

```dotenv
DB_DSN=postgres://user:password@db.example.com:5432/gift_suggestion?sslmode=require
AUTH_JWT_SECRET=replace-with-strong-secret
APP_ENV=production
LOG_LEVEL=info
ML_GRPC_ENABLED=false
EMAIL_ENABLED=false
```

## Full environment inventory

### CI-only

- `VITE_API_BASE_URL`
- `GO_VERSION_FILE`
- `NODE_VERSION`
- `GOLANGCI_LINT_VERSION`

### Backend runtime

- `APP_NAME`
- `APP_ENV`
- `LOG_LEVEL`
- `HTTP_HOST`
- `HTTP_PORT`
- `HTTP_READ_TIMEOUT`
- `HTTP_WRITE_TIMEOUT`
- `HTTP_IDLE_TIMEOUT`
- `HTTP_SHUTDOWN_TIMEOUT`
- `DB_DSN`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `DB_CONN_MAX_LIFETIME`
- `DB_PING_TIMEOUT`
- `DB_MIGRATIONS_ENABLED`
- `IMPORT_MAX_FILE_SIZE_BYTES`
- `ML_GRPC_ENABLED`
- `ML_GRPC_ADDR`
- `ML_GRPC_DIAL_TIMEOUT`
- `ML_GRPC_REQUEST_TIMEOUT`
- `ML_GRPC_MAX_RETRIES`
- `VK_ENABLED`
- `VK_REQUEST_TIMEOUT`
- `VK_TOKEN_ENCRYPTION_KEY`
- `EMAIL_ENABLED`
- `EMAIL_PROVIDER`
- `EMAIL_FROM_EMAIL`
- `EMAIL_FROM_NAME`
- `FRONTEND_BASE_URL`
- `EMAIL_SEND_TIMEOUT`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_USE_TLS`
- `AUTH_JWT_SECRET`
- `AUTH_JWT_ISSUER`
- `AUTH_JWT_AUDIENCE`
- `AUTH_ACCESS_TTL`
- `AUTH_REFRESH_TTL`
- `AUTH_PASSWORD_RESET_TTL`
- `AUTH_EMAIL_VERIFICATION_TTL`
- `AUTH_REFRESH_COOKIE_NAME`
- `AUTH_REFRESH_COOKIE_PATH`
- `AUTH_REFRESH_COOKIE_DOMAIN`
- `AUTH_REFRESH_COOKIE_SECURE`

### Frontend runtime/build

- `VITE_API_BASE_URL` (build-arg собранного образа frontend)
- `FRONTEND_PUBLIC_URL` (подстановка compose на сервере)
- Переменные GitHub / GitLab Environment для frontend-сборки (имена сохранены как ниже — **не используйте** `VK_APP_ID` во встроенном CD):
  - `VK_ID` → в Dockerfile пробрасывается как `VITE_VK_APP_ID`
  - `VK_REDIRECT_URI` → `VITE_VK_REDIRECT_URI`
  - **`VK_SECRET`**: ключ приложения ВК относится к серверу; **не задаётся** шагами CD для образа frontend и не записывается в `VITE_*`. Храните в `DEPLOY_RUNTIME_ENV_FILE`/секретах backend-процесса, когда добавите поддержку на API.

При деплое в конец `.deploy.env` дописываются `VITE_VK_APP_ID` и `VITE_VK_REDIRECT_URI` из переменных `VK_ID`/`VK_REDIRECT_URI` окружения GitHub, чтобы они совпадали с собранным образом и попадали в `frontend.environment` (см. `services/frontend/docker/99-print-frontend-url.sh`).

### CD / deployment

- `STAGING_DEPLOY_ENABLED`
- `STAGING_DEPLOY_BRANCH`
- `PRODUCTION_DEPLOY_ENABLED`
- `PRODUCTION_DEPLOY_BRANCH`
- `DOCKER_REGISTRY`
- `DOCKER_BACKEND_IMAGE`
- `DOCKER_FRONTEND_IMAGE`
- `DOCKER_REGISTRY_USERNAME`
- `DOCKER_REGISTRY_PASSWORD`
- `DEPLOY_HOST`
- `DEPLOY_PORT`
- `DEPLOY_USER`
- `DEPLOY_PATH`
- `DEPLOY_PROJECT_NAME`
- `DEPLOY_COMPOSE_FILENAME`
- `DEPLOY_SSH_PRIVATE_KEY`
- `DEPLOY_SSH_KNOWN_HOSTS`
- `DEPLOY_RUNTIME_ENV_FILE`
- `BACKEND_PORT`
- `FRONTEND_PORT`
- `FRONTEND_PUBLIC_URL`

## Manual rollout checklist

1. Create GitHub Environments: `staging` and/or `production`
2. Fill environment-scoped variables and secrets
3. Verify the remote host has:
   - Docker Engine
   - Docker Compose v2
   - registry access
4. Verify `DEPLOY_RUNTIME_ENV_FILE` matches the backend's current env contract
5. Run `cd` via `workflow_dispatch` before enabling automatic branch-based deploys
