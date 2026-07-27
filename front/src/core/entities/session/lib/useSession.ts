'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { $userId, sessionHydrated } from '../model';

export const useSession = (): {
  isAuthenticated: boolean;
  userId: null | string;
} => {
  const [userId, hydrate] = useUnit([$userId, sessionHydrated]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    isAuthenticated: userId !== null,
    userId,
  };
};
