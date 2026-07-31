'use client';

import clsx from 'clsx';
import { useRouter } from 'next/navigation';

import type { UseHeaderSearchResult } from '../../model/types';

import { IconSearch, IconSupport } from '@/core/shared/icons';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getProductPath } from '@/core/shared/router/paths';

import styles from './HeaderSearchSuggest.module.css';

type HeaderSearchSuggestProps = {
  search: UseHeaderSearchResult;
};

/** Живые подсказки под поисковой строкой + переход к ИИ-помощнику. */
export const HeaderSearchSuggest = ({ search }: HeaderSearchSuggestProps): JSX.Element | null => {
  const router = useRouter();

  if (!search.hasSuggest) {
    return null;
  }

  const hasQuery = search.query.trim().length > 0;
  const hasSuggestions = search.suggestions.length > 0;

  return (
    <div className={clsx(styles.root)}>
      {!hasQuery && search.recentQueries.length > 0 ? (
        <section className={clsx(styles.section)}>
          <p className={clsx(styles.sectionTitle)}>Недавние запросы</p>
          {search.recentQueries.map((recent) => (
            <button
              className={clsx(styles.row)}
              key={recent}
              onMouseDown={(event) => {
                event.preventDefault();
                search.onRecentSelect(recent);
              }}
              type="button"
            >
              <IconSearch className={clsx(styles.rowIcon)} height={14} width={14} />
              <span>{recent}</span>
            </button>
          ))}
        </section>
      ) : null}

      {hasQuery && hasSuggestions ? (
        <section className={clsx(styles.section)}>
          <p className={clsx(styles.sectionTitle)}>Найдено в каталоге</p>
          {search.suggestions.map((product) => (
            <button
              className={clsx(styles.row)}
              key={product.id}
              onMouseDown={(event) => {
                event.preventDefault();
                search.onSuggestClose();
                router.push(getProductPath(product.id));
              }}
              type="button"
            >
              <span className={clsx(styles.productMain)}>
                <span className={clsx(styles.productName)}>{product.name}</span>
                <span className={clsx(styles.productMeta)}>
                  {product.sku}
                  {product.gost ? ` · ${product.gost}` : ''}
                  {product.stock_qty > 0 ? ` · в наличии ${product.stock_qty}` : ' · под заказ'}
                </span>
              </span>
              <span className={clsx(styles.productPrice)}>
                {formatPrice(product.client_price || product.base_price)}
              </span>
            </button>
          ))}
        </section>
      ) : null}

      {hasQuery && !hasSuggestions && !search.isSuggestPending ? (
        <p className={clsx(styles.empty)}>По запросу ничего не найдено в каталоге.</p>
      ) : null}

      {hasQuery && search.isSuggestPending ? (
        <p className={clsx(styles.empty)}>Ищем…</p>
      ) : null}

      <button
        className={clsx(styles.assistant)}
        onMouseDown={(event) => {
          event.preventDefault();
          search.onAskAssistant();
        }}
        type="button"
      >
        <IconSupport className={clsx(styles.rowIcon)} height={14} width={14} />
        <span>
          {hasQuery
            ? `Спросить помощника: «${search.query.trim()}»`
            : 'Открыть подбор инструмента'}
        </span>
      </button>
    </div>
  );
};
