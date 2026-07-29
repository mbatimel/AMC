'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';

import styles from './CatalogEmptyState.module.css';

type CatalogEmptyStateProps = {
  onReset: () => void;
};

export const CatalogEmptyState = ({ onReset }: CatalogEmptyStateProps): JSX.Element => {
  return (
    <div className={clsx(styles.root)}>
      <div aria-hidden="true" className={clsx(styles.icon)}>
        ×
      </div>
      <h2 className={clsx(styles.title)}>Ничего не найдено</h2>
      <p className={clsx(styles.text)}>Попробуйте изменить фильтры или поисковый запрос</p>
      <Button className={clsx(styles.button)} onPress={onReset} variant="outline">
        Сбросить фильтры
      </Button>
    </div>
  );
};
