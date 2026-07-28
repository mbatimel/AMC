import type { CityItem } from '@/core/shared/api/orders';

export const resolveDefaultCityId = (cities: CityItem[]): string => {
  return cities[0].id;
};
