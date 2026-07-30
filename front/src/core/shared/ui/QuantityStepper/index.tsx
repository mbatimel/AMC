'use client';

import { NumberField } from '@heroui/react';
import clsx from 'clsx';

import styles from './QuantityStepper.module.css';

type QuantityStepperProps = {
  className?: string;
  disabled?: boolean;
  onChange: (value: number) => void;
  step?: number;
  value: number;
};

export const QuantityStepper = ({
  className,
  disabled = false,
  onChange,
  step = 1,
  value,
}: QuantityStepperProps): JSX.Element => {
  const safeStep = Math.max(1, step);

  return (
    <NumberField
      aria-label="Количество"
      className={clsx(styles.root, className)}
      isDisabled={disabled}
      minValue={0}
      onChange={(next) => {
        onChange(typeof next === 'number' && Number.isFinite(next) ? next : 0);
      }}
      step={safeStep}
      value={value}
    >
      <NumberField.Group className={clsx(styles.group)}>
        <NumberField.DecrementButton
          aria-label="Уменьшить количество"
          className={clsx(styles.button)}
        >
          −
        </NumberField.DecrementButton>
        <NumberField.Input className={clsx(styles.input)} />
        <NumberField.IncrementButton
          aria-label="Увеличить количество"
          className={clsx(styles.button)}
        >
          +
        </NumberField.IncrementButton>
      </NumberField.Group>
    </NumberField>
  );
};
