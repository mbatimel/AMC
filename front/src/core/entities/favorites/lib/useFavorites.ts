'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { $favoriteIds, favoritesHydrated, favoriteToggled } from '../model';

export const useFavorites = () => {
  const [favoriteIds, hydrate, toggleFavorite] = useUnit([
    $favoriteIds,
    favoritesHydrated,
    favoriteToggled,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    favoriteIds,
    isFavorite: (productId: string) => favoriteIds.includes(productId),
    toggleFavorite,
  };
};
