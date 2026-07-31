'use client';

import { useUnit } from 'effector-react';
import { useRouter } from 'next/navigation';
import { useCallback, useEffect, useRef, useState } from 'react';

import type { ProductListItem } from '@/core/shared/api/products';

import { assistantMessageSent, assistantOpened } from '@/core/entities/assistant';
import { listProductsRequest } from '@/core/shared/api/products';
import { AppPath } from '@/core/shared/router/paths';

import { HEADER_MOBILE_SEARCH_PLACEHOLDER, HEADER_SEARCH_PLACEHOLDER } from '../constants';

const SUGGEST_DEBOUNCE_MS = 250;
const SUGGEST_MIN_LENGTH = 2;
const SUGGEST_LIMIT = 6;
const RECENT_STORAGE_KEY = 'amc_search_history';
const RECENT_LIMIT = 5;

const readRecentQueries = (): string[] => {
  if (typeof window === 'undefined') {
    return [];
  }

  try {
    const raw = window.localStorage.getItem(RECENT_STORAGE_KEY);
    const parsed: unknown = raw ? JSON.parse(raw) : [];

    return Array.isArray(parsed) ? parsed.filter((item): item is string => typeof item === 'string') : [];
  } catch {
    return [];
  }
};

const writeRecentQueries = (queries: string[]): void => {
  try {
    window.localStorage.setItem(RECENT_STORAGE_KEY, JSON.stringify(queries));
  } catch {
    // приватный режим — история не сохраняется
  }
};

/**
 * Поиск в шапке: живые подсказки по каталогу, история запросов и
 * передача запроса ИИ-помощнику, если ничего не нашлось.
 */
export const useHeaderSearch = () => {
  const router = useRouter();
  const [openAssistant, sendToAssistant] = useUnit([assistantOpened, assistantMessageSent]);
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<ProductListItem[]>([]);
  const [recentQueries, setRecentQueries] = useState<string[]>([]);
  const [isSuggestOpen, setIsSuggestOpen] = useState(false);
  const [isSuggestPending, setIsSuggestPending] = useState(false);
  const requestIdRef = useRef(0);

  useEffect(() => {
    setRecentQueries(readRecentQueries());
  }, []);

  useEffect(() => {
    const trimmed = query.trim();

    if (trimmed.length < SUGGEST_MIN_LENGTH) {
      setSuggestions([]);
      setIsSuggestPending(false);

      return;
    }

    setIsSuggestPending(true);

    const requestId = requestIdRef.current + 1;

    requestIdRef.current = requestId;

    const timer = setTimeout(() => {
      listProductsRequest({ limit: SUGGEST_LIMIT, q: trimmed })
        .then((result) => {
          if (requestIdRef.current === requestId) {
            setSuggestions(result.items);
          }
        })
        .catch(() => {
          if (requestIdRef.current === requestId) {
            setSuggestions([]);
          }
        })
        .finally(() => {
          if (requestIdRef.current === requestId) {
            setIsSuggestPending(false);
          }
        });
    }, SUGGEST_DEBOUNCE_MS);

    return () => {
      clearTimeout(timer);
    };
  }, [query]);

  const rememberQuery = useCallback((value: string): void => {
    setRecentQueries((previous) => {
      const next = [value, ...previous.filter((item) => item !== value)].slice(0, RECENT_LIMIT);

      writeRecentQueries(next);

      return next;
    });
  }, []);

  const onQueryChange = useCallback((value: string): void => {
    setQuery(value);
    setIsSuggestOpen(true);
  }, []);

  const onSuggestOpen = useCallback((): void => {
    setIsSuggestOpen(true);
  }, []);

  const onSuggestClose = useCallback((): void => {
    setIsSuggestOpen(false);
  }, []);

  const runSearch = useCallback(
    (value: string): void => {
      const trimmed = value.trim();
      const params = new URLSearchParams();

      if (trimmed) {
        params.set('q', trimmed);
        rememberQuery(trimmed);
      }

      const search = params.toString();

      setIsSuggestOpen(false);
      router.push(search ? `${AppPath.Catalog}?${search}` : AppPath.Catalog);
    },
    [rememberQuery, router],
  );

  const onSearchSubmit = useCallback(
    (event: React.FormEvent<HTMLFormElement>): void => {
      event.preventDefault();
      runSearch(query);
    },
    [query, runSearch],
  );

  /** Передаём запрос ИИ-помощнику — он подберёт позиции и подключит оператора. */
  const onAskAssistant = useCallback((): void => {
    const trimmed = query.trim();

    setIsSuggestOpen(false);

    if (!trimmed) {
      openAssistant();

      return;
    }

    rememberQuery(trimmed);
    openAssistant();
    sendToAssistant(trimmed);
  }, [openAssistant, query, rememberQuery, sendToAssistant]);

  const onRecentSelect = useCallback(
    (value: string): void => {
      setQuery(value);
      runSearch(value);
    },
    [runSearch],
  );

  return {
    desktopPlaceholder: HEADER_SEARCH_PLACEHOLDER,
    hasSuggest:
      isSuggestOpen && (query.trim().length >= SUGGEST_MIN_LENGTH || recentQueries.length > 0),
    isSuggestPending,
    mobilePlaceholder: HEADER_MOBILE_SEARCH_PLACEHOLDER,
    onAskAssistant,
    onQueryChange,
    onRecentSelect,
    onSearchSubmit,
    onSuggestClose,
    onSuggestOpen,
    query,
    recentQueries,
    suggestions,
  };
};
