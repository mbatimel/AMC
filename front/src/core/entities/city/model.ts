import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { CityItem } from '@/core/shared/api/orders';

import { getCitiesRequest } from '@/core/shared/api/orders';

import { resolveDefaultCityId } from './lib/resolveDefaultCityId';
import { readSelectedCityId, writeSelectedCityId } from './lib/storage';

export const citiesHydrated = createEvent();
export const citySelected = createEvent<string>();

export const fetchCitiesFx = createEffect<void, CityItem[], Error>(() => getCitiesRequest());

const persistSelectedCityFx = createEffect((cityId: string) => {
  writeSelectedCityId(cityId);
});

export const $cities = createStore<CityItem[]>([]).on(
  fetchCitiesFx.doneData,
  (_, cities) => cities,
);

export const $selectedCityId = createStore<null | string>(null).on(
  citySelected,
  (_, cityId) => cityId,
);

export const $isCitiesPending = createStore(false)
  .on(fetchCitiesFx, () => true)
  .on(fetchCitiesFx.finally, () => false);

export const $citiesError = createStore<null | string>(null)
  .on(fetchCitiesFx, () => null)
  .on(fetchCitiesFx.failData, (_, error) => error.message)
  .on(citySelected, () => null);

export const $selectedCityName = createStore<null | string>(null);

const $shouldFetchCities = combine(
  $cities,
  fetchCitiesFx.pending,
  (cities, pending) => cities.length === 0 && !pending,
);

sample({
  clock: citiesHydrated,
  filter: $shouldFetchCities,
  target: fetchCitiesFx,
});

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> fn -> target */
sample({
  clock: fetchCitiesFx.doneData,
  source: $selectedCityId,
  fn: (selectedCityId, cities) => {
    const storedCityId = readSelectedCityId();
    const preferredCityId = selectedCityId ?? storedCityId;

    if (preferredCityId && cities.some((city) => city.id === preferredCityId)) {
      return preferredCityId;
    }

    return resolveDefaultCityId(cities);
  },
  target: citySelected,
});

sample({
  clock: citySelected,
  source: $cities,
  fn: (cities, cityId) => cities.find((city) => city.id === cityId)?.city ?? null,
  target: $selectedCityName,
});
/* eslint-enable perfectionist/sort-objects */

sample({
  clock: citySelected,
  target: persistSelectedCityFx,
});
