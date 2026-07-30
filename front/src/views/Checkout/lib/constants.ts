import type { DeliveryType } from './types';

export const PICKUP_WAREHOUSE_ADDRESS =
  'г. Волжский, ул. Пушкина, 45, склад ООО ПО «Волжский инструмент»';

export const DELIVERY_OPTIONS: {
  label: (cityName: string) => string;
  value: DeliveryType;
}[] = [
  {
    label: (cityName) => `Самовывоз — ${cityName}, склад (бесплатно)`,
    value: 'pickup',
  },
  {
    label: () =>
      'Доставка «Волжский инструмент» (по городу, 500 ₽ — бесплатно при сумме от 10 000 ₽)',
    value: 'courier',
  },
  {
    label: () => 'Транспортная компания (по тарифам перевозчика)',
    value: 'transport',
  },
];

/** Короткая подпись для success-модалки (как в Figma). */
export const deliveryTypeShortLabel = (type: string): string => {
  switch (type) {
    case 'courier': {
      return 'Доставка по городу';
    }
    case 'pickup': {
      return 'Самовывоз';
    }
    case 'transport': {
      return 'Транспортная компания';
    }
    default: {
      return type;
    }
  }
};
