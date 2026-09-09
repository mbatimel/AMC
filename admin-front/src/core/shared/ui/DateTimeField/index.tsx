'use client';

import type { TimeValue } from '@heroui/react';

import { Calendar, DateField, DatePicker, Label, TimeField } from '@heroui/react';
import { CalendarDate, CalendarDateTime } from '@internationalized/date';
import clsx from 'clsx';
import { useRef, useState } from 'react';

import styles from './DateTimeField.module.css';

export type DateTimeFieldProps = {
  className?: string;
  error?: string;
  /** `day` → `YYYY-MM-DD`, `minute` (по умолчанию) → `YYYY-MM-DDTHH:mm` */
  granularity?: 'day' | 'minute';
  isInvalid?: boolean;
  label: string;
  onChange: (value: string) => void;
  value: string;
};

type DateParts = {
  day: number;
  hour?: number;
  minute?: number;
  month: number;
  year: number;
};

const pad = (value: number): string => String(value).padStart(2, '0');

const datePart = (value: string): string => {
  if (!value) {
    return '';
  }

  return value.includes('T') ? value.slice(0, 10) : value.slice(0, 10);
};

/** `YYYY-MM-DD` или `YYYY-MM-DDTHH:mm` → CalendarDate / CalendarDateTime */
const toDateValue = (local: string, withTime: boolean): CalendarDate | CalendarDateTime | null => {
  const date = datePart(local);

  if (!date) {
    return null;
  }

  const [year, month, day] = date.split('-').map(Number);

  if (!year || !month || !day) {
    return null;
  }

  if (!withTime) {
    return new CalendarDate(year, month, day);
  }

  const timePart = local.includes('T') ? (local.split('T')[1] ?? '00:00') : '00:00';
  const [hour = 0, minute = 0] = timePart.split(':').map(Number);

  return new CalendarDateTime(year, month, day, hour, minute);
};

/** DateParts → `YYYY-MM-DD` или `YYYY-MM-DDTHH:mm` */
const fromDateValue = (value: DateParts | null, withTime: boolean): string => {
  if (!value) {
    return '';
  }

  const date = `${value.year}-${pad(value.month)}-${pad(value.day)}`;

  if (!withTime) {
    return date;
  }

  const hour = value.hour ?? 0;
  const minute = value.minute ?? 0;

  return `${date}T${pad(hour)}:${pad(minute)}`;
};

/**
 * Дата (и опционально время) на базе HeroUI DatePicker.
 * - `granularity="day"` → `YYYY-MM-DD`
 * - `granularity="minute"` → `YYYY-MM-DDTHH:mm`
 */
export const DateTimeField = ({
  className,
  error,
  granularity = 'minute',
  isInvalid,
  label,
  onChange,
  value,
}: DateTimeFieldProps): JSX.Element => {
  const fieldRef = useRef<HTMLDivElement>(null);
  const [popoverWidth, setPopoverWidth] = useState<number>();
  const invalid = Boolean(isInvalid || error);
  const withTime = granularity === 'minute';

  return (
    <DatePicker
      className={clsx(styles.root, className)}
      granularity={granularity}
      hourCycle={24}
      isInvalid={invalid}
      onChange={(next) => onChange(fromDateValue(next, withTime))}
      onOpenChange={(isOpen) => {
        if (isOpen && fieldRef.current) {
          setPopoverWidth(fieldRef.current.getBoundingClientRect().width);
        }
      }}
      value={toDateValue(value, withTime)}
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
            {withTime ? (
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
            ) : null}
          </DatePicker.Popover>
        </>
      )}
    </DatePicker>
  );
};
