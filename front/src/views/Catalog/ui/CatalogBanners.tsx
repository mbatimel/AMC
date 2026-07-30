'use client';

import { Alert, Button, Chip } from '@heroui/react';
import clsx from 'clsx';

import styles from './CatalogBanners.module.css';

type BrandFilterChipProps = {
  brandName: string;
  onClear: () => void;
};

export const BrandFilterChip = ({ brandName, onClear }: BrandFilterChipProps): JSX.Element => {
  return (
    <div className={clsx(styles.chip)}>
      <Chip variant="soft">
        <Chip.Label>Бренд: {brandName}</Chip.Label>
      </Chip>
      <Button
        aria-label="Сбросить бренд"
        className={clsx(styles.chipClose)}
        isIconOnly
        onPress={onClear}
        size="sm"
        variant="ghost"
      >
        ×
      </Button>
    </div>
  );
};

export const OneCUnavailableBanner = (): JSX.Element => {
  return (
    <Alert className={clsx(styles.banner)} status="warning">
      <Alert.Indicator />
      <Alert.Content>
        <Alert.Title>1С недоступна</Alert.Title>
        <Alert.Description>Каталог показывает последние актуальные данные</Alert.Description>
      </Alert.Content>
    </Alert>
  );
};
