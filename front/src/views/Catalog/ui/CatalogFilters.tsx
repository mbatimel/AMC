'use client';

import { Button, Checkbox, Chip, Input, Label, TextField } from '@heroui/react';
import clsx from 'clsx';
import { useState } from 'react';

import type { CatalogFilters } from '../lib/filters';

import styles from './CatalogFilters.module.css';

type CatalogFiltersProps = {
  activeCount: number;
  filters: CatalogFilters;
  onChange: (patch: Partial<CatalogFilters>) => void;
  onReset: () => void;
};

export const CatalogFiltersPanel = ({
  activeCount,
  filters,
  onChange,
  onReset,
}: CatalogFiltersProps): JSX.Element => {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <section className={clsx(styles.root)}>
      <div className={clsx(styles.header)}>
        <div className={clsx(styles.headerLeft)}>
          <h2 className={clsx(styles.title)}>Фильтры</h2>
          {activeCount > 0 ? (
            <Chip className={clsx(styles.badge)} color="danger" size="sm">
              <Chip.Label>{activeCount}</Chip.Label>
            </Chip>
          ) : null}
        </div>
        <Button onPress={() => setIsOpen((value) => !value)} size="sm" variant="ghost">
          {isOpen ? 'Свернуть' : 'Развернуть'}
        </Button>
      </div>

      {isOpen ? (
        <>
          <div className={clsx(styles.grid)}>
            <TextField
              className={clsx(styles.field)}
              name="material"
              onChange={(value) => onChange({ material: value || undefined })}
              value={filters.material ?? ''}
            >
              <Label className={clsx(styles.label)}>Материал</Label>
              <Input className={clsx(styles.input)} placeholder="Все" />
            </TextField>
            <TextField
              className={clsx(styles.field)}
              name="size"
              onChange={(value) => onChange({ size: value || undefined })}
              value={filters.size ?? ''}
            >
              <Label className={clsx(styles.label)}>Размер / диаметр</Label>
              <Input className={clsx(styles.input)} placeholder="Все" />
            </TextField>
            <TextField
              className={clsx(styles.field)}
              name="gost"
              onChange={(value) => onChange({ gost: value || undefined })}
              value={filters.gost ?? ''}
            >
              <Label className={clsx(styles.label)}>ГОСТ / ТУ</Label>
              <Input className={clsx(styles.input)} placeholder="Все" />
            </TextField>
          </div>

          <div className={clsx(styles.footer)}>
            <Checkbox
              className={clsx(styles.checkbox)}
              isSelected={Boolean(filters.inStock)}
              onChange={(isSelected) => onChange({ inStock: isSelected || undefined })}
            >
              <Checkbox.Content>
                <Checkbox.Control>
                  <Checkbox.Indicator />
                </Checkbox.Control>
                Только в наличии
              </Checkbox.Content>
            </Checkbox>
            <Button onPress={onReset} size="sm" variant="outline">
              Сбросить
            </Button>
          </div>

          <div className={clsx(styles.profile)}>
            <div className={clsx(styles.profileHeader)}>
              <span>Профильные фильтры</span>
              <Chip className={clsx(styles.devBadge)} color="warning" size="sm">
                <Chip.Label>В разработке</Chip.Label>
              </Chip>
            </div>
            <p className={clsx(styles.profileHint)}>
              Выберите категорию — появятся профильные свойства.
            </p>
          </div>
        </>
      ) : null}
    </section>
  );
};
