import { createEffect, createEvent, createStore, sample } from 'effector';

const STORAGE_KEY = 'amc_favorites';

const readFavorites = (): string[] => {
  if (typeof window === 'undefined') {
    return [];
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);

    if (!raw) {
      return [];
    }

    const parsed: unknown = JSON.parse(raw);

    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter((item): item is string => typeof item === 'string');
  } catch {
    return [];
  }
};

const writeFavorites = (ids: string[]): void => {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
};

export const favoritesHydrated = createEvent();
export const favoriteToggled = createEvent<string>();

const persistFavoritesFx = createEffect((ids: string[]) => {
  writeFavorites(ids);
});

export const $favoriteIds = createStore<string[]>([])
  .on(favoritesHydrated, () => readFavorites())
  .on(favoriteToggled, (state, productId) =>
    state.includes(productId) ? state.filter((id) => id !== productId) : [...state, productId],
  );

sample({
  clock: favoriteToggled,
  source: $favoriteIds,
  target: persistFavoritesFx,
});
