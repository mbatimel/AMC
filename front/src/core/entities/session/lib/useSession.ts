'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { $isSessionHydrated, $userId, sessionHydrated } from '../model';

export const useSession = (): {
  isAuthenticated: boolean;
  isHydrated: boolean;
  userId: null | string;
} => {
  const [userId, isHydrated, hydrate] = useUnit([$userId, $isSessionHydrated, sessionHydrated]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    isAuthenticated: userId !== null,
    isHydrated,
    userId,
  };
};
