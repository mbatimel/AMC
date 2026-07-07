### Импорты

**Используйте абсолютные импорты с алиасами** для внутренних модулей проекта.

```typescript
// ✅ Правильно
import { AppPath } from '@/core/shared/router';
import { Home } from '@/views/Home';

// ❌ Неправильно - относительные импорты между слоями
import { Key } from '../../constants/key';
import useLayoutType from '../hooks/useLayoutType';
```

**Правила:**
- `@/` — алиас для `src/` (настроен в `tsconfig.json`)
- Относительные импорты — для файлов внутри одного слайса (`./Home.module.css`, `./ui/Sidebar.tsx`)
- Порядок импортов поддерживает `eslint-plugin-perfectionist` (правило `sort-imports`)

### Экспорты

**Используйте именованные экспорты** для компонентов и утилит.

```typescript
// ✅ Правильно
export const Dropdown = <T,>({ ... }: Props<T>): JSX.Element => {
  // ...
};

export type Props<T> = { ... };

// ❌ Неправильно - default export
const Dropdown = ({ ... }: Props): JSX.Element => {
  // ...
};
export default Dropdown;
```

**Исключение для Next.js:** файлы маршрутов (`app/page.tsx`, `app/not-found.tsx`) экспортируют `default`. Слайсы в `pages/` — по-прежнему с **именованным** экспортом.

**Code splitting** — `dynamic` из `next/dynamic`:

```typescript
import dynamic from 'next/dynamic';

const Home = dynamic(() => import('@/views/Home').then((module) => ({ default: module.Home })));

const Page = (): JSX.Element => <Home />;

export default Page;
```

### Стрелочные функции vs Function Declaration

**Используйте стрелочные функции** для компонентов React и большинства функций.

```jsx
// ✅ Правильно
const Component = ({ prop }: Props): JSX.Element => {
  return <div>{prop}</div>;
};

// ❌ Неправильно
function Component({ prop }: Props): JSX.Element {
  return <div>{prop}</div>;
}
```

**Спорный вопрос:** Для вложенных функций можно использовать function declaration, если это улучшает читаемость. В текущей кодовой базе используются оба подхода.

### Вложенные функции

**Объявляйте вложенные функции снизу** (перед использованием), если они не являются обработчиками событий.

```jsx
// ✅ Правильно
const Component = ({ onClick }: Props): JSX.Element => {
  const handleClick = useCallback(() => {
    processData();
  }, []);

  const processData = (): void => {
    // обработка данных
  };

  return <button onClick={handleClick}>Click</button>;
};
```

Для обработчиков событий (handlers) используйте `useCallback` и объявляйте их в начале компонента.

### Количество параметров функции

**Если функция принимает более 3 параметров, используйте объект** для передачи параметров.

```typescript
// ✅ Правильно
const createUser = ({
  email,
  name,
  password,
  role,
}: {
  email: string;
  name: string;
  password: string;
  role: string;
}): void => {
  // ...
};

// ❌ Неправильно - слишком много параметров
const createUser = (
  email: string,
  name: string,
  password: string,
  role: string
): void => {
  // ...
};
```

### Type vs Interface

**Используйте `type` вместо `interface`** для определения типов.

```typescript
// ✅ Правильно
export type Props<T> = {
  className?: string;
  children: React.ReactNode;
};

// ❌ Неправильно
export interface Props<T> {
  className?: string;
  children: React.ReactNode;
}
```

### Явное указание типов

**Явно указывайте тип возвращаемого значения функции**, особенно для компонентов React.

```jsx
// ✅ Правильно
const Component = ({ prop }: Props): JSX.Element => {
  return <div>{prop}</div>;
};

// ❌ Неправильно - оставлять на откуп TypeScript
const Component = ({ prop }: Props) => {
  return <div>{prop}</div>;
};
```

Для обычных функций можно полагаться на вывод типов TypeScript, если тип очевиден из контекста.

### Enum

**Используйте `enum`** для наборов именованных констант — они работают и как тип, и как контейнер значений.

```typescript
// ✅ Правильно
enum OrderingType {
  ASC = 'asc',
  DESC = 'desc',
}

export const foo = (orderingType: OrderingType) => {
  switch (orderingType) {
    case OrderingType.ASC:
      // ...
    case OrderingType.DESC:
      // ...
    default:
      // ...
  }
};
```

Для констант, которые не нужны в рантайме, используйте `const enum`:

```typescript
export const enum Cookie {
  CSRF_TOKEN = 'csrftoken',
  SESSION_ID = 'sessionid',
}
```

Альтернатива — union type, если нужен только тип без runtime-значений:

```typescript
type OrderingType = 'asc' | 'desc';
```

```typescript
// ❌ Неправильно — дублировать enum через as const + type
const OrderingType = {
  ASC: 'asc',
  DESC: 'desc',
} as const;

type OrderingType = (typeof OrderingType)[keyof typeof OrderingType];
```

### Типы пропсов

**Не выносите типы пропсов в отдельный файл**, если они используются только в одном компоненте. Определяйте их в том же файле, что и компонент.

```typescript
// ✅ Правильно
type Props = {
  className?: string;
  children: React.ReactNode;
};

const Component = ({ className, children }: Props): JSX.Element => {
  // ...
};

// ❌ Неправильно - отдельный файл для простых пропсов
// types.ts
export type Props = { ... };
```

Если тип пропсов используется в нескольких местах или является частью публичного API, можно вынести его в отдельный файл.

### Сортировка свойств

**Сортируйте свойства объектов, типов и JSX-пропсов** — в проекте это обеспечивает [`eslint-plugin-perfectionist`](https://perfectionist.dev/). Не сортируйте вручную: запустите `eslint --fix`.

Плагин покрывает, в частности:
- `sort-imports`, `sort-named-imports`, `sort-named-exports`
- `sort-object-types`, `sort-objects`
- `sort-jsx-props`
- `sort-enums`, `sort-union-types`, `sort-intersection-types`
- `sort-switch-case`

```typescript
// ✅ Правильно — после eslint --fix
type Props = {
  className?: string;
  disabled?: boolean;
  label?: string;
  onClick: () => void;
  value: string;
};

// ❌ Неправильно — неупорядоченные свойства
type Props = {
  value: string;
  onClick: () => void;
  className?: string;
  label?: string;
  disabled?: boolean;
};
```

### Именование переменных

#### Обычные переменные

**Используйте camelCase** для обычных переменных и констант внутри функций.

```typescript
// ✅ Правильно
const intl = useIntl();
const isMobile = useLayoutType() === LayoutType.MOBILE;
const debounceTimeoutHandle = useRef<number | null>(null);
const activeItemRef = useRef<HTMLDivElement>(null);
const isDropdownWithContent = !!(list.length || emptyText || loading);

// ❌ Неправильно
const Intl = useIntl();
const is_mobile = useLayoutType() === LayoutType.MOBILE;
const DebounceTimeoutHandle = useRef<number | null>(null);
```

#### Константы

**Используйте UPPER_SNAKE_CASE** для констант, которые не изменяются.

```typescript
// ✅ Правильно
const DEBOUNCE_DELAY = 400;
const INVALID_TIMEZONE_COOKIE_NAME = 'prtnrInvalidTimezone';
const AB_ARTICLES_SLUG = 'ab_selections_articles_slug';

// ❌ Неправильно
const debounceDelay = 400;
const invalidTimezoneCookieName = 'prtnrInvalidTimezone';
const AbArticlesSlug = 'ab_selections_articles_slug';
```

**Когда использовать:**
- Магические числа и строки
- Конфигурационные значения
- Имена для cookies, localStorage keys
- Feature flags и эксперименты

#### Boolean переменные

**Используйте префиксы `is`, `has`, `should`, `can`** для boolean переменных.

```typescript
// ✅ Правильно
const isMobile = useLayoutType() === LayoutType.MOBILE;
const hasFocus = useRef(false);
const isDropdownWithContent = !!(list.length || emptyText || loading);
const isManualSubmission = onSubmit !== undefined;
const shouldShow = condition && otherCondition;
const canEdit = user?.permissions?.edit;

// ❌ Неправильно
const mobile = useLayoutType() === LayoutType.MOBILE;
const focus = useRef(false);
const dropdownWithContent = !!(list.length || emptyText || loading);
```

**Правила:**
- `is` - для состояний и проверок (`isLoading`, `isVisible`, `isActive`)
- `has` - для наличия чего-либо (`hasError`, `hasData`, `hasFocus`)
- `should` - для условий (`shouldShow`, `shouldRender`)
- `can` - для разрешений (`canEdit`, `canDelete`)

#### Refs

**Используйте суффикс `Ref`** для переменных, содержащих refs.

```typescript
// ✅ Правильно
const inputRef = useRef<HTMLInputElement>(null);
const activeItemRef = useRef<HTMLDivElement>(null);
const debounceTimeoutHandle = useRef<number | null>(null);

// ❌ Неправильно
const input = useRef<HTMLInputElement>(null);
const activeItem = useRef<HTMLDivElement>(null);
```

#### Обработчики событий

**Используйте префикс `on` для пропсов-обработчиков** и `handle` для внутренних обработчиков.

```typescript
// ✅ Правильно - пропсы компонента
type Props = {
  onClick: () => void;
  onFocus?: () => void;
  onChange: (value: string) => void;
};

// ✅ Правильно - внутренние обработчики
const handleClick = useCallback(() => {
  // логика
}, []);

const onInputFocus = useCallback((event: React.FocusEvent) => {
  // логика
}, []);

// ✅ Правильно - внутренние обработчики с суффиксом Internal
const onDropdownItemClickInternal = useCallback((index: number) => {
  onDropdownItemClick(index);
  onInputBlur();
}, []);

// ❌ Неправильно
const clickHandler = () => {}; // не используем такой формат
const onClickHandler = () => {}; // не используем такой формат
```

**Правила:**
- `on*` - для пропсов компонента (что передается извне)
- `handle*` или `on*` - для внутренних обработчиков
- `*Internal` - для внутренних обработчиков, которые оборачивают внешние

#### Функции

**Используйте camelCase** для имен функций.

```typescript
// ✅ Правильно
const getFreeCancellationTime = (date: Date): string => {
  // ...
};

const formatVatRate = (vatRate: number): string => {
  // ...
};

export const useBookLink = () => {
  // ...
};

// ❌ Неправильно
const GetFreeCancellationTime = () => {};
const format_vat_rate = () => {};
```

#### Типы и интерфейсы

**Используйте PascalCase** для типов, интерфейсов и type aliases.

```typescript
// ✅ Правильно
type Props = {
  className?: string;
};

type UserProfile = {
  name: string;
  email: string;
};

type EventTrackerContext = {
  trackEvent: () => void;
};

// ❌ Неправильно
type props = {};
type userProfile = {};
type eventTrackerContext = {};
```

#### Компоненты

**Используйте PascalCase** для компонентов React.

```typescript
// ✅ Правильно
export const Dropdown = <T,>({ ... }: Props<T>): JSX.Element => {
  // ...
};

export const TransferRates = ({ rates }: Props): JSX.Element => {
  // ...
};

// ❌ Неправильно
export const dropdown = () => {};
export const transferRates = () => {};
```

#### Переменные с объектами/массивами

**Используйте множественное число** для массивов и описательные имена для объектов.

```typescript
// ✅ Правильно
const rates: Transfer[] = [];
const users: User[] = [];
const settings: Settings = {};
const messages = defineMessages({});

// ❌ Неправильно
const rate: Transfer[] = [];
const user: User[] = [];
const setting: Settings = {};
```

#### Временные переменные

**Используйте короткие, но понятные имена** для временных переменных в циклах и коротких блоках.

```typescript
// ✅ Правильно
list.map((item, index) => (
  <div key={index}>{item.name}</div>
));

for (let i = 0; i < items.length; i += 1) {
  // ...
}

// ✅ Правильно - более описательные имена для сложной логики
const nextItemIndex = isUp
  ? dropdownListState.activeItemIndex - 1
  : dropdownListState.activeItemIndex + 1;

// ❌ Неправильно - слишком короткие имена вне циклов
const n = isUp ? idx - 1 : idx + 1;
```

### Props + JSX.Element vs `React.FC<Props>`

**Используйте явное указание типа возвращаемого значения `JSX.Element`** вместо `React.FC<Props>` или `FC<Props>`.

```jsx
// ✅ Правильно
export const Component = ({ prop }: Props): JSX.Element => {
  return <div>{prop}</div>;
};

// ❌ Неправильно
export const Component: FC<Props> = ({ prop }) => {
  return <div>{prop}</div>;
};
```

**Причины:**
- Более явное указание типа возвращаемого значения
- Не добавляет неявные пропсы (children) в типы
- Лучше работает с generics

### Generics в компонентах

Используйте generics для компонентов, которые работают с разными типами данных.

```typescript
// ✅ Правильно
export const Dropdown = <T,>({
  list,
  children,
}: Props<T>): JSX.Element => {
  // ...
};
```

### Названия переменных и пропсов

**Используйте понятные, описательные имена.**

```typescript
// ✅ Правильно
type Props = {
  isLoading: boolean;
  onUserClick: (userId: string) => void;
  selectedItems: string[];
};

// ❌ Неправильно
type Props = {
  loading: boolean;
  onClick: (id: string) => void;
  items: string[];
};
```

### Названия обработчиков

**Используйте префикс `on` для пропсов-обработчиков** и `handle` для внутренних обработчиков.

```jsx
// ✅ Правильно
type Props = {
  onClick: () => void; // проп компонента
  onFocus?: () => void;
};

const Component = ({ onClick }: Props): JSX.Element => {
  const handleClick = useCallback(() => {
    // внутренняя логика
    onClick();
  }, [onClick]);

  return <button onClick={handleClick}>Click</button>;
};

// ❌ Неправильно
type Props = {
  clickHandler: () => void; // не используем такой формат
};
```

**Правила:**
- `onClick`, `onFocus`, `onChange` - пропсы компонента (что передается извне)
- `handleClick`, `handleFocus`, `handleChange` - внутренние обработчики (логика внутри компонента)

### Структура компонента

#### Страницы и виджеты

**Корневой компонент** — `index.tsx` в папке слайса (`PascalCase`). Стили — `PageName.module.css` рядом.

```
views/Home/
  index.tsx              # export const Home
  Home.module.css
  ui/                    # только дополнительные части
    PromoBanner.tsx
  model.ts
  messages.ts

views/NotFound/
  index.tsx
  NotFound.module.css    # опционально
  ui/
    ErrorDetails.tsx
```

```typescript
// views/Home/index.tsx
import clsx from 'clsx';
import { Button } from '@heroui/react';

import styles from './Home.module.css';

export const Home = (): JSX.Element => (
  <main className={clsx(styles.root)}>
    <Button variant="primary">AMC</Button>
  </main>
);
```

```typescript
// views/NotFound/index.tsx
import { ErrorDetails } from './ui/ErrorDetails';

export const NotFound = (): JSX.Element => (
  <main>
    <ErrorDetails />
  </main>
);
```

#### Переиспользуемые компоненты

**Стандартная структура директории** (`core/shared/ui`, виджеты при выделении части в shared):

```
ComponentName/
  index.tsx              # Основной файл компонента
  ComponentName.module.css  # Стили компонента
  messages.ts            # Переводы компонента
  ComponentName.stories.tsx  # Storybook stories (опционально)
  assets/                # Изображения и другие ресурсы (опционально)
    image.svg
    icon.png
```

#### Файл стилей

**Стили компонента всегда хранятся в CSS Modules файле** с суффиксом `.module.css`.

Для **страниц/виджетов** — `PageName.module.css` в папке слайса (`views/Home/Home.module.css`).

Для **переиспользуемых компонентов** — `ComponentName.module.css` внутри папки компонента.

```typescript
// ✅ Правильно
// ComponentName/ComponentName.module.css
import clsx from 'clsx';
import styles from './ComponentName.module.css';

// ❌ Неправильно
import styles from './styles.css';
import styles from './Component.module.css'; // если директория называется ComponentName
```

#### Файл переводов

**Переводы для компонента всегда хранятся в файле `messages.ts`** в директории компонента.

```typescript
// messages.ts
import { defineMessages } from 'react-intl';

const messages = defineMessages({
  title: {
    id: 'component.title',
    defaultMessage: 'Заголовок',
  },
  description: {
    id: 'component.description',
    defaultMessage: 'Описание',
  },
});

export default messages;
```

**Использование переводов:**

```jsx
// В компоненте
import { useIntl } from 'react-intl';
import { FormattedMessage } from 'react-intl';
import messages from './messages';

const Component = (): JSX.Element => {
  const intl = useIntl();

  return (
    <div>
      {/* Через intl.formatMessage */}
      <h1>{intl.formatMessage(messages.title)}</h1>
      
      {/* Через FormattedMessage */}
      <p>
        <FormattedMessage {...messages.description} />
      </p>
    </div>
  );
};
```

**Когда использовать:**
- `intl.formatMessage` - для текста в атрибутах, переменных, сложных случаях
- `<FormattedMessage>` - для простого текста в JSX, когда нужна поддержка форматирования

#### Изображения и ресурсы

**Изображения и другие ресурсы кладем в директорию `assets/` внутри компонента.**

```jsx
// ✅ Правильно
import expiredImg from './assets/expired-pic.svg';
import icon from './assets/icon.png';

const Component = (): JSX.Element => {
  return <img src={expiredImg} alt="expired" />;
};

// ❌ Неправильно - не класть в общую папку assets
import icon from '../../assets/icons/icon.svg';
```

**Структура:**
```
ComponentName/
  assets/
    image.svg
    icon.png
    background.jpg
```

#### CSS Modules и clsx

**Используйте `clsx` для сборки `className` с CSS Modules.**

```jsx
import clsx from 'clsx';
import styles from './ComponentName.module.css';

type Props = {
  className?: string;
  isActive?: boolean;
  size?: 's' | 'm';
};

const Component = ({ className, isActive, size }: Props): JSX.Element => {
  return (
    <div
      className={clsx(
        styles.container,
        isActive && styles.container_active,
        size && styles[`container_size_${size}`],
        className,
      )}
    >
      Content
    </div>
  );
};
```

**Паттерны:**
- Базовый класс: `styles.block`
- Boolean-модификатор: `isActive && styles.block_active`
- Enum/union-модификатор: `size && styles[\`block_size_${size}\`]`
- Внешний `className` из пропсов — последним аргументом в `clsx`

---

#### Файл Storybook stories

**Storybook stories для компонента хранятся в файле `ComponentName.stories.tsx`** в директории компонента (опционально).

```jsx
// ComponentName.stories.tsx
import React from 'react';

import { Meta, StoryFn } from '@storybook/react';

import ComponentName from '.';

export default {
  component: ComponentName,
  title: 'components/ComponentName',
} as Meta<typeof ComponentName>;

export const Default: StoryFn<typeof ComponentName> = () => (
  <ComponentName prop="value" />
);
```

**Структура stories файла:**

```jsx
import React from 'react';

import { Meta, StoryFn } from '@storybook/react';

import ComponentName from '.';

// Метаданные story
export default {
  component: ComponentName,
  title: 'components/ComponentName', // Путь в Storybook
  argTypes: {
    // Настройки для Controls
    prop: {
      control: { type: 'text' },
      defaultValue: 'default value',
    },
  },
} as Meta<typeof ComponentName>;

// Основная story
export const Default: StoryFn<typeof ComponentName> = () => (
  <ComponentName prop="value" />
);

// Дополнительные варианты
export const WithCustomProps: StoryFn<typeof ComponentName> = () => (
  <ComponentName prop="custom" otherProp={123} />
);
```

**Правила:**
- Название файла: `ComponentName.stories.tsx` (название компонента + `.stories.tsx`)
- `title` должен отражать путь к компоненту в Storybook
- Используйте `StoryFn<typeof ComponentName>` для типизации stories
- Экспортируйте `Default` story как основную
- Добавляйте дополнительные stories для разных состояний компонента

**Когда создавать stories:**
- Для переиспользуемых компонентов
- Для компонентов с разными состояниями (loading, error, empty)
- Для компонентов, которые нужно демонстрировать в документации

---

### Custom Hooks

**Именование:** Все хуки начинаются с префикса `use`.

```typescript
// ✅ Правильно
export const useFetch = () => { ... };
export const useManager = () => { ... };
export const useTimeout = () => { ... };

// ❌ Неправильно
export const fetch = () => { ... };
export const manager = () => { ... };
```

**Структура хука:**

```typescript
import { useState, useEffect } from 'react';

export const useCustomHook = (param: string) => {
  const [state, setState] = useState<State>(initialState);

  useEffect(() => {
    // логика хука
    return () => {
      // cleanup
    };
  }, [param]);

  return { state, setState };
};
```

**Паттерны:**
- Возвращайте объект для множественных значений
- Используйте `useMemo` для мемоизации вычислений
- Очищайте ресурсы в `useEffect` cleanup (timeouts, subscriptions)

### Managers и API

**Managers - это классы для работы с API**, которые наследуются от базового `Manager`.

#### Структура Manager

```typescript
import Manager from '@managers/Manager';
import HttpClient from '@helpers/HttpClient';

export default class SelectionsManager extends Manager {
  getSelection(uid: string): Promise<Selection> {
    return this.httpClient.get<Selection>(
      `/selections/api/public/v1/selections/${uid}/`,
      undefined,
      undefined,
      'selection'
    );
  }

  updateSelection(uid: string, data: UpdateData): Promise<Selection> {
    return this.httpClient.post<Selection>(
      `/selections/api/public/v1/selections/${uid}/`,
      data,
      undefined,
      undefined,
      'updateSelection'
    );
  }
}
```

#### Использование в компонентах

**Используйте хук `useManager` для создания менеджеров:**

```typescript
import { useManager } from '@hooks/useManager';
import SelectionsManager from '@managers/Selections';

const Component = (): JSX.Element => {
  const manager = useManager(SelectionsManager);

  const handleLoad = async () => {
    const selection = await manager.getSelection(uid);
    // ...
  };
};
```

**Паттерны:**
- Все методы возвращают `Promise<T>`
- Используйте `CancellablePromise` для отмены запросов
- Обрабатывайте ошибки через `HttpClientError`

### Константы

**Организуйте константы по категориям** в `src/constants/`.

#### Структура

```
src/constants/
  cookie.ts          # Cookie имена
  url.ts            # API endpoints
  page.ts           # Типы страниц
  key.ts            # Клавиши
  maxLength.ts      # Максимальные длины
  Rate/             # Константы для рейтов
    amenity.ts
    meal.ts
```

#### Именование

```typescript
// ✅ Правильно — enum для именованных констант
export const enum Cookie {
  CSRF_TOKEN = 'csrftoken',
  SESSION_ID = 'sessionid',
}

// ✅ Правильно — объект с числовыми лимитами
export const maxLength = {
  NAME: 100,
  DESCRIPTION: 500,
};
```

### Модели данных

**Модели разделены на `backend/` и `frontend/`:**

```
src/models/
  backend/          # Типы данных из API
    selections/
      selection.ts
  frontend/         # Типы для UI
    Selection.ts
    User.ts
```

**Паттерны:**
- Backend модели соответствуют структуре API
- Frontend модели оптимизированы для UI
- Используйте сериализаторы для преобразования данных

### Парсеры

**Парсеры находятся в `src/parsers/`:**

```typescript
export const parseQueryParam = (param: string | string[] | undefined): string => {
  if (Array.isArray(param)) {
    return param[0];
  }
  return param || '';
};
```

**Паттерны:**
- Парсеры валидируют и трансформируют данные
- Обрабатывают edge cases (undefined, массивы)
- Возвращают типизированные значения

### Хелперы/Утилиты

**Утилиты находятся в `src/helpers/`:**

```typescript
// helpers/cookie.ts
export const getCookie = (name: string): string | undefined => {
  // ...
};

export const setCookie = (name: string, value: string, options?: CookieOptions): void => {
  // ...
};
```

**Паттерны:**
- Утилиты - чистые функции
- Экспортируйте именованными экспортами
- Группируйте связанные функции в один файл

### Динамические импорты

**Используйте `dynamic` из `next/dynamic` для code splitting страниц:**

```tsx
// app/page.tsx
import dynamic from 'next/dynamic';

const Home = dynamic(() => import('@/views/Home').then((module) => ({ default: module.Home })));

const Page = (): JSX.Element => <Home />;

export default Page;
```

**Когда использовать:**
- Страницы с тяжёлыми клиентскими зависимостями (HeroUI, графики)
- Виджеты, не нужные при первом рендере

### Типы и утилиты TypeScript

**Используйте utility types для работы с типами:**

```typescript
// Omit - исключить свойства
type PropsWithoutClassName = Omit<Props, 'className'>;

// Pick - выбрать свойства
type ButtonProps = Pick<Props, 'onClick' | 'disabled'>;

// Record - объект с ключами
type Config = Record<string, string>;
```

**Паттерны:**
- Используйте generics для переиспользуемых типов
- Создавайте утилитные типы для сложных случаев
- Документируйте сложные типы через JSDoc

---

## Other

### Обработка ошибок

#### ErrorBoundary

Используйте ErrorBoundary для обработки ошибок в React-компонентах.

```jsx
import { Component, ErrorInfo, ReactNode } from 'react';

class ErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean }
> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): { hasError: boolean } {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Логирование ошибки
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render(): ReactNode {
    if (this.state.hasError) {
      return <Error500 />;
    }

    return this.props.children;
  }
}
```

## Линтинг

В проекте используется ESLint с плагином [`eslint-plugin-perfectionist`](https://perfectionist.dev/) для единообразной сортировки кода. Все правила плагина autofix — не правьте порядок вручную, запускайте `eslint --fix`.

### Установка

```bash
yarn add -D eslint eslint-plugin-perfectionist
```

### Конфигурация

Рекомендуемый пресет — `recommended-natural` (natural sort, ascending):

```javascript
// eslint.config.js
import perfectionist from 'eslint-plugin-perfectionist';

export default [
  perfectionist.configs['recommended-natural'],
];
```

Альтернативы: `recommended-alphabetical`, `recommended-line-length`, `recommended-custom` — см. [документацию](https://perfectionist.dev/configs/).

### Что сортируется

| Правило | Что покрывает |
| --- | --- |
| `sort-imports` | Импорты и их группы |
| `sort-named-imports` / `sort-named-exports` | Именованные импорты и экспорты |
| `sort-object-types` / `sort-objects` | Свойства типов и объектов |
| `sort-jsx-props` | JSX-пропсы |
| `sort-enums` | Члены enum |
| `sort-union-types` / `sort-intersection-types` | Union и intersection types |
| `sort-switch-case` | Ветки `switch` |

### Практика

- Включите автофикс ESLint в редакторе (format on save) или в pre-commit hook.
- При ревью не придирайтесь к порядку свойств — это зона ответственности линтера.
- Если нужно исключение, настройте конкретное правило в `eslint.config.js`, а не отключайте плагин целиком.

