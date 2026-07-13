'use client';

import { useHeaderNav } from './useHeaderNav';
import { useHeaderSearch } from './useHeaderSearch';
import { useMobileMenu } from './useMobileMenu';

export const useHeader = () => {
  const nav = useHeaderNav();
  const search = useHeaderSearch();
  const mobileMenu = useMobileMenu();

  return {
    mobileMenu,
    nav,
    search,
  };
};
