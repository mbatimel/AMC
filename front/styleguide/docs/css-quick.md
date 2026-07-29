### CSS (Quick rules)

#### CSS-переменные

- Переменные в `src/core/shared/styles/theme.css`: `--palette-*`, `--border`, `--header-*`, `--link-*`, status (`--alert-*`, `--success-*`, `--warning-*`), `--focus-outline`, `--z-index-*`.
- Не хардкоди цвета и `z-index` — бери `var(...)`.

#### HeroUI и Tailwind

- HeroUI v3 — **основная UI-библиотека проекта** (`@heroui/react`).
- При реализации экранов и компонентов **сначала ищи готовый примитив в HeroUI** (Button, Input, TextField, Select, Dropdown, Modal, Tabs, Checkbox, Switch, Badge и т.д.) и используй его.
- Свой HTML-контрол или полностью кастомный виджет — только если подходящего компонента в HeroUI нет или его API не закрывает нужное поведение.
- Кастомизация HeroUI — через пропсы и `className` из CSS Module, не через переписывание с нуля.
- Tailwind v4 — только инфраструктура для HeroUI в сборке.
- **В своём коде Tailwind utility-классы не пишем** — CSS Modules + `theme.css` + пропсы HeroUI.

#### Логические свойства

- Используй `margin-inline-start/end`, `padding-inline-start/end`, `border-inline-start/end` вместо физических.

#### Адаптивность (брейкпоинты)

- Миксины: `mobileOnly` (≤680px), `tabletOnly` (681–1074px), `tabletUp` (≥681px), `desktopOnly` (1075–1450px), `desktopUp` (≥1075px), `wideDesktopUp` (≥1451px).

#### Z-index и состояния

- Попапы: `--z-index-popup`, модалки: `--z-index-popup-important`
- Фокус: `--focus-outline`
- Статусы: `--alert-*`, `--success-*`, `--warning-*`
