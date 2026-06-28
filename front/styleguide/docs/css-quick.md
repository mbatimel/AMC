### CSS (Quick rules)

#### CSS-переменные
- Переменные в `src/core/shared/styles/theme.css`: `--palette-*`, `--border`, `--header-*`, `--link-*`, status (`--alert-*`, `--success-*`, `--warning-*`), `--focus-outline`, `--z-index-*`.
- Не хардкоди цвета и `z-index` — бери `var(...)`.

#### HeroUI и Tailwind
- HeroUI v3 подключён; Tailwind v4 — только инфраструктура для HeroUI в сборке.
- **В своём коде Tailwind utility-классы не пишем** — CSS Modules + `theme.css` + пропсы HeroUI.

#### Логические свойства
- Используй `margin-inline-start/end`, `padding-inline-start/end`, `border-inline-start/end` вместо физических.

#### Адаптивность (брейкпоинты)
- Миксины: `mobileOnly` (≤680px), `tabletOnly` (681–1074px), `tabletUp` (≥681px), `desktopOnly` (1075–1450px), `desktopUp` (≥1075px), `wideDesktopUp` (≥1451px).

#### Z-index и состояния
- Попапы: `--z-index-popup`, модалки: `--z-index-popup-important`
- Фокус: `--focus-outline`
- Статусы: `--alert-*`, `--success-*`, `--warning-*`
