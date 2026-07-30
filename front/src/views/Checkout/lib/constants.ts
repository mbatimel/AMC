import type { DeliveryType } from './types';

export const PICKUP_WAREHOUSE_ADDRESS =
  'г. Волжский, ул. Пушкина, 45, склад ООО ПО «Волжский инструмент»';

export const DELIVERY_OPTIONS: {
  description: string;
  label: string;
  value: DeliveryType;
}[] = [
  {
    description: 'Заберёте со склада в согласованный день',
    label: 'Самовывоз',
    value: 'pickup',
  },
  {
    description: 'Доставка собственной службой',
    label: 'Курьер',
    value: 'courier',
  },
  {
    description: 'Отправка транспортной компанией',
    label: 'Транспортная компания',
    value: 'transport',
  },
];

export const deliveryTypeLabel = (type: string): string =>
  DELIVERY_OPTIONS.find((option) => option.value === type)?.label ?? type;
