import { SELECTED_CITY_ID_STORAGE_KEY } from './constants';

export const readSelectedCityId = (): null | string => {
  if (typeof window === 'undefined') {
    return null;
  }

  const value = window.localStorage.getItem(SELECTED_CITY_ID_STORAGE_KEY);

  return value && value.length > 0 ? value : null;
};

export const writeSelectedCityId = (cityId: string): void => {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.setItem(SELECTED_CITY_ID_STORAGE_KEY, cityId);
};
