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
    PromoBanner.tsx
  model.ts               # опционально
  api/                   # опционально
```

**Переиспользуемые компоненты** (`core/shared/ui` и т.п.):

- `ComponentName/index.tsx`, `ComponentName.module.css`, `messages.ts`
- CSS Modules + `clsx`; **без Tailwind utility-классов**
- Переводы: `messages.ts` через `react-intl`

#### Формы
- Для форм используй `react-hook-form`.
- Ошибки показываются по blur (не при фокусе) — это поведение `Controller`.

