export const pluralize = (count: number, one: string, few: string, many: string): string => {
  const absolute = Math.abs(count) % 100;
  const lastDigit = absolute % 10;

  if (absolute > 10 && absolute < 20) {
    return many;
  }

  if (lastDigit === 1) {
    return one;
  }

  if (lastDigit >= 2 && lastDigit <= 4) {
    return few;
  }

  return many;
};

export const formatPositionsCount = (count: number): string =>
  `${count} ${pluralize(count, 'позиция', 'позиции', 'позиций')}`;

export const formatOrdersCount = (count: number): string =>
  `${count} ${pluralize(count, 'заказ', 'заказа', 'заказов')}`;
