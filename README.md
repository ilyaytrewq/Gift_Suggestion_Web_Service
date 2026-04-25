# Gift Suggestion Web Service

MVP web service for personalized gift suggestions. The repository currently contains:

- Go/Gin backend in `services/backend`
- React/Vite frontend in `services/frontend`
- PostgreSQL-based local environment through `docker-compose.yml`

## Local Run

From the repository root:

```bash
docker compose up --build postgres backend frontend
```

Default local URLs:

- Backend health: `http://localhost:8080/health/ready`
- Frontend: `http://localhost:5173`
- PostgreSQL: `localhost:5432`

The frontend container also prints the public URL on startup, so it is visible in `docker compose up` output and `docker compose logs frontend`.

Backend migrations run on startup when `DB_MIGRATIONS_ENABLED=true`.
ML ranking is disabled by default in compose (`ML_GRPC_ENABLED=false`), so recommendations use backend fallback ranking.
VK integration is disabled by default (`VK_ENABLED=false`) and is implemented as a safe scaffold.

## Backend

```bash
cd services/backend
go test ./...
task lint
go run ./cmd/api
```

Backend API contracts are documented in:

```bash
services/backend/docs/openapi/backend.yaml
```

## Frontend

```bash
cd services/frontend
npm install
npm run generate:api
npm run lint
npm run build
npm run dev
```

`npm run generate:api` regenerates TypeScript API types from the backend OpenAPI file.

## Known MVP Limits

- Password reset is implemented as backend token foundation; reset confirmation flow is deferred.
- Recommendation ranking uses deterministic fallback unless a real ML gRPC service is enabled.
- Catalog import is synchronous; async/background import is deferred.
- Tracking events are ingested through a dedicated endpoint; automatic server-side event emission is deferred.
- VK OAuth/API integration is a scaffold; real external VK client behavior is deferred.
