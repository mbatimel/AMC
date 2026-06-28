### Effector (Quick rules)

#### Именование
- Stores: префикс `$` (например, `$user`, `$settings`).
- Events: прошедшее время (например, `userUpdated`, `buttonClicked`), без `set*`/`toggle*`.
- Effects: суффикс `Fx` (например, `loadDataFx`).

#### Паттерны логики
- `sample` — для сложной логики
- `.on()` — для простых обновлений store
- `attach` — когда effect должен брать данные из stores
- В компонентах используй `useUnit`

#### API-запросы
- Через Farfetched: `createJsonQuery`, `createQuery`, `createMutation`.

