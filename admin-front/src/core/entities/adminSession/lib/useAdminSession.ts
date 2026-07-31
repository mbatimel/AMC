'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { $adminUserId, $isAdminSessionHydrated, adminSessionEnded, adminSessionHydrated } from '../model';

export const useAdminSession = (): {
  adminUserId: null | string;
  isAuthenticated: boolean;
  isHydrated: boolean;
  logout: () => void;
} => {
  const [adminUserId, isHydrated, hydrate, logout] = useUnit([
    $adminUserId,
    $isAdminSessionHydrated,
    adminSessionHydrated,
    adminSessionEnded,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    adminUserId,
    isAuthenticated: adminUserId !== null,
    isHydrated,
    logout,
  };
};
