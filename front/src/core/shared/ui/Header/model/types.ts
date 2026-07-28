import type { RefObject } from 'react';

import type { useCity } from '@/core/entities/city';

import type { useHeaderNav } from './useHeaderNav';
import type { useHeaderSearch } from './useHeaderSearch';
import type { useMobileMenu } from './useMobileMenu';

export type HeaderAccountLink = {
  href: string;
  isAuthenticated: boolean;
  label: string;
};

export type HeaderMobileMenuProps = {
  account: HeaderAccountLink;
  closeMenu: () => void;
  drawerId: string;
  drawerRef: RefObject<HTMLElement | null>;
  isOpen: boolean;
  nav: UseHeaderNavResult;
  search: UseHeaderSearchResult;
};

export type HeaderMobileViewProps = HeaderViewProps & {
  mobileMenu: UseMobileMenuResult;
};

export type HeaderViewProps = {
  account: HeaderAccountLink;
  city: UseCityResult;
  nav: UseHeaderNavResult;
  search: UseHeaderSearchResult;
};

export type UseCityResult = ReturnType<typeof useCity>;

export type UseHeaderNavResult = ReturnType<typeof useHeaderNav>;

export type UseHeaderSearchResult = ReturnType<typeof useHeaderSearch>;

export type UseMobileMenuResult = ReturnType<typeof useMobileMenu>;
