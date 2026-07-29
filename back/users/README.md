# Users

Сервис пользователей, профиля, клиентских кабинетов и избранных товаров.

## Запуск

Сначала примените общие миграции:

```bash
cd back/migrations
PG_ADDRESS=localhost:5432 PG_DB=amc PG_USER=postgres PG_PASSWORD=postgres go run ./cmd
```

Затем запустите сервис:

```bash
cd back/users
PG_HOST=localhost PG_PORT=5432 PG_DB=amc PG_USER=postgres PG_PASSWORD=postgres \
ACCESS_URL=http://localhost:8080 BIND_ADDR=:8083 go run ./cmd
```

Личные ручки используют `X-User-Id`. Операции CRUD, меняющие роль, используют
`X-Admin-User-Id` и делегируют изменение роли access-сервису.
