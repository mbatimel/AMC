'use client';

/* eslint-disable perfectionist/sort-imports -- react-scan must be imported before react */
import { scan } from 'react-scan';
import { useEffect } from 'react';

export const ReactScan = (): null => {
  useEffect(() => {
    scan({
      enabled: process.env.NODE_ENV === 'development',
    });
  }, []);

  return null;
};
