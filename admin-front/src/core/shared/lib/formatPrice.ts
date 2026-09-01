export const formatPrice = (value: number): string => {
  const formatted = new Intl.NumberFormat('ru-RU', {
    maximumFractionDigits: 2,
    minimumFractionDigits: 2,
  }).format(value);

  return `${formatted}\u00A0₽`;
};

export const normalizeQtyToPackage = (qty: number, packageQty: number): number => {
  const step = Math.max(1, packageQty);

  if (qty <= 0) {
    return 0;
  }

  return Math.ceil(qty / step) * step;
};
