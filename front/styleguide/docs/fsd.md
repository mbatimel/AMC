https://feature-sliced.github.io/documentation/

FSD - это ряд правил и соглашений для организации кода.

В нашем случае мы стараемся не налегать на строгость и фокусируемся только на приятных нам частях методологии.

## Next.js

Проект на **Next.js App Router**. Папка `src/app/` совмещает слой FSD `app` и маршруты Next.js:

```bash
src/app/
  layout.tsx       # корневой layout, глобальные стили
  page.tsx         # маршрут / → подключает слайс views/Home
  not-found.tsx    # 404 → подключает слайс views/NotFound
```

UI и логика страниц — в `src/views/<PageName>/`. Файлы маршрутов — тонкие обёртки с `default export`.

UI и логика страниц живут в `src/views/<PageName>/` (слой FSD `pages`; папка `views`, потому что `src/pages/` зарезервирована Next.js).

Константы путей — `core/shared/router/paths.ts` (`AppPath` enum) для `Link` и редиректов.

## Слои

- проект горизонтально делится на слои от верхнего к нижнему:    
    ```bash
    project/
    ├─ app/
    ├─ pages/
    ├─ widgets/
    ├─ entities/ # не все слои обязательны.
    └─ shared/
    ```

- импортировать можно только из нижнего слоя в верхний:
  ```ts
  // widgets/status/model.ts
  import {SomeType} from '../pages/booking';  // 🚫 widgets -> pages
  import {SomeType} from '../shared/types';   // ✅ widgets -> shared
  ```
- внутри слоя импорты между модулями не допустимы\*:
  ```ts
  // pages/status/model.ts
  import {$orderToken} from '../pages/booking';   // 🚫 page -> page
  import {$orderToken} from '../entities/order';  // ✅ page -> entity
  ```
  _\* в shared и entities можно_

## Стримы

В нашем случае проекты еще вертикально делятся на стримы, где внутри каждого своя группа слоев.

Импорты между стримами запрещены, кроме импортов из core.

```bash
project/
├─ app/
├─ <stream_1>/
│   ├─ pages/
│   ├─ widgets/
│   ├─ entities/
│   └─ shared/
│
├─ <stream_2>/
│   ├─ pages/
│   └─ shared/
│
└─ core/ # (shared)
    ├─ entities/
    └─ shared/
```

## Сегменты

- модуль внутри слоя, например страница или виджет, делится на +- фиксированный список сегментов:

    ```bash
    project/views/<PageName>/
    ├─ index.tsx       # корневой компонент страницы (единственный вход)
    ├─ api/
    ├─ model.ts
    ├─ ui/             # только дополнительные компоненты
    │   └─ Sidebar.tsx
    ├─ lib.ts
    ├─ messages.ts
    └─ types.ts
    ```

    ```bash
    project/<stream>/widgets/<WidgetName>/
    ├─ index.tsx       # корневой компонент виджета
    ├─ api/
    ├─ model.ts
    ├─ ui/
    │   └─ WidgetPart.tsx
    ├─ lib.ts
    ├─ messages.ts
    └─ types.ts
    ```

- **корневой компонент** — `views/Home/index.tsx`; папка слайса в `PascalCase`; отдельный `index.ts` не нужен.
- все сегменты опциональны, каждый может быть как файлом так и директорией, в зависимости от ситуации и специфики конкретного модуля.
- вложенных сегментов не должно быть:
    ```bash
        some-widget/ui/
        ├─ api.ts        🚫
        ├─ model.ts      🚫
        └─ SubPart.tsx   ✅
    ```
    _Если компоненту просится свой model — скорее всего ему надо переезжать в виджеты._

- ui сегмент желательно держать плоским (избегать вложенности компонентов)

- у shared-слоя сегментов может быть много, в нашем случае мы еще напихиваем туда то, что по совести должно быть в entities:
    ```bash
    project/shared/
    ├─ config/ 
    ├─ router/ 
    ├─ ui/        # компоненты кандидаты на переезд в b2a-frontend-shared
    ├─ lib/       # кандидаты на переезд в b2a-frontend-shared/packages
    ├─ profile.ts # entity по факту, но для простоты можно и тут
    └─ notifications.ts
    ```

## Colocation

> Place code as close to where it's relevant as possible

_⚠️ Кажется у этого понятия нет распространённого имени.
Colocation из [этой статьи](https://kentcdodds.com/blog/colocation#the-principle) мне нравится._

FSD методология всегда относилась с уважением к этому принципу, а в [последней редакции он закреплён как основной](https://github.com/feature-sliced/documentation/releases/tag/v2.1):

- по умолчанию весь код относящийся к странице должен лежать в её сегментах, и выделяться на слои ниже по необходимости, а не по умолчанию.

- если пишете новую функциональность - не надо сразу разносить ее по проекту:

    ```bash
    project/
    ├─ entities/.../   # 🚫 запросы и модель
    ├─ widgets/.../    # 🚫 ui
    ├─ features/.../
    ├─ shared/lib/.../ # 🚫 хелперы
    │
    └─ /pages/<page>/
        └─ index.ts    # просто entry point
    ```

   Делайте все на странице: 

    ```bash
    project/
    ├─ ...
    └─ /views/<PageName>/
        ├─ index.tsx   # ✅ корневой компонент
        ├─ api/
        ├─ model.ts
        ├─ ui/         # ✅ только доп. компоненты
        └─ lib.ts
    ```
- Colocation не только про страницы, а вообще про всё: хелпер-функция для модели, компонента или любого другого модуля _может_ жить в том же файле очень долго.
