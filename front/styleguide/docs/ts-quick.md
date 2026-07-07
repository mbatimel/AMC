### TypeScript & React (Quick rules)

#### Imports & aliases
- Используй абсолютные импорты с алиасом `@/` → `src/`.
- Относительные импорты — внутри одного слайса (`./Home.module.css`, `./ui/Banner.tsx`).

#### Exports
- Именованные экспорты в слайсах `views/`, `widgets/`, `core/shared/ui`.
- `default export` — только в файлах маршрутов Next.js (`app/page.tsx`, `app/layout.tsx`, `app/not-found.tsx`).
- Code splitting — `dynamic` из `next/dynamic` + адаптер для именованного экспорта слайса.

#### Компоненты и функции
- Для React-компонентов используй стрелочные функции.
- Для компонентов явно указывай возвращаемый тип `JSX.Element`.
- Вложенные функции объявляй снизу (перед использованием), кроме event handlers.
- Если параметров больше 3 — передавай объектом.

#### Разметка текста
- **Отображаемый текст оборачивай в `<span>`** — не оставляй голый текст в JSX.
- **Исключение:** текст внутри `<button>` — можно без `<span>`.
- Семантический контейнер сохраняй (`h1`–`h6`, `label`), стили и строки — на `<span>` внутри.

```tsx
// 🚫
<p className={clsx(styles.description)}>{text}</p>
<h2 className={clsx(styles.title)}>{title}</h2>

// ✅
<button className={clsx(styles.button)} type="button">
  Войти
</button>
<h2 className={clsx(styles.title)}>
  <span>{title}</span>
</h2>
<span className={clsx(styles.description)}>{text}</span>
```

- Для UI-текста не используй `<p>` — только `<span>` с нужным классом.

#### Типы
- Используй `type` вместо `interface`.
- Используй `enum` для наборов именованных констант; `const enum` — когда значения не нужны в рантайме.
- Типы пропсов держи в том же файле, если они используются только в одном компоненте.
- Сортируй свойства объектов/типов по алфавиту — это делает `eslint-plugin-perfectionist` (см. Linting).

#### Linting
- `eslint-plugin-perfectionist` с конфигом `recommended-natural`.
- Сортировку не делай вручную — все правила плагина autofix через `eslint --fix`.

#### Naming conventions
- Переменные: `camelCase`
- Константы: `UPPER_SNAKE_CASE`
- Типы/компоненты: `PascalCase`
- Boolean: `is` / `has` / `should` / `can`
- Refs: суффикс `*Ref`
- Пропсы обработчики: префикс `on*`
- Внутренние обработчики: `handle*` или `on*Internal`

#### Структура

**Страницы** (`views/`):

```
views/Home/
  index.tsx              # export const Home
  Home.module.css
  ui/                    # опционально — только доп. компоненты
    PromoBanner/
      index.tsx
      PromoBanner.module.css
  model.ts               # опционально
  api/                   # опционально
```

**Компонент** (в `ui/` слайса или `core/shared/ui/`):

```
ComponentName/
  index.tsx              # export const ComponentName
  ComponentName.module.css
  messages.ts            # опционально
```

- Папка компонента — `PascalCase`, совпадает с именем компонента.
- Стили — `{ComponentName}.module.css` рядом с `index.tsx`.
- **Не** создавай `ComponentName.tsx` + `ComponentName.module.css` без папки.

**Иконки** (`core/shared/icons/`):

```
icons/
  types.ts               # IconProps
  CartIcon.tsx
  ShieldIcon.tsx
  index.ts               # barrel export
```

- Один файл на иконку: `IconName.tsx` с `export const IconName`.
- Пропсы — общий `IconProps`: `height`, `width`, `currentColor`, `className`.
- `currentColor` передаётся в атрибут `fill` (и `stroke` для контурных иконок).
- **Не** создавай папку `IconName/index.tsx`.

**Переиспользуемые компоненты** (`core/shared/ui` и т.п.):

- Та же структура: `ComponentName/index.tsx`, `ComponentName/ComponentName.module.css`, `messages.ts`
- CSS Modules + `clsx`; **без Tailwind utility-классов**
- CSS-классы — **`camelCase`**
- Переводы: `messages.ts` через `react-intl`
