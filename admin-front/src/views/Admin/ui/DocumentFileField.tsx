'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useEffect, useRef, useState } from 'react';

import {
  assertDocumentFile,
  DOCUMENT_FILE_ACCEPT,
  DOCUMENT_FILE_HINT,
} from '@/core/shared/lib/fileToBase64';

import fieldStyles from '../Admin.module.css';
import styles from './DocumentFileField.module.css';
import { FileThumb } from './FileThumb';

type DocumentFileFieldProps = {
  currentFileHref?: string;
  currentFileLabel?: string;
  file: File | null;
  hint?: string;
  id: string;
  isDisabled?: boolean;
  isRequired?: boolean;
  label: string;
  onChange: (file: File | null) => void;
};

const ACCEPTED_EXTENSIONS = DOCUMENT_FILE_ACCEPT.split(',');

const isAcceptedName = (name: string): boolean => {
  const lowerName = name.toLowerCase();

  return ACCEPTED_EXTENSIONS.some((extension) => lowerName.endsWith(extension));
};

const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) {
    return `${bytes} Б`;
  }

  const kilobytes = bytes / 1024;

  if (kilobytes < 1024) {
    return `${kilobytes < 10 ? kilobytes.toFixed(1) : Math.round(kilobytes)} КБ`;
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`;
};

const fileExtension = (name: string): string => {
  const index = name.lastIndexOf('.');

  return index >= 0 ? name.slice(index + 1).toUpperCase() : 'ФАЙЛ';
};

export const DocumentFileField = ({
  currentFileHref,
  currentFileLabel = 'Текущий файл',
  file,
  hint = DOCUMENT_FILE_HINT,
  id,
  isDisabled = false,
  isRequired = false,
  label,
  onChange,
}: DocumentFileFieldProps): JSX.Element => {
  const inputRef = useRef<HTMLInputElement>(null);
  const dragCountRef = useRef(0);
  const [isDragging, setIsDragging] = useState(false);
  const [localError, setLocalError] = useState<null | string>(null);
  const hasCurrentFile = Boolean(currentFileHref);
  const hasFile = Boolean(file);

  useEffect(() => {
    if (!file && inputRef.current) {
      inputRef.current.value = '';
    }
  }, [file]);

  const openPicker = (): void => {
    if (isDisabled) {
      return;
    }

    inputRef.current?.click();
  };

  const applyFile = (next: File | null): void => {
    if (!next) {
      setLocalError(null);
      onChange(null);
      if (inputRef.current) {
        inputRef.current.value = '';
      }
      return;
    }

    try {
      if (!isAcceptedName(next.name)) {
        throw new Error('Можно загрузить PDF, JPG, PNG или WEBP');
      }
      assertDocumentFile(next);
      setLocalError(null);
      onChange(next);
    } catch (error) {
      setLocalError(error instanceof Error ? error.message : 'Не удалось выбрать файл');
    }
  };

  const handleDragEnter = (event: React.DragEvent<HTMLDivElement>): void => {
    event.preventDefault();
    if (isDisabled) {
      return;
    }
    dragCountRef.current += 1;
    setIsDragging(true);
  };

  const handleDragLeave = (event: React.DragEvent<HTMLDivElement>): void => {
    event.preventDefault();
    dragCountRef.current -= 1;
    if (dragCountRef.current <= 0) {
      dragCountRef.current = 0;
      setIsDragging(false);
    }
  };

  const handleDragOver = (event: React.DragEvent<HTMLDivElement>): void => {
    event.preventDefault();
  };

  const handleDrop = (event: React.DragEvent<HTMLDivElement>): void => {
    event.preventDefault();
    dragCountRef.current = 0;
    setIsDragging(false);
    if (isDisabled) {
      return;
    }

    const next = event.dataTransfer.files[0];

    if (next) {
      applyFile(next);
    }
  };

  const handleZoneClick = (event: React.MouseEvent<HTMLDivElement>): void => {
    if (hasFile || hasCurrentFile || isDisabled) {
      return;
    }
    if ((event.target as HTMLElement).closest('a, button')) {
      return;
    }
    openPicker();
  };

  return (
    <div className={clsx(fieldStyles.field)}>
      <label className={clsx(fieldStyles.label)} htmlFor={id}>
        {label}
        {isRequired ? <span className={clsx(styles.required)}> *</span> : null}
      </label>
      <div
        className={clsx(
          styles.dropzone,
          isDragging && styles.dropzoneActive,
          (hasFile || hasCurrentFile) && styles.dropzoneFilled,
          isDisabled && styles.dropzoneDisabled,
        )}
        onClick={handleZoneClick}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
      >
        <input
          accept={DOCUMENT_FILE_ACCEPT}
          className={clsx(styles.input)}
          disabled={isDisabled}
          id={id}
          onChange={(event) => applyFile(event.target.files?.[0] ?? null)}
          ref={inputRef}
          tabIndex={-1}
          type="file"
        />
        {file ? (
          <div className={clsx(styles.selected)}>
            <FileThumb file={file} href={currentFileHref} />
            <div className={clsx(styles.meta)}>
              <p className={clsx(styles.fileName)}>{file.name}</p>
              <p className={clsx(styles.fileSize)}>
                {fileExtension(file.name)} · {formatFileSize(file.size)}
              </p>
            </div>
            <div className={clsx(styles.actions)}>
              <Button isDisabled={isDisabled} onPress={openPicker} size="sm" variant="outline">
                Заменить
              </Button>
              <Button
                isDisabled={isDisabled}
                onPress={() => applyFile(null)}
                size="sm"
                variant="secondary"
              >
                Убрать
              </Button>
            </div>
          </div>
        ) : currentFileHref ? (
          <div className={clsx(styles.selected)}>
            <FileThumb href={currentFileHref} />
            <div className={clsx(styles.meta)}>
              <a
                className={clsx(styles.current)}
                href={currentFileHref}
                onClick={(event) => event.stopPropagation()}
                rel="noopener noreferrer"
                target="_blank"
              >
                {currentFileLabel}
              </a>
              <p className={clsx(styles.fileSize)}>Загруженный файл</p>
            </div>
            <div className={clsx(styles.actions)}>
              <Button isDisabled={isDisabled} onPress={openPicker} size="sm" variant="outline">
                Заменить
              </Button>
            </div>
          </div>
        ) : (
          <div className={clsx(styles.empty)}>
            <p className={clsx(styles.emptyTitle)}>
              {isDragging ? 'Отпустите файл, чтобы загрузить' : 'Перетащите файл сюда'}
            </p>
            <Button isDisabled={isDisabled} onPress={openPicker} size="sm" variant="outline">
              Выбрать файл
            </Button>
            <p className={clsx(fieldStyles.hint)}>{hint}</p>
          </div>
        )}
      </div>
      {localError ? <p className={clsx(fieldStyles.error)}>{localError}</p> : null}
    </div>
  );
};
