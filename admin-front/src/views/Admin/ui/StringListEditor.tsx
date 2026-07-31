'use client';

import clsx from 'clsx';

import styles from '../Admin.module.css';

type StringListEditorProps = {
  addLabel?: string;
  label: string;
  onChange: (items: string[]) => void;
  placeholder?: string;
  values: string[];
};

/** Редактор списка строк (пункты, преимущества, перечни документов). */
export const StringListEditor = ({
  addLabel = 'Добавить пункт',
  label,
  onChange,
  placeholder,
  values,
}: StringListEditorProps): JSX.Element => {
  const update = (index: number, value: string): void => {
    onChange(values.map((item, itemIndex) => (itemIndex === index ? value : item)));
  };

  const remove = (index: number): void => {
    onChange(values.filter((_, itemIndex) => itemIndex !== index));
  };

  return (
    <div className={clsx(styles.field)}>
      <span className={clsx(styles.label)}>{label}</span>
      <div className={clsx(styles.listEditor)}>
        {values.map((value, index) => (
          <div className={clsx(styles.listRow)} key={`item-${index}`}>
            <input
              aria-label={`${label}, пункт ${index + 1}`}
              className={clsx(styles.input)}
              onChange={(event) => update(index, event.target.value)}
              placeholder={placeholder}
              value={value}
            />
            <button
              className={clsx(styles.smallButton, styles.smallButtonDanger)}
              onClick={() => remove(index)}
              type="button"
            >
              Удалить
            </button>
          </div>
        ))}
        <button
          className={clsx(styles.smallButton)}
          onClick={() => onChange([...values, ''])}
          type="button"
        >
          {addLabel}
        </button>
      </div>
    </div>
  );
};
