'use client';

import { Chip, CloseButton } from '@heroui/react';
import clsx from 'clsx';

import styles from './CatalogBanners.module.css';

type PromoFilterChipProps = {
  onClear: () => void;
  promotionName: string;
};

export const PromoFilterChip = ({ onClear, promotionName }: PromoFilterChipProps): JSX.Element => {
  return (
    <Chip className={clsx(styles.promoChip)} color="danger" variant="soft">
      <Chip.Label>Акция: {promotionName}</Chip.Label>
      <CloseButton
        aria-label="Сбросить акцию"
        className={clsx(styles.promoChipClose)}
        onPress={onClear}
      />
    </Chip>
  );
};
