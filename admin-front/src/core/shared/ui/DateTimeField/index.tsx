'use client';

import type { TimeValue } from '@heroui/react';
import type { DateValue } from '@internationalized/date';

import { Calendar, DateField, DatePicker, Label, TimeField } from '@heroui/react';
import { CalendarDateTime } from '@internationalized/date';
import clsx from 'clsx';
import { useRef, useState } from 'react';

import styles from './DateTimeField.module.css';

export type DateTimeFieldProps = {
  className?: string;
  error?: string;
  isInvalid?: boolean;
  label: string;
  onChange: (value: string) => void;
  value: string;
};

const pad = (value: number): string => String(value).padStart(2, '0');

/** `YYYY-MM-DDTHH:mm` → CalendarDateTime */
const toDateValue = (local: string): CalendarDateTime | null => {
  if (!local) {
    return null;
  }

  const [datePart, timePart = '00:00'] = local.split('T');
  const [year, month, day] = datePart.split('-').map(Number);
  const [hour = 0, minute = 0] = timePart.split(':').map(Number);

  if (!year || !month || !day) {
    return null;
  }

  return new CalendarDateTime(year, month, day, hour, minute);
};

/** CalendarDateTime → `YYYY-MM-DDTHH:mm` */
const fromDateValue = (value: DateValue | null): string => {
  if (!value) {
    return '';
  }

  const hour = 'hour' in value ? value.hour : 0;
  const minute = 'minute' in value ? value.minute : 0;

  return `${value.year}-${pad(value.month)}-${pad(value.day)}T${pad(hour)}:${pad(minute)}`;
};

/**
 * Дата и время (минуты) на базе HeroUI DatePicker.
 * Значение — локальная строка `YYYY-MM-DDTHH:mm`.
 */
export const DateTimeField = ({
  className,
  error,
  isInvalid,
  label,
  onChange,
  value,
}: DateTimeFieldProps): JSX.Element => {
  const fieldRef = useRef<HTMLDivElement>(null);
  const [popoverWidth, setPopoverWidth] = useState<number>();
  const invalid = Boolean(isInvalid || error);

  return (
    <DatePicker
      className={clsx(styles.root, className)}
      granularity="minute"
      hourCycle={24}
      isInvalid={invalid}
      onChange={(next) => onChange(fromDateValue(next))}
      onOpenChange={(isOpen) => {
        if (isOpen && fieldRef.current) {
          setPopoverWidth(fieldRef.current.getBoundingClientRect().width);
        }
      }}
      value={toDateValue(value)}
    >
      {({ state }) => (
        <>
          <Label className={clsx(styles.label)}>{label}</Label>
          <DateField.Group
            className={clsx(styles.group, invalid && styles.groupInvalid)}
            fullWidth
            ref={fieldRef}
          >
            <DateField.Input className={clsx(styles.input)}>
              {(segment) => <DateField.Segment segment={segment} />}
            </DateField.Input>
            <DateField.Suffix className={clsx(styles.suffix)}>
              <DatePicker.Trigger aria-label={`Открыть календарь: ${label}`}>
                <DatePicker.TriggerIndicator />
              </DatePicker.Trigger>
            </DateField.Suffix>
          </DateField.Group>
          {error ? <p className={clsx(styles.error)}>{error}</p> : null}
          <DatePicker.Popover
            className={clsx(styles.popover)}
            style={popoverWidth ? { width: popoverWidth } : undefined}
          >
            <Calendar aria-label={label} className={clsx(styles.calendar)}>
              <Calendar.Header>
                <Calendar.YearPickerTrigger aria-label={`Выбор года: ${label}`}>
                  <Calendar.YearPickerTriggerHeading />
                  <Calendar.YearPickerTriggerIndicator />
                </Calendar.YearPickerTrigger>
                <Calendar.NavButton aria-label="Предыдущий месяц" slot="previous" />
                <Calendar.NavButton aria-label="Следующий месяц" slot="next" />
              </Calendar.Header>
              <Calendar.Grid>
                <Calendar.GridHeader>
                  {(day) => <Calendar.HeaderCell>{day}</Calendar.HeaderCell>}
                </Calendar.GridHeader>
                <Calendar.GridBody>{(date) => <Calendar.Cell date={date} />}</Calendar.GridBody>
              </Calendar.Grid>
              <Calendar.YearPickerGrid>
                <Calendar.YearPickerGridBody>
                  {({ year }) => <Calendar.YearPickerCell year={year} />}
                </Calendar.YearPickerGridBody>
              </Calendar.YearPickerGrid>
            </Calendar>
            <div className={clsx(styles.timeBlock)}>
              <Label className={clsx(styles.label)}>Время</Label>
              <TimeField
                aria-label={`${label}: время`}
                granularity="minute"
                hourCycle={24}
                onChange={(next) => state.setTimeValue(next as TimeValue)}
                value={state.timeValue}
              >
                <TimeField.Group className={clsx(styles.group)} fullWidth>
                  <TimeField.Input className={clsx(styles.input)}>
                    {(segment) => <TimeField.Segment segment={segment} />}
                  </TimeField.Input>
                </TimeField.Group>
              </TimeField>
            </div>
          </DatePicker.Popover>
        </>
      )}
    </DatePicker>
  );
};
