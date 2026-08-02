Используйте Effector для управления состоянием приложения. Ниже описаны правила и паттерны написания логики для проекта AMC.

## Именование

**Stores**

- префикс `$`:

  ```typescript
  export const $theme = createStore<Theme>({});
  export const $user = createStore<User | null>(null);
  export const $settings = createStore<Settings>({});
  ```

- избегаем обобщенных названий:
  ```typescript
  export const $store = createStore<Settings>({}); // 🚫
  export const $settings = createStore<Settings>({}); // ✅
  ```

**Events**

- стараемся именовать в прошедшем времени, чтобы имя отражало "что произошло", также избегаем сеттеров (`set<name>`)

  ```typescript
  export const setShiftR = createEvent<boolean>(); // 🚫
  export const shiftRUpdated = createEvent<boolean>(); // ✅

  export const toggleShiftR = createEvent<boolean>(); // 🚫
  export const shiftRToggled = createEvent<boolean>(); // ✅
  // если событие вызывается в ui слое - не зазорно это упомянуть
  export const shiftRButtonClicked = createEvent<boolean>(); // ✅
  ```

- желательно чтобы имя относилось к побудившему действию, а не к реализации
  ```ts
  const initIntl = createEvent(); // 🚫
  const appStarted = createEvent(); // ✅
  ```

**Effects** - суффикс `Fx`:

```typescript
const setShiftCookieFx = createEffect(() => {
  setCookie(Cookie.SHIFT_R, '1', { maxAge: 3600 });
});

const setupIntlFx = attach({
  source: { l10n: $l10n },
  effect: ({ l10n }) => createIntl({ locale: l10n.locale, messages: l10n.messages }),
});
```

## Создание stores

**Простой store:**

```typescript
import { createStore } from 'effector';

export const $page = createStore<Page | null>(null);
export const $user = createStore<User | null>(null);
```

**Store с начальным значением:**

```typescript
export const $settings = createStore<Settings>({
  [Setting.BLOG_FEED_URL]: '',
  [Setting.HELP_CENTER_URL]: '',
  // ...
});
```

## Обновление stores

**Через `.on()` для простых случаев:**

```typescript
export const $shiftR = createStore(false);
export const shiftRUpdated = createEvent<boolean>();
export const shiftRToggled = createEvent();

$shiftR.on(shiftRUpdated, (_, payload) => payload);
$shiftR.on(shiftRToggled, (state) => !state);
```

**Через `.on()` с трансформацией:**

```typescript
$theme.on(updateMainButtonColors, (theme, payload) => {
  const mainButtonColor = getMainButtonColors(payload);
  return {
    ...theme,
    'button-primary-bg': mainButtonColor.bg,
    // ...
  };
});
```

**Через `sample()` для сложной логики:**

```typescript
sample({
  clock: resetTheme,
  source: $defaultTheme,
  fn: (defaultTheme) => defaultTheme,
  target: $theme,
});
```

## Sample паттерны

**Простая передача события:**

```typescript
sample({
  clock: createNotification,
  target: createNotificationFx,
});
```

**С фильтрацией:**

```typescript
sample({
  clock: shiftRToggled,
  source: $shiftR,
  filter: (isActive) => isActive,
  target: setShiftCookieFx,
});

// Также filter может принимать значение типа Store
const $isShiftRApplicable = $userAreaLogin.map(
  (userAreaLogin) => !!userAreaLogin && isShiftRApplicable(userAreaLogin.currentContract),
);

sample({
  clock: shiftREvent,
  filter: $isShiftRApplicable,
  target: shiftRToggled,
});
```

**С трансформацией:**

```typescript
sample({
  clock: setupIntlFx.doneData,
  fn: (data) => mapData(data),
  target: $intl,
});
```

**С несколькими источниками:**

```typescript
sample({
  clock: add,
  source: {
    storeA: $storeA,
    storeB: $storeB,
  },
  fn: ({ storeA, storeB }, addValue) => ({
    combined: storeA + storeB,
    eventValue: addValue,
  }),
  target: calculateFx,
});
```

## Combine для вычисляемых значений

**Используйте `combine` для значений, зависящих от нескольких stores:**

```typescript
import { combine } from 'effector';

export const $areTransfersEnabled = combine(
  $settings,
  isTest(AB_TRANSFERS_SLUG),
  (settings, isInTest) => !!(settings[Setting.OWL_SELECTIONS_TRANSFERS_ENABLED] && isInTest),
);
```

## Effects

**Простой effect:**

```typescript
const setShiftCookieFx = createEffect(() => {
  setCookie(Cookie.SHIFT_R, '1', {
    maxAge: SHIFT_R_COOKIE_MAX_AGE,
    path: '/',
  });
});
```

**Effect через `attach` (для использования значений из stores):**

```typescript
const setupIntlFx = attach({
  source: { l10n: $l10n },
  effect: ({ l10n }) => {
    return createIntl({ locale: l10n.locale, messages: l10n.messages }, cache);
  },
});
```

**Effect с типами:**

```typescript
export const createNotificationFx = createEffect<NotificationConfig, void, Error>();

// Созданному таким способом эффекту хендлер можно добавить в другом месте
createNotificationFx.use(handler);
```

⚠️ явное указание типов юнитам - это вырожденный случай, обычно лучше позволить эффектору самому выводить типы.

## Использование в компонентах

**Один store:**

```typescript
import { useUnit } from 'effector-react';
import { $theme } from '@shared/model';

const Component = (): JSX.Element => {
  const theme = useUnit($theme);
  // ...
};
```

**Несколько stores:**

```typescript
const [settings, user, page] = useUnit([$settings, $user, $page]);
```

**Event:**

```jsx
const shiftRToggled = useUnit(shiftRToggled);

// Использование
<button onClick={() => shiftRToggled()}>Toggle</button>;
```

## Scope и изоляция состояния

**Создание scope через `fork`:**

```jsx
import { fork } from 'effector';
import { Provider } from 'effector-react';

const scope = fork({
  values: [
    [$user, user],
    [$settings, settings],
    [$theme, theme],
  ],
});

return <Provider value={scope}>{children}</Provider>;
```

**Выполнение эффектов в scope:**

```typescript
import { allSettled } from 'effector';

await allSettled(initIntl, { scope });
```

## Структура файлов

**Модели хранятся в `src/shared/model/`:**

```
src/shared/model/
  theme.ts      # $theme, updateMainButtonColors, resetTheme
  user.ts       # $user
  settings.ts   # $settings
  shiftR.ts     # $shiftR, shiftRUpdated, shiftRToggled
  intl.ts       # $l10n, $intl, initIntl
```

**Каждый файл экспортирует:**

- Stores (с префиксом `$`)
- Events (для обновления stores)
- Effects (если нужны сайд-эффекты)
- Computed stores через `combine` (если нужны)

**Пример структуры файла:**

```typescript
import { createEvent, createStore, sample } from 'effector';

// Типы
export type Theme = Record<string, string>;

// Stores
export const $theme = createStore<Theme>({});
export const $defaultTheme = createStore<Theme>({});

// Events
export const updateMainButtonColors = createEvent<SelectionMainColor>();
export const resetTheme = createEvent();

// Обновление stores
$theme.on(updateMainButtonColors, (theme, payload) => {
  // логика обновления
});

sample({
  clock: resetTheme,
  source: $defaultTheme,
  fn: (defaultTheme) => defaultTheme,
  target: $theme,
});
```

## API

Для работы с http-запросами мы используем [Farfetched](https://ff.effector.dev/)

Farfetched оперирует такими понятиями как query и mutation, и помимо выполнения запросов, берет на себя вопросы хранения данных, их трансформации и кеширования.

### Фабрики

Мы можем создать квери/мутации тремя способами:

**createJsonQuery**

```ts
export const bookingFormOrderQuery = createJsonQuery({
  params: declareParams<{ rateHash: string }>(),
  request: {
    method: 'GET',
    url: ({ rateHash }) => `/api/v2/orders/booking_form/${rateHash}/site/`,
  },
  response: {
    contract: noopContract<BookingFormResult>(),
  },
});
```

**createQuery** c кастомным хендлером или эффектом, например для использования с существующими менеджерами или для `multipart/form-data` запросов

```ts
export const brgMutation = createMutation({
  handler: async () => {
    // any async logic
  },
});

// wrap manager
const $manager = createStore(new SubAgentsManager());

export const loadSubAgentsQuery = createQuery({
  effect: attach({
    source: $manager,
    effect: async (manager, value: string) => {
      const data = await manager.getSubAgents(value);
      return data;
    },
  }),
});
```

**createQuery + createApiEffect** - для работы с готовой типизацией на основе openapi

```ts
import { paths } from './types'; // https://openapi-ts.dev/

const { createApiEffect } = createEffectorClient<paths>();

export const settingsQuery = createQuery({
  ...createApiEffect('get', '/partner/user_area/v2/common/login/settings'),
});
```

**operators**

Базовые операторы которые могут понадобиться при создании квери: `concurrency`, `retry`

```ts
export const settingsQuery = createQuery({
  ...createApiEffect('get', '/partner/user_area/v2/common/login/settings'),
});
concurrency(settingsQuery, { strategy: 'TAKE_EVERY' });

// у concurrency есть суперважная опция для отмены запросов
concurrency(settingsQuery, {
  strategy: 'TAKE_EVERY',
  abortAll: cancelQuery, // ⚠️
});

// ⚠️ у retry много опций, тут они опущены
retry(settingsQuery, { times: 5, delay: 3000 });
```

### Использование

```ts
// События и основные сторы

bookingFormOrderQuery.finished.success
bookingFormOrderQuery.finished.failure
bookingFormOrderQuery.finished.finally

bookingFormOrderQuery.$data
bookingFormOrderQuery.$error
bookingFormOrderQuery.$status // initial | pending | done | fail

sample({
  clock: orderUpdateMutation.finished.failure,
  fn: ({ intl, status }) => ({ type: 'error', title: 'Oops!' }) as const,
  target: showNotification,
});

createAction({
  // ⚠️ все опции за раз
  // done status: ok | done status: x | fail
  clock: orderStatusQuery.finished.finally,
  target: {
    statusReceived,
    statusCheckFailed,
  },
  fn(target, res) {
    if (res.status === 'done') {
      if (res.result.status === 'ok') {
        target.statusReceived({ status: 'ok', result: res.result });
      } else {
        ...
      }
    } else if (res.status === 'fail') {
      target.statusCheckFailed('http');
    }
  },
});
```

### Конфигурация

При создании квери или мутации можно сделать много всего, чтобы потом не заниматься этим в модели.

```ts
export const bookingFormOrderQuery = createJsonQuery({
  params: declareParams<{ rateHash: string }>(),
  request: {
    method: 'GET',
    url: ({ rateHash }) => `/api/v2/orders/booking_form/${rateHash}/site/`,
    headers: {
      // ⚠️ так можно добавить данные из сторов
      source: $contractSlug,
      fn: (_, slug) => ({ 'X-Partners-Contract-Slug': slug }),
    },
  },
  response: {
    // можно валидировать с помощью Contract API
    contract: noopContract<BookingFormResult>(),
    // ⚠️ c помощью validate колбека
    // можно обратить ответ 200 { status: 'error' } в ошибку
    validate: ({ result }) => result?.status !== 'ok',
    // трансформировать $data, отбросить лишнее
    mapData: ({ result }) => result.data,
  },
  // можно убрать nullable у $data: Data | null -> $data: Data
  // чтобы не делать лишний optional chaining в модели
  initialData: defaultState(),
});
```

⚠️ Конфигурации для `createQuery` и `createJsonQuery` отличаются, но по возможности настройки они идентичны.

```ts
export const settingsQuery = createQuery({
  name: 'settingsQuery',
  ...createApiEffect('get', '/partner/user_area/v2/common/login/settings', {
    mapParams: {
      source: $contractSlug,
      fn: (slug, init) => mergeInitHeaders(init, { 'X-Partners-Contract-Slug': slug }),
    },
  }),
  validate: validateEnvelope,
});
```
