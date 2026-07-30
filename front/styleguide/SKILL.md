---
name: styleguide
description: 'AMC Frontend стайлгайд — соглашения по TypeScript, React, CSS, Effector и FSD. Используй при написании, ревью или рефакторинге фронтенд-кода.'
---

# AMC Frontend Styleguide

Набор правил и соглашений для написания консистентного фронтенд-кода. Каждый раздел содержит примеры и обоснования.

## Когда применять

Используй этот skill когда:

- Пишешь или рефакторишь компоненты React
- Работаешь со стилями и адаптивной вёрсткой
- Реализуешь или ревьюишь логику на Effector
- Создаёшь API-запросы через Farfetched
- Организуешь структуру проекта по FSD
- Работаешь с формами (React Hook Form) и валидацией
- Создаёшь переводы и интернационализацию

## Разделы

| Раздел             | Файл                                             | Фокус                                                     |
| ------------------ | ------------------------------------------------ | --------------------------------------------------------- |
| TypeScript & React | [docs/ts-quick.md](docs/ts-quick.md)             | Импорты, экспорты, именование, компоненты, хуки, формы    |
| CSS                | [docs/css-quick.md](docs/css-quick.md)           | CSS-переменные, логические свойства, миксины, брейкпоинты |
| Effector           | [docs/effector-quick.md](docs/effector-quick.md) | Сторы, события, эффекты, sample, Farfetched               |
| FSD                | [docs/fsd-quick.md](docs/fsd-quick.md)           | Слои, стримы, сегменты, правила импортов, colocation      |

## Краткая справка

### TypeScript & React

- Именованные экспорты в слайсах; `default export` — только в `app/*.tsx` (маршруты Next.js)
- Стрелочные функции для компонентов, возвращаемый тип `JSX.Element`
- `type` вместо `interface`; для наборов констант — `enum` (предпочтительно) или `const enum`
- Тип пропсов в том же файле; сортировка свойств — через `eslint-plugin-perfectionist`
- Boolean: префиксы `is`/`has`/`should`/`can`
- Константы: `UPPER_SNAKE_CASE`; переменные: `camelCase`; типы/компоненты: `PascalCase`
- Пропсы-обработчики: `on*`; внутренние обработчики: `handle*` или `on*Internal`
- Рефы: суффикс `*Ref`
- Больше 3 параметров → объект
- CSS Modules: `import clsx from 'clsx'` + `import styles from '*.module.css'` + `className={clsx(styles.foo, ...)}`
- Переводы в `messages.ts` через `react-intl`
- **Страницы/виджеты**: `views/Home/index.tsx` + `Home.module.css`; папка слайса — `PascalCase`; `ui/` — только доп. компоненты
- **Переиспользуемые компоненты** (`core/shared/ui`): `ComponentName/index.tsx`, `ComponentName.module.css`, `messages.ts`

### CSS

- CSS-переменные из `theme.css` (`--palette-*`, `--border`, `--link`, `--focus-outline` и т.д.)
- **UI-компоненты** — обязательно опирайся на **HeroUI** (`@heroui/react`): Button, Input, Select, Modal, Tabs, Checkbox и т.д. Свой `<button>` / `<input>` / кастомный dropdown — только если в библиотеке нет подходящего примитива или поведение принципиально не покрывается
- **Tailwind utility-классы в своём коде не используем** — только CSS Modules + HeroUI
- Логические свойства (`margin-inline-start`) вместо физических (`margin-left`)
- Адаптивность через брейкпоинты-миксины: `mobileOnly` (≤680px), `tabletUp` (≥681px), `desktopUp` (≥1075px)

### Effector

- Сторы: префикс `$` (`$user`, `$settings`)
- События: прошедшее время (`userUpdated`, `buttonClicked`), без `set*`/`toggle*`
- Эффекты: суффикс `Fx` (`loadDataFx`)
- `sample` для сложной логики, `.on()` для простых обновлений
- `attach` для доступа к сторам в эффектах
- `useUnit` в компонентах
- API через Farfetched: `createJsonQuery`, `createQuery`, `createMutation`

### FSD (Feature-Sliced Design)

- Слои сверху вниз: `app/ → pages/ → widgets/ → entities/ → shared/`
- Импорт только из нижнего слоя в верхний
- Кросс-импорты внутри слоя запрещены (кроме shared/entities)
- Стримы — вертикальное деление; импорты между стримами запрещены (кроме core)
- Сегменты: `api/`, `model.ts`, `ui/`, `lib.ts`, `messages.ts`, `types.ts` — все опциональны
- **Страницы/виджеты**: корневой компонент в `PageName/index.tsx` в `views/`; папка слайса — `PascalCase`; `ui/` — только для вспомогательных компонентов
- **Colocation**: код рядом с местом использования; не разносить преждевременно по слоям

### Linting

- ESLint + [`eslint-plugin-perfectionist`](https://perfectionist.dev/) с конфигом `recommended-natural`
- Сортировка импортов, экспортов, типов, JSX-пропсов, enum-членов и т.д. — автоматически через `eslint --fix`

## Детальная информация

**Используй краткую справку выше** — её достаточно для большинства задач.

Детальные файлы (`docs/ts.md`, `docs/css.md`, `docs/effector.md`, `docs/fsd.md`) читай только когда нужно разобрать конкретное замечание по коду или нестандартную ситуацию, и правил из `*-quick.md` недостаточно.
