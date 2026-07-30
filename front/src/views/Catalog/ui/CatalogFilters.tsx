'use client';

import { Checkbox, Input, Label, TextField } from '@heroui/react';
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
          {activeCount > 0 ? <span className={clsx(styles.activeCount)}>{activeCount}</span> : null}
        </div>
        <button
          className={clsx(styles.textAction)}
          onClick={() => setIsOpen((value) => !value)}
          type="button"
        >
          {isOpen ? 'Свернуть' : 'Развернуть'}
        </button>
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
              <Checkbox.Content className={clsx(styles.checkboxContent)}>
                <Checkbox.Control className={clsx(styles.checkboxControl)}>
                  <Checkbox.Indicator />
                </Checkbox.Control>
                <span className={clsx(styles.checkboxLabel)}>Только в наличии</span>
              </Checkbox.Content>
            </Checkbox>
            <button className={clsx(styles.textAction)} onClick={onReset} type="button">
              Сбросить
            </button>
          </div>

          <div className={clsx(styles.profile)}>
            <div className={clsx(styles.profileHeader)}>
              <span>Профильные фильтры</span>
              <span className={clsx(styles.devBadge)}>В разработке</span>
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
