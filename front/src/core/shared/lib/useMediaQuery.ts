'use client';

import { useCallback, useSyncExternalStore } from 'react';

/** Подписка на CSS media query. На SSR возвращает `serverSnapshot` (по умолчанию false). */
export const useMediaQuery = (query: string, serverSnapshot = false): boolean => {
  const subscribe = useCallback(
    (onStoreChange: () => void): (() => void) => {
      const mediaQueryList = window.matchMedia(query);

      mediaQueryList.addEventListener('change', onStoreChange);

      return () => {
        mediaQueryList.removeEventListener('change', onStoreChange);
      };
    },
    [query],
  );

  const getSnapshot = useCallback((): boolean => window.matchMedia(query).matches, [query]);

  const getServerSnapshot = useCallback((): boolean => serverSnapshot, [serverSnapshot]);

  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
};
