'use client';

import { useEffect, useState } from 'react';

/**
 * Возвращает значение с задержкой `delayMs`.
 * При `delayMs <= 0` отдаёт актуальное значение сразу (удобно для мгновенного сброса).
 */
export const useDebouncedValue = <T>(value: T, delayMs: number): T => {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timeoutId = window.setTimeout(
      () => {
        setDebounced(value);
      },
      Math.max(0, delayMs),
    );

    return () => window.clearTimeout(timeoutId);
  }, [delayMs, value]);

  return delayMs <= 0 ? value : debounced;
};
