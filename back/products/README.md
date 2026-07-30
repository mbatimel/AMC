# Products

Сервис каталога продукции. HTTP API по умолчанию слушает `:8081`, health-check — `:9091`.

Обязательные переменные окружения:

- `PG_DB`
- `PG_USER`
- `PG_PASSWORD`

Опциональные переменные:

- `PG_HOST` (по умолчанию `localhost`)
- `PG_PORT` (по умолчанию `5432`)
- `BIND_ADDR` (по умолчанию `:8081`)
- `HEALTH_ADDR` (по умолчанию `:9091`)
- `ACCESS_URL` (по умолчанию `http://localhost:8080`)

Запуск PostgreSQL из корня репозитория (после создания `deploy/.env` из
`deploy/.env.example`):

```sh
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d postgres
```

Применение миграций:

```sh
cd back/migrations
PG_ADDRESS=localhost:5432 PG_DB=amc PG_USER=amc PG_PASSWORD=secret go run ./cmd
```

Запуск products:

```sh
cd back/products
PG_DB=amc PG_USER=amc PG_PASSWORD=secret ACCESS_URL=http://localhost:8080 \
go run ./cmd
```
