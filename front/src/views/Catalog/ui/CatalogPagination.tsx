'use client';

import clsx from 'clsx';

import { IconChevronRight } from '@/core/shared/icons/IconChevronRight';

import styles from './CatalogPagination.module.css';

type CatalogPaginationProps = {
  onPageChange: (page: number) => void;
  page: number;
  pageCount: number;
};

const buildPages = (page: number, pageCount: number): Array<'ellipsis' | number> => {
  if (pageCount <= 7) {
    return Array.from({ length: pageCount }, (_, index) => index + 1);
  }

  const pages = new Set<number>([1, page, page + 1, page - 1, pageCount]);
  const sorted = [...pages]
    .filter((value) => value >= 1 && value <= pageCount)
    .sort((a, b) => a - b);
  const result: Array<'ellipsis' | number> = [];

  sorted.forEach((value, index) => {
    const previous = sorted[index - 1];

    if (previous !== undefined && value - previous > 1) {
      result.push('ellipsis');
    }

    result.push(value);
  });

  return result;
};

export const CatalogPagination = ({
  onPageChange,
  page,
  pageCount,
}: CatalogPaginationProps): JSX.Element | null => {
  if (pageCount <= 1) {
    return null;
  }

  const pages = buildPages(page, pageCount);

  return (
    <nav aria-label="Страницы каталога" className={clsx(styles.root)}>
      <button
        aria-label="Предыдущая страница"
        className={clsx(styles.navButton)}
        disabled={page <= 1}
        onClick={() => onPageChange(page - 1)}
        type="button"
      >
        <IconChevronRight className={clsx(styles.chevronPrev)} height={14} width={14} />
        Назад
      </button>

      <ul className={clsx(styles.pages)}>
        {pages.map((item, index) =>
          item === 'ellipsis' ? (
            <li aria-hidden className={clsx(styles.ellipsis)} key={`ellipsis-${index}`}>
              …
            </li>
          ) : (
            <li key={item}>
              <button
                aria-current={item === page ? 'page' : undefined}
                className={clsx(styles.pageButton, item === page && styles.pageButtonActive)}
                onClick={() => onPageChange(item)}
                type="button"
              >
                {item}
              </button>
            </li>
          ),
        )}
      </ul>

      <button
        aria-label="Следующая страница"
        className={clsx(styles.navButton)}
        disabled={page >= pageCount}
        onClick={() => onPageChange(page + 1)}
        type="button"
      >
        Вперёд
        <IconChevronRight height={14} width={14} />
      </button>
    </nav>
  );
};
