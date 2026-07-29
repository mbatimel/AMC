'use client';

import { useRouter } from 'next/navigation';
import { useCallback, useState } from 'react';

import { AppPath } from '@/core/shared/router/paths';

import { HEADER_MOBILE_SEARCH_PLACEHOLDER, HEADER_SEARCH_PLACEHOLDER } from '../constants';

export const useHeaderSearch = () => {
  const router = useRouter();
  const [query, setQuery] = useState('');

  const onQueryChange = useCallback((value: string): void => {
    setQuery(value);
  }, []);

  const onSearchSubmit = useCallback(
    (event: React.FormEvent<HTMLFormElement>): void => {
      event.preventDefault();

      const trimmed = query.trim();
      const params = new URLSearchParams();

      if (trimmed) {
        params.set('q', trimmed);
      }

      const search = params.toString();

      router.push(search ? `${AppPath.Catalog}?${search}` : AppPath.Catalog);
    },
    [query, router],
  );

  return {
    desktopPlaceholder: HEADER_SEARCH_PLACEHOLDER,
    mobilePlaceholder: HEADER_MOBILE_SEARCH_PLACEHOLDER,
    onQueryChange,
    onSearchSubmit,
    query,
  };
};
