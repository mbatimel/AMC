'use client';

import { usePathname } from 'next/navigation';
import { useCallback } from 'react';

import { HEADER_CABINET_ITEMS, HEADER_NAV_ITEMS } from '../constants';

export const useHeaderNav = () => {
  const pathname = usePathname();

  const getIsActive = useCallback((href: string): boolean => href === pathname, [pathname]);

  return {
    cabinetItems: HEADER_CABINET_ITEMS,
    getIsActive,
    navItems: HEADER_NAV_ITEMS,
    pathname,
  };
};
