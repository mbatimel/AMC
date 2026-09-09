import { createEffect, createEvent, createStore, sample } from 'effector';

import { $userId, sessionEnded } from '@/core/entities/session';
import {
  addFavoriteRequest,
  deleteFavoriteRequest,
  listFavoritesRequest,
} from '@/core/shared/api/favorites';

const STORAGE_KEY = 'amc_favorites';

const isUserId = (userId: null | string): userId is string =>
  typeof userId === 'string' && userId.length > 0;

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

const favoritesLocalLoaded = createEvent<string[]>();

const persistFavoritesFx = createEffect((ids: string[]) => {
  writeFavorites(ids);
});

const fetchRemoteFavoritesFx = createEffect((userId: string) => listFavoritesRequest(userId));

const syncFavoriteFx = createEffect(
  async ({
    productID,
    shouldAdd,
    userId,
  }: {
    productID: string;
    shouldAdd: boolean;
    userId: string;
  }) => {
    if (shouldAdd) {
      await addFavoriteRequest(userId, productID);

      return;
    }

    await deleteFavoriteRequest(userId, productID);
  },
);

/** Локальный hydrate — один раз на вкладку (useFavorites может монтироваться многократно). */
const $isLocalHydrated = createStore(false)
  .on(favoritesLocalLoaded, () => true)
  .reset(sessionEnded);

/** Для какого userId уже забрали remote-список. */
const $remoteFetchedForUserId = createStore<null | string>(null)
  .on(fetchRemoteFavoritesFx.done, (_, { params }) => params)
  .reset(sessionEnded);

export const $favoriteIds = createStore<string[]>([])
  .on(favoritesLocalLoaded, (_, ids) => ids)
  .on(fetchRemoteFavoritesFx.doneData, (_, ids) => ids)
  .on(favoriteToggled, (state, productId) =>
    state.includes(productId) ? state.filter((id) => id !== productId) : [...state, productId],
  )
  .reset(sessionEnded);

sample({
  clock: favoriteToggled,
  source: $favoriteIds,
  target: persistFavoritesFx,
});

sample({
  clock: fetchRemoteFavoritesFx.doneData,
  target: persistFavoritesFx,
});

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> fn -> target */

sample({
  clock: favoritesHydrated,
  source: $isLocalHydrated,
  filter: (hydrated) => !hydrated,
  fn: () => readFavorites(),
  target: favoritesLocalLoaded,
});

sample({
  clock: [favoritesLocalLoaded, $userId],
  source: {
    fetchedFor: $remoteFetchedForUserId,
    pending: fetchRemoteFavoritesFx.pending,
    userId: $userId,
  },
  filter: ({ fetchedFor, pending, userId }) =>
    !pending && isUserId(userId) && userId !== fetchedFor,
  fn: ({ userId }) => userId as string,
  target: fetchRemoteFavoritesFx,
});

sample({
  clock: favoriteToggled,
  source: {
    ids: $favoriteIds,
    userId: $userId,
  },
  filter: ({ userId }) => isUserId(userId),
  fn: ({ ids, userId }, productID) => ({
    productID,
    shouldAdd: ids.includes(productID),
    userId: userId as string,
  }),
  target: syncFavoriteFx,
});

/* eslint-enable perfectionist/sort-objects */
