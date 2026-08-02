import type { Promotion } from '@/core/shared/api/promotions';

export const formatPromoEndsAt = (iso: string): string => {
  const date = new Date(iso);

  if (Number.isNaN(date.getTime())) {
    return '—';
  }

  return date.toLocaleString('ru-RU', {
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    month: '2-digit',
    year: 'numeric',
  });
};

export const formatProductsCount = (count: number): string => {
  const mod10 = count % 10;
  const mod100 = count % 100;

  if (mod10 === 1 && mod100 !== 11) {
    return `${count} товар в акции`;
  }

  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) {
    return `${count} товара в акции`;
  }

  return `${count} товаров в акции`;
};

export const promoNeedsQty = (promo: Promotion): boolean =>
  promo.products.some((product) => product.min_qty > 1);

export const promoConditionText = (promo: Promotion): string => {
  if (promoNeedsQty(promo)) {
    return `Скидка ${promo.discount_percent}% при покупке от минимального количества по товару`;
  }

  return `Скидка ${promo.discount_percent}% на выбранные товары в период акции`;
};
