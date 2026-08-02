'use client';

import clsx from 'clsx';

import { formatPositionsCount } from '@/core/shared/lib/pluralize';

import type { CatalogViewMode } from '../lib/filters';

import styles from './CatalogToolbar.module.css';

type CatalogToolbarProps = {
  canExport: boolean;
  hideViewToggle?: boolean;
  onExportXls: () => void;
  onPrint: () => void;
  onViewChange: (view: CatalogViewMode) => void;
  total: number;
  view: CatalogViewMode;
};

const IconTable = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M2.5 3.5h11v9h-11v-9Z"
      stroke="currentColor"
      strokeLinejoin="round"
      strokeWidth="1.4"
    />
    <path d="M2.5 6.5h11M6.5 3.5v9" stroke="currentColor" strokeWidth="1.4" />
  </svg>
);

const IconCards = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M3 3h4v4H3V3ZM9 3h4v4H9V3ZM3 9h4v4H3V9ZM9 9h4v4H9V9Z"
      stroke="currentColor"
      strokeLinejoin="round"
      strokeWidth="1.4"
    />
  </svg>
);

const IconDownload = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M8 2.5v7m0 0 2.5-2.5M8 9.5 5.5 7M3.5 12.5h9"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.4"
    />
  </svg>
);

const IconPrint = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M4.5 6V3.5h7V6M4.5 10.5H3A1.5 1.5 0 0 1 1.5 9V7A1.5 1.5 0 0 1 3 5.5h10A1.5 1.5 0 0 1 14.5 7v2A1.5 1.5 0 0 1 13 10.5h-1.5M4.5 9.5h7v3h-7v-3Z"
      stroke="currentColor"
      strokeLinejoin="round"
      strokeWidth="1.4"
    />
  </svg>
);

export const CatalogToolbar = ({
  canExport,
  hideViewToggle = false,
  onExportXls,
  onPrint,
  onViewChange,
  total,
  view,
}: CatalogToolbarProps): JSX.Element => {
  return (
    <div className={clsx(styles.root)}>
      <h1 className={clsx(styles.title)}>Каталог</h1>

      <div className={clsx(styles.controls)}>
        <span className={clsx(styles.count)}>{formatPositionsCount(total)}</span>

        {hideViewToggle ? null : (
          <div aria-label="Вид каталога" className={clsx(styles.viewToggle)} role="group">
            <button
              aria-pressed={view === 'table'}
              className={clsx(styles.viewButton, view === 'table' && styles.viewButtonActive)}
              onClick={() => onViewChange('table')}
              type="button"
            >
              <IconTable />
              B2B-таблица
            </button>
            <button
              aria-pressed={view === 'cards'}
              className={clsx(styles.viewButton, view === 'cards' && styles.viewButtonActive)}
              onClick={() => onViewChange('cards')}
              type="button"
            >
              <IconCards />
              Карточки
            </button>
          </div>
        )}

        <button
          className={clsx(styles.actionButton)}
          disabled={!canExport}
          onClick={onExportXls}
          type="button"
        >
          <IconDownload />
          Прайс XLS
        </button>
        <button className={clsx(styles.actionButton)} onClick={onPrint} type="button">
          <IconPrint />
          Печать
        </button>
      </div>
    </div>
  );
};
