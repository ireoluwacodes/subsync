# SubSync Backend

Managed subscriptions billing engine for Nomba-powered SaaS, fintech, and marketplace teams.

## Stack

| Layer | Choice |
|-------|--------|
| Runtime | Go 1.22 |
| HTTP | Gin |
| Database | PostgreSQL 15 + GORM |
| Migrations | Goose (schema); GORM for queries |
| Queue | Redis + asynq |
| Deployment | Docker / Cloud Run |

## Quick start

### Prerequisites

- Go 1.22+
- Local PostgreSQL
- Docker (Redis only, via `make up`)
- Make

### Database setup (one-time)

PostgreSQL runs on your machine. Create the database once (as a superuser if `cierge_user` lacks `CREATEDB`):

```bash
createdb -h localhost -p 5432 -U postgres subsync
# or: psql -U postgres -c "CREATE DATABASE subsync OWNER cierge_user;"
```

Credentials are in `.env` (`DB_HOST`, `DB_USER`, etc.). `make db-create` will try this automatically if your user has permission.

### Local development

```bash
cp .env.example .env   # already configured for local postgres + redis

make up                # start Redis (docker)
make db-create         # skip if you created the DB manually
make migrate-up
make run-api           # terminal 1
make run-worker        # terminal 2
```

Or all-in-one (after DB exists):

```bash
make dev
```

### Health checks

```bash
curl http://localhost:8080/health   # liveness
curl http://localhost:8080/ready    # readiness (postgres + redis)
```

## Project layout

```
cmd/
  api/          HTTP server entrypoint
  worker/       asynq background worker
internal/
  domain/       Pure business logic (structs, FSM, repo interfaces)
  db/           PostgreSQL + GORM (models/, repos)
  api/          Gin router, middleware, handlers, DTOs
  service/      Orchestration layer
  jobs/         asynq task handlers (billing, dunning, lifecycle)
  email/        Resend strategy + templates
  storage/      Cloudinary strategy for invoice PDFs
  nomba/        Nomba API client (hackathon phase)
  config/       Environment configuration
  queue/        asynq + Redis client
migrations/     Goose SQL migrations
```

## API response envelope

All responses follow:

```json
{
  "data": {},
  "meta": { "request_id": "..." },
  "error": { "code": "not_found", "message": "..." }
}
```

Error codes: `invalid_request`, `validation_failed`, `unauthorized`, `forbidden`, `not_found`, `conflict`, `transition_not_allowed`, `internal_error`.

## Migrations

```bash
make migrate-up       # apply all
make migrate-down     # roll back one
make migrate-status   # show current version
make migrate-create name=add_foo  # create new migration
```

## Environment variables

See [`.env.example`](.env.example) for all configuration options.

| Variable | Required | Description |
|----------|----------|-------------|
| `APP_ENV` | No | `development` / `staging` / `production` |
| `HTTP_PORT` | No | API port (default: 8080) |
| `DB_HOST` | No | PostgreSQL host (default: localhost) |
| `DB_PORT` | No | PostgreSQL port (default: 5432) |
| `DB_USER` | No | PostgreSQL user (default: cierge_user) |
| `DB_PASSWORD` | No | PostgreSQL password |
| `DB_NAME` | No | Database name (default: subsync) |
| `POSTGRES_DSN` | No | Full DSN; overrides `DB_*` if set |
| `REDIS_URL` | Yes (prod) | Redis connection string |
| `NOMBA_CREDENTIALS_ENCRYPTION_KEY` | Prod | AES-256 key for encrypting tenant Nomba secrets at rest |
| `JWT_SECRET` | Prod | HS256 signing key for merchant dashboard JWTs |
| `JWT_ACCESS_TTL` | No | Access token lifetime (default: 24h) |
| `JWT_REFRESH_TTL` | No | Refresh token lifetime (default: 7d) |
| `BILLING_MOCK_RESULT` | No | Leave empty for live Nomba charges. Set to `success` or `failure` to mock invoice charges locally |
| `RESEND_API_KEY` | No | Resend API key for transactional email (Phase 3) |
| `RESEND_FROM_EMAIL` | No | Default from address for Resend |
| `CLOUDINARY_CLOUD_NAME` | No | Cloudinary cloud for invoice PDF storage |
| `CLOUDINARY_API_KEY` | No | Cloudinary API key |
| `CLOUDINARY_API_SECRET` | No | Cloudinary API secret |
| `CLOUDINARY_FOLDER` | No | Cloudinary upload folder (default: `subsync/invoices`) |
| `NOMBA_WEBHOOK_SIGNING_KEY` | No | Dev fallback for inbound Nomba webhooks |
| `WEBHOOK_SIGNING_SECRET` | Phase 4 | Outbound SubSync webhook signing |

See [nomba-integration.md](nomba-integration.md) for API validation details.

## Build

```bash
make build          # outputs bin/api and bin/worker
docker build --build-arg TARGET=api -t subsync-api .
docker build --build-arg TARGET=worker -t subsync-worker .
```

## Architecture: per-merchant Nomba

Each merchant stores their own Nomba OAuth credentials (`client_id`, encrypted `client_secret`, `account_id`, optional `sub_account_id`, `nomba_env`). SubSync orchestrates billing and calls Nomba on each merchant's behalf — it never pools funds or shares a parent account.

## API auth

**Self-serve (dashboard):** `POST /api/v1/auth/register` then `POST /api/v1/auth/login` → use `Authorization: Bearer <access_jwt>` or the returned `api_key`.

Use the returned `api_key` or JWT as `Authorization: Bearer <token>` for `/api/v1/*` routes.

## Testing

```bash
make test                                    # unit tests
make test-integration                        # postgres + redis integration tests
```

## Roadmap
1. **Phase 1** — Tenant auth, plans/customers/payment-methods CRUD (implemented)
2. **Phase 2** — JWT auth, settings, subscriptions, invoices (per-merchant Nomba credentials) (implemented)
3. **Phase 3** — Background jobs (billing, dunning) with tenant-scoped Nomba calls (implemented)
4. **Phase 4** — Inbound Nomba webhooks, outbound webhooks, customer portal (implemented)
5. **Phase 5** — Analytics + live Nomba charge swap

