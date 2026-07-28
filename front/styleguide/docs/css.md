## CSS-переменные

**Используйте CSS-переменные из темы** для цветов, отступов и других значений. Переменные определены в `src/core/shared/styles/theme.css`.

```css
/* ✅ Правильно */
.container {
  background-color: var(--palette-white);
  border: 1px solid var(--border);
  color: var(--palette-black);
  z-index: var(--z-index-popup-important);
}

/* ❌ Неправильно — хардкод */
.container {
  background-color: #fff;
  color: #111;
  z-index: 1000;
}
```

### Palette

Базовые цвета:

- `--palette-black`, `--palette-dark`, `--palette-gray`, `--palette-light`, `--palette-ulight`, `--palette-white`
- `--palette-disabled` — неактивные элементы
- `--palette-main` — акцентный цвет темы
- `--palette-sec` — вторичный цвет

**Когда использовать:**

- `--palette-white`, `--palette-ulight` — фоны
- `--palette-black`, `--palette-dark` — текст
- `--palette-main` — акценты, брендовые элементы
- `--palette-gray`, `--palette-disabled` — второстепенный контент, disabled-состояния

### Base

- `--border` — цвет границ
- `--disabled-overlay` — оверлей для disabled-состояний
- `--disabled-filter` — фильтр для disabled-состояний

### Header

- `--header-bg`, `--header-bg-secondary` — фоны шапки
- `--header-typo`, `--header-menu-typo`, `--header-menu-typo-secondary` — текст
- `--header-main`, `--header-secondary`, `--header-active` — акцентные цвета
- `--header-height`, `--header-height-secondary` — высота
- `--header-logo-main-max-height`, `--header-logo-partner-max-height` — логотипы

### Link

- `--link`, `--link-hover`, `--link-active` — основные ссылки
- `--link-alt`, `--link-alt-hover`, `--link-alt-active` — альтернативные ссылки (на тёмном фоне)

```css
.link {
  color: var(--link);
  transition: color 0.16s ease;

  &:hover,
  &:focus {
    color: var(--link-hover);
  }

  &:active {
    color: var(--link-active);
  }
}
```

Для ссылок также есть миксин `@mixin link` в `src/core/shared/styles/mixins/link.css`.

### Status

- `--alert`, `--alert-bg`, `--alert-alt`, `--alert-outline` — ошибки
- `--success`, `--success-bg`, `--success-alt` — успех
- `--warning`, `--warning-bg`, `--warning-alt` — предупреждения
- `--focus-outline` — обводка при фокусе (accessibility)
- `--highlight-bg` — подсветка

### Z-index

- `--z-index-start` (1), `--z-index-tip` (2)
- `--z-index-popup` (12) — dropdown, tooltip, обычные попапы
- `--z-index-popup-important` (22) — модальные окна

```css
.popup {
  z-index: var(--z-index-popup);
}

.modal {
  z-index: var(--z-index-popup-important);
}
```

### Примеры

```css
.card {
  background-color: var(--palette-white);
  border: 1px solid var(--border);
  color: var(--palette-black);
}

.sidebar {
  background-color: var(--palette-ulight);
}

.hint {
  color: var(--palette-dark);
}

.input:focus {
  box-shadow: 0 0 0 2px var(--focus-outline);
}

.alert {
  background-color: var(--alert-bg);
  color: var(--alert);
}
```

### Что не хардкодить

- ❌ `background-color: #fff` → ✅ `var(--palette-white)`
- ❌ `color: #111` → ✅ `var(--palette-black)`
- ❌ `border: 1px solid #e5e5e5` → ✅ `var(--border)`
- ❌ `z-index: 1000` → ✅ `var(--z-index-popup-important)`

**Исключения:** служебные/отладочные стили, уникальные градиенты вне темы.

## HeroUI и Tailwind

Проект использует **HeroUI v3**. Библиотека построена на Tailwind CSS v4 — он подключён в сборке (`@import "tailwindcss"` и `@import "@heroui/styles"` в `src/core/shared/styles/index.css`).

**В своём коде Tailwind utility-классы не используем** (`flex`, `p-6`, `text-sm` и т.п.).

**Как стилизуем:**

- **Своя вёрстка** — CSS Modules (`.module.css`) + переменные из `theme.css` + логические свойства
- **UI-компоненты** — HeroUI через пропсы (`variant`, `size`, `isDisabled` и т.д.)
- **Кастомизация HeroUI** — `className` с классами из CSS Module, не Tailwind

```tsx
// ✅ Правильно
import clsx from 'clsx';
import { Button } from '@heroui/react';

import styles from './Home.module.css';

export const Home = (): JSX.Element => (
  <main className={clsx(styles.root)}>
    <Button variant="primary">AMC</Button>
  </main>
);
```

```tsx
// ❌ Неправильно — Tailwind в прикладном коде
<main className="flex min-h-screen items-center justify-center p-6">
```

Tailwind остаётся только как инфраструктурная зависимость HeroUI, а не как способ вёрстки в проекте.

## Логические свойства

**Используйте логические свойства** вместо физических для поддержки RTL и улучшения семантики.

```css
/* ✅ Правильно */
.element {
  margin-inline-start: 8px;
  margin-inline-end: 12px;
  padding-inline: 16px;
}

/* ❌ Неправильно */
.element {
  margin-left: 8px;
  margin-right: 12px;
  padding-left: 16px;
  padding-right: 16px;
}
```

**Маппинг:**

- `margin-left` → `margin-inline-start`
- `margin-right` → `margin-inline-end`
- `padding-left` → `padding-inline-start`
- `padding-right` → `padding-inline-end`
- `border-left` → `border-inline-start`
- `border-right` → `border-inline-end`

**Примечание:** В проекте настроен линтер `stylelint-use-logical`, который проверяет использование логических свойств.

## Миксины

**Используйте миксины** для переиспользуемых стилей.

### Миксин `link`

Для стилизации ссылок:

```css
@mixin link;
```

### Миксин `layout`

Для адаптивной разметки:

```css
.container {
  @mixin layout 8px, 20px; /* компактные/широкие отступы */
}
```

### Брейкпоинты

Используйте миксины брейкпоинтов для адаптивных стилей. Применяй нужный mixin под целевой диапазон ширины.

**Доступные брейкпоинты:**

#### `mobileOnly`

**Значение:** `max-width: 680px`

**Назначение:** Стили применяются **только на мобильных устройствах** (до 680px включительно).

```css
.element {
  padding: 8px;

  @mixin mobileOnly {
    padding: 4px;
    font-size: 12px;
  }
}
```

**Когда использовать:**

- Скрытие элементов на мобильных (`display: none`)
- Изменение layout на мобильных (flex-direction: column)
- Уменьшение отступов и размеров шрифтов
- Специфичные стили только для мобильных устройств

#### `tabletUp`

**Значение:** `min-width: 681px`

**Назначение:** Стили применяются **на планшетах и десктопах** (от 681px и выше).

```css
.element {
  padding: 8px;

  @mixin tabletUp {
    padding: 20px;
    display: flex;
  }
}
```

**Когда использовать:**

- Показ элементов на планшетах и десктопах (`display: block`, `display: flex`)
- Изменение layout для больших экранов
- Увеличение отступов и размеров
- Горизонтальная ориентация элементов (flex-direction: row)

#### `desktopUp`

**Значение:** `min-width: 1075px`

**Назначение:** Стили применяются **только на десктопах** (от 1075px и выше).

```css
.element {
  padding: 16px;

  @mixin desktopUp {
    padding: 24px;
    max-width: 1200px;
    column-count: 2;
  }
}
```

**Когда использовать:**

- Специфичные стили только для десктопов
- Многоколоночная верстка (column-count)
- Увеличенные отступы и размеры для больших экранов
- Оптимизация layout для десктопов

**Диапазоны размеров экранов:**

В проекте принято разделение на 4 типа девайсов:

- **Mobile:** 320px - 680px
- **Tablet:** 681px - 1074px
- **Desktop:** 1075px - 1450px
- **Wide Desktop:** 1451px и более

**Примеры использования:**

```css
/* Скрытие на мобильных, показ на планшетах и выше */
.tablet-up {
  display: none;

  @mixin tabletUp {
    display: block;
  }
}

/* Показ только на мобильных */
.mobile-only {
  display: block;

  @mixin tabletUp {
    display: none;
  }
}

/* Адаптивная верстка */
.container {
  flex-direction: column;
  padding: 8px;

  @mixin tabletUp {
    flex-direction: row;
    padding: 16px;
  }

  @mixin desktopUp {
    padding: 24px;
    max-width: 1200px;
  }
}

/* Многоколоночная верстка на десктопе */
.content {
  @mixin desktopUp {
    column-count: 2;
    column-gap: 24px;
  }
}
```
