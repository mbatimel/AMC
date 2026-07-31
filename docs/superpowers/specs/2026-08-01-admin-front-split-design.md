# Вынесение админки в отдельный проект admin-front

## Контекст

Сейчас `front` — единое Next.js-приложение: публичный сайт + админка под
`/admin/*` в одном билде, одном контейнере, на одном домене
(`wk.amctechgroup.ru`). Нужно разделить их на два независимых Next.js-проекта,
чтобы админку можно было собирать и деплоить отдельно, на своём домене
(`admin.wk.amctechgroup.ru`).

Ключевая сложность: обе части сейчас читают/пишут один и тот же временный
файловый "бэкенд" `portal-api` (контент, баннеры, legal, support, feedback,
portal-users, signup-requests, audit-log, ассистент) — `src/app/portal-api/*`
+ `src/core/shared/server/portal/*`, хранящий состояние в `portal.json` на
volume. Это заглушка на время, пока не появится `back/admin` в полном объёме
(см. `front/src/app/portal-api/README.md`).

## Решения (согласованы с пользователем)

1. **Общий код** (api-клиенты, часть UI-кита, иконки, лib-утилиты) —
   **дублируется** в `admin-front` копированием файлов, без yarn workspaces
   и общего пакета. Дешевле сейчас; расхождение допустимо, т.к. `portal-api`
   всё равно временный и будет заменён на `back/admin`.
2. **`portal-api`** (роуты + файловое хранилище) остаётся **только во
   `front`**. Публичный сайт не должен зависеть от аптайма админки.
   `admin-front` обращается к нему через прокси (nginx + fallback-rewrite
   в `next.config.ts`), физически данные хранятся в одном месте — без
   split-brain по `portal.json`.
3. **Домен**: `admin.wk.amctechgroup.ru`.
4. Деплой заводится полностью: `Dockerfile`, сервис в
   `deploy/docker-compose.yml`, nginx server block.

## Архитектура

Два независимых Next.js-приложения — соседние директории в одном репозитории
(`front/`, `admin-front/`), каждое со своим `package.json`, сборкой,
Docker-образом. Общий backend gateway (`API_PROXY_TARGET`,
`wk.amctechgroup.ru` → access/auth/orders/products/users/admin Go-сервисы)
доступен обоим одинаково — это НЕ временный слой, оба проекта настраивают
свой `next.config.ts` рядом с ним независимо.

`portal-api` (временный слой) доступен только через `front`; `admin-front`
ходит к нему по HTTP (см. раздел «Сеть»).

## Структура admin-front

Новый проект, тот же tooling, что у `front` (Next 16, React 19, TypeScript,
effector/effector-react, `@heroui/react`, Tailwind v4, ESLint/Prettier/
Stylelint конфиги — копии текущих `front/eslint.config.js`,
`postcss.config.js`, `prettier.config.js`, `stylelint.config.js`,
`tsconfig.json` с тем же alias `@/*`).

### Переезжает (move — удаляется из front)

| Откуда (front) | Куда (admin-front) |
|---|---|
| `src/app/admin/*` (кроме `layout.tsx`) | `src/app/*` — см. де-префиксацию маршрутов ниже |
| `src/views/Admin/*` | `src/views/Admin/*` (без изменений внутри) |
| `src/core/entities/adminSession/*` | `src/core/entities/adminSession/*` |
| `src/core/shared/api/admin.ts` | `src/core/shared/api/admin.ts` |
| `src/core/shared/api/portalUsers.ts` | `src/core/shared/api/portalUsers.ts` |
| `src/core/shared/api/signupRequests.ts` | `src/core/shared/api/signupRequests.ts` |
| ветка `/admin` в `src/proxy.ts` | собственный `src/proxy.ts` (охраняет все пути, кроме `/login`, куки `ADMIN_USER_ID_COOKIE`) |

### Дублируется (копия — остаётся и во front, и в admin-front)

Всё это используется и публичным сайтом, и админкой одновременно:

- `core/shared/api/{content,feedback,support,products}.ts`
- `core/shared/api/{portalClient,parseApiError}.ts`
- `core/shared/ui/{AuthShell,FormSelect,Toast}`
- `core/shared/icons/IconKey.tsx` (+ `icons/types.ts`, если нужен как тип)
- `core/shared/lib/{formatPrice,readFormString}.ts`
- `core/shared/styles/*` (fonts.css, global.css, normalize.css, theme.css, mixins) — копируется целиком, файлы маленькие
- `public/` шрифты/статика, используемая в styles

### Не переносится и не дублируется

- `src/app/portal-api/*`, `src/core/shared/server/portal/*`,
  `src/core/shared/server/assistant/*` — остаются только во front.
- Все публичные вьюхи/entities (Cart, Catalog, Cabinet и т.д.) — не нужны
  admin-front.

### Маршруты: де-префиксация

Т.к. админка переезжает на свой домен, префикс `/admin` в путях больше не
нужен. `admin-front` получает собственный `core/shared/router/paths.ts`:

```
/            (было /admin)
/login       (было /admin/login)
/products    (было /admin/products)
/products/:id
/categories
/content
/content/:key
/banners
/legal
/feedback
/support
/signup-requests
/users
/audit-log
```

`views/Admin/lib/nav.ts` и все внутренние ссылки на `AppPath.Admin*`
обновляются на новые пути без префикса. `front/src/core/shared/router/paths.ts`
теряет все `Admin*`-значения и `getAdminContentPath`/`getAdminProductPath`.

## Сеть: как admin-front достаёт данные

Два разных внешних сервиса, две разные схемы:

1. **Реальный backend** (`/api/v1/*` — products/users/orders/access/admin
   Go-сервисы). `admin-front/next.config.ts` получает те же rewrite'ы, что
   сейчас в `front/next.config.ts` (`API_PROXY_TARGET`, по умолчанию
   `https://wk.amctechgroup.ru`). Ничего специфичного — это не завязано на
   `front`-приложение.

2. **`portal-api`** (временный слой, живёт только во front). Клиентский код
   в admin-front (`content.ts`, `feedback.ts`, `support.ts`, `admin.ts`,
   `portalUsers.ts`, `signupRequests.ts`) продолжает ходить по
   относительному пути `/portal-api/...` через `portalClient.ts` без
   изменений — прокси прозрачен для кода:
   - в проде маршрутизацию делает **nginx**: на домене
     `admin.wk.amctechgroup.ru` location `/portal-api/` проксируется на
     `http://front:3000/portal-api/` (front и admin-front — контейнеры в
     одной docker-сети);
   - для локальной разработки (`yarn dev` без nginx) в
     `admin-front/next.config.ts` добавляется fallback-rewrite
     `/portal-api/:path*` → `${PORTAL_API_PROXY_TARGET}/portal-api/:path*`
     (по умолчанию `https://wk.amctechgroup.ru`) — по аналогии с уже
     существующим `API_PROXY_TARGET` в `front/next.config.ts`.

## Деплой

- `admin-front/Dockerfile` — копия шаблона `front/Dockerfile` (multi-stage,
  standalone output), без блока `PORTAL_DATA_FILE`/volume — своего
  хранилища у admin-front нет.
- `deploy/docker-compose.yml`: новый сервис `admin-front`
  (`image: ghcr.io/mbatimel/amc-admin-front:${IMAGE_TAG}`, `PORT=3000`,
  `HOSTNAME=0.0.0.0`, `API_PROXY_TARGET`), добавляется в `depends_on` у
  `nginx`.
- `deploy/nginx/conf.d/admin.wk.amctechgroup.ru.conf` — новый server block:
  - `location /api/v1/*` — те же блоки, что в `wk.amctechgroup.ru.conf`
    (прямой проксинг на access/auth/orders/products/users/admin);
  - `location /portal-api/` → `proxy_pass http://front:3000/portal-api/;`
  - `location /` → `proxy_pass http://admin-front:3000;`

## Аутентификация

Без изменений в логике: `ADMIN_USER_ID_COOKIE` (`admin_user_id`) как и
раньше выставляется без `Domain=`, то есть host-only для
`admin.wk.amctechgroup.ru`. Кука сессии обычного сайта (`USER_ID_COOKIE`) на
`admin-front` не используется и не нужна — админка и сайт теперь полностью
раздельные сессии, что и так было по сути (разные куки), просто раньше жили
в одном origin.

## Проверка после реализации

- `admin-front`: `yarn build` проходит, `yarn typecheck`, `yarn lint`.
- `front`: после удаления admin-кода `yarn build`/`yarn typecheck`/`yarn lint`
  по-прежнему проходят (нет мёртвых импортов на удалённые файлы).
- Ручная проверка: логин в админку на `admin.wk.amctechgroup.ru`, CRUD по
  контенту/баннерам — изменения видны на публичном сайте (общий
  `portal-api`).
