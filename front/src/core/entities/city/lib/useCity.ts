'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import type { CityItem } from '@/core/shared/api/orders';

import {
  $cities,
  $citiesError,
  $isCitiesPending,
  $selectedCityId,
  $selectedCityName,
  citiesHydrated,
  citySelected,
} from '../model';

export const useCity = (): {
  cities: CityItem[];
  citiesError: null | string;
  isCitiesPending: boolean;
  selectCity: (cityId: string) => void;
  selectedCityId: null | string;
  selectedCityName: string;
} => {
  const [
    cities,
    citiesError,
    isCitiesPending,
    selectedCityId,
    selectedCityName,
    hydrate,
    selectCity,
  ] = useUnit([
    $cities,
    $citiesError,
    $isCitiesPending,
    $selectedCityId,
    $selectedCityName,
    citiesHydrated,
    citySelected,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    cities,
    citiesError,
    isCitiesPending,
    selectCity,
    selectedCityId,
    selectedCityName: selectedCityName ?? cities[0]?.city ?? '',
  };
};
