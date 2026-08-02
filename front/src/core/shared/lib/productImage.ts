import type { ProductImage } from '@/core/shared/api/products';

/** Главное фото товара: `is_primary`, иначе первое в массиве. */
export const getPrimaryProductImage = (
  images: null | ProductImage[] | undefined,
): null | ProductImage => {
  if (!images || images.length === 0) {
    return null;
  }

  return images.find((item) => item.is_primary) ?? images[0] ?? null;
};

export const getPrimaryProductImageUrl = (
  images: null | ProductImage[] | undefined,
): null | string => getPrimaryProductImage(images)?.url ?? null;
