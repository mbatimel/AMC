'use client';

import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import {
  $adminUserId,
  $isAdminSessionHydrated,
  adminLogoutFx,
  adminSessionHydrated,
} from '../model';

export const useAdminSession = (): {
  adminUserId: null | string;
  isAuthenticated: boolean;
  isHydrated: boolean;
  logout: () => Promise<void>;
} => {
  const [adminUserId, isHydrated, hydrate, logoutFx] = useUnit([
    $adminUserId,
    $isAdminSessionHydrated,
    adminSessionHydrated,
    adminLogoutFx,
  ]);

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return {
    adminUserId,
    isAuthenticated: adminUserId !== null,
    isHydrated,
    logout: async () => {
      await logoutFx(adminUserId);
    },
  };
};
