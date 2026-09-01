'use client';

import clsx from 'clsx';
import { useRef } from 'react';

import { sanitizeHtml } from '@/core/shared/lib/sanitizeHtml';

import styles from '../Admin.module.css';
import editorStyles from './HtmlEditor.module.css';

type HtmlEditorProps = {
  label: string;
  onChange: (value: string) => void;
  rows?: number;
  value: string;
};

const wrapSelection = (
  value: string,
  start: number,
  end: number,
  before: string,
  after: string,
): string => {
  const selected = value.slice(start, end) || 'текст';

  return `${value.slice(0, start)}${before}${selected}${after}${value.slice(end)}`;
};

export const HtmlEditor = ({ label, onChange, rows = 12, value }: HtmlEditorProps): JSX.Element => {
  const areaRef = useRef<HTMLTextAreaElement>(null);

  const applyWrap = (before: string, after: string): void => {
    const area = areaRef.current;
    const start = area?.selectionStart ?? value.length;
    const end = area?.selectionEnd ?? value.length;
    const next = sanitizeHtml(wrapSelection(value, start, end, before, after));

    onChange(next);
  };

  const applyLink = (): void => {
    const href = window.prompt('Ссылка (https://, mailto: или путь вида /catalog)', 'https://');

    if (!href) {
      return;
    }

    applyWrap(`<a href="${href}">`, '</a>');
  };

  return (
    <div className={clsx(styles.field)}>
      <span className={clsx(styles.label)}>{label}</span>
      <div className={clsx(editorStyles.toolbar)}>
        <button
          className={clsx(editorStyles.tool)}
          onClick={() => applyWrap('<strong>', '</strong>')}
          type="button"
        >
          Жирный
        </button>
        <button
          className={clsx(editorStyles.tool)}
          onClick={() => applyWrap('<em>', '</em>')}
          type="button"
        >
          Курсив
        </button>
        <button className={clsx(editorStyles.tool)} onClick={applyLink} type="button">
          Ссылка
        </button>
      </div>
      <textarea
        className={clsx(styles.textarea, editorStyles.area)}
        onBlur={() => onChange(sanitizeHtml(value))}
        onChange={(event) => onChange(event.target.value)}
        ref={areaRef}
        rows={rows}
        value={value}
      />
      <p className={clsx(styles.hint)}>
        Можно выделить текст и нажать «Жирный» или «Ссылка». Разрешены теги strong, em, a, списки и
        абзацы.
      </p>
    </div>
  );
};
