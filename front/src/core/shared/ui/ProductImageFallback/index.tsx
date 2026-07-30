'use client';

import clsx from 'clsx';

import styles from './ProductImageFallback.module.css';

type ProductImageFallbackProps = {
  categoryName?: string;
  className?: string;
  label?: string;
};

export const ProductImageFallback = ({
  categoryName,
  className,
  label,
}: ProductImageFallbackProps): JSX.Element => {
  const text = label ?? categoryName ?? 'Товар';

  return (
    <div aria-hidden="true" className={clsx(styles.root, className)}>
      <svg className={clsx(styles.icon)} fill="none" viewBox="0 0 64 64">
        <rect fill="#F0F2F5" height="64" rx="12" width="64" />
        <path
          d="M18 42L28 28L36 36L42 30L48 42H18Z"
          stroke="#9AA3B2"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
        />
        <circle cx="26" cy="22" fill="#9AA3B2" r="3" />
      </svg>
      <span className={clsx(styles.label)}>{text}</span>
    </div>
  );
};
