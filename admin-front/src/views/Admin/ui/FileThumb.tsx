'use client';

import clsx from 'clsx';
import { useEffect, useMemo, useState } from 'react';

import { isPdfFileUrl } from '@/core/shared/lib/fileToBase64';

import styles from './DocumentFileField.module.css';

type FileThumbProps = {
  alt?: string;
  file?: File | null;
  href?: string;
};

const extensionLabel = (name: string): string => {
  const index = name.lastIndexOf('.');

  return index >= 0 ? name.slice(index + 1).toUpperCase() : 'ФАЙЛ';
};

export const FileThumb = ({ alt = '', file, href }: FileThumbProps): JSX.Element | null => {
  const [brokenSrc, setBrokenSrc] = useState('');
  const objectUrl = useMemo(() => {
    if (!file?.type.startsWith('image/')) {
      return null;
    }

    return URL.createObjectURL(file);
  }, [file]);

  useEffect(() => {
    if (!objectUrl) {
      return;
    }

    return () => URL.revokeObjectURL(objectUrl);
  }, [objectUrl]);

  const isPdf = file
    ? file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf')
    : Boolean(href && isPdfFileUrl(href));
  const imageSrc = objectUrl ?? (href && !isPdf ? href : null);

  if (imageSrc && imageSrc !== brokenSrc) {
    return (
      <img
        alt={alt}
        className={clsx(styles.preview)}
        onError={() => setBrokenSrc(imageSrc)}
        src={imageSrc}
      />
    );
  }

  if (file || href) {
    return (
      <span className={clsx(styles.badge)}>
        {file ? extensionLabel(file.name) : isPdf ? 'PDF' : 'ФАЙЛ'}
      </span>
    );
  }

  return null;
};
