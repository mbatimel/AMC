'use client';

import {
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownPopover,
  DropdownTrigger,
} from '@heroui/react';
import clsx from 'clsx';

import { IconChevronDown } from '@/core/shared/icons';

import styles from './FormSelect.module.css';

export type FormSelectOption = {
  label: string;
  value: string;
};

type FormSelectProps = {
  ariaLabel: string;
  className?: string;
  isDisabled?: boolean;
  label?: string;
  onChange: (value: string) => void;
  options: FormSelectOption[];
  placeholder?: string;
  value: string;
};

/**
 * Выпадающий список на базе HeroUI `Dropdown`.
 * Используется там, где нужен «select»: фильтры админки, формы поддержки и т.п.
 */
export const FormSelect = ({
  ariaLabel,
  className,
  isDisabled,
  label,
  onChange,
  options,
  placeholder = 'Не выбрано',
  value,
}: FormSelectProps): JSX.Element => {
  const selected = options.find((option) => option.value === value);

  return (
    <div className={clsx(styles.root, className)}>
      {label ? <span className={clsx(styles.label)}>{label}</span> : null}
      <Dropdown>
        <DropdownTrigger
          aria-label={ariaLabel}
          className={clsx(styles.trigger)}
          isDisabled={isDisabled || options.length === 0}
        >
          <span className={clsx(styles.value, !selected && styles.placeholder)}>
            {selected?.label ?? placeholder}
          </span>
          <IconChevronDown className={clsx(styles.icon)} height={12} width={12} />
        </DropdownTrigger>

        <DropdownPopover className={clsx(styles.popover)}>
          <DropdownMenu
            aria-label={ariaLabel}
            onSelectionChange={(keys) => {
              const next = Array.from(keys)[0];

              if (typeof next === 'string') {
                onChange(next);
              }
            }}
            selectedKeys={value ? [value] : []}
            selectionMode="single"
          >
            {options.map((option) => (
              <DropdownItem id={option.value} key={option.value} textValue={option.label}>
                {option.label}
              </DropdownItem>
            ))}
          </DropdownMenu>
        </DropdownPopover>
      </Dropdown>
    </div>
  );
};
