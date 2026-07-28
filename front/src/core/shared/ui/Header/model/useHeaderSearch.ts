'use client';

import { useCallback, useState } from 'react';

import { HEADER_MOBILE_SEARCH_PLACEHOLDER, HEADER_SEARCH_PLACEHOLDER } from '../constants';

export const useHeaderSearch = () => {
  const [query, setQuery] = useState('');

  const onQueryChange = useCallback((value: string): void => {
    setQuery(value);
  }, []);

  const onSearchSubmit = useCallback((event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
  }, []);

  return {
    desktopPlaceholder: HEADER_SEARCH_PLACEHOLDER,
    mobilePlaceholder: HEADER_MOBILE_SEARCH_PLACEHOLDER,
    onQueryChange,
    onSearchSubmit,
    query,
  };
};
