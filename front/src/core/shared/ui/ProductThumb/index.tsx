'use client';

import clsx from 'clsx';
import { useState } from 'react';

import type { ProductImage } from '@/core/shared/api/products';

import { getPrimaryProductImage } from '@/core/shared/lib/productImage';
import { ProductImageFallback } from '@/core/shared/ui/ProductImageFallback';

import styles from './ProductThumb.module.css';

type ProductThumbProps = {
  alt: string;
  categoryName?: string;
  className?: string;
  fallbackClassName?: string;
  images?: null | ProductImage[];
  /** Готовый URL, если фото уже известно (корзина / кэш). */
  src?: null | string;
};

export const ProductThumb = ({
  alt,
  categoryName,
  className,
  fallbackClassName,
  images,
  src,
}: ProductThumbProps): JSX.Element => {
  const fromImages = getPrimaryProductImage(images)?.url;
  const url = src || fromImages || null;
  const [failed, setFailed] = useState(false);

  if (!url || failed) {
    return (
      <ProductImageFallback
        categoryName={categoryName}
        className={clsx(styles.fallback, fallbackClassName)}
        label=""
      />
    );
  }

  return (
    <img
      alt={alt}
      className={clsx(styles.image, className)}
      onError={() => setFailed(true)}
      src={url}
    />
  );
};
