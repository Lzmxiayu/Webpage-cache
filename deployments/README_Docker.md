# Docker Compose Quick Start

## 1. Start All Services

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

Services started:
- `app` (Go API): `http://localhost:8080`
- `mysql` (MySQL 8.4): `localhost:3306`
- `redis` (Redis 7): `localhost:6379`

Database migrations are auto-applied from:
- `internal/repository/migrations/*.sql`

## 2. Verify

```bash
curl http://localhost:8080/metrics
```

```bash
curl -X POST http://localhost:8080/screenshot \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

## 3. Check Logs

```bash
docker compose -f deployments/docker-compose.yml logs -f app
```

## 4. Stop

```bash
docker compose -f deployments/docker-compose.yml down
```

To remove persistent volumes too:

```bash
docker compose -f deployments/docker-compose.yml down -v
```

## Troubleshooting

If `POST /screenshot` returns `INTERNAL_ERROR` and MySQL has no `tasks` table:

1. Check app logs:

```bash
docker compose -f deployments/docker-compose.yml logs -f app
```

2. Restart app to trigger startup migrations:

```bash
docker compose -f deployments/docker-compose.yml restart app
```

3. If you want full re-init from scratch (destructive to DB data):

```bash
docker compose -f deployments/docker-compose.yml down -v
docker compose -f deployments/docker-compose.yml up -d --build
```
